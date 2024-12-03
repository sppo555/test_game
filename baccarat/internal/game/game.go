package game

import (
	"math/rand"
	"time"
)

type Card struct {
	Suit  int // 0: 黑桃, 1: 紅心, 2: 方塊, 3: 梅花
	Value int // 1-13 (A-K)
}

type Hand struct {
	Cards []Card
}

type Game struct {
	PlayerHand   Hand
	BankerHand   Hand
	PlayerScore  int
	BankerScore  int
	Winner       string
	IsLuckySix   bool
	LuckySixType string
	Deck         []Card
}

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

func (g *Game) DealThirdCard() {
	playerScore := g.PlayerScore
	bankerScore := g.BankerScore

	// 檢查是否需要補牌
	if playerScore <= 5 {
		// 閒家補牌
		g.PlayerHand.Cards = append(g.PlayerHand.Cards, g.Deck[0])
		g.Deck = g.Deck[1:]
		g.PlayerScore = calculateScore(g.PlayerHand.Cards)

		// 根據閒家補牌結果決定莊家是否補牌
		playerThirdCard := g.PlayerHand.Cards[2].Value
		if needBankerThirdCard(bankerScore, playerThirdCard) {
			g.BankerHand.Cards = append(g.BankerHand.Cards, g.Deck[0])
			g.Deck = g.Deck[1:]
			g.BankerScore = calculateScore(g.BankerHand.Cards)
		}
	} else if bankerScore <= 5 {
		// 閒家不補牌，莊家補牌
		g.BankerHand.Cards = append(g.BankerHand.Cards, g.Deck[0])
		g.Deck = g.Deck[1:]
		g.BankerScore = calculateScore(g.BankerHand.Cards)
	}
}

func (g *Game) DetermineWinner() {
	// 檢查幸運6
	if g.BankerScore == 6 {
		g.IsLuckySix = true
		if len(g.BankerHand.Cards) == 2 {
			g.LuckySixType = "2cards"
		} else {
			g.LuckySixType = "3cards"
		}
	}

	// 決定贏家
	if g.PlayerScore == g.BankerScore {
		g.Winner = "Tie"
	} else if g.PlayerScore > g.BankerScore {
		g.Winner = "Player"
	} else {
		g.Winner = "Banker"
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
