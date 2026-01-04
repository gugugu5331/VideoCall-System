# 🏗️ 后端服务架构详解

## 📊 系统架构总览

```mermaid
graph TB
    subgraph Client["🖥️ 客户端层"]
        Qt6["Qt6 桌面客户端"]
        Web["Web 浏览器"]
        Mobile["移动端"]
    end

    subgraph Gateway["🌐 网关层"]
        Nginx["Nginx 负载均衡<br/>8800/443"]
        APIGateway["API 网关<br/>路由/限流/认证"]
    end

    subgraph Services["🎯 微服务层"]
        UserSvc["👤 用户服务<br/>:8080 / gRPC:50051<br/>认证/授权/用户管理"]
        MeetingSvc["📞 会议服务<br/>:8082 / gRPC:50052<br/>会议管理/参与者"]
        SignalSvc["📡 信令服务<br/>:8081<br/>WebSocket/媒体协商"]
        MediaSvc["🎬 媒体服务<br/>:8083<br/>SFU转发/录制"]
        AISvc["🤖 AI推理服务<br/>:8085 / gRPC:9085<br/>AI分析请求"]
        NotifySvc["🔔 通知服务<br/>:8085<br/>邮件/短信/推送"]
    end

    subgraph SharedLayer["🔧 共享层"]
        Config["配置管理"]
        Logger["日志系统"]
        Metrics["指标收集"]
        Tracing["链路追踪"]
        Discovery["服务发现"]
        Queue["消息队列"]
        Storage["存储管理"]
    end

    subgraph AILayer["🤖 AI推理层"]
        AIInference["Triton Inference Server<br/>HTTP:8000 / gRPC:8001<br/>TensorRT/CUDA"]
    end

    subgraph DataLayer["💾 数据层"]
        PostgreSQL["🗄️ PostgreSQL<br/>用户/会议/参与者"]
        Redis["⚡ Redis<br/>缓存/队列/Session"]
        MongoDB["📊 MongoDB<br/>AI结果/分析数据"]
        MinIO["📦 MinIO<br/>录制/媒体文件"]
        Etcd["🔧 etcd<br/>配置/服务发现"]
    end

    subgraph Observability["📊 可观测性"]
        Prometheus["Prometheus<br/>监控指标"]
        Grafana["Grafana<br/>可视化"]
        Jaeger["Jaeger<br/>链路追踪"]
        Loki["Loki<br/>日志聚合"]
    end

    Client -->|HTTP/WebSocket| Nginx
    Nginx --> APIGateway
    
    APIGateway -->|HTTP| UserSvc
    APIGateway -->|HTTP| MeetingSvc
    APIGateway -->|WebSocket| SignalSvc
    APIGateway -->|HTTP| MediaSvc
    APIGateway -->|HTTP| AISvc
    APIGateway -->|HTTP| NotifySvc

    UserSvc -.->|gRPC| MeetingSvc
    MeetingSvc -.->|gRPC| SignalSvc
    SignalSvc -.->|gRPC| MediaSvc
    MediaSvc -.->|gRPC| AISvc
    AISvc -.->|HTTP/gRPC| AIInference

    UserSvc --> SharedLayer
    MeetingSvc --> SharedLayer
    SignalSvc --> SharedLayer
    MediaSvc --> SharedLayer
    AISvc --> SharedLayer
    NotifySvc --> SharedLayer

    UserSvc -->|SQL| PostgreSQL
    MeetingSvc -->|SQL| PostgreSQL
    SignalSvc -->|Redis| Redis
    MediaSvc -->|SQL| PostgreSQL
    AISvc -->|NoSQL| MongoDB
    NotifySvc -->|Redis| Redis

    PostgreSQL -.->|缓存| Redis
    MongoDB -.->|存储| MinIO
    UserSvc -.->|配置| Etcd

    UserSvc -.->|metrics| Prometheus
    MeetingSvc -.->|metrics| Prometheus
    SignalSvc -.->|metrics| Prometheus
    MediaSvc -.->|metrics| Prometheus
    AISvc -.->|metrics| Prometheus
    NotifySvc -.->|metrics| Prometheus

    Prometheus --> Grafana
    
    UserSvc -.->|traces| Jaeger
    MeetingSvc -.->|traces| Jaeger
    SignalSvc -.->|traces| Jaeger
    MediaSvc -.->|traces| Jaeger
    AISvc -.->|traces| Jaeger
    NotifySvc -.->|traces| Jaeger

    UserSvc -.->|logs| Loki
    MeetingSvc -.->|logs| Loki
    SignalSvc -.->|logs| Loki
    MediaSvc -.->|logs| Loki
    AISvc -.->|logs| Loki
    NotifySvc -.->|logs| Loki

    classDef client fill:#e1f5ff,stroke:#01579b,stroke-width:2px
    classDef gateway fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef service fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef shared fill:#f0f4c3,stroke:#827717,stroke-width:2px
    classDef ai fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px
    classDef data fill:#fce4ec,stroke:#880e4f,stroke-width:2px
    classDef obs fill:#f1f8e9,stroke:#33691e,stroke-width:2px

    class Qt6,Web,Mobile client
    class Nginx,APIGateway gateway
    class UserSvc,MeetingSvc,SignalSvc,MediaSvc,AISvc,NotifySvc service
    class Config,Logger,Metrics,Tracing,Discovery,Queue,Storage shared
    class AIInference ai
    class PostgreSQL,Redis,MongoDB,MinIO,Etcd data
    class Prometheus,Grafana,Jaeger,Loki obs
```

