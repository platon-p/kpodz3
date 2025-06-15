package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/platon-p/kpodz3/payments/application/services"
	pb "github.com/platon-p/kpodz3/proto"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type TaskWorker struct {
	accountRepo    services.AccountRepo
	accountService services.AccountService

	ch    *amqp091.Channel
	queue *amqp091.Queue
}

func NewTaskWorker(accountService services.AccountService, accountRepo services.AccountRepo, ch *amqp091.Channel, queue *amqp091.Queue) *TaskWorker {
	return &TaskWorker{
		accountRepo:    accountRepo,
		accountService: accountService,
		ch:             ch,
		queue:          queue,
	}
}

func (w *TaskWorker) Run(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := w.periodicTask(ctx); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (w *TaskWorker) periodicTask(ctx context.Context) error {
	var event pb.Event
	err := w.accountRepo.PopEvent(ctx, "inbox", &event)
	if errors.Is(err, services.ErrNoEvents) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("pop event: %w", err)
	}

	evt := event.GetOrderCreated()
	err = w.accountService.Withdraw(context.Background(), int(evt.UserId), int(evt.Amount))

	if err != nil && !errors.Is(err, services.ErrInsufficientBalance) && !errors.Is(err, services.ErrAccountNotFound) {
		w.accountRepo.UnpopEvent(ctx, "inbox", &event)
		return fmt.Errorf("withdraw: %w", err)
	}

	if err == nil {
		evtSuccess := &pb.Event{
			Type: pb.Event_TypeOrderSuccess,
			Data: &pb.Event_OrderSuccess{
				OrderSuccess: &pb.Event_Order_Success{
					Name:   evt.Name,
					UserId: evt.UserId,
				},
			},
		}
		serialized, _ := proto.Marshal(evtSuccess)
		return w.ch.Publish("", w.queue.Name, false, false, amqp091.Publishing{
			ContentType: "application/protobuf",
			Body:        serialized,
		})
	}

	reason := pb.Event_Order_Failed_ReasonUnknown
	if errors.Is(err, services.ErrInsufficientBalance) {
		reason = pb.Event_Order_Failed_ReasonInsufficientBalance
	} else if errors.Is(err, services.ErrAccountNotFound) {
		reason = pb.Event_Order_Failed_ReasonAccountNotFound
	}

	evtFailed := &pb.Event{
		Type: pb.Event_TypeOrderFailed,
		Data: &pb.Event_OrderFailed{
			OrderFailed: &pb.Event_Order_Failed{
				Name:   evt.Name,
				UserId: evt.UserId,
				Reason: reason,
			},
		},
	}
	serialized, _ := proto.Marshal(evtFailed)

	return w.ch.Publish("", w.queue.Name, false, false, amqp091.Publishing{
		ContentType: "application/protobuf",
		Body:        serialized,
	})
}
