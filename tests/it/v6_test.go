package it

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "gin-tutorial/app/grpc/pb"
	grpcserver "gin-tutorial/app/grpc/server"
	rdb "gin-tutorial/app/redis"
	"gin-tutorial/app/router"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// setupV6IT は miniredis + 実 TCP gRPC サーバーを起動し、
// GRPC_ADDR 環境変数でハンドラーの接続先をテスト用サーバーに向ける。
// router.New() で全ルートを起動し、シード用 gRPC クライアントも返す。
func setupV6IT(t *testing.T) (pb.ArticleServiceClient, func(method, path string, body []byte) *httptest.ResponseRecorder) {
	t.Helper()

	mr := miniredis.RunT(t)
	t.Setenv("REDIS_HOST", mr.Host())
	t.Setenv("REDIS_PORT", mr.Port())
	require.NoError(t, rdb.Init())

	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s := grpcserver.NewServer()
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(func() { s.GracefulStop() })

	t.Setenv("GRPC_ADDR", lis.Addr().String())

	seedConn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { seedConn.Close() })

	r := router.New()
	do := func(method, path string, body []byte) *httptest.ResponseRecorder {
		var req *http.Request
		if body != nil {
			req = httptest.NewRequest(method, path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(method, path, nil)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	return pb.NewArticleServiceClient(seedConn), do
}

// jsonBody は map を JSON バイトスライスに変換するヘルパー。
func jsonBody(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// --- ListArticles ---

func TestV6IT_ListArticles_Empty(t *testing.T) {
	_, do := setupV6IT(t)

	w := do(http.MethodGet, "/api/v6/articles", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(0), data["total"])
}

func TestV6IT_ListArticles_WithData(t *testing.T) {
	grpcClient, do := setupV6IT(t)

	for _, title := range []string{"IT記事1", "IT記事2", "IT記事3"} {
		_, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{Title: title})
		require.NoError(t, err)
	}

	w := do(http.MethodGet, "/api/v6/articles?page=1&per_page=2", nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(3), data["total"])
	articles := data["articles"].([]any)
	assert.Len(t, articles, 2)
}

// --- CreateArticle ---

func TestV6IT_CreateArticle_OK(t *testing.T) {
	_, do := setupV6IT(t)

	w := do(http.MethodPost, "/api/v6/articles",
		jsonBody(t, map[string]string{"title": "IT新規記事", "body": "本文", "author": "Dave"}))

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseBody(t, w)
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]any)
	assert.Equal(t, "IT新規記事", data["title"])
	assert.NotZero(t, data["id"])
}

func TestV6IT_CreateArticle_MissingTitle(t *testing.T) {
	_, do := setupV6IT(t)

	w := do(http.MethodPost, "/api/v6/articles",
		jsonBody(t, map[string]string{"body": "タイトルなし"}))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseBody(t, w)
	assert.False(t, resp["success"].(bool))
}

// --- GetArticle ---

func TestV6IT_GetArticle_OK(t *testing.T) {
	grpcClient, do := setupV6IT(t)

	created, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "IT取得テスト", Author: "Eve",
	})
	require.NoError(t, err)

	w := do(http.MethodGet, fmt.Sprintf("/api/v6/articles/%d", created.Article.Id), nil)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]any)
	assert.Equal(t, "IT取得テスト", data["title"])
	assert.Equal(t, "Eve", data["author"])
}

func TestV6IT_GetArticle_NotFound(t *testing.T) {
	_, do := setupV6IT(t)

	w := do(http.MethodGet, "/api/v6/articles/99999", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resp := parseBody(t, w)
	assert.False(t, resp["success"].(bool))
}

// --- UpdateArticle ---

func TestV6IT_UpdateArticle_OK(t *testing.T) {
	grpcClient, do := setupV6IT(t)

	created, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "IT元タイトル", Author: "Frank",
	})
	require.NoError(t, err)

	w := do(http.MethodPut,
		fmt.Sprintf("/api/v6/articles/%d", created.Article.Id),
		jsonBody(t, map[string]string{"title": "IT更新タイトル"}))

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]any)
	assert.Equal(t, "IT更新タイトル", data["title"])
	assert.Equal(t, "Frank", data["author"]) // author は変更されない
}

func TestV6IT_UpdateArticle_NotFound(t *testing.T) {
	_, do := setupV6IT(t)

	w := do(http.MethodPut, "/api/v6/articles/99999",
		jsonBody(t, map[string]string{"title": "x"}))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- DeleteArticle ---

func TestV6IT_DeleteArticle_OK(t *testing.T) {
	grpcClient, do := setupV6IT(t)

	created, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "IT削除対象",
	})
	require.NoError(t, err)

	w := do(http.MethodDelete, fmt.Sprintf("/api/v6/articles/%d", created.Article.Id), nil)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// 削除後は 404 になる
	w = do(http.MethodGet, fmt.Sprintf("/api/v6/articles/%d", created.Article.Id), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestV6IT_DeleteArticle_NotFound(t *testing.T) {
	_, do := setupV6IT(t)

	w := do(http.MethodDelete, "/api/v6/articles/99999", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
