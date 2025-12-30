package errors

import "fmt"

type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrNotFound          = New("NOT_FOUND", "Resource not found")
	ErrInvalidInput      = New("INVALID_INPUT", "Invalid input provided")
	ErrUnauthorized      = New("UNAUTHORIZED", "Unauthorized access")
	ErrForbidden         = New("FORBIDDEN", "Forbidden access")
	ErrInternalServer    = New("INTERNAL_SERVER_ERROR", "Internal server error")
	ErrDatabaseError     = New("DATABASE_ERROR", "Database operation failed")
	ErrInsufficientFunds = New("INSUFFICIENT_FUNDS", "Insufficient funds for transaction")
	ErrDuplicateEntry    = New("DUPLICATE_ENTRY", "Duplicate entry")
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Invalid credentials")
)
