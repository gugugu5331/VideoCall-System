package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"meeting-system/shared/config"
	"meeting-system/shared/grpc"
	"meeting-system/shared/logger"
)

// ServiceInteractionExample 服务交互示例
type ServiceInteractionExample struct {
	grpcClients *grpc.ServiceClients
}

// NewServiceInteractionExample 创建服务交互示例
func NewServiceInteractionExample() *ServiceInteractionExample {
	cfg := &config.Config{
		// 配置信息
	}

	grpcClients := grpc.NewServiceClients(cfg)
	if err := grpcClients.Initialize(); err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}

	return &ServiceInteractionExample{
		grpcClients: grpcClients,
	}
}

// UserJoinMeetingFlow 用户加入会议的完整流程
func (s *ServiceInteractionExample) UserJoinMeetingFlow(userID, meetingID uint32, token string) error {
	ctx := context.Background()

	// 1. 验证用户令牌
	logger.Info("Step 1: Validating user token")
	tokenResponse, err := s.grpcClients.ValidateToken(ctx, token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if !tokenResponse.Valid {
		return fmt.Errorf("invalid token: %s", tokenResponse.Error)
	}

	logger.Info("Token validated successfully", 
		logger.Uint32("user_id", tokenResponse.UserId),
		logger.String("username", tokenResponse.Username))

	// 2. 获取用户信息
	logger.Info("Step 2: Getting user information")
	userResponse, err := s.grpcClients.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	logger.Info("User information retrieved", 
		logger.String("username", userResponse.Username),
		logger.String("email", userResponse.Email))

	// 3. 验证会议访问权限
	logger.Info("Step 3: Validating meeting access")
	accessResponse, err := s.grpcClients.ValidateUserAccess(ctx, userID, meetingID)
	if err != nil {
		return fmt.Errorf("access validation failed: %w", err)
	}

	if !accessResponse.HasAccess {
		return fmt.Errorf("user does not have access to meeting: %s", accessResponse.Error)
	}

	logger.Info("Meeting access validated", 
		logger.String("role", accessResponse.Role))

	// 4. 获取会议信息
	logger.Info("Step 4: Getting meeting information")
	meetingResponse, err := s.grpcClients.GetMeeting(ctx, meetingID)
	if err != nil {
		return fmt.Errorf("failed to get meeting info: %w", err)
	}

	logger.Info("Meeting information retrieved", 
		logger.String("title", meetingResponse.Title),
		logger.String("status", meetingResponse.Status))

	// 5. 通知用户加入房间
	roomID := fmt.Sprintf("meeting_%d", meetingID)
	logger.Info("Step 5: Notifying user joined room")
	
	_, err = s.grpcClients.SignalingClient.NotifyUserJoined(ctx, &grpc.NotifyUserJoinedRequest{
		RoomId:   roomID,
		UserId:   userID,
		Username: userResponse.Username,
		PeerId:   fmt.Sprintf("peer_%d_%d", userID, time.Now().Unix()),
	})
	if err != nil {
		logger.Error("Failed to notify user joined", logger.Err(err))
		// 不返回错误，继续执行
	}

	// 6. 通知媒体服务准备录制
	logger.Info("Step 6: Notifying media service for recording")
	_, err = s.grpcClients.MediaClient.NotifyRecordingStarted(ctx, &grpc.NotifyRecordingStartedRequest{
		RoomId:      roomID,
		UserId:      userID,
		RecordingId: fmt.Sprintf("rec_%d_%d", meetingID, time.Now().Unix()),
		Title:       fmt.Sprintf("%s - %s", meetingResponse.Title, userResponse.Username),
	})
	if err != nil {
		logger.Error("Failed to notify recording started", logger.Err(err))
		// 不返回错误，继续执行
	}

	// 7. 发送欢迎通知
	logger.Info("Step 7: Sending welcome notification")
	err = s.grpcClients.SendNotification(ctx, userID, "meeting_joined", 
		"Welcome to Meeting", 
		fmt.Sprintf("You have successfully joined the meeting: %s", meetingResponse.Title))
	if err != nil {
		logger.Error("Failed to send welcome notification", logger.Err(err))
		// 不返回错误，继续执行
	}

	// 8. 广播用户加入消息给房间内其他用户
	logger.Info("Step 8: Broadcasting user joined message")
	_, err = s.grpcClients.SignalingClient.BroadcastMessage(ctx, &grpc.BroadcastMessageRequest{
		RoomId:      roomID,
		FromUserId:  userID,
		MessageType: "user_joined",
		Content:     fmt.Sprintf("%s joined the meeting", userResponse.Username),
	})
	if err != nil {
		logger.Error("Failed to broadcast user joined message", logger.Err(err))
		// 不返回错误，继续执行
	}

	logger.Info("User successfully joined meeting", 
		logger.Uint32("user_id", userID),
		logger.Uint32("meeting_id", meetingID),
		logger.String("room_id", roomID))

	return nil
}

// MediaProcessingFlow 媒体处理流程
func (s *ServiceInteractionExample) MediaProcessingFlow(userID uint32, roomID string, audioData, videoData []byte) error {
	ctx := context.Background()

	// 1. 发送音频数据到AI服务进行分析
	logger.Info("Step 1: Processing audio data with AI service")
	audioResponse, err := s.grpcClients.AIClient.ProcessAudioData(ctx, &grpc.ProcessAudioDataRequest{
		AudioData:  audioData,
		Format:     "opus",
		SampleRate: 48000,
		Channels:   2,
		RoomId:     roomID,
		UserId:     userID,
	})
	if err != nil {
		logger.Error("Failed to process audio data", logger.Err(err))
	} else {
		logger.Info("Audio processing started", 
			logger.String("task_id", audioResponse.TaskId),
			logger.String("status", audioResponse.Status))
	}

	// 2. 发送视频帧到AI服务进行分析
	logger.Info("Step 2: Processing video frame with AI service")
	videoResponse, err := s.grpcClients.AIClient.ProcessVideoFrame(ctx, &grpc.ProcessVideoFrameRequest{
		FrameData: videoData,
		Format:    "h264",
		Width:     1280,
		Height:    720,
		RoomId:    roomID,
		UserId:    userID,
	})
	if err != nil {
		logger.Error("Failed to process video frame", logger.Err(err))
	} else {
		logger.Info("Video processing started", 
			logger.String("task_id", videoResponse.TaskId),
			logger.String("status", videoResponse.Status))
	}

	// 3. 通知媒体服务处理进度
	logger.Info("Step 3: Notifying media processing progress")
	_, err = s.grpcClients.MediaClient.NotifyMediaProcessing(ctx, &grpc.NotifyMediaProcessingRequest{
		UserId:   userID,
		JobId:    "media_job_" + roomID,
		JobType:  "realtime_processing",
		Progress: 50.0,
		Status:   "processing",
	})
	if err != nil {
		logger.Error("Failed to notify media processing", logger.Err(err))
	}

	// 4. 获取媒体统计信息
	logger.Info("Step 4: Getting media statistics")
	statsResponse, err := s.grpcClients.MediaClient.GetMediaStats(ctx, &grpc.GetMediaStatsRequest{
		RoomId: roomID,
	})
	if err != nil {
		logger.Error("Failed to get media stats", logger.Err(err))
	} else {
		logger.Info("Media statistics retrieved", 
			logger.Int32("active_streams", statsResponse.ActiveStreams),
			logger.Int32("recording_count", statsResponse.RecordingCount),
			logger.Float64("total_bandwidth", statsResponse.TotalBandwidth))
	}

	return nil
}

// NotificationFlow 通知流程
func (s *ServiceInteractionExample) NotificationFlow(userIDs []uint32, notificationType, title, content string) error {
	ctx := context.Background()

	// 1. 发送单个通知
	if len(userIDs) == 1 {
		logger.Info("Sending single notification")
		err := s.grpcClients.SendNotification(ctx, userIDs[0], notificationType, title, content)
		if err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}
	} else {
		// 2. 发送批量通知
		logger.Info("Sending bulk notifications", logger.Int("user_count", len(userIDs)))
		_, err := s.grpcClients.NotificationClient.SendBulkNotifications(ctx, &grpc.SendBulkNotificationsRequest{
			UserIds:           userIDs,
			Type:              notificationType,
			Title:             title,
			Content:           content,
			PushNotification:  true,
			EmailNotification: false,
		})
		if err != nil {
			return fmt.Errorf("failed to send bulk notifications: %w", err)
		}
	}

	// 3. 获取通知历史
	if len(userIDs) > 0 {
		logger.Info("Getting notification history for first user")
		historyResponse, err := s.grpcClients.NotificationClient.GetNotificationHistory(ctx, &grpc.GetNotificationHistoryRequest{
			UserId: userIDs[0],
			Limit:  10,
			Offset: 0,
		})
		if err != nil {
			logger.Error("Failed to get notification history", logger.Err(err))
		} else {
			logger.Info("Notification history retrieved", 
				logger.Int32("total_count", historyResponse.TotalCount),
				logger.Int("notifications", len(historyResponse.Notifications)))
		}
	}

	return nil
}

