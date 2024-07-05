package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"meeting-system/shared/logger"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 记录请求日志
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("query", param.Request.URL.RawQuery),
			zap.String("ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("error", param.ErrorMessage),
		)
		return ""
	})
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 获取状态码
		status := c.Writer.Status()

		// 记录日志
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		}

		// 如果有错误，添加错误信息
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// 根据状态码选择日志级别
		if status >= 500 {
			logger.Error("HTTP Request", fields...)
		} else if status >= 400 {
			logger.Warn("HTTP Request", fields...)
		} else {
			logger.Info("HTTP Request", fields...)
		}
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		c.JSON(500, gin.H{
			"code":    500,
			"message": "Internal server error",
		})
	})
}
