# 端到端测试指南（队列/信令/AI）

## 目标

用仓库自带脚本验证注册/登录、会议创建/加入、信令连接以及可选的 AI 请求，确保 Redis 队列和核心服务正常协同。

## 前置条件

- `docker compose up -d` 已启动基础设施和微服务（PostgreSQL、Redis、etcd、Nginx、user/meeting/signaling/media/ai、triton 等）。
- 网关可访问 `http://localhost:8800`，如有自定义域名/端口，请在脚本中调整 `BASE_URL` 相关变量。

## 可用脚本（meeting-system/tests）

- `e2e_queue_integration_test.sh`：bash 版本，使用 `curl` 走完整流程。
- `e2e_queue_integration_test.py`：python 版本，输出更详细的日志/报告。
- `check_service_logs.sh`：汇总容器日志，检查初始化与错误。

> 运行前可 `chmod +x *.sh`。Python 版需 `pip install requests redis`。

## 推荐执行步骤

```bash
cd meeting-system/tests
./e2e_queue_integration_test.sh        # 或运行 python 版本
./check_service_logs.sh                # 如需额外核对日志
```

脚本通常会：
1) 注册并登录测试用户，获取 JWT  
2) 创建会议并加入  
3) 建立 WS 信令连接 `/ws/signaling?...token=<jwt>`  
4) （可选）调用 `/api/v1/ai/*` 进行 ASR/情绪/合成检测  

所有 HTTP 请求需返回 2xx；如启用 AI，需确保 `ai-inference-service` 与 `triton` 健康。

## 手动检查

- 网关健康：`curl http://localhost:8800/health`
- 用户/会议/信令/媒体健康：访问各服务 `/health`
- AI：`curl http://localhost:8085/api/v1/ai/health`
- 信令：使用 `wscat` 或浏览器连接 `ws://localhost:8800/ws/signaling?...`

## 常见问题

1. **401/403**：确认设置了 `JWT_SECRET` 且前端携带了最新 Token；需要 CSRF 时调用 `/api/v1/csrf-token`。
2. **WS 连接失败**：检查 `signaling-service` 容器、Redis 连通性、Nginx WS 配置。
3. **AI 请求报错**：确认 `triton` 启动且模型仓库已挂载，`backend/ai-inference-service/config/ai-inference-service.yaml` 与 Nginx upstream 一致。
4. **录制/上传异常**：核对 MinIO 凭据和桶名称与 `media-service` 配置是否一致。

## 参考

- `meeting-system/tests/*`
- `backend/config/*.yaml`
- `docker-compose.yml`
