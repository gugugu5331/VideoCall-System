package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MongoDB  MongoConfig    `mapstructure:"mongodb"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	ZMQ      ZMQConfig      `mapstructure:"zmq"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	WebRTC   WebRTCConfig   `mapstructure:"webrtc"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name         string        `mapstructure:"name"`
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig PostgreSQL数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  time.Duration `mapstructure:"max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// MongoConfig MongoDB配置
type MongoConfig struct {
	URI        string `mapstructure:"uri"`
	Database   string `mapstructure:"database"`
	Collection string `mapstructure:"collection"`
}

// MinIOConfig MinIO对象存储配置
type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketName      string `mapstructure:"bucket_name"`
}

// ZMQConfig ZMQ通信配置
type ZMQConfig struct {
	ServiceName     string            `mapstructure:"service_name"`
	UnitManagerURL  string            `mapstructure:"unit_manager_url"`
	PublisherURL    string            `mapstructure:"publisher_url"`
	SubscriberURL   string            `mapstructure:"subscriber_url"`
	ReplierURL      string            `mapstructure:"replier_url"`
	Endpoints       map[string]string `mapstructure:"endpoints"`
	Timeout         time.Duration     `mapstructure:"timeout"`
}

// JWTConfig JWT认证配置
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Issuer     string        `mapstructure:"issuer"`
	Expiration time.Duration `mapstructure:"expiration"`
}

// WebRTCConfig WebRTC配置
type WebRTCConfig struct {
	STUNServers []string `mapstructure:"stun_servers"`
	TURNServers []string `mapstructure:"turn_servers"`
	ICEServers  []ICEServer `mapstructure:"ice_servers"`
}

// ICEServer ICE服务器配置
type ICEServer struct {
	URLs       []string `mapstructure:"urls"`
	Username   string   `mapstructure:"username,omitempty"`
	Credential string   `mapstructure:"credential,omitempty"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, text
	Output     string `mapstructure:"output"` // stdout, file
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`    // MB
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`     // days
	Compress   bool   `mapstructure:"compress"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// 数据库默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "meeting_system")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", "1h")

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.idle_timeout", "5m")

	// MongoDB默认配置
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "meeting_system")
	viper.SetDefault("mongodb.collection", "meeting_records")

	// MinIO默认配置
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.access_key_id", "minioadmin")
	viper.SetDefault("minio.secret_access_key", "minioadmin")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_name", "meeting-recordings")

	// ZMQ默认配置
	viper.SetDefault("zmq.service_name", "meeting-service")
	viper.SetDefault("zmq.unit_manager_url", "tcp://localhost:5555")
	viper.SetDefault("zmq.publisher_url", "tcp://*:5556")
	viper.SetDefault("zmq.subscriber_url", "tcp://localhost:5557")
	viper.SetDefault("zmq.replier_url", "tcp://*:5558")
	viper.SetDefault("zmq.timeout", "30s")

	// JWT默认配置
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.issuer", "meeting-system")
	viper.SetDefault("jwt.expiration", "24h")

	// WebRTC默认配置
	viper.SetDefault("webrtc.stun_servers", []string{
		"stun:stun.l.google.com:19302",
		"stun:stun1.l.google.com:19302",
	})

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)
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

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if c.ZMQ.ServiceName == "" {
		return fmt.Errorf("ZMQ service name is required")
	}

	return nil
}
