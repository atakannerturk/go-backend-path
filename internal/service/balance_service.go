package service

import (
	"context"
	"sync"

	"github.com/atakannerturk/go-backend-path/internal/domain"
	"github.com/atakannerturk/go-backend-path/internal/repository"
	"github.com/atakannerturk/go-backend-path/pkg/errors"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type BalanceService struct {
	balanceRepo repository.BalanceRepository
	mu          sync.RWMutex
	balances    map[int64]*domain.Balance
}

func NewBalanceService(balanceRepo repository.BalanceRepository) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		balances:    make(map[int64]*domain.Balance),
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (*domain.Balance, error) {
	s.mu.RLock()
	if balance, ok := s.balances[userID]; ok {
		s.mu.RUnlock()
		return balance, nil
	}
	s.mu.RUnlock()

	balance, err := s.balanceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.balances[userID] = balance
	s.mu.Unlock()

	return balance, nil
}

func (s *BalanceService) Credit(ctx context.Context, userID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Credit amount must be positive")
	}

	logger.Info("Crediting balance", "user_id", userID, "amount", amount)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.balanceRepo.Credit(ctx, userID, amount); err != nil {
		logger.Error("Failed to credit balance", "user_id", userID, "error", err)
		return err
	}

	if balance, ok := s.balances[userID]; ok {
		balance.Credit(amount)
	}

	logger.Info("Balance credited successfully", "user_id", userID, "amount", amount)
	return nil
}

func (s *BalanceService) Debit(ctx context.Context, userID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Debit amount must be positive")
	}

	logger.Info("Debiting balance", "user_id", userID, "amount", amount)

	s.mu.Lock()
	defer s.mu.Unlock()

	balance, err := s.balanceRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get balance", "user_id", userID, "error", err)
		return err
	}

	if balance.Amount < amount {
		logger.Warn("Insufficient funds", "user_id", userID, "balance", balance.Amount, "amount", amount)
		return errors.ErrInsufficientFunds
	}

	if err := s.balanceRepo.Debit(ctx, userID, amount); err != nil {
		logger.Error("Failed to debit balance", "user_id", userID, "error", err)
		return err
	}

	if cachedBalance, ok := s.balances[userID]; ok {
		cachedBalance.Debit(amount)
	}

	logger.Info("Balance debited successfully", "user_id", userID, "amount", amount)
	return nil
}

func (s *BalanceService) Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64) error {
	if amount <= 0 {
		return errors.New("INVALID_AMOUNT", "Transfer amount must be positive")
	}

	if fromUserID == toUserID {
		return errors.New("INVALID_TRANSFER", "Cannot transfer to the same account")
	}

	logger.Info("Transferring balance", "from_user_id", fromUserID, "to_user_id", toUserID, "amount", amount)

	s.mu.Lock()
	defer s.mu.Unlock()

	fromBalance, err := s.balanceRepo.GetByUserID(ctx, fromUserID)
	if err != nil {
		logger.Error("Failed to get from balance", "user_id", fromUserID, "error", err)
		return err
	}

	if fromBalance.Amount < amount {
		logger.Warn("Insufficient funds for transfer", "user_id", fromUserID, "balance", fromBalance.Amount, "amount", amount)
		return errors.ErrInsufficientFunds
	}

	if err := s.balanceRepo.Debit(ctx, fromUserID, amount); err != nil {
		logger.Error("Failed to debit from account", "user_id", fromUserID, "error", err)
		return err
	}

	if err := s.balanceRepo.Credit(ctx, toUserID, amount); err != nil {
		logger.Error("Failed to credit to account, attempting rollback", "user_id", toUserID, "error", err)
		if rollbackErr := s.balanceRepo.Credit(ctx, fromUserID, amount); rollbackErr != nil {
			logger.Error("CRITICAL: Failed to rollback debit", "user_id", fromUserID, "error", rollbackErr)
		}
		return err
	}

	if cachedBalance, ok := s.balances[fromUserID]; ok {
		cachedBalance.Debit(amount)
	}
	if cachedBalance, ok := s.balances[toUserID]; ok {
		cachedBalance.Credit(amount)
	}

	logger.Info("Balance transferred successfully", "from_user_id", fromUserID, "to_user_id", toUserID, "amount", amount)
	return nil
}

func (s *BalanceService) InvalidateCache(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.balances, userID)
}

func (s *BalanceService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances = make(map[int64]*domain.Balance)
	logger.Info("Balance cache cleared")
}
