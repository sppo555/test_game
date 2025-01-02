package game

import (
    "baccarat/config"
)

// Hand 代表一手牌
type Hand struct {
	Cards []Card
}

// Game 代表百家樂遊戲
type Game struct {
	Deck         *Deck
	PlayerHand   Hand
	BankerHand   Hand
	PlayerScore  int
	BankerScore  int
	Winner       string
	IsLuckySix   bool
	LuckySixType string             // "2cards" 或 "3cards"
	Payouts      map[string]float64 // 各種投注的賠率
}

// NewGame 創建新遊戲
func NewGame() *Game {
	deck := NewDeck()
	deck.Shuffle()
	return &Game{
		Deck:    deck,
		Payouts: make(map[string]float64),
	}
}

// Deal 發牌
func (g *Game) Deal() {
	// 初始發牌：閒家和莊家各發兩張牌
	g.PlayerHand.Cards = []Card{g.Deck.DrawCard(), g.Deck.DrawCard()}
	g.BankerHand.Cards = []Card{g.Deck.DrawCard(), g.Deck.DrawCard()}
	g.calculateScores()
}

// calculateScores 計算點數
func (g *Game) calculateScores() {
	g.PlayerScore = 0
	g.BankerScore = 0

	// 計算閒家點數
	for _, card := range g.PlayerHand.Cards {
		g.PlayerScore = (g.PlayerScore + card.GetCardValue()) % 10
	}

	// 計算莊家點數
	for _, card := range g.BankerHand.Cards {
		g.BankerScore = (g.BankerScore + card.GetCardValue()) % 10
	}
}

// NeedThirdCard 判斷是否需要第三張牌
func (g *Game) NeedThirdCard() bool {
	// 如果任一方為8或9點，不需要補牌
	if g.PlayerScore >= 8 || g.BankerScore >= 8 {
		return false
	}

	// 閒家補牌規則
	if g.PlayerScore <= 5 {
		return true
	}

	return false
}

// DealThirdCard 發第三張牌
func (g *Game) DealThirdCard() {
	// 閒家補牌
	if g.PlayerScore <= 5 {
		playerThirdCard := g.Deck.DrawCard()
		g.PlayerHand.Cards = append(g.PlayerHand.Cards, playerThirdCard)
		g.calculateScores()

		// 莊家補牌規則
		playerThirdValue := playerThirdCard.GetCardValue()
		if g.shouldBankerDrawThird(playerThirdValue) {
			g.BankerHand.Cards = append(g.BankerHand.Cards, g.Deck.DrawCard())
			g.calculateScores()
		}
	} else if g.BankerScore <= 5 { // 閒家不補牌，莊家點數<=5時補牌
		g.BankerHand.Cards = append(g.BankerHand.Cards, g.Deck.DrawCard())
		g.calculateScores()
	}
}

// shouldBankerDrawThird 判斷莊家是否需要補第三張牌
func (g *Game) shouldBankerDrawThird(playerThirdValue int) bool {
	switch g.BankerScore {
	case 0, 1, 2:
		return true
	case 3:
		return playerThirdValue != 8
	case 4:
		return playerThirdValue >= 2 && playerThirdValue <= 7
	case 5:
		return playerThirdValue >= 4 && playerThirdValue <= 7
	case 6:
		return playerThirdValue == 6 || playerThirdValue == 7
	default:
		return false
	}
}

// DetermineWinner 判斷勝負
func (g *Game) DetermineWinner() {
	// 先判斷勝負
	if g.PlayerScore > g.BankerScore {
		g.Winner = "Player"
	} else if g.BankerScore > g.PlayerScore {
		g.Winner = "Banker"
	} else {
		g.Winner = "Tie"
	}

	// 再判斷幸運6（只有莊家贏且點數為6時才可能是幸運6）
	if g.Winner == "Banker" && g.BankerScore == 6 {
		g.IsLuckySix = true
		if len(g.BankerHand.Cards) == 2 {
			g.LuckySixType = "2cards"
		} else {
			g.LuckySixType = "3cards"
		}
	} else {
		g.IsLuckySix = false
		g.LuckySixType = ""
	}

	// 計算賠率
	g.CalculatePayouts()
}

