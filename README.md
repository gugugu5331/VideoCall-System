# 多人视频会议系统 - 带伪造音视频检测

## 项目概述

这是一个基于微服务架构的多人视频会议系统，具备伪造音视频检测功能。系统采用Go语言开发后端服务，Qt开发跨平台客户端，使用WebRTC进行实时音视频传输，FFmpeg进行音视频处理。

## 技术栈

### 后端
- **语言**: Go 1.21+
- **框架**: Gin (HTTP), gRPC (微服务通信)
- **数据库**: PostgreSQL (主数据库), Redis (缓存), MongoDB (日志存储)
- **消息队列**: RabbitMQ
- **服务发现**: Consul
- **负载均衡**: Nginx
- **容器化**: Docker + Docker Compose

### 前端
- **框架**: Qt 6.5+ (C++)
- **音视频**: WebRTC, FFmpeg
- **UI**: QML + C++

### AI检测
- **深度学习**: TensorFlow/PyTorch
- **模型**: FaceSwap检测、语音合成检测
- **部署**: TensorFlow Serving

## 快速开始

### 1. 环境准备

确保您的Windows 11系统已安装以下软件：

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Git](https://git-scm.com/downloads)
- [Go 1.21+](https://golang.org/dl/) (可选，用于本地开发)
- [Qt 6.5+](https://www.qt.io/download) (可选，用于前端开发)

### 2. 克隆项目

```bash
git clone https://github.com/your-repo/video-conference-system.git
cd video-conference-system
```

### 3. 一键部署

```bash
# 给脚本执行权限
chmod +x scripts/deploy.sh

# 运行部署脚本
./scripts/deploy.sh
```

### 4. 验证部署

```bash
# 运行测试脚本
chmod +x scripts/test.sh
./scripts/test.sh
```

### 5. 访问系统

部署完成后，您可以通过以下地址访问系统：

- **Web界面**: http://localhost
- **API文档**: http://localhost/api/docs
- **管理界面**:
  - RabbitMQ: http://localhost:15672 (admin/password123)
  - Consul: http://localhost:8500

## 开发指南

### 开发环境部署

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 查看服务状态
docker-compose -f docker-compose.dev.yml ps
```

### 后端开发

```bash
# 进入后端目录
cd backend

# 安装依赖
go mod download

# 运行特定服务（以用户服务为例）
cd services/user
go run main.go
```

### 前端开发

```bash
# 进入前端目录
cd frontend

# 配置Qt环境
# 确保Qt 6.5+已安装并配置环境变量

# 构建项目
mkdir build && cd build
cmake ..
make

# 运行应用
./VideoConferenceClient
```

### AI模型开发

```bash
# 进入AI检测目录
cd ai-detection

# 创建Python虚拟环境
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 安装依赖
pip install -r requirements.txt

# 运行开发服务器
python app.py
```

## 系统架构

### 微服务划分

1. **用户服务 (user-service)** - 端口 8081
   - 用户注册、登录、认证
   - 用户信息管理
   - JWT令牌管理

2. **会议服务 (meeting-service)** - 端口 8082
   - 会议创建、管理
   - 会议室状态管理
   - 参会者管理

3. **信令服务 (signaling-service)** - 端口 8083
   - WebRTC信令处理
   - 实时通信协调
   - 连接状态管理

4. **媒体服务 (media-service)** - 端口 8084
   - 音视频流处理
   - FFmpeg编解码
   - 流转发和录制

5. **检测服务 (detection-service)** - 端口 8085
   - 伪造音视频检测
   - AI模型推理
   - 检测结果存储

6. **记录服务 (record-service)** - 端口 8086
   - 通讯记录存储
   - 会议记录管理
   - 历史数据查询

7. **通知服务 (notification-service)** - 端口 8087
   - 实时消息推送
   - 邮件通知
   - 系统通知

8. **网关服务 (gateway-service)** - 端口 8080
   - API网关
   - 路由转发
   - 认证授权

9. **AI检测服务 (ai-detection)** - 端口 8501
   - 深度学习模型推理
   - 人脸伪造检测
   - 语音合成检测

## 数据库设计

### PostgreSQL (主数据库)
- 用户信息
- 会议信息
- 检测结果
- 系统配置

### MongoDB (日志数据库)
- 通讯记录
- 会议记录
- 操作日志
- 检测日志

### Redis (缓存)
- 会话缓存
- 会议状态缓存
- 检测结果缓存

## API文档

详细的API文档请参考：
- [用户服务API](docs/api/user-service.md)
- [会议服务API](docs/api/meeting-service.md)
- [检测服务API](docs/api/detection-service.md)
- [完整API设计](docs/api-design.md)

## 部署架构

```
[负载均衡器 Nginx]
    ↓
[API网关]
    ↓
[微服务集群]
    ├── 用户服务 (多实例)
    ├── 会议服务 (多实例)
    ├── 信令服务 (多实例)
    ├── 媒体服务 (多实例)
    ├── 检测服务 (多实例)
    ├── 记录服务 (多实例)
    └── 通知服务 (多实例)
    ↓
[数据层]
    ├── PostgreSQL 集群
    ├── MongoDB 集群
    ├── Redis 集群
    └── RabbitMQ 集群
```

## 项目结构

```
video-conference-system/
├── backend/                 # Go后端服务
│   ├── services/           # 微服务
│   │   ├── user/           # 用户服务
│   │   ├── meeting/        # 会议服务
│   │   ├── signaling/      # 信令服务
│   │   ├── media/          # 媒体服务
│   │   ├── detection/      # 检测服务
│   │   ├── record/         # 记录服务
│   │   ├── notification/   # 通知服务
│   │   └── gateway/        # 网关服务
│   ├── shared/             # 共享库
│   │   ├── config/         # 配置管理
│   │   ├── database/       # 数据库连接
│   │   ├── auth/           # 认证工具
│   │   └── models/         # 数据模型
│   ├── proto/              # gRPC协议定义
│   └── deploy/             # 部署配置
├── frontend/               # Qt前端应用
│   ├── src/                # 源代码
│   │   ├── services/       # 服务层
│   │   ├── models/         # 数据模型
│   │   ├── controllers/    # 控制器
│   │   ├── ui/             # UI组件
│   │   └── utils/          # 工具类
│   ├── qml/                # QML文件
│   ├── ui/                 # UI文件
│   └── resources/          # 资源文件
├── ai-detection/           # AI检测模块
│   ├── models/             # 训练模型
│   ├── inference/          # 推理服务
│   ├── training/           # 训练脚本
│   └── app.py              # Flask应用
├── docs/                   # 文档
│   ├── api/                # API文档
│   ├── deployment/         # 部署文档
│   └── development/        # 开发文档
├── scripts/                # 部署脚本
│   ├── deploy.sh           # 部署脚本
│   ├── test.sh             # 测试脚本
│   └── backup.sh           # 备份脚本
├── docker-compose.yml      # 生产环境容器编排
├── docker-compose.dev.yml  # 开发环境容器编排
└── README.md               # 项目说明
```

## 核心功能

### 1. 多人视频会议
- 支持最多50人同时在线
- 高清音视频传输 (1080p@30fps)
- 屏幕共享
- 文字聊天
- 会议录制

### 2. 伪造音视频检测
- 实时人脸伪造检测 (FaceSwap, Deepfake)
- 语音合成检测 (TTS, Voice Cloning)
- 检测结果实时显示
- 检测历史记录
- 可疑活动告警

### 3. 会议管理
- 会议预约和调度
- 会议权限控制
- 参会者管理
- 会议录制和回放
- 会议统计分析

### 4. 数据记录
- 完整的通讯记录
- 详细的会议记录
- 用户行为日志
- 系统监控日志
- 检测结果归档

## 性能指标

- **并发用户**: 10,000+
- **音视频延迟**: <100ms
- **检测响应时间**: <500ms
- **系统可用性**: 99.9%
- **数据处理能力**: 1TB/day

## 安全特性

- **身份认证**: JWT + OAuth 2.0
- **传输加密**: HTTPS/WSS (TLS 1.3)
- **数据加密**: AES-256 数据库加密
- **访问控制**: RBAC权限模型
- **审计日志**: 完整的操作审计
- **安全扫描**: 自动化安全漏洞检测

## 监控和运维

### 日志管理
- **集中日志**: ELK Stack (Elasticsearch + Logstash + Kibana)
- **日志级别**: DEBUG, INFO, WARN, ERROR, FATAL
- **日志轮转**: 按大小和时间自动轮转

### 监控指标
- **系统监控**: Prometheus + Grafana
- **应用监控**: 自定义业务指标
- **告警通知**: 邮件、短信、钉钉

### 健康检查
- **服务健康**: HTTP健康检查端点
- **数据库健康**: 连接池状态监控
- **依赖服务**: 外部服务可用性检查

## 常用命令

### 部署相关
```bash
# 完整部署
./scripts/deploy.sh

# 启动服务
./scripts/deploy.sh start

# 停止服务
./scripts/deploy.sh stop

# 重启服务
./scripts/deploy.sh restart

# 查看状态
./scripts/deploy.sh status

# 查看日志
./scripts/deploy.sh logs [service-name]

# 清理环境
./scripts/deploy.sh cleanup
```

### 开发相关
```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 查看开发环境状态
docker-compose -f docker-compose.dev.yml ps

# 查看特定服务日志
docker-compose -f docker-compose.dev.yml logs -f user-service-dev

# 进入服务容器
docker-compose -f docker-compose.dev.yml exec user-service-dev bash

# 重启特定服务
docker-compose -f docker-compose.dev.yml restart user-service-dev
```

### 测试相关
```bash
# 运行所有测试
./scripts/test.sh

# 运行特定类型测试
./scripts/test.sh health      # 健康检查测试
./scripts/test.sh database    # 数据库测试
./scripts/test.sh api         # API测试
./scripts/test.sh ai          # AI检测测试
./scripts/test.sh performance # 性能测试
./scripts/test.sh security    # 安全测试
```

## 故障排除

### 常见问题

1. **服务启动失败**
   ```bash
   # 检查Docker状态
   docker ps -a

   # 查看服务日志
   docker-compose logs [service-name]

   # 重启服务
   docker-compose restart [service-name]
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose exec postgres pg_isready

   # 查看数据库日志
   docker-compose logs postgres
   ```

3. **AI检测服务异常**
   ```bash
   # 检查AI服务状态
   curl http://localhost:8501/health

   # 查看AI服务日志
   docker-compose logs ai-detection
   ```

### 性能优化

1. **数据库优化**
   - 定期执行 `VACUUM` 和 `ANALYZE`
   - 监控慢查询日志
   - 优化索引策略

2. **缓存优化**
   - 合理设置Redis过期时间
   - 监控缓存命中率
   - 定期清理无效缓存

3. **服务优化**
   - 调整Go服务的GOMAXPROCS
   - 优化数据库连接池大小
   - 配置合适的超时时间

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目维护者: [luoxx](mailto:luoxx@stu.xju.edu.cn)
- 项目主页: [GitHub Repository](https://github.com/your-repo/video-conference-system)
- 问题反馈: [Issues](https://github.com/your-repo/video-conference-system/issues)

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 基础视频会议功能
- AI检测功能
- 微服务架构实现

---

**注意**: 这是一个演示项目，生产环境使用前请确保进行充分的安全评估和性能测试。
