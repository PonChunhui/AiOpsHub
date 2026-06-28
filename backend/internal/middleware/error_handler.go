package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/aiops/AiOpsHub/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(fmt.Sprintf("Panic recovered: %v\n%s", err, debug.Stack()))

				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal Server Error",
					"message": "An unexpected error occurred",
					"code":    500,
				})
				c.Abort()
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			logger.Error(fmt.Sprintf("Request error: %v", err))

			status := c.Writer.Status()
			if status == 0 {
				status = http.StatusInternalServerError
			}

			code := status
			message := err.Error()

			if status >= 400 && status < 500 {
				message = "Bad Request"
			} else if status >= 500 {
				message = "Internal Server Error"
			}

			c.JSON(status, gin.H{
				"error":   http.StatusText(status),
				"message": message,
				"code":    code,
				"path":    c.Request.URL.Path,
			})
		}
	}
}

func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": fmt.Sprintf("Path %s not found", c.Request.URL.Path),
			"code":    404,
		})
	}
}

func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":   "Method Not Allowed",
			"message": fmt.Sprintf("Method %s not allowed", c.Request.Method),
			"code":    405,
		})
	}
}
