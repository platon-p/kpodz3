package services

import (
	"context"
	"fmt"

	"github.com/platon-p/kpodz3/orders/domain"
	pb "github.com/platon-p/kpodz3/proto"
)

type TXable[T any] interface {
	TxBegin(ctx context.Context) T
	TxCommit(ctx context.Context) error
	TxRollback(ctx context.Context) error
}

type OrderRepo interface {
	TXable[OrderRepo]

	Create(ctx context.Context, order domain.Order) error
	GetAll(ctx context.Context) ([]domain.Order, error)
	Get(ctx context.Context, name string) (domain.Order, error)

	PushEvent(ctx context.Context, key string, event *pb.Event) error
	PopEvent(ctx context.Context, key string, dest *pb.Event) error
}

type OrderService interface {
	CreateOrder(ctx context.Context, userId int, title string, amount int) (domain.Order, error)
	GetAllOrders(ctx context.Context) ([]domain.Order, error)
	GetOrder(ctx context.Context, name string) (domain.Order, error)
}

type OrderServiceImpl struct {
	repo OrderRepo
}

func NewOrderService(repo OrderRepo) *OrderServiceImpl {
	return &OrderServiceImpl{repo: repo}
}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, userId int, title string, amount int) (domain.Order, error) {
	order := domain.Order{
		UserId: userId,
		Name:   title,
		Amount: amount,
	}
	repo := s.repo.TxBegin(ctx)
	err := func() error {
		if err := repo.Create(ctx, order); err != nil {
			return err
		}
		evt := &pb.Event{
			Type: pb.Event_TypeOrderCreated,
			Data: &pb.Event_OrderCreated{
				OrderCreated: &pb.Event_Order_Created{
					Name:   order.Name,
					UserId: int32(order.UserId),
					Amount: int32(order.Amount),
				},
			},
		}
		if err := repo.PushEvent(ctx, "order_created", evt); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		if rollbackErr := repo.TxRollback(ctx); rollbackErr != nil {
			return domain.Order{}, rollbackErr
		}
		return domain.Order{}, fmt.Errorf("failed to create order: %w", err)
	}
	if err := repo.TxCommit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return order, nil
}
