package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config はアプリケーション全体の設定です。
type Config struct {
	WebSocket WebSocketConfig  `yaml:"websocket"`
	Filter    FilterConfig     `yaml:"filter"`
	Notifiers []NotifierConfig `yaml:"notifiers"`
}

// WebSocketConfig は WebSocket 接続の設定です。
type WebSocketConfig struct {
	URL                     string        `yaml:"url"`
	ReconnectInitialInterval time.Duration `yaml:"reconnect_initial_interval"`
	ReconnectMaxInterval    time.Duration `yaml:"reconnect_max_interval"`
}

// FilterConfig は通知フィルターの設定です。
type FilterConfig struct {
	NotifyCodes []int `yaml:"notify_codes"`
	MinScale    int   `yaml:"min_scale"`
}

// NotifierConfig は各通知先の設定です。
type NotifierConfig struct {
	Type string `yaml:"type"`
	URL  string `yaml:"url"`
}

// Load は設定ファイルを読み込み、デフォルト値を補完して返します。
// ファイルパスは環境変数 CONFIG_FILE で指定できます（デフォルト: config.yaml）。
func Load() (*Config, error) {
	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = "config.yaml"
	}

	f, err := os.Open(path) // #nosec G304 — ユーザーが明示的に指定するパス
	if err != nil {
		return nil, fmt.Errorf("設定ファイルを開けません (%s): %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルのパースに失敗しました: %w", err)
	}

	setDefaults(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.WebSocket.URL == "" {
		cfg.WebSocket.URL = "wss://api.p2pquake.net/v2/ws"
	}
	if cfg.WebSocket.ReconnectInitialInterval == 0 {
		cfg.WebSocket.ReconnectInitialInterval = 1 * time.Second
	}
	if cfg.WebSocket.ReconnectMaxInterval == 0 {
		cfg.WebSocket.ReconnectMaxInterval = 60 * time.Second
	}
	if len(cfg.Filter.NotifyCodes) == 0 {
		cfg.Filter.NotifyCodes = []int{551, 552, 556}
	}
}

func validate(cfg *Config) error {
	if len(cfg.Notifiers) == 0 {
		return fmt.Errorf("notifiers が設定されていません")
	}
	for i, n := range cfg.Notifiers {
		if n.URL == "" {
			return fmt.Errorf("notifiers[%d]: url が設定されていません", i)
		}
		if n.Type != "slack" && n.Type != "discord" {
			return fmt.Errorf("notifiers[%d]: 未対応の type です: %s", i, n.Type)
		}
	}
	return nil
}
