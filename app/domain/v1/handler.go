package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes は v1 デモ用ルートをまとめて登録する
func RegisterRoutes(rg *gin.RouterGroup) {
	// クエリパラメータ
	rg.GET("/welcome", welcome)

	// POSTフォームデータ
	rg.POST("/form_post", formPost)

	// クエリパラメータ + POSTフォームデータの組み合わせ
	rg.POST("/post", post)

	// クエリマップ + フォームマップ
	rg.POST("/form_map", formMap)

	// multipart/form-data（ファイルアップロード）
	rg.POST("/multipart", multipartUpload)

	// リミット・オフセットによるページネーション
	rg.GET("/articles", listWithOffset)

	// カーソルベースのページネーション
	rg.GET("/events", listWithCursor)
}

func welcome(c *gin.Context) {
	firstname := c.DefaultQuery("firstname", "Guest")
	lastname := c.Query("lastname")
	c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
}

func formPost(c *gin.Context) {
	message := c.PostForm("message")
	nick := c.DefaultPostForm("nick", "anonymous")
	c.JSON(http.StatusOK, gin.H{
		"status":  "posted",
		"message": message,
		"nick":    nick,
	})
}

func post(c *gin.Context) {
	id := c.Query("id")
	page := c.DefaultQuery("page", "0")
	name := c.PostForm("name")
	message := c.PostForm("message")
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"page":    page,
		"name":    name,
		"message": message,
	})
}

// formMap はクエリマップ + フォームマップのサンプル
// 例: POST /api/v1/form_map?ids[first]=1&ids[second]=2 (Body: names[a]=foo&names[b]=bar)
func formMap(c *gin.Context) {
	ids := c.QueryMap("ids")
	names := c.PostFormMap("names")
	c.JSON(http.StatusOK, gin.H{
		"ids":   ids,
		"names": names,
	})
}

func multipartUpload(c *gin.Context) {
	message := c.PostForm("message")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  message,
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// listWithOffset はリミット・オフセットによるページネーションのサンプル
// GET /api/v1/articles?limit=20&offset=0
func listWithOffset(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 100 {
		limit = 100
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []gin.H{},
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  0,
		},
	})
}

// listWithCursor はカーソルベースのページネーションのサンプル
// GET /api/v1/events?cursor=<last_id>&limit=20
func listWithCursor(c *gin.Context) {
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}
	_ = cursor
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"data":        []gin.H{},
		"next_cursor": "",
	})
}
