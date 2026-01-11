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

// KafkaPubSubConfig Kafka 发布订阅配置
type KafkaPubSubConfig struct {
	Brokers     []string
	TopicPrefix string
	GroupID     string
	Transport   *kafka.Transport
}

// KafkaPubSub 基于 Kafka 的事件总线
type KafkaPubSub struct {
	cfg KafkaPubSubConfig

	writer        *kafka.Writer
	subscriptions map[string][]PubSubHandler
	readers       map[string]*kafka.Reader

	subMutex sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	stats struct {
		sync.RWMutex
		totalPublished uint64
		totalReceived  uint64
		totalProcessed uint64
		totalFailed    uint64
	}
}

// NewKafkaPubSub 创建 Kafka 发布订阅实例
func NewKafkaPubSub(cfg KafkaPubSubConfig) *KafkaPubSub {
	if cfg.Transport == nil {
		cfg.Transport = &kafka.Transport{}
	}

	ctx, cancel := context.WithCancel(context.Background())

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           20 * time.Millisecond,
		Transport:              cfg.Transport,
	}

	return &KafkaPubSub{
		cfg:           cfg,
		writer:        writer,
		subscriptions: make(map[string][]PubSubHandler),
		readers:       make(map[string]*kafka.Reader),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Subscribe 订阅频道
func (b *KafkaPubSub) Subscribe(channel string, handler PubSubHandler) error {
	b.subMutex.Lock()
	defer b.subMutex.Unlock()

	// 记录处理器
	b.subscriptions[channel] = append(b.subscriptions[channel], handler)

	// 为新频道启动 reader
	if _, exists := b.readers[channel]; !exists {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     b.cfg.Brokers,
			GroupID:     b.cfg.GroupID,
			Topic:       b.topicForChannel(channel),
			Dialer:      &kafka.Dialer{Timeout: 10 * time.Second, DualStack: true, TLS: b.cfg.Transport.TLS, SASLMechanism: b.cfg.Transport.SASL},
			MinBytes:    1e3,
			MaxBytes:    10e6,
			StartOffset: kafka.FirstOffset,
		})
		b.readers[channel] = reader

		b.wg.Add(1)
		go b.consumeLoop(channel, reader)
	}

	logger.Info(fmt.Sprintf("Kafka subscribed to channel: %s", channel))
	return nil
}

// Publish 发布消息
func (b *KafkaPubSub) Publish(ctx context.Context, channel string, msg *PubSubMessage) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if msg.MessageID == "" {
		msg.MessageID = generateMessageID()
	}
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal pubsub message: %w", err)
	}

	kMsg := kafka.Message{
		Topic: b.topicForChannel(channel),
		Key:   []byte(msg.Type),
		Value: payload,
		Time:  time.Now(),
	}

	if err := b.writer.WriteMessages(ctx, kMsg); err != nil {
		return fmt.Errorf("failed to publish Kafka pubsub message: %w", err)
	}

	atomic.AddUint64(&b.stats.totalPublished, 1)
	return nil
}

// Start 启动事件总线（lazy reader 已在订阅时创建）
func (b *KafkaPubSub) Start() error {
	logger.Info("Kafka pubsub initialized")
	return nil
}

// Stop 停止事件总线
func (b *KafkaPubSub) Stop() error {
	logger.Info("Stopping Kafka pubsub...")
	b.cancel()

	for _, reader := range b.readers {
		_ = reader.Close()
	}
	b.wg.Wait()

	if b.writer != nil {
		_ = b.writer.Close()
	}

	logger.Info("Kafka pubsub stopped")
	return nil
}

func (b *KafkaPubSub) consumeLoop(channel string, reader *kafka.Reader) {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
		}

		kMsg, err := reader.FetchMessage(b.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			logger.Warn(fmt.Sprintf("Kafka pubsub fetch error: %v", err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var msg PubSubMessage
		if err := json.Unmarshal(kMsg.Value, &msg); err != nil {
			logger.Error(fmt.Sprintf("Failed to decode Kafka pubsub message: %v", err))
			_ = reader.CommitMessages(context.Background(), kMsg)
			continue
		}

		atomic.AddUint64(&b.stats.totalReceived, 1)
		b.dispatch(channel, &msg)

		if err := reader.CommitMessages(context.Background(), kMsg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to commit Kafka pubsub message: %v", err))
		}
	}
}

func (b *KafkaPubSub) dispatch(channel string, msg *PubSubMessage) {
	b.subMutex.RLock()
	handlers := b.subscriptions[channel]
	b.subMutex.RUnlock()

	if len(handlers) == 0 {
		logger.Debug(fmt.Sprintf("No Kafka pubsub handlers for channel: %s", channel))
		return
	}

	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h PubSubHandler) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
			defer cancel()

			if err := h(ctx, msg); err != nil {
				logger.Error(fmt.Sprintf("Kafka pubsub handler error on %s: %v", channel, err))
				atomic.AddUint64(&b.stats.totalFailed, 1)
			} else {
				atomic.AddUint64(&b.stats.totalProcessed, 1)
			}
		}(handler)
	}

	wg.Wait()
}

func (b *KafkaPubSub) topicForChannel(channel string) string {
	return fmt.Sprintf("%s.events.%s", b.cfg.TopicPrefix, channel)
}

// GetStats 返回统计信息
func (b *KafkaPubSub) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_published": atomic.LoadUint64(&b.stats.totalPublished),
		"total_received":  atomic.LoadUint64(&b.stats.totalReceived),
		"total_processed": atomic.LoadUint64(&b.stats.totalProcessed),
		"total_failed":    atomic.LoadUint64(&b.stats.totalFailed),
	}
}

// NewKafkaPubSubFromConfig 基于配置初始化
func NewKafkaPubSubFromConfig(cfg config.Config) *KafkaPubSub {
	transport := buildKafkaTransport(cfg.Kafka)
	return NewKafkaPubSub(KafkaPubSubConfig{
		Brokers:     cfg.Kafka.Brokers,
		TopicPrefix: cfg.Kafka.TopicPrefix,
		GroupID:     cfg.Kafka.GroupID,
		Transport:   transport,
	})
}
