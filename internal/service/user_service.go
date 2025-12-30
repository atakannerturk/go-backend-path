package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/atakannerturk/go-backend-path/internal/domain"
	"github.com/atakannerturk/go-backend-path/internal/repository"
	"github.com/atakannerturk/go-backend-path/pkg/errors"
	"github.com/atakannerturk/go-backend-path/pkg/logger"
)

type UserService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

type AuthResponse struct {
	User  *domain.User `json:"user"`
	Token string       `json:"token"`
}

type Claims struct {
	UserID int64            `json:"user_id"`
	Email  string           `json:"email"`
	Role   domain.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *UserService) Register(ctx context.Context, input domain.CreateUserInput) (*AuthResponse, error) {
	logger.Info("Registering new user", "username", input.Username, "email", input.Email)

	if err := input.Validate(); err != nil {
		return nil, err
	}

	existingUser, _ := s.userRepo.GetByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, errors.New("USER_ALREADY_EXISTS", "User with this email already exists")
	}

	existingUser, _ = s.userRepo.GetByUsername(ctx, input.Username)
	if existingUser != nil {
		return nil, errors.New("USER_ALREADY_EXISTS", "User with this username already exists")
	}

	user, err := domain.NewUser(input)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", "error", err)
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		logger.Error("Failed to generate token", "error", err)
		return nil, err
	}

	logger.Info("User registered successfully", "user_id", user.ID, "username", user.Username)

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, input domain.LoginInput) (*AuthResponse, error) {
	logger.Info("User login attempt", "email", input.Email)

	if err := input.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		logger.Warn("Login failed: user not found", "email", input.Email)
		return nil, errors.ErrInvalidCredentials
	}

	if !user.CheckPassword(input.Password) {
		logger.Warn("Login failed: invalid password", "email", input.Email)
		return nil, errors.ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		logger.Error("Failed to generate token", "error", err)
		return nil, err
	}

	logger.Info("User logged in successfully", "user_id", user.ID, "username", user.Username)

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		logger.Error("Failed to update user", "user_id", user.ID, "error", err)
		return err
	}

	logger.Info("User updated successfully", "user_id", user.ID)
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		logger.Error("Failed to delete user", "user_id", userID, "error", err)
		return err
	}

	logger.Info("User deleted successfully", "user_id", userID)
	return nil
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to list users", "error", err)
		return nil, err
	}
	return users, nil
}

func (s *UserService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "INVALID_TOKEN", "Invalid token")
	}

	if !token.Valid {
		return nil, errors.New("INVALID_TOKEN", "Invalid token")
	}

	return claims, nil
}

func (s *UserService) generateToken(user *domain.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
