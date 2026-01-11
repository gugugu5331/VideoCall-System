package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/discovery"
	pb "meeting-system/shared/grpc"
	"meeting-system/shared/logger"
	"meeting-system/shared/metrics"
	"meeting-system/shared/middleware"
	"meeting-system/shared/queue"
	"meeting-system/shared/tracing"
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

	// 初始化 Jaeger 追踪
	logger.Info("Initializing Jaeger tracer...")
	log.Println("Initializing Jaeger tracer...")
	tracer, closer, err := tracing.InitJaeger("user-service")
	if err != nil {
		logger.Warn("Failed to initialize Jaeger tracer: " + err.Error())
		log.Println("Failed to initialize Jaeger tracer:", err.Error())
	} else {
		defer closer.Close()
		logger.Info("Jaeger tracer initialized successfully")
		log.Println("Jaeger tracer initialized successfully")
	}
	_ = tracer // 避免未使用变量警告

	// 初始化服务注册中心
	registry, err := discovery.NewServiceRegistry(cfg.Etcd)
	if err != nil {
		logger.Fatal("Failed to connect etcd service registry: " + err.Error())
	}
	defer registry.Close()

	advertiseHost := resolveAdvertiseHost(cfg.Server.Host)
	var httpInstanceID string
	var grpcInstanceID string

	// 初始化数据库
	logger.Info("Initializing database...")
	log.Println("Initializing database...")
	if err := database.InitDB(cfg.Database); err != nil {
		logger.Fatal("Failed to initialize database: " + err.Error())
	}
	defer database.CloseDB()

	// 为数据库添加追踪插件
	db := database.GetDB()
	if err := db.Use(&tracing.GormTracingPlugin{}); err != nil {
		logger.Warn("Failed to register GORM tracing plugin: " + err.Error())
	} else {
		logger.Info("GORM tracing plugin registered successfully")
	}

	logger.Info("Database initialized successfully")
	log.Println("Database initialized successfully")

	logger.Info("Initializing Redis...")
	log.Println("Initializing Redis...")
	var redisInitialized bool
	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Warn("Failed to initialize Redis: " + err.Error())
		log.Println("Failed to initialize Redis:", err.Error())
		// Redis初始化失败不影响压力测试
	} else {
		redisInitialized = true
		logger.Info("Redis initialized successfully")
		log.Println("Redis initialized successfully")
	}
	defer func() {
		if !redisInitialized {
			return
		}
		if err := database.CloseRedis(); err != nil {
			logger.Warn("Failed to close Redis client: " + err.Error())
		}
	}()

	// 初始化消息队列系统（Kafka 优先，不依赖 Redis）
	logger.Info("Initializing message queue system...")
	log.Println("Initializing message queue system...")
	queueManager, err := queue.InitializeQueueSystem(cfg)
	if err != nil {
		logger.Warn("Failed to initialize queue system: " + err.Error())
		log.Printf("Failed to initialize queue system: %v", err)
	} else {
		defer queueManager.Stop()
		logger.Info("Message queue system initialized successfully")
		log.Println("Message queue system initialized successfully")

		// 注册用户任务处理器
		registerUserTaskHandlers(queueManager)
	}

	// 跳过自动迁移（表已存在且结构正确）
	logger.Info("Skipping User table migration (table already exists)")
	log.Println("Skipping User table migration (table already exists)")

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
	r.Use(middleware.Tracing("user-service")) // Jaeger 追踪中间件
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(metrics.PrometheusMiddleware("user-service"))
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

	// 启动HTTP服务器
	go func() {
		logger.Info("User service HTTP server starting...")
		log.Println("User service HTTP server starting...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()
	log.Println("HTTP server goroutine started")

	logger.Info("User service HTTP started successfully on " + addr)
	log.Println("User service HTTP started successfully on", addr)

	// 启动gRPC服务器
	grpcPort := cfg.GRPC.Port
	if grpcPort == 0 {
		grpcPort = 50051 // 默认端口
	}
	grpcHost := cfg.Server.Host
	if grpcHost == "" {
		grpcHost = "0.0.0.0"
	}
	grpcAddr := fmt.Sprintf("%s:%d", grpcHost, grpcPort)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("Failed to listen on gRPC port: " + err.Error())
	}

	// 创建 gRPC 服务器，添加追踪拦截器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(tracing.UnaryServerInterceptor()),
	)
	userGRPCServer := NewUserGRPCServer()
	pb.RegisterUserServiceServer(grpcServer, userGRPCServer)

	go func() {
		logger.Info("User service gRPC server starting on " + grpcAddr)
		log.Println("User service gRPC server starting on", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to start gRPC server: " + err.Error())
		}
	}()

	logger.Info("User service gRPC started successfully on " + grpcAddr)
	log.Println("User service gRPC started successfully on", grpcAddr)

	// 注册HTTP服务实例
	metadata := map[string]string{
		"protocol":  "http",
		"grpc_port": strconv.Itoa(grpcPort),
	}
	httpInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
		Name:     "user-service",
		Host:     advertiseHost,
		Port:     cfg.Server.Port,
		Protocol: "http",
		Metadata: metadata,
	})
	if err != nil {
		logger.Fatal("Failed to register user-service http instance: " + err.Error())
	}

	defer func() {
		if httpInstanceID != "" {
			if err := registry.DeregisterService("user-service", httpInstanceID); err != nil {
				logger.Warn("Failed to deregister user-service http instance: " + err.Error())
			}
		}
	}()

	grpcInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
		Name:     "user-service",
		Host:     advertiseHost,
		Port:     grpcPort,
		Protocol: "grpc",
		Metadata: map[string]string{
			"protocol": "grpc",
		},
	})
	if err != nil {
		logger.Warn("Failed to register user-service grpc instance: " + err.Error())
	} else {
		defer func() {
			if grpcInstanceID != "" {
				if err := registry.DeregisterService("user-service", grpcInstanceID); err != nil {
					logger.Warn("Failed to deregister user-service grpc instance: " + err.Error())
				}
			}
		}()
	}

	// 等待中断信号
	log.Println("Waiting for interrupt signal...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down user service...")

	// 优雅关闭gRPC服务器
	logger.Info("Stopping gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped")

	// 优雅关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: " + err.Error())
		if err := server.Close(); err != nil {
			logger.Error("Failed to force close server: " + err.Error())
		}
	}

	logger.Info("User service stopped")
}

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
func setupRoutes(r *gin.Engine, userHandler *handlers.UserHandler) {
	// 应用全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	// 测试环境：禁用全局限流
	// r.Use(middleware.GlobalRateLimit(100, 200))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "user-service",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Prometheus指标端点
	r.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// CSRF token 获取端点
	r.GET("/api/v1/csrf-token", middleware.GetCSRFToken)

	// API版本1
	v1 := r.Group("/api/v1")
	v1.Use(middleware.CSRFTokenGenerator()) // 为所有请求生成 CSRF token
	{
		// 公开接口（不需要认证，但需要智能 CSRF 保护）
		auth := v1.Group("/auth")
		// 测试环境临时关闭 CSRF 保护，便于自动化脚本直接调用
		// auth.Use(middleware.SmartCSRFProtection()) // 智能 CSRF 保护：JWT Token 请求跳过
		{
			// 测试环境：禁用限流
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
			auth.POST("/forgot-password", userHandler.ForgotPassword)
			auth.POST("/reset-password", userHandler.ResetPassword)
		}

		// 需要认证的接口
		protected := v1.Group("/users")
		protected.Use(middleware.JWTAuth())
		// protected.Use(middleware.SmartCSRFProtection()) // 智能 CSRF 保护
		// 测试环境：禁用限流
		// protected.Use(middleware.UserRateLimit(50, 100))
		{
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)
			protected.POST("/change-password", userHandler.ChangePassword)
			protected.POST("/upload-avatar", userHandler.UploadAvatar)
			protected.DELETE("/account", userHandler.DeleteAccount)
		}

		// 用户管理接口（管理员）
		admin := v1.Group("/admin/users")
		admin.Use(middleware.JWTAuth())
		admin.Use(middleware.RequireAdmin())        // 要求管理员权限
		admin.Use(middleware.SmartCSRFProtection()) // 智能 CSRF 保护
		// 测试环境：禁用限流
		// admin.Use(middleware.UserRateLimit(30, 60))
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

// registerUserTaskHandlers 注册用户任务处理器
func registerUserTaskHandlers(qm *queue.QueueManager) {
	logger.Info("Registering user task handlers...")

	// 注册Kafka消息队列处理器
	if kafkaQueue := qm.GetKafkaMessageQueue(); kafkaQueue != nil {
		// 用户注册任务
		kafkaQueue.RegisterHandler("user_register", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing user register task: %s", msg.ID))

			username := msg.Payload["username"]
			email := msg.Payload["email"]

			logger.Info(fmt.Sprintf("Registering user: username=%v, email=%v", username, email))

			// 发布用户注册完成事件
			if bus := qm.GetKafkaEventBus(); bus != nil {
				bus.Publish(ctx, "user_events", &queue.PubSubMessage{
					Type: "user.registered",
					Payload: map[string]interface{}{
						"task_id":  msg.ID,
						"username": username,
						"email":    email,
					},
					Source: "user-service",
				})
			}

			return nil
		})

		// 用户登录任务
		kafkaQueue.RegisterHandler("user_login", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing user login task: %s", msg.ID))

			username := msg.Payload["username"]
			userID := msg.Payload["user_id"]

			logger.Info(fmt.Sprintf("User login: username=%v, user_id=%v", username, userID))

			// 发布用户登录事件
			if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
				pubsub.Publish(ctx, "user_events", &queue.PubSubMessage{
					Type: "user.logged_in",
					Payload: map[string]interface{}{
						"task_id":  msg.ID,
						"username": username,
						"user_id":  userID,
					},
					Source: "user-service",
				})
			}

			return nil
		})

		// 用户资料更新任务
		kafkaQueue.RegisterHandler("user_profile_update", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing user profile update task: %s", msg.ID))

			userID := msg.Payload["user_id"]
			updates := msg.Payload["updates"]

			logger.Info(fmt.Sprintf("Updating user profile: user_id=%v, updates=%v", userID, updates))

			// 发布用户资料更新事件
			if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
				pubsub.Publish(ctx, "user_events", &queue.PubSubMessage{
					Type: "user.profile_updated",
					Payload: map[string]interface{}{
						"task_id": msg.ID,
						"user_id": userID,
						"updates": updates,
					},
					Source: "user-service",
				})
			}

			return nil
		})

		// 用户状态变更任务
		kafkaQueue.RegisterHandler("user_status_change", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing user status change task: %s", msg.ID))

			userID := msg.Payload["user_id"]
			status := msg.Payload["status"]

			logger.Info(fmt.Sprintf("Changing user status: user_id=%v, status=%v", userID, status))

			// 发布用户状态变更事件
			if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
				pubsub.Publish(ctx, "user_events", &queue.PubSubMessage{
					Type: "user.status_changed",
					Payload: map[string]interface{}{
						"task_id": msg.ID,
						"user_id": userID,
						"status":  status,
					},
					Source: "user-service",
				})
			}

			return nil
		})

		logger.Info("Kafka message queue handlers registered")
	}

	// 注册发布订阅处理器
	if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
		// 订阅会议事件
		pubsub.Subscribe("meeting_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received meeting event: %s", msg.Type))

			switch msg.Type {
			case "meeting.user_joined":
				// 用户加入会议时更新用户状态
				userID := msg.Payload["user_id"]
				meetingID := msg.Payload["meeting_id"]
				logger.Info(fmt.Sprintf("User %v joined meeting %v", userID, meetingID))
			case "meeting.user_left":
				// 用户离开会议时更新用户状态
				userID := msg.Payload["user_id"]
				meetingID := msg.Payload["meeting_id"]
				logger.Info(fmt.Sprintf("User %v left meeting %v", userID, meetingID))
			case "meeting.created":
				// 会议创建时记录创建者信息
				logger.Info("Meeting created event received")
			case "meeting.ended":
				// 会议结束时更新参与用户的会议历史
				logger.Info("Meeting ended event received")
			}

			return nil
		})

		// 订阅AI事件（如果需要）
		pubsub.Subscribe("ai_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received AI event: %s", msg.Type))

			switch msg.Type {
			case "speech_recognition.completed":
				// AI处理完成，可能需要通知用户
				logger.Info("Speech recognition completed")
			case "emotion_detection.completed":
				// 情绪检测完成
				logger.Info("Emotion detection completed")
			}

			return nil
		})

		logger.Info("PubSub handlers registered")
	}

	// 注册本地事件总线处理器
	if localBus := qm.GetLocalEventBus(); localBus != nil {
		localBus.On("user_registered", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - User registered: %v", event.Payload))
			return nil
		})

		localBus.On("user_logged_in", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - User logged in: %v", event.Payload))
			return nil
		})

		localBus.On("user_profile_updated", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - User profile updated: %v", event.Payload))
			return nil
		})

		localBus.On("user_status_changed", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - User status changed: %v", event.Payload))
			return nil
		})

		logger.Info("Local event bus handlers registered")
	}

	logger.Info("All user task handlers registered successfully")
}
