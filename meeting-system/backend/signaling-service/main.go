package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/discovery"
	sharedgrpc "meeting-system/shared/grpc"
	"meeting-system/shared/logger"
	"meeting-system/shared/metrics"
	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/shared/queue"
	"meeting-system/shared/tracing"
	"meeting-system/shared/zmq"
	"meeting-system/signaling-service/handlers"
	"meeting-system/signaling-service/services"
)

var (
	configPath = flag.String("config", "../config/signaling-service.yaml", "配置文件路径")
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED: %v", r)
			fmt.Printf("PANIC RECOVERED: %v\n", r)
		}
	}()

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

	logger.Info("Starting signaling service...")
	log.Println("Starting signaling service...")

	// 初始化 Jaeger 追踪
	logger.Info("Initializing Jaeger tracer...")
	tracer, closer, err := tracing.InitJaeger("signaling-service")
	if err != nil {
		logger.Warn("Failed to initialize Jaeger tracer: " + err.Error())
	} else {
		defer closer.Close()
		logger.Info("Jaeger tracer initialized successfully")
	}
	_ = tracer

	// 初始化服务注册中心
	registry, err := discovery.NewServiceRegistry(cfg.Etcd)
	if err != nil {
		logger.Fatal("Failed to connect etcd service registry: " + err.Error())
	}
	defer registry.Close()

	advertiseHost := resolveAdvertiseHost(cfg.Server.Host)
	var httpInstanceID string

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
	}

	logger.Info("Database initialized successfully")
	log.Println("Database initialized successfully")

	// 初始化Redis
	var redisInitialized bool
	logger.Info("Initializing Redis...")
	log.Println("Initializing Redis...")
	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Warn("Failed to initialize Redis: " + err.Error())
		log.Printf("Failed to initialize Redis: %v", err)
	} else {
		redisInitialized = true
		logger.Info("Redis initialized successfully")
		log.Println("Redis initialized successfully")
	}
	defer func() {
		if redisInitialized {
			database.CloseRedis()
		}
	}()

	// 初始化消息队列系统
	var queueManager *queue.QueueManager
	if redisInitialized {
		logger.Info("Initializing message queue system...")
		log.Println("Initializing message queue system...")
		redisClient := database.GetRedis()
		var err error
		queueManager, err = queue.InitializeQueueSystem(cfg, redisClient)
		if err != nil {
			logger.Warn("Failed to initialize queue system: " + err.Error())
			log.Printf("Failed to initialize queue system: %v", err)
		} else {
			defer queueManager.Stop()
			logger.Info("Message queue system initialized successfully")
			log.Println("Message queue system initialized successfully")

			// 注册信令任务处理器
			registerSignalingTaskHandlers(queueManager)
		}
	}

	// MongoDB和MinIO不是信令服务的必需依赖，跳过初始化
	logger.Info("MongoDB and MinIO initialization skipped (not required for signaling service)")
	log.Println("MongoDB and MinIO initialization skipped (not required for signaling service)")

	// 自动迁移数据库表（异步执行，不阻塞服务启动）
	go func() {
		logger.Info("Migrating database tables...")
		log.Println("Migrating database tables...")
		if err := database.AutoMigrate(
			&models.SignalingSession{},
			&models.SignalingMessage{},
		); err != nil {
			logger.Error("Failed to migrate database: " + err.Error())
			log.Printf("Failed to migrate database: %v", err)
		} else {
			logger.Info("Database migration completed")
			log.Println("Database migration completed")
		}
	}()

	// 异步初始化ZMQ客户端（用于AI功能）
	// ZMQ初始化可能需要30秒，不应阻塞HTTP服务器启动
	var zmqInitialized bool
	go func() {
		logger.Info("Starting ZMQ initialization in background...")
		log.Println("Starting ZMQ initialization in background...")
		if err := zmq.InitZMQ(cfg.ZMQ); err != nil {
			logger.Warn("Failed to initialize ZMQ client: " + err.Error())
			log.Printf("Failed to initialize ZMQ client: %v", err)
		} else {
			zmqInitialized = true
			logger.Info("ZMQ client initialized successfully")
			log.Println("ZMQ client initialized successfully")
		}
	}()
	defer func() {
		if zmqInitialized {
			zmq.CloseZMQ()
		}
	}()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(middleware.Tracing("signaling-service")) // Jaeger 追踪
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 初始化gRPC客户端
	grpcClients := sharedgrpc.NewServiceClients(cfg)
	if err := grpcClients.Initialize(); err != nil {
		logger.Fatal("Failed to initialize gRPC clients: " + err.Error())
	}
	defer grpcClients.Close()

	// 初始化服务
	logger.Info("Initializing signaling service components...")
	signalingService := services.NewSignalingService(grpcClients)
	wsHandler := handlers.NewWebSocketHandler(signalingService)
	logger.Info("Signaling service components initialized")

	// 注册路由
	logger.Info("Setting up routes...")
	setupRoutes(r, wsHandler, signalingService)
	logger.Info("Routes configured")

	// 启动定期清理任务
	go startCleanupTasks(signalingService)

	// 注册HTTP服务实例
	metadata := map[string]string{
		"protocol": "http",
	}
	httpInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
		Name:     "signaling-service",
		Host:     advertiseHost,
		Port:     cfg.Server.Port,
		Protocol: "http",
		Metadata: metadata,
	})
	if err != nil {
		logger.Fatal("Failed to register signaling-service http instance: " + err.Error())
	}

	defer func() {
		if httpInstanceID != "" {
			if err := registry.DeregisterService("signaling-service", httpInstanceID); err != nil {
				logger.Warn("Failed to deregister signaling-service http instance: " + err.Error())
			}
		}
	}()

	logger.Info("Signaling service registered to etcd", logger.String("instance_id", httpInstanceID))

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
		logger.Info("Signaling service started and listening on " + addr)
		log.Println("Signaling service started and listening on", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down signaling service...")

	// 停止WebSocket处理器
	wsHandler.Stop()

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.Err(err))
		if err := server.Close(); err != nil {
			logger.Error("Failed to force close server", logger.Err(err))
		}
	}

	logger.Info("Signaling service stopped")
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine, wsHandler *handlers.WebSocketHandler, signalingService *services.SignalingService) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		activeCount, _ := signalingService.GetActiveSessionCount()
		roomStats := wsHandler.GetRoomStats()

		c.JSON(http.StatusOK, gin.H{
			"status":            "ok",
			"service":           "signaling-service",
			"time":              time.Now().Format(time.RFC3339),
			"active_sessions":   activeCount,
			"connected_clients": wsHandler.GetClientCount(),
			"active_rooms":      len(roomStats),
		})
	})

	// Prometheus指标端点
	r.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// WebSocket信令接口
	r.GET("/ws/signaling", wsHandler.HandleWebSocket)

	// API版本1
	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth()) // 所有API都需要认证
	{
		// 会话管理
		sessions := v1.Group("/sessions")
		{
			sessions.GET("/:session_id", getSession(signalingService))
			sessions.GET("/room/:meeting_id", getRoomSessions(signalingService))
		}

		// 消息历史
		messages := v1.Group("/messages")
		{
			messages.GET("/history/:meeting_id", getMessageHistory(signalingService))
		}

		// 统计信息
		stats := v1.Group("/stats")
		{
			stats.GET("/overview", getStatsOverview(signalingService, wsHandler))
			stats.GET("/rooms", getRoomStats(wsHandler))
		}
	}

	// 管理接口
	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuth()) // TODO: 添加管理员权限检查
	{
		admin.POST("/cleanup/sessions", cleanupSessions(signalingService))
		admin.GET("/sessions", listAllSessions(signalingService))
	}
}

