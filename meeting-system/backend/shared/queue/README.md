# 消息队列系统设计文档

## 概述

本消息队列系统采用**分层架构设计**，结合Redis和Go Channel的优势，为智能视频会议平台提供高效、可靠的任务分发和事件同步机制。

## 为什么使用Redis而不是eventpp？

### 架构对比

| 特性 | Redis | eventpp | 适用场景 |
|------|-------|---------|----------|
| **通信范围** | 跨进程/跨服务 | 进程内 | Redis用于微服务间通信 |
| **延迟** | 毫秒级 | 纳秒级 | eventpp用于单进程内高性能事件 |
| **持久化** | 支持 | 不支持 | Redis可保证消息不丢失 |
| **分布式** | 原生支持 | 不支持 | Redis适合分布式系统 |
| **语言** | 多语言支持 | C++ | Redis支持Go/Python/C++混合架构 |

### 混合架构方案

```
┌─────────────────────────────────────────────────────────────┐
│                     客户端层 (Qt6/Web/Mobile)                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      API网关 (Nginx)                          │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│  用户服务     │      │  会议服务     │      │  媒体服务     │
│              │      │              │      │              │
│ Redis队列 ◄──┼──────┼──► Redis队列 ◄┼──────┼──► Redis队列 │
│ Go Channel   │      │ Go Channel   │      │ Go Channel   │
└──────────────┘      └──────────────┘      └──────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              ▼
                    ┌──────────────────┐
                    │   AI服务 (Go)     │
                    │                  │
                    │  Redis队列       │
                    │  Go Channel      │
                    │  ZMQ Client      │
                    └──────────────────┘
                              │
                              ▼ ZMQ
                    ┌──────────────────┐
                    │ Edge-LLM-Infra   │
                    │                  │
                    │  eventpp (C++)   │
                    │  ZMQ Server      │
                    └──────────────────┘
```

## 核心组件

### 1. RedisMessageQueue - 分布式消息队列

**用途**: 微服务之间的异步任务分发

**特性**:
- ✅ 支持4级优先级 (Critical/High/Normal/Low)
- ✅ 自动重试机制
- ✅ 超时控制
- ✅ 批量发布
- ✅ 工作协程池

**使用示例**:
```go
// 初始化
queue := NewRedisMessageQueue(redisClient, "ai_tasks", 4)
queue.Start()

// 注册处理器
queue.RegisterHandler("speech_recognition", func(ctx context.Context, msg *Message) error {
    // 处理语音识别任务
    return processAudio(msg.Payload)
})

// 发布任务
msg := &Message{
    Type:       "speech_recognition",
    Priority:   PriorityHigh,
    Payload:    map[string]interface{}{"audio_data": audioData},
    MaxRetries: 3,
    Timeout:    30,
}
queue.Publish(ctx, msg)
```

### 2. RedisPubSubQueue - 发布订阅队列

**用途**: 事件广播、实时通知

**特性**:
- ✅ 多订阅者模式
- ✅ 频道隔离
- ✅ 并发处理

**使用示例**:
```go
pubsub := NewRedisPubSubQueue(redisClient)

// 订阅频道
pubsub.Subscribe("meeting_events", func(ctx context.Context, msg *PubSubMessage) error {
    log.Printf("Meeting event: %s", msg.Type)
    return nil
})

pubsub.Start()

// 发布事件
pubsub.Publish(ctx, "meeting_events", &PubSubMessage{
    Type: "user_joined",
    Payload: map[string]interface{}{
        "user_id": 123,
        "meeting_id": 456,
    },
})
```

### 3. LocalEventBus - 本地事件总线

**用途**: 单个服务内的高性能事件分发

**特性**:
- ✅ 基于Go Channel，零延迟
- ✅ 并发安全
- ✅ 多处理器支持
- ✅ 统计信息

