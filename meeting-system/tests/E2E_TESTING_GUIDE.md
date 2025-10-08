# 端到端消息队列集成测试指南

## 概述

本指南介绍如何执行端到端（E2E）测试，验证消息队列系统在 meeting-system 各个微服务中的集成和运行状态。

## 测试场景

**主要测试场景**：三个用户注册、加入同一会议室并调用 AI 服务

**测试目标**：
1. 验证消息队列系统在所有服务中正确初始化
2. 验证任务处理器正确注册和执行
3. 验证事件在服务间正确流转
4. 验证 Redis 队列的消息发布和消费
5. 验证系统的端到端功能

## 前置条件

### 1. 服务运行状态

确保以下服务正在运行：

```bash
# 检查 Docker 容器状态
docker ps

# 应该看到以下容器：
# - user-service
# - meeting-service
# - media-service
# - signaling-service
# - ai-service
# - redis
# - nginx
```

### 2. 启动所有服务

如果服务未运行，使用以下命令启动：

```bash
cd meeting-system
docker-compose up -d
```

### 3. 验证 Redis 连接

```bash
redis-cli ping
# 应该返回: PONG
```

### 4. 验证 Nginx 配置

```bash
curl http://localhost/health
# 或者检查 Nginx 是否在监听 80 端口
netstat -tuln | grep :80
```

## 测试工具

提供了三个测试脚本：

### 1. Bash 测试脚本 (`e2e_queue_integration_test.sh`)

**特点**：
- 使用 curl 发送 HTTP 请求
- 轻量级，无需额外依赖
- 适合快速测试

**使用方法**：

```bash
cd meeting-system/tests
chmod +x e2e_queue_integration_test.sh
./e2e_queue_integration_test.sh
```

**输出**：
- `e2e_test_YYYYMMDD_HHMMSS.log` - 详细日志
- `e2e_test_report_YYYYMMDD_HHMMSS.md` - 测试报告

### 2. Python 测试脚本 (`e2e_queue_integration_test.py`)

**特点**：
- 更好的 JSON 处理
- 详细的测试报告
- 更友好的输出格式

**依赖安装**：

```bash
pip install requests redis
```

**使用方法**：

```bash
cd meeting-system/tests
chmod +x e2e_queue_integration_test.py
python3 e2e_queue_integration_test.py
```

**输出**：
- `e2e_test_YYYYMMDD_HHMMSS.log` - 详细日志
- `e2e_test_report_YYYYMMDD_HHMMSS.md` - 测试报告

### 3. 服务日志检查脚本 (`check_service_logs.sh`)

**特点**：
- 检查各服务的日志文件
- 验证队列系统初始化状态
- 统计任务和事件处理数量

**使用方法**：

```bash
cd meeting-system/tests
chmod +x check_service_logs.sh
./check_service_logs.sh
```

**输出**：
- `service_logs_check_YYYYMMDD_HHMMSS.md` - 日志检查报告

## 测试步骤详解

### 阶段 1: 用户注册

**操作**：
- 注册 3 个测试用户

**验证点**：
- HTTP 响应状态码为 200/201
- user-service 日志中有 "Processing user register task"
- Redis 队列中有消息发布记录
- `user_events` 频道发布了 `user.registered` 事件

**检查命令**：

```bash
# 检查 user-service 日志
tail -f backend/user-service/logs/service.log | grep "register"

# 检查 Redis 队列
redis-cli LLEN meeting_system:normal_queue
```

### 阶段 2: 用户登录

**操作**：
- 3 个用户分别登录
- 获取认证 token

**验证点**：
- 成功获取 JWT token
- user-service 日志中有 "User login" 记录
- `user_events` 频道发布了 `user.logged_in` 事件

### 阶段 3: 创建会议

**操作**：
- User1 创建一个会议室

**验证点**：
- 成功创建会议，获取 meeting_id
- meeting-service 日志中有 "Processing meeting create task"
- `meeting_events` 频道发布了 `meeting.created` 事件
- ai-service、media-service、signaling-service 接收到该事件

**检查命令**：

```bash
# 检查 meeting-service 日志
tail -f backend/meeting-service/logs/service.log | grep "meeting"

# 检查其他服务是否接收到事件
tail -f backend/ai-service/logs/service.log | grep "Received meeting event"
tail -f backend/media-service/logs/service.log | grep "Received meeting event"
tail -f backend/signaling-service/logs/service.log | grep "Received meeting event"
```

### 阶段 4: 用户加入会议

**操作**：
- User1、User2、User3 依次加入会议

**验证点**：
- 所有用户成功加入会议
- meeting-service 日志中有 "Processing meeting user join task"
- `meeting_events` 频道发布了 `meeting.user_joined` 事件
- media-service 和 signaling-service 接收到用户加入事件

**检查命令**：

```bash
# 检查用户加入事件
tail -f backend/meeting-service/logs/service.log | grep "user_join"

# 检查 media-service 响应
tail -f backend/media-service/logs/service.log | grep "User.*joined"
```

### 阶段 5: 调用 AI 服务

**操作**：
- User1: 语音识别
- User2: 情绪检测
- User3: 音频降噪

**验证点**：
- AI 任务成功提交
- ai-service 日志中有任务处理记录
- `ai_events` 频道发布了完成事件
- meeting-service 接收到 AI 处理完成事件

