package queue

import (
	"fmt"
	"sync"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// QueueManager 统一管理所有队列组件
type QueueManager struct {
	config *config.Config

	// Kafka 队列
	kafkaMessageQueue *KafkaMessageQueue
	kafkaEventBus     *KafkaPubSub

	// 内存队列（向后兼容）
	memoryMessageQueue *MemoryMessageQueue

	// 本地事件总线
	localEventBus *LocalEventBus

	// 任务调度器
	taskScheduler *TaskScheduler

	// 任务分发器
	taskDispatcher *TaskDispatcher

	started bool
	mu      sync.Mutex
}

// NewQueueManager 创建队列管理器
func NewQueueManager(cfg *config.Config) *QueueManager {
	qm := &QueueManager{
		config: cfg,
	}

	return qm
}

// InitKafkaMessageQueue 初始化 Kafka 消息队列
func (qm *QueueManager) InitKafkaMessageQueue(kafkaCfg config.KafkaConfig, mqCfg config.MessageQueueConfig) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	transport := buildKafkaTransport(kafkaCfg)
	qm.kafkaMessageQueue = NewKafkaMessageQueue(KafkaMessageQueueConfig{
		Brokers:         kafkaCfg.Brokers,
		Topic:           fmt.Sprintf("%s.tasks", kafkaCfg.TopicPrefix),
		DeadLetterTopic: fmt.Sprintf("%s.tasks.dlq", kafkaCfg.TopicPrefix),
		GroupID:         kafkaCfg.GroupID,
		Workers:         mqCfg.Workers,
		Transport:       transport,
	})
}

// InitKafkaEventBus 初始化 Kafka 事件总线
func (qm *QueueManager) InitKafkaEventBus(kafkaCfg config.KafkaConfig) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.kafkaEventBus = NewKafkaPubSub(KafkaPubSubConfig{
		Brokers:     kafkaCfg.Brokers,
		TopicPrefix: kafkaCfg.TopicPrefix,
		GroupID:     kafkaCfg.GroupID,
		Transport:   buildKafkaTransport(kafkaCfg),
	})
}

// InitMessageQueue 初始化内存消息队列（向后兼容）
func (qm *QueueManager) InitMessageQueue(queueName string, bufferSize, workers int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.memoryMessageQueue = NewMemoryMessageQueue(queueName, bufferSize, workers)
}

// InitLocalEventBus 初始化本地事件总线
func (qm *QueueManager) InitLocalEventBus(bufferSize, workers int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.localEventBus = NewLocalEventBus(bufferSize, workers)
}

// InitTaskScheduler 初始化任务调度器
func (qm *QueueManager) InitTaskScheduler(bufferSize, workers int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.taskScheduler = NewTaskScheduler(bufferSize, workers)
}

// InitTaskDispatcher 初始化任务分发器
func (qm *QueueManager) InitTaskDispatcher() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.taskDispatcher = NewTaskDispatcher()
}

// Start 启动所有组件
func (qm *QueueManager) Start() error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if qm.started {
		return nil
	}

	// 启动 Kafka 组件
	if qm.kafkaMessageQueue != nil {
		if err := qm.kafkaMessageQueue.Start(); err != nil {
			return err
		}
	}
	if qm.kafkaEventBus != nil {
		if err := qm.kafkaEventBus.Start(); err != nil {
			return err
		}
	}

	// 启动内存队列
	if qm.memoryMessageQueue != nil {
		_ = qm.memoryMessageQueue.Start()
	}

	// 启动本地事件总线
	if qm.localEventBus != nil {
		qm.localEventBus.Start()
	}

	// 启动任务调度器
	if qm.taskScheduler != nil {
		qm.taskScheduler.Start()
	}

	// 启动任务分发器
	if qm.taskDispatcher != nil {
		_ = qm.taskDispatcher.Start()
	}

	qm.started = true
	logger.Info("Queue manager started")
	return nil
}

// Stop 停止所有组件
func (qm *QueueManager) Stop() error {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if !qm.started {
		return nil
	}

	// 停止任务分发器
	if qm.taskDispatcher != nil {
		_ = qm.taskDispatcher.Stop()
	}

	// 停止任务调度器
	if qm.taskScheduler != nil {
		qm.taskScheduler.Stop()
	}

	// 停止本地事件总线
	if qm.localEventBus != nil {
		qm.localEventBus.Stop()
	}

	// 停止 Kafka 组件
	if qm.kafkaEventBus != nil {
		if err := qm.kafkaEventBus.Stop(); err != nil {
			logger.Error("Failed to stop Kafka event bus: " + err.Error())
		}
	}
	if qm.kafkaMessageQueue != nil {
		if err := qm.kafkaMessageQueue.Stop(); err != nil {
			logger.Error("Failed to stop Kafka message queue: " + err.Error())
		}
	}

	// 停止内存队列
	if qm.memoryMessageQueue != nil {
		_ = qm.memoryMessageQueue.Stop()
	}

	qm.started = false
	logger.Info("Queue manager stopped")
	return nil
}

// GetKafkaMessageQueue 获取 Kafka 消息队列
func (qm *QueueManager) GetKafkaMessageQueue() *KafkaMessageQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.kafkaMessageQueue
}

// GetKafkaEventBus 获取 Kafka 事件总线
func (qm *QueueManager) GetKafkaEventBus() *KafkaPubSub {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.kafkaEventBus
}

// GetMessageQueue 获取内存消息队列（向后兼容）
func (qm *QueueManager) GetMessageQueue() *MemoryMessageQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.memoryMessageQueue
}

// GetLocalEventBus 获取本地事件总线
func (qm *QueueManager) GetLocalEventBus() *LocalEventBus {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.localEventBus
}

// GetTaskScheduler 获取任务调度器
func (qm *QueueManager) GetTaskScheduler() *TaskScheduler {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.taskScheduler
}

// GetTaskDispatcher 获取任务分发器
func (qm *QueueManager) GetTaskDispatcher() *TaskDispatcher {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.taskDispatcher
}

// GetStats 获取所有组件的统计信息
func (qm *QueueManager) GetStats() map[string]interface{} {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	stats := make(map[string]interface{})

	if qm.kafkaMessageQueue != nil {
		stats["kafka_message_queue"] = qm.kafkaMessageQueue.GetStats()
	}
	if qm.kafkaEventBus != nil {
		stats["kafka_event_bus"] = qm.kafkaEventBus.GetStats()
	}
	if qm.memoryMessageQueue != nil {
		stats["memory_message_queue"] = qm.memoryMessageQueue.GetStats()
	}
	if qm.localEventBus != nil {
		stats["local_event_bus"] = qm.localEventBus.GetStats()
	}
	if qm.taskScheduler != nil {
		stats["task_scheduler"] = qm.taskScheduler.GetStats()
	}
	if qm.taskDispatcher != nil {
		stats["task_dispatcher"] = qm.taskDispatcher.GetStats()
	}

	return stats
}

var (
	globalQueueManager *QueueManager
	queueManagerOnce   sync.Once
)

// InitGlobalQueueManager 初始化全局队列管理器
func InitGlobalQueueManager(cfg *config.Config) {
	queueManagerOnce.Do(func() {
		globalQueueManager = NewQueueManager(cfg)
	})
}

// GetGlobalQueueManager 获取全局队列管理器
func GetGlobalQueueManager() *QueueManager {
	return globalQueueManager
}

// CloseGlobalQueueManager 关闭全局队列管理器
func CloseGlobalQueueManager() error {
	if globalQueueManager != nil {
		return globalQueueManager.Stop()
	}
	return nil
}
