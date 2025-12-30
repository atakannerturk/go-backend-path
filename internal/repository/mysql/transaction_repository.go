package mysql

import (
	"context"
	"database/sql"

	"github.com/atakannerturk/go-backend-path/internal/domain"
	"github.com/atakannerturk/go-backend-path/internal/repository"
	"github.com/atakannerturk/go-backend-path/pkg/errors"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) repository.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (from_user_id, to_user_id, amount, type, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		tx.FromUserID,
		tx.ToUserID,
		tx.Amount,
		tx.Type,
		tx.Status,
		tx.CreatedAt,
		tx.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to create transaction")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to get last insert ID")
	}

	tx.ID = id
	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	query := `
		SELECT id, from_user_id, to_user_id, amount, type, status, created_at, updated_at
		FROM transactions
		WHERE id = ?
	`

	tx := &domain.Transaction{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID,
		&tx.FromUserID,
		&tx.ToUserID,
		&tx.Amount,
		&tx.Type,
		&tx.Status,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to get transaction by ID")
	}

	return tx, nil
}

func (r *transactionRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, from_user_id, to_user_id, amount, type, status, created_at, updated_at
		FROM transactions
		WHERE from_user_id = ? OR to_user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, userID, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to get transactions by user ID")
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, id int64, status domain.TransactionStatus) error {
	query := `UPDATE transactions SET status = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return errors.Wrap(err, "DATABASE_ERROR", "Failed to update transaction status")
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

func (r *transactionRepository) List(ctx context.Context, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, from_user_id, to_user_id, amount, type, status, created_at, updated_at
		FROM transactions
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to list transactions")
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) GetPendingTransactions(ctx context.Context, limit int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, from_user_id, to_user_id, amount, type, status, created_at, updated_at
		FROM transactions
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to get pending transactions")
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) scanTransactions(rows *sql.Rows) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		if err := rows.Scan(
			&tx.ID,
			&tx.FromUserID,
			&tx.ToUserID,
			&tx.Amount,
			&tx.Type,
			&tx.Status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		); err != nil {
			return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to scan transaction")
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "DATABASE_ERROR", "Failed to iterate transactions")
	}

	return transactions, nil
}
