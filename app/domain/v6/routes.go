package v6

import "github.com/gin-gonic/gin"

// RegisterRoutes は v6 のエンドポイントを登録する。
// v6 では Gin ハンドラーが内部で gRPC クライアントとして動作し、
// gRPC サーバーに処理を委譲する（API Gateway パターンのデモ）。
func RegisterRoutes(rg *gin.RouterGroup) {
	articles := rg.Group("/articles")
	articles.GET("", ListArticles)
	articles.POST("", CreateArticle)
	articles.GET("/:id", GetArticle)
	articles.PUT("/:id", UpdateArticle)
	articles.DELETE("/:id", DeleteArticle)
}
