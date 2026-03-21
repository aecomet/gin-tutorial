package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck はヘルスチェックのサンプル
// Accept-Version ヘッダーによるバージョン切り替えも兼ねる（v1 or v2）
func HealthCheck(c *gin.Context) {
	version := c.GetString("api_version")
	switch version {
	case "v2":
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "v2"})
	default:
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
