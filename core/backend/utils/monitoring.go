package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SystemMetrics 系统指标
type SystemMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage"`
	Goroutines    int       `json:"goroutines"`
	ActiveConns   int       `json:"active_connections"`
	DBConnections int       `json:"db_connections"`
	RedisConnections int    `json:"redis_connections"`
}

// StartMetricsServer 启动指标服务器
func StartMetricsServer(addr string) {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := getSystemMetrics()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})
	
	log.Printf("Metrics server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Metrics server error: %v", err)
	}
}

// StartHealthCheck 启动健康检查
func StartHealthCheck(ctx context.Context, db *gorm.DB, redisClient *redis.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkHealth(db, redisClient)
		}
	}
}

// MonitorConnectionPools 监控连接池
func MonitorConnectionPools(ctx context.Context, db *gorm.DB, redisClient *redis.Client) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			monitorPools(db, redisClient)
		}
	}
}

// getSystemMetrics 获取系统指标
func getSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// 计算内存使用率
	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100
	
	// 获取CPU使用率（简化版本）
	cpuUsage := getCPUUsage()
	
	return SystemMetrics{
		Timestamp:   time.Now(),
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		Goroutines:  runtime.NumGoroutine(),
	}
}

// getCPUUsage 获取CPU使用率
func getCPUUsage() float64 {
	// 这里可以实现更复杂的CPU使用率计算
	// 暂时返回一个模拟值
	return 15.5 // 模拟15.5%的CPU使用率
}

// checkHealth 检查系统健康状态
func checkHealth(db *gorm.DB, redisClient *redis.Client) {
	// 检查数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Health check: Database connection error: %v", err)
		return
	}
	
	if err := sqlDB.Ping(); err != nil {
		log.Printf("Health check: Database ping failed: %v", err)
	} else {
		log.Printf("Health check: Database OK")
	}
	
	// 检查Redis连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Health check: Redis ping failed: %v", err)
	} else {
		log.Printf("Health check: Redis OK")
	}
}

// monitorPools 监控连接池
func monitorPools(db *gorm.DB, redisClient *redis.Client) {
	// 监控数据库连接池
	sqlDB, err := db.DB()
	if err == nil {
		stats := sqlDB.Stats()
		log.Printf("DB Pool Stats: Open=%d, InUse=%d, Idle=%d, WaitCount=%d",
			stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
	}
	
	// 监控Redis连接池
	poolStats := redisClient.PoolStats()
	log.Printf("Redis Pool Stats: TotalConns=%d, IdleConns=%d, StaleConns=%d",
		poolStats.TotalConns, poolStats.IdleConns, poolStats.StaleConns)
}

// GetConnectionInfo 获取连接信息
func GetConnectionInfo(db *gorm.DB, redisClient *redis.Client) map[string]interface{} {
	info := make(map[string]interface{})
	
	// 数据库连接信息
	if sqlDB, err := db.DB(); err == nil {
		stats := sqlDB.Stats()
		info["database"] = map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":          stats.InUse,
			"idle":            stats.Idle,
			"wait_count":      stats.WaitCount,
			"wait_duration":   stats.WaitDuration.String(),
		}
	}
	
	// Redis连接信息
	poolStats := redisClient.PoolStats()
	info["redis"] = map[string]interface{}{
		"total_connections": poolStats.TotalConns,
		"idle_connections":  poolStats.IdleConns,
		"stale_connections": poolStats.StaleConns,
	}
	
	// 系统信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info["system"] = map[string]interface{}{
		"goroutines":    runtime.NumGoroutine(),
		"memory_alloc":  m.Alloc,
		"memory_total":  m.TotalAlloc,
		"memory_sys":    m.Sys,
		"num_gc":        m.NumGC,
	}
	
	return info
} 