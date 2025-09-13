package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"video-conference-system/shared/config"
	"video-conference-system/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CommunicationRecord 通讯记录
type CommunicationRecord struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	MeetingID string                 `bson:"meeting_id" json:"meeting_id"`
	UserID    string                 `bson:"user_id" json:"user_id"`
	Username  string                 `bson:"username" json:"username"`
	Type      string                 `bson:"message_type" json:"message_type"` // text, audio, video, file, system
	Content   map[string]interface{} `bson:"content" json:"content"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// MeetingRecord 会议记录
type MeetingRecord struct {
	ID               primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	MeetingID        string                 `bson:"meeting_id" json:"meeting_id"`
	Title            string                 `bson:"title" json:"title"`
	StartTime        time.Time              `bson:"start_time" json:"start_time"`
	EndTime          *time.Time             `bson:"end_time,omitempty" json:"end_time,omitempty"`
	Participants     []ParticipantRecord    `bson:"participants" json:"participants"`
	Recording        *RecordingInfo         `bson:"recording,omitempty" json:"recording,omitempty"`
	DetectionSummary *DetectionSummary      `bson:"detection_summary,omitempty" json:"detection_summary,omitempty"`
	Statistics       *MeetingStatistics     `bson:"statistics,omitempty" json:"statistics,omitempty"`
	CreatedAt        time.Time              `bson:"created_at" json:"created_at"`
}

// ParticipantRecord 参与者记录
type ParticipantRecord struct {
	UserID          string    `bson:"user_id" json:"user_id"`
	Username        string    `bson:"username" json:"username"`
	JoinTime        time.Time `bson:"join_time" json:"join_time"`
	LeaveTime       *time.Time `bson:"leave_time,omitempty" json:"leave_time,omitempty"`
	TotalDuration   int       `bson:"total_duration" json:"total_duration"`     // 秒
	SpeakingTime    int       `bson:"speaking_time" json:"speaking_time"`       // 秒
	DetectionAlerts int       `bson:"detection_alerts" json:"detection_alerts"` // 检测告警次数
}

// RecordingInfo 录制信息
type RecordingInfo struct {
	FileURL  string `bson:"file_url" json:"file_url"`
	FileSize int64  `bson:"file_size" json:"file_size"`
	Duration int    `bson:"duration" json:"duration"` // 秒
	Format   string `bson:"format" json:"format"`
}

// DetectionSummary 检测摘要
type DetectionSummary struct {
	TotalDetections     int                      `bson:"total_detections" json:"total_detections"`
	FakeDetections      int                      `bson:"fake_detections" json:"fake_detections"`
	SuspiciousActivities []SuspiciousActivity     `bson:"suspicious_activities" json:"suspicious_activities"`
}

// SuspiciousActivity 可疑活动
type SuspiciousActivity struct {
	UserID     string                 `bson:"user_id" json:"user_id"`
	Timestamp  time.Time              `bson:"timestamp" json:"timestamp"`
	Type       string                 `bson:"type" json:"type"`
	Confidence float64                `bson:"confidence" json:"confidence"`
	Details    map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
}

// MeetingStatistics 会议统计
type MeetingStatistics struct {
	PeakParticipants   int                    `bson:"peak_participants" json:"peak_participants"`
	TotalMessages      int                    `bson:"total_messages" json:"total_messages"`
	TotalFilesShared   int                    `bson:"total_files_shared" json:"total_files_shared"`
	NetworkQuality     map[string]interface{} `bson:"network_quality" json:"network_quality"`
}

// SystemLog 系统日志
type SystemLog struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Level     string                 `bson:"level" json:"level"` // debug, info, warn, error, fatal
	Service   string                 `bson:"service" json:"service"`
	Message   string                 `bson:"message" json:"message"`
	Details   map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
	UserID    string                 `bson:"user_id,omitempty" json:"user_id,omitempty"`
	MeetingID string                 `bson:"meeting_id,omitempty" json:"meeting_id,omitempty"`
	RequestID string                 `bson:"request_id,omitempty" json:"request_id,omitempty"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// RecordService 记录服务
type RecordService struct {
	db *database.MongoDB
}

// NewRecordService 创建记录服务
func NewRecordService(db *database.MongoDB) *RecordService {
	return &RecordService{db: db}
}

// SaveCommunicationRecord 保存通讯记录
func (s *RecordService) SaveCommunicationRecord(record *CommunicationRecord) error {
	collection := s.db.Collection("communications")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	record.Timestamp = time.Now()
	_, err := collection.InsertOne(ctx, record)
	return err
}

