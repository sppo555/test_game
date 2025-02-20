# 分布式追蹤系統整合文檔

## 專案概述
本專案實現了一個基於 OpenResty 和 Golang 的微服務架構，並整合了 Jaeger 分布式追蹤系統，用於監控和分析服務間的調用關係。

## 系統架構
系統主要由以下組件構成：
- **OpenResty 網關**：作為系統的入口點，處理請求路由和負載均衡
- **Golang 服務**：包含 Webhook 處理和其他業務邏輯
- **Jaeger**：分布式追蹤系統，用於監控和診斷微服務架構

## 主要功能
1. **OpenResty 網關**
   - 請求轉發和負載均衡
   - 通過 Lua 模組實現分布式追蹤
   - 自定義請求處理邏輯

2. **Golang 服務**
   - Webhook 處理服務
   - 與 Jaeger 的原生集成
   - 健康檢查接口

3. **分布式追蹤**
   - 完整的請求鏈路追蹤
   - 性能監控和問題診斷
   - 跨服務調用可視化

## 快速開始

### 環境要求
- Docker 和 Docker Compose
- OpenResty
- Golang 1.20+
- Jaeger

### 啟動服務
```bash
# 在專案根目錄執行
docker-compose up -d
```

### 驗證部署
1. 訪問 Jaeger UI：
   ```
   http://localhost:16686
   ```

2. 檢查服務健康狀態：
   ```
   curl http://localhost:8080/health
   ```

## 配置說明

### OpenResty 配置
OpenResty 的配置主要在 `nginx.conf` 和 `default.conf` 中定義，包含：
- 路由規則
- 負載均衡設置
- Lua 模組配置

### Golang 服務配置
主要環境變量：
- `JAEGER_ENDPOINT`：Jaeger 收集器地址（默認：http://jaeger:14268/api/traces）
- `SERVICE_NAME`：服務名稱

### Jaeger 配置
Jaeger 通過 Docker Compose 進行配置，主要端口：
- UI 界面：16686
- 收集器：14268
- Agent：6831/UDP

## 監控和追蹤

### 查看追蹤數據
1. 打開 Jaeger UI（http://localhost:16686）
2. 選擇需要查看的服務
3. 設置時間範圍
4. 查看調用鏈路圖

### 常見監控指標
- 請求延遲
- 錯誤率
- 服務依賴關係

## 故障排除
1. 服務無法啟動
   - 檢查 Docker 容器狀態
   - 查看服務日誌

2. 追蹤數據未顯示
   - 確認 Jaeger 服務狀態
   - 檢查服務配置是否正確

3. 網關連接問題
   - 檢查 OpenResty 配置
   - 確認網絡連接狀態

## 開發指南

### 添加新的追蹤點
1. OpenResty 中添加追蹤：
```lua
local trace_ctx = {
    trace_id = ngx.var.trace_id,
    parent_id = ngx.var.parent_span_id
}
jaeger_exporter.export_span(trace_ctx, "operation_name", ngx.now(), duration, tags)
```

2. Golang 中添加追蹤：
```go
span := tracer.Start(ctx, "operation_name")
defer span.End()
```

## 維護和更新
- 定期檢查並更新依賴
- 監控系統性能
- 根據追蹤數據優化系統

## 注意事項
1. 生產環境部署前請確保：
   - 所有敏感信息已經正確配置
   - 服務間的網絡連接正常
   - 監控告警已經設置

2. 安全建議：
   - 限制 Jaeger UI 訪問權限
   - 定期更新依賴包
   - 監控異常訪問行為
