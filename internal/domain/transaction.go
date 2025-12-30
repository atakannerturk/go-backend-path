package domain

import (
	"database/sql"
	"time"

	"github.com/atakannerturk/go-backend-path/pkg/errors"
)

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeCredit   TransactionType = "credit"
	TransactionTypeDebit    TransactionType = "debit"
	TransactionTypeTransfer TransactionType = "transfer"
)

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
)

type Transaction struct {
	ID         int64             `json:"id" db:"id"`
	FromUserID sql.NullInt64     `json:"from_user_id,omitempty" db:"from_user_id"`
	ToUserID   sql.NullInt64     `json:"to_user_id,omitempty" db:"to_user_id"`
	Amount     float64           `json:"amount" db:"amount"`
	Type       TransactionType   `json:"type" db:"type"`
	Status     TransactionStatus `json:"status" db:"status"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
}

type CreateTransactionInput struct {
	FromUserID *int64          `json:"from_user_id,omitempty"`
	ToUserID   *int64          `json:"to_user_id,omitempty"`
	Amount     float64         `json:"amount"`
	Type       TransactionType `json:"type"`
}

type TransferInput struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Transaction amount must be positive")
	}

	if !t.Type.IsValid() {
		return errors.New("INVALID_TYPE", "Invalid transaction type")
	}

	if !t.Status.IsValid() {
		return errors.New("INVALID_STATUS", "Invalid transaction status")
	}

	switch t.Type {
	case TransactionTypeCredit:
		if !t.ToUserID.Valid {
			return errors.New("INVALID_TRANSACTION", "Credit transaction requires to_user_id")
		}
		if t.FromUserID.Valid {
			return errors.New("INVALID_TRANSACTION", "Credit transaction should not have from_user_id")
		}
	case TransactionTypeDebit:
		if !t.FromUserID.Valid {
			return errors.New("INVALID_TRANSACTION", "Debit transaction requires from_user_id")
		}
		if t.ToUserID.Valid {
			return errors.New("INVALID_TRANSACTION", "Debit transaction should not have to_user_id")
		}
	case TransactionTypeTransfer:
		if !t.FromUserID.Valid || !t.ToUserID.Valid {
			return errors.New("INVALID_TRANSACTION", "Transfer transaction requires both from_user_id and to_user_id")
		}
		if t.FromUserID.Int64 == t.ToUserID.Int64 {
			return errors.New("INVALID_TRANSACTION", "Cannot transfer to the same account")
		}
	}

	return nil
}

func (t *Transaction) CanTransition(newStatus TransactionStatus) bool {
	switch t.Status {
	case TransactionStatusPending:
		return newStatus == TransactionStatusProcessing || newStatus == TransactionStatusFailed
	case TransactionStatusProcessing:
		return newStatus == TransactionStatusCompleted || newStatus == TransactionStatusFailed || newStatus == TransactionStatusRolledBack
	case TransactionStatusCompleted:
		return newStatus == TransactionStatusRolledBack
	case TransactionStatusFailed, TransactionStatusRolledBack:
		return false
	default:
		return false
	}
}

func (t *Transaction) UpdateStatus(newStatus TransactionStatus) error {
	if !t.CanTransition(newStatus) {
		return errors.New("INVALID_STATUS_TRANSITION",
			"Cannot transition from " + string(t.Status) + " to " + string(newStatus))
	}
	t.Status = newStatus
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Transaction) MarkAsProcessing() error {
	return t.UpdateStatus(TransactionStatusProcessing)
}

func (t *Transaction) MarkAsCompleted() error {
	return t.UpdateStatus(TransactionStatusCompleted)
}

func (t *Transaction) MarkAsFailed() error {
	return t.UpdateStatus(TransactionStatusFailed)
}

func (t *Transaction) MarkAsRolledBack() error {
	return t.UpdateStatus(TransactionStatusRolledBack)
}

func (tt TransactionType) IsValid() bool {
	return tt == TransactionTypeCredit || tt == TransactionTypeDebit || tt == TransactionTypeTransfer
}

func (ts TransactionStatus) IsValid() bool {
	return ts == TransactionStatusPending ||
		   ts == TransactionStatusProcessing ||
		   ts == TransactionStatusCompleted ||
		   ts == TransactionStatusFailed ||
		   ts == TransactionStatusRolledBack
}

func (input *CreateTransactionInput) Validate() error {
	if input.Amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Transaction amount must be positive")
	}

	if !input.Type.IsValid() {
		return errors.New("INVALID_TYPE", "Invalid transaction type")
	}

	switch input.Type {
	case TransactionTypeCredit:
		if input.ToUserID == nil {
			return errors.New("INVALID_TRANSACTION", "Credit transaction requires to_user_id")
		}
	case TransactionTypeDebit:
		if input.FromUserID == nil {
			return errors.New("INVALID_TRANSACTION", "Debit transaction requires from_user_id")
		}
	case TransactionTypeTransfer:
		if input.FromUserID == nil || input.ToUserID == nil {
			return errors.New("INVALID_TRANSACTION", "Transfer transaction requires both from_user_id and to_user_id")
		}
		if *input.FromUserID == *input.ToUserID {
			return errors.New("INVALID_TRANSACTION", "Cannot transfer to the same account")
		}
	}

	return nil
}

func (input *TransferInput) Validate() error {
	if input.Amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Transfer amount must be positive")
	}
	if input.FromUserID == input.ToUserID {
		return errors.New("INVALID_TRANSACTION", "Cannot transfer to the same account")
	}
	return nil
}

func NewTransaction(input CreateTransactionInput) (*Transaction, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	tx := &Transaction{
		Amount:    input.Amount,
		Type:      input.Type,
		Status:    TransactionStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if input.FromUserID != nil {
		tx.FromUserID = sql.NullInt64{Int64: *input.FromUserID, Valid: true}
	}

	if input.ToUserID != nil {
		tx.ToUserID = sql.NullInt64{Int64: *input.ToUserID, Valid: true}
	}

	return tx, nil
}
