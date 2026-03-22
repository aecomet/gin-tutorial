// package server_test は gRPC サーバー実装のユニットテスト。
// miniredis でRedisをモックし、bufconn でインプロセス gRPC 通信を実現する。
package server_test

import (
	"context"
	"net"
	"testing"

	pb "gin-tutorial/app/grpc/pb"
	grpcserver "gin-tutorial/app/grpc/server"
	rdb "gin-tutorial/app/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// setupTestServer は miniredis + bufconn を使ったテスト用 gRPC サーバーを起動し、
// クライアントを返す。t.Cleanup でサーバー・接続のクローズを自動登録する。
func setupTestServer(t *testing.T) pb.ArticleServiceClient {
	t.Helper()

	// miniredis でインメモリ Redis を起動する。
	// テスト終了時に自動停止される。
	mr := miniredis.RunT(t)

	// t.Setenv で環境変数を設定する。テスト終了時に自動的に元の値に戻る。
	t.Setenv("REDIS_HOST", mr.Host())
	t.Setenv("REDIS_PORT", mr.Port())
	require.NoError(t, rdb.Init())

	// bufconn はインプロセスの TCP 代替リスナー。
	// 実際のネットワークを使わないため高速でポート競合が起きない。
	lis := bufconn.Listen(bufSize)

	s := grpcserver.NewServer()
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(func() {
		s.Stop()
		lis.Close()
	})

	// bufconn に接続する際は WithContextDialer でカスタムダイヤラーを渡す。
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return pb.NewArticleServiceClient(conn)
}

// --- CreateArticle ---

func TestCreateArticle_OK(t *testing.T) {
	client := setupTestServer(t)

	resp, err := client.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title:  "テストタイトル",
		Body:   "テスト本文",
		Author: "Alice",
	})

	require.NoError(t, err)
	assert.NotZero(t, resp.Article.Id)
	assert.Equal(t, "テストタイトル", resp.Article.Title)
	assert.Equal(t, "テスト本文", resp.Article.Body)
	assert.Equal(t, "Alice", resp.Article.Author)
}

func TestCreateArticle_MissingTitle(t *testing.T) {
	client := setupTestServer(t)

	_, err := client.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "", // title は必須
	})

	// gRPC エラーコードの検証
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// --- GetArticle ---

func TestGetArticle_OK(t *testing.T) {
	client := setupTestServer(t)

	// 先に記事を作成してから取得する
	created, err := client.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "取得テスト", Author: "Bob",
	})
	require.NoError(t, err)

	resp, err := client.GetArticle(context.Background(), &pb.GetArticleRequest{
		Id: created.Article.Id,
	})

	require.NoError(t, err)
	assert.Equal(t, created.Article.Id, resp.Article.Id)
	assert.Equal(t, "取得テスト", resp.Article.Title)
}

func TestGetArticle_NotFound(t *testing.T) {
	client := setupTestServer(t)

	_, err := client.GetArticle(context.Background(), &pb.GetArticleRequest{Id: 9999})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// --- ListArticles ---

func TestListArticles_OK(t *testing.T) {
	client := setupTestServer(t)

	// 2件作成する
	for _, title := range []string{"記事A", "記事B"} {
		_, err := client.CreateArticle(context.Background(), &pb.CreateArticleRequest{Title: title})
		require.NoError(t, err)
	}

	resp, err := client.ListArticles(context.Background(), &pb.ListArticlesRequest{
		Page: 1, PerPage: 10,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(2), resp.Total)
	assert.Len(t, resp.Articles, 2)
}

func TestListArticles_Empty(t *testing.T) {
	client := setupTestServer(t)

	resp, err := client.ListArticles(context.Background(), &pb.ListArticlesRequest{
		Page: 1, PerPage: 10,
	})

	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Total)
	assert.Empty(t, resp.Articles)
}

// --- UpdateArticle ---

func TestUpdateArticle_OK(t *testing.T) {
	client := setupTestServer(t)

	created, err := client.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "元のタイトル", Author: "Alice",
	})
	require.NoError(t, err)

	resp, err := client.UpdateArticle(context.Background(), &pb.UpdateArticleRequest{
		Id:    created.Article.Id,
		Title: "更新後のタイトル",
	})

	require.NoError(t, err)
	assert.Equal(t, "更新後のタイトル", resp.Article.Title)
	assert.Equal(t, "Alice", resp.Article.Author) // Author は変更されない
}

func TestUpdateArticle_NotFound(t *testing.T) {
	client := setupTestServer(t)

	_, err := client.UpdateArticle(context.Background(), &pb.UpdateArticleRequest{
		Id: 9999, Title: "更新",
	})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// --- DeleteArticle ---

func TestDeleteArticle_OK(t *testing.T) {
	client := setupTestServer(t)

	created, err := client.CreateArticle(context.Background(), &pb.CreateArticleRequest{
		Title: "削除対象",
	})
	require.NoError(t, err)

	_, err = client.DeleteArticle(context.Background(), &pb.DeleteArticleRequest{
		Id: created.Article.Id,
	})
	require.NoError(t, err)

	// 削除後は NotFound になる
	_, err = client.GetArticle(context.Background(), &pb.GetArticleRequest{
		Id: created.Article.Id,
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteArticle_NotFound(t *testing.T) {
	client := setupTestServer(t)

	_, err := client.DeleteArticle(context.Background(), &pb.DeleteArticleRequest{Id: 9999})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
