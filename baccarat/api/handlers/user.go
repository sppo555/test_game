package handlers

import (
	"baccarat/api/middleware"
	"baccarat/db"
	"baccarat/pkg/utils"
	"baccarat/pkg/validation"
	"database/sql"
	"net/http"
	"strconv"
)

type UserHandler struct {
	db *sql.DB
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
		utils.UnauthorizedError(w)
		return
	}

	balance, err := db.GetUserBalance(userID)
	if err != nil {
		utils.ServerError(w, "Error retrieving balance")
		return
	}

	utils.SuccessResponse(w, map[string]float64{"balance": balance})
}

// Deposit 處理用戶存款
func (h *UserHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.UnauthorizedError(w)
		return
	}

	amountStr := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		utils.ValidationError(w, "Invalid amount format")
		return
	}

	// 驗證存款金額
	if err := validation.ValidateAmount(amount); err != nil {
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
		utils.ServerError(w, "Error processing deposit")
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": "Deposit successful",
		"amount":  amount,
	})
}

// GetTransactions 獲取用戶交易記錄
func (h *UserHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
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
		utils.ServerError(w, "Error retrieving transactions")
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var amount float64
		var transactionType, createdAt string
		if err := rows.Scan(&amount, &transactionType, &createdAt); err != nil {
			utils.ServerError(w, "Error scanning transaction")
			return
		}
		transactions = append(transactions, map[string]interface{}{
			"amount":          amount,
			"transactionType": transactionType,
			"createdAt":       createdAt,
		})
	}

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

	utils.SuccessResponse(w, map[string]interface{}{
		"page":     page,
		"pageSize": pageSize,
		"bets":     bets,
	})
}
