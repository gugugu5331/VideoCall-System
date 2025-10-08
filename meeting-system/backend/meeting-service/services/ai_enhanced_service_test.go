package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"meeting-system/shared/config"
)

func TestNewAIEnhancedService(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	assert.NotNil(t, service)
	assert.NotNil(t, service.meetingService)
	assert.NotNil(t, service.aiClient)
	assert.NotNil(t, service.processingTasks)
	assert.True(t, service.fallbackEnabled)
	assert.Equal(t, 3, service.maxRetries)
	assert.Equal(t, time.Second, service.retryDelay)
}

func TestAIEnhancedService_ProcessSpeechRecognition(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	ctx := context.Background()
	audioData := []byte("mock audio data")
	
	result, err := service.ProcessSpeechRecognition(ctx, 1, 1, audioData, "wav", 16000)
	
	// 由于AI服务可能不可用，我们期望得到降级响应
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "speech_recognition", result.Type)
	assert.NotEmpty(t, result.TaskID)
	
	// 检查是否是降级响应
	if !result.Success {
		assert.NotEmpty(t, result.Error)
	} else if result.Data != nil {
		// 如果成功，检查数据结构
		if fallback, exists := result.Data["fallback"]; exists && fallback.(bool) {
			assert.Contains(t, result.Data, "message")
		}
	}
}

func TestAIEnhancedService_ProcessEmotionDetection(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	ctx := context.Background()
	imageData := []byte("mock image data")
	
	result, err := service.ProcessEmotionDetection(ctx, 1, 1, imageData, "jpg", 640, 480)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "emotion_detection", result.Type)
	assert.NotEmpty(t, result.TaskID)
}

func TestAIEnhancedService_ProcessAudioDenoising(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	ctx := context.Background()
	audioData := []byte("mock audio data")
	
	result, err := service.ProcessAudioDenoising(ctx, 1, 1, audioData, "wav", 16000)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "audio_denoising", result.Type)
	assert.NotEmpty(t, result.TaskID)
}

func TestAIEnhancedService_ProcessVideoEnhancement(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	ctx := context.Background()
	videoData := []byte("mock video data")
	
	result, err := service.ProcessVideoEnhancement(ctx, 1, 1, videoData, "mp4", 1920, 1080, 30)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "video_enhancement", result.Type)
	assert.NotEmpty(t, result.TaskID)
}

func TestAIEnhancedService_GetFallbackResponse(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	testCases := []struct {
		requestType string
		expected    map[string]interface{}
	}{
		{
			requestType: "speech_recognition",
			expected: map[string]interface{}{
				"text":       "",
				"confidence": 0.0,
				"fallback":   true,
				"message":    "Speech recognition service unavailable",
			},
		},
		{
			requestType: "emotion_detection",
			expected: map[string]interface{}{
				"emotion":    "neutral",
				"confidence": 0.0,
				"fallback":   true,
				"message":    "Emotion detection service unavailable",
			},
		},
		{
			requestType: "audio_denoising",
			expected: map[string]interface{}{
				"processed": false,
				"fallback":  true,
				"message":   "Audio denoising service unavailable",
			},
		},
		{
			requestType: "video_enhancement",
			expected: map[string]interface{}{
				"enhanced": false,
				"fallback": true,
				"message":  "Video enhancement service unavailable",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.requestType, func(t *testing.T) {
			response := service.getFallbackResponse(tc.requestType, nil)
			
			assert.Equal(t, tc.requestType, response.Type)
			assert.Equal(t, "fallback", response.Status)
			assert.Equal(t, tc.expected, response.Data)
		})
	}
}

func TestAIEnhancedService_TaskManagement(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	// 创建测试任务
	task := &AITask{
		ID:        "test_task_001",
		Type:      "speech_recognition",
		MeetingID: 1,
		UserID:    1,
		Status:    "pending",
		StartTime: time.Now(),
	}
	
	// 跟踪任务
	service.trackTask(task)
	
	// 获取任务状态
	retrievedTask, err := service.GetTaskStatus("test_task_001")
	require.NoError(t, err)
	assert.Equal(t, task.ID, retrievedTask.ID)
	assert.Equal(t, task.Type, retrievedTask.Type)
	
	// 列出活跃任务
	activeTasks := service.ListActiveTasks()
	assert.Len(t, activeTasks, 1)
	assert.Equal(t, "test_task_001", activeTasks[0].ID)
	
	// 移除任务
	service.removeTask("test_task_001")
	
	// 验证任务已移除
	_, err = service.GetTaskStatus("test_task_001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
	
	// 验证活跃任务列表为空
	activeTasks = service.ListActiveTasks()
	assert.Len(t, activeTasks, 0)
}

func TestAIEnhancedService_GetAIProcessingStats(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	// 添加一些测试任务
	tasks := []*AITask{
		{
			ID:     "task1",
			Type:   "speech_recognition",
			Status: "processing",
		},
		{
			ID:     "task2",
			Type:   "emotion_detection",
			Status: "completed",
		},
		{
			ID:     "task3",
			Type:   "speech_recognition",
			Status: "failed",
		},
	}
	
	for _, task := range tasks {
		service.trackTask(task)
	}
	
	// 获取统计信息
	stats := service.GetAIProcessingStats()
	
	assert.Equal(t, 3, stats["active_tasks"])
	
	taskTypes := stats["task_types"].(map[string]int)
	assert.Equal(t, 2, taskTypes["speech_recognition"])
	assert.Equal(t, 1, taskTypes["emotion_detection"])
	
	taskStatus := stats["task_status"].(map[string]int)
	assert.Equal(t, 1, taskStatus["processing"])
	assert.Equal(t, 1, taskStatus["completed"])
	assert.Equal(t, 1, taskStatus["failed"])
}

func TestAIEnhancedService_GenerateTaskID(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	taskID1 := service.generateTaskID("speech_recognition", 1, 1)
	taskID2 := service.generateTaskID("speech_recognition", 1, 1)
	
	// 任务ID应该不同（因为包含纳秒时间戳）
	assert.NotEqual(t, taskID1, taskID2)
	
	// 任务ID应该包含类型和ID信息
	assert.Contains(t, taskID1, "speech_recognition")
	assert.Contains(t, taskID1, "1_1")
}

func TestAIEnhancedService_ContextCancellation(t *testing.T) {
	cfg := &config.Config{}
	service := NewAIEnhancedService(cfg)
	
	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	
	// 立即取消上下文
	cancel()
	
	audioData := []byte("mock audio data")
	
	// 处理应该快速返回（由于上下文已取消）
	start := time.Now()
	result, err := service.ProcessSpeechRecognition(ctx, 1, 1, audioData, "wav", 16000)
	duration := time.Since(start)
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// 应该很快返回（不应该等待重试延迟）
	assert.Less(t, duration, 5*time.Second)
	
	// 应该是失败的结果
	assert.False(t, result.Success)
}