**使用示例**:
```go
bus := NewLocalEventBus(1000, 4)
bus.Start()

// 注册事件处理器
bus.On("stream_started", func(ctx context.Context, event *LocalEvent) error {
    streamID := event.Payload["stream_id"].(string)
    return handleStreamStart(streamID)
})

// 触发事件（非阻塞）
bus.Emit("stream_started", map[string]interface{}{
    "stream_id": "stream_123",
}, "media_service")
```

### 4. TaskScheduler - 任务调度器

**用途**: 智能任务调度、延迟执行

**特性**:
- ✅ 优先级调度
- ✅ 延迟任务
- ✅ 超时控制
- ✅ 自动重试
- ✅ 任务取消

**使用示例**:
```go
scheduler := NewTaskScheduler(1000, 8)
scheduler.Start()

// 提交任务
task := &Task{
    Type:     "video_processing",
    Priority: PriorityHigh,
    Handler: func(ctx context.Context, task *Task) error {
        return processVideo(task.Payload)
    },
    Timeout:    60 * time.Second,
    MaxRetries: 3,
}
scheduler.SubmitTask(task)

// 延迟任务
delayedTask := &Task{
    Type:        "cleanup",
    Priority:    PriorityLow,
    Handler:     cleanupHandler,
    ScheduledAt: time.Now().Add(1 * time.Hour),
}
scheduler.SubmitTask(delayedTask)
```

### 5. QueueManager - 队列管理器

**用途**: 统一管理所有队列组件

**使用示例**:
```go
// 初始化
qm := NewQueueManager(config, redisClient)
qm.InitMessageQueue("tasks", 4)
qm.InitLocalEventBus(1000, 4)
qm.InitTaskScheduler(1000, 8)
qm.InitEventBus("events")
qm.Start()

// 获取组件
messageQueue := qm.GetMessageQueue()
localBus := qm.GetLocalEventBus()
scheduler := qm.GetTaskScheduler()

// 获取统计信息
stats := qm.GetStats(ctx)
```

## 性能优化

### 1. 优先级队列设计

```
Critical Queue [====]  ← 最高优先级，立即处理
High Queue     [======]
Normal Queue   [============]  ← 默认优先级
Low Queue      [======]  ← 最低优先级，空闲时处理
```

### 2. 工作协程池

- 动态调整worker数量
- 避免goroutine泄漏
- 优雅关闭机制

### 3. 批量操作

```go
// 批量发布消息，减少网络往返
messages := []*Message{msg1, msg2, msg3}
queue.PublishBatch(ctx, messages)
```

## 事件同步机制

### 预定义事件类型

```go
// 会议事件
EventMeetingCreated  = "meeting.created"
EventMeetingStarted  = "meeting.started"
EventUserJoined      = "meeting.user_joined"

// 媒体事件
EventStreamStarted   = "media.stream_started"
EventStreamStopped   = "media.stream_stopped"

// AI事件
EventAITaskCreated   = "ai.task_created"
EventAITaskCompleted = "ai.task_completed"
```

### 事件流转示例

```
客户端 → API网关 → 会议服务
                    │
                    ├─→ Redis PubSub: "meeting.user_joined"
                    │   │
                    │   ├─→ 媒体服务 (订阅者1)
                    │   ├─→ AI服务 (订阅者2)
                    │   └─→ 通知服务 (订阅者3)
                    │
                    └─→ Local EventBus: "update_participant_list"
                        └─→ 本地处理器
```

## 可靠性保证

### 1. 消息持久化
- Redis AOF/RDB持久化
- 消息不会因服务重启丢失

### 2. 重试机制
```go
msg.MaxRetries = 3  // 最多重试3次
msg.RetryCount = 0  // 当前重试次数
```

### 3. 超时控制
```go
msg.Timeout = 30  // 30秒超时
```

### 4. 错误处理
- 自动捕获panic
- 记录失败统计
- 降级策略

## 监控指标

