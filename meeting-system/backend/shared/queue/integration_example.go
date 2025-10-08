package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// IntegrationExample 展示如何集成和使用消息队列系统
type IntegrationExample struct {
	queueManager *QueueManager
	redisClient  *redis.Client
}

// NewIntegrationExample 创建集成示例
func NewIntegrationExample(cfg *config.Config, redisClient *redis.Client) *IntegrationExample {
	return &IntegrationExample{
		queueManager: NewQueueManager(cfg, redisClient),
		redisClient:  redisClient,
	}
}

// SetupQueues 设置所有队列组件
func (ie *IntegrationExample) SetupQueues() error {
	// 1. 初始化Redis消息队列
	ie.queueManager.InitRedisMessageQueue(ie.redisClient, RedisMessageQueueConfig{
		QueueName:         "ai_tasks",
		Workers:           4,
		VisibilityTimeout: 30 * time.Second,
		PollInterval:      100 * time.Millisecond,
	})
	
	// 2. 初始化Redis发布订阅队列
	ie.queueManager.InitRedisPubSubQueue(ie.redisClient)
	
	// 3. 初始化本地事件总线
	ie.queueManager.InitLocalEventBus(1000, 4)
	
	// 4. 初始化任务调度器
	ie.queueManager.InitTaskScheduler(1000, 8)
	
	// 5. 初始化任务分发器
	ie.queueManager.InitTaskDispatcher()
	
	// 启动所有组件
	if err := ie.queueManager.Start(); err != nil {
		return fmt.Errorf("failed to start queue manager: %w", err)
	}
	
	logger.Info("All queue components initialized and started")
	return nil
}

// RegisterHandlers 注册消息处理器
func (ie *IntegrationExample) RegisterHandlers() {
	// 1. 注册Redis消息队列处理器
	redisQueue := ie.queueManager.GetRedisMessageQueue()
	if redisQueue != nil {
		// AI任务处理器
		redisQueue.RegisterHandler("speech_recognition", ie.handleSpeechRecognition)
		redisQueue.RegisterHandler("emotion_detection", ie.handleEmotionDetection)
		redisQueue.RegisterHandler("audio_denoising", ie.handleAudioDenoising)
		
		logger.Info("Registered Redis message queue handlers")
	}
	
	// 2. 注册发布订阅处理器
	pubsubQueue := ie.queueManager.GetRedisPubSubQueue()
	if pubsubQueue != nil {
		// 会议事件处理器
		pubsubQueue.Subscribe("meeting_events", ie.handleMeetingEvent)
		pubsubQueue.Subscribe("media_events", ie.handleMediaEvent)
		pubsubQueue.Subscribe("ai_events", ie.handleAIEvent)
		
		logger.Info("Registered PubSub handlers")
	}
	
	// 3. 注册本地事件总线处理器
	localBus := ie.queueManager.GetLocalEventBus()
	if localBus != nil {
		localBus.On("stream_started", ie.handleStreamStarted)
		localBus.On("stream_stopped", ie.handleStreamStopped)
		
		logger.Info("Registered local event bus handlers")
	}
}

// 示例：发布消息到Redis队列
func (ie *IntegrationExample) PublishAITask(taskType string, payload map[string]interface{}) error {
	redisQueue := ie.queueManager.GetRedisMessageQueue()
	if redisQueue == nil {
		return fmt.Errorf("redis message queue not initialized")
	}
	
	msg := &Message{
		Type:       taskType,
		Priority:   PriorityHigh,
		Payload:    payload,
		MaxRetries: 3,
		Timeout:    30,
		Source:     "ai-service",
	}
	
	ctx := context.Background()
	return redisQueue.Publish(ctx, msg)
}

// 示例：发布事件到PubSub
func (ie *IntegrationExample) PublishMeetingEvent(eventType string, payload map[string]interface{}) error {
	pubsubQueue := ie.queueManager.GetRedisPubSubQueue()
	if pubsubQueue == nil {
		return fmt.Errorf("redis pubsub queue not initialized")
	}
	
	msg := &PubSubMessage{
		Type:    eventType,
		Payload: payload,
		Source:  "meeting-service",
	}
	
	ctx := context.Background()
	return pubsubQueue.Publish(ctx, "meeting_events", msg)
}

// 示例：触发本地事件
func (ie *IntegrationExample) EmitLocalEvent(eventType string, payload map[string]interface{}) error {
	localBus := ie.queueManager.GetLocalEventBus()
	if localBus == nil {
		return fmt.Errorf("local event bus not initialized")
	}
	
	return localBus.Emit(eventType, payload, "media-service")
}

// 示例：提交任务到调度器
func (ie *IntegrationExample) ScheduleTask(taskType string, payload map[string]interface{}, delay time.Duration) error {
	scheduler := ie.queueManager.GetTaskScheduler()
	if scheduler == nil {
		return fmt.Errorf("task scheduler not initialized")
	}
	
	task := &Task{
		Type:        taskType,
		Priority:    PriorityNormal,
		Payload:     payload,
		Handler:     ie.handleScheduledTask,
		Timeout:     60 * time.Second,
		MaxRetries:  3,
		ScheduledAt: time.Now().Add(delay),
	}
	
	return scheduler.SubmitTask(task)
}

