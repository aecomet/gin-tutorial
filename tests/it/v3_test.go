package it

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"gin-tutorial/app/router"

	"github.com/stretchr/testify/assert"
)

func TestV3_CreateUser(t *testing.T) {
	// Arrange
	r := router.New()
	body := bytes.NewBufferString(`{"name":"Bob","email":"bob@example.com","age":30,"password":"password123"}`)

	// Act
	w := doRequest(r, http.MethodPost, "/api/v3/users", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV3_GetUser(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/users/5", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["id"])
}

func TestV3_Search(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/search?keyword=gin", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "gin", data["keyword"])
}

func TestV3_Login(t *testing.T) {
	// Arrange
	r := router.New()
	body := strings.NewReader("username=alice&password=secret123")

	// Act
	w := doRequest(r, http.MethodPost, "/api/v3/login", body, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV3_ListPosts(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/posts", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV3_GetMe(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v3/me", nil, map[string]string{
		"Authorization": "Bearer token123",
		"X-Request-Id":  "550e8400-e29b-41d4-a716-446655440000",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}
