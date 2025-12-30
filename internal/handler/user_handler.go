package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/atakannerturk/go-backend-path/internal/middleware"
	"github.com/atakannerturk/go-backend-path/internal/service"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get user", "user_id", id, "error", err)
		middleware.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	middleware.RespondSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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

	users, err := h.userService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to list users", "error", err)
		middleware.RespondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	middleware.RespondSuccess(w, http.StatusOK, map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		middleware.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if currentUserID == id {
		middleware.RespondError(w, http.StatusBadRequest, "Cannot delete your own account")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		logger.Error("Failed to delete user", "user_id", id, "error", err)
		middleware.RespondError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	logger.Info("User deleted", "user_id", id, "deleted_by", currentUserID)
	middleware.RespondSuccess(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}
