# 会议系统项目技术面试问题清单

**项目**: meeting-system-server (智能视频会议平台)  
**面试时间**: 预计 60-90 分钟  
**难度级别**: 由浅入深，循序渐进

---

## 第一部分：项目背景与架构理解 (15分钟)

### 1.1 项目概述

**Q1: 请用3-5分钟简要介绍一下这个会议系统项目，包括它的核心功能和你在项目中的角色。**

**考察目的**: 
- 候选人对项目的整体把握能力
- 表达和总结能力
- 在项目中的参与度

**优秀答案要点**:
- 项目是基于 Edge-LLM-Infra 构建的企业级视频会议系统
- 采用 SFU (Selective Forwarding Unit) 架构实现低延迟音视频通话
- 集成分布式 AI 推理服务（语音识别、情绪检测、合成检测）
- 微服务架构：5个核心业务服务 + AI推理层 + 基础设施层
- 技术栈：Go (后端) + WebRTC (音视频) + Edge-LLM-Infra (AI) + Docker (部署)
- 个人角色：明确说明负责的模块和贡献

---

### 1.2 架构设计

**Q2: 请画出或描述这个系统的整体架构图，包括各个服务之间的交互关系。**

**考察目的**:
- 系统架构设计能力
- 对分布式系统的理解
- 服务拆分的合理性

**优秀答案要点**:
```
客户端层 → 网关层(Nginx) → 微服务层 → AI推理层 → 数据层

微服务层包括:
- user-service (8080): 用户认证、权限管理
- meeting-service (8082): 会议管理、参与者管理
- signaling-service (8081): WebSocket信令、WebRTC协商
- media-service (8083): SFU媒体转发、录制
- ai-inference-service (8085): AI推理网关

AI推理层:
- edge-model-infra (10001): C++ 单元管理器
- ai-inference-worker (5010): Python 推理节点
- 支持 ASR、情绪检测、合成检测

数据层:
- PostgreSQL: 用户、会议等结构化数据
- MongoDB: AI分析结果、聊天记录
- Redis: 缓存、消息队列、会话管理
- MinIO: 录制文件、媒体存储
```

---

**Q3: 为什么选择微服务架构而不是单体架构？这个选择带来了哪些好处和挑战？**

**考察目的**:
- 架构决策能力
- 对微服务优缺点的理解
- 实际问题解决经验

**优秀答案要点**:

**好处**:
- 独立部署：各服务可独立升级，不影响其他服务
- 技术异构：Go服务 + C++ AI节点 + Python推理，各取所长
- 水平扩展：媒体服务和AI服务可根据负载独立扩展
- 故障隔离：单个服务故障不会导致整个系统崩溃
- 团队协作：不同团队可并行开发不同服务

**挑战**:
- 分布式事务：需要使用消息队列实现最终一致性
- 服务间通信：需要处理网络延迟、超时、重试
- 链路追踪：使用 Jaeger 实现分布式追踪
- 配置管理：使用 etcd 进行服务发现和配置中心
- 部署复杂度：使用 Docker Compose 和 Kubernetes 简化部署

---

### 1.3 技术选型

**Q4: 为什么选择 Go 语言作为后端开发语言？相比 Java 或 Node.js 有什么优势？**

**考察目的**:
- 技术选型的合理性
- 对不同语言特性的理解
- 性能和并发模型的认知

**优秀答案要点**:
- **并发模型**: Goroutine 轻量级协程，适合高并发场景（WebSocket连接、媒体流处理）
- **性能**: 编译型语言，性能接近 C++，远超 Node.js
- **内存管理**: GC 优化良好，延迟可控
- **标准库**: net/http、WebSocket 支持完善
- **部署简单**: 单一二进制文件，无需运行时环境
- **生态**: Gin、GORM、gRPC 等成熟框架

**对比**:
- vs Java: 更轻量，启动快，内存占用小
- vs Node.js: 更适合 CPU 密集型任务，类型安全

---

## 第二部分：核心技术深度 (25分钟)

### 2.1 WebRTC 与 SFU 架构

**Q5: 请详细解释 SFU (Selective Forwarding Unit) 架构的工作原理，以及它与 MCU 和 P2P 架构的区别。**

**考察目的**:
- 对 WebRTC 架构的深入理解
- 音视频技术专业知识
- 架构选型的权衡能力

**优秀答案要点**:

**SFU 工作原理**:
- 客户端将媒体流发送到 SFU 服务器
- SFU 选择性转发流到其他参与者，不进行转码
- 每个客户端接收多个独立的媒体流
- 客户端可以选择订阅哪些流（Simulcast/SVC）

**架构对比**:

| 特性 | P2P | SFU | MCU |
|------|-----|-----|-----|
| 服务器负载 | 无 | 中等 | 高 |
| 客户端负载 | 高 | 中等 | 低 |
| 延迟 | 低 | 低 | 中 |
| 扩展性 | 差(2-4人) | 好(10-50人) | 好(100+人) |
| 带宽消耗 | 高 | 中 | 低 |
| 实现复杂度 | 低 | 中 | 高 |

**选择 SFU 的原因**:
- 平衡了性能和扩展性
- 不需要转码，降低服务器成本
- 支持 10-50 人的中型会议
- 客户端可以灵活控制接收的流

---

**Q6: 在你的项目中，WebRTC 信令是如何实现的？请描述从客户端加入会议到建立媒体连接的完整流程。**

**考察目的**:
- WebRTC 信令流程的掌握
- 实际实现经验
- 对协议细节的理解

**优秀答案要点**:

**完整流程**:

1. **客户端加入会议**:
   ```
   POST /api/v1/meetings/{id}/join
   → 返回 ICE servers、room_id、session_id
   ```

2. **建立 WebSocket 连接**:
   ```
   ws://gateway/ws/signaling?token={jwt}&meeting_id={id}&user_id={uid}&peer_id={pid}
   → 信令服务验证 JWT，建立持久连接
   ```

3. **创建 PeerConnection**:
   ```javascript
   pc = new RTCPeerConnection({ iceServers })
   pc.addTrack(localAudioTrack)
   pc.addTrack(localVideoTrack)
   ```

4. **发送 Offer**:
   ```
   offer = await pc.createOffer()
   await pc.setLocalDescription(offer)
   → WebSocket 发送 {type: "offer", sdp: offer.sdp}
   ```

5. **接收 Answer**:
   ```
   ← WebSocket 接收 {type: "answer", sdp: answer.sdp}
   await pc.setRemoteDescription(answer)
   ```

6. **ICE 候选交换**:
   ```
   pc.onicecandidate → WebSocket 发送 candidate
   ← WebSocket 接收 candidate → pc.addIceCandidate()
   ```

7. **媒体流建立**:
   ```
   pc.ontrack → 接收远程媒体流
   pc.onconnectionstatechange → 监控连接状态
   ```

**关键实现细节**:
- 使用 Redis 存储会话状态，支持信令服务水平扩展
- 心跳机制：每 30 秒发送 ping，检测连接活性
- 重连机制：断线后自动重连，恢复会话
- 消息队列：使用 Redis 队列处理信令消息，保证顺序

---

### 2.2 Edge-LLM-Infra 集成

**Q7: 请详细介绍 Edge-LLM-Infra 是什么，以及你是如何将它集成到会议系统中的？**

**考察目的**:
- 对复杂第三方框架的集成能力
- 跨语言通信的实现经验
- 系统集成的架构设计

