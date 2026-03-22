package it

import (
	"net/http"
	"testing"

	"gin-tutorial/app/router"

	"github.com/stretchr/testify/assert"
)

func TestV2_CreateUser(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodPost, "/api/v2/users", nil, nil)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestV2_ListUsers(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/users", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV2_GetUserByID(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/users/42", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestV2_Products_SortPrice(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/products?sort=price&order=asc", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	filters := resp["filters"].(map[string]interface{})
	assert.Equal(t, "price", filters["sort"])
}

func TestV2_GetItem_Found(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/items/1", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV2_GetItem_NotFound(t *testing.T) {
	// Arrange
	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v2/items/0", nil, nil)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseBody(t, w)
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errInfo["code"])
}
