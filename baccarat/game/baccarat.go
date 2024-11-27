package game

// Hand 代表一手牌
type Hand struct {
	Cards []Card
}

// Game 代表百家樂遊戲
type Game struct {
	Deck          *Deck
	PlayerHand    Hand
	BankerHand    Hand
	PlayerScore   int
	BankerScore   int
	Winner        string
	IsLuckySix    bool
}

// NewGame 創建新遊戲
func NewGame() *Game {
	deck := NewDeck()
	deck.Shuffle()
	return &Game{
		Deck: deck,
	}
}

// Deal 發牌
func (g *Game) Deal() {
	// 發給閒家第一張牌
	g.PlayerHand.Cards = append(g.PlayerHand.Cards, g.Deck.DrawCard())
	// 發給莊家第一張牌
	g.BankerHand.Cards = append(g.BankerHand.Cards, g.Deck.DrawCard())
	// 發給閒家第二張牌
	g.PlayerHand.Cards = append(g.PlayerHand.Cards, g.Deck.DrawCard())
	// 發給莊家第二張牌
	g.BankerHand.Cards = append(g.BankerHand.Cards, g.Deck.DrawCard())

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
	// 檢查幸運6
	g.IsLuckySix = g.BankerScore == 6 && len(g.BankerHand.Cards) == 2

	if g.PlayerScore > g.BankerScore {
		g.Winner = "Player"
	} else if g.BankerScore > g.PlayerScore {
		g.Winner = "Banker"
	} else {
		g.Winner = "Tie"
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
