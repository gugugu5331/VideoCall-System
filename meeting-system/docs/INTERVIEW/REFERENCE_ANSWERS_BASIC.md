# 基础知识题标准参考答案 (Q1-Q5)

**文档说明**: 本文档提供技术面试题中"基础知识题 (20%)"部分的标准参考答案，采用口语化表达，模拟真实面试场景中优秀候选人的回答方式。

**评分标准**: 所有答案均达到"优秀"等级（9-10分）标准，体现深入的技术理解和实战经验。

---

## Q1: Go 语言中的 goroutine 和 channel，以及它们在本项目中的应用场景

### 标准答案（口语化版本）

**核心概括（30秒电梯演讲）**:

我的理解是，goroutine 是 Go 语言的轻量级线程，它比操作系统线程轻得多，初始栈只有 2KB，可以轻松创建成千上万个。而 channel 是 goroutine 之间通信的管道，遵循"不要通过共享内存来通信，而要通过通信来共享内存"的设计哲学。在我们的会议系统项目中，这两个特性被大量使用在 WebSocket 消息处理、SFU 媒体转发、以及 AI 任务队列等场景。

---

**技术细节深入展开**:

首先说 goroutine，它的核心优势在于调度机制。Go 运行时使用 M:N 调度模型，也就是 M 个 goroutine 映射到 N 个操作系统线程上。这意味着即使我创建了 10000 个 goroutine，实际上可能只用了几十个系统线程。这在我们的信令服务中特别有用，因为每个 WebSocket 连接都需要两个 goroutine：一个负责读消息，一个负责写消息。

举个具体例子，在 `signaling-service/handlers/websocket_handler.go` 中，当一个新客户端连接进来时，我们会这样处理：

```go
func (h *WebSocketHandler) handleClient(client *Client) {
    // 启动两个 goroutine
    go h.handleClientMessages(client)  // 读取客户端消息
    go h.handleClientWrites(client)    // 发送消息给客户端
}
```

这里的关键是，`handleClientMessages` 会阻塞在 `conn.ReadMessage()` 上等待客户端发送消息，而 `handleClientWrites` 会阻塞在 `client.send` channel 上等待要发送的消息。如果用传统的线程模型，1000 个连接就需要 2000 个线程，这会消耗大量内存（每个线程至少 1MB 栈空间）。但用 goroutine，2000 个 goroutine 可能只占用几十 MB 内存。

---

再说 channel，它解决了并发编程中最头疼的问题：数据竞争。在我们的项目中，channel 主要有三种用法：

**用法1: 消息传递**（最常见）

在 WebSocket 处理中，我们用 buffered channel 来解耦消息的接收和发送：

```go
type Client struct {
    send chan []byte  // 缓冲 256 条消息
}

// 发送消息时，不直接写 WebSocket，而是写 channel
client.send <- messageBytes

// 另一个 goroutine 从 channel 读取并发送
for message := range client.send {
    conn.WriteMessage(websocket.TextMessage, message)
}
```

这样做的好处是，即使客户端网络慢，也不会阻塞消息的接收。如果 channel 满了，我们可以选择丢弃消息或者关闭连接，避免慢客户端拖垮整个系统。

**用法2: 任务队列**

在 AI 推理服务中，我们用 channel 实现了一个简单但高效的任务队列：

```go
processingQueue := make(chan *ProcessingTask, 1000)

// 启动 8 个 worker goroutine
for i := 0; i < 8; i++ {
    go func() {
        for task := range processingQueue {
            processAITask(task)
        }
    }()
}

// 提交任务
processingQueue <- &ProcessingTask{...}
```

这比用 Redis 队列简单多了，而且性能更好，因为没有网络开销。当然，缺点是任务不持久化，服务重启会丢失。

**用法3: 信号通知**

在 SFU 媒体转发中，我们用 channel 来通知 goroutine 退出：

```go
stopCh := make(chan struct{})

go func() {
    for {
        select {
        case <-stopCh:
            return  // 收到停止信号，退出
        default:
            // 继续转发 RTP 包
            forwardRTP()
        }
    }
}()

// 需要停止时
close(stopCh)  // 关闭 channel 会通知所有监听者
```

---

**项目中的实际应用总结**:

1. **WebSocket 消息处理**（signaling-service）: 每个连接 2 个 goroutine + 1 个 buffered channel
2. **SFU 媒体转发**（media-service）: 每个 RTP 流 1 个 goroutine，用 channel 控制生命周期
3. **AI 任务队列**（ai-inference-service）: 8 个 worker goroutine + 1 个容量 1000 的 channel
4. **心跳保活**（所有服务）: 定时器 + channel 实现优雅退出

