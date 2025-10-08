package queue

import (
	"sync"

	"github.com/redis/go-redis/v9"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// QueueManager 统一管理所有队列组件
type QueueManager struct {
	config *config.Config

	// Redis队列
	redisMessageQueue *RedisMessageQueue
	redisPubSubQueue  *RedisPubSubQueue

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
func NewQueueManager(cfg *config.Config, redisClient *redis.Client) *QueueManager {
	qm := &QueueManager{
		config: cfg,
	}

	// 如果提供了Redis客户端，初始化Redis队列
	if redisClient != nil {
		qm.initRedisQueues(redisClient)
	}

	return qm
}

// initRedisQueues 初始化Redis队列
func (qm *QueueManager) initRedisQueues(redisClient *redis.Client) {
	// 初始化Redis消息队列
	qm.redisMessageQueue = NewRedisMessageQueue(redisClient, RedisMessageQueueConfig{
		QueueName:         "default",
		Workers:           4,
		VisibilityTimeout: 30,
		PollInterval:      100,
	})

	// 初始化Redis发布订阅队列
	qm.redisPubSubQueue = NewRedisPubSubQueue(redisClient)

	logger.Info("Redis queues initialized")
}

// InitRedisMessageQueue 初始化Redis消息队列（自定义配置）
func (qm *QueueManager) InitRedisMessageQueue(redisClient *redis.Client, config RedisMessageQueueConfig) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.redisMessageQueue = NewRedisMessageQueue(redisClient, config)
}

// InitRedisPubSubQueue 初始化Redis发布订阅队列
func (qm *QueueManager) InitRedisPubSubQueue(redisClient *redis.Client) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.redisPubSubQueue = NewRedisPubSubQueue(redisClient)
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

	// 启动Redis队列
	if qm.redisMessageQueue != nil {
		if err := qm.redisMessageQueue.Start(); err != nil {
			return err
		}
	}
	if qm.redisPubSubQueue != nil {
		if err := qm.redisPubSubQueue.Start(); err != nil {
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

	// 停止内存队列
	if qm.memoryMessageQueue != nil {
		_ = qm.memoryMessageQueue.Stop()
	}

	// 停止Redis队列
	if qm.redisPubSubQueue != nil {
		if err := qm.redisPubSubQueue.Stop(); err != nil {
			logger.Error("Failed to stop Redis PubSub queue: " + err.Error())
		}
	}
	if qm.redisMessageQueue != nil {
		if err := qm.redisMessageQueue.Stop(); err != nil {
			logger.Error("Failed to stop Redis message queue: " + err.Error())
		}
	}

	qm.started = false
	logger.Info("Queue manager stopped")
	return nil
}

// GetRedisMessageQueue 获取Redis消息队列
func (qm *QueueManager) GetRedisMessageQueue() *RedisMessageQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.redisMessageQueue
}

// GetRedisPubSubQueue 获取Redis发布订阅队列
func (qm *QueueManager) GetRedisPubSubQueue() *RedisPubSubQueue {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	return qm.redisPubSubQueue
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

	if qm.redisMessageQueue != nil {
		stats["redis_message_queue"] = qm.redisMessageQueue.GetStats()
	}
	if qm.redisPubSubQueue != nil {
		stats["redis_pubsub_queue"] = qm.redisPubSubQueue.GetStats()
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
func InitGlobalQueueManager(cfg *config.Config, redisClient *redis.Client) {
	queueManagerOnce.Do(func() {
		globalQueueManager = NewQueueManager(cfg, redisClient)
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
