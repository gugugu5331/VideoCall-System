package queue

import (
	"context"
	"fmt"
	"time"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// InitializeQueueSystem 初始化完整的队列系统（仅依赖 Kafka/内存队列）
func InitializeQueueSystem(cfg *config.Config) (*QueueManager, error) {
	logger.Info("Initializing queue system...")

	// 创建队列管理器
	qm := NewQueueManager(cfg)

	// 根据配置初始化各个组件
	if cfg.MessageQueue.Enabled && cfg.Kafka.Enabled {
		logger.Info("Initializing Kafka message queue...")
		qm.InitKafkaMessageQueue(cfg.Kafka, cfg.MessageQueue)
	}

	// 初始化发布订阅队列
	if cfg.EventBus.Enabled && cfg.Kafka.Enabled {
		logger.Info("Initializing Kafka event bus...")
		qm.InitKafkaEventBus(cfg.Kafka)
	}

	// 同时初始化本地事件总线（用于服务内部事件）
	logger.Info("Initializing local event bus...")
	qm.InitLocalEventBus(cfg.EventBus.BufferSize, cfg.EventBus.Workers)

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
func InitializeGlobalQueueSystem(cfg *config.Config) error {
	InitGlobalQueueManager(cfg)

	qm := GetGlobalQueueManager()
	if qm == nil {
		return fmt.Errorf("failed to get global queue manager")
	}

	// 根据配置初始化各个组件
	if cfg.MessageQueue.Enabled {
		if cfg.Kafka.Enabled {
			qm.InitKafkaMessageQueue(cfg.Kafka, cfg.MessageQueue)
		}
	}

	if cfg.EventBus.Enabled {
		if cfg.Kafka.Enabled {
			qm.InitKafkaEventBus(cfg.Kafka)
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
	// Kafka 事件总线
	if kafkaBus := qm.GetKafkaEventBus(); kafkaBus != nil {
		logger.Info(fmt.Sprintf("[%s] Registering Kafka event bus handlers", serviceName))

		kafkaBus.Subscribe("system_events", func(ctx context.Context, msg *PubSubMessage) error {
			logger.Info(fmt.Sprintf("[%s] Received Kafka system event: %s", serviceName, msg.Type))
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
	msg := &PubSubMessage{
		Type:    eventType,
		Payload: payload,
		Source:  "system",
	}

	if kafkaBus := qm.GetKafkaEventBus(); kafkaBus != nil {
		return kafkaBus.Publish(context.Background(), "system_events", msg)
	}

	return fmt.Errorf("no event bus initialized")
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
	if kafkaQueue := qm.GetKafkaMessageQueue(); kafkaQueue != nil {
		msg := &Message{
			Type:       taskType,
			Priority:   priority,
			Payload:    payload,
			Source:     source,
			MaxRetries: 3,
			Timeout:    30,
		}
		return kafkaQueue.Publish(context.Background(), msg)
	}

	// Kafka 未启用时回退到内存队列
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
