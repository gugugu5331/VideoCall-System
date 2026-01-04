# 📚 文档中心

涵盖当前仓库内的架构、API、部署、开发测试和 Web 客户端说明，全部基于 `meeting-system` 目录下的实际代码与配置。

---

## 🚀 导航

- **总览**：`../README.md`（仓库根）、`../README.md`（meeting-system 说明）
- **架构**：`ARCHITECTURE_DIAGRAM.md`、`BACKEND_ARCHITECTURE.md`、`DATABASE_SCHEMA.md`
- **API**：`API/API_DOCUMENTATION.md`
- **部署**：`DEPLOYMENT/README.md`（含远程、GPU/模型指南）
- **开发/测试**：`DEVELOPMENT/README.md`
- **客户端**：`CLIENT/README.md`（Web 客户端说明，静态资源位于 `frontend/dist`）

---

## 📁 分类

### 🔌 API (`API/`)
- `API_DOCUMENTATION.md`：用户/会议/信令/媒体/AI 现有接口列表与示例。

### 🚀 部署 (`DEPLOYMENT/`)
- `README.md`：本地与生产部署概览。
- `REMOTE_DEPLOYMENT_GUIDE.md`：远程主机部署变量与命令。
- `AI_MODELS_DEPLOYMENT_GUIDE.md`、`GPU_AI_NODES.md`：Triton/AI 节点与模型准备。

### 🔧 开发 (`DEVELOPMENT/`)
- 队列/任务/AI 说明：`QUEUE_SYSTEM.md`、`QUEUE_SYSTEM_USAGE_GUIDE.md`、`TASK_DISPATCHER_GUIDE.md`、`AI_INFERENCE_SERVICE.md`
- 测试：`TESTING_GUIDE.md`、`E2E_TESTING_GUIDE.md`

### 💻 客户端 (`CLIENT/`)
- Web 客户端调用与协议：`API_USAGE_GUIDE.md`、`COMMUNICATION_DESIGN.md`
- AI/特效：`AI_FEATURES.md`、`VIDEO_EFFECTS_SEI.md`、`STICKER_FEATURE.md`

---

## 🧭 快速上手

```bash
cd ..                # 仓库根
docker compose -f meeting-system/docker-compose.yml up -d
```

默认入口：`http://localhost:8800`（API + 前端）。更多服务端口见 `docker-compose.yml`。

---

## 📝 维护说明

- 文件位置与名称以当前目录结构为准；不存在 `INTERVIEW/` 或 Qt 客户端文档。
- 更新文档时同步校验链接与端口配置是否与 `docker-compose.yml`、服务源码一致。
- 需要新文档时按现有分类新增，并在本索引补充链接。
