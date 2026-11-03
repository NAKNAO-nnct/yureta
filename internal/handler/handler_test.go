package handler

import (
	"context"
	"fmt"
	"testing"
)

type recordingNotifier struct {
	messages []string
}

func (n *recordingNotifier) Notify(_ context.Context, message string) error {
	n.messages = append(n.messages, message)
	return nil
}

func TestImmediateNotifyPrefsOverrideMinScale(t *testing.T) {
	tests := []struct {
		name        string
		prefs       []string
		points      string
		maxScale    int
		notifyCodes []int
		wantCount   int
	}{
		{
			name:        "matching prefecture notifies below minimum scale",
			prefs:       []string{"東京都"},
			points:      `[{"pref":"東京都","scale":10}]`,
			maxScale:    10,
			notifyCodes: []int{551},
			wantCount:   1,
		},
		{
			name:        "non-matching prefecture does not override minimum scale",
			prefs:       []string{"東京都"},
			points:      `[{"pref":"大阪府","scale":10}]`,
			maxScale:    10,
			notifyCodes: []int{551},
			wantCount:   0,
		},
		{
			name:        "no configured prefectures keeps minimum scale",
			points:      `[{"pref":"東京都","scale":10}]`,
			maxScale:    10,
			notifyCodes: []int{551},
			wantCount:   0,
		},
		{
			name:        "minimum scale still notifies without matching prefecture",
			prefs:       []string{"東京都"},
			points:      `[{"pref":"大阪府","scale":30}]`,
			maxScale:    30,
			notifyCodes: []int{551},
			wantCount:   1,
		},
		{
			name:        "notify code setting remains effective",
			prefs:       []string{"東京都"},
			points:      `[{"pref":"東京都","scale":10}]`,
			maxScale:    10,
			notifyCodes: []int{552},
			wantCount:   0,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &recordingNotifier{}
			h := New(n, tt.notifyCodes, 30, tt.prefs)
			data := []byte(fmt.Sprintf(
				`{"_id":"test-%d","code":551,"earthquake":{"time":"2026-01-01 00:00:00","hypocenter":{"name":"東京湾","depth":10,"magnitude":3.0},"maxScale":%d,"domesticTsunami":"None"},"points":%s}`,
				i, tt.maxScale, tt.points,
			))

			h.Handle(data)

			if got := len(n.messages); got != tt.wantCount {
				t.Fatalf("notification count = %d, want %d", got, tt.wantCount)
			}
		})
	}
}
