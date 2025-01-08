package game

import (
	"testing"
)

func TestShouldBankerDrawThird(t *testing.T) {
	// 建立一個 Game 物件（僅為了呼叫 shouldBankerDrawThird）
	g := &Game{}

	// 表格測試：輸入為 (莊家分數, 閒家第三張點數)，期望輸出為 shouldDraw (true/false)
	tests := []struct {
		bankerScore      int
		playerThirdValue int
		expected         bool
	}{
		// case 1: 莊家 0,1,2 => 任何閒家第三張都要補
		{0, 0, true},
		{2, 9, true},

		// case 2: 莊家 3 => 閒家第三張是 8 不補，其它都補
		{3, 8, false},
		{3, 0, true},
		{3, 7, true},

		// case 3: 莊家 4 => 閒家第三張 0,1,8,9 不補，其它 (2~7) 補
		{4, 0, false},
		{4, 1, false},
		{4, 8, false},
		{4, 9, false},
		{4, 2, true},
		{4, 7, true},

		// case 4: 莊家 5 => 閒家第三張 0,1,2,3,8,9 不補，其它 (4~7) 補
		{5, 3, false},
		{5, 8, false},
		{5, 4, true},
		{5, 7, true},

		// case 5: 莊家 6 => 閒家第三張 6,7 才補，其它不補
		{6, 6, true},
		{6, 7, true},
		{6, 5, false},
		{6, 9, false},

		// case 6: 莊家 >= 7 => 不補
		{7, 0, false},
		{8, 5, false},
		{9, 9, false},
	}

	for _, tt := range tests {
		g.BankerScore = tt.bankerScore
		got := g.shouldBankerDrawThird(tt.playerThirdValue)
		if got != tt.expected {
			t.Errorf("BankerScore=%d, PlayerThird=%d => shouldDraw=%v, but got=%v",
				tt.bankerScore, tt.playerThirdValue, tt.expected, got)
		}
	}
}

func TestNeedThirdCard(t *testing.T) {
	g := &Game{}

	tests := []struct {
		playerScore int
		bankerScore int
		expected    bool
	}{
		// 任一方 >= 8 => false (天牌不補)
		{8, 0, false},
		{0, 9, false},
		{9, 9, false},

		// 閒家 <=5 => true
		{0, 0, true},
		{5, 5, true},

		// 閒家 6,7 且莊家 6,7,8,9 => false
		{6, 6, false},
		{7, 7, false},
		{6, 8, false},
		{7, 9, false},

		// 閒家 6,7 且莊家 0-5 => true（莊家必須補牌）
		{6, 0, true},
		{7, 1, true},
		{6, 2, true},
		{7, 3, true},
		{6, 4, true},
		{7, 5, true},
	}

	for _, tt := range tests {
		g.PlayerScore = tt.playerScore
		g.BankerScore = tt.bankerScore
		got := g.NeedThirdCard()
		if got != tt.expected {
			t.Errorf("PlayerScore=%d, BankerScore=%d => NeedThird=%v, but got=%v",
				tt.playerScore, tt.bankerScore, tt.expected, got)
		}
	}
}

type MockDeck struct {
	Cards []Card
	Index int
}

func (md *MockDeck) DrawCard() Card {
	card := md.Cards[md.Index]
	md.Index++
	return card
}

func (md *MockDeck) Shuffle() {}

// GetCards 實現 Deck 接口
func (md *MockDeck) GetCards() []Card {
	return md.Cards
}

// 其餘方法如有需要可自行補充

func TestGameFlow(t *testing.T) {
	// 模擬以下發牌順序：
	// 閒家：2、3，莊家：4、0 => 整理後點數：閒家 5 點，莊家 4 點
	// 閒家第三張：4 (閒家 9 點)
	// 莊家根據閒家的第三張 4 => 因為莊家前兩張是 4 點，閒家第三張是 4，屬於 "4 => 2~7 補牌"
	// 莊家再拿一張：5 => 莊家最終 9 點
	// => 結果可能是 和局 或 其他狀況

	mockDeck := &MockDeck{
		Cards: []Card{
			// 預先寫死要抽的卡，按 Deal() 函式的抽卡順序
			// 1. Player
			{Suit: Spades, Value: 2},
			// 2. Player
			{Suit: Spades, Value: 3},
			// 3. Banker
			{Suit: Hearts, Value: 4},
			// 4. Banker
			{Suit: Clubs, Value: 10}, // 10 => 卡值=0
			// 5. Player 第三張
			{Suit: Diamonds, Value: 4},
			// 6. Banker 第三張
			{Suit: Hearts, Value: 5},
		},
	}

	g := &Game{
		Deck:    mockDeck,
		Payouts: make(map[string]float64),
	}

	g.Play()

	// 驗證結果
	// 1) 閒家前兩張 => 2+3=5
	// 2) 莊家前兩張 => 4+0=4
	// 閒家 <=5 => 補 -> 第三張=4 => 閒家總分= (5+4)=9( mod10)
	// 莊家看到閒家第三張=4 => 自身是4 => 依照 shouldBankerDrawThird => true => 拿 5 => (4+5)=9
	// 最終雙方 9 vs 9 => 和局

	if g.PlayerScore != 9 || g.BankerScore != 9 {
		t.Errorf("Expected PlayerScore=9, BankerScore=9, got %d, %d", g.PlayerScore, g.BankerScore)
	}
	if g.Winner != "Tie" {
		t.Errorf("Expected Tie, got %s", g.Winner)
	}
}

