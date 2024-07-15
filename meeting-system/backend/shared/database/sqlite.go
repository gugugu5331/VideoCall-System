package database

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"meeting-system/shared/config"
	appLogger "meeting-system/shared/logger"
	"meeting-system/shared/models"
)

// InitSQLite 初始化SQLite数据库连接（用于压力测试）
func InitSQLite(config config.DatabaseConfig) error {
	appLogger.Info("Starting SQLite database initialization...")
	appLogger.Info("SQLite DSN: " + config.DSN)

	// 配置GORM日志
	gormLogger := logger.New(
		&GormLogWriter{},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info, // 显示更多日志用于调试
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	appLogger.Info("Opening SQLite database connection...")
	// 连接SQLite数据库
	db, err := gorm.Open(sqlite.Open(config.DSN), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	if err != nil {
		appLogger.Error("Failed to connect to SQLite database: " + err.Error())
		return err
	}
	appLogger.Info("SQLite database connection established")

	// 配置连接池
	appLogger.Info("Configuring database connection pool...")
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("Failed to get underlying sql.DB: " + err.Error())
		return err
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	appLogger.Info("Database connection pool configured")

	// 自动迁移数据库表
	appLogger.Info("Starting database migration...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Meeting{},
		&models.MeetingParticipant{},
		&models.MeetingRecording{},
		&models.MediaStream{},
		&models.MeetingRoom{},
	)

	if err != nil {
		appLogger.Error("Failed to migrate database: " + err.Error())
		return err
	}
	appLogger.Info("Database migration completed successfully")

	DB = db
	appLogger.Info("SQLite database connected successfully")
	return nil
}

// InitDB 通用数据库初始化函数
func InitDB(config config.DatabaseConfig) error {
	if config.Driver == "sqlite" {
		return InitSQLite(config)
	} else {
		return InitPostgreSQL(config)
	}
}
