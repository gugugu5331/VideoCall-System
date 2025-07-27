package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"videocall-backend/config"
	"videocall-backend/database"
	"videocall-backend/handlers"
	"videocall-backend/middleware"
	"videocall-backend/routes"
	"videocall-backend/utils"

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

// 全局并发控制
var (
	// 限制同时处理的请求数量
	requestLimiter *middleware.ConcurrencyLimiter
	// 全局上下文
	globalCtx context.Context
	// 取消函数
	cancelFunc context.CancelFunc
)

func main() {
	// 设置最大CPU核心数
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 创建全局上下文
	globalCtx, cancelFunc = context.WithCancel(context.Background())
	defer cancelFunc()

	// 初始化并发控制
	maxConcurrentRequests := int64(1000) // 最大并发请求数
	requestLimiter = middleware.NewConcurrencyLimiter(maxConcurrentRequests)

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
	r := gin.New() // 使用gin.New()而不是gin.Default()以获得更好的性能

	// 添加自定义中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.RateLimit(redisClient))           // 添加限流中间件
	r.Use(middleware.ConcurrencyLimit(requestLimiter)) // 添加并发限制中间件
	r.Use(middleware.Metrics())                        // 添加监控中间件

	// 初始化处理器
	handlers.InitHandlers(db, redisClient, cfg)

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 启动监控协程
	go utils.StartMetricsServer(":8080")

	// 启动健康检查协程
	go utils.StartHealthCheck(globalCtx, db, redisClient)

	// 启动连接池监控协程
	go utils.MonitorConnectionPools(globalCtx, db, redisClient)

	log.Printf("Server starting on port %s with %d CPU cores", port, runtime.NumCPU())
	log.Printf("Max concurrent requests: %d", maxConcurrentRequests)

	// 优雅关闭
	go func() {
		<-globalCtx.Done()
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
