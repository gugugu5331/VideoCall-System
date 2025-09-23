package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"video-conference-system/shared/config"
	"video-conference-system/shared/database"
	"video-conference-system/shared/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// DetectionRequest 检测请求
type DetectionRequest struct {
	MeetingID string `json:"meeting_id,omitempty"`
	UserID    string `json:"user_id" binding:"required"`
	FileType  string `json:"file_type" binding:"required,oneof=image audio video"`
	Priority  int    `json:"priority,omitempty"`
}

// DetectionResponse 检测响应
type DetectionResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// AIDetectionResult AI检测结果
type AIDetectionResult struct {
	IsFake     bool                   `json:"is_fake"`
	Confidence float64                `json:"confidence"`
	Type       string                 `json:"detection_type"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// DetectionService 检测服务
type DetectionService struct {
	db          *database.MongoDB
	rabbitConn  *amqp.Connection
	rabbitCh    *amqp.Channel
	aiServiceURL string
	storagePath  string
}

// NewDetectionService 创建检测服务
func NewDetectionService(db *database.MongoDB, rabbitURL, aiServiceURL, storagePath string) (*DetectionService, error) {
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
		"detection_tasks", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// 创建存储目录
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	service := &DetectionService{
		db:           db,
		rabbitConn:   conn,
		rabbitCh:     ch,
		aiServiceURL: aiServiceURL,
		storagePath:  storagePath,
	}

	// 启动消息消费者
	go service.consumeDetectionTasks()

	return service, nil
}

// Close 关闭服务
func (s *DetectionService) Close() error {
	if s.rabbitCh != nil {
		s.rabbitCh.Close()
	}
	if s.rabbitConn != nil {
		s.rabbitConn.Close()
	}
	return nil
}

// SubmitDetectionTask 提交检测任务
func (s *DetectionService) SubmitDetectionTask(req DetectionRequest, filePath string) (*models.DetectionTask, error) {
	// 创建检测任务
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	task := &models.DetectionTask{
		UserID:   userID,
		FilePath: filePath,
		FileType: req.FileType,
		Status:   models.TaskStatusPending,
		Priority: req.Priority,
	}

	if req.MeetingID != "" {
		meetingID, err := uuid.Parse(req.MeetingID)
		if err != nil {
			return nil, fmt.Errorf("invalid meeting ID: %w", err)
		}
		task.MeetingID = &meetingID
	}

	if task.Priority == 0 {
		task.Priority = 5 // 默认优先级
	}

	// 保存到数据库
	collection := s.db.Collection("detection_tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// 发送到消息队列
	taskData, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	err = s.rabbitCh.Publish(
		"",                // exchange
		"detection_tasks", // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskData,
			Priority:    uint8(task.Priority),
		})

	if err != nil {
		return nil, fmt.Errorf("failed to publish task: %w", err)
	}

	log.Printf("Detection task %s submitted", task.ID.String())
	return task, nil
}

// GetDetectionTask 获取检测任务
func (s *DetectionService) GetDetectionTask(taskID string) (*models.DetectionTask, error) {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID: %w", err)
	}

	collection := s.db.Collection("detection_tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var task models.DetectionTask
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// consumeDetectionTasks 消费检测任务
func (s *DetectionService) consumeDetectionTasks() {
	msgs, err := s.rabbitCh.Consume(
		"detection_tasks", // queue
		"",                // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	for msg := range msgs {
		var task models.DetectionTask
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			msg.Nack(false, false)
			continue
		}

		// 处理任务
		if err := s.processDetectionTask(&task); err != nil {
			log.Printf("Failed to process task %s: %v", task.ID.String(), err)
			msg.Nack(false, true) // 重新入队
		} else {
			msg.Ack(false)
		}
	}
}

// processDetectionTask 处理检测任务
func (s *DetectionService) processDetectionTask(task *models.DetectionTask) error {
	log.Printf("Processing detection task %s", task.ID.String())

	// 更新任务状态为处理中
	if err := s.updateTaskStatus(task.ID.String(), models.TaskStatusProcessing); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 调用AI检测服务
	result, err := s.callAIDetectionService(task.FilePath, task.FileType)
	if err != nil {
		// 更新任务状态为失败
		s.updateTaskStatus(task.ID.String(), models.TaskStatusFailed)
		return fmt.Errorf("AI detection failed: %w", err)
	}

	// 保存检测结果
	detectionResult := &models.DetectionResult{
		TaskID:         task.ID,
		IsFake:         result.IsFake,
		Confidence:     result.Confidence,
		DetectionType:  result.Type,
		ProcessingTime: 1000, // 示例处理时间
	}

	if result.Details != nil {
		detailsJSON, _ := json.Marshal(result.Details)
		detectionResult.Details = string(detailsJSON)
	}

	// 保存到数据库
	collection := s.db.Collection("detection_results")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, detectionResult)
	if err != nil {
		return fmt.Errorf("failed to save detection result: %w", err)
	}

	// 更新任务状态为完成
	if err := s.updateTaskStatus(task.ID.String(), models.TaskStatusCompleted); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 如果检测到伪造，发送告警
	if result.IsFake && result.Confidence > 0.7 {
		s.sendDetectionAlert(task, result)
	}

	log.Printf("Detection task %s completed", task.ID.String())
	return nil
}

// callAIDetectionService 调用AI检测服务
func (s *DetectionService) callAIDetectionService(filePath, fileType string) (*AIDetectionResult, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// 添加文件类型
	writer.WriteField("type", fileType)
	writer.Close()

	// 发送HTTP请求
	url := fmt.Sprintf("%s/detect", s.aiServiceURL)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("AI service returned status %d", resp.StatusCode)
	}

	// 解析响应
	var response struct {
		TaskID string             `json:"task_id"`
		Status string             `json:"status"`
		Result *AIDetectionResult `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 如果是异步处理，轮询结果
	if response.Status == "processing" {
		return s.pollAIResult(response.TaskID)
	}

	if response.Result == nil {
		return nil, fmt.Errorf("no result returned from AI service")
	}

	return response.Result, nil
}

