package infra

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/platon-p/kpodz3/payments/application/services"
	"github.com/redis/go-redis/v9"
)

var ErrAccountNotFound = fmt.Errorf("account not found")

var _ services.AccountRepo = (*RedisAccountRepo)(nil)

type RedisAccountRepo struct {
	client *redis.Client
}

func (r *RedisAccountRepo) CreateAccount(ctx context.Context, userId int) error {
	// check if the account already exists
	key := r.buildAccountKey(userId)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil // Account already exists, no need to create it
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
		return ErrAccountNotFound
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
		return 0, ErrAccountNotFound // Account does not exist
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

func (r *RedisAccountRepo) buildAccountKey(userId int) string {
	return fmt.Sprintf("account:%d", userId)
}
