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
| 400 | `VALIDATION_ERROR` | バリデーション失敗 |
| 401 | `UNAUTHORIZED` | 認証が必要 |
| 404 | `NOT_FOUND` | リソースが存在しない |
| 500 | `INTERNAL_ERROR` | サーバー内部エラー |

---

## システムエンドポイント

### GET /api/healthcheck

ヘルスチェック。`Accept-Version` ヘッダーによるバージョニングのデモ。

**リクエストヘッダー**

| Name | 型 | 必須 | 値 | 説明 |
|------|----|------|----|------|
| `Accept-Version` | string | No | `v1` / `v2` | 省略時は `v1` として動作 |

**リクエスト例 — v1**
```
GET /api/healthcheck
```

**リクエスト例 — v2**
```
GET /api/healthcheck
Accept-Version: v2
```

**レスポンス例 (200) — v1**
```json
{ "status": "ok" }
```

**レスポンス例 (200) — v2**
```json
{ "status": "ok", "version": "v2" }
```

---

### GET /api/routes

登録済みルート一覧を返す。パラメータなし。

**リクエスト例**
```
GET /api/routes
```

---

## v1 — Gin 基本機能デモ

### GET /api/v1/welcome

クエリパラメータの基本取得デモ。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 説明 |
|------|----|------|-----------|------|
| `firstname` | string | No | `Guest` | 名 |
| `lastname` | string | No | `""` (空文字) | 姓 |

**リクエスト例**
```
GET /api/v1/welcome?firstname=John&lastname=Doe
```

**レスポンス例 (200)**
```
Hello John Doe
```

---

### POST /api/v1/form_post

`application/x-www-form-urlencoded` のデモ。

**リクエストボディ** (`Content-Type: application/x-www-form-urlencoded`)

| Name | 型 | 必須 | デフォルト | 説明 |
|------|----|------|-----------|------|
| `message` | string | No | `""` | メッセージ本文 |
| `nick` | string | No | `anonymous` | ニックネーム |

**リクエスト例**
```
message=hello&nick=Alice
```

**レスポンス例 (200)**
```json
{ "status": "posted", "message": "hello", "nick": "Alice" }
```

---

### POST /api/v1/post

クエリパラメータ + フォームデータの組み合わせデモ。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 説明 |
|------|----|------|-----------|------|
| `id` | string | No | `""` | リソースID |
| `page` | string | No | `0` | ページ番号 |

**リクエストボディ** (`Content-Type: application/x-www-form-urlencoded`)

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `name` | string | No | 名前 |
| `message` | string | No | メッセージ |

**リクエスト例**
```
POST /api/v1/post?id=42&page=3
Body: name=Bob&message=hello
```

**レスポンス例 (200)**
```json
{ "id": "42", "page": "3", "name": "Bob", "message": "hello" }
```

---

### POST /api/v1/form_map

QueryMap と PostFormMap のデモ。

**クエリパラメータ（マップ形式）**

| Name | 型 | 説明 |
|------|----|------|
| `ids[<key>]` | string | 任意のキーで複数値を渡す |

**リクエストボディ** (`Content-Type: application/x-www-form-urlencoded`、マップ形式)

| Name | 型 | 説明 |
|------|----|------|
| `names[<key>]` | string | 任意のキーで複数値を渡す |

**リクエスト例**
```
POST /api/v1/form_map?ids[a]=1&ids[b]=2
Body: names[first]=Alice&names[last]=Smith
```

**レスポンス例 (200)**
```json
{
  "ids":   { "a": "1", "b": "2" },
  "names": { "first": "Alice", "last": "Smith" }
}
```

---

### POST /api/v1/multipart

`multipart/form-data` によるファイルアップロードのデモ。

**リクエストボディ** (`Content-Type: multipart/form-data`)

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `file` | file | **Yes** | アップロードするファイル |
| `message` | string | No | 付随メッセージ |

**リクエスト例**
```
POST /api/v1/multipart
Content-Type: multipart/form-data; boundary=----FormBoundary

------FormBoundary
Content-Disposition: form-data; name="message"

hello
------FormBoundary
Content-Disposition: form-data; name="file"; filename="test.txt"
Content-Type: text/plain

(file content)
------FormBoundary--
```