// TestNaturalHand 測試天牌情況
func TestNaturalHand(t *testing.T) {
	tests := []struct {
		name        string
		playerCards []Card
		bankerCards []Card
		expected    string // 預期贏家
	}{
		{
			name: "閒家天牌8點",
			playerCards: []Card{
				{Suit: Spades, Value: 3},
				{Suit: Hearts, Value: 5}, // 3+5=8
			},
			bankerCards: []Card{
				{Suit: Diamonds, Value: 2},
				{Suit: Clubs, Value: 3}, // 2+3=5
			},
			expected: "Player",
		},
		{
			name: "莊家天牌9點",
			playerCards: []Card{
				{Suit: Spades, Value: 4},
				{Suit: Hearts, Value: 3}, // 4+3=7
			},
			bankerCards: []Card{
				{Suit: Diamonds, Value: 4},
				{Suit: Clubs, Value: 5}, // 4+5=9
			},
			expected: "Banker",
		},
		{
			name: "雙方都天牌，莊9閒8",
			playerCards: []Card{
				{Suit: Spades, Value: 3},
				{Suit: Hearts, Value: 5}, // 3+5=8
			},
			bankerCards: []Card{
				{Suit: Diamonds, Value: 4},
				{Suit: Clubs, Value: 5}, // 4+5=9
			},
			expected: "Banker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDeck := &MockDeck{
				Cards: append(tt.playerCards, tt.bankerCards...),
			}

			g := &Game{
				Deck:    mockDeck,
				Payouts: make(map[string]float64),
			}

			g.Play()

			// 驗證沒有補牌（天牌不補）
			if len(g.PlayerHand.Cards) != 2 || len(g.BankerHand.Cards) != 2 {
				t.Errorf("Natural hand should not draw third card, got Player:%d cards, Banker:%d cards",
					len(g.PlayerHand.Cards), len(g.BankerHand.Cards))
			}

			if g.Winner != tt.expected {
				t.Errorf("Expected winner %s, got %s", tt.expected, g.Winner)
			}
		})
	}
}

