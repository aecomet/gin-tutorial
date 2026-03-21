package middleware

import "github.com/gin-gonic/gin"

// Version は Accept-Version ヘッダーからAPIバージョンを読み取り、コンテキストに設定する
func Version() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("Accept-Version")
		if version == "" {
			version = "v1"
		}
		c.Set("api_version", version)
		c.Next()
	}
}
