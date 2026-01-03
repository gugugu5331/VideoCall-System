# 🚀 部署指南

本目录包含系统部署、配置和运维相关的文档。

## 📖 文档列表

### 部署指南
- **[REMOTE_DEPLOYMENT_GUIDE.md](REMOTE_DEPLOYMENT_GUIDE.md)** - 远程服务器部署指南
- **[GPU_AI_NODES.md](GPU_AI_NODES.md)** - 多台 GPU 服务器 AI 推理节点部署
- **[AI_MODELS_DEPLOYMENT_GUIDE.md](AI_MODELS_DEPLOYMENT_GUIDE.md)** - AI 模型部署和配置

## 🎯 部署场景

### 本地开发环境
```bash
cd meeting-system
docker-compose up -d
```

### 远程生产环境
请参考 [REMOTE_DEPLOYMENT_GUIDE.md](REMOTE_DEPLOYMENT_GUIDE.md)

### AI 功能启用
请参考 [AI_MODELS_DEPLOYMENT_GUIDE.md](AI_MODELS_DEPLOYMENT_GUIDE.md)

## 📋 部署检查清单

- [ ] 系统要求满足（Docker 20.0+, Docker Compose 2.0+）
- [ ] 环境变量配置正确
- [ ] 数据库初始化完成
- [ ] 所有服务启动成功
- [ ] 健康检查通过
- [ ] API 网关可访问
- [ ] 监控系统正常运行

## 🔍 验证部署

### 检查服务状态
```bash
docker-compose ps
```

### 检查健康状态
```bash
curl http://localhost:8800/health
```

### 查看日志
```bash
docker-compose logs -f
```

## 🛠️ 常见问题

### Q: 如何重启服务？
```bash
docker-compose restart
```

### Q: 如何查看特定服务的日志？
```bash
docker-compose logs -f <service_name>
```

### Q: 如何更新配置？
1. 修改配置文件
2. 重启相应服务
3. 验证配置生效

## 📚 相关文档

- [开发指南](../DEVELOPMENT/README.md) - 开发和测试
- [API 文档](../API/README.md) - API 接口参考
- [文档中心](../README.md) - 所有文档

## 🔗 相关链接

- [项目主 README](../../README.md)
- [后端系统 README](../README.md)
