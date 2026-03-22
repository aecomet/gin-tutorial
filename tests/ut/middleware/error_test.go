package middleware_test

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

func TestErrorHandler_AppError(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(handler.ErrNotFound)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, false, body["success"])
	errInfo := body["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errInfo["code"])
	assert.Equal(t, "resource not found", errInfo["message"])
}

func TestErrorHandler_UnknownError(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(assert.AnError)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, false, body["success"])
	errInfo := body["error"].(map[string]interface{})
	assert.Equal(t, "INTERNAL", errInfo["code"])
}
