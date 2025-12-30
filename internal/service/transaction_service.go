package service

import (
	"context"

	"github.com/atakannerturk/go-backend-path/internal/domain"
	"github.com/atakannerturk/go-backend-path/internal/repository"
	"github.com/atakannerturk/go-backend-path/internal/worker"
	"github.com/atakannerturk/go-backend-path/pkg/errors"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type TransactionService struct {
	txRepo         repository.TransactionRepository
	balanceService *BalanceService
	workerPool     *worker.WorkerPool
}

func NewTransactionService(
	txRepo repository.TransactionRepository,
	balanceService *BalanceService,
	workerPool *worker.WorkerPool,
) *TransactionService {
	return &TransactionService{
		txRepo:         txRepo,
		balanceService: balanceService,
		workerPool:     workerPool,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, input domain.CreateTransactionInput) (*domain.Transaction, error) {
	logger.Info("Creating transaction", "type", input.Type, "amount", input.Amount)

	tx, err := domain.NewTransaction(input)
	if err != nil {
		return nil, err
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		logger.Error("Failed to create transaction", "error", err)
		return nil, err
	}

	logger.Info("Transaction created", "transaction_id", tx.ID, "type", tx.Type)

	submitted := s.workerPool.Submit(func(workerCtx context.Context) error {
		return s.processTransaction(workerCtx, tx.ID)
	})

	if !submitted {
		logger.Error("Failed to submit transaction to worker pool", "transaction_id", tx.ID)
		s.txRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed)
		return nil, errors.New("QUEUE_FULL", "Transaction queue is full, please try again later")
	}

	return tx, nil
}

func (s *TransactionService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	input := domain.CreateTransactionInput{
		ToUserID: &userID,
		Amount:   amount,
		Type:     domain.TransactionTypeCredit,
	}
	return s.CreateTransaction(ctx, input)
}

func (s *TransactionService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	input := domain.CreateTransactionInput{
		FromUserID: &userID,
		Amount:     amount,
		Type:       domain.TransactionTypeDebit,
	}
	return s.CreateTransaction(ctx, input)
}

func (s *TransactionService) Transfer(ctx context.Context, input domain.TransferInput) (*domain.Transaction, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	txInput := domain.CreateTransactionInput{
		FromUserID: &input.FromUserID,
		ToUserID:   &input.ToUserID,
		Amount:     input.Amount,
		Type:       domain.TransactionTypeTransfer,
	}

	return s.CreateTransaction(ctx, txInput)
}

func (s *TransactionService) processTransaction(ctx context.Context, txID int64) error {
	logger.Info("Processing transaction", "transaction_id", txID)

	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		logger.Error("Failed to get transaction", "transaction_id", txID, "error", err)
		return err
	}

	if tx.Status != domain.TransactionStatusPending {
		logger.Warn("Transaction is not in pending status", "transaction_id", txID, "status", tx.Status)
		return nil
	}

	if err := s.txRepo.UpdateStatus(ctx, txID, domain.TransactionStatusProcessing); err != nil {
		logger.Error("Failed to update transaction status to processing", "transaction_id", txID, "error", err)
		return err
	}

	var processErr error

	switch tx.Type {
	case domain.TransactionTypeCredit:
		if tx.ToUserID.Valid {
			processErr = s.balanceService.Credit(ctx, tx.ToUserID.Int64, tx.Amount)
		} else {
			processErr = errors.New("INVALID_TRANSACTION", "Credit transaction requires to_user_id")
		}

	case domain.TransactionTypeDebit:
		if tx.FromUserID.Valid {
			processErr = s.balanceService.Debit(ctx, tx.FromUserID.Int64, tx.Amount)
		} else {
			processErr = errors.New("INVALID_TRANSACTION", "Debit transaction requires from_user_id")
		}

	case domain.TransactionTypeTransfer:
		if tx.FromUserID.Valid && tx.ToUserID.Valid {
			processErr = s.balanceService.Transfer(ctx, tx.FromUserID.Int64, tx.ToUserID.Int64, tx.Amount)
		} else {
			processErr = errors.New("INVALID_TRANSACTION", "Transfer transaction requires both from_user_id and to_user_id")
		}

	default:
		processErr = errors.New("INVALID_TYPE", "Invalid transaction type")
	}

	if processErr != nil {
		logger.Error("Failed to process transaction", "transaction_id", txID, "error", processErr)
		s.txRepo.UpdateStatus(ctx, txID, domain.TransactionStatusFailed)
		return processErr
	}

	if err := s.txRepo.UpdateStatus(ctx, txID, domain.TransactionStatusCompleted); err != nil {
		logger.Error("Failed to update transaction status to completed", "transaction_id", txID, "error", err)
		return err
	}

	logger.Info("Transaction processed successfully", "transaction_id", txID)
	return nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, txID int64) (*domain.Transaction, error) {
	return s.txRepo.GetByID(ctx, txID)
}

func (s *TransactionService) GetUserTransactions(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error) {
	return s.txRepo.GetByUserID(ctx, userID, limit, offset)
}

func (s *TransactionService) ListTransactions(ctx context.Context, limit, offset int) ([]*domain.Transaction, error) {
	return s.txRepo.List(ctx, limit, offset)
}

func (s *TransactionService) RollbackTransaction(ctx context.Context, txID int64) error {
	logger.Info("Rolling back transaction", "transaction_id", txID)

	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return err
	}

	if tx.Status != domain.TransactionStatusCompleted {
		return errors.New("INVALID_STATUS", "Can only rollback completed transactions")
	}

	var rollbackErr error

	switch tx.Type {
	case domain.TransactionTypeCredit:
		if tx.ToUserID.Valid {
			rollbackErr = s.balanceService.Debit(ctx, tx.ToUserID.Int64, tx.Amount)
		}

	case domain.TransactionTypeDebit:
		if tx.FromUserID.Valid {
			rollbackErr = s.balanceService.Credit(ctx, tx.FromUserID.Int64, tx.Amount)
		}

	case domain.TransactionTypeTransfer:
		if tx.FromUserID.Valid && tx.ToUserID.Valid {
			rollbackErr = s.balanceService.Transfer(ctx, tx.ToUserID.Int64, tx.FromUserID.Int64, tx.Amount)
		}
	}

	if rollbackErr != nil {
		logger.Error("Failed to rollback transaction", "transaction_id", txID, "error", rollbackErr)
		return rollbackErr
	}

	if err := s.txRepo.UpdateStatus(ctx, txID, domain.TransactionStatusRolledBack); err != nil {
		logger.Error("Failed to update transaction status to rolled_back", "transaction_id", txID, "error", err)
		return err
	}

	logger.Info("Transaction rolled back successfully", "transaction_id", txID)
	return nil
}
