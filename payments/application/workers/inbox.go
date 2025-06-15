package workers

import (
	"context"
	"errors"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	"github.com/platon-p/kpodz3/payments/application/services"
	pb "github.com/platon-p/kpodz3/proto"
)

var ErrUnknownEventType = fmt.Errorf("unknown event type")

type InboxWorker struct {
	ch    *amqp091.Channel
	queue *amqp091.Queue

	accountRepo    services.AccountRepo
	accountService services.AccountService
}

func NewInboxWorker(ch *amqp091.Channel, queue *amqp091.Queue, accountRepo services.AccountRepo, accountService services.AccountService) *InboxWorker {
	return &InboxWorker{ch: ch, queue: queue, accountRepo: accountRepo, accountService: accountService}
}

func (w *InboxWorker) Run(ctx context.Context) {
	d, err := w.ch.Consume(w.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-d:
			go func() {
				err := w.handleMessage(&msg)
				if err != nil {
					fmt.Println(err)
				}
			}()

		}
	}
}

func (w *InboxWorker) handleMessage(msg *amqp091.Delivery) error {
	var e pb.Event
	err := proto.Unmarshal(msg.Body, &e)
	if err != nil {
		return fmt.Errorf("message to event unmarshal: %w", err)
	}

	err = w.handleEvent(msg, &e)
	if errors.Is(err, ErrUnknownEventType) { // TODO:
		return err
	}
	if err != nil {
		return fmt.Errorf("handle event: %w", err)
	}
	return nil
}

func (w *InboxWorker) handleEvent(msg *amqp091.Delivery, e *pb.Event) error {
	switch e.Type {
	case pb.Event_TypeOrderCreated:
		return w.handleOrderCreated(msg, e)
	default:
		return ErrUnknownEventType
	}
}

func (w *InboxWorker) handleOrderCreated(msg *amqp091.Delivery, e *pb.Event) error {
	evt := e.GetOrderCreated()
	err := w.accountService.Withdraw(context.Background(), int(evt.UserId), int(evt.Amount))

	msg.Ack(false)
	fmt.Println("ack")

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
