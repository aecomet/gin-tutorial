package v2_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "gin-tutorial/app/domain/v2"
	"gin-tutorial/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newV2Engine() *gin.Engine {
	r := gin.New()
	v2.RegisterRoutes(r.Group("/v2"))
	return r
}

func newV2EngineWithErrorHandler() *gin.Engine {
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	v2.RegisterRoutes(r.Group("/v2"))
	return r
}

func TestListUsers(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "list_users", resp["action"])
}

func TestCreateUser_V2(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v2/users", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "create_user", resp["action"])
}

func TestGetUserByID_V2(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/users/42", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "get_user", resp["action"])
}

func TestUpdateUser(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v2/users/42", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "update_user", resp["action"])
}

func TestDeleteUser(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v2/users/42", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestListProducts_Defaults(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/products", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	filters := resp["filters"].(map[string]interface{})
	assert.Equal(t, "created_at", filters["sort"])
	assert.Equal(t, "desc", filters["order"])
}

func TestListProducts_ValidSortOrder(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/products?sort=price&order=asc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	filters := resp["filters"].(map[string]interface{})
	assert.Equal(t, "price", filters["sort"])
	assert.Equal(t, "asc", filters["order"])
}

func TestListProducts_InvalidSort(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/products?sort=invalid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	filters := resp["filters"].(map[string]interface{})
	assert.Equal(t, "created_at", filters["sort"])
}

func TestListProducts_InvalidOrder(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/products?order=invalid", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	filters := resp["filters"].(map[string]interface{})
	assert.Equal(t, "desc", filters["order"])
}

func TestListOrders(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/orders", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "list_orders", resp["action"])
}

func TestCreateOrder(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v2/orders", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "create_order", resp["action"])
}

func TestGetOrderByID(t *testing.T) {
	r := newV2Engine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/orders/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "get_order", resp["action"])
}

func TestGetItemByID_NotFound(t *testing.T) {
	r := newV2EngineWithErrorHandler()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/items/0", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errInfo["code"])
}

func TestGetItemByID_Found(t *testing.T) {
	r := newV2EngineWithErrorHandler()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v2/items/1", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}