// getSession 获取会话信息
func getSession(service *services.SignalingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("session_id")

		session, err := service.GetSession(sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": session})
	}
}

// getRoomSessions 获取房间会话列表
func getRoomSessions(service *services.SignalingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		meetingIDStr := c.Param("meeting_id")
		meetingID, err := parseUint(meetingIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting_id"})
			return
		}

		sessions, err := service.GetRoomSessions(uint(meetingID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room sessions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": sessions})
	}
}

// getMessageHistory 获取消息历史
func getMessageHistory(service *services.SignalingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		meetingIDStr := c.Param("meeting_id")
		meetingID, err := parseUint(meetingIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting_id"})
			return
		}

		limit := 50 // 默认限制
		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := parseUint(limitStr); err == nil && l > 0 && l <= 1000 {
				limit = int(l)
			}
		}

		messages, err := service.GetMessageHistory(uint(meetingID), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get message history"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": messages})
	}
}

// getStatsOverview 获取统计概览
func getStatsOverview(service *services.SignalingService, wsHandler *handlers.WebSocketHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		activeCount, _ := service.GetActiveSessionCount()
		roomStats := wsHandler.GetRoomStats()

		c.JSON(http.StatusOK, gin.H{
			"active_sessions":   activeCount,
			"connected_clients": wsHandler.GetClientCount(),
			"active_rooms":      len(roomStats),
			"room_details":      roomStats,
		})
	}
}

// getRoomStats 获取房间统计
func getRoomStats(wsHandler *handlers.WebSocketHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := wsHandler.GetRoomStats()
		c.JSON(http.StatusOK, gin.H{"data": stats})
	}
}

// cleanupSessions 清理过期会话
func cleanupSessions(service *services.SignalingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := service.CleanupExpiredSessions(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup sessions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Sessions cleaned up successfully"})
	}
}

// listAllSessions 列出所有会话（管理员功能）
func listAllSessions(service *services.SignalingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现分页查询所有会话
		c.JSON(http.StatusOK, gin.H{"message": "Not implemented yet"})
	}
}

// startCleanupTasks 启动定期清理任务
func startCleanupTasks(service *services.SignalingService) {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		if err := service.CleanupExpiredSessions(); err != nil {
			logger.Error("Failed to cleanup expired sessions", logger.Err(err))
		}
	}
}

