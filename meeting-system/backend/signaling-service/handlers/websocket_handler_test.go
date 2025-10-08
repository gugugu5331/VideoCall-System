package handlers

import (
	"context"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	sharedgrpc "meeting-system/shared/grpc"
	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/signaling-service/services"
	"sync"
)

// WebSocketHandlerTestSuite WebSocket处理器测试套件
type WebSocketHandlerTestSuite struct {
	suite.Suite
	handler       *WebSocketHandler
	service       *services.SignalingService
	db            *gorm.DB
	server        *httptest.Server
	meetingClient *mockMeetingClient
}

type mockMeetingClient struct {
	sharedgrpc.MeetingServiceClient
	mu    sync.Mutex
	calls map[uint32]map[uint32]int
}

func (m *mockMeetingClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make(map[uint32]map[uint32]int)
}

func (m *mockMeetingClient) wasCalledWith(userID, meetingID uint32) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inner, ok := m.calls[meetingID]; ok {
		_, exists := inner[userID]
		return exists
	}
	return false
}

func (m *mockMeetingClient) ValidateUserAccess(ctx context.Context, req *sharedgrpc.ValidateUserAccessRequest, opts ...grpc.CallOption) (*sharedgrpc.ValidateUserAccessResponse, error) {
	m.mu.Lock()
	if m.calls == nil {
		m.calls = make(map[uint32]map[uint32]int)
	}
	inner, ok := m.calls[req.GetMeetingId()]
	if !ok {
		inner = make(map[uint32]int)
		m.calls[req.GetMeetingId()] = inner
	}
	inner[req.GetUserId()]++
	m.mu.Unlock()
	return &sharedgrpc.ValidateUserAccessResponse{
		HasAccess: true,
		Role:      "participant",
	}, nil
}

func (m *mockMeetingClient) countCalls(meetingID uint32) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	if inner, ok := m.calls[meetingID]; ok {
		for range inner {
			total++
		}
	}
	return total
}

