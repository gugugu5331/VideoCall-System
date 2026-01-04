# 智能视频会议平台 API 文档

**基础 URL**：`http://localhost:8800`（默认经 Nginx，同源前端）  
**协议**：HTTP/HTTPS + WebSocket  
**认证**：JWT Bearer（`Authorization: Bearer <token>`），CSRF Token：`GET /api/v1/csrf-token`  
**响应格式**：`{"code":200,"message":"success","data":{...}}`；错误返回 `code`/`message`。

---

## 目录

1. [用户服务](#用户服务-user-service)
2. [会议服务](#会议服务-meeting-service)
3. [信令服务](#信令服务-signaling-service)
4. [媒体服务](#媒体服务-media-service)
5. [AI 推理服务](#ai-推理服务-ai-inference-service)
6. [健康检查与指标](#健康检查与指标)

---

## 用户服务 (user-service)

- `POST /api/v1/auth/register`：注册
- `POST /api/v1/auth/login`：登录
- `POST /api/v1/auth/refresh`：刷新 Token
- `POST /api/v1/auth/forgot-password` / `reset-password`：找回/重置密码
- `GET /api/v1/csrf-token`：获取 CSRF Token
- 认证接口（需 JWT + CSRF）：
  - `GET /api/v1/users/profile` / `PUT /api/v1/users/profile`
  - `POST /api/v1/users/change-password`
  - `POST /api/v1/users/upload-avatar`
  - `DELETE /api/v1/users/account`
- 管理员：
  - `GET /api/v1/admin/users`
  - `GET /api/v1/admin/users/:id`
  - `PUT /api/v1/admin/users/:id`
  - `DELETE /api/v1/admin/users/:id`
  - `POST /api/v1/admin/users/:id/ban`
  - `POST /api/v1/admin/users/:id/unban`

示例（登录）：
```http
POST /api/v1/auth/login
Content-Type: application/json

{"username":"demo","password":"secret"}
```

---

## 会议服务 (meeting-service)

> 所有接口需 JWT。

- 会议管理：
  - `POST /api/v1/meetings`
  - `GET /api/v1/meetings`
  - `GET /api/v1/meetings/:id`
  - `PUT /api/v1/meetings/:id`
  - `DELETE /api/v1/meetings/:id`
- 会议控制：
  - `POST /api/v1/meetings/:id/start`
  - `POST /api/v1/meetings/:id/end`
  - `POST /api/v1/meetings/:id/join`
  - `POST /api/v1/meetings/:id/leave`
- 参与者管理：
  - `GET /api/v1/meetings/:id/participants`
  - `POST /api/v1/meetings/:id/participants`
  - `DELETE /api/v1/meetings/:id/participants/:user_id`
  - `PUT /api/v1/meetings/:id/participants/:user_id/role`
- 房间与录制：
  - `GET /api/v1/meetings/:id/room`
  - `POST /api/v1/meetings/:id/room`
  - `DELETE /api/v1/meetings/:id/room`
  - `POST /api/v1/meetings/:id/recording/start`
  - `POST /api/v1/meetings/:id/recording/stop`
  - `GET /api/v1/meetings/:id/recordings`
- 聊天：
  - `GET /api/v1/meetings/:id/messages`
  - `POST /api/v1/meetings/:id/messages`
- 我的会议：
  - `GET /api/v1/my/meetings`
  - `GET /api/v1/my/meetings/upcoming`
  - `GET /api/v1/my/meetings/history`
- 管理员：
  - `GET /api/v1/admin/meetings`
  - `GET /api/v1/admin/meetings/stats`
  - `POST /api/v1/admin/meetings/:id/force-end`

---

## 信令服务 (signaling-service)

- WebSocket：`GET /ws/signaling`
  - 客户端发送：`join/offer/answer/candidate/leave/chat` 等类型（见前端 `app.js`）。
- REST（需 JWT）：
  - `GET /api/v1/sessions/:session_id`
  - `GET /api/v1/sessions/room/:meeting_id`
  - `GET /api/v1/messages/history/:meeting_id`
  - `GET /api/v1/stats/overview`
  - `GET /api/v1/stats/rooms`
- 管理：
  - `POST /admin/cleanup/sessions`
  - `GET /admin/sessions`

---

## 媒体服务 (media-service)

- 媒体管理：
  - `POST /api/v1/media/upload`
  - `GET /api/v1/media`（列表）
  - `GET /api/v1/media/download/:id`
  - `GET /api/v1/media/info/:id`
  - `POST /api/v1/media/process`
  - `DELETE /api/v1/media/:id`
- WebRTC/SFU 辅助：
  - `POST /api/v1/webrtc/answer`
  - `POST /api/v1/webrtc/ice-candidate`
  - `POST /api/v1/webrtc/room/:roomId/join`
  - `POST /api/v1/webrtc/room/:roomId/leave`
  - `GET /api/v1/webrtc/room/:roomId/peers`
  - `GET /api/v1/webrtc/room/:roomId/stats`
  - `POST /api/v1/webrtc/peer/:peerId/media`
  - `GET /api/v1/webrtc/peer/:peerId/status`
  - `GET /api/v1/webrtc/peer/:peerId/ice-candidates`
  - `GET /api/v1/webrtc/peer/:peerId/offer`
  - `POST /api/v1/webrtc/peer/:peerId/answer`
- 录制：
  - `POST /api/v1/recording/start`
  - `POST /api/v1/recording/stop`
  - `GET /api/v1/recording/status/:id`
  - `GET /api/v1/recording/list`
  - `GET /api/v1/recording/download/:id`
  - `DELETE /api/v1/recording/:id`
- FFmpeg（缩略图）：
  - `POST /api/v1/ffmpeg/thumbnail`
  - `GET /api/v1/ffmpeg/job/:id/status`
- AI 状态（媒体处理侧）：
  - `GET /api/v1/ai/connectivity`
  - `GET /api/v1/ai/streams`
  - `GET /api/v1/ai/streams/:stream_id`

---

## AI 推理服务 (ai-inference-service)

- 健康与信息：
  - `GET /health`
  - `GET /api/v1/ai/health`
  - `GET /api/v1/ai/info`
- 推理：
  - `POST /api/v1/ai/asr`（音频 base64，字段：`audio_data`、`format`、`sample_rate`）
  - `POST /api/v1/ai/emotion`（音频或文本，`audio_data` 或 `text`）
  - `POST /api/v1/ai/synthesis`（音频深度伪造检测）
  - `POST /api/v1/ai/setup`（预热指定 `meeting_id` 的模型）
  - `POST /api/v1/ai/analyze`（通用任务：`task_type` + `input_data`）
  - `POST /api/v1/ai/batch`（批量任务列表）

返回示例（成功）：
```json
{"code":200,"message":"success","data":{"text":"...","confidence":0.95}}
```

---

## 健康检查与指标

- `/health`：各服务健康（user/meeting/signaling/media/ai）
- `/metrics`：Prometheus 指标
- `/status`（media-service）：服务状态
- Nginx 网关健康：`http://localhost:8800/health`

---

## 错误与限流

- 认证失败：HTTP 401，`{"code":401,"message":"unauthorized"}`。
- 参数错误：HTTP 400。
- 业务错误：HTTP 4xx/5xx，`code` 与 `message` 描述原因。
- 限流：当前限流中间件在部分接口默认关闭（测试场景），生产可在配置中开启。
