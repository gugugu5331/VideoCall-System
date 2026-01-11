package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// KafkaMessageQueueConfig 配置
type KafkaMessageQueueConfig struct {
	Brokers         []string
	Topic           string
	DeadLetterTopic string
	GroupID         string
	Workers         int
	Transport       *kafka.Transport
}

// KafkaMessageQueue 基于 Kafka 的消息队列
type KafkaMessageQueue struct {
	cfg      KafkaMessageQueueConfig
	writer   *kafka.Writer
	readers  []*kafka.Reader
	handlers map[string]MessageHandler

	handlersMutex sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup

	stats struct {
		sync.RWMutex
		totalPublished  uint64
		totalProcessed  uint64
		totalFailed     uint64
		totalRetried    uint64
		totalDeadLetter uint64
	}
}

// NewKafkaMessageQueue 创建 Kafka 消息队列
func NewKafkaMessageQueue(cfg KafkaMessageQueueConfig) *KafkaMessageQueue {
	if cfg.Workers <= 0 {
		cfg.Workers = 3
	}
	if cfg.DeadLetterTopic == "" && cfg.Topic != "" {
		cfg.DeadLetterTopic = cfg.Topic + ".dlq"
	}
	if cfg.Transport == nil {
		cfg.Transport = &kafka.Transport{}
	}

	ctx, cancel := context.WithCancel(context.Background())

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           50 * time.Millisecond,
		Transport:              cfg.Transport,
	}

	return &KafkaMessageQueue{
		cfg:      cfg,
		writer:   writer,
		handlers: make(map[string]MessageHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// RegisterHandler 注册消息处理器
func (q *KafkaMessageQueue) RegisterHandler(messageType string, handler MessageHandler) {
	q.handlersMutex.Lock()
	defer q.handlersMutex.Unlock()
	q.handlers[messageType] = handler
	logger.Info(fmt.Sprintf("Registered Kafka handler for message type: %s", messageType))
}

// Publish 发布消息
func (q *KafkaMessageQueue) Publish(ctx context.Context, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	q.prepareMessage(msg)

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kMsg := kafka.Message{
		Topic: q.cfg.Topic,
		Key:   []byte(msg.Type),
		Value: payload,
		Time:  time.Now(),
	}

	if err := q.writer.WriteMessages(ctx, kMsg); err != nil {
		return fmt.Errorf("failed to publish message to Kafka: %w", err)
	}

	atomic.AddUint64(&q.stats.totalPublished, 1)
	return nil
}

// PublishBatch 批量发布消息
func (q *KafkaMessageQueue) PublishBatch(ctx context.Context, messages []*Message) error {
	if len(messages) == 0 {
		return nil
	}

	kMessages := make([]kafka.Message, 0, len(messages))
	for _, msg := range messages {
		q.prepareMessage(msg)
		payload, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message %s: %w", msg.ID, err)
		}
		kMessages = append(kMessages, kafka.Message{
			Topic: q.cfg.Topic,
			Key:   []byte(msg.Type),
			Value: payload,
			Time:  time.Now(),
		})
	}

	if err := q.writer.WriteMessages(ctx, kMessages...); err != nil {
		return fmt.Errorf("failed to publish batch to Kafka: %w", err)
	}

	atomic.AddUint64(&q.stats.totalPublished, uint64(len(messages)))
	return nil
}

// Start 启动消费者
func (q *KafkaMessageQueue) Start() error {
	logger.Info(fmt.Sprintf("Starting Kafka message queue '%s' with group '%s' (%d workers)", q.cfg.Topic, q.cfg.GroupID, q.cfg.Workers))

	for i := 0; i < q.cfg.Workers; i++ {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     q.cfg.Brokers,
			GroupID:     q.cfg.GroupID,
			Topic:       q.cfg.Topic,
			Dialer:      &kafka.Dialer{Timeout: 10 * time.Second, DualStack: true, TLS: q.cfg.Transport.TLS, SASLMechanism: q.cfg.Transport.SASL},
			MinBytes:    1e3,
			MaxBytes:    10e6,
			StartOffset: kafka.FirstOffset,
		})
		q.readers = append(q.readers, reader)

		q.wg.Add(1)
		go q.consumeLoop(reader)
	}

	return nil
}

