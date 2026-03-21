package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Welcome(c *gin.Context) {
	firstname := c.DefaultQuery("firstname", "Guest")
	lastname := c.Query("lastname")
	c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
}

func FormPost(c *gin.Context) {
	message := c.PostForm("message")
	nick := c.DefaultPostForm("nick", "anonymous")
	c.JSON(http.StatusOK, gin.H{
		"status":  "posted",
		"message": message,
		"nick":    nick,
	})
}

func Post(c *gin.Context) {
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

// multipart/form-data および application/x-www-form-urlencoded の両方に対応
func MultipartForm(c *gin.Context) {
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

// クエリマップ + フォームマップ
// 例: GET /form_map?ids[first]=1&ids[second]=2 (Body: names[a]=foo&names[b]=bar)
func FormMap(c *gin.Context) {
	ids := c.QueryMap("ids")
	names := c.PostFormMap("names")
	c.JSON(http.StatusOK, gin.H{
		"ids":   ids,
		"names": names,
	})
}
