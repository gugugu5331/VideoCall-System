package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// ServiceClients 服务客户端管理器
type ServiceClients struct {
	config *config.Config

	// gRPC连接
	userConn      *grpc.ClientConn
	meetingConn   *grpc.ClientConn
	signalingConn *grpc.ClientConn
	mediaConn     *grpc.ClientConn
	aiConn        *grpc.ClientConn
	notifyConn    *grpc.ClientConn

	// 客户端实例
	UserClient         UserServiceClient
	MeetingClient      MeetingServiceClient
	SignalingClient    SignalingServiceClient
	MediaClient        MediaServiceClient
	AIClient           AIServiceClient
	NotificationClient NotificationServiceClient

	// 连接状态
	connections map[string]*grpc.ClientConn
	mutex       sync.RWMutex
}

// NewServiceClients 创建服务客户端管理器
func NewServiceClients(cfg *config.Config) *ServiceClients {
	return &ServiceClients{
		config:      cfg,
		connections: make(map[string]*grpc.ClientConn),
	}
}

// Initialize 初始化所有服务连接（支持可选服务）
func (sc *ServiceClients) Initialize() error {
	// 尝试连接用户服务（必需）
	if err := sc.connectToUserService(); err != nil {
		return fmt.Errorf("failed to connect to required user service: %w", err)
	}

	// 尝试连接其他服务（可选）
	optionalServices := []struct {
		name string
		fn   func() error
	}{
		{"meeting-service", sc.connectToMeetingService},
		{"signaling-service", sc.connectToSignalingService},
		{"media-service", sc.connectToMediaService},
		{"ai-service", sc.connectToAIService},
		{"notification-service", sc.connectToNotificationService},
	}

	for _, svc := range optionalServices {
		if err := svc.fn(); err != nil {
			logger.Warn(fmt.Sprintf("Failed to connect to optional service %s: %v", svc.name, err))
		} else {
			logger.Info(fmt.Sprintf("Successfully connected to optional service %s", svc.name))
		}
	}

	logger.Info("gRPC service connections initialized")
	return nil
}

// connectToUserService 连接用户服务
func (sc *ServiceClients) connectToUserService() error {
	var svcCfg config.ServiceConfig
	if sc.config != nil {
		svcCfg = sc.config.Services.UserService
	}
	address, timeout := sc.resolveServiceEndpoint(svcCfg, "127.0.0.1", 8080, 5*time.Second)
	conn, err := sc.createConnection("user-service", address, timeout)
	if err != nil {
		return err
	}

	sc.userConn = conn
	sc.UserClient = NewUserServiceClient(conn)
	return nil
}

// connectToMeetingService 连接会议服务
func (sc *ServiceClients) connectToMeetingService() error {
	var svcCfg config.ServiceConfig
	if sc.config != nil {
		svcCfg = sc.config.Services.MeetingService
	}
	address, timeout := sc.resolveServiceEndpoint(svcCfg, "127.0.0.1", 8082, 5*time.Second)
	conn, err := sc.createConnection("meeting-service", address, timeout)
	if err != nil {
		return err
	}

	sc.meetingConn = conn
	sc.MeetingClient = NewMeetingServiceClient(conn)
	return nil
}

// connectToSignalingService 连接信令服务
func (sc *ServiceClients) connectToSignalingService() error {
	var svcCfg config.ServiceConfig
	if sc.config != nil {
		svcCfg = sc.config.Services.SignalingService
	}
	address, timeout := sc.resolveServiceEndpoint(svcCfg, "127.0.0.1", 8081, 5*time.Second)
	conn, err := sc.createConnection("signaling-service", address, timeout)
	if err != nil {
		return err
	}

	sc.signalingConn = conn
	sc.SignalingClient = NewSignalingServiceClient(conn)
	return nil
}

// connectToMediaService 连接媒体服务
func (sc *ServiceClients) connectToMediaService() error {
	var svcCfg config.ServiceConfig
	if sc.config != nil {
		svcCfg = sc.config.Services.MediaService
	}
	address, timeout := sc.resolveServiceEndpoint(svcCfg, "127.0.0.1", 8083, 5*time.Second)
	conn, err := sc.createConnection("media-service", address, timeout)
	if err != nil {
		return err
	}

	sc.mediaConn = conn
	sc.MediaClient = NewMediaServiceClient(conn)
	return nil
}

// connectToAIService 连接AI服务
func (sc *ServiceClients) connectToAIService() error {
	var svcCfg config.ServiceConfig
	if sc.config != nil {
		svcCfg = sc.config.Services.AIService
	}
	address, timeout := sc.resolveServiceEndpoint(svcCfg, "127.0.0.1", 8084, 10*time.Second)
	conn, err := sc.createConnection("ai-service", address, timeout)
	if err != nil {
		return err
	}

	sc.aiConn = conn
	sc.AIClient = NewAIServiceClient(conn)
	return nil
}

