# 遊戲驗證系統

[![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

本系統專屬於Alex Lin開發的百家樂遊戲記錄進行完整性檢查，支援 SQL 直接驗證和 API 接口驗證兩種模式。


## 主要功能

### 🛠️ 核心特性
- **雙模式驗證引擎**
  - **🔍 SQL 直接驗證**：直接對 MySQL 資料庫執行預定義驗證規則
  - **🌐 API 接口驗證**：通過 REST API 獲取遊戲詳情進行驗證
- **📝 動態規則配置**
  - JSON 格式驗證規則 (路徑可通過 `VALIDATION_RULES_PATH` 配置) 相容SQL驗證使用
  - 支援中/英雙語規則描述
  - 即時啟用/停用單個規則
- **⚙️ 環境驅動配置**
  - 數據庫憑證管理
  - API 令牌配置
  - 驗證模式切換 (單遊戲/批量模式)

## 🚀 快速開始

### 前置需求
- Go 1.20+
- MySQL 8.0+
- 遊戲服務 API 訪問權限

```bash
# 1. 克隆存儲庫
git clone https://github.com/sppo555/test_game.git
cd test_game/verify

# 2. 初始化環境
cp .env.example .env
vim .env  # 編輯配置

# 3. 啟動驗證
go run cmd/main.go
```

## ⚙️ 環境配置
編輯 `.env` 文件：
```ini
# 數據庫配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=admin
DB_PASSWORD=secret
DB_NAME=game_db

# 驗證模式選擇
SQL_VERIFY_MODE=true          # 啟用 SQL 驗證
SINGLE_GAME_MODE=false        # 單遊戲驗證模式
SINGLE_GAME_ID=""            # 指定驗證的遊戲 ID（僅在 SINGLE_GAME_MODE=true 時使用）
GAME_ID_LIMIT=100            # 批量驗證時的遊戲數量限制

# 驗證規則配置
VALIDATION_RULES_PATH=/tmp/config/rules.json

# API 配置
API_URL="http://localhost:8080/api/game/details"
API_TOKEN=your_api_key_here
```

## 🎮 驗證模式
系統支援兩種驗證模式：

### 1. 批量驗證模式
適合一次性驗證多筆遊戲記錄：
```bash
# 在 .env 設置
SINGLE_GAME_MODE=false
GAME_ID_LIMIT=100  # 設置要驗證的遊戲數量

# 賠率設定
PLAYER_PAYOUT=2.0 # 本金加賠率
BANKER_PAYOUT=2.0 # 本金加賠率
TIE_PAYOUT=8.0
LUCKY6_2CARDS_PAYOUT=12.0
LUCKY6_3CARDS_PAYOUT=20.0
BANKER_LUCKY6_2CARDS_PAYOUT=1.5
BANKER_LUCKY6_3CARDS_PAYOUT=1.95

# 執行驗證
go run cmd/main.go
```

### 2. 單遊戲驗證模式
用於針對特定遊戲 ID 進行詳細驗證：
```bash
# 在 .env 設置
SINGLE_GAME_MODE=true
SINGLE_GAME_ID="506fe804-0adf-43dd-87b4-c30c2c2a7a53"

# 執行驗證
go run cmd/main.go
```

單遊戲模式會輸出更詳細的驗證資訊，包括：
- 完整的遊戲記錄
- 詳細的驗證步驟
- 所有違規項目的具體說明

## 📜 SQL驗證規則配置
VALIDATION_RULES_PATH 配置範例規則文件 (`config/rules.json`)：
```json
{
  "rules": [
    {
      "name": "banker_score_3_rule",
      "chinese_name": "莊家分數3抽牌規則",
      "query": "SELECT game_id FROM game_records WHERE banker_score=3 AND player_third_card NOT IN (0,1,8,9)",
      "description": "Banker should draw when score is 3 only if player's third card is 0,1,8,9",
      "chinese_description": "莊家分數3時，僅在玩家第三張牌為0/1/8/9時抽牌",
      "enabled": true
    },
    {
      "name": "payout_validation",
      "chinese_name": "賠率驗證",
      "query": "SELECT game_id FROM payouts WHERE ABS(actual_payout - expected_payout) > 0.01",
      "description": "Verify payout accuracy within 1% tolerance",
      "chinese_description": "檢查實際賠率與預期值差異是否超過1%",
      "enabled": true
    }
  ]
}
```

## 📂 項目結構
```
game-validator/
├── cmd/                  # CLI 入口點
│   └── main.go
├── internal/
│   ├── api/              # API 客戶端實現
│   ├── config/           # 配置載入邏輯
│   ├── db/               # 數據庫模組
│   ├── verify/           # 核心驗證引擎
│   └── validator/        # 遊戲規則驗證器
├── scripts/
│   └── init.sql          # 數據庫初始化腳本
└── .env                  # 環境配置
```

## 🚨 常見問題
### 錯誤："Invalid BATCH_SIZE"
解決方案：已移除 BATCH_SIZE 配置，無需處理

### 錯誤："VALIDATION_RULES_PATH not set"
```bash
# 在 .env 添加配置
echo 'VALIDATION_RULES_PATH=config/rules.json' >> .env
```

## 📄 許可證
MIT License - 詳見 [LICENSE](LICENSE) 文件
