package infra

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/platon-p/kpodz3/payments/application/services"
	"github.com/platon-p/kpodz3/payments/domain"
	pb "github.com/platon-p/kpodz3/proto"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

var _ services.AccountRepo = (*RedisAccountRepo)(nil)

type RedisAccountRepo struct {
	client *redis.Client
}

func NewRedisAccountRepo(client *redis.Client) *RedisAccountRepo {
	return &RedisAccountRepo{client: client}
}

func (r *RedisAccountRepo) CreateAccount(ctx context.Context, userId int) error {
	// check if the account already exists
	key := r.buildAccountKey(userId)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return services.ErrAccountAlreadyExists
	}

	err = r.client.Set(ctx, key, 0, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to create account for user %d: %w", userId, err)
	}

	return nil
}

func (r *RedisAccountRepo) TopUp(ctx context.Context, userId int, amount int) error {
	key := r.buildAccountKey(userId)
	// check if the account exists
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return services.ErrAccountNotFound
	}

	// Increment the account balance
	_, err = r.client.IncrBy(ctx, key, int64(amount)).Result()
	if err != nil {
		return fmt.Errorf("failed to top up account for user %d: %w", userId, err)
	}

	return nil
}

func (r *RedisAccountRepo) GetBalance(ctx context.Context, userId int) (int, error) {
	key := r.buildAccountKey(userId)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, services.ErrAccountNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get balance for user %d: %w", userId, err)
	}
	balance, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid balance value for user %d: %w", userId, err)
	}
	return balance, nil
}

func (r *RedisAccountRepo) GetAccount(ctx context.Context, userId int) (domain.Account, error) {
	var account domain.Account
	key := r.buildAccountKey(userId)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return account, services.ErrAccountNotFound // Account does not exist
	}
	if err != nil {
		return account, fmt.Errorf("failed to get account of user %d: %w", userId, err)
	}
	balance, err := strconv.Atoi(val)
	if err != nil {
		return account, fmt.Errorf("invalid balance value for user %d: %w", userId, err)
	}
	account = domain.Account{UserId: userId, Balance: balance}
	return account, nil
}

func (r *RedisAccountRepo) Withdraw(ctx context.Context, userId int, amount int) error {
	retries := 3
	key := r.buildAccountKey(userId)
	for range retries {
		err := r.client.Watch(ctx, func(tx *redis.Tx) error {
			balanceStr, err := tx.Get(ctx, key).Result()
			if errors.Is(err, redis.Nil) {
				return services.ErrAccountNotFound
			}
			if err != nil {
				return fmt.Errorf("failed to get balance for user %d: %w", userId, err)
			}
			balance, err := strconv.Atoi(balanceStr)
			if err != nil {
				return fmt.Errorf("invalid balance value for user %d: %w", userId, err)
			}
			if balance < amount {
				return services.ErrInsufficientBalance
			}
			_, err = tx.DecrBy(ctx, key, int64(amount)).Result()
			if err != nil {
				return fmt.Errorf("failed to withdraw from account for user %d: %w", userId, err)
			}
			return nil
		}, key)
		if errors.Is(err, redis.TxFailedErr) {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		return err
	}
	return fmt.Errorf("withdraw failed after %d retries", retries)
}

func (r *RedisAccountRepo) PushEvent(ctx context.Context, key string, event *pb.Event) error {
	serialized, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	return r.client.LPush(ctx, fmt.Sprintf("event:%s", key), serialized).Err()
}

func (r *RedisAccountRepo) PopEvent(ctx context.Context, key string, event *pb.Event) error {
	key = fmt.Sprintf("event:%s", key)
	serialized, err := r.client.LMove(ctx, key, fmt.Sprintf("%s:processing", key), "LEFT", "RIGHT").Result()
	if errors.Is(err, redis.Nil) {
		return services.ErrNoEvents
	}
	if err != nil {
		return err
	}
	if err := proto.Unmarshal([]byte(serialized), event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return nil
}

func (r *RedisAccountRepo) UnpopEvent(ctx context.Context, key string, event *pb.Event) error {
	serialized, _ := proto.Marshal(event)
	key = fmt.Sprintf("event:%s", key)
	tmp := fmt.Sprintf("%s:processing", key)
	tx := r.client.TxPipeline()
	err := tx.LRem(ctx, tmp, 1, serialized).Err()
	if err != nil {
		return fmt.Errorf("failed to remove event from processing queue: %w", err)
	}
	err = tx.LPush(ctx, key, serialized).Err()
	if err != nil {
		tx.Discard()
		return fmt.Errorf("failed to push event back to queue: %w", err)
	}
	_, err = tx.Exec(ctx)
	return err
}

func (r *RedisAccountRepo) buildAccountKey(userId int) string {
	return fmt.Sprintf("account:%d", userId)
}
