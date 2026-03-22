// package server は gRPC サーバーの起動と ArticleService の実装を提供する。
// データは Redis に保存し、複数のサービスインスタンス間で共有できる構成のデモ。
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"

	pb "gin-tutorial/app/grpc/pb"
	rdb "gin-tutorial/app/redis"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Start は指定ポートで gRPC サーバーをブロッキングで起動する。
// main.go から goroutine で呼び出すことを想定している。
func Start(port string) error {
	// TCP リスナーを作成する。gRPC はデフォルトで HTTP/2 を使用する。
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	// grpc.NewServer でサーバーインスタンスを生成する。
	// オプションでインターセプター（ミドルウェア相当）を追加できる。
	s := grpc.NewServer()

	// 生成されたコードの RegisterArticleServiceServer で
	// サービス実装をサーバーに登録する。
	pb.RegisterArticleServiceServer(s, &articleServiceServer{})

	slog.Info("gRPC server starting", slog.String("addr", ":"+port))

	// Serve はリスナーを受け取り、接続をブロッキングで処理し続ける。
	return s.Serve(lis)
}

// articleServiceServer は pb.ArticleServiceServer インターフェースの実装。
// Redis をストレージとして使用する。
type articleServiceServer struct {
	// UnimplementedArticleServiceServer を埋め込むことで、
	// 未実装のメソッドに対してデフォルトの「未実装」エラーを返す。
	// これにより proto に新しいメソッドが追加されてもコンパイルエラーにならない。
	pb.UnimplementedArticleServiceServer
}

// Redis キーの定数定義
const (
	keyPrefix  = "grpc:article:"  // 記事ハッシュのキープレフィックス（例: grpc:article:1）
	keyIDSet   = "grpc:article:ids" // 全記事IDを管理する Redis Set
	keyCounter = "grpc:article:seq" // ID採番用カウンター
)

// articleKey は記事IDからRedisキーを生成する
func articleKey(id uint32) string {
	return fmt.Sprintf("%s%d", keyPrefix, id)
}

// toArticle は Redis に保存した JSON を pb.Article に変換する
func toArticle(data string) (*pb.Article, error) {
	var a pb.Article
	if err := json.Unmarshal([]byte(data), &a); err != nil {
		return nil, err
	}
	return &a, nil
}

// ListArticles は記事一覧を返す。
// context.Context は締め切り・キャンセル・タイムアウトの伝播に使う（gRPCの慣習）。
func (s *articleServiceServer) ListArticles(ctx context.Context, req *pb.ListArticlesRequest) (*pb.ListArticlesResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	perPage := req.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	rc := rdb.Client()

	// RedisのSetから全IDを取得する
	ids, err := rc.SMembers(ctx, keyIDSet).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list articles: %v", err)
	}

	total := int32(len(ids))
	start := int((page - 1) * perPage)
	if start >= len(ids) {
		return &pb.ListArticlesResponse{Articles: []*pb.Article{}, Total: total}, nil
	}
	end := start + int(perPage)
	if end > len(ids) {
		end = len(ids)
	}

	articles := make([]*pb.Article, 0, end-start)
	for _, id := range ids[start:end] {
		data, err := rc.Get(ctx, keyPrefix+id).Result()
		if err != nil {
			continue
		}
		a, err := toArticle(data)
		if err != nil {
			continue
		}
		articles = append(articles, a)
	}

	return &pb.ListArticlesResponse{Articles: articles, Total: total}, nil
}

// GetArticle は指定IDの記事を返す。
// 存在しない場合は gRPC の status パッケージでエラーコードを付けて返す。
// HTTP の 404 に相当するのが codes.NotFound。
func (s *articleServiceServer) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	rc := rdb.Client()

	data, err := rc.Get(ctx, articleKey(req.Id)).Result()
	if err != nil {
		// status.Errorf でエラーコードとメッセージを付けて返す。
		// クライアント側では status.Code(err) でコードを取り出せる。
		return nil, status.Errorf(codes.NotFound, "article not found: id=%d", req.Id)
	}

	a, err := toArticle(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to decode article: %v", err)
	}
	return &pb.GetArticleResponse{Article: a}, nil
}

// CreateArticle は新しい記事を作成する。
func (s *articleServiceServer) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	rc := rdb.Client()

	// INCR でスレッドセーフに採番する
	id64, err := rc.Incr(ctx, keyCounter).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate id: %v", err)
	}
	id := uint32(id64)

	a := &pb.Article{
		Id:     id,
		Title:  req.Title,
		Body:   req.Body,
		Author: req.Author,
	}

	data, err := json.Marshal(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to encode article: %v", err)
	}

	// パイプラインで SET と SADD をまとめて実行する（往復を1回に削減）
	pipe := rc.Pipeline()
	pipe.Set(ctx, articleKey(id), string(data), 0)
	pipe.SAdd(ctx, keyIDSet, fmt.Sprintf("%d", id))
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save article: %v", err)
	}

	return &pb.CreateArticleResponse{Article: a}, nil
}

// UpdateArticle は指定IDの記事を更新する。
func (s *articleServiceServer) UpdateArticle(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	rc := rdb.Client()

	data, err := rc.Get(ctx, articleKey(req.Id)).Result()
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "article not found: id=%d", req.Id)
	}

	a, err := toArticle(data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to decode article: %v", err)
	}

	if req.Title != "" {
		a.Title = req.Title
	}
	if req.Body != "" {
		a.Body = req.Body
	}
	if req.Author != "" {
		a.Author = req.Author
	}

	updated, err := json.Marshal(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to encode article: %v", err)
	}
	if err := rc.Set(ctx, articleKey(req.Id), string(updated), 0).Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update article: %v", err)
	}

	return &pb.UpdateArticleResponse{Article: a}, nil
}

// DeleteArticle は指定IDの記事を削除する。
func (s *articleServiceServer) DeleteArticle(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleResponse, error) {
	rc := rdb.Client()

	if rc.Exists(ctx, articleKey(req.Id)).Val() == 0 {
		return nil, status.Errorf(codes.NotFound, "article not found: id=%d", req.Id)
	}

	pipe := rc.Pipeline()
	pipe.Del(ctx, articleKey(req.Id))
	pipe.SRem(ctx, keyIDSet, fmt.Sprintf("%d", req.Id))
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete article: %v", err)
	}

	return &pb.DeleteArticleResponse{}, nil
}