// HealthCheckFlow 健康检查流程
func (s *ServiceInteractionExample) HealthCheckFlow() {
	logger.Info("Performing health check on all services")
	
	status := s.grpcClients.HealthCheck()
	
	for serviceName, isHealthy := range status {
		if isHealthy {
			logger.Info("Service is healthy", logger.String("service", serviceName))
		} else {
			logger.Error("Service is unhealthy", logger.String("service", serviceName))
		}
	}
}

// Close 关闭所有连接
func (s *ServiceInteractionExample) Close() {
	s.grpcClients.Close()
}

func main() {
	// 初始化日志
	logger.InitLogger(logger.LogConfig{
		Level:      "info",
		Filename:   "logs/service_interaction_example.log",
		MaxSize:    10,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
	})

	// 创建服务交互示例
	example := NewServiceInteractionExample()
	defer example.Close()

	// 执行健康检查
	example.HealthCheckFlow()

	// 模拟用户加入会议
	userID := uint32(123)
	meetingID := uint32(456)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

	if err := example.UserJoinMeetingFlow(userID, meetingID, token); err != nil {
		logger.Error("User join meeting flow failed", logger.Err(err))
	}

	// 模拟媒体处理
	roomID := fmt.Sprintf("meeting_%d", meetingID)
	audioData := make([]byte, 1024) // 模拟音频数据
	videoData := make([]byte, 4096) // 模拟视频数据

	if err := example.MediaProcessingFlow(userID, roomID, audioData, videoData); err != nil {
		logger.Error("Media processing flow failed", logger.Err(err))
	}

	// 模拟通知发送
	userIDs := []uint32{123, 456, 789}
	if err := example.NotificationFlow(userIDs, "meeting_update", "Meeting Update", "The meeting has been updated"); err != nil {
		logger.Error("Notification flow failed", logger.Err(err))
	}

	logger.Info("Service interaction example completed successfully")
}
