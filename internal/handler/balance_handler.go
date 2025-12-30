package handler

import (
	"net/http"

	"github.com/atakannerturk/go-backend-path/internal/middleware"
	"github.com/atakannerturk/go-backend-path/internal/service"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type BalanceHandler struct {
	balanceService *service.BalanceService
}

func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

func (h *BalanceHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get balance", "user_id", userID, "error", err)
		middleware.RespondError(w, http.StatusNotFound, "Balance not found")
		return
	}

	middleware.RespondSuccess(w, http.StatusOK, balance)
}
