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
	APIURL     string
	Token      string
	GameIDLimit int
	// 單次驗證配置
	SingleGameMode bool
	SingleGameID   string
	// SQL驗證配置
	SQLVerifyMode bool
	// 賠率配置
	PlayerPayout            float64
	BankerPayout           float64
	TiePayout              float64
	Lucky6_2CardsPayout    float64
	Lucky6_3CardsPayout    float64
	BankerLucky6_2CardsPayout float64
	BankerLucky6_3CardsPayout float64
	BatchSize               int // 添加 BatchSize 字段
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
		log.Fatal("Error: Could not find .env file in any of the search paths")
	}

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatal("Invalid DB_PORT")
	}

	gameIDLimit, err := strconv.Atoi(os.Getenv("GAME_ID_LIMIT"))
	if err != nil {
		log.Fatal("Invalid GAME_ID_LIMIT")
	}

	singleGameMode, err := strconv.ParseBool(os.Getenv("SINGLE_GAME_MODE"))
	if err != nil {
		log.Fatal("Invalid SINGLE_GAME_MODE")
	}

	playerPayout, err := strconv.ParseFloat(os.Getenv("PLAYER_PAYOUT"), 64)
	if err != nil {
		log.Fatal("Invalid PLAYER_PAYOUT")
	}

	bankerPayout, err := strconv.ParseFloat(os.Getenv("BANKER_PAYOUT"), 64)
	if err != nil {
		log.Fatal("Invalid BANKER_PAYOUT")
	}

	tiePayout, err := strconv.ParseFloat(os.Getenv("TIE_PAYOUT"), 64)
	if err != nil {
		log.Fatal("Invalid TIE_PAYOUT")
	}

	lucky6_2CardsPayout, err := strconv.ParseFloat(os.Getenv("LUCKY6_2CARDS_PAYOUT"), 64)
	if err != nil {
		log.Fatal("Invalid LUCKY6_2CARDS_PAYOUT")
	}

	lucky6_3CardsPayout, err := strconv.ParseFloat(os.Getenv("LUCKY6_3CARDS_PAYOUT"), 64)
	if err != nil {
		log.Fatal("Invalid LUCKY6_3CARDS_PAYOUT")
	}

	bankerLucky6_2CardsPayout, err := strconv.ParseFloat(os.Getenv("BANKER_LUCKY6_2CARDS_PAYOUT"), 64)
	if err != nil {
		log.Fatal("Invalid BANKER_LUCKY6_2CARDS_PAYOUT")
	}

	bankerLucky6_3CardsPayout, err := strconv.ParseFloat(os.Getenv("BANKER_LUCKY6_3CARDS_PAYOUT"), 64)
	if err != nil {
		log.Fatal("Invalid BANKER_LUCKY6_3CARDS_PAYOUT")
	}

	sqlVerifyMode, err := strconv.ParseBool(os.Getenv("SQL_VERIFY_MODE"))
	if err != nil {
		log.Printf("Warning: Invalid SQL_VERIFY_MODE, defaulting to false")
		sqlVerifyMode = false
	}

	batchSize, err := strconv.Atoi(os.Getenv("BATCH_SIZE")) // 添加 BatchSize 的讀取
	if err != nil {
		log.Fatal("Invalid BATCH_SIZE")
	}

	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     dbPort,
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		TimeZone:   os.Getenv("TIME_ZONE"),
		APIURL:     os.Getenv("API_URL"),
		Token:      os.Getenv("API_TOKEN"),
		GameIDLimit: gameIDLimit,
		SingleGameMode: singleGameMode,
		SingleGameID:   os.Getenv("SINGLE_GAME_ID"),
		SQLVerifyMode: sqlVerifyMode,
		PlayerPayout:            playerPayout,
		BankerPayout:           bankerPayout,
		TiePayout:              tiePayout,
		Lucky6_2CardsPayout:    lucky6_2CardsPayout,
		Lucky6_3CardsPayout:    lucky6_3CardsPayout,
		BankerLucky6_2CardsPayout: bankerLucky6_2CardsPayout,
		BankerLucky6_3CardsPayout: bankerLucky6_3CardsPayout,
		BatchSize: batchSize, // 添加 BatchSize 到返回的配置中
	}
}
