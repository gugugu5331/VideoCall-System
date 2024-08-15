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

	"meeting-system/meeting-service/handlers"
	"meeting-system/meeting-service/services"
	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/shared/zmq"
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

	logger.Info("Starting meeting service...")
	log.Println("Starting meeting service...")

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
	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Warn("Failed to initialize Redis: " + err.Error())
		// Redis初始化失败不影响压力测试
	} else {
		defer database.CloseRedis()
		logger.Info("Redis initialized successfully")
	}

	// 初始化MongoDB
	logger.Info("Initializing MongoDB...")
	if err := database.InitMongoDB(cfg.MongoDB); err != nil {
		logger.Warn("Failed to initialize MongoDB: " + err.Error())
		// MongoDB初始化失败不影响压力测试
	} else {
		defer database.CloseMongoDB()
		logger.Info("MongoDB initialized successfully")
	}

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		&models.Meeting{},
		&models.MeetingParticipant{},
		&models.MeetingRoom{},
		&models.MediaStream{},
		&models.MeetingRecording{},
	); err != nil {
		logger.Fatal("Failed to migrate database: " + err.Error())
	}

	// 初始化ZMQ客户端（用于AI功能）
	if err := zmq.InitZMQ(cfg.ZMQ); err != nil {
		logger.Warn("Failed to initialize ZMQ client: " + err.Error())
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

	// 优雅关闭
	go func() {
		logger.Info("Meeting service started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down meeting service...")

	// 优雅关闭服务器
	if err := server.Close(); err != nil {
		logger.Error("Server forced to shutdown", logger.Err(err))
	}

	logger.Info("Meeting service stopped")
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
