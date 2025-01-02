# Verify Service

這是一個遊戲驗證服務，用於驗證遊戲結果和支付。

## 目錄結構

```
verify/
├── cmd/            # 主程序
├── internal/       # 內部包
│   ├── api/       # API 客戶端
│   ├── config/    # 配置
│   ├── db/        # 數據庫
│   └── validator/ # 驗證邏輯
```

## 開始使用

1. 複製 `.env.example` 到 `.env` 並設置必要的環境變量
2. 運行服務：
   ```bash
   go run cmd/main.go
   ```
