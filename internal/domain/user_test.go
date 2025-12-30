package domain

import (
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "john_doe", false},
		{"Valid username numbers", "user123", false},
		{"Too short", "ab", true},
		{"Too long", "this_is_a_very_long_username", true},
		{"Invalid characters", "john-doe", true},
		{"Empty username", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "test@example.com", false},
		{"Valid email with subdomain", "test@mail.example.com", false},
		{"Invalid email no @", "testexample.com", true},
		{"Invalid email no domain", "test@", true},
		{"Empty email", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "password123", false},
		{"Valid long password", "this_is_a_very_secure_password_123", false},
		{"Too short", "pass", true},
		{"Empty password", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserSetPassword(t *testing.T) {
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}

	password := "password123"
	err := user.SetPassword(password)
	if err != nil {
		t.Errorf("SetPassword() error = %v", err)
	}

	if user.PasswordHash == "" {
		t.Error("Password hash should not be empty")
	}

	if user.PasswordHash == password {
		t.Error("Password hash should not equal plain password")
	}
}

func TestUserCheckPassword(t *testing.T) {
	user := &User{
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleUser,
	}

	password := "password123"
	user.SetPassword(password)

	if !user.CheckPassword(password) {
		t.Error("CheckPassword should return true for correct password")
	}

	if user.CheckPassword("wrongpassword") {
		t.Error("CheckPassword should return false for incorrect password")
	}
}

func TestNewUser(t *testing.T) {
	input := CreateUserInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     RoleUser,
	}

	user, err := NewUser(input)
	if err != nil {
		t.Errorf("NewUser() error = %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.PasswordHash == "" {
		t.Error("Password hash should not be empty")
	}

	if !user.CheckPassword("password123") {
		t.Error("Password should be correctly hashed")
	}
}
