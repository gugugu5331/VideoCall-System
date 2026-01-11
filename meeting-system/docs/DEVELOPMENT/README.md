# 🔧 开发指南

面向后端开发与联调，涵盖核心模块、测试入口与常见任务。仓库主代码位于 `meeting-system/backend`。

## 文档索引

- AI 推理服务：`AI_INFERENCE_SERVICE.md`
- 任务分发/队列：`TASK_DISPATCHER_GUIDE.md`
- 测试：`TESTING_GUIDE.md`、`E2E_TESTING_GUIDE.md`

## 核心模块概览

- **队列/事件**：Kafka 为默认实现（`message_queue.type=kafka`、`event_bus.type=kafka`），封装在 `backend/shared/queue`。内存模式仅用于本地开发。
- **AI 推理**：`backend/ai-inference-service` 通过 Triton 提供 ASR/情绪/合成检测，需独立部署或使用 GPU compose。
- **WebRTC/SFU**：媒体链路由 `media-service` 与 `signaling-service` 协作；录制与媒资落地 Postgres + MinIO。
- **配置**：服务配置位于 `backend/config/*.yaml`（AI 在 `backend/ai-inference-service/config`），可用环境变量覆盖。

## 本地开发流程

1. 启动依赖
   ```bash
   docker compose up -d postgres redis kafka etcd minio
   ```
2. 启动单个服务
   ```bash
   cd meeting-system/backend/user-service
   go run . -config=../config/config.yaml
   ```
3. 前端同源访问 `/api/v1/*` 与 `/ws/signaling`；如需直连服务，调整浏览器地址即可。
4. 如需 AI，单独启动 `deployment/gpu-ai/docker-compose.gpu-ai.yml` 或远程节点。

> 配置可通过环境变量覆盖 `backend/config/*.yaml` 中的字段；如需本地 `.env`，确保未提交敏感信息。

## 测试入口

- 集成测试：`backend/tests/run_all_tests.sh`、`quick_integration_test.sh`、`test_nginx_gateway.sh`
- E2E（含信令/队列）：`tests/e2e_queue_integration_test.{sh,py}`，详见 `E2E_TESTING_GUIDE.md`
- AI 自测：`backend/ai-inference-service/test_ai_service.py`、`scripts/e2e_stream_pcm.sh`

## 常见开发任务

- **新增 API**：实现业务逻辑 → handler → 路由注册 → 测试 → 更新 `docs/API/*`
- **添加队列任务**：在 `shared/queue` 注册处理器，发布任务/事件，确保 Kafka 配置正确
- **接入新模型**：准备 Triton 模型仓库，更新 `ai-inference-service` 配置，补充前后处理逻辑与文档

提交前建议运行：`go test ./...`（对应服务或 shared），必要时运行 `backend/tests/quick_integration_test.sh`。

## 参考链接

- API 文档：`../API/README.md`
- 部署：`../DEPLOYMENT/README.md`
- 客户端：`../CLIENT/README.md`
- 架构：`../ARCHITECTURE_DIAGRAM.md`
