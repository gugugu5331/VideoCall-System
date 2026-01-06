# 🎥 Meeting System

企业级 WebRTC 会议系统，提供安全的音视频会议、AI 语音识别/情感/合成检测，支持容器化一键启动。项目当前包含后端微服务、预构建的 Web 客户端（`meeting-system/frontend/dist`），以及完整的运维监控栈；不再包含 Qt6 客户端或其他未提交的子项目。

## 🏗️ 当前架构

- **网关**：Nginx（HTTP/HTTPS，端口 `8800/443`）
- **微服务**：`user-service` (8080)、`signaling-service` (8081)、`meeting-service` (8082)、`media-service` (8083)、`ai-inference-service` (8085)
- **基础设施**：PostgreSQL、Redis、MongoDB、MinIO、etcd
- **AI 推理**：Triton Inference Server（GPU，端口 8000）+ Go AI Inference Service
- **可观测性**：Prometheus (8801)、Alertmanager (8802)、Jaeger (8803)、Grafana (8804)、Loki/Promtail (8805)
- **前端**：`frontend/dist` 静态 Web 客户端，通过 Nginx 提供

详见 `meeting-system/docker-compose.yml` 与 `meeting-system/docs/ARCHITECTURE_DIAGRAM.md`。

## 🚀 快速启动

```bash
git clone https://github.com/gugugu5331/VideoCall-System.git
cd VideoCall-System/meeting-system
docker compose up -d   # 需 Docker 20+ / Compose 2+
```

访问入口：
- 网关 & Web 客户端：`http://localhost:8800`
- Prometheus/Grafana/Jaeger：`http://localhost:8801/8804/8803`
- MinIO 控制台：`http://localhost:9001`（账号密码均为 `minioadmin`）

> 生产环境请通过环境变量设置 `JWT_SECRET`、允许的 CORS 源和 TLS 证书（位于 `meeting-system/nginx/ssl/`）。


## 🔌 主要能力

- WebRTC SFU 信令与房间管理（WebSocket `/ws/signaling`）
- 会议创建/加入/退出、参与者管理、录制元数据
- 对象存储 (MinIO) 与媒体上传/管理
- AI 推理接口：`/api/v1/ai/{asr,emotion,synthesis,health,info,batch}`
- 监控与日志：Prometheus + Grafana + Loki + Jaeger，容器健康检查

## 📚 文档

- 入口索引：`meeting-system/docs/README.md`
- 架构说明：`meeting-system/docs/ARCHITECTURE_DIAGRAM.md`
- API：`meeting-system/docs/API/API_DOCUMENTATION.md`
- 部署：`meeting-system/docs/DEPLOYMENT/README.md`
- 开发与测试：`meeting-system/docs/DEVELOPMENT/README.md`
- Web 客户端：`meeting-system/docs/CLIENT/README.md`


