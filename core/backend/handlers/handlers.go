package handlers

import (
	"videocall-backend/config"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
	Config      *config.Config
)

// InitHandlers 初始化所有处理器
func InitHandlers(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) {
	DB = db
	RedisClient = redisClient
	Config = cfg
} 