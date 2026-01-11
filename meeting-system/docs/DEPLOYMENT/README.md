# 🚀 部署指南

覆盖本地开发、远程主机、GPU AI 节点与 Kubernetes 示例。所有端口/变量以对应 compose/kustomize 文件为准。

## 方案一览

- **本地快速启动**：`docker compose up -d`（基础链路，默认不启用 AI）
- **启用 AI（单机 GPU）**：`deployment/gpu-ai/docker-compose.gpu-ai.yml`
- **多 GPU 节点上游**：参考 `GPU_AI_NODES.md`，通过 Nginx upstream 扩展
- **远程一键部署**：`quick-deploy-remote.sh`，详见 `REMOTE_DEPLOYMENT_GUIDE.md`
- **Kubernetes Demo**：`deployment/k8s/`，内置 Kafka/监控，需自建镜像与存储，并按需接入外部队列/存储

## 前置要求

- Docker 20+ / Docker Compose v2+（或兼容运行时）
- 可用端口：网关 8800/443，监控 8801~8805，MinIO 9000/9001，Kafka 9092 等
- 强随机 `JWT_SECRET`，正确的数据库/对象存储凭据
- TLS 证书放在 `nginx/ssl/`（生产必需）

## 本地基础版

```bash
cd meeting-system
JWT_SECRET="change-me" ALLOWED_ORIGINS="http://localhost:8800" docker compose up -d
docker compose ps
curl http://localhost:8800/health
```

默认入口：`http://localhost:8800`。Prometheus/Jaeger/Grafana/Loki 分别暴露在 8801/8803/8804/8805。AI 组件在本 compose 中保持注释，避免拉取大镜像。

示例 `.env` 片段：
```
JWT_SECRET=please-change-me
ALLOWED_ORIGINS=http://localhost:8800
POSTGRES_PASSWORD=change-postgres
MINIO_ROOT_PASSWORD=change-minio
```

常用环境变量（不同 compose/kustomize 可覆盖）：
- 数据库：`POSTGRES_USER`、`POSTGRES_PASSWORD`、`POSTGRES_DB`
- MinIO：`MINIO_ROOT_USER`、`MINIO_ROOT_PASSWORD`
- Kafka：`KAFKA_BROKER_ID`、`KAFKA_ADVERTISED_LISTENERS`
- 网关：`ALLOWED_ORIGINS`、证书路径（见 `nginx/ssl`）
- AI：`MODEL_DIR`、`AI_HTTP_PORT` 等（GPU 方案）

## 启用 AI

- 单机 GPU：`docker compose -f deployment/gpu-ai/docker-compose.gpu-ai.yml up -d --build`
- 多节点：使用 `nginx/scripts/gen_ai_inference_service_servers_conf.sh` 生成本地 upstream，详见 `GPU_AI_NODES.md`
- 远程/生产：使用 `docker-compose.remote.yml`（已包含 `ai-inference-service` 与 Triton）或在 GPU 节点单独部署

确保 `backend/ai-inference-service/config/ai-inference-service.yaml` 中的模型名/输入输出与 Triton 模型仓库一致。

## Kubernetes 示例

参考 `deployment/k8s/README.md`：
1. 先构建并推送镜像（用户/会议/信令/媒体/AI/Nginx）
2. 更新 `kustomization.yaml` 中的密钥与镜像名
3. `kubectl apply -k deployment/k8s`

默认使用空存储卷与单节点 Kafka，生产需改为持久化 PVC 和高可用队列。

## 部署检查清单

- [ ] `JWT_SECRET` 已配置且非默认值
- [ ] 数据库/MinIO/Kafka/Redis 凭据已修改
- [ ] 需要的端口已放行（或通过 NAT 映射）
- [ ] 所有容器健康（`docker compose ps` / `kubectl get pods`）
- [ ] 网关健康：`curl http://<host>:8800/health`
- [ ] 监控可访问（Prometheus/Grafana/Jaeger/Loki）
- [ ] 如启用 AI：Triton `/v2/health/ready` 与 `/api/v1/ai/health` 正常

## 常见问题

- **401/403**：检查 `JWT_SECRET`、前端是否携带最新 Token、是否需要 CSRF。
- **AI 报错**：确认 `ai-inference-service` 已启用并能访问 Triton；模型仓库路径/名称与配置一致。
- **录制/上传失败**：核对 MinIO 凭据、桶名称和访问策略。
- **Kafka 未启动**：队列/事件功能会受影响；确保 `kafka` 容器健康或关闭相关功能（配置 `message_queue.type=memory`）。
- **K8s PVC/Ingress 缺失**：示例使用 `emptyDir` 与 LoadBalancer；生产请改用 PVC + Ingress + 证书，并可将 Kafka/DB 指向托管服务以减少运维。

更多细节请查看各子指南：`REMOTE_DEPLOYMENT_GUIDE.md`、`GPU_AI_NODES.md`、`AI_MODELS_DEPLOYMENT_GUIDE.md`。
