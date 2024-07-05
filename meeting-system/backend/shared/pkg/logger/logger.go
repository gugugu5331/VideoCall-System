package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"meeting-system/shared/pkg/config"
)

// Logger 日志管理器
type Logger struct {
	*logrus.Logger
}

// NewLogger 创建新的日志管理器
func NewLogger(cfg *config.LogConfig) (*Logger, error) {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)

	// 设置日志格式
	switch cfg.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "file":
		if cfg.Filename == "" {
			cfg.Filename = "app.log"
		}
		
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}

		output = &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,    // MB
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,     // days
			Compress:   cfg.Compress,
		}
	default:
		output = os.Stdout
	}

	logger.SetOutput(output)

	return &Logger{Logger: logger}, nil
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError 添加错误字段
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithService 添加服务名称字段
func (l *Logger) WithService(serviceName string) *logrus.Entry {
	return l.Logger.WithField("service", serviceName)
}

// WithRequestID 添加请求ID字段
func (l *Logger) WithRequestID(requestID string) *logrus.Entry {
	return l.Logger.WithField("request_id", requestID)
}

// WithUserID 添加用户ID字段
func (l *Logger) WithUserID(userID string) *logrus.Entry {
	return l.Logger.WithField("user_id", userID)
}

// WithMeetingID 添加会议ID字段
func (l *Logger) WithMeetingID(meetingID string) *logrus.Entry {
	return l.Logger.WithField("meeting_id", meetingID)
}
