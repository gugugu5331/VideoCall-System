package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"meeting-system/shared/logger"
)

// getAllowedOrigins 从环境变量获取允许的源列表
func getAllowedOrigins() []string {
	// 从环境变量读取，多个源用逗号分隔
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if originsEnv == "" {
		// 默认允许本地开发环境
		return []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
		}
	}

	origins := strings.Split(originsEnv, ",")
	// 去除空格
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

// isOriginAllowed 检查源是否在白名单中
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			// 仅在开发环境允许 *
			if os.Getenv("GIN_MODE") == "debug" {
				return true
			}
			continue
		}
		if origin == allowed {
			return true
		}
	}
	return false
}

// CORS 跨域中间件（使用白名单机制）
func CORS() gin.HandlerFunc {
	allowedOrigins := getAllowedOrigins()

	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 检查源是否在白名单中
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Cache-Control, X-File-Name")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400") // 24小时
		} else if origin != "" {
			// 记录被拒绝的源
			logger.Warn("CORS request from unauthorized origin",
				logger.String("origin", origin),
				logger.String("path", c.Request.URL.Path))
		}

		// 处理预检请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// CORSWithOrigins 自定义源列表的CORS中间件
func CORSWithOrigins(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		// 检查源是否在白名单中
		if origin != "" && isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Cache-Control, X-File-Name")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		} else if origin != "" {
			logger.Warn("CORS request from unauthorized origin",
				logger.String("origin", origin),
				logger.String("path", c.Request.URL.Path))
		}

		// 处理预检请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