**优秀答案要点**:

**Edge-LLM-Infra 简介**:
- 分布式 AI 推理基础设施框架
- C++ 实现的高性能推理节点管理器
- 支持 ZMQ/TCP 多协议通信
- 提供 unit-manager 进行节点注册和负载均衡

**集成架构**:
```
Go AI Service (8085)
    ↓ TCP 连接
unit-manager (10001, C++)
    ↓ IPC Socket (/tmp/llm)
AI Inference Node (C++)
    ↓ ZMQ (tcp://5010)
Python Worker (PyTorch)
    ↓
AI Models (Whisper, HuBERT, etc.)
```

**关键实现**:

1. **Go-TCP 桥接**:
```go
type EdgeLLMClient struct {
    host string
    port int
    conn net.Conn
}

func (c *EdgeLLMClient) Setup(modelType string) (*InferenceSession, error) {
    // 建立 TCP 连接
    conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.host, c.port), 5*time.Second)
    
    // 发送 setup 请求
    req := &EdgeLLMRequest{
        RequestID: generateRequestID(),
        WorkID:    "llm",
        Action:    "setup",
        Object:    "llm.setup",
        Data: map[string]interface{}{
            "model": modelType,
            "response_format": "llm.utf-8.stream",
        },
    }
    
    // JSON 序列化并发送
    data, _ := json.Marshal(req)
    conn.Write(append(data, '\n'))
    
    // 读取响应
    reader := bufio.NewReader(conn)
    response, _ := reader.ReadString('\n')
    
    return &InferenceSession{Conn: conn, WorkID: "llm"}, nil
}
```

2. **流式数据传输**:
- 大文件分块传输（每块 64KB）
- 支持音频流式识别
- 使用 `finish` 标志表示最后一块

3. **会话管理**:
- Setup → Inference → Exit 生命周期
- 连接池复用，减少建立连接开销
- 超时控制：setup 10s, inference 30s

**遇到的挑战**:
- C++ 和 Go 之间的数据格式对齐
- ZMQ 消息格式的理解和调试
- 连接超时和重试机制的设计

---

**Q8: 你们使用了哪些 AI 模型？这些模型是如何加载和推理的？**

**考察目的**:
- AI/ML 技术栈的了解
- 模型部署经验
- 性能优化意识

**优秀答案要点**:

**AI 模型清单**:

| 功能 | 模型 | 框架 | 大小 | 推理时间 |
|------|------|------|------|----------|
| 语音识别 | openai/whisper-base | PyTorch | 74MB | ~30s/min |
| 情绪检测 | superb/hubert-base-superb-er | PyTorch | 95MB | ~0.2s |
| 合成检测 | dima806/deepfake-detection | PyTorch | 85MB | ~0.7s |

**模型加载流程**:

1. **下载模型**:
```python
from transformers import WhisperProcessor, WhisperForConditionalGeneration

# 从 HuggingFace 下载
processor = WhisperProcessor.from_pretrained("openai/whisper-base")
model = WhisperForConditionalGeneration.from_pretrained("openai/whisper-base")

# 缓存到 /models 目录
model.save_pretrained("/models/whisper-base")
```

2. **推理服务启动**:
```python
class InferenceWorker:
    def __init__(self):
        self.models = {}
        self.load_models()
    
    def load_models(self):
        # 预加载所有模型到内存
        self.models['whisper'] = WhisperForConditionalGeneration.from_pretrained(
            "/models/whisper-base",
            torch_dtype=torch.float16  # 使用 FP16 加速
        ).to("cuda")  # GPU 推理
        
        self.models['emotion'] = AutoModelForAudioClassification.from_pretrained(
            "/models/hubert-emotion"
        ).to("cuda")
```

3. **推理执行**:
```python
def process_asr(self, audio_data):
    # 解码音频
    audio = base64.b64decode(audio_data)
    waveform, sr = librosa.load(io.BytesIO(audio), sr=16000)
    
    # 预处理
    inputs = self.processor(waveform, sampling_rate=16000, return_tensors="pt")
    inputs = inputs.to("cuda")
    
    # 推理
    with torch.no_grad():
        generated_ids = self.models['whisper'].generate(inputs.input_features)
    
    # 后处理
    transcription = self.processor.batch_decode(generated_ids, skip_special_tokens=True)[0]
    
    return {"text": transcription, "confidence": 0.95}
```

**性能优化**:
- 使用 FP16 半精度推理，速度提升 2x
- GPU 推理（CUDA）
- 批处理：多个请求合并推理
- 模型预加载：避免每次请求加载模型
- 结果缓存：相同输入直接返回缓存结果

---

### 2.3 数据库设计

**Q9: 请介绍一下数据库的设计，为什么使用了 PostgreSQL、MongoDB、Redis 三种数据库？**

**考察目的**:
- 数据库选型能力
- 多数据源管理经验
- 数据模型设计能力

**优秀答案要点**:

**数据库选型理由**:

1. **PostgreSQL (主数据库)**:
   - **用途**: 用户、会议、参与者等结构化数据
   - **优势**: ACID 事务、复杂查询、外键约束
   - **表设计**:
     ```sql
     -- 用户表
     CREATE TABLE users (
         id SERIAL PRIMARY KEY,
         username VARCHAR(50) UNIQUE NOT NULL,
         email VARCHAR(100) UNIQUE NOT NULL,
         password_hash VARCHAR(255) NOT NULL,
         role INTEGER DEFAULT 1,
         status INTEGER DEFAULT 1,
         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
         INDEX idx_username (username),
         INDEX idx_email (email)
     );
     
     -- 会议表
     CREATE TABLE meetings (
         id SERIAL PRIMARY KEY,
         title VARCHAR(255) NOT NULL,
         creator_id INTEGER REFERENCES users(id),
         start_time TIMESTAMP NOT NULL,
         end_time TIMESTAMP NOT NULL,
         status INTEGER DEFAULT 1,
         settings JSONB,  -- 灵活的会议设置
         INDEX idx_creator_id (creator_id),
         INDEX idx_start_time (start_time)
     );
     
     -- 参与者表
     CREATE TABLE meeting_participants (
         id SERIAL PRIMARY KEY,
         meeting_id INTEGER REFERENCES meetings(id) ON DELETE CASCADE,
         user_id INTEGER REFERENCES users(id),
         role INTEGER DEFAULT 1,
         joined_at TIMESTAMP,
         UNIQUE(meeting_id, user_id)
     );
     ```

2. **MongoDB (文档存储)**:
   - **用途**: AI 分析结果、聊天记录、会议事件
   - **优势**: 灵活的 Schema、嵌套文档、高写入性能
   - **集合设计**:
     ```javascript
     // AI 分析结果
     db.ai_analysis_results.insertOne({
         meeting_id: "12345",
         analysis_type: "emotion_detection",
         timestamp: ISODate(),
         results: {
             user_id: 1,
             emotion: "happy",
             confidence: 0.85,
             frame_index: 120
         },
         metadata: {
             model: "hubert-emotion",
             version: "1.0"
         }
     });
     
     // 索引
     db.ai_analysis_results.createIndex({ meeting_id: 1, timestamp: -1 });
     db.ai_analysis_results.createIndex({ analysis_type: 1 });
     ```

