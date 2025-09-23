package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"video-conference-system/shared/config"
	"video-conference-system/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// SmartEditingService AI智能剪辑服务
type SmartEditingService struct {
	mongodb    *database.MongoDB
	aiAnalyzer *AIContentAnalyzer
	editor     *AutomaticEditor
	storage    string
}

// EditingTask 剪辑任务
type EditingTask struct {
	ID          string                 `json:"id" bson:"_id"`
	MeetingID   string                 `json:"meeting_id" bson:"meeting_id"`
	VideoPath   string                 `json:"video_path" bson:"video_path"`
	Status      string                 `json:"status" bson:"status"` // pending, analyzing, editing, completed, failed
	Config      EditingConfig          `json:"config" bson:"config"`
	Analysis    *ContentAnalysis       `json:"analysis,omitempty" bson:"analysis,omitempty"`
	Result      *EditingResult         `json:"result,omitempty" bson:"result,omitempty"`
	Progress    float64                `json:"progress" bson:"progress"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
	Error       string                 `json:"error,omitempty" bson:"error,omitempty"`
}

// EditingConfig 剪辑配置
type EditingConfig struct {
	Style           string            `json:"style"`           // highlight, summary, full, custom
	Duration        int               `json:"duration"`        // 目标时长（秒）
	IncludeAudio    bool              `json:"include_audio"`   // 是否包含音频
	AddSubtitles    bool              `json:"add_subtitles"`   // 是否添加字幕
	AddMusic        bool              `json:"add_music"`       // 是否添加背景音乐
	Quality         string            `json:"quality"`         // low, medium, high, ultra
	Format          string            `json:"format"`          // mp4, webm, avi
	Filters         []FilterConfig    `json:"filters"`         // 视频滤镜
	CustomSettings  map[string]interface{} `json:"custom_settings"` // 自定义设置
}

// FilterConfig 滤镜配置
type FilterConfig struct {
	Type      string  `json:"type"`      // beauty, blur, vintage, etc.
	Intensity float64 `json:"intensity"` // 0.0 - 1.0
	Enabled   bool    `json:"enabled"`
}

// ContentAnalysis 内容分析结果
type ContentAnalysis struct {
	Duration        float64           `json:"duration"`
	Participants    []Participant     `json:"participants"`
	Highlights      []Highlight       `json:"highlights"`
	AudioAnalysis   AudioAnalysis     `json:"audio_analysis"`
	VideoAnalysis   VideoAnalysis     `json:"video_analysis"`
	TextAnalysis    TextAnalysis      `json:"text_analysis"`
	OverallScore    float64           `json:"overall_score"`
	Recommendations []string          `json:"recommendations"`
}

// Participant 参与者信息
type Participant struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	SpeakingTime float64 `json:"speaking_time"`
	Engagement   float64 `json:"engagement"`   // 参与度评分
	Emotions     map[string]float64 `json:"emotions"` // 情绪分布
}

// Highlight 精彩片段
type Highlight struct {
	StartTime   float64           `json:"start_time"`
	EndTime     float64           `json:"end_time"`
	Type        string            `json:"type"`        // discussion, decision, presentation, reaction
	Score       float64           `json:"score"`       // 重要性评分
	Participants []string         `json:"participants"` // 涉及的参与者
	Keywords    []string          `json:"keywords"`    // 关键词
	Summary     string            `json:"summary"`     // 片段摘要
	Emotions    map[string]float64 `json:"emotions"`   // 情绪强度
}

// AudioAnalysis 音频分析
type AudioAnalysis struct {
	VoiceActivity    []VoiceSegment    `json:"voice_activity"`
	SpeakerChanges   []float64         `json:"speaker_changes"`
	VolumeProfile    []float64         `json:"volume_profile"`
	SilencePeriods   []TimeRange       `json:"silence_periods"`
	BackgroundNoise  float64           `json:"background_noise"`
	AudioQuality     float64           `json:"audio_quality"`
}

// VideoAnalysis 视频分析
type VideoAnalysis struct {
	SceneChanges     []float64         `json:"scene_changes"`
	MotionIntensity  []float64         `json:"motion_intensity"`
	FaceDetections   []FaceDetection   `json:"face_detections"`
	VisualQuality    float64           `json:"visual_quality"`
	Brightness       []float64         `json:"brightness"`
	ColorProfile     ColorProfile      `json:"color_profile"`
}

// TextAnalysis 文本分析（语音转文字）
type TextAnalysis struct {
	Transcript       []TranscriptSegment `json:"transcript"`
	Keywords         []Keyword          `json:"keywords"`
	Topics           []Topic            `json:"topics"`
	Sentiment        SentimentAnalysis  `json:"sentiment"`
	ImportantPhrases []string           `json:"important_phrases"`
}

// VoiceSegment 语音片段
type VoiceSegment struct {
	StartTime   float64 `json:"start_time"`
	EndTime     float64 `json:"end_time"`
	SpeakerID   string  `json:"speaker_id"`
	Confidence  float64 `json:"confidence"`
	Volume      float64 `json:"volume"`
	Pitch       float64 `json:"pitch"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// FaceDetection 人脸检测
type FaceDetection struct {
	Timestamp   float64           `json:"timestamp"`
	Faces       []Face            `json:"faces"`
}

// Face 人脸信息
type Face struct {
	PersonID    string            `json:"person_id"`
	BoundingBox BoundingBox       `json:"bounding_box"`
	Emotions    map[string]float64 `json:"emotions"`
	Landmarks   []Point           `json:"landmarks"`
	Pose        HeadPose          `json:"pose"`
}

// BoundingBox 边界框
type BoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Point 坐标点
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// HeadPose 头部姿态
type HeadPose struct {
	Yaw   float64 `json:"yaw"`
	Pitch float64 `json:"pitch"`
	Roll  float64 `json:"roll"`
}

// ColorProfile 颜色配置
type ColorProfile struct {
	Dominant    []string `json:"dominant"`
	Brightness  float64  `json:"brightness"`
	Contrast    float64  `json:"contrast"`
	Saturation  float64  `json:"saturation"`
}

// TranscriptSegment 转录片段
type TranscriptSegment struct {
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	SpeakerID  string  `json:"speaker_id"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// Keyword 关键词
type Keyword struct {
	Word      string  `json:"word"`
	Frequency int     `json:"frequency"`
	Relevance float64 `json:"relevance"`
	Timestamps []float64 `json:"timestamps"`
}

// Topic 主题
type Topic struct {
	Name       string    `json:"name"`
	Score      float64   `json:"score"`
	Keywords   []string  `json:"keywords"`
	TimeRanges []TimeRange `json:"time_ranges"`
}

// SentimentAnalysis 情感分析
type SentimentAnalysis struct {
	Overall    float64           `json:"overall"`    // -1 to 1
	Timeline   []SentimentPoint  `json:"timeline"`
	Emotions   map[string]float64 `json:"emotions"`
}

// SentimentPoint 情感时间点
type SentimentPoint struct {
	Timestamp float64 `json:"timestamp"`
	Score     float64 `json:"score"`
	Emotion   string  `json:"emotion"`
}

// EditingResult 剪辑结果
type EditingResult struct {
	OutputPath      string            `json:"output_path"`
	Duration        float64           `json:"duration"`
	FileSize        int64             `json:"file_size"`
	Resolution      string            `json:"resolution"`
	Bitrate         int               `json:"bitrate"`
	Segments        []EditedSegment   `json:"segments"`
	Transitions     []Transition      `json:"transitions"`
	Effects         []Effect          `json:"effects"`
	Subtitles       []Subtitle        `json:"subtitles"`
	BackgroundMusic string            `json:"background_music,omitempty"`
	Thumbnail       string            `json:"thumbnail"`
	Preview         string            `json:"preview"`
}

// EditedSegment 剪辑片段
type EditedSegment struct {
	OriginalStart float64 `json:"original_start"`
	OriginalEnd   float64 `json:"original_end"`
	EditedStart   float64 `json:"edited_start"`
	EditedEnd     float64 `json:"edited_end"`
	Type          string  `json:"type"`
	Importance    float64 `json:"importance"`
}

// Transition 转场效果
type Transition struct {
	Position float64 `json:"position"`
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
}

// Effect 特效
type Effect struct {
	StartTime float64                `json:"start_time"`
	EndTime   float64                `json:"end_time"`
	Type      string                 `json:"type"`
	Params    map[string]interface{} `json:"params"`
}

// Subtitle 字幕
type Subtitle struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Position  string  `json:"position"`
	Style     SubtitleStyle `json:"style"`
}

// SubtitleStyle 字幕样式
type SubtitleStyle struct {
	FontSize   int    `json:"font_size"`
	FontColor  string `json:"font_color"`
	Background string `json:"background"`
	Alignment  string `json:"alignment"`
}

// NewSmartEditingService 创建智能剪辑服务
func NewSmartEditingService(mongodb *database.MongoDB, storagePath string) (*SmartEditingService, error) {
	aiAnalyzer, err := NewAIContentAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI analyzer: %w", err)
	}

	editor, err := NewAutomaticEditor(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create automatic editor: %w", err)
	}

	return &SmartEditingService{
		mongodb:    mongodb,
		aiAnalyzer: aiAnalyzer,
		editor:     editor,
		storage:    storagePath,
	}, nil
}

// SubmitEditingTask 提交剪辑任务
func (s *SmartEditingService) SubmitEditingTask(task EditingTask) error {
	task.ID = uuid.New().String()
	task.Status = "pending"
	task.Progress = 0.0
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	collection := s.mongodb.Database.Collection("editing_tasks")
	_, err := collection.InsertOne(context.Background(), task)
	if err != nil {
		return fmt.Errorf("failed to insert editing task: %w", err)
	}

	// 异步处理任务
	go s.processEditingTask(task.ID)

	return nil
}

// processEditingTask 处理剪辑任务
func (s *SmartEditingService) processEditingTask(taskID string) {
	log.Printf("开始处理剪辑任务: %s", taskID)

	// 获取任务
	task, err := s.getEditingTask(taskID)
	if err != nil {
		log.Printf("获取任务失败: %v", err)
		return
	}

	// 更新状态为分析中
	s.updateTaskStatus(taskID, "analyzing", 10.0, "")

	// 1. 内容分析阶段
	analysis, err := s.aiAnalyzer.AnalyzeContent(task.VideoPath)
	if err != nil {
		s.updateTaskStatus(taskID, "failed", 0.0, err.Error())
		return
	}

	task.Analysis = analysis
	s.updateTask(task)
	s.updateTaskStatus(taskID, "analyzing", 50.0, "")

	// 2. 智能剪辑阶段
	s.updateTaskStatus(taskID, "editing", 60.0, "")

	result, err := s.editor.CreateSmartEdit(task.VideoPath, *analysis, task.Config)
	if err != nil {
		s.updateTaskStatus(taskID, "failed", 0.0, err.Error())
		return
	}

	task.Result = result
	s.updateTask(task)
	s.updateTaskStatus(taskID, "completed", 100.0, "")

	log.Printf("剪辑任务完成: %s", taskID)
}

// getEditingTask 获取剪辑任务
func (s *SmartEditingService) getEditingTask(taskID string) (*EditingTask, error) {
	collection := s.mongodb.Database.Collection("editing_tasks")
	var task EditingTask

	err := collection.FindOne(context.Background(), map[string]string{"_id": taskID}).Decode(&task)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &task, nil
}

// updateTaskStatus 更新任务状态
func (s *SmartEditingService) updateTaskStatus(taskID string, status string, progress float64, errorMsg string) {
	collection := s.mongodb.Database.Collection("editing_tasks")

	update := map[string]interface{}{
		"status":     status,
		"progress":   progress,
		"updated_at": time.Now(),
	}

	if errorMsg != "" {
		update["error"] = errorMsg
	}

	_, err := collection.UpdateOne(
		context.Background(),
		map[string]string{"_id": taskID},
		map[string]interface{}{"$set": update},
	)

	if err != nil {
		log.Printf("Failed to update task status: %v", err)
	}
}

// updateTask 更新完整任务
func (s *SmartEditingService) updateTask(task *EditingTask) {
	collection := s.mongodb.Database.Collection("editing_tasks")
	task.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(
		context.Background(),
		map[string]string{"_id": task.ID},
		task,
	)

	if err != nil {
		log.Printf("Failed to update task: %v", err)
	}
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

	// 创建智能剪辑服务
	editingService, err := NewSmartEditingService(mongodb, "./storage/editing")
	if err != nil {
		log.Fatalf("Failed to create editing service: %v", err)
	}

	// 设置路由
	router := setupRoutes(editingService)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("Smart editing service starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(service *SmartEditingService) *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "smart-editing-service",
			"timestamp": time.Now(),
		})
	})

	// API路由
	api := router.Group("/api/v1")
	{
		// 提交剪辑任务
		api.POST("/editing/submit", func(c *gin.Context) {
			var task EditingTask
			if err := c.ShouldBindJSON(&task); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := service.SubmitEditingTask(task); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"task_id": task.ID, "status": "submitted"})
		})

		// 获取任务状态
		api.GET("/editing/status/:task_id", func(c *gin.Context) {
			taskID := c.Param("task_id")
			task, err := service.getEditingTask(taskID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			c.JSON(http.StatusOK, task)
		})

		// 获取任务列表
		api.GET("/editing/tasks", func(c *gin.Context) {
			meetingID := c.Query("meeting_id")
			status := c.Query("status")

			tasks, err := service.getEditingTasks(meetingID, status)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"tasks": tasks})
		})

		// 取消任务
		api.DELETE("/editing/cancel/:task_id", func(c *gin.Context) {
			taskID := c.Param("task_id")

			if err := service.cancelEditingTask(taskID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Task cancelled"})
		})

		// 下载结果
		api.GET("/editing/download/:task_id", func(c *gin.Context) {
			taskID := c.Param("task_id")

			task, err := service.getEditingTask(taskID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			if task.Status != "completed" || task.Result == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Task not completed"})
				return
			}

			c.File(task.Result.OutputPath)
		})
	}

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
