package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/trompot/yure-bot/internal/model"
	"github.com/trompot/yure-bot/internal/notifier"
)

const deduplicateCacheSize = 200

// Handler はメッセージを受信し、重複排除・フィルタリングを行って通知します。
type Handler struct {
	notifier    notifier.Notifier
	notifyCodes map[int]struct{}
	minScale    int

	mu      sync.Mutex
	seenIDs []string // 直近 deduplicateCacheSize 件
	seenSet map[string]struct{}
}

// New は Handler を生成します。
func New(n notifier.Notifier, notifyCodes []int, minScale int) *Handler {
	codes := make(map[int]struct{}, len(notifyCodes))
	for _, c := range notifyCodes {
		codes[c] = struct{}{}
	}
	return &Handler{
		notifier:    n,
		notifyCodes: codes,
		minScale:    minScale,
		seenIDs:     make([]string, 0, deduplicateCacheSize),
		seenSet:     make(map[string]struct{}, deduplicateCacheSize),
	}
}

// Handle は WebSocket から受信した生の JSON を処理します。
func (h *Handler) Handle(data []byte) {
	var base model.BasicData
	if err := json.Unmarshal(data, &base); err != nil {
		slog.Warn("メッセージのパースに失敗しました", "error", err)
		return
	}

	// 重複排除
	if h.isDuplicate(base.ID) {
		slog.Debug("重複メッセージをスキップします", "id", base.ID, "code", base.Code)
		return
	}

	// コードフィルター
	if _, ok := h.notifyCodes[base.Code]; !ok {
		return
	}

	msg, err := h.format(base.Code, data)
	if err != nil {
		slog.Warn("メッセージのフォーマットに失敗しました", "code", base.Code, "error", err)
		return
	}
	if msg == "" {
		return // フィルターで除外
	}

	ctx := context.Background()
	if err := h.notifier.Notify(ctx, msg); err != nil {
		slog.Error("通知に失敗しました", "error", err)
	}
}

// isDuplicate は id が既出かどうかを確認し、新規なら記録します。
func (h *Handler) isDuplicate(id string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.seenSet[id]; exists {
		return true
	}

	// キャッシュが満杯なら最古のエントリを削除
	if len(h.seenIDs) >= deduplicateCacheSize {
		oldest := h.seenIDs[0]
		h.seenIDs = h.seenIDs[1:]
		delete(h.seenSet, oldest)
	}

	h.seenIDs = append(h.seenIDs, id)
	h.seenSet[id] = struct{}{}
	return false
}

// format はコードに応じてメッセージ文字列を生成します。
// フィルター条件を満たさない場合は空文字を返します。
func (h *Handler) format(code int, data []byte) (string, error) {
	switch code {
	case 551:
		return h.formatJMAQuake(data)
	case 552:
		return h.formatJMATsunami(data)
	case 554:
		return h.formatEEWDetection(data)
	case 556:
		return h.formatEEW(data)
	default:
		return "", nil
	}
}

func (h *Handler) formatJMAQuake(data []byte) (string, error) {
	var q model.JMAQuake
	if err := json.Unmarshal(data, &q); err != nil {
		return "", fmt.Errorf("JMAQuake のパースに失敗: %w", err)
	}
	if q.Earthquake == nil {
		return "", nil
	}
	eq := q.Earthquake

	// 震度フィルター
	if eq.MaxScale < h.minScale {
		return "", nil
	}

	tsunami := eq.DomesticTsunami
	if tsunami == "None" {
		tsunami = "なし"
	}

	return fmt.Sprintf(
		"【地震情報】\n震源: %s\nM%.1f 深さ %dkm\n最大震度: %s\n津波: %s\n発生時刻: %s",
		eq.Hypocenter.Name,
		eq.Hypocenter.Magnitude,
		eq.Hypocenter.Depth,
		model.ScaleLabel(eq.MaxScale),
		tsunami,
		eq.Time,
	), nil
}

func (h *Handler) formatJMATsunami(data []byte) (string, error) {
	var t model.JMATsunami
	if err := json.Unmarshal(data, &t); err != nil {
		return "", fmt.Errorf("JMATsunami のパースに失敗: %w", err)
	}
	if t.Cancelled {
		return "【津波予報】津波予報が解除されました。", nil
	}

	gradeLabel := map[string]string{
		"MajorWarning": "大津波警報",
		"Warning":      "津波警報",
		"Watch":        "津波注意報",
	}
	gradeOrder := []string{"MajorWarning", "Warning", "Watch"}
	gradeAreas := make(map[string][]string)
	for _, a := range t.Areas {
		gradeAreas[a.Grade] = append(gradeAreas[a.Grade], a.Name)
	}

	var lines []string
	for _, grade := range gradeOrder {
		names := gradeAreas[grade]
		if len(names) == 0 {
			continue
		}
		label, ok := gradeLabel[grade]
		if !ok {
			label = grade
		}
		lines = append(lines, fmt.Sprintf("%s: %s", label, strings.Join(names, "、")))
	}

	return fmt.Sprintf("【津波予報】\n発表: %s\n%s", t.Issue.Time, strings.Join(lines, "\n")), nil
}

func (h *Handler) formatEEWDetection(data []byte) (string, error) {
	var e model.EEWDetection
	if err := json.Unmarshal(data, &e); err != nil {
		return "", fmt.Errorf("EEWDetection のパースに失敗: %w", err)
	}
	return fmt.Sprintf("【緊急地震速報】発表を検出しました (%s)", e.Type), nil
}

func (h *Handler) formatEEW(data []byte) (string, error) {
	var e model.EEW
	if err := json.Unmarshal(data, &e); err != nil {
		return "", fmt.Errorf("EEW のパースに失敗: %w", err)
	}
	if e.Test {
		return "", nil // テスト電文は通知しない
	}
	if e.Cancelled {
		return fmt.Sprintf("【緊急地震速報】取消 (イベントID: %s)", e.Issue.EventID), nil
	}
	if e.Earthquake == nil {
		return "", nil
	}

	eq := e.Earthquake
	return fmt.Sprintf(
		"【緊急地震速報（警報）】第%s報\n震源: %s\nM%.1f 深さ %dkm\n発表時刻: %s",
		e.Issue.Serial,
		eq.Hypocenter.Name,
		eq.Hypocenter.Magnitude,
		eq.Hypocenter.Depth,
		e.Issue.Time,
	), nil
}
