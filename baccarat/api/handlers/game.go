package handlers

import (
	"baccarat/api/middleware"
	"baccarat/config"
	"baccarat/db"
	"baccarat/internal/game"
	"baccarat/pkg/logger"
	"baccarat/pkg/utils"
	"baccarat/pkg/validation"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type GameHandler struct {
	db *sql.DB
}

func NewGameHandler(db *sql.DB) *GameHandler {
	return &GameHandler{
		db: db,
	}
}

// PlayGame 進行一局遊戲
func (h *GameHandler) PlayGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		logger.Warn("Unauthorized access to PlayGame")
		utils.UnauthorizedError(w)
		return
	}

	logger.Debug("Starting game for user", userID)

	// 解析投注信息
	var bets struct {
		Player    float64 `json:"player"`
		Banker    float64 `json:"banker"`
		Tie       float64 `json:"tie"`
		LuckySix  float64 `json:"luckySix"`
		RUN_TIMES string  `json:"RUN_TIMES"`
	}

	if err := json.NewDecoder(r.Body).Decode(&bets); err != nil {
		logger.Warn("Invalid bet data")
		utils.ValidationError(w, "Invalid bet data")
		return
	}

	// 檢查是否同時下注莊家和閒家
	if bets.Player > 0 && bets.Banker > 0 {
		logger.Warn("Cannot bet on both Player and Banker")
		utils.ValidationError(w, "不能同時下注莊家和閒家")
		return
	}

	// 驗證每個投注金額
	if bets.Player > 0 {
		if err := validation.ValidateAmount(bets.Player); err != nil {
			logger.Warn("Invalid player bet for user", userID, "Error:", err)
			utils.ValidationError(w, "Invalid player bet: "+err.Error())
			return
		}
	}
	if bets.Banker > 0 {
		if err := validation.ValidateAmount(bets.Banker); err != nil {
			logger.Warn("Invalid banker bet for user", userID, "Error:", err)
			utils.ValidationError(w, "Invalid banker bet: "+err.Error())
			return
		}
	}
	if bets.Tie > 0 {
		if err := validation.ValidateAmount(bets.Tie); err != nil {
			logger.Warn("Invalid tie bet for user", userID, "Error:", err)
			utils.ValidationError(w, "Invalid tie bet: "+err.Error())
			return
		}
	}
	if bets.LuckySix > 0 {
		if err := validation.ValidateAmount(bets.LuckySix); err != nil {
			logger.Warn("Invalid lucky six bet for user", userID, "Error:", err)
			utils.ValidationError(w, "Invalid lucky six bet: "+err.Error())
			return
		}
	}

	// 計算總投注額
	totalBet := bets.Player + bets.Banker + bets.Tie + bets.LuckySix
	if totalBet <= 0 {
		logger.Warn("No bets placed for user", userID)
		utils.ValidationError(w, "No bets placed")
		return
	}

	// 設置運行次數，默認為1次
	runTimes := 1
	if bets.RUN_TIMES != "" {
		var err error
		runTimes, err = strconv.Atoi(bets.RUN_TIMES)
		if err != nil || runTimes <= 0 {
			logger.Warn("Invalid RUN_TIMES value for user", userID)
			utils.ValidationError(w, "Invalid RUN_TIMES value")
			return
		}
	}

	// 檢查用戶餘額是否足夠支付所有運行次數的投注
	balance, err := db.GetUserBalance(userID)
	if err != nil {
		logger.Error("Error checking balance for user", userID, "Error:", err)
		utils.ServerError(w, "Error checking balance")
		return
	}

	if balance < totalBet*float64(runTimes) {
		logger.Warn("Insufficient balance for user", userID)
		utils.ValidationError(w, "Insufficient balance for all runs")
		return
	}

	// 存儲所有遊戲結果
	var allGameResults []map[string]interface{}

	// 循環執行指定次數的遊戲
	for i := 0; i < runTimes; i++ {
		// 定義遊戲相關變量
		var (
			gameID      string
			g           *game.Game
			totalPayout float64
			payouts     map[string]float64
		)

		// 開始事務
		err = db.Transaction(func(tx *sql.Tx) error {
			// 扣除投注金額
			if err := db.UpdateUserBalance(tx, userID, -totalBet); err != nil {
				return err
			}

			// 進行遊戲
			gameID = uuid.New().String()
			g = game.NewGame()
			g.Deal()
			g.DealThirdCard()
			g.DetermineWinner()

			// 計算賠付
			payouts = calculatePayouts(g, bets)
			totalPayout = calculateTotalPayout(payouts)

			// 更新用戶餘額（加上賠付金額）
			if totalPayout > 0 {
				if err := db.UpdateUserBalance(tx, userID, totalPayout); err != nil {
					return err
				}
			}

			// 保存遊戲記錄
			if err := saveGameRecord(tx, g, gameID, payouts); err != nil {
				return err
			}

			// 保存投注記錄
			if err := saveBets(tx, userID, gameID, bets); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			logger.Error("Error processing game for user", userID, "Error:", err)
			utils.ServerError(w, "Error processing game")
			return
		}

		// 返回遊戲結果
		gameResult := map[string]interface{}{
			"gameId":       gameID,
			"playerCards":  formatCards(g.PlayerHand.Cards),
			"bankerCards":  formatCards(g.BankerHand.Cards),
			"playerScore":  g.PlayerScore,
			"bankerScore":  g.BankerScore,
			"winner":       g.Winner,
			"isLuckySix":   g.IsLuckySix,
			"luckySixType": g.LuckySixType,
			// 下注明細
			"bets": map[string]float64{
				"player":   bets.Player,
				"banker":   bets.Banker,
				"tie":      bets.Tie,
				"luckySix": bets.LuckySix,
			},
			// 賠付明細（不含本金）
			"payouts": map[string]float64{
				"player":   payouts["player"],
				"banker":   payouts["banker"],
				"tie":      payouts["tie"],
				"luckySix": payouts["luckySix"],
			},
			// 本金返還明細
			"principalReturns": map[string]float64{
				"player":   payouts["player_principal"],
				"banker":   payouts["banker_principal"],
				"tie":      payouts["tie_principal"],
				"luckySix": payouts["luckySix_principal"],
			},
			"totalBet":    totalBet,
			"totalPayout": totalPayout,
		}

		allGameResults = append(allGameResults, gameResult)

		// 每次遊戲完成後立即輸出日志
		logger.Info("Successfully processed game for user", userID, "GameID:", gameID, "Round:", i+1, "of", runTimes)
	}

	utils.SuccessResponse(w, allGameResults)
}

