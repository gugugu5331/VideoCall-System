package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"meeting-system/shared/ai"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// AIEnhancedService AI增强的会议服务
type AIEnhancedService struct {
	meetingService *MeetingService
	aiClient       *ai.AIClient
	config         *config.Config
	
	// AI处理状态跟踪
	processingTasks map[string]*AITask
	tasksMutex      sync.RWMutex
	
	// 降级策略配置
	fallbackEnabled bool
	maxRetries      int
	retryDelay      time.Duration
}

// AITask AI处理任务
type AITask struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	MeetingID   uint                   `json:"meeting_id"`
	UserID      uint                   `json:"user_id"`
	Status      string                 `json:"status"` // pending, processing, completed, failed
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Retries     int                    `json:"retries"`
}

// AIProcessingResult AI处理结果
type AIProcessingResult struct {
	TaskID    string                 `json:"task_id"`
	Type      string                 `json:"type"`
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Latency   float64                `json:"latency"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewAIEnhancedService 创建AI增强服务
func NewAIEnhancedService(config *config.Config) *AIEnhancedService {
	return &AIEnhancedService{
		meetingService:  NewMeetingService(),
		aiClient:        ai.NewAIClient(config),
		config:          config,
		processingTasks: make(map[string]*AITask),
		fallbackEnabled: true,
		maxRetries:      3,
		retryDelay:      time.Second,
	}
}

// ProcessSpeechRecognition 处理语音识别
func (s *AIEnhancedService) ProcessSpeechRecognition(ctx context.Context, meetingID, userID uint, audioData []byte, format string, sampleRate int) (*AIProcessingResult, error) {
	taskID := s.generateTaskID("speech_recognition", meetingID, userID)
	
	// 创建任务
	task := &AITask{
		ID:        taskID,
		Type:      "speech_recognition",
		MeetingID: meetingID,
		UserID:    userID,
		Status:    "pending",
		StartTime: time.Now(),
		Input: map[string]interface{}{
			"audio_format": format,
			"sample_rate":  sampleRate,
			"audio_size":   len(audioData),
		},
	}
	
	s.trackTask(task)
	defer s.removeTask(taskID)
	
	// 执行AI处理
	result := s.executeWithFallback(ctx, func() (*ai.AIResponse, error) {
		task.Status = "processing"
		return s.aiClient.SpeechRecognition(ctx, audioData, format, sampleRate)
	}, task)
	
	// 构建结果
	processingResult := &AIProcessingResult{
		TaskID:    taskID,
		Type:      "speech_recognition",
		Success:   result.Error == "",
		Latency:   result.Latency,
		Timestamp: time.Now(),
	}

	if result.Error != "" {
		processingResult.Error = result.Error
		task.Status = "failed"
		task.Error = result.Error
	} else {
		processingResult.Data = result.Data
		task.Status = "completed"
		task.Output = result.Data
	}
	
	now := time.Now()
	task.EndTime = &now
	
	// 记录处理结果
	s.logAIProcessing(processingResult)
	
	return processingResult, nil
}

// ProcessEmotionDetection 处理情绪识别
func (s *AIEnhancedService) ProcessEmotionDetection(ctx context.Context, meetingID, userID uint, imageData []byte, format string, width, height int) (*AIProcessingResult, error) {
	taskID := s.generateTaskID("emotion_detection", meetingID, userID)
	
	task := &AITask{
		ID:        taskID,
		Type:      "emotion_detection",
		MeetingID: meetingID,
		UserID:    userID,
		Status:    "pending",
		StartTime: time.Now(),
		Input: map[string]interface{}{
			"image_format": format,
			"width":        width,
			"height":       height,
			"image_size":   len(imageData),
		},
	}
	
	s.trackTask(task)
	defer s.removeTask(taskID)
	
	// 执行AI处理
	result := s.executeWithFallback(ctx, func() (*ai.AIResponse, error) {
		task.Status = "processing"
		return s.aiClient.EmotionDetection(ctx, imageData, format, width, height)
	}, task)
	
	// 构建结果
	processingResult := &AIProcessingResult{
		TaskID:    taskID,
		Type:      "emotion_detection",
		Success:   result.Error == "",
		Latency:   result.Latency,
		Timestamp: time.Now(),
	}

	if result.Error != "" {
		processingResult.Error = result.Error
		task.Status = "failed"
		task.Error = result.Error
	} else {
		processingResult.Data = result.Data
		task.Status = "completed"
		task.Output = result.Data
	}
	
	now := time.Now()
	task.EndTime = &now
	
	s.logAIProcessing(processingResult)
	return processingResult, nil
}

// ProcessAudioDenoising 处理音频降噪
func (s *AIEnhancedService) ProcessAudioDenoising(ctx context.Context, meetingID, userID uint, audioData []byte, format string, sampleRate int) (*AIProcessingResult, error) {
	taskID := s.generateTaskID("audio_denoising", meetingID, userID)
	
	task := &AITask{
		ID:        taskID,
		Type:      "audio_denoising",
		MeetingID: meetingID,
		UserID:    userID,
		Status:    "pending",
		StartTime: time.Now(),
		Input: map[string]interface{}{
			"audio_format": format,
			"sample_rate":  sampleRate,
			"audio_size":   len(audioData),
		},
	}
	
	s.trackTask(task)
	defer s.removeTask(taskID)
	
	// 执行AI处理
	result := s.executeWithFallback(ctx, func() (*ai.AIResponse, error) {
		task.Status = "processing"
		return s.aiClient.AudioDenoising(ctx, audioData, format, sampleRate)
	}, task)
	
	// 构建结果
	processingResult := &AIProcessingResult{
		TaskID:    taskID,
		Type:      "audio_denoising",
		Success:   result.Error == "",
		Latency:   result.Latency,
		Timestamp: time.Now(),
	}

	if result.Error != "" {
		processingResult.Error = result.Error
		task.Status = "failed"
		task.Error = result.Error
	} else {
		processingResult.Data = result.Data
		task.Status = "completed"
		task.Output = result.Data
	}
	
	now := time.Now()
	task.EndTime = &now
	
	s.logAIProcessing(processingResult)
	return processingResult, nil
}

// ProcessVideoEnhancement 处理视频增强
func (s *AIEnhancedService) ProcessVideoEnhancement(ctx context.Context, meetingID, userID uint, videoData []byte, format string, width, height, fps int) (*AIProcessingResult, error) {
	taskID := s.generateTaskID("video_enhancement", meetingID, userID)
	
	task := &AITask{
		ID:        taskID,
		Type:      "video_enhancement",
		MeetingID: meetingID,
		UserID:    userID,
		Status:    "pending",
		StartTime: time.Now(),
		Input: map[string]interface{}{
			"video_format": format,
			"width":        width,
			"height":       height,
			"fps":          fps,
			"video_size":   len(videoData),
		},
	}
	
	s.trackTask(task)
	defer s.removeTask(taskID)
	
	// 执行AI处理
	result := s.executeWithFallback(ctx, func() (*ai.AIResponse, error) {
		task.Status = "processing"
		return s.aiClient.VideoEnhancement(ctx, videoData, format, width, height, fps)
	}, task)
	
	// 构建结果
	processingResult := &AIProcessingResult{
		TaskID:    taskID,
		Type:      "video_enhancement",
		Success:   result.Error == "",
		Latency:   result.Latency,
		Timestamp: time.Now(),
	}

	if result.Error != "" {
		processingResult.Error = result.Error
		task.Status = "failed"
		task.Error = result.Error
	} else {
		processingResult.Data = result.Data
		task.Status = "completed"
		task.Output = result.Data
	}
	
	now := time.Now()
	task.EndTime = &now
	
	s.logAIProcessing(processingResult)
	return processingResult, nil
}

// executeWithFallback 执行AI请求并支持降级策略
func (s *AIEnhancedService) executeWithFallback(ctx context.Context, aiFunc func() (*ai.AIResponse, error), task *AITask) *ai.AIResponse {
	var lastErr error

	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			// 重试延迟
			select {
			case <-ctx.Done():
				return &ai.AIResponse{Error: "context cancelled"}
			case <-time.After(s.retryDelay * time.Duration(attempt)):
			}

			task.Retries = attempt
			logger.Warn(fmt.Sprintf("Retrying AI request (attempt %d/%d) for task %s", attempt+1, s.maxRetries+1, task.ID))
		}

		// 设置超时
		_, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		response, err := aiFunc()

		if err == nil && response.Error == "" {
			return response
		}

		lastErr = err
		if err != nil {
			logger.Error(fmt.Sprintf("AI request failed (attempt %d): %v", attempt+1, err))
		} else if response.Error != "" {
			logger.Error(fmt.Sprintf("AI request returned error (attempt %d): %s", attempt+1, response.Error))
		}
	}

	// 所有重试都失败，返回降级响应
	if s.fallbackEnabled {
		return s.getFallbackResponse(task.Type, lastErr)
	}

	errorMsg := "AI request failed after all retries"
	if lastErr != nil {
		errorMsg = lastErr.Error()
	}

	return &ai.AIResponse{
		RequestID: task.ID,
		Type:      task.Type,
		Status:    "error",
		Error:     errorMsg,
		Latency:   0,
	}
}

// getFallbackResponse 获取降级响应
func (s *AIEnhancedService) getFallbackResponse(requestType string, err error) *ai.AIResponse {
	logger.Warn(fmt.Sprintf("Using fallback response for %s due to: %v", requestType, err))

	var fallbackData map[string]interface{}

	switch requestType {
	case "speech_recognition":
		fallbackData = map[string]interface{}{
			"text":       "",
			"confidence": 0.0,
			"fallback":   true,
			"message":    "Speech recognition service unavailable",
		}
	case "emotion_detection":
		fallbackData = map[string]interface{}{
			"emotion":    "neutral",
			"confidence": 0.0,
			"fallback":   true,
			"message":    "Emotion detection service unavailable",
		}
	case "audio_denoising":
		fallbackData = map[string]interface{}{
			"processed":  false,
			"fallback":   true,
			"message":    "Audio denoising service unavailable",
		}
	case "video_enhancement":
		fallbackData = map[string]interface{}{
			"enhanced":   false,
			"fallback":   true,
			"message":    "Video enhancement service unavailable",
		}
	default:
		fallbackData = map[string]interface{}{
			"fallback": true,
			"message":  "AI service unavailable",
		}
	}

	return &ai.AIResponse{
		Type:    requestType,
		Status:  "fallback",
		Data:    fallbackData,
		Latency: 0,
	}
}

// trackTask 跟踪任务
func (s *AIEnhancedService) trackTask(task *AITask) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()
	s.processingTasks[task.ID] = task
}

// removeTask 移除任务
func (s *AIEnhancedService) removeTask(taskID string) {
	s.tasksMutex.Lock()
	defer s.tasksMutex.Unlock()
	delete(s.processingTasks, taskID)
}

// GetTaskStatus 获取任务状态
func (s *AIEnhancedService) GetTaskStatus(taskID string) (*AITask, error) {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	task, exists := s.processingTasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListActiveTasks 列出活跃任务
func (s *AIEnhancedService) ListActiveTasks() []*AITask {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	tasks := make([]*AITask, 0, len(s.processingTasks))
	for _, task := range s.processingTasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// generateTaskID 生成任务ID
func (s *AIEnhancedService) generateTaskID(taskType string, meetingID, userID uint) string {
	return fmt.Sprintf("%s_%d_%d_%d", taskType, meetingID, userID, time.Now().UnixNano())
}

// logAIProcessing 记录AI处理结果
func (s *AIEnhancedService) logAIProcessing(result *AIProcessingResult) {
	// 记录到日志
	logData, _ := json.Marshal(result)
	logger.Info(fmt.Sprintf("AI processing result: %s", string(logData)))

	// 这里可以添加更多的日志记录逻辑，比如：
	// - 发送到监控系统
	// - 保存到数据库
	// - 发送到消息队列
}

// GetAIProcessingStats 获取AI处理统计
func (s *AIEnhancedService) GetAIProcessingStats() map[string]interface{} {
	s.tasksMutex.RLock()
	defer s.tasksMutex.RUnlock()

	stats := map[string]interface{}{
		"active_tasks": len(s.processingTasks),
		"task_types":   make(map[string]int),
		"task_status":  make(map[string]int),
	}

	taskTypes := stats["task_types"].(map[string]int)
	taskStatus := stats["task_status"].(map[string]int)

	for _, task := range s.processingTasks {
		taskTypes[task.Type]++
		taskStatus[task.Status]++
	}

	return stats
}
