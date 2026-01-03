package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"meeting-system/shared/logger"
)

// MemoryMessageQueue 基于内存的消息队列（向后兼容/开发测试用）
type MemoryMessageQueue struct {
	queueName string
	buffer    chan *Message

	handlers      map[string]MessageHandler
	handlersMutex sync.RWMutex

	workers int
	stopCh  chan struct{}
	wg      sync.WaitGroup

	stats struct {
		totalPublished  uint64
		totalProcessed  uint64
		totalFailed     uint64
		totalRetried    uint64
		totalDeadLetter uint64
	}
}

// NewMemoryMessageQueue 创建内存消息队列
func NewMemoryMessageQueue(queueName string, bufferSize, workers int) *MemoryMessageQueue {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	if workers <= 0 {
		workers = 1
	}

	return &MemoryMessageQueue{
		queueName: queueName,
		buffer:    make(chan *Message, bufferSize),
		handlers:  make(map[string]MessageHandler),
		workers:   workers,
		stopCh:    make(chan struct{}),
	}
}

// RegisterHandler 注册消息处理器
func (q *MemoryMessageQueue) RegisterHandler(messageType string, handler MessageHandler) {
	q.handlersMutex.Lock()
	defer q.handlersMutex.Unlock()
	q.handlers[messageType] = handler
}

// Publish 发布消息
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
	if msg.Timeout == 0 {
		msg.Timeout = 30
	}

	select {
	case <-q.stopCh:
		return fmt.Errorf("queue stopped")
	case q.buffer <- msg:
		atomic.AddUint64(&q.stats.totalPublished, 1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	return nil
}

// Stop 停止队列
func (q *MemoryMessageQueue) Stop() error {
	select {
	case <-q.stopCh:
		// already closed
	default:
		close(q.stopCh)
	}
	q.wg.Wait()
	return nil
}

func (q *MemoryMessageQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.stopCh:
			return
		case msg := <-q.buffer:
			if msg == nil {
				continue
			}
			q.handleMessage(msg)
		}
	}
}

func (q *MemoryMessageQueue) handleMessage(msg *Message) {
	q.handlersMutex.RLock()
	handler, exists := q.handlers[msg.Type]
	q.handlersMutex.RUnlock()
	if !exists {
		atomic.AddUint64(&q.stats.totalFailed, 1)
		logger.Warn(fmt.Sprintf("No handler registered for message type: %s", msg.Type))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(msg.Timeout)*time.Second)
	defer cancel()

	if err := handler(ctx, msg); err != nil {
		atomic.AddUint64(&q.stats.totalFailed, 1)
		logger.Warn(fmt.Sprintf("Memory queue message %s failed: %v", msg.ID, err))
		return
	}

	atomic.AddUint64(&q.stats.totalProcessed, 1)
}

// GetStats 获取统计信息
func (q *MemoryMessageQueue) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"queue_name":        q.queueName,
		"workers":           q.workers,
		"buffer_size":       cap(q.buffer),
		"buffer_len":        len(q.buffer),
		"total_published":   atomic.LoadUint64(&q.stats.totalPublished),
		"total_processed":   atomic.LoadUint64(&q.stats.totalProcessed),
		"total_failed":      atomic.LoadUint64(&q.stats.totalFailed),
		"total_retried":     atomic.LoadUint64(&q.stats.totalRetried),
		"total_dead_letter": atomic.LoadUint64(&q.stats.totalDeadLetter),
	}

	return stats
}
