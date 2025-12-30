package mysql

import (
	"context"
	"database/sql"

	"github.com/atakannerturk/go-backend-path/internal/domain"
	"github.com/atakannerturk/go-backend-path/internal/repository"
	"github.com/atakannerturk/go-backend-path/pkg/errors"
)

type balanceRepository struct {
	db *sql.DB
}

func NewBalanceRepository(db *sql.DB) repository.BalanceRepository {
	return &balanceRepository{db: db}
}

func (r *balanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	query := `
		INSERT INTO balances (user_id, amount, last_updated_at)
		VALUES (?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		balance.UserID,
		balance.Amount,
		balance.LastUpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to create balance")
	}

	return nil
}

func (r *balanceRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	query := `
		SELECT user_id, amount, last_updated_at
		FROM balances
		WHERE user_id = ?
	`

	balance := &domain.Balance{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.Amount,
		&balance.LastUpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to get balance by user ID")
	}

	return balance, nil
}

func (r *balanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	query := `
		UPDATE balances
		SET amount = ?, last_updated_at = ?
		WHERE user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		balance.Amount,
		balance.LastUpdatedAt,
		balance.UserID,
	)
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to update balance")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *balanceRepository) Credit(ctx context.Context, userID int64, amount float64) error {
	query := `
		UPDATE balances
		SET amount = amount + ?, last_updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to credit balance")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *balanceRepository) Debit(ctx context.Context, userID int64, amount float64) error {
	query := `
		UPDATE balances
		SET amount = amount - ?, last_updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND amount >= ?
	`

	result, err := r.db.ExecContext(ctx, query, amount, userID, amount)
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to debit balance")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to get rows affected")
	}

	if rowsAffected == 0 {
		balance, err := r.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if balance.Amount < amount {
			return errors.ErrInsufficientFunds
		}
		return errors.ErrNotFound
	}

	return nil
}

func (r *balanceRepository) GetMultiple(ctx context.Context, userIDs []int64) ([]*domain.Balance, error) {
	if len(userIDs) == 0 {
		return []*domain.Balance{}, nil
	}

	query := `
		SELECT user_id, amount, last_updated_at
		FROM balances
		WHERE user_id IN (?)
	`

	rows, err := r.db.QueryContext(ctx, query, userIDs)
	if err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to get multiple balances")
	}
	defer rows.Close()

	var balances []*domain.Balance
	for rows.Next() {
		balance := &domain.Balance{}
		if err := rows.Scan(
			&balance.UserID,
			&balance.Amount,
			&balance.LastUpdatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to scan balance")
		}
		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to iterate balances")
	}

	return balances, nil
}
