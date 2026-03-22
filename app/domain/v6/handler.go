// package v6 は gRPC クライアントとして動作する Gin ハンドラー群を提供する。
// REST クライアントからのリクエストを受け取り、内部の gRPC サーバーに転送する。
// これは API Gateway パターンのデモ実装。
package v6

import (
	"net/http"
	"strconv"

	pb "gin-tutorial/app/grpc/pb"
	"gin-tutorial/app/handler"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// newGRPCClient は gRPC サーバーへの接続とクライアントスタブを生成する。
// insecure.NewCredentials() はTLSなし接続（開発・デモ用）。
// 本番では TLS クレデンシャルを使用すること。
func newGRPCClient() (pb.ArticleServiceClient, *grpc.ClientConn, error) {
	// grpc.NewClient でコネクションを確立する。
	// ダイヤルはバックグラウンドで行われ、最初のRPC呼び出し時に接続される。
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}
	// pb.NewArticleServiceClient でサービス定義から自動生成されたクライアントを生成する。
	// このクライアントを通じて proto で定義したメソッドを呼び出せる。
	return pb.NewArticleServiceClient(conn), conn, nil
}

// grpcCodeToHTTP は gRPC のエラーコードを HTTP ステータスコードに変換する。
func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.AlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// handleGRPCError は gRPC エラーを Gin のレスポンスに変換する。
func handleGRPCError(c *gin.Context, err error) {
	st, _ := status.FromError(err)
	handler.Fail(c, grpcCodeToHTTP(st.Code()), st.Code().String(), st.Message())
}

// ListArticles は GET /api/v6/articles を処理する。
// gRPC の ListArticles を呼び出して結果を返す。
func ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	client, conn, err := newGRPCClient()
	if err != nil {
		handler.Fail(c, http.StatusInternalServerError, handler.ErrInternal.Code, handler.ErrInternal.Message)
		return
	}
	defer conn.Close()

	resp, err := client.ListArticles(c.Request.Context(), &pb.ListArticlesRequest{
		Page:    int32(page),
		PerPage: int32(perPage),
	})
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	handler.OK(c, gin.H{
		"articles": resp.Articles,
		"total":    resp.Total,
		"page":     page,
		"per_page": perPage,
	})
}

// GetArticle は GET /api/v6/articles/:id を処理する。
func GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handler.Fail(c, http.StatusBadRequest, handler.ErrBadRequest.Code, handler.ErrBadRequest.Message)
		return
	}

	client, conn, grpcErr := newGRPCClient()
	if grpcErr != nil {
		handler.Fail(c, http.StatusInternalServerError, handler.ErrInternal.Code, handler.ErrInternal.Message)
		return
	}
	defer conn.Close()

	resp, grpcErr := client.GetArticle(c.Request.Context(), &pb.GetArticleRequest{Id: uint32(id)})
	if grpcErr != nil {
		handleGRPCError(c, grpcErr)
		return
	}

	handler.OK(c, resp.Article)
}

// CreateArticleInput はリクエストボディのバインド用構造体。
type CreateArticleInput struct {
	Title  string `json:"title"  binding:"required"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

// CreateArticle は POST /api/v6/articles を処理する。
func CreateArticle(c *gin.Context) {
	var input CreateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handler.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	client, conn, err := newGRPCClient()
	if err != nil {
		handler.Fail(c, http.StatusInternalServerError, handler.ErrInternal.Code, handler.ErrInternal.Message)
		return
	}
	defer conn.Close()

	resp, err := client.CreateArticle(c.Request.Context(), &pb.CreateArticleRequest{
		Title:  input.Title,
		Body:   input.Body,
		Author: input.Author,
	})
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": resp.Article})
}

// UpdateArticleInput はリクエストボディのバインド用構造体。
type UpdateArticleInput struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

// UpdateArticle は PUT /api/v6/articles/:id を処理する。
func UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handler.Fail(c, http.StatusBadRequest, handler.ErrBadRequest.Code, handler.ErrBadRequest.Message)
		return
	}

	var input UpdateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handler.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	client, conn, grpcErr := newGRPCClient()
	if grpcErr != nil {
		handler.Fail(c, http.StatusInternalServerError, handler.ErrInternal.Code, handler.ErrInternal.Message)
		return
	}
	defer conn.Close()

	resp, grpcErr := client.UpdateArticle(c.Request.Context(), &pb.UpdateArticleRequest{
		Id:     uint32(id),
		Title:  input.Title,
		Body:   input.Body,
		Author: input.Author,
	})
	if grpcErr != nil {
		handleGRPCError(c, grpcErr)
		return
	}

	handler.OK(c, resp.Article)
}

// DeleteArticle は DELETE /api/v6/articles/:id を処理する。
func DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		handler.Fail(c, http.StatusBadRequest, handler.ErrBadRequest.Code, handler.ErrBadRequest.Message)
		return
	}

	client, conn, grpcErr := newGRPCClient()
	if grpcErr != nil {
		handler.Fail(c, http.StatusInternalServerError, handler.ErrInternal.Code, handler.ErrInternal.Message)
		return
	}
	defer conn.Close()

	_, grpcErr = client.DeleteArticle(c.Request.Context(), &pb.DeleteArticleRequest{Id: uint32(id)})
	if grpcErr != nil {
		handleGRPCError(c, grpcErr)
		return
	}

	c.Status(http.StatusNoContent)
}
