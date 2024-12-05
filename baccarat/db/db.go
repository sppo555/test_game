package db

import (
	"baccarat/config"
	"baccarat/pkg/logger"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName,
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
	playerThirdCard, bankerThirdCard sql.NullString,
	playerFinalScore, bankerFinalScore int,
	winner string, isLuckySix bool,
	luckySixType sql.NullString,
	payouts map[string]float64) error {

	return Transaction(func(tx *sql.Tx) error {
		query := `
			INSERT INTO game_records (
				game_id, player_initial_cards, banker_initial_cards,
				player_third_card, banker_third_card,
				player_final_score, banker_final_score,
				winner, is_lucky_six, lucky_six_type,
				player_payout, banker_payout, tie_payout, lucky_six_payout
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		var playerPayout, bankerPayout, tiePayout, luckySixPayout sql.NullFloat64

		if p, ok := payouts["player"]; ok {
			playerPayout = sql.NullFloat64{Float64: p, Valid: true}
		}
		if p, ok := payouts["banker"]; ok {
			bankerPayout = sql.NullFloat64{Float64: p, Valid: true}
		}
		if p, ok := payouts["tie"]; ok {
			tiePayout = sql.NullFloat64{Float64: p, Valid: true}
		}
		if p, ok := payouts["luckySix"]; ok {
			luckySixPayout = sql.NullFloat64{Float64: p, Valid: true}
		}

		_, err := tx.Exec(query,
			gameID, playerInitialCards, bankerInitialCards,
			playerThirdCard, bankerThirdCard,
			playerFinalScore, bankerFinalScore,
			winner, isLuckySix, luckySixType,
			playerPayout, bankerPayout, tiePayout, luckySixPayout,
		)

		return err
	})
}

// SaveCardDetail saves individual card details to database
func SaveCardDetail(gameID string, position string, suit, value int) error {
	return Transaction(func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			INSERT INTO card_details (game_id, position, suit, value)
			VALUES (?, ?, ?, ?)
		`, gameID, position, suit, value)
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
