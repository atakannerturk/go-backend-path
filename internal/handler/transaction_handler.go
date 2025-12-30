package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/atakannerturk/go-backend-path/internal/domain"
	"github.com/atakannerturk/go-backend-path/internal/middleware"
	"github.com/atakannerturk/go-backend-path/internal/service"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type TransactionHandler struct {
	txService *service.TransactionService
}

func NewTransactionHandler(txService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		txService: txService,
	}
}

type CreditRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

type DebitRequest struct {
	UserID int64   `json:"user_id"`
	Amount float64 `json:"amount"`
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) {
	var req CreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Amount <= 0 {
		middleware.RespondError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}

	tx, err := h.txService.Credit(r.Context(), req.UserID, req.Amount)
	if err != nil {
		logger.Error("Credit transaction failed", "user_id", req.UserID, "error", err)
		middleware.RespondErrorWithCode(w, http.StatusBadRequest, "TRANSACTION_FAILED", err.Error())
		return
	}

	logger.Info("Credit transaction created", "transaction_id", tx.ID, "user_id", req.UserID, "amount", req.Amount)
	middleware.RespondSuccess(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	var req DebitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Amount <= 0 {
		middleware.RespondError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}

	tx, err := h.txService.Debit(r.Context(), req.UserID, req.Amount)
	if err != nil {
		logger.Error("Debit transaction failed", "user_id", req.UserID, "error", err)
		middleware.RespondErrorWithCode(w, http.StatusBadRequest, "TRANSACTION_FAILED", err.Error())
		return
	}

	logger.Info("Debit transaction created", "transaction_id", tx.ID, "user_id", req.UserID, "amount", req.Amount)
	middleware.RespondSuccess(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var input domain.TransferInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if ok {
		input.FromUserID = userID
	}

	tx, err := h.txService.Transfer(r.Context(), input)
	if err != nil {
		logger.Error("Transfer transaction failed", "from", input.FromUserID, "to", input.ToUserID, "error", err)
		middleware.RespondErrorWithCode(w, http.StatusBadRequest, "TRANSACTION_FAILED", err.Error())
		return
	}

	logger.Info("Transfer transaction created", "transaction_id", tx.ID, "from", input.FromUserID, "to", input.ToUserID, "amount", input.Amount)
	middleware.RespondSuccess(w, http.StatusCreated, tx)
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	tx, err := h.txService.GetTransaction(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get transaction", "transaction_id", id, "error", err)
		middleware.RespondError(w, http.StatusNotFound, "Transaction not found")
		return
	}

	middleware.RespondSuccess(w, http.StatusOK, tx)
}

func (h *TransactionHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	transactions, err := h.txService.GetUserTransactions(r.Context(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get transaction history", "user_id", userID, "error", err)
		middleware.RespondError(w, http.StatusInternalServerError, "Failed to get transaction history")
		return
	}

	middleware.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	transactions, err := h.txService.ListTransactions(r.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to list transactions", "error", err)
		middleware.RespondError(w, http.StatusInternalServerError, "Failed to list transactions")
		return
	}

	middleware.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *TransactionHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	if err := h.txService.RollbackTransaction(r.Context(), id); err != nil {
		logger.Error("Failed to rollback transaction", "transaction_id", id, "error", err)
		middleware.RespondErrorWithCode(w, http.StatusBadRequest, "ROLLBACK_FAILED", err.Error())
		return
	}

	logger.Info("Transaction rolled back", "transaction_id", id)
	middleware.RespondSuccess(w, http.StatusOK, map[string]string{
		"message": "Transaction rolled back successfully",
	})
}
