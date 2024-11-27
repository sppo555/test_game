package game

import (
	"math/rand"
	"time"
)

// Suit 代表撲克牌花色
type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

// Card 代表一張撲克牌
type Card struct {
	Suit  Suit
	Value int // 1-13 代表 A-K
}

// Deck 代表一副牌
type Deck struct {
	Cards []Card
}

// NewDeck 創建一副新牌
func NewDeck() *Deck {
	cards := make([]Card, 52)
	index := 0
	for suit := Spades; suit <= Clubs; suit++ {
		for value := 1; value <= 13; value++ {
			cards[index] = Card{Suit: suit, Value: value}
			index++
		}
	}
	return &Deck{Cards: cards}
}

// Shuffle 洗牌
func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// DrawCard 抽一張牌
func (d *Deck) DrawCard() Card {
	if len(d.Cards) == 0 {
		panic("No cards left in deck")
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	return card
}

// GetCardValue 獲取牌面點數（百家樂規則）
func (c Card) GetCardValue() int {
	if c.Value >= 10 {
		return 0
	}
	return c.Value
}
