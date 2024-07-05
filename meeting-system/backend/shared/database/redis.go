package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

var RedisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis(config config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.GetRedisAddr(),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully")
	return nil
}

// GetRedis 获取Redis客户端
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// RedisCache Redis缓存操作封装
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache() *RedisCache {
	return &RedisCache{
		client: RedisClient,
	}
}

// Set 设置缓存
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取缓存
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del 删除缓存
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// HSet 设置哈希字段
func (r *RedisCache) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func (r *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// SAdd 添加集合成员
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SRem 移除集合成员
func (r *RedisCache) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember 检查是否为集合成员
func (r *RedisCache) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// ZAdd 添加有序集合成员
func (r *RedisCache) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRem 移除有序集合成员
func (r *RedisCache) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围成员
func (r *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore 按分数范围获取有序集合成员
func (r *RedisCache) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, opt).Result()
}

// Publish 发布消息
func (r *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (r *RedisCache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// LPush 从左侧推入列表
func (r *RedisCache) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 从右侧推入列表
func (r *RedisCache) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop 从左侧弹出列表元素
func (r *RedisCache) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出列表元素
func (r *RedisCache) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (r *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// LRange 获取列表范围元素
func (r *RedisCache) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}
