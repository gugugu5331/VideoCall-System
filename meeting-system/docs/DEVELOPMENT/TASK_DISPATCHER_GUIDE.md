# 任务分发器使用指南

## 概述

任务分发器（Task Dispatcher）是一个高效的消息队列系统，用于从客户端接收任务并分发到指定的微服务。它提供了以下核心功能：

- ✅ **任务路由** - 自动将任务路由到正确的服务
- ✅ **优先级队列** - 支持4级优先级（Critical/High/Normal/Low）
- ✅ **Redis持久化** - 确保消息不丢失
- ✅ **异步处理** - 支持同步和异步任务提交
- ✅ **回调机制** - 任务完成后自动回调
- ✅ **超时控制** - 自动检测和处理超时任务
- ✅ **重试机制** - 失败任务自动重试
- ✅ **死信队列** - 处理失败的消息
- ✅ **事件同步** - 通过Redis Pub/Sub实现服务间事件同步

## 架构设计

```
┌─────────────────┐
│  客户端 (Web/Native) │
└────────┬────────┘
         │ WebSocket
         ▼
┌─────────────────────────┐
│  ClientTaskManager      │  ← 客户端任务管理器
│  - 会话管理             │
│  - 任务转换             │
│  - 响应推送             │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  TaskDispatcher         │  ← 任务分发器
│  - 任务路由             │
│  - 优先级队列           │
│  - 回调管理             │
│  - 超时检查             │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Redis Message Queue    │  ← Redis持久化队列
│  - Critical Queue       │
│  - High Queue           │
│  - Normal Queue         │
│  - Low Queue            │
│  - Processing Queue     │
│  - Dead Letter Queue    │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  微服务                 │
│  - AI Service           │
│  - Media Service        │
│  - Meeting Service      │
│  - Signaling Service    │
└─────────────────────────┘
```

## 快速开始

### 1. 初始化任务分发器

```go
package main

import (
    "context"
    "github.com/redis/go-redis/v9"
    "meeting-system/shared/queue"
    "meeting-system/shared/logger"
)

func main() {
    // 初始化Redis客户端
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 创建任务分发器
    dispatcher := queue.NewTaskDispatcher(redisClient)
    
    // 启动分发器
    if err := dispatcher.Start(); err != nil {
        logger.Fatal("Failed to start dispatcher: " + err.Error())
    }
    defer dispatcher.Stop()
    
    logger.Info("Task dispatcher started successfully")
}
```

### 2. 注册自定义路由

```go
// 注册自定义服务路由
dispatcher.RegisterRoute(&queue.ServiceRoute{
    ServiceName: "custom-service",
    QueueName:   "custom_tasks",
    TaskTypes: []queue.TaskType{
        "custom_task_type_1",
        "custom_task_type_2",
    },
})
```

### 3. 分发任务

#### 同步分发

```go
// 创建任务请求
request := &queue.TaskRequest{
    Type:       queue.TaskTypeSpeechRecognition,
    Priority:   queue.PriorityHigh,
    UserID:     123,
    MeetingID:  456,
    SessionID:  "session-abc",
    Payload: map[string]interface{}{
        "audio_data": "base64_encoded_audio",
        "format":     "wav",
        "sample_rate": 16000,
    },
    Timeout:    30,  // 30秒超时
    MaxRetries: 3,   // 最多重试3次
}

// 分发任务
ctx := context.Background()
response, err := dispatcher.DispatchTask(ctx, request)
if err != nil {
    logger.Error("Failed to dispatch task: " + err.Error())
    return
}

logger.Info("Task dispatched: " + response.TaskID)
```

#### 异步分发（带回调）

```go
// 异步分发任务
err := dispatcher.DispatchTaskAsync(ctx, request, func(ctx context.Context, response *queue.TaskResponse) error {
    if response.Status == queue.TaskStatusCompleted {
        logger.Info("Task completed successfully")
        logger.Info("Result: %v", response.Result)
    } else {
        logger.Error("Task failed: " + response.Error)
    }
    return nil
})
```

### 4. 批量分发任务

```go
requests := []*queue.TaskRequest{
    {
        Type:     queue.TaskTypeSpeechRecognition,
        Priority: queue.PriorityHigh,
        UserID:   1,
        Payload:  map[string]interface{}{"data": "audio1"},
    },
    {
        Type:     queue.TaskTypeEmotionDetection,
        Priority: queue.PriorityNormal,
        UserID:   2,
        Payload:  map[string]interface{}{"data": "image1"},
    },
}

responses, err := dispatcher.DispatchBatchTasks(ctx, requests)
if err != nil {
    logger.Error("Batch dispatch failed: " + err.Error())
}
```

## 客户端任务管理器

### 1. 初始化客户端任务管理器

```go
// 创建客户端任务管理器
clientManager := queue.NewClientTaskManager(dispatcher, redisClient)

// 启动管理器
if err := clientManager.Start(); err != nil {
    logger.Fatal("Failed to start client manager: " + err.Error())
}
defer clientManager.Stop()
```

### 2. 注册客户端会话

```go
// 当客户端连接时注册会话
clientManager.RegisterSession("session-123", userID, meetingID)

// 当客户端断开时注销会话
defer clientManager.UnregisterSession("session-123")
```

### 3. 提交客户端任务

```go
// 创建客户端任务请求
clientRequest := &queue.ClientTaskRequest{
    Type:      queue.ClientTaskAISpeechRecognition,
    UserID:    123,
    MeetingID: 456,
    SessionID: "session-123",
    Data: map[string]interface{}{
        "audio_data": "base64_encoded_audio",
        "format":     "wav",
    },
    Priority: "high",
    Timeout:  30,
}

// 提交任务
response, err := clientManager.SubmitTask(ctx, clientRequest)
if err != nil {
    logger.Error("Failed to submit task: " + err.Error())
    return
}

logger.Info("Client task submitted: " + response.RequestID)
```

