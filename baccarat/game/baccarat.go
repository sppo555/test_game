package game

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
	// 確保賠率鍵名與數據庫一致
	g.Payouts["Player"] = g.Payouts["PAYOUT_PLAYER"]
	g.Payouts["Banker"] = g.Payouts["PAYOUT_BANKER"]
	g.Payouts["Tie"] = g.Payouts["PAYOUT_TIE"]
	g.Payouts["LuckySix"] = 0 // 初始化 LuckySix 賠率

	// 根據遊戲結果設置賠率
	switch g.Winner {
	case "Player":
		g.Payouts["Player"] = g.Payouts["PAYOUT_PLAYER"]
	case "Banker":
		if g.IsLuckySix {
			if g.LuckySixType == "2cards" {
				g.Payouts["Banker"] = g.Payouts["PAYOUT_BANKER_LUCKY6_2CARDS"]
				g.Payouts["LuckySix"] = g.Payouts["PAYOUT_LUCKY6_2CARDS"]
			} else {
				g.Payouts["Banker"] = g.Payouts["PAYOUT_BANKER_LUCKY6_3CARDS"]
				g.Payouts["LuckySix"] = g.Payouts["PAYOUT_LUCKY6_3CARDS"]
			}
		} else {
			g.Payouts["Banker"] = g.Payouts["PAYOUT_BANKER"]
		}
	case "Tie":
		g.Payouts["Tie"] = g.Payouts["PAYOUT_TIE"]
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
