package validation

import (
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{"Empty username", "", ErrEmptyUsername},
		{"Too short username", "ab", ErrUsernameTooShort},
		{"Too long username", "abcdefghijklmnopqrstuvwxyz", ErrUsernameTooLong},
		{"Invalid characters", "user@name", ErrInvalidUsername},
		{"Valid username", "user123", nil},
		{"Valid username with underscore", "user_123", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if err != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"Empty password", "", ErrEmptyPassword},
		{"Too short password", "Aa1", ErrPasswordTooShort},
		{"Too long password", "Aa1" + string(make([]byte, 50)), ErrPasswordTooLong},
		{"No uppercase", "password123", ErrInvalidPassword},
		{"No lowercase", "PASSWORD123", ErrInvalidPassword},
		{"No number", "PasswordABC", ErrInvalidPassword},
		{"Valid password", "Password123", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr error
	}{
		{"Zero amount", 0, ErrInvalidAmount},
		{"Negative amount", -1, ErrInvalidAmount},
		{"Valid amount", 100, nil},
		{"Small valid amount", 0.01, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if err != tt.wantErr {
				t.Errorf("ValidateAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBetType(t *testing.T) {
	tests := []struct {
		name    string
		betType string
		wantErr error
	}{
		{"Empty bet type", "", ErrInvalidBetType},
		{"Invalid bet type", "invalid", ErrInvalidBetType},
		{"Valid player bet", "player", nil},
		{"Valid banker bet", "banker", nil},
		{"Valid tie bet", "tie", nil},
		{"Valid lucky six bet", "luckySix", nil},
		{"Case insensitive test", "PLAYER", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBetType(tt.betType)
			if err != tt.wantErr {
				t.Errorf("ValidateBetType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGameID(t *testing.T) {
	tests := []struct {
		name    string
		gameID  string
		wantErr error
	}{
		{"Empty game ID", "", ErrInvalidGameID},
		{"Invalid length game ID", "123", ErrInvalidGameID},
		{"Valid game ID", "550e8400-e29b-41d4-a716-446655440000", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGameID(tt.gameID)
			if err != tt.wantErr {
				t.Errorf("ValidateGameID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRegistration(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{"Invalid username", "a", ErrUsernameTooShort},
		{"Invalid password", "weak", ErrPasswordTooShort},
		{"Valid registration", "user123", "Password123", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRegistration(tt.username, tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidateRegistration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBet(t *testing.T) {
	tests := []struct {
		name    string
		betType string
		amount  float64
		wantErr error
	}{
		{"Invalid bet type", "invalid", 100, ErrInvalidBetType},
		{"Invalid amount", "player", 0, ErrInvalidAmount},
		{"Valid bet", "player", 100, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBet(tt.betType, tt.amount)
			if err != tt.wantErr {
				t.Errorf("ValidateBet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
