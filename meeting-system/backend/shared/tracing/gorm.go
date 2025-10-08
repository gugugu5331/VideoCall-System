package tracing

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gorm.io/gorm"
)

const (
	gormSpanKey = "gorm:span"
)

// GormTracingPlugin GORM 追踪插件
type GormTracingPlugin struct{}

// Name 插件名称
func (p *GormTracingPlugin) Name() string {
	return "gorm:tracing"
}

// Initialize 初始化插件
func (p *GormTracingPlugin) Initialize(db *gorm.DB) error {
	// 注册回调
	if err := db.Callback().Create().Before("gorm:create").Register("tracing:before_create", p.before); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").Register("tracing:after_create", p.after); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("gorm:query").Register("tracing:before_query", p.before); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register("tracing:after_query", p.after); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:update").Register("tracing:before_update", p.before); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("tracing:after_update", p.after); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("gorm:delete").Register("tracing:before_delete", p.before); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("tracing:after_delete", p.after); err != nil {
		return err
	}

	if err := db.Callback().Row().Before("gorm:row").Register("tracing:before_row", p.before); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gorm:row").Register("tracing:after_row", p.after); err != nil {
		return err
	}

	if err := db.Callback().Raw().Before("gorm:raw").Register("tracing:before_raw", p.before); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gorm:raw").Register("tracing:after_raw", p.after); err != nil {
		return err
	}

	return nil
}

// before 在操作前创建 span
func (p *GormTracingPlugin) before(db *gorm.DB) {
	// 从 context 获取父 span
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	var parentSpan opentracing.SpanContext
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		parentSpan = parent.Context()
	}

	// 创建 span
	tracer := opentracing.GlobalTracer()
	operationName := fmt.Sprintf("gorm:%s", db.Statement.Table)
	span := tracer.StartSpan(
		operationName,
		opentracing.ChildOf(parentSpan),
	)

	// 设置标签
	ext.DBType.Set(span, "sql")
	ext.DBInstance.Set(span, db.Statement.Table)
	span.SetTag("db.statement", db.Statement.SQL.String())

	// 将 span 存储到 context
	db.InstanceSet(gormSpanKey, span)
}

// after 在操作后完成 span
func (p *GormTracingPlugin) after(db *gorm.DB) {
	// 获取 span
	spanValue, exists := db.InstanceGet(gormSpanKey)
	if !exists {
		return
	}

	span, ok := spanValue.(opentracing.Span)
	if !ok {
		return
	}
	defer span.Finish()

	// 设置行数
	span.SetTag("db.rows_affected", db.Statement.RowsAffected)

	// 如果有错误，标记 span
	if db.Error != nil {
		ext.Error.Set(span, true)
		span.SetTag("error.message", db.Error.Error())
	}
}

// WithContext 为 GORM 操作添加追踪上下文
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx)
}

// WithSpan 为 GORM 操作添加 span
func WithSpan(span opentracing.Span, db *gorm.DB) *gorm.DB {
	if span == nil {
		return db
	}
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	return db.WithContext(ctx)
}

