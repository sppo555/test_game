// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"github.com/letron/verify/internal/validator"
	"github.com/letron/verify/internal/config"
	"github.com/letron/verify/internal/db"
	"github.com/letron/verify/internal/api"
)

func validateSingleGame(cfg *config.Config, v *validator.Validator, apiClient *api.APIClient) {
	fmt.Printf("\n=== Validating Single Game ID: %s ===\n", cfg.SingleGameID)
	
	// 從 API 獲取遊戲詳情
	gameDetails, err := apiClient.GetGameDetails(cfg.SingleGameID)
	if err != nil {
		log.Fatalf("Error fetching game details: %v\n", err)
	}

	// 輸出遊戲詳情
	prettyJSON, err := json.MarshalIndent(gameDetails, "", "  ")
	if err != nil {
		log.Printf("Error formatting game details: %v\n", err)
	} else {
		fmt.Printf("\nAPI Response:\n%s\n", string(prettyJSON))
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

		// 輸出遊戲詳情
		prettyJSON, err := json.MarshalIndent(gameDetails, "", "  ")
		if err != nil {
			log.Printf("Error formatting game details: %v\n", err)
		} else {
			fmt.Printf("\nAPI Response:\n%s\n", string(prettyJSON))
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

func main() {
	// 讀取配置
	cfg := config.LoadConfig()

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
