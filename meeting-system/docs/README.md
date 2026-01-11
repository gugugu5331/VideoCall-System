# 📚 文档中心（meeting-system）

围绕当前代码与配置的官方文档索引：架构、数据、API、部署、开发/测试、Web 客户端与 AI。默认入口参考 `../README.md`。

## 导航速览

- 总览：`../../README.md`（仓库根）、`../README.md`（后端/运维）
- 架构与数据：`ARCHITECTURE_DIAGRAM.md`、`BACKEND_ARCHITECTURE.md`、`DATABASE_SCHEMA.md`
- API：`API/API_DOCUMENTATION.md`
- 部署：`DEPLOYMENT/README.md`（本地/远程/K8s/GPU AI、模型与上游说明）
- 开发与测试：`DEVELOPMENT/README.md`
- 客户端：`CLIENT/README.md`（调用、通信、AI/特效）

## 分类

- **架构**：系统拓扑、服务职责、数据与可观测性
- **数据**：PostgreSQL/Redis/Mongo/MinIO/etcd 结构与约定
- **API**：用户、会议、信令、媒体、AI 端点清单与示例
- **部署**：本地 compose、远程一键部署、GPU/Triton 节点、K8s
- **开发/测试**：任务调度、AI 服务实现、集成/E2E 测试流程
- **客户端**：Web 端调用约定、WS/WebRTC、AI 特性、SEI/贴图扩展

## 快速上手

```bash
cd ..  # repo 根
docker compose -f meeting-system/docker-compose.yml up -d
```

默认入口 `http://localhost:8800`。验证：

```bash
curl http://localhost:8800/health
docker compose -f meeting-system/docker-compose.yml ps
```

AI 组件需参考部署文档启用并准备 Triton 模型。端口/凭据以对应 compose 文件为准。

## 贡献与更新

- 更新端口、上游、队列或存储配置时，请同步修改相关文档与 compose/kustomize。
- 新增文档遵循现有分类，并在本索引补充链接。
- 文档中的命令默认在仓库根或 `meeting-system/` 目录执行，需结合自身环境调整变量。

## 维护提示

- 端口、上游与环境变量以 compose 与 `backend/config/*.yaml` 为准；调整后同步更新相关文档。
- 修改队列类型（Kafka/内存）、AI 上游或存储凭据时，需同步修改对应配置、Nginx upstream 与部署清单。
- 新增文档请按分类补充到本索引；不再存在 Qt/原生客户端说明。
