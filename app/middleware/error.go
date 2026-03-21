package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"gin-tutorial/app/handler"

	"github.com/gin-gonic/gin"
)

// ErrorHandler は c.Error() でセットされたエラーを構造化レスポンスに変換するミドルウェア
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *handler.AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{
				"success": false,
				"error":   gin.H{"code": appErr.Code, "message": appErr.Message},
			})
		} else {
			slog.Error("unhandled error",
				slog.String("error", err.Error()),
				slog.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   gin.H{"code": "INTERNAL", "message": "an unexpected error occurred"},
			})
		}
	}
}