3. **Redis (缓存和队列)**:
   - **用途**: 
     - 会话管理（JWT token、WebSocket 连接）
     - 消息队列（AI 推理任务、异步处理）
     - 缓存（用户信息、会议信息）
     - 实时数据（在线用户、房间状态）
   - **数据结构**:
     ```
     # 会话管理
     SET session:{token} {user_data} EX 3600
     
     # 消息队列（优先级队列）
     RPUSH queue:ai:critical {task_json}
     RPUSH queue:ai:high {task_json}
     BLPOP queue:ai:critical queue:ai:high 5
     
     # 缓存
     SETEX user:{id} 300 {user_json}
     
     # 实时数据
     SADD room:{meeting_id}:users {user_id}
     HSET room:{meeting_id}:state participant_count 5
     ```

**数据一致性策略**:
- PostgreSQL 作为 Source of Truth
- MongoDB 和 Redis 作为衍生数据
- 使用消息队列实现最终一致性
- 定期同步任务修复不一致

---

## 第三部分：实现细节与难点 (20分钟)

### 3.1 并发处理

**Q10: 在高并发场景下（比如1000个用户同时在线），你是如何保证系统性能和稳定性的？**

**考察目的**:
- 并发编程能力
- 性能优化经验
- 系统容量规划

**优秀答案要点**:

**1. 连接池管理**:
```go
// 数据库连接池
db.SetMaxOpenConns(100)      // 最大打开连接数
db.SetMaxIdleConns(10)       // 最大空闲连接数
db.SetConnMaxLifetime(3600)  // 连接最大生命周期

// Redis 连接池
redisClient := redis.NewClient(&redis.Options{
    PoolSize:     50,
    MinIdleConns: 10,
    PoolTimeout:  30 * time.Second,
})
```

**2. Goroutine 池**:
```go
// 限制并发 Goroutine 数量
type WorkerPool struct {
    workers   int
    taskQueue chan Task
    wg        sync.WaitGroup
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()
    for task := range p.taskQueue {
        task.Execute()
    }
}
```

**3. 消息队列异步处理**:
```go
// AI 推理任务异步化
type RedisMessageQueue struct {
    client    *redis.Client
    workers   int
    handlers  map[string]MessageHandler
}

// 发布任务
func (q *RedisMessageQueue) Publish(msg *Message) error {
    // 根据优先级选择队列
    queue := q.selectQueue(msg.Priority)
    data, _ := json.Marshal(msg)
    return q.client.RPush(ctx, queue, data).Err()
}

// 消费任务（多个 worker 并发处理）
func (q *RedisMessageQueue) worker(id int) {
    for {
        // 按优先级顺序获取任务
        result, err := q.client.BLPop(ctx, 5*time.Second,
            q.criticalQueue, q.highQueue, q.normalQueue, q.lowQueue).Result()

        if err == nil && len(result) >= 2 {
            var msg Message
            json.Unmarshal([]byte(result[1]), &msg)
            q.handleMessage(&msg)
        }
    }
}
```

**4. 缓存策略**:
```go
// 多级缓存
type CacheService struct {
    localCache  *sync.Map           // 本地内存缓存
    redisCache  *redis.Client       // Redis 分布式缓存
}

func (c *CacheService) Get(key string) (interface{}, error) {
    // L1: 本地缓存
    if val, ok := c.localCache.Load(key); ok {
        return val, nil
    }

    // L2: Redis 缓存
    val, err := c.redisCache.Get(ctx, key).Result()
    if err == nil {
        c.localCache.Store(key, val)
        return val, nil
    }

    // L3: 数据库
    val, err = c.loadFromDB(key)
    if err == nil {
        c.redisCache.Set(ctx, key, val, 5*time.Minute)
        c.localCache.Store(key, val)
    }
    return val, err
}
```

**5. 限流和熔断**:
```go
// Nginx 限流配置
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_req zone=api_limit burst=50 nodelay;

// 服务端限流
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiters sync.Map
}

func (r *RateLimiter) Allow(userID string) bool {
    limiter, _ := r.limiters.LoadOrStore(userID, rate.NewLimiter(10, 20))
    return limiter.(*rate.Limiter).Allow()
}
```

**6. 水平扩展**:
- 无状态服务设计：所有状态存储在 Redis/DB
- Nginx 负载均衡：轮询、最少连接
- 服务发现：使用 etcd 动态注册和发现
- 数据库读写分离：主从复制

**性能指标**:
- 单机支持 1000+ WebSocket 并发连接
- API 响应时间 < 100ms (P95)
- AI 推理吞吐量 > 50 req/s
- 数据库连接池利用率 < 80%

---

### 3.2 实时通信优化

**Q11: WebSocket 连接管理是如何实现的？如何处理断线重连和消息可靠性？**

**考察目的**:
- WebSocket 实战经验
- 网络异常处理能力
- 消息可靠性保证

**优秀答案要点**:

**1. 连接管理**:
```go
type ConnectionManager struct {
    connections sync.Map  // map[userID]*Connection
    mu          sync.RWMutex
}

type Connection struct {
    UserID      string
    MeetingID   string
    Conn        *websocket.Conn
    SendCh      chan []byte
    LastPing    time.Time
    mu          sync.Mutex
}

func (cm *ConnectionManager) AddConnection(userID string, conn *Connection) {
    cm.connections.Store(userID, conn)
    go conn.readPump()
    go conn.writePump()
}

func (c *Connection) readPump() {
    defer c.Close()

    c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.Conn.SetPongHandler(func(string) error {
        c.LastPing = time.Now()
        c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })

    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            break
        }
        c.handleMessage(message)
    }
}

func (c *Connection) writePump() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case message := <-c.SendCh:
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            c.Conn.WriteMessage(websocket.TextMessage, message)

        case <-ticker.C:
            // 发送心跳
            c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            c.Conn.WriteMessage(websocket.PingMessage, nil)
        }
    }
}
```

**2. 断线重连**:
```javascript
// 客户端重连逻辑
class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.connect();
    }

    connect() {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
            this.startHeartbeat();
        };

        this.ws.onclose = () => {
            console.log('WebSocket closed');
            this.stopHeartbeat();
            this.reconnect();
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    reconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnect attempts reached');
            return;
        }

        this.reconnectAttempts++;
        const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

        console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
        setTimeout(() => this.connect(), delay);
    }

    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            if (this.ws.readyState === WebSocket.OPEN) {
                this.ws.send(JSON.stringify({ type: 'ping' }));
            }
        }, 30000);
    }
}
```

**3. 消息可靠性**:
```go
// 消息确认机制
type Message struct {
    ID        string    `json:"id"`
    Type      int       `json:"type"`
    Payload   string    `json:"payload"`
    Timestamp time.Time `json:"timestamp"`
    Ack       bool      `json:"ack"`
}

// 发送消息并等待确认
func (c *Connection) SendWithAck(msg *Message) error {
    msg.ID = generateMessageID()
    msg.Ack = false

    // 存储到 Redis（等待确认）
    c.redis.HSet(ctx, fmt.Sprintf("pending:%s", c.UserID), msg.ID, msg)
    c.redis.Expire(ctx, fmt.Sprintf("pending:%s", c.UserID), 5*time.Minute)

    // 发送消息
    data, _ := json.Marshal(msg)
    c.SendCh <- data

    // 等待确认（超时重发）
    go c.waitForAck(msg.ID, 3)

    return nil
}

func (c *Connection) waitForAck(msgID string, retries int) {
    for i := 0; i < retries; i++ {
        time.Sleep(2 * time.Second)

        // 检查是否已确认
        exists, _ := c.redis.HExists(ctx, fmt.Sprintf("pending:%s", c.UserID), msgID).Result()
        if !exists {
            return  // 已确认
        }

        // 重发
        msg, _ := c.redis.HGet(ctx, fmt.Sprintf("pending:%s", c.UserID), msgID).Result()
        c.SendCh <- []byte(msg)
    }
}

// 处理确认消息
func (c *Connection) handleAck(msgID string) {
    c.redis.HDel(ctx, fmt.Sprintf("pending:%s", c.UserID), msgID)
}
```

