package router

import (
	v1 "gin-tutorial/app/domain/v1"
	v2 "gin-tutorial/app/domain/v2"
	v3 "gin-tutorial/app/domain/v3"
	v4 "gin-tutorial/app/domain/v4"
	v5 "gin-tutorial/app/domain/v5"
	"gin-tutorial/app/handler"
	"gin-tutorial/app/middleware"

	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.Version())

	api := r.Group("/api")

	// ヘルスチェック（ヘッダーバージョニングのサンプル: Accept-Version: v1 or v2）
	api.GET("/healthcheck", handler.HealthCheck)

	v1.RegisterRoutes(api.Group("/v1"))
	v2.RegisterRoutes(api.Group("/v2"))
	v3.RegisterRoutes(api.Group("/v3"))
	v4.RegisterRoutes(api.Group("/v4"))
	v5.RegisterRoutes(api.Group("/v5"))

	// 登録済みルート一覧
	api.GET("/routes", func(c *gin.Context) {
		routes := r.Routes()
		list := make([]gin.H, 0, len(routes))
		for _, route := range routes {
			list = append(list, gin.H{
				"method": route.Method,
				"path":   route.Path,
			})
		}
		c.JSON(200, list)
	})

	return r
}
