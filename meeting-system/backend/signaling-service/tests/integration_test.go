package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/signaling-service/handlers"
	"meeting-system/signaling-service/services"
)

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
	suite.Suite
	server  *httptest.Server
	handler *handlers.WebSocketHandler
	service *services.SignalingService
	db      *gorm.DB
}

// SetupSuite 设置测试套件
func (suite *IntegrationTestSuite) SetupSuite() {
	// 初始化测试配置
	config.GlobalConfig = &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8083,
			Mode: "test",
		},
		WebSocket: config.WebSocketConfig{
			Path:            "/ws/signaling",
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     true,
			PingPeriod:      54,
			PongWait:        60,
			WriteWait:       10,
			MaxMessageSize:  512,
		},
		Redis: config.RedisConfig{
			SessionPrefix: "test:signaling:session:",
			RoomPrefix:    "test:signaling:room:",
			SessionTTL:    3600,
		},
		Signaling: config.SignalingConfig{
			Room: config.RoomConfig{
				MaxParticipants: 100,
				CleanupInterval: 300,
				InactiveTimeout: 1800,
			},
			Session: config.SessionConfig{
				HeartbeatInterval: 30,
				ConnectionTimeout: 60,
			},
			ICEServers: []config.ICEServer{
				{URLs: "stun:stun.l.google.com:19302"},
			},
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpireTime: 24,
		},
		Log: config.LogConfig{
			Level: "info",
		},
	}

	// 初始化日志
	logger.InitLogger(logger.LogConfig{
		Level:    "info",
		Filename: "",
	})

	// 初始化内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(
		&models.User{},
		&models.Meeting{},
		&models.MeetingParticipant{},
		&models.SignalingSession{},
		&models.SignalingMessage{},
	)
	suite.Require().NoError(err)

	suite.db = db
	database.SetDB(db)

	// 创建测试数据
	suite.createTestData()

	// 初始化服务和处理器
	suite.service = services.NewSignalingService(nil)
	suite.handler = handlers.NewWebSocketHandler(suite.service)

	// 创建测试服务器
	gin.SetMode(gin.TestMode)
	router := suite.setupRoutes()
	suite.server = httptest.NewServer(router)
}

// TearDownSuite 清理测试套件
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}

	if suite.handler != nil {
		suite.handler.Stop()
	}

	if sqlDB, err := suite.db.DB(); err == nil {
		sqlDB.Close()
	}
}

// setupRoutes 设置路由
func (suite *IntegrationTestSuite) setupRoutes() *gin.Engine {
	r := gin.New()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "signaling-service",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// WebSocket信令接口
	r.GET("/ws/signaling", suite.handler.HandleWebSocket)

	// API接口
	v1 := r.Group("/api/v1")
	{
		sessions := v1.Group("/sessions")
		{
			sessions.GET("/:session_id", func(c *gin.Context) {
				sessionID := c.Param("session_id")
				session, err := suite.service.GetSession(sessionID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"data": session})
			})
		}
	}

	return r
}

// createTestData 创建测试数据
func (suite *IntegrationTestSuite) createTestData() {
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
		Title:           "Integration Test Meeting",
		Description:     "Test meeting for integration tests",
		CreatorID:       1,
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(time.Hour),
		MaxParticipants: 10,
		Status:          models.MeetingStatusScheduled,
		MeetingType:     models.MeetingTypeVideo,
	}
	suite.db.Create(&meeting)

	participants := []models.MeetingParticipant{
		{
			MeetingID: 1,
			UserID:    1,
			Role:      models.ParticipantRoleHost,
			Status:    models.ParticipantStatusJoined,
		},
		{
			MeetingID: 1,
			UserID:    2,
			Role:      models.ParticipantRoleParticipant,
			Status:    models.ParticipantStatusJoined,
		},
	}

	for _, participant := range participants {
		suite.db.Create(&participant)
	}
}

