# Web 客户端 AI 功能说明

前端可在通话中触发实时或手动 AI 检测，依赖已部署的 `ai-inference-service` 与 Triton。未启用 AI 时相关按钮应隐藏/禁用。

## 能力

- **实时检测**：对当前说话人定期截取音频，调用 `/api/v1/ai/{asr,emotion,synthesis}`，在 UI 中显示字幕/情绪/伪造标签
- **手动检测**：
  - 文本情绪：`POST /api/v1/ai/emotion`（`text`）
  - 文件 ASR：`POST /api/v1/ai/asr`（`audio_data` base64，`format`、`sample_rate`）
  - 深度伪造：`POST /api/v1/ai/synthesis`
- **健康信息**：`GET /api/v1/ai/{health,info}`

## 数据流（浏览器侧）

```
麦克风/远端音频 → MediaRecorder → PCM/WAV → base64 → /api/v1/ai/*
AI 响应 → 字幕/标签 → 参与者列表标注
```

## 关键实现位置

- `frontend/dist/app.js`：`apiFetch` 封装、AI 按钮事件、实时检测逻辑
- `frontend/dist/index.html`：AI 控件（实时开关、健康检查、文件上传）
- `frontend/dist/styles.css`：AI 状态徽标、字幕样式

## 调用约定

- 默认无需 JWT；若后端开启鉴权，按其他 API 一致带上 Token
- 常用字段：`audio_data`（base64）、`format`=`wav`、`sample_rate`=16000、`text`（情绪文本）
- 典型返回字段：ASR `text/ confidence`；Emotion `emotion/ emotions/ confidence`；Synthesis `is_synthetic/ confidence/ score`

## 使用建议

- 生产场景强制 HTTPS，避免浏览器阻止音视频采集
- 推荐 16kHz 单声道 WAV，减少带宽并保持推理准确度
- 前端默认超时 10s；模型冷启动时可在服务端调整 `timeout_ms`
- 控制实时检测开关，避免并发发送过多请求；失败时应降级为手动检测
- 未启用 AI 时，UI 应禁用相关按钮并提示“AI 未启用”，以免用户误触发错误请求
