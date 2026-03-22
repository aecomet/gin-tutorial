package it

import (
	"net/http"
	"testing"

	"gin-tutorial/app/router"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV4_Profile_Authenticated(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v4/profile", nil, map[string]string{
		"Authorization": basicAuthHeader("admin", "secret"),
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "admin", resp["user"])
}

func TestV4_Async(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v4/async", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	tasks, ok := resp["tasks"].([]interface{})
	require.True(t, ok)
	assert.Len(t, tasks, 3)
}
