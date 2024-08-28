# 🎥 智能会议系统 - Meeting System

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![C++ Standard](https://img.shields.io/badge/C++-17-blue.svg)](https://isocpp.org/)
[![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](https://www.docker.com/)
[![WebRTC](https://img.shields.io/badge/WebRTC-SFU-green.svg)](https://webrtc.org/)

基于SFU架构的企业级智能音视频会议系统，集成Edge-LLM-Infra分布式AI推理框架，提供实时AI检测、音视频增强、智能分析等功能。

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
- **手势识别**: 实时手势检测和交互
- **智能摘要**: 会议内容自动总结

### 🎨 视频特效
- **实时滤镜**: OpenCV + OpenGL实现的视频滤镜
- **虚拟背景**: AI背景分割和替换
- **美颜功能**: 实时面部美化和调整
- **贴图特效**: 动态贴图和AR效果

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
- **Web客户端**: HTML5 + JavaScript + WebRTC
- **移动端**: React Native (规划中)
- **管理界面**: Vue.js + Element Plus

### 部署技术
- **容器化**: Docker + Docker Compose
- **编排**: Kubernetes (可选)
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

# 切换到重构分支
git checkout reconstruct

# 启动所有服务
cd meeting-system/deployment/docker
docker-compose up -d

# 初始化数据库
cd ../scripts
chmod +x init-database.sh
./init-database.sh
```

### 构建系统
```bash
# 构建所有组件
cd meeting-system/deployment/scripts
chmod +x build.sh
./build.sh

# 或者分别构建
./build.sh go          # 构建Go微服务
./build.sh ai          # 构建AI节点
./build.sh docker      # 构建Docker镜像
./build.sh frontend    # 构建前端应用
```

### 访问系统
- **Web客户端**: http://localhost
- **管理界面**: http://localhost/admin
- **API文档**: http://localhost/api/docs
- **监控面板**: http://localhost:3000 (Grafana)
- **指标监控**: http://localhost:9090 (Prometheus)

## 📁 项目结构

```
meeting-system/
├── backend/                 # Go微服务后端
│   ├── shared/             # 共享库和工具
│   ├── user-service/       # 用户服务
│   ├── meeting-service/    # 会议服务
│   ├── signaling-service/  # 信令服务
│   ├── media-service/      # 媒体服务(SFU)
│   ├── ai-service/         # AI检测服务
│   └── notification-service/ # 通知服务
├── ai-node/                # C++ AI推理节点
│   ├── include/            # 头文件
│   ├── src/                # 源代码
│   └── config/             # 配置文件
├── frontend/               # 前端应用
│   ├── qt/                 # Qt6桌面客户端
│   ├── web/                # Web浏览器客户端
│   └── mobile/             # 移动端客户端
├── admin-web/              # Web管理界面
├── deployment/             # 部署配置
│   ├── docker/             # Docker配置
│   ├── k8s/                # Kubernetes配置
│   └── scripts/            # 部署脚本
└── docs/                   # 文档
```

## 🔧 开发指南

### 本地开发环境
```bash
# 启动开发环境数据库
docker-compose -f deployment/docker/docker-compose.yml up -d postgres redis mongodb minio

# 启动用户服务
cd backend/user-service
go run main.go

# 启动其他服务...
```

### API接口
- **用户管理**: `/api/v1/users/`
- **会议管理**: `/api/v1/meetings/`
- **信令通信**: `/ws/signaling/`
- **AI服务**: `/api/v1/ai/`
- **文件上传**: `/api/v1/upload/`

### 配置说明
主要配置文件位于 `backend/config/config.yaml`，包含：
- 数据库连接配置
- Redis缓存配置
- JWT认证配置
- AI服务配置
- 日志配置

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
