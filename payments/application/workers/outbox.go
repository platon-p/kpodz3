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

type OutboxWorker struct {
	accountRepo services.AccountRepo
	ch          *amqp091.Channel
	queue       *amqp091.Queue
}

func NewOutboxWorker(accountRepo services.AccountRepo, ch *amqp091.Channel, queue *amqp091.Queue) *OutboxWorker {
	return &OutboxWorker{
		accountRepo: accountRepo,
		ch:          ch,
		queue:       queue,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			err := w.check(ctx)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (w *OutboxWorker) check(ctx context.Context) error {
	evt := new(pb.Event)
	err := w.accountRepo.PopEvent(ctx, "outbox", evt)
	if errors.Is(err, services.ErrNoEvents) {
		return nil
	}
	if err != nil {
		return err
	}
	err = func() error {
		data, err := proto.Marshal(evt)
		if err != nil {
			return err
		}
		return w.ch.Publish("", w.queue.Name, false, false, amqp091.Publishing{
			ContentType: "application/protobuf",
			Body:        data,
		})
	}()
	if err != nil {
		w.accountRepo.UnpopEvent(ctx, "outbox", evt)
	}
	return nil
}
