# gRPC 動作確認ガイド

このリポジトリの v6 エンドポイントは、Gin が gRPC クライアントとして内部の gRPC サーバーを呼び出す **API Gateway パターン**のデモです。

## アーキテクチャ

```
REST クライアント（curl など）
        ↓ HTTP :8080
  GET /api/v6/articles
        ↓ gRPC (localhost:50051)
  ArticleService.ListArticles()
        ↓ Redis
  grpc:article:* キー
```

gRPC サーバーは **:50051** で独立して動作しており、Gin（:8080）とは別のポートです。  
REST 経由と gRPC 直接呼び出しの両方で動作確認できます。

---

## 事前準備

### サーバー起動

```bash
# Docker で起動（MySQL + Redis + アプリ）
docker compose up -d

# またはローカル起動（.env を用意した上で）
cp .env.example .env
go run main.go
```

### grpcurl のインストール

```bash
# macOS
brew install grpcurl

# Go
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

---

## REST 経由での動作確認（curl）

v6 エンドポイントを REST で呼ぶと、内部で gRPC サーバーに転送されます。

### 記事一覧の取得

```bash
curl http://localhost:8080/api/v6/articles
```

```bash
# ページネーション指定
curl "http://localhost:8080/api/v6/articles?page=1&per_page=5"
```

### 記事の作成

```bash
curl -X POST http://localhost:8080/api/v6/articles \
  -H "Content-Type: application/json" \
  -d '{"title":"gRPCとは","body":"HTTP/2上で動作するRPCフレームワーク","author":"Alice"}'
```

### 記事の取得

```bash
curl http://localhost:8080/api/v6/articles/1
```

### 記事の更新

```bash
curl -X PUT http://localhost:8080/api/v6/articles/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"更新後のタイトル"}'
```

### 記事の削除

```bash
curl -X DELETE http://localhost:8080/api/v6/articles/1
```

---

## gRPC 直接呼び出し（grpcurl）

grpcurl を使うと Gin を経由せずに gRPC サーバーを直接呼び出せます。

### サービス・メソッド一覧の確認

```bash
# サービス一覧
grpcurl -plaintext localhost:50051 list
# => article.ArticleService

# メソッド一覧
grpcurl -plaintext localhost:50051 list article.ArticleService
# => article.ArticleService.CreateArticle
# => article.ArticleService.DeleteArticle
# => ...

# メソッドの定義（引数・戻り値の型）を確認
grpcurl -plaintext localhost:50051 describe article.ArticleService.ListArticles
```

### ListArticles（記事一覧）

```bash
grpcurl -plaintext \
  -d '{"page": 1, "per_page": 10}' \
  localhost:50051 article.ArticleService/ListArticles
```

**レスポンス例:**
```json
{
  "articles": [
    {
      "id": 1,
      "title": "gRPCとGinで作るAPI",
      "body": "gRPCはHTTP/2上で動作する高速なRPCフレームワークです。",
      "author": "Alice"
    }
  ],
  "total": 1
}
```

### GetArticle（記事取得）

```bash
grpcurl -plaintext \
  -d '{"id": 1}' \
  localhost:50051 article.ArticleService/GetArticle
```

**存在しないIDの場合（codes.NotFound）:**
```bash
grpcurl -plaintext \
  -d '{"id": 999}' \
  localhost:50051 article.ArticleService/GetArticle
# ERROR:
#   Code: NotFound
#   Message: article not found: id=999
```

### CreateArticle（記事作成）

```bash
grpcurl -plaintext \
  -d '{"title": "Protocol Buffersとは", "body": "protoファイルでスキーマを定義します", "author": "Bob"}' \
  localhost:50051 article.ArticleService/CreateArticle
```

### UpdateArticle（記事更新）

```bash
# title のみ更新（空文字フィールドは更新しない）
grpcurl -plaintext \
  -d '{"id": 1, "title": "新しいタイトル"}' \
  localhost:50051 article.ArticleService/UpdateArticle
```

### DeleteArticle（記事削除）

```bash
grpcurl -plaintext \
  -d '{"id": 1}' \
  localhost:50051 article.ArticleService/DeleteArticle
```

---

## proto ファイルからコードを再生成する

`.proto` ファイルを変更した場合は以下のコマンドで Go コードを再生成してください。

```bash
# 必要なツールのインストール（初回のみ）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# コード生成
protoc \
  --go_out=app/grpc/pb --go_opt=paths=source_relative \
  --go-grpc_out=app/grpc/pb --go-grpc_opt=paths=source_relative \
  -I app/grpc/proto \
  app/grpc/proto/article.proto
```

> **注意:** `app/grpc/pb/` 配下のファイルは自動生成コードです。手動で編集しないでください。

---

## gRPC エラーコードと HTTP ステータスの対応

v6 ハンドラーは gRPC エラーコードを HTTP ステータスに変換します。

| gRPC コード | HTTP ステータス | 説明 |
|------------|----------------|------|
| `OK` | 200 | 成功 |
| `NOT_FOUND` | 404 | リソースが存在しない |
| `INVALID_ARGUMENT` | 400 | リクエストパラメータ不正 |
| `ALREADY_EXISTS` | 409 | リソースが既に存在する |
| その他 | 500 | サーバー内部エラー |

---

## Redis でデータを確認する

gRPC サーバーのデータは Redis に保存されます。`redis-cli` で直接確認できます。

```bash
# Docker 環境
docker compose exec redis redis-cli

# ローカル環境
redis-cli
```

```bash
# 全記事IDを確認
SMEMBERS grpc:article:ids

# 特定記事のデータを確認
GET grpc:article:1

# 現在のIDカウンターを確認
GET grpc:article:seq

# 全データをリセット
DEL grpc:article:ids grpc:article:seq
KEYS grpc:article:* | xargs DEL
```
