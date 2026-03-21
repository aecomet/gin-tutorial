package v3

import (
	"net/http"

	"gin-tutorial/app/handler"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes は v3 モデルバインディング・バリデーションのデモルートを登録する
func RegisterRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		// JSON バインディング + バリデーション
		users.POST("", createUser)
		// URI バインディング + バリデーション
		users.GET("/:id", getUser)
	}
	// クエリパラメータ バインディング + バリデーション
	rg.GET("/search", search)
	// フォーム バインディング + バリデーション
	rg.POST("/login", login)
	// デフォルト値付き バインディング
	rg.GET("/posts", listPosts)
	// ヘッダー バインディング
	rg.GET("/me", getMe)
}

// CreateUserRequest は JSON バインディングのサンプル構造体
type CreateUserRequest struct {
	Name     string `json:"name"     binding:"required,min=2,max=50"`
	Email    string `json:"email"    binding:"required,email"`
	Age      int    `json:"age"      binding:"required,gte=1,lte=130"`
	Password string `json:"password" binding:"required,min=8"`
}

// GetUserUriRequest は URI バインディングのサンプル構造体
type GetUserUriRequest struct {
	ID int `uri:"id" binding:"required,gt=0"`
}

// SearchQuery はクエリパラメータ バインディングのサンプル構造体
type SearchQuery struct {
	Keyword string `form:"keyword"  binding:"required"`
	Page    int    `form:"page"     binding:"omitempty,gte=1"`
	PerPage int    `form:"per_page" binding:"omitempty,gte=1,lte=100"`
}

// LoginRequest はフォーム バインディングのサンプル構造体
type LoginRequest struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required,min=8"`
}

// createUser は JSON バインディングと入力バリデーションのサンプル
// POST /api/v3/users
// Body: {"name":"Alice","email":"alice@example.com","age":30,"password":"secret123"}
func createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	handler.OK(c, gin.H{
		"name":  req.Name,
		"email": req.Email,
		"age":   req.Age,
	})
}

// getUser は URI バインディングと入力バリデーションのサンプル
// GET /api/v3/users/:id  ※ id は gt=0 の整数
func getUser(c *gin.Context) {
	var uri GetUserUriRequest
	if err := c.ShouldBindUri(&uri); err != nil {
		handler.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	handler.OK(c, gin.H{"id": uri.ID})
}

// search はクエリパラメータ バインディングと入力バリデーションのサンプル
// GET /api/v3/search?keyword=gin&page=1&per_page=20
func search(c *gin.Context) {
	var q SearchQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		handler.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PerPage == 0 {
		q.PerPage = 20
	}
	handler.OK(c, gin.H{
		"keyword":  q.Keyword,
		"page":     q.Page,
		"per_page": q.PerPage,
	})
}

// login はフォーム バインディングと入力バリデーションのサンプル
// POST /api/v3/login
// Body (application/x-www-form-urlencoded): username=alice&password=secret123
func login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		handler.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	handler.OK(c, gin.H{"username": req.Username, "message": "login successful"})
}

// ListPostsQuery はデフォルト値付きクエリバインディングのサンプル構造体
// default タグで未指定時の値を定義する
type ListPostsQuery struct {
	Page     int    `form:"page"     binding:"omitempty,gte=1"             default:"1"`
	PerPage  int    `form:"per_page" binding:"omitempty,gte=1,lte=100"     default:"20"`
	Sort     string `form:"sort"     binding:"omitempty,oneof=created_at updated_at title" default:"created_at"`
	Order    string `form:"order"    binding:"omitempty,oneof=asc desc"    default:"desc"`
	Status   string `form:"status"   binding:"omitempty,oneof=draft published archived" default:"published"`
}

// listPosts はデフォルト値付き binding model のサンプル
// パラメータを省略した場合、default タグの値が適用される
// GET /api/v3/posts
// GET /api/v3/posts?page=2&per_page=5&sort=title&order=asc&status=draft
func listPosts(c *gin.Context) {
	var q ListPostsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		handler.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	handler.OK(c, gin.H{
		"data": []gin.H{},
		"meta": gin.H{
			"page":     q.Page,
			"per_page": q.PerPage,
			"sort":     q.Sort,
			"order":    q.Order,
			"status":   q.Status,
		},
	})
}

// MeHeader はヘッダー バインディングのサンプル構造体
// header タグでHTTPリクエストヘッダーをフィールドにマッピングする
type MeHeader struct {
	Authorization string `header:"Authorization"  binding:"required"`
	XRequestID    string `header:"X-Request-Id"   binding:"required,uuid4"`
	AcceptLang    string `header:"Accept-Language" binding:"omitempty"`
}

// getMe はヘッダー バインディングと入力バリデーションのサンプル
// GET /api/v3/me
// Headers: Authorization: Bearer <token>, X-Request-Id: <uuid>
func getMe(c *gin.Context) {
	var h MeHeader
	if err := c.ShouldBindHeader(&h); err != nil {
		handler.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	handler.OK(c, gin.H{
		"authorization": h.Authorization,
		"request_id":    h.XRequestID,
		"accept_lang":   h.AcceptLang,
	})
}
