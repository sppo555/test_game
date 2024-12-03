package main

import (
	"baccarat/api"
	"baccarat/config"
	"baccarat/db"
	"log"
	"net/http"
)

func main() {
	// 載入配置
	if err := config.LoadConfig(); err != nil {
		log.Fatal("Error loading config:", err)
	}

	// 初始化數據庫連接
	if err := db.InitDB(); err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.DB.Close()

	// 設置路由
	router := api.NewRouter(db.DB)
	router.SetupRoutes()

	// 啟動服務器
	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
