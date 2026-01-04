# Web 客户端 AI 功能说明

前端已构建的 `frontend/dist` 在通话时可触发实时/手动 AI 检测，依赖 `ai-inference-service` 暴露的 `/api/v1/ai/*` 接口。

## 支持的能力

- **实时检测（当前说话人）**：ASR + 情绪 + 合成检测；结果以字幕/标签形式展示。
- **手动检测**：
  - 文本情绪：`POST /api/v1/ai/emotion`（`text` 字段）
  - 文件 ASR：`POST /api/v1/ai/asr`（音频 base64，`format`、`sample_rate`）
  - 文件合成检测：`POST /api/v1/ai/synthesis`
- **健康与信息**：`GET /api/v1/ai/{health,info}` 按钮触发。

## 数据流（前端）

```
麦克风/远端音频 → MediaRecorder → PCM/WAV → base64 → /api/v1/ai/{asr,emotion,synthesis}
AI 响应 → UI 字幕/标签 → 参与者列表标注
```

## 关键实现位置

- `frontend/dist/app.js`：`apiFetch`（同源请求），AI 按钮事件与实时检测逻辑。
- `frontend/dist/index.html`：AI 控件（实时检测开关、健康检查、文件上传）。
- `frontend/dist/styles.css`：AI 状态标签、字幕样式。

## 调用约定

- 所有 AI 接口默认不需要 JWT（当前实现）；如在生产环境收紧权限，可在后端开启 JWT 校验。
- 请求体字段：
  - `audio_data`：base64 编码音频（WAV/PCM）
  - `format`：`wav` / `pcm`（默认为 `wav`）
  - `sample_rate`：默认 `16000`
  - `text`：情绪分析文本（可选）
- 响应字段常见：
  - ASR：`text`, `confidence`
  - Emotion：`emotion`, `emotions`(score map), `confidence`
  - Synthesis：`is_synthetic`, `confidence`, `score`

## 使用建议

- **HTTPS 环境**：浏览器启用麦克风/摄像头需安全上下文（自签证书亦可）。
- **音频格式**：推荐 16kHz 单声道 WAV，降低带宽与延迟。
- **超时与重试**：前端默认 10s 超时；若模型加载较慢可在后端配置加长 `timeout_ms`。
- **并发**：实时检测开关在 UI 中控制，避免同时发起过多请求。
