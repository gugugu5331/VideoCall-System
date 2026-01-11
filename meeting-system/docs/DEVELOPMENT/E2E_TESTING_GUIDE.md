# 端到端测试指南（队列/信令/AI）

使用 `meeting-system/tests` 下脚本验证注册登录、会议创建/加入、信令连通性以及可选 AI 请求，确保 Kafka 队列与核心服务协同。

## 前置条件

- `docker compose up -d` 已启动基础设施与微服务（PostgreSQL、Redis、Kafka、etcd、Nginx、user/meeting/signaling/media，若测 AI 则含 Triton/ai-inference-service）
- 网关可访问 `http://localhost:8800`（或自定义域名/端口）
- 如使用远程环境，设置 `BASE_URL` / `REMOTE_BASE_URL` 指向实际网关；AI 节点地址可通过环境变量覆盖

## 可用脚本

- `e2e_queue_integration_test.sh`：Bash 版，基于 `curl` 跑完整流程
- `e2e_queue_integration_test.py`：Python 版，日志更详细
- `comprehensive_e2e_test.py`：扩展检查
- `check_service_logs.sh`：汇总容器日志，便于快速排错

> Python 版若需 Kafka 额外检查，可 `pip install requests kafka-python`；未安装库时核心 HTTP 流程仍可运行。

## 推荐执行

```bash
cd meeting-system/tests
./e2e_queue_integration_test.sh          # 或运行 python 版本
./check_service_logs.sh                  # 可选日志检查
```

脚本流程：
1. 注册并登录测试用户，获取 JWT  
2. 创建并加入会议  
3. 建立 WS 信令 `/ws/signaling?...token=<jwt>`  
4. （可选）调用 `/api/v1/ai/*` 做 ASR/情绪/合成检测  

## 手动检查

- 网关健康：`curl http://localhost:8800/health`
- 服务健康：各服务 `/health`
- AI：`curl http://localhost:8085/api/v1/ai/health`（启用后）
- 信令：`wscat -c "ws://localhost:8800/ws/signaling?..."` 或浏览器开发者工具

## 常见问题

1. **401/403**：确认 `JWT_SECRET` 已设置且 Token 最新；需要 CSRF 时调用 `/api/v1/csrf-token`
2. **WS 连接失败**：检查 `signaling-service` 与 Redis 连通性，确认网关 WS 代理配置
3. **AI 失败**：确认 Triton 模型就绪，`ai-inference-service` 配置与上游匹配
4. **录制/上传异常**：核对 MinIO 凭据和桶名称与 `media-service` 配置
