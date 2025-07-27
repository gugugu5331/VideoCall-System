package middleware

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 并发限制器
type ConcurrencyLimiter struct {
	current int64
	max     int64
}

// NewConcurrencyLimiter 创建新的并发限制器
func NewConcurrencyLimiter(max int64) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		current: 0,
		max:     max,
	}
}

// ConcurrencyLimit 并发限制中间件
func ConcurrencyLimit(limiter *ConcurrencyLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试增加当前并发数
		current := atomic.AddInt64(&limiter.current, 1)
		if current > limiter.max {
			// 超过限制，减少计数并返回错误
			atomic.AddInt64(&limiter.current, -1)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service temporarily unavailable",
				"message": "Too many concurrent requests",
			})
			c.Abort()
			return
		}
		
		// 确保在请求结束时减少计数
		defer atomic.AddInt64(&limiter.current, -1)
		
		// 设置请求超时
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}

// RateLimit 限流中间件
func RateLimit(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP
		
		// 使用Redis实现滑动窗口限流
		ctx := c.Request.Context()
		current, err := redisClient.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		
		// 设置过期时间（1分钟窗口）
		if current == 1 {
			redisClient.Expire(ctx, key, time.Minute)
		}
		
		// 限制每分钟100个请求
		if current > 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests per minute",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Metrics 监控中间件
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 处理请求
		c.Next()
		
		// 记录指标
		duration := time.Since(start)
		status := c.Writer.Status()
		
		// 这里可以发送指标到监控系统
		// 例如：Prometheus, StatsD等
		recordMetrics(c.Request.Method, c.Request.URL.Path, status, duration)
	}
}

// recordMetrics 记录监控指标
func recordMetrics(method, path string, status int, duration time.Duration) {
	// 这里可以实现具体的指标记录逻辑
	// 例如发送到Prometheus或StatsD
	// 暂时使用日志记录
	log.Printf("Metrics: %s %s %d %v", method, path, status, duration)
} 