// TestLuckySix 測試幸運6情況
func TestLuckySix(t *testing.T) {
	tests := []struct {
		name           string
		playerCards    []Card
		bankerCards    []Card
		playerThird    *Card     // 可選的閒家第三張
		bankerThird    *Card     // 可選的莊家第三張
		expectedType   string    // 預期幸運6類型：2cards 或 3cards
		expectedWinner string
	}{
		{
			name: "莊家兩張牌幸運6",
			playerCards: []Card{
				{Suit: Spades, Value: 2},
				{Suit: Hearts, Value: 3}, // 2+3=5
			},
			bankerCards: []Card{
				{Suit: Diamonds, Value: 1},
				{Suit: Clubs, Value: 5}, // 1+5=6
			},
			playerThird:    &Card{Suit: Diamonds, Value: 0}, // 補牌後 5+0=5
			bankerThird:    nil,  // 莊家6點不補牌
			expectedType:   "2cards",
			expectedWinner: "Banker",
		},
		{
			name: "莊家三張牌幸運6",
			playerCards: []Card{
				{Suit: Spades, Value: 2},
				{Suit: Hearts, Value: 3}, // 2+3=5
			},
			bankerCards: []Card{
				{Suit: Clubs, Value: 4},
				{Suit: Spades, Value: 7}, // 4+7=1
			},
			playerThird:    &Card{Suit: Diamonds, Value: 0}, // 補牌後 5+0=5
			bankerThird:    &Card{Suit: Hearts, Value: 5},   // 補牌後 1+5=6
			expectedType:   "3cards",
			expectedWinner: "Banker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 按照發牌順序構建牌組
			cards := append([]Card{}, tt.playerCards...)
			cards = append(cards, tt.bankerCards...)
			if tt.playerThird != nil {
				cards = append(cards, *tt.playerThird)
			}
			if tt.bankerThird != nil {
				cards = append(cards, *tt.bankerThird)
			}
			
			mockDeck := &MockDeck{
				Cards: cards,
			}

			g := &Game{
				Deck:    mockDeck,
				Payouts: make(map[string]float64),
			}

			g.Play()

			// 輸出詳細的遊戲狀態，幫助調試
			t.Logf("Game state - Player: %v (%d), Banker: %v (%d)", 
				g.PlayerHand.Cards, g.PlayerScore, 
				g.BankerHand.Cards, g.BankerScore)

			if !g.IsLuckySix {
				t.Errorf("Expected Lucky Six, but got false. Banker score: %d", g.BankerScore)
			}

			if g.LuckySixType != tt.expectedType {
				t.Errorf("Expected Lucky Six type %s, got %s", tt.expectedType, g.LuckySixType)
			}

			if g.Winner != tt.expectedWinner {
				t.Errorf("Expected winner %s, got %s. Banker score: %d, Player score: %d", 
					tt.expectedWinner, g.Winner, g.BankerScore, g.PlayerScore)
			}

			if g.BankerScore != 6 {
				t.Errorf("Lucky Six must have banker score 6, got %d", g.BankerScore)
			}
		})
	}
}

// TestPlayerThirdTo8or9 測試閒家補牌後變成8或9點的情況
func TestPlayerThirdTo8or9(t *testing.T) {
	tests := []struct {
		name           string
		playerCards    []Card
		bankerCards    []Card
		playerThird    Card
		bankerThird    Card
		bankerShouldDraw bool
		expectedWinner string
	}{
		{
			name: "閒家補牌後8點，莊家4點要補牌",
			playerCards: []Card{
				{Suit: Spades, Value: 2},
				{Suit: Hearts, Value: 3}, // 2+3=5
			},
			bankerCards: []Card{
				{Suit: Clubs, Value: 2},
				{Suit: Spades, Value: 2}, // 2+2=4
			},
			playerThird: Card{Suit: Diamonds, Value: 3}, // 補牌後 5+3=8
			bankerThird: Card{Suit: Hearts, Value: 3},   // 補牌後 4+3=7
			bankerShouldDraw: true,
			expectedWinner: "Player",
		},
		{
			name: "閒家補牌後9點，莊家3點要補牌",
			playerCards: []Card{
				{Suit: Spades, Value: 2},
				{Suit: Hearts, Value: 2}, // 2+2=4
			},
			bankerCards: []Card{
				{Suit: Clubs, Value: 2},
				{Suit: Spades, Value: 1}, // 2+1=3
			},
			playerThird: Card{Suit: Diamonds, Value: 5}, // 補牌後 4+5=9
			bankerThird: Card{Suit: Hearts, Value: 2},   // 補牌後 3+2=5
			bankerShouldDraw: true,
			expectedWinner: "Player",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := append([]Card{}, tt.playerCards...)
			cards = append(cards, tt.bankerCards...)
			cards = append(cards, tt.playerThird)
			if tt.bankerShouldDraw {
				cards = append(cards, tt.bankerThird)
			}
			
			mockDeck := &MockDeck{
				Cards: cards,
			}

			g := &Game{
				Deck:    mockDeck,
				Payouts: make(map[string]float64),
			}

			g.Play()

			// 驗證閒家有補第三張牌
			if len(g.PlayerHand.Cards) != 3 {
				t.Error("Player should draw third card")
			}

			// 驗證莊家補牌情況
			expectedBankerCards := 2
			if tt.bankerShouldDraw {
				expectedBankerCards = 3
			}
			if len(g.BankerHand.Cards) != expectedBankerCards {
				t.Errorf("Expected banker to have %d cards, got %d. Banker score: %d, Player score: %d", 
					expectedBankerCards, len(g.BankerHand.Cards), g.BankerScore, g.PlayerScore)
			}

			if g.Winner != tt.expectedWinner {
				t.Errorf("Expected winner %s, got %s. Banker score: %d, Player score: %d", 
					tt.expectedWinner, g.Winner, g.BankerScore, g.PlayerScore)
			}
		})
	}
}

