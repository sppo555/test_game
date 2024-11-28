package main

import (
	"baccarat/db"
	"baccarat/game"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var (
	sessions     = make(map[string]int) // session token to user ID mapping
	sessionsLock sync.Mutex
	jwtKey       = []byte("your_secret_key")
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
	payouts["PAYOUT_PLAYER"], _ = strconv.ParseFloat(os.Getenv("PLAYER_PAYOUT"), 64)
	payouts["PAYOUT_BANKER"], _ = strconv.ParseFloat(os.Getenv("BANKER_PAYOUT"), 64)
	payouts["PAYOUT_TIE"], _ = strconv.ParseFloat(os.Getenv("TIE_PAYOUT"), 64)

	// 幸運6賠率
	payouts["PAYOUT_LUCKY6_2CARDS"], _ = strconv.ParseFloat(os.Getenv("LUCKY6_2CARDS_PAYOUT"), 64)
	payouts["PAYOUT_LUCKY6_3CARDS"], _ = strconv.ParseFloat(os.Getenv("LUCKY6_3CARDS_PAYOUT"), 64)
	payouts["PAYOUT_BANKER_LUCKY6_2CARDS"], _ = strconv.ParseFloat(os.Getenv("BANKER_LUCKY6_2CARDS_PAYOUT"), 64)
	payouts["PAYOUT_BANKER_LUCKY6_3CARDS"], _ = strconv.ParseFloat(os.Getenv("BANKER_LUCKY6_3CARDS_PAYOUT"), 64)

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

	// 加載環境變量
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	payouts := loadPayouts()
	game := game.NewGame()
	game.Payouts = payouts
	game.Deal()
	game.DealThirdCard() // 確保補牌邏輯被調用
	game.DetermineWinner()
	game.CalculatePayouts()

	if showLog {
		// 初始發牌
		fmt.Println("=== 初始發牌 ===")
		fmt.Printf("閒家牌: %s, 初始點數: %d\n", printHand(game.PlayerHand.Cards), game.PlayerScore)
		fmt.Printf("莊家牌: %s, 初始點數: %d\n", printHand(game.BankerHand.Cards), game.BankerScore)

		// 補牌階段
		fmt.Println("\n=== 補牌階段 ===")
		fmt.Printf("閒家牌: %s, 最終點數: %d\n", printHand(game.PlayerHand.Cards), game.PlayerScore)
		fmt.Printf("莊家牌: %s, 最終點數: %d\n", printHand(game.BankerHand.Cards), game.BankerScore)
		fmt.Printf("贏家: %s\n", getWinnerString(game.Winner))

		// 顯示賠率信息
		if game.IsLuckySix {
			fmt.Printf("\n🎉 恭喜！獲得%s幸運6！\n",
				map[string]string{"2cards": "兩張牌", "3cards": "三張牌"}[game.LuckySixType])
		}

		fmt.Println("\n=== 賠率信息 ===")
		if game.Winner == "Banker" && game.IsLuckySix {
			if game.LuckySixType == "2cards" {
				fmt.Printf("幸運6賠率: %.2f:1\n", game.Payouts["PAYOUT_LUCKY6_2CARDS"])
				fmt.Printf("莊家賠率: %.2f:1\n", game.Payouts["PAYOUT_BANKER_LUCKY6_2CARDS"])
			} else {
				fmt.Printf("幸運6賠率: %.2f:1\n", game.Payouts["PAYOUT_LUCKY6_3CARDS"])
				fmt.Printf("莊家賠率: %.2f:1\n", game.Payouts["PAYOUT_BANKER_LUCKY6_3CARDS"])
			}
		} else {
			fmt.Printf("和局賠率: %.2f:1\n", game.Payouts["PAYOUT_TIE"])
			fmt.Printf("閒家賠率: %.2f:1\n", game.Payouts["PAYOUT_PLAYER"])
			fmt.Printf("莊家賠率: %.2f:1\n", game.Payouts["PAYOUT_BANKER"])
		}
	}

	// 初始化數據庫連接
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 保存遊戲記錄到資料庫
	if err := saveGame(game, gameID); err != nil {
		log.Printf("Error saving game record: %v\n", err)
	} else if showLog {
		fmt.Printf("\n遊戲記錄已保存，遊戲ID: %s\n", gameID)
	}
}

func generateJWT(userID int) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func authenticateJWT(r *http.Request) (int, bool) {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	} else {
		return 0, false
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return 0, false
	}
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return 0, false
		}
		return userID, true
	}
	return 0, false
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Account created successfully")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var userID int
	var passwordHash string
	err := db.DB.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", username).Scan(&userID, &passwordHash)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(userID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Login successful. Token: %s", token)
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := authenticateJWT(r)
	if !authenticated {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var balance float64
	err := db.DB.QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&balance)
	if err != nil {
		http.Error(w, "Error retrieving balance", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Current balance: %.2f", balance)
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := authenticateJWT(r)
	if !authenticated {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.DB.Query("SELECT amount, created_at FROM transactions WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Error retrieving transactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var amount float64
		var createdAt string
		if err := rows.Scan(&amount, &createdAt); err != nil {
			http.Error(w, "Error scanning transaction", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Amount: %.2f, Date: %s\n", amount, createdAt)
	}
}

func betsHandler(w http.ResponseWriter, r *http.Request) {
	userID, authenticated := authenticateJWT(r)
	if !authenticated {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.DB.Query("SELECT game_id, bet_amount, bet_type, created_at FROM bets WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Error retrieving bets", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var gameID, betType, createdAt string
		var betAmount float64
		if err := rows.Scan(&gameID, &betAmount, &betType, &createdAt); err != nil {
			http.Error(w, "Error scanning bet", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Game ID: %s, Bet Amount: %.2f, Bet Type: %s, Date: %s\n", gameID, betAmount, betType, createdAt)
	}
}

func depositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, authenticated := authenticateJWT(r)
	if !authenticated {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	amountStr := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid deposit amount", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, userID)
	if err != nil {
		http.Error(w, "Error processing deposit", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("INSERT INTO transactions (user_id, amount) VALUES (?, ?)", userID, amount)
	if err != nil {
		http.Error(w, "Error recording transaction", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Deposit successful: %.2f", amount)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// JWT does not need to be deleted from the server-side, as it is stateless
	fmt.Fprintln(w, "Logout successful")
}

func main() {
	// 載入環境變數
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 初始化數據庫連接
	if err := db.InitDB(); err != nil {
		log.Fatal("Error initializing database:", err)
	}

	// 設置路由
	http.HandleFunc("/api/register", registerHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/user/balance", balanceHandler)
	http.HandleFunc("/api/user/transactions", transactionsHandler)
	http.HandleFunc("/api/user/bets", betsHandler)
	http.HandleFunc("/api/user/deposit", depositHandler)
	http.HandleFunc("/api/logout", logoutHandler)

	// 啟動服務器
	log.Fatal(http.ListenAndServe(":8080", nil))
}
