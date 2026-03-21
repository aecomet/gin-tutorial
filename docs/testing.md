# テスト設計資料

## 達成目標

| 指標 | 目標値 | 現状 |
|------|--------|------|
| UT カバレッジ（`./app/...`） | **90% 以上** | **95.9%** ✅ |
| UT テスト数 | 全ハンドラー・ミドルウェアの主要パスを網羅 | **54 ケース** ✅ |
| IT テスト数 | 全エンドポイントの正常系を網羅 | **21 ケース** ✅ |
| テストパターン | すべてのケースに AAA パターンを適用 | ✅ |
| CI 要件 | `go test ./tests/...` がコミット前に全件通過すること | ✅ |

---

## テストコマンド

```bash
# 全テスト実行
go test ./tests/...

# UT のみ（カバレッジ計測付き）
go test ./tests/ut/... -cover -coverpkg=./app/...

# IT のみ
go test ./tests/it/...

# 詳細出力
go test ./tests/... -v
```

---

## テスト構成

```
tests/
├── ut/                       # ユニットテスト（54 ケース）
│   ├── middleware_test.go    # ミドルウェア
│   ├── v1_handler_test.go    # v1 ハンドラー
│   ├── v2_handler_test.go    # v2 ハンドラー
│   ├── v3_handler_test.go    # v3 ハンドラー
│   └── v4_handler_test.go    # v4 ハンドラー
└── it/                       # インテグレーションテスト（21 ケース）
    └── api_test.go           # 全エンドポイントの正常系シナリオ
```

---

## ユニットテスト（tests/ut）

**方針**: `router.New()` を使わず、テスト対象のハンドラーを最小構成の `gin.Engine` に登録して個別に検証する。外部依存なし。

### テストパターン（AAA）

すべてのテストケースは以下のコメントで構造化する。

```go
// Arrange — テスト対象のセットアップ・入力データの準備
// Act     — テスト対象の実行（HTTP リクエスト送信）
// Assert  — レスポンスの検証
```

### middleware_test.go（6 ケース）

| テスト名 | 検証内容 |
|---------|---------|
| `TestErrorHandler_AppError` | `c.Error(ErrNotFound)` → JSON `success=false, code=NOT_FOUND`, status 404 |
| `TestErrorHandler_UnknownError` | 不明エラー → JSON `code=INTERNAL`, status 500 |
| `TestLogger_PassThrough` | リクエストが正常に通過し status 200 を返す |
| `TestRecovery_Panic` | ハンドラー内 panic → status 500、プロセスが落ちない |
| `TestVersion_WithHeader` | `Accept-Version: v2` → コンテキストに `api_version=v2` がセットされる |
| `TestVersion_DefaultsToV1` | ヘッダーなし → `api_version=v1` がデフォルト適用 |

### v1_handler_test.go（14 ケース）

| テスト名 | 検証内容 |
|---------|---------|
| `TestWelcome_DefaultGuest` | クエリなし → `Hello Guest ` |
| `TestWelcome_CustomName` | `?firstname=Gin&lastname=Gopher` → `Hello Gin Gopher` |
| `TestFormPost_WithNick` | nick 指定 → レスポンスに nick が反映される |
| `TestFormPost_DefaultNick` | nick 未指定 → `nick=anonymous` |
| `TestPost_QueryAndForm` | クエリ id/page + フォーム name/message → 全フィールド返却 |
| `TestFormMap` | `ids[first]=1` + `names[a]=foo` → QueryMap / PostFormMap が正しく返る |
| `TestMultipartUpload_WithFile` | ファイル添付 → 200、filename/size が返る |
| `TestMultipartUpload_WithoutFile` | ファイルなし → 400 |
| `TestListWithOffset_Defaults` | パラメータなし → `limit=20, offset=0` |
| `TestListWithOffset_Custom` | `?limit=5&offset=10` → そのまま反映 |
| `TestListWithOffset_LimitCapped` | `?limit=200` → `limit=100` に上限適用 |
| `TestListWithCursor_Defaults` | パラメータなし → `limit=20` |
| `TestListWithCursor_CustomCursorAndLimit` | `?cursor=abc&limit=5` → 正常レスポンス |
| `TestListWithCursor_LimitCapped` | `?limit=200` → `limit=100` に上限適用 |

### v2_handler_test.go（13 ケース）

| テスト名 | 検証内容 |
|---------|---------|
| `TestListUsers` | GET /users → 200 `action=list_users` |
| `TestCreateUser_V2` | POST /users → 201 `action=create_user` |
| `TestGetUserByID_V2` | GET /users/:id → 200 `action=get_user` |
| `TestUpdateUser` | PUT /users/:id → 200 `action=update_user` |
| `TestDeleteUser` | DELETE /users/:id → 204 ボディなし |
| `TestListProducts_Defaults` | クエリなし → `sort=created_at, order=desc` |
| `TestListProducts_ValidSortOrder` | `?sort=price&order=asc` → そのまま反映 |
| `TestListProducts_InvalidSort` | `?sort=invalid` → `sort=created_at` にフォールバック |
| `TestListProducts_InvalidOrder` | `?order=invalid` → `order=desc` にフォールバック |
| `TestListOrders` | GET /orders → 200 `action=list_orders` |
| `TestCreateOrder` | POST /orders → 201 `action=create_order` |
| `TestGetOrderByID` | GET /orders/:id → 200 `action=get_order` |
| `TestGetItemByID_NotFound` | id=0 → `c.Error(ErrNotFound)` → 404 NOT_FOUND（ErrorHandler チェーン） |
| `TestGetItemByID_Found` | id=1 → 200 `success=true, data.id=1` |

### v3_handler_test.go（14 ケース）

