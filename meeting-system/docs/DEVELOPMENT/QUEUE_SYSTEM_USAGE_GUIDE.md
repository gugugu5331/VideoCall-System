# 消息队列系统使用指南

## 概述

meeting-system 的消息队列系统提供了完整的异步任务处理、事件分发和任务调度功能，支持 Redis 和内存两种实现方式。

## 核心组件

### 1. RedisMessageQueue - Redis消息队列

基于 Redis 的持久化消息队列，支持优先级、重试、死信队列等高级特性。

**特性**：
- ✅ 4级优先级（Critical/High/Normal/Low）
- ✅ 自动重试机制
- ✅ 死信队列（DLQ）
- ✅ 可见性超时
- ✅ 批量操作
- ✅ 工作协程池

**使用示例**：

```go
import (
    "context"
    "meeting-system/shared/queue"
    "meeting-system/shared/database"
)

// 初始化
redisClient := database.GetRedis()
msgQueue := queue.NewRedisMessageQueue(redisClient, queue.RedisMessageQueueConfig{
    QueueName:         "ai_tasks",
    Workers:           4,
    VisibilityTimeout: 30 * time.Second,
    PollInterval:      100 * time.Millisecond,
})

// 注册处理器
msgQueue.RegisterHandler("speech_recognition", func(ctx context.Context, msg *queue.Message) error {
    // 处理语音识别任务
    audioData := msg.Payload["audio_data"].(string)
    result := processAudio(audioData)
    return nil
})

// 启动队列
msgQueue.Start()
defer msgQueue.Stop()

// 发布消息
msg := &queue.Message{
    Type:       "speech_recognition",
    Priority:   queue.PriorityHigh,
    Payload:    map[string]interface{}{"audio_data": "base64..."},
    MaxRetries: 3,
    Timeout:    30,
}
msgQueue.Publish(context.Background(), msg)
```

### 2. RedisPubSubQueue - Redis发布订阅

基于 Redis Pub/Sub 的事件广播系统，支持多订阅者模式。

**特性**：
- ✅ 多订阅者模式
- ✅ 频道隔离
- ✅ 并发处理
- ✅ 自动重连

**使用示例**：

```go
// 初始化
pubsubQueue := queue.NewRedisPubSubQueue(redisClient)

// 订阅频道
pubsubQueue.Subscribe("meeting_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
    log.Printf("Meeting event: %s", msg.Type)
    return nil
})

// 启动
pubsubQueue.Start()
defer pubsubQueue.Stop()

// 发布事件
pubsubQueue.Publish(context.Background(), "meeting_events", &queue.PubSubMessage{
    Type: "user_joined",
    Payload: map[string]interface{}{
        "user_id":    123,
        "meeting_id": 456,
    },
})
```

### 3. LocalEventBus - 本地事件总线

基于 Go Channel 的高性能本地事件分发系统。

**特性**：
- ✅ 零延迟（纳秒级）
- ✅ 并发安全
- ✅ 多处理器支持
- ✅ 统计信息

**使用示例**：

```go
// 初始化
eventBus := queue.NewLocalEventBus(1000, 4)

// 注册事件处理器
eventBus.On("stream_started", func(ctx context.Context, event *queue.LocalEvent) error {
    streamID := event.Payload["stream_id"].(string)
    return handleStreamStart(streamID)
})

// 启动
eventBus.Start()
defer eventBus.Stop()

// 触发事件（非阻塞）
eventBus.Emit("stream_started", map[string]interface{}{
    "stream_id": "stream_123",
}, "media_service")
```

### 4. TaskScheduler - 任务调度器

支持延迟执行、优先级调度的任务调度系统。

**特性**：
- ✅ 优先级调度
- ✅ 延迟任务
- ✅ 超时控制
- ✅ 自动重试
- ✅ 任务取消

**使用示例**：

```go
// 初始化
scheduler := queue.NewTaskScheduler(1000, 8)

// 启动
scheduler.Start()
defer scheduler.Stop()

// 提交立即执行的任务
task := &queue.Task{
    Type:     "video_processing",
    Priority: queue.PriorityHigh,
    Handler: func(ctx context.Context, task *queue.Task) error {
        return processVideo(task.Payload)
    },
    Timeout:    60 * time.Second,
    MaxRetries: 3,
}
scheduler.SubmitTask(task)

// 提交延迟任务
delayedTask := &queue.Task{
    Type:        "cleanup",
    Priority:    queue.PriorityLow,
    Handler:     cleanupHandler,
    ScheduledAt: time.Now().Add(1 * time.Hour),
}
scheduler.SubmitTask(delayedTask)
```

### 5. TaskDispatcher - 任务分发器

智能任务路由和分发系统。

**特性**：
- ✅ 自动任务路由
- ✅ 任务状态跟踪
- ✅ 回调机制
- ✅ 批量分发

**使用示例**：

```go
// 初始化
dispatcher := queue.NewTaskDispatcher()

// 启动
dispatcher.Start()
defer dispatcher.Stop()

// 分发任务
request := &queue.TaskRequest{
    Type:       queue.TaskTypeSpeechRecognition,
    Priority:   queue.PriorityHigh,
    Payload:    map[string]interface{}{"audio": "data"},
    Timeout:    30,
    MaxRetries: 3,
}

response, err := dispatcher.DispatchTask(context.Background(), request)
if err != nil {
    log.Fatal(err)
}

log.Printf("Task %s status: %s", response.TaskID, response.Status)
```

