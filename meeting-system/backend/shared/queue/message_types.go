package queue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// MessagePriority 表示消息优先级
type MessagePriority int

const (
	PriorityLow      MessagePriority = 0
	PriorityNormal   MessagePriority = 1
	PriorityHigh     MessagePriority = 2
	PriorityCritical MessagePriority = 3
)

// Message 消息结构（用于任务队列）
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

	ProcessingStartTime int64  `json:"processing_start_time,omitempty"`
	VisibilityTimeout   int64  `json:"visibility_timeout,omitempty"` // 秒
	DeadLetterQueue     string `json:"dead_letter_queue,omitempty"`
}

// MessageHandler 定义消息处理函数
type MessageHandler func(ctx context.Context, msg *Message) error

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

// generateMessageID 生成唯一消息ID
func generateMessageID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