// 示例：使用任务分发器
func (ie *IntegrationExample) DispatchTask(taskType TaskType, payload map[string]interface{}) (*TaskResponse, error) {
	dispatcher := ie.queueManager.GetTaskDispatcher()
	if dispatcher == nil {
		return nil, fmt.Errorf("task dispatcher not initialized")
	}
	
	request := &TaskRequest{
		Type:       taskType,
		Priority:   PriorityHigh,
		Payload:    payload,
		Timeout:    30,
		MaxRetries: 3,
	}
	
	ctx := context.Background()
	return dispatcher.DispatchTask(ctx, request)
}

// 获取统计信息
func (ie *IntegrationExample) GetStats() map[string]interface{} {
	return ie.queueManager.GetStats()
}

// 停止所有组件
func (ie *IntegrationExample) Shutdown() error {
	return ie.queueManager.Stop()
}

// ========== 消息处理器示例 ==========

func (ie *IntegrationExample) handleSpeechRecognition(ctx context.Context, msg *Message) error {
	logger.Info(fmt.Sprintf("Processing speech recognition task: %s", msg.ID))
	
	// 模拟处理
	time.Sleep(100 * time.Millisecond)
	
	// 处理完成后发布事件
	ie.PublishMeetingEvent("ai.speech_recognition.completed", map[string]interface{}{
		"task_id": msg.ID,
		"result":  "transcription result",
	})
	
	return nil
}

func (ie *IntegrationExample) handleEmotionDetection(ctx context.Context, msg *Message) error {
	logger.Info(fmt.Sprintf("Processing emotion detection task: %s", msg.ID))
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ie *IntegrationExample) handleAudioDenoising(ctx context.Context, msg *Message) error {
	logger.Info(fmt.Sprintf("Processing audio denoising task: %s", msg.ID))
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ie *IntegrationExample) handleMeetingEvent(ctx context.Context, msg *PubSubMessage) error {
	logger.Info(fmt.Sprintf("Received meeting event: %s", msg.Type))
	return nil
}

func (ie *IntegrationExample) handleMediaEvent(ctx context.Context, msg *PubSubMessage) error {
	logger.Info(fmt.Sprintf("Received media event: %s", msg.Type))
	return nil
}

func (ie *IntegrationExample) handleAIEvent(ctx context.Context, msg *PubSubMessage) error {
	logger.Info(fmt.Sprintf("Received AI event: %s", msg.Type))
	return nil
}

func (ie *IntegrationExample) handleStreamStarted(ctx context.Context, event *LocalEvent) error {
	logger.Info(fmt.Sprintf("Stream started: %v", event.Payload))
	return nil
}

func (ie *IntegrationExample) handleStreamStopped(ctx context.Context, event *LocalEvent) error {
	logger.Info(fmt.Sprintf("Stream stopped: %v", event.Payload))
	return nil
}

func (ie *IntegrationExample) handleScheduledTask(ctx context.Context, task *Task) error {
	logger.Info(fmt.Sprintf("Executing scheduled task: %s", task.Type))
	time.Sleep(100 * time.Millisecond)
	return nil
}

// ========== 完整使用示例 ==========

// RunCompleteExample 运行完整示例
func RunCompleteExample(cfg *config.Config, redisClient *redis.Client) error {
	// 1. 创建集成示例
	example := NewIntegrationExample(cfg, redisClient)
	
	// 2. 设置队列
	if err := example.SetupQueues(); err != nil {
		return err
	}
	defer example.Shutdown()
	
	// 3. 注册处理器
	example.RegisterHandlers()
	
	// 4. 发布AI任务
	example.PublishAITask("speech_recognition", map[string]interface{}{
		"audio_data": "base64_encoded_audio",
		"language":   "zh-CN",
	})
	
	// 5. 发布会议事件
	example.PublishMeetingEvent("meeting.user_joined", map[string]interface{}{
		"user_id":    123,
		"meeting_id": 456,
	})
	
	// 6. 触发本地事件
	example.EmitLocalEvent("stream_started", map[string]interface{}{
		"stream_id": "stream_123",
	})
	
	// 7. 调度延迟任务
	example.ScheduleTask("cleanup", map[string]interface{}{
		"resource": "temp_files",
	}, 5*time.Minute)
	
	// 8. 分发任务
	response, err := example.DispatchTask(TaskTypeSpeechRecognition, map[string]interface{}{
		"audio": "data",
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to dispatch task: %v", err))
	} else {
		logger.Info(fmt.Sprintf("Task dispatched: %s, status: %s", response.TaskID, response.Status))
	}
	
	// 9. 获取统计信息
	stats := example.GetStats()
	logger.Info(fmt.Sprintf("Queue stats: %+v", stats))
	
	// 等待一段时间让任务处理完成
	time.Sleep(2 * time.Second)
	
	return nil
}

