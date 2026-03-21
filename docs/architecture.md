# アーキテクチャ設計資料

## ディレクトリ構成

```
gin-tutorial/
├── main.go              # エントリーポイント。サーバー起動のみ
└── app/
    ├── router/
    │   └── router.go        # ルーティング定義。全グループをここで組み立てる
    ├── handler/             # Gin汎用ユーティリティ（特定ドメインに依存しない共通処理）
    │   ├── errors.go        # AppError 型・共通エラー変数
    │   ├── health.go        # ヘルスチェックハンドラー
    │   └── response.go      # 統一レスポンス型 (Response, OK, Fail)
    ├── middleware/          # Ginミドルウェア
    │   ├── error.go         # エラーハンドリングミドルウェア（AppError → JSON変換）
    │   └── version.go       # Accept-Version ヘッダーバージョニングミドルウェア
    └── domain/              # ドメイン・バージョン別ハンドラー
        ├── v1/
        │   └── handler.go   # v1 APIデモ（Gin機能サンプル群）
        └── v2/
            └── handler.go   # v2 ドメインルート（users, products, orders, items）
        └── v3/
            └── handler.go   # v3 モデルバインディング・バリデーションのデモ
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
└── /v3                        # モデルバインディング・バリデーションデモ
    ├── POST /users            # JSON バインディング（required / email / min / max / gte）
    ├── GET  /users/:id        # URI バインディング（gt=0）
    ├── GET  /search           # クエリ バインディング（omitempty / gte / lte）
    └── POST /login            # フォーム バインディング（required / min）
    ├── /users
    │   ├── GET    /users                    # 一覧
    │   ├── POST   /users                    # 作成
    │   ├── GET    /users/:id                # 取得
    │   ├── PUT    /users/:id                # 更新
    │   ├── DELETE /users/:id                # 削除
    │   ├── GET    /users/:name/hello        # パスパラメータデモ
    │   └── GET    /users/:name/*action      # ワイルドカードパスパラメータデモ
    ├── /products
    │   └── GET    /products                 # フィルタリング・ソートデモ
    ├── /orders
    │   ├── GET    /orders                   # 一覧
    │   ├── POST   /orders                   # 作成
    │   └── GET    /orders/:id              # 取得
    └── /items
        └── GET    /items/:id               # カスタムエラーハンドリングデモ
```

## 設計方針

### パッケージ分割
- **`handler/`** は特定ドメインに依存しない共通処理のみを置く。レスポンス形式・エラー型・ヘルスチェックが該当する。
- **`domain/v1/`** はGin機能を示すサンプルAPIをまとめる。フォーム・ページネーション等のデモが対象。
- **`domain/v2/`** はリソースベースのドメインAPIをまとめる。users / products / orders / items の各リソースハンドラーを1ファイルに集約する。
- **`middleware/`** はリクエスト横断的な処理を置く。

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
