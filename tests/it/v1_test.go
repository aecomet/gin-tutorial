package it

import (
	"net/http"
	"strings"
	"testing"

	"gin-tutorial/app/router"

	"github.com/stretchr/testify/assert"
)

func TestV1_Welcome(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/welcome?firstname=World", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Hello World")
}

func TestV1_Articles_DefaultLimit(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/articles?limit=50&offset=10", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(50), meta["limit"])
}

func TestV1_Articles_LimitCapped(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/articles?limit=200", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(100), meta["limit"])
}

func TestV1_Events(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v1/events?cursor=abc&limit=5", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV1_FormPost(t *testing.T) {
	// Arrange
	r := router.New()
	body := strings.NewReader("nick=Alice&message=Hi")

	// Act
	w := doRequest(r, http.MethodPost, "/api/v1/form_post", body, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, "Alice", resp["nick"])
}
