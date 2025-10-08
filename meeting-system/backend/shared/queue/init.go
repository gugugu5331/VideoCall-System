package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// InitializeQueueSystem 初始化完整的队列系统
func InitializeQueueSystem(cfg *config.Config, redisClient *redis.Client) (*QueueManager, error) {
	logger.Info("Initializing queue system...")
	
	// 创建队列管理器
	qm := NewQueueManager(cfg, redisClient)
	
	// 根据配置初始化各个组件
	if cfg.MessageQueue.Enabled {
		if cfg.MessageQueue.Type == "redis" && redisClient != nil {
			logger.Info("Initializing Redis message queue...")
			qm.InitRedisMessageQueue(redisClient, RedisMessageQueueConfig{
				QueueName:         cfg.MessageQueue.QueueName,
				Workers:           cfg.MessageQueue.Workers,
				VisibilityTimeout: time.Duration(cfg.MessageQueue.VisibilityTimeout) * time.Second,
				PollInterval:      time.Duration(cfg.MessageQueue.PollInterval) * time.Millisecond,
			})
		} else {
			logger.Info("Initializing memory message queue...")
			qm.InitMessageQueue(cfg.MessageQueue.QueueName, 1000, cfg.MessageQueue.Workers)
		}
	}
	
	// 初始化发布订阅队列
	if cfg.EventBus.Enabled {
		if cfg.EventBus.Type == "redis_pubsub" && redisClient != nil {
			logger.Info("Initializing Redis PubSub queue...")
			qm.InitRedisPubSubQueue(redisClient)
		}
		
		// 同时初始化本地事件总线（用于服务内部事件）
		logger.Info("Initializing local event bus...")
		qm.InitLocalEventBus(cfg.EventBus.BufferSize, cfg.EventBus.Workers)
	}
	
	// 初始化任务调度器
	if cfg.TaskScheduler.Enabled {
		logger.Info("Initializing task scheduler...")
		qm.InitTaskScheduler(cfg.TaskScheduler.BufferSize, cfg.TaskScheduler.Workers)
	}
	
	// 初始化任务分发器
	if cfg.TaskDispatcher.Enabled {
		logger.Info("Initializing task dispatcher...")
		qm.InitTaskDispatcher()
	}
	
	// 启动所有组件
	if err := qm.Start(); err != nil {
		return nil, fmt.Errorf("failed to start queue manager: %w", err)
	}
	
	logger.Info("Queue system initialized successfully")
	return qm, nil
}

// InitializeGlobalQueueSystem 初始化全局队列系统
func InitializeGlobalQueueSystem(cfg *config.Config, redisClient *redis.Client) error {
	InitGlobalQueueManager(cfg, redisClient)
	
	qm := GetGlobalQueueManager()
	if qm == nil {
		return fmt.Errorf("failed to get global queue manager")
	}
	
	// 根据配置初始化各个组件
	if cfg.MessageQueue.Enabled {
		if cfg.MessageQueue.Type == "redis" && redisClient != nil {
			qm.InitRedisMessageQueue(redisClient, RedisMessageQueueConfig{
				QueueName:         cfg.MessageQueue.QueueName,
				Workers:           cfg.MessageQueue.Workers,
				VisibilityTimeout: time.Duration(cfg.MessageQueue.VisibilityTimeout) * time.Second,
				PollInterval:      time.Duration(cfg.MessageQueue.PollInterval) * time.Millisecond,
			})
		} else {
			qm.InitMessageQueue(cfg.MessageQueue.QueueName, 1000, cfg.MessageQueue.Workers)
		}
	}
	
	if cfg.EventBus.Enabled {
		if cfg.EventBus.Type == "redis_pubsub" && redisClient != nil {
			qm.InitRedisPubSubQueue(redisClient)
		}
		qm.InitLocalEventBus(cfg.EventBus.BufferSize, cfg.EventBus.Workers)
	}
	
	if cfg.TaskScheduler.Enabled {
		qm.InitTaskScheduler(cfg.TaskScheduler.BufferSize, cfg.TaskScheduler.Workers)
	}
	
	if cfg.TaskDispatcher.Enabled {
		qm.InitTaskDispatcher()
	}
	
	// 启动所有组件
	if err := qm.Start(); err != nil {
		return fmt.Errorf("failed to start global queue manager: %w", err)
	}
	
	logger.Info("Global queue system initialized successfully")
	return nil
}

