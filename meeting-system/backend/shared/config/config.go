package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const defaultJWTSecret = "meeting-system-secret-key"

// Config 全局配置结构
type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	GRPC           GRPCConfig           `mapstructure:"grpc"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Redis          RedisConfig          `mapstructure:"redis"`
	MongoDB        MongoConfig          `mapstructure:"mongodb"`
	MinIO          MinIOConfig          `mapstructure:"minio"`
	JWT            JWTConfig            `mapstructure:"jwt"`
	ZMQ            ZMQConfig            `mapstructure:"zmq"`
	AI             AIConfig             `mapstructure:"ai"`
	Log            LogConfig            `mapstructure:"log"`
	WebSocket      WebSocketConfig      `mapstructure:"websocket"`
	Signaling      SignalingConfig      `mapstructure:"signaling"`
	WebRTC         WebRTCConfig         `mapstructure:"webrtc"`
	Services       ServicesConfig       `mapstructure:"services"`
	Etcd           EtcdConfig           `mapstructure:"etcd"`
	MessageQueue   MessageQueueConfig   `mapstructure:"message_queue"`
	TaskScheduler  TaskSchedulerConfig  `mapstructure:"task_scheduler"`
	EventBus       EventBusConfig       `mapstructure:"event_bus"`
	TaskDispatcher TaskDispatcherConfig `mapstructure:"task_dispatcher"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"` // debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// GRPCConfig gRPC配置
