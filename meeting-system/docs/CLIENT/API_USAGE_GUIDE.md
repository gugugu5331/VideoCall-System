# 客户端 API 使用指南（Web）

前端通过同源的 Nginx 访问后端，所有请求均以 `/api/v1/...` 相对路径发起，WebSocket 信令走 `/ws/signaling`。

## 基础约定

- **Base URL**：`window.location.origin`（默认 `http://localhost:8800`）
- **认证**：`Authorization: Bearer <jwt>`，CSRF Token（`X-CSRF-Token`）可从 `GET /api/v1/csrf-token` 获取
- **默认头**：`Content-Type: application/json`，`Accept: application/json`

## 认证流程

1) **注册**
```http
POST /api/v1/auth/register
{"username":"demo","email":"demo@example.com","password":"Passw0rd!","nickname":"Demo"}
```
2) **登录**
```http
POST /api/v1/auth/login
{"username":"demo","password":"Passw0rd!"}
```
返回 `data.token`（JWT）。前端保存后用于后续请求与 WebSocket。

3) **刷新**
```http
POST /api/v1/auth/refresh
{"token":"<old-jwt>"}
```

4) **获取 CSRF Token（变更接口推荐携带）**
```http
GET /api/v1/csrf-token
```

## 用户接口

- `GET /api/v1/users/profile`
- `PUT /api/v1/users/profile`
- `POST /api/v1/users/change-password` (`old_password`,`new_password`)
- `POST /api/v1/users/upload-avatar`（未实现上传逻辑，返回占位信息）
- `DELETE /api/v1/users/account`
- 管理员：`/api/v1/admin/users` + `/:id` + `/ban` `/unban`

## 会议接口

> 需 JWT；示例参数按 `meeting-service` 实现。

- `POST /api/v1/meetings`（创建）
- `GET /api/v1/meetings`（列表）
- `GET /api/v1/meetings/:id`
- `PUT /api/v1/meetings/:id`
- `DELETE /api/v1/meetings/:id`
- `POST /api/v1/meetings/:id/join` / `leave`
- `GET /api/v1/meetings/:id/participants`
- `POST /api/v1/meetings/:id/participants`
- `DELETE /api/v1/meetings/:id/participants/:user_id`
- `PUT /api/v1/meetings/:id/participants/:user_id/role`
- 房间/录制/聊天：`/room`、`/recording/*`、`/messages`
- 我的会议：`/api/v1/my/meetings`、`/upcoming`、`/history`
- 管理：`/api/v1/admin/meetings`、`/stats`、`/force-end`

## 信令（WebSocket）

- URL：`ws(s)://<host>/ws/signaling?user_id=<uid>&meeting_id=<mid>&peer_id=<uuid>&token=<jwt>`
- 消息类型见 `frontend/dist/app.js`（`WS_TYPES`）：`JOIN`、`OFFER`、`ANSWER`、`ICE`、`CHAT` 等。
- 服务端状态接口（需 JWT）：`/api/v1/sessions/*`、`/api/v1/stats/*`。

## 媒体服务

- 媒体：`/api/v1/media/upload|download/:id|info/:id|process|delete`
- WebRTC 辅助：`/api/v1/webrtc/{answer,ice-candidate,room/*,peer/*}`
- 录制：`/api/v1/recording/{start,stop,status/:id,list,download/:id}`、`DELETE /api/v1/recording/:id`
- 缩略图：`POST /api/v1/ffmpeg/thumbnail`
- AI 状态：`GET /api/v1/ai/{connectivity,streams,streams/:id}`

## AI 推理

直接请求 `ai-inference-service` 或经网关：
- `GET /api/v1/ai/health`
- `GET /api/v1/ai/info`
- `POST /api/v1/ai/asr`（`audio_data` base64，`format`=`wav`，`sample_rate`=16000）
- `POST /api/v1/ai/emotion`（音频或 `text`）
- `POST /api/v1/ai/synthesis`
- `POST /api/v1/ai/setup`（预热：`meeting_id`、`models` 可选）
- `POST /api/v1/ai/analyze`（通用任务：`task_type` + `input_data`）
- `POST /api/v1/ai/batch`（批量任务数组）

## 前端调用示例（fetch）

```js
async function apiFetch(path, {method="GET", body, auth=true, csrf=false} = {}) {
  const headers = {"Accept":"application/json"};
  if (body) headers["Content-Type"] = "application/json";
  if (auth && token) headers["Authorization"] = `Bearer ${token}`;
  if (csrf) headers["X-CSRF-Token"] = await getCsrfToken();
  const res = await fetch(path, {method, headers, body: body ? JSON.stringify(body) : undefined});
  const json = await res.json();
  if (!res.ok || json.code >= 400) throw new Error(json.message || res.statusText);
  return json.data;
}
```

## 错误处理

- 401/403：Token 过期或缺失 → 重新登录/刷新。
- 400：参数错误，按 `message` 提示修正。
- 5xx：服务异常；检查后端日志、Triton 状态或依赖服务。
