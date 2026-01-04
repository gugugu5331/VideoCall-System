# 🔧 开发指南

本目录包含核心模块设计、实现和测试相关的文档。

## 📖 文档列表

- **核心**：`AI_INFERENCE_SERVICE.md`、`TASK_DISPATCHER_GUIDE.md`
- **测试**：`TESTING_GUIDE.md`、`E2E_TESTING_GUIDE.md`

## 🏗️ 核心模块

### 任务分发与队列
代码位于 `backend/shared/queue`，使用 Redis 实现消息队列与 Pub/Sub；调度与任务分发参考 `TASK_DISPATCHER_GUIDE.md`。未提供单独队列文档，使用时以源码为准。

### AI 推理服务
基于 Triton/TensorRT 的 AI 推理微服务，支持：
- 语音识别 (ASR)
- 情感检测
- 合成检测

**相关文档**:
- [AI_INFERENCE_SERVICE.md](AI_INFERENCE_SERVICE.md) - 服务文档

### WebRTC SFU
媒体转发单元实现，支持：
- 多人音视频通话
- 媒体流转发
- 会议录制

## 🧪 测试

### 运行单元测试
```bash
cd meeting-system/backend
go test ./...
```

### 运行集成测试
```bash
cd meeting-system/backend/tests
./run_all_tests.sh          # 或 quick_integration_test.sh / test_nginx_gateway.sh
```

### 运行 E2E 测试
```bash
cd meeting-system/tests
./e2e_queue_integration_test.sh   # 结合队列/信令
# 其他 python 脚本参考目录说明
```

## 📚 开发流程

1. **理解架构** - 阅读相关模块文档
2. **本地开发** - 在本地环境中开发和测试
3. **单元测试** - 编写和运行单元测试
4. **集成测试** - 运行集成测试验证
5. **E2E 测试** - 执行端到端测试
6. **提交代码** - 提交 Git 提交

## 🔍 常见开发任务

### 添加新的 API 端点
1. 在相应的 service 中实现业务逻辑
2. 在 handler 中添加 HTTP 处理器
3. 在路由中注册端点
4. 编写测试用例
5. 更新 API 文档

### 添加新的消息队列任务
1. 定义任务类型
2. 实现任务处理器
3. 在需要的地方发布任务
4. 编写测试用例

### 集成新的 AI 模型
1. 准备 Triton 模型仓库并放入模型目录
2. 在 AI Inference Service 中配置 `model_name` / 输入输出节点
3. 在处理器中接入预处理/后处理逻辑
4. 编写测试用例
5. 更新文档

## 📚 相关文档

- [API 文档](../API/README.md) - API 接口参考
- [部署指南](../DEPLOYMENT/README.md) - 部署和配置
- [客户端文档](../CLIENT/README.md) - 客户端相关
- [文档中心](../README.md) - 所有文档

## 🔗 相关链接

- [项目主 README](../../README.md)
- [后端系统 README](../README.md)
- [Go 官方文档](https://golang.org/doc/)
- [Gin Web 框架](https://gin-gonic.com/)
