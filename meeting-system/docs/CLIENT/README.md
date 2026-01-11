# 💻 Web 客户端文档

仓库已包含构建好的前端（`frontend/dist`），由 Nginx 同源发布，默认入口 `http://localhost:8800`。无需单独编译即可体验。

## 文档索引

- [API_USAGE_GUIDE.md](API_USAGE_GUIDE.md)：调用流程与示例
- [COMMUNICATION_DESIGN.md](COMMUNICATION_DESIGN.md)：HTTP/WS/WebRTC 设计
- [AI_FEATURES.md](AI_FEATURES.md)：实时与手动 AI 功能
- [VIDEO_EFFECTS_SEI.md](VIDEO_EFFECTS_SEI.md)：H264 SEI 美颜/滤镜参数示例
- [STICKER_FEATURE.md](STICKER_FEATURE.md)：贴图/虚拟形象扩展现状

## 客户端概览

- **位置**：`meeting-system/frontend/dist`（`index.html`、`app.js`、`styles.css`）
- **能力**：注册/登录、创建/加入会议、音视频/屏幕共享、聊天、房间和参与者状态、AI 字幕/标签、可选 SEI 美颜示例
- **依赖**：同源 API `/api/v1/*` 与 `ws(s)://<host>/ws/signaling`；生产场景建议 HTTPS 以启用摄像头/麦克风
- **浏览器要求**：现代 Chromium/Firefox；SEI 示例依赖 Encoded Insertable Streams，仅在支持的浏览器上可用

## 使用步骤

1. 在 `meeting-system` 目录执行 `docker compose up -d`
2. 打开 `http://localhost:8800`（生产请使用 HTTPS 域名）
3. 注册或登录，创建/加入会议；若需 AI 功能，请确保后端已启用 `ai-inference-service` + Triton 上游

## 调试提示

- 前端使用同源相对路径，无需额外环境变量；调整网关域名/端口即可。
- 浏览器不支持 Encoded Streams 时，SEI 示例会被自动降级，通话功能不受影响。
- 401/403 多因缺少 JWT 或密钥不一致；确认 `JWT_SECRET` 已在后端设置并重新登录。
- 页面异常可通过浏览器控制台查看网络/WS 报错；必要时检查 `nginx` 与 `signaling-service` 日志。

更多接口、部署与测试说明请参考上层文档：`../API/README.md`、`../DEPLOYMENT/README.md`、`../DEVELOPMENT/README.md`。
