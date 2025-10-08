package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	// TracingSpanKey 用于在 gin.Context 中存储 span 的键
	TracingSpanKey = "tracing-span"
)

// Tracing 追踪中间件
func Tracing(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer := opentracing.GlobalTracer()

		// 尝试从请求头提取 span context
		spanCtx, _ := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		// 创建新的 span
		operationName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		span := tracer.StartSpan(
			operationName,
			ext.RPCServerOption(spanCtx),
		)
		defer span.Finish()

		// 设置 span 标签
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.String())
		ext.Component.Set(span, serviceName)

		// 将 span 存储到 context
		c.Set(TracingSpanKey, span)

		// 继续处理请求
		c.Next()

		// 设置响应状态码
		ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))

		// 如果有错误，标记 span
		if len(c.Errors) > 0 {
			ext.Error.Set(span, true)
			span.SetTag("error.message", c.Errors.String())
		}
	}
}

// GetSpan 从 gin.Context 获取当前 span
func GetSpan(c *gin.Context) opentracing.Span {
	if span, exists := c.Get(TracingSpanKey); exists {
		if s, ok := span.(opentracing.Span); ok {
			return s
		}
	}
	return nil
}

// StartChildSpan 从当前请求的 span 创建子 span
func StartChildSpan(c *gin.Context, operationName string) opentracing.Span {
	parentSpan := GetSpan(c)
	if parentSpan == nil {
		return opentracing.StartSpan(operationName)
	}

	return opentracing.StartSpan(
		operationName,
		opentracing.ChildOf(parentSpan.Context()),
	)
}

// TraceFunction 追踪函数执行
func TraceFunction(c *gin.Context, operationName string, fn func() error) error {
	span := StartChildSpan(c, operationName)
	defer span.Finish()

	err := fn()
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("error.message", err.Error())
	}

	return err
}

