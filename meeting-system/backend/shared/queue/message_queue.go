package queue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"meeting-system/shared/logger"
)

// MessagePriority 表示消息优先级
type MessagePriority int

const (
	PriorityLow      MessagePriority = 0
	PriorityNormal   MessagePriority = 1
	PriorityHigh     MessagePriority = 2
	PriorityCritical MessagePriority = 3
)

// Message 消息结构
type Message struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    MessagePriority        `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   int64                  `json:"timestamp"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	Timeout     int64                  `json:"timeout"` // 秒
	Source      string                 `json:"source"`
	Destination string                 `json:"destination"`

	// 新增字段用于可靠性保证
	ProcessingStartTime int64  `json:"processing_start_time,omitempty"`
	VisibilityTimeout   int64  `json:"visibility_timeout,omitempty"` // 秒
	DeadLetterQueue     string `json:"dead_letter_queue,omitempty"`
}

// MessageHandler 定义消息处理函数
type MessageHandler func(ctx context.Context, msg *Message) error

// generateMessageID 生成唯一消息ID
func generateMessageID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RedisMessageQueue 基于Redis的消息队列
type RedisMessageQueue struct {
	client    *redis.Client
	queueName string

	// 优先级队列名称
	criticalQueue string
	highQueue     string
	normalQueue   string
	lowQueue      string

	// 处理中队列和死信队列
	processingQueue string
	deadLetterQueue string

	// 消息处理器
	handlers      map[string]MessageHandler
	handlersMutex sync.RWMutex

	// 工作协程控制
	workers int
	stopCh  chan struct{}
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc

	// 统计信息
	stats struct {
		sync.RWMutex
		totalPublished uint64
		totalProcessed uint64
		totalFailed    uint64
		totalRetried   uint64
		totalDeadLetter uint64
	}

	// 配置
	visibilityTimeout time.Duration // 消息可见性超时
	pollInterval      time.Duration // 轮询间隔
}

// RedisMessageQueueConfig Redis消息队列配置
type RedisMessageQueueConfig struct {
	QueueName         string
	Workers           int
	VisibilityTimeout time.Duration
	PollInterval      time.Duration
}

// NewRedisMessageQueue 创建Redis消息队列
func NewRedisMessageQueue(client *redis.Client, config RedisMessageQueueConfig) *RedisMessageQueue {
	if config.Workers <= 0 {
		config.Workers = 4
	}
	if config.VisibilityTimeout <= 0 {
		config.VisibilityTimeout = 30 * time.Second
	}
	if config.PollInterval <= 0 {
		config.PollInterval = 100 * time.Millisecond
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RedisMessageQueue{
		client:            client,
		queueName:         config.QueueName,
		criticalQueue:     fmt.Sprintf("%s:critical", config.QueueName),
		highQueue:         fmt.Sprintf("%s:high", config.QueueName),
		normalQueue:       fmt.Sprintf("%s:normal", config.QueueName),
		lowQueue:          fmt.Sprintf("%s:low", config.QueueName),
		processingQueue:   fmt.Sprintf("%s:processing", config.QueueName),
		deadLetterQueue:   fmt.Sprintf("%s:dlq", config.QueueName),
		handlers:          make(map[string]MessageHandler),
		workers:           config.Workers,
		stopCh:            make(chan struct{}),
		ctx:               ctx,
		cancel:            cancel,
		visibilityTimeout: config.VisibilityTimeout,
		pollInterval:      config.PollInterval,
	}
}

// RegisterHandler 注册消息处理器
func (q *RedisMessageQueue) RegisterHandler(messageType string, handler MessageHandler) {
	q.handlersMutex.Lock()
	defer q.handlersMutex.Unlock()
	q.handlers[messageType] = handler
	logger.Info(fmt.Sprintf("Registered handler for message type: %s", messageType))
}

// Publish 发布消息到队列
func (q *RedisMessageQueue) Publish(ctx context.Context, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// 设置默认值
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}
	if msg.Timeout == 0 {
		msg.Timeout = 30 // 默认30秒超时
	}
	if msg.VisibilityTimeout == 0 {
		msg.VisibilityTimeout = int64(q.visibilityTimeout.Seconds())
	}
	msg.DeadLetterQueue = q.deadLetterQueue

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 根据优先级选择队列
	var targetQueue string
	switch msg.Priority {
	case PriorityCritical:
		targetQueue = q.criticalQueue
	case PriorityHigh:
		targetQueue = q.highQueue
	case PriorityNormal:
		targetQueue = q.normalQueue
	case PriorityLow:
		targetQueue = q.lowQueue
	default:
		targetQueue = q.normalQueue
	}

	// 推送到Redis列表
	if err := q.client.RPush(ctx, targetQueue, data).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	atomic.AddUint64(&q.stats.totalPublished, 1)
	logger.Debug(fmt.Sprintf("Published message %s to queue %s", msg.ID, targetQueue))

	return nil
}

