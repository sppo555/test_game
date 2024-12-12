package auth

import (
	"baccarat/config"
	"baccarat/pkg/logger"
	"errors"
	"fmt"
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
	logger.Debug("Generating JWT token for user:", userID)
	
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		logger.Error("Error signing JWT token:", err)
		return "", fmt.Errorf("error signing token: %v", err)
	}

	logger.Debug("JWT token generated successfully for user:", userID)
	return tokenString, nil
}

// ValidateToken 驗證並解析JWT令牌
func (s *JWTService) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		logger.Warn("Error parsing JWT token:", err)
		return 0, ErrInvalidToken
	}

	if !token.Valid {
		logger.Warn("Invalid JWT token")
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		logger.Warn("Invalid JWT token claims")
		return 0, ErrInvalidToken
	}

	// 檢查是否過期
	if claims.ExpiresAt.Time.Before(time.Now()) {
		logger.Warn("JWT token has expired")
		return 0, ErrExpiredToken
	}

	// 解析用戶ID
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		logger.Error("Error converting user_id to int:", err)
		return 0, ErrInvalidToken
	}

	logger.Debug("JWT token validated successfully for user:", userID)
	return userID, nil
}

// ExtractBearerToken 從Authorization header提取Bearer token
func ExtractBearerToken(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
