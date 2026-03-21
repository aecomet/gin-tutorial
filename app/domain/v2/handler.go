package v2

import (
	"net/http"

	"gin-tutorial/app/handler"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes は v2 ドメインルートをまとめて登録する
func RegisterRoutes(rg *gin.RouterGroup) {
	registerUserRoutes(rg)
	registerProductRoutes(rg)
	registerOrderRoutes(rg)
	registerItemRoutes(rg)
}

// --- users ---

func registerUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", listUsers)
		users.POST("", createUser)
		users.GET("/:id", getUserByID)
		users.PUT("/:id", updateUser)
		users.DELETE("/:id", deleteUser)
	}
}

func listUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "list_users"})
}

func createUser(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"action": "create_user"})
}

func getUserByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "get_user"})
}

func updateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "update_user"})
}

func deleteUser(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// --- products ---

func registerProductRoutes(rg *gin.RouterGroup) {
	products := rg.Group("/products")
	{
		products.GET("", listProducts)
	}
}

// listProducts はクエリパラメータによるフィルタリング・ソートのサンプル
// GET /api/v2/products?category=electronics&min_price=10&sort=price&order=asc
func listProducts(c *gin.Context) {
	category := c.Query("category")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	// インジェクション防止のため許可リストで検証
	allowed := map[string]bool{"created_at": true, "price": true, "name": true}
	if !allowed[sortBy] {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	_ = category
	_ = minPrice
	_ = maxPrice
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []gin.H{},
		"filters": gin.H{
			"category":  category,
			"min_price": minPrice,
			"max_price": maxPrice,
			"sort":      sortBy,
			"order":     order,
		},
	})
}

// --- orders ---

func registerOrderRoutes(rg *gin.RouterGroup) {
	orders := rg.Group("/orders")
	{
		orders.GET("", listOrders)
		orders.POST("", createOrder)
		orders.GET("/:id", getOrderByID)
	}
}

func listOrders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "list_orders"})
}

func createOrder(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"action": "create_order"})
}

func getOrderByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"action": "get_order"})
}

// --- items ---

func registerItemRoutes(rg *gin.RouterGroup) {
	items := rg.Group("/items")
	{
		items.GET("/:id", getItemByID)
	}
}

// getItemByID はカスタムエラー型を使ったエラーハンドリングのサンプル
// id=0 はNOT_FOUNDエラーを返す
func getItemByID(c *gin.Context) {
	id := c.Param("id")
	if id == "0" {
		_ = c.Error(handler.ErrNotFound)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"id": id}})
}
