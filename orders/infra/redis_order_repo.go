package infra

import (
	"context"

	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/domain"
	"github.com/redis/go-redis/v9"
)

var _ services.OrderRepo = (*RedisOrderRepo)(nil)

type RedisOrderRepo struct {
	client *redis.Client
}

func NewRedisOrderRepo(client *redis.Client) *RedisOrderRepo {
	return &RedisOrderRepo{client: client}
}

func (r *RedisOrderRepo) Create(ctx context.Context, order domain.Order) error {
	return nil
}

func (r *RedisOrderRepo) GetAll(ctx context.Context) ([]domain.Order, error) {
	return nil, nil
}

func (r *RedisOrderRepo) Get(ctx context.Context, name string) (domain.Order, error) {
	return domain.Order{}, nil
}