// Stop 停止消费者并关闭连接
func (q *KafkaMessageQueue) Stop() error {
	logger.Info("Stopping Kafka message queue...")
	q.cancel()

	for _, reader := range q.readers {
		_ = reader.Close()
	}
	q.wg.Wait()

	if q.writer != nil {
		_ = q.writer.Close()
	}

	logger.Info("Kafka message queue stopped")
	return nil
}

func (q *KafkaMessageQueue) consumeLoop(reader *kafka.Reader) {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		default:
		}

		kMsg, err := reader.FetchMessage(q.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			logger.Warn(fmt.Sprintf("Kafka fetch error: %v", err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var msg Message
		if err := json.Unmarshal(kMsg.Value, &msg); err != nil {
			logger.Error(fmt.Sprintf("Failed to decode Kafka message: %v", err))
			_ = reader.CommitMessages(context.Background(), kMsg)
			continue
		}

		q.handleMessage(&msg)
		if err := reader.CommitMessages(context.Background(), kMsg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to commit Kafka message: %v", err))
		}
	}
}

func (q *KafkaMessageQueue) handleMessage(msg *Message) {
	handler := q.getHandler(msg.Type)
	if handler == nil {
		logger.Debug(fmt.Sprintf("No Kafka handler for message type: %s", msg.Type))
		return
	}

	timeout := time.Duration(msg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(q.ctx, timeout)
	defer cancel()

	if err := handler(ctx, msg); err != nil {
		logger.Error(fmt.Sprintf("Kafka message handler error for %s: %v", msg.Type, err))
		atomic.AddUint64(&q.stats.totalFailed, 1)
		q.retryOrDeadLetter(ctx, msg)
		return
	}

	atomic.AddUint64(&q.stats.totalProcessed, 1)
}

func (q *KafkaMessageQueue) retryOrDeadLetter(ctx context.Context, msg *Message) {
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 3
	}

	if msg.RetryCount < msg.MaxRetries {
		msg.RetryCount++
		if err := q.Publish(ctx, msg); err != nil {
			logger.Error(fmt.Sprintf("Failed to requeue Kafka message %s: %v", msg.ID, err))
		} else {
			atomic.AddUint64(&q.stats.totalRetried, 1)
		}
		return
	}

	atomic.AddUint64(&q.stats.totalDeadLetter, 1)
	if q.cfg.DeadLetterTopic == "" {
		return
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshal message for DLQ: %v", err))
		return
	}

	dlqMsg := kafka.Message{
		Topic: q.cfg.DeadLetterTopic,
		Key:   []byte(msg.Type),
		Value: payload,
		Time:  time.Now(),
	}

	if err := q.writer.WriteMessages(ctx, dlqMsg); err != nil {
		logger.Error(fmt.Sprintf("Failed to publish to DLQ: %v", err))
	}
}

func (q *KafkaMessageQueue) getHandler(messageType string) MessageHandler {
	q.handlersMutex.RLock()
	defer q.handlersMutex.RUnlock()
	return q.handlers[messageType]
}

func (q *KafkaMessageQueue) prepareMessage(msg *Message) {
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}
	if msg.Timeout == 0 {
		msg.Timeout = 30
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 3
	}
	if msg.DeadLetterQueue == "" {
		msg.DeadLetterQueue = q.cfg.DeadLetterTopic
	}
}

// GetStats 返回统计信息
func (q *KafkaMessageQueue) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_published":   atomic.LoadUint64(&q.stats.totalPublished),
		"total_processed":   atomic.LoadUint64(&q.stats.totalProcessed),
		"total_failed":      atomic.LoadUint64(&q.stats.totalFailed),
		"total_retried":     atomic.LoadUint64(&q.stats.totalRetried),
		"total_dead_letter": atomic.LoadUint64(&q.stats.totalDeadLetter),
		"topic":             q.cfg.Topic,
		"group_id":          q.cfg.GroupID,
	}
}

// NewKafkaQueueFromConfig 基于全局配置初始化
func NewKafkaQueueFromConfig(cfg config.Config) *KafkaMessageQueue {
	transport := buildKafkaTransport(cfg.Kafka)
	return NewKafkaMessageQueue(KafkaMessageQueueConfig{
		Brokers:         cfg.Kafka.Brokers,
		Topic:           fmt.Sprintf("%s.tasks", cfg.Kafka.TopicPrefix),
		DeadLetterTopic: fmt.Sprintf("%s.tasks.dlq", cfg.Kafka.TopicPrefix),
		GroupID:         cfg.Kafka.GroupID,
		Workers:         cfg.MessageQueue.Workers,
		Transport:       transport,
	})
}
