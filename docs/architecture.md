# アーキテクチャ設計資料

## ディレクトリ構成

```
gin-tutorial/
├── main.go              # エントリーポイント。サーバー起動のみ
├── Dockerfile           # マルチステージビルド（scratch ベース）
├── docker-compose.yml   # docker compose up -d でアプリ起動
└── app/
    ├── router/
    │   └── router.go        # ルーティング定義。全グループをここで組み立てる
    ├── handler/             # 共通ユーティリティ（特定ドメインに依存しない）
    │   ├── errors.go        # AppError 型・共通エラー変数
    │   ├── health.go        # ヘルスチェックハンドラー
    │   └── response.go      # 統一レスポンス型 (Response, OK, Fail)
    ├── middleware/          # Ginミドルウェア
    │   ├── error.go         # エラーハンドリング（AppError → JSON変換）
    │   └── version.go       # Accept-Version ヘッダーバージョニング
    └── domain/              # バージョン別ハンドラー
        ├── v1/
        │   └── handler.go   # Gin基本機能デモ（フォーム・ページネーション等）
        ├── v2/
        │   └── handler.go   # リソースベースCRUD（users / products / orders / items）
        ├── v3/
        │   └── handler.go   # モデルバインディング・バリデーションデモ
        └── v4/
            └── handler.go   # Basic認証・goroutine非同期処理デモ
```

## ルーティング構成

```
/api
├── GET  /healthcheck          # ヘルスチェック（Accept-Versionヘッダーバージョニングのデモ）
├── GET  /routes               # 登録済みルート一覧
│
├── /v1                        # Gin機能デモ
│   ├── GET  /welcome          # クエリパラメータ
│   ├── POST /form_post        # POSTフォームデータ
│   ├── POST /post             # クエリパラメータ + POSTフォームデータ
│   ├── POST /form_map         # QueryMap + PostFormMap
│   ├── POST /multipart        # multipart/form-data（ファイルアップロード）
│   ├── GET  /articles         # ページネーション（limit/offset）
│   └── GET  /events           # ページネーション（カーソルベース）
│
├── /v2                        # ドメインルート
│   ├── GET/POST       /users          # ユーザー一覧・作成
│   ├── GET/PUT/DELETE /users/:id      # ユーザー取得・更新・削除
│   ├── GET            /products       # フィルタリング・ソートデモ
│   ├── GET/POST       /orders         # オーダー一覧・作成
│   ├── GET            /orders/:id     # オーダー取得
│   └── GET            /items/:id      # カスタムエラーハンドリングデモ
│
├── /v3                        # モデルバインディング・バリデーションデモ
│   ├── POST /users            # JSON バインディング（required / email / min / max / gte）
│   ├── GET  /users/:id        # URI バインディング（gt=0）
│   ├── GET  /search           # クエリ バインディング（omitempty / gte / lte）
│   ├── POST /login            # フォーム バインディング（required / min）
│   ├── GET  /posts            # デフォルト値付きクエリ バインディング（default タグ / oneof）
│   └── GET  /me               # ヘッダー バインディング（required / uuid4）
│
└── /v4                        # Basic認証・goroutine非同期処理デモ
    ├── GET /profile           # Basic認証（gin.BasicAuth ミドルウェア）
    ├── GET /secret            # Basic認証保護リソース
    └── GET /async             # goroutine による並列タスク実行
```

## 設計方針

### パッケージ分割
- **`handler/`** は特定ドメインに依存しない共通処理のみを置く。レスポンス形式・エラー型・ヘルスチェックが該当する。
- **`middleware/`** はリクエスト横断的な処理を置く。ErrorHandler・Version ミドルウェアが該当する。
- **`domain/v1/`** はGin基本機能のサンプルAPIをまとめる。クエリパラメータ・フォームデータ・ページネーション等のデモが対象。
- **`domain/v2/`** はリソースベースのCRUD APIをまとめる。users / products / orders / items の各リソースハンドラーを1ファイルに集約する。フィルタリング・ソート・カスタムエラーハンドリングもここでデモする。
- **`domain/v3/`** はモデルバインディングとバリデーションのサンプルAPIをまとめる。JSON / URI / クエリ / フォーム / ヘッダー / デフォルト値の各バインディングパターンをカバーする。
- **`domain/v4/`** はGinの認証と非同期処理のサンプルAPIをまとめる。`gin.BasicAuth` ミドルウェアと `sync.WaitGroup` を使った goroutine 並列実行が対象。

### エラーハンドリング
`handler.AppError` 型を `c.Error()` にセットし、`middleware.ErrorHandler` が一括してJSONレスポンスに変換する。ハンドラー内で直接 `c.JSON` を呼ぶ必要はない。

### バージョニング戦略
2種類のバージョニングをデモとして実装している。
- **URLパスバージョニング**: `/api/v1/...`、`/api/v2/...` でバージョンを分離する。
- **ヘッダーバージョニング**: `Accept-Version: v2` ヘッダーで挙動を切り替える（`/api/healthcheck` で確認可能）。

### レスポンス形式
`handler.Response` 型でAPIレスポンスを統一する。

```json
{
  "success": true,
  "data": { ... },
  "error": { "code": "NOT_FOUND", "message": "resource not found" },
  "meta": { "page": 1, "per_page": 20, "total": 100 }
}
```
