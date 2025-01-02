// config.go
package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 儲存所有環境變數
type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	TimeZone   string
	Token      string
	GameIDLimit int
	PlayerPayout float64
	BankerPayout float64
	TiePayout float64
	Lucky6_2CardsPayout float64
	Lucky6_3CardsPayout float64
	BankerLucky6_2CardsPayout float64
	BankerLucky6_3CardsPayout float64
}

// LoadConfig 從 .env 讀取配置
func LoadConfig() *Config {
	// 獲取當前執行檔案的目錄
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not get executable path: %v", err)
	}

	// 嘗試多個可能的 .env 位置
	envPaths := []string{
		".env",                                    // 當前目錄
		"../.env",                                // 上一層目錄
		"../../.env",                             // 上兩層目錄
		filepath.Join(filepath.Dir(execPath), ".env"), // 執行檔案所在目錄
	}

	// 嘗試載入 .env 文件
	var loaded bool
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		log.Printf("Warning: Could not load .env file")
	}

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		dbPort = 3306 // 默認 MySQL 端口
	}

	gameIDLimit, err := strconv.Atoi(os.Getenv("GAME_ID_LIMIT"))
	if err != nil {
		gameIDLimit = 100 // 默認限制
	}

	playerPayout, _ := strconv.ParseFloat(os.Getenv("PLAYER_PAYOUT"), 64)
	bankerPayout, _ := strconv.ParseFloat(os.Getenv("BANKER_PAYOUT"), 64)
	tiePayout, _ := strconv.ParseFloat(os.Getenv("TIE_PAYOUT"), 64)
	lucky6_2CardsPayout, _ := strconv.ParseFloat(os.Getenv("LUCKY6_2CARDS_PAYOUT"), 64)
	lucky6_3CardsPayout, _ := strconv.ParseFloat(os.Getenv("LUCKY6_3CARDS_PAYOUT"), 64)
	bankerLucky6_2CardsPayout, _ := strconv.ParseFloat(os.Getenv("BANKER_LUCKY6_2CARDS_PAYOUT"), 64)
	bankerLucky6_3CardsPayout, _ := strconv.ParseFloat(os.Getenv("BANKER_LUCKY6_3CARDS_PAYOUT"), 64)

	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     dbPort,
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		TimeZone:   os.Getenv("TIME_ZONE"),
		Token:      os.Getenv("API_TOKEN"),
		GameIDLimit: gameIDLimit,
		PlayerPayout: playerPayout,
		BankerPayout: bankerPayout,
		TiePayout: tiePayout,
		Lucky6_2CardsPayout: lucky6_2CardsPayout,
		Lucky6_3CardsPayout: lucky6_3CardsPayout,
		BankerLucky6_2CardsPayout: bankerLucky6_2CardsPayout,
		BankerLucky6_3CardsPayout: bankerLucky6_3CardsPayout,
	}
}
