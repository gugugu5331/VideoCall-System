# 会议系统项目面试快速参考

**项目**: meeting-system-server (智能视频会议平台)  
**面试准备时间**: 30 分钟速查版

---

## 一、项目核心信息

### 项目概述
- **定位**: 基于 Edge-LLM-Infra 的企业级视频会议系统
- **架构**: 微服务 + SFU + 分布式 AI 推理
- **规模**: 支持 1000+ 并发用户，10-50 人会议
- **技术栈**: Go + WebRTC + Edge-LLM-Infra + Docker

### 核心功能
1. **音视频通话**: SFU 架构，低延迟 (<100ms)
2. **AI 功能**: 语音识别、情绪检测、合成检测
3. **会议管理**: 创建、加入、录制、回放
4. **实时通信**: WebSocket 信令、聊天、屏幕共享

### 系统架构
```
客户端 → Nginx → 微服务层 → AI推理层 → 数据层

微服务层 (5个):
- user-service (8080): 用户认证
- meeting-service (8082): 会议管理
- signaling-service (8081): WebSocket 信令
- media-service (8083): SFU 媒体转发
- ai-inference-service (8085): AI 推理网关

AI推理层:
- edge-model-infra (10001): C++ 单元管理器
- ai-inference-worker (5010): Python 推理节点

数据层:
- PostgreSQL: 用户、会议数据
- MongoDB: AI 分析结果
- Redis: 缓存、消息队列
- MinIO: 录制文件存储
```

---

## 二、必答问题准备

### Q1: 为什么选择 SFU 架构？

**答案要点**:
- **SFU 优势**: 不转码，服务器负载低，延迟低
- **vs P2P**: 扩展性更好（支持 10-50 人）
- **vs MCU**: 成本更低，不需要转码
- **适用场景**: 中小型会议，平衡性能和成本

### Q2: WebRTC 信令流程？

**答案要点**:
1. 客户端加入会议 → 获取 ICE servers
2. 建立 WebSocket 连接 → 信令服务验证 JWT
3. 创建 PeerConnection → 添加本地媒体流
4. 发送 Offer → 信令服务转发
5. 接收 Answer → 设置远程描述
6. ICE 候选交换 → 建立媒体连接
7. 媒体流传输 → SFU 转发

### Q3: Edge-LLM-Infra 集成？

**答案要点**:
- **架构**: Go → TCP → unit-manager (C++) → IPC → AI Node (C++) → ZMQ → Python Worker
- **通信协议**: JSON over TCP
- **消息格式**: {request_id, work_id, action, object, data}
- **生命周期**: Setup → Inference → Exit
- **挑战**: 跨语言通信、数据格式对齐、超时控制

### Q4: 如何保证高并发性能？

**答案要点**:
1. **连接池**: 数据库、Redis 连接池
2. **Goroutine 池**: 限制并发数量
3. **消息队列**: 异步处理 AI 推理
4. **缓存**: 多级缓存（本地 + Redis）
5. **限流**: Nginx 限流 + 服务端限流
6. **水平扩展**: 无状态服务 + 负载均衡

### Q5: 遇到的最大技术难点？

**答案要点**:
- **问题**: Edge-LLM-Infra 集成，C++ 和 Go 通信
- **挑战**: 数据格式不匹配、ZMQ 消息理解、超时控制
- **解决**: 
  - 详细阅读源码和文档
  - 使用 tcpdump 抓包分析
  - 编写测试脚本验证
  - 逐步调试，最终成功集成

---

## 三、技术深度问题

### 数据库设计

**为什么用三种数据库？**
- **PostgreSQL**: 结构化数据（用户、会议），ACID 事务
- **MongoDB**: 灵活 Schema（AI 分析结果、聊天记录）
- **Redis**: 高性能缓存、消息队列、会话管理

**关键表设计**:
```sql
-- 用户表
users (id, username, email, password_hash, role, status)
索引: username, email

-- 会议表
meetings (id, title, creator_id, start_time, end_time, status, settings)
索引: creator_id, start_time

-- 参与者表
meeting_participants (id, meeting_id, user_id, role, status)
索引: meeting_id, user_id
唯一约束: (meeting_id, user_id)
```

### 性能优化案例

**优化 1: 数据库查询**
- 问题: N+1 查询，2s 延迟
- 方案: 使用 Preload 预加载，添加索引
- 效果: 2s → 50ms，提升 40x

**优化 2: Redis 缓存**
- 问题: 频繁查询数据库
- 方案: 添加 Redis 缓存，TTL 5 分钟
- 效果: 缓存命中率 95%，响应时间 10ms → 1ms

**优化 3: 消息队列异步化**
- 问题: AI 推理阻塞 HTTP 请求
- 方案: 使用 Redis 队列异步处理
- 效果: API 响应 30s → 50ms，吞吐量 2 → 50 req/s

### 安全措施

1. **认证**: JWT Token，24 小时过期
2. **授权**: RBAC 权限控制
3. **加密**: bcrypt 密码加密，HTTPS 传输
4. **防护**: SQL 注入防护、XSS 防护、限流防 DDoS
5. **审计**: 操作日志记录