// parseUint 解析无符号整数
func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 32)
}

// resolveAdvertiseHost 解析广播地址
func resolveAdvertiseHost(host string) string {
	if host == "" || host == "0.0.0.0" {
		return "localhost"
	}
	return host
}

// registerSignalingTaskHandlers 注册信令任务处理器
func registerSignalingTaskHandlers(qm *queue.QueueManager) {
	logger.Info("Registering signaling task handlers...")

	// 注册Redis消息队列处理器
	if redisQueue := qm.GetRedisMessageQueue(); redisQueue != nil {
		// WebRTC信令处理任务
		redisQueue.RegisterHandler("signaling_webrtc_offer", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing WebRTC offer task: %s", msg.ID))

			sessionID := msg.Payload["session_id"]
			offer := msg.Payload["offer"]

			logger.Info(fmt.Sprintf("Processing WebRTC offer for session: %v", sessionID))

			// 发布offer处理完成事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				pubsub.Publish(ctx, "signaling_events", &queue.PubSubMessage{
					Type: "webrtc.offer",
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"session_id": sessionID,
						"offer":      offer,
					},
					Source: "signaling-service",
				})
			}

			return nil
		})

		// WebRTC answer处理任务
		redisQueue.RegisterHandler("signaling_webrtc_answer", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing WebRTC answer task: %s", msg.ID))

			sessionID := msg.Payload["session_id"]
			answer := msg.Payload["answer"]

			logger.Info(fmt.Sprintf("Processing WebRTC answer for session: %v", sessionID))

			// 发布answer处理完成事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				pubsub.Publish(ctx, "signaling_events", &queue.PubSubMessage{
					Type: "webrtc.answer",
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"session_id": sessionID,
						"answer":     answer,
					},
					Source: "signaling-service",
				})
			}

			return nil
		})

		// ICE candidate处理任务
		redisQueue.RegisterHandler("signaling_ice_candidate", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing ICE candidate task: %s", msg.ID))

			sessionID := msg.Payload["session_id"]
			candidate := msg.Payload["candidate"]

			logger.Info(fmt.Sprintf("Processing ICE candidate for session: %v", sessionID))

			// 发布ICE candidate事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				pubsub.Publish(ctx, "signaling_events", &queue.PubSubMessage{
					Type: "webrtc.ice_candidate",
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"session_id": sessionID,
						"candidate":  candidate,
					},
					Source: "signaling-service",
				})
			}

			return nil
		})

		// 连接管理任务
		redisQueue.RegisterHandler("signaling_connection_manage", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing connection management task: %s", msg.ID))

			action := msg.Payload["action"]
			sessionID := msg.Payload["session_id"]

			logger.Info(fmt.Sprintf("Managing connection: action=%v, session=%v", action, sessionID))

			// 发布连接管理事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				eventType := "webrtc.connected"
				if action == "disconnect" {
					eventType = "webrtc.disconnected"
				}

				pubsub.Publish(ctx, "signaling_events", &queue.PubSubMessage{
					Type: eventType,
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"session_id": sessionID,
					},
					Source: "signaling-service",
				})
			}

			return nil
		})

		logger.Info("Redis message queue handlers registered")
	}

	// 注册发布订阅处理器
	if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
		// 订阅会议事件
		pubsub.Subscribe("meeting_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received meeting event: %s", msg.Type))

			switch msg.Type {
			case "meeting.started":
				// 会议开始时准备信令资源
				logger.Info("Meeting started, preparing signaling resources")
			case "meeting.ended":
				// 会议结束时清理信令会话
				logger.Info("Meeting ended, cleaning up signaling sessions")
			case "meeting.user_joined":
				// 用户加入时建立信令连接
				logger.Info("User joined meeting, establishing signaling connection")
			case "meeting.user_left":
				// 用户离开时断开信令连接
				logger.Info("User left meeting, closing signaling connection")
			}

			return nil
		})

		// 订阅媒体事件
		pubsub.Subscribe("media_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received media event: %s", msg.Type))

			switch msg.Type {
			case "stream.started":
				// 媒体流开始时更新信令状态
				logger.Info("Media stream started, updating signaling state")
			case "stream.stopped":
				// 媒体流停止时更新信令状态
				logger.Info("Media stream stopped, updating signaling state")
			}

			return nil
		})

		logger.Info("PubSub handlers registered")
	}

	// 注册本地事件总线处理器
	if localBus := qm.GetLocalEventBus(); localBus != nil {
		localBus.On("session_created", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Session created: %v", event.Payload))
			return nil
		})

		localBus.On("session_closed", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Session closed: %v", event.Payload))
			return nil
		})

		localBus.On("ice_candidate_added", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - ICE candidate added: %v", event.Payload))
			return nil
		})

		logger.Info("Local event bus handlers registered")
	}

	logger.Info("All signaling task handlers registered successfully")
}
