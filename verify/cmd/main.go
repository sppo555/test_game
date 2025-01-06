// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	
	"github.com/letron/verify/internal/validator"
	"github.com/letron/verify/internal/config"
	"github.com/letron/verify/internal/db"
	"github.com/letron/verify/internal/api"
)

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

	// 查詢最新的 game_ids
	gameIDs, err := db.GetLatestGameIDs(cfg.GameIDLimit)
	if err != nil {
		log.Fatalf("Error fetching game IDs: %v", err)
	}

	if len(gameIDs) == 0 {
		log.Println("No game records found.")
		return
	}

	// 儲存驗證結果
	var results []validator.ValidationResult

	// 遍歷每個 game_id，進行驗證
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
		validationResult := v.ValidateGame(gameDetails)
		results = append(results, validationResult)

		// 輸出驗證結果
		if validationResult.IsValid {
			fmt.Printf("\nValidation Result: Valid \n")
		} else {
			fmt.Printf("\nValidation Result: Invalid \n")
			for _, errMsg := range validationResult.Errors {
				fmt.Printf("Error: %s\n", errMsg)
			}
		}
		fmt.Println(strings.Repeat("-", 50))
	}

	// 總結驗證結果
	var invalidCount int
	for _, res := range results {
		if !res.IsValid {
			invalidCount++
		}
	}

	fmt.Printf("\nValidation Summary:\n")
	fmt.Printf("Total games: %d\n", len(results))
	fmt.Printf("Valid games: %d\n", len(results)-invalidCount)
	fmt.Printf("Invalid games: %d\n", invalidCount)
}
