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
	winner string, isLuckySix bool) error {

	query := `
		INSERT INTO game_records (
			game_id, player_initial_cards, banker_initial_cards,
			player_third_card, banker_third_card,
			player_final_score, banker_final_score,
			winner, is_lucky_six
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := DB.Exec(query,
		gameID, playerInitialCards, bankerInitialCards,
		playerThirdCard, bankerThirdCard,
		playerFinalScore, bankerFinalScore,
		winner, isLuckySix,
	)

	if err != nil {
		return fmt.Errorf("error saving game record: %v", err)
	}

	return nil
}

// SaveCardDetail saves individual card details to database
func SaveCardDetail(gameID string, position string, suit, value int) error {
	query := `
		INSERT INTO card_details (
			game_id, card_position, suit, value
		) VALUES (?, ?, ?, ?)
	`

	_, err := DB.Exec(query, gameID, position, suit, value)
	if err != nil {
		return fmt.Errorf("error saving card detail: %v", err)
	}

	return nil
}
