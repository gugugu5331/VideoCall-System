package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"meeting-system/media-service/handlers"
	"meeting-system/media-service/services"
	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/discovery"
	"meeting-system/shared/logger"
	"meeting-system/shared/metrics"
	"meeting-system/shared/middleware"
	"meeting-system/shared/queue"
	"meeting-system/shared/storage"
	"meeting-system/shared/tracing"
)

func main() {
	// 添加panic恢复
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC: %v\n", r)
		}
	}()

	// 初始化配置
	config.InitConfig("config/media-service.yaml")
	cfg := config.GetConfig()
	fmt.Println("✅ Config loaded")

	// 初始化日志
	logger.InitLogger(logger.LogConfig{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxAge:     cfg.Log.MaxAge,
		MaxBackups: cfg.Log.MaxBackups,
		Compress:   cfg.Log.Compress,
	})
	fmt.Println("✅ Logger initialized")

	logger.Info("Starting Media Service...")
	fmt.Println("✅ Starting Media Service...")

	// 初始化 Jaeger 追踪
	logger.Info("Initializing Jaeger tracer...")
	tracer, closer, err := tracing.InitJaeger("media-service")
	if err != nil {
		logger.Warn("Failed to initialize Jaeger tracer: " + err.Error())
		fmt.Printf("⚠️  Jaeger tracer initialization failed: %v\n", err)
	} else {
		defer closer.Close()
		logger.Info("Jaeger tracer initialized successfully")
		fmt.Println("✅ Jaeger tracer initialized")
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
	fmt.Println("Initializing database...")
	if err := database.InitDB(cfg.Database); err != nil {
		fmt.Printf("❌ Failed to initialize database: %v\n", err)
		logger.Fatal("Failed to initialize database: " + err.Error())
	}
	db := database.GetDB()

	// 为数据库添加追踪插件
	if err := db.Use(&tracing.GormTracingPlugin{}); err != nil {
		logger.Warn("Failed to register GORM tracing plugin: " + err.Error())
	}

	fmt.Println("✅ Database initialized")

	// 初始化Redis
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

			// 注册媒体任务处理器
			registerMediaTaskHandlers(queueManager)
		}
	}

	// 初始化MinIO存储
	fmt.Println("Initializing MinIO...")
	if err := storage.InitMinIO(cfg.MinIO); err != nil {
		fmt.Printf("❌ Failed to initialize MinIO: %v\n", err)
		logger.Fatal("Failed to initialize MinIO: " + err.Error())
	}
	logger.Info("MinIO initialized successfully")
	fmt.Println("✅ MinIO initialized")

	// 跳过信令服务客户端初始化（可选功能）
	signalingClient := services.NewSignalingClient(cfg)
	logger.Info("Signaling client created (initialization skipped)")

	// 初始化媒体服务
	mediaService := services.NewMediaService(cfg, db, signalingClient)
	mediaService.SetStorageClient(services.NewMinIOStorageAdapter(cfg.MinIO.BucketName, 0))
	logger.Info("Media service created (initialization skipped)")

	// 初始化FFmpeg服务
	ffmpegService := services.NewFFmpegService(cfg, mediaService, signalingClient)
	logger.Info("FFmpeg service created (initialization skipped)")

	// 初始化AI客户端
	aiClient := services.NewAIClient(cfg)
	logger.Info("AI client created")

	// 初始化媒体处理器
	mediaProcessor := services.NewMediaProcessor(cfg, aiClient, ffmpegService)
	logger.Info("Media processor created")

	// 初始化WebRTC服务
	webrtcService := services.NewWebRTCService(cfg, mediaService, mediaProcessor)
	if err := webrtcService.Initialize(); err != nil {
		logger.Error("Failed to initialize WebRTC service: " + err.Error())
		panic(err)
	}
	logger.Info("WebRTC service initialized successfully")

	// 初始化录制服务
	recordingService := services.NewRecordingService(cfg, mediaService, ffmpegService, signalingClient)
	logger.Info("Recording service created (initialization skipped)")

	// 设置路由
	router := setupRouter(mediaService, webrtcService, ffmpegService, recordingService, mediaProcessor, aiClient)

	// 注册HTTP服务实例
	metadata := map[string]string{
		"protocol": "http",
	}
	httpInstanceID, err = registry.RegisterService(&discovery.ServiceInfo{
		Name:     "media-service",
		Host:     advertiseHost,
		Port:     cfg.Server.Port,
		Protocol: "http",
		Metadata: metadata,
	})
	if err != nil {
		logger.Fatal("Failed to register media-service http instance: " + err.Error())
	}

	defer func() {
		if httpInstanceID != "" {
			if err := registry.DeregisterService("media-service", httpInstanceID); err != nil {
				logger.Warn("Failed to deregister media-service http instance: " + err.Error())
			}
		}
	}()

	logger.Info("Media service registered to etcd", logger.String("instance_id", httpInstanceID))

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		logger.Info(fmt.Sprintf("Media Service listening on %s:%d", cfg.Server.Host, cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Media Service...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止服务（跳过，因为服务未完全初始化）
	logger.Info("Services cleanup skipped (not fully initialized)")

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: " + err.Error())
	}

	logger.Info("Media Service stopped")
}

