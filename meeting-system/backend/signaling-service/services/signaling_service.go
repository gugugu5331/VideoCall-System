package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/grpc"
	"meeting-system/shared/logger"
	"meeting-system/shared/models"
)

// SignalingService 信令服务
type SignalingService struct {
	db          *gorm.DB
	cache       *redis.Client
	grpcClients *grpc.ServiceClients
}

var errMeetingValidationUnavailable = errors.New("meeting service validation unavailable")

// NewSignalingService 创建信令服务实例
func NewSignalingService(grpcClients *grpc.ServiceClients) *SignalingService {
	return &SignalingService{
		db:          database.GetDB(),
		cache:       database.GetRedis(),
		grpcClients: grpcClients,
	}
}

// ValidateUserAccess 检查用户是否有权加入会议
func (s *SignalingService) ValidateUserAccess(userID, meetingID uint) error {
	if err := s.validateUserAccessViaMeetingService(userID, meetingID); err != nil {
		if errors.Is(err, errMeetingValidationUnavailable) {
			logger.Warn("Meeting service unavailable, falling back to direct database check",
				logger.Uint("user_id", userID),
				logger.Uint("meeting_id", meetingID),
				logger.Err(err))
			return s.validateUserAccessFromDB(userID, meetingID)
		}
		return err
	}
	return nil
}

// CreateSession 创建信令会话
func (s *SignalingService) CreateSession(sessionID string, userID, meetingID uint, peerID string) error {
	// 检查会议是否存在
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return fmt.Errorf("meeting not found: %w", err)
	}

	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 创建数据库会话记录
	session := &models.SignalingSession{
		SessionID: sessionID,
		UserID:    userID,
		MeetingID: meetingID,
		PeerID:    peerID,
		Status:    models.SessionStatusConnecting,
		JoinedAt:  time.Now(),
	}

	if err := s.db.Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// 缓存会话信息（如果Redis可用）
	if s.cache != nil {
		cfg := config.GlobalConfig
		sessionKey := cfg.Redis.SessionPrefix + sessionID
		sessionData := map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
			"meeting_id": meetingID,
			"peer_id":    peerID,
			"status":     int(models.SessionStatusConnecting),
			"joined_at":  time.Now().Unix(),
		}

		sessionJSON, _ := json.Marshal(sessionData)
		ctx := context.Background()
		if err := s.cache.Set(ctx, sessionKey, string(sessionJSON), time.Duration(cfg.Redis.SessionTTL)*time.Second).Err(); err != nil {
			logger.Warn("Failed to cache session", logger.Err(err))
		}

		// 添加到房间
		roomKey := cfg.Redis.RoomPrefix + fmt.Sprintf("%d", meetingID)
		if err := s.cache.SAdd(ctx, roomKey, sessionID).Err(); err != nil {
			logger.Warn("Failed to add session to room", logger.Err(err))
		}
	}

	logger.Info(fmt.Sprintf("Created signaling session: %s", sessionID))
	return nil
}

