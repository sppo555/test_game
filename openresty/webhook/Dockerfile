FROM golang:1.19-alpine

WORKDIR /app

# 安裝必要的系統依賴
RUN apk add --no-cache curl

# 複製源代碼
COPY . .

# 編譯應用
RUN go build -o webhook

# 設置執行權限
RUN chmod +x webhook

# 暴露端口
EXPOSE 8080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 運行應用
CMD ["./webhook", "-port", "8080"]
