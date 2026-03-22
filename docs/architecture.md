# アーキテクチャ設計資料

## ディレクトリ構成

```
gin-tutorial/
├── main.go              # エントリーポイント。DB初期化・マイグレーション・サーバー起動
├── Dockerfile           # マルチステージビルド（scratch ベース）
├── docker-compose.yml   # docker compose up -d でアプリ + MySQL 起動
├── logs/                # ログ出力ディレクトリ（.gitignore: *.log）
├── tests/               # テストコード
│   ├── ut/              # ユニットテスト（app/ のパッケージ構成を反映）
│   │   ├── db/          # app/db のテスト
│   │   ├── logger/      # app/logger のテスト
│   │   ├── handler/     # app/handler のテスト（errors / health / response）
│   │   ├── middleware/  # app/middleware のテスト（error / logger / recovery / version）
│   │   └── domain/      # app/domain のテスト
│   │       ├── v1/ ~ v4/   # 各バージョンのハンドラーテスト
│   │       └── v5/      # handler / migrate / seed のテスト
│   └── it/              # インテグレーションテスト（router.New() を使った E2E シナリオ）
└── app/
    ├── db/
    │   └── db.go            # GORM DBインスタンス初期化（環境変数からDSN構築）
    ├── logger/
    │   └── logger.go        # slog JSONハンドラーの初期化（logs/app.log へ出力）
    ├── router/
    │   └── router.go        # ルーティング定義。全グループをここで組み立てる
    ├── handler/             # 共通ユーティリティ（特定ドメインに依存しない）
    │   ├── errors.go        # AppError 型・共通エラー変数
    │   ├── health.go        # ヘルスチェックハンドラー
    │   └── response.go      # 統一レスポンス型 (Response, OK, Fail)
    ├── middleware/          # Ginミドルウェア
    │   ├── error.go         # エラーハンドリング（AppError → JSON変換）
    │   ├── logger.go        # HTTPリクエストログ（slog JSON出力）
    │   ├── recovery.go      # panicリカバリー（slog.Error で記録）
    │   └── version.go       # Accept-Version ヘッダーバージョニング
    └── domain/              # バージョン別ハンドラー
        ├── v1/
        │   └── handler.go   # Gin基本機能デモ（フォーム・ページネーション等）
        ├── v2/
        │   └── handler.go   # リソースベースCRUD（users / products / orders / items）
        ├── v3/
        │   └── handler.go   # モデルバインディング・バリデーションデモ
        ├── v4/
        │   └── handler.go   # Basic認証・goroutine非同期処理デモ
        └── v5/
            ├── model.go     # Article GORMモデル（バインディング入力型を含む）
            ├── handler.go   # GORM連携CRUDハンドラー（articles リソース）
            ├── migrate.go   # AutoMigrateによるDDL自動適用
            └── seed.go      # 初期データ投入（DB_SEED=true で実行）
```

## ルーティング構成

各エンドポイントの詳細な API 仕様は **[docs/api.md](api.md)** を参照してください。

ルートの概要:

| プレフィックス | 説明 |
|--------------|------|
| `GET /api/healthcheck` | ヘルスチェック（Accept-Version ヘッダーバージョニングのデモ） |
| `GET /api/routes` | 登録済みルート一覧 |
| `/api/v1/...` | Gin 基本機能デモ（クエリパラメータ・フォーム・ページネーション） |
| `/api/v2/...` | リソースベース CRUD デモ（users / products / orders / items） |
| `/api/v3/...` | モデルバインディング・バリデーションデモ |
| `/api/v4/...` | Basic 認証・goroutine 非同期処理デモ |
| `/api/v5/...` | GORM + MySQL CRUD デモ（articles リソース） |

## 設計方針

### テスト戦略
テストは `tests/` 配下の `ut` / `it` パッケージに分離している。すべてのテストケースは AAA（Arrange / Act / Assert）パターンで記述する。

- **`tests/ut/`**: 各ハンドラー・ミドルウェアを `httptest.NewRecorder()` で個別に検証する。`router.New()` を使わず最小限の `gin.Engine` を構築するため、外部依存がない。UT 単体で **約 96%** のカバレッジを達成している。
- **`tests/it/`**: `router.New()` で実際のルーター全体を起動し、エンドポイントの正常系シナリオを E2E で検証する。

テスト実行:
```bash
go test ./tests/...                        # 全テスト
go test ./tests/ut/... -cover -coverpkg=./app/...  # UT カバレッジ計測
```

### ログ出力
`log/slog`（Go 1.21 標準ライブラリ）を使い、すべてのログを JSON 形式で `logs/app.log` に出力する。

- **`logger.Init()`** をサーバー起動前に呼び出し、`slog.SetDefault` でプロセス全体のロガーを設定する。
- **`middleware.Logger()`** がリクエストごとに `method / path / status / latency / ip / user_agent` を `slog.Info` で記録する。
- **`middleware.Recovery()`** が panic 発生時に `slog.Error` で記録し、500 を返す。
- ログファイルは `logs/app.log` に追記される。Docker 環境では `docker-compose.yml` の volume mount（`./logs:/app/logs`）でホストからも確認できる。`logs/*.log` は `.gitignore` で除外し、ディレクトリのみ `logs/.gitkeep` で管理する。

### パッケージ分割
- **`logger/`** はアプリ起動時に一度だけ呼ぶ初期化処理を置く。`slog.NewJSONHandler` で `logs/app.log` への JSON ログを設定する。ファイルクローズ用の cleanup 関数を返す。
- **`handler/`** は特定ドメインに依存しない共通処理のみを置く。レスポンス形式・エラー型・ヘルスチェックが該当する。
- **`middleware/`** はリクエスト横断的な処理を置く。ErrorHandler・Version ミドルウェアが該当する。
- **`domain/v1/`** はGin基本機能のサンプルAPIをまとめる。クエリパラメータ・フォームデータ・ページネーション等のデモが対象。
- **`domain/v2/`** はリソースベースのCRUD APIをまとめる。users / products / orders / items の各リソースハンドラーを1ファイルに集約する。フィルタリング・ソート・カスタムエラーハンドリングもここでデモする。
- **`domain/v3/`** はモデルバインディングとバリデーションのサンプルAPIをまとめる。JSON / URI / クエリ / フォーム / ヘッダー / デフォルト値の各バインディングパターンをカバーする。
- **`domain/v4/`** はGinの認証と非同期処理のサンプルAPIをまとめる。`gin.BasicAuth` ミドルウェアと `sync.WaitGroup` を使った goroutine 並列実行が対象。
- **`domain/v5/`** はGORMを使ったMySQL連携CRUDのサンプルAPIをまとめる。articles リソースのCRUD・ページネーション・論理削除・AutoMigrate・Seed が対象。
- **`db/`** はGORMのDBインスタンスをシングルトンとして管理する。環境変数（`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`）からDSNを構築し、`db.DB` を通じてアプリ全体で共有する。

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