// TestBankerMustDrawButNotDraw 測試莊家該補牌卻沒有補牌的情況
func TestBankerMustDrawButNotDraw(t *testing.T) {
	tests := []struct {
		name           string
		playerCards    []Card
		bankerCards    []Card
		playerThird    *Card    // 閒家第三張牌，nil 表示不補牌
		bankerMustDraw bool     // 莊家是否必須補牌
		bankerThird    Card     // 莊家的第三張牌
	}{
		{
			name: "閒家6點不補牌，莊家2點必須補牌",
			playerCards: []Card{
				{Suit: Hearts, Value: 3},    // ♥3
				{Suit: Clubs, Value: 3},     // ♣3
			},
			bankerCards: []Card{
				{Suit: Hearts, Value: 1},    // ♥A
				{Suit: Spades, Value: 1},    // ♠A
			},
			playerThird: nil,  // 閒家6點，不補牌
			bankerMustDraw: true,  // 莊家2點，必須補牌
			bankerThird: Card{Suit: Diamonds, Value: 3},  // 莊家補3點
		},
		{
			name: "閒家補4點，莊家4點看到4點必須補牌",
			playerCards: []Card{
				{Suit: Hearts, Value: 2},    // ♥2
				{Suit: Clubs, Value: 2},     // ♣2
			},
			bankerCards: []Card{
				{Suit: Hearts, Value: 2},    // ♥2
				{Suit: Spades, Value: 2},    // ♠2
			},
			playerThird: &Card{Suit: Diamonds, Value: 4},  // 閒家補4點
			bankerMustDraw: true,  // 莊家4點看到4點必須補牌
			bankerThird: Card{Suit: Hearts, Value: 5},  // 莊家補5點
		},
		{
			name: "閒家補6點，莊家6點看到6點必須補牌",
			playerCards: []Card{
				{Suit: Hearts, Value: 2},    // ♥2
				{Suit: Clubs, Value: 3},     // ♣3
			},
			bankerCards: []Card{
				{Suit: Hearts, Value: 3},    // ♥3
				{Suit: Spades, Value: 3},    // ♠3
			},
			playerThird: &Card{Suit: Diamonds, Value: 6},  // 閒家補6點
			bankerMustDraw: true,  // 莊家6點看到6點必須補牌
			bankerThird: Card{Suit: Clubs, Value: 4},  // 莊家補4點
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 準備牌組
			cards := append([]Card{}, tt.playerCards...)
			cards = append(cards, tt.bankerCards...)
			if tt.playerThird != nil {
				cards = append(cards, *tt.playerThird)
			}
			if tt.bankerMustDraw {
				cards = append(cards, tt.bankerThird)
			}
			
			mockDeck := &MockDeck{
				Cards: cards,
			}

			g := &Game{
				Deck:    mockDeck,
				Payouts: make(map[string]float64),
			}

			g.Play()

			// 檢查莊家是否有補牌
			if tt.bankerMustDraw && len(g.BankerHand.Cards) != 3 {
				t.Errorf("%s: Banker must draw third card but didn't. Banker cards: %v, score: %d", 
					tt.name, g.BankerHand.Cards, g.BankerScore)
			}

			// 輸出詳細的遊戲狀態，幫助調試
			t.Logf("Game state - Player: %v (%d), Banker: %v (%d)", 
				g.PlayerHand.Cards, g.PlayerScore, 
				g.BankerHand.Cards, g.BankerScore)
		})
	}
}

