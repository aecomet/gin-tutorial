package it

import (
	"net/http"
	"testing"

	"gin-tutorial/app/router"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck_V1(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/healthcheck", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "ok", resp["status"])
}

func TestHealthCheck_V2Header(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/healthcheck", nil, map[string]string{"Accept-Version": "v2"})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "v2", resp["version"])
}