---

## 🎯 微服务详解

### 1️⃣ 用户服务 (User Service)

**端口**: 8080 (HTTP) / 50051 (gRPC)

**职责**:
- 用户注册、登录、认证
- JWT Token 生成和验证
- 用户资料管理
- 权限控制和授权
- 用户角色管理

**依赖**:
- PostgreSQL: 用户数据存储
- Redis: Session 缓存、Token 黑名单
- etcd: 服务发现、配置管理

**通信方式**:
- HTTP REST API (客户端)
- gRPC (服务间通信)

**关键接口**:
```
POST   /api/v1/auth/register      # 用户注册
POST   /api/v1/auth/login         # 用户登录
POST   /api/v1/auth/logout        # 用户登出
GET    /api/v1/users/:id          # 获取用户信息
PUT    /api/v1/users/:id          # 更新用户信息
POST   /api/v1/auth/refresh       # 刷新Token
```

---

### 2️⃣ 会议服务 (Meeting Service)

**端口**: 8082 (HTTP) / 50052 (gRPC)

**职责**:
- 会议创建、更新、删除
- 会议参与者管理
- 会议权限控制
- 会议状态管理
- 参与者邀请

**依赖**:
- PostgreSQL: 会议数据存储
- Redis: 会议状态缓存
- etcd: 服务发现
- gRPC: 与用户服务通信

**通信方式**:
- HTTP REST API (客户端)
- gRPC (服务间通信)

**关键接口**:
```
POST   /api/v1/meetings           # 创建会议
GET    /api/v1/meetings/:id       # 获取会议信息
PUT    /api/v1/meetings/:id       # 更新会议
DELETE /api/v1/meetings/:id       # 删除会议
POST   /api/v1/meetings/:id/join  # 加入会议
POST   /api/v1/meetings/:id/leave # 离开会议
```

---

### 3️⃣ 信令服务 (Signaling Service)

**端口**: 8081 (HTTP/WebSocket)

**职责**:
- WebSocket 连接管理
- 媒体协商 (SDP/ICE)
- 房间管理
- 消息转发
- 连接状态管理

**依赖**:
- Redis: 房间状态、消息队列
- etcd: 服务发现
- gRPC: 与其他服务通信

**通信方式**:
- WebSocket (客户端实时通信)
- gRPC (服务间通信)

