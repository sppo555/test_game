package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	return nil
}

// SaveGameRecord saves the game record to database
func SaveGameRecord(gameID string, playerInitialCards, bankerInitialCards string,
	playerThirdCard, bankerThirdCard sql.NullString,
	playerFinalScore, bankerFinalScore int,
	winner string, isLuckySix bool,
	luckySixType sql.NullString,
	payouts map[string]float64) error {

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

	if p, ok := payouts["Player"]; ok {
		playerPayout = sql.NullFloat64{Float64: p, Valid: true}
	}
	if p, ok := payouts["Banker"]; ok {
		bankerPayout = sql.NullFloat64{Float64: p, Valid: true}
	}
	if p, ok := payouts["Tie"]; ok {
		tiePayout = sql.NullFloat64{Float64: p, Valid: true}
	}
	if p, ok := payouts["LuckySix"]; ok {
		luckySixPayout = sql.NullFloat64{Float64: p, Valid: true}
	}

	_, err := DB.Exec(query,
		gameID, playerInitialCards, bankerInitialCards,
		playerThirdCard, bankerThirdCard,
		playerFinalScore, bankerFinalScore,
		winner, isLuckySix, luckySixType,
		playerPayout, bankerPayout, tiePayout, luckySixPayout,
	)

	if err != nil {
		return fmt.Errorf("error saving game record: %v", err)
	}

	return nil
}

// SaveCardDetail saves individual card details to database
func SaveCardDetail(gameID string, position string, suit, value int) error {
	stmt, err := DB.Prepare(`
		INSERT INTO card_details (game_id, position, suit, value)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(gameID, position, suit, value)
	if err != nil {
		return fmt.Errorf("error executing statement: %v", err)
	}

	return nil
}
