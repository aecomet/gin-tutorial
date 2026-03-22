package v4_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v4 "gin-tutorial/app/domain/v4"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newV4Engine() *gin.Engine {
	r := gin.New()
	v4.RegisterRoutes(r.Group("/v4"))
	return r
}

func basicAuth(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func TestGetProfile_Authenticated(t *testing.T) {
	r := newV4Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v4/profile", nil)
	req.Header.Set("Authorization", basicAuth("admin", "secret"))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "admin", resp["user"])
	assert.Equal(t, "authenticated successfully", resp["message"])
}

func TestGetProfile_Unauthenticated(t *testing.T) {
	r := newV4Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v4/profile", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetSecret_Authenticated(t *testing.T) {
	r := newV4Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v4/secret", nil)
	req.Header.Set("Authorization", basicAuth("user", "password"))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "user", resp["user"])
	assert.Equal(t, "this is confidential", resp["secret"])
}

func TestGetSecret_Unauthenticated(t *testing.T) {
	r := newV4Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v4/secret", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAsyncTasks(t *testing.T) {
	r := newV4Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v4/async", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	tasks, ok := resp["tasks"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tasks, 3)
	for _, task := range tasks {
		item := task.(map[string]interface{})
		assert.NotEmpty(t, item["task"])
		assert.NotEmpty(t, item["result"])
		assert.NotEmpty(t, item["duration"])
	}
}
