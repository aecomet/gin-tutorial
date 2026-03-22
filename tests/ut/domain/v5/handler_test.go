package v5_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gin-tutorial/app/db"
	v5 "gin-tutorial/app/domain/v5"
	"gin-tutorial/app/handler"
	"gin-tutorial/app/middleware"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupV5MockDB(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	mock.ExpectQuery("SELECT VERSION()").
		WillReturnRows(sqlmock.NewRows([]string{"VERSION()"}).AddRow("8.0.0"))
	dialector := mysqldriver.New(mysqldriver.Config{Conn: sqlDB})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)
	db.DB = gormDB
	t.Cleanup(func() { _ = sqlDB.Close() })
	return mock
}

func newV5Engine() *gin.Engine {
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	v5.RegisterRoutes(r.Group("/v5"))
	return r
}

var articleColumns = []string{"id", "title", "body", "author", "created_at", "updated_at", "deleted_at"}

func articleRow(id int, title, body, author string) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows(articleColumns).
		AddRow(id, title, body, author, now, now, sql.NullTime{})
}

// --- listArticles ---

func TestListArticles_OK(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	now := time.Now()
	rows := sqlmock.NewRows(articleColumns).
		AddRow(1, "Title 1", "Body 1", "Alice", now, now, sql.NullTime{}).
		AddRow(2, "Title 2", "Body 2", "Bob", now, now, sql.NullTime{})
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(rows)
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v5/articles?page=1&per_page=10", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	require.NotNil(t, resp.Meta)
	assert.Equal(t, 2, resp.Meta.Total)
}

func TestListArticles_DefaultPagination(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT count\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(articleColumns))
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v5/articles", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	require.NotNil(t, resp.Meta)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 10, resp.Meta.PerPage)
}

// --- createArticle ---

func TestCreateArticle_OK(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	body := `{"title":"Test Article","body":"Test Body","author":"Alice"}`
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v5/articles", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestCreateArticle_MissingTitle(t *testing.T) {
	body := `{"body":"Test Body","author":"Alice"}`
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v5/articles", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
}

func TestCreateArticle_InvalidJSON(t *testing.T) {
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v5/articles", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- getArticleByID ---

func TestGetArticleByID_OK(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(articleRow(1, "Title 1", "Body 1", "Alice"))
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v5/articles/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestGetArticleByID_NotFound(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(sqlmock.NewRows(articleColumns))
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v5/articles/999", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
}

func TestGetArticleByID_InvalidID(t *testing.T) {
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v5/articles/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
}

// --- updateArticle ---

func TestUpdateArticle_OK(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(articleRow(1, "Old Title", "Old Body", "Alice"))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(articleRow(1, "New Title", "Old Body", "Alice"))
	body := `{"title":"New Title"}`
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v5/articles/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateArticle_InvalidID(t *testing.T) {
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v5/articles/xyz", bytes.NewBufferString(`{"title":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateArticle_InvalidJSON(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(articleRow(1, "Title", "Body", "Alice"))
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v5/articles/1", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- deleteArticle ---

func TestDeleteArticle_OK(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(articleRow(1, "Title", "Body", "Alice"))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v5/articles/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestDeleteArticle_InvalidID(t *testing.T) {
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v5/articles/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteArticle_NotFound(t *testing.T) {
	mock := setupV5MockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).WillReturnRows(sqlmock.NewRows(articleColumns))
	r := newV5Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v5/articles/999", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
