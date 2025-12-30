package handler

import (
	"encoding/json"
	"net/http"

	"github.com/atakannerturk/go-backend-path/internal/domain"
	"github.com/atakannerturk/go-backend-path/internal/middleware"
	"github.com/atakannerturk/go-backend-path/internal/service"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Warn("Invalid request body", "error", err)
		middleware.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	authResponse, err := h.userService.Register(r.Context(), input)
	if err != nil {
		logger.Error("Registration failed", "error", err)
		middleware.RespondErrorWithCode(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	logger.Info("User registered successfully", "user_id", authResponse.User.ID)
	middleware.RespondSuccess(w, http.StatusCreated, authResponse)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input domain.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Warn("Invalid request body", "error", err)
		middleware.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	authResponse, err := h.userService.Login(r.Context(), input)
	if err != nil {
		logger.Warn("Login failed", "email", input.Email, "error", err)
		middleware.RespondErrorWithCode(w, http.StatusUnauthorized, "LOGIN_FAILED", "Invalid credentials")
		return
	}

	logger.Info("User logged in successfully", "user_id", authResponse.User.ID)
	middleware.RespondSuccess(w, http.StatusOK, authResponse)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		middleware.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user", "user_id", userID, "error", err)
		middleware.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	middleware.RespondSuccess(w, http.StatusOK, user)
}
