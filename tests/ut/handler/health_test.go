package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-tutorial/app/handler"
	"gin-tutorial/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthCheck_DefaultV1(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.Version())
	r.GET("/healthcheck", handler.HealthCheck)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
	_, hasVersion := resp["version"]
	assert.False(t, hasVersion, "v1レスポンスにversionフィールドが含まれないこと")
}

func TestHealthCheck_V2Header(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.Version())
	r.GET("/healthcheck", handler.HealthCheck)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	req.Header.Set("Accept-Version", "v2")

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "v2", resp["version"])
}
