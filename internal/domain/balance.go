package domain

import (
	"sync"
	"time"

	"github.com/atakannerturk/go-backend-path/pkg/errors"
)

type Balance struct {
	UserID        int64     `json:"user_id" db:"user_id"`
	Amount        float64   `json:"amount" db:"amount"`
	LastUpdatedAt time.Time `json:"last_updated_at" db:"last_updated_at"`
	mu            sync.RWMutex `json:"-" db:"-"`
}

type BalanceChange struct {
	UserID int64
	Amount float64
	Type   BalanceChangeType
}

type BalanceChangeType string

const (
	BalanceChangeTypeCredit BalanceChangeType = "credit"
	BalanceChangeTypeDebit  BalanceChangeType = "debit"
)

func NewBalance(userID int64) *Balance {
	return &Balance{
		UserID:        userID,
		Amount:        0,
		LastUpdatedAt: time.Now(),
	}
}

func (b *Balance) Credit(amount float64) error {
	if amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Credit amount must be positive")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.Amount += amount
	b.LastUpdatedAt = time.Now()
	return nil
}

func (b *Balance) Debit(amount float64) error {
	if amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Debit amount must be positive")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Amount < amount {
		return errors.ErrInsufficientFunds
	}

	b.Amount -= amount
	b.LastUpdatedAt = time.Now()
	return nil
}

func (b *Balance) GetAmount() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Amount
}

func (b *Balance) SetAmount(amount float64) error {
	if amount < 0 {
		return errors.New("INVALID_AMOUNT", "Balance amount cannot be negative")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.Amount = amount
	b.LastUpdatedAt = time.Now()
	return nil
}

func (b *Balance) HasSufficientFunds(amount float64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Amount >= amount
}

func (bc *BalanceChange) Validate() error {
	if bc.Amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Balance change amount must be positive")
	}
	if bc.Type != BalanceChangeTypeCredit && bc.Type != BalanceChangeTypeDebit {
		return errors.New("INVALID_TYPE", "Invalid balance change type")
	}
	return nil
}