type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"` // postgres, sqlite
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	DSN             string `mapstructure:"dsn"` // 用于SQLite
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host          string `mapstructure:"host"`
	Port          int    `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	DB            int    `mapstructure:"db"`
	PoolSize      int    `mapstructure:"pool_size"`
	SessionPrefix string `mapstructure:"session_prefix"`
	RoomPrefix    string `mapstructure:"room_prefix"`
	MessagePrefix string `mapstructure:"message_prefix"`
	SessionTTL    int    `mapstructure:"session_ttl"`
}

// MongoConfig MongoDB配置
type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
	Timeout  int    `mapstructure:"timeout"`
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

// ZMQConfig ZeroMQ配置（遗留配置，当前未使用）
type ZMQConfig struct {
	UnitManagerHost string `mapstructure:"unit_manager_host"`
	UnitManagerPort int    `mapstructure:"unit_manager_port"`
	UnitName        string `mapstructure:"unit_name"`
	Timeout         int    `mapstructure:"timeout"` // 秒
}

// AIConfig AI 推理配置（ai-inference-service 使用）
type AIConfig struct {
	Runtime   AIRuntimeConfig   `mapstructure:"runtime"`
	Models    AIModelsConfig     `mapstructure:"models"`
	Request   AIRequestConfig    `mapstructure:"request"`
	Streaming AIStreamingConfig  `mapstructure:"streaming"`
}

type AIModelsConfig struct {
	ASR       AIModelConfig `mapstructure:"asr"`
	Emotion   AIModelConfig `mapstructure:"emotion"`
	Synthesis AIModelConfig `mapstructure:"synthesis"`
}

// AIRuntimeConfig defines inference runtime settings.
type AIRuntimeConfig struct {
	Backend        string   `mapstructure:"backend"`
	Providers      []string `mapstructure:"providers"`
	DeviceID       int      `mapstructure:"device_id"`
	LibraryPath    string   `mapstructure:"library_path"`
	IntraOpThreads int      `mapstructure:"intra_op_threads"`
	InterOpThreads int      `mapstructure:"inter_op_threads"`
	EnableFP16     bool     `mapstructure:"enable_fp16"`
	EnableTensorRT bool     `mapstructure:"enable_tensorrt"`
	Triton         TritonConfig `mapstructure:"triton"`
}

// TritonConfig defines the remote Triton inference server settings.
type TritonConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	TimeoutMs int    `mapstructure:"timeout_ms"`
}

type AIModelConfig struct {
	ModelName          string   `mapstructure:"model_name"`
	ModelPath          string   `mapstructure:"model_path"`
	InputName          string   `mapstructure:"input_name"`
	OutputNames        []string `mapstructure:"output_names"`
	InputType          string   `mapstructure:"input_type"`
	DecoderPath        string   `mapstructure:"decoder_path"`
	DecoderInputNames  []string `mapstructure:"decoder_input_names"`
	DecoderOutputNames []string `mapstructure:"decoder_output_names"`
	TokenizerPath      string   `mapstructure:"tokenizer_path"`
	SpecialTokensPath  string   `mapstructure:"special_tokens_path"`
	ConfigPath         string   `mapstructure:"config_path"`
	LabelsPath         string   `mapstructure:"labels_path"`
	SampleRate         int      `mapstructure:"sample_rate"`
	Channels           int      `mapstructure:"channels"`
	Timeout            int      `mapstructure:"timeout"`        // 秒
	MaxConcurrent      int      `mapstructure:"max_concurrent"` // 并发上限（预留）
}

type AIRequestConfig struct {
	MaxRetries int `mapstructure:"max_retries"`
	RetryDelay int `mapstructure:"retry_delay"` // 毫秒
	Timeout    int `mapstructure:"timeout"`     // 秒
}

// AIStreamingConfig defines stream aggregation behavior for gRPC audio streams.
type AIStreamingConfig struct {
	FlushIntervalMs int `mapstructure:"flush_interval_ms"`
	MaxBufferMs     int `mapstructure:"max_buffer_ms"`
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

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Path            string `mapstructure:"path"`
	ReadBufferSize  int    `mapstructure:"read_buffer_size"`
	WriteBufferSize int    `mapstructure:"write_buffer_size"`
	CheckOrigin     bool   `mapstructure:"check_origin"`
	PingPeriod      int    `mapstructure:"ping_period"`
	PongWait        int    `mapstructure:"pong_wait"`
	WriteWait       int    `mapstructure:"write_wait"`
	MaxMessageSize  int    `mapstructure:"max_message_size"`
}

// SignalingConfig 信令服务配置
type SignalingConfig struct {
	Room       RoomConfig    `mapstructure:"room"`
	Session    SessionConfig `mapstructure:"session"`
	Message    MessageConfig `mapstructure:"message"`
	ICEServers []ICEServer   `mapstructure:"ice_servers"`
	Media      MediaConfig   `mapstructure:"media"`
}

// RoomConfig 房间配置
type RoomConfig struct {
	MaxParticipants int `mapstructure:"max_participants"`
	CleanupInterval int `mapstructure:"cleanup_interval"`
	InactiveTimeout int `mapstructure:"inactive_timeout"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	HeartbeatInterval    int `mapstructure:"heartbeat_interval"`
	ConnectionTimeout    int `mapstructure:"connection_timeout"`
	MaxReconnectAttempts int `mapstructure:"max_reconnect_attempts"`
}

// MessageConfig 消息配置
type MessageConfig struct {
	MaxQueueSize  int `mapstructure:"max_queue_size"`
	BatchSize     int `mapstructure:"batch_size"`
	RetryAttempts int `mapstructure:"retry_attempts"`
	RetryDelay    int `mapstructure:"retry_delay"`
}

// ICEServer ICE服务器配置
type ICEServer struct {
	URLs       string `mapstructure:"urls"`
	Username   string `mapstructure:"username,omitempty"`
	Credential string `mapstructure:"credential,omitempty"`
}

// WebRTCConfig WebRTC（SFU/ICE）配置
type WebRTCConfig struct {
	ICEServers []WebRTCICEServer `mapstructure:"ice_servers"`
}

// WebRTCICEServer WebRTC ICE server 配置（支持 urls 为数组）
type WebRTCICEServer struct {
	URLs       []string `mapstructure:"urls"`
	Username   string   `mapstructure:"username,omitempty"`
	Credential string   `mapstructure:"credential,omitempty"`
}

// MediaConfig 媒体配置
type MediaConfig struct {
	Video VideoConfig `mapstructure:"video"`
	Audio AudioConfig `mapstructure:"audio"`
}

