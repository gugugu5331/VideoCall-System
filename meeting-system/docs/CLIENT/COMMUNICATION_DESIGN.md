# Web 客户端通信设计

面向当前 Web 前端（`frontend/dist`），涵盖 HTTP、WebSocket、WebRTC 与可选的 AI 请求。

## 总览

```
Browser
  ├─ HTTP(S) → Nginx → user/meeting/media/ai
  ├─ WebSocket → Nginx → signaling-service (/ws/signaling)
  └─ WebRTC → P2P/SFU（通过信令协商）
```

- **基址**：`window.location.origin`（默认 `http://localhost:8800`）
- **认证**：JWT `Authorization: Bearer <token>`；变更接口可带 `X-CSRF-Token`
- **错误格式**：`{code, message, data}`，HTTP 非 2xx 视为失败

## HTTP API

- 用户：`/api/v1/auth/*`、`/api/v1/users/*`、`/api/v1/admin/users/*`
- 会议：`/api/v1/meetings/*`、`/api/v1/my/*`、`/api/v1/admin/meetings/*`
- 媒体/录制：`/api/v1/media/*`、`/api/v1/recording/*`、`/api/v1/webrtc/*`、`/api/v1/ffmpeg/*`
- AI（可选）：`/api/v1/ai/{health,info,asr,emotion,synthesis,setup,batch,analyze}`

请求统一走同源相对路径，默认 JSON；上传接口按需使用表单/文件流。

## WebSocket 信令

- URL：`ws(s)://<host>/ws/signaling?user_id=<uid>&meeting_id=<mid>&peer_id=<uuid>&token=<jwt>`
- 消息结构：`{id, type, peer_id, payload}`，`type` 取值与前端 `WS_TYPES` 对应（JOIN/OFFER/ANSWER/ICE/LEAVE/CHAT 等）
- 心跳/重连：前端维护基本心跳和状态提示，后端 Redis 持久化房间状态

## WebRTC 媒体

- ICE/STUN/TURN 配置来自 `signaling-service`（`backend/config/signaling-service.yaml`）
- SDP/ICE 通过 WS 交换；浏览器直接建立 P2P/SFU 连接
- 录制与媒资由 `media-service` 提供独立接口，不依赖信令通道

## AI 请求（可选）

- 实时或手动调用 `/api/v1/ai/{asr,emotion,synthesis}`，音频以 base64（WAV/PCM）提交
- 健康/信息：`/api/v1/ai/{health,info}`
- 上游依赖 `ai-inference-service` + Triton；未部署时前端应禁用相关按钮

## 安全与部署提示

- 生产必须启用 HTTPS，确保 `getUserMedia` 和 WebRTC 正常工作
- 与后端保持同源，调整域名/端口时更新 Nginx 反代即可
- JWT 秘钥不一致会导致 401/403；如跨域访问需同步调整 `ALLOWED_ORIGINS`
