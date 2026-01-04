# 🔌 API 文档

本目录提供当前实现的接口列表。基础入口与路径均以 `docker-compose.yml` 和服务源码为准。

## 🚀 快速开始

- **基础 URL**：`http://localhost:8800`（经 Nginx，同源前端）
- **认证**：除公开接口外，需携带 `Authorization: Bearer <jwt>`。CSRF Token 可从 `GET /api/v1/csrf-token` 获取（user-service）。
- **响应格式**：`{"code":200,"message":"success","data":...}`；错误包含 `code`/`message`。

## 📄 文档

- 详细接口列表与示例：`API_DOCUMENTATION.md`
- Web 客户端调用示例：`../CLIENT/API_USAGE_GUIDE.md`
- 通信设计：`../CLIENT/COMMUNICATION_DESIGN.md`

## 🧭 涉及服务

- 用户：`user-service` (`/api/v1/auth/*`, `/api/v1/users/*`, `/api/v1/admin/users`)
- 会议：`meeting-service` (`/api/v1/meetings/*`, `/api/v1/my/*`, `/api/v1/admin/meetings/*`)
- 信令：`signaling-service` (`/ws/signaling`, `/api/v1/sessions/*`, `/api/v1/stats/*`)
- 媒体：`media-service` (`/api/v1/media/*`, `/api/v1/recording/*`, `/api/v1/webrtc/*`, `/api/v1/ffmpeg/*`, `/api/v1/ai/*`)
- AI 推理：`ai-inference-service` (`/api/v1/ai/{asr,emotion,synthesis,setup,batch,health,info,analyze}`)