// setupRouter 设置路由
func setupRouter(
	mediaService *services.MediaService,
	webrtcService *services.WebRTCService,
	ffmpegService *services.FFmpegService,
	recordingService *services.RecordingService,
	mediaProcessor *services.MediaProcessor,
	aiClient *services.AIClient,
) *gin.Engine {
	// 设置Gin模式
	if config.GetConfig().Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.Tracing("media-service")) // Jaeger 追踪
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "media-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Prometheus指标端点
	router.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// 服务状态
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "media-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"uptime":    time.Since(time.Now()).String(),
		})
	})

	// API路由组
	api := router.Group("/api/v1")
	{
		// 媒体处理相关
		media := api.Group("/media")
		{
			media.POST("/upload", handlers.NewMediaHandler(mediaService).UploadMedia)
			media.GET("/download/:id", handlers.NewMediaHandler(mediaService).DownloadMedia)
			media.GET("", handlers.NewMediaHandler(mediaService).ListMedia)
			media.POST("/process", handlers.NewMediaHandler(mediaService).ProcessMedia)
			media.GET("/info/:id", handlers.NewMediaHandler(mediaService).GetMediaInfo)
			media.DELETE("/:id", handlers.NewMediaHandler(mediaService).DeleteMedia)
		}

		// WebRTC相关（SFU架构）
		// 注意：SDP Offer/Answer/ICE候选的交换应该通过信令服务的WebSocket进行
		// 这里提供HTTP API仅用于测试或特殊场景
		webrtc := api.Group("/webrtc")
		{
			// SFU核心功能：接收客户端Offer并创建Answer
			webrtc.POST("/answer", handlers.NewWebRTCHandler(webrtcService).HandleOfferAndCreateAnswer)

			// ICE候选处理
			webrtc.POST("/ice-candidate", handlers.NewWebRTCHandler(webrtcService).HandleICECandidate)

			// 房间管理
			webrtc.POST("/room/:roomId/join", handlers.NewWebRTCHandler(webrtcService).JoinRoom)
			webrtc.POST("/room/:roomId/leave", handlers.NewWebRTCHandler(webrtcService).LeaveRoom)
			webrtc.GET("/room/:roomId/peers", handlers.NewWebRTCHandler(webrtcService).GetRoomPeers)
			webrtc.GET("/room/:roomId/stats", handlers.NewWebRTCHandler(webrtcService).GetRoomStats)

			// 媒体控制
			webrtc.POST("/peer/:peerId/media", handlers.NewWebRTCHandler(webrtcService).UpdatePeerMedia)
			webrtc.GET("/peer/:peerId/status", handlers.NewWebRTCHandler(webrtcService).GetPeerStatus)

			// SFU renegotiation / trickle ICE
			webrtc.GET("/peer/:peerId/ice-candidates", handlers.NewWebRTCHandler(webrtcService).GetICECandidates)
			webrtc.GET("/peer/:peerId/offer", handlers.NewWebRTCHandler(webrtcService).GetPendingOffer)
			webrtc.POST("/peer/:peerId/answer", handlers.NewWebRTCHandler(webrtcService).HandlePeerAnswer)
		}

		// SFU 架构：FFmpeg 转码相关路由已禁用
		// 原因：SFU 不应进行服务端转码、格式转换等操作
		// 保留缩略图生成和任务状态查询（用于录制后处理）
		ffmpeg := api.Group("/ffmpeg")
		{
			// 转码功能已禁用（违反 SFU 架构）
			// ffmpeg.POST("/transcode", handlers.NewFFmpegHandler(ffmpegService).TranscodeMedia)
			// ffmpeg.POST("/extract-audio", handlers.NewFFmpegHandler(ffmpegService).ExtractAudio)
			// ffmpeg.POST("/extract-video", handlers.NewFFmpegHandler(ffmpegService).ExtractVideo)
			// ffmpeg.POST("/merge", handlers.NewFFmpegHandler(ffmpegService).MergeMedia)

			// 保留缩略图生成（用于录制后处理，非实时）
			ffmpeg.POST("/thumbnail", handlers.NewFFmpegHandler(ffmpegService).GenerateThumbnail)
			ffmpeg.GET("/job/:id/status", handlers.NewFFmpegHandler(ffmpegService).GetJobStatus)
		}

		// 录制相关
		recording := api.Group("/recording")
		{
			recording.POST("/start", handlers.NewRecordingHandler(recordingService).StartRecording)
			recording.POST("/stop", handlers.NewRecordingHandler(recordingService).StopRecording)
			recording.GET("/status/:id", handlers.NewRecordingHandler(recordingService).GetRecordingStatus)
			recording.GET("/list", handlers.NewRecordingHandler(recordingService).ListRecordings)
			recording.GET("/download/:id", handlers.NewRecordingHandler(recordingService).DownloadRecording)
			recording.DELETE("/:id", handlers.NewRecordingHandler(recordingService).DeleteRecording)
		}

		// SFU 架构：流媒体路由已删除
		// 原因：SFU不应进行服务端流媒体处理（RTMP/HLS等），仅负责WebRTC RTP转发
		// 已删除路由：/streaming/*

		// AI处理相关（仅保留监控和状态查询）
		ai := api.Group("/ai")
		{
			aiHandler := handlers.NewAIHandler(mediaProcessor, aiClient)

			// 连通性检测
			ai.GET("/connectivity", aiHandler.CheckAIConnectivity)

			// 流状态管理（仅查询）
			ai.GET("/streams", aiHandler.ListActiveStreams)
			ai.GET("/streams/:stream_id", aiHandler.GetStreamStatus)

			// SFU 架构：服务端AI处理路由已删除
			// 原因：客户端应直接调用AI服务接口
			// 已删除路由：
			// - POST /streams/:stream_id/enable
			// - POST /streams/:stream_id/disable
			// - POST /process/audio
			// - POST /process/video
			// - POST /process/multimodal

			// 统计信息
			ai.GET("/stats", aiHandler.GetAIProcessingStats)
		}

		// SFU 架构：滤镜和美颜相关路由已完全禁用
		// 原因：SFU 架构要求所有滤镜、美颜等视觉效果在客户端处理
		// 服务端仅负责媒体流的选择性转发
		//
		// 替代方案：
		// - 在用户服务中存储用户的滤镜偏好设置
		// - 通过 WebSocket 信令将滤镜参数发送给客户端
		// - 客户端本地应用滤镜效果后再发送媒体流
		//
		// filters := api.Group("/filters")
		// {
		// 	filters.POST("/apply", handlers.NewFilterHandler(mediaService).ApplyFilter)
		// 	filters.GET("/list", handlers.NewFilterHandler(mediaService).ListFilters)
		// 	filters.POST("/beauty", handlers.NewFilterHandler(mediaService).ApplyBeautyFilter)
		// 	filters.POST("/custom", handlers.NewFilterHandler(mediaService).ApplyCustomFilter)
		// }
	}

	// 注意：WebSocket连接由信令服务统一处理，媒体服务不直接处理WebSocket

	return router
}

