package v1_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "gin-tutorial/app/domain/v1"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newV1Engine() *gin.Engine {
	r := gin.New()
	v1.RegisterRoutes(r.Group("/v1"))
	return r
}

func TestWelcome_DefaultGuest(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/welcome", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello Guest ", w.Body.String())
}

func TestWelcome_CustomName(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/welcome?firstname=John&lastname=Doe", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello John Doe", w.Body.String())
}

func TestFormPost_WithNick(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	body := strings.NewReader("message=hello&nick=Alice")
	req := httptest.NewRequest(http.MethodPost, "/v1/form_post", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "Alice", resp["nick"])
	assert.Equal(t, "hello", resp["message"])
}

func TestFormPost_DefaultNick(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	body := strings.NewReader("message=hi")
	req := httptest.NewRequest(http.MethodPost, "/v1/form_post", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "anonymous", resp["nick"])
}

func TestPost_QueryAndForm(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	body := strings.NewReader("name=Bob&message=test")
	req := httptest.NewRequest(http.MethodPost, "/v1/post?id=42&page=3", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "42", resp["id"])
	assert.Equal(t, "3", resp["page"])
	assert.Equal(t, "Bob", resp["name"])
	assert.Equal(t, "test", resp["message"])
}

func TestFormMap(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	body := strings.NewReader("names[first]=Alice&names[last]=Smith")
	req := httptest.NewRequest(http.MethodPost, "/v1/form_map?ids[a]=1&ids[b]=2", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	ids := resp["ids"].(map[string]interface{})
	assert.Equal(t, "1", ids["a"])
	names := resp["names"].(map[string]interface{})
	assert.Equal(t, "Alice", names["first"])
}

func TestMultipartUpload_WithFile(t *testing.T) {
	r := newV1Engine()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("message", "upload test")
	fw, err := mw.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = fw.Write([]byte("file content"))
	require.NoError(t, err)
	mw.Close()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/multipart", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "test.txt", resp["filename"])
	assert.Equal(t, "upload test", resp["message"])
}

func TestMultipartUpload_WithoutFile(t *testing.T) {
	r := newV1Engine()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("message", "no file")
	mw.Close()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/multipart", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListWithOffset_Defaults(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/articles", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(20), meta["limit"])
	assert.Equal(t, float64(0), meta["offset"])
}

func TestListWithOffset_Custom(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/articles?limit=50&offset=10", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(50), meta["limit"])
	assert.Equal(t, float64(10), meta["offset"])
}

func TestListWithOffset_LimitCapped(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/articles?limit=200", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(100), meta["limit"])
}

func TestListWithCursor_Defaults(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}

func TestListWithCursor_CustomCursorAndLimit(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/events?cursor=abc&limit=5", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}

func TestListWithCursor_LimitCapped(t *testing.T) {
	r := newV1Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/events?limit=200", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}
