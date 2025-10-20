# 💻 客户端文档

本目录包含 Qt6 客户端相关的文档和指南。

## 📖 文档列表

### 客户端指南
- **[API_USAGE_GUIDE.md](API_USAGE_GUIDE.md)** - 客户端 API 使用指南
- **[COMMUNICATION_DESIGN.md](COMMUNICATION_DESIGN.md)** - 客户端-服务器通信架构

### 功能文档
- **[AI_FEATURES.md](AI_FEATURES.md)** - AI 功能实现和使用
- **[STICKER_FEATURE.md](STICKER_FEATURE.md)** - 视频贴图特效功能

## 🎯 客户端架构

### 主要模块
- **Application** - 应用程序入口和初始化
- **Network Layer** - HTTP API 和 WebSocket 通信
- **Services Layer** - 业务服务层
- **WebRTC Layer** - WebRTC 实现和媒体管理
- **UI Layer** - 用户界面控制器

### 通信方式
- **HTTP/REST** - 用于 API 调用
- **WebSocket** - 用于实时信令
- **WebRTC** - 用于音视频通信

## 🚀 快速开始

### 构建客户端

**Windows**:
```powershell
cd qt6-client
.\setup-and-build.ps1
```

**Linux/macOS**:
```bash
cd qt6-client
./build.sh
```

### 运行客户端
```bash
./build/MeetingSystemClient
```

## 📚 主要功能

### 用户认证
- 用户注册
- 用户登录
- Token 管理

### 会议管理
- 创建会议
- 加入会议
- 离开会议
- 参与者管理

### 音视频通话
- 音频输入/输出
- 视频采集/显示
- 屏幕共享
- 媒体流管理

### AI 功能
- 语音识别
- 情感检测
- 音频降噪
- 视频增强

### 视频特效
- 实时滤镜
- 虚拟背景
- 美颜功能
- 贴图特效

## 🔧 开发指南

### 添加新的 API 调用
1. 在 `ApiClient` 中添加方法
2. 在相应的 Service 中调用
3. 在 UI Controller 中处理响应
4. 更新文档

### 添加新的 UI 界面
1. 创建 QML 文件
2. 创建对应的 Controller
3. 连接信号和槽
4. 集成到主窗口

### 集成新的 AI 功能
1. 在 `AIService` 中添加 API 调用
2. 在 UI 中添加相应的控件
3. 处理 AI 结果
4. 更新文档

## 📚 相关文档

- [API 文档](../API/README.md) - API 接口参考
- [开发指南](../DEVELOPMENT/README.md) - 后端开发
- [部署指南](../DEPLOYMENT/README.md) - 系统部署
- [文档中心](../README.md) - 所有文档

## 🔗 相关链接

- [Qt6 客户端 README](../../qt6-client/README.md)
- [项目主 README](../../README.md)
- [后端系统 README](../README.md)
- [Qt6 官方文档](https://doc.qt.io/qt-6/)
- [WebRTC 文档](https://webrtc.org/)

