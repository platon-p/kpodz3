package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/platon-p/kpodz3/orders/application/services"
	pb "github.com/platon-p/kpodz3/proto"
	"google.golang.org/protobuf/proto"

	"github.com/rabbitmq/amqp091-go"
)

type EventPublishWorker struct {
	ch        *amqp091.Channel
	queue     *amqp091.Queue
	orderRepo services.OrderRepo
}

func NewEventPublishWorker(orderRepo services.OrderRepo, ch *amqp091.Channel, queue *amqp091.Queue) *EventPublishWorker {
	return &EventPublishWorker{orderRepo: orderRepo, ch: ch, queue: queue}
}

func (w *EventPublishWorker) publishOrderCreated(evt *pb.Event) error {
	return w.publishEvent(evt)
}

func (w *EventPublishWorker) publishEvent(event *pb.Event) error {
	// serialize
	bytes, err := proto.Marshal(event)
	if err != nil {
		return err
	}
	err = w.ch.Publish("", w.queue.Name, false, false, amqp091.Publishing{
		Body:        bytes,
		ContentType: "application/protobuf",
	})
	return err
}

func (w *EventPublishWorker) Run(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			evt := new(pb.Event)
			err := w.orderRepo.PopEvent(ctx, "order_created", evt)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = w.publishOrderCreated(evt)
			fmt.Println(err)
		}
	}
}
