package services

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/platon-p/kpodz3/proto"
)

var ErrInsufficientBalance = errors.New("insufficient balance")
var ErrAccountNotFound = fmt.Errorf("account not found")

type AccountRepo interface {
	CreateAccount(ctx context.Context, userId int) error
	TopUp(ctx context.Context, userId int, amount int) error
	GetBalance(ctx context.Context, userId int) (int, error)

	// expect it is safe
	Withdraw(ctx context.Context, userId int, amount int) error

	PushEvent(ctx context.Context, key string, event *pb.Event) error
	PopEvent(ctx context.Context, key string, dest *pb.Event) error
}

type AccountService interface {
	CreateAccount(ctx context.Context, userId int) error
	TopUp(ctx context.Context, userId int, amount int) error
	GetBalance(ctx context.Context, userId int) (int, error)

	Withdraw(ctx context.Context, userId int, amount int) error
}

type AccountServiceImpl struct {
	repo AccountRepo
}

func NewAccountService(repo AccountRepo) *AccountServiceImpl {
	return &AccountServiceImpl{repo: repo}
}

func (s *AccountServiceImpl) CreateAccount(ctx context.Context, userId int) error {
	return s.repo.CreateAccount(ctx, userId)
}

func (s *AccountServiceImpl) TopUp(ctx context.Context, userId int, amount int) error {
	return s.repo.TopUp(ctx, userId, amount)
}

func (s *AccountServiceImpl) GetBalance(ctx context.Context, userId int) (int, error) {
	balance, err := s.repo.GetBalance(ctx, userId)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (s *AccountServiceImpl) Withdraw(ctx context.Context, userId int, amount int) error {
	return s.repo.Withdraw(ctx, userId, amount)
}
