# AI Inference Service

AI 推理微服务，集成 Edge-LLM-Infra 框架，为 meeting-system 提供 AI 推理能力。

## 功能特性

### 支持的 AI 功能

1. **ASR (Automatic Speech Recognition)** - 语音识别
   - 将音频转换为文本
   - 支持多种音频格式
   - 返回识别文本和置信度

2. **Emotion Detection** - 情感检测
   - 分析文本的情感倾向
   - 返回主要情感和所有情感分数
   - 支持多种情感类别

3. **Synthesis Detection** - 深度伪造检测
   - 检测音频是否为 AI 合成
   - 返回合成概率和置信度
   - 用于音频真实性验证

### 架构特点

- **RESTful API**: 提供用户友好的 HTTP 接口
- **Edge-LLM-Infra 集成**: 通过 TCP 连接到 unit-manager
- **完整的请求流程**: 自动处理 setup → inference → exit 流程
- **资源管理**: 确保每次请求后正确释放资源
- **错误处理**: 完整的错误处理和超时机制
- **服务注册**: 集成 etcd 服务发现
- **消息队列**: 支持 Redis 消息队列和发布订阅
- **分布式追踪**: 集成 Jaeger 追踪
- **监控指标**: 提供 Prometheus 指标

## API 文档

### 基础信息

- **Base URL**: `http://localhost:8085`
- **Content-Type**: `application/json`

### 端点列表

#### 1. 健康检查

```http
GET /health
GET /api/v1/ai/health
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": "healthy",
    "service": "ai-inference-service",
    "timestamp": 1696000000
  }
}
```

#### 2. 服务信息

```http
GET /api/v1/ai/info
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "service": "ai-inference-service",
    "version": "1.0.0",
    "capabilities": ["speech_recognition", "emotion_detection", "synthesis_detection"],
    "models": {
      "asr": "asr-model",
      "emotion": "emotion-model",
      "synthesis": "synthesis-model"
    }
  }
}
```

#### 3. 语音识别 (ASR)

```http
POST /api/v1/ai/asr
```

**请求体**:
```json
{
  "audio_data": "base64_encoded_audio_data",
  "format": "wav",
  "sample_rate": 16000,
  "language": "en"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "text": "Transcribed text from audio",
    "confidence": 0.95,
    "language": "en",
    "duration_ms": 125.5
  }
}
```

#### 4. 情感检测

```http
POST /api/v1/ai/emotion
```

**请求体**:
```json
{
  "text": "I am very happy today!"
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "emotion": "happy",
    "confidence": 0.92,
    "emotions": {
      "happy": 0.92,
      "neutral": 0.05,
      "sad": 0.02,
      "angry": 0.01
    },
    "duration_ms": 45.2
  }
}
```

#### 5. 深度伪造检测

```http
POST /api/v1/ai/synthesis
```

**请求体**:
```json
{
  "audio_data": "base64_encoded_audio_data",
  "format": "wav",
  "sample_rate": 16000
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "is_synthetic": false,
    "confidence": 0.88,
    "score": 0.15,
    "duration_ms": 98.3
  }
}
```

#### 6. 批量推理

```http
POST /api/v1/ai/batch
```

**请求体**:
```json
{
  "tasks": [
    {
      "task_id": "task_1",
      "type": "asr",
      "data": {
        "audio_data": "base64_encoded_audio",
        "format": "wav",
        "sample_rate": 16000
      }
    },
    {
      "task_id": "task_2",
      "type": "emotion",
      "data": {
        "text": "Sample text for emotion detection"
      }
    }
  ]
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "results": [
      {
        "task_id": "task_1",
        "type": "asr",
        "result": { "text": "...", "confidence": 0.95 }
      },
      {
        "task_id": "task_2",
        "type": "emotion",
        "result": { "emotion": "happy", "confidence": 0.92 }
      }
    ],
    "total": 2
  }
}
```

## 部署指南

### 前置条件

1. **Edge-LLM-Infra 框架**
   - unit-manager 运行在 `localhost:19001`
   - llm 节点已启动并注册

2. **基础设施**
   - PostgreSQL (可选)
   - Redis
   - Etcd
   - Jaeger (可选)

### 本地开发

1. **安装依赖**:
```bash
cd meeting-system/backend/ai-inference-service
go mod download
```

