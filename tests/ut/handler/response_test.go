package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-tutorial/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOK_ReturnsSuccessTrue(t *testing.T) {
	// Arrange
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handler.OK(c, nil)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Nil(t, resp.Error)
}

func TestOK_WithData(t *testing.T) {
	// Arrange
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handler.OK(c, gin.H{"key": "value"})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	data, ok := raw["data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value", data["key"])
}

func TestFail_SetsStatusCodeAndCode(t *testing.T) {
	// Arrange
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handler.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid input")
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handler.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
	assert.Equal(t, "invalid input", resp.Error.Message)
}

func TestFail_NoDataField(t *testing.T) {
	// Arrange
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "server error")
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// Act
	r.ServeHTTP(w, req)

	// Assert
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	_, hasData := raw["data"]
	assert.False(t, hasData, "エラーレスポンスにdataフィールドが含まれないこと")
}
