package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/shared/zmq"
	"meeting-system/user-service/handlers"
	"meeting-system/user-service/services"
)

var (
	configPath = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 初始化配置
	config.InitConfig(*configPath)
	cfg := config.GlobalConfig

	// 初始化日志
	if err := logger.InitLogger(logger.LogConfig{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxAge:     cfg.Log.MaxAge,
		MaxBackups: cfg.Log.MaxBackups,
		Compress:   cfg.Log.Compress,
	}); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting user service...")

	// 初始化数据库
	if err := database.InitPostgreSQL(cfg.Database); err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", logger.Error(err))
	}
	defer database.CloseDB()

	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Fatal("Failed to initialize Redis", logger.Error(err))
	}
	defer database.CloseRedis()

	// 自动迁移数据库表
	if err := database.AutoMigrate(&models.User{}); err != nil {
		logger.Fatal("Failed to migrate database", logger.Error(err))
	}

	// 初始化ZMQ客户端（可选，用于AI功能）
	if err := zmq.InitZMQ(cfg.ZMQ); err != nil {
		logger.Warn("Failed to initialize ZMQ client", logger.Error(err))
		// ZMQ初始化失败不影响用户服务启动
	}
	defer zmq.CloseZMQ()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 初始化服务
	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)

	// 注册路由
	setupRoutes(r, userHandler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 优雅关闭
	go func() {
		logger.Info("User service started", logger.String("address", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down user service...")

	// 优雅关闭服务器
	if err := server.Close(); err != nil {
		logger.Error("Server forced to shutdown", logger.Error(err))
	}

	logger.Info("User service stopped")
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine, userHandler *handlers.UserHandler) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "user-service",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API版本1
	v1 := r.Group("/api/v1")
	{
		// 公开接口（不需要认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/forgot-password", userHandler.ForgotPassword)
			auth.POST("/reset-password", userHandler.ResetPassword)
		}

		// 需要认证的接口
		protected := v1.Group("/users")
		protected.Use(middleware.JWTAuth())
		{
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)
			protected.POST("/change-password", userHandler.ChangePassword)
			protected.POST("/upload-avatar", userHandler.UploadAvatar)
			protected.DELETE("/account", userHandler.DeleteAccount)
		}

		// 用户管理接口（管理员）
		admin := v1.Group("/admin/users")
		admin.Use(middleware.JWTAuth()) // TODO: 添加管理员权限检查
		{
			admin.GET("", userHandler.ListUsers)
			admin.GET("/:id", userHandler.GetUser)
			admin.PUT("/:id", userHandler.UpdateUser)
			admin.DELETE("/:id", userHandler.DeleteUser)
			admin.POST("/:id/ban", userHandler.BanUser)
			admin.POST("/:id/unban", userHandler.UnbanUser)
		}
	}
}
