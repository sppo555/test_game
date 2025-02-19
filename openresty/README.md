            # OpenResty 追蹤與監控系統

## 目錄
- [系統概述](#系統概述)
  - [架構設計](#架構設計)
  - [主要功能](#主要功能)
- [快速開始](#快速開始)
  - [環境要求](#環境要求)
  - [安裝步驟](#安裝步驟)
  - [基本配置](#基本配置)
- [系統組件](#系統組件)
  - [OpenResty](#openresty)
  - [Webhook 服務](#webhook-服務)
  - [Trace Context](#trace-context)
- [配置說明](#配置說明)
  - [環境變量](#環境變量)
  - [Metrics 收集](#metrics-收集)
  - [數據脫敏](#數據脫敏)
  - [性能調優](#性能調優)
- [使用示例](#使用示例)
- [最佳實踐](#最佳實踐)
- [故障排除](#故障排除)
- [Docker 部署](#docker-部署)
  - [目錄結構](#目錄結構)
  - [Docker 環境配置](#docker-環境配置)
  - [部署步驟](#部署步驟)
  - [配置說明](#配置說明-1)
  - [健康檢查](#健康檢查)
  - [性能調優](#性能調優-1)
  - [常見問題](#常見問題-1)
  - [備份和恢復](#備份和恢復)

## 系統概述

### 架構設計

本系統由以下主要組件組成：

1. **OpenResty 網關**
   - 處理所有入站請求
   - 生成和傳遞 trace context
   - 收集性能指標
   - 實現請求過濾和轉發

2. **Webhook 服務**
   - 接收並處理 trace 數據
   - 提供 metrics 查詢接口
   - 支持數據持久化
   - 提供告警機制

3. **監控系統**
   - 實時數據展示
   - 性能指標分析
   - 異常檢測
   - 報告生成

### 主要功能

- W3C Trace Context 支持
- 分布式請求追蹤
- 性能指標收集
- 數據脫敏處理
- 可配置的日誌格式
- 靈活的開關控制
- 告警機制

## 快速開始

### 環境要求

- Docker 20.10+
- Go 1.19+
- OpenResty 1.21.4.1+
- Redis 6.0+ (可選，用於數據緩存)

### 安裝步驟

1. **克隆代碼倉庫**
```bash
git clone [repository-url]
cd openresty
```

2. **構建 Docker 鏡像**
```bash
docker build -t openresty-trace .
```

3. **啟動服務**
```bash
docker-compose up -d
```

### 基本配置

1. **設置環境變量**
```bash
# .env 文件
ENABLE_TRACE=on
ENABLE_TRACE_METRICS=on
TRACE_BACKEND_URL=http://webhook:8080
```

2. **配置 Webhook**
```bash
cd webhook
go build
./webhook -port 8080
```

## 系統組件

### OpenResty

OpenResty 作為系統的入口網關，主要職責包括：

1. **請求處理**
   - URL 路由
   - 請求轉發
   - 負載均衡
   - 訪問控制

2. **追蹤功能**
   - 生成 trace_id
   - 傳遞 trace context
   - 記錄請求信息
   - 性能數據採集

3. **配置示例**
```nginx
http {
    init_by_lua_block {
        require "trace_context"
    }
    
    server {
        location /api {
            set $service_name 'api-service';
            set $enable_trace 'on';
            proxy_pass http://backend;
        }
    }
}
```

### Webhook 服務

Webhook 服務負責處理和存儲追蹤數據：

1. **數據處理**
   - 接收 trace 數據
   - 解析 trace context
   - 數據驗證和清理
   - 指標計算

2. **API 端點**
   - `/trace` - 接收追蹤數據
   - `/metrics` - 提供指標查詢
   - `/health` - 健康檢查
   - `/config` - 配置管理

3. **使用方法**
```bash
# 啟動 webhook 服務
./webhook -config config.yaml

# 測試追蹤功能
curl -H "Traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01" \
     http://localhost:8080/trace
```

### Trace Context

W3C Trace Context 實現：

1. **格式規範**
```
traceparent: 00-[trace-id]-[parent-id]-[flags]
```

2. **示例值**
```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

3. **處理流程**
   - 檢查請求頭
   - 提取或生成 trace-id
   - 創建 span context
   - 傳遞給下游服務

## 配置說明

### 環境變量

| 變量名 | 說明 | 可選值 | 默認值 |
|--------|------|--------|--------|
| ENABLE_TRACE | 控制追蹤功能 | on/off | on |
| ENABLE_TRACE_METRICS | 控制指標收集 | on/off | off |

### Metrics 收集

收集的指標包括：

| 指標名稱 | 說明 | 單位 |
|----------|------|------|
| requests_total | 總請求數 | 計數 |
| errors_total | 錯誤請求數 | 計數 |
| processing_time_total | 總處理時間 | 毫秒 |
| backend_errors | 後端錯誤數 | 計數 |
| batch_size_current | 當前批次大小 | 數量 |
| traces_sent_total | 已發送追蹤數據量 | 計數 |

### 數據脫敏

脫敏配置選項：

```nginx
# 啟用脫敏
set $enable_trace_sanitization 'on';

# 脫敏規則
set $sanitize_auth 'on';      # Authorization headers
set $sanitize_cookie 'on';    # Cookie 值
set $sanitize_token 'on';     # Bearer tokens
set $sanitize_custom_fields 'api-key,session-id';  # 自定義字段
```

### 性能調優

關鍵參數建議值：

| 參數 | 建議值 | 說明 |
|------|--------|------|
| trace_timeout | 500-2000ms | 追蹤超時時間 |
| trace_retry_times | 3-5 | 重試次數 |
| trace_retry_delay | 1-5s | 重試間隔 |
| trace_batch_size | 50-200 | 批次大小 |
| trace_flush_interval | 5-15s | 發送間隔 |

## 使用示例

1. 基本追蹤配置：
```nginx
location /api {
    set $service_name 'api-service';
    set $enable_trace 'on';
    set $enable_trace_metrics 'off';
    proxy_pass http://backend;
}
```

2. 完整監控配置：
```nginx
location /admin {
    set $service_name 'admin-service';
    set $enable_trace 'on';
    set $enable_trace_metrics 'on';
    set $enable_trace_sanitization 'on';
    set $sanitize_custom_fields 'session-id,csrf-token';
    proxy_pass http://admin_backend;
}
```

## 最佳實踐

1. **追蹤配置**
   - 對重要服務啟用完整追蹤
   - 對公開 API 關閉敏感信息收集
   - 根據服務重要性調整超時和重試

2. **性能優化**
   - 適當調整批次大小和發送間隔
   - 避免過頻繁的重試
   - 根據記憶體使用情況調整參數

3. **安全建議**
   - 始終對敏感接口啟用脫敏
   - 定期審查脫敏規則
   - 避免在日誌中記錄敏感信息

4. **監控建議**
   - 對關鍵服務啟用完整 metrics
   - 設置合適的告警閾值
   - 定期分析性能指標

## 故障排除

### 常見問題

1. **Trace ID 未生成**
   - 檢查 `ENABLE_TRACE` 設置
   - 確認 trace_context 模塊已加載
   - 查看 OpenResty 錯誤日誌

2. **Metrics 未收集**
   - 檢查 `ENABLE_TRACE_METRICS` 設置
   - 確認 Webhook 服務運行狀態
   - 檢查網絡連接

3. **性能問題**
   - 調整批次處理參數
   - 檢查日誌級別設置
   - 優化數據存儲方式

### 診斷工具

1. **日誌查看**
```bash
# OpenResty 錯誤日誌
tail -f logs/error.log

# Webhook 服務日誌
tail -f webhook/logs/app.log
```

2. **性能測試**
```bash
# 壓力測試腳本
./benchmark.sh -c 100 -n 10000
```

3. **配置檢查**
```bash
# 檢查 OpenResty 配置
nginx -t

# 檢查 Webhook 配置
./webhook -check-config
```

## Docker 部署

### 目錄結構
```
.
├── docker-compose.yml          # Docker 編排配置
├── dockerfile                  # OpenResty 鏡像配置
├── nginx.conf                  # Nginx 主配置
├── default.conf               # Nginx 默認站點配置
├── logs/                      # 日誌目錄
└── webhook/
    ├── Dockerfile            # Webhook 服務鏡像配置
    ├── webhook.go           # Webhook 服務源碼
    └── logs/                # Webhook 日誌目錄
```

### Docker 環境配置

1. **環境要求**
   - Docker 20.10+
   - Docker Compose 2.0+
   - 至少 2GB 可用記憶體
   - 至少 10GB 磁盤空間

2. **網絡配置**
   - OpenResty: 8080 端口 (外部訪問)
   - Webhook: 8081 端口 (外部訪問)
   - 內部網絡: trace-network (容器間通信)

### 部署步驟

1. **準備環境**
```bash
# 創建必要的目錄
mkdir -p logs webhook/logs
chmod 777 logs webhook/logs
```

2. **構建和啟動服務**
```bash
# 構建鏡像
docker-compose build

# 啟動服務
docker-compose up -d

# 查看服務狀態
docker-compose ps
```

3. **查看日誌**
```bash
# 查看 OpenResty 日誌
docker-compose logs openresty

# 查看 Webhook 日誌
docker-compose logs webhook

# 實時跟蹤日誌
docker-compose logs -f
```

4. **服務管理**
```bash
# 停止服務
docker-compose stop

# 重啟服務
docker-compose restart

# 停止並移除容器
docker-compose down

# 停止並移除容器及卷
docker-compose down -v
```

### 配置說明

1. **OpenResty 容器配置**
```yaml
environment:
  - ENABLE_TRACE=on           # 啟用追蹤
  - ENABLE_TRACE_METRICS=on   # 啟用指標收集
  - TRACE_BACKEND_URL=http://webhook:8080  # Webhook 服務地址
```

2. **Webhook 容器配置**
```yaml
environment:
  - LOG_LEVEL=info           # 日誌級別
  - METRICS_RETENTION=24h    # 指標保留時間
```

### 健康檢查

系統內建了健康檢查機制：

1. **檢查端點**
   - OpenResty: `http://localhost:8080/health`
   - Webhook: `http://localhost:8081/health`

2. **手動檢查**
```bash
# 檢查 OpenResty 狀態
curl http://localhost:8080/health

# 檢查 Webhook 狀態
curl http://localhost:8081/health

# 檢查 Metrics
curl http://localhost:8081/metrics
```

### 性能調優

1. **容器資源限制**
```yaml
# docker-compose.yml 中添加
services:
  openresty:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

2. **日誌配置**
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "100m"
    max-file: "3"
```

### 常見問題

1. **容器無法啟動**
   - 檢查端口衝突
   - 檢查目錄權限
   - 查看容器日誌

2. **服務間無法通信**
   - 確認網絡配置
   - 檢查服務名稱解析
   - 驗證防火牆設置

3. **性能問題**
   - 調整容器資源限制
   - 檢查日誌輸出量
   - 監控系統資源使用

### 備份和恢復

1. **配置備份**
```bash
# 備份配置文件
tar -czf config-backup.tar.gz nginx.conf default.conf

# 備份日誌
tar -czf logs-backup.tar.gz logs/
```

2. **數據卷備份**
```bash
# 創建數據卷備份
docker run --rm -v openresty_data:/data -v $(pwd):/backup alpine tar -czf /backup/data-backup.tar.gz /data
```