**4. 会话恢复**:
```go
// 断线后恢复会话
func (s *SignalingService) ResumeSession(userID, sessionID string) error {
    // 从 Redis 恢复会话状态
    sessionData, err := s.redis.Get(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
    if err != nil {
        return errors.New("session not found")
    }

    var session Session
    json.Unmarshal([]byte(sessionData), &session)

    // 恢复 WebRTC 连接状态
    // 重新发送未确认的消息
    pendingMsgs, _ := s.redis.HGetAll(ctx, fmt.Sprintf("pending:%s", userID)).Result()
    for _, msg := range pendingMsgs {
        // 重发消息
    }

    return nil
}
```

---

### 3.3 媒体处理

**Q12: 会议录制功能是如何实现的？如何保证录制文件的质量和存储效率？**

**考察目的**:
- 音视频处理经验
- FFmpeg 使用能力
- 存储优化策略

**优秀答案要点**:

**1. 录制架构**:
```
WebRTC Stream → SFU → Recording Service → FFmpeg → MinIO
```

**2. 录制实现**:
```go
type RecordingService struct {
    ffmpegPath string
    storagePath string
    minioClient *minio.Client
}

func (r *RecordingService) StartRecording(meetingID string, streams []MediaStream) error {
    // 创建 FFmpeg 进程
    outputFile := fmt.Sprintf("/tmp/recording_%s.mp4", meetingID)

    cmd := exec.Command(r.ffmpegPath,
        "-f", "webm",                    // 输入格式
        "-i", streams[0].AudioURL,       // 音频流
        "-f", "webm",
        "-i", streams[0].VideoURL,       // 视频流
        "-c:v", "libx264",               // 视频编码器
        "-preset", "medium",             // 编码速度
        "-crf", "23",                    // 质量（0-51，越小越好）
        "-c:a", "aac",                   // 音频编码器
        "-b:a", "128k",                  // 音频码率
        "-movflags", "+faststart",       // 优化流式播放
        outputFile,
    )

    // 启动录制
    if err := cmd.Start(); err != nil {
        return err
    }

    // 保存进程信息
    r.saveRecordingProcess(meetingID, cmd.Process.Pid, outputFile)

    // 等待录制完成
    go r.waitForRecording(meetingID, cmd, outputFile)

    return nil
}

func (r *RecordingService) waitForRecording(meetingID string, cmd *exec.Cmd, outputFile string) {
    cmd.Wait()

    // 上传到 MinIO
    r.uploadToMinIO(meetingID, outputFile)

    // 删除临时文件
    os.Remove(outputFile)

    // 更新数据库
    r.updateMeetingRecording(meetingID, fmt.Sprintf("recordings/%s.mp4", meetingID))
}

func (r *RecordingService) uploadToMinIO(meetingID, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    stat, _ := file.Stat()

    _, err = r.minioClient.PutObject(
        context.Background(),
        "meeting-recordings",
        fmt.Sprintf("%s.mp4", meetingID),
        file,
        stat.Size(),
        minio.PutObjectOptions{
            ContentType: "video/mp4",
        },
    )

    return err
}
```

**3. 多流合成**:
```go
// 合成多个参与者的视频流
func (r *RecordingService) ComposeMultipleStreams(streams []MediaStream) error {
    // 使用 FFmpeg filter_complex 合成画面
    filterComplex := `
        [0:v]scale=640:360[v0];
        [1:v]scale=640:360[v1];
        [2:v]scale=640:360[v2];
        [3:v]scale=640:360[v3];
        [v0][v1][v2][v3]xstack=inputs=4:layout=0_0|w0_0|0_h0|w0_h0[vout]
    `

    cmd := exec.Command(r.ffmpegPath,
        "-i", streams[0].URL,
        "-i", streams[1].URL,
        "-i", streams[2].URL,
        "-i", streams[3].URL,
        "-filter_complex", filterComplex,
        "-map", "[vout]",
        "-c:v", "libx264",
        "output.mp4",
    )

    return cmd.Run()
}
```

**4. 存储优化**:
- **编码参数优化**: CRF 23（平衡质量和大小）
- **分辨率自适应**: 根据参与者数量调整分辨率
- **分段存储**: 每 10 分钟一个文件，便于断点续传
- **压缩**: 使用 H.264 High Profile
- **CDN 加速**: MinIO 配合 CDN 分发录制文件

**存储成本**:
- 1 小时 1080p 录制 ≈ 1.5GB
- 使用 MinIO 对象存储，成本低
- 冷数据归档到 S3 Glacier

---

### 3.4 安全性

**Q13: 系统的安全性是如何保证的？包括认证、授权、数据加密等方面。**

**考察目的**:
- 安全意识
- 认证授权机制
- 数据保护能力

**优秀答案要点**:

**1. 认证机制 (JWT)**:
```go
type JWTService struct {
    secretKey []byte
    expireTime time.Duration
}

func (j *JWTService) GenerateToken(userID int, username string) (string, error) {
    claims := jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "role":     "user",
        "exp":      time.Now().Add(j.expireTime).Unix(),
        "iat":      time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(j.secretKey)
}

func (j *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return j.secretKey, nil
    })
}

// 中间件
func AuthMiddleware(jwtService *JWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwtService.ValidateToken(tokenString)
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        claims := token.Claims.(jwt.MapClaims)
        c.Set("user_id", int(claims["user_id"].(float64)))
        c.Set("username", claims["username"].(string))
        c.Next()
    }
}
```

**2. 权限控制 (RBAC)**:
```go
type Role int

const (
    RoleGuest Role = iota
    RoleUser
    RoleModerator
    RoleAdmin
    RoleSuperAdmin
)

type Permission struct {
    Resource string
    Action   string
}

var rolePermissions = map[Role][]Permission{
    RoleUser: {
        {Resource: "meeting", Action: "create"},
        {Resource: "meeting", Action: "join"},
        {Resource: "meeting", Action: "leave"},
    },
    RoleModerator: {
        {Resource: "meeting", Action: "create"},
        {Resource: "meeting", Action: "manage"},
        {Resource: "user", Action: "mute"},
        {Resource: "user", Action: "kick"},
    },
    RoleAdmin: {
        {Resource: "*", Action: "*"},
    },
}

func CheckPermission(userRole Role, resource, action string) bool {
    permissions := rolePermissions[userRole]
    for _, perm := range permissions {
        if (perm.Resource == "*" || perm.Resource == resource) &&
           (perm.Action == "*" || perm.Action == action) {
            return true
        }
    }
    return false
}

// 权限中间件
func RequirePermission(resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetInt("user_role")
        if !CheckPermission(Role(userRole), resource, action) {
            c.JSON(403, gin.H{"error": "permission denied"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**3. 数据加密**:
```go
// 密码加密（bcrypt）
import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// 敏感数据加密（AES-256）
import "crypto/aes"
import "crypto/cipher"

