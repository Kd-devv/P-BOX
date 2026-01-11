package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if query != "" {
			path = path + "?" + query
		}

		// 简单日志输出
		if status >= 400 {
			gin.DefaultErrorWriter.Write([]byte(
				time.Now().Format("2006/01/02 15:04:05") +
					" | " + c.Request.Method +
					" | " + path +
					" | " + latency.String() +
					"\n",
			))
		}
	}
}
