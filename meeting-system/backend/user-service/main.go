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
	"meeting-system/user-service/handlers"
	"meeting-system/user-service/services"
)

var (
	configPath = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 初始化配置
	log.Println("Initializing configuration...")
	config.InitConfig(*configPath)
	cfg := config.GlobalConfig
	log.Println("Configuration initialized successfully")

	// 初始化日志
	log.Println("Initializing logger...")
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
	log.Println("Logger initialized successfully")

	logger.Info("Starting user service...")
	log.Println("Starting user service...")

	// 初始化数据库
	logger.Info("Initializing database...")
	log.Println("Initializing database...")
	if err := database.InitDB(cfg.Database); err != nil {
		logger.Fatal("Failed to initialize database: " + err.Error())
	}
	defer database.CloseDB()
	logger.Info("Database initialized successfully")
	log.Println("Database initialized successfully")

	logger.Info("Initializing Redis...")
	log.Println("Initializing Redis...")
	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Warn("Failed to initialize Redis: " + err.Error())
		log.Println("Failed to initialize Redis:", err.Error())
		// Redis初始化失败不影响压力测试
	} else {
		defer database.CloseRedis()
		logger.Info("Redis initialized successfully")
		log.Println("Redis initialized successfully")
	}

	// 强制迁移User表以确保表存在
	logger.Info("Migrating User table...")
	log.Println("Migrating User table...")
	if err := database.AutoMigrate(&models.User{}); err != nil {
		logger.Fatal("Failed to migrate database: " + err.Error())
	}
	logger.Info("User table migration completed")
	log.Println("User table migration completed")

	// 跳过ZMQ初始化（压力测试时不需要）
	logger.Info("Skipping ZMQ initialization for stress testing")
	log.Println("Skipping ZMQ initialization for stress testing")

	// 设置Gin模式
	logger.Info("Setting up HTTP server...")
	log.Println("Setting up HTTP server...")
	gin.SetMode(cfg.Server.Mode)
	log.Println("Gin mode set to:", cfg.Server.Mode)

	// 创建Gin引擎
	log.Println("Creating Gin engine...")
	r := gin.New()
	log.Println("Gin engine created")

	// 添加中间件
	log.Println("Adding middleware...")
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	log.Println("Middleware added")

	// 初始化服务
	logger.Info("Initializing user service components...")
	log.Println("Initializing user service components...")
	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)
	logger.Info("User service components initialized")
	log.Println("User service components initialized")

	// 注册路由
	logger.Info("Setting up routes...")
	log.Println("Setting up routes...")
	setupRoutes(r, userHandler)
	logger.Info("Routes configured")
	log.Println("Routes configured")

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Starting HTTP server on " + addr)
	log.Println("Starting HTTP server on", addr)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}
	log.Println("HTTP server configured")

	// 优雅关闭
	go func() {
		logger.Info("User service HTTP server starting...")
		log.Println("User service HTTP server starting...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()
	log.Println("HTTP server goroutine started")

	logger.Info("User service started successfully and listening on " + addr)
	log.Println("User service started successfully and listening on", addr)

	// 等待中断信号
	log.Println("Waiting for interrupt signal...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down user service...")

	// 优雅关闭服务器
	if err := server.Close(); err != nil {
		logger.Error("Server forced to shutdown: " + err.Error())
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
