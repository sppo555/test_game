package main

import (
	"baccarat/db"
	"baccarat/game"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func printCard(card game.Card) string {
	suits := []string{"黑桃", "紅心", "方塊", "梅花"}
	values := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	return suits[card.Suit] + values[card.Value-1]
}

func printHand(cards []game.Card) string {
	result := ""
	for i, card := range cards {
		if i > 0 {
			result += ", "
		}
		result += printCard(card)
	}
	return result
}

func getWinnerString(winner string) string {
	switch winner {
	case "Player":
		return "閒家"
	case "Banker":
		return "莊家"
	case "Tie":
		return "和局"
	default:
		return "未知"
	}
}

func loadPayouts() map[string]float64 {
	payouts := make(map[string]float64)

	// 基本賠率
	payouts["PAYOUT_PLAYER"], _ = strconv.ParseFloat(os.Getenv("PAYOUT_PLAYER"), 64)
	payouts["PAYOUT_BANKER"], _ = strconv.ParseFloat(os.Getenv("PAYOUT_BANKER"), 64)
	payouts["PAYOUT_TIE"], _ = strconv.ParseFloat(os.Getenv("PAYOUT_TIE"), 64)
	
	// 幸運6賠率
	payouts["PAYOUT_LUCKY6_2CARDS"], _ = strconv.ParseFloat(os.Getenv("PAYOUT_LUCKY6_2CARDS"), 64)
	payouts["PAYOUT_LUCKY6_3CARDS"], _ = strconv.ParseFloat(os.Getenv("PAYOUT_LUCKY6_3CARDS"), 64)
	payouts["PAYOUT_BANKER_LUCKY6_2CARDS"], _ = strconv.ParseFloat(os.Getenv("PAYOUT_BANKER_LUCKY6_2CARDS"), 64)
	payouts["PAYOUT_BANKER_LUCKY6_3CARDS"], _ = strconv.ParseFloat(os.Getenv("PAYOUT_BANKER_LUCKY6_3CARDS"), 64)

	return payouts
}

func saveGame(g *game.Game, gameID string) error {
	// 準備初始牌的字符串
	playerInitialCards := printHand(g.PlayerHand.Cards[:2])
	bankerInitialCards := printHand(g.BankerHand.Cards[:2])

	// 準備補牌的字符串（如果有的話）
	var playerThirdCard, bankerThirdCard sql.NullString
	if len(g.PlayerHand.Cards) > 2 {
		playerThirdCard = sql.NullString{
			String: printCard(g.PlayerHand.Cards[2]),
			Valid:  true,
		}
	}
	if len(g.BankerHand.Cards) > 2 {
		bankerThirdCard = sql.NullString{
			String: printCard(g.BankerHand.Cards[2]),
			Valid:  true,
		}
	}

	// 準備幸運6類型
	var luckySixType sql.NullString
	if g.IsLuckySix {
		luckySixType = sql.NullString{
			String: g.LuckySixType,
			Valid:  true,
		}
	}

	// 保存遊戲記錄
	err := db.SaveGameRecord(
		gameID,
		playerInitialCards,
		bankerInitialCards,
		playerThirdCard,
		bankerThirdCard,
		g.PlayerScore,
		g.BankerScore,
		g.Winner,
		g.IsLuckySix,
		luckySixType,
		g.Payouts,
	)
	if err != nil {
		return err
	}

	// 保存每張牌的詳細信息
	positions := []string{"PlayerInitial1", "PlayerInitial2", "BankerInitial1", "BankerInitial2"}
	cards := append(g.PlayerHand.Cards[:2], g.BankerHand.Cards[:2]...)

	for i, card := range cards {
		err = db.SaveCardDetail(gameID, positions[i], int(card.Suit), card.Value)
		if err != nil {
			return err
		}
	}

	// 保存補牌信息（如果有的話）
	if len(g.PlayerHand.Cards) > 2 {
		err = db.SaveCardDetail(gameID, "PlayerThird", int(g.PlayerHand.Cards[2].Suit), g.PlayerHand.Cards[2].Value)
		if err != nil {
			return err
		}
	}
	if len(g.BankerHand.Cards) > 2 {
		err = db.SaveCardDetail(gameID, "BankerThird", int(g.BankerHand.Cards[2].Suit), g.BankerHand.Cards[2].Value)
		if err != nil {
			return err
		}
	}

	return nil
}

func playOneGame(showLog bool) {
	// 生成遊戲ID
	gameID := uuid.New().String()

	// 創建新遊戲並設置賠率
	g := game.NewGame()
	payouts := loadPayouts()
	for k, v := range payouts {
		g.Payouts[k] = v
	}

	// 初始發牌
	g.Deal()
	if showLog {
		fmt.Println("=== 初始發牌 ===")
		fmt.Printf("閒家牌: %s, 初始點數: %d\n", printHand(g.PlayerHand.Cards), g.PlayerScore)
		fmt.Printf("莊家牌: %s, 初始點數: %d\n", printHand(g.BankerHand.Cards), g.BankerScore)
	}

	// 補牌階段
	if g.NeedThirdCard() {
		if showLog {
			fmt.Println("\n=== 補牌階段 ===")
		}
		initialPlayerCards := len(g.PlayerHand.Cards)
		initialBankerCards := len(g.BankerHand.Cards)

		g.DealThirdCard()

		if showLog {
			// 檢查閒家是否補牌
			if len(g.PlayerHand.Cards) > initialPlayerCards {
				fmt.Printf("閒家補牌: %s\n", printCard(g.PlayerHand.Cards[len(g.PlayerHand.Cards)-1]))
				fmt.Printf("閒家最終點數: %d\n", g.PlayerScore)
			} else {
				fmt.Println("閒家不需要補牌")
			}

			// 檢查莊家是否補牌
			if len(g.BankerHand.Cards) > initialBankerCards {
				fmt.Printf("莊家補牌: %s\n", printCard(g.BankerHand.Cards[len(g.BankerHand.Cards)-1]))
				fmt.Printf("莊家最終點數: %d\n", g.BankerScore)
			} else {
				fmt.Println("莊家不需要補牌")
			}
		}
	} else if showLog {
		fmt.Println("\n莊閒皆不需要補牌")
	}

	// 判定勝負
	g.DetermineWinner()

	if showLog {
		// 最終結果
		fmt.Println("\n=== 最終結果 ===")
		fmt.Printf("閒家最終牌: %s, 最終點數: %d\n", printHand(g.PlayerHand.Cards), g.PlayerScore)
		fmt.Printf("莊家最終牌: %s, 最終點數: %d\n", printHand(g.BankerHand.Cards), g.BankerScore)
		fmt.Printf("贏家: %s\n", getWinnerString(g.Winner))

		// 顯示賠率信息
		if g.IsLuckySix {
			fmt.Printf("\n🎉 恭喜！獲得%s幸運6！\n", 
				map[string]string{"2cards": "兩張牌", "3cards": "三張牌"}[g.LuckySixType])
		}

		fmt.Println("\n=== 賠率信息 ===")
		if g.Winner == "Banker" && g.IsLuckySix {
			if g.LuckySixType == "2cards" {
				fmt.Printf("幸運6賠率: %.2f:1\n", g.Payouts["PAYOUT_LUCKY6_2CARDS"])
				fmt.Printf("莊家賠率: %.2f:1\n", g.Payouts["PAYOUT_BANKER_LUCKY6_2CARDS"])
			} else {
				fmt.Printf("幸運6賠率: %.2f:1\n", g.Payouts["PAYOUT_LUCKY6_3CARDS"])
				fmt.Printf("莊家賠率: %.2f:1\n", g.Payouts["PAYOUT_BANKER_LUCKY6_3CARDS"])
			}
		} else {
			fmt.Printf("和局賠率: %.2f:1\n", g.Payouts["PAYOUT_TIE"])
			fmt.Printf("閒家賠率: %.2f:1\n", g.Payouts["PAYOUT_PLAYER"])
			fmt.Printf("莊家賠率: %.2f:1\n", g.Payouts["PAYOUT_BANKER"])
		}
	}

	// 保存遊戲記錄到資料庫
	if err := saveGame(g, gameID); err != nil {
		log.Printf("Error saving game record: %v\n", err)
	} else if showLog {
		fmt.Printf("\n遊戲記錄已保存，遊戲ID: %s\n", gameID)
	}
}

func main() {
	// 載入環境變數
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 初始化資料庫連接
	if err := db.InitDB(); err != nil {
		log.Fatal("Error initializing database:", err)
	}

	// 加載賠率設置
	payouts := loadPayouts()

	// 檢查是否設定了運行次數
	if runTimes := os.Getenv("RUN_TIMES"); runTimes != "" {
		times, err := strconv.Atoi(runTimes)
		if err != nil {
			log.Fatal("Invalid RUN_TIMES value:", err)
		}

		// 統計變數
		bankerWins := 0
		playerWins := 0
		ties := 0
		lucky6Count := 0

		fmt.Printf("執行 %d 次遊戲中...\n", times)
		for i := 1; i <= times; i++ {
			g := game.NewGame()
			g.Payouts = payouts // 設置賠率
			g.Play() // 直接使用 Play 方法來運行一局完整遊戲

			// 統計結果
			switch g.Winner {
			case "Banker":
				bankerWins++
				if g.IsLuckySix {
					lucky6Count++
				}
			case "Player":
				playerWins++
			case "Tie":
				ties++
			}

			// 保存遊戲記錄
			gameID := uuid.New().String()
			if err := saveGame(g, gameID); err != nil {
				log.Printf("Error saving game record: %v\n", err)
			}
		}

		// 輸出統計結果
		fmt.Printf("\n=== 遊戲統計 ===\n")
		fmt.Printf("總局數: %d\n", times)
		fmt.Printf("莊家贏: %d (%.2f%%)\n", bankerWins, float64(bankerWins)/float64(times)*100)
		fmt.Printf("閒家贏: %d (%.2f%%)\n", playerWins, float64(playerWins)/float64(times)*100)
		fmt.Printf("和局: %d (%.2f%%)\n", ties, float64(ties)/float64(times)*100)
		fmt.Printf("幸運6: %d (%.2f%%)\n", lucky6Count, float64(lucky6Count)/float64(times)*100)
		fmt.Printf("完成 %d 次遊戲\n", times)
	} else {
		// 互動模式保持不變
		fmt.Println("請輸入要執行的次數（直接按 Enter 執行一次）：")
		var input string
		fmt.Scanln(&input)

		times := 1
		if input != "" {
			var err error
			times, err = strconv.Atoi(input)
			if err != nil || times < 1 {
				log.Fatal("請輸入有效的次數")
			}
		}

		for i := 1; i <= times; i++ {
			fmt.Printf("\n=== 第 %d 局 ===\n", i)
			playOneGame(true) // 互動模式顯示日誌
		}
	}
}