| テスト名 | 検証内容 |
|---------|---------|
| `TestCreateUser_V3_Valid` | 正常 JSON → 200 `name/email/age` 返却 |
| `TestCreateUser_V3_MissingName` | name 未指定 → 400 VALIDATION_ERROR |
| `TestCreateUser_V3_InvalidEmail` | email 不正 → 400 VALIDATION_ERROR |
| `TestGetUser_V3_Valid` | id=5 → 200 `data.id=5` |
| `TestGetUser_V3_ZeroID` | id=0 → 400（`gt=0` バリデーション失敗） |
| `TestSearch_WithKeyword` | `?keyword=gin` → 200 `keyword=gin` |
| `TestSearch_MissingKeyword` | keyword なし → 400 VALIDATION_ERROR |
| `TestSearch_DefaultPagePerPage` | keyword のみ → `page=1, per_page=20` デフォルト適用 |
| `TestLogin_Valid` | username/password 正常 → 200 `login successful` |
| `TestLogin_MissingUsername` | username なし → 400 VALIDATION_ERROR |
| `TestListPosts_Defaults` | パラメータなし → `page=1, per_page=20, sort=created_at, order=desc, status=published` |
| `TestListPosts_CustomParams` | 全パラメータ指定 → 指定値が反映される |
| `TestGetMe_Valid` | Authorization + UUID v4 X-Request-Id → 200 |
| `TestGetMe_MissingAuthorization` | Authorization なし → 400 VALIDATION_ERROR |
| `TestGetMe_InvalidRequestID` | X-Request-Id が UUID v4 形式でない → 400 VALIDATION_ERROR |

### v4_handler_test.go（5 ケース）

| テスト名 | 検証内容 |
|---------|---------|
| `TestGetProfile_Authenticated` | `admin:secret` で認証 → 200 `user=admin` |
| `TestGetProfile_Unauthenticated` | 認証なし → 401 |
| `TestGetSecret_Authenticated` | `user:password` で認証 → 200 `user=user, secret` |
| `TestGetSecret_Unauthenticated` | 認証なし → 401 |
| `TestAsyncTasks` | GET /async → 200、`tasks` 配列に 3 要素（task/result/duration） |

---

## インテグレーションテスト（tests/it）

**方針**: `router.New()` で実際のルーター全体を起動し、エンドポイントの正常系を E2E で検証する。ログ出力は `io.Discard` に向けてテスト出力を抑制する。

### api_test.go（21 ケース）

| テスト名 | 検証内容 |
|---------|---------|
| `TestHealthCheck_V1` | GET /api/healthcheck → 200 `status=ok` |
| `TestHealthCheck_V2Header` | Accept-Version: v2 → 200 `version=v2, detail` |
| `TestV1_Welcome` | GET /api/v1/welcome?firstname=World → `Hello World` |
| `TestV1_Articles_DefaultLimit` | GET /api/v1/articles?limit=50&offset=10 → `meta.limit=50` |
| `TestV1_Articles_LimitCapped` | limit=200 → `meta.limit=100` |
| `TestV1_Events` | GET /api/v1/events?cursor=abc&limit=5 → 200 |
| `TestV1_FormPost` | POST /api/v1/form_post nick=Alice → 200 `nick=Alice` |
| `TestV2_ListUsers` | GET /api/v2/users → 200 |
| `TestV2_CreateUser` | POST /api/v2/users → 201 |
| `TestV2_GetUserByID` | GET /api/v2/users/42 → 200 |
| `TestV2_Products_SortPrice` | GET /api/v2/products?sort=price&order=asc → `filters.sort=price` |
| `TestV2_GetItem_Found` | GET /api/v2/items/1 → 200 `success=true` |
| `TestV2_GetItem_NotFound` | GET /api/v2/items/0 → 404 `error.code=NOT_FOUND` |
| `TestV3_CreateUser` | POST /api/v3/users 正常 JSON → 200 |
| `TestV3_GetUser` | GET /api/v3/users/5 → 200 `data.id=5` |
| `TestV3_Search` | GET /api/v3/search?keyword=gin → 200 `data.keyword=gin` |
| `TestV3_Login` | POST /api/v3/login 正常フォーム → 200 |
| `TestV3_ListPosts` | GET /api/v3/posts → 200 デフォルト値適用 |
| `TestV3_GetMe` | GET /api/v3/me 正常ヘッダー → 200 |
| `TestV4_Profile_Authenticated` | GET /api/v4/profile admin:secret → 200 |
| `TestV4_Async` | GET /api/v4/async → 200 `tasks` 配列あり |

---

## カバレッジ詳細

```
gin-tutorial/app/domain/v1/handler.go  ... 100.0%
gin-tutorial/app/domain/v2/handler.go  ... 100.0%
gin-tutorial/app/domain/v3/handler.go  ...  ~97%  (listPosts の default タグ分岐)
gin-tutorial/app/domain/v4/handler.go  ... 100.0%
gin-tutorial/app/handler/errors.go     ...  ~50%  (Error() メソッドは間接呼び出しのみ)
gin-tutorial/app/handler/health.go     ...   0%   (IT でカバー・UT スコープ外)
gin-tutorial/app/handler/response.go   ... 100.0%
gin-tutorial/app/middleware/error.go   ... 100.0%
gin-tutorial/app/middleware/logger.go  ...  90%   (query あり/なし分岐の一方)
gin-tutorial/app/middleware/recovery.go... 100.0%
gin-tutorial/app/middleware/version.go ... 100.0%
─────────────────────────────────────────────────
total                                     95.9%
```

> `handler/health.go` は IT テストで網羅されているが、UT の `-coverpkg` 計測スコープ外のため 0% となっている。UT + IT 合算では実質 100% カバー済み。
