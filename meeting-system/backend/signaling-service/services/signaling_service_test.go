package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/models"
)

// SignalingServiceTestSuite 信令服务测试套件
type SignalingServiceTestSuite struct {
	suite.Suite
	service *SignalingService
	db      *gorm.DB
}

// SetupSuite 设置测试套件
func (suite *SignalingServiceTestSuite) SetupSuite() {
	// 初始化测试配置
	config.GlobalConfig = &config.Config{
		Redis: config.RedisConfig{
			SessionPrefix: "test:signaling:session:",
			RoomPrefix:    "test:signaling:room:",
			SessionTTL:    3600,
		},
		Signaling: config.SignalingConfig{
			Session: config.SessionConfig{
				ConnectionTimeout: 60,
			},
		},
	}

	// 初始化内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(
		&models.User{},
		&models.Meeting{},
		&models.SignalingSession{},
		&models.SignalingMessage{},
	)
	suite.Require().NoError(err)

	suite.db = db
	database.SetDB(db)

	// 创建测试数据
	suite.createTestData()

	// 初始化服务
	suite.service = NewSignalingService(nil)
}

// TearDownSuite 清理测试套件
func (suite *SignalingServiceTestSuite) TearDownSuite() {
	// 清理数据库连接
	if sqlDB, err := suite.db.DB(); err == nil {
		sqlDB.Close()
	}
}

// createTestData 创建测试数据
func (suite *SignalingServiceTestSuite) createTestData() {
	// 创建测试用户
	users := []models.User{
		{
			ID:       1,
			Username: "testuser1",
			Email:    "test1@example.com",
			Password: "hashedpassword",
			Status:   models.UserStatusActive,
		},
		{
			ID:       2,
			Username: "testuser2",
			Email:    "test2@example.com",
			Password: "hashedpassword",
			Status:   models.UserStatusActive,
		},
	}

	for _, user := range users {
		suite.db.Create(&user)
	}

	// 创建测试会议
	meeting := models.Meeting{
		ID:              1,
		Title:           "Test Meeting",
		Description:     "Test meeting description",
		CreatorID:       1,
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(time.Hour),
		MaxParticipants: 10,
		Status:          models.MeetingStatusScheduled,
		MeetingType:     models.MeetingTypeVideo,
	}
	suite.db.Create(&meeting)
}

// TestCreateSession 测试创建会话
func (suite *SignalingServiceTestSuite) TestCreateSession() {
	sessionID := "test_session_1"
	userID := uint(1)
	meetingID := uint(1)
	peerID := "peer_1"

	err := suite.service.CreateSession(sessionID, userID, meetingID, peerID)
	suite.NoError(err)

	// 验证会话是否创建成功
	var session models.SignalingSession
	err = suite.db.Where("session_id = ?", sessionID).First(&session).Error
	suite.NoError(err)
	suite.Equal(sessionID, session.SessionID)
	suite.Equal(userID, session.UserID)
	suite.Equal(meetingID, session.MeetingID)
	suite.Equal(peerID, session.PeerID)
	suite.Equal(models.SessionStatusConnecting, session.Status)
}

// TestCreateSessionWithInvalidMeeting 测试创建会话时会议不存在
func (suite *SignalingServiceTestSuite) TestCreateSessionWithInvalidMeeting() {
	sessionID := "test_session_invalid"
	userID := uint(1)
	meetingID := uint(999) // 不存在的会议ID
	peerID := "peer_invalid"

	err := suite.service.CreateSession(sessionID, userID, meetingID, peerID)
	suite.Error(err)
	suite.Contains(err.Error(), "meeting not found")
}

// TestCreateSessionWithInvalidUser 测试创建会话时用户不存在
func (suite *SignalingServiceTestSuite) TestCreateSessionWithInvalidUser() {
	sessionID := "test_session_invalid_user"
	userID := uint(999) // 不存在的用户ID
	meetingID := uint(1)
	peerID := "peer_invalid_user"

	err := suite.service.CreateSession(sessionID, userID, meetingID, peerID)
	suite.Error(err)
	suite.Contains(err.Error(), "user not found")
}

// TestUpdateSessionStatus 测试更新会话状态
func (suite *SignalingServiceTestSuite) TestUpdateSessionStatus() {
	// 先创建会话
	sessionID := "test_session_update"
	userID := uint(1)
	meetingID := uint(1)
	peerID := "peer_update"

	err := suite.service.CreateSession(sessionID, userID, meetingID, peerID)
	suite.NoError(err)

	// 更新状态
	err = suite.service.UpdateSessionStatus(sessionID, models.SessionStatusConnected)
	suite.NoError(err)

	// 验证状态是否更新
	var session models.SignalingSession
	err = suite.db.Where("session_id = ?", sessionID).First(&session).Error
	suite.NoError(err)
	suite.Equal(models.SessionStatusConnected, session.Status)
	suite.NotNil(session.LastPingAt)
}

