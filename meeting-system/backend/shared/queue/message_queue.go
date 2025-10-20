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
