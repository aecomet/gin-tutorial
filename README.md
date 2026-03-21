# gin-tutorial

Gin フレームワークの機能を学ぶためのチュートリアルリポジトリ。

## 利用技術

| 技術 | バージョン |
|------|-----------|
| Go | 1.26.1 |
| [gin-gonic/gin](https://github.com/gin-gonic/gin) | 1.12.0 |
| Docker | マルチステージビルド（scratch ベース） |

## アプリケーションの実行方法

### ローカル実行

#### 前提条件

- Go 1.26.1 以上がインストールされていること

#### 起動

```bash
go run main.go
```

### Docker で実行

#### 前提条件

- Docker がインストールされていること

#### 起動

```bash
docker compose up -d
```

#### 停止

```bash
docker compose down
```

サーバーは `http://localhost:8080` で起動します。

### 動作確認

```bash
# ルート一覧を確認
curl http://localhost:8080/api/routes

# ヘルスチェック
curl http://localhost:8080/api/healthcheck

# ヘッダーバージョニング（v2）
curl -H "Accept-Version: v2" http://localhost:8080/api/healthcheck
```

## ドキュメント

- [アーキテクチャ設計資料](docs/architecture.md)
