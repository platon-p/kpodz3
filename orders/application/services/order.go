package services

import (
	"context"

	"github.com/platon-p/kpodz3/orders/domain"
)

type OrderRepo interface {
	Create(ctx context.Context, order domain.Order) error
	GetAll(ctx context.Context) ([]domain.Order, error)
	Get(ctx context.Context, name string) (domain.Order, error)
}

type OrderService interface {
	CreateOrder(ctx context.Context, userId int, title string, amount int) (domain.Order, error)
	GetAllOrders(ctx context.Context) ([]domain.Order, error)
	GetOrder(ctx context.Context, name string) (domain.Order, error)
}

type OrderServiceImpl struct {
	repo                   OrderRepo
	onOrderCreatedHandlers []func(domain.Order)
}

func NewOrderService(repo OrderRepo) *OrderServiceImpl {
	return &OrderServiceImpl{repo: repo, onOrderCreatedHandlers: make([]func(domain.Order), 0)}
}

func (s *OrderServiceImpl) SubscribeOrderCreated(handler func(domain.Order)) {
	s.onOrderCreatedHandlers = append(s.onOrderCreatedHandlers, handler)
}

func (s *OrderServiceImpl) onOrderCreated(order domain.Order) {
	for _, handler := range s.onOrderCreatedHandlers {
		handler(order)
	}
}

func (s *OrderServiceImpl) CreateOrder(ctx context.Context, userId int, title string, amount int) (domain.Order, error) {
	order := domain.Order{
		UserId: userId,
		Name:   title,
		Amount: amount,
	}
	err := s.repo.Create(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}
	s.onOrderCreated(order)
	return order, nil
}
