package queue

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// setupTestRedis 创建测试用的Redis实例
func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	
	return client, mr
}

func TestRedisMessageQueue_PublishAndConsume(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()
	
	// 创建队列
	queue := NewRedisMessageQueue(client, RedisMessageQueueConfig{
		QueueName:         "test_queue",
		Workers:           2,
		VisibilityTimeout: 5 * time.Second,
		PollInterval:      10 * time.Millisecond,
	})
	
	// 注册处理器
	processed := make(chan string, 1)
	queue.RegisterHandler("test_task", func(ctx context.Context, msg *Message) error {
		processed <- msg.ID
		return nil
	})
	
	// 启动队列
	err := queue.Start()
	assert.NoError(t, err)
	defer queue.Stop()
	
	// 发布消息
	msg := &Message{
		Type:     "test_task",
		Priority: PriorityNormal,
		Payload: map[string]interface{}{
			"data": "test",
		},
	}
	
	err = queue.Publish(context.Background(), msg)
	assert.NoError(t, err)
	
	// 等待消息被处理
	select {
	case msgID := <-processed:
		assert.Equal(t, msg.ID, msgID)
	case <-time.After(2 * time.Second):
		t.Fatal("Message not processed within timeout")
	}
	
	// 检查统计信息
	stats := queue.GetStats()
	assert.Equal(t, uint64(1), stats["total_published"])
	assert.Equal(t, uint64(1), stats["total_processed"])
}

func TestRedisMessageQueue_Priority(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()
	
	queue := NewRedisMessageQueue(client, RedisMessageQueueConfig{
		QueueName:         "test_queue",
		Workers:           1,
		VisibilityTimeout: 5 * time.Second,
		PollInterval:      10 * time.Millisecond,
	})
	
	processed := make(chan MessagePriority, 3)
	queue.RegisterHandler("test_task", func(ctx context.Context, msg *Message) error {
		processed <- msg.Priority
		return nil
	})
	
	err := queue.Start()
	assert.NoError(t, err)
	defer queue.Stop()
	
	// 发布不同优先级的消息
	priorities := []MessagePriority{PriorityLow, PriorityNormal, PriorityCritical}
	for _, priority := range priorities {
		msg := &Message{
			Type:     "test_task",
			Priority: priority,
			Payload:  map[string]interface{}{},
		}
		err = queue.Publish(context.Background(), msg)
		assert.NoError(t, err)
	}
	
	// 验证处理顺序（应该是Critical -> Normal -> Low）
	expectedOrder := []MessagePriority{PriorityCritical, PriorityNormal, PriorityLow}
	for i := 0; i < 3; i++ {
		select {
		case priority := <-processed:
			assert.Equal(t, expectedOrder[i], priority)
		case <-time.After(2 * time.Second):
			t.Fatal("Message not processed within timeout")
		}
	}
}

func TestRedisMessageQueue_Retry(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()
	
	queue := NewRedisMessageQueue(client, RedisMessageQueueConfig{
		QueueName:         "test_queue",
		Workers:           1,
		VisibilityTimeout: 5 * time.Second,
		PollInterval:      10 * time.Millisecond,
	})
	
	attempts := 0
	queue.RegisterHandler("test_task", func(ctx context.Context, msg *Message) error {
		attempts++
		if attempts < 3 {
			return assert.AnError // 前两次失败
		}
		return nil // 第三次成功
	})
	
	err := queue.Start()
	assert.NoError(t, err)
	defer queue.Stop()
	
	msg := &Message{
		Type:       "test_task",
		Priority:   PriorityNormal,
		Payload:    map[string]interface{}{},
		MaxRetries: 3,
	}
	
	err = queue.Publish(context.Background(), msg)
	assert.NoError(t, err)
	
	// 等待重试完成
	time.Sleep(5 * time.Second)
	
	// 验证重试次数
	assert.Equal(t, 3, attempts)
}

