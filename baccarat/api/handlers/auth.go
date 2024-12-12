package handlers

import (
	"baccarat/api/middleware"
	"baccarat/db"
	"baccarat/internal/auth"
	"baccarat/pkg/logger"
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
		logger.Warn("Invalid method for Register:", r.Method)
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Warn("Invalid request body for Register:", err)
		utils.ValidationError(w, "Invalid request body")
		return
	}

	// 驗證註冊信息
	if err := validation.ValidateRegistration(input.Username, input.Password); err != nil {
		logger.Warn("Invalid registration data:", err)
		utils.ValidationError(w, err.Error())
		return
	}

	// 檢查用戶名是否已存在
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", input.Username).Scan(&exists)
	if err != nil {
		logger.Error("Error checking username existence:", err)
		utils.ServerError(w, "Error checking username")
		return
	}
	if exists {
		logger.Warn("Username already exists:", input.Username)
		utils.ValidationError(w, "Username already exists")
		return
	}

	// 加密密碼
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Error hashing password:", err)
		utils.ServerError(w, "Error creating account")
		return
	}

	// 創建用戶
	if err := db.CreateUser(input.Username, passwordHash); err != nil {
		logger.Error("Error creating user:", err)
		utils.ServerError(w, "Error creating account")
		return
	}

	logger.Info("Successfully created user:", input.Username)
	utils.SuccessResponse(w, map[string]string{"message": "Account created successfully"})
}

// Login 處理用戶登入
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Warn("Invalid method for Login:", r.Method)
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Warn("Invalid request body for Login:", err)
		utils.ValidationError(w, "Invalid request body")
		return
	}

	// 驗證輸入
	if err := validation.ValidateUsername(input.Username); err != nil {
		logger.Warn("Invalid username:", err)
		utils.ValidationError(w, err.Error())
		return
	}
	if err := validation.ValidatePassword(input.Password); err != nil {
		logger.Warn("Invalid password format")
		utils.ValidationError(w, err.Error())
		return
	}

	// 獲取用戶信息
	userID, hashedPassword, err := db.GetUserByUsername(input.Username)
	if err == sql.ErrNoRows {
		logger.Warn("User not found:", input.Username)
		utils.ValidationError(w, "Invalid username or password")
		return
	}
	if err != nil {
		logger.Error("Error retrieving user:", err)
		utils.ServerError(w, "Error during login")
		return
	}

	// 驗證密碼
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(input.Password)); err != nil {
		logger.Warn("Invalid password for user:", input.Username)
		utils.ValidationError(w, "Invalid username or password")
		return
	}

	// 生成 JWT token
	token, err := h.jwtService.GenerateToken(userID)
	if err != nil {
		logger.Error("Error generating token:", err)
		utils.ServerError(w, "Error during login")
		return
	}

	logger.Info("User logged in successfully:", input.Username)
	utils.SuccessResponse(w, map[string]string{"token": token})
}

// Logout 處理用戶登出
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Warn("Invalid method for Logout:", r.Method)
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Note: GetUserID function is not defined in the provided code,
	// so I assume it's defined in the middleware package.
	// If not, you need to define it or use a different way to get the user ID.
	userID, ok := middleware.GetUserID(r)
	if !ok {
		logger.Warn("Unauthorized access to Logout")
		utils.UnauthorizedError(w)
		return
	}

	logger.Info("User logged out successfully, UserID:", userID)
	utils.SuccessResponse(w, map[string]string{"message": "Logged out successfully"})
}
