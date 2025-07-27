# VideoCall System - 并发性能指南

## 🚀 概述

本文档介绍VideoCall系统的多线程和高并发实现，包括性能优化策略、监控指标和最佳实践。

## 📊 系统架构

### 后端服务 (Golang)
- **并发模型**: Goroutines + 信号量控制
- **最大并发请求**: 1000
- **连接池优化**: 
  - PostgreSQL: 20-200连接
  - Redis: 50连接池
- **限流机制**: Redis滑动窗口
- **监控**: 实时指标收集

### AI服务 (Python)
- **并发模型**: asyncio + 线程池 + 进程池
- **线程池大小**: CPU核心数 × 2
- **进程池大小**: CPU核心数
- **异步处理**: 批量检测支持
- **资源监控**: CPU/内存/磁盘监控

## 🔧 性能优化特性

### 1. 后端服务优化

#### 并发控制
```go
// 全局并发限制
maxConcurrentRequests := int64(1000)
requestSemaphore = semaphore.NewWeighted(maxConcurrentRequests)

// 中间件实现
func ConcurrencyLimit(sem *semaphore.Weighted) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := sem.Acquire(c.Request.Context(), 1); err != nil {
            c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Too many concurrent requests"})
            c.Abort()
            return
        }
        defer sem.Release(1)
        c.Next()
    }
}
```

#### 限流机制
```go
// Redis滑动窗口限流
func RateLimit(redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        clientIP := c.ClientIP()
        key := "rate_limit:" + clientIP
        
        current, err := redisClient.Incr(ctx, key).Result()
        if current == 1 {
            redisClient.Expire(ctx, key, time.Minute)
        }
        
        if current > 100 { // 每分钟100个请求
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

#### 连接池优化
```go
// PostgreSQL连接池
sqlDB.SetMaxIdleConns(20)
sqlDB.SetMaxOpenConns(200)
sqlDB.SetConnMaxLifetime(time.Hour)
sqlDB.SetConnMaxIdleTime(30 * time.Minute)

// Redis连接池
client := redis.NewClient(&redis.Options{
    PoolSize:     50,
    MinIdleConns: 10,
    MaxRetries:   3,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolTimeout:  4 * time.Second,
})
```

### 2. AI服务优化

#### 异步处理架构
```python
# 线程池和进程池
thread_pool = ThreadPoolExecutor(max_workers=multiprocessing.cpu_count() * 2)
process_pool = ProcessPoolExecutor(max_workers=multiprocessing.cpu_count())

# 异步检测
async def detect_spoofing(request: DetectionRequest):
    loop = asyncio.get_event_loop()
    result = await loop.run_in_executor(
        thread_pool,
        detection_service.detect_sync,
        request.detection_id,
        request.detection_type,
        request.audio_data,
        request.video_data,
        request.metadata
    )
    return result
```

#### 批量处理
```python
@app.post("/detect/batch")
async def detect_spoofing_batch(requests: list[DetectionRequest]):
    tasks = []
    for request in requests:
        task = asyncio.create_task(detect_single_request(request))
        tasks.append(task)
    
    results = await asyncio.gather(*tasks, return_exceptions=True)
    return process_batch_results(results)
```

#### 资源监控
```python
async def monitor_system_resources():
    while True:
        cpu_percent = psutil.cpu_percent(interval=5)
        memory = psutil.virtual_memory()
        
        if cpu_percent > 80:
            logger.warning(f"High CPU usage: {cpu_percent}%")
        
        if memory.percent > 80:
            logger.warning(f"High memory usage: {memory.percent}%")
        
        await asyncio.sleep(30)