// UpdateSessionStatus 更新会话状态
func (s *SignalingService) UpdateSessionStatus(sessionID string, status models.SessionStatus) error {
	// 更新数据库
	if err := s.db.Model(&models.SignalingSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":       status,
			"last_ping_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	// 更新缓存（如果Redis可用）
	if s.cache != nil {
		cfg := config.GlobalConfig
		sessionKey := cfg.Redis.SessionPrefix + sessionID
		ctx := context.Background()

		// 获取现有会话数据
		sessionData, err := s.cache.Get(ctx, sessionKey).Result()
		if err != nil {
			logger.Warn("Failed to get session from cache", logger.Err(err))
			return nil
		}

		var sessionMap map[string]interface{}
		if err := json.Unmarshal([]byte(sessionData), &sessionMap); err != nil {
			logger.Warn("Failed to unmarshal session data", logger.Err(err))
			return nil
		}

		// 更新状态
		sessionMap["status"] = int(status)
		sessionMap["last_ping_at"] = time.Now().Unix()

		updatedJSON, _ := json.Marshal(sessionMap)
		if err := s.cache.Set(ctx, sessionKey, string(updatedJSON), time.Duration(cfg.Redis.SessionTTL)*time.Second).Err(); err != nil {
			logger.Warn("Failed to update session in cache", logger.Err(err))
		}
	}

	return nil
}

// DisconnectSession 断开会话
func (s *SignalingService) DisconnectSession(sessionID string) error {
	// 获取会话信息
	var session models.SignalingSession
	if err := s.db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// 更新数据库
	now := time.Now()
	if err := s.db.Model(&session).Updates(map[string]interface{}{
		"status":          models.SessionStatusDisconnected,
		"disconnected_at": &now,
	}).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// 从缓存中移除（如果Redis可用）
	if s.cache != nil {
		cfg := config.GlobalConfig
		ctx := context.Background()

		// 从会话缓存中移除
		sessionKey := cfg.Redis.SessionPrefix + sessionID
		if err := s.cache.Del(ctx, sessionKey).Err(); err != nil {
			logger.Warn("Failed to delete session from cache", logger.Err(err))
		}

		// 从房间中移除
		roomKey := cfg.Redis.RoomPrefix + fmt.Sprintf("%d", session.MeetingID)
		if err := s.cache.SRem(ctx, roomKey, sessionID).Err(); err != nil {
			logger.Warn("Failed to remove session from room", logger.Err(err))
		}
	}

	logger.Info(fmt.Sprintf("Disconnected signaling session: %s", sessionID))
	return nil
}

// GetSession 获取会话信息
func (s *SignalingService) GetSession(sessionID string) (*models.SignalingSession, error) {
	var session models.SignalingSession
	if err := s.db.Where("session_id = ?", sessionID).
		Preload("User").
		Preload("Meeting").
		First(&session).Error; err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	return &session, nil
}

// GetRoomSessions 获取房间内的所有会话
func (s *SignalingService) GetRoomSessions(meetingID uint) ([]*models.SignalingSession, error) {
	var sessions []*models.SignalingSession
	if err := s.db.Where("meeting_id = ? AND status IN ?", meetingID,
		[]models.SessionStatus{
			models.SessionStatusConnected,
			models.SessionStatusOffering,
			models.SessionStatusAnswering,
			models.SessionStatusStable,
		}).
		Preload("User").
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get room sessions: %w", err)
	}

	return sessions, nil
}

// SaveMessage 保存信令消息
func (s *SignalingService) SaveMessage(message *models.WebSocketMessage) error {
	// 只保存重要的消息类型
	if !s.shouldSaveMessage(message.Type) {
		return nil
	}

	payloadJSON, err := json.Marshal(message.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	signalingMessage := &models.SignalingMessage{
		MessageID:   message.ID,
		SessionID:   message.SessionID,
		FromUserID:  message.FromUserID,
		ToUserID:    message.ToUserID,
		MeetingID:   message.MeetingID,
		MessageType: message.Type,
		Payload:     string(payloadJSON),
		Status:      models.MessageStatusSent,
	}

	if err := s.db.Create(signalingMessage).Error; err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// shouldSaveMessage 判断是否应该保存消息
func (s *SignalingService) shouldSaveMessage(messageType models.MessageType) bool {
	switch messageType {
	case models.MessageTypeOffer,
		models.MessageTypeAnswer,
		models.MessageTypeChat,
		models.MessageTypeJoinRoom,
		models.MessageTypeLeaveRoom:
		return true
	default:
		return false
	}
}

// GetUserInfo 获取用户信息
func (s *SignalingService) GetUserInfo(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// GetMeetingInfo 获取会议信息
func (s *SignalingService) GetMeetingInfo(meetingID uint) (*models.Meeting, error) {
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return nil, fmt.Errorf("meeting not found: %w", err)
	}

	return &meeting, nil
}

// GetActiveSessionCount 获取活跃会话数量
func (s *SignalingService) GetActiveSessionCount() (int64, error) {
	var count int64
	if err := s.db.Model(&models.SignalingSession{}).
		Where("status IN ?", []models.SessionStatus{
			models.SessionStatusConnected,
			models.SessionStatusOffering,
			models.SessionStatusAnswering,
			models.SessionStatusStable,
		}).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}

	return count, nil
}

// GetRoomParticipantCount 获取房间参与者数量
func (s *SignalingService) GetRoomParticipantCount(meetingID uint) (int64, error) {
	var count int64
	if err := s.db.Model(&models.SignalingSession{}).
		Where("meeting_id = ? AND status IN ?", meetingID, []models.SessionStatus{
			models.SessionStatusConnected,
			models.SessionStatusOffering,
			models.SessionStatusAnswering,
			models.SessionStatusStable,
		}).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count room participants: %w", err)
	}

	return count, nil
}