// pollAIResult 轮询AI检测结果
func (s *DetectionService) pollAIResult(taskID string) (*AIDetectionResult, error) {
	url := fmt.Sprintf("%s/result/%s", s.aiServiceURL, taskID)
	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < 30; i++ { // 最多等待5分钟
		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to poll result: %w", err)
		}

		var response struct {
			Status string             `json:"status"`
			Result *AIDetectionResult `json:"result"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		if response.Status == "completed" && response.Result != nil {
			return response.Result, nil
		}

		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("AI detection timeout")
}

// updateTaskStatus 更新任务状态
func (s *DetectionService) updateTaskStatus(taskID, status string) error {
	id, err := uuid.Parse(taskID)
	if err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	collection := s.db.Collection("detection_tasks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"status": status}}
	if status == models.TaskStatusProcessing {
		update["$set"].(bson.M)["started_at"] = time.Now()
	} else if status == models.TaskStatusCompleted || status == models.TaskStatusFailed {
		update["$set"].(bson.M)["completed_at"] = time.Now()
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// sendDetectionAlert 发送检测告警
func (s *DetectionService) sendDetectionAlert(task *models.DetectionTask, result *AIDetectionResult) {
	alert := models.DetectionAlert{
		UserID:        task.UserID.String(),
		DetectionType: result.Type,
		Confidence:    result.Confidence,
		Timestamp:     time.Now(),
	}

	if task.MeetingID != nil {
		alert.MeetingID = task.MeetingID.String()
	}

	if result.Details != nil {
		alert.Details = result.Details
	}

	// 发送到消息队列进行进一步处理
	alertData, _ := json.Marshal(alert)
	s.rabbitCh.Publish(
		"",               // exchange
		"detection_alerts", // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        alertData,
		})

	log.Printf("Detection alert sent for task %s", task.ID.String())
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

	// 创建检测服务
	detectionService, err := NewDetectionService(
		mongodb,
		cfg.RabbitMQ.URL,
		cfg.AI.ServiceURL,
		"./storage/detection",
	)
	if err != nil {
		log.Fatalf("Failed to create detection service: %v", err)
	}
	defer detectionService.Close()

	// 设置路由
	router := setupRoutes(detectionService)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("Detection service starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(detectionService *DetectionService) *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "detection-service",
			"timestamp": time.Now(),
		})
	})

	// API路由
	api := router.Group("/api/v1")
	{
		// 提交检测任务
		api.POST("/detection/analyze", func(c *gin.Context) {
			// 解析表单数据
			var req DetectionRequest
			if err := c.ShouldBind(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 处理文件上传
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
				return
			}

			// 生成唯一文件名
			filename := uuid.New().String() + filepath.Ext(file.Filename)
			filePath := filepath.Join(detectionService.storagePath, "uploads", filename)

			// 创建目录
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
				return
			}

			// 保存文件
			if err := c.SaveUploadedFile(file, filePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
				return
			}

			// 提交检测任务
			task, err := detectionService.SubmitDetectionTask(req, filePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusAccepted, DetectionResponse{
				TaskID:  task.ID.String(),
				Status:  "pending",
				Message: "Detection task submitted",
			})
		})

		// 获取检测结果
		api.GET("/detection/results/:task_id", func(c *gin.Context) {
			taskID := c.Param("task_id")
			task, err := detectionService.GetDetectionTask(taskID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"task": task.ToResponse(),
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
