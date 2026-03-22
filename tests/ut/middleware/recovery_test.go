package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-tutorial/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
