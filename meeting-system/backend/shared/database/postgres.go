package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"meeting-system/shared/config"
	appLogger "meeting-system/shared/logger"
)

var DB *gorm.DB

// InitPostgreSQL 初始化PostgreSQL数据库连接
func InitPostgreSQL(config config.DatabaseConfig) error {
	dsn := config.GetDSN()

	// 配置GORM日志
	gormLogger := logger.New(
		&GormLogWriter{},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	connMaxLifetime := time.Duration(config.ConnMaxLifetime) * time.Second
	if connMaxLifetime <= 0 {
		connMaxLifetime = time.Hour
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	appLogger.Info("PostgreSQL database connected successfully")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// SetDB 设置数据库实例（用于测试）
func SetDB(db *gorm.DB) {
	DB = db
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

// GormLogWriter GORM日志写入器
type GormLogWriter struct{}

func (w *GormLogWriter) Printf(format string, args ...interface{}) {
	appLogger.Info(fmt.Sprintf(format, args...))
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	for _, model := range models {
		if err := DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}

	appLogger.Info("Database migration completed successfully")
	return nil
}

// Transaction 执行事务
func Transaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}