// TestBankerMustDrawWhenScoreIs0to2 測試莊家點數為0-2時必須補牌
func TestBankerMustDrawWhenScoreIs0to2(t *testing.T) {
	tests := []struct {
		name           string
		playerCards    []Card
		bankerCards   []Card
		bankerThirdCard Card  // 新增：莊家第三張牌
		expectedDraw   bool
		expectedScore int
	}{
		{
			name: "莊家點數為0時必須補牌",
			playerCards: []Card{
				{Value: 4, Suit: 1}, // D4
				{Value: 3, Suit: 2}, // C3
			}, // 閒家7點，不補牌
			bankerCards: []Card{
				{Value: 3, Suit: 0}, // S3
				{Value: 7, Suit: 1}, // D7
			}, // 莊家0點，必須補牌
			bankerThirdCard: Card{Value: 5, Suit: 3}, // C5
			expectedDraw: true,
			expectedScore: 0,
		},
		{
			name: "莊家點數為2時必須補牌",
			playerCards: []Card{
				{Value: 9, Suit: 2}, // H9
				{Value: 8, Suit: 0}, // S8
			}, // 閒家7點，不補牌
			bankerCards: []Card{
				{Value: 13, Suit: 2}, // HK
				{Value: 2, Suit: 1},  // D2
			}, // 莊家2點，必須補牌
			bankerThirdCard: Card{Value: 4, Suit: 3}, // C4
			expectedDraw: true,
			expectedScore: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 創建一個新遊戲
			g := NewGame()

			// 使用測試牌替換真實的牌組
			allCards := append(tt.playerCards, tt.bankerCards...)
			allCards = append(allCards, tt.bankerThirdCard) // 添加莊家第三張牌
			mockDeck := &MockDeck{
				Cards: allCards,
			}
			g.Deck = mockDeck

			// 發牌
			g.Deal()
			g.calculateScores()

			// 驗證莊家初始點數
			if g.BankerScore != tt.expectedScore {
				t.Errorf("莊家初始點數錯誤，期望 %d，得到 %d", tt.expectedScore, g.BankerScore)
			}

			// 驗證閒家不是天牌
			if g.PlayerScore >= 8 {
				t.Errorf("閒家不應該是天牌，但得到點數 %d", g.PlayerScore)
			}

			// 補牌
			g.DealThirdCard()

			// 驗證莊家是否有補牌
			if tt.expectedDraw && len(g.BankerHand.Cards) != 3 {
				t.Errorf("莊家點數為 %d 時應該補牌，但沒有補牌", tt.expectedScore)
			}

			// 輸出遊戲狀態以便調試
			t.Logf("Game state - Player: %v (%d), Banker: %v (%d)",
				g.PlayerHand.Cards, g.PlayerScore,
				g.BankerHand.Cards, g.BankerScore)
		})
	}
}

// TestBankerMustDrawWhenPlayerStands 測試當閒家不補牌時（6或7點），莊家點數0-5必須補牌的情況
func TestBankerMustDrawWhenPlayerStands(t *testing.T) {
	tests := []struct {
		name           string
		playerCards    []Card
		bankerCards   []Card
		bankerThirdCard Card
		expectedDraw   bool
		expectedScore int
	}{
		{
			name: "閒家6點不補牌，莊家1點必須補牌",
			playerCards: []Card{
				{Value: 13, Suit: 1}, // DK
				{Value: 6, Suit: 0},  // S6
			}, // 閒家6點，不補牌
			bankerCards: []Card{
				{Value: 7, Suit: 0}, // S7
				{Value: 4, Suit: 0}, // S4
			}, // 莊家1點，必須補牌
			bankerThirdCard: Card{Value: 5, Suit: 3}, // C5
			expectedDraw: true,
			expectedScore: 1,
		},
		{
			name: "閒家7點不補牌，莊家5點必須補牌",
			playerCards: []Card{
				{Value: 3, Suit: 1}, // D3
				{Value: 4, Suit: 0}, // S4
			}, // 閒家7點，不補牌
			bankerCards: []Card{
				{Value: 2, Suit: 0}, // S2
				{Value: 3, Suit: 0}, // S3
			}, // 莊家5點，必須補牌
			bankerThirdCard: Card{Value: 6, Suit: 3}, // C6
			expectedDraw: true,
			expectedScore: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 創建一個新遊戲
			g := NewGame()

			// 使用測試牌替換真實的牌組
			allCards := append(tt.playerCards, tt.bankerCards...)
			allCards = append(allCards, tt.bankerThirdCard)
			mockDeck := &MockDeck{
				Cards: allCards,
			}
			g.Deck = mockDeck

			// 發牌
			g.Deal()
			g.calculateScores()

			// 驗證莊家初始點數
			if g.BankerScore != tt.expectedScore {
				t.Errorf("莊家初始點數錯誤，期望 %d，得到 %d", tt.expectedScore, g.BankerScore)
			}

			// 驗證閒家不是天牌且不需要補牌（6或7點）
			if g.PlayerScore < 6 || g.PlayerScore > 7 {
				t.Errorf("閒家點數應該是6或7點，但得到點數 %d", g.PlayerScore)
			}

			// 補牌
			if g.NeedThirdCard() {
				g.DealThirdCard()
			}

			// 驗證莊家是否有補牌
			if tt.expectedDraw && len(g.BankerHand.Cards) != 3 {
				t.Errorf("莊家點數為 %d 時應該補牌，但沒有補牌", tt.expectedScore)
			}

			// 輸出遊戲狀態以便調試
			t.Logf("Game state - Player: %v (%d), Banker: %v (%d)",
				g.PlayerHand.Cards, g.PlayerScore,
				g.BankerHand.Cards, g.BankerScore)
		})
	}
}
