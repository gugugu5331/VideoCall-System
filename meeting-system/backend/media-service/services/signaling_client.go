package services

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// SignalingClient 信令服务客户端
type SignalingClient struct {
	config *config.Config
	conn   *grpc.ClientConn
	// client signaling.SignalingServiceClient // 这里应该导入信令服务的gRPC客户端
}

// NotificationRequest 通知请求
type NotificationRequest struct {
	Type      string                 `json:"type"`
	RoomID    string                 `json:"room_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// NewSignalingClient 创建信令服务客户端
func NewSignalingClient(config *config.Config) *SignalingClient {
	return &SignalingClient{
		config: config,
	}
}

// Initialize 初始化信令服务客户端
func (c *SignalingClient) Initialize() error {
	if c == nil || c.config == nil {
		logger.Warn("Signaling client config missing; skip initialization")
		return nil
	}

	host := c.config.Server.Host
	if host == "" {
		host = "localhost"
	}

	port := c.config.GRPC.Port
	if port == 0 {
		port = 50053
	}

	signalingAddr := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.Dial(signalingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to signaling service: %w", err)
	}

	c.conn = conn
	// c.client = signaling.NewSignalingServiceClient(conn)

	logger.Info("Signaling client initialized successfully")
	return nil
}

// Close 关闭连接
func (c *SignalingClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// NotifyRecordingStarted 通知录制开始
func (c *SignalingClient) NotifyRecordingStarted(roomID, userID, recordingID string) error {
	req := &NotificationRequest{
		Type:   "recording_started",
		RoomID: roomID,
		UserID: userID,
		Data: map[string]interface{}{
			"recording_id": recordingID,
			"status":       "recording",
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyRecordingStopped 通知录制停止
func (c *SignalingClient) NotifyRecordingStopped(roomID, userID, recordingID string) error {
	req := &NotificationRequest{
		Type:   "recording_stopped",
		RoomID: roomID,
		UserID: userID,
		Data: map[string]interface{}{
			"recording_id": recordingID,
			"status":       "stopped",
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyRecordingCompleted 通知录制完成
func (c *SignalingClient) NotifyRecordingCompleted(roomID, userID, recordingID, filePath string, fileSize int64) error {
	req := &NotificationRequest{
		Type:   "recording_completed",
		RoomID: roomID,
		UserID: userID,
		Data: map[string]interface{}{
			"recording_id": recordingID,
			"status":       "completed",
			"file_path":    filePath,
			"file_size":    fileSize,
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyStreamStarted 通知流媒体开始
func (c *SignalingClient) NotifyStreamStarted(roomID, userID, streamID, streamURL string) error {
	req := &NotificationRequest{
		Type:   "stream_started",
		RoomID: roomID,
		UserID: userID,
		Data: map[string]interface{}{
			"stream_id":  streamID,
			"stream_url": streamURL,
			"status":     "active",
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyStreamStopped 通知流媒体停止
func (c *SignalingClient) NotifyStreamStopped(roomID, userID, streamID string) error {
	req := &NotificationRequest{
		Type:   "stream_stopped",
		RoomID: roomID,
		UserID: userID,
		Data: map[string]interface{}{
			"stream_id": streamID,
			"status":    "stopped",
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyMediaProcessingStarted 通知媒体处理开始
func (c *SignalingClient) NotifyMediaProcessingStarted(userID, jobID, jobType string) error {
	req := &NotificationRequest{
		Type:   "media_processing_started",
		UserID: userID,
		Data: map[string]interface{}{
			"job_id":   jobID,
			"job_type": jobType,
			"status":   "processing",
			"progress": 0,
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyMediaProcessingProgress 通知媒体处理进度
func (c *SignalingClient) NotifyMediaProcessingProgress(userID, jobID string, progress float64) error {
	req := &NotificationRequest{
		Type:   "media_processing_progress",
		UserID: userID,
		Data: map[string]interface{}{
			"job_id":   jobID,
			"progress": progress,
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyMediaProcessingCompleted 通知媒体处理完成
func (c *SignalingClient) NotifyMediaProcessingCompleted(userID, jobID, outputPath string) error {
	req := &NotificationRequest{
		Type:   "media_processing_completed",
		UserID: userID,
		Data: map[string]interface{}{
			"job_id":      jobID,
			"status":      "completed",
			"output_path": outputPath,
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyMediaProcessingFailed 通知媒体处理失败
func (c *SignalingClient) NotifyMediaProcessingFailed(userID, jobID, errorMsg string) error {
	req := &NotificationRequest{
		Type:   "media_processing_failed",
		UserID: userID,
		Data: map[string]interface{}{
			"job_id": jobID,
			"status": "failed",
			"error":  errorMsg,
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyMediaUploaded 通知媒体文件上传完成
func (c *SignalingClient) NotifyMediaUploaded(userID, fileID, fileName string, fileSize int64) error {
	req := &NotificationRequest{
		Type:   "media_uploaded",
		UserID: userID,
		Data: map[string]interface{}{
			"file_id":   fileID,
			"file_name": fileName,
			"file_size": fileSize,
			"status":    "uploaded",
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// NotifyFilterApplied 通知滤镜应用完成
func (c *SignalingClient) NotifyFilterApplied(userID, fileID, filterID string) error {
	req := &NotificationRequest{
		Type:   "filter_applied",
		UserID: userID,
		Data: map[string]interface{}{
			"file_id":   fileID,
			"filter_id": filterID,
			"status":    "completed",
		},
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// sendNotification 发送通知到信令服务
func (c *SignalingClient) sendNotification(req *NotificationRequest) error {
	if c.conn == nil {
		return fmt.Errorf("signaling client not initialized")
	}

	// 这里应该调用信令服务的gRPC方法
	// 由于我们还没有定义信令服务的gRPC接口，这里先模拟
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 模拟gRPC调用
	logger.Info(fmt.Sprintf("Sending notification to signaling service: %s", req.Type))

	// 实际实现应该是：
	// _, err := c.client.SendNotification(ctx, &signaling.NotificationRequest{
	//     Type:      req.Type,
	//     RoomId:    req.RoomID,
	//     UserId:    req.UserID,
	//     Data:      req.Data,
	//     Timestamp: req.Timestamp,
	// })
	// return err

	// 模拟成功
	_ = ctx
	return nil
}

// BroadcastToRoom 向房间广播消息
func (c *SignalingClient) BroadcastToRoom(roomID string, messageType string, data map[string]interface{}) error {
	req := &NotificationRequest{
		Type:      messageType,
		RoomID:    roomID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// SendToUser 向特定用户发送消息
func (c *SignalingClient) SendToUser(userID string, messageType string, data map[string]interface{}) error {
	req := &NotificationRequest{
		Type:      messageType,
		UserID:    userID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	return c.sendNotification(req)
}

// GetConnectionStatus 获取连接状态
func (c *SignalingClient) GetConnectionStatus() string {
	if c.conn == nil {
		return "disconnected"
	}

	state := c.conn.GetState()
	return state.String()
}

// Reconnect 重新连接
func (c *SignalingClient) Reconnect() error {
	if c.conn != nil {
		c.conn.Close()
	}

	return c.Initialize()
}
