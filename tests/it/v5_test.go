package it

import (
	"bytes"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"gin-tutorial/app/router"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestV5_ListArticles(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT count\(\*\) FROM`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Title 1", "Body 1", "Alice", now, now, sql.NullTime{}).
			AddRow(2, "Title 2", "Body 2", "Bob", now, now, sql.NullTime{}))

	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v5/articles?page=1&per_page=5", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
	meta := resp["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["total"])
	assert.Equal(t, float64(1), meta["page"])
}

func TestV5_CreateArticle(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO`).WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectCommit()

	r := router.New()
	body := bytes.NewBufferString(`{"title":"Integration Test Article","body":"IT Body","author":"Tester"}`)

	// Act
	w := doRequest(r, http.MethodPost, "/api/v5/articles", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV5_CreateArticle_BadRequest(t *testing.T) {
	// Arrange - titleはrequiredのためDBアクセスなし
	r := router.New()
	body := bytes.NewBufferString(`{"body":"no title here"}`)

	// Act
	w := doRequest(r, http.MethodPost, "/api/v5/articles", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, false, resp["success"])
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "BAD_REQUEST", errInfo["code"])
}

func TestV5_GetArticleByID_Found(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Title", "Body", "Alice", now, now, sql.NullTime{}))

	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v5/articles/1", nil, nil)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV5_GetArticleByID_NotFound(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns))

	r := router.New()

	// Act
	w := doRequest(r, http.MethodGet, "/api/v5/articles/999", nil, nil)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, false, resp["success"])
	errInfo := resp["error"].(map[string]interface{})
	assert.Equal(t, "NOT_FOUND", errInfo["code"])
}

func TestV5_UpdateArticle(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Old Title", "Body", "Alice", now, now, sql.NullTime{}))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// 更新後のリロード
	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Updated Title", "Body", "Alice", now, now, sql.NullTime{}))

	r := router.New()
	body := bytes.NewBufferString(`{"title":"Updated Title"}`)

	// Act
	w := doRequest(r, http.MethodPut, "/api/v5/articles/1", body, map[string]string{
		"Content-Type": "application/json",
	})

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.Equal(t, true, resp["success"])
}

func TestV5_DeleteArticle(t *testing.T) {
	// Arrange
	mock := setupITMockDB(t)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM`).
		WillReturnRows(sqlmock.NewRows(itArticleColumns).
			AddRow(1, "Title", "Body", "Alice", now, now, sql.NullTime{}))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r := router.New()

	// Act
	w := doRequest(r, http.MethodDelete, "/api/v5/articles/1", nil, nil)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}
