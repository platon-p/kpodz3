package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/domain"
	"github.com/redis/go-redis/v9"
)

var _ services.OrderRepo = (*RedisOrderRepo)(nil)

type RedisOrderRepo struct {
	client redis.Cmdable
}

func NewRedisOrderRepo(client redis.Cmdable) *RedisOrderRepo {
	return &RedisOrderRepo{client: client}
}

func (r *RedisOrderRepo) Create(ctx context.Context, order domain.Order) error {
	key := fmt.Sprintf("order:%s", order.Name)
	value, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}
	if err := r.client.Set(ctx, key, value, 0).Err(); err != nil {
		return fmt.Errorf("failed to set order in redis: %w", err)
	}
	return nil
}

func (r *RedisOrderRepo) GetAll(ctx context.Context) ([]domain.Order, error) {
	entries, err := r.client.Keys(ctx, "order:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}
	var orders []domain.Order
	for _, key := range entries {
		order, err := r.getOrderByKey(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get order by key %s: %w", key, err)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *RedisOrderRepo) Get(ctx context.Context, name string) (domain.Order, error) {
	key := fmt.Sprintf("order:%s", name)
	order, err := r.getOrderByKey(ctx, key)
	if err != nil {
		return domain.Order{}, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (r *RedisOrderRepo) getOrderByKey(ctx context.Context, key string) (domain.Order, error) {
	var order domain.Order

	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return order, fmt.Errorf("key %s does not exist", key)
	}
	if err != nil {
		return order, fmt.Errorf("failed to get order: %w", err)
	}
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return order, fmt.Errorf("failed to unmarshal order: %w", err)
	}
	return order, nil
}

func (r *RedisOrderRepo) PushEvent(ctx context.Context, key string, event any) error {
	key = fmt.Sprintf("event:%s", key)
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return r.client.LPush(ctx, key, value).Err()
}

func (r *RedisOrderRepo) PopEvent(ctx context.Context, key string, dest any) error {
	key = fmt.Sprintf("event:%s", key)
	// check length
	if length, err := r.client.LLen(ctx, key).Result(); err != nil {
		return fmt.Errorf("failed to get length of queue %s: %w", key, err)
	} else if length == 0 {
		fmt.Println(12)
		return fmt.Errorf("no events in queue %s", key)
	}
	value, err := r.client.RPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("no events in queue %s", key)
	}
	if err != nil {
		return fmt.Errorf("failed to pop event: %w", err)
	}
	if err := json.Unmarshal([]byte(value), dest); err != nil {
		fmt.Println(key)
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return nil
}

// NOTE: After end of transaction, use previous client
func (r *RedisOrderRepo) TxBegin(ctx context.Context) services.OrderRepo {
	return NewRedisOrderRepo(r.client.TxPipeline())
}

func (r *RedisOrderRepo) TxCommit(ctx context.Context) error {
	if tx, ok := r.client.(redis.Pipeliner); !ok {
		return fmt.Errorf("client does not support transactions")
	} else {
		_, err := tx.TxPipeline().Exec(ctx)
		return err
	}
}

func (r *RedisOrderRepo) TxRollback(ctx context.Context) error {
	if tx, ok := r.client.(redis.Pipeliner); !ok {
		return fmt.Errorf("client does not support transactions")
	} else {
		tx.Discard()
		return nil
	}
}