// GetCommunicationRecords 获取通讯记录
func (s *RecordService) GetCommunicationRecords(meetingID string, page, limit int) ([]CommunicationRecord, int64, error) {
	collection := s.db.Collection("communications")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}
	if meetingID != "" {
		filter["meeting_id"] = meetingID
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []CommunicationRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// SaveMeetingRecord 保存会议记录
func (s *RecordService) SaveMeetingRecord(record *MeetingRecord) error {
	collection := s.db.Collection("meeting_records")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := collection.InsertOne(ctx, record)
	return err
}

// GetMeetingRecord 获取会议记录
func (s *RecordService) GetMeetingRecord(meetingID string) (*MeetingRecord, error) {
	collection := s.db.Collection("meeting_records")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var record MeetingRecord
	err := collection.FindOne(ctx, bson.M{"meeting_id": meetingID}).Decode(&record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// UpdateMeetingRecord 更新会议记录
func (s *RecordService) UpdateMeetingRecord(meetingID string, update bson.M) error {
	collection := s.db.Collection("meeting_records")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"meeting_id": meetingID},
		bson.M{"$set": update},
	)
	return err
}

// GetMeetingRecords 获取会议记录列表
func (s *RecordService) GetMeetingRecords(userID string, page, limit int) ([]MeetingRecord, int64, error) {
	collection := s.db.Collection("meeting_records")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}
	if userID != "" {
		filter["participants.user_id"] = userID
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "start_time", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var records []MeetingRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// SaveSystemLog 保存系统日志
func (s *RecordService) SaveSystemLog(log *SystemLog) error {
	collection := s.db.Collection("system_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Timestamp = time.Now()
	_, err := collection.InsertOne(ctx, log)
	return err
}

// GetSystemLogs 获取系统日志
func (s *RecordService) GetSystemLogs(level, service string, page, limit int) ([]SystemLog, int64, error) {
	collection := s.db.Collection("system_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}
	if level != "" {
		filter["level"] = level
	}
	if service != "" {
		filter["service"] = service
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []SystemLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 连接MongoDB
	mongodb, err := database.NewMongoDB(
		"mongodb://admin:password123@mongodb:27017",
		"video_conference",
	)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	// 创建记录服务
	recordService := NewRecordService(mongodb)

	// 设置路由
	router := setupRoutes(recordService)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("Record service starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(recordService *RecordService) *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "record-service",
			"timestamp": time.Now(),
		})
	})

	// API路由
	api := router.Group("/api/v1")
	{
		// 通讯记录
		communications := api.Group("/records/communications")
		{
			communications.POST("", func(c *gin.Context) {
				var record CommunicationRecord
				if err := c.ShouldBindJSON(&record); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := recordService.SaveCommunicationRecord(&record); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"message": "Communication record saved",
					"id":      record.ID.Hex(),
				})
			})

			communications.GET("", func(c *gin.Context) {
				meetingID := c.Query("meeting_id")
				page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
				limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

				records, total, err := recordService.GetCommunicationRecords(meetingID, page, limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"records": records,
					"total":   total,
					"page":    page,
					"limit":   limit,
				})
			})
		}

		// 会议记录
		meetings := api.Group("/records/meetings")
		{
			meetings.POST("", func(c *gin.Context) {
				var record MeetingRecord
				if err := c.ShouldBindJSON(&record); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := recordService.SaveMeetingRecord(&record); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"message": "Meeting record saved",
					"id":      record.ID.Hex(),
				})
			})

			meetings.GET("/:meeting_id", func(c *gin.Context) {
				meetingID := c.Param("meeting_id")
				record, err := recordService.GetMeetingRecord(meetingID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Meeting record not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"record": record,
				})
			})

			meetings.PUT("/:meeting_id", func(c *gin.Context) {
				meetingID := c.Param("meeting_id")
				var update map[string]interface{}
				if err := c.ShouldBindJSON(&update); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := recordService.UpdateMeetingRecord(meetingID, update); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "Meeting record updated",
				})
			})

			meetings.GET("", func(c *gin.Context) {
				userID := c.Query("user_id")
				page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
				limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

				records, total, err := recordService.GetMeetingRecords(userID, page, limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"records": records,
					"total":   total,
					"page":    page,
					"limit":   limit,
				})
			})
		}

		// 系统日志
		logs := api.Group("/logs")
		{
			logs.POST("", func(c *gin.Context) {
				var log SystemLog
				if err := c.ShouldBindJSON(&log); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if err := recordService.SaveSystemLog(&log); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"message": "System log saved",
					"id":      log.ID.Hex(),
				})
			})

			logs.GET("", func(c *gin.Context) {
				level := c.Query("level")
				service := c.Query("service")
				page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
				limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

				logs, total, err := recordService.GetSystemLogs(level, service, page, limit)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"logs":  logs,
					"total": total,
					"page":  page,
					"limit": limit,
				})
			})
		}
	}

	return router
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
