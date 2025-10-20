package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"meeting-system/ai-inference-service/handlers"
	"meeting-system/ai-inference-service/services"
	pb "meeting-system/shared/grpc"
	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/discovery"
	"meeting-system/shared/logger"
	"meeting-system/shared/metrics"
	"meeting-system/shared/middleware"
	"meeting-system/shared/queue"
	"meeting-system/shared/tracing"
	"meeting-system/shared/zmq"

	"net"
)

var (
	configPath = flag.String("config", "config/ai-inference-service.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 初始化配置
	fmt.Println("Initializing configuration...")
	config.InitConfig(*configPath)
	cfg := config.GlobalConfig
	fmt.Println("✅ Configuration initialized")

	// 初始化日志
	fmt.Println("Initializing logger...")
	if err := logger.InitLogger(logger.LogConfig{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxAge:     cfg.Log.MaxAge,
		MaxBackups: cfg.Log.MaxBackups,
		Compress:   cfg.Log.Compress,
	}); err != nil {
		fmt.Printf("❌ Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	fmt.Println("✅ Logger initialized")

	logger.Info("Starting AI Inference Service...")

	// 初始化 Jaeger 追踪
	logger.Info("Initializing Jaeger tracer...")
	tracer, closer, err := tracing.InitJaeger("ai-inference-service")
	if err != nil {
		logger.Warn("Failed to initialize Jaeger tracer: " + err.Error())
		fmt.Printf("⚠️  Jaeger tracer initialization failed: %v\n", err)
	} else {
		defer closer.Close()
		logger.Info("Jaeger tracer initialized successfully")
		fmt.Println("✅ Jaeger tracer initialized")
	}
	_ = tracer

	// 初始化服务注册中心（可选）
	registry, err := discovery.NewServiceRegistry(cfg.Etcd)
	if err != nil {
		logger.Warn("Failed to connect etcd service registry (continuing without service discovery): " + err.Error())
		fmt.Println("⚠️  Service registry connection failed (continuing without service discovery)")
		registry = nil
	} else {
		defer registry.Close()
		fmt.Println("✅ Service registry connected")
	}

	// 初始化 ZeroMQ（用于与 Edge-LLM-Infra 通信）
	fmt.Println("Initializing ZeroMQ client...")
	if err := zmq.InitZMQ(cfg.ZMQ); err != nil {
		logger.Warn("Failed to initialize ZeroMQ: " + err.Error())
		fmt.Println("⚠️  ZeroMQ initialization failed: ", err)
	} else {
		defer zmq.CloseZMQ()
		fmt.Println("✅ ZeroMQ initialized")
	}

	advertiseHost := resolveAdvertiseHost(cfg.Server.Host)
	var httpInstanceID string
	var grpcInstanceID string
	grpcPort := cfg.GRPC.Port
	if grpcPort == 0 {
		grpcPort = 9085
	}

	// 初始化数据库（可选）
	fmt.Println("Initializing database...")
	if err := database.InitDB(cfg.Database); err != nil {
		logger.Warn("Failed to initialize database: " + err.Error())
		fmt.Printf("⚠️  Database initialization failed: %v\n", err)
	} else {
		defer database.CloseDB()
		db := database.GetDB()

		// 为数据库添加追踪插件
		if err := db.Use(&tracing.GormTracingPlugin{}); err != nil {
			logger.Warn("Failed to register GORM tracing plugin: " + err.Error())
		}
		fmt.Println("✅ Database initialized")
	}

	// 初始化 Redis
	fmt.Println("Initializing Redis...")
	var redisInitialized bool
	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Warn("Failed to initialize Redis: " + err.Error())
		fmt.Printf("⚠️  Redis initialization failed: %v\n", err)
	} else {
		redisInitialized = true
		defer database.CloseRedis()
		logger.Info("Redis initialized successfully")
		fmt.Println("✅ Redis initialized")
	}

	// 初始化消息队列系统
	var queueManager *queue.QueueManager
	if redisInitialized {
		logger.Info("Initializing message queue system...")
		fmt.Println("Initializing message queue system...")
		redisClient := database.GetRedis()
		var err error
		queueManager, err = queue.InitializeQueueSystem(cfg, redisClient)
		if err != nil {
			logger.Warn("Failed to initialize queue system: " + err.Error())
			fmt.Printf("⚠️  Queue system initialization failed: %v\n", err)
		} else {
			defer queueManager.Stop()
			logger.Info("Message queue system initialized successfully")
			fmt.Println("✅ Message queue system initialized")

			// 注册 AI 任务处理器
			registerAITaskHandlers(queueManager)
		}
	}

	// 初始化 AI 推理服务
	fmt.Println("Initializing AI Inference Service...")
	aiService := services.NewAIInferenceService(cfg)
	logger.Info("AI Inference Service initialized")
	fmt.Println("✅ AI Inference Service initialized")

	// 初始化处理器
	aiHandler := handlers.NewAIHandler(aiService)

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建 Gin 引擎
	r := gin.New()

	// 添加中间件
	r.Use(middleware.Tracing("ai-inference-service")) // Jaeger 追踪
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 启动 gRPC 服务器（AIService）
	if _, err := startAIGrpcServer(grpcPort, aiService); err != nil {
		logger.Fatal("Failed to start AI gRPC server: " + err.Error())
	}


	// 注册路由
	setupRoutes(r, aiHandler)

	// 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		logger.Info("AI Inference Service HTTP started on " + addr)
		fmt.Printf("✅ AI Inference Service listening on %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	// 注册 HTTP/GRPC 服务实例（如果 registry 可用）
	if registry != nil {
		// HTTP
		httpMeta := map[string]string{"protocol": "http"}
		httpInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
			Name:     "ai-inference-service",
			Host:     advertiseHost,
			Port:     cfg.Server.Port,
			Protocol: "http",
			Metadata: httpMeta,
		})
		if err != nil {
			logger.Warn("Failed to register ai-inference-service http instance: " + err.Error())
			fmt.Println("⚠️  Service registration failed (continuing without service discovery)")
		}

		// gRPC
		grpcMeta := map[string]string{"protocol": "grpc"}
		grpcInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
			Name:     "ai-inference-service",
			Host:     advertiseHost,
			Port:     grpcPort,
			Protocol: "grpc",
			Metadata: grpcMeta,
		})
		if err != nil {
			logger.Warn("Failed to register ai-inference-service grpc instance: " + err.Error())
		}

		defer func() {
			if httpInstanceID != "" {
				if err := registry.DeregisterService("ai-inference-service", httpInstanceID); err != nil {
					logger.Warn("Failed to deregister ai-inference-service http instance: " + err.Error())
				}
			}
			if grpcInstanceID != "" {
				if err := registry.DeregisterService("ai-inference-service", grpcInstanceID); err != nil {
					logger.Warn("Failed to deregister ai-inference-service grpc instance: " + err.Error())
				}
			}
		}()

		logger.Info("AI Inference Service registered to etcd", logger.String("http_instance_id", httpInstanceID), logger.String("grpc_instance_id", grpcInstanceID))
		fmt.Printf("✅ Service registered to etcd (http: %s, grpc: %s)\n", httpInstanceID, grpcInstanceID)
	} else {
		fmt.Println("⚠️  Skipping service registration (no registry available)")
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down AI Inference Service...")
	fmt.Println("\n🛑 Shutting down AI Inference Service...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: " + err.Error())
		fmt.Printf("❌ Server forced to shutdown: %v\n", err)
	}

	logger.Info("AI Inference Service stopped")
	fmt.Println("✅ AI Inference Service stopped")
}