// TestDisconnectSession 测试断开会话
func (suite *SignalingServiceTestSuite) TestDisconnectSession() {
	// 先创建会话
	sessionID := "test_session_disconnect"
	userID := uint(1)
	meetingID := uint(1)
	peerID := "peer_disconnect"

	err := suite.service.CreateSession(sessionID, userID, meetingID, peerID)
	suite.NoError(err)

	// 断开会话
	err = suite.service.DisconnectSession(sessionID)
	suite.NoError(err)

	// 验证会话状态
	var session models.SignalingSession
	err = suite.db.Where("session_id = ?", sessionID).First(&session).Error
	suite.NoError(err)
	suite.Equal(models.SessionStatusDisconnected, session.Status)
	suite.NotNil(session.DisconnectedAt)
}

// TestGetSession 测试获取会话信息
func (suite *SignalingServiceTestSuite) TestGetSession() {
	// 先创建会话
	sessionID := "test_session_get"
	userID := uint(1)
	meetingID := uint(1)
	peerID := "peer_get"

	err := suite.service.CreateSession(sessionID, userID, meetingID, peerID)
	suite.NoError(err)

	// 获取会话
	session, err := suite.service.GetSession(sessionID)
	suite.NoError(err)
	suite.NotNil(session)
	suite.Equal(sessionID, session.SessionID)
	suite.Equal(userID, session.UserID)
	suite.Equal(meetingID, session.MeetingID)
	suite.Equal(peerID, session.PeerID)
}

// TestGetRoomSessions 测试获取房间会话列表
func (suite *SignalingServiceTestSuite) TestGetRoomSessions() {
	meetingID := uint(1)

	// 清理现有会话
	suite.db.Where("meeting_id = ?", meetingID).Delete(&models.SignalingSession{})

	// 创建多个会话
	sessions := []struct {
		sessionID string
		userID    uint
		peerID    string
		status    models.SessionStatus
	}{
		{"room_session_1", 1, "peer_1", models.SessionStatusConnected},
		{"room_session_2", 2, "peer_2", models.SessionStatusConnected},
		{"room_session_3", 1, "peer_3", models.SessionStatusDisconnected}, // 已断开，不应该包含
	}

	for _, s := range sessions {
		err := suite.service.CreateSession(s.sessionID, s.userID, meetingID, s.peerID)
		suite.NoError(err)

		if s.status != models.SessionStatusConnecting {
			err = suite.service.UpdateSessionStatus(s.sessionID, s.status)
			suite.NoError(err)
		}
	}

	// 获取房间会话
	roomSessions, err := suite.service.GetRoomSessions(meetingID)
	suite.NoError(err)
	suite.Len(roomSessions, 2) // 只有连接状态的会话

	// 验证会话信息
	sessionIDs := make([]string, len(roomSessions))
	for i, session := range roomSessions {
		sessionIDs[i] = session.SessionID
	}
	suite.Contains(sessionIDs, "room_session_1")
	suite.Contains(sessionIDs, "room_session_2")
	suite.NotContains(sessionIDs, "room_session_3")
}

// TestSaveMessage 测试保存消息
func (suite *SignalingServiceTestSuite) TestSaveMessage() {
	// 创建测试消息
	message := &models.WebSocketMessage{
		ID:         "test_message_1",
		Type:       models.MessageTypeOffer,
		FromUserID: 1,
		ToUserID:   &[]uint{2}[0],
		MeetingID:  1,
		SessionID:  "test_session",
		Payload: models.WebRTCOffer{
			SDP:  "test_sdp",
			Type: "offer",
		},
		Timestamp: time.Now(),
	}

	err := suite.service.SaveMessage(message)
	suite.NoError(err)

	// 验证消息是否保存
	var savedMessage models.SignalingMessage
	err = suite.db.Where("message_id = ?", message.ID).First(&savedMessage).Error
	suite.NoError(err)
	suite.Equal(message.ID, savedMessage.MessageID)
	suite.Equal(message.Type, savedMessage.MessageType)
	suite.Equal(message.FromUserID, savedMessage.FromUserID)
	suite.Equal(*message.ToUserID, *savedMessage.ToUserID)
	suite.Equal(message.MeetingID, savedMessage.MeetingID)
}