```

## 📈 性能指标

### 监控端点

#### 后端服务
- `GET /health` - 健康检查
- `GET /metrics` - 系统指标 (端口8080)

#### AI服务
- `GET /health` - 健康检查 (包含资源使用情况)
- `GET /metrics` - 详细系统指标

### 关键指标

#### 响应时间
- 平均响应时间
- 中位数响应时间
- 95th百分位响应时间
- 99th百分位响应时间

#### 吞吐量
- 请求/秒 (RPS)
- 并发连接数
- 成功/失败率

#### 资源使用
- CPU使用率
- 内存使用率
- 连接池状态
- Goroutine数量

## 🧪 性能测试

### 并发测试脚本
```bash
# 健康检查并发测试
python scripts/testing/test_concurrency.py --requests 100 --type health

# 检测服务并发测试
python scripts/testing/test_concurrency.py --requests 50 --type detection

# 批量检测测试
python scripts/testing/test_concurrency.py --requests 200 --type batch
```

### 测试参数
- `--requests`: 并发请求数量
- `--type`: 测试类型 (health/detection/batch)
- `--backend-url`: 后端服务URL
- `--ai-url`: AI服务URL

### 测试结果分析
- 成功率统计
- 响应时间分布
- 吞吐量计算
- 错误分析

## 🔍 监控和告警

### 系统监控
```python
# 系统指标结构
{
    "timestamp": 1640995200,
    "cpu": {
        "usage_percent": 15.5,
        "cores": 8
    },
    "memory": {
        "total": 8589934592,
        "used": 4294967296,
        "percent": 50.0
    },
    "disk": {
        "total": 107374182400,
        "used": 53687091200,
        "percent": 50.0
    }
}
```

### 连接池监控
```go
// 数据库连接池状态
{
    "open_connections": 15,
    "in_use": 8,
    "idle": 7,
    "wait_count": 0,
    "wait_duration": "0s"
}

// Redis连接池状态
{
    "total_connections": 25,
    "idle_connections": 20,
    "stale_connections": 0
}
```

## 🛠️ 配置调优

### 环境变量配置
```bash
# 后端服务配置
MAX_CONCURRENT_REQUESTS=1000
DB_MAX_OPEN_CONNS=200
DB_MAX_IDLE_CONNS=20
REDIS_POOL_SIZE=50
RATE_LIMIT_PER_MINUTE=100

# AI服务配置
THREAD_POOL_SIZE=16  # CPU核心数 × 2
PROCESS_POOL_SIZE=8   # CPU核心数
MAX_BATCH_SIZE=10
```

### 性能调优建议

#### 高并发场景
1. 增加连接池大小
2. 调整并发限制
3. 优化数据库查询
4. 使用缓存策略

#### 低延迟场景
1. 减少中间件数量
2. 优化序列化
3. 使用连接复用
4. 预加载数据

#### 高可用场景
1. 实现熔断机制
2. 添加重试逻辑
3. 监控资源使用
4. 自动扩缩容

## 📋 最佳实践

### 1. 并发控制
- 使用信号量限制并发数
- 实现优雅降级
- 监控资源使用情况

### 2. 错误处理
- 超时控制
- 重试机制
- 错误分类处理

### 3. 监控告警
- 实时指标收集
- 阈值告警
- 性能趋势分析

### 4. 资源管理
- 连接池管理
- 内存使用优化
- CPU使用监控

## 🚨 故障排查

### 常见问题

#### 高延迟
1. 检查数据库连接池
2. 监控Redis性能
3. 分析慢查询
4. 检查网络延迟

#### 高错误率
1. 检查服务健康状态
2. 监控资源使用
3. 分析错误日志
4. 检查依赖服务

#### 内存泄漏
1. 监控内存使用趋势
2. 检查goroutine泄漏
3. 分析对象引用
4. 优化数据结构

### 调试工具
- `pprof` - Go性能分析
- `netstat` - 网络连接分析
- `top` - 系统资源监控
- `redis-cli` - Redis性能分析

## 📚 参考资料

- [Go Concurrency Patterns](https://golang.org/doc/effective_go.html#concurrency)
- [FastAPI Performance](https://fastapi.tiangolo.com/tutorial/performance/)
- [Redis Performance](https://redis.io/topics/optimization)
- [PostgreSQL Performance](https://www.postgresql.org/docs/current/performance.html) 