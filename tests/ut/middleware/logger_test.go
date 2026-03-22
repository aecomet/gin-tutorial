package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-tutorial/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