// connectToNotificationService 连接通知服务
func (sc *ServiceClients) connectToNotificationService() error {
	var svcCfg config.ServiceConfig
	if sc.config != nil {
		svcCfg = sc.config.Services.NotificationService
	}
	address, timeout := sc.resolveServiceEndpoint(svcCfg, "127.0.0.1", 8085, 5*time.Second)
	conn, err := sc.createConnection("notification-service", address, timeout)
	if err != nil {
		return err
	}

	sc.notifyConn = conn
	sc.NotificationClient = NewNotificationServiceClient(conn)
	return nil
}

// createConnection 创建gRPC连接
func (sc *ServiceClients) createConnection(serviceName, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	// 检查是否已存在连接
	if conn, exists := sc.connections[serviceName]; exists {
		return conn, nil
	}

	// 连接配置
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
			grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
		),
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s at %s: %w", serviceName, address, err)
	}

	// 测试连接
	ctxCheck, cancelCheck := context.WithTimeout(context.Background(), timeout)
	conn.WaitForStateChange(ctxCheck, connectivity.Idle)
	cancelCheck()

	sc.connections[serviceName] = conn
	logger.Info(fmt.Sprintf("Connected to %s at %s", serviceName, address))

	return conn, nil
}

func (sc *ServiceClients) resolveServiceEndpoint(cfg config.ServiceConfig, defaultHost string, defaultPort int, defaultTimeout time.Duration) (string, time.Duration) {
	host := cfg.Host
	if host == "" {
		host = defaultHost
	}
	port := cfg.Port
	if port == 0 {
		port = defaultPort
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return fmt.Sprintf("%s:%d", host, port), timeout
}

// Close 关闭所有连接
func (sc *ServiceClients) Close() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	for serviceName, conn := range sc.connections {
		if err := conn.Close(); err != nil {
			logger.Error(fmt.Sprintf("Failed to close connection to %s: %v", serviceName, err))
		} else {
			logger.Info(fmt.Sprintf("Closed connection to %s", serviceName))
		}
	}

	sc.connections = make(map[string]*grpc.ClientConn)
}

// HealthCheck 健康检查
func (sc *ServiceClients) HealthCheck() map[string]bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	status := make(map[string]bool)

	for serviceName, conn := range sc.connections {
		state := conn.GetState()
		status[serviceName] = state == connectivity.Ready || state == connectivity.Idle
	}

	return status
}

// GetConnection 获取指定服务的连接
func (sc *ServiceClients) GetConnection(serviceName string) (*grpc.ClientConn, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	conn, exists := sc.connections[serviceName]
	return conn, exists
}

// 便捷方法：调用用户服务
func (sc *ServiceClients) GetUser(ctx context.Context, userID uint32) (*GetUserResponse, error) {
	if sc.UserClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	return sc.UserClient.GetUser(ctx, &GetUserRequest{
		UserId: userID,
	})
}

// 便捷方法：验证令牌
func (sc *ServiceClients) ValidateToken(ctx context.Context, token string) (*ValidateTokenResponse, error) {
	if sc.UserClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	return sc.UserClient.ValidateToken(ctx, &ValidateTokenRequest{
		Token: token,
	})
}

// 便捷方法：获取会议信息
func (sc *ServiceClients) GetMeeting(ctx context.Context, meetingID uint32) (*GetMeetingResponse, error) {
	if sc.MeetingClient == nil {
		return nil, fmt.Errorf("meeting service client not initialized")
	}

	return sc.MeetingClient.GetMeeting(ctx, &GetMeetingRequest{
		MeetingId: meetingID,
	})
}

// 便捷方法：验证用户访问权限
func (sc *ServiceClients) ValidateUserAccess(ctx context.Context, userID, meetingID uint32) (*ValidateUserAccessResponse, error) {
	if sc.MeetingClient == nil {
		return nil, fmt.Errorf("meeting service client not initialized")
	}

	return sc.MeetingClient.ValidateUserAccess(ctx, &ValidateUserAccessRequest{
		UserId:    userID,
		MeetingId: meetingID,
	})
}

// 便捷方法：发送通知
func (sc *ServiceClients) SendNotification(ctx context.Context, userID uint32, notificationType, title, content string) error {
	if sc.NotificationClient == nil {
		return fmt.Errorf("notification service client not initialized")
	}

	_, err := sc.NotificationClient.SendNotification(ctx, &SendNotificationRequest{
		UserId:            userID,
		Type:              notificationType,
		Title:             title,
		Content:           content,
		PushNotification:  true,
		EmailNotification: false,
	})

	return err
}

// 便捷方法：广播消息
func (sc *ServiceClients) BroadcastMessage(ctx context.Context, roomID string, fromUserID uint32, messageType, content string) error {
	if sc.SignalingClient == nil {
		return fmt.Errorf("signaling service client not initialized")
	}

	_, err := sc.SignalingClient.BroadcastMessage(ctx, &BroadcastMessageRequest{
		RoomId:      roomID,
		FromUserId:  fromUserID,
		MessageType: messageType,
		Content:     content,
	})

	return err
}
