package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/domain"
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

func (w *EventPublishWorker) publishOrderCreated(order domain.Order) error {
	err := w.ch.Publish("", w.queue.Name, false, false, amqp091.Publishing{
		Body:        []byte(order.Name),
		ContentType: "text/plain",
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
			order := domain.Order{}
			tx := w.orderRepo.TxBegin(ctx)
			err := func() error {
				err := tx.PopEvent(ctx, "order_created", &order)
				if err != nil {
					return err
				}
				if err := w.publishOrderCreated(order); err != nil {
					return err
				}
				return nil
			}()
			// if err != nil {
			// 	if rollbackErr := tx.TxRollback(ctx); rollbackErr != nil {
			// 		fmt.Printf("failed to rollback transaction: %v\n", rollbackErr)
			// 	}
			// 	fmt.Printf("failed to process order: %v\n", err)
			// 	continue
			// }
			// if err := tx.TxCommit(ctx); err != nil {
			// 	fmt.Printf("failed to commit transaction: %v\n", err)
			// 	continue
			// }

			fmt.Println(err)
		}
	}
}
