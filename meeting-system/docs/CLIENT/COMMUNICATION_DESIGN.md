# Web 客户端通信设计

适用于当前 Web 前端（`frontend/dist`）与后端微服务的通信方式，覆盖 HTTP、WebSocket、WebRTC 与 AI 请求。

## 总览

```
Browser (Web UI)
  ├─ HTTP(S) → Nginx → user/meeting/media/ai services
  ├─ WebSocket → Nginx → signaling-service (/ws/signaling)
  └─ WebRTC (P2P/SFU) → 通过信令协商媒体
```

- **基址**：同源 `window.location.origin`（默认 `http://localhost:8800`）。
- **认证**：JWT Bearer；CSRF Token `GET /api/v1/csrf-token`（stateful 变更接口使用 `X-CSRF-Token`）。
- **错误处理**：响应包含 `code`/`message`，HTTP 非 2xx 视为失败。

## HTTP API

- 用户：`/api/v1/auth/*`、`/api/v1/users/*`、`/api/v1/admin/users/*`
- 会议：`/api/v1/meetings/*`、`/api/v1/my/*`、`/api/v1/admin/meetings/*`
- 媒体/录制：`/api/v1/media/*`、`/api/v1/recording/*`、`/api/v1/webrtc/*`、`/api/v1/ffmpeg/*`
- AI 推理：`/api/v1/ai/{health,info,asr,emotion,synthesis,setup,batch,analyze}`

`app.js` 的 `apiFetch` 封装了默认头、JWT、可选 CSRF 与超时。

## WebSocket 信令

- URL：`ws(s)://<host>/ws/signaling?user_id=<uid>&meeting_id=<mid>&peer_id=<uuid>&token=<jwt>`
- 消息结构：`{id, type, peer_id, payload}`，`type` 对应 `WS_TYPES`（JOIN/OFFER/ANSWER/ICE/LEAVE/CHAT 等）。
- 服务器广播房间事件和错误；前端维护 `wsState`、心跳与重连提示。

## WebRTC 媒体

- ICE/STUN/TURN：来自 `signaling-service` 配置（`backend/config/signaling-service.yaml`），前端在 `join` 后获取并应用。
- SDP/ICE 交换：通过 WebSocket 消息 `OFFER/ANSWER/ICE`。
- 媒体控制：本地按钮控制音视频/屏幕共享，信令仅传输必要控制与聊天。

## AI 请求

- 实时检测：录制当前说话人音频，定期调用 `/api/v1/ai/{asr,emotion,synthesis}`。
- 手动检测：文件上传/文本输入触发对应 API。
- 健康/信息：`/api/v1/ai/health`、`/api/v1/ai/info`。

## 安全与部署注意

- 生产启用 HTTPS；否则浏览器可能拒绝 `getUserMedia`。
- 必须在后端设置强 `JWT_SECRET`，调整 `ALLOWED_ORIGINS` 以匹配实际域名。
- 若 AI 节点独立部署，更新 Nginx `ai_inference_service` 上游配置。
