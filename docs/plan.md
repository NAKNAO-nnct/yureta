# 地震速報 bot

## 概要

P2P地震情報の WebSocket API に接続し、地震・津波・緊急地震速報をリアルタイムに受信して Slack などへ通知するボット。

- データソース: [P2P地震情報 JSON API v2](https://www.p2pquake.net/develop/json_api_v2/)
- WebSocket エンドポイント: `wss://api.p2pquake.net/v2/ws`
- サンドボックス（テスト用）: `wss://api-realtime-sandbox.p2pquake.net/v2/ws`

## 言語・技術選定

**Go** を採用する。

- goroutine による再接続ループと通知の並行処理が容易
- 長時間稼働プロセスとして安定性・メモリ効率が高い
- 単一バイナリで Docker デプロイがシンプル

## ディレクトリ構成

```
yure-bot/
├── cmd/bot/
│   └── main.go              # エントリポイント・DI
├── internal/
│   ├── model/
│   │   └── p2pquake.go      # API レスポンスの型定義
│   ├── client/
│   │   └── websocket.go     # WS 接続・自動再接続ループ
│   ├── handler/
│   │   └── handler.go       # メッセージ受信・重複排除・ルーティング
│   └── notifier/
│       ├── notifier.go      # Notifier インターフェース & MultiNotifier
│       ├── slack.go         # Slack Incoming Webhook 実装
│       └── discord.go       # Discord Webhook 実装（将来）
├── config/
│   └── config.go            # 設定ファイル読み込み（YAML）
├── config.example.yaml
├── .env.example
├── go.mod
└── Dockerfile
```

## アーキテクチャ・データフロー

```
WS Server (api.p2pquake.net)
        │  テキストフレーム (JSON)
        ▼
  client/websocket.go
  ・指数バックオフで自動再接続
  ・受信メッセージを ch chan []byte へ送信
        │
        ▼
  handler/handler.go
  ・code フィールドで型判別
  ・id フィールドで重複排除（直近 200 件を map で保持）
  ・対象コードのみ通知対象とする
        │
        ▼
  MultiNotifier（notifier.go）
  ・[]Notifier を持ち、全て並行 goroutine で呼び出す
  ・各 Notifier のエラーは収集してログ出力（1 件失敗しても他は継続）
        │
        ├── slack.go    （Slack Incoming Webhook）
        ├── discord.go  （Discord Webhook、将来）
        └── ...         （LINE Notify など追加可能）
```

## 対応する情報コード

| コード | 内容 | 優先度 |
|--------|------|--------|
| 551 | 地震情報 (JMAQuake) | 高 |
| 552 | 津波予報 (JMATsunami) | 高 |
| 554 | 緊急地震速報 発表検出 (EEWDetection) | 中 |
| 556 | 緊急地震速報・警報 (EEW) | 高 |
| 9611 | 地震感知評価 (UserquakeEvaluation) | 低（設定で on/off） |

## 設定（YAML ファイル）

複数の送信先を柔軟に管理するため、環境変数ではなく YAML ファイルで設定する。  
ファイルパスは環境変数 `CONFIG_FILE`（デフォルト: `config.yaml`）で指定する。

```yaml
websocket:
  url: wss://api.p2pquake.net/v2/ws
  reconnect_initial_interval: 1s
  reconnect_max_interval: 60s

filter:
  notify_codes: [551, 552, 556]   # 通知するコード
  min_scale: 30                   # 最小震度（0=全て, 30=震度3以上）

notifiers:
  - type: slack
    url: https://hooks.slack.com/services/XXX/YYY/ZZZ
  - type: slack
    url: https://hooks.slack.com/services/AAA/BBB/CCC  # 別チャンネルも可
  - type: discord
    url: https://discord.com/api/webhooks/XXX/YYY
```

### Notifier の型一覧

| `type` | 説明 |
|--------|------|
| `slack` | Slack Incoming Webhook |
| `discord` | Discord Webhook（将来実装） |

## 自動再接続ロジック

指数バックオフを採用する。接続成功時はインターバルを初期値にリセットする。

```
interval = initialInterval
for {
    err := connect()
    if err == nil {
        interval = initialInterval
        continue  // 正常切断 or サーバ主導切断後にすぐ再接続
    }
    sleep(interval)
    interval = min(interval * 2, maxInterval)
}
```

## 重複排除

WebSocket 仕様上、同一メッセージが複数回配信される可能性がある。  
受信した `id` フィールドを直近 200 件分 `map[string]struct{}` に保持し、既出の場合は処理をスキップする。

## 主要な依存ライブラリ

| ライブラリ | 用途 |
|-----------|------|
| `nhooyr.io/websocket` | WebSocket クライアント |
| `gopkg.in/yaml.v3` | YAML 設定ファイル読み込み |

## 実装ステップ

1. `go mod init` + 依存ライブラリ追加
2. `model/p2pquake.go` — 型定義（BasicData, JMAQuake, EEW, JMATsunami）
3. `config/config.go` — YAML 設定ファイル読み込み
4. `notifier/notifier.go` — `Notifier` インターフェース と `MultiNotifier` 実装
5. `notifier/slack.go` — Slack Webhook 通知
6. `client/websocket.go` — 接続・再接続ループ
7. `handler/handler.go` — 受信処理・重複排除・通知振り分け
8. `cmd/bot/main.go` — 設定を読んで Notifier を動的に組み立て・起動
9. `Dockerfile` — マルチステージビルドでコンテナ化
