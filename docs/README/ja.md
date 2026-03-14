# unispeedtest (日本語)

[English](../../README.md)

`unispeedtest` は、回線品質を計測する CLI ツールです。現在は Cloudflare のスピードテスト用エンドポイントを利用していますが、将来的に他プロバイダへ拡張しやすい構成を想定しています。

## 主な機能

- ダウンロード速度（サンプル Mbps の 90 パーセンタイル）
- アップロード速度（サンプル Mbps の 90 パーセンタイル）
- 無負荷レイテンシ（20 サンプルの中央値）
- ダウンロード/アップロード負荷時レイテンシ
- ジッター（無負荷レイテンシ連続差分の絶対値平均）
- パケットロス（1000 リクエスト、同時実行 50）
- ネットワーク情報（Cloudflare colo、ASN/AS 組織名、グローバル IP）

## インストール

### 1) GitHub Releases からインストール（推奨）

[Releases](https://github.com/hsblabs/universal-speedtest-cli/releases) から最新版アーカイブを取得し、`unispeedtest` を `PATH` の通った場所に配置してください。

またはインストーラスクリプトを利用できます。

```sh
curl -fsSL https://raw.githubusercontent.com/hsblabs/universal-speedtest-cli/main/install.sh | sh
```

デフォルトのインストール先は `/usr/local/bin` です（`INSTALL_DIR` で変更可能）。
インストーラスクリプトは release の `checksums.txt` を使って SHA-256 を検証します。

### 2) Go でインストール

```sh
go install github.com/hsblabs/universal-speedtest-cli/cmd/unispeedtest@latest
```

この方法ではバイナリ名は `unispeedtest` になります。

## 使い方

```sh
unispeedtest
```

オプション:

- `-json`: JSON（1行）で出力
- `-pretty`: 整形済み JSON で出力（`-json` を含む）

例:

```sh
unispeedtest -json
unispeedtest -pretty
```

## JSON 出力形式

一部の測定だけ失敗した場合、該当メトリクスは `null` で出力され、`warnings` 配列に理由が入ります。これにより、本当の `0` と欠損値を区別できます。

```json
{
  "download_mbps": 225.14,
  "upload_mbps": 102.87,
  "latency_ms": {
    "unloaded": 12.41,
    "loaded_down": 35.09,
    "loaded_up": 41.22,
    "jitter": 1.98
  },
  "packet_loss_percent": 0.1,
  "server_colo": "Tokyo",
  "network_asn": "AS2516",
  "network_as_org": "KDDI CORPORATION",
  "ip": "203.0.113.10",
  "warnings": [
    "upload loaded latency unavailable: no samples collected"
  ]
}
```

## 開発

テスト:

```sh
go test ./...
```

ビルド:

```sh
go build -trimpath -ldflags="-s -w" -o dist/unispeedtest ./cmd/unispeedtest
```

補足:

- `NO_COLOR=1` を設定すると ANSI カラー出力を無効化できます。

## ライセンス

MIT
