package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackNotifier は Slack Incoming Webhook へ通知します。
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier は SlackNotifier を生成します。
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

type slackPayload struct {
	Text string `json:"text"`
}

// Notify は Slack へメッセージを送信します。
func (s *SlackNotifier) Notify(ctx context.Context, message string) error {
	payload := slackPayload{Text: message}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: ペイロードのエンコードに失敗しました: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: リクエストの生成に失敗しました: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack: リクエストの送信に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: 予期しないステータスコード: %d", resp.StatusCode)
	}

	return nil
}
