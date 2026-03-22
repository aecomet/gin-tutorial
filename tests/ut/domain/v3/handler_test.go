package v3_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v3 "gin-tutorial/app/domain/v3"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newV3Engine() *gin.Engine {
	r := gin.New()
	v3.RegisterRoutes(r.Group("/v3"))
	return r
}

func TestCreateUser_V3_Valid(t *testing.T) {
	r := newV3Engine()
	body := `{"name":"Alice","email":"alice@example.com","age":25,"password":"secret123"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Alice", data["name"])
	assert.Equal(t, "alice@example.com", data["email"])
	assert.Equal(t, float64(25), data["age"])
}

func TestCreateUser_V3_MissingName(t *testing.T) {
	r := newV3Engine()
	body := `{"email":"alice@example.com","age":25,"password":"secret123"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errInfo["code"])
}

func TestCreateUser_V3_InvalidEmail(t *testing.T) {
	r := newV3Engine()
	body := `{"name":"Alice","email":"not-an-email","age":25,"password":"secret123"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errInfo["code"])
}

func TestGetUser_V3_Valid(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users/5", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["id"])
}

func TestGetUser_V3_ZeroID(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users/0", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errInfo["code"])
}

func TestSearch_WithKeyword(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/search?keyword=gin", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "gin", data["keyword"])
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(20), data["per_page"])
}

func TestSearch_MissingKeyword(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/search", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errInfo["code"])
}

func TestSearch_DefaultPagePerPage(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/search?keyword=test", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["page"])
	assert.Equal(t, float64(20), data["per_page"])
}

func TestLogin_Valid(t *testing.T) {
	r := newV3Engine()
	body := strings.NewReader("username=alice&password=secret123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "alice", data["username"])
	assert.Equal(t, "login successful", data["message"])
}

func TestLogin_MissingUsername(t *testing.T) {
	r := newV3Engine()
	body := strings.NewReader("password=secret123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errInfo["code"])
}

func TestListPosts_Defaults(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/posts", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	meta := data["meta"].(map[string]interface{})
	assert.Equal(t, float64(0), meta["page"])
	assert.Equal(t, float64(0), meta["per_page"])
	assert.Equal(t, "", meta["sort"])
	assert.Equal(t, "", meta["order"])
	assert.Equal(t, "", meta["status"])
}

func TestListPosts_CustomParams(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/posts?page=2&per_page=10&sort=title&order=asc&status=draft", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	meta := data["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["page"])
	assert.Equal(t, float64(10), meta["per_page"])
	assert.Equal(t, "title", meta["sort"])
	assert.Equal(t, "asc", meta["order"])
	assert.Equal(t, "draft", meta["status"])
}

func TestGetMe_Valid(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/me", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Request-Id", "550e8400-e29b-41d4-a716-446655440000")
	req.Header.Set("Accept-Language", "en-US")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "Bearer token123", data["authorization"])
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", data["request_id"])
}

func TestGetMe_MissingAuthorization(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/me", nil)
	req.Header.Set("X-Request-Id", "550e8400-e29b-41d4-a716-446655440000")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errInfo["code"])
}

func TestGetMe_InvalidRequestID(t *testing.T) {
	r := newV3Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/me", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("X-Request-Id", "not-a-uuid")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errInfo["code"])
}
