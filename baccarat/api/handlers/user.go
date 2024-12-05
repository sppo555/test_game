package handlers

import (
	"baccarat/api/middleware"
	"baccarat/db"
	"baccarat/pkg/logger"
	"baccarat/pkg/utils"
	"baccarat/pkg/validation"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

type UserHandler struct {
	db *sql.DB
}

type DepositRequest struct {
	Amount string `json:"amount"`
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

// GetBalance 獲取用戶餘額
func (h *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		logger.Warn("Unauthorized access to GetBalance")
		utils.UnauthorizedError(w)
		return
	}

	logger.Debug("Retrieving balance for user", userID)
	balance, err := db.GetUserBalance(userID)
	if err != nil {
		logger.Error("Error retrieving balance for user", userID, "Error:", err)
		utils.ServerError(w, "Error retrieving balance")
		return
	}

	logger.Info("Successfully retrieved balance for user", userID)
	utils.SuccessResponse(w, map[string]float64{"balance": balance})
}

// Deposit 處理用戶存款
func (h *UserHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Warn("Invalid method for Deposit:", r.Method)
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		logger.Warn("Unauthorized access to Deposit")
		utils.UnauthorizedError(w)
		return
	}

	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Invalid request body for Deposit:", err)
		utils.ValidationError(w, "Invalid request format")
		return
	}
	defer r.Body.Close()

	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		logger.Warn("Invalid amount format for Deposit:", req.Amount)
		utils.ValidationError(w, "Invalid amount format")
		return
	}

	// 驗證存款金額
	if err := validation.ValidateAmount(amount); err != nil {
		logger.Warn("Invalid amount for Deposit:", amount, "Error:", err)
		utils.ValidationError(w, err.Error())
		return
	}

	// 使用事務處理存款
	err = db.Transaction(func(tx *sql.Tx) error {
		// 更新餘額
		if err := db.UpdateUserBalance(tx, userID, amount); err != nil {
			return err
		}

		// 記錄交易
		if err := db.SaveTransaction(tx, userID, amount, "deposit"); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error("Error processing deposit for user", userID, "Error:", err)
		utils.ServerError(w, "Error processing deposit")
		return
	}

	logger.Info("Successfully processed deposit for user", userID)
	utils.SuccessResponse(w, map[string]interface{}{
		"message": "Deposit successful",
		"amount":  amount,
	})
}

// GetTransactions 獲取用戶交易記錄
func (h *UserHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		logger.Warn("Unauthorized access to GetTransactions")
		utils.UnauthorizedError(w)
		return
	}

	// 獲取分頁參數
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	offset := (page - 1) * pageSize

	rows, err := h.db.Query(`
		SELECT amount, transaction_type, created_at 
		FROM transactions 
		WHERE user_id = ? 
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, 
		userID, pageSize, offset)
	if err != nil {
		logger.Error("Error retrieving transactions for user", userID, "Error:", err)
		utils.ServerError(w, "Error retrieving transactions")
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var amount float64
		var transactionType, createdAt string
		if err := rows.Scan(&amount, &transactionType, &createdAt); err != nil {
			logger.Error("Error scanning transaction for user", userID, "Error:", err)
			utils.ServerError(w, "Error scanning transaction")
			return
		}
		transactions = append(transactions, map[string]interface{}{
			"amount":          amount,
			"transactionType": transactionType,
			"createdAt":       createdAt,
		})
	}

	logger.Info("Successfully retrieved transactions for user", userID)
	utils.SuccessResponse(w, map[string]interface{}{
		"page":         page,
		"pageSize":     pageSize,
		"transactions": transactions,
	})
}

// GetBets 獲取用戶投注記錄
func (h *UserHandler) GetBets(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		logger.Warn("Unauthorized access to GetBets")
		utils.UnauthorizedError(w)
		return
	}

	// 獲取分頁參數
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	offset := (page - 1) * pageSize

	rows, err := h.db.Query(`
		SELECT b.game_id, b.bet_amount, b.bet_type, b.created_at,
			   g.winner, g.is_lucky_six, g.lucky_six_type
		FROM bets b
		LEFT JOIN game_records g ON b.game_id = g.game_id
		WHERE b.user_id = ? 
		ORDER BY b.created_at DESC
		LIMIT ? OFFSET ?`, 
		userID, pageSize, offset)
	if err != nil {
		logger.Error("Error retrieving bets for user", userID, "Error:", err)
		utils.ServerError(w, "Error retrieving bets")
		return
	}
	defer rows.Close()

	var bets []map[string]interface{}
	for rows.Next() {
		var gameID, betType, createdAt, winner string
		var betAmount float64
		var isLuckySix bool
		var luckySixType sql.NullString

		if err := rows.Scan(&gameID, &betAmount, &betType, &createdAt,
			&winner, &isLuckySix, &luckySixType); err != nil {
			logger.Error("Error scanning bet for user", userID, "Error:", err)
			utils.ServerError(w, "Error scanning bet")
			return
		}

		bet := map[string]interface{}{
			"gameId":    gameID,
			"amount":    betAmount,
			"betType":   betType,
			"createdAt": createdAt,
			"winner":    winner,
		}

		if isLuckySix {
			bet["isLuckySix"] = true
			if luckySixType.Valid {
				bet["luckySixType"] = luckySixType.String
			}
		}

		bets = append(bets, bet)
	}

	logger.Info("Successfully retrieved bets for user", userID)
	utils.SuccessResponse(w, map[string]interface{}{
		"page":     page,
		"pageSize": pageSize,
		"bets":     bets,
	})
}
