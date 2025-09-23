package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"videocall-backend/config"
	"videocall-backend/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数 - 优化并发性能
	sqlDB.SetMaxIdleConns(20)        // 增加空闲连接数
	sqlDB.SetMaxOpenConns(200)       // 增加最大连接数
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 设置空闲连接超时

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	DB = db
	log.Println("Database connected successfully")
	return db, nil
}

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     50,              // 增加连接池大小
		MinIdleConns: 10,              // 最小空闲连接数
		MaxRetries:   3,               // 最大重试次数
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 连接池超时
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	RedisClient = client
	log.Println("Redis connected successfully")
	return client, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Call{},
		&models.SecurityDetection{},
		&models.UserSession{},
		&models.SystemConfig{},
		&models.ModelVersion{},
	)
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
} 