// VideoConfig 视频配置
type VideoConfig struct {
	MaxBitrate   int `mapstructure:"max_bitrate"`
	MaxFramerate int `mapstructure:"max_framerate"`
}

// AudioConfig 音频配置
type AudioConfig struct {
	MaxBitrate int `mapstructure:"max_bitrate"`
	SampleRate int `mapstructure:"sample_rate"`
}

// ServicesConfig 外部服务配置
type ServicesConfig struct {
	SignalingService    ServiceConfig `mapstructure:"signaling_service"`
	UserService         ServiceConfig `mapstructure:"user_service"`
	MeetingService      ServiceConfig `mapstructure:"meeting_service"`
	AIService           ServiceConfig `mapstructure:"ai_service"`
	MediaService        ServiceConfig `mapstructure:"media_service"`
	NotificationService ServiceConfig `mapstructure:"notification_service"`
}

// ServiceConfig 单个服务配置
type ServiceConfig struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	GrpcPort int           `mapstructure:"grpc_port"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// EtcdConfig etcd配置
type EtcdConfig struct {
	Endpoints   []string `mapstructure:"endpoints"`
	DialTimeout int      `mapstructure:"dial_timeout"`
	TTL         int      `mapstructure:"ttl"`
	Username    string   `mapstructure:"username"`
	Password    string   `mapstructure:"password"`
	Namespace   string   `mapstructure:"namespace"`
}

// MessageQueueConfig 消息队列配置
type MessageQueueConfig struct {
	Enabled               bool   `mapstructure:"enabled"`
	Type                  string `mapstructure:"type"` // redis, memory
	QueueName             string `mapstructure:"queue_name"`
	Workers               int    `mapstructure:"workers"`
	VisibilityTimeout     int    `mapstructure:"visibility_timeout"` // 秒
	PollInterval          int    `mapstructure:"poll_interval"`      // 毫秒
	MaxRetries            int    `mapstructure:"max_retries"`
	EnableDeadLetterQueue bool   `mapstructure:"enable_dead_letter_queue"`
}

// TaskSchedulerConfig 任务调度器配置
type TaskSchedulerConfig struct {
	Enabled            bool `mapstructure:"enabled"`
	BufferSize         int  `mapstructure:"buffer_size"`
	Workers            int  `mapstructure:"workers"`
	EnableDelayedTasks bool `mapstructure:"enable_delayed_tasks"`
}

// EventBusConfig 事件总线配置
type EventBusConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Type       string `mapstructure:"type"` // redis_pubsub, local
	BufferSize int    `mapstructure:"buffer_size"`
	Workers    int    `mapstructure:"workers"`
}

// TaskDispatcherConfig 任务分发器配置
type TaskDispatcherConfig struct {
	Enabled         bool `mapstructure:"enabled"`
	EnableRouting   bool `mapstructure:"enable_routing"`
	EnableCallbacks bool `mapstructure:"enable_callbacks"`
}

var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 支持使用环境变量覆盖，形如 JWT_SECRET
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if len(config.Etcd.Endpoints) == 0 {
		if endpoints := viper.GetStringSlice("etcd.endpoints"); len(endpoints) > 0 {
			config.Etcd.Endpoints = endpoints
		}
	}

	// 显式处理 ETCD_ENDPOINTS（逗号分隔），避免在不同 Viper 版本/环境下 slice 解析不一致导致无法覆盖配置文件。
	if endpointsEnv := strings.TrimSpace(viper.GetString("ETCD_ENDPOINTS")); endpointsEnv != "" {
		parts := strings.Split(endpointsEnv, ",")
		endpoints := make([]string, 0, len(parts))
		for _, part := range parts {
			value := strings.TrimSpace(part)
			if value != "" {
				endpoints = append(endpoints, value)
			}
		}
		if len(endpoints) > 0 {
			config.Etcd.Endpoints = endpoints
		}
	}

	// 优先从环境变量读取JWT密钥
	if jwtSecret := viper.GetString("JWT_SECRET"); jwtSecret != "" {
		config.JWT.Secret = jwtSecret
		log.Println("JWT secret loaded from environment variable")
	}

	// 验证JWT密钥
	if secret := strings.TrimSpace(config.JWT.Secret); secret == "" || secret == defaultJWTSecret {
		return nil, fmt.Errorf("jwt.secret must be set via JWT_SECRET environment variable or config file (do not use default value)")
	}

	// 验证JWT密钥长度（至少32字符）
	if len(config.JWT.Secret) < 32 {
		return nil, fmt.Errorf("jwt.secret must be at least 32 characters long for security")
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
	viper.SetDefault("database.conn_max_lifetime", 3600)

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// MongoDB默认配置
	viper.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongodb.database", "meeting_system")
	viper.SetDefault("mongodb.timeout", 30)

	// MinIO默认配置
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.access_key_id", "minioadmin")
	viper.SetDefault("minio.secret_access_key", "minioadmin")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket_name", "meeting-system")

	// JWT默认配置
	viper.SetDefault("jwt.secret", defaultJWTSecret)
	viper.SetDefault("jwt.expire_time", 24)

	// ZMQ默认配置
	viper.SetDefault("zmq.unit_manager_host", "localhost")
	viper.SetDefault("zmq.unit_manager_port", 5001)
	viper.SetDefault("zmq.unit_name", "meeting_ai_service")
	viper.SetDefault("zmq.timeout", 30)

	// AI 默认配置（仅 ai-inference-service 使用；其他服务可忽略）
	viper.SetDefault("ai.runtime.backend", "triton")
	viper.SetDefault("ai.runtime.providers", []string{"cuda", "cpu"})
	viper.SetDefault("ai.runtime.device_id", 0)
	viper.SetDefault("ai.runtime.library_path", "/opt/onnxruntime/lib/libonnxruntime.so")
	viper.SetDefault("ai.runtime.intra_op_threads", 2)
	viper.SetDefault("ai.runtime.inter_op_threads", 2)
	viper.SetDefault("ai.runtime.enable_fp16", false)
	viper.SetDefault("ai.runtime.enable_tensorrt", false)
	viper.SetDefault("ai.runtime.triton.endpoint", "http://localhost:8000")
	viper.SetDefault("ai.runtime.triton.timeout_ms", 60000)

	viper.SetDefault("ai.models.asr.model_name", "whisper")
	viper.SetDefault("ai.models.asr.model_path", "")
	viper.SetDefault("ai.models.asr.input_name", "mel")
	viper.SetDefault("ai.models.asr.output_names", []string{"encoder_output"})
	viper.SetDefault("ai.models.asr.input_type", "mel")
	viper.SetDefault("ai.models.asr.decoder_path", "")
	viper.SetDefault("ai.models.asr.decoder_input_names", []string{"tokens", "encoder_output"})
	viper.SetDefault("ai.models.asr.decoder_output_names", []string{"logits"})
	viper.SetDefault("ai.models.asr.tokenizer_path", "")
	viper.SetDefault("ai.models.asr.special_tokens_path", "")
	viper.SetDefault("ai.models.asr.config_path", "")
	viper.SetDefault("ai.models.asr.sample_rate", 16000)
	viper.SetDefault("ai.models.asr.channels", 1)
	viper.SetDefault("ai.models.asr.timeout", 30)
	viper.SetDefault("ai.models.asr.max_concurrent", 10)
	viper.SetDefault("ai.models.emotion.model_name", "emotion")
	viper.SetDefault("ai.models.emotion.model_path", "")
	viper.SetDefault("ai.models.emotion.input_name", "audio_input")
	viper.SetDefault("ai.models.emotion.output_names", []string{"logits"})
	viper.SetDefault("ai.models.emotion.input_type", "waveform")
	viper.SetDefault("ai.models.emotion.labels_path", "")
	viper.SetDefault("ai.models.emotion.sample_rate", 16000)
	viper.SetDefault("ai.models.emotion.channels", 1)
	viper.SetDefault("ai.models.emotion.timeout", 15)
	viper.SetDefault("ai.models.emotion.max_concurrent", 20)
	viper.SetDefault("ai.models.synthesis.model_name", "synthesis")
	viper.SetDefault("ai.models.synthesis.model_path", "")
	viper.SetDefault("ai.models.synthesis.input_name", "audio_input")
	viper.SetDefault("ai.models.synthesis.output_names", []string{"synthesis_output"})
	viper.SetDefault("ai.models.synthesis.input_type", "mel")
	viper.SetDefault("ai.models.synthesis.sample_rate", 16000)
	viper.SetDefault("ai.models.synthesis.channels", 1)
	viper.SetDefault("ai.models.synthesis.timeout", 20)
	viper.SetDefault("ai.models.synthesis.max_concurrent", 15)
	viper.SetDefault("ai.request.max_retries", 3)
	viper.SetDefault("ai.request.retry_delay", 1000)
	viper.SetDefault("ai.request.timeout", 30)
	viper.SetDefault("ai.streaming.flush_interval_ms", 0)
	viper.SetDefault("ai.streaming.max_buffer_ms", 10000)

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.filename", "logs/app.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.max_backups", 10)
	viper.SetDefault("log.compress", true)

	// 服务默认配置
	viper.SetDefault("services.signaling_service.host", "127.0.0.1")
	viper.SetDefault("services.signaling_service.port", 8081)
	viper.SetDefault("services.signaling_service.timeout", "5s")
	viper.SetDefault("services.user_service.host", "127.0.0.1")
	viper.SetDefault("services.user_service.port", 8080)
	viper.SetDefault("services.user_service.timeout", "5s")
	viper.SetDefault("services.meeting_service.host", "127.0.0.1")
	viper.SetDefault("services.meeting_service.port", 8082)
	viper.SetDefault("services.meeting_service.timeout", "5s")
	viper.SetDefault("services.ai_service.host", "ai-inference-service")
	viper.SetDefault("services.ai_service.port", 8085)
	viper.SetDefault("services.ai_service.grpc_port", 9085)
	viper.SetDefault("services.ai_service.timeout", "10s")

	// etcd默认配置
	viper.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	viper.SetDefault("etcd.dial_timeout", 5)
	viper.SetDefault("etcd.ttl", 30)
	viper.SetDefault("etcd.namespace", "/services")

	// 消息队列默认配置
	viper.SetDefault("message_queue.enabled", true)
	viper.SetDefault("message_queue.type", "redis")
	viper.SetDefault("message_queue.queue_name", "meeting_system")
	viper.SetDefault("message_queue.workers", 4)
	viper.SetDefault("message_queue.visibility_timeout", 30)
	viper.SetDefault("message_queue.poll_interval", 100)
	viper.SetDefault("message_queue.max_retries", 3)
	viper.SetDefault("message_queue.enable_dead_letter_queue", true)

	// 任务调度器默认配置
	viper.SetDefault("task_scheduler.enabled", true)
	viper.SetDefault("task_scheduler.buffer_size", 1000)
	viper.SetDefault("task_scheduler.workers", 8)
	viper.SetDefault("task_scheduler.enable_delayed_tasks", true)

	// 事件总线默认配置
	viper.SetDefault("event_bus.enabled", true)
	viper.SetDefault("event_bus.type", "redis_pubsub")
	viper.SetDefault("event_bus.buffer_size", 1000)
	viper.SetDefault("event_bus.workers", 4)

	// 任务分发器默认配置
	viper.SetDefault("task_dispatcher.enabled", true)
	viper.SetDefault("task_dispatcher.enable_routing", true)
	viper.SetDefault("task_dispatcher.enable_callbacks", true)
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

// GetConfig 获取全局配置
func GetConfig() *Config {
	if GlobalConfig == nil {
		log.Fatal("Config not initialized, call InitConfig first")
	}
	return GlobalConfig
}