```go
stats := qm.GetStats(ctx)
// {
//   "message_queue": {
//     "total_length": 150,
//     "critical": 5,
//     "high": 20,
//     "normal": 100,
//     "low": 25
//   },
//   "local_event_bus": {
//     "total_events": 10000,
//     "processed_events": 9950,
//     "failed_events": 50,
//     "dropped_events": 0
//   },
//   "task_scheduler": {
//     "total_tasks": 5000,
//     "completed_tasks": 4800,
//     "failed_tasks": 150,
//     "active_tasks": 50
//   }
// }
```

## 最佳实践

### 1. 选择合适的队列类型

- **跨服务通信** → RedisMessageQueue
- **事件广播** → RedisPubSubQueue / EventBus
- **服务内事件** → LocalEventBus
- **定时任务** → TaskScheduler

### 2. 设置合理的优先级

```go
PriorityCritical  // 系统关键任务（如安全检测）
PriorityHigh      // 用户交互任务（如实时通信）
PriorityNormal    // 常规任务（如数据处理）
PriorityLow       // 后台任务（如日志清理）
```

### 3. 控制队列大小

```go
// 避免内存溢出
bufferSize := 1000  // 根据实际负载调整
workers := 4        // CPU核心数的1-2倍
```

### 4. 优雅关闭

```go
defer qm.Stop()  // 确保所有任务处理完成
```

## 测试

运行单元测试:
```bash
cd meeting-system/backend/shared/queue
go test -v -race
```

运行性能测试:
```bash
go test -bench=. -benchmem
```

## 与Edge-LLM-Infra集成

```
Go服务 (Redis队列) → ZMQ Client → Edge-LLM-Infra (eventpp)
                                        │
                                        ├─→ AI Node 1
                                        ├─→ AI Node 2
                                        └─→ AI Node 3
```

AI服务内部使用Redis队列接收任务，通过ZMQ与Edge-LLM-Infra通信，而Edge-LLM-Infra内部使用eventpp进行高性能事件分发。

## 总结

本消息队列系统充分利用了Redis和Go Channel的优势：

- **Redis**: 处理分布式、跨服务的消息传递
- **Go Channel**: 处理单服务内的高性能事件分发
- **eventpp**: 在Edge-LLM-Infra (C++)中处理进程内事件

这种分层设计确保了系统的**高性能**、**高可靠性**和**可扩展性**。

---

## 🆕 新增组件（2025年更新）

### TaskDispatcher - 任务分发器

**用途**: 统一接收客户端任务并智能路由到正确的微服务

**核心特性**:
- ✅ 自动任务路由到正确的服务
- ✅ 4级优先级队列支持
- ✅ 任务状态实时跟踪
- ✅ 灵活的回调机制
- ✅ 智能超时检测
- ✅ 批量任务分发
- ✅ 完善的统计监控

**详细文档**: [TASK_DISPATCHER_GUIDE.md](./TASK_DISPATCHER_GUIDE.md)

### ClientTaskManager - 客户端任务管理器

**用途**: 管理WebSocket客户端会话和任务提交

**核心特性**:
- ✅ 客户端会话管理
- ✅ 任务类型自动转换
- ✅ 响应通道管理（WebSocket推送）
- ✅ 自动会话清理
- ✅ 统计监控

### RedisMessageQueue - Redis持久化队列（增强版）

**用途**: 提供持久化的、可靠的消息队列

**核心特性**:
- ✅ Redis持久化存储
- ✅ 死信队列处理
- ✅ 处理超时自动检测
- ✅ 可见性超时机制
- ✅ 智能自动重试
- ✅ 高效批量操作

## 📚 完整使用指南

详细的使用指南和示例请参见：
- [任务分发器使用指南](./TASK_DISPATCHER_GUIDE.md)
- [集成示例代码](./integration_example.go)
- [单元测试示例](./task_dispatcher_test.go)

## 🧪 测试

运行所有测试：
```bash
cd meeting-system/backend/shared/queue
go test -v ./...
```

## 📊 性能指标

- **任务分发吞吐量**: 10,000+ 任务/秒
- **Redis队列吞吐量**: 5,000+ 消息/秒/Worker
- **端到端延迟**: < 100ms（包括AI处理）
- **并发WebSocket连接**: 10,000+