func (suite *IntegrationTestSuite) generateJWT(user *models.User) string {
	claims := middleware.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
	suite.Require().NoError(err)
	return signed
}

// connectWebSocket 连接WebSocket
func (suite *IntegrationTestSuite) connectWebSocket(userID, meetingID uint, peerID string) (*websocket.Conn, error) {
	// 构建WebSocket URL
	u, err := url.Parse(suite.server.URL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/ws/signaling"

	// 添加查询参数
	q := u.Query()
	q.Set("user_id", strconv.FormatUint(uint64(userID), 10))
	q.Set("meeting_id", strconv.FormatUint(uint64(meetingID), 10))
	q.Set("peer_id", peerID)
	u.RawQuery = q.Encode()

	var user models.User
	suite.Require().NoError(suite.db.First(&user, userID).Error)
	token := suite.generateJWT(&user)

	headers := http.Header{}
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// 连接WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	return conn, err
}

// TestHealthCheck 测试健康检查
func (suite *IntegrationTestSuite) TestHealthCheck() {
	resp, err := http.Get(suite.server.URL + "/health")
	suite.NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)
}

// TestWebSocketConnection 测试WebSocket连接
func (suite *IntegrationTestSuite) TestWebSocketConnection() {
	conn, err := suite.connectWebSocket(1, 1, "integration_test_peer")
	suite.NoError(err)
	suite.NotNil(conn)
	defer conn.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证连接统计
	suite.Equal(1, suite.handler.GetClientCount())
}

// TestSignalingFlow 测试完整的信令流程
func (suite *IntegrationTestSuite) TestSignalingFlow() {
	// 连接两个客户端
	conn1, err := suite.connectWebSocket(1, 1, "peer_1")
	suite.NoError(err)
	defer conn1.Close()

	conn2, err := suite.connectWebSocket(2, 1, "peer_2")
	suite.NoError(err)
	defer conn2.Close()

	// 等待连接建立
	time.Sleep(200 * time.Millisecond)

	// 验证两个客户端都连接成功
	suite.Equal(2, suite.handler.GetClientCount())

	// 测试Offer/Answer交换
	suite.testOfferAnswerExchange(conn1, conn2)

	// 测试ICE候选交换
	suite.testICECandidateExchange(conn1, conn2)

	// 测试聊天消息
	suite.testChatMessage(conn1, conn2)
}

