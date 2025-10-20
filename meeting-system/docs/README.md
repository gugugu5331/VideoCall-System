# 📚 会议系统文档中心

欢迎来到智能视频会议系统的文档中心。本目录包含系统的所有技术文档、API 参考、部署指南和开发资源。

---

## 🚀 快速导航

### 📖 新手入门
- **[项目主 README](../../README.md)** - 项目总体介绍和快速开始
- **[后端系统 README](../README.md)** - 后端服务详细说明
- **[客户端 README](../../qt6-client/README.md)** - Qt6 客户端文档

### 🎯 核心文档
| 文档 | 描述 |
|------|------|
| [系统架构图](ARCHITECTURE_DIAGRAM.md) | 系统整体架构设计（含详细分层说明） |
| [后端架构详解](BACKEND_ARCHITECTURE.md) | 微服务架构、服务间通信、数据流、可观测性 |
| [技术栈](../../README.md#-技术栈) | 使用的技术和框架 |
| [项目结构](../../README.md#-项目结构) | 代码组织和目录说明 |

---

## 📁 文档分类

### 🔌 API 文档 (`API/`)
完整的 API 接口文档和参考资料。

| 文档 | 描述 |
|------|------|
| [API 文档](API/API_DOCUMENTATION.md) | 所有 API 端点的完整文档 |

**相关资源**:
- [客户端 API 使用指南](CLIENT/API_USAGE_GUIDE.md) - 如何在客户端中调用 API

---

### 🚀 部署指南 (`DEPLOYMENT/`)
系统部署、配置和运维相关文档。

| 文档 | 描述 | 适用场景 |
|------|------|---------|
| [远程部署指南](DEPLOYMENT/REMOTE_DEPLOYMENT_GUIDE.md) | 在远程服务器上部署系统 | 生产环境部署 |
| [AI 模型部署指南](DEPLOYMENT/AI_MODELS_DEPLOYMENT_GUIDE.md) | AI 模型的下载和配置 | AI 功能启用 |

**快速开始**:
```bash
# 本地开发环境
docker-compose up -d

# 查看服务状态
docker-compose ps
```

---

### 🔧 开发指南 (`DEVELOPMENT/`)
核心模块设计、实现和测试文档。

| 文档 | 描述 | 适用对象 |
|------|------|---------|
| [消息队列系统](DEVELOPMENT/QUEUE_SYSTEM.md) | Redis 消息队列架构设计 | 后端开发者 |
| [消息队列使用指南](DEVELOPMENT/QUEUE_SYSTEM_USAGE_GUIDE.md) | 如何使用消息队列 | 后端开发者 |
| [任务分发器指南](DEVELOPMENT/TASK_DISPATCHER_GUIDE.md) | 任务分发和调度 | 后端开发者 |
| [AI 推理服务](DEVELOPMENT/AI_INFERENCE_SERVICE.md) | AI 推理微服务文档 | 后端开发者 |
| [测试指南](DEVELOPMENT/TESTING_GUIDE.md) | 后端集成测试 | QA/开发者 |
| [E2E 测试指南](DEVELOPMENT/E2E_TESTING_GUIDE.md) | 端到端测试执行 | QA/开发者 |

**核心模块**:
- **消息队列**: 基于 Redis 的分布式消息队列系统
- **AI 推理**: 集成 Edge-LLM-Infra 的 AI 推理服务
- **WebRTC SFU**: 媒体转发单元实现

---

### 💻 客户端文档 (`CLIENT/`)
Qt6 客户端相关的文档和指南。

| 文档 | 描述 |
|------|------|
| [API 使用指南](CLIENT/API_USAGE_GUIDE.md) | 客户端如何调用后端 API |
| [通信设计](CLIENT/COMMUNICATION_DESIGN.md) | 客户端-服务器通信架构 |
| [AI 功能](CLIENT/AI_FEATURES.md) | 客户端 AI 功能实现 |
| [贴图特效](CLIENT/STICKER_FEATURE.md) | 视频贴图特效功能 |

---

### 🎓 面试参考 (`INTERVIEW/`)
技术面试相关的参考资料和题库。

| 文档 | 描述 |
|------|------|
| [快速参考](INTERVIEW/QUICK_REFERENCE.md) | 面试快速参考手册 |
| [技术问题](INTERVIEW/TECHNICAL_QUESTIONS.md) | 常见技术面试题 |
| [基础知识答案](INTERVIEW/REFERENCE_ANSWERS_BASIC.md) | 基础知识参考答案 |
| [项目实践答案](INTERVIEW/REFERENCE_ANSWERS_PRACTICE.md) | 项目实践参考答案 |
| [通信模式](INTERVIEW/COMMUNICATION_PATTERNS.md) | 系统通信模式总结 |
| [同步 vs 异步](INTERVIEW/SYNC_VS_ASYNC.md) | 同步异步通信对比 |

---

## 🔗 相关链接

### 项目文档
- [项目主 README](../../README.md) - 项目总体介绍
- [后端系统 README](../README.md) - 后端详细说明
- [Qt6 客户端 README](../../qt6-client/README.md) - 客户端文档
- [Edge-LLM-Infra README](../Edge-LLM-Infra-master/node/llm/README.md) - AI 推理框架

### 外部资源
- [Go 官方文档](https://golang.org/doc/)
- [Gin Web 框架](https://gin-gonic.com/)
- [WebRTC 文档](https://webrtc.org/)
- [Qt6 文档](https://doc.qt.io/qt-6/)

---

## 📝 文档维护指南

### 文档结构规范
```
docs/
├── README.md                    # 文档中心索引（本文件）
├── API/                         # API 文档
├── DEPLOYMENT/                  # 部署指南
├── DEVELOPMENT/                 # 开发指南
├── CLIENT/                      # 客户端文档
└── INTERVIEW/                   # 面试参考
```

### 文档编写规范
1. **使用 Markdown 格式** - 所有文档使用 `.md` 扩展名
2. **清晰的结构** - 使用标题、列表、表格等组织内容
3. **代码示例** - 提供实际的代码示例和用法
4. **更新链接** - 修改文件位置时更新所有相关链接
5. **版本信息** - 在文档中标注适用的版本号

### 文档更新流程
1. 在相应的目录中编辑或创建文档
2. 更新本 README.md 中的链接和索引
3. 提交 Git 提交，使用清晰的提交信息
4. 推送到远程仓库

### 命名规范
- 文件名使用大写字母和下划线: `QUEUE_SYSTEM.md`
- 避免使用特殊字符和空格
- 使用有意义的名称，清晰表达内容

---

## 📊 文档统计

| 分类 | 文档数 | 描述 |
|------|--------|------|
| API | 1 | API 接口文档 |
| DEPLOYMENT | 2 | 部署和配置指南 |
| DEVELOPMENT | 7 | 开发和测试指南 |
| CLIENT | 4 | 客户端相关文档 |
| INTERVIEW | 6 | 面试参考资料 |
| **总计** | **20** | **所有文档** |

---

## ❓ 常见问题

### Q: 如何快速开始开发？
A: 请参考 [项目主 README](../../README.md) 的快速开始部分。

### Q: 如何部署到生产环境？
A: 请参考 [远程部署指南](DEPLOYMENT/REMOTE_DEPLOYMENT_GUIDE.md)。

### Q: 如何使用消息队列？
A: 请参考 [消息队列使用指南](DEVELOPMENT/QUEUE_SYSTEM_USAGE_GUIDE.md)。

### Q: 如何运行测试？
A: 请参考 [测试指南](DEVELOPMENT/TESTING_GUIDE.md) 和 [E2E 测试指南](DEVELOPMENT/E2E_TESTING_GUIDE.md)。

### Q: 如何调用 API？
A: 请参考 [API 文档](API/API_DOCUMENTATION.md) 和 [客户端 API 使用指南](CLIENT/API_USAGE_GUIDE.md)。

---

## 📞 获取帮助

- 📖 查看相关文档
- 🔍 搜索关键词
- 💬 提交 Issue
- 📧 联系开发团队

---

**最后更新**: 2025-10-20
**文档版本**: 1.0
**维护者**: 开发团队

