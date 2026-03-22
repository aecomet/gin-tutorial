# API 仕様書

全エンドポイントの仕様を Swagger 準拠の表形式でまとめます。

## 共通仕様

### ベース URL

```
http://localhost:8080/api
```

### レスポンス形式

```json
{
  "success": true,
  "data": { ... },
  "error": { "code": "NOT_FOUND", "message": "resource not found" },
  "meta": { "page": 1, "per_page": 20, "total": 100, "total_pages": 5 }
}
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `success` | boolean | リクエスト成否 |
| `data` | any | 成功時のレスポンスボディ（省略可） |
| `error` | object | エラー時の詳細（省略可） |
| `meta` | object | ページネーション情報（省略可） |

### 共通エラーコード

| HTTP Status | code | 説明 |
|-------------|------|------|
| 400 | `BAD_REQUEST` | リクエストパラメータが不正 |
| 401 | `UNAUTHORIZED` | 認証が必要 |
| 404 | `NOT_FOUND` | リソースが存在しない |
| 500 | `INTERNAL_ERROR` | サーバー内部エラー |

---

## システムエンドポイント

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/healthcheck` | ヘルスチェック（`Accept-Version` ヘッダーバージョニングのデモ） |
| GET | `/api/routes` | 登録済みルート一覧 |

### GET /api/healthcheck

**レスポンス例 (200)**
```json
{ "version": "v1", "status": "ok" }
```

**ヘッダー**

| Name | 必須 | 説明 |
|------|------|------|
| `Accept-Version` | No | `v1` または `v2`（省略時は `v1`） |

---

## v1 — Gin 基本機能デモ

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/v1/welcome` | クエリパラメータ取得 |
| POST | `/api/v1/form_post` | POSTフォームデータ |
| POST | `/api/v1/post` | クエリパラメータ + POSTフォームデータ |
| POST | `/api/v1/form_map` | QueryMap + PostFormMap |
| POST | `/api/v1/multipart` | multipart/form-data（ファイルアップロード） |
| GET | `/api/v1/articles` | ページネーション（limit / offset） |
| GET | `/api/v1/events` | ページネーション（カーソルベース） |

---

## v2 — リソースベース CRUD デモ

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/v2/users` | ユーザー一覧 |
| POST | `/api/v2/users` | ユーザー作成 |
| GET | `/api/v2/users/:id` | ユーザー取得 |
| PUT | `/api/v2/users/:id` | ユーザー更新 |
| DELETE | `/api/v2/users/:id` | ユーザー削除 |
| GET | `/api/v2/products` | 商品一覧（フィルタリング・ソートデモ） |
| GET | `/api/v2/orders` | オーダー一覧 |
| POST | `/api/v2/orders` | オーダー作成 |
| GET | `/api/v2/orders/:id` | オーダー取得 |
| GET | `/api/v2/items/:id` | アイテム取得（カスタムエラーハンドリングデモ） |

### GET /api/v2/products — クエリパラメータ

| Name | 型 | デフォルト | 説明 |
|------|----|-----------|------|
| `category` | string | — | カテゴリフィルター |
| `min_price` | string | — | 最低価格 |
| `max_price` | string | — | 最高価格 |
| `sort` | string | `created_at` | ソートキー（`created_at` / `price` / `name`） |
| `order` | string | `desc` | ソート順（`asc` / `desc`） |

---

## v3 — モデルバインディング・バリデーションデモ

| Method | Path | 説明 |
|--------|------|------|
| POST | `/api/v3/users` | JSONバインディング（required / email / min / max / gte） |
| GET | `/api/v3/users/:id` | URIバインディング（gt=0） |
| GET | `/api/v3/search` | クエリバインディング（omitempty / gte / lte） |
| POST | `/api/v3/login` | フォームバインディング（required / min） |
| GET | `/api/v3/posts` | デフォルト値付きクエリバインディング（default / oneof） |
| GET | `/api/v3/me` | ヘッダーバインディング（required / uuid4） |

---

## v4 — Basic認証・goroutine 非同期処理デモ

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/v4/profile` | Basic認証（`gin.BasicAuth` ミドルウェア） |
| GET | `/api/v4/secret` | Basic認証保護リソース |
| GET | `/api/v4/async` | goroutine による並列タスク実行 |

---

## v5 — GORM + MySQL CRUD（articles リソース）

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/v5/articles` | 記事一覧（ページネーション対応） |
| POST | `/api/v5/articles` | 記事作成 |
| GET | `/api/v5/articles/:id` | 記事取得 |
| PUT | `/api/v5/articles/:id` | 記事更新（部分更新対応） |
| DELETE | `/api/v5/articles/:id` | 記事削除（論理削除） |

### Article モデル

| フィールド | 型 | 制約 | 説明 |
|-----------|-----|------|------|
| `id` | uint | PK, auto | 記事ID |
| `title` | string | NOT NULL, max=255 | タイトル |
| `body` | text | — | 本文 |
| `author` | string | max=100 | 著者名 |
| `created_at` | datetime | auto | 作成日時 |
| `updated_at` | datetime | auto | 更新日時 |
| `deleted_at` | datetime | — | 論理削除日時 |

### GET /api/v5/articles — クエリパラメータ

| Name | 型 | デフォルト | 制約 | 説明 |
|------|----|-----------|------|------|
| `page` | int | `1` | ≥1 | ページ番号 |
| `per_page` | int | `10` | 1〜100 | 1ページあたりの件数 |

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": [
    { "id": 1, "title": "タイトル", "body": "本文", "author": "Alice", "created_at": "...", "updated_at": "..." }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 1, "total_pages": 1 }
}
```

### POST /api/v5/articles — リクエストボディ

```json
{ "title": "タイトル（必須）", "body": "本文", "author": "著者名" }
```

| フィールド | 型 | 必須 | 制約 | 説明 |
|-----------|-----|------|------|------|
| `title` | string | ✓ | max=255 | タイトル |
| `body` | string | — | — | 本文 |
| `author` | string | — | max=100 | 著者名 |

**レスポンス (201)**
```json
{ "success": true, "data": { "id": 1, "title": "...", ... } }
```

### PUT /api/v5/articles/:id — リクエストボディ

指定フィールドのみ更新（省略したフィールドは変更なし）

```json
{ "title": "新しいタイトル", "body": "新しい本文", "author": "新しい著者" }
```

| フィールド | 型 | 必須 | 制約 | 説明 |
|-----------|-----|------|------|------|
| `title` | string | — | max=255 | タイトル |
| `body` | string | — | — | 本文 |
| `author` | string | — | max=100 | 著者名 |

### DELETE /api/v5/articles/:id

論理削除（`deleted_at` に削除日時を設定）。成功時は **204 No Content** を返す。
