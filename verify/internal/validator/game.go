// game_logic.go
package validator

import (
	"math/rand"
	"time"
)

// Card 代表一張牌
type Card struct {
	Suit  int // 0: 黑桃, 1: 紅心, 2: 方塊, 3: 梅花
	Value int // 1-13 (A-K)
}

// Hand 代表一手牌
type Hand struct {
	Cards []Card
}

// Game 代表百家樂遊戲
type Game struct {
	PlayerHand   Hand
	BankerHand   Hand
	PlayerScore  int
	BankerScore  int
	Winner       string
	IsLuckySix   bool
	LuckySixType string // "2cards" 或 "3cards"
	Deck         []Card
}

// NewGame 創建新遊戲
func NewGame() *Game {
	g := &Game{
		Deck: make([]Card, 52),
	}
	g.initializeDeck()
	g.shuffleDeck()
	return g
}

func (g *Game) initializeDeck() {
	index := 0
	for suit := 0; suit < 4; suit++ {
		for value := 1; value <= 13; value++ {
			g.Deck[index] = Card{Suit: suit, Value: value}
			index++
		}
	}
}

func (g *Game) shuffleDeck() {
	rand.Seed(time.Now().UnixNano())
	for i := len(g.Deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		g.Deck[i], g.Deck[j] = g.Deck[j], g.Deck[i]
	}
}

func (g *Game) Deal() {
	// 發兩張牌給閒家和莊家
	g.PlayerHand.Cards = append(g.PlayerHand.Cards, g.Deck[0], g.Deck[1])
	g.BankerHand.Cards = append(g.BankerHand.Cards, g.Deck[2], g.Deck[3])
	g.Deck = g.Deck[4:]

	// 計算初始點數
	g.PlayerScore = calculateScore(g.PlayerHand.Cards)
	g.BankerScore = calculateScore(g.BankerHand.Cards)
}

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

func (g *Game) DealThirdCard() {
	// 如果閒家需要補牌
	if g.PlayerScore <= 5 {
		playerThirdCard := g.Deck[0]
		g.PlayerHand.Cards = append(g.PlayerHand.Cards, playerThirdCard)
		g.Deck = g.Deck[1:]
		g.PlayerScore = calculateScore(g.PlayerHand.Cards)

		// 根據閒家第三張牌的值，決定莊家是否補牌
		playerThirdValue := playerThirdCard.Value
		if needBankerThirdCard(g.BankerScore, playerThirdValue) {
			bankerThirdCard := g.Deck[0]
			g.BankerHand.Cards = append(g.BankerHand.Cards, bankerThirdCard)
			g.Deck = g.Deck[1:]
			g.BankerScore = calculateScore(g.BankerHand.Cards)
		}
	} else if g.BankerScore <= 5 { // 閒家不補牌，莊家點數 <=5 時補牌
		bankerThirdCard := g.Deck[0]
		g.BankerHand.Cards = append(g.BankerHand.Cards, bankerThirdCard)
		g.Deck = g.Deck[1:]
		g.BankerScore = calculateScore(g.BankerHand.Cards)
	}
}

func (g *Game) DetermineWinner() {
	// 決定贏家
	if g.PlayerScore > g.BankerScore {
		g.Winner = "Player"
	} else if g.BankerScore > g.PlayerScore {
		g.Winner = "Banker"
	} else {
		g.Winner = "Tie"
	}

	// 檢查幸運6：莊家6點且贏得遊戲（不是和局）
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
}

func calculateScore(cards []Card) int {
	total := 0
	for _, card := range cards {
		value := card.Value
		if value > 9 {
			value = 0
		}
		total += value
	}
	return total % 10
}

func needBankerThirdCard(bankerScore int, playerThirdCard int) bool {
	switch bankerScore {
	case 0, 1, 2:
		return true
	case 3:
		return playerThirdCard != 8
	case 4:
		return playerThirdCard >= 2 && playerThirdCard <= 7
	case 5:
		return playerThirdCard >= 4 && playerThirdCard <= 7
	case 6:
		return playerThirdCard == 6 || playerThirdCard == 7
	default:
		return false
	}
}
