# 会议系统远程部署指南

## 概述

本指南介绍如何将会议系统部署到远程服务器 `js1.blockelite.cn`，并从本地执行集成测试。

## 远程服务器信息

- **SSH 主机**: js1.blockelite.cn
- **SSH 端口**: 22124
- **SSH 用户**: root
- **SSH 密码**: beip3ius
- **部署目录**: /root/meeting-system-server

## 端口映射配置

远程服务器已配置以下 NAT 端口映射（外网端口 → 服务器内网端口）：

| 外网端口 | 内网端口 | 服务 |
|---------|---------|------|
| 22176 | 8800 | Nginx 网关 (主要 HTTP 入口) |
| 22177 | 8801 | Jaeger UI (分布式追踪) |
| 22178 | 8802 | Prometheus (监控指标) |
| 22179 | 8803 | Alertmanager (告警管理) |
| 22180 | 8804 | Grafana (监控仪表板) |
| 22181 | 8805 | Loki (日志聚合) |

## 快速开始

### 方法 1: 一键部署和测试（推荐）

```bash
cd /root/meeting-system-server/meeting-system
./deploy-and-test-remote.sh
```

这个脚本会自动完成：
1. ✅ 部署代码到远程服务器
2. ✅ 启动所有 Docker 服务
3. ✅ 验证服务可访问性
4. ✅ 执行集成测试
5. ✅ 收集性能指标
6. ✅ 生成测试报告

### 方法 2: 分步执行

#### 步骤 1: 部署到远程服务器

```bash
cd /root/meeting-system-server/meeting-system
./deploy-to-remote.sh
```

这个脚本会：
- 检查依赖工具（rsync, sshpass）
- 测试 SSH 连接
- 准备远程服务器环境
- 同步代码到远程服务器
- 使用 docker-compose.remote.yml 启动所有服务
- 验证服务健康状态

#### 步骤 2: 执行集成测试

```bash
cd /root/meeting-system-server/meeting-system
./run-remote-integration-test.sh
```

这个脚本会：
- 检查 Python 环境
- 测试远程服务器连接
- 运行远程集成测试脚本
- 显示测试结果

## 部署的服务清单

### 基础设施服务
- PostgreSQL (数据库)
- MongoDB (文档存储)
- Redis (缓存和消息队列)
- MinIO (对象存储)
- etcd (服务发现)
- Jaeger (分布式追踪)

### 业务微服务
- user-service (用户服务，内网端口 8080)
- meeting-service (会议服务，内网端口 8082)
- signaling-service (信令服务，内网端口 8081)
- media-service (媒体服务，内网端口 8083)
- ai-service (AI 服务，内网端口 8084)

### AI 推理基础设施
- edge-model-infra (Edge-LLM-Infra 单元管理器，端口 10001)
- ai-inference-worker (AI 推理服务，使用 IPC socket)

### 网关和监控
- Nginx (反向代理网关，内网端口 80)
- Prometheus (监控指标收集)
- Grafana (监控仪表板)
- Alertmanager (告警管理)
- Loki (日志聚合)
- Promtail (日志收集)

## 访问服务

### 外网访问地址

- **Nginx 网关**: http://js1.blockelite.cn:22176
- **Jaeger UI**: http://js1.blockelite.cn:22177
- **Prometheus**: http://js1.blockelite.cn:22178
- **Alertmanager**: http://js1.blockelite.cn:22179
- **Grafana**: http://js1.blockelite.cn:22180
  - 默认用户名: admin
  - 默认密码: admin123
- **Loki**: http://js1.blockelite.cn:22181

### API 访问示例

```bash
# 健康检查
curl http://js1.blockelite.cn:22176/health

# 获取 CSRF Token
curl http://js1.blockelite.cn:22176/api/v1/csrf-token

# 用户注册
curl -X POST http://js1.blockelite.cn:22176/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test@Pass123","email":"test@example.com"}'

# 用户登录
curl -X POST http://js1.blockelite.cn:22176/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test@Pass123"}'
```

## 集成测试

### 测试脚本

远程集成测试脚本位于：
```
backend/tests/complete_integration_test_remote.py
```

### 测试覆盖范围

- ✅ 用户注册和登录
- ✅ 会议创建和加入
- ✅ WebRTC 信令交换
- ✅ AI 服务功能（情绪识别、语音识别）
- ✅ 服务间通信和数据一致性

### 查看测试日志

```bash
cat backend/tests/logs/remote_integration_test.log
```

