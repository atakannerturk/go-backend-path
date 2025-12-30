package repository

import (
	"context"

	"github.com/atakannerturk/go-backend-path/internal/domain"
)

type BalanceRepository interface {
	Create(ctx context.Context, balance *domain.Balance) error
	GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error)
	Update(ctx context.Context, balance *domain.Balance) error
	Credit(ctx context.Context, userID int64, amount float64) error
	Debit(ctx context.Context, userID int64, amount float64) error
	GetMultiple(ctx context.Context, userIDs []int64) ([]*domain.Balance, error)
}
