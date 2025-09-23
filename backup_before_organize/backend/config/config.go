package config

import (
	"os"
	"strconv"
)

// Config 应用配置结构
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	AI       AIConfig
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	ExpireTime int // 过期时间（小时）
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string
	Mode         string
	AllowOrigins []string
}

// AIConfig AI服务配置
type AIConfig struct {
	ServiceURL string
	Timeout    int
}

// Load 加载配置
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "videocall"),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "videocall123"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-here"),
			ExpireTime: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		},
		Server: ServerConfig{
			Port:         getEnv("PORT", "8000"),
			Mode:         getEnv("GIN_MODE", "debug"),
			AllowOrigins: getEnvAsSlice("ALLOW_ORIGINS", []string{"*"}),
		},
		AI: AIConfig{
			ServiceURL: getEnv("AI_SERVICE_URL", "http://localhost:5000"),
			Timeout:    getEnvAsInt("AI_SERVICE_TIMEOUT", 30),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsSlice 获取环境变量并转换为字符串切片
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// 简单的逗号分隔实现，可以根据需要扩展
		if value == "*" {
			return []string{"*"}
		}
		// 这里可以实现更复杂的解析逻辑
		return defaultValue
	}
	return defaultValue
} 