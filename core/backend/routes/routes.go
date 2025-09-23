package routes

import (
	"videocall-backend/handlers"
	"videocall-backend/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置路由
func SetupRoutes(r *gin.Engine) {
	// 创建处理器实例
	userHandler := handlers.NewUserHandler()

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关路由（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// 需要认证的路由
		authenticated := v1.Group("/")
		authenticated.Use(middleware.AuthMiddleware())
		{
			// 用户相关
			user := authenticated.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.PUT("/profile", userHandler.UpdateProfile)
			}

			// 用户搜索
			users := authenticated.Group("/users")
			{
				users.GET("/search", userHandler.SearchUsers)
			}

			// 通话相关
			calls := authenticated.Group("/calls")
			{
				calls.POST("/start", handlers.StartCall)
				calls.POST("/end", handlers.EndCall)
				calls.GET("/history", handlers.GetCallHistory)
				calls.GET("/:id", handlers.GetCallDetails)
				calls.GET("/active", handlers.GetActiveCalls)
			}

			// 安全检测相关
			security := authenticated.Group("/security")
			{
				security.POST("/detect", handlers.TriggerDetection)
				security.GET("/status/:callId", handlers.GetDetectionStatus)
				security.GET("/history", handlers.GetDetectionHistory)
			}
		}
	}

	// WebSocket 路由（暂时不需要认证，用于测试）
	r.GET("/ws/call/:callId", handlers.WebSocketHandler)

	// 通知WebSocket路由
	r.GET("/ws/notifications", handlers.NotificationWebSocketHandler)

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "VideoCall Backend is running",
		})
	})

	// 根路径
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "VideoCall Backend API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
		})
	})
}