**レスポンス例 (200)**
```json
{ "message": "hello", "filename": "test.txt", "size": 1024 }
```

**エラーレスポンス (400)** — `file` フィールドがない場合
```json
{ "error": "http: no such file" }
```

---

### GET /api/v1/articles

リミット・オフセットによるページネーションデモ。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 制約 | 説明 |
|------|----|------|-----------|------|------|
| `limit` | integer | No | `20` | 1〜100（超過時は `100` に切り捨て） | 取得件数 |
| `offset` | integer | No | `0` | ≥0 | 読み飛ばし件数 |

**リクエスト例**
```
GET /api/v1/articles?limit=50&offset=10
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": [],
  "meta": { "limit": 50, "offset": 10, "total": 0 }
}
```

---

### GET /api/v1/events

カーソルベースのページネーションデモ。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 制約 | 説明 |
|------|----|------|-----------|------|------|
| `cursor` | string | No | `""` | — | 前回レスポンスの `next_cursor` 値 |
| `limit` | integer | No | `20` | 1〜100（超過時は `100` に切り捨て） | 取得件数 |

**リクエスト例**
```
GET /api/v1/events?cursor=eyJpZCI6MTAwfQ&limit=10
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": [],
  "next_cursor": ""
}
```

---

## v2 — リソースベース CRUD デモ

### GET /api/v2/users

ユーザー一覧を返す。パラメータなし（スタブ実装）。

**リクエスト例**
```
GET /api/v2/users
```

**レスポンス例 (200)**
```json
{ "action": "list_users" }
```

---

### POST /api/v2/users

ユーザーを作成する（スタブ実装）。

**リクエスト例**
```json
{ "name": "Alice", "email": "alice@example.com" }
```

**レスポンス例 (201)**
```json
{ "action": "create_user" }
```

---

### GET /api/v2/users/:id

ユーザーを取得する（スタブ実装）。

**パスパラメータ**

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `id` | string | **Yes** | ユーザーID |

**リクエスト例**
```
GET /api/v2/users/1
```

**レスポンス例 (200)**
```json
{ "action": "get_user" }
```

---

### PUT /api/v2/users/:id

ユーザーを更新する（スタブ実装）。

**パスパラメータ**

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `id` | string | **Yes** | ユーザーID |

**リクエスト例**
```json
{ "name": "Alice Updated" }
```

**レスポンス例 (200)**
```json
{ "action": "update_user" }
```

---

### DELETE /api/v2/users/:id

ユーザーを削除する（スタブ実装）。成功時は **204 No Content**。

**パスパラメータ**

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `id` | string | **Yes** | ユーザーID |

**リクエスト例**
```
DELETE /api/v2/users/1
```

**レスポンス (204)** — ボディなし

---

### GET /api/v2/products

商品一覧。フィルタリング・ソートのデモ。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 制約 | 説明 |
|------|----|------|-----------|------|------|
| `category` | string | No | `""` | — | カテゴリフィルター |
| `min_price` | string | No | `""` | — | 最低価格 |
| `max_price` | string | No | `""` | — | 最高価格 |
| `sort` | string | No | `created_at` | `created_at` / `price` / `name`（許可外は `created_at` にフォールバック） | ソートキー |
| `order` | string | No | `desc` | `asc` / `desc`（許可外は `desc` にフォールバック） | ソート順 |

**リクエスト例**
```
GET /api/v2/products?category=electronics&min_price=100&sort=price&order=asc
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": [],
  "filters": {
    "category":  "electronics",
    "min_price": "100",
    "max_price": "",
    "sort":      "price",
    "order":     "asc"
  }
}
```

---

### GET /api/v2/orders

オーダー一覧（スタブ実装）。

**リクエスト例**
```
GET /api/v2/orders
```

**レスポンス例 (200)**
```json
{ "action": "list_orders" }
```

---

### POST /api/v2/orders

オーダー作成（スタブ実装）。

**リクエスト例**
```json
{ "product_id": 1, "quantity": 2 }
```

**レスポンス例 (201)**
```json
{ "action": "create_order" }
```

---

### GET /api/v2/orders/:id

オーダー取得（スタブ実装）。

**パスパラメータ**

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `id` | string | **Yes** | オーダーID |

**リクエスト例**
```
GET /api/v2/orders/1
```

**レスポンス例 (200)**
```json
{ "action": "get_order" }
```

