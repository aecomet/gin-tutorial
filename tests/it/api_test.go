package it

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gin-tutorial/app/db"
	"gin-tutorial/app/router"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

// setupITMockDB はインテグレーションテスト用のsqlmockを設定する
func setupITMockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectQuery("SELECT VERSION()").
		WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.0"))

	dialector := mysqldriver.New(mysqldriver.Config{Conn: sqlDB})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	db.DB = gormDB
	t.Cleanup(func() { sqlDB.Close() })
	return mock
}

var itArticleColumns = []string{"id", "title", "body", "author", "created_at", "updated_at", "deleted_at"}

func basicAuthHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func doRequest(r http.Handler, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)
	return w
}

func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func TestHealthCheck_V1(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/healthcheck", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "ok", resp["status"])
}

func TestHealthCheck_V2Header(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/healthcheck", nil, map[string]string{"Accept-Version": "v2"})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "v2", resp["version"])
}

func TestV1_Welcome(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/welcome?firstname=World", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Hello World")
}

func TestV1_Articles_DefaultLimit(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/articles?limit=50&offset=10", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(50), meta["limit"])
}

func TestV1_Articles_LimitCapped(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/articles?limit=200", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(100), meta["limit"])
}

func TestV1_Events(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/events?cursor=abc&limit=5", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV1_FormPost(t *testing.T) {
	// Arrange
	r := router.New()
	body := strings.NewReader("nick=Alice&message=Hi")

	// Act
	w := doRequest(r, http.MethodPost, "/api/v1/form_post", body, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "Alice", resp["nick"])
}

func TestV2_CreateUser(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodPost, "/api/v2/users", nil, nil)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestV2_ListUsers(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/users", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV2_GetUserByID(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/users/42", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV2_Products_SortPrice(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/products?sort=price&order=asc", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	filters := resp["filters"].(map[string]interface{})
	assert.Equal(t, "price", filters["sort"])
}

func TestV2_GetItem_Found(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/items/1", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV2_GetItem_NotFound(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/items/0", nil, nil)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseBody(t, w)
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errInfo["code"])
}

func TestV3_CreateUser(t *testing.T) {
	// Arrange
	r := router.New()
	body := bytes.NewBufferString(`{"name":"Bob","email":"bob@example.com","age":30,"password":"password123"}`)

	// Act
	w := doRequest(r, http.MethodPost, "/api/v3/users", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV3_GetUser(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/users/5", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["id"])
}

func TestV3_Search(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/search?keyword=gin", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "gin", data["keyword"])
}

func TestV3_Login(t *testing.T) {
	// Arrange
	r := router.New()
	body := strings.NewReader("username=alice&password=secret123")

	// Act
	w := doRequest(r, http.MethodPost, "/api/v3/login", body, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV3_ListPosts(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/posts", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV3_GetMe(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/me", nil, map[string]string{
		"Authorization": "Bearer token123",
		"X-Request-Id":  "550e8400-e29b-41d4-a716-446655440000",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV4_Profile_Authenticated(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v4/profile", nil, map[string]string{
		"Authorization": basicAuthHeader("admin", "secret"),
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "admin", resp["user"])
}

func TestV4_Async(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v4/async", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	tasks, ok := resp["tasks"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tasks, 3)
}

// --- v5 integration tests ---

func TestV5_ListArticles(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Title 1", "Body 1", "Alice", now, now, sql.NullTime{}).
			AddRow(2, "Title 2", "Body 2", "Bob", now, now, sql.NullTime{}))

	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v5/articles?page=1&per_page=5", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["total"])
	assert.Equal(t, float64(1), meta["page"])
}

func TestV5_CreateArticle(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO`).WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectCommit()

	r := router.New()
	body := bytes.NewBufferString(`{"title":"Integration Test Article","body":"IT Body","author":"Tester"}`)

	// Act
	w := doRequest(r, http.MethodPost, "/api/v5/articles", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV5_CreateArticle_BadRequest(t *testing.T) {
	// Arrange - no DB mock needed (validation fails before DB access)
	r := router.New()
	body := bytes.NewBufferString(`{"body":"no title here"}`)

	// Act
	w := doRequest(r, http.MethodPost, "/api/v5/articles", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, false, resp["success"])
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errInfo["code"])
}

func TestV5_GetArticleByID_Found(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Title", "Body", "Alice", now, now, sql.NullTime{}))

	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v5/articles/1", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV5_GetArticleByID_NotFound(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns))

	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v5/articles/999", nil, nil)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, false, resp["success"])
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errInfo["code"])
}

func TestV5_UpdateArticle(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Old Title", "Body", "Alice", now, now, sql.NullTime{}))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// 更新後のリロード
	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Updated Title", "Body", "Alice", now, now, sql.NullTime{}))

	r := router.New()
	body := bytes.NewBufferString(`{"title":"Updated Title"}`)

	// Act
	w := doRequest(r, http.MethodPut, "/api/v5/articles/1", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV5_DeleteArticle(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Title", "Body", "Alice", now, now, sql.NullTime{}))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r := router.New()

	// Act
	w := doRequest(r, http.MethodDelete, "/api/v5/articles/1", nil, nil)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}