// CalculatePayouts 計算賠率
func (g *Game) CalculatePayouts() {
    // 初始化賠率，所有賠率預設為0（表示輸掉全部押注）
    g.Payouts = make(map[string]float64)
    g.Payouts["Player"] = 0    // 閒家賠率（不含本金）
    g.Payouts["Banker"] = 0    // 莊家賠率（不含本金）
    g.Payouts["Tie"] = 0       // 和局賠率（不含本金）
    g.Payouts["LuckySix"] = 0  // 幸運6賠率（不含本金）

    // 根據遊戲結果設置賠率
    switch g.Winner {
    case "Player":
        // 閒家贏，只有押閒家的會贏錢，其他全輸
        g.Payouts["Player"] = config.AppConfig.PlayerPayout
        // 其他投注保持為0，表示輸掉全部押注
        
    case "Banker":
        // 莊家贏
        if g.IsLuckySix {
            // 如果是幸運6，莊家和幸運6都贏
            if g.LuckySixType == "2cards" {
                g.Payouts["Banker"] = config.AppConfig.BankerLucky6_2Cards
                g.Payouts["LuckySix"] = config.AppConfig.Lucky6_2CardsPayout
            } else {
                g.Payouts["Banker"] = config.AppConfig.BankerLucky6_3Cards
                g.Payouts["LuckySix"] = config.AppConfig.Lucky6_3CardsPayout
            }
        } else {
            // 普通莊家贏
            g.Payouts["Banker"] = config.AppConfig.BankerPayout
        }
        // 其他投注保持為0，表示輸掉全部押注
        
    case "Tie":
        // 和局情況
        g.Payouts["Tie"] = config.AppConfig.TiePayout  // 和局的賠率
        g.Payouts["Player"] = 1.0  // 閒家押注返回本金（不輸不贏）
        g.Payouts["Banker"] = 1.0  // 莊家押注返回本金（不輸不贏）
        g.Payouts["LuckySix"] = 1.0  // 幸運6押注也返回本金（因為不是莊家贏，所以不輸不贏）
    }
}

// Play 進行一局遊戲
func (g *Game) Play() {
	g.Deal()
	if g.NeedThirdCard() {
		g.DealThirdCard()
	}
	g.DetermineWinner()
}

// GetPlayerHand 获取闲家手牌
func (g *Game) GetPlayerHand() []Card {
	return g.PlayerHand.Cards
}

// GetBankerHand 获取庄家手牌
func (g *Game) GetBankerHand() []Card {
	return g.BankerHand.Cards
}

// GetPlayerScore 获取闲家点数
func (g *Game) GetPlayerScore() int {
	return g.PlayerScore
}

// GetBankerScore 获取庄家点数
func (g *Game) GetBankerScore() int {
	return g.BankerScore
}

// GetWinner 获取赢家
func (g *Game) GetWinner() string {
	return g.Winner
}

// GetIsLuckySix 获取是否幸运6
func (g *Game) GetIsLuckySix() bool {
	return g.IsLuckySix
}

// GetLuckySixType 获取幸运6类型
func (g *Game) GetLuckySixType() string {
	return g.LuckySixType
}

// GetPayouts 获取赔付
func (g *Game) GetPayouts(bets struct {
	Player    float64 `json:"player"`
	Banker    float64 `json:"banker"`
	Tie       float64 `json:"tie"`
	LuckySix  float64 `json:"luckySix"`
	RUN_TIMES string  `json:"RUN_TIMES"`
}) map[string]float64 {
	payouts := make(map[string]float64)

	// 记录投注金额
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

	// 初始化本金返还（预设为0）
	payouts["player_principal"] = 0
	payouts["banker_principal"] = 0
	payouts["tie_principal"] = 0
	payouts["luckySix_principal"] = 0

	switch g.Winner {
	case "Player":
		if bets.Player > 0 {
			payouts["player"] = bets.Player * (1 + config.AppConfig.PlayerPayout)
		}
	case "Banker":
		if bets.Banker > 0 {
			if g.IsLuckySix {
				if g.LuckySixType == "2cards" {
					payouts["banker"] = bets.Banker * (1 + config.AppConfig.BankerLucky6_2Cards)
				} else {
					payouts["banker"] = bets.Banker * (1 + config.AppConfig.BankerLucky6_3Cards)
				}
			} else {
				payouts["banker"] = bets.Banker * (1 + config.AppConfig.BankerPayout)
			}
		}
	case "Tie":
		if bets.Tie > 0 {
			payouts["tie"] = bets.Tie * config.AppConfig.TiePayout
		}
		payouts["player_principal"] = bets.Player
		payouts["banker_principal"] = bets.Banker
		payouts["luckySix_principal"] = bets.LuckySix
	}

	// 处理幸运6
	if g.IsLuckySix && bets.LuckySix > 0 {
		if g.LuckySixType == "2cards" {
			payouts["luckySix"] = bets.LuckySix * config.AppConfig.Lucky6_2CardsPayout
		} else {
			payouts["luckySix"] = bets.LuckySix * config.AppConfig.Lucky6_3CardsPayout
		}
	}

	return payouts
}

// GetTotalPayout 获取总赔付
func (g *Game) GetTotalPayout(payouts map[string]float64) float64 {
	total := 0.0

	// 计算赔付（已包含本金）
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

	// 和局时的本金返还
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