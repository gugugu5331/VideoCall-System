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

	"meeting-system/meeting-service/handlers"
	"meeting-system/meeting-service/services"
	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/discovery"
	pb "meeting-system/shared/grpc"
	"meeting-system/shared/logger"
	"meeting-system/shared/metrics"
	"meeting-system/shared/middleware"
	"meeting-system/shared/queue"
	"meeting-system/shared/tracing"
)

var (
	configPath = flag.String("config", "config/meeting-service.yaml", "配置文件路径")
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

	logger.Info("Starting meeting service...")
	log.Println("Starting meeting service...")

	// 初始化 Jaeger 追踪
	logger.Info("Initializing Jaeger tracer...")
	tracer, closer, err := tracing.InitJaeger("meeting-service")
	if err != nil {
		logger.Warn("Failed to initialize Jaeger tracer: " + err.Error())
	} else {
		defer closer.Close()
		logger.Info("Jaeger tracer initialized successfully")
	}
	_ = tracer

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
	}

	logger.Info("Database initialized successfully")
	log.Println("Database initialized successfully")

	var redisInitialized bool
	logger.Info("Initializing Redis...")
	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Warn("Failed to initialize Redis: " + err.Error())
		// Redis初始化失败不影响压力测试
	} else {
		redisInitialized = true
		logger.Info("Redis initialized successfully")
	}
	defer func() {
		if !redisInitialized {
			return
		}
		if err := database.CloseRedis(); err != nil {
			logger.Warn("Failed to close Redis client: " + err.Error())
		}
	}()

	// 初始化消息队列系统（Kafka 优先）
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

		// 注册会议任务处理器
		registerMeetingTaskHandlers(queueManager)
	}

	// 跳过MongoDB初始化（可选功能）
	logger.Info("Skipping MongoDB initialization (optional feature)")
	log.Println("Skipping MongoDB initialization (optional feature)")

	// 跳过自动迁移（表已存在）
	logger.Info("Skipping database migration (tables already exist)")
	log.Println("Skipping database migration (tables already exist)")

	// 跳过ZMQ初始化（可选功能）
	logger.Info("Skipping ZMQ initialization (optional feature)")
	log.Println("Skipping ZMQ initialization (optional feature)")

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(middleware.Tracing("meeting-service")) // Jaeger 追踪
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 初始化服务
	meetingService := services.NewMeetingService()
	meetingHandler := handlers.NewMeetingHandler(meetingService)

	// 注册路由
	setupRoutes(r, meetingHandler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 启动HTTP服务器
	go func() {
		logger.Info("Meeting service HTTP started on " + addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	// 启动gRPC服务器
	grpcPort := cfg.GRPC.Port
	if grpcPort == 0 {
		grpcPort = 50052 // 默认端口
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
	pb.RegisterMeetingServiceServer(grpcServer, NewMeetingGRPCServer())

	go func() {
		logger.Info("Meeting service gRPC started on " + grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to start gRPC server: " + err.Error())
		}
	}()

	metadata := map[string]string{
		"protocol":  "http",
		"grpc_port": strconv.Itoa(grpcPort),
	}
	httpInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
		Name:     "meeting-service",
		Host:     advertiseHost,
		Port:     cfg.Server.Port,
		Protocol: "http",
		Metadata: metadata,
	})
	if err != nil {
		logger.Fatal("Failed to register meeting-service http instance: " + err.Error())
	}
	defer func() {
		if httpInstanceID != "" {
			if err := registry.DeregisterService("meeting-service", httpInstanceID); err != nil {
				logger.Warn("Failed to deregister meeting-service http instance: " + err.Error())
			}
		}
	}()

	grpcInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
		Name:     "meeting-service",
		Host:     advertiseHost,
		Port:     grpcPort,
		Protocol: "grpc",
		Metadata: map[string]string{
			"protocol": "grpc",
		},
	})
	if err != nil {
		logger.Warn("Failed to register meeting-service grpc instance: " + err.Error())
	} else {
		defer func() {
			if grpcInstanceID != "" {
				if err := registry.DeregisterService("meeting-service", grpcInstanceID); err != nil {
					logger.Warn("Failed to deregister meeting-service grpc instance: " + err.Error())
				}
			}
		}()
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down meeting service...")

	// 优雅关闭gRPC服务器
	logger.Info("Stopping gRPC server...")
	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped")

	// 优雅关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logger.Err(err))
		if err := server.Close(); err != nil {
			logger.Error("Failed to force close server", logger.Err(err))
		}
	}

	logger.Info("Meeting service stopped")
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
func setupRoutes(r *gin.Engine, meetingHandler *handlers.MeetingHandler) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "meeting-service",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Prometheus指标端点
	r.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// API版本1
	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth()) // 所有会议接口都需要认证
	{
		// 会议管理
		meetings := v1.Group("/meetings")
		{
			meetings.POST("", meetingHandler.CreateMeeting)
			meetings.GET("", meetingHandler.ListMeetings)
			meetings.GET("/:id", meetingHandler.GetMeeting)
			meetings.PUT("/:id", meetingHandler.UpdateMeeting)
			meetings.DELETE("/:id", meetingHandler.DeleteMeeting)

			// 会议控制
			meetings.POST("/:id/start", meetingHandler.StartMeeting)
			meetings.POST("/:id/end", meetingHandler.EndMeeting)
			meetings.POST("/:id/join", meetingHandler.JoinMeeting)
			meetings.POST("/:id/leave", meetingHandler.LeaveMeeting)

			// 参与者管理
			meetings.GET("/:id/participants", meetingHandler.GetParticipants)
			meetings.POST("/:id/participants", meetingHandler.AddParticipant)
			meetings.DELETE("/:id/participants/:user_id", meetingHandler.RemoveParticipant)
			meetings.PUT("/:id/participants/:user_id/role", meetingHandler.UpdateParticipantRole)

			// 会议室管理
			meetings.GET("/:id/room", meetingHandler.GetMeetingRoom)
			meetings.POST("/:id/room", meetingHandler.CreateMeetingRoom)
			meetings.DELETE("/:id/room", meetingHandler.CloseMeetingRoom)

			// 录制管理
			meetings.POST("/:id/recording/start", meetingHandler.StartRecording)
			meetings.POST("/:id/recording/stop", meetingHandler.StopRecording)
			meetings.GET("/:id/recordings", meetingHandler.GetRecordings)

			// 聊天消息
			meetings.GET("/:id/messages", meetingHandler.GetChatMessages)
			meetings.POST("/:id/messages", meetingHandler.SendChatMessage)
		}

		// 我的会议
		my := v1.Group("/my")
		{
			my.GET("/meetings", meetingHandler.GetMyMeetings)
			my.GET("/meetings/upcoming", meetingHandler.GetUpcomingMeetings)
			my.GET("/meetings/history", meetingHandler.GetMeetingHistory)
		}

		// 管理员接口
		admin := v1.Group("/admin/meetings")
		// TODO: 添加管理员权限检查中间件
		{
			admin.GET("", meetingHandler.AdminListMeetings)
			admin.GET("/stats", meetingHandler.GetMeetingStats)
			admin.POST("/:id/force-end", meetingHandler.ForceEndMeeting)
		}
	}
}

