# ビルドステージ
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o yure-bot ./cmd/bot

# 実行ステージ
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/yure-bot /yure-bot

ENV CONFIG_FILE=/config.yaml

ENTRYPOINT ["/yure-bot"]