**检查命令**：

```bash
# 检查 AI 任务处理
tail -f backend/ai-service/logs/service.log | grep "Processing.*task"

# 检查 AI 事件发布
tail -f backend/ai-service/logs/service.log | grep "ai_events"

# 检查 meeting-service 接收 AI 事件
tail -f backend/meeting-service/logs/service.log | grep "Received AI event"
```

## 验证事件流转

### 完整的事件流转链路

```
Meeting Service → meeting_events → AI/Media/Signaling Services
AI Service → ai_events → Meeting/Media Services
Media Service → media_events → Meeting/Signaling Services
Signaling Service → signaling_events → Meeting/Media Services
User Service → user_events → Meeting Service
```

### 验证方法

**1. 使用 Redis Monitor**：

```bash
redis-cli monitor | grep "PUBLISH"
```

**2. 检查服务日志**：

```bash
# 检查所有服务的事件接收
for service in user-service meeting-service media-service signaling-service ai-service; do
    echo "=== $service ==="
    grep "Received.*event" backend/$service/logs/service.log | tail -5
done
```

**3. 检查队列统计**：

```bash
# 使用 Redis CLI
redis-cli
> LLEN meeting_system:critical_queue
> LLEN meeting_system:high_queue
> LLEN meeting_system:normal_queue
> LLEN meeting_system:low_queue
> LLEN meeting_system:dead_letter_queue
```

## 性能指标

### 关键指标

1. **消息发布延迟**：从发布到进入队列的时间
2. **消息处理延迟**：从队列取出到处理完成的时间
3. **端到端延迟**：从用户请求到事件完成的总时间
4. **吞吐量**：每秒处理的消息数
5. **队列长度**：各优先级队列的消息积压情况
6. **失败率**：进入死信队列的消息比例

### 监控命令

```bash
# 实时监控队列长度
watch -n 1 'redis-cli LLEN meeting_system:normal_queue'

# 监控死信队列
watch -n 1 'redis-cli LLEN meeting_system:dead_letter_queue'

# 查看队列中的消息（不移除）
redis-cli LRANGE meeting_system:normal_queue 0 -1
```

## 故障排查

### 常见问题

#### 1. 服务未初始化队列系统

**症状**：日志中没有 "Initializing message queue system"

**解决方法**：
- 检查 Redis 是否正常运行
- 检查配置文件中的 Redis 连接信息
- 重启服务

#### 2. 任务未被处理

**症状**：队列长度持续增长，但没有处理日志

**解决方法**：
- 检查任务处理器是否正确注册
- 检查工作协程是否启动
- 查看错误日志

#### 3. 事件未流转

**症状**：发布了事件但其他服务未接收

**解决方法**：
- 检查 Pub/Sub 订阅是否成功
- 检查频道名称是否正确
- 使用 `redis-cli monitor` 监控 Pub/Sub 消息

#### 4. 消息进入死信队列

**症状**：dead_letter_queue 长度增加

**解决方法**：
- 查看死信队列中的消息内容
- 检查任务处理器的错误日志
- 修复处理逻辑后重新发布消息

```bash
# 查看死信队列消息
redis-cli LRANGE meeting_system:dead_letter_queue 0 -1

# 移动死信消息回正常队列（谨慎操作）
redis-cli RPOPLPUSH meeting_system:dead_letter_queue meeting_system:normal_queue
```

## 清理测试数据

### 清理 Redis 队列

```bash
redis-cli FLUSHDB
```

### 清理测试用户

```bash
# 连接到数据库
mysql -u root -p meeting_system

# 删除测试用户
DELETE FROM users WHERE username LIKE 'test_user_%';
```

### 清理测试会议

```bash
# 删除测试会议
DELETE FROM meetings WHERE title LIKE 'E2E Test%';
```

## 测试报告示例

测试完成后，会生成类似以下的报告：

```markdown
# 端到端消息队列集成测试报告

**测试时间**: 2025-01-15 10:30:00
**测试时长**: 45.23 秒
**测试结果**: 15/15 成功

## 测试步骤和结果

### 1. 用户注册阶段
- ✅ test_user_1 (user1@test.com)
- ✅ test_user_2 (user2@test.com)
- ✅ test_user_3 (user3@test.com)

### 2. 用户登录阶段
- ✅ test_user_1
- ✅ test_user_2
- ✅ test_user_3

### 3. 创建会议室阶段
- ✅ 会议 ID: 123

### 4. 用户加入会议阶段
- ✅ test_user_1
- ✅ test_user_2
- ✅ test_user_3

### 5. 调用 AI 服务阶段
- ✅ test_user_1: speech_recognition
- ✅ test_user_2: emotion_detection
- ✅ test_user_3: audio_denoising

## Redis 队列统计

critical_queue: 0
high_queue: 0
normal_queue: 2
low_queue: 0
dead_letter_queue: 0

## 测试结论

- 总测试数: 15
- 成功: 15
- 失败: 0
- 成功率: 100.00%
```

## 总结

通过本指南，您可以：
1. 执行完整的端到端测试
2. 验证消息队列系统的集成状态
3. 监控系统性能和健康状况
4. 排查和解决常见问题

如有问题，请查看各服务的详细日志或联系开发团队。

