package main

import (
	"baccarat/config"
	"baccarat/db"
	"baccarat/pkg/logger"
	"database/sql"
	"fmt"
	"time"
)

func main() {
	// 加載配置
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// 初始化日志
	logger.InitLogger()

	// 連接數據庫
	if err := db.InitDB(); err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}
	defer db.DB.Close()

	// 創建測試表（如果不存在）
	_, err := db.DB.Exec(`
		CREATE TABLE IF NOT EXISTS timezone_test (
			id INT AUTO_INCREMENT PRIMARY KEY,
			test_name VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Printf("Error creating test table: %v\n", err)
		return
	}

	// 插入測試數據
	result, err := db.DB.Exec("INSERT INTO timezone_test (test_name) VALUES (?)", "timezone_test")
	if err != nil {
		fmt.Printf("Error inserting test data: %v\n", err)
		return
	}

	// 獲取插入的記錄ID
	id, _ := result.LastInsertId()

	// 查詢插入的數據
	var (
		testName  string
		createdAt time.Time
	)
	err = db.DB.QueryRow("SELECT test_name, created_at FROM timezone_test WHERE id = ?", id).Scan(&testName, &createdAt)
	if err != nil {
		fmt.Printf("Error querying test data: %v\n", err)
		return
	}

	// 打印時區信息
	fmt.Printf("Database Timezone Setting: %s\n", config.AppConfig.DBTimezone)
	fmt.Printf("Current Local Time: %v\n", time.Now().Format("2006-01-02 15:04:05 -0700 MST"))
	fmt.Printf("Database Record Time: %v\n", createdAt.Format("2006-01-02 15:04:05 -0700 MST"))
	
	// 獲取系統時區設置
	localZone, _ := time.Now().Zone()
	fmt.Printf("System Timezone: %s\n", localZone)

	// 直接從數據庫獲取當前時間
	var dbTime sql.NullString
	err = db.DB.QueryRow("SELECT NOW()").Scan(&dbTime)
	if err != nil {
		fmt.Printf("Error getting database time: %v\n", err)
		return
	}
	fmt.Printf("Direct Database NOW(): %s\n", dbTime.String)
}
