package tracing

import (
	"fmt"
	"io"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	"meeting-system/shared/logger"
)

// InitJaeger 初始化 Jaeger 追踪
func InitJaeger(serviceName string) (opentracing.Tracer, io.Closer, error) {
	// 从环境变量读取配置
	agentHost := os.Getenv("JAEGER_AGENT_HOST")
	if agentHost == "" {
		agentHost = "localhost"
	}

	agentPort := os.Getenv("JAEGER_AGENT_PORT")
	if agentPort == "" {
		agentPort = "6831"
	}

	samplerType := os.Getenv("JAEGER_SAMPLER_TYPE")
	if samplerType == "" {
		samplerType = "const"
	}

	samplerParam := os.Getenv("JAEGER_SAMPLER_PARAM")
	if samplerParam == "" {
		samplerParam = "1"
	}

	// 配置 Jaeger
	cfg := &config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  samplerType,
			Param: 1, // 1 = 100% 采样率
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: fmt.Sprintf("%s:%s", agentHost, agentPort),
		},
	}

	// 创建 tracer
	tracer, closer, err := cfg.NewTracer(
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create tracer: %w", err)
	}

	// 设置全局 tracer
	opentracing.SetGlobalTracer(tracer)

	logger.Info("Jaeger tracer initialized",
		logger.String("service", serviceName),
		logger.String("agent", fmt.Sprintf("%s:%s", agentHost, agentPort)))

	return tracer, closer, nil
}

// StartSpan 开始一个新的 span
func StartSpan(operationName string) opentracing.Span {
	return opentracing.StartSpan(operationName)
}

// StartSpanFromContext 从上下文开始一个新的 span
func StartSpanFromContext(parent opentracing.SpanContext, operationName string) opentracing.Span {
	return opentracing.StartSpan(
		operationName,
		opentracing.ChildOf(parent),
	)
}

