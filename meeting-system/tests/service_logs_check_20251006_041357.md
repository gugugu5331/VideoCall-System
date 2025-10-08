# 服务日志检查报告

**检查时间**: 2025-10-06 04:13:58

## 总结

- ⚠️ user-service
- ⚠️ meeting-service
- ⚠️ media-service
- ⚠️ signaling-service
- ⚠️ ai-service

---


# 服务日志详细检查

## user-service

❌ 未找到队列系统初始化日志
❌ 未找到任务处理器注册日志
❌ 未找到 Redis 消息队列处理器日志
❌ 未找到 PubSub 处理器日志
❌ 未找到本地事件总线日志
📊 处理任务数: 0
0
📊 接收事件数: 0
0
✅ 无错误

## meeting-service

❌ 未找到队列系统初始化日志
❌ 未找到任务处理器注册日志
❌ 未找到 Redis 消息队列处理器日志
❌ 未找到 PubSub 处理器日志
❌ 未找到本地事件总线日志
📊 处理任务数: 0
0
📊 接收事件数: 0
0
✅ 无错误

## media-service

❌ 未找到队列系统初始化日志
❌ 未找到任务处理器注册日志
❌ 未找到 Redis 消息队列处理器日志
❌ 未找到 PubSub 处理器日志
❌ 未找到本地事件总线日志
📊 处理任务数: 0
0
📊 接收事件数: 0
0
⚠️ 错误数: 1

### 最近的错误
```
2025/10/02 02:37:17 Failed to load config: failed to unmarshal config: decoding failed due to the following error(s):
```

## signaling-service

❌ 未找到队列系统初始化日志
❌ 未找到任务处理器注册日志
❌ 未找到 Redis 消息队列处理器日志
❌ 未找到 PubSub 处理器日志
❌ 未找到本地事件总线日志
📊 处理任务数: 0
0
📊 接收事件数: 0
0
⚠️ 错误数: 3

### 最近的错误
```
2025/10/02 02:37:52 ERROR: [transport] Client received GoAway with error code ENHANCE_YOUR_CALM and debug data equal to ASCII "too_many_pings".
2025/10/02 02:37:52 ERROR: [transport] Client received GoAway with error code ENHANCE_YOUR_CALM and debug data equal to ASCII "too_many_pings".
2025/10/02 02:44:26 ERROR: [transport] Client received GoAway with error code ENHANCE_YOUR_CALM and debug data equal to ASCII "too_many_pings".
```

## ai-service

❌ 未找到队列系统初始化日志
❌ 未找到任务处理器注册日志
❌ 未找到 Redis 消息队列处理器日志
❌ 未找到 PubSub 处理器日志
❌ 未找到本地事件总线日志
📊 处理任务数: 0
0
📊 接收事件数: 0
0
✅ 无错误

# Redis 队列统计

## 队列长度

| 队列名称 | 长度 |
|---------|------|
| critical_queue | 0 |
| high_queue | 0 |
| normal_queue | 0 |
| low_queue | 0 |
| dead_letter_queue | 0 |

## Pub/Sub 频道

```

```

# Docker 容器日志检查

## user-service
```

```

## meeting-service
```

```

## media-service
```

```

## signaling-service
```

```

## ai-service
```

```

