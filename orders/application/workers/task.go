package workers

import (
	"context"
	"fmt"

	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/domain"
	pb "github.com/platon-p/kpodz3/proto"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type TaskWorker struct {
	ch           *amqp091.Channel
	queue        *amqp091.Queue
	orderService services.OrderService
	logger       *zap.Logger
}

func NewTaskWorker(logger *zap.Logger, orderService services.OrderService, ch *amqp091.Channel, queue *amqp091.Queue) *TaskWorker {
	return &TaskWorker{
		logger:       logger,
		orderService: orderService,
		ch:           ch,
		queue:        queue,
	}
}

func (w *TaskWorker) Run(ctx context.Context) {
	msg, err := w.ch.Consume(w.queue.Name, "", false, false, false, false, nil)
	if err != nil {
		w.logger.Error("failed to start consuming messages", zap.Error(err))
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-msg:
			go func() {
				if err := w.handleMessage(&m); err != nil {
					w.logger.Error("failed to handle message", zap.Error(err))
				}
			}()
		}
	}
}

func (w *TaskWorker) handleMessage(msg *amqp091.Delivery) error {
	var event pb.Event
	if err := proto.Unmarshal(msg.Body, &event); err != nil {
		return err
	}
	w.logger.Info("received event", zap.Stringer("event_type", event.Type))
	switch event.Type {
	case pb.Event_TypeOrderCreated:
		// ignore
		return nil
	case pb.Event_TypeOrderSuccess:
		return w.handleOrderSuccess(msg, &event)
	case pb.Event_TypeOrderFailed:
		return w.handleOrderFailed(msg, &event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (w *TaskWorker) handleOrderSuccess(msg *amqp091.Delivery, event *pb.Event) error {
	e := event.GetOrderSuccess()
	err := w.orderService.SetOrderStatus(context.Background(), e.Name, domain.OrderStatusSuccess)

	if err != nil {
		msg.Nack(false, true)
		w.logger.Info("nack order failed", zap.String("order_name", e.Name), zap.Error(err))
		return fmt.Errorf("failed to handle event: %w", err)
	}
	msg.Ack(false)
	w.logger.Info("ack order success", zap.String("order_name", e.Name))
	return nil
}

func (w *TaskWorker) handleOrderFailed(msg *amqp091.Delivery, event *pb.Event) error {
	e := event.GetOrderFailed()
	err := w.orderService.SetOrderStatus(context.Background(), e.Name, domain.OrderStatusFailed)
	if err != nil {
		msg.Nack(false, true)
		w.logger.Info("nack order failed", zap.String("order_name", e.Name), zap.Error(err))
		return fmt.Errorf("failed to handle event: %w", err)
	}
	msg.Ack(false)
	w.logger.Info("ack order failed", zap.String("order_name", e.Name), zap.Stringer("reason", e.Reason))
	return nil
}
