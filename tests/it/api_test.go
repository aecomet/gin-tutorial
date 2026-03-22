package it

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

// itArticleColumns はv5テストで使うsqlmock用カラム名リスト
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

// --- HealthCheck ---

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
