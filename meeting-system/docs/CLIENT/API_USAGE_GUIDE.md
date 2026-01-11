# 客户端 API 使用指南（Web）

前端通过 Nginx 同源访问后端，所有请求均以 `/api/v1/...` 相对路径发起；信令通过 `/ws/signaling`。

## 基础约定

- **Base URL**：`window.location.origin`（默认 `http://localhost:8800`）
- **认证**：`Authorization: Bearer <jwt>`；变更接口可带 `X-CSRF-Token`（`GET /api/v1/csrf-token`）
- **默认头**：`Content-Type: application/json`，`Accept: application/json`

## 登录流程

1) 注册（可选）  
`POST /api/v1/auth/register`，参数：`username`、`email`、`password`、`nickname`

2) 登录  
`POST /api/v1/auth/login` → 保存返回的 `data.token`（JWT）

3) 刷新  
`POST /api/v1/auth/refresh`，携带旧 JWT 获取新 Token

## 常用接口

- **用户**：`GET/PUT /api/v1/users/profile`、`POST /api/v1/users/change-password`、`DELETE /api/v1/users/account`
- **会议**：`POST /api/v1/meetings` 创建 → `POST /api/v1/meetings/:id/join` 加入 → `GET /api/v1/meetings/:id/participants` 查询；离开/结束对应 `leave`/`end`
- **录制/房间**：`POST /api/v1/meetings/:id/room` 创建房间；`POST /recording/{start,stop}` 控制录制；列表 `GET /api/v1/meetings/:id/recordings`
- **聊天**：`GET/POST /api/v1/meetings/:id/messages`
- **信令 WebSocket**：`ws(s)://<host>/ws/signaling?user_id=<uid>&meeting_id=<mid>&peer_id=<uuid>&token=<jwt>`
- **AI（可选）**：`POST /api/v1/ai/{asr,emotion,synthesis}`；启用前须部署 `ai-inference-service`

创建会议示例：

```bash
curl -X POST /api/v1/meetings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Demo","start_time":"2025-06-01T10:00:00Z"}'
```

## 统一 fetch 封装示例

```js
async function apiFetch(path, {method="GET", body, auth=true, csrf=false} = {}) {
  const headers = {"Accept": "application/json"};
  if (body) headers["Content-Type"] = "application/json";
  if (auth && token) headers["Authorization"] = `Bearer ${token}`;
  if (csrf) headers["X-CSRF-Token"] = await getCsrfToken();
  const res = await fetch(path, {method, headers, body: body ? JSON.stringify(body) : undefined});
  const json = await res.json();
  if (!res.ok || json.code >= 400) throw new Error(json.message || res.statusText);
  return json.data;
}
```

## WebSocket 事件（概览）

- 连接：附带 `user_id`、`meeting_id`、`peer_id`、JWT
- 消息类型：`JOIN`、`OFFER`、`ANSWER`、`ICE`、`LEAVE`、`CHAT`（参见前端 `WS_TYPES` 定义）
- 错误处理：服务端会返回错误消息；前端应提示并可尝试重连

## 错误处理提示

- 401/403：Token 缺失或失效 → 重新登录/刷新
- 400：参数错误 → 按 `message` 修正输入
- AI 请求失败：确认已部署 `ai-inference-service` + Triton，并检查音频编码参数

更多协议细节见 `COMMUNICATION_DESIGN.md`，AI 使用见 `AI_FEATURES.md`。
