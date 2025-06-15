package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/platon-p/kpodz3/orders/application/services"
	pb "github.com/platon-p/kpodz3/proto"
	"google.golang.org/protobuf/proto"

	"github.com/rabbitmq/amqp091-go"
)

type OutboxWorker struct {
	ch        *amqp091.Channel
	queue     *amqp091.Queue
	orderRepo services.OrderRepo
}

func NewOutboxWorker(orderRepo services.OrderRepo, ch *amqp091.Channel, queue *amqp091.Queue) *OutboxWorker {
	return &OutboxWorker{orderRepo: orderRepo, ch: ch, queue: queue}
}

func (w *OutboxWorker) publishEvent(event *pb.Event) error {
	// serialize
	bytes, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = w.ch.Publish("", w.queue.Name, false, false, amqp091.Publishing{
		Body:        bytes,
		ContentType: "application/protobuf",
	})
	fmt.Println("published event:", event.Type, err)
	return err
}

func (w *OutboxWorker) Run(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			evt, err := w.orderRepo.PopEvent(ctx, "outbox")
			if errors.Is(err, services.ErrNoEvents) {
				continue
			}
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = w.publishEvent(evt)
			if err != nil {
				fmt.Println("failed to publish event:", err)
				continue
			}
		}
	}
}
