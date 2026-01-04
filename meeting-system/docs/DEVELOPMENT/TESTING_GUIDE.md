# 微服务集成测试

## 快速开始

```bash
cd meeting-system/backend/tests
./run_all_tests.sh          # 完整验证（推荐）
./quick_integration_test.sh # 仅连通性/基础检查
./test_nginx_gateway.sh     # 网关路由校验
```

运行前确保依赖容器已启动（PostgreSQL、Redis、etcd、Nginx 等），可直接使用 `docker compose up -d`。

## 测试脚本说明

- **run_all_tests.sh**：整合服务发现、网关、HTTP 端点与基础业务流程。
- **quick_integration_test.sh**：快速连通性，耗时短，适合开发自检。
- **test_nginx_gateway.sh**：验证网关转发与 upstream 配置。
- 其他脚本：`test_services_direct.sh`、`verify_ai_service.sh`、`complete_integration_test.py` 等覆盖特定场景。

## 参考流程

1. `docker compose up -d`（或按需启动基础设施/服务）
2. `cd meeting-system/backend/tests && ./run_all_tests.sh`
3. 查看输出/日志；若失败，排查容器状态与配置（JWT_SECRET、端口等）

## 故障排查

1. 检查容器：`docker compose ps`
2. 查看服务日志：`docker compose logs -f <service>`
3. 直接调用健康检查：`curl http://localhost:8800/health`
4. 重新运行测试，确保已拉起依赖

测试结果未固化在文档中，请按需执行并记录到自己的环境。

## 服务列表

### 基础设施服务
- PostgreSQL (数据库)
- Redis (缓存)
- MongoDB (文档存储)
- etcd (服务注册中心)
- MinIO (对象存储)

### 微服务
- user-service (用户服务) - 端口 8080
- meeting-service (会议服务) - 端口 8082
- signaling-service (信令服务) - 端口 8081
- media-service (媒体服务) - 端口 8083
- ai-inference-service (AI 推理服务) - 端口 8085

### API 网关
- Nginx (API 网关) - 端口 8800

---

## 故障排查

### 如果测试失败

1. **检查 Docker 容器状态**
   ```bash
   docker ps -a
   ```

2. **检查服务日志**
   ```bash
   docker compose logs -f <container-name>
   ```

3. **重启服务**
   ```bash
   cd meeting-system
   docker compose restart
   ```

4. **重新运行测试**
   ```bash
   cd meeting-system/backend/tests
   ./run_all_tests.sh
   ```

---

## 详细报告

查看完整的测试报告：
```bash
cat INTEGRATION_TEST_REPORT.md
```

---

## 生产就绪

微服务架构已经具备以下生产环境能力：
- ✅ 服务自动注册与发现 (etcd)
- ✅ API 网关路由 (Nginx)
- ✅ 健康检查与故障恢复
- ✅ 多实例部署支持
- ✅ 真实的数据持久化
- ✅ 完整的服务间通信
- ✅ 负载均衡 (Nginx upstream)

---

**最后更新**: 2025-10-05 00:54  
**测试版本**: v1.1.0