func Encrypt(plaintext, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}
```

**4. 安全头部 (Nginx)**:
```nginx
# 安全头部
add_header X-Frame-Options "DENY" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Strict-Transport-Security "max-age=31536000" always;
add_header Content-Security-Policy "default-src 'self'" always;

# CORS 配置
add_header Access-Control-Allow-Origin "https://meeting.example.com" always;
add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
add_header Access-Control-Allow-Headers "Authorization, Content-Type" always;
```

**5. 输入验证**:
```go
import "github.com/go-playground/validator/v10"

type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=100"`
}

func ValidateRequest(req interface{}) error {
    validate := validator.New()
    return validate.Struct(req)
}

// SQL 注入防护（使用参数化查询）
func GetUser(db *gorm.DB, username string) (*User, error) {
    var user User
    // 使用 ? 占位符，GORM 自动转义
    err := db.Where("username = ?", username).First(&user).Error
    return &user, err
}
```

**6. 限流防护**:
```nginx
# Nginx 限流
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=5r/s;

location /api/v1/auth {
    limit_req zone=auth_limit burst=10 nodelay;
}
```

**安全检查清单**:
- ✅ JWT 认证
- ✅ RBAC 权限控制
- ✅ 密码 bcrypt 加密
- ✅ HTTPS/TLS 传输加密
- ✅ SQL 注入防护
- ✅ XSS 防护
- ✅ CSRF 防护
- ✅ 限流防 DDoS
- ✅ 输入验证
- ✅ 安全审计日志

---

## 第四部分：问题排查与优化 (15分钟)

### 4.1 问题定位

**Q14: 在开发过程中遇到过哪些比较棘手的 Bug？你是如何定位和解决的？**

**考察目的**:
- 问题排查能力
- 调试技巧
- 解决问题的思路

**优秀答案要点**:

**案例 1: Edge-LLM-Infra 集成问题**

**问题描述**:
- AI 推理请求发送后，unit-manager 返回 "bad_any_cast" 错误
- 无法正常调用 AI 模型

**排查过程**:
1. **查看日志**:
   ```bash
   docker logs meeting-edge-model-infra
   # 发现: terminate called after throwing an instance of 'std::bad_any_cast'
   ```

2. **分析原因**:
   - C++ 的 `std::any` 类型转换失败
   - 可能是配置文件格式不匹配

3. **检查配置**:
   ```bash
   docker exec meeting-edge-model-infra cat /app/master_config.json
   # 发现: 配置文件路径错误，导致加载失败
   ```

4. **修复方案**:
   ```dockerfile
   # 修改 Dockerfile，确保配置文件正确复制
   COPY Edge-LLM-Infra-master/unit-manager/master_config.json /app/

   # 修改启动命令，指定配置文件路径
   CMD ["./unit-manager", "--config", "/app/master_config.json"]
   ```

5. **验证修复**:
   ```bash
   # 重新构建镜像
   docker-compose build edge-model-infra

   # 重启服务
   docker-compose up -d edge-model-infra

   # 测试 AI 推理
   curl -X POST http://localhost:8085/api/v1/ai/asr -d '{"audio_data":"..."}'
   ```

**经验总结**:
- 跨语言集成要特别注意数据格式和配置
- 使用详细的日志记录，便于问题定位
- 容器化环境要确保文件路径正确

---

**案例 2: WebSocket 连接频繁断开**

**问题描述**:
- 用户反馈会议中频繁掉线
- WebSocket 连接每隔几分钟就断开

**排查过程**:
1. **监控指标**:
   ```bash
   # 查看 Nginx 日志
   tail -f /var/log/nginx/access.log | grep "ws/signaling"
   # 发现: 大量 499 状态码（客户端主动断开）
   ```

2. **抓包分析**:
   ```bash
   tcpdump -i any -w websocket.pcap port 8081
   # 分析: 60 秒后连接被重置
   ```

3. **定位原因**:
   - Nginx 默认 `proxy_read_timeout` 为 60 秒
   - 没有心跳机制，连接被认为空闲而关闭

4. **修复方案**:
   ```nginx
   # Nginx 配置
   location /ws/signaling {
       proxy_pass http://signaling_service;
       proxy_http_version 1.1;
       proxy_set_header Upgrade $http_upgrade;
       proxy_set_header Connection "upgrade";

       # 增加超时时间
       proxy_read_timeout 86400s;  # 24 小时
       proxy_send_timeout 86400s;
       proxy_connect_timeout 60s;
   }
   ```

   ```go
   // 服务端心跳
   func (c *Connection) writePump() {
       ticker := time.NewTicker(30 * time.Second)
       defer ticker.Stop()

       for {
           select {
           case <-ticker.C:
               c.Conn.WriteMessage(websocket.PingMessage, nil)
           }
       }
   }
   ```

   ```javascript
   // 客户端心跳响应
   ws.addEventListener('ping', () => {
       ws.pong();
   });
   ```

5. **效果验证**:
   - 连接稳定性提升到 99.9%
   - 平均连接时长从 5 分钟提升到 2+ 小时

---

**案例 3: AI 推理性能瓶颈**

**问题描述**:
- AI 推理请求响应时间过长（> 60s）
- 高并发时出现超时

**排查过程**:
1. **性能分析**:
   ```bash
   # 查看 AI 服务日志
   docker logs meeting-ai-inference-worker
   # 发现: 每次请求都在加载模型
   ```

2. **代码审查**:
   ```python
   # 问题代码
   def process_asr(audio_data):
       # 每次都加载模型！
       model = WhisperForConditionalGeneration.from_pretrained("openai/whisper-base")
       # ...
   ```

3. **优化方案**:
   ```python
   # 优化后：模型预加载
   class InferenceWorker:
       def __init__(self):
           self.models = {}
           self.load_models()

       def load_models(self):
           # 启动时加载所有模型到内存
           self.models['whisper'] = WhisperForConditionalGeneration.from_pretrained(
               "/models/whisper-base",
               torch_dtype=torch.float16  # FP16 加速
           ).to("cuda")

           self.models['emotion'] = AutoModelForAudioClassification.from_pretrained(
               "/models/hubert-emotion"
           ).to("cuda")

       def process_asr(self, audio_data):
           # 直接使用预加载的模型
           model = self.models['whisper']
           # ...
   ```

4. **性能对比**:
   - 优化前: 30-60s/请求
   - 优化后: 2-5s/请求
   - 性能提升 10x+

---

### 4.2 性能优化

**Q15: 你做过哪些性能优化？效果如何？**

**考察目的**:
- 性能优化经验
- 性能分析能力
- 量化思维

**优秀答案要点**:

**优化 1: 数据库查询优化**

**问题**: 会议列表查询慢（> 2s）

**优化前**:
```go
func GetMeetings(userID int) ([]Meeting, error) {
    var meetings []Meeting
    db.Where("creator_id = ?", userID).Find(&meetings)

    // N+1 查询问题
    for i := range meetings {
        db.Model(&meetings[i]).Association("Participants").Find(&meetings[i].Participants)
    }

    return meetings, nil
}
```

**优化后**:
```go
func GetMeetings(userID int) ([]Meeting, error) {
    var meetings []Meeting

    // 使用 Preload 预加载关联数据
    db.Preload("Participants").
       Preload("Creator").
       Where("creator_id = ?", userID).
       Find(&meetings)

    return meetings, nil
}