// RegisterCommonHandlers 注册通用的消息处理器
func RegisterCommonHandlers(qm *QueueManager, serviceName string) {
	// 注册Redis消息队列处理器
	if redisQueue := qm.GetRedisMessageQueue(); redisQueue != nil {
		logger.Info(fmt.Sprintf("[%s] Registering Redis message queue handlers", serviceName))
		
		// 这里可以注册通用的处理器
		// 具体的处理器应该在各个服务中注册
	}
	
	// 注册发布订阅处理器
	if pubsubQueue := qm.GetRedisPubSubQueue(); pubsubQueue != nil {
		logger.Info(fmt.Sprintf("[%s] Registering PubSub handlers", serviceName))
		
		// 订阅系统级事件
		pubsubQueue.Subscribe("system_events", func(ctx context.Context, msg *PubSubMessage) error {
			logger.Info(fmt.Sprintf("[%s] Received system event: %s", serviceName, msg.Type))
			return nil
		})
	}
	
	// 注册本地事件总线处理器
	if localBus := qm.GetLocalEventBus(); localBus != nil {
		logger.Info(fmt.Sprintf("[%s] Registering local event bus handlers", serviceName))
		
		// 注册通用的本地事件处理器
		localBus.On("service_started", func(ctx context.Context, event *LocalEvent) error {
			logger.Info(fmt.Sprintf("[%s] Service started event received", serviceName))
			return nil
		})
		
		localBus.On("service_stopping", func(ctx context.Context, event *LocalEvent) error {
			logger.Info(fmt.Sprintf("[%s] Service stopping event received", serviceName))
			return nil
		})
	}
}

// PublishSystemEvent 发布系统事件
func PublishSystemEvent(qm *QueueManager, eventType string, payload map[string]interface{}) error {
	pubsubQueue := qm.GetRedisPubSubQueue()
	if pubsubQueue == nil {
		return fmt.Errorf("redis pubsub queue not initialized")
	}
	
	msg := &PubSubMessage{
		Type:    eventType,
		Payload: payload,
		Source:  "system",
	}
	
	return pubsubQueue.Publish(context.Background(), "system_events", msg)
}

// EmitLocalEvent 触发本地事件
func EmitLocalEvent(qm *QueueManager, eventType string, payload map[string]interface{}, source string) error {
	localBus := qm.GetLocalEventBus()
	if localBus == nil {
		return fmt.Errorf("local event bus not initialized")
	}
	
	return localBus.Emit(eventType, payload, source)
}

// PublishTask 发布任务到队列
func PublishTask(qm *QueueManager, taskType string, priority MessagePriority, payload map[string]interface{}, source string) error {
	redisQueue := qm.GetRedisMessageQueue()
	if redisQueue == nil {
		// 回退到内存队列
		memQueue := qm.GetMessageQueue()
		if memQueue == nil {
			return fmt.Errorf("no message queue available")
		}
		
		msg := &Message{
			Type:     taskType,
			Priority: priority,
			Payload:  payload,
			Source:   source,
		}
		
		return memQueue.Publish(context.Background(), msg)
	}
	
	msg := &Message{
		Type:       taskType,
		Priority:   priority,
		Payload:    payload,
		Source:     source,
		MaxRetries: 3,
		Timeout:    30,
	}
	
	return redisQueue.Publish(context.Background(), msg)
}

// ScheduleDelayedTask 调度延迟任务
func ScheduleDelayedTask(qm *QueueManager, taskType string, priority MessagePriority, payload map[string]interface{}, delay time.Duration, handler TaskHandler) error {
	scheduler := qm.GetTaskScheduler()
	if scheduler == nil {
		return fmt.Errorf("task scheduler not initialized")
	}
	
	task := &Task{
		Type:        taskType,
		Priority:    priority,
		Payload:     payload,
		Handler:     handler,
		Timeout:     60 * time.Second,
		MaxRetries:  3,
		ScheduledAt: time.Now().Add(delay),
	}
	
	return scheduler.SubmitTask(task)
}

// GetQueueStats 获取队列统计信息
func GetQueueStats(qm *QueueManager) map[string]interface{} {
	return qm.GetStats()
}

// ShutdownQueueSystem 关闭队列系统
func ShutdownQueueSystem(qm *QueueManager) error {
	logger.Info("Shutting down queue system...")
	
	if err := qm.Stop(); err != nil {
		return fmt.Errorf("failed to stop queue manager: %w", err)
	}
	
	logger.Info("Queue system shut down successfully")
	return nil
}

// ShutdownGlobalQueueSystem 关闭全局队列系统
func ShutdownGlobalQueueSystem() error {
	logger.Info("Shutting down global queue system...")
	
	if err := CloseGlobalQueueManager(); err != nil {
		return fmt.Errorf("failed to close global queue manager: %w", err)
	}
	
	logger.Info("Global queue system shut down successfully")
	return nil
}

