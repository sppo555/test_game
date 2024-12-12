// package handlers

// import (
//     "database/sql"
//     "encoding/json"
//     "errors"
//     "fmt"
//     "net/http"
//     "regexp"
//     "time"

//     "baccarat/pkg/logger"
//     "github.com/google/uuid"
// )

// var (
//     validBetTypes = map[string]bool{
//         "Player":  true,
//         "Banker":  true,
//         "Tie":     true,
//         "Lucky6":  true,
//     }

//     gameIDRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
// )

// type AutoGameStatus struct {
//     GameID    string    `json:"game_id"`
//     Status    string    `json:"status"`
//     StartTime time.Time `json:"start_time"`
//     EndTime   time.Time `json:"end_time"`
// }

// type AutoGameBet struct {
//     GameID  string  `json:"game_id"`
//     BetType string  `json:"bet_type"`
//     Amount  float64 `json:"amount"`
// }

// type AutoGameHandler struct {
//     DB *sql.DB
// }

// func NewAutoGameHandler(db *sql.DB) *AutoGameHandler {
//     return &AutoGameHandler{
//         DB: db,
//     }
// }

// func (h *AutoGameHandler) validateBet(bet *AutoGameBet) error {
//     if !gameIDRegex.MatchString(bet.GameID) {
//         return errors.New("invalid game ID format")
//     }

//     if !validBetTypes[bet.BetType] {
//         return fmt.Errorf("invalid bet type: %s", bet.BetType)
//     }

//     if bet.Amount <= 0 || bet.Amount > 10000 { // 假設最大下注限額為10000
//         return errors.New("bet amount must be between 0 and 10000")
//     }

//     return nil
// }

// func (h *AutoGameHandler) GetAutoGameStatus(w http.ResponseWriter, r *http.Request) {
//     if r.Method != http.MethodGet {
//         http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//         return
//     }

//     var status AutoGameStatus
//     err := h.DB.QueryRow(`
//         SELECT game_id, game_status, betting_start_time, betting_end_time
//         FROM auto_game_records
//         WHERE game_status IN ('pending', 'betting', 'closed')
//         ORDER BY created_at DESC LIMIT 1
//     `).Scan(&status.GameID, &status.Status, &status.StartTime, &status.EndTime)

//     if err == sql.ErrNoRows {
//         json.NewEncoder(w).Encode(map[string]string{
//             "status": "no_active_game",
//         })
//         return
//     }

//     if err != nil {
//         http.Error(w, "Database error", http.StatusInternalServerError)
//         logger.Error("Database error in GetAutoGameStatus:", err)
//         return
//     }

//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(status)
// }

// func (h *AutoGameHandler) PlaceAutoBet(w http.ResponseWriter, r *http.Request) {
//     if r.Method != http.MethodPost {
//         http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//         return
//     }

//     // 安全的用戶ID檢查
//     userIDValue := r.Context().Value("userID")
//     if userIDValue == nil {
//         http.Error(w, "Unauthorized", http.StatusUnauthorized)
//         logger.Warn("Unauthorized access to PlaceAutoBet")
//         return
//     }
//     userID, ok := userIDValue.(int)
//     if !ok {
//         http.Error(w, "Invalid user ID", http.StatusInternalServerError)
//         logger.Error("Invalid user ID type in PlaceAutoBet")
//         return
//     }

//     // 找出當前活躍的遊戲ID
//     var gameID string
//     var gameStatus string
//     err := h.DB.QueryRow(`
//         SELECT game_id, game_status FROM auto_game_records
//         WHERE game_status = 'betting'
//         ORDER BY created_at DESC
//         LIMIT 1
//     `).Scan(&gameID, &gameStatus)

//     if err == sql.ErrNoRows {
//         // 如果沒有活躍遊戲，自動創建新遊戲
//         gameID = uuid.New().String()
//         _, err = h.DB.Exec(`
//             INSERT INTO auto_game_records (game_id, game_status, betting_start_time)
//             VALUES (?, 'betting', NOW())
//         `, gameID)
//         if err != nil {
//             http.Error(w, "Failed to create auto game", http.StatusInternalServerError)
//             logger.Error("Failed to create auto game:", err)
//             return
//         }
//     } else if gameStatus != "betting" {
//         http.Error(w, "Betting is not open", http.StatusBadRequest)
//         logger.Warn("Attempted to bet on a non-betting game")
//         return
//     }

//     var bet AutoGameBet
//     if err := json.NewDecoder(r.Body).Decode(&bet); err != nil {
//         http.Error(w, "Invalid request body", http.StatusBadRequest)
//         logger.Warn("Invalid request body in PlaceAutoBet:", err)
//         return
//     }

//     // 使用找到/創建的遊戲ID
//     bet.GameID = gameID

//     // 驗證下注信息
//     if err := h.validateBet(&bet); err != nil {
//         http.Error(w, err.Error(), http.StatusBadRequest)
//         logger.Warn("Invalid bet in PlaceAutoBet:", err)
//         return
//     }

//     // 驗證用戶餘額
//     var balance float64
//     err = h.DB.QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&balance)
//     if err != nil {
//         http.Error(w, "Database error", http.StatusInternalServerError)
//         logger.Error("Database error in balance check:", err)
//         return
//     }

//     if balance < bet.Amount {
//         http.Error(w, "Insufficient balance", http.StatusBadRequest)
//         logger.Warn("Insufficient balance for user", userID)
//         return
//     }

//     // 開始事務
//     tx, err := h.DB.Begin()
//     if err != nil {
//         http.Error(w, "Transaction error", http.StatusInternalServerError)
//         logger.Error("Transaction start error:", err)
//         return
//     }
//     defer tx.Rollback() // 確保事務最終會被回滾，除非顯式提交

//     // 扣除用戶餘額
//     _, err = tx.Exec("UPDATE users SET balance = balance - ? WHERE id = ?", bet.Amount, userID)
//     if err != nil {
//         http.Error(w, "Update balance failed", http.StatusInternalServerError)
//         logger.Error("Balance update error:", err)
//         return
//     }

//     // 記錄下注
//     _, err = tx.Exec(`
//         INSERT INTO auto_game_bets (game_id, user_id, bet_type, amount, status)
//         VALUES (?, ?, ?, ?, 'pending')
//     `, bet.GameID, userID, bet.BetType, bet.Amount)

//     if err != nil {
//         http.Error(w, "Record bet failed", http.StatusInternalServerError)
//         logger.Error("Bet record error:", err)
//         return
//     }

//     // 提交事務
//     if err = tx.Commit(); err != nil {
//         http.Error(w, "Commit transaction failed", http.StatusInternalServerError)
//         logger.Error("Transaction commit error:", err)
//         return
//     }

//     // 返回成功響應
//     response := map[string]interface{}{
//         "game_id": bet.GameID,
//         "bet_type": bet.BetType,
//         "amount": bet.Amount,
//         "message": "Bet placed successfully",
//     }

//     w.Header().Set("Content-Type", "application/json")
//     json.NewEncoder(w).Encode(response)

//     logger.Info("Successfully placed auto bet for user", userID, "Game ID:", bet.GameID)
// }
