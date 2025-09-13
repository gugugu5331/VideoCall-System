package main

import (
	"log"
	"net/http"
	"time"

	"video-conference-system/services/meeting/handlers"
	"video-conference-system/services/meeting/repository"
	"video-conference-system/services/meeting/service"
	"video-conference-system/shared/auth"
	"video-conference-system/shared/config"
	"video-conference-system/shared/database"
	"video-conference-system/shared/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 连接数据库
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 自动迁移
	if err := db.AutoMigrate(&models.Meeting{}, &models.MeetingParticipant{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 连接Redis
	redis, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// 创建JWT管理器
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireTime)

	// 创建仓储层
	meetingRepo := repository.NewMeetingRepository(db.DB)

	// 创建服务层
	meetingService := service.NewMeetingService(meetingRepo, redis.Client)

	// 创建处理器
	meetingHandler := handlers.NewMeetingHandler(meetingService)

	// 设置路由
	router := setupRoutes(meetingHandler, jwtManager)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("Meeting service starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(meetingHandler *handlers.MeetingHandler, jwtManager *auth.JWTManager) *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "meeting-service",
			"timestamp": time.Now(),
		})
	})

	// API路由
	api := router.Group("/api/v1")
	api.Use(authMiddleware(jwtManager))
	{
		meetings := api.Group("/meetings")
		{
			meetings.POST("", meetingHandler.CreateMeeting)
			meetings.GET("", meetingHandler.ListMeetings)
			meetings.GET("/:id", meetingHandler.GetMeeting)
			meetings.PUT("/:id", meetingHandler.UpdateMeeting)
			meetings.DELETE("/:id", meetingHandler.DeleteMeeting)
			meetings.POST("/:id/join", meetingHandler.JoinMeeting)
			meetings.POST("/:id/leave", meetingHandler.LeaveMeeting)
			meetings.GET("/:id/participants", meetingHandler.GetParticipants)
			meetings.POST("/:id/participants", meetingHandler.AddParticipant)
			meetings.DELETE("/:id/participants/:user_id", meetingHandler.RemoveParticipant)
		}
	}

	return router
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

// authMiddleware 认证中间件
func authMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := authHeader[len(bearerPrefix):]
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
