package auth

import (
	"baccarat/config"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type JWTService struct {
	secretKey []byte
	expiry    time.Duration
}

// NewJWTService 創建新的JWT服務
func NewJWTService() *JWTService {
	return &JWTService{
		secretKey: []byte(config.AppConfig.JWTSecret),
		expiry:    time.Duration(config.AppConfig.JWTExpiry) * time.Hour,
	}
}

// GenerateToken 生成JWT令牌
func (s *JWTService) GenerateToken(userID int) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken 驗證並解析JWT令牌
func (s *JWTService) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		return 0, ErrInvalidToken
	}

	if !token.Valid {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return 0, ErrInvalidToken
	}

	// 檢查是否過期
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return 0, ErrExpiredToken
	}

	// 解析用戶ID
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return 0, ErrInvalidToken
	}

	return userID, nil
}

// ExtractBearerToken 從Authorization header提取Bearer token
func ExtractBearerToken(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
