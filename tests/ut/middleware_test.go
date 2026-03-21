package ut

import (
	"encoding/json"
	"errors"
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

// ErrorHandler tests

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
		_ = c.Error(errors.New("something went wrong"))
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

// Logger tests

func TestLogger_PassThrough(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.Logger())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

// Recovery tests

func TestRecovery_Panic(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Version tests

func TestVersion_WithHeader(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.Version())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("api_version"))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Version", "v2")

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v2", w.Body.String())
}

func TestVersion_DefaultsToV1(t *testing.T) {
	// Arrange
	r := gin.New()
	r.Use(middleware.Version())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("api_version"))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v1", w.Body.String())
}
