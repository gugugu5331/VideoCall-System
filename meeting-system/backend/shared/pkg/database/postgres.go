package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"meeting-system/shared/pkg/config"
)

// PostgresDB PostgreSQL数据库管理器
type PostgresDB struct {
	*gorm.DB
}

// NewPostgresDB 创建新的PostgreSQL连接
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	dsn := cfg.GetDSN()
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

// Close 关闭数据库连接
func (db *PostgresDB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate 自动迁移数据库表
func (db *PostgresDB) AutoMigrate(models ...interface{}) error {
	return db.DB.AutoMigrate(models...)
}

// Transaction 执行事务
func (db *PostgresDB) Transaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

// HealthCheck 健康检查
func (db *PostgresDB) HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
