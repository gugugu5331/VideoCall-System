package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"meeting-system/shared/logger"
	"meeting-system/shared/response"
)

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	// RequestsPerSecond 每秒允许的请求数
	RequestsPerSecond int
	// BurstSize 突发请求数（令牌桶大小）
	BurstSize int
	// KeyFunc 获取限流键的函数（默认使用 IP）
	KeyFunc func(*gin.Context) string
}

// tokenBucket 令牌桶
type tokenBucket struct {
	tokens         int       // 当前令牌数
	maxTokens      int       // 最大令牌数
	refillRate     int       // 每秒补充的令牌数
	lastRefillTime time.Time // 上次补充时间
	mu             sync.Mutex
}

// newTokenBucket 创建令牌桶
func newTokenBucket(maxTokens, refillRate int) *tokenBucket {
	return &tokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// tryConsume 尝试消费一个令牌
func (tb *tokenBucket) tryConsume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tokensToAdd := int(elapsed * float64(tb.refillRate))

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		tb.lastRefillTime = now
	}

	// 尝试消费令牌
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// rateLimiter 限流器
type rateLimiter struct {
	buckets map[string]*tokenBucket
	config  RateLimiterConfig
	mu      sync.RWMutex
}

// newRateLimiter 创建限流器
func newRateLimiter(config RateLimiterConfig) *rateLimiter {
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP()
		}
	}

	limiter := &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		config:  config,
	}

	// 启动清理过期桶的 goroutine
	go limiter.cleanupExpiredBuckets()

	return limiter
}

// getBucket 获取或创建令牌桶
func (rl *rateLimiter) getBucket(key string) *tokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return bucket
	}

	// 创建新桶
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查
	if bucket, exists := rl.buckets[key]; exists {
		return bucket
	}

	bucket = newTokenBucket(rl.config.BurstSize, rl.config.RequestsPerSecond)
	rl.buckets[key] = bucket
	return bucket
}

// cleanupExpiredBuckets 清理过期的令牌桶
func (rl *rateLimiter) cleanupExpiredBuckets() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			bucket.mu.Lock()
			// 如果桶超过10分钟没有使用，删除它
			if now.Sub(bucket.lastRefillTime) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimit 限流中间件
func RateLimit(config RateLimiterConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)
		bucket := limiter.getBucket(key)

		if !bucket.tryConsume() {
			logger.Warn("Rate limit exceeded",
				logger.String("key", key),
				logger.String("path", c.Request.URL.Path),
				logger.String("method", c.Request.Method))

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "1")

			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		// 设置限流信息头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerSecond))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", bucket.tokens))

		c.Next()
	}
}

// IPRateLimit 基于 IP 的限流中间件
func IPRateLimit(requestsPerSecond, burstSize int) gin.HandlerFunc {
	return RateLimit(RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	})
}

// UserRateLimit 基于用户 ID 的限流中间件
func UserRateLimit(requestsPerSecond, burstSize int) gin.HandlerFunc {
	return RateLimit(RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		KeyFunc: func(c *gin.Context) string {
			// 尝试从上下文获取用户 ID
			if userID, exists := c.Get("user_id"); exists {
				return fmt.Sprintf("user:%v", userID)
			}
			// 如果没有用户 ID，使用 IP
			return fmt.Sprintf("ip:%s", c.ClientIP())
		},
	})
}

// EndpointRateLimit 基于端点的限流中间件
func EndpointRateLimit(requestsPerSecond, burstSize int) gin.HandlerFunc {
	return RateLimit(RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		KeyFunc: func(c *gin.Context) string {
			// 组合 IP 和端点路径
			return fmt.Sprintf("%s:%s", c.ClientIP(), c.Request.URL.Path)
		},
	})
}

// StrictRateLimit 严格限流中间件（用于敏感操作）
// 例如：登录、注册、密码重置等
func StrictRateLimit(requestsPerMinute int) gin.HandlerFunc {
	return RateLimit(RateLimiterConfig{
		RequestsPerSecond: requestsPerMinute / 60,
		BurstSize:         requestsPerMinute / 60,
		KeyFunc: func(c *gin.Context) string {
			// 组合 IP 和端点，更严格的限制
			return fmt.Sprintf("strict:%s:%s", c.ClientIP(), c.Request.URL.Path)
		},
	})
}

// GlobalRateLimit 全局限流中间件
// 限制整个服务的请求速率
func GlobalRateLimit(requestsPerSecond, burstSize int) gin.HandlerFunc {
	return RateLimit(RateLimiterConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		KeyFunc: func(c *gin.Context) string {
			return "global"
		},
	})
}

