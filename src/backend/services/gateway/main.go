package main

import (
	"log"
	"net/http"
	"time"

	"video-conference-system/shared/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置路由
	router := setupRoutes()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("Gateway service starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes() *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "gateway-service",
			"timestamp": time.Now(),
		})
	})

	// API路由代理
	api := router.Group("/api/v1")
	{
		// 用户服务路由
		users := api.Group("/users")
		{
			users.Any("/*path", proxyToService("user-service", "8080"))
		}

		// 会议服务路由
		meetings := api.Group("/meetings")
		{
			meetings.Any("/*path", proxyToService("meeting-service", "8080"))
		}

		// 检测服务路由
		detection := api.Group("/detection")
		{
			detection.Any("/*path", proxyToService("detection-service", "8080"))
		}

		// 记录服务路由
		records := api.Group("/records")
		{
			records.Any("/*path", proxyToService("record-service", "8080"))
		}
	}

	// 信令服务WebSocket代理
	router.Any("/signaling/*path", proxyToService("signaling-service", "8080"))

	return router
}

// proxyToService 代理请求到指定服务
func proxyToService(serviceName, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单的代理实现
		c.JSON(http.StatusOK, gin.H{
			"message": "Gateway proxy to " + serviceName,
			"path":    c.Request.URL.Path,
		})
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