## 远程服务器管理

### SSH 连接

```bash
ssh -p 22124 root@js1.blockelite.cn
# 密码: beip3ius
```

### 查看服务状态

```bash
# 查看所有容器
docker ps --filter 'name=meeting-'

# 查看特定服务日志
docker logs meeting-nginx --tail 100 -f
docker logs meeting-ai-service --tail 100 -f
docker logs meeting-edge-model-infra --tail 100 -f
```

### 管理服务

```bash
cd /root/meeting-system-server/meeting-system

# 启动所有服务
docker compose -f docker-compose.remote.yml up -d

# 停止所有服务
docker compose -f docker-compose.remote.yml down

# 重启特定服务
docker compose -f docker-compose.remote.yml restart nginx
docker compose -f docker-compose.remote.yml restart ai-service

# 查看服务状态
docker compose -f docker-compose.remote.yml ps

# 查看服务日志
docker compose -f docker-compose.remote.yml logs -f [service-name]
```

### 检查资源使用

```bash
# 查看 Docker 容器资源使用
docker stats --no-stream

# 查看系统资源
top
htop
df -h
```

## 故障排查

### 问题 1: 无法从外网访问服务

**症状**: curl 请求超时或连接被拒绝

**排查步骤**:
1. 检查远程服务器防火墙
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'iptables -L -n'
   ```

2. 验证端口映射
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'netstat -tlnp | grep 8800'
   ```

3. 检查 Nginx 容器状态
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'docker logs meeting-nginx --tail 50'
   ```

### 问题 2: AI 服务返回 "unit call false"

**症状**: AI 服务 API 返回错误

**排查步骤**:
1. 检查 edge-model-infra 是否运行
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'docker ps | grep edge-model-infra'
   ```

2. 验证 IPC socket 文件
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'ls -la /tmp/llm/'
   ```

3. 检查 ai-inference-worker 进程
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'docker exec meeting-ai-inference-worker ps aux | grep python'
   ```

4. 查看日志
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'docker logs meeting-edge-model-infra --tail 100'
   ssh -p 22124 root@js1.blockelite.cn 'docker logs meeting-ai-service --tail 100'
   ```

### 问题 3: 数据库连接失败

**症状**: 服务启动失败，日志显示数据库连接错误

**排查步骤**:
1. 检查 PostgreSQL 容器
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'docker logs meeting-postgres --tail 50'
   ```

2. 测试数据库连接
   ```bash
   ssh -p 22124 root@js1.blockelite.cn 'docker exec meeting-postgres psql -U postgres -d meeting_system -c "SELECT 1;"'
   ```

## 性能优化建议

1. **数据库优化**
   - 配置 PostgreSQL 连接池
   - 添加适当的索引
   - 定期执行 VACUUM

2. **缓存优化**
   - 增加 Redis 内存限制
   - 配置 Redis 持久化策略

3. **网络优化**
   - 启用 Nginx gzip 压缩
   - 配置 HTTP/2
   - 使用 CDN 加速静态资源

4. **监控告警**
   - 配置 Prometheus 告警规则
   - 设置 Grafana 仪表板
   - 启用日志聚合和分析

## 安全注意事项

⚠️ **重要**: 以下配置仅用于测试环境，生产环境需要加强安全措施：

1. **修改默认密码**
   - PostgreSQL: postgres/password
   - MongoDB: admin/password
   - MinIO: minioadmin/minioadmin
   - Grafana: admin/admin123

2. **启用 HTTPS**
   - 配置 SSL/TLS 证书
   - 强制 HTTPS 重定向

3. **网络安全**
   - 限制外网访问的端口
   - 配置防火墙规则
   - 使用 VPN 或堡垒机

4. **数据备份**
   - 定期备份数据库
   - 备份配置文件
   - 测试恢复流程

## 文件说明

- `docker-compose.remote.yml` - 远程服务器专用的 Docker Compose 配置
- `deploy-to-remote.sh` - 远程部署脚本
- `run-remote-integration-test.sh` - 远程集成测试脚本
- `deploy-and-test-remote.sh` - 一键部署和测试脚本
- `backend/tests/complete_integration_test_remote.py` - 远程集成测试 Python 脚本

## 支持

如有问题，请查看：
- 部署日志: `logs/remote_deploy_*.log`
- 测试日志: `logs/remote_test_*.log`
- 测试报告: `logs/remote_deployment_report_*.md`

---

**最后更新**: 2025-10-06