// registerMeetingTaskHandlers 注册会议任务处理器
func registerMeetingTaskHandlers(qm *queue.QueueManager) {
	logger.Info("Registering meeting task handlers...")

	// 注册Kafka消息队列处理器
	if kafkaQueue := qm.GetKafkaMessageQueue(); kafkaQueue != nil {
		// 会议创建任务
		kafkaQueue.RegisterHandler("meeting_create", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing meeting create task: %s", msg.ID))

			// 从payload中获取会议信息
			meetingData := msg.Payload
			logger.Info(fmt.Sprintf("Creating meeting: %+v", meetingData))

			// 发布会议创建完成事件
			if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
				pubsub.Publish(ctx, "meeting_events", &queue.PubSubMessage{
					Type: "meeting.created",
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"meeting_id": meetingData["meeting_id"],
					},
					Source: "meeting-service",
				})
			}

			return nil
		})

		// 会议结束任务
		kafkaQueue.RegisterHandler("meeting_end", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing meeting end task: %s", msg.ID))

			meetingID := msg.Payload["meeting_id"]
			logger.Info(fmt.Sprintf("Ending meeting: %v", meetingID))

			// 发布会议结束事件
			if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
				pubsub.Publish(ctx, "meeting_events", &queue.PubSubMessage{
					Type: "meeting.ended",
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"meeting_id": meetingID,
					},
					Source: "meeting-service",
				})
			}

			return nil
		})

		// 用户加入会议任务
		kafkaQueue.RegisterHandler("meeting_user_join", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing user join task: %s", msg.ID))

			meetingID := msg.Payload["meeting_id"]
			userID := msg.Payload["user_id"]

			logger.Info(fmt.Sprintf("User %v joining meeting %v", userID, meetingID))

			// 发布用户加入事件
			if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
				pubsub.Publish(ctx, "meeting_events", &queue.PubSubMessage{
					Type: "meeting.user_joined",
					Payload: map[string]interface{}{
						"meeting_id": meetingID,
						"user_id":    userID,
					},
					Source: "meeting-service",
				})
			}

			return nil
		})

		// 用户离开会议任务
		kafkaQueue.RegisterHandler("meeting_user_leave", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing user leave task: %s", msg.ID))

			meetingID := msg.Payload["meeting_id"]
			userID := msg.Payload["user_id"]

			logger.Info(fmt.Sprintf("User %v leaving meeting %v", userID, meetingID))

			// 发布用户离开事件
			if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
				pubsub.Publish(ctx, "meeting_events", &queue.PubSubMessage{
					Type: "meeting.user_left",
					Payload: map[string]interface{}{
						"meeting_id": meetingID,
						"user_id":    userID,
					},
					Source: "meeting-service",
				})
			}

			return nil
		})

		logger.Info("Kafka message queue handlers registered")
	}

	// 注册发布订阅处理器
	if pubsub := qm.GetKafkaEventBus(); pubsub != nil {
		// 订阅AI事件
		pubsub.Subscribe("ai_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received AI event: %s", msg.Type))

			switch msg.Type {
			case "speech_recognition.completed":
				// 处理语音识别完成事件
				logger.Info("Speech recognition completed for meeting")
			case "emotion_detection.completed":
				// 处理情绪检测完成事件
				logger.Info("Emotion detection completed for meeting")
			}

			return nil
		})

		// 订阅媒体事件
		pubsub.Subscribe("media_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received media event: %s", msg.Type))

			switch msg.Type {
			case "recording.started":
				// 处理录制开始事件
				logger.Info("Recording started for meeting")
			case "recording.stopped":
				// 处理录制停止事件
				logger.Info("Recording stopped for meeting")
			case "stream.started":
				// 处理媒体流开始事件
				logger.Info("Media stream started")
			case "stream.stopped":
				// 处理媒体流停止事件
				logger.Info("Media stream stopped")
			}

			return nil
		})

		// 订阅信令事件
		pubsub.Subscribe("signaling_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received signaling event: %s", msg.Type))

			switch msg.Type {
			case "webrtc.connected":
				// 处理WebRTC连接建立事件
				logger.Info("WebRTC connection established")
			case "webrtc.disconnected":
				// 处理WebRTC连接断开事件
				logger.Info("WebRTC connection disconnected")
			}

			return nil
		})

		logger.Info("PubSub handlers registered")
	}

	// 注册本地事件总线处理器
	if localBus := qm.GetLocalEventBus(); localBus != nil {
		localBus.On("meeting_created", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Meeting created: %v", event.Payload))
			return nil
		})

		localBus.On("meeting_ended", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Meeting ended: %v", event.Payload))
			return nil
		})

		logger.Info("Local event bus handlers registered")
	}

	logger.Info("All meeting task handlers registered successfully")
}
