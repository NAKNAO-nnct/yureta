package client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nhooyr.io/websocket"
)

// MessageHandler は受信したメッセージを処理する関数型です。
type MessageHandler func(data []byte)

// Client は WebSocket クライアントです。
type Client struct {
	url                     string
	reconnectInitialInterval time.Duration
	reconnectMaxInterval    time.Duration
	handler                 MessageHandler
}

// NewClient は Client を生成します。
func NewClient(url string, initialInterval, maxInterval time.Duration, handler MessageHandler) *Client {
	return &Client{
		url:                     url,
		reconnectInitialInterval: initialInterval,
		reconnectMaxInterval:    maxInterval,
		handler:                 handler,
	}
}

// Run は WebSocket に接続し、切断されたら指数バックオフで再接続し続けます。
// ctx がキャンセルされると終了します。
func (c *Client) Run(ctx context.Context) {
	interval := c.reconnectInitialInterval

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		slog.Info("WebSocket に接続します", "url", c.url)
		err := c.connect(ctx)
		if err == nil || ctx.Err() != nil {
			// 正常終了 or コンテキストキャンセル
			interval = c.reconnectInitialInterval
			if ctx.Err() != nil {
				return
			}
			slog.Info("WebSocket が切断されました。再接続します")
			continue
		}

		slog.Error("WebSocket 接続エラー。再接続を待機します", "error", err, "interval", interval)
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}

		interval = min(interval*2, c.reconnectMaxInterval)
	}
}

// connect は WebSocket に接続し、メッセージを受信し続けます。
func (c *Client) connect(ctx context.Context) error {
	conn, _, err := websocket.Dial(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("接続失敗: %w", err)
	}
	defer conn.CloseNow()

	slog.Info("WebSocket に接続しました")

	for {
		msgType, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("受信エラー: %w", err)
		}
		if msgType != websocket.MessageText {
			continue
		}
		c.handler(data)
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
