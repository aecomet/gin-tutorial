// package v6_test は v6 Gin ハンドラーのユニットテスト。
// miniredis でRedisをモックし、実 TCP ポートで gRPC サーバーを起動する。
// GRPC_ADDR 環境変数でハンドラーの接続先をテスト用サーバーに向ける。
package v6_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	v6 "gin-tutorial/app/domain/v6"
	pb "gin-tutorial/app/grpc/pb"
	grpcserver "gin-tutorial/app/grpc/server"
	"gin-tutorial/app/middleware"
	rdb "gin-tutorial/app/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupV6 は miniredis + 実 TCP gRPC サーバーを起動し、
// GRPC_ADDR 環境変数でハンドラーの接続先をテスト用サーバーに向ける。
// Gin エンジンと gRPC クライアント（シード用）を返す。
func setupV6(t *testing.T) (*gin.Engine, pb.ArticleServiceClient) {
	t.Helper()

	mr := miniredis.RunT(t)
	t.Setenv("REDIS_HOST", mr.Host())
	t.Setenv("REDIS_PORT", mr.Port())
	require.NoError(t, rdb.Init())

	// :0 でランダムな空きポートを取得する。
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s := grpcserver.NewServer()
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(func() { s.GracefulStop() })

	// GRPC_ADDR でハンドラーの接続先をテスト用サーバーに向ける。
	t.Setenv("GRPC_ADDR", lis.Addr().String())

	// テスト用 gRPC クライアント（シード・検証用）
	seedConn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { seedConn.Close() })

	r := gin.New()
	r.Use(middleware.ErrorHandler())
	v6.RegisterRoutes(r.Group("/v6"))

	return r, pb.NewArticleServiceClient(seedConn)
}

func post(t *testing.T, r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- ListArticles ---

func TestV6_ListArticles_Empty(t *testing.T) {
	r, _ := setupV6(t)

	req := httptest.NewRequest(http.MethodGet, "/v6/articles", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.True(t, body["success"].(bool))
}

func TestV6_ListArticles_WithData(t *testing.T) {
	r, grpcClient := setupV6(t)

	_, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{Title: "記事1"})
	require.NoError(t, err)
	_, err = grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{Title: "記事2"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/v6/articles?page=1&per_page=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(2), data["total"])
}

// --- GetArticle ---

func TestV6_GetArticle_OK(t *testing.T) {
	r, grpcClient := setupV6(t)

	created, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "取得テスト", Author: "Alice",
	})
	require.NoError(t, err)
	id := created.Article.Id

	req := httptest.NewRequest(http.MethodGet, "/v6/articles/"+itoa(id), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	assert.Equal(t, "取得テスト", data["title"])
}

func TestV6_GetArticle_NotFound(t *testing.T) {
	r, _ := setupV6(t)

	req := httptest.NewRequest(http.MethodGet, "/v6/articles/9999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.False(t, body["success"].(bool))
}

func TestV6_GetArticle_InvalidID(t *testing.T) {
	r, _ := setupV6(t)

	req := httptest.NewRequest(http.MethodGet, "/v6/articles/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- CreateArticle ---

func TestV6_CreateArticle_OK(t *testing.T) {
	r, _ := setupV6(t)

	w := post(t, r, "/v6/articles", map[string]string{
		"title": "新規記事", "body": "本文", "author": "Bob",
	})

	assert.Equal(t, http.StatusCreated, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.True(t, body["success"].(bool))
	data := body["data"].(map[string]any)
	assert.Equal(t, "新規記事", data["title"])
}

func TestV6_CreateArticle_MissingTitle(t *testing.T) {
	r, _ := setupV6(t)

	w := post(t, r, "/v6/articles", map[string]string{"body": "本文のみ"})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- UpdateArticle ---

func TestV6_UpdateArticle_OK(t *testing.T) {
	r, grpcClient := setupV6(t)

	created, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "元タイトル", Author: "Carol",
	})
	require.NoError(t, err)
	id := created.Article.Id

	req := httptest.NewRequest(http.MethodPut, "/v6/articles/"+itoa(id), bytes.NewBufferString(`{"title":"更新タイトル"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	data := body["data"].(map[string]any)
	assert.Equal(t, "更新タイトル", data["title"])
	assert.Equal(t, "Carol", data["author"]) // author は変更されない
}

func TestV6_UpdateArticle_NotFound(t *testing.T) {
	r, _ := setupV6(t)

	req := httptest.NewRequest(http.MethodPut, "/v6/articles/9999", bytes.NewBufferString(`{"title":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- DeleteArticle ---

func TestV6_DeleteArticle_OK(t *testing.T) {
	r, grpcClient := setupV6(t)

	created, err := grpcClient.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "削除対象",
	})
	require.NoError(t, err)
	id := created.Article.Id

	req := httptest.NewRequest(http.MethodDelete, "/v6/articles/"+itoa(id), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestV6_DeleteArticle_NotFound(t *testing.T) {
	r, _ := setupV6(t)

	req := httptest.NewRequest(http.MethodDelete, "/v6/articles/9999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// itoa は uint32 を文字列に変換するヘルパー。
func itoa(n uint32) string {
	return fmt.Sprintf("%d", n)
}
