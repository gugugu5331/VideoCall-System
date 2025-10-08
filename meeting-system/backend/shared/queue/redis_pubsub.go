package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"meeting-system/shared/logger"
)

// PubSubMessage 发布订阅消息
type PubSubMessage struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp int64                  `json:"timestamp"`
	Source    string                 `json:"source"`
	MessageID string                 `json:"message_id"`
}

// PubSubHandler 发布订阅处理函数
type PubSubHandler func(ctx context.Context, msg *PubSubMessage) error

// RedisPubSubQueue Redis发布订阅队列
type RedisPubSubQueue struct {
	client *redis.Client
	
	// 订阅管理
	subscriptions map[string][]PubSubHandler
	subMutex      sync.RWMutex
	
	// 订阅连接
	pubsubs map[string]*redis.PubSub
	psMutex sync.RWMutex
	
	// 控制
	stopCh chan struct{}
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	
	// 统计
	stats struct {
		sync.RWMutex
		totalPublished  uint64
		totalReceived   uint64
		totalProcessed  uint64
		totalFailed     uint64
	}
}

// NewRedisPubSubQueue 创建Redis发布订阅队列
func NewRedisPubSubQueue(client *redis.Client) *RedisPubSubQueue {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RedisPubSubQueue{
		client:        client,
		subscriptions: make(map[string][]PubSubHandler),
		pubsubs:       make(map[string]*redis.PubSub),
		stopCh:        make(chan struct{}),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Subscribe 订阅频道
func (q *RedisPubSubQueue) Subscribe(channel string, handler PubSubHandler) error {
	q.subMutex.Lock()
	defer q.subMutex.Unlock()
	
	// 添加处理器
	if _, exists := q.subscriptions[channel]; !exists {
		q.subscriptions[channel] = make([]PubSubHandler, 0)
	}
	q.subscriptions[channel] = append(q.subscriptions[channel], handler)
	
	// 如果是第一次订阅该频道，创建订阅连接
	q.psMutex.Lock()
	if _, exists := q.pubsubs[channel]; !exists {
		pubsub := q.client.Subscribe(q.ctx, channel)
		q.pubsubs[channel] = pubsub
		
		// 启动消息接收协程
		q.wg.Add(1)
		go q.receiveMessages(channel, pubsub)
	}
	q.psMutex.Unlock()
	
	logger.Info(fmt.Sprintf("Subscribed to channel: %s", channel))
	return nil
}

// Unsubscribe 取消订阅频道
func (q *RedisPubSubQueue) Unsubscribe(channel string) error {
	q.subMutex.Lock()
	delete(q.subscriptions, channel)
	q.subMutex.Unlock()
	
	q.psMutex.Lock()
	defer q.psMutex.Unlock()
	
	if pubsub, exists := q.pubsubs[channel]; exists {
		if err := pubsub.Unsubscribe(q.ctx, channel); err != nil {
			return fmt.Errorf("failed to unsubscribe from channel %s: %w", channel, err)
		}
		if err := pubsub.Close(); err != nil {
			return fmt.Errorf("failed to close pubsub for channel %s: %w", channel, err)
		}
		delete(q.pubsubs, channel)
		logger.Info(fmt.Sprintf("Unsubscribed from channel: %s", channel))
	}
	
	return nil
}

// Publish 发布消息到频道
func (q *RedisPubSubQueue) Publish(ctx context.Context, channel string, msg *PubSubMessage) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	
	if msg.MessageID == "" {
		msg.MessageID = generateMessageID()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	if err := q.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	atomic.AddUint64(&q.stats.totalPublished, 1)
	logger.Debug(fmt.Sprintf("Published message to channel %s", channel))
	
	return nil
}

// PublishBatch 批量发布消息
func (q *RedisPubSubQueue) PublishBatch(ctx context.Context, channel string, messages []*PubSubMessage) error {
	pipe := q.client.Pipeline()
	
	for _, msg := range messages {
		if msg.MessageID == "" {
			msg.MessageID = generateMessageID()
		}
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().Unix()
		}
		
		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		
		pipe.Publish(ctx, channel, data)
	}
	
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to publish batch: %w", err)
	}
	
	atomic.AddUint64(&q.stats.totalPublished, uint64(len(messages)))
	logger.Debug(fmt.Sprintf("Published %d messages to channel %s", len(messages), channel))
	
	return nil
}

