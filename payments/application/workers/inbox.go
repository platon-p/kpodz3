package workers

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/platon-p/kpodz3/payments/application/services"
	pb "github.com/platon-p/kpodz3/proto"
)

var ErrUnknownEventType = fmt.Errorf("unknown event type")

type InboxWorker struct {
	logger *zap.Logger

	ch    *amqp091.Channel
	queue *amqp091.Queue

	accountRepo    services.AccountRepo
	accountService services.AccountService
}

func NewInboxWorker(
	logger *zap.Logger,
	ch *amqp091.Channel,
	queue *amqp091.Queue,
	accountRepo services.AccountRepo,
	accountService services.AccountService,
) *InboxWorker {
	return &InboxWorker{logger: logger, ch: ch, queue: queue, accountRepo: accountRepo, accountService: accountService}
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
	if err != nil {
		return fmt.Errorf("handle event: %w", err)
	}
	return nil
}

func (w *InboxWorker) handleEvent(msg *amqp091.Delivery, e *pb.Event) error {
	w.logger.Info("handle event", zap.String("type", e.Type.String()))
	switch e.Type {
	case pb.Event_TypeOrderCreated:
		return w.handleOrderCreated(msg, e)
	case pb.Event_TypeOrderSuccess, pb.Event_TypeOrderFailed:
		// ignore
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownEventType, e.Type)
	}
}

func (w *InboxWorker) handleOrderCreated(msg *amqp091.Delivery, e *pb.Event) error {
	err := w.accountRepo.PushEvent(context.Background(), "inbox", e)
	if err != nil {
		msg.Nack(false, true)
		return fmt.Errorf("push event to inbox: %w", err)
	}
	msg.Ack(false)
	return nil
}