// CleanupExpiredSessions 清理过期会话
func (s *SignalingService) CleanupExpiredSessions() error {
	cfg := config.GlobalConfig
	expiredTime := time.Now().Add(-time.Duration(cfg.Signaling.Session.ConnectionTimeout) * time.Second)

	// 更新过期会话状态
	if err := s.db.Model(&models.SignalingSession{}).
		Where("last_ping_at < ? AND status NOT IN ?", expiredTime, []models.SessionStatus{
			models.SessionStatusDisconnected,
			models.SessionStatusFailed,
		}).
		Updates(map[string]interface{}{
			"status":          models.SessionStatusDisconnected,
			"disconnected_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	logger.Info("Cleaned up expired sessions")
	return nil
}

// GetMessageHistory 获取消息历史
func (s *SignalingService) GetMessageHistory(meetingID uint, limit int) ([]*models.SignalingMessage, error) {
	var messages []*models.SignalingMessage
	if err := s.db.Where("meeting_id = ?", meetingID).
		Preload("FromUser").
		Preload("ToUser").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	return messages, nil
}

// ValidateUserToken 验证用户令牌 (调用用户服务)
func (s *SignalingService) ValidateUserToken(ctx context.Context, token string) (*grpc.ValidateTokenResponse, error) {
	if s.grpcClients == nil || s.grpcClients.UserClient == nil {
		return nil, fmt.Errorf("user service client not available")
	}

	response, err := s.grpcClients.ValidateToken(ctx, token)
	if err != nil {
		logger.Error("Failed to validate token via gRPC", logger.Err(err))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	return response, nil
}

// ValidateUserMeetingAccess 验证用户会议访问权限 (调用会议服务)
func (s *SignalingService) ValidateUserMeetingAccess(userID, meetingID uint) error {
	return s.validateUserAccessViaMeetingService(userID, meetingID)
}

func (s *SignalingService) validateUserAccessViaMeetingService(userID, meetingID uint) error {
	if s.grpcClients == nil || s.grpcClients.MeetingClient == nil {
		return errMeetingValidationUnavailable
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := s.grpcClients.ValidateUserAccess(ctx, uint32(userID), uint32(meetingID))
	if err != nil {
		logger.Error("Failed to validate user access via meeting service",
			logger.Uint("user_id", userID),
			logger.Uint("meeting_id", meetingID),
			logger.Err(err))
		return fmt.Errorf("%w: %v", errMeetingValidationUnavailable, err)
	}

	if !response.HasAccess {
		reason := response.Error
		if reason == "" {
			reason = "access denied"
		}
		return fmt.Errorf("user %d does not have access to meeting %d: %s", userID, meetingID, reason)
	}

	return nil
}

func (s *SignalingService) validateUserAccessFromDB(userID, meetingID uint) error {
	var meeting models.Meeting
	if err := s.db.First(&meeting, meetingID).Error; err != nil {
		return fmt.Errorf("meeting not found: %w", err)
	}

	if meeting.CreatorID == userID {
		return nil
	}

	var participant models.MeetingParticipant
	err := s.db.Where("meeting_id = ? AND user_id = ?", meetingID, userID).First(&participant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not in meeting")
		}
		return fmt.Errorf("failed to query participant: %w", err)
	}

	if participant.Status == models.ParticipantStatusRejected {
		return fmt.Errorf("user is rejected from meeting")
	}

	return nil
}
