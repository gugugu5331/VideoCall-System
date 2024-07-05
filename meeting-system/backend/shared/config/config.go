package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MongoDB  MongoConfig    `mapstructure:"mongodb"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	ZMQ      ZMQConfig      `mapstructure:"zmq"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig PostgreSQL数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// MongoConfig MongoDB配置
type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

// MinIOConfig MinIO对象存储配置
type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expire_time"` // 小时
}

// ZMQConfig ZeroMQ配置 (与Edge-LLM-Infra集成)
type ZMQConfig struct {
	UnitManagerHost string `mapstructure:"unit_manager_host"`
	UnitManagerPort int    `mapstructure:"unit_manager_port"`
	UnitName        string `mapstructure:"unit_name"`
	Timeout         int    `mapstructure:"timeout"` // 秒
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = &config
	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("server.write_timeout", 60)

	// 数据库默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "meeting_system")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// MongoDB默认配置
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "meeting_system")

	// MinIO默认配置
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.access_key_id", "minioadmin")
	viper.SetDefault("minio.secret_access_key", "minioadmin")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_name", "meeting-system")

	// JWT默认配置
	viper.SetDefault("jwt.secret", "meeting-system-secret-key")
	viper.SetDefault("jwt.expire_time", 24)

	// ZMQ默认配置
	viper.SetDefault("zmq.unit_manager_host", "localhost")
	viper.SetDefault("zmq.unit_manager_port", 5001)
	viper.SetDefault("zmq.unit_name", "meeting_ai_service")
	viper.SetDefault("zmq.timeout", 30)

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.filename", "logs/app.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.max_backups", 10)
	viper.SetDefault("log.compress", true)
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetZMQAddr 获取ZMQ地址
func (c *ZMQConfig) GetZMQAddr() string {
	return fmt.Sprintf("tcp://%s:%d", c.UnitManagerHost, c.UnitManagerPort)
}

// InitConfig 初始化配置
func InitConfig(configPath string) {
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	GlobalConfig = config
	log.Printf("Config loaded successfully from %s", configPath)
}