// TestSaveMessageShouldNotSave 测试不应该保存的消息类型
func (suite *SignalingServiceTestSuite) TestSaveMessageShouldNotSave() {
	// 创建心跳消息（不应该保存）
	message := &models.WebSocketMessage{
		ID:         "test_ping_message",
		Type:       models.MessageTypePing,
		FromUserID: 1,
		MeetingID:  1,
		SessionID:  "test_session",
		Payload:    nil,
		Timestamp:  time.Now(),
	}

	err := suite.service.SaveMessage(message)
	suite.NoError(err) // 不应该报错

	// 验证消息没有被保存
	var count int64
	suite.db.Model(&models.SignalingMessage{}).Where("message_id = ?", message.ID).Count(&count)
	suite.Equal(int64(0), count)
}

// TestGetActiveSessionCount 测试获取活跃会话数量
func (suite *SignalingServiceTestSuite) TestGetActiveSessionCount() {
	meetingID := uint(1)

	// 创建不同状态的会话
	sessions := []struct {
		sessionID string
		userID    uint
		status    models.SessionStatus
	}{
		{"active_session_1", 1, models.SessionStatusConnected},
		{"active_session_2", 2, models.SessionStatusStable},
		{"inactive_session_1", 1, models.SessionStatusDisconnected},
		{"inactive_session_2", 2, models.SessionStatusFailed},
	}

	for _, s := range sessions {
		err := suite.service.CreateSession(s.sessionID, s.userID, meetingID, "peer_"+s.sessionID)
		suite.NoError(err)

		err = suite.service.UpdateSessionStatus(s.sessionID, s.status)
		suite.NoError(err)
	}

	// 获取活跃会话数量
	count, err := suite.service.GetActiveSessionCount()
	suite.NoError(err)
	suite.Equal(int64(2), count) // 只有连接和稳定状态的会话
}

// TestGetMessageHistory 测试获取消息历史
func (suite *SignalingServiceTestSuite) TestGetMessageHistory() {
	meetingID := uint(1)

	// 创建测试消息
	messages := []*models.WebSocketMessage{
		{
			ID:         "history_msg_1",
			Type:       models.MessageTypeChat,
			FromUserID: 1,
			MeetingID:  meetingID,
			SessionID:  "session_1",
			Payload:    models.ChatMessage{Content: "Hello", UserID: 1, Username: "user1"},
			Timestamp:  time.Now().Add(-time.Hour),
		},
		{
			ID:         "history_msg_2",
			Type:       models.MessageTypeChat,
			FromUserID: 2,
			MeetingID:  meetingID,
			SessionID:  "session_2",
			Payload:    models.ChatMessage{Content: "Hi", UserID: 2, Username: "user2"},
			Timestamp:  time.Now().Add(-30 * time.Minute),
		},
	}

	for _, msg := range messages {
		err := suite.service.SaveMessage(msg)
		suite.NoError(err)
	}

	// 获取消息历史
	history, err := suite.service.GetMessageHistory(meetingID, 10)
	suite.NoError(err)
	suite.Len(history, 2)

	// 验证消息按时间倒序排列
	suite.Equal("history_msg_2", history[0].MessageID) // 更新的消息在前
	suite.Equal("history_msg_1", history[1].MessageID)
}

// TestCleanupExpiredSessions 测试清理过期会话
func (suite *SignalingServiceTestSuite) TestCleanupExpiredSessions() {
	meetingID := uint(1)

	// 创建过期会话
	expiredSessionID := "expired_session"
	err := suite.service.CreateSession(expiredSessionID, 1, meetingID, "expired_peer")
	suite.NoError(err)

	// 手动设置过期时间
	expiredTime := time.Now().Add(-2 * time.Hour)
	suite.db.Model(&models.SignalingSession{}).
		Where("session_id = ?", expiredSessionID).
		Update("last_ping_at", expiredTime)

	// 创建正常会话
	normalSessionID := "normal_session"
	err = suite.service.CreateSession(normalSessionID, 2, meetingID, "normal_peer")
	suite.NoError(err)

	// 执行清理
	err = suite.service.CleanupExpiredSessions()
	suite.NoError(err)

	// 验证过期会话被标记为断开
	var expiredSession models.SignalingSession
	err = suite.db.Where("session_id = ?", expiredSessionID).First(&expiredSession).Error
	suite.NoError(err)
	suite.Equal(models.SessionStatusDisconnected, expiredSession.Status)

	// 验证正常会话未受影响
	var normalSession models.SignalingSession
	err = suite.db.Where("session_id = ?", normalSessionID).First(&normalSession).Error
	suite.NoError(err)
	suite.Equal(models.SessionStatusConnecting, normalSession.Status)
}

// TestSignalingServiceTestSuite 运行测试套件
func TestSignalingServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SignalingServiceTestSuite))
}