// resolveAdvertiseHost 解析广播地址
func resolveAdvertiseHost(host string) string {
	if host == "" || host == "0.0.0.0" {
		return "localhost"
	}
	return host
}

// registerMediaTaskHandlers 注册媒体任务处理器
func registerMediaTaskHandlers(qm *queue.QueueManager) {
	logger.Info("Registering media task handlers...")

	// 注册Redis消息队列处理器
	if redisQueue := qm.GetRedisMessageQueue(); redisQueue != nil {
		// 媒体流处理任务
		redisQueue.RegisterHandler("media_stream_process", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing media stream task: %s", msg.ID))

			streamID := msg.Payload["stream_id"]
			logger.Info(fmt.Sprintf("Processing stream: %v", streamID))

			// 发布流处理完成事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				pubsub.Publish(ctx, "media_events", &queue.PubSubMessage{
					Type: "stream.processed",
					Payload: map[string]interface{}{
						"task_id":   msg.ID,
						"stream_id": streamID,
					},
					Source: "media-service",
				})
			}

			return nil
		})

		// 录制任务
		redisQueue.RegisterHandler("media_recording_start", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing recording start task: %s", msg.ID))

			meetingID := msg.Payload["meeting_id"]
			logger.Info(fmt.Sprintf("Starting recording for meeting: %v", meetingID))

			// 发布录制开始事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				pubsub.Publish(ctx, "media_events", &queue.PubSubMessage{
					Type: "recording.started",
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"meeting_id": meetingID,
					},
					Source: "media-service",
				})
			}

			return nil
		})

		// 录制停止任务
		redisQueue.RegisterHandler("media_recording_stop", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing recording stop task: %s", msg.ID))

			meetingID := msg.Payload["meeting_id"]
			logger.Info(fmt.Sprintf("Stopping recording for meeting: %v", meetingID))

			// 发布录制停止事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				pubsub.Publish(ctx, "media_events", &queue.PubSubMessage{
					Type: "recording.stopped",
					Payload: map[string]interface{}{
						"task_id":    msg.ID,
						"meeting_id": meetingID,
					},
					Source: "media-service",
				})
			}

			return nil
		})

		// 转码任务
		redisQueue.RegisterHandler("media_transcode", func(ctx context.Context, msg *queue.Message) error {
			logger.Info(fmt.Sprintf("Processing transcode task: %s", msg.ID))

			videoID := msg.Payload["video_id"]
			format := msg.Payload["format"]

			logger.Info(fmt.Sprintf("Transcoding video %v to format %v", videoID, format))

			// 发布转码完成事件
			if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
				pubsub.Publish(ctx, "media_events", &queue.PubSubMessage{
					Type: "transcode.completed",
					Payload: map[string]interface{}{
						"task_id":  msg.ID,
						"video_id": videoID,
						"format":   format,
					},
					Source: "media-service",
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
				// 会议开始时准备媒体资源
				logger.Info("Meeting started, preparing media resources")
			case "meeting.ended":
				// 会议结束时清理媒体资源
				logger.Info("Meeting ended, cleaning up media resources")
			case "meeting.user_joined":
				// 用户加入时准备媒体流
				logger.Info("User joined meeting, preparing media stream")
			case "meeting.user_left":
				// 用户离开时清理媒体流
				logger.Info("User left meeting, cleaning up media stream")
			}

			return nil
		})

		// 订阅AI事件
		pubsub.Subscribe("ai_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received AI event: %s", msg.Type))

			switch msg.Type {
			case "audio_denoising.completed":
				// 音频降噪完成
				logger.Info("Audio denoising completed")
			case "video_enhancement.completed":
				// 视频增强完成
				logger.Info("Video enhancement completed")
			}

			return nil
		})

		// 订阅信令事件
		pubsub.Subscribe("signaling_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
			logger.Info(fmt.Sprintf("Received signaling event: %s", msg.Type))

			switch msg.Type {
			case "webrtc.offer":
				// 处理WebRTC offer
				logger.Info("Received WebRTC offer")
			case "webrtc.answer":
				// 处理WebRTC answer
				logger.Info("Received WebRTC answer")
			case "webrtc.ice_candidate":
				// 处理ICE candidate
				logger.Info("Received ICE candidate")
			}

			return nil
		})

		logger.Info("PubSub handlers registered")
	}

	// 注册本地事件总线处理器
	if localBus := qm.GetLocalEventBus(); localBus != nil {
		localBus.On("stream_started", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Stream started: %v", event.Payload))
			return nil
		})

		localBus.On("stream_stopped", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Stream stopped: %v", event.Payload))
			return nil
		})

		localBus.On("recording_started", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Recording started: %v", event.Payload))
			return nil
		})

		localBus.On("recording_stopped", func(ctx context.Context, event *queue.LocalEvent) error {
			logger.Info(fmt.Sprintf("Local event - Recording stopped: %v", event.Payload))
			return nil
		})

		logger.Info("Local event bus handlers registered")
	}

	logger.Info("All media task handlers registered successfully")
}
