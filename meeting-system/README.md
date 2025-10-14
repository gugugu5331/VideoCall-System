# 🎥 Meeting System Backend - 后端服务文档

## 📋 目录

- [系统概述](#-系统概述)
- [微服务架构](#-微服务架构)
- [技术栈](#-技术栈)
- [快速开始](#-快速开始)
- [服务详解](#-服务详解)
- [数据库设计](#-数据库设计)
- [API 文档](#-api-文档)
- [配置说明](#-配置说明)
- [部署指南](#-部署指南)

---

## 📖 系统概述

Meeting System Backend 是一个基于 Go 语言的微服务架构视频会议系统后端，采用 SFU (Selective Forwarding Unit) 媒体转发架构，集成 Edge-LLM-Infra 分布式 AI 推理框架。

**核心特性：**
- 🏗️ **微服务架构**: 5个独立的 Go 微服务 + AI 推理服务
- 🔐 **安全认证**: JWT + CSRF 保护 + 限流
- 📡 **实时通信**: WebSocket 信令 + WebRTC 媒体传输
- 🤖 **AI 集成**: ZeroMQ 连接 Edge-LLM-Infra
- 📊 **完整监控**: Prometheus + Jaeger + Loki
- 🔄 **服务发现**: etcd 服务注册与发现
- 🐳 **容器化**: Docker Compose 一键部署

---

## 🏗️ 微服务架构

### 服务组件

```
┌─────────────────────────────────────────────────────────────┐
│                      Nginx API Gateway                       │
│                         (:8800)                              │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
┌───────▼────────┐  ┌────────▼────────┐  ┌────────▼────────┐
│  User Service  │  │Meeting Service  │  │Signaling Service│
│     :8080      │  │     :8082       │  │     :8081       │
│                │  │                 │  │                 │
│ - 用户注册登录  │  │ - 会议管理      │  │ - WebSocket     │
│ - JWT 认证     │  │ - 参与者管理    │  │ - 信令转发      │
│ - 用户资料     │  │ - 会议状态      │  │ - 房间管理      │
└────────────────┘  └─────────────────┘  └─────────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
┌───────▼────────┐  ┌────────▼────────┐  ┌────────▼────────┐
│  Media Service │  │   AI Service    │  │AI Infer Service │
│     :8083      │  │     :8084       │  │     :8085       │
│                │  │                 │  │                 │
│ - SFU 转发     │  │ - AI 分析       │  │ - 模型推理      │
│ - 媒体录制     │  │ - 结果存储      │  │ - ZMQ 通信      │
│ - MinIO 存储   │  │ - MongoDB       │  │ - Unit Manager  │
└────────────────┘  └─────────────────┘  └─────────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Service Registry (etcd)                   │
│                    Message Queue (Redis)                     │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      Data Layer                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │PostgreSQL│  │  Redis   │  │ MongoDB  │  │  MinIO   │   │
│  │  :5432   │  │  :6379   │  │ :27017   │  │  :9000   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 服务职责

| 服务 | 端口 | 职责 | 依赖 |
|------|------|------|------|
| **user-service** | 8080 | 用户认证、资料管理、权限控制 | PostgreSQL, Redis, etcd |
| **meeting-service** | 8082 | 会议创建、管理、参与者控制 | PostgreSQL, Redis, etcd |
| **signaling-service** | 8081 | WebSocket 信令、房间管理 | Redis, etcd |
| **media-service** | 8083 | SFU 媒体转发、录制、存储 | PostgreSQL, MinIO |
| **ai-service** | 8084 | AI 分析请求、结果管理 | MongoDB, ZMQ |
| **ai-inference-service** | 8085 | AI 模型推理、ZMQ 通信 | PostgreSQL, Redis, ZMQ |

---

## 🛠️ 技术栈

### 核心框架
| 技术 | 版本 | 用途 |
|------|------|------|
| **Go** | 1.24.0+ | 主要开发语言 |
| **Gin** | 1.9.1 | HTTP Web 框架 |
| **GORM** | 1.31.0 | ORM 数据库框架 |
| **gRPC** | 1.75.1 | 服务间 RPC 通信 |

### 通信协议
| 技术 | 版本 | 用途 |
|------|------|------|
| **WebSocket** | gorilla/websocket 1.5.3 | 实时信令通信 |
| **ZeroMQ** | pebbe/zmq4 1.4.0 | AI 服务高性能通信 |
| **HTTP/2** | - | RESTful API |

### 数据存储
| 技术 | 版本 | 用途 |
|------|------|------|
| **PostgreSQL** | 15-alpine | 用户数据、会议数据 |
| **Redis** | 7-alpine | 缓存、消息队列、会话 |
| **MongoDB** | 6.0.14 | AI 分析结果存储 |
| **MinIO** | latest | 对象存储（录制文件） |

### 基础设施
| 技术 | 版本 | 用途 |
|------|------|------|
| **etcd** | 3.6.5 | 服务注册与发现 |
| **Nginx** | alpine | API 网关、反向代理 |
| **Docker** | 20.0+ | 容器化部署 |

### 监控与追踪
| 技术 | 版本 | 用途 |
|------|------|------|
| **Prometheus** | 2.48.0 | 指标收集 |
| **Jaeger** | 1.51 | 分布式追踪 |
| **Grafana** | 10.2.2 | 可视化面板 |
| **Loki** | 2.9.3 | 日志聚合 |

---

## 🚀 快速开始

### 环境要求

- **Docker**: 20.0+
- **Docker Compose**: 2.0+
- **Go**: 1.24.0+ (本地开发)
- **Make**: (可选)

### 一键启动（Docker Compose）

```bash
# 1. 进入项目目录
cd meeting-system

# 2. 启动所有服务
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f user-service
```

### 本地开发启动

```bash
# 1. 启动基础设施服务
docker-compose up -d postgres redis mongodb minio etcd jaeger

# 2. 编译并启动用户服务
cd backend/user-service
go build -o user-service
./user-service -config=../config/config.yaml

# 3. 启动其他服务
cd ../meeting-service
go run main.go -config=../config/meeting-service.yaml

cd ../signaling-service
go run main.go -config=../config/signaling-service.yaml

# 或使用脚本启动所有服务
cd ../scripts
./start_all_services.sh
```

### 验证服务

```bash
# 检查用户服务健康状态
curl http://localhost:8080/health

# 检查会议服务
curl http://localhost:8082/health

# 检查信令服务
curl http://localhost:8081/health

# 查看 Prometheus 指标
curl http://localhost:8080/metrics
```

---

## 🔍 服务详解

### 1. User Service (用户服务)

**端口**: 8080
**职责**: 用户认证、资料管理、权限控制

**主要功能**:
- ✅ 用户注册与登录
- ✅ JWT Token 生成与验证
- ✅ CSRF 保护
- ✅ 用户资料 CRUD
- ✅ 头像上传
- ✅ 密码修改
- ✅ 用户封禁/解封（管理员）
- ✅ 请求限流

**技术实现**:
- Gin Web 框架
- GORM ORM
- JWT 认证 (golang-jwt/jwt v5)
- PostgreSQL 用户数据存储
- Redis 会话缓存
- etcd 服务注册

**API 端点**:
```
POST   /api/v1/register          # 用户注册
POST   /api/v1/login             # 用户登录
POST   /api/v1/refresh-token     # 刷新 Token
GET    /api/v1/profile           # 获取用户资料
PUT    /api/v1/profile           # 更新用户资料
POST   /api/v1/change-password   # 修改密码
POST   /api/v1/upload-avatar     # 上传头像
DELETE /api/v1/account           # 删除账户
GET    /api/v1/admin/users       # 管理员：用户列表
```

**配置文件**: `backend/config/config.yaml`

---

### 2. Meeting Service (会议服务)

**端口**: 8082
**职责**: 会议管理、参与者控制

**主要功能**:
- ✅ 会议创建/删除
- ✅ 会议列表查询
- ✅ 参与者加入/离开
- ✅ 参与者管理（踢出、静音）
- ✅ 会议状态管理
- ✅ 会议权限控制

**技术实现**:
- Gin Web 框架
- GORM ORM
- PostgreSQL 会议数据存储
- Redis 会议状态缓存
- gRPC 服务间通信
- etcd 服务注册

**API 端点**:
```
POST   /api/v1/meetings                    # 创建会议
GET    /api/v1/meetings                    # 获取会议列表
GET    /api/v1/meetings/:id                # 获取会议详情
PUT    /api/v1/meetings/:id                # 更新会议
DELETE /api/v1/meetings/:id                # 删除会议
POST   /api/v1/meetings/:id/join           # 加入会议
POST   /api/v1/meetings/:id/leave          # 离开会议
GET    /api/v1/meetings/:id/participants   # 参与者列表
POST   /api/v1/meetings/:id/participants/:uid/kick  # 踢出参与者
```

**配置文件**: `backend/config/meeting-service.yaml`

---

### 3. Signaling Service (信令服务)

**端口**: 8081
**职责**: WebSocket 信令、房间管理

**主要功能**:
- ✅ WebSocket 连接管理
- ✅ 信令消息转发（offer/answer/candidate）
- ✅ 房间状态管理
- ✅ 客户端心跳检测
- ✅ 连接统计

**技术实现**:
- Gin Web 框架
- gorilla/websocket
- Redis Pub/Sub 消息分发
- 内存房间管理
- etcd 服务注册

**WebSocket 协议**:
```json
// 客户端 -> 服务器
{
  "type": "join",
  "room_id": "meeting-123",
  "user_id": "user-456"
}

{
  "type": "offer",
  "target": "user-789",
  "sdp": "..."
}

{
  "type": "candidate",
  "target": "user-789",
  "candidate": "..."
}

// 服务器 -> 客户端
{
  "type": "user-joined",
  "user_id": "user-789",
  "user_info": {...}
}

{
  "type": "offer",
  "from": "user-456",
  "sdp": "..."
}
```

**API 端点**:
```
GET    /ws/signaling             # WebSocket 连接
GET    /api/v1/stats             # 统计信息
GET    /api/v1/rooms/stats       # 房间统计
```

**配置文件**: `backend/config/signaling-service.yaml`

---

### 4. Media Service (媒体服务)

**端口**: 8083
**职责**: SFU 媒体转发、录制、存储

**主要功能**:
- ✅ 媒体文件上传/下载
- ✅ 会议录制
- ✅ MinIO 对象存储集成
- ✅ 录制文件管理

**技术实现**:
- Gin Web 框架
- MinIO Go SDK
- PostgreSQL 媒体元数据
- FFmpeg 媒体处理（计划）

**API 端点**:
```
POST   /api/v1/media/upload      # 上传媒体文件
GET    /api/v1/media/:id         # 获取媒体文件
DELETE /api/v1/media/:id         # 删除媒体文件
GET    /api/v1/recordings        # 录制列表
POST   /api/v1/recordings/start  # 开始录制
POST   /api/v1/recordings/stop   # 停止录制
```

**配置文件**: `backend/config/media-service.yaml`

---

### 5. AI Service (AI 服务)

**端口**: 8084
**职责**: AI 分析请求、结果管理

**主要功能**:
- ✅ AI 分析任务提交
- ✅ 分析结果查询
- ✅ MongoDB 结果存储

**技术实现**:
- Gin Web 框架
- MongoDB Go Driver
- ZMQ 通信（与 AI Inference Service）

**API 端点**:
```
POST   /api/v1/ai/analyze        # 提交分析任务
GET    /api/v1/ai/results/:id    # 获取分析结果
```

**配置文件**: `backend/config/ai-service.yaml`

---

### 6. AI Inference Service (AI 推理服务)

**端口**: 8085
**职责**: AI 模型推理、ZMQ 通信

**主要功能**:
- ✅ 推理任务调度
- ✅ ZMQ 连接 Unit Manager
- ✅ 模型列表查询
- ✅ 推理结果返回

**技术实现**:
- Gin Web 框架
- ZeroMQ (pebbe/zmq4)
- 连接宿主机 Unit Manager (:19001)

**API 端点**:
```
POST   /api/v1/inference/submit  # 提交推理任务
GET    /api/v1/inference/:id     # 获取推理结果
GET    /api/v1/models            # 可用模型列表
```

**配置文件**: `backend/config/ai-inference-service.yaml`

---

## 🗄️ 数据库设计

### PostgreSQL 表结构

#### users 表（用户表）
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(100),
    avatar_url VARCHAR(255),
    role VARCHAR(20) DEFAULT 'user',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### meetings 表（会议表）
```sql
CREATE TABLE meetings (
    id SERIAL PRIMARY KEY,
    meeting_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    creator_id INTEGER REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'scheduled',
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    max_participants INTEGER DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### participants 表（参与者表）
```sql
CREATE TABLE participants (
    id SERIAL PRIMARY KEY,
    meeting_id INTEGER REFERENCES meetings(id),
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'participant',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    UNIQUE(meeting_id, user_id)
);
```

### Redis 数据结构

```
# 用户会话
session:{user_id} -> {token, expires_at}

# 会议状态
meeting:{meeting_id}:status -> {active|ended}
meeting:{meeting_id}:participants -> Set{user_id1, user_id2, ...}

# 在线用户
online:users -> Set{user_id1, user_id2, ...}

# 限流
ratelimit:{user_id}:{endpoint} -> counter
```

### MongoDB 集合

```javascript
// AI 分析结果
{
  _id: ObjectId,
  task_id: "task-123",
  meeting_id: "meeting-456",
  user_id: "user-789",
  type: "emotion|transcription|quality",
  result: {...},
  created_at: ISODate
}
```

## 📝 API 文档

### 通用响应格式

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {...}
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "error message",
  "error": "detailed error"
}
```

### 认证方式

所有需要认证的接口都需要在 Header 中携带 JWT Token：

```
Authorization: Bearer <jwt_token>
```

### 用户服务 API

详见 [服务详解 - User Service](#1-user-service-用户服务)

### 会议服务 API

详见 [服务详解 - Meeting Service](#2-meeting-service-会议服务)

### 信令服务 API

详见 [服务详解 - Signaling Service](#3-signaling-service-信令服务)

---

## ⚙️ 配置说明

### 配置文件位置

所有配置文件位于 `backend/config/` 目录：

```
backend/config/
├── config.yaml                 # user-service 配置
├── meeting-service.yaml        # meeting-service 配置
├── signaling-service.yaml      # signaling-service 配置
├── media-service.yaml          # media-service 配置
├── ai-service.yaml             # ai-service 配置
└── ai-inference-service.yaml   # ai-inference-service 配置
```

### 配置文件示例 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # debug | release

database:
  host: "postgres"
  port: 5432
  user: "postgres"
  password: "password"
  dbname: "meeting_system"
  sslmode: "disable"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  host: "redis"
  port: 6379
  password: ""
  db: 0
  pool_size: 10

etcd:
  endpoints:
    - "etcd:2379"
  dial_timeout: 5

jwt:
  secret: "your-secret-key-change-in-production"
  expire_hours: 24
  refresh_expire_hours: 168

log:
  level: "info"
  filename: "logs/user-service.log"
  max_size: 100
  max_age: 30
  max_backups: 10
  compress: true

cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allowed_headers:
    - "Origin"
    - "Content-Type"
    - "Authorization"
```

### 环境变量

可以通过环境变量覆盖配置文件：

```bash
# 数据库配置
export DATABASE_HOST=postgres
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASSWORD=password

# Redis 配置
export REDIS_HOST=redis
export REDIS_PORT=6379

# JWT 配置
export JWT_SECRET=your-super-secret-key

# etcd 配置
export ETCD_ENDPOINTS=etcd:2379

# ZMQ 配置（AI 服务）
export ZMQ_UNIT_MANAGER_HOST=host.docker.internal
export ZMQ_UNIT_MANAGER_PORT=19001
```

---

## 🐳 部署指南

### Docker Compose 部署

**完整部署**（推荐）:
```bash
cd meeting-system
docker-compose up -d
```

**分步部署**:
```bash
# 1. 启动基础设施
docker-compose up -d postgres redis mongodb minio etcd

# 2. 启动监控服务
docker-compose up -d prometheus grafana jaeger loki promtail

# 3. 启动业务服务
docker-compose up -d user-service meeting-service signaling-service media-service

# 4. 启动 AI 服务
docker-compose up -d ai-service ai-inference-service

# 5. 启动网关
docker-compose up -d nginx
```

### 服务健康检查

```bash
# 检查所有服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f user-service

# 检查服务健康
curl http://localhost:8800/api/v1/health
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

---

## 🔧 开发指南

### 添加新的微服务

1. **创建服务目录**:
```bash
cd backend
mkdir new-service
cd new-service
```

2. **初始化 Go 模块**:
```bash
go mod init meeting-system/new-service
```

3. **创建 main.go**:
```go
package main

import (
    "github.com/gin-gonic/gin"
    "meeting-system/shared/config"
    "meeting-system/shared/logger"
)

func main() {
    config.InitConfig("../config/new-service.yaml")
    logger.InitLogger(...)

    r := gin.Default()
    r.GET("/health", healthCheck)
    r.Run(":8086")
}
```

4. **创建 Dockerfile**:
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o new-service

FROM alpine:latest
COPY --from=builder /app/new-service /app/
CMD ["/app/new-service"]
```

5. **添加到 docker-compose.yml**:
```yaml
new-service:
  build:
    context: ./backend
    dockerfile: new-service/Dockerfile
  container_name: meeting-new-service
  ports:
    - "8086:8086"
  networks:
    - meeting-network
```

### 共享库使用

所有微服务共享 `backend/shared/` 目录下的库：

```go
import (
    "meeting-system/shared/config"      // 配置管理
    "meeting-system/shared/database"    // 数据库连接
    "meeting-system/shared/logger"      // 日志工具
    "meeting-system/shared/middleware"  // Gin 中间件
    "meeting-system/shared/models"      // 数据模型
    "meeting-system/shared/discovery"   // 服务发现
    "meeting-system/shared/metrics"     // Prometheus 指标
    "meeting-system/shared/tracing"     // Jaeger 追踪
)
```

### 代码规范

- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 遵循 Go 官方代码规范
- 添加必要的注释和文档

---

## 🧪 测试

### 单元测试

```bash
cd backend/user-service
go test ./... -v
```

### 集成测试

```bash
cd meeting-system/scripts
./test_integration.sh
```

### E2E 测试

```bash
cd meeting-system/scripts
./run_e2e_test.sh
```

### 压力测试

```bash
cd backend/stress-test
go run main.go -config=../config/stress-test-config.yaml
```

---

## 📊 监控与日志

### Prometheus 指标

访问: http://localhost:8801

**可用指标**:
- `http_requests_total`: HTTP 请求总数
- `http_request_duration_seconds`: 请求延迟
- `grpc_server_handled_total`: gRPC 调用统计
- `db_connections`: 数据库连接数
- `active_users`: 在线用户数
- `active_meetings`: 活跃会议数

### Grafana 面板

访问: http://localhost:8804 (admin/admin123)

**预配置面板**:
1. 服务概览
2. 数据库性能
3. Redis 性能
4. 系统资源
5. 业务指标

### Jaeger 追踪

访问: http://localhost:8803

查看分布式调用链路和性能分析。

### Loki 日志

在 Grafana 中通过 Explore 查询日志：

```
{container_name="meeting-user-service"} |= "error"
```

---

## 🔗 相关链接

- [项目主页](https://github.com/gugugu5331/VideoCall-System)
- [Qt6 客户端文档](../qt6-client/README.md)
- [部署文档](docs/deployment/)
- [测试文档](docs/testing/)
- [Edge-LLM-Infra](Edge-LLM-Infra-master/)

---

## 📄 许可证

MIT License

---

**最后更新**: 2025-10-08
