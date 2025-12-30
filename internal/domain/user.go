package domain

import (
	"database/sql"
	"regexp"
	"strings"
	"time"

	"github.com/atakannerturk/go-backend-path/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           int64        `json:"id" db:"id"`
	Username     string       `json:"username" db:"username"`
	Email        string       `json:"email" db:"email"`
	PasswordHash string       `json:"-" db:"password_hash"`
	Role         UserRole     `json:"role" db:"role"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

type CreateUserInput struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
)

func (u *User) Validate() error {
	if err := ValidateUsername(u.Username); err != nil {
		return err
	}
	if err := ValidateEmail(u.Email); err != nil {
		return err
	}
	if !u.Role.IsValid() {
		return errors.New("INVALID_ROLE", "Invalid user role")
	}
	return nil
}

func (u *User) SetPassword(password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "PASSWORD_HASH_ERROR", "Failed to hash password")
	}

	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (r UserRole) IsValid() bool {
	return r == RoleUser || r == RoleAdmin
}

func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 20 {
		return errors.New("INVALID_USERNAME", "Username must be between 3 and 20 characters")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("INVALID_USERNAME", "Username can only contain letters, numbers, and underscores")
	}
	return nil
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("INVALID_EMAIL", "Email is required")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("INVALID_EMAIL", "Invalid email format")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("INVALID_PASSWORD", "Password must be at least 8 characters long")
	}
	if len(password) > 72 {
		return errors.New("INVALID_PASSWORD", "Password must not exceed 72 characters")
	}
	return nil
}

func (input *CreateUserInput) Validate() error {
	if err := ValidateUsername(input.Username); err != nil {
		return err
	}
	if err := ValidateEmail(input.Email); err != nil {
		return err
	}
	if err := ValidatePassword(input.Password); err != nil {
		return err
	}
	if input.Role == "" {
		input.Role = RoleUser
	}
	if !input.Role.IsValid() {
		return errors.New("INVALID_ROLE", "Invalid user role")
	}
	return nil
}

func (input *LoginInput) Validate() error {
	if err := ValidateEmail(input.Email); err != nil {
		return err
	}
	if input.Password == "" {
		return errors.New("INVALID_PASSWORD", "Password is required")
	}
	return nil
}

func NewUser(input CreateUserInput) (*User, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	user := &User{
		Username:  strings.TrimSpace(input.Username),
		Email:     strings.ToLower(strings.TrimSpace(input.Email)),
		Role:      input.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.SetPassword(input.Password); err != nil {
		return nil, err
	}

	return user, nil
}

type NullUser struct {
	User  User
	Valid bool
}

func (nu *NullUser) Scan(value interface{}) error {
	if value == nil {
		nu.Valid = false
		return nil
	}
	nu.Valid = true
	return nil
}

func NewNullUser(user *User) NullUser {
	if user == nil {
		return NullUser{Valid: false}
	}
	return NullUser{User: *user, Valid: true}
}

func ScanNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
