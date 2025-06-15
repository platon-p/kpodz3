package workers

import (
	"context"
	"fmt"

	"github.com/platon-p/kpodz3/orders/domain"
	"github.com/rabbitmq/amqp091-go"
)

type EventPublishWorker struct {
	mq *amqp091.Channel
}

func NewEventPublishWorker(mq *amqp091.Channel) *EventPublishWorker {
	return &EventPublishWorker{mq: mq}
}

func (w *EventPublishWorker) PublishOrderCreated(order domain.Order) {
	w.mq.Publish("exchange", buildOrderCreatedEventKey(order), false, false, amqp091.Publishing{
		Body:        []byte(order.Name),
		ContentType: "text/plain",
	})
}

func (w *EventPublishWorker) Run(ctx context.Context) {
	defer w.mq.Close()
	<-ctx.Done()
}

func buildOrderCreatedEventKey(order domain.Order) string {
	return fmt.Sprintf("order.created.%s", order.Name)
}