// PublishBatch 批量发布消息
func (q *RedisMessageQueue) PublishBatch(ctx context.Context, messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()

	for _, msg := range messages {
		if msg.ID == "" {
			msg.ID = generateMessageID()
		}
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().Unix()
		}
		if msg.Timeout == 0 {
			msg.Timeout = 30
		}
		if msg.VisibilityTimeout == 0 {
			msg.VisibilityTimeout = int64(q.visibilityTimeout.Seconds())
		}
		msg.DeadLetterQueue = q.deadLetterQueue

		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message %s: %w", msg.ID, err)
		}

		var targetQueue string
		switch msg.Priority {
		case PriorityCritical:
			targetQueue = q.criticalQueue
		case PriorityHigh:
			targetQueue = q.highQueue
		case PriorityNormal:
			targetQueue = q.normalQueue
		case PriorityLow:
			targetQueue = q.lowQueue
		default:
			targetQueue = q.normalQueue
		}

		pipe.RPush(ctx, targetQueue, data)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	atomic.AddUint64(&q.stats.totalPublished, uint64(len(messages)))
	logger.Debug(fmt.Sprintf("Published %d messages in batch", len(messages)))

	return nil
}

// Start 启动消息队列工作协程
func (q *RedisMessageQueue) Start() error {
	logger.Info(fmt.Sprintf("Starting Redis message queue '%s' with %d workers", q.queueName, q.workers))

	// 启动工作协程
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	// 启动超时检测协程
	q.wg.Add(1)
	go q.timeoutChecker()

	return nil
}

// Stop 停止消息队列
func (q *RedisMessageQueue) Stop() error {
	logger.Info(fmt.Sprintf("Stopping Redis message queue '%s'...", q.queueName))
	q.cancel()
	close(q.stopCh)
	q.wg.Wait()
	logger.Info(fmt.Sprintf("Redis message queue '%s' stopped", q.queueName))
	return nil
}

// worker 工作协程
func (q *RedisMessageQueue) worker(id int) {
	defer q.wg.Done()
	logger.Debug(fmt.Sprintf("Worker %d started for queue '%s'", id, q.queueName))

	for {
		select {
		case <-q.stopCh:
			logger.Debug(fmt.Sprintf("Worker %d stopped", id))
			return
		case <-q.ctx.Done():
			logger.Debug(fmt.Sprintf("Worker %d context cancelled", id))
			return
		default:
			// 按优先级顺序处理消息
			if q.processNextMessage() {
				continue
			}

			// 没有消息时短暂休眠
			time.Sleep(q.pollInterval)
		}
	}
}

// processNextMessage 处理下一条消息
func (q *RedisMessageQueue) processNextMessage() bool {
	ctx := q.ctx

	// 按优先级顺序尝试获取消息
	queues := []string{q.criticalQueue, q.highQueue, q.normalQueue, q.lowQueue}

	for _, queue := range queues {
		// 使用BLPOP从队列中获取消息（阻塞式）
		result, err := q.client.BLPop(ctx, q.pollInterval, queue).Result()
		if err != nil {
			if err != redis.Nil {
				logger.Error(fmt.Sprintf("Failed to pop message from %s: %v", queue, err))
			}
			continue
		}

		if len(result) < 2 {
			continue
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
			logger.Error(fmt.Sprintf("Failed to unmarshal message: %v", err))
			continue
		}

		// 处理消息
		q.handleMessage(&msg)
		return true
	}

	return false
}

