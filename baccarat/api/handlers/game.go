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
		Player   float64 `json:"player"`
		Banker   float64 `json:"banker"`
		Tie      float64 `json:"tie"`
		LuckySix float64 `json:"luckySix"`
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

	// 檢查用戶餘額
	balance, err := db.GetUserBalance(userID)
	if err != nil {
		logger.Error("Error checking balance for user", userID, "Error:", err)
		utils.ServerError(w, "Error checking balance")
		return
	}

	if balance < totalBet {
		logger.Warn("Insufficient balance for user", userID)
		utils.ValidationError(w, "Insufficient balance")
		return
	}

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
	response := map[string]interface{}{
		"gameId":        gameID,
		"playerCards":   formatCards(g.PlayerHand.Cards),
		"bankerCards":   formatCards(g.BankerHand.Cards),
		"playerScore":   g.PlayerScore,
		"bankerScore":   g.BankerScore,
		"winner":        g.Winner,
		"isLuckySix":    g.IsLuckySix,
		"luckySixType":  g.LuckySixType,
		"totalPayout":   totalPayout,
		"payoutDetails": payouts,
	}

	logger.Info("Successfully processed game for user", userID, "GameID:", gameID)
	utils.SuccessResponse(w, response)
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
	Player   float64 `json:"player"`
	Banker   float64 `json:"banker"`
	Tie      float64 `json:"tie"`
	LuckySix float64 `json:"luckySix"`
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

	switch g.Winner {
	case "Player":
		if bets.Player > 0 {
			payouts["player"] = bets.Player * config.AppConfig.PlayerPayout
		}
	case "Banker":
		if bets.Banker > 0 {
			if g.IsLuckySix {
				if g.LuckySixType == "2cards" {
					payouts["banker"] = bets.Banker * config.AppConfig.BankerLucky6_2Cards
				} else {
					payouts["banker"] = bets.Banker * config.AppConfig.BankerLucky6_3Cards
				}
			} else {
				payouts["banker"] = bets.Banker * config.AppConfig.BankerPayout
			}
		}
	case "Tie":
		if bets.Tie > 0 {
			payouts["tie"] = bets.Tie * config.AppConfig.TiePayout
		}
	}

	// 處理幸運6
	if g.IsLuckySix && bets.LuckySix > 0 {
		if g.LuckySixType == "2cards" {
			payouts["luckySix"] = bets.LuckySix * config.AppConfig.Lucky6_2CardsPayout
		} else {
			payouts["luckySix"] = bets.LuckySix * config.AppConfig.Lucky6_3CardsPayout
		}
	}

	return payouts
}

// 計算總賠付
func calculateTotalPayout(payouts map[string]float64) float64 {
	total := 0.0
	for _, amount := range payouts {
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
	Player   float64 `json:"player"`
	Banker   float64 `json:"banker"`
	Tie      float64 `json:"tie"`
	LuckySix float64 `json:"luckySix"`
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