我觉得 Go 的并发模型最大的优势是**简单**。不需要像 Java 那样处理复杂的线程池、锁、条件变量，大部分场景用 goroutine + channel 就能优雅解决。当然，也要注意 goroutine 泄漏的问题，比如忘记关闭 channel 或者 goroutine 里有死循环，这在我们项目的早期版本中就出现过。

---

## Q2: GORM 事务处理机制，以及在会议创建流程中为什么需要使用事务

### 标准答案（口语化版本）

**核心概括（30秒电梯演讲）**:

GORM 的事务机制其实就是对数据库 ACID 特性的封装。简单来说，事务保证一组数据库操作要么全部成功，要么全部失败，不会出现中间状态。在我们的会议创建流程中，需要同时创建会议记录和参与者记录，如果只创建了会议但没创建参与者，就会出现"孤儿会议"，这是不可接受的。所以必须用事务来保证数据一致性。

---

**技术细节深入展开**:

GORM 提供了两种事务使用方式，我更推荐第一种自动事务管理：

```go
// 方式1: 自动事务管理（推荐）
err := db.Transaction(func(tx *gorm.DB) error {
    // 操作1
    if err := tx.Create(&meeting).Error; err != nil {
        return err  // 自动回滚
    }

    // 操作2
    if err := tx.Create(&participant).Error; err != nil {
        return err  // 自动回滚
    }

    return nil  // 提交事务
})

// 方式2: 手动事务管理（更灵活但容易出错）
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&meeting).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&participant).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

第一种方式的好处是，如果回调函数返回 error，GORM 会自动回滚；如果返回 nil，会自动提交。这样就不用担心忘记调用 `Rollback()` 或 `Commit()`。

---

**在会议创建流程中的实际应用**:

在 `meeting-service/services/meeting_service.go` 中，我们的 `CreateMeeting` 方法涉及多个数据库操作：

```go
func (s *MeetingService) CreateMeeting(req *CreateMeetingRequest) (*models.Meeting, error) {
    var meeting *models.Meeting

    err := s.db.Transaction(func(tx *gorm.DB) error {
        // 步骤1: 创建会议记录
        meeting = &models.Meeting{
            Title:       req.Title,
            CreatorID:   req.CreatorID,
            StartTime:   req.StartTime,
            Duration:    req.Duration,
            Status:      models.MeetingStatusScheduled,
            MeetingCode: generateMeetingCode(),
        }

        if err := tx.Create(meeting).Error; err != nil {
            return fmt.Errorf("failed to create meeting: %w", err)
        }

        // 步骤2: 添加创建者为主持人
        participant := &models.MeetingParticipant{
            MeetingID: meeting.ID,
            UserID:    req.CreatorID,
            Role:      models.ParticipantRoleHost,
            Status:    models.ParticipantStatusAccepted,
        }

        if err := tx.Create(participant).Error; err != nil {
            return fmt.Errorf("failed to create participant: %w", err)
        }

        // 步骤3: 如果有邀请用户，批量创建参与者记录
        if len(req.InvitedUserIDs) > 0 {
            participants := make([]*models.MeetingParticipant, len(req.InvitedUserIDs))
            for i, userID := range req.InvitedUserIDs {
                participants[i] = &models.MeetingParticipant{
                    MeetingID: meeting.ID,
                    UserID:    userID,
                    Role:      models.ParticipantRoleAttendee,
                    Status:    models.ParticipantStatusPending,
                }
            }

            if err := tx.Create(&participants).Error; err != nil {
                return fmt.Errorf("failed to create invited participants: %w", err)
            }
        }

        return nil  // 所有操作成功，提交事务
    })

    if err != nil {
        return nil, err
    }

    return meeting, nil
}
```

---

**为什么必须使用事务**:

让我举个反例，如果不用事务会发生什么：

```go
// 错误示例：不使用事务
func (s *MeetingService) CreateMeetingWrong(req *CreateMeetingRequest) (*models.Meeting, error) {
    // 步骤1: 创建会议
    meeting := &models.Meeting{...}
    s.db.Create(meeting)  // 成功，meeting.ID = 123

    // 步骤2: 创建参与者
    participant := &models.MeetingParticipant{
        MeetingID: meeting.ID,
        UserID:    req.CreatorID,
    }

    // 假设这里数据库连接断了，或者违反了唯一约束
    err := s.db.Create(participant).Error
    if err != nil {
        // 糟糕！会议已经创建了，但参与者创建失败
        // 现在数据库里有一个没有主持人的会议
        return nil, err
    }

    return meeting, nil
}
```

这会导致几个严重问题：

1. **数据不一致**: 会议存在但没有主持人，违反业务规则
2. **孤儿数据**: 如果后续操作失败，前面的数据无法回滚
3. **并发问题**: 如果两个请求同时创建会议，可能出现竞态条件

---

**事务的 ACID 特性在项目中的体现**:

- **原子性 (Atomicity)**: 会议创建和参与者创建是一个原子操作，要么都成功，要么都失败
- **一致性 (Consistency)**: 保证数据库从一个一致状态转换到另一个一致状态（每个会议必须有主持人）
- **隔离性 (Isolation)**: 并发创建会议时，事务之间互不干扰（通过数据库锁实现）
- **持久性 (Durability)**: 事务提交后，数据永久保存，即使系统崩溃也不会丢失

---

**性能考虑**:

事务虽然保证了数据一致性，但也有性能开销：

1. **锁开销**: 事务期间会持有行锁或表锁，影响并发性能
2. **日志开销**: 数据库需要写 WAL（Write-Ahead Log）
3. **回滚开销**: 如果事务失败，需要回滚所有操作

所以在实际项目中，我们会注意：

- **事务尽量短**: 不要在事务中执行耗时操作（如调用外部 API）
- **减少锁冲突**: 使用乐观锁而不是悲观锁
- **批量操作**: 多个插入操作合并成一个批量插入

```go
// 优化：批量插入参与者
tx.Create(&participants)  // 一次插入多条，而不是循环插入
```

总的来说，事务是保证数据一致性的核心机制，在涉及多表操作的场景中必不可少。我们项目中除了会议创建，还在用户注册、会议结束（更新会议状态 + 生成统计数据）等场景使用了事务。

---

## Q3: PostgreSQL、Redis、MongoDB 三种数据库的特点和在项目中的应用场景

### 标准答案（口语化版本）

**核心概括（30秒电梯演讲）**:

我们项目采用了多数据库架构，每种数据库都有其最适合的场景。PostgreSQL 是关系型数据库，用于存储结构化的核心业务数据，比如用户、会议、参与者，因为这些数据之间有复杂的关联关系，需要 ACID 事务保证。Redis 是内存数据库，用于缓存热点数据、消息队列、会话管理，追求极致的读写性能。MongoDB 是文档型数据库，用于存储 AI 分析结果和聊天记录，因为这些数据的 schema 比较灵活，不适合用固定的表结构。

---

**技术细节深入展开**:

**1. PostgreSQL - 关系型数据库的选择**

我们选择 PostgreSQL 而不是 MySQL，主要有几个原因：

首先是**更强的 ACID 保证**。PostgreSQL 使用 MVCC（多版本并发控制），在高并发场景下性能更好。比如在会议创建时，多个用户同时创建会议，PostgreSQL 可以通过 MVCC 避免锁冲突。

其次是**更丰富的数据类型**。PostgreSQL 支持 JSON、数组、范围类型等，这在我们存储会议配置时很有用：

```sql
CREATE TABLE meetings (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    creator_id INTEGER REFERENCES users(id),
    start_time TIMESTAMP NOT NULL,
    duration INTEGER NOT NULL,
    settings JSONB,  -- 存储会议配置（允许录制、允许屏幕共享等）
    participants INTEGER[],  -- 参与者 ID 数组（冗余字段，加速查询）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 可以直接查询 JSON 字段
SELECT * FROM meetings WHERE settings->>'allow_recording' = 'true';

-- 可以使用数组操作
SELECT * FROM meetings WHERE 123 = ANY(participants);
```

第三是**强大的索引支持**。PostgreSQL 支持 B-Tree、Hash、GiST、GIN 等多种索引类型。我们在会议查询中用到了复合索引：

```sql
-- 创建复合索引，加速"查询某用户创建的进行中的会议"
CREATE INDEX idx_meetings_creator_status ON meetings(creator_id, status);

-- 创建 GIN 索引，加速 JSONB 查询
CREATE INDEX idx_meetings_settings ON meetings USING GIN(settings);
```

---

**在项目中的应用**:

PostgreSQL 存储所有核心业务数据：

```go
// user-service: 用户表
type User struct {
    ID           uint      `gorm:"primaryKey"`
    Username     string    `gorm:"uniqueIndex;size:50"`
    Email        string    `gorm:"uniqueIndex;size:100"`
    PasswordHash string    `gorm:"size:255"`
    FullName     string    `gorm:"size:100"`
    Status       string    `gorm:"size:20;default:'active'"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// meeting-service: 会议表
type Meeting struct {
    ID          uint      `gorm:"primaryKey"`
    Title       string    `gorm:"size:255"`
    CreatorID   uint      `gorm:"index"`
    StartTime   time.Time `gorm:"index"`
    Duration    int
    Status      string    `gorm:"index;size:20"`
    MeetingCode string    `gorm:"uniqueIndex;size:20"`
    Settings    datatypes.JSON  // PostgreSQL JSONB
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// 关联查询
type MeetingParticipant struct {
    ID        uint `gorm:"primaryKey"`
    MeetingID uint `gorm:"index"`
    UserID    uint `gorm:"index"`
    Role      string
    Status    string

    // 关联关系
    Meeting Meeting `gorm:"foreignKey:MeetingID"`
    User    User    `gorm:"foreignKey:UserID"`
}
```

---

**2. Redis - 内存数据库的三种用法**

Redis 在我们项目中有三种典型应用：

**用法1: 缓存热点数据**

```go
// 缓存会议信息，减轻数据库压力
func (s *MeetingService) GetMeeting(meetingID uint) (*models.Meeting, error) {
    cacheKey := fmt.Sprintf("meeting:%d", meetingID)

    // 1. 先查 Redis 缓存
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var meeting models.Meeting
        json.Unmarshal([]byte(cached), &meeting)
        return &meeting, nil  // 缓存命中，直接返回
    }

    // 2. 缓存未命中，查数据库
    var meeting models.Meeting
    if err := s.db.First(&meeting, meetingID).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存，过期时间 10 分钟
    meetingJSON, _ := json.Marshal(meeting)
    s.redis.Set(ctx, cacheKey, meetingJSON, 10*time.Minute)

    return &meeting, nil
}
```

这里有个经典问题：**缓存穿透、缓存击穿、缓存雪崩**。我们的解决方案：

- **缓存穿透**（查询不存在的数据）: 使用布隆过滤器或缓存空值
- **缓存击穿**（热点数据过期）: 使用互斥锁或永不过期
- **缓存雪崩**（大量缓存同时过期）: 设置随机过期时间

```go
// 防止缓存雪崩：随机过期时间
expireTime := 10*time.Minute + time.Duration(rand.Intn(60))*time.Second
s.redis.Set(ctx, cacheKey, meetingJSON, expireTime)
```

**用法2: 消息队列**

```go
// 生产者：提交 AI 任务到 Redis 队列
func (s *AIService) SubmitTask(task *AITask) error {
    taskJSON, _ := json.Marshal(task)
    return s.redis.LPush(ctx, "ai_tasks", taskJSON).Err()
}

// 消费者：从 Redis 队列获取任务
func (w *Worker) ProcessTasks() {
    for {
        // BRPOP 阻塞式弹出，超时时间 0 表示永久阻塞
        result, err := w.redis.BRPop(ctx, 0, "ai_tasks").Result()
        if err != nil {
            continue
        }

        var task AITask
        json.Unmarshal([]byte(result[1]), &task)
        w.processTask(&task)
    }
}
```

**用法3: 会话管理**

```go
// 存储用户会话（JWT token）
func (s *UserService) SaveSession(userID uint, token string) error {
    sessionKey := fmt.Sprintf("session:%d", userID)
    return s.redis.Set(ctx, sessionKey, token, 24*time.Hour).Err()
}

// 验证会话
func (s *UserService) ValidateSession(userID uint, token string) (bool, error) {
    sessionKey := fmt.Sprintf("session:%d", userID)
    storedToken, err := s.redis.Get(ctx, sessionKey).Result()
    if err != nil {
        return false, err
    }
    return storedToken == token, nil
}
```

---

**3. MongoDB - 文档型数据库的灵活性**

MongoDB 最大的优势是 **schema-less**，适合存储结构不固定的数据。

**在项目中的应用**:

```go
// AI 分析结果（每种分析类型的结果结构都不同）
{
    "_id": ObjectId("..."),
    "meeting_id": "123",
    "user_id": "456",
    "analysis_type": "emotion_detection",
    "timestamp": ISODate("2025-10-09T10:00:00Z"),
    "results": {
        // 情绪检测结果
        "emotion": "happy",
        "confidence": 0.95,
        "emotions": {
            "happy": 0.95,
            "neutral": 0.03,
            "sad": 0.02
        }
    }
}

{
    "_id": ObjectId("..."),
    "meeting_id": "123",
    "user_id": "456",
    "analysis_type": "speech_recognition",
    "timestamp": ISODate("2025-10-09T10:00:05Z"),
    "results": {
        // 语音识别结果
        "text": "大家好，今天我们讨论项目进度",
        "language": "zh-CN",
        "confidence": 0.98,
        "words": [
            {"word": "大家好", "start": 0.0, "end": 0.5},
            {"word": "今天", "start": 0.6, "end": 0.9}
        ]
    }
}
```

如果用 PostgreSQL 存储这些数据，要么创建很多表（emotion_results, speech_results），要么用 JSONB 字段（失去类型检查）。而 MongoDB 可以灵活存储不同结构的文档。

---

**三种数据库对比总结**:

| 特性 | PostgreSQL | Redis | MongoDB |
|------|-----------|-------|---------|
| **类型** | 关系型 | 键值型 | 文档型 |
| **存储** | 磁盘 | 内存 | 磁盘 |
| **事务** | 完整 ACID | 有限支持 | 有限支持 |
| **查询** | SQL，复杂查询 | 简单 KV 查询 | 类 SQL 查询 |
| **性能** | 中等（磁盘 I/O） | 极高（内存） | 高（索引优化） |
| **扩展** | 垂直扩展为主 | 集群模式 | 分片集群 |
| **适用场景** | 核心业务数据 | 缓存、队列、会话 | 非结构化数据 |

在实际项目中，我们遵循"**用合适的工具做合适的事**"的原则，而不是用一种数据库解决所有问题。这种多数据库架构虽然增加了复杂度，但能充分发挥每种数据库的优势。

---

## Q4: Redis 的三种应用模式（缓存、消息队列、发布订阅）及其区别

### 标准答案（口语化版本）

**核心概括（30秒电梯演讲）**:

Redis 在我们项目中扮演了三种不同的角色。第一种是**缓存**，用来加速热点数据的读取，比如会议信息、用户信息，这是最常见的用法。第二种是**消息队列**，用来异步处理 AI 任务，保证任务不丢失，支持多个消费者。第三种是**发布订阅**，用来实时广播事件，比如会议开始、用户加入，所有订阅者都能立即收到通知。这三种模式的核心区别在于：缓存是主动读取，消息队列是持久化的一对一消费，发布订阅是非持久化的一对多广播。

---

**技术细节深入展开**:

**模式1: 缓存（Cache）**

缓存是 Redis 最经典的应用，核心思想是"**用空间换时间**"。我们在项目中大量使用缓存来减轻数据库压力。

**实际代码示例**:

```go
// meeting-service/services/meeting_service.go

func (s *MeetingService) GetMeeting(meetingID uint) (*models.Meeting, error) {
    cacheKey := fmt.Sprintf("meeting:%d", meetingID)

    // 1. 尝试从缓存读取
    cached, err := s.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        // 缓存命中
        var meeting models.Meeting
        json.Unmarshal([]byte(cached), &meeting)
        logger.Debug(fmt.Sprintf("Cache hit for meeting %d", meetingID))
        return &meeting, nil
    }

    // 2. 缓存未命中，查询数据库
    var meeting models.Meeting
    if err := s.db.First(&meeting, meetingID).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存
    meetingJSON, _ := json.Marshal(meeting)
    s.redis.Set(ctx, cacheKey, meetingJSON, 10*time.Minute)
    logger.Debug(fmt.Sprintf("Cache miss for meeting %d, loaded from DB", meetingID))

    return &meeting, nil
}
```

**缓存的关键问题**:

1. **缓存穿透**（查询不存在的数据）:
```go
// 解决方案：缓存空值
if err := s.db.First(&meeting, meetingID).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        // 缓存空值，防止穿透
        s.redis.Set(ctx, cacheKey, "null", 5*time.Minute)
    }
    return nil, err
}
```

2. **缓存击穿**（热点数据过期，大量请求打到数据库）:
```go
// 解决方案：使用互斥锁
lockKey := fmt.Sprintf("lock:meeting:%d", meetingID)
if s.redis.SetNX(ctx, lockKey, "1", 10*time.Second).Val() {
    // 获取锁成功，查询数据库
    defer s.redis.Del(ctx, lockKey)
    // ... 查询数据库并缓存
} else {
    // 获取锁失败，等待其他线程加载缓存
    time.Sleep(100 * time.Millisecond)
    return s.GetMeeting(meetingID)  // 重试
}
```

3. **缓存雪崩**（大量缓存同时过期）:
```go
// 解决方案：随机过期时间
baseExpire := 10 * time.Minute
randomExpire := time.Duration(rand.Intn(60)) * time.Second
s.redis.Set(ctx, cacheKey, meetingJSON, baseExpire+randomExpire)
```

---

**模式2: 消息队列（Message Queue）**

消息队列用于异步任务处理，保证任务不丢失。我们在 AI 推理服务中大量使用。

**实际代码示例**:

```go
// shared/queue/redis_message_queue.go

// 生产者：提交 AI 任务
func (s *AIService) SubmitSpeechRecognitionTask(audioData []byte, meetingID, userID uint) error {
    task := &AITask{
        TaskID:    uuid.NewString(),
        TaskType:  "speech_recognition",
        MeetingID: meetingID,
        UserID:    userID,
        AudioData: base64.StdEncoding.EncodeToString(audioData),
        CreatedAt: time.Now(),
    }

    taskJSON, _ := json.Marshal(task)

    // 使用 LPUSH 将任务推入队列
    if err := s.redis.LPush(ctx, "ai_tasks", taskJSON).Err(); err != nil {
        return fmt.Errorf("failed to submit task: %w", err)
    }

    logger.Info(fmt.Sprintf("Task submitted: %s", task.TaskID))
    return nil
}

// 消费者：处理 AI 任务
func (w *Worker) ProcessTasks() {
    for {
        // BRPOP 阻塞式弹出，超时时间 0 表示永久阻塞
        result, err := w.redis.BRPop(ctx, 0, "ai_tasks").Result()
        if err != nil {
            logger.Error(fmt.Sprintf("Failed to pop task: %v", err))
            time.Sleep(1 * time.Second)
            continue
        }

        // result[0] 是队列名，result[1] 是任务数据
        var task AITask
        if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
            logger.Error(fmt.Sprintf("Failed to unmarshal task: %v", err))
            continue
        }

        logger.Info(fmt.Sprintf("Processing task: %s", task.TaskID))

        // 处理任务
        if err := w.processTask(&task); err != nil {
            logger.Error(fmt.Sprintf("Task failed: %s, error: %v", task.TaskID, err))

            // 重试机制：将任务重新放回队列
            task.RetryCount++
            if task.RetryCount < 3 {
                taskJSON, _ := json.Marshal(task)
                w.redis.LPush(ctx, "ai_tasks", taskJSON)
            } else {
                // 超过重试次数，放入死信队列
                w.redis.LPush(ctx, "ai_tasks_dead_letter", result[1])
            }
        } else {
            logger.Info(fmt.Sprintf("Task completed: %s", task.TaskID))
        }
    }
}
```

**消息队列的特点**:

1. **持久化**: 任务存储在 Redis 中，即使消费者宕机，任务也不会丢失
2. **FIFO**: 先进先出，保证任务按顺序处理
3. **一对一消费**: 每个任务只会被一个消费者处理
4. **支持重试**: 失败的任务可以重新放回队列

---

**模式3: 发布订阅（Pub/Sub）**

发布订阅用于实时事件广播，所有订阅者都能收到消息。

**实际代码示例**:

```go
// shared/queue/redis_pubsub.go

// 发布者：会议服务发布事件
func (s *MeetingService) StartMeeting(meetingID uint) error {
    // 1. 更新数据库
    if err := s.db.Model(&models.Meeting{}).Where("id = ?", meetingID).Update("status", "ongoing").Error; err != nil {
        return err
    }

    // 2. 发布事件到 Redis Pub/Sub
    event := &Event{
        Type:      "meeting.started",
        MeetingID: meetingID,
        Timestamp: time.Now().Unix(),
        Data: map[string]interface{}{
            "meeting_id": meetingID,
            "start_time": time.Now(),
        },
    }

    eventJSON, _ := json.Marshal(event)
    if err := s.redis.Publish(ctx, "meeting_events", eventJSON).Err(); err != nil {
        logger.Error(fmt.Sprintf("Failed to publish event: %v", err))
    }

    logger.Info(fmt.Sprintf("Meeting started event published: %d", meetingID))
    return nil
}
```

**三种模式对比总结**:

| 特性 | 缓存 | 消息队列 | 发布订阅 |
|------|------|----------|----------|
| **数据结构** | String/Hash | List | Pub/Sub Channel |
| **持久化** | 有过期时间 | 持久化（AOF） | 不持久化 |
| **读取方式** | 主动 GET | 阻塞 BRPOP | 被动推送 |
| **消费模式** | 多次读取 | 一次消费 | 多次广播 |
| **消息丢失** | 过期删除 | 不丢失（除非 Redis 宕机） | 订阅者离线会丢失 |
| **适用场景** | 热点数据加速 | 异步任务处理 | 实时事件通知 |
| **项目应用** | 会议信息、用户信息 | AI 任务队列 | 会议事件、在线状态 |

在我们项目中，这三种模式经常组合使用。比如会议开始时：
- 更新数据库状态
- 删除缓存（Cache Invalidation）
- 发布事件到 Pub/Sub（通知所有订阅者）
- 提交 AI 任务到消息队列（异步处理）

这样既保证了数据一致性，又实现了实时通知和异步处理。

---



## Q5: Docker Compose 中的 healthcheck、depends_on、networks 配置及其作用

### 标准答案（口语化版本）

**核心概括（30秒电梯演讲）**:

Docker Compose 的这三个配置是容器编排的核心。`healthcheck` 用来检查容器是否真正可用，而不仅仅是启动了；`depends_on` 定义服务启动顺序，确保依赖服务先启动；`networks` 隔离容器网络，让服务之间可以通过服务名通信。在我们的会议系统中，这三个配置保证了服务按正确顺序启动、健康检查通过后才接受流量、以及服务间的网络隔离和通信。

---

**技术细节深入展开**:

**配置1: healthcheck（健康检查）**

很多人以为容器启动了就代表服务可用，其实不是。比如 PostgreSQL 容器启动了，但数据库初始化可能还需要 10 秒。如果这时候 user-service 连接数据库，就会失败。

**实际配置示例**:

```yaml
# docker-compose.yml

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres123
      POSTGRES_DB: meeting_system
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s      # 每 10 秒检查一次
      timeout: 5s        # 单次检查超时时间
      retries: 5         # 失败 5 次后标记为 unhealthy
      start_period: 30s  # 启动后 30 秒内失败不计入 retries
    ports:
      - "5432:5432"
```

**健康检查的工作原理**:

1. 容器启动后，等待 `start_period`（30秒）
2. 每隔 `interval`（10秒）执行一次 `test` 命令
3. 如果命令在 `timeout`（5秒）内返回 0，标记为 healthy
4. 如果连续失败 `retries`（5次），标记为 unhealthy

**查看健康状态**:

```bash
# 查看容器健康状态
docker ps
# CONTAINER ID   IMAGE              STATUS
# abc123         postgres:15        Up 2 minutes (healthy)

# 查看健康检查日志
docker inspect postgres | jq '.[0].State.Health'
```

**不同服务的健康检查命令**:

```yaml
# PostgreSQL
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U postgres"]

# Redis
healthcheck:
  test: ["CMD", "redis-cli", "ping"]

# HTTP 服务
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]

# gRPC 服务
healthcheck:
  test: ["CMD", "grpc_health_probe", "-addr=:9090"]
```

**在项目中的应用**:

我们的每个微服务都提供了 `/health` 端点：

```go
// user-service/main.go

func healthCheckHandler(c *gin.Context) {
    // 检查数据库连接
    if err := db.DB().Ping(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error":  "database connection failed",
        })
        return
    }

    // 检查 Redis 连接
    if err := redisClient.Ping(ctx).Err(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error":  "redis connection failed",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "timestamp": time.Now(),
    })
}

r.GET("/health", healthCheckHandler)
```

---

**配置2: depends_on（依赖关系）**

`depends_on` 定义服务启动顺序，但**只保证启动顺序，不保证服务可用**。这是很多人容易误解的地方。

**基础用法**:

```yaml
services:
  user-service:
    build: ./backend/user-service
    depends_on:
      - postgres
      - redis
      - etcd
    # user-service 会在 postgres、redis、etcd 启动后再启动
```

**进阶用法（配合 healthcheck）**:

```yaml
services:
  postgres:
    image: postgres:15-alpine
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  user-service:
    build: ./backend/user-service
    depends_on:
      postgres:
        condition: service_healthy  # 等待 postgres 健康检查通过
      redis:
        condition: service_healthy
      etcd:
        condition: service_started  # 只等待启动，不等健康检查
```

**condition 的三种取值**:

- `service_started`: 容器启动即可（默认）
- `service_healthy`: 等待健康检查通过
- `service_completed_successfully`: 等待容器成功退出（用于初始化容器）

**为什么需要 depends_on？**

如果不设置依赖关系，所有服务会并行启动，可能导致：

```
user-service 启动 → 连接 postgres → 失败（postgres 还没启动）
user-service 退出 → Docker Compose 认为启动失败
```

设置依赖后：

```
postgres 启动 → 健康检查通过 → user-service 启动 → 连接成功
```

**实际项目中的依赖关系**:

```yaml
services:
  # 基础设施层（最先启动）
  postgres:
    image: postgres:15-alpine
    healthcheck: ...

  redis:
    image: redis:7-alpine
    healthcheck: ...

  etcd:
    image: quay.io/coreos/etcd:v3.5.15
    healthcheck: ...

  # 核心服务层（依赖基础设施）
  user-service:
    depends_on:
      postgres: {condition: service_healthy}
      redis: {condition: service_healthy}
      etcd: {condition: service_healthy}

  meeting-service:
    depends_on:
      postgres: {condition: service_healthy}
      redis: {condition: service_healthy}
      etcd: {condition: service_healthy}

  # 媒体服务层（依赖核心服务）
  media-service:
    depends_on:
      postgres: {condition: service_healthy}
      redis: {condition: service_healthy}
      user-service: {condition: service_healthy}
      meeting-service: {condition: service_healthy}

  # API 网关（最后启动）
  nginx:
    depends_on:
      user-service: {condition: service_healthy}
      meeting-service: {condition: service_healthy}
      media-service: {condition: service_healthy}
```

---

**配置3: networks（网络配置）**

Docker Compose 默认会为每个项目创建一个网络，但我们可以自定义网络来实现更精细的控制。

**基础用法**:

```yaml
networks:
  meeting-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.25.0.0/16

services:
  user-service:
    networks:
      - meeting-network
    # 可以通过服务名访问其他服务
    # 例如: http://postgres:5432
```

**网络的核心功能**:

1. **服务发现（DNS）**: 容器可以通过服务名互相访问

```go
// user-service 连接 postgres
dsn := "host=postgres user=postgres password=postgres123 dbname=meeting_system"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// 而不是
dsn := "host=172.25.0.10 user=postgres ..."  // 硬编码 IP
```

2. **网络隔离**: 不同网络的容器无法互相访问

```yaml
networks:
  frontend-network:  # 前端网络
  backend-network:   # 后端网络
  database-network:  # 数据库网络

services:
  nginx:
    networks:
      - frontend-network
      - backend-network
    # nginx 可以访问前端和后端，但不能直接访问数据库

  user-service:
    networks:
      - backend-network
      - database-network
    # user-service 可以访问后端和数据库

  postgres:
    networks:
      - database-network
    # postgres 只能被后端服务访问，前端无法直接访问
```

3. **固定 IP（可选）**:

```yaml
services:
  postgres:
    networks:
      meeting-network:
        ipv4_address: 172.25.0.10
```

**实际项目中的网络配置**:

```yaml
# docker-compose.yml

networks:
  meeting-network:
    driver: bridge
    ipam:
      driver: default
      config:
        - subnet: 172.25.0.0/16
          gateway: 172.25.0.1

services:
  postgres:
    image: postgres:15-alpine
    container_name: meeting-postgres
    networks:
      - meeting-network
    # 其他服务可以通过 postgres:5432 访问

  user-service:
    build: ./backend/user-service
    container_name: meeting-user-service
    networks:
      - meeting-network
    environment:
      - DATABASE_HOST=postgres  # 使用服务名
      - REDIS_HOST=redis
      - ETCD_ENDPOINTS=etcd:2379

  nginx:
    image: nginx:alpine
    container_name: meeting-nginx
    networks:
      - meeting-network
    ports:
      - "8800:80"  # 只有 nginx 暴露到宿主机
```

**网络的优势**:

1. **简化配置**: 不需要硬编码 IP 地址
2. **动态扩展**: 新增服务自动加入网络
3. **安全隔离**: 数据库不直接暴露到宿主机
4. **负载均衡**: 同一服务的多个副本可以共享服务名

---

**三个配置的协同工作**:

在我们的项目中，这三个配置是这样协同工作的：

```yaml
services:
  postgres:
    image: postgres:15-alpine
    networks:
      - meeting-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    # 1. 启动 postgres
    # 2. 执行健康检查
    # 3. 健康检查通过后，标记为 healthy

  user-service:
    build: ./backend/user-service
    networks:
      - meeting-network
    depends_on:
      postgres:
        condition: service_healthy
    # 4. 等待 postgres 健康检查通过
    # 5. 启动 user-service
    # 6. 通过服务名 postgres:5432 连接数据库
```

**启动流程**:

```
1. docker-compose up -d
2. 创建 meeting-network 网络
3. 启动 postgres 容器，加入 meeting-network
4. 执行 postgres 健康检查（每 10 秒一次）
5. 健康检查通过后，启动 user-service
6. user-service 通过 DNS 解析 postgres 服务名
7. user-service 连接 postgres:5432 成功
```

**总结**:

这三个配置解决了容器编排的三个核心问题：
- **healthcheck**: 确保服务真正可用
- **depends_on**: 确保启动顺序正确
- **networks**: 确保服务间可以通信

在实际项目中，我们必须同时使用这三个配置，才能保证系统稳定启动。如果只用 `depends_on` 不用 `healthcheck`，可能会出现服务启动了但还没准备好的情况；如果不用 `networks`，服务间无法通过服务名通信，只能硬编码 IP。

---

**文档版本**: v1.0
**最后更新**: 2025-10-09
**维护者**: Meeting System Team
