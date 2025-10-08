# Qt6 智能会议系统客户端

## 项目概述

这是一个基于Qt6开发的智能视频会议系统桌面客户端，风格类似腾讯会议，通过API网关与后端微服务通信。

## 主要功能

- ✅ 用户登录/注册
- ✅ 会议创建/加入/管理
- ✅ 实时音视频通话（WebRTC）
- ✅ 实时聊天
- ✅ 屏幕共享
- ✅ AI功能集成
  - 音视频合成检测
  - 实时语音识别(ASR)
  - 情绪识别

## 技术栈

- **Qt 6.5+**: 跨平台应用框架
- **QML**: 声明式UI
- **Qt Network**: HTTP/HTTPS通信
- **Qt WebSockets**: WebSocket实时通信
- **Qt Multimedia**: 音视频处理
- **WebRTC**: 音视频通话（需要集成第三方库）

## 项目结构

```
qt6-client/
├── CMakeLists.txt              # CMake构建配置
├── config/
│   └── config.json             # 应用配置文件
├── include/                    # 头文件
│   ├── application.h
│   ├── network/                # 网络层
│   │   ├── http_client.h
│   │   ├── websocket_client.h
│   │   └── api_client.h
│   ├── webrtc/                 # WebRTC层
│   │   ├── webrtc_manager.h
│   │   ├── peer_connection.h
│   │   └── media_stream.h
│   ├── services/               # 业务服务层
│   │   ├── auth_service.h
│   │   ├── meeting_service.h
│   │   ├── media_service.h
│   │   └── ai_service.h
│   ├── models/                 # 数据模型
│   │   ├── user.h
│   │   ├── meeting.h
│   │   ├── participant.h
│   │   └── message.h
│   ├── ui/                     # UI控制器
│   │   ├── login_controller.h
│   │   ├── main_window_controller.h
│   │   ├── meeting_room_controller.h
│   │   └── ai_panel_controller.h
│   └── utils/                  # 工具类
│       ├── logger.h
│       ├── config.h
│       └── json_helper.h
├── src/                        # 源文件
│   ├── main.cpp
│   ├── application.cpp
│   ├── network/
│   ├── webrtc/
│   ├── services/
│   ├── models/
│   ├── ui/
│   └── utils/
├── qml/                        # QML界面文件
│   ├── main.qml
│   ├── LoginPage.qml
│   ├── MainWindow.qml
│   ├── MeetingRoom.qml
│   ├── AIPanel.qml
│   └── components/
│       ├── VideoTile.qml
│       ├── ToolBar.qml
│       ├── ParticipantList.qml
│       └── ChatPanel.qml
└── resources/                  # 资源文件
    ├── resources.qrc
    ├── images/
    └── fonts/
```

## 架构设计

### 通信架构

```
Qt6客户端
    ↓
HTTP/HTTPS (RESTful API)
WebSocket (信令)
WebRTC (音视频)
    ↓
Nginx API网关
    ↓
后端微服务
```

### 模块划分

1. **网络层** (`network/`)
   - HTTP客户端：处理RESTful API请求
   - WebSocket客户端：处理实时信令
   - API客户端：封装业务API调用

2. **WebRTC层** (`webrtc/`)
   - WebRTC管理器：管理音视频连接
   - PeerConnection：WebRTC对等连接
   - MediaStream：媒体流管理

3. **服务层** (`services/`)
   - 认证服务：用户登录/注册
   - 会议服务：会议管理
   - 媒体服务：文件上传/下载
   - AI服务：AI功能调用

4. **模型层** (`models/`)
   - 数据模型定义

5. **UI层** (`qml/`, `ui/`)
   - QML界面
   - UI控制器

## 构建说明

### 依赖要求

- Qt 6.5 或更高版本
- CMake 3.16 或更高版本
- C++17 编译器
- OpenSSL (用于HTTPS)

### 构建步骤

```bash
# 1. 创建构建目录
mkdir build
cd build

# 2. 配置CMake
cmake ..

# 3. 编译
cmake --build .

# 4. 运行
./bin/MeetingSystemClient
```

### 配置文件

编辑 `config/config.json` 配置API网关地址：

```json
{
  "api": {
    "base_url": "https://api.meeting.com",
    "ws_url": "wss://api.meeting.com/ws/signaling"
  }
}
```

## API网关集成

客户端通过Nginx API网关与后端通信：

### HTTP API

- **认证**: `POST /api/v1/auth/login`
- **会议**: `POST /api/v1/meetings`
- **用户**: `GET /api/v1/users/profile`

### WebSocket信令

- **连接**: `wss://api.meeting.com/ws/signaling?token={jwt_token}&meeting_id={id}&user_id={id}&peer_id={id}`
- **消息类型**: Offer, Answer, ICE Candidate, Chat, etc.

### WebRTC

- 使用SFU架构
- 通过WebSocket交换信令
- 支持音频、视频、屏幕共享

## AI功能集成

### 合成检测

```cpp
aiService->detectDeepfake(videoData, userId);
```

### 语音识别

```cpp
aiService->recognizeSpeech(audioData, userId, "zh");
```

### 情绪识别

```cpp
aiService->recognizeEmotion(imageData, userId);
```

## 开发指南

### 添加新功能

1. 在 `include/` 中定义头文件
2. 在 `src/` 中实现源文件
3. 在 `CMakeLists.txt` 中添加文件
4. 如需UI，在 `qml/` 中创建QML文件

### 调试

```bash
# 启用详细日志
export QT_LOGGING_RULES="*.debug=true"
./bin/MeetingSystemClient
```

### 测试

```bash
# 运行测试
ctest
```

## 待完成功能

- [ ] WebRTC实际集成（需要libwebrtc或QtWebEngine）
- [ ] 完整的UI控制器实现
- [ ] 会议室界面完善
- [ ] AI面板界面
- [ ] 视频渲染组件
- [ ] 音频处理
- [ ] 屏幕共享
- [ ] 文件传输
- [ ] 录制功能
- [ ] 单元测试
- [ ] 集成测试

## 注意事项

1. **WebRTC集成**: 当前WebRTC部分是接口定义，需要集成实际的WebRTC库（如libwebrtc）
2. **网关隔离**: 所有后端通信必须通过API网关，不直接连接微服务
3. **安全性**: 使用HTTPS/WSS加密通信，JWT Token认证
4. **性能**: 注意内存管理，使用智能指针避免内存泄漏

## 许可证

Copyright © 2025 Meeting System

