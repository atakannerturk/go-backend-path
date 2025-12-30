package repository

import (
	"context"

	"github.com/atakannerturk/go-backend-path/internal/domain"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	GetByID(ctx context.Context, id int64) (*domain.Transaction, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error)
	UpdateStatus(ctx context.Context, id int64, status domain.TransactionStatus) error
	List(ctx context.Context, limit, offset int) ([]*domain.Transaction, error)
	GetPendingTransactions(ctx context.Context, limit int) ([]*domain.Transaction, error)
}
