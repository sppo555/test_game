# Build stage
FROM golang:1.20-alpine AS builder

# 安裝編譯工具
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# 安裝必要的運行時依賴
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# 設置時區
ENV TZ=Asia/Taipei

WORKDIR /app
COPY --from=builder /app/main .

# 設置環境變量
ENV PORT=8081
ENV JAEGER_ENDPOINT=http://jaeger:14268/api/traces

EXPOSE 8081

# 使用非 root 用戶運行
RUN adduser -D appuser
USER appuser

CMD ["./main"]
