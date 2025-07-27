package main

import (
	"log"
	"os"

	"videocall-backend/config"
	"videocall-backend/database"
	"videocall-backend/handlers"
	"videocall-backend/middleware"
	"videocall-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title 音视频通话系统 API
// @version 1.0
// @description 基于深度学习的音视频通话系统后端API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 初始化配置
	cfg := config.Load()

	// 初始化数据库连接
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化Redis连接
	redisClient, err := database.InitRedis(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// 设置Gin模式
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 添加基础中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// 初始化处理器
	handlers.InitHandlers(db, redisClient, cfg)

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Basic backend server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 