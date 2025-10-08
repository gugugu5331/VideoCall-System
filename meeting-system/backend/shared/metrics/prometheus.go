package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP请求指标
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code", "service"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "service"},
	)

	// gRPC请求指标
	grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status", "service"},
	)

	grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "service"},
	)

	// 数据库连接指标
	dbConnectionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"service"},
	)

	dbConnectionsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
		[]string{"service"},
	)

	// 业务指标
	activeUsers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of active users",
		},
		[]string{"service"},
	)

	activeMeetings = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_meetings_total",
			Help: "Number of active meetings",
		},
		[]string{"service"},
	)

	webrtcConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "webrtc_connections_total",
			Help: "Number of active WebRTC connections",
		},
		[]string{"service"},
	)

	// AI推理指标
	aiInferenceRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_inference_requests_total",
			Help: "Total number of AI inference requests",
		},
		[]string{"model", "status", "service"},
	)

	aiInferenceDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_inference_duration_seconds",
			Help:    "AI inference duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
		[]string{"model", "service"},
	)

	// 系统资源指标
	memoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
		[]string{"service"},
	)

	cpuUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpu_usage_percent",
			Help: "CPU usage percentage",
		},
		[]string{"service"},
	)
)

// init 初始化Prometheus指标
func init() {
	// 注册所有指标
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		grpcRequestsTotal,
		grpcRequestDuration,
		dbConnectionsActive,
		dbConnectionsIdle,
		activeUsers,
		activeMeetings,
		webrtcConnections,
		aiInferenceRequests,
		aiInferenceDuration,
		memoryUsage,
		cpuUsage,
	)
}

// PrometheusMiddleware Gin中间件，用于收集HTTP指标
func PrometheusMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
			serviceName,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			serviceName,
		).Observe(duration)
	}
}

// RecordGRPCRequest 记录gRPC请求指标
func RecordGRPCRequest(serviceName, method, status string, duration time.Duration) {
	grpcRequestsTotal.WithLabelValues(method, status, serviceName).Inc()
	grpcRequestDuration.WithLabelValues(method, serviceName).Observe(duration.Seconds())
}

// UpdateDatabaseConnections 更新数据库连接指标
func UpdateDatabaseConnections(serviceName string, active, idle int) {
	dbConnectionsActive.WithLabelValues(serviceName).Set(float64(active))
	dbConnectionsIdle.WithLabelValues(serviceName).Set(float64(idle))
}

// UpdateActiveUsers 更新活跃用户数
func UpdateActiveUsers(serviceName string, count int) {
	activeUsers.WithLabelValues(serviceName).Set(float64(count))
}

// UpdateActiveMeetings 更新活跃会议数
func UpdateActiveMeetings(serviceName string, count int) {
	activeMeetings.WithLabelValues(serviceName).Set(float64(count))
}

// UpdateWebRTCConnections 更新WebRTC连接数
func UpdateWebRTCConnections(serviceName string, count int) {
	webrtcConnections.WithLabelValues(serviceName).Set(float64(count))
}

// RecordAIInference 记录AI推理指标
func RecordAIInference(serviceName, model, status string, duration time.Duration) {
	aiInferenceRequests.WithLabelValues(model, status, serviceName).Inc()
	aiInferenceDuration.WithLabelValues(model, serviceName).Observe(duration.Seconds())
}

// UpdateSystemMetrics 更新系统资源指标
func UpdateSystemMetrics(serviceName string, memoryBytes uint64, cpuPercent float64) {
	memoryUsage.WithLabelValues(serviceName).Set(float64(memoryBytes))
	cpuUsage.WithLabelValues(serviceName).Set(cpuPercent)
}

// MetricsHandler 返回Prometheus指标处理器
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// StartMetricsServer 启动指标服务器
func StartMetricsServer(port int) error {
	http.Handle("/metrics", MetricsHandler())
	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	CollectMetrics() map[string]float64
}

// ServiceMetrics 服务指标结构
type ServiceMetrics struct {
	ServiceName string
	Collectors  []MetricsCollector
}

// NewServiceMetrics 创建服务指标收集器
func NewServiceMetrics(serviceName string) *ServiceMetrics {
	return &ServiceMetrics{
		ServiceName: serviceName,
		Collectors:  make([]MetricsCollector, 0),
	}
}

// AddCollector 添加指标收集器
func (sm *ServiceMetrics) AddCollector(collector MetricsCollector) {
	sm.Collectors = append(sm.Collectors, collector)
}

// StartCollection 开始指标收集
func (sm *ServiceMetrics) StartCollection(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			sm.collectAllMetrics()
		}
	}()
}

// collectAllMetrics 收集所有指标
func (sm *ServiceMetrics) collectAllMetrics() {
	for _, collector := range sm.Collectors {
		metrics := collector.CollectMetrics()
		for name, value := range metrics {
			switch name {
			case "active_users":
				UpdateActiveUsers(sm.ServiceName, int(value))
			case "active_meetings":
				UpdateActiveMeetings(sm.ServiceName, int(value))
			case "webrtc_connections":
				UpdateWebRTCConnections(sm.ServiceName, int(value))
			case "memory_usage":
				UpdateSystemMetrics(sm.ServiceName, uint64(value), 0)
			case "cpu_usage":
				UpdateSystemMetrics(sm.ServiceName, 0, value)
			case "db_connections_active":
				dbConnectionsActive.WithLabelValues(sm.ServiceName).Set(value)
			case "db_connections_idle":
				dbConnectionsIdle.WithLabelValues(sm.ServiceName).Set(value)
			}
		}
	}
}

// HealthMetrics 健康检查指标
var (
	serviceHealthStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health_status",
			Help: "Service health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service", "endpoint"},
	)

	serviceUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "Service uptime in seconds",
		},
		[]string{"service"},
	)
)

func init() {
	prometheus.MustRegister(serviceHealthStatus, serviceUptime)
}

// UpdateServiceHealth 更新服务健康状态
func UpdateServiceHealth(serviceName, endpoint string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	serviceHealthStatus.WithLabelValues(serviceName, endpoint).Set(status)
}

// UpdateServiceUptime 更新服务运行时间
func UpdateServiceUptime(serviceName string, startTime time.Time) {
	uptime := time.Since(startTime).Seconds()
	serviceUptime.WithLabelValues(serviceName).Set(uptime)
}
