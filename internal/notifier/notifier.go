package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Notifier は通知を送信するインターフェースです。
type Notifier interface {
	// Notify はメッセージを送信します。
	Notify(ctx context.Context, message string) error
}

// MultiNotifier は複数の Notifier へ並行して通知します。
// 一部が失敗しても残りは継続して送信されます。
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier は MultiNotifier を生成します。
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers}
}

// Notify は全ての Notifier へ並行して通知します。
// 全ての送信が完了するまで待機します。エラーは個別にログ出力されます。
func (m *MultiNotifier) Notify(ctx context.Context, message string) error {
	var wg sync.WaitGroup
	errs := make([]error, len(m.notifiers))

	for i, n := range m.notifiers {
		wg.Add(1)
		go func(idx int, notifier Notifier) {
			defer wg.Done()
			if err := notifier.Notify(ctx, message); err != nil {
				errs[idx] = err
				slog.Error("通知の送信に失敗しました", "index", idx, "error", err)
			}
		}(i, n)
	}

	wg.Wait()

	// 1件でも失敗したら最初のエラーを返す（呼び出し元でのログ用）
	for _, err := range errs {
		if err != nil {
			return fmt.Errorf("一部の通知先への送信に失敗しました: %w", err)
		}
	}
	return nil
}
