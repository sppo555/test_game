package handlers

import (
	"baccarat/db"
	"baccarat/internal/auth"
	"baccarat/pkg/utils"
	"baccarat/pkg/validation"
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db         *sql.DB
	jwtService *auth.JWTService
}

func NewAuthHandler(db *sql.DB, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
	}
}

// Register 處理用戶註冊
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.ValidationError(w, "Invalid request body")
		return
	}

	// 驗證註冊信息
	if err := validation.ValidateRegistration(input.Username, input.Password); err != nil {
		utils.ValidationError(w, err.Error())
		return
	}

	// 檢查用戶名是否已存在
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", input.Username).Scan(&exists)
	if err != nil {
		utils.ServerError(w, "Error checking username")
		return
	}
	if exists {
		utils.ValidationError(w, "Username already exists")
		return
	}

	// 加密密碼
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ServerError(w, "Error creating account")
		return
	}

	// 創建用戶
	if err := db.CreateUser(input.Username, passwordHash); err != nil {
		utils.ServerError(w, "Error creating account")
		return
	}

	utils.SuccessResponse(w, map[string]string{"message": "Account created successfully"})
}

// Login 處理用戶登入
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.ValidationError(w, "Invalid request body")
		return
	}

	// 驗證輸入
	if err := validation.ValidateUsername(input.Username); err != nil {
		utils.ValidationError(w, err.Error())
		return
	}
	if err := validation.ValidatePassword(input.Password); err != nil {
		utils.ValidationError(w, err.Error())
		return
	}

	// 獲取用戶信息
	userID, passwordHash, err := db.GetUserByUsername(input.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.ValidationError(w, "Invalid username or password")
			return
		}
		utils.ServerError(w, "Error during login")
		return
	}

	// 驗證密碼
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		utils.ValidationError(w, "Invalid username or password")
		return
	}

	// 生成JWT令牌
	token, err := h.jwtService.GenerateToken(userID)
	if err != nil {
		utils.ServerError(w, "Error generating token")
		return
	}

	utils.SuccessResponse(w, map[string]string{
		"message": "Login successful",
		"token":   token,
	})
}

// Logout 處理用戶登出
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 在客戶端，只需要刪除JWT令牌即可
	utils.SuccessResponse(w, map[string]string{
		"message": "Logged out successfully",
	})
}
