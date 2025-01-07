package db

import (
	"baccarat/config"
	"baccarat/pkg/logger"
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=%s",
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName,
		url.QueryEscape(config.AppConfig.TimeZone),
	)

	logger.Debug("Attempting to connect to database with DSN:", dsn)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		logger.Error("Error opening database:", err)
		return fmt.Errorf("error opening database: %v", err)
	}

	logger.Debug("Database connection opened successfully")

	if err = DB.Ping(); err != nil {
		logger.Error("Error pinging database:", err)
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	logger.Info("Successfully connected to database")

	// 設置連接池參數
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	return nil
}

// Transaction 執行事務
func Transaction(fn func(*sql.Tx) error) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// SaveGameRecord saves the game record to database
func SaveGameRecord(gameID string, playerInitialCards, bankerInitialCards string,
	playerInitialScore, bankerInitialScore int,
	playerThirdCard, bankerThirdCard sql.NullString,
	playerFinalScore, bankerFinalScore int,
	winner string, isLuckySix bool,
	luckySixType sql.NullString,
	payouts map[string]float64) error {

	return Transaction(func(tx *sql.Tx) error {
		query := `
			INSERT INTO game_records (
				game_id, player_initial_cards, banker_initial_cards,
				player_initial_score, banker_initial_score,
				player_third_card, banker_third_card,
				player_final_score, banker_final_score,
				winner, is_lucky_six, lucky_six_type,
				player_payout, banker_payout, tie_payout, lucky_six_payout,
				total_bets, total_payouts
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		var playerPayout, bankerPayout, tiePayout, luckySixPayout sql.NullFloat64
		totalBets := 0.0
		totalPayouts := 0.0

		if p, ok := payouts["player"]; ok {
			playerPayout = sql.NullFloat64{Float64: p, Valid: true}
			totalPayouts += p
		}
		if p, ok := payouts["banker"]; ok {
			bankerPayout = sql.NullFloat64{Float64: p, Valid: true}
			totalPayouts += p
		}
		if p, ok := payouts["tie"]; ok {
			tiePayout = sql.NullFloat64{Float64: p, Valid: true}
			totalPayouts += p
		}
		if p, ok := payouts["luckySix"]; ok {
			luckySixPayout = sql.NullFloat64{Float64: p, Valid: true}
			totalPayouts += p
		}

		// 計算總投注額
		if p, ok := payouts["player_bet"]; ok {
			totalBets += p
		}
		if p, ok := payouts["banker_bet"]; ok {
			totalBets += p
		}
		if p, ok := payouts["tie_bet"]; ok {
			totalBets += p
		}
		if p, ok := payouts["luckySix_bet"]; ok {
			totalBets += p
		}

		_, err := tx.Exec(query,
			gameID, playerInitialCards, bankerInitialCards,
			playerInitialScore, bankerInitialScore,
			playerThirdCard, bankerThirdCard,
			playerFinalScore, bankerFinalScore,
			winner, isLuckySix, luckySixType,
			playerPayout, bankerPayout, tiePayout, luckySixPayout,
			totalBets, totalPayouts,
		)

		return err
	})
}

// GetUserBalance 獲取用戶餘額
func GetUserBalance(userID int) (float64, error) {
	var balance float64
	err := DB.QueryRow("SELECT balance FROM users WHERE id = ?", userID).Scan(&balance)
	return balance, err
}

// UpdateUserBalance 更新用戶餘額
func UpdateUserBalance(tx *sql.Tx, userID int, amount float64) error {
	_, err := tx.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, userID)
	return err
}

// CreateUser 創建新用戶
func CreateUser(username string, passwordHash []byte) error {
	return Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			"INSERT INTO users (username, password_hash, balance) VALUES (?, ?, 0)",
			username, passwordHash,
		)
		return err
	})
}

// GetUserByUsername 通過用戶名獲取用戶信息
func GetUserByUsername(username string) (int, string, error) {
	var id int
	var passwordHash string
	err := DB.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = ?",
		username,
	).Scan(&id, &passwordHash)
	return id, passwordHash, err
}

// SaveTransaction 保存交易記錄
func SaveTransaction(tx *sql.Tx, userID int, amount float64, transactionType string) error {
	_, err := tx.Exec(
		"INSERT INTO transactions (user_id, amount, transaction_type) VALUES (?, ?, ?)",
		userID, amount, transactionType,
	)
	return err
}

// SaveBet 保存投注記錄
func SaveBet(tx *sql.Tx, userID int, gameID string, amount float64, betType string) error {
	_, err := tx.Exec(
		"INSERT INTO bets (user_id, game_id, bet_amount, bet_type) VALUES (?, ?, ?, ?)",
		userID, gameID, amount, betType,
	)
	return err
}

// GameResult 遊戲結果完整信息
type GameResult struct {
	GameID             string         `json:"game_id"`
	Winner             string         `json:"winner"`
	PlayerInitialScore int            `json:"player_initial_score"`
	BankerInitialScore int            `json:"banker_initial_score"`
	PlayerScore        int            `json:"player_score"`
	BankerScore        int            `json:"banker_score"`
	IsLuckySix         bool           `json:"is_lucky_six"`
	LuckySixType       sql.NullString `json:"lucky_six_type"`
	PlayerCards        string         `json:"player_cards"`
	BankerCards        string         `json:"banker_cards"`
	PlayerThirdCard    sql.NullString `json:"player_third_card"`
	BankerThirdCard    sql.NullString `json:"banker_third_card"`
	PlayerPayout       sql.NullFloat64 `json:"player_payout"`
	BankerPayout       sql.NullFloat64 `json:"banker_payout"`
	TiePayout          sql.NullFloat64 `json:"tie_payout"`
	LuckySixPayout     sql.NullFloat64 `json:"lucky_six_payout"`
	Bets               []BetDetail    `json:"bets"`
	TotalBets          float64        `json:"total_bets"`
	TotalPayouts       float64        `json:"total_payouts"`
}

// BetDetail 下注詳情
type BetDetail struct {
	Username  string         `json:"username"`
	BetType   string         `json:"bet_type"`
	BetAmount float64        `json:"bet_amount"`
	Payout    sql.NullFloat64 `json:"payout"`
}

// GetGameDetails 獲取單局遊戲的詳細信息
func GetGameDetails(gameID string) (*GameResult, error) {
	// 首先查詢遊戲基本信息
	baseQuery := `
		SELECT 
			game_id,
			winner,
			player_initial_score,
			banker_initial_score,
			player_final_score,
			banker_final_score,
			is_lucky_six,
			lucky_six_type,
			player_initial_cards,
			banker_initial_cards,
			player_third_card,
			banker_third_card,
			player_payout,
			banker_payout,
			tie_payout,
			lucky_six_payout
		FROM game_records
		WHERE game_id = ?`

	var result GameResult
	err := DB.QueryRow(baseQuery, gameID).Scan(
		&result.GameID,
		&result.Winner,
		&result.PlayerInitialScore,
		&result.BankerInitialScore,
		&result.PlayerScore,
		&result.BankerScore,
		&result.IsLuckySix,
		&result.LuckySixType,
		&result.PlayerCards,
		&result.BankerCards,
		&result.PlayerThirdCard,
		&result.BankerThirdCard,
		&result.PlayerPayout,
		&result.BankerPayout,
		&result.TiePayout,
		&result.LuckySixPayout,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("遊戲不存在")
		}
		return nil, fmt.Errorf("查詢遊戲記錄失敗: %v", err)
	}

	// 查詢所有下注記錄
	betsQuery := `
		SELECT 
			u.username,
			b.bet_type,
			b.bet_amount,
			CASE b.bet_type 
				WHEN 'player' THEN gr.player_payout
				WHEN 'banker' THEN gr.banker_payout
				WHEN 'tie' THEN gr.tie_payout
				WHEN 'luckySix' THEN gr.lucky_six_payout
			END as payout
		FROM bets b
		JOIN users u ON b.user_id = u.id
		JOIN game_records gr ON b.game_id = gr.game_id
		WHERE b.game_id = ?
		ORDER BY u.username, b.bet_type`

	rows, err := DB.Query(betsQuery, gameID)
	if err != nil {
		return nil, fmt.Errorf("查詢下注記錄失敗: %v", err)
	}
	defer rows.Close()

	// 讀取所有下注記錄
	var bets []BetDetail
	for rows.Next() {
		var bet BetDetail
		err := rows.Scan(
			&bet.Username,
			&bet.BetType,
			&bet.BetAmount,
			&bet.Payout,
		)
		if err != nil {
			return nil, fmt.Errorf("讀取下注記錄失敗: %v", err)
		}
		bets = append(bets, bet)
	}

	// 計算總計
	var totalBets, totalPayouts float64
	for _, bet := range bets {
		totalBets += bet.BetAmount
		if bet.Payout.Valid {
			totalPayouts += bet.Payout.Float64
		}
	}

	result.Bets = bets
	result.TotalBets = totalBets
	result.TotalPayouts = totalPayouts

	return &result, nil
}
