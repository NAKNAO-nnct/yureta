package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// DiscordNotifier は Discord Webhook へ通知します。
type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewDiscordNotifier は DiscordNotifier を生成します。
func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

type discordPayload struct {
	Content string `json:"content"`
}

// Notify は Discord へメッセージを送信します。
func (d *DiscordNotifier) Notify(ctx context.Context, message string) error {
	payload := discordPayload{Content: message}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: ペイロードのエンコードに失敗しました: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: リクエストの生成に失敗しました: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("discord: リクエストの送信に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	// Discord Webhook は成功時に 204 No Content を返す
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord: 予期しないステータスコード: %d", resp.StatusCode)
	}

	return nil
}
