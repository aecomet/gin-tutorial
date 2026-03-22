package v5

import (
	"errors"
	"net/http"
	"strconv"

	"gin-tutorial/app/db"
	"gin-tutorial/app/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes は v5 ドメインルートをまとめて登録する
func RegisterRoutes(rg *gin.RouterGroup) {
	articles := rg.Group("/articles")
	{
		articles.GET("", listArticles)
		articles.POST("", createArticle)
		articles.GET("/:id", getArticleByID)
		articles.PUT("/:id", updateArticle)
		articles.DELETE("/:id", deleteArticle)
	}
}

// listArticles は記事一覧を返す (ページネーション対応)
// GET /api/v5/articles?page=1&per_page=10
func listArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	var total int64
	var articles []Article

	if result := db.DB.Model(&Article{}).Count(&total); result.Error != nil {
		handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to count articles")
		return
	}
	if result := db.DB.Offset(offset).Limit(perPage).Find(&articles); result.Error != nil {
		handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch articles")
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, handler.Response{
		Success: true,
		Data:    articles,
		Meta: &handler.Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: totalPages,
		},
	})
}

// createArticle は新規記事を作成する
// POST /api/v5/articles
func createArticle(c *gin.Context) {
	var input CreateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handler.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	article := Article{
		Title:  input.Title,
		Body:   input.Body,
		Author: input.Author,
	}
	if result := db.DB.Create(&article); result.Error != nil {
		handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create article")
		return
	}

	c.JSON(http.StatusCreated, handler.Response{
		Success: true,
		Data:    article,
	})
}

// getArticleByID は指定IDの記事を返す
// GET /api/v5/articles/:id
func getArticleByID(c *gin.Context) {
	article, ok := findArticle(c)
	if !ok {
		return
	}
	handler.OK(c, article)
}

// updateArticle は指定IDの記事を更新する
// PUT /api/v5/articles/:id
func updateArticle(c *gin.Context) {
	article, ok := findArticle(c)
	if !ok {
		return
	}

	var input UpdateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		handler.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	updates := map[string]interface{}{}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Body != nil {
		updates["body"] = *input.Body
	}
	if input.Author != nil {
		updates["author"] = *input.Author
	}

	if result := db.DB.Model(&article).Updates(updates); result.Error != nil {
		handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update article")
		return
	}

	// 更新後の最新データをDBから取得して返す
	if result := db.DB.First(&article, article.ID); result.Error != nil {
		handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch updated article")
		return
	}

	handler.OK(c, article)
}

// deleteArticle は指定IDの記事を論理削除する
// DELETE /api/v5/articles/:id
func deleteArticle(c *gin.Context) {
	article, ok := findArticle(c)
	if !ok {
		return
	}

	if result := db.DB.Delete(&article); result.Error != nil {
		handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete article")
		return
	}

	c.Status(http.StatusNoContent)
}

// findArticle はパスパラメータ :id から記事を取得するヘルパー
func findArticle(c *gin.Context) (*Article, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		handler.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "id must be a positive integer")
		return nil, false
	}

	var article Article
	if result := db.DB.First(&article, id); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			handler.Fail(c, http.StatusNotFound, "NOT_FOUND", "article not found")
		} else {
			handler.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch article")
		}
		return nil, false
	}

	return &article, true
}
