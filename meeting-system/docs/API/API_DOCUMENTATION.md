# 智能视频会议平台 API 文档

- **Base URL**：`http://localhost:8800`（经 Nginx，同源前端）
- **协议**：HTTP/HTTPS + WebSocket
- **认证**：`Authorization: Bearer <jwt>`；需要状态变更时可加 `X-CSRF-Token`（`GET /api/v1/csrf-token`）
- **响应格式**：`{"code":200,"message":"success","data":{...}}`

> AI 端点需启用 `ai-inference-service`。所有端口与路径以实际 compose 与配置文件为准。

通用错误约定：
- 401/403：Token 缺失或过期
- 400：参数错误，`message` 描述详情
- 5xx：内部错误，需查看对应服务日志

---

## 用户服务（user-service）

公开：
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/reset-password`
- `GET /api/v1/csrf-token`

需 JWT（部分建议带 CSRF）：
- `GET/PUT /api/v1/users/profile`
- `POST /api/v1/users/change-password`
- `POST /api/v1/users/upload-avatar`
- `DELETE /api/v1/users/account`

管理员：
- `GET /api/v1/admin/users`
- `GET/PUT/DELETE /api/v1/admin/users/:id`
- `POST /api/v1/admin/users/:id/{ban,unban}`

---

## 会议服务（meeting-service）

全部需 JWT：
- 会议 CRUD：`POST/GET/GET:id/PUT/DELETE /api/v1/meetings`
- 控制：`POST /api/v1/meetings/:id/{start,end,join,leave}`
- 参与者：`GET /api/v1/meetings/:id/participants`、`POST` 新增、`DELETE /:user_id` 移除、`PUT /:user_id/role`
- 房间/录制：`GET|POST|DELETE /api/v1/meetings/:id/room`，`POST /api/v1/meetings/:id/recording/{start,stop}`，`GET /api/v1/meetings/:id/recordings`
- 聊天：`GET/POST /api/v1/meetings/:id/messages`
- 我的会议：`/api/v1/my/meetings`、`/upcoming`、`/history`
- 管理：`/api/v1/admin/meetings`、`/stats`、`POST /api/v1/admin/meetings/:id/force-end`

---

## 信令服务（signaling-service）

- WebSocket：`GET /ws/signaling?user_id=<uid>&meeting_id=<mid>&peer_id=<uuid>&token=<jwt>`
  - 消息类型：JOIN/OFFER/ANSWER/ICE/LEAVE/CHAT 等（参见前端 `WS_TYPES`）
- REST（需 JWT）：
  - `GET /api/v1/sessions/:session_id`
  - `GET /api/v1/sessions/room/:meeting_id`
  - `GET /api/v1/messages/history/:meeting_id`
  - `GET /api/v1/stats/{overview,rooms}`
- 管理：
  - `POST /admin/cleanup/sessions`
  - `GET /admin/sessions`

---

## 媒体服务（media-service）

- 媒体：`POST /api/v1/media/upload`、`GET /api/v1/media`、`GET /api/v1/media/{download|info}/:id`、`POST /api/v1/media/process`、`DELETE /api/v1/media/:id`
- WebRTC/SFU：`POST /api/v1/webrtc/{answer,ice-candidate}`、`POST /api/v1/webrtc/room/:roomId/{join,leave}`、`GET /api/v1/webrtc/room/:roomId/{peers,stats}`、`POST /api/v1/webrtc/peer/:peerId/media`、`GET /api/v1/webrtc/peer/:peerId/{status,ice-candidates,offer}`、`POST /api/v1/webrtc/peer/:peerId/answer`
- 录制：`POST /api/v1/recording/{start,stop}`、`GET /api/v1/recording/{status/:id,list,download/:id}`、`DELETE /api/v1/recording/:id`
- FFmpeg：`POST /api/v1/ffmpeg/thumbnail`、`GET /api/v1/ffmpeg/job/:id/status`
- AI 状态（媒体侧观测）：`GET /api/v1/ai/{connectivity,streams,streams/:id}`

---

## AI 推理服务（ai-inference-service，可选）

- 健康/信息：`GET /health`、`GET /api/v1/ai/{health,info}`
- 推理：
  - `POST /api/v1/ai/asr`（`audio_data` base64，`format`=`wav`，`sample_rate`=16000）
  - `POST /api/v1/ai/emotion`（音频或 `text`，取决于模型配置）
  - `POST /api/v1/ai/synthesis`（音频深度伪造检测）
  - `POST /api/v1/ai/setup`（可选预热）
  - `POST /api/v1/ai/analyze`（通用任务：`task_type` + `input_data`）
  - `POST /api/v1/ai/batch`（批量任务数组）

返回示例：
```json
{"code":200,"message":"success","data":{"text":"hello","confidence":0.95}}
```

---

## 健康与指标

- `/health`：服务健康检查（user/meeting/signaling/media/ai）
- `/metrics`：Prometheus 指标
- `/status`（media-service）：服务状态
- Nginx 网关健康：`http://localhost:8800/health`

## 错误与限流

- 401/403：JWT 缺失或过期
- 400：参数错误；参考 `message`
- 5xx：服务内部异常，请检查容器日志或上游依赖
- 限流：按配置启用，默认开发环境较宽松

调试提示：
- 若接口返回 CSRF 相关错误，请先获取 `X-CSRF-Token`。
- 通过网关访问即可；如需直连服务，使用相应容器端口（如 8080/8082 等）。
