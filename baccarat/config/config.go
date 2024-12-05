package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// 數據庫配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBTimezone string

	// 日志配置
	LogLevel string

	// JWT配置
	JWTSecret string
	JWTExpiry int // 小時

	// 遊戲配置
	PlayerPayout           float64
	BankerPayout          float64
	TiePayout             float64
	Lucky6_2CardsPayout   float64
	Lucky6_3CardsPayout   float64
	BankerLucky6_2Cards   float64
	BankerLucky6_3Cards   float64
}

var AppConfig Config

// LoadConfig 載入配置
func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	AppConfig = Config{
		// 數據庫配置
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBTimezone: os.Getenv("DB_TIMEZONE"),

		// 日志配置
		LogLevel: os.Getenv("LOG_LEVEL"),

		// JWT配置
		JWTSecret: os.Getenv("JWT_SECRET"),
		JWTExpiry: getEnvAsInt("JWT_EXPIRY", 24), // 默認24小時

		// 遊戲配置
		PlayerPayout:         getEnvAsFloat("PLAYER_PAYOUT", 1.0),
		BankerPayout:        getEnvAsFloat("BANKER_PAYOUT", 1.0),
		TiePayout:           getEnvAsFloat("TIE_PAYOUT", 8.0),
		Lucky6_2CardsPayout: getEnvAsFloat("LUCKY6_2CARDS_PAYOUT", 12.0),
		Lucky6_3CardsPayout: getEnvAsFloat("LUCKY6_3CARDS_PAYOUT", 20.0),
		BankerLucky6_2Cards: getEnvAsFloat("BANKER_LUCKY6_2CARDS_PAYOUT", 0.5),
		BankerLucky6_3Cards: getEnvAsFloat("BANKER_LUCKY6_3CARDS_PAYOUT", 0.95),
	}

	return nil
}

// getEnvAsInt 獲取環境變數的整數值
func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// getEnvAsFloat 獲取環境變數的浮點數值
func getEnvAsFloat(key string, defaultVal float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}