---

### GET /api/v2/items/:id

カスタムエラーハンドリングのデモ。`id=0` は `NOT_FOUND` を返す。

**パスパラメータ**

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `id` | string | **Yes** | アイテムID（`0` を指定すると 404） |

**レスポンス例 (200)** — `id` ≥ 1
```json
{ "success": true, "data": { "id": "1" } }
```

**エラーレスポンス (404)** — `id=0`
```json
{ "success": false, "error": { "code": "NOT_FOUND", "message": "resource not found" } }
```

---

## v3 — モデルバインディング・バリデーションデモ

### POST /api/v3/users

JSON バインディング + バリデーションデモ。

**リクエストボディ** (`Content-Type: application/json`)

| フィールド | 型 | 必須 | 制約 | 説明 |
|-----------|-----|------|------|------|
| `name` | string | **Yes** | min=2, max=50 | 名前 |
| `email` | string | **Yes** | 有効なメールアドレス形式 | メールアドレス |
| `age` | integer | **Yes** | 1〜130 | 年齢 |
| `password` | string | **Yes** | min=8 | パスワード（レスポンスには含まれない） |

**リクエスト例**
```json
{
  "name":     "Alice",
  "email":    "alice@example.com",
  "age":      25,
  "password": "secret123"
}
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": { "name": "Alice", "email": "alice@example.com", "age": 25 }
}
```

**エラーレスポンス (400)**
```json
{
  "success": false,
  "error": { "code": "VALIDATION_ERROR", "message": "Key: 'CreateUserRequest.Name' Error:..." }
}
```

---

### GET /api/v3/users/:id

URI バインディング + バリデーションデモ。

**パスパラメータ**

| Name | 型 | 必須 | 制約 | 説明 |
|------|----|------|------|------|
| `id` | integer | **Yes** | gt=0（0以下はバリデーションエラー） | ユーザーID |

**リクエスト例**
```
GET /api/v3/users/5
```

**レスポンス例 (200)**
```json
{ "success": true, "data": { "id": 5 } }
```

**エラーレスポンス (400)** — `id=0`
```json
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "..." } }
```

---

### GET /api/v3/search

クエリパラメータ バインディング + バリデーションデモ。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 制約 | 説明 |
|------|----|------|-----------|------|------|
| `keyword` | string | **Yes** | — | — | 検索キーワード |
| `page` | integer | No | `1` | ≥1 | ページ番号 |
| `per_page` | integer | No | `20` | 1〜100 | 1ページあたりの件数 |

**リクエスト例**
```
GET /api/v3/search?keyword=gin&page=2&per_page=5
```

**レスポンス例 (200)**
```json
{ "success": true, "data": { "keyword": "gin", "page": 2, "per_page": 5 } }
```

**エラーレスポンス (400)** — `keyword` 未指定
```json
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "..." } }
```

---

### POST /api/v3/login

フォームバインディング + バリデーションデモ。

**リクエストボディ** (`Content-Type: application/x-www-form-urlencoded`)

| Name | 型 | 必須 | 制約 | 説明 |
|------|----|------|------|------|
| `username` | string | **Yes** | — | ユーザー名 |
| `password` | string | **Yes** | min=8 | パスワード |

**リクエスト例**
```
username=alice&password=secret123
```

**レスポンス例 (200)**
```json
{ "success": true, "data": { "username": "alice", "message": "login successful" } }
```

---

### GET /api/v3/posts

デフォルト値付きクエリバインディングデモ。全パラメータ省略可。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 制約 | 説明 |
|------|----|------|-----------|------|------|
| `page` | integer | No | `1` | ≥1 | ページ番号 |
| `per_page` | integer | No | `20` | 1〜100 | 1ページあたりの件数 |
| `sort` | string | No | `created_at` | `created_at` / `updated_at` / `title` | ソートキー |
| `order` | string | No | `desc` | `asc` / `desc` | ソート順 |
| `status` | string | No | `published` | `draft` / `published` / `archived` | 投稿ステータス |

**リクエスト例**
```
GET /api/v3/posts?page=2&per_page=10&sort=title&order=asc&status=draft
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": {
    "data": [],
    "meta": { "page": 2, "per_page": 10, "sort": "title", "order": "asc", "status": "draft" }
  }
}
```