**WebSocket 消息类型**:
```
join_room          # 加入房间
leave_room         # 离开房间
offer              # WebRTC Offer
answer             # WebRTC Answer
ice_candidate      # ICE 候选
```

---

### 4️⃣ 媒体服务 (Media Service)

**端口**: 8083 (HTTP)

**职责**:
- SFU 媒体转发
- 会议录制
- 媒体处理 (FFmpeg)
- 媒体统计
- 录制文件管理

**依赖**:
- PostgreSQL: 录制元数据
- MinIO: 录制文件存储
- FFmpeg: 媒体处理
- gRPC: 与其他服务通信

**通信方式**:
- HTTP REST API
- gRPC (服务间通信)
- WebRTC (媒体传输)

**关键接口**:
```
POST   /api/v1/recordings         # 开始录制
POST   /api/v1/recordings/:id/stop # 停止录制
GET    /api/v1/recordings/:id     # 获取录制信息
GET    /api/v1/media/stats        # 获取媒体统计
```

---

### 5️⃣ AI 服务 (AI Service)

**端口**: 8084 (HTTP) / 9084 (gRPC)

**职责**:
- AI 分析请求处理
- 模型管理
- 推理结果存储
- 节点健康检查
- 负载均衡

**依赖**:
- MongoDB: AI 结果存储
- Redis: 缓存、队列
- PostgreSQL: 配置存储
- AI Inference Service: 推理执行（HTTP/gRPC）

**通信方式**:
- HTTP REST API
- gRPC (服务间通信)
- HTTP/gRPC (与 AI 推理服务通信)

**支持的 AI 功能**:
- 语音识别 (ASR)
- 情感检测
- 合成检测 (Deepfake)
- 音频降噪
- 视频增强

---

### 6️⃣ 通知服务 (Notification Service)

**端口**: 8085 (HTTP)

**职责**:
- 邮件发送
- 短信发送
- 推送通知
- 通知队列管理
- 通知历史记录

**依赖**:
- Redis: 消息队列
- PostgreSQL: 通知历史
- 第三方服务: 邮件、短信、推送

**通信方式**:
- HTTP REST API
- 消息队列 (Redis)

---

## 🔧 共享层 (Shared Layer)

所有微服务共享的通用功能:

| 模块 | 功能 |
|------|------|
| **config** | 配置管理、环境变量处理 |
| **logger** | 日志记录、日志级别控制 |
| **database** | 数据库连接、连接池管理 |
| **grpc** | gRPC 客户端、服务器、拦截器 |
| **metrics** | Prometheus 指标收集 |
| **tracing** | Jaeger 链路追踪 |
| **middleware** | HTTP 中间件、CORS、认证 |
| **models** | 数据模型定义 |
| **queue** | 消息队列、Redis 操作 |
| **storage** | 文件存储、MinIO 操作 |
| **discovery** | 服务发现、etcd 操作 |

---

## 📊 数据流示例

### 用户加入会议流程

```
1. 客户端 → 用户服务: 登录请求
2. 用户服务 → PostgreSQL: 验证用户
3. 用户服务 → Redis: 存储 Session
4. 用户服务 → 客户端: 返回 JWT Token

5. 客户端 → 会议服务: 加入会议请求
6. 会议服务 → PostgreSQL: 查询会议信息
7. 会议服务 → Redis: 更新会议状态
8. 会议服务 → 信令服务: 通知新用户加入

9. 客户端 → 信令服务: WebSocket 连接
10. 信令服务 → Redis: 存储房间状态
11. 信令服务 → 客户端: 返回房间信息

12. 客户端 → 媒体服务: WebRTC 连接
13. 媒体服务 → PostgreSQL: 记录媒体流
14. 媒体服务 → 其他客户端: 转发媒体流
```

---

## 🔄 服务间通信

### gRPC 通信

用于服务间的同步通信:
- 用户服务 ↔ 会议服务
- 会议服务 ↔ 信令服务
- 媒体服务 ↔ AI 服务

### HTTP/gRPC 通信

