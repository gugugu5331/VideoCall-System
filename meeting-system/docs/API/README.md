# 🔌 API 文档

所有接口均以网关同源暴露，基础行为以 `docker-compose*.yml` 与服务源码为准。

## 快速使用

- **Base URL**：`http://localhost:8800`
- **认证**：大部分接口需要 `Authorization: Bearer <jwt>`；需要状态变更时可携带 `X-CSRF-Token`（`GET /api/v1/csrf-token`）。
- **响应格式**：`{"code":200,"message":"success","data":{...}}`，错误返回对应 `code/message`。
- **WS 信令**：`ws://<host>/ws/signaling?user_id=...&meeting_id=...&peer_id=...&token=<jwt>`
- **AI**：启用 `ai-inference-service` 后可访问 `/api/v1/ai/*`。

示例（登录获取 JWT）：

```bash
curl -X POST http://localhost:8800/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"Passw0rd!"}'
```

提示：
- 变更类接口若返回 CSRF 相关错误，先调用 `GET /api/v1/csrf-token`，在请求头带上 `X-CSRF-Token`。
- 默认使用 JSON；上传/下载接口按需使用表单或文件流。

## 文档索引

- 端点清单与示例：`API_DOCUMENTATION.md`
- 前端调用示例：`../CLIENT/API_USAGE_GUIDE.md`
- 通信/协议设计：`../CLIENT/COMMUNICATION_DESIGN.md`

## 服务映射

- 用户：`/api/v1/auth/*`、`/api/v1/users/*`、`/api/v1/admin/users/*`
- 会议：`/api/v1/meetings/*`、`/api/v1/my/*`、`/api/v1/admin/meetings/*`
- 信令：`/ws/signaling`、`/api/v1/sessions/*`、`/api/v1/stats/*`
- 媒体：`/api/v1/media/*`、`/api/v1/recording/*`、`/api/v1/webrtc/*`、`/api/v1/ffmpeg/*`
- AI（可选）：`/api/v1/ai/{health,info,asr,emotion,synthesis,setup,batch,analyze}`