2. **配置文件**:
编辑 `config/ai-inference-service.yaml`，确保 `zmq.unit_manager_host` 和 `zmq.unit_manager_port` 正确。

3. **启动服务**:
```bash
go run main.go --config config/ai-inference-service.yaml
```

4. **测试服务**:
```bash
# 健康检查
curl http://localhost:8085/health

# 运行测试脚本
python3 test_ai_service.py --host localhost --port 8085
```

### Docker 部署

1. **构建镜像**:
```bash
docker build -t ai-inference-service:latest .
```

2. **运行容器**:
```bash
docker run -d \
  --name ai-inference-service \
  --network meeting-system-network \
  -p 8085:8085 \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  ai-inference-service:latest
```

### Docker Compose 部署

在 `docker-compose.yml` 中添加：

```yaml
ai-inference-service:
  build:
    context: ./backend/ai-inference-service
    dockerfile: Dockerfile
  container_name: meeting-ai-inference-service
  ports:
    - "8085:8085"
  environment:
    - SERVICE_ADVERTISE_HOST=ai-inference-service
  volumes:
    - ./backend/ai-inference-service/config:/app/config
    - ./backend/ai-inference-service/logs:/app/logs
  networks:
    - meeting-system-network
  depends_on:
    - postgres
    - redis
    - etcd
  restart: unless-stopped
```

启动服务：
```bash
docker-compose up -d ai-inference-service
```

## 测试

### 单元测试

```bash
go test ./...
```

### 集成测试

```bash
# 确保 unit-manager 和 llm 节点正在运行
cd /root/meeting-system-server/meeting-system/Edge-LLM-Infra-master/unit-manager/build
./unit_manager &

cd /root/meeting-system-server/meeting-system/Edge-LLM-Infra-master/node/llm/build
./llm &

# 运行 AI 服务测试
cd /root/meeting-system-server/meeting-system/backend/ai-inference-service
python3 test_ai_service.py
```

### 压力测试

```bash
# 使用 Apache Bench
ab -n 1000 -c 10 -p test_data.json -T application/json \
  http://localhost:8085/api/v1/ai/asr

# 使用 wrk
wrk -t4 -c100 -d30s --latency \
  -s test_script.lua \
  http://localhost:8085/api/v1/ai/asr
```

## 监控和日志

### 日志

日志文件位置：`logs/ai-inference-service.log`

查看日志：
```bash
tail -f logs/ai-inference-service.log
```

### Prometheus 指标

访问：`http://localhost:8085/metrics`

主要指标：
- `http_requests_total` - HTTP 请求总数
- `http_request_duration_seconds` - 请求延迟
- `ai_inference_requests_total` - AI 推理请求总数
- `ai_inference_duration_seconds` - AI 推理延迟

### Jaeger 追踪

访问 Jaeger UI：`http://localhost:16686`

搜索服务：`ai-inference-service`

## 故障排查

### 常见问题

1. **无法连接到 unit-manager**
   - 检查 unit-manager 是否运行：`netstat -tlnp | grep 19001`
   - 检查配置文件中的 `zmq.unit_manager_host` 和 `zmq.unit_manager_port`
   - 检查防火墙设置

2. **推理请求超时**
   - 检查 llm 节点是否正常运行
   - 增加 `zmq.timeout` 配置
   - 查看 llm 节点日志

3. **资源未释放**
   - 检查 exit 请求是否成功
   - 查看 unit-manager 日志中的 "release work_id success" 消息
   - 重启 unit-manager 和 llm 节点

## 开发指南

### 添加新的 AI 功能

1. 在 `services/ai_inference_service.go` 中添加新方法
2. 在 `handlers/ai_handler.go` 中添加新的 HTTP 处理器
3. 在 `main.go` 的 `setupRoutes` 中注册新路由
4. 更新 API 文档

### 代码结构

```
ai-inference-service/
├── config/                 # 配置文件
├── handlers/               # HTTP 处理器
│   └── ai_handler.go
├── services/               # 业务逻辑
│   ├── edge_llm_client.go  # Edge-LLM-Infra 客户端
│   └── ai_inference_service.go
├── models/                 # 数据模型（可选）
├── main.go                 # 主程序
├── Dockerfile              # Docker 配置
├── go.mod                  # Go 模块
└── README.md               # 文档
```

## 许可证

MIT License