---

### GET /api/v3/me

ヘッダーバインディング + バリデーションデモ。

**リクエストヘッダー**

| Name | 型 | 必須 | 制約 | 説明 |
|------|----|------|------|------|
| `Authorization` | string | **Yes** | — | 認証トークン（例: `Bearer <token>`） |
| `X-Request-Id` | string | **Yes** | UUID v4 形式 | リクエスト追跡ID |
| `Accept-Language` | string | No | — | 言語設定（例: `ja-JP`） |

**リクエスト例**
```
GET /api/v3/me
Authorization: Bearer eyJhbGc...
X-Request-Id: 550e8400-e29b-41d4-a716-446655440000
Accept-Language: ja-JP
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": {
    "authorization": "Bearer eyJhbGc...",
    "request_id":    "550e8400-e29b-41d4-a716-446655440000",
    "accept_lang":   "ja-JP"
  }
}
```

**エラーレスポンス (400)** — `X-Request-Id` が UUID 形式でない場合
```json
{ "success": false, "error": { "code": "VALIDATION_ERROR", "message": "..." } }
```

---

## v4 — Basic 認証・goroutine 非同期処理デモ

### 認証情報

| ユーザー名 | パスワード |
|-----------|-----------|
| `admin` | `secret` |
| `user` | `password` |

Basic 認証ヘッダーの形式: `Authorization: Basic <base64(username:password)>`

---

### GET /api/v4/profile

Basic 認証デモ。認証ユーザー情報を返す。

**リクエストヘッダー**

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `Authorization` | string | **Yes** | `Basic <base64(user:pass)>` |

**リクエスト例**
```
GET /api/v4/profile
Authorization: Basic YWRtaW46c2VjcmV0
```
> `YWRtaW46c2VjcmV0` は `admin:secret` を Base64 エンコードした値

**レスポンス例 (200)**
```json
{ "user": "admin", "message": "authenticated successfully" }
```

**エラーレスポンス (401)** — 認証ヘッダーなし／認証失敗
```
HTTP 401 Unauthorized
```

---

### GET /api/v4/secret

Basic 認証で保護された機密リソース。

**リクエストヘッダー**

| Name | 型 | 必須 | 説明 |
|------|----|------|------|
| `Authorization` | string | **Yes** | `Basic <base64(user:pass)>` |

**リクエスト例**
```
GET /api/v4/secret
Authorization: Basic dXNlcjpwYXNzd29yZA==
```
> `dXNlcjpwYXNzd29yZA==` は `user:password` を Base64 エンコードした値

**レスポンス例 (200)**
```json
{ "user": "user", "secret": "this is confidential" }
```

---

### GET /api/v4/async

goroutine による並列タスク実行デモ。パラメータなし。

内部で 3 つのタスク（task-a / task-b / task-c）を goroutine で並列実行し、全完了後に結果を返す。

**リクエスト例**
```
GET /api/v4/async
```

**レスポンス例 (200)**
```json
{
  "tasks": [
    { "task": "task-a", "result": "task-a completed", "duration": "50.123ms" },
    { "task": "task-b", "result": "task-b completed", "duration": "80.456ms" },
    { "task": "task-c", "result": "task-c completed", "duration": "30.789ms" }
  ]
}
```

---

## v5 — GORM + MySQL CRUD（articles リソース）

### Article モデル

| フィールド | 型 | DB 制約 | JSON キー | 説明 |
|-----------|-----|---------|-----------|------|
| `id` | uint | PK, AUTO_INCREMENT | `id` | 記事ID |
| `title` | string | NOT NULL, size=255 | `title` | タイトル |
| `body` | text | — | `body` | 本文 |
| `author` | string | size=100 | `author` | 著者名 |
| `created_at` | datetime | auto | `created_at` | 作成日時（ISO 8601） |
| `updated_at` | datetime | auto | `updated_at` | 更新日時（ISO 8601） |
| `deleted_at` | datetime | index, nullable | (omitted) | 論理削除日時（レスポンスには含まれない） |

---

### GET /api/v5/articles

記事一覧を返す（ページネーション対応）。

**クエリパラメータ**

| Name | 型 | 必須 | デフォルト | 制約 | 説明 |
|------|----|------|-----------|------|------|
| `page` | integer | No | `1` | ≥1（下限未満は `1` に補正） | ページ番号 |
| `per_page` | integer | No | `10` | 1〜100（範囲外は `10` に補正） | 1ページあたりの件数 |