// 添加索引
CREATE INDEX idx_meetings_creator_id ON meetings(creator_id);
CREATE INDEX idx_meeting_participants_meeting_id ON meeting_participants(meeting_id);
```

**效果**:
- 查询时间: 2s → 50ms
- 数据库查询次数: N+1 → 1
- 性能提升 40x

---

**优化 2: Redis 缓存**

**问题**: 用户信息频繁查询数据库

**优化前**:
```go
func GetUser(userID int) (*User, error) {
    var user User
    db.First(&user, userID)
    return &user, nil
}
```

**优化后**:
```go
func GetUser(userID int) (*User, error) {
    // 先查缓存
    cacheKey := fmt.Sprintf("user:%d", userID)
    cached, err := redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 缓存未命中，查数据库
    var user User
    db.First(&user, userID)

    // 写入缓存
    data, _ := json.Marshal(user)
    redis.Set(ctx, cacheKey, data, 5*time.Minute)

    return &user, nil
}
```

**效果**:
- 缓存命中率: 95%
- 响应时间: 10ms → 1ms
- 数据库负载降低 90%

---

**优化 3: 消息队列异步化**

**问题**: AI 推理阻塞 HTTP 请求

**优化前**:
```go
func HandleASR(c *gin.Context) {
    var req ASRRequest
    c.BindJSON(&req)

    // 同步调用，阻塞 30 秒
    result, err := aiService.SpeechRecognition(req)

    c.JSON(200, result)
}
```

**优化后**:
```go
func HandleASR(c *gin.Context) {
    var req ASRRequest
    c.BindJSON(&req)

    // 生成任务 ID
    taskID := generateTaskID()

    // 发布到消息队列
    messageQueue.Publish(&Message{
        ID:       taskID,
        Type:     "asr",
        Payload:  req,
        Priority: PriorityHigh,
    })

    // 立即返回任务 ID
    c.JSON(202, gin.H{
        "task_id": taskID,
        "status":  "processing",
    })
}

// 客户端轮询结果
func GetTaskResult(c *gin.Context) {
    taskID := c.Param("task_id")

    result, err := redis.Get(ctx, fmt.Sprintf("task:%s", taskID)).Result()
    if err != nil {
        c.JSON(200, gin.H{"status": "processing"})
        return
    }

    c.JSON(200, gin.H{"status": "completed", "result": result})
}
```

**效果**:
- API 响应时间: 30s → 50ms
- 吞吐量: 2 req/s → 50 req/s
- 用户体验提升（不阻塞）

---

**优化 4: Goroutine 池**

**问题**: 高并发时 Goroutine 数量爆炸

**优化前**:
```go
func HandleRequest(c *gin.Context) {
    // 每个请求创建一个 Goroutine
    go processRequest(c)
}
```

**优化后**:
```go
type WorkerPool struct {
    workers   int
    taskQueue chan Task
    wg        sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    pool := &WorkerPool{
        workers:   workers,
        taskQueue: make(chan Task, 1000),
    }
    pool.Start()
    return pool
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()
    for task := range p.taskQueue {
        task.Execute()
    }
}

func (p *WorkerPool) Submit(task Task) {
    p.taskQueue <- task
}
```

**效果**:
- Goroutine 数量: 10000+ → 100
- 内存占用: 2GB → 500MB
- GC 压力降低 80%

---

**性能优化总结**:

| 优化项 | 优化前 | 优化后 | 提升 |
|--------|--------|--------|------|
| 数据库查询 | 2s | 50ms | 40x |
| 用户信息查询 | 10ms | 1ms | 10x |
| AI 推理吞吐量 | 2 req/s | 50 req/s | 25x |
| 内存占用 | 2GB | 500MB | 4x |
| API P95 延迟 | 500ms | 80ms | 6x |

---

### 4.3 监控与告警

**Q16: 系统的监控和告警是如何实现的？你关注哪些关键指标？**

**考察目的**:
- 可观测性意识
- 监控体系建设
- 运维能力

**优秀答案要点**:

**1. 监控架构**:
```
应用 → Prometheus (指标) → Grafana (可视化)
     → Jaeger (链路追踪)
     → Loki (日志聚合) → Grafana
     → Alertmanager (告警)
```

**2. 关键指标**:

**业务指标**:
```go
// 使用 Prometheus 客户端
import "github.com/prometheus/client_golang/prometheus"

var (
    // 会议创建数
    meetingCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "meeting_created_total",
            Help: "Total number of meetings created",
        },
        []string{"status"},
    )

    // 在线用户数
    onlineUsers = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "online_users",
            Help: "Number of online users",
        },
    )

    // API 请求延迟
    apiDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_request_duration_seconds",
            Help:    "API request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )

    // AI 推理延迟
    aiInferenceDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ai_inference_duration_seconds",
            Help:    "AI inference duration",
            Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
        },
        []string{"model", "status"},
    )
)

// 中间件记录指标
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        duration := time.Since(start).Seconds()
        apiDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            fmt.Sprintf("%d", c.Writer.Status()),
        ).Observe(duration)
    }
}
```

**系统指标**:
- CPU 使用率
- 内存使用率
- 磁盘 I/O
- 网络带宽
- Goroutine 数量
- GC 暂停时间

**数据库指标**:
- 连接池使用率
- 查询延迟
- 慢查询数量
- 死锁数量

**3. 告警规则**:
```yaml
# alert_rules.yml
groups:
  - name: api_alerts
    interval: 30s
    rules:
      # API 错误率告警
      - alert: HighAPIErrorRate
        expr: |
          sum(rate(api_request_duration_seconds_count{status=~"5.."}[5m]))
          /
          sum(rate(api_request_duration_seconds_count[5m]))
          > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High API error rate (> 5%)"
          description: "API error rate is {{ $value | humanizePercentage }}"

      # API 延迟告警
      - alert: HighAPILatency
        expr: |
          histogram_quantile(0.95,
            sum(rate(api_request_duration_seconds_bucket[5m])) by (le, endpoint)
          ) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API latency (P95 > 1s)"
          description: "Endpoint {{ $labels.endpoint }} P95 latency is {{ $value }}s"

      # 在线用户数异常
      - alert: OnlineUsersDropped
        expr: |
          (online_users - online_users offset 5m) / online_users offset 5m < -0.5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Online users dropped by 50%"
          description: "Current: {{ $value }}, 5m ago: {{ $value offset 5m }}"

      # 数据库连接池告警
      - alert: DatabaseConnectionPoolExhausted
        expr: |
          database_connections_in_use / database_connections_max > 0.9
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool almost exhausted (> 90%)"
          description: "{{ $value | humanizePercentage }} connections in use"
```

**4. Grafana 仪表板**:
- **系统概览**: CPU、内存、网络、磁盘
- **业务指标**: 在线用户、会议数、消息数
- **API 性能**: QPS、延迟分布、错误率
- **AI 推理**: 推理延迟、吞吐量、队列长度
- **数据库**: 连接数、查询延迟、慢查询

**5. 分布式追踪 (Jaeger)**:
```go
import "github.com/opentracing/opentracing-go"

