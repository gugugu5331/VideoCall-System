# 🎥 智能会议系统 - Meeting System

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![C++ Standard](https://img.shields.io/badge/C++-17-blue.svg)](https://isocpp.org/)
[![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](https://www.docker.com/)
[![WebRTC](https://img.shields.io/badge/WebRTC-SFU-green.svg)](https://webrtc.org/)

基于SFU架构的企业级智能音视频会议系统，集成分布式AI推理框架，提供实时AI检测、音视频增强、智能分析等功能。

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端层                              │
├─────────────────┬─────────────────┬─────────────────────────┤
│   Qt6 桌面客户端  │   Web 浏览器客户端  │      移动端客户端        │
└─────────────────┴─────────────────┴─────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                        网关层                                │
├─────────────────────────────┬───────────────────────────────┤
│      Nginx 负载均衡          │         API 网关              │
└─────────────────────────────┴───────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      微服务层                                │
├──────┬──────┬──────┬──────┬──────┬──────┬─────────────────┤
│用户服务│会议服务│信令服务│媒体服务│AI检测服务│通知服务│                 │
└──────┴──────┴──────┴──────┴──────┴──────┴─────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    AI推理层                                  │
├─────────────────┬─────────────────┬─────────────────────────┤
│ Edge-LLM-Infra  │   模型管理器      │    推理节点集群          │
└─────────────────┴─────────────────┴─────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                      数据层                                  │
├──────────┬──────────┬──────────┬──────────┬─────────────────┤
│PostgreSQL│  Redis   │ MongoDB  │  MinIO   │                 │
└──────────┴──────────┴──────────┴──────────┴─────────────────┘
```

## ✨ 核心特性

### 🎯 音视频会议
- **SFU架构**: 基于Selective Forwarding Unit的高效媒体路由
- **WebRTC通信**: 低延迟P2P和多方音视频通话
- **实时信令**: WebSocket信令服务器处理连接协商
- **媒体处理**: FFmpeg音视频编解码和处理
- **屏幕共享**: 支持桌面和应用程序共享

### 🤖 AI智能功能
- **语音识别**: 实时语音转文字，支持多语言
- **情绪检测**: 基于面部表情的情绪分析
- **音频降噪**: AI驱动的实时音频质量优化
- **视频增强**: 智能视频质量提升和美颜
- **智能摘要**: 会议内容自动总结
- **合成检测**: 检测参会者是否为数字人
### 🎨 视频特效
- **实时滤镜**: OpenCV + OpenGL实现的视频滤镜
- **虚拟背景**: AI背景分割和替换
- **美颜功能**: 实时面部美化和调整
- **贴图特效**: 动态贴图

### 🔒 安全与认证
- **JWT认证**: 基于Token的用户认证
- **权限管理**: 细粒度的角色权限控制
- **数据加密**: 端到端加密通信
- **安全审计**: 完整的操作日志记录

## 🛠️ 技术栈

### 后端技术
- **语言**: Go 1.21+, C++ 17
- **框架**: Gin, GORM, gRPC
- **数据库**: PostgreSQL, Redis, MongoDB
- **存储**: MinIO对象存储
- **通信**: ZeroMQ, WebSocket, WebRTC
- **AI框架**: Edge-LLM-Infra

### 前端技术
- **桌面客户端**: Qt6 + QML
- **移动端**: React Native (规划中)


### 部署技术
- **容器化**: Docker + Docker Compose
- **消息队列**：采用redis实现了一个基于内存的消息队列，实现任务的调度
- **负载均衡**: Nginx
- **监控**: Prometheus + Grafana
- **CI/CD**: GitHub Actions

## 🚀 快速开始

### 环境要求
- Docker 20.0+
- Docker Compose 2.0+
- Go 1.21+ (开发环境)
- Qt6 (桌面客户端开发)
- CMake 3.10+ (AI节点编译)

### 一键部署
```bash
# 克隆项目
git clone https://github.com/gugugu5331/VideoCall-System.git
cd VideoCall-System

# 启动所有服务
cd meeting-system
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 开发环境启动
```bash
# 启动基础服务（数据库、缓存等）
cd meeting-system
docker-compose up -d postgres redis mongodb minio

# 启动后端服务
cd backend
./start_services.sh

# 构建Qt6客户端
cd ../qt6-client
mkdir build && cd build
cmake ..
cmake --build .
```

### 访问系统
- **Web客户端**: http://localhost
- **管理界面**: http://localhost/admin
- **API文档**: http://localhost/api/docs
- **监控面板**: http://localhost:3000 (Grafana)
- **指标监控**: http://localhost:9090 (Prometheus)

## 📁 项目结构

```
VideoCall-System/
├── meeting-system/          # 后端服务系统
│   ├── backend/            # Go微服务后端
│   │   ├── shared/         # 共享库和工具
│   │   ├── user-service/   # 用户服务 (8080)
│   │   ├── meeting-service/ # 会议服务 (8082)
│   │   ├── signaling-service/ # 信令服务 (8081)
│   │   ├── media-service/  # 媒体服务/SFU (8083)
│   │   └── ai-inference-service/ # AI推理服务 (8085)
│   ├── Edge-LLM-Infra/ # AI推理框架
│   │   ├── unit-manager/   # 单元管理器 (10001)
│   │   ├── node/           # AI推理节点
│   │   └── network/        # 网络通信层
│   ├── deployment/         # 部署配置
│   │   ├── docker/         # Docker配置
│   │   └── scripts/        # 部署脚本
│   ├── nginx/              # Nginx配置
│   ├── monitoring/         # 监控配置
│   ├── docs/               # 文档
│   │   ├── deployment/     # 部署文档
│   │   ├── testing/        # 测试文档
│   │   └── interview/      # 面试参考
│   └── docker-compose.yml  # Docker编排文件
│
└── qt6-client/             # Qt6桌面客户端
    ├── src/                # C++源代码
    ├── include/            # 头文件
    ├── qml/                # QML界面
    ├── resources/          # 资源文件
    ├── tests/              # 测试文件
    └── docs/               # 客户端文档
```

## 🔧 开发指南

### 本地开发环境
```bash
# 启动开发环境数据库
cd meeting-system
docker-compose up -d postgres redis mongodb minio

# 启动用户服务
cd backend/user-service
go run main.go

# 启动会议服务
cd ../meeting-service
go run main.go

# 启动信令服务
cd ../signaling-service
go run main.go

# 或使用脚本启动所有服务
cd backend
./start_services.sh
```

### API接口
所有API通过Nginx网关访问 (http://localhost)

- **用户管理**: `POST /api/v1/auth/login`, `POST /api/v1/auth/register`
- **会议管理**: `POST /api/v1/meetings`, `GET /api/v1/meetings/:id`
- **信令通信**: `WS /ws/signaling?token={jwt}&meeting_id={id}`
- **AI服务**: `POST /api/v1/ai/asr`, `POST /api/v1/ai/emotion`
- **媒体服务**: `POST /api/v1/media/upload`

### 配置说明
各服务配置文件位于 `meeting-system/backend/config/`：
- `user-service.yaml` - 用户服务配置
- `meeting-service.yaml` - 会议服务配置
- `signaling-service.yaml` - 信令服务配置
- `media-service.yaml` - 媒体服务配置
- `ai-service.yaml` - AI服务配置

主要配置项：
- 数据库连接 (PostgreSQL, MongoDB, Redis)
- JWT认证密钥
- AI推理节点地址
- 日志级别和路径

## 📊 性能指标

- **并发用户**: 支持1000+并发用户
- **会议规模**: 单会议支持100+参与者
- **延迟**: 端到端延迟 < 200ms
- **可用性**: 99.9%系统可用性
- **扩展性**: 水平扩展支持

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Edge-LLM-Infra](https://github.com/gugugu5331/Edge-LLM-Infra) - 分布式AI推理框架
- [WebRTC](https://webrtc.org/) - 实时通信技术
- [Go](https://golang.org/) - 后端开发语言
- [Qt](https://www.qt.io/) - 跨平台GUI框架

## 📞 联系我们

- 项目主页: https://github.com/gugugu5331/VideoCall-System
- 问题反馈: https://github.com/gugugu5331/VideoCall-System/issues
- 邮箱: gugugu5331@example.com

---

⭐ 如果这个项目对你有帮助，请给我们一个星标！