func (m *mockMeetingClient) GetMeeting(ctx context.Context, in *sharedgrpc.GetMeetingRequest, opts ...grpc.CallOption) (*sharedgrpc.GetMeetingResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *mockMeetingClient) UpdateMeetingStatus(ctx context.Context, in *sharedgrpc.UpdateMeetingStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *mockMeetingClient) GetActiveMeetings(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*sharedgrpc.GetActiveMeetingsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// SetupSuite 设置测试套件
func (suite *WebSocketHandlerTestSuite) SetupSuite() {
	// 初始化测试配置
	config.GlobalConfig = &config.Config{
		WebSocket: config.WebSocketConfig{
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
	}

	// 初始化内存数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	suite.Require().NoError(err)
	sqlDB, err := db.DB()
	suite.Require().NoError(err)
	sqlDB.SetMaxOpenConns(1)
	db.Exec("PRAGMA busy_timeout = 5000")

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
	suite.meetingClient = &mockMeetingClient{}
	grpcClients := &sharedgrpc.ServiceClients{MeetingClient: suite.meetingClient}
	suite.service = services.NewSignalingService(grpcClients)
	suite.handler = NewWebSocketHandler(suite.service)

	// 创建测试服务器
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws/signaling", suite.handler.HandleWebSocket)
	suite.server = httptest.NewServer(router)
}

// TearDownSuite 清理测试套件
func (suite *WebSocketHandlerTestSuite) TearDownSuite() {
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

// createTestData 创建测试数据
func (suite *WebSocketHandlerTestSuite) createTestData() {
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

func (suite *WebSocketHandlerTestSuite) generateJWT(user *models.User) string {
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
func (suite *WebSocketHandlerTestSuite) connectWebSocket(userID, meetingID uint, peerID string) (*websocket.Conn, error) {
	suite.meetingClient.reset()
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

// TestWebSocketConnection 测试WebSocket连接
func (suite *WebSocketHandlerTestSuite) TestWebSocketConnection() {
	conn, err := suite.connectWebSocket(1, 1, "test_peer_1")
	suite.NoError(err)
	suite.NotNil(conn)
	defer conn.Close()
	suite.True(suite.meetingClient.wasCalledWith(1, 1))

	// 验证客户端数量
	suite.Equal(1, suite.handler.GetClientCount())

	// 验证房间统计
	roomStats := suite.handler.GetRoomStats()
	suite.Equal(1, roomStats[1])
}

// TestWebSocketConnectionWithMissingParams 测试缺少参数的WebSocket连接
func (suite *WebSocketHandlerTestSuite) TestWebSocketConnectionWithMissingParams() {
	// 构建不完整的URL
	u, err := url.Parse(suite.server.URL)
	suite.NoError(err)
	u.Scheme = "ws"
	u.Path = "/ws/signaling"

	// 只设置部分参数
	q := u.Query()
	q.Set("user_id", "1")
	// 缺少meeting_id和peer_id
	u.RawQuery = q.Encode()

	// 尝试连接，应该失败
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if conn != nil {
		conn.Close()
	}

	// 应该返回400错误
	suite.Error(err)
	if resp != nil {
		suite.Equal(http.StatusBadRequest, resp.StatusCode)
	}
}

// TestMultipleConnections 测试多个连接
func (suite *WebSocketHandlerTestSuite) TestMultipleConnections() {
	// 连接多个客户端
	conn1, err := suite.connectWebSocket(1, 1, "peer_1")
	suite.NoError(err)
	defer conn1.Close()

	conn2, err := suite.connectWebSocket(2, 1, "peer_2")
	suite.NoError(err)
	defer conn2.Close()
	suite.True(suite.meetingClient.wasCalledWith(2, 1))

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证客户端数量
	suite.Equal(2, suite.handler.GetClientCount())

	// 验证房间统计
	roomStats := suite.handler.GetRoomStats()
	suite.Equal(2, roomStats[1])
}

// TestRoomUpdatesOnNewParticipant 测试新参会者加入时的实时更新
func (suite *WebSocketHandlerTestSuite) TestRoomUpdatesOnNewParticipant() {
	// 建立第一个用户的连接
	conn1, err := suite.connectWebSocket(1, 1, "peer_1")
	suite.NoError(err)
	defer conn1.Close()

	// 确保连接完成
	time.Sleep(100 * time.Millisecond)

	// 发送加入房间消息
	joinMsg1 := models.WebSocketMessage{
		ID:         fmt.Sprintf("join_%d", time.Now().UnixNano()),
		Type:       models.MessageTypeJoinRoom,
		FromUserID: 1,
		MeetingID:  1,
		PeerID:     "peer_1",
		Timestamp:  time.Now(),
		Payload: models.JoinRoomRequest{
			MeetingID: 1,
			UserID:    1,
			PeerID:    "peer_1",
		},
	}

	data, err := json.Marshal(joinMsg1)
	suite.NoError(err)
	suite.NoError(conn1.WriteMessage(websocket.TextMessage, data))

	// 第一个用户应收到房间信息，参与人数为1
	conn1.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, rawMessage, err := conn1.ReadMessage()
	suite.NoError(err)

	var roomInfo models.WebSocketMessage
	suite.NoError(json.Unmarshal(rawMessage, &roomInfo))
	suite.Equal(models.MessageTypeRoomInfo, roomInfo.Type)

	payload, ok := roomInfo.Payload.(map[string]interface{})
	suite.True(ok)
	participantCount, ok := payload["participant_count"].(float64)
	suite.True(ok)
	suite.Equal(1, int(participantCount))

	participantsRaw, ok := payload["participants"].([]interface{})
	suite.True(ok)
	suite.Equal(1, len(participantsRaw))
	participantEntry, ok := participantsRaw[0].(map[string]interface{})
	suite.True(ok)
	userIDValue, ok := participantEntry["user_id"].(float64)
	suite.True(ok)
	suite.Equal(1, int(userIDValue))
	suite.Equal("testuser1", participantEntry["username"])
	selfFlag, ok := participantEntry["is_self"].(bool)
	suite.True(ok)
	suite.True(selfFlag)

	// 建立第二个用户的连接
	conn2, err := suite.connectWebSocket(2, 1, "peer_2")
	suite.NoError(err)
	defer conn2.Close()

	// 等待连接稳定
	time.Sleep(100 * time.Millisecond)

	// 第二个用户发送加入房间消息
	joinMsg2 := models.WebSocketMessage{
		ID:         fmt.Sprintf("join_%d", time.Now().UnixNano()),
		Type:       models.MessageTypeJoinRoom,
		FromUserID: 2,
		MeetingID:  1,
		PeerID:     "peer_2",
		Timestamp:  time.Now(),
		Payload: models.JoinRoomRequest{
			MeetingID: 1,
			UserID:    2,
			PeerID:    "peer_2",
		},
	}

	data, err = json.Marshal(joinMsg2)
	suite.NoError(err)
	suite.NoError(conn2.WriteMessage(websocket.TextMessage, data))

	// 第二个用户应收到包含两人信息的房间更新（跳过可能的初始RoomInfo）
	conn2.SetReadDeadline(time.Now().Add(5 * time.Second))
	var roomInfo2 models.WebSocketMessage
	for i := 0; i < 3; i++ {
		_, rawMessage, err = conn2.ReadMessage()
		suite.NoError(err)

		suite.NoError(json.Unmarshal(rawMessage, &roomInfo2))
		if roomInfo2.Type != models.MessageTypeRoomInfo {
			continue
		}

		payload2, ok := roomInfo2.Payload.(map[string]interface{})
		suite.True(ok)
		participantCount2, ok := payload2["participant_count"].(float64)
		suite.True(ok)
		if int(participantCount2) == 2 {
			participantsRaw2, ok := payload2["participants"].([]interface{})
			suite.True(ok)
			suite.Len(participantsRaw2, 2)
			break
		}
	}
	suite.Equal(models.MessageTypeRoomInfo, roomInfo2.Type)
	payload2, ok := roomInfo2.Payload.(map[string]interface{})
	suite.True(ok)
	participantCount2, ok := payload2["participant_count"].(float64)
	suite.True(ok)
	suite.Equal(2, int(participantCount2))
	participantsRaw2, ok := payload2["participants"].([]interface{})
	suite.True(ok)
	suite.Len(participantsRaw2, 2)
	participantPresence := map[int]bool{}
	for _, entry := range participantsRaw2 {
		entryMap, ok := entry.(map[string]interface{})
		suite.True(ok)
		userIDFloat, ok := entryMap["user_id"].(float64)
		suite.True(ok)
		username, ok := entryMap["username"].(string)
		suite.True(ok)
		isSelf, ok := entryMap["is_self"].(bool)
		suite.True(ok)
		switch int(userIDFloat) {
		case 1:
			suite.Equal("testuser1", username)
			suite.False(isSelf)
		case 2:
			suite.Equal("testuser2", username)
			suite.True(isSelf)
		default:
			suite.Fail("unexpected participant id")
		}
		participantPresence[int(userIDFloat)] = true
	}
	suite.True(participantPresence[1])
	suite.True(participantPresence[2])

	conn1.SetReadDeadline(time.Now().Add(5 * time.Second))
	receivedJoin := false
	receivedRoomInfo := false
	for i := 0; i < 3; i++ {
		_, broadcastData, err := conn1.ReadMessage()
		suite.NoError(err)
		var msg models.WebSocketMessage
		suite.NoError(json.Unmarshal(broadcastData, &msg))
		switch msg.Type {
		case models.MessageTypeUserJoined:
			suite.Equal(uint(2), msg.FromUserID)
			notification, ok := msg.Payload.(map[string]interface{})
			suite.True(ok)
			joinedUserID, ok := notification["user_id"].(float64)
			suite.True(ok)
			suite.Equal(2, int(joinedUserID))
			suite.Equal("testuser2", notification["username"])
			receivedJoin = true
		case models.MessageTypeRoomInfo:
			payloadRoom, ok := msg.Payload.(map[string]interface{})
			suite.True(ok)
			count, ok := payloadRoom["participant_count"].(float64)
			suite.True(ok)
			if int(count) != 2 {
				continue
			}
			suite.Equal(2, int(count))
			participantsRaw, ok := payloadRoom["participants"].([]interface{})
			suite.True(ok)
			suite.Len(participantsRaw, 2)
			for _, entry := range participantsRaw {
				entryMap, ok := entry.(map[string]interface{})
				suite.True(ok)
				userIDFloat, ok := entryMap["user_id"].(float64)
				suite.True(ok)
				username, ok := entryMap["username"].(string)
				suite.True(ok)
				isSelf, ok := entryMap["is_self"].(bool)
				suite.True(ok)
				switch int(userIDFloat) {
				case 1:
					suite.Equal("testuser1", username)
					suite.True(isSelf)
				case 2:
					suite.Equal("testuser2", username)
					suite.False(isSelf)
				default:
					suite.Fail("unexpected participant id")
				}
			}
			receivedRoomInfo = true
		}
		if receivedJoin && receivedRoomInfo {
			break
		}
	}
	suite.True(receivedJoin)
	suite.True(receivedRoomInfo)
}

// TestMessageRouting 测试消息路由
func (suite *WebSocketHandlerTestSuite) TestMessageRouting() {
	// 连接两个客户端
	conn1, err := suite.connectWebSocket(1, 1, "peer_1")
	suite.NoError(err)
	defer conn1.Close()

	conn2, err := suite.connectWebSocket(2, 1, "peer_2")
	suite.NoError(err)
	defer conn2.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 从conn1发送聊天消息
	chatMessage := models.WebSocketMessage{
		ID:        "test_chat_1",
		Type:      models.MessageTypeChat,
		MeetingID: 1,
		Payload: models.ChatMessage{
			Content:  "Hello from user 1",
			UserID:   1,
			Username: "testuser1",
		},
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(chatMessage)
	suite.NoError(err)

	err = conn1.WriteMessage(websocket.TextMessage, messageData)
	suite.NoError(err)

	// 从conn2接收消息
	conn2.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedData, err := conn2.ReadMessage()
	suite.NoError(err)

	var receivedMessage models.WebSocketMessage
	err = json.Unmarshal(receivedData, &receivedMessage)
	suite.NoError(err)

	// 验证消息内容
	suite.Equal(models.MessageTypeChat, receivedMessage.Type)
	suite.Equal(uint(1), receivedMessage.FromUserID)
	suite.Equal(uint(1), receivedMessage.MeetingID)
}

// TestOfferAnswerExchange 测试Offer/Answer交换
func (suite *WebSocketHandlerTestSuite) TestOfferAnswerExchange() {
	// 连接两个客户端
	conn1, err := suite.connectWebSocket(1, 1, "peer_1")
	suite.NoError(err)
	defer conn1.Close()

	conn2, err := suite.connectWebSocket(2, 1, "peer_2")
	suite.NoError(err)
	defer conn2.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 从conn1发送Offer给用户2
	targetUserID := uint(2)
	offerMessage := models.WebSocketMessage{
		ID:        "test_offer_1",
		Type:      models.MessageTypeOffer,
		ToUserID:  &targetUserID,
		MeetingID: 1,
		Payload: models.WebRTCOffer{
			SDP:  "test_sdp_offer",
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

	// 从conn2发送Answer给用户1
	answerTargetUserID := uint(1)
	answerMessage := models.WebSocketMessage{
		ID:        "test_answer_1",
		Type:      models.MessageTypeAnswer,
		ToUserID:  &answerTargetUserID,
		MeetingID: 1,
		Payload: models.WebRTCAnswer{
			SDP:  "test_sdp_answer",
			Type: "answer",
		},
		Timestamp: time.Now(),
	}

	answerData, err := json.Marshal(answerMessage)
	suite.NoError(err)

	err = conn2.WriteMessage(websocket.TextMessage, answerData)
	suite.NoError(err)

	// 从conn1接收Answer
	conn1.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedAnswerData, err := conn1.ReadMessage()
	suite.NoError(err)

	var receivedAnswer models.WebSocketMessage
	err = json.Unmarshal(receivedAnswerData, &receivedAnswer)
	suite.NoError(err)

	// 验证Answer消息
	suite.Equal(models.MessageTypeAnswer, receivedAnswer.Type)
	suite.Equal(uint(2), receivedAnswer.FromUserID)
	suite.Equal(&answerTargetUserID, receivedAnswer.ToUserID)
}

// TestPingPong 测试心跳机制
func (suite *WebSocketHandlerTestSuite) TestPingPong() {
	conn, err := suite.connectWebSocket(1, 1, "ping_peer")
	suite.NoError(err)
	defer conn.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 发送Ping消息
	pingMessage := models.WebSocketMessage{
		ID:        "test_ping",
		Type:      models.MessageTypePing,
		MeetingID: 1,
		Timestamp: time.Now(),
	}

	messageData, err := json.Marshal(pingMessage)
	suite.NoError(err)

	err = conn.WriteMessage(websocket.TextMessage, messageData)
	suite.NoError(err)

	// 接收Pong响应
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedData, err := conn.ReadMessage()
	suite.NoError(err)

	var pongMessage models.WebSocketMessage
	err = json.Unmarshal(receivedData, &pongMessage)
	suite.NoError(err)

	// 验证Pong消息
	suite.Equal(models.MessageTypePong, pongMessage.Type)
	suite.Equal(uint(1), pongMessage.FromUserID)
}

// TestConnectionCleanup 测试连接清理
func (suite *WebSocketHandlerTestSuite) TestConnectionCleanup() {
	conn, err := suite.connectWebSocket(1, 1, "cleanup_peer")
	suite.NoError(err)

	// 验证连接建立
	time.Sleep(100 * time.Millisecond)
	suite.Equal(1, suite.handler.GetClientCount())

	// 关闭连接
	conn.Close()

	// 等待清理完成
	time.Sleep(200 * time.Millisecond)

	// 验证连接已清理
	suite.Equal(0, suite.handler.GetClientCount())
}

// TestInvalidMessage 测试无效消息处理
func (suite *WebSocketHandlerTestSuite) TestInvalidMessage() {
	conn, err := suite.connectWebSocket(1, 1, "invalid_peer")
	suite.NoError(err)
	defer conn.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 发送无效JSON
	err = conn.WriteMessage(websocket.TextMessage, []byte("invalid json"))
	suite.NoError(err)

	// 接收错误消息
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, receivedData, err := conn.ReadMessage()
	suite.NoError(err)

	var errorMessage models.WebSocketMessage
	err = json.Unmarshal(receivedData, &errorMessage)
	suite.NoError(err)

	// 验证错误消息
	suite.Equal(models.MessageTypeError, errorMessage.Type)
}

// TestWebSocketHandlerTestSuite 运行测试套件
func TestWebSocketHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(WebSocketHandlerTestSuite))
}
