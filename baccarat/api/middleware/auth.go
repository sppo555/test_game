package middleware

import (
	"baccarat/internal/auth"
	"baccarat/pkg/utils"
	"context"
	"net/http"
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
			utils.UnauthorizedError(w)
			return
		}

		tokenString := auth.ExtractBearerToken(authHeader)
		if tokenString == "" {
			utils.UnauthorizedError(w)
			return
		}

		userID, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			utils.UnauthorizedError(w)
			return
		}

		// 將用戶ID添加到請求上下文中
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID 從請求上下文中獲取用戶ID
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	return userID, ok
}