### 4. 异步提交（WebSocket推送）

```go
// 异步提交任务
err := clientManager.SubmitTaskAsync(ctx, clientRequest)
if err != nil {
    logger.Error("Failed to submit async task: " + err.Error())
    return
}

// 获取响应通道
responseChan, exists := clientManager.GetResponseChannel(clientRequest.RequestID)
if exists {
    go func() {
        select {
        case response := <-responseChan:
            // 通过WebSocket推送响应给客户端
            sendToWebSocket(clientRequest.SessionID, response)
        case <-time.After(60 * time.Second):
            logger.Warn("Response timeout")
        }
    }()
}
```

## Redis持久化队列

### 1. 创建Redis队列

```go
// 创建Redis消息队列
redisQueue := queue.NewRedisMessageQueue(redisClient, &queue.RedisQueueConfig{
    QueueName:         "ai_tasks",
    Workers:           8,
    PollInterval:      1 * time.Second,
    VisibilityTimeout: 30 * time.Second,
    MaxRetries:        3,
})

// 注册消息处理器
redisQueue.RegisterHandler("speech_recognition", func(ctx context.Context, msg *queue.Message) error {
    // 处理语音识别任务
    logger.Info("Processing speech recognition task: " + msg.ID)
    
    // 执行AI推理
    result, err := processSpeechRecognition(msg.Payload)
    if err != nil {
        return err
    }
    
    // 发布结果
    publishResult(msg.ID, result)
    return nil
})

// 启动队列
if err := redisQueue.Start(); err != nil {
    logger.Fatal("Failed to start Redis queue: " + err.Error())
}
defer redisQueue.Stop()
```

### 2. 发布消息

```go
msg := &queue.Message{
    Type:     "speech_recognition",
    Priority: queue.PriorityHigh,
    Payload: map[string]interface{}{
        "audio_data": "base64_encoded_audio",
    },
    Timeout:    30,
    MaxRetries: 3,
}

err := redisQueue.Publish(ctx, msg)
if err != nil {
    logger.Error("Failed to publish message: " + err.Error())
}
```

### 3. 管理死信队列

```go
// 获取死信队列消息数量
count, err := redisQueue.GetDeadLetterCount(ctx)
logger.Info("Dead letter count: %d", count)

// 重新入队死信消息
requeued, err := redisQueue.RequeueDeadLetters(ctx, 10)
logger.Info("Requeued %d messages", requeued)

// 清除死信队列
err = redisQueue.PurgeDeadLetters(ctx)
```

## 监控和统计

### 1. 获取任务分发器统计

```go
stats := dispatcher.GetStats()
logger.Info("Dispatcher Stats:")
logger.Info("  Total Dispatched: %d", stats["total_dispatched"])
logger.Info("  Total Completed: %d", stats["total_completed"])
logger.Info("  Total Failed: %d", stats["total_failed"])
logger.Info("  Total Timeout: %d", stats["total_timeout"])
logger.Info("  Active Tasks: %d", stats["active_tasks"])
```

### 2. 获取队列统计

```go
stats := redisQueue.GetStats(ctx)
logger.Info("Queue Stats:")
logger.Info("  Total Published: %d", stats["total_published"])
logger.Info("  Total Processed: %d", stats["total_processed"])
logger.Info("  Critical Pending: %d", stats["critical_pending"])
logger.Info("  High Pending: %d", stats["high_pending"])
logger.Info("  Normal Pending: %d", stats["normal_pending"])
logger.Info("  Low Pending: %d", stats["low_pending"])
logger.Info("  Processing Count: %d", stats["processing_count"])
logger.Info("  Dead Letter Count: %d", stats["dead_letter_count"])
```

### 3. 获取客户端管理器统计

```go
stats := clientManager.GetStats()
logger.Info("Client Manager Stats:")
logger.Info("  Total Requests: %d", stats["total_requests"])
logger.Info("  Total Success: %d", stats["total_success"])
logger.Info("  Total Failed: %d", stats["total_failed"])
logger.Info("  Active Sessions: %d", stats["active_sessions"])
```

## 最佳实践

### 1. 任务优先级设置

- **Critical**: 紧急任务（如紧急通知、系统告警）
- **High**: 高优先级任务（如实时AI处理、用户交互）
- **Normal**: 普通任务（如常规数据处理）
- **Low**: 低优先级任务（如日志记录、统计分析）

### 2. 超时设置

- AI任务: 30-60秒
- 媒体处理: 60-120秒
- 数据库操作: 5-10秒
- 网络请求: 10-30秒

### 3. 重试策略

- 网络错误: 重试3次
- 临时错误: 重试2次
- 业务错误: 不重试

### 4. 会话管理

- 定期清理不活跃会话（30分钟）
- 限制每个会话的并发任务数
- 记录会话活动日志

## 故障处理

### 1. 任务超时

任务超时会自动标记为`TaskStatusTimeout`，并触发回调。

### 2. 处理失败

失败的任务会根据`MaxRetries`自动重试，超过最大重试次数后移入死信队列。

### 3. Redis连接失败

系统会自动尝试重连，期间任务会缓存在本地队列。

### 4. 死信队列处理

定期检查死信队列，分析失败原因，必要时重新入队或清除。

## 性能优化

1. **调整Worker数量**: 根据CPU核心数和任务类型调整
2. **批量操作**: 使用批量发布减少Redis往返
3. **连接池**: 使用Redis连接池提高并发性能
4. **监控告警**: 设置队列长度告警阈值
5. **定期清理**: 清理已完成的任务和不活跃会话