// handleMessage 处理消息
func (q *RedisMessageQueue) handleMessage(msg *Message) {
	startTime := time.Now()
	msg.ProcessingStartTime = startTime.Unix()

	// 将消息移到处理中队列
	data, _ := json.Marshal(msg)
	q.client.HSet(q.ctx, q.processingQueue, msg.ID, data)

	defer func() {
		// 处理完成后从处理中队列移除
		q.client.HDel(q.ctx, q.processingQueue, msg.ID)
	}()

	// 获取处理器
	q.handlersMutex.RLock()
	handler, exists := q.handlers[msg.Type]
	q.handlersMutex.RUnlock()

	if !exists {
		logger.Error(fmt.Sprintf("No handler registered for message type: %s", msg.Type))
		q.moveToDeadLetterQueue(msg, fmt.Errorf("no handler registered"))
		return
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(q.ctx, time.Duration(msg.Timeout)*time.Second)
	defer cancel()

	// 执行处理器
	err := handler(ctx, msg)

	duration := time.Since(startTime)

	if err != nil {
		logger.Error(fmt.Sprintf("Message %s processing failed: %v (took %v)", msg.ID, err, duration))

		// 重试逻辑
		if msg.MaxRetries > 0 && msg.RetryCount < msg.MaxRetries {
			msg.RetryCount++
			atomic.AddUint64(&q.stats.totalRetried, 1)

			logger.Warn(fmt.Sprintf("Retrying message %s (%d/%d)", msg.ID, msg.RetryCount, msg.MaxRetries))

			// 延迟重试
			time.Sleep(time.Duration(msg.RetryCount) * time.Second)
			q.Publish(q.ctx, msg)
		} else {
			// 超过最大重试次数，移到死信队列
			q.moveToDeadLetterQueue(msg, err)
		}

		atomic.AddUint64(&q.stats.totalFailed, 1)
	} else {
		atomic.AddUint64(&q.stats.totalProcessed, 1)
		logger.Debug(fmt.Sprintf("Message %s processed successfully (took %v)", msg.ID, duration))
	}
}

// moveToDeadLetterQueue 将消息移到死信队列
func (q *RedisMessageQueue) moveToDeadLetterQueue(msg *Message, err error) {
	msg.Payload["error"] = err.Error()
	msg.Payload["failed_at"] = time.Now().Unix()

	data, _ := json.Marshal(msg)
	if err := q.client.RPush(q.ctx, q.deadLetterQueue, data).Err(); err != nil {
		logger.Error(fmt.Sprintf("Failed to move message %s to DLQ: %v", msg.ID, err))
	} else {
		atomic.AddUint64(&q.stats.totalDeadLetter, 1)
		logger.Warn(fmt.Sprintf("Message %s moved to dead letter queue", msg.ID))
	}
}

// timeoutChecker 检查处理超时的消息
func (q *RedisMessageQueue) timeoutChecker() {
	defer q.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopCh:
			return
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			q.checkTimeouts()
		}
	}
}

// checkTimeouts 检查超时消息
func (q *RedisMessageQueue) checkTimeouts() {
	ctx := q.ctx

	// 获取所有处理中的消息
	processingMsgs, err := q.client.HGetAll(ctx, q.processingQueue).Result()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get processing messages: %v", err))
		return
	}

	now := time.Now().Unix()

	for msgID, msgData := range processingMsgs {
		var msg Message
		if err := json.Unmarshal([]byte(msgData), &msg); err != nil {
			logger.Error(fmt.Sprintf("Failed to unmarshal processing message: %v", err))
			continue
		}

		// 检查是否超时
		if now-msg.ProcessingStartTime > msg.VisibilityTimeout {
			logger.Warn(fmt.Sprintf("Message %s processing timeout, re-queuing", msgID))

			// 从处理中队列移除
			q.client.HDel(ctx, q.processingQueue, msgID)

			// 重新入队
			if msg.MaxRetries > 0 && msg.RetryCount < msg.MaxRetries {
				msg.RetryCount++
				q.Publish(ctx, &msg)
			} else {
				q.moveToDeadLetterQueue(&msg, fmt.Errorf("processing timeout"))
			}
		}
	}
}