### 6. QueueManager - 队列管理器

统一管理所有队列组件的管理器。

**使用示例**：

```go
import (
    "meeting-system/shared/config"
    "meeting-system/shared/database"
    "meeting-system/shared/queue"
)

// 加载配置
cfg := config.GetConfig()
redisClient := database.GetRedis()

// 初始化队列系统
qm, err := queue.InitializeQueueSystem(cfg, redisClient)
if err != nil {
    log.Fatal(err)
}
defer qm.Stop()

// 注册处理器
queue.RegisterCommonHandlers(qm, "my-service")

// 获取各个组件
redisQueue := qm.GetRedisMessageQueue()
pubsubQueue := qm.GetRedisPubSubQueue()
localBus := qm.GetLocalEventBus()
scheduler := qm.GetTaskScheduler()
dispatcher := qm.GetTaskDispatcher()

// 获取统计信息
stats := qm.GetStats()
log.Printf("Queue stats: %+v", stats)
```

## 快速开始

### 1. 在服务中集成

```go
package main

import (
    "meeting-system/shared/config"
    "meeting-system/shared/database"
    "meeting-system/shared/queue"
    "meeting-system/shared/logger"
)

func main() {
    // 加载配置
    cfg, _ := config.LoadConfig("config.yaml")
    
    // 初始化数据库
    database.InitDB(cfg.Database)
    database.InitRedis(cfg.Redis)
    redisClient := database.GetRedis()
    
    // 初始化队列系统
    qm, err := queue.InitializeQueueSystem(cfg, redisClient)
    if err != nil {
        logger.Fatal("Failed to initialize queue system: " + err.Error())
    }
    defer qm.Stop()
    
    // 注册消息处理器
    registerHandlers(qm)
    
    // 启动服务...
}

func registerHandlers(qm *queue.QueueManager) {
    // Redis消息队列处理器
    if redisQueue := qm.GetRedisMessageQueue(); redisQueue != nil {
        redisQueue.RegisterHandler("my_task", handleMyTask)
    }
    
    // 发布订阅处理器
    if pubsub := qm.GetRedisPubSubQueue(); pubsub != nil {
        pubsub.Subscribe("my_events", handleMyEvent)
    }
    
    // 本地事件处理器
    if localBus := qm.GetLocalEventBus(); localBus != nil {
        localBus.On("local_event", handleLocalEvent)
    }
}
```

### 2. 配置文件

在 `config.yaml` 中添加：

```yaml
message_queue:
  enabled: true
  type: "redis"  # redis 或 memory
  queue_name: "my_service"
  workers: 4
  visibility_timeout: 30
  poll_interval: 100
  max_retries: 3
  enable_dead_letter_queue: true

task_scheduler:
  enabled: true
  buffer_size: 1000
  workers: 8
  enable_delayed_tasks: true

event_bus:
  enabled: true
  type: "redis_pubsub"
  buffer_size: 1000
  workers: 4

task_dispatcher:
  enabled: true
  enable_routing: true
  enable_callbacks: true
```

## 最佳实践

### 1. 选择合适的队列类型

- **跨服务通信** → RedisMessageQueue
- **事件广播** → RedisPubSubQueue
- **服务内事件** → LocalEventBus
- **定时任务** → TaskScheduler
- **任务路由** → TaskDispatcher

### 2. 设置合理的优先级

```go
queue.PriorityCritical  // 系统关键任务（如安全检测）
queue.PriorityHigh      // 用户交互任务（如实时通信）
queue.PriorityNormal    // 常规任务（如数据处理）
queue.PriorityLow       // 后台任务（如日志清理）
```

### 3. 错误处理

```go
redisQueue.RegisterHandler("my_task", func(ctx context.Context, msg *queue.Message) error {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(fmt.Sprintf("Handler panic: %v", r))
        }
    }()
    
    // 处理任务
    if err := processTask(msg); err != nil {
        // 返回错误会触发重试
        return err
    }
    
    return nil
})
```

### 4. 监控和日志

```go
// 定期获取统计信息
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        stats := qm.GetStats()
        logger.Info(fmt.Sprintf("Queue stats: %+v", stats))
    }
}()
```

## 故障排查

### 1. 消息未被处理

- 检查处理器是否正确注册
- 检查队列是否已启动
- 查看日志中的错误信息

### 2. 消息重复处理

- 确保处理器是幂等的
- 检查可见性超时设置

### 3. 死信队列积压

```go
// 查看死信队列
dlqCount := redisQueue.GetDeadLetterCount()
log.Printf("Dead letter queue count: %d", dlqCount)

// 重新入队死信消息
count, err := redisQueue.RequeueDeadLetterMessages(100)
log.Printf("Requeued %d messages", count)
```

## 性能优化

1. **调整工作协程数**：根据 CPU 核心数和任务类型调整
2. **批量操作**：使用 `PublishBatch` 减少网络往返
3. **合理设置超时**：避免任务长时间占用资源
4. **监控队列长度**：及时发现积压问题

## 参考资料

- [README.md](./README.md) - 系统设计文档
- [TASK_DISPATCHER_GUIDE.md](./TASK_DISPATCHER_GUIDE.md) - 任务分发器详细指南
- [integration_example.go](./integration_example.go) - 完整集成示例

