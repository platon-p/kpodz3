package services

import "context"

type AccountRepo interface {
	CreateAccount(ctx context.Context, userId int) error
	TopUp(ctx context.Context, userId int, amount int) error
	GetBalance(ctx context.Context, userId int) (int, error)
}

type AccountService interface {
	CreateAccount(ctx context.Context, userId int) error
	TopUp(ctx context.Context, userId int, amount int) error
	GetBalance(ctx context.Context, userId int) (int, error)
}

type AccountServiceImpl struct {
	repo AccountRepo
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