// GameDetailsRequest 遊戲詳情請求結構
type GameDetailsRequest struct {
	GameID string `json:"game_id"`
}

// GetGameDetails 獲取遊戲詳情
func (h *GameHandler) GetGameDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析JSON請求
	var req struct {
		GameID string `json:"game_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// 驗證game_id
	if req.GameID == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Missing game_id")
		return
	}

	// 獲取遊戲詳情
	result, err := db.GetGameDetails(req.GameID)
	if err != nil {
		logger.Error("獲取遊戲詳情失敗:", err)
		utils.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(w, result)
}

// 格式化牌
func formatCards(cards []game.Card) []string {
	result := make([]string, len(cards))
	for i, card := range cards {
		result[i] = formatCard(card)
	}
	return result
}

// 格式化單張牌
func formatCard(card game.Card) string {
	suits := []string{"S", "H", "D", "C"} // Spades, Hearts, Diamonds, Clubs
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	return suits[card.Suit] + values[card.Value-1] // 例如：S7, CK
}

// 格式化牌組為字符串
func formatCardsToString(cards []string) string {
	return strings.Join(cards, ",") // 例如：S7,CK
}

// 計算賠付
func calculatePayouts(g *game.Game, bets struct {
	Player    float64 `json:"player"`
	Banker    float64 `json:"banker"`
	Tie       float64 `json:"tie"`
	LuckySix  float64 `json:"luckySix"`
	RUN_TIMES string  `json:"RUN_TIMES"`
}) map[string]float64 {
	payouts := make(map[string]float64)

	// 記錄投注金額
	if bets.Player > 0 {
		payouts["player_bet"] = bets.Player
	}
	if bets.Banker > 0 {
		payouts["banker_bet"] = bets.Banker
	}
	if bets.Tie > 0 {
		payouts["tie_bet"] = bets.Tie
	}
	if bets.LuckySix > 0 {
		payouts["luckySix_bet"] = bets.LuckySix
	}

	// 初始化本金返還（預設為0）
	payouts["player_principal"] = 0
	payouts["banker_principal"] = 0
	payouts["tie_principal"] = 0
	payouts["luckySix_principal"] = 0

	switch g.Winner {
	case "Player":
		if bets.Player > 0 {
			// 贏家賠付已包含本金
			payouts["player"] = bets.Player * (1 + config.AppConfig.PlayerPayout)
		}
	case "Banker":
		if bets.Banker > 0 {
			if g.IsLuckySix {
				if g.LuckySixType == "2cards" {
					// 幸運6時莊家賠付0.5倍
					payouts["banker"] = bets.Banker * (1 + config.AppConfig.BankerLucky6_2Cards)
				} else {
					// 幸運6時莊家賠付0.95倍
					payouts["banker"] = bets.Banker * (1 + config.AppConfig.BankerLucky6_3Cards)
				}
			} else {
				payouts["banker"] = bets.Banker * (1 + config.AppConfig.BankerPayout)
			}
		}
	case "Tie":
		if bets.Tie > 0 {
			// 和局贏家賠付不包含本金
			payouts["tie"] = bets.Tie * config.AppConfig.TiePayout
		}
		// 和局時，只有莊閒和幸運6需要返還本金，和局贏家不需要返還本金
		payouts["player_principal"] = bets.Player
		payouts["banker_principal"] = bets.Banker
		payouts["luckySix_principal"] = bets.LuckySix
	}

	// 處理幸運6
	if g.IsLuckySix && bets.LuckySix > 0 {
		if g.LuckySixType == "2cards" {
			// 幸運6賠付已包含本金，兩張牌12倍
			payouts["luckySix"] = bets.LuckySix * config.AppConfig.Lucky6_2CardsPayout
		} else {
			// 幸運6賠付已包含本金，三張牌20倍
			payouts["luckySix"] = bets.LuckySix * config.AppConfig.Lucky6_3CardsPayout
		}
	}

	return payouts
}

// 計算總賠付
func calculateTotalPayout(payouts map[string]float64) float64 {
	total := 0.0

	// 計算賠付（已包含本金）
	if amount, ok := payouts["player"]; ok {
		total += amount
	}
	if amount, ok := payouts["banker"]; ok {
		total += amount
	}
	if amount, ok := payouts["tie"]; ok {
		total += amount
	}
	if amount, ok := payouts["luckySix"]; ok {
		total += amount
	}

	// 和局時的本金返還
	if amount, ok := payouts["player_principal"]; ok {
		total += amount
	}
	if amount, ok := payouts["banker_principal"]; ok {
		total += amount
	}
	if amount, ok := payouts["luckySix_principal"]; ok {
		total += amount
	}

	return total
}

// 保存遊戲記錄
func saveGameRecord(tx *sql.Tx, g *game.Game, gameID string, payouts map[string]float64) error {
	// 格式化初始牌
	playerInitialCards := formatCards(g.PlayerHand.Cards[:2])
	bankerInitialCards := formatCards(g.BankerHand.Cards[:2])

	// 處理第三張牌
	var playerThirdCard, bankerThirdCard sql.NullString
	if len(g.PlayerHand.Cards) > 2 {
		playerThirdCard = sql.NullString{
			String: formatCard(g.PlayerHand.Cards[2]),
			Valid:  true,
		}
	}
	if len(g.BankerHand.Cards) > 2 {
		bankerThirdCard = sql.NullString{
			String: formatCard(g.BankerHand.Cards[2]),
			Valid:  true,
		}
	}

	// 處理幸運6類型
	var luckySixType sql.NullString
	if g.IsLuckySix {
		luckySixType = sql.NullString{
			String: g.LuckySixType,
			Valid:  true,
		}
	}

	// 保存遊戲記錄
	if err := db.SaveGameRecord(
		gameID,
		formatCardsToString(playerInitialCards),
		formatCardsToString(bankerInitialCards),
		playerThirdCard,
		bankerThirdCard,
		g.PlayerScore,
		g.BankerScore,
		g.Winner,
		g.IsLuckySix,
		luckySixType,
		payouts,
	); err != nil {
		return err
	}

	return nil
}

// 保存投注記錄
func saveBets(tx *sql.Tx, userID int, gameID string, bets struct {
	Player    float64 `json:"player"`
	Banker    float64 `json:"banker"`
	Tie       float64 `json:"tie"`
	LuckySix  float64 `json:"luckySix"`
	RUN_TIMES string  `json:"RUN_TIMES"`
}) error {
	// 保存每個投注
	if bets.Player > 0 {
		if err := db.SaveBet(tx, userID, gameID, bets.Player, "player"); err != nil {
			return err
		}
	}
	if bets.Banker > 0 {
		if err := db.SaveBet(tx, userID, gameID, bets.Banker, "banker"); err != nil {
			return err
		}
	}
	if bets.Tie > 0 {
		if err := db.SaveBet(tx, userID, gameID, bets.Tie, "tie"); err != nil {
			return err
		}
	}
	if bets.LuckySix > 0 {
		if err := db.SaveBet(tx, userID, gameID, bets.LuckySix, "luckySix"); err != nil {
			return err
		}
	}

	return nil
}
