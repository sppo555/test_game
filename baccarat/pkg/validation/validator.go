package validation

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrEmptyUsername     = errors.New("用戶名不能為空")
	ErrInvalidUsername   = errors.New("用戶名只能包含字母、數字和下劃線")
	ErrUsernameTooShort  = errors.New("用戶名長度不能少於3個字符")
	ErrUsernameTooLong   = errors.New("用戶名長度不能超過20個字符")
	ErrEmptyPassword     = errors.New("密碼不能為空")
	ErrPasswordTooShort  = errors.New("密碼長度不能少於6個字符")
	ErrPasswordTooLong   = errors.New("密碼長度不能超過50個字符")
	ErrInvalidPassword   = errors.New("密碼必須包含至少一個大寫字母、一個小寫字母和一個數字")
	ErrInvalidAmount     = errors.New("金額必須大於0")
	ErrInvalidBetType    = errors.New("無效的投注類型")
	ErrInvalidGameID     = errors.New("無效的遊戲ID")
)

// 用戶名正則表達式
var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

// 密碼規則正則表達式
var (
	hasUpperCase = regexp.MustCompile(`[A-Z]`)
	hasLowerCase = regexp.MustCompile(`[a-z]`)
	hasNumber    = regexp.MustCompile(`[0-9]`)
)

// ValidateUsername 驗證用戶名
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return ErrEmptyUsername
	}
	
	if len(username) < 3 {
		return ErrUsernameTooShort
	}
	
	if len(username) > 20 {
		return ErrUsernameTooLong
	}
	
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	
	return nil
}

// ValidatePassword 驗證密碼
func ValidatePassword(password string) error {
	if password == "" {
		return ErrEmptyPassword
	}
	
	if len(password) < 6 {
		return ErrPasswordTooShort
	}
	
	if len(password) > 50 {
		return ErrPasswordTooLong
	}
	
	if !hasUpperCase.MatchString(password) ||
		!hasLowerCase.MatchString(password) ||
		!hasNumber.MatchString(password) {
		return ErrInvalidPassword
	}
	
	return nil
}

// ValidateAmount 驗證金額
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// ValidateBetType 驗證投注類型
func ValidateBetType(betType string) error {
	validBetTypes := map[string]bool{
		"player":    true,
		"banker":    true,
		"tie":       true,
		"luckySix":  true,
	}
	
	if !validBetTypes[strings.ToLower(betType)] {
		return ErrInvalidBetType
	}
	return nil
}

// ValidateGameID 驗證遊戲ID
func ValidateGameID(gameID string) error {
	if len(gameID) != 36 { // UUID長度
		return ErrInvalidGameID
	}
	return nil
}

// ValidateRegistration 驗證註冊信息
func ValidateRegistration(username, password string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	return ValidatePassword(password)
}

// ValidateBet 驗證投注信息
func ValidateBet(betType string, amount float64) error {
	if err := ValidateBetType(betType); err != nil {
		return err
	}
	return ValidateAmount(amount)
}
