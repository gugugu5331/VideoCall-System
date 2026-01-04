# 文档同步记录

本次调整将文档与当前仓库实际内容保持一致，清理了缺失目录/过时引用，并更新了端口、服务名称和客户端形态。

## 关键变更

- **入口文档**：根 README 与 `meeting-system/README.md` 现描述真实的服务列表、端口、前端位置（`frontend/dist`）与部署方式。
- **索引修正**：`docs/README.md` 去除不存在的目录（如 `INTERVIEW/`、Qt 客户端），仅保留现有文件。
- **架构文档**：`ARCHITECTURE_DIAGRAM.md` 与 `BACKEND_ARCHITECTURE.md` 对齐 `docker-compose.yml`，删除通知服务/Qt 客户端等无效节点。
- **API 文档**：`API/README.md`、`API/API_DOCUMENTATION.md` 更新为实际端点（user/meeting/signaling/media/ai），基础 URL 改为 `http://localhost:8800`。
- **客户端文档**：改为 Web 客户端说明，移除 Qt/C++ API；新增当前 AI/通信/贴图状态说明。

## 现有文档结构

```
docs/
├── README.md
├── ARCHITECTURE_DIAGRAM.md
├── BACKEND_ARCHITECTURE.md
├── DATABASE_SCHEMA.md
├── API/
├── DEPLOYMENT/
├── DEVELOPMENT/
└── CLIENT/
```

## 后续维护建议

- 新增或修改端口/服务时同步更新 `docker-compose.yml` 与相关文档。
- 若引入新客户端形态（Native/移动端），单独补充目录并在 `docs/README.md` 增加索引。
- 生产部署前确认环境变量（`JWT_SECRET` 等）已设置，避免文档与配置偏差。
