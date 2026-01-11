# 💾 数据与存储设计

> 权威 schema 位于 `backend/shared/database/schema.sql`。本页概述各存储的职责、关键表与约定，便于对照服务实现。

## 总览

- **PostgreSQL**：核心业务数据（用户、会议、信令、录制、配置、AI 任务）。
- **Redis**：会话/房间状态缓存、速率限制、短期数据。
- **Kafka**：任务队列与事件总线（主题前缀 `meeting.*`，配置见 `backend/config/*.yaml`）。
- **MinIO**：录制文件、媒资上传、头像等对象。
- **MongoDB**：AI 结果或分析数据（可选）。
- **etcd**：服务注册/配置命名空间（仅内部使用）。

## PostgreSQL 主要表

| 表 | 用途 | 关键字段/索引 |
| --- | --- | --- |
| `users` | 用户账号、角色、状态 | `username`/`email` 唯一；`role`/`status` 索引 |
| `meetings` | 会议基本信息/设置 | `creator_id`、`status`、`start_time` 索引 |
| `meeting_participants` | 参会者关系 | `(meeting_id, user_id)` 唯一；状态/角色索引 |
| `meeting_rooms` | WebRTC 房间/节点 | `room_id` 唯一，`status`/`meeting_id` 索引 |
| `media_streams` | 音视频流元数据 | `room_id`/`user_id`/`stream_type` 索引 |
| `meeting_recordings` | 录制文件元数据 | `meeting_id`/`status` 索引，记录路径/时长/格式 |
| `signaling_sessions` | WS 会话/房间状态快照 | `session_id` 唯一，`meeting_id`/`status` 索引 |
| `signaling_messages` | 部分信令/聊天持久化 | `message_id` 唯一，`meeting_id`/`from_user_id` 索引 |
| `ai_tasks` | AI 推理任务追踪 | `task_id` 唯一，`task_type`/`status`/`priority` 索引 |
| `system_configs` | 运行时配置键值 | `config_key` 唯一，公共配置标记 |
| `operation_logs` | 管理/操作审计 | `user_id`、资源类型/ID、时间索引 |
| 视图 `active_meetings_stats` | 活跃会议统计 | 便于报表/监控 |

所有表包含 `created_at/updated_at` 触发器，部分表具备软删除列 `deleted_at`。

## Redis 约定

- **会话/房间**：`session:{id}`、`room:{meeting_id}`，保存当前参会者/心跳；TTL 按场景设置（会议房间默认天级）。
- **限流/锁**：按需在服务配置启用，键格式 `rate:{user}`、`lock:{resource}`。
- **缓存**：热点会议/用户信息短期缓存；写后更新或依赖 TTL 自动失效。

## Kafka 主题（队列/事件）

- `meeting.tasks` / `meeting.tasks.dlq`：任务队列与死信。
- `meeting.system_events` 等自定义事件主题：由 `kafka.topic_prefix` 控制。
- 消费组、重试与优先级由 `backend/shared/queue` 管理，使用时参考 `docs/DEVELOPMENT/TASK_DISPATCHER_GUIDE.md`。

## MinIO 结构（建议）

```
meeting-system/
├── recordings/<meeting_id>/xxx.m3u8|mp4|log
├── media/uploads/<uuid>.<ext>
├── media/avatars/<user_id>.jpg
└── temp/
```

桶与目录在 `backend/config/media-service.yaml` 中配置；上传/录制接口返回对象路径供前端访问。

## MongoDB（可选）

AI 结果或分析可落在 `ai_results` 等集合，字段由实际模型输出决定；未启用 MongoDB 时不会影响主业务链路。

## etcd 命名空间（示例）

```
/meeting-system/config/services/<service>/{host,port,grpc_port}
/meeting-system/services/<service>/instance_<id> -> {host,port,metadata}
```

服务发现/注册可按需开启；生产部署可改用外部配置中心。

## 数据安全与备份

- 替换默认数据库/MinIO/Kafka/Redis 凭据；生产仅暴露必要端口。
- 定期备份 PostgreSQL（全量+增量）与 MinIO；etcd 建议定时快照。
- 清理策略：录制/日志按存储配额定期归档或删除，避免桶无限增长。

## 连接与性能建议

- PostgreSQL 连接池：`max_idle_conns=10`、`max_open_conns=100`、`conn_max_lifetime=3600s` 可作为起点，根据负载调整。
- 索引维护：定期 `VACUUM ANALYZE`，对高频写表监控膨胀；新增查询路径时补充覆盖索引。
- Redis：为会话/房间设置合理 TTL，避免键数量无限增长；限流/锁务必设置过期时间。
- Kafka：监控消费滞后并按需扩容分区/消费者；死信队列需定期消费或清理。
