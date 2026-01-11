# 微服务测试指南

覆盖集成测试脚本、执行顺序与排查方法。测试脚本位于 `meeting-system/backend/tests`。

## 前置条件

- 依赖容器已启动：`docker compose up -d`（至少 PostgreSQL/Redis/Kafka/etcd/MinIO/Nginx）
- `JWT_SECRET` 等环境变量已设置
- 需要 AI 测试时，确保 `ai-inference-service` 与 Triton 可访问

## 主要脚本

- `run_all_tests.sh`：覆盖网关与主要业务流程的集成测试
- `quick_integration_test.sh`：快速连通性与基础 API 验证
- `test_nginx_gateway.sh`：网关路由与 upstream 校验
- 其他：`test_services_direct.sh`、`verify_ai_service.sh`、`complete_integration_test.py`（更细粒度或 AI 专用）

执行示例：

```bash
cd meeting-system/backend/tests
./run_all_tests.sh
```

## 推荐流程

1. `docker compose up -d` 启动依赖与服务
2. `quick_integration_test.sh`（快速检查）
3. `run_all_tests.sh` 或按需运行 AI/网关专项脚本
4. 查看输出与容器日志，确认健康检查通过

> 部分脚本允许覆盖基础地址（`BASE_URL`/`GATEWAY_URL`）、账号、AI 上游等参数，可在执行前通过环境变量传入或临时 `.env` 文件加载。

## 故障排查

- **容器未就绪**：`docker compose ps` 查看状态；必要时 `docker compose logs -f <service>`
- **认证失败**：确认 `JWT_SECRET` 配置一致；重置测试数据或重新注册登录
- **AI 相关错误**：检查 `http://<triton>:8000/v2/health/ready` 与 `/api/v1/ai/health`
- **队列异常**：确认 Kafka 端口/健康，或在配置中将队列切换为 `memory` 进行定位

> 本目录不包含固定的测试报告文件；请根据当前运行结果记录。

## 性能巡检 / CI 快捷入口

- `scripts/perf_smoke.sh`：串行执行网关性能（`nginx/scripts/test-gateway.sh --performance`）、信令快压（`backend/signaling-service/run_stress_test.sh --quick`）、业务 HTTP 压测（`backend/stress-test`）。输出落在 `perf-results/<timestamp>/`，可用于 GitHub Actions/cron。
- 常用环境变量：`SIGNALING_URL`、`SIGNALING_SECRET`、`SIGNALING_MEETING_ID`、`USER_SERVICE_URL`、`MEETING_SERVICE_URL`、`STRESS_CONCURRENT_USERS`（逗号分隔），`STABILITY_USERS`、`STABILITY_DURATION`。

## 深挖瓶颈（pprof / Prom）

- pprof 建议在单服务中临时开启：
  ```go
  import _ "net/http/pprof"
  go func() { _ = http.ListenAndServe(":6060", nil) }()
  ```
  然后用 `go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30` 抓热点，或 `pprof -http=:8088` 可视化。
- Prometheus GoCollector：在服务 init 前注册 `prometheus.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))`，便于抓取 GC/协程数等 runtime 指标。

## 长时间稳压

- `backend/stress-test` 长压时间与用户数可通过环境变量调整，示例：`STABILITY_USERS=300 STABILITY_DURATION=1h go run .`。
- 建议结合 `GODEBUG=gctrace=1`、`docker stats`、`ss -s` 观察 GC 抖动、内存是否回收、TIME_WAIT 是否堆积；将关键指标（p95、错误率、CPU/内存）与 5/30/60 分钟时刻对比。
