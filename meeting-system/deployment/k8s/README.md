# Kubernetes + Kafka 快速体验

Kustomize 覆盖运行核心组件（PostgreSQL/Redis/etcd/Kafka/MinIO/MongoDB + Go 服务 + Nginx + AI 可选），用于在集群中演示队列/事件与监控链路。

## 前置

- 可用的 Kubernetes 集群（kubectl 已配置）
- 镜像仓库推送权限（替换 `<registry>`）
- 默认使用 `emptyDir`，生产需提供 PVC/Ingress/证书
- 如需外部 Kafka/数据库/对象存储，可在 `services.yaml` 中替换相应主机与凭据，并移除内置组件

## 构建与推送镜像

从仓库根目录：

```bash
docker build -t <registry>/meeting-system/user-service:latest -f backend/user-service/Dockerfile backend
docker build -t <registry>/meeting-system/signaling-service:latest -f backend/signaling-service/Dockerfile backend
docker build -t <registry>/meeting-system/meeting-service:latest -f backend/meeting-service/Dockerfile backend
docker build -t <registry>/meeting-system/media-service:latest -f backend/media-service/Dockerfile backend
docker build -t <registry>/meeting-system/ai-inference-service:latest -f backend/ai-inference-service/Dockerfile backend
# Nginx 镜像需包含 frontend/dist 与 nginx 配置，构建后替换 services.yaml 中的镜像
```

## 应用清单

```bash
# 更新 kustomization.yaml 中的密钥 (JWT_SECRET 等) 与镜像名
kubectl apply -k deployment/k8s
# 本地调试可端口转发
kubectl -n meeting-system port-forward svc/meeting-nginx 8800:80
```

ConfigMap 由仓库配置生成（`backend/config/*.yaml`、`ai-inference-service/config`、`nginx/conf*`）。如需调整端口/上游，请先修改源配置文件。

AI 可选：如不需要，移除或注释 `services.yaml` 中的 `ai-inference-service` 部分，或在镜像构建阶段跳过相关镜像。

验证：
```bash
kubectl -n meeting-system get pods
kubectl -n meeting-system get svc
curl http://localhost:8800/health   # 若已做端口转发
```

Kafka 说明：
- 默认单节点 KRaft（`kafka:9092`），服务使用 `MESSAGE_QUEUE_TYPE=kafka`、`EVENT_BUS_TYPE=kafka`。
- 切换到外部 Kafka：修改 `services.yaml` 中相关环境变量（如 `KAFKA_BROKERS`），并移除内置 Kafka StatefulSet/Service。
- 若使用 Redis 队列，请同步修改服务环境变量为 `memory`/`local` 模式，并调整 ConfigMap。

## 注意事项

- Kafka：单节点 KRaft，服务默认 `MESSAGE_QUEUE_TYPE=kafka` 与 `EVENT_BUS_TYPE=kafka`。若要使用 Redis 队列，请修改 `services.yaml` 对应环境变量。
- 存储：当前使用 `emptyDir`，生产需要为 Postgres/MinIO/MongoDB/Kafka 配置 PVC。
- 前端：Nginx 部署期望镜像已包含 `frontend/dist`，或挂载卷提供静态资源。
- TLS/Ingress：示例使用 LoadBalancer 暴露 `meeting-nginx`；生产请改用 Ingress + 证书管理。
