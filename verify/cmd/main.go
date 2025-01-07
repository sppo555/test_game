package main

import (
	"fmt"
	"log"
	
	"github.com/letron/verify/internal/validator"
	"github.com/letron/verify/internal/config"
	"github.com/letron/verify/internal/db"
	"github.com/letron/verify/internal/api"
	"github.com/letron/verify/internal/verify"
)

func main() {
	// 讀取配置
	cfg := config.LoadConfig()

	// 如果啟用了 SQL 驗證模式，執行 SQL 驗證
	if cfg.SQLVerifyMode {
		fmt.Println("Running in SQL verification mode...")
		verifier, err := verify.NewSQLVerifier(cfg)
		if err != nil {
			log.Fatalf("Error creating SQL verifier: %v", err)
		}
		defer verifier.Close()

		if err := verifier.VerifyGames(); err != nil {
			log.Fatalf("Error verifying games: %v", err)
		}
		return
	}

	// 否則執行標準驗證
	// 建立資料庫連線
	db := db.NewDB(cfg)
	defer db.Close()

	// 建立 API 客戶端
	apiClient := api.NewAPIClient(cfg.APIURL, cfg.Token)

	// 建立驗證器
	v := validator.NewValidator(cfg, apiClient)

	if cfg.SingleGameMode {
		validateSingleGame(cfg, v, apiClient)
	} else {
		validateMultipleGames(cfg, v, apiClient, db)
	}
}

func validateSingleGame(cfg *config.Config, v *validator.Validator, apiClient *api.APIClient) {
	fmt.Printf("\n=== Validating Single Game ID: %s ===\n", cfg.SingleGameID)
	
	// 從 API 獲取遊戲詳情
	gameDetails, err := apiClient.GetGameDetails(cfg.SingleGameID)
	if err != nil {
		log.Fatalf("Error fetching game details: %v\n", err)
	}

	// 驗證遊戲
	result := v.ValidateGame(gameDetails)
	
	// 輸出驗證結果
	if len(result.InvalidGameIDs) == 0 {
		fmt.Printf("\nValidation Result: Valid\n")
	} else {
		fmt.Printf("\nValidation Result: Invalid\n")
		for _, errMsg := range result.ErrorDetails {
			fmt.Printf("Error: %s\n", errMsg)
		}
	}
}

func validateMultipleGames(cfg *config.Config, v *validator.Validator, apiClient *api.APIClient, db *db.DB) {
	// 查詢最新的 game_ids
	gameIDs, err := db.GetLatestGameIDs(cfg.GameIDLimit)
	if err != nil {
		log.Fatalf("Error fetching game IDs: %v", err)
	}

	if len(gameIDs) == 0 {
		log.Println("No game records found.")
		return
	}

	// 驗證所有遊戲
	var totalResult validator.ValidationResult
	totalResult.TotalGames = len(gameIDs)

	for _, gameID := range gameIDs {
		fmt.Printf("\n=== Game ID: %s ===\n", gameID)
		
		// 從 API 獲取遊戲詳情
		gameDetails, err := apiClient.GetGameDetails(gameID)
		if err != nil {
			log.Printf("Error fetching game details for %s: %v\n", gameID, err)
			continue
		}

		// 驗證遊戲
		result := v.ValidateGame(gameDetails)
		
		// 更新總結果
		if len(result.InvalidGameIDs) > 0 {
			totalResult.InvalidGames++
			totalResult.InvalidGameIDs = append(totalResult.InvalidGameIDs, result.InvalidGameIDs...)
			totalResult.ErrorDetails = append(totalResult.ErrorDetails, result.ErrorDetails...)
		} else {
			totalResult.ValidGames++
		}

		// 輸出單個遊戲的驗證結果
		if len(result.InvalidGameIDs) == 0 {
			fmt.Printf("\nValidation Result: Valid\n")
		} else {
			fmt.Printf("\nValidation Result: Invalid\n")
			for _, errMsg := range result.ErrorDetails {
				fmt.Printf("Error: %s\n", errMsg)
			}
		}
		fmt.Println("--------------------------------------------------")
	}

	// 輸出總結果
	fmt.Println(totalResult.String())
}
