# 远程部署 - 下一步操作指南

## 当前状态

✅ **已完成**:
- 创建了所有必要的部署脚本和配置文件
- 代码已成功传输到远程服务器
- Docker Compose 构建命令已执行

🔄 **进行中**:
- Docker 镜像构建和服务启动（预计需要 10-20 分钟）

## 立即执行的操作

### 步骤 1: 检查 Docker 构建进度

```bash
# 检查正在运行的容器数量
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker ps --filter 'name=meeting-' | wc -l"

# 查看所有容器状态
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker ps -a --filter 'name=meeting-' --format 'table {{.Names}}\t{{.Status}}'"
```

**预期结果**: 应该看到约 20+ 个容器（包括基础设施和微服务）

### 步骤 2: 如果构建还在进行中

```bash
# 等待 5 分钟后再次检查
sleep 300

# 再次检查容器状态
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker ps --filter 'name=meeting-' --format 'table {{.Names}}\t{{.Status}}'"
```

### 步骤 3: 如果构建失败或卡住

```bash
# 查看 docker-compose 日志
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && docker-compose -f docker-compose.remote.yml logs --tail=100"

# 重新启动部署
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && docker-compose -f docker-compose.remote.yml down && docker-compose -f docker-compose.remote.yml up -d"
```

## 服务启动后的操作

### 步骤 4: 下载 AI 模型

AI 模型文件未包含在代码同步中，需要在远程服务器上下载：

```bash
# 方法 1: 使用 download_models.py 脚本
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && python3 download_models.py"

# 如果 python3 或 pip3 不可用，使用容器内的 Python
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker exec meeting-ai-inference-worker python3 /app/download_models.py"
```

**注意**: 模型下载可能需要 10-30 分钟，取决于网络速度。

**需要下载的模型**:
- `openai/whisper-tiny` (语音识别, ~39MB)
- `j-hartmann/emotion-english-distilroberta-base` (情绪检测, ~82MB)
- `sshleifer/distilbart-cnn-6-6` (文本摘要, ~306MB)
- 其他辅助模型

### 步骤 5: 验证服务可访问性

```bash
# 从本地测试远程服务
curl http://js1.blockelite.cn:22176/health
curl http://js1.blockelite.cn:22177/
curl http://js1.blockelite.cn:22178/
curl http://js1.blockelite.cn:22180/
```

**预期结果**:
- Nginx (22176): 返回 HTTP 200
- Jaeger (22177): 返回 HTTP 200
- Prometheus (22178): 返回 HTTP 200
- Grafana (22180): 返回 HTTP 200 或 302

### 步骤 6: 验证 AI 服务

```bash
cd /root/meeting-system-server/meeting-system
./verify-ai-service-remote.sh
```

**关键检查点**:
- ✅ edge-model-infra 容器运行中
- ✅ ai-inference-worker 容器运行中
- ✅ ai-service 容器运行中
- ✅ IPC socket 文件存在: `/tmp/llm/5010.sock`
- ✅ 模型文件已下载到 `/models/`

### 步骤 7: 执行集成测试

```bash
cd /root/meeting-system-server/meeting-system
./run-remote-integration-test.sh
```

**测试覆盖范围**:
- 用户注册和登录
- 会议创建和加入
- AI 服务（情绪识别、语音识别）
- 服务间通信

### 步骤 8: 查看测试结果

```bash
# 查看测试日志
cat backend/tests/logs/remote_integration_test.log

# 如果测试失败，查看服务日志
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker logs meeting-ai-service --tail 100"
```

## 故障排查

### 问题 1: 容器启动失败

```bash
# 查看失败的容器
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker ps -a --filter 'status=exited' --filter 'name=meeting-'"

# 查看特定容器日志
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker logs meeting-[service-name]"

# 重启特定服务
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && docker-compose -f docker-compose.remote.yml restart [service-name]"
```

### 问题 2: AI 服务返回 "unit call false"

这通常表示 edge-model-infra 和 ai-inference-worker 之间的连接问题。

```bash
# 检查 edge-model-infra 日志
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker logs meeting-edge-model-infra --tail 50"

# 检查 IPC socket
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "ls -la /tmp/llm/"

# 检查 ai-inference-worker 进程
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker exec meeting-ai-inference-worker ps aux | grep python"

# 重启 AI 相关服务
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && docker-compose -f docker-compose.remote.yml restart edge-model-infra ai-inference-worker ai-service"
```

### 问题 3: 模型未加载

```bash
# 检查模型目录
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "ls -lh /models/"

# 手动下载模型
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker exec meeting-ai-inference-worker python3 -c 'from huggingface_hub import snapshot_download; snapshot_download(\"openai/whisper-tiny\", local_dir=\"/models/speech_recognition\")'"
```

### 问题 4: 端口不可访问

```bash
# 检查远程服务器端口监听
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "netstat -tlnp | grep -E ':(8800|8801|8802|8803|8804|8805)'"

# 检查防火墙
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "iptables -L -n | grep -E '8800|8801|8802|8803|8804|8805'"
```

## 完整的重新部署流程

如果需要完全重新部署：

```bash
# 1. 停止所有服务
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && docker-compose -f docker-compose.remote.yml down -v"

# 2. 清理旧镜像（可选）
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker system prune -af"

# 3. 重新传输代码
cd /root/meeting-system-server/meeting-system
tar czf /tmp/meeting-system.tar.gz --exclude='node_modules' --exclude='.git' --exclude='venv' --exclude='__pycache__' --exclude='*.pyc' --exclude='data' --exclude='logs' --exclude='/models' --exclude='*.bin' --exclude='*.safetensors' .
sshpass -p "beip3ius" scp -P 22124 /tmp/meeting-system.tar.gz root@js1.blockelite.cn:/tmp/
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && tar xzf /tmp/meeting-system.tar.gz"

# 4. 启动服务
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && docker-compose -f docker-compose.remote.yml up -d"

# 5. 等待服务启动
sleep 120

# 6. 下载模型
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker exec meeting-ai-inference-worker python3 /app/download_models.py"

# 7. 验证服务
./verify-ai-service-remote.sh

# 8. 运行测试
./run-remote-integration-test.sh
```

## 监控和日志

### 实时查看日志

```bash
# 查看所有服务日志
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "cd /root/meeting-system-server/meeting-system && docker-compose -f docker-compose.remote.yml logs -f"

# 查看特定服务日志
sshpass -p "beip3ius" ssh -p 22124 root@js1.blockelite.cn \
  "docker logs -f meeting-[service-name]"
```

### 访问监控界面

- **Jaeger**: http://js1.blockelite.cn:22177
  - 查看分布式追踪
  - 分析服务调用链

- **Prometheus**: http://js1.blockelite.cn:22178
  - 查看系统指标
  - 监控资源使用

- **Grafana**: http://js1.blockelite.cn:22180
  - 用户名: admin
  - 密码: admin123
  - 可视化监控仪表板

## 成功标准

部署成功的标志：

- ✅ 所有容器状态为 "Up"
- ✅ Nginx 网关可从外网访问 (22176)
- ✅ 监控服务可访问 (Jaeger, Prometheus, Grafana)
- ✅ AI 模型已下载到 `/models/`
- ✅ AI 服务 API 返回正常（无 "unit call false" 错误）
- ✅ 集成测试 100% 通过

## 联系和支持

如遇到问题，请检查：
1. `DEPLOYMENT_STATUS.md` - 部署状态文档
2. `REMOTE_DEPLOYMENT_GUIDE.md` - 完整部署指南
3. 服务日志和错误信息

---

**文档版本**: 1.0
**最后更新**: 2025-10-06

