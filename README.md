# 地震情報配信サーバ

[P2P地震情報 JSON API v2](https://www.p2pquake.net/develop/json_api_v2/) の WebSocket に接続し、地震・津波・緊急地震速報をリアルタイムに Slack / Discord へ通知する Bot です。

## 機能

- P2P地震情報 WebSocket API へのリアルタイム接続・自動再接続
- Slack Incoming Webhook / Discord Webhook への通知
- 通知コード（地震情報・津波情報・緊急地震速報）によるフィルタリング
- 最小震度によるフィルタリング
- 複数通知先への同時送信
- 重複メッセージの排除

## 動作確認済み環境

- Go 1.22 以上
- Docker

## セットアップ

### 1. 設定ファイルの準備

```bash
cp config.example.yaml config.yaml
```

`config.yaml` を編集して Slack または Discord の Webhook URL を設定します。

```yaml
websocket:
  url: wss://api.p2pquake.net/v2/ws
  reconnect_initial_interval: 1s
  reconnect_max_interval: 60s

filter:
  notify_codes: [551, 552, 556]   # 551=地震情報, 552=津波情報, 556=緊急地震速報
  min_scale: 30                   # 0=全て, 10=震度1以上, 30=震度3以上, 40=震度4以上

notifiers:
  - type: slack
    url: https://hooks.slack.com/services/XXX/YYY/ZZZ
  # - type: discord
  #   url: https://discord.com/api/webhooks/XXX/YYY
```

### 2. ローカルで実行（Go）

```bash
go mod download
go run ./cmd/bot
```

設定ファイルのパスは環境変数 `CONFIG_FILE` で変更できます（デフォルト: `config.yaml`）。

```bash
CONFIG_FILE=/path/to/config.yaml go run ./cmd/bot
```

### 3. ビルドして実行

```bash
go build -trimpath -ldflags="-s -w" -o yure-bot ./cmd/bot
./yure-bot
```

### 4. Docker で実行

```bash
docker build -t yure-bot .
docker run --rm -v $(pwd)/config.yaml:/config.yaml yure-bot
```

## テスト用サンドボックス

実際の地震がなくてもテストデータが流れるサンドボックス環境が利用できます。`config.yaml` の `websocket.url` を以下に変更してください。

```yaml
websocket:
  url: wss://api-realtime-sandbox.p2pquake.net/v2/ws
```

## 設定リファレンス

| キー | デフォルト値 | 説明 |
|------|-------------|------|
| `websocket.url` | `wss://api.p2pquake.net/v2/ws` | 接続先 WebSocket URL |
| `websocket.reconnect_initial_interval` | `1s` | 再接続の初期待機時間 |
| `websocket.reconnect_max_interval` | `60s` | 再接続の最大待機時間 |
| `filter.notify_codes` | `[551, 552, 556]` | 通知するイベントコード |
| `filter.min_scale` | `0` | 通知する最小震度スケール値 |
| `notifiers[].type` | — | `slack` または `discord` |
| `notifiers[].url` | — | Webhook URL |

### notify_codes の値

| コード | 種別 |
|--------|------|
| `551` | 地震情報 |
| `552` | 津波情報 |
| `556` | 緊急地震速報（警報） |

### min_scale の値

| 値 | 対応震度 |
|----|---------|
| `0` | すべて |
| `10` | 震度1以上 |
| `20` | 震度2以上 |
| `30` | 震度3以上 |
| `40` | 震度4以上 |
| `45` | 震度4.5以上 |
| `50` | 震度5弱以上 |

## データソース

- [P2P地震情報 JSON API v2 ドキュメント](https://www.p2pquake.net/develop/json_api_v2/)