func HandleRequest(c *gin.Context) {
    // 创建 span
    span := opentracing.StartSpan("handle_request")
    defer span.Finish()

    // 调用其他服务
    ctx := opentracing.ContextWithSpan(context.Background(), span)

    // 数据库查询
    dbSpan, _ := opentracing.StartSpanFromContext(ctx, "db_query")
    db.WithContext(ctx).Find(&users)
    dbSpan.Finish()

    // AI 推理
    aiSpan, _ := opentracing.StartSpanFromContext(ctx, "ai_inference")
    aiService.Process(ctx, data)
    aiSpan.Finish()
}
```

**关键指标阈值**:
- API P95 延迟 < 100ms
- API 错误率 < 1%
- 数据库连接池使用率 < 80%
- CPU 使用率 < 70%
- 内存使用率 < 80%
- AI 推理队列长度 < 100

---

## 第五部分：扩展性与架构演进 (10分钟)

### 5.1 系统扩展性

**Q17: 如果用户量从 1000 增长到 100万，你会如何扩展这个系统？**

**考察目的**:
- 系统扩展能力
- 架构演进思路
- 容量规划能力

**优秀答案要点**:

**1. 水平扩展策略**:

**应用层扩展**:
```yaml
# Kubernetes 部署
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 10  # 从 2 个扩展到 10 个
  template:
    spec:
      containers:
      - name: user-service
        image: meeting-system/user-service:latest
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: user-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: user-service
  minReplicas: 5
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

**数据库扩展**:
```
主从复制 + 读写分离:
┌─────────┐
│ Master  │ ← 写操作
└────┬────┘
     │ 复制
     ├──────┬──────┬──────┐
     ↓      ↓      ↓      ↓
  Slave1 Slave2 Slave3 Slave4 ← 读操作

分库分表:
- 用户表: 按 user_id % 16 分 16 个库
- 会议表: 按 meeting_id % 32 分 32 个库
- 时间范围分表: meetings_2024_01, meetings_2024_02...
```

```go
// 分库分表实现
type ShardingRouter struct {
    shards []*gorm.DB
}

func (r *ShardingRouter) GetShard(userID int) *gorm.DB {
    shardIndex := userID % len(r.shards)
    return r.shards[shardIndex]
}

func (r *ShardingRouter) GetUser(userID int) (*User, error) {
    db := r.GetShard(userID)
    var user User
    err := db.First(&user, userID).Error
    return &user, err
}
```

**缓存扩展**:
```
Redis 集群:
- 使用 Redis Cluster 模式
- 16384 个 hash slot
- 3 主 3 从配置
- 自动故障转移

redis-cluster:
  nodes:
    - redis-1:6379 (master, slots 0-5460)
    - redis-2:6379 (master, slots 5461-10922)
    - redis-3:6379 (master, slots 10923-16383)
    - redis-4:6379 (slave of redis-1)
    - redis-5:6379 (slave of redis-2)
    - redis-6:6379 (slave of redis-3)
```

**2. 架构优化**:

**CDN 加速**:
- 静态资源（前端、录制文件）使用 CDN
- 就近访问，降低延迟
- 减轻源站压力

**消息队列集群**:
```
Kafka 集群:
- 3 个 broker
- 分区数: 32
- 副本数: 3
- 吞吐量: 100k msg/s

topics:
  - ai-inference-tasks (32 partitions)
  - meeting-events (16 partitions)
  - notification-queue (8 partitions)
```

**微服务拆分**:
```
原有服务:
- user-service
- meeting-service
- signaling-service
- media-service
- ai-service

新增服务:
- notification-service (通知服务)
- recording-service (录制服务)
- analytics-service (分析服务)
- billing-service (计费服务)
```

**3. 性能优化**:

**数据库优化**:
- 冷热数据分离（历史会议归档）
- 索引优化（覆盖索引、联合索引）
- 查询优化（避免全表扫描）
- 连接池调优

**缓存优化**:
- 多级缓存（本地缓存 + Redis）
- 缓存预热（热点数据提前加载）
- 缓存穿透防护（布隆过滤器）
- 缓存雪崩防护（随机过期时间）

**AI 推理优化**:
- 模型量化（FP16 → INT8）
- 批处理推理（batch size = 8）
- GPU 集群（多卡并行）
- 模型缓存（热门模型常驻内存）

**4. 容量规划**:

| 指标 | 1000 用户 | 100万 用户 | 扩展倍数 |
|------|-----------|------------|----------|
| 应用服务器 | 2 台 | 50 台 | 25x |
| 数据库 | 1 主 2 从 | 16 主 32 从 | 16x |
| Redis | 1 主 1 从 | 3 主 3 从 (集群) | 3x |
| AI 推理节点 | 2 台 | 20 台 | 10x |
| 带宽 | 100 Mbps | 10 Gbps | 100x |
| 存储 | 1 TB | 100 TB | 100x |

**成本估算**:
- 服务器成本: $500/月 → $50,000/月
- 带宽成本: $100/月 → $10,000/月
- 存储成本: $50/月 → $5,000/月
- **总成本**: $650/月 → $65,000/月

---

### 5.2 技术债务

**Q18: 在项目开发过程中，有哪些技术债务？你打算如何偿还？**

**考察目的**:
- 对技术债务的认识
- 代码质量意识
- 重构能力

**优秀答案要点**:

**技术债务清单**:

**1. 代码质量**:
- **问题**: 部分代码缺少单元测试
- **影响**: 重构风险高，回归测试困难
- **偿还计划**:
  ```go
  // 补充单元测试
  func TestUserService_CreateUser(t *testing.T) {
      // 使用 testify 框架
      suite.Run(t, new(UserServiceTestSuite))
  }

  // 目标: 测试覆盖率 > 80%
  go test -cover ./...
  ```

**2. 架构设计**:
- **问题**: 服务间耦合度较高
- **影响**: 难以独立部署和扩展
- **偿还计划**:
  - 引入事件驱动架构（Event Sourcing）
  - 使用消息队列解耦
  - 定义清晰的服务边界

**3. 性能优化**:
- **问题**: 部分 API 未做缓存
- **影响**: 数据库压力大
- **偿还计划**:
  - 识别热点 API
  - 添加 Redis 缓存
  - 实施缓存预热

**4. 监控告警**:
- **问题**: 部分服务缺少监控指标
- **影响**: 问题发现不及时
- **偿还计划**:
  - 补充 Prometheus 指标
  - 完善 Grafana 仪表板
  - 添加关键告警规则

**5. 文档**:
- **问题**: API 文档不完整
- **影响**: 前后端协作效率低
- **偿还计划**:
  - 使用 Swagger 自动生成 API 文档
  - 编写架构设计文档
  - 补充运维手册

**偿还优先级**:
1. **P0 (立即)**: 安全漏洞、性能瓶颈
2. **P1 (本周)**: 监控告警、关键 Bug
3. **P2 (本月)**: 单元测试、代码重构
4. **P3 (本季度)**: 文档完善、架构优化

---

### 5.3 未来规划

**Q19: 如果让你继续负责这个项目，你会在哪些方面进行改进？**

**考察目的**:
- 技术视野
- 创新能力
- 产品思维

**优秀答案要点**:

**1. 功能增强**:

**AI 能力升级**:
- 实时字幕翻译（多语言支持）
- 会议智能摘要（自动生成会议纪要）
- 虚拟背景（AI 背景分割）
- 噪音抑制（深度学习降噪）
- 手势识别（举手、点赞等）

**协作功能**:
- 白板协作（实时绘图）
- 文档共享（在线编辑）
- 投票功能（实时投票）
- 分组讨论（Breakout Rooms）

**2. 技术升级**:

**WebRTC 优化**:
- 支持 Simulcast（多码率自适应）
- 支持 SVC（可伸缩视频编码）
- 优化 ICE 候选收集
- 实现 WebRTC 录制

