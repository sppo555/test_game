package main

import (
	"baccarat/api"
	"baccarat/config"
	"baccarat/db"
	"baccarat/pkg/logger"
	"log"
	"net/http"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("无法加载配置: %v", err)
	}

	// 初始化日志
	logger.InitLogger()

	// 连接数据库
	err := db.InitDB()
	if err != nil {
		logger.Fatal("数据库连接失败: ", err)
	}
	defer db.DB.Close()

	logger.Info("服务器启动成功")

	// 设置路由
	router := api.NewRouter(db.DB)

	// 启动服务器
	logger.Info("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		logger.Fatal("Error starting server:", err)
	}
}