// GetQueueLength 获取指定优先级队列的长度
func (q *RedisMessageQueue) GetQueueLength(priority MessagePriority) int {
	var queue string
	switch priority {
	case PriorityCritical:
		queue = q.criticalQueue
	case PriorityHigh:
		queue = q.highQueue
	case PriorityNormal:
		queue = q.normalQueue
	case PriorityLow:
		queue = q.lowQueue
	default:
		return 0
	}

	length, err := q.client.LLen(q.ctx, queue).Result()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get queue length: %v", err))
		return 0
	}

	return int(length)
}

// GetTotalQueueLength 获取所有队列的总长度
func (q *RedisMessageQueue) GetTotalQueueLength() int {
	total := 0
	queues := []string{q.criticalQueue, q.highQueue, q.normalQueue, q.lowQueue}

	for _, queue := range queues {
		length, err := q.client.LLen(q.ctx, queue).Result()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get queue length for %s: %v", queue, err))
			continue
		}
		total += int(length)
	}

	return total
}

// GetProcessingCount 获取正在处理的消息数量
func (q *RedisMessageQueue) GetProcessingCount() int {
	count, err := q.client.HLen(q.ctx, q.processingQueue).Result()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get processing count: %v", err))
		return 0
	}
	return int(count)
}

// GetDeadLetterCount 获取死信队列的消息数量
func (q *RedisMessageQueue) GetDeadLetterCount() int {
	count, err := q.client.LLen(q.ctx, q.deadLetterQueue).Result()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get dead letter count: %v", err))
		return 0
	}
	return int(count)
}

// GetStats 获取统计信息
func (q *RedisMessageQueue) GetStats() map[string]interface{} {
	q.stats.RLock()
	defer q.stats.RUnlock()

	return map[string]interface{}{
		"queue_name":         q.queueName,
		"total_published":    atomic.LoadUint64(&q.stats.totalPublished),
		"total_processed":    atomic.LoadUint64(&q.stats.totalProcessed),
		"total_failed":       atomic.LoadUint64(&q.stats.totalFailed),
		"total_retried":      atomic.LoadUint64(&q.stats.totalRetried),
		"total_dead_letter":  atomic.LoadUint64(&q.stats.totalDeadLetter),
		"queue_length": map[string]int{
			"critical": q.GetQueueLength(PriorityCritical),
			"high":     q.GetQueueLength(PriorityHigh),
			"normal":   q.GetQueueLength(PriorityNormal),
			"low":      q.GetQueueLength(PriorityLow),
			"total":    q.GetTotalQueueLength(),
		},
		"processing_count":   q.GetProcessingCount(),
		"dead_letter_count":  q.GetDeadLetterCount(),
		"workers":            q.workers,
	}
}

// ClearQueue 清空所有队列
func (q *RedisMessageQueue) ClearQueue() error {
	ctx := q.ctx
	queues := []string{
		q.criticalQueue,
		q.highQueue,
		q.normalQueue,
		q.lowQueue,
		q.processingQueue,
	}

	for _, queue := range queues {
		if err := q.client.Del(ctx, queue).Err(); err != nil {
			return fmt.Errorf("failed to clear queue %s: %w", queue, err)
		}
	}

	logger.Info(fmt.Sprintf("Cleared all queues for '%s'", q.queueName))
	return nil
}

