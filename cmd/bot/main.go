package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/trompot/yure-bot/config"
	"github.com/trompot/yure-bot/internal/client"
	"github.com/trompot/yure-bot/internal/handler"
	"github.com/trompot/yure-bot/internal/notifier"
)

func main() {
	logOpts := &slog.HandlerOptions{Level: slog.LevelInfo}
	var logHandler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		logHandler = slog.NewJSONHandler(os.Stdout, logOpts)
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, logOpts)
	}
	slog.SetDefault(slog.New(logHandler))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("設定の読み込みに失敗しました", "error", err)
		os.Exit(1)
	}

	// 通知先を構築
	multi, err := buildNotifier(cfg)
	if err != nil {
		slog.Error("Notifier の初期化に失敗しました", "error", err)
		os.Exit(1)
	}

	// ハンドラーを構築
	h := handler.New(multi, cfg.Filter.NotifyCodes, cfg.Filter.MinScale)

	// WebSocket クライアントを構築
	wsClient := client.NewClient(
		cfg.WebSocket.URL,
		cfg.WebSocket.ReconnectInitialInterval,
		cfg.WebSocket.ReconnectMaxInterval,
		h.Handle,
	)

	// シグナルハンドリング
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("yure-bot を起動します", "url", cfg.WebSocket.URL, "notifiers", len(cfg.Notifiers))
	wsClient.Run(ctx)
	slog.Info("yure-bot を終了します")
}

// buildNotifier は設定から Notifier のリストを構築して MultiNotifier を返します。
func buildNotifier(cfg *config.Config) (*notifier.MultiNotifier, error) {
	notifiers := make([]notifier.Notifier, 0, len(cfg.Notifiers))

	for i, nc := range cfg.Notifiers {
		switch nc.Type {
		case "slack":
			notifiers = append(notifiers, notifier.NewSlackNotifier(nc.URL))
			slog.Info("Slack Notifier を登録しました", "index", i)
		case "discord":
			notifiers = append(notifiers, notifier.NewDiscordNotifier(nc.URL))
			slog.Info("Discord Notifier を登録しました", "index", i)
		default:
			return nil, fmt.Errorf("未対応の notifier type: %s", nc.Type)
		}
	}

	return notifier.NewMultiNotifier(notifiers...), nil
}
