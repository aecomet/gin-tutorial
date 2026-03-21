package router

import (
	"gin-tutorial/handler"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.Default()

	// パスパラメータ
	r.GET("/user/:name", handler.GetUser)

	// ワイルドカードパスパラメータ
	r.GET("/user/:name/*action", handler.GetUserAction)

	// クエリパラメータ
	r.GET("/welcome", handler.Welcome)

	// POSTフォームデータ
	r.POST("/form_post", handler.FormPost)

	// クエリパラメータ + POSTフォームデータの組み合わせ
	r.POST("/post", handler.Post)

	// multipart/form-data（ファイルアップロード）
	r.POST("/multipart", handler.MultipartForm)

	// クエリマップ + フォームマップ
	r.POST("/form_map", handler.FormMap)

	return r
}
