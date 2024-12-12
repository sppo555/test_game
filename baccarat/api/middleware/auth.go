package middleware

import (
	"baccarat/internal/auth"
	"baccarat/pkg/logger"
	"baccarat/pkg/utils"
	"context"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
}

func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// Authenticate JWT認證中間件
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			utils.UnauthorizedError(w)
			return
		}

		// 檢查 Bearer token 格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid Authorization header format")
			utils.UnauthorizedError(w)
			return
		}

		token := parts[1]
		userID, err := m.jwtService.ValidateToken(token)
		if err != nil {
			logger.Warn("Invalid token:", err)
			utils.UnauthorizedError(w)
			return
		}

		// 將用戶ID添加到請求上下文中
		ctx := context.WithValue(r.Context(), "userID", userID)
		logger.Debug("User authenticated, UserID:", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID 從請求上下文中獲取用戶ID
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		logger.Warn("Failed to get userID from context")
		return 0, false
	}
	return userID, true
}
