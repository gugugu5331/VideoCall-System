package database

import (
	"context"
	"fmt"
	"log"

	"video-conference-system/shared/config"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient 创建Redis连接
func NewRedisClient(cfg *config.RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")
	return &RedisClient{Client: rdb}, nil
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