// testOfferAnswerExchange 测试Offer/Answer交换
func (suite *IntegrationTestSuite) testOfferAnswerExchange(conn1, conn2 *websocket.Conn) {
	// 从conn1发送Offer给用户2
	targetUserID := uint(2)
	offerMessage := models.WebSocketMessage{
		ID:        "integration_offer_1",
		Type:      models.MessageTypeOffer,
		ToUserID:  &targetUserID,
		MeetingID: 1,
		Payload: models.WebRTCOffer{
			SDP:  "integration_test_sdp_offer",
			Type: "offer",
		},
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(offerMessage)
	suite.NoError(err)

	err = conn1.WriteMessage(websocket.TextMessage, messageData)
	suite.NoError(err)

	// 从conn2接收Offer
	conn2.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedData, err := conn2.ReadMessage()
	suite.NoError(err)

	var receivedOffer models.WebSocketMessage
	err = json.Unmarshal(receivedData, &receivedOffer)
	suite.NoError(err)

	// 验证Offer消息
	suite.Equal(models.MessageTypeOffer, receivedOffer.Type)
	suite.Equal(uint(1), receivedOffer.FromUserID)
	suite.Equal(&targetUserID, receivedOffer.ToUserID)
}

// testICECandidateExchange 测试ICE候选交换
func (suite *IntegrationTestSuite) testICECandidateExchange(conn1, conn2 *websocket.Conn) {
	// 从conn1发送ICE候选给用户2
	targetUserID := uint(2)
	iceMessage := models.WebSocketMessage{
		ID:        "integration_ice_1",
		Type:      models.MessageTypeICECandidate,
		ToUserID:  &targetUserID,
		MeetingID: 1,
		Payload: models.ICECandidate{
			Candidate:     "candidate:1 1 UDP 2130706431 192.168.1.100 54400 typ host",
			SDPMid:        "0",
			SDPMLineIndex: 0,
		},
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(iceMessage)
	suite.NoError(err)

	err = conn1.WriteMessage(websocket.TextMessage, messageData)
	suite.NoError(err)

	// 从conn2接收ICE候选
	conn2.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedData, err := conn2.ReadMessage()
	suite.NoError(err)

	var receivedICE models.WebSocketMessage
	err = json.Unmarshal(receivedData, &receivedICE)
	suite.NoError(err)

	// 验证ICE候选消息
	suite.Equal(models.MessageTypeICECandidate, receivedICE.Type)
	suite.Equal(uint(1), receivedICE.FromUserID)
}

// testChatMessage 测试聊天消息
func (suite *IntegrationTestSuite) testChatMessage(conn1, conn2 *websocket.Conn) {
	// 从conn1发送聊天消息
	chatMessage := models.WebSocketMessage{
		ID:        "integration_chat_1",
		Type:      models.MessageTypeChat,
		MeetingID: 1,
		Payload: models.ChatMessage{
			Content:   "Hello from integration test!",
			UserID:    1,
			Username:  "testuser1",
			MeetingID: 1,
		},
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(chatMessage)
	suite.NoError(err)

	err = conn1.WriteMessage(websocket.TextMessage, messageData)
	suite.NoError(err)

	// 从conn2接收聊天消息
	conn2.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedData, err := conn2.ReadMessage()
	suite.NoError(err)

	var receivedChat models.WebSocketMessage
	err = json.Unmarshal(receivedData, &receivedChat)
	suite.NoError(err)

	// 验证聊天消息
	suite.Equal(models.MessageTypeChat, receivedChat.Type)
	suite.Equal(uint(1), receivedChat.FromUserID)
}

// TestMultipleRooms 测试多房间场景
func (suite *IntegrationTestSuite) TestMultipleRooms() {
	// 创建第二个会议
	meeting2 := models.Meeting{
		ID:              2,
		Title:           "Second Test Meeting",
		Description:     "Second test meeting",
		CreatorID:       1,
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(time.Hour),
		MaxParticipants: 10,
		Status:          models.MeetingStatusScheduled,
		MeetingType:     models.MeetingTypeVideo,
	}
	suite.db.Create(&meeting2)
	suite.db.Create(&models.MeetingParticipant{
		MeetingID: 2,
		UserID:    2,
		Role:      models.ParticipantRoleParticipant,
		Status:    models.ParticipantStatusJoined,
	})

	// 连接到不同房间的客户端
	conn1, err := suite.connectWebSocket(1, 1, "room1_peer1")
	suite.NoError(err)
	defer conn1.Close()

	conn2, err := suite.connectWebSocket(2, 2, "room2_peer1")
	suite.NoError(err)
	defer conn2.Close()

	// 等待连接建立
	time.Sleep(200 * time.Millisecond)

	// 验证房间统计
	roomStats := suite.handler.GetRoomStats()
	suite.Equal(1, roomStats[1]) // 房间1有1个客户端
	suite.Equal(1, roomStats[2]) // 房间2有1个客户端
}

// TestConnectionCleanup 测试连接清理
func (suite *IntegrationTestSuite) TestConnectionCleanup() {
	conn, err := suite.connectWebSocket(1, 1, "cleanup_test_peer")
	suite.NoError(err)

	// 验证连接建立
	time.Sleep(100 * time.Millisecond)
	suite.Equal(1, suite.handler.GetClientCount())

	// 关闭连接
	conn.Close()

	// 等待清理完成
	time.Sleep(300 * time.Millisecond)

	// 验证连接已清理
	suite.Equal(0, suite.handler.GetClientCount())
}

// TestIntegrationTestSuite 运行集成测试套件
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