**リクエスト例**
```
GET /api/v5/articles?page=2&per_page=5
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "GORMとGinで作るREST API",
      "body": "GORMはGoで最も使われているORMライブラリです。",
      "author": "Alice",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 2, "per_page": 5, "total": 3, "total_pages": 1 }
}
```

---

### POST /api/v5/articles

新規記事を作成する。

**リクエストボディ** (`Content-Type: application/json`)

| フィールド | 型 | 必須 | 制約 | 説明 |
|-----------|-----|------|------|------|
| `title` | string | **Yes** | max=255 | タイトル |
| `body` | string | No | — | 本文 |
| `author` | string | No | max=100 | 著者名 |

**リクエスト例**
```json
{
  "title":  "Dockerでローカル開発環境を構築する",
  "body":   "Docker Composeを使うことで開発環境を簡単に再現できます。",
  "author": "Bob"
}
```

**レスポンス例 (201)**
```json
{
  "success": true,
  "data": {
    "id": 4,
    "title":  "Dockerでローカル開発環境を構築する",
    "body":   "Docker Composeを使うことで開発環境を簡単に再現できます。",
    "author": "Bob",
    "created_at": "2024-06-01T12:00:00Z",
    "updated_at": "2024-06-01T12:00:00Z"
  }
}
```

**エラーレスポンス (400)** — `title` が未指定
```json
{ "success": false, "error": { "code": "BAD_REQUEST", "message": "Key: 'CreateArticleInput.Title' Error:..." } }
```

---

### GET /api/v5/articles/:id

指定 ID の記事を返す。

**パスパラメータ**

| Name | 型 | 必須 | 制約 | 説明 |
|------|----|------|------|------|
| `id` | integer | **Yes** | 正の整数（非整数は 400） | 記事ID |

**リクエスト例**
```
GET /api/v5/articles/1
```

**レスポンス例 (200)**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title":  "GORMとGinで作るREST API",
    "body":   "GORMはGoで最も使われているORMライブラリです。",
    "author": "Alice",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**エラーレスポンス (400)** — `id` が整数でない
```json
{ "success": false, "error": { "code": "BAD_REQUEST", "message": "id must be a positive integer" } }
```

**エラーレスポンス (404)** — 存在しない ID
```json
{ "success": false, "error": { "code": "NOT_FOUND", "message": "article not found" } }
```

---

### PUT /api/v5/articles/:id

指定 ID の記事を部分更新する。指定したフィールドのみ更新し、省略フィールドは変更しない。

**パスパラメータ**

| Name | 型 | 必須 | 制約 | 説明 |
|------|----|------|------|------|
| `id` | integer | **Yes** | 正の整数 | 記事ID |

**リクエストボディ** (`Content-Type: application/json`)

| フィールド | 型 | 必須 | 制約 | 説明 |
|-----------|-----|------|------|------|
| `title` | string | No | min=1, max=255 | タイトル（空文字列は不可） |
| `body` | string | No | — | 本文 |
| `author` | string | No | min=1, max=100 | 著者名（空文字列は不可） |

> すべてのフィールドがポインタ型（`*string`）のため、省略（`null`/未指定）したフィールドは DB 更新対象外になる。

**リクエスト例** — タイトルのみ更新
```json
{ "title": "更新後のタイトル" }
```

**レスポンス例 (200)** — 更新後の最新データを返す
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title":  "更新後のタイトル",
    "body":   "GORMはGoで最も使われているORMライブラリです。",
    "author": "Alice",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-06-01T15:30:00Z"
  }
}
```

---

### DELETE /api/v5/articles/:id

指定 ID の記事を論理削除する（`deleted_at` に削除日時を設定）。

**パスパラメータ**

| Name | 型 | 必須 | 制約 | 説明 |
|------|----|------|------|------|
| `id` | integer | **Yes** | 正の整数 | 記事ID |

**リクエスト例**
```
DELETE /api/v5/articles/1
```

**レスポンス (204)** — ボディなし

**エラーレスポンス (404)** — 存在しない ID
```json
{ "success": false, "error": { "code": "NOT_FOUND", "message": "article not found" } }
```
