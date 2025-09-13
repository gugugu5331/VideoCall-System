package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"video-conference-system/shared/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// MediaTask 媒体处理任务
type MediaTask struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // encode, decode, transcode, record
	InputPath   string                 `json:"input_path"`
	OutputPath  string                 `json:"output_path"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"` // pending, processing, completed, failed
	Progress    float64                `json:"progress"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Error       string                 `json:"error,omitempty"`
}

// MediaService 媒体服务
type MediaService struct {
	taskQueue   chan MediaTask
	tasks       map[string]*MediaTask
	rabbitConn  *amqp.Connection
	rabbitCh    *amqp.Channel
	workersNum  int
	storagePath string
}

// NewMediaService 创建媒体服务
func NewMediaService(rabbitURL string, workersNum int, storagePath string) (*MediaService, error) {
	// 连接RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明队列
	_, err = ch.QueueDeclare(
		"media_tasks", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// 创建存储目录
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	service := &MediaService{
		taskQueue:   make(chan MediaTask, 100),
		tasks:       make(map[string]*MediaTask),
		rabbitConn:  conn,
		rabbitCh:    ch,
		workersNum:  workersNum,
		storagePath: storagePath,
	}

	// 启动工作协程
	for i := 0; i < workersNum; i++ {
		go service.worker(i)
	}

	// 启动消息消费者
	go service.consumeMessages()

	return service, nil
}

// Close 关闭服务
func (s *MediaService) Close() error {
	if s.rabbitCh != nil {
		s.rabbitCh.Close()
	}
	if s.rabbitConn != nil {
		s.rabbitConn.Close()
	}
	return nil
}

// SubmitTask 提交任务
func (s *MediaService) SubmitTask(task MediaTask) error {
	task.ID = uuid.New().String()
	task.Status = "pending"
	task.CreatedAt = time.Now()

	s.tasks[task.ID] = &task

	// 发送到队列
	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = s.rabbitCh.Publish(
		"",            // exchange
		"media_tasks", // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	return err
}

// GetTask 获取任务状态
func (s *MediaService) GetTask(taskID string) (*MediaTask, bool) {
	task, exists := s.tasks[taskID]
	return task, exists
}

// worker 工作协程
func (s *MediaService) worker(id int) {
	log.Printf("Media worker %d started", id)

	for task := range s.taskQueue {
		log.Printf("Worker %d processing task %s", id, task.ID)

		// 更新任务状态
		if taskPtr, exists := s.tasks[task.ID]; exists {
			taskPtr.Status = "processing"
			now := time.Now()
			taskPtr.StartedAt = &now
		}

		// 处理任务
		err := s.processTask(&task)

		// 更新任务状态
		if taskPtr, exists := s.tasks[task.ID]; exists {
			now := time.Now()
			taskPtr.CompletedAt = &now

			if err != nil {
				taskPtr.Status = "failed"
				taskPtr.Error = err.Error()
				log.Printf("Task %s failed: %v", task.ID, err)
			} else {
				taskPtr.Status = "completed"
				taskPtr.Progress = 100.0
				log.Printf("Task %s completed", task.ID)
			}
		}
	}
}

// consumeMessages 消费RabbitMQ消息
func (s *MediaService) consumeMessages() {
	msgs, err := s.rabbitCh.Consume(
		"media_tasks", // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	for msg := range msgs {
		var task MediaTask
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			continue
		}

		select {
		case s.taskQueue <- task:
		default:
			log.Printf("Task queue is full, dropping task %s", task.ID)
		}
	}
}

// processTask 处理任务
func (s *MediaService) processTask(task *MediaTask) error {
	switch task.Type {
	case "encode":
		return s.encodeVideo(task)
	case "decode":
		return s.decodeVideo(task)
	case "transcode":
		return s.transcodeVideo(task)
	case "record":
		return s.recordStream(task)
	case "extract_frames":
		return s.extractFrames(task)
	case "extract_audio":
		return s.extractAudio(task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// encodeVideo 编码视频
func (s *MediaService) encodeVideo(task *MediaTask) error {
	codec := task.Parameters["codec"].(string)
	bitrate := task.Parameters["bitrate"].(string)
	resolution := task.Parameters["resolution"].(string)

	args := []string{
		"-i", task.InputPath,
		"-c:v", codec,
		"-b:v", bitrate,
		"-s", resolution,
		"-y", task.OutputPath,
	}

	return s.runFFmpeg(args, task)
}

// decodeVideo 解码视频
func (s *MediaService) decodeVideo(task *MediaTask) error {
	args := []string{
		"-i", task.InputPath,
		"-c:v", "rawvideo",
		"-pix_fmt", "yuv420p",
		"-y", task.OutputPath,
	}

	return s.runFFmpeg(args, task)
}

// transcodeVideo 转码视频
func (s *MediaService) transcodeVideo(task *MediaTask) error {
	outputFormat := task.Parameters["format"].(string)
	quality := task.Parameters["quality"].(string)

	args := []string{
		"-i", task.InputPath,
		"-c:v", "libx264",
		"-preset", quality,
		"-f", outputFormat,
		"-y", task.OutputPath,
	}

	return s.runFFmpeg(args, task)
}

// recordStream 录制流
func (s *MediaService) recordStream(task *MediaTask) error {
	duration := task.Parameters["duration"].(float64)
	format := task.Parameters["format"].(string)

	args := []string{
		"-i", task.InputPath,
		"-t", fmt.Sprintf("%.2f", duration),
		"-c", "copy",
		"-f", format,
		"-y", task.OutputPath,
	}

	return s.runFFmpeg(args, task)
}

// extractFrames 提取帧
func (s *MediaService) extractFrames(task *MediaTask) error {
	fps := task.Parameters["fps"].(float64)
	format := task.Parameters["format"].(string)

	args := []string{
		"-i", task.InputPath,
		"-vf", fmt.Sprintf("fps=%.2f", fps),
		"-f", "image2",
		"-q:v", "2",
		task.OutputPath + "/frame_%04d." + format,
	}

	// 创建输出目录
	if err := os.MkdirAll(task.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return s.runFFmpeg(args, task)
}

// extractAudio 提取音频
func (s *MediaService) extractAudio(task *MediaTask) error {
	codec := task.Parameters["codec"].(string)
	bitrate := task.Parameters["bitrate"].(string)

	args := []string{
		"-i", task.InputPath,
		"-vn", // 不包含视频
		"-c:a", codec,
		"-b:a", bitrate,
		"-y", task.OutputPath,
	}

	return s.runFFmpeg(args, task)
}

// runFFmpeg 运行FFmpeg命令
func (s *MediaService) runFFmpeg(args []string, task *MediaTask) error {
	cmd := exec.Command("ffmpeg", args...)

	// 设置环境变量
	cmd.Env = os.Environ()

	// 创建管道获取进度信息
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// 监控进度
	go s.monitorProgress(stderr, task)

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg command failed: %w", err)
	}

	return nil
}

// monitorProgress 监控FFmpeg进度
func (s *MediaService) monitorProgress(stderr io.ReadCloser, task *MediaTask) {
	defer stderr.Close()

	buf := make([]byte, 1024)
	for {
		n, err := stderr.Read(buf)
		if err != nil {
			break
		}

		output := string(buf[:n])
		
		// 解析进度信息
		if strings.Contains(output, "time=") {
			progress := s.parseProgress(output)
			if taskPtr, exists := s.tasks[task.ID]; exists {
				taskPtr.Progress = progress
			}
		}
	}
}

// parseProgress 解析FFmpeg进度
func (s *MediaService) parseProgress(output string) float64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "time=") {
			parts := strings.Split(line, "time=")
			if len(parts) > 1 {
				timeStr := strings.Fields(parts[1])[0]
				// 这里可以根据总时长计算百分比
				// 简化处理，返回固定值
				return 50.0
			}
		}
	}
	return 0.0
}

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 创建媒体服务
	mediaService, err := NewMediaService(
		cfg.RabbitMQ.URL,
		4, // 4个工作协程
		"./storage/media",
	)
	if err != nil {
		log.Fatalf("Failed to create media service: %v", err)
	}
	defer mediaService.Close()

	// 设置路由
	router := setupRoutes(mediaService)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("Media service starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(mediaService *MediaService) *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "media-service",
			"timestamp": time.Now(),
		})
	})

	// API路由
	api := router.Group("/api/v1")
	{
		// 提交任务
		api.POST("/tasks", func(c *gin.Context) {
			var task MediaTask
			if err := c.ShouldBindJSON(&task); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := mediaService.SubmitTask(task); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{
				"message": "Task submitted successfully",
				"task_id": task.ID,
			})
		})

		// 获取任务状态
		api.GET("/tasks/:id", func(c *gin.Context) {
			taskID := c.Param("id")
			task, exists := mediaService.GetTask(taskID)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"task": task,
			})
		})

		// 文件上传
		api.POST("/upload", func(c *gin.Context) {
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
				return
			}

			// 生成唯一文件名
			filename := uuid.New().String() + filepath.Ext(file.Filename)
			filepath := filepath.Join(mediaService.storagePath, "uploads", filename)

			// 创建目录
			if err := os.MkdirAll(filepath.Dir(filepath), 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
				return
			}

			// 保存文件
			if err := c.SaveUploadedFile(file, filepath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  "File uploaded successfully",
				"filename": filename,
				"path":     filepath,
			})
		})
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
