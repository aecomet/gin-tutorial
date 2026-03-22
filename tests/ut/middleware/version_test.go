package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-tutorial/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