// resolveAdvertiseHost 解析广播地址
func resolveAdvertiseHost(defaultHost string) string {
	hostEnvs := []string{"SERVICE_ADVERTISE_HOST", "POD_IP", "HOST_IP"}
	for _, key := range hostEnvs {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	if defaultHost == "" || defaultHost == "0.0.0.0" {
		return "localhost"
	}
	return defaultHost
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine, aiHandler *handlers.AIHandler) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "ai-inference-service",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// API 版本 1
	v1 := r.Group("/api/v1")
	{
		// AI 推理接口
		ai := v1.Group("/ai")
		{
			// 基础推理接口
			ai.POST("/asr", aiHandler.SpeechRecognition)
			ai.POST("/emotion", aiHandler.EmotionDetection)
			ai.POST("/synthesis", aiHandler.SynthesisDetection)
			ai.POST("/analyze", aiHandler.Analyze)

			// 批量推理
			ai.POST("/batch", aiHandler.BatchInference)

			// 服务信息
			ai.GET("/health", aiHandler.HealthCheck)
			ai.GET("/info", aiHandler.GetServiceInfo)
		}
	}
}

// registerAITaskHandlers 注册 AI 任务处理器
func registerAITaskHandlers(qm *queue.QueueManager) {
	logger.Info("Registering AI task handlers...")

	// 注册 Redis 消息队列处理器
	if redisQueue := qm.GetRedisMessageQueue(); redisQueue != nil {
		// ASR 任务
		redisQueue.RegisterHandler("ai_asr", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing ASR task: %s", msg.ID))
			// 处理 ASR 任务逻辑
			return nil
		})

		// 情感检测任务
		redisQueue.RegisterHandler("ai_emotion", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing emotion detection task: %s", msg.ID))
			// 处理情感检测任务逻辑
			return nil
		})

		// 深度伪造检测任务
		redisQueue.RegisterHandler("ai_synthesis", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing synthesis detection task: %s", msg.ID))
			// 处理深度伪造检测任务逻辑
			return nil
		})

		logger.Info("Redis message queue handlers registered")
	}

	// 注册发布订阅处理器
	if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
		// 订阅会议事件
		pubsub.Subscribe("meeting_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received meeting event: %s", msg.Type))
			// 处理会议相关的 AI 任务
			return nil
		})

		// 订阅媒体事件
		pubsub.Subscribe("media_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received media event: %s", msg.Type))
			// 处理媒体相关的 AI 任务
			return nil
		})

		logger.Info("PubSub handlers registered")
	}

	logger.Info("All AI task handlers registered successfully")
}

