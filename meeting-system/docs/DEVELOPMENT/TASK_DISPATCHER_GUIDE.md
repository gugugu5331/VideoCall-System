# 任务分发器使用指南（Kafka）

队列/事件总线默认使用 Kafka（单节点 KRaft，Broker `kafka:9092`）。本指南说明如何在服务内使用共享的 `QueueManager` 注册处理器、发布任务与事件。

## 架构要点

- 主题前缀：`meeting.*`（由 `kafka.topic_prefix` 控制）
- 任务队列：`<prefix>.tasks`，死信：`<prefix>.tasks.dlq`
- 事件总线：按通道命名，例如 `meeting.system_events`
- 内存模式：`message_queue.type=memory`、`event_bus.type=local` 可用于无 Kafka 的本地开发

## 快速示例（Go）

```go
import (
    "context"
    "meeting-system/shared/config"
    "meeting-system/shared/logger"
    "meeting-system/shared/queue"
)

func main() {
    config.InitConfig("config/config.yaml")
    logger.InitLogger(logger.LogConfig{Level: "info"})

    qm, err := queue.InitializeQueueSystem(config.GlobalConfig)
    if err != nil { panic(err) }
    defer qm.Stop()

    if mq := qm.GetKafkaMessageQueue(); mq != nil {
        mq.RegisterHandler("speech_recognition", func(ctx context.Context, msg *queue.Message) error {
            // 处理任务
            return nil
        })
    }

    if bus := qm.GetKafkaEventBus(); bus != nil {
        bus.Subscribe("system_events", func(ctx context.Context, msg *queue.PubSubMessage) error {
            logger.Info("收到事件: " + msg.Type)
            return nil
        })
    }

    _ = queue.PublishTask(qm, "speech_recognition", queue.PriorityHigh, map[string]any{
        "audio": "base64_data",
    }, "test-service")
}
```

## 运行时配置

- `message_queue.type`: `kafka`（默认）或 `memory`
- `kafka.brokers`: 默认 `["kafka:9092"]`，可通过环境变量覆盖
- `kafka.topic_prefix`: 默认 `meeting`
- 如需 SASL/TLS，填写 `kafka.sasl`、`kafka.tls` 字段
- 消费组与重试：`queue_manager` 内部管理重试次数/延迟，可在配置中调节，确保死信策略符合业务预期

## 运维提示

- 列出主题：`docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka:9092 --list`
- 查看消费滞后：`kafka-consumer-groups.sh --bootstrap-server kafka:9092 --describe --group <group>`
- 死信处理：`<prefix>.tasks.dlq`，按需消费或清理
- 本地无 Kafka 时，可切换到内存模式运行测试，但不具备持久化与多实例能力
- K8s/外部 Kafka：若使用托管 Kafka，确保网络可达并配置 SASL/TLS；必要时调整 `message_queue`/`event_bus` 超时与重试参数匹配云服务限制。
