# 文档重写记录

此次更新统一了仓库内所有文档，聚焦当前代码与部署形态（Go 微服务 + Nginx + Kafka + 可选 AI/Triton），清理了过时的 Python 推理路径和缺失目录引用。

## 主要改动

- **总览与架构**：重写根 README、`meeting-system/README.md`、`docs/ARCHITECTURE_DIAGRAM.md`、`BACKEND_ARCHITECTURE.md`，明确可选 AI 组件、Kafka 队列与端口映射。
- **API 与数据**：更新 `docs/API/*`、`DATABASE_SCHEMA.md`，对齐现有端点、存储职责与 Kafka/Redis 约定。
- **客户端**：重新编写 Web 调用/通信/AI/SEI/贴图文档，强调同源访问与降级策略。
- **部署**：新增完整的本地/远程/GPU/K8s 说明，替换过期的模型部署指南为 Triton 方案，整理 GPU 节点与自动化脚本用法。
- **开发与测试**：刷新测试流程、队列指南、AI 服务文档，去除不存在的报告文件与历史日期。
- **附加**：更新媒体测试文件说明、清理无效示例列表。

## 细化补充（本次）

- 增加前置要求、环境变量与故障排查提示（根 README、meeting-system/README、部署文档）。
- 在架构、数据、API、客户端文档中补充调用示例、浏览器要求、索引与性能建议。
- 部署与 AI 模型文档添加验证命令、.env 示例、上游负载与健康检查步骤。

## 后续维护建议

- 修改端口、上游或配置时同步更新对应文档与 compose/kustomize。
- 准备或更换 AI 模型时先更新 `ai-inference-service` 配置与 Triton 仓库，再补充文档说明。
- 引入新客户端形态或存储结构时按现有分类新增文档并补充索引。