func TestRedisMessageQueue_BatchPublish(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()
	
	queue := NewRedisMessageQueue(client, RedisMessageQueueConfig{
		QueueName:         "test_queue",
		Workers:           2,
		VisibilityTimeout: 5 * time.Second,
		PollInterval:      10 * time.Millisecond,
	})
	
	processed := make(chan string, 10)
	queue.RegisterHandler("test_task", func(ctx context.Context, msg *Message) error {
		processed <- msg.ID
		return nil
	})
	
	err := queue.Start()
	assert.NoError(t, err)
	defer queue.Stop()
	
	// 批量发布消息
	messages := make([]*Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = &Message{
			Type:     "test_task",
			Priority: PriorityNormal,
			Payload:  map[string]interface{}{"index": i},
		}
	}
	
	err = queue.PublishBatch(context.Background(), messages)
	assert.NoError(t, err)
	
	// 等待所有消息被处理
	processedCount := 0
	timeout := time.After(5 * time.Second)
	for processedCount < 10 {
		select {
		case <-processed:
			processedCount++
		case <-timeout:
			t.Fatalf("Only %d messages processed within timeout", processedCount)
		}
	}
	
	assert.Equal(t, 10, processedCount)
}

func TestRedisPubSubQueue_PublishAndSubscribe(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()
	
	queue := NewRedisPubSubQueue(client)
	
	received := make(chan string, 1)
	queue.Subscribe("test_channel", func(ctx context.Context, msg *PubSubMessage) error {
		received <- msg.Type
		return nil
	})
	
	err := queue.Start()
	assert.NoError(t, err)
	defer queue.Stop()
	
	// 等待订阅建立
	time.Sleep(100 * time.Millisecond)
	
	// 发布消息
	msg := &PubSubMessage{
		Type:    "test_event",
		Payload: map[string]interface{}{"data": "test"},
		Source:  "test",
	}
	
	err = queue.Publish(context.Background(), "test_channel", msg)
	assert.NoError(t, err)
	
	// 等待消息接收
	select {
	case eventType := <-received:
		assert.Equal(t, "test_event", eventType)
	case <-time.After(2 * time.Second):
		t.Fatal("Message not received within timeout")
	}
}

func TestMemoryMessageQueue_PublishAndConsume(t *testing.T) {
	queue := NewMemoryMessageQueue("test_queue", 100, 2)
	
	processed := make(chan string, 1)
	queue.RegisterHandler("test_task", func(ctx context.Context, msg *Message) error {
		processed <- msg.ID
		return nil
	})
	
	err := queue.Start()
	assert.NoError(t, err)
	defer queue.Stop()
	
	msg := &Message{
		Type:     "test_task",
		Priority: PriorityNormal,
		Payload:  map[string]interface{}{"data": "test"},
	}
	
	err = queue.Publish(context.Background(), msg)
	assert.NoError(t, err)
	
	// 内存队列是同步的，消息应该已经被处理
	select {
	case msgID := <-processed:
		assert.Equal(t, msg.ID, msgID)
	case <-time.After(1 * time.Second):
		t.Fatal("Message not processed within timeout")
	}
}

func TestQueueManager_Integration(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()
	
	// 创建队列管理器
	qm := NewQueueManager(nil, client)
	
	// 初始化各个组件
	qm.InitRedisMessageQueue(client, RedisMessageQueueConfig{
		QueueName:         "test_queue",
		Workers:           2,
		VisibilityTimeout: 5 * time.Second,
		PollInterval:      10 * time.Millisecond,
	})
	qm.InitRedisPubSubQueue(client)
	qm.InitLocalEventBus(100, 2)
	qm.InitTaskScheduler(100, 2)
	qm.InitTaskDispatcher()
	
	// 启动所有组件
	err := qm.Start()
	assert.NoError(t, err)
	defer qm.Stop()
	
	// 验证所有组件都已初始化
	assert.NotNil(t, qm.GetRedisMessageQueue())
	assert.NotNil(t, qm.GetRedisPubSubQueue())
	assert.NotNil(t, qm.GetLocalEventBus())
	assert.NotNil(t, qm.GetTaskScheduler())
	assert.NotNil(t, qm.GetTaskDispatcher())
	
	// 获取统计信息
	stats := qm.GetStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "redis_message_queue")
	assert.Contains(t, stats, "redis_pubsub_queue")
	assert.Contains(t, stats, "local_event_bus")
	assert.Contains(t, stats, "task_scheduler")
	assert.Contains(t, stats, "task_dispatcher")
}

