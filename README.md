# gin-tutorial

Gin フレームワークの機能を学ぶためのチュートリアルリポジトリ。

## 利用技術

| 技術 | バージョン |
|------|-----------|
| Go | 1.26.1 |
| [gin-gonic/gin](https://github.com/gin-gonic/gin) | 1.12.0 |
| [gorm.io/gorm](https://gorm.io/) | 1.31.1 |
| [gorm.io/driver/mysql](https://github.com/go-gorm/mysql) | 1.6.0 |
| [google.golang.org/grpc](https://grpc.io/) | 1.79.3 |
| [redis/go-redis](https://github.com/redis/go-redis) | 9.x |
| MySQL | 8.0 (Docker) |
| Redis | 7 (Docker) |
| Docker | マルチステージビルド（scratch ベース） |

## アプリケーションの実行方法

### ローカル実行

#### 前提条件

- Go 1.26.1 以上がインストールされていること
- MySQL 8.0 が起動していること（または後述の Docker 経由で起動する）
- Redis 7 が起動していること（または後述の Docker 経由で起動する）

#### 環境変数（DB・Redis接続設定）

| 変数名 | デフォルト | 説明 |
|-------|-----------|------|
| `DB_HOST` | `localhost` | MySQLホスト |
| `DB_PORT` | `3306` | MySQLポート |
| `DB_USER` | `root` | DBユーザー |
| `DB_PASSWORD` | `root` | DBパスワード |
| `DB_NAME` | `gin_tutorial` | DB名 |
| `DB_SEED` | `false` | `true` で初期データを投入 |
| `REDIS_HOST` | `localhost` | Redisホスト |
| `REDIS_PORT` | `6379` | Redisポート |

#### 起動

```bash
# .env を作成して設定を編集
cp .env.example .env

DB_SEED=true go run main.go
```

### Docker で実行

#### 前提条件

- Docker がインストールされていること

#### 起動（MySQL + Redis + アプリを同時起動）

```bash
docker compose up -d
```

MySQL・Redis のヘルスチェックが完了してからアプリが起動します。初回起動時は自動でマイグレーションとシードが実行されます。

#### 停止

```bash
docker compose down
```

#### データを含めてリセット

```bash
docker compose down -v
```

サーバーは `http://localhost:8080` で起動します。

## テスト

### 全テスト実行

```bash
go test ./tests/...
```

### ユニットテスト（カバレッジ計測付き）

```bash
go test ./tests/ut/... -cover -coverpkg=./app/...
```

### インテグレーションテスト

```bash
go test ./tests/it/...
```

### テスト構成

| パッケージ | 内容 |
|-----------|------|
| `tests/ut` | 各ハンドラー・ミドルウェアの単体テスト。AAA パターン適用 |
| `tests/it` | `router.New()` を使った全エンドポイントの正常系 E2E テスト |

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
- [API 仕様書](docs/api.md)
- [gRPC 動作確認ガイド](docs/grpc.md)
- [テスト設計資料](docs/testing.md)
