package verify

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/letron/verify/internal/config"
)

type SQLVerifier struct {
	db  *sql.DB
	cfg *config.Config
}

func NewSQLVerifier(cfg *config.Config) (*SQLVerifier, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	// 測試連線
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return &SQLVerifier{
		db:  db,
		cfg: cfg,
	}, nil
}

func (v *SQLVerifier) Close() {
	if v.db != nil {
		v.db.Close()
	}
}

type ValidationResult struct {
	GameID      string
	Rule        string
	Description string
}

type ValidationResults struct {
	ValidGames   []string
	InvalidGames []ValidationResult
}

type ValidationRule struct {
	Name              string `json:"name"`
	ChineseName       string `json:"chinese_name"`
	Query             string `json:"query"`
	Description       string `json:"description"`
	ChineseDescription string `json:"chinese_description"`
	Enabled           bool   `json:"enabled"`
}

type ValidationRules struct {
	Rules []ValidationRule `json:"rules"`
}

func LoadValidationRules(configPath string) (*ValidationRules, error) {
	// 嘗試從多個可能的位置讀取配置文件
	possiblePaths := []string{
		configPath,                                    // 直接使用傳入的路徑
		filepath.Join(".", configPath),               // 相對於當前目錄
		filepath.Join("..", configPath),              // 相對於上級目錄
		filepath.Join("..", "..", configPath),        // 相對於上上級目錄
		filepath.Join("..", "..", "..", configPath),  // 相對於上上上級目錄
	}

	var data []byte
	var err error
	var loadedPath string

	// 嘗試每個可能的路徑
	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			loadedPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("讀取配置文件失敗，嘗試過以下路徑: %v, 錯誤: %v", possiblePaths, err)
	}

	// 解析 JSON
	var rules ValidationRules
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("解析配置文件失敗 (%s): %v", loadedPath, err)
	}

	return &rules, nil
}

func ValidateGames(db *sql.DB, rules *ValidationRules) (*ValidationResults, error) {
	results := &ValidationResults{
		ValidGames:   make([]string, 0),
		InvalidGames: make([]ValidationResult, 0),
	}

	// 用於追踪每個遊戲違反的規則
	invalidGameMap := make(map[string]bool)

	// 對每個啟用的規則執行驗證
	for _, rule := range rules.Rules {
		// 跳過未啟用的規則
		if !rule.Enabled {
			continue
		}

		rows, err := db.Query(rule.Query)
		if err != nil {
			return nil, fmt.Errorf("執行驗證規則 '%s' 失敗: %v", rule.Name, err)
		}
		defer rows.Close()

		for rows.Next() {
			var gameID string
			var otherCols []sql.NullString
			scanArgs := make([]interface{}, 1)
			scanArgs[0] = &gameID
			
			// 動態創建掃描參數
			cols, _ := rows.Columns()
			for range cols[1:] {
				var col sql.NullString
				scanArgs = append(scanArgs, &col)
				otherCols = append(otherCols, col)
			}

			if err := rows.Scan(scanArgs...); err != nil {
				return nil, fmt.Errorf("掃描結果失敗: %v", err)
			}

			invalidGameMap[gameID] = true
			results.InvalidGames = append(results.InvalidGames, ValidationResult{
				GameID:      gameID,
				Rule:        rule.ChineseName,  // 使用中文名稱
				Description: rule.ChineseDescription,  // 使用中文描述
			})
		}
	}

	// 獲取所有遊戲ID
	rows, err := db.Query("SELECT game_id FROM game_records")
	if err != nil {
		return nil, fmt.Errorf("獲取所有遊戲記錄失敗: %v", err)
	}
	defer rows.Close()

	// 將未出現在無效遊戲列表中的遊戲添加到有效遊戲列表
	for rows.Next() {
		var gameID string
		if err := rows.Scan(&gameID); err != nil {
			return nil, fmt.Errorf("掃描遊戲ID失敗: %v", err)
		}
		if !invalidGameMap[gameID] {
			results.ValidGames = append(results.ValidGames, gameID)
		}
	}

	return results, nil
}

func (v *SQLVerifier) VerifyGames() error {
	// 從環境變數讀取驗證規則文件路徑
	rulesPath := os.Getenv("VALIDATION_RULES_PATH")
	if rulesPath == "" {
		return fmt.Errorf("VALIDATION_RULES_PATH environment variable is not set")
	}

	// 加載驗證規則
	rules, err := LoadValidationRules(rulesPath)
	if err != nil {
		return fmt.Errorf("加載驗證規則失敗: %v", err)
	}

	results, err := ValidateGames(v.db, rules)
	if err != nil {
		return fmt.Errorf("驗證遊戲失敗: %v", err)
	}

	// 獲取總遊戲數
	var totalGames int
	err = v.db.QueryRow("SELECT COUNT(*) FROM game_records").Scan(&totalGames)
	if err != nil {
		return fmt.Errorf("獲取總遊戲數失敗: %v", err)
	}

	// 計算實際的無效遊戲數（去重）
	uniqueInvalidGames := make(map[string]bool)
	for _, result := range results.InvalidGames {
		uniqueInvalidGames[result.GameID] = true
	}

	fmt.Println("\nRunning in SQL verification mode...")

	if len(results.InvalidGames) > 0 {
		fmt.Println("\n按規則列出無效遊戲:")
		ruleMap := make(map[string][]ValidationResult)
		for _, result := range results.InvalidGames {
			ruleMap[result.Rule] = append(ruleMap[result.Rule], result)
		}

		// 按規則名稱排序
		for _, rule := range rules.Rules {
			if ruleResults, ok := ruleMap[rule.ChineseName]; ok {
				fmt.Printf("\n%s:\n", rule.ChineseName)
				for _, result := range ruleResults {
					fmt.Printf("- 遊戲 %s: %s\n", result.GameID, result.Description)
				}
			}
		}
	}

	// 在最後輸出統計結果
	fmt.Println("\n=== SQL 驗證結果 ===")
	fmt.Printf("檢查遊戲總數: %d\n", totalGames)
	fmt.Printf("有效遊戲數: %d\n", len(results.ValidGames))
	fmt.Printf("無效遊戲數: %d\n", len(uniqueInvalidGames))

	return nil
}