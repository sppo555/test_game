// db.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	
	_ "github.com/go-sql-driver/mysql"
	"github.com/letron/verify/internal/config"
)

// DB 封裝資料庫連線
type DB struct {
	conn   *sql.DB
	config *config.Config
}

// NewDB 建立新的資料庫連線
func NewDB(cfg *config.Config) *DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	return &DB{
		conn:   db,
		config: cfg,
	}
}

// GetLatestGameIDs 從資料庫中查詢最新的 game_id
func (db *DB) GetLatestGameIDs(limit int) ([]string, error) {
	tableName := os.Getenv("DB_TABLE")
	if tableName == "" {
		tableName = "games" // 默認表名
	}
	
	query := fmt.Sprintf("SELECT game_id FROM %s ORDER BY created_at DESC LIMIT ?", tableName)
	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gameIDs []string
	for rows.Next() {
		var gameID string
		if err := rows.Scan(&gameID); err != nil {
			return nil, err
		}
		gameIDs = append(gameIDs, gameID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return gameIDs, nil
}

// Close 關閉資料庫連線
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}