// ClearDeadLetterQueue 清空死信队列
func (q *RedisMessageQueue) ClearDeadLetterQueue() error {
	if err := q.client.Del(q.ctx, q.deadLetterQueue).Err(); err != nil {
		return fmt.Errorf("failed to clear dead letter queue: %w", err)
	}

	logger.Info(fmt.Sprintf("Cleared dead letter queue for '%s'", q.queueName))
	return nil
}

// RequeueDeadLetterMessages 重新入队死信消息
func (q *RedisMessageQueue) RequeueDeadLetterMessages(maxCount int) (int, error) {
	ctx := q.ctx
	count := 0

	for i := 0; i < maxCount; i++ {
		result, err := q.client.LPop(ctx, q.deadLetterQueue).Result()
		if err != nil {
			if err == redis.Nil {
				break
			}
			return count, fmt.Errorf("failed to pop from DLQ: %w", err)
		}

		var msg Message
		if err := json.Unmarshal([]byte(result), &msg); err != nil {
			logger.Error(fmt.Sprintf("Failed to unmarshal DLQ message: %v", err))
			continue
		}

		// 重置重试计数
		msg.RetryCount = 0
		delete(msg.Payload, "error")
		delete(msg.Payload, "failed_at")

		if err := q.Publish(ctx, &msg); err != nil {
			logger.Error(fmt.Sprintf("Failed to requeue DLQ message %s: %v", msg.ID, err))
			continue
		}

		count++
	}

	logger.Info(fmt.Sprintf("Requeued %d messages from dead letter queue", count))
	return count, nil
}

// MemoryMessageQueue 内存消息队列（用于向后兼容）
type MemoryMessageQueue struct {
	queueName string
	mu        sync.RWMutex
	handlers  map[string]MessageHandler
	stats     struct {
		totalPublished uint64
		totalProcessed uint64
		totalFailed    uint64
	}
}

// NewMemoryMessageQueue 创建内存消息队列
func NewMemoryMessageQueue(queueName string, _ int, _ int) *MemoryMessageQueue {
	return &MemoryMessageQueue{
		queueName: queueName,
		handlers:  make(map[string]MessageHandler),
	}
}

// RegisterHandler 注册消息处理器
func (q *MemoryMessageQueue) RegisterHandler(messageType string, handler MessageHandler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[messageType] = handler
}

// Publish 同步发布消息
func (q *MemoryMessageQueue) Publish(ctx context.Context, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}

	q.mu.RLock()
	handler := q.handlers[msg.Type]
	q.mu.RUnlock()

	atomic.AddUint64(&q.stats.totalPublished, 1)

	if handler == nil {
		return fmt.Errorf("no handler registered for message type: %s", msg.Type)
	}

	if err := handler(ctx, msg); err != nil {
		atomic.AddUint64(&q.stats.totalFailed, 1)
		return err
	}

	atomic.AddUint64(&q.stats.totalProcessed, 1)
	return nil
}

// PublishBatch 批量发布消息
func (q *MemoryMessageQueue) PublishBatch(ctx context.Context, messages []*Message) error {
	for _, msg := range messages {
		if err := q.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Start 启动队列
func (q *MemoryMessageQueue) Start() error {
	return nil
}

// Stop 停止队列
func (q *MemoryMessageQueue) Stop() error {
	return nil
}

// GetQueueLength 获取队列长度
func (q *MemoryMessageQueue) GetQueueLength(priority MessagePriority) int {
	return 0
}

// GetTotalQueueLength 获取总队列长度
func (q *MemoryMessageQueue) GetTotalQueueLength() int {
	return 0
}

// GetStats 获取统计信息
func (q *MemoryMessageQueue) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"queue_name":      q.queueName,
		"total_published": atomic.LoadUint64(&q.stats.totalPublished),
		"total_processed": atomic.LoadUint64(&q.stats.totalProcessed),
		"total_failed":    atomic.LoadUint64(&q.stats.totalFailed),
		"total_pending":   0,
	}
}

// ClearQueue 清空队列
func (q *MemoryMessageQueue) ClearQueue() {}