**AI 模型升级**:
- 使用更大的模型（Whisper Large）
- 支持更多语言
- 提升识别准确率
- 降低推理延迟

**架构演进**:
```
当前架构: 单体 SFU
↓
目标架构: 分布式 SFU 集群

┌─────────┐     ┌─────────┐     ┌─────────┐
│ SFU-1   │────│ SFU-2   │────│ SFU-3   │
└─────────┘     └─────────┘     └─────────┘
     │               │               │
     └───────────────┴───────────────┘
                     │
              ┌──────▼──────┐
              │ Coordinator │
              └─────────────┘

优势:
- 支持更大规模会议（100+ 人）
- 跨区域低延迟
- 高可用性
```

**3. 运维改进**:

**CI/CD 流程**:
```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  script:
    - go test -v -cover ./...
    - go vet ./...
    - golangci-lint run

build:
  stage: build
  script:
    - docker build -t meeting-system/user-service:$CI_COMMIT_SHA .
    - docker push meeting-system/user-service:$CI_COMMIT_SHA

deploy:
  stage: deploy
  script:
    - kubectl set image deployment/user-service user-service=meeting-system/user-service:$CI_COMMIT_SHA
    - kubectl rollout status deployment/user-service
```

**灰度发布**:
```yaml
# Istio VirtualService
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: user-service
spec:
  hosts:
  - user-service
  http:
  - match:
    - headers:
        x-version:
          exact: "v2"
    route:
    - destination:
        host: user-service
        subset: v2
  - route:
    - destination:
        host: user-service
        subset: v1
      weight: 90
    - destination:
        host: user-service
        subset: v2
      weight: 10
```

**4. 商业化**:

**计费系统**:
- 按会议时长计费
- 按参与人数计费
- 按 AI 功能使用量计费
- 订阅制（月费/年费）

**企业版功能**:
- SSO 单点登录
- 企业级安全（数据加密、审计）
- 自定义品牌
- API 集成
- 专属客服

---

## 第六部分：综合能力评估 (5分钟)

### 6.1 团队协作

**Q20: 在这个项目中，你是如何与团队成员协作的？遇到过哪些协作上的挑战？**

**考察目的**:
- 团队协作能力
- 沟通能力
- 冲突解决能力

**优秀答案要点**:

**协作方式**:
- **代码审查**: 所有代码必须经过 Code Review
- **技术分享**: 每周技术分享会
- **文档协作**: 使用 Confluence 维护文档
- **任务管理**: 使用 Jira 跟踪任务
- **即时沟通**: Slack/钉钉

**协作挑战**:

**案例: 前后端接口对接**
- **问题**: 前端需要的数据格式与后端返回不一致
- **解决**:
  - 制定统一的 API 规范
  - 使用 Swagger 自动生成文档
  - 前后端联调测试

**案例: 跨团队协作**
- **问题**: AI 团队和后端团队对接口理解不一致
- **解决**:
  - 组织联合设计会议
  - 明确接口定义和数据格式
  - 编写详细的集成文档

---

### 6.2 学习能力

**Q21: 在这个项目中，你学到了哪些新技术？是如何学习的？**

**考察目的**:
- 学习能力
- 技术热情
- 成长潜力

**优秀答案要点**:

**学到的新技术**:
1. **WebRTC**: 从零学习音视频通信
2. **Edge-LLM-Infra**: C++ 框架集成
3. **AI 模型部署**: PyTorch 模型推理
4. **Kubernetes**: 容器编排
5. **分布式追踪**: Jaeger 使用

**学习方法**:
- **官方文档**: 阅读 WebRTC、Kubernetes 官方文档
- **开源项目**: 研究 Janus、Jitsi 等开源项目
- **技术博客**: 阅读技术博客和论文
- **实践**: 动手实现 Demo
- **社区交流**: 参与技术社区讨论

---

### 6.3 项目总结

**Q22: 请总结一下这个项目的亮点和不足，以及你的收获。**

**考察目的**:
- 总结能力
- 自我反思
- 项目理解深度

**优秀答案要点**:

**项目亮点**:
1. **技术栈先进**: Go + WebRTC + AI，技术栈现代化
2. **架构合理**: 微服务架构，易于扩展
3. **AI 集成**: 成功集成 Edge-LLM-Infra，实现真实 AI 推理
4. **性能优秀**: 支持 1000+ 并发，延迟 < 100ms
5. **可观测性**: 完善的监控告警体系

**项目不足**:
1. **测试覆盖**: 单元测试覆盖率不足
2. **文档**: 部分模块缺少详细文档
3. **扩展性**: 部分模块耦合度较高
4. **安全性**: 需要进一步加强安全审计

**个人收获**:
1. **技术能力**: 掌握了 WebRTC、AI 模型部署等新技术
2. **架构能力**: 理解了微服务架构设计
3. **工程能力**: 提升了代码质量和工程实践
4. **问题解决**: 锻炼了问题定位和解决能力
5. **团队协作**: 提升了团队协作和沟通能力

---

## 附录：技术栈总结

### 后端技术栈
- **语言**: Go 1.21+
- **框架**: Gin (HTTP), GORM (ORM), gRPC
- **数据库**: PostgreSQL, MongoDB, Redis
- **存储**: MinIO (对象存储)
- **消息队列**: Redis (优先级队列)
- **服务发现**: etcd
- **监控**: Prometheus, Grafana, Jaeger, Loki

### AI 技术栈
- **框架**: Edge-LLM-Infra (C++)
- **模型**: Whisper, HuBERT, Deepfake Detection
- **推理**: PyTorch, ONNX Runtime
- **通信**: ZMQ, TCP

### 前端技术栈
- **桌面**: Qt6 + QML
- **Web**: React/Vue + WebRTC
- **移动**: React Native

### 运维技术栈
- **容器**: Docker, Docker Compose
- **编排**: Kubernetes
- **网关**: Nginx
- **CI/CD**: GitLab CI, Jenkins

### 开发工具
- **IDE**: GoLand, VS Code
- **版本控制**: Git, GitLab
- **API 测试**: Postman, curl
- **性能分析**: pprof, Jaeger

---

## 面试评分标准

### 优秀 (90-100分)
- 对项目有深入理解，能清晰描述架构和实现细节
- 能够独立解决复杂技术问题
- 有性能优化和系统扩展经验
- 代码质量高，有良好的工程实践
- 学习能力强，技术视野广

### 良好 (75-89分)
- 对项目有较好理解，能描述主要功能和技术栈
- 能够在指导下解决技术问题
- 有一定的性能优化意识
- 代码质量较好
- 学习能力较强

### 合格 (60-74分)
- 对项目有基本了解，能描述自己负责的模块
- 能够完成基本的开发任务
- 代码质量一般
- 需要持续学习和提升

### 不合格 (<60分)
- 对项目理解不深，无法清晰描述技术细节
- 解决问题能力弱
- 代码质量差
- 学习能力不足

---

**面试官备注**:
- 根据候选人的实际情况，灵活调整问题难度和深度
- 鼓励候选人画图说明，更直观地展示理解
- 关注候选人的思考过程，而不仅仅是答案
- 给予候选人充分的表达时间
- 适当追问，深入了解候选人的真实水平

---

**文档版本**: v1.0
**最后更新**: 2025-10-08
**适用岗位**: 后端开发工程师、全栈工程师、架构师


