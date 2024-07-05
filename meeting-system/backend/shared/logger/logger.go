package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// LogConfig 日志配置
type LogConfig struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

// InitLogger 初始化日志
func InitLogger(config LogConfig) error {
	// 创建日志目录
	if err := os.MkdirAll(filepath.Dir(config.Filename), 0755); err != nil {
		return err
	}

	// 设置日志级别
	level := zapcore.InfoLevel
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// 配置日志轮转
	writer := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
		MaxBackups: config.MaxBackups,
		Compress:   config.Compress,
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		level,
	)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// GetLogger 获取logger实例
func GetLogger() *zap.Logger {
	if Logger == nil {
		// 如果logger未初始化，使用默认配置
		config := LogConfig{
			Level:      "info",
			Filename:   "logs/app.log",
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 10,
			Compress:   true,
		}
		if err := InitLogger(config); err != nil {
			panic(err)
		}
	}
	return Logger
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Sync 同步日志
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
