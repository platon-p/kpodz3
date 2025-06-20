package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/platon-p/kpodz3/orders/application/services"
	"github.com/platon-p/kpodz3/orders/domain"
	pb "github.com/platon-p/kpodz3/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

var _ services.OrderRepo = (*RedisOrderRepo)(nil)

type RedisOrderRepo struct {
	client redis.Cmdable
}

func NewRedisOrderRepo(client redis.Cmdable) *RedisOrderRepo {
	return &RedisOrderRepo{client: client}
}

func (r *RedisOrderRepo) Create(ctx context.Context, order domain.Order) error {
	return r.setOrder(ctx, order)
}

func (r *RedisOrderRepo) setOrder(ctx context.Context, order domain.Order) error {
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
	return r.getOrderByKey(ctx, key)
}

func (r *RedisOrderRepo) getOrderByKey(ctx context.Context, key string) (domain.Order, error) {
	var order domain.Order

	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return order, services.ErrNoOrder
	}
	if err != nil {
		return order, fmt.Errorf("failed to get order: %w", err)
	}
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return order, fmt.Errorf("failed to unmarshal order: %w", err)
	}
	return order, nil
}

func (r *RedisOrderRepo) PushEvent(ctx context.Context, key string, event *pb.Event) error {
	key = fmt.Sprintf("event:%s", key)
	value, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return r.client.LPush(ctx, key, value).Err()
}

func (r *RedisOrderRepo) PopEvent(ctx context.Context, key string) (*pb.Event, error) {
	key = fmt.Sprintf("event:%s", key)
	value, err := r.client.LMove(ctx, key, fmt.Sprintf("tmp:%s", key), "LEFT", "RIGHT").Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("%w: %s is nil", services.ErrNoEvents, key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to pop event: %w", err)
	}
	dest := new(pb.Event)
	if err := proto.Unmarshal([]byte(value), dest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return dest, nil
}

func (r *RedisOrderRepo) SetStatus(ctx context.Context, name string, status domain.OrderStatus) error {
	order, err := r.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	order.Status = status
	return r.setOrder(ctx, order)
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