用于 AI 服务与 AI 推理服务的同步通信:
- 请求/应答模式
- gRPC 流式音频

### Redis 消息队列

用于异步任务处理:
- 通知队列
- 媒体处理队列
- 日志队列

---

## 📈 可观测性

### Prometheus 指标

每个服务收集:
- HTTP 请求数、延迟、错误率
- gRPC 请求数、延迟、错误率
- 数据库连接数、查询时间
- 缓存命中率

### Jaeger 链路追踪

追踪完整的请求链路:
- 跨服务调用
- 数据库查询
- 缓存操作

### Loki 日志聚合

收集所有服务的日志:
- 应用日志
- 错误日志
- 审计日志

---

## � 服务交互流程图

### 会议创建流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant Gateway as API网关
    participant UserSvc as 用户服务
    participant MeetingSvc as 会议服务
    participant DB as PostgreSQL
    participant Redis as Redis
    participant Etcd as etcd

    Client->>Gateway: POST /meetings (JWT Token)
    Gateway->>UserSvc: 验证Token
    UserSvc->>Redis: 查询Session
    Redis-->>UserSvc: Session数据
    UserSvc-->>Gateway: Token有效

    Gateway->>MeetingSvc: 创建会议请求
    MeetingSvc->>DB: 插入会议记录
    DB-->>MeetingSvc: 会议ID
    MeetingSvc->>Redis: 缓存会议信息
    Redis-->>MeetingSvc: OK
    MeetingSvc->>Etcd: 注册会议
    Etcd-->>MeetingSvc: OK

    MeetingSvc-->>Gateway: 会议创建成功
    Gateway-->>Client: 返回会议信息
```

### 用户加入会议流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant SignalSvc as 信令服务
    participant MeetingSvc as 会议服务
    participant MediaSvc as 媒体服务
    participant Redis as Redis
    participant DB as PostgreSQL

    Client->>SignalSvc: WebSocket连接
    SignalSvc->>Redis: 创建房间
    Redis-->>SignalSvc: 房间ID

    Client->>SignalSvc: 加入房间请求
    SignalSvc->>MeetingSvc: 验证权限(gRPC)
    MeetingSvc->>DB: 查询会议
    DB-->>MeetingSvc: 会议信息
    MeetingSvc-->>SignalSvc: 权限验证通过

    SignalSvc->>Redis: 更新房间成员
    Redis-->>SignalSvc: OK
    SignalSvc->>MediaSvc: 通知新成员(gRPC)
    MediaSvc->>DB: 记录媒体流
    DB-->>MediaSvc: OK

    SignalSvc-->>Client: 加入成功
    SignalSvc->>Client: 广播新成员加入
```

### AI分析请求流程

```mermaid
sequenceDiagram
    participant MediaSvc as 媒体服务
    participant AISvc as AI推理服务(API)
    participant AIInference as Triton 推理服务
    participant MongoDB as MongoDB

    MediaSvc->>AISvc: 发送分析请求(gRPC)
    AISvc->>MongoDB: 查询模型配置
    MongoDB-->>AISvc: 模型信息

AISvc->>AIInference: 发起推理(HTTP/gRPC)
AIInference-->>AISvc: 推理结果

    AISvc->>MongoDB: 存储分析结果
    MongoDB-->>AISvc: OK
    AISvc-->>MediaSvc: 返回分析结果
```

---

## �🚀 部署架构

```
Docker Compose 编排:
├── user-service (容器)
├── meeting-service (容器)
├── signaling-service (容器)
├── media-service (容器)
├── ai-inference-service (容器)
├── notification-service (容器)
├── PostgreSQL (容器)
├── Redis (容器)
├── MongoDB (容器)
├── MinIO (容器)
├── etcd (容器)
├── Nginx (容器)
├── Prometheus (容器)
├── Grafana (容器)
├── Jaeger (容器)
└── Loki (容器)
```

所有服务通过 Docker 网络互联，支持水平扩展。
