# 💻 Web 客户端文档

当前仓库仅包含已构建的 Web 客户端（`frontend/dist`），由 Nginx 同源提供。无需单独编译，容器启动后即可通过 `http://localhost:8800` 访问。

## 📖 文档列表

- **[API_USAGE_GUIDE.md](API_USAGE_GUIDE.md)**：前端调用后端 API 的流程与示例
- **[COMMUNICATION_DESIGN.md](COMMUNICATION_DESIGN.md)**：HTTP/WebSocket/WebRTC 通信设计
- **[AI_FEATURES.md](AI_FEATURES.md)**：实时 AI/手动检测能力（ASR/情感/合成检测）
- **[VIDEO_EFFECTS_SEI.md](VIDEO_EFFECTS_SEI.md)**：H264 SEI 携带滤镜/美颜参数
- **[STICKER_FEATURE.md](STICKER_FEATURE.md)**：贴图/虚拟形象说明

## 🏗️ 客户端概览

- **资源位置**：`meeting-system/frontend/dist`（`index.html`, `app.js`, `styles.css`）
- **主要能力**：
  - 登录/注册、会话保持（JWT）
  - 创建/加入会议，音视频开关、屏幕共享
  - WebSocket 信令 `/ws/signaling`，显示房间/参与者状态
  - 聊天消息、基础控制台数据
  - 实时 AI：语音识别、情绪、合成检测，字幕/标签展示
  - H264 SEI 透传本地美颜/滤镜参数（浏览器支持 Encoded Streams 时）
- **依赖**：同源 API (`http://<host>:8800`)，HTTPS 环境可启用摄像头/麦克风。

## 🚀 使用

1. 启动后端与网关：`docker compose up -d`（在 `meeting-system`）。
2. 浏览器访问 `http://localhost:8800`（生产请使用 HTTPS）。
3. 注册或使用已有账号登录，创建/加入会议进行通话和 AI 检测。

## 🔧 配置与调试

- API 基础路径：相对同源（`/api/v1/...`），无需修改前端配置。
- 若自定义域名/端口，保持 Nginx 同源反代即可；信令使用 `ws(s)://<host>/ws/signaling`。
- 浏览器需允许获取摄像头/麦克风；非 HTTPS 环境可能被阻止。

## 📚 相关文档

- [API 文档](../API/README.md)
- [部署指南](../DEPLOYMENT/README.md)
- [开发/测试](../DEVELOPMENT/README.md)
- [架构](../ARCHITECTURE_DIAGRAM.md)