---

## 四、监控与运维

### 监控指标

**业务指标**:
- 在线用户数
- 会议创建数
- AI 推理请求数

**性能指标**:
- API P95 延迟 < 100ms
- API 错误率 < 1%
- AI 推理延迟 < 5s

**系统指标**:
- CPU 使用率 < 70%
- 内存使用率 < 80%
- 数据库连接池使用率 < 80%

### 告警规则

```yaml
# 高错误率告警
- alert: HighAPIErrorRate
  expr: error_rate > 0.05
  for: 5m
  severity: critical

# 高延迟告警
- alert: HighAPILatency
  expr: p95_latency > 1s
  for: 5m
  severity: warning
```

### 问题排查

**工具**:
- 日志: `docker logs <container>`
- 指标: Grafana 仪表板
- 追踪: Jaeger 分布式追踪
- 抓包: tcpdump, Wireshark

**流程**:
1. 查看监控告警
2. 分析日志和指标
3. 定位问题服务
4. 查看分布式追踪
5. 复现问题
6. 修复验证

---

## 五、扩展性设计

### 水平扩展

**应用层**:
- Kubernetes HPA 自动扩展
- 最小 5 个副本，最大 50 个副本
- CPU 70% 触发扩展

**数据库**:
- 主从复制 + 读写分离
- 分库分表（按 user_id 分 16 个库）
- 时间范围分表

**缓存**:
- Redis Cluster（3 主 3 从）
- 16384 个 hash slot
- 自动故障转移

### 容量规划

| 指标 | 当前 | 目标 | 扩展方案 |
|------|------|------|----------|
| 用户数 | 1000 | 100万 | 水平扩展 |
| 应用服务器 | 2 台 | 50 台 | K8s HPA |
| 数据库 | 1 主 2 从 | 16 主 32 从 | 分库分表 |
| Redis | 1 主 1 从 | 3 主 3 从 | 集群模式 |

---

## 六、项目亮点

### 技术亮点
1. ✅ **真实 AI 推理**: 集成 Whisper、HuBERT 等真实模型
2. ✅ **SFU 架构**: 低延迟、高性能媒体转发
3. ✅ **微服务架构**: 易于扩展和维护
4. ✅ **完善监控**: Prometheus + Grafana + Jaeger
5. ✅ **容器化部署**: Docker + Kubernetes

### 性能指标
- 支持 1000+ 并发用户
- API P95 延迟 < 100ms
- AI 推理吞吐量 > 50 req/s
- 系统可用性 99.9%

### 测试覆盖
- 单元测试
- 集成测试
- E2E 测试
- 压力测试

---

## 七、常见追问

### Q: 如果 Redis 挂了怎么办？
**答**: 
- Redis 主从复制，自动故障转移
- 应用层降级：直接查数据库
- 限流保护：防止数据库被打垮

### Q: 如何保证消息不丢失？
**答**:
- 消息持久化到 Redis
- 消息确认机制（ACK）
- 超时重试（最多 3 次）
- 死信队列（DLQ）

### Q: 如何处理热点数据？
**答**:
- 本地缓存（sync.Map）
- Redis 缓存
- 缓存预热
- 限流保护

### Q: 如何保证数据一致性？
**答**:
- PostgreSQL 作为 Source of Truth
- 使用消息队列实现最终一致性
- 定期同步任务修复不一致
- 分布式事务（Saga 模式）

---

## 八、面试技巧

### 回答结构
1. **背景**: 简要说明问题背景
2. **方案**: 详细描述解决方案
3. **效果**: 量化说明优化效果
4. **反思**: 总结经验教训

### 加分项
- 画图说明架构
- 展示代码片段
- 量化性能指标
- 分享踩坑经验
- 提出改进建议

### 注意事项
- 不要夸大自己的贡献
- 承认不足和技术债务
- 展示学习能力和成长
- 保持谦虚和诚实

---

## 九、快速记忆卡片

### 技术栈
- **后端**: Go + Gin + GORM + gRPC
- **前端**: Qt6 + React + WebRTC
- **数据库**: PostgreSQL + MongoDB + Redis + MinIO
- **AI**: Edge-LLM-Infra + PyTorch + Whisper
- **运维**: Docker + K8s + Nginx + Prometheus

### 端口映射
- 8080: user-service
- 8081: signaling-service
- 8082: meeting-service
- 8083: media-service
- 8085: ai-inference-service
- 10001: edge-model-infra
- 5010: ai-inference-worker

### 关键指标
- 并发用户: 1000+
- API 延迟: < 100ms (P95)
- 错误率: < 1%
- 可用性: 99.9%
- AI 推理: < 5s

---

**准备建议**:
1. 熟读本文档 2-3 遍
2. 准备好架构图（手绘或电子版）
3. 准备 2-3 个技术难点案例
4. 准备 1-2 个性能优化案例
5. 准备项目 Demo 演示（可选）

**面试前 10 分钟**:
- 深呼吸，放松心态
- 快速浏览本文档
- 回顾项目架构图
- 准备好纸笔（画图用）

**祝面试顺利！** 🎉