// receiveMessages 接收消息
func (q *RedisPubSubQueue) receiveMessages(channel string, pubsub *redis.PubSub) {
	defer q.wg.Done()
	
	ch := pubsub.Channel()
	
	for {
		select {
		case <-q.stopCh:
			return
		case <-q.ctx.Done():
			return
		case redisMsg, ok := <-ch:
			if !ok {
				logger.Warn(fmt.Sprintf("Channel %s closed", channel))
				return
			}
			
			atomic.AddUint64(&q.stats.totalReceived, 1)
			
			// 解析消息
			var msg PubSubMessage
			if err := json.Unmarshal([]byte(redisMsg.Payload), &msg); err != nil {
				logger.Error(fmt.Sprintf("Failed to unmarshal message from channel %s: %v", channel, err))
				continue
			}
			
			// 处理消息
			q.handleMessage(channel, &msg)
		}
	}
}

// handleMessage 处理消息
func (q *RedisPubSubQueue) handleMessage(channel string, msg *PubSubMessage) {
	q.subMutex.RLock()
	handlers, exists := q.subscriptions[channel]
	q.subMutex.RUnlock()
	
	if !exists || len(handlers) == 0 {
		logger.Debug(fmt.Sprintf("No handlers for channel: %s", channel))
		return
	}
	
	// 并发执行所有处理器
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h PubSubHandler) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Error(fmt.Sprintf("PubSub handler panic: %v", r))
					atomic.AddUint64(&q.stats.totalFailed, 1)
				}
			}()
			
			ctx, cancel := context.WithTimeout(q.ctx, 10*time.Second)
			defer cancel()
			
			if err := h(ctx, msg); err != nil {
				logger.Error(fmt.Sprintf("PubSub handler error for channel %s: %v", channel, err))
				atomic.AddUint64(&q.stats.totalFailed, 1)
			} else {
				atomic.AddUint64(&q.stats.totalProcessed, 1)
			}
		}(handler)
	}
	
	wg.Wait()
}

// Start 启动发布订阅队列
func (q *RedisPubSubQueue) Start() error {
	logger.Info("Starting Redis PubSub queue")
	return nil
}

// Stop 停止发布订阅队列
func (q *RedisPubSubQueue) Stop() error {
	logger.Info("Stopping Redis PubSub queue...")
	q.cancel()
	close(q.stopCh)
	
	// 关闭所有订阅
	q.psMutex.Lock()
	for channel, pubsub := range q.pubsubs {
		pubsub.Unsubscribe(q.ctx, channel)
		pubsub.Close()
	}
	q.pubsubs = make(map[string]*redis.PubSub)
	q.psMutex.Unlock()
	
	q.wg.Wait()
	logger.Info("Redis PubSub queue stopped")
	return nil
}

// GetStats 获取统计信息
func (q *RedisPubSubQueue) GetStats() map[string]interface{} {
	q.subMutex.RLock()
	channelCount := len(q.subscriptions)
	q.subMutex.RUnlock()
	
	return map[string]interface{}{
		"total_published": atomic.LoadUint64(&q.stats.totalPublished),
		"total_received":  atomic.LoadUint64(&q.stats.totalReceived),
		"total_processed": atomic.LoadUint64(&q.stats.totalProcessed),
		"total_failed":    atomic.LoadUint64(&q.stats.totalFailed),
		"channel_count":   channelCount,
	}
}

// GetSubscribedChannels 获取已订阅的频道列表
func (q *RedisPubSubQueue) GetSubscribedChannels() []string {
	q.subMutex.RLock()
	defer q.subMutex.RUnlock()
	
	channels := make([]string, 0, len(q.subscriptions))
	for channel := range q.subscriptions {
		channels = append(channels, channel)
	}
	
	return channels
}

