package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"meeting-system/shared/config"
	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/models"
	"meeting-system/signaling-service/handlers"
	"meeting-system/signaling-service/services"
)

// ConcurrentStressTestSuite 并发压力测试套件
type ConcurrentStressTestSuite struct {
	suite.Suite
	server  *httptest.Server
	handler *handlers.WebSocketHandler
	service *services.SignalingService
	metrics *StressTestMetrics
}

// StressTestMetrics 压力测试指标
type StressTestMetrics struct {
	TotalConnections      int64
	SuccessfulConnections int64
	FailedConnections     int64
	MessagesSent          int64
	MessagesReceived      int64
	MessagesFailed        int64
	TotalLatency          int64 // 微秒
	MaxLatency            int64 // 微秒
	MinLatency            int64 // 微秒
	ConnectionErrors      int64
	MessageErrors         int64
	StartTime             time.Time
	EndTime               time.Time
	mu                    sync.RWMutex
}

// SetupSuite 测试套件初始化
func (suite *ConcurrentStressTestSuite) SetupSuite() {
	// 初始化配置
	config.InitConfig("../config/signaling-service.yaml")
	cfg := config.GlobalConfig

	// 初始化日志
	logger.InitLogger(logger.LogConfig{
		Level:      "info",
		Filename:   "logs/stress_test.log",
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
	})

	// 初始化数据库
	err := database.InitDB(cfg.Database)
	suite.NoError(err)

	// 初始化Redis
	err = database.InitRedis(cfg.Redis)
	if err != nil {
		logger.Warn("Redis initialization failed, continuing without Redis: " + err.Error())
	}

	// 自动迁移
	db := database.GetDB()
	err = db.AutoMigrate(
		&models.User{},
		&models.Meeting{},
		&models.MeetingParticipant{},
		&models.SignalingSession{},
		&models.SignalingMessage{},
	)
	suite.NoError(err)

	// 创建测试数据
	suite.createTestData()

	// 创建服务
	suite.service = services.NewSignalingService(nil)
	suite.handler = handlers.NewWebSocketHandler(suite.service)

	// 创建测试服务器
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws/signaling", suite.handler.HandleWebSocket)
	suite.server = httptest.NewServer(router)

	// 初始化指标
	suite.metrics = &StressTestMetrics{
		MinLatency: int64(^uint64(0) >> 1), // 最大int64值
	}
}

// TearDownSuite 测试套件清理
func (suite *ConcurrentStressTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.handler != nil {
		suite.handler.Stop()
	}
	database.CloseDB()
	if database.GetRedis() != nil {
		database.CloseRedis()
	}
}

// createTestData 创建测试数据
func (suite *ConcurrentStressTestSuite) createTestData() {
	db := database.GetDB()

	// 创建测试用户
	for i := 1; i <= 1000; i++ {
		user := &models.User{
			Username: fmt.Sprintf("stress_user_%d", i),
			Email:    fmt.Sprintf("stress_user_%d@test.com", i),
			Password: "hashed_password",
			Status:   models.UserStatusActive,
		}
		db.Create(user)
	}

	// 创建测试会议
	for i := 1; i <= 100; i++ {
		meeting := &models.Meeting{
			Title:           fmt.Sprintf("Stress Test Meeting %d", i),
			Description:     "Stress test meeting",
			CreatorID:       1,
			Status:          models.MeetingStatusOngoing,
			MaxParticipants: 1000,
		}
		db.Create(meeting)

		// 为每个会议添加参与者
		for j := 1; j <= 1000; j++ {
			participant := &models.MeetingParticipant{
				MeetingID: uint(i),
				UserID:    uint(j),
				Role:      models.ParticipantRoleParticipant,
				Status:    models.ParticipantStatusAccepted,
			}
			db.Create(participant)
		}
	}
}

// connectWebSocket 连接WebSocket
func (suite *ConcurrentStressTestSuite) connectWebSocket(userID, meetingID uint, peerID string) (*websocket.Conn, error) {
	wsURL := "ws" + suite.server.URL[4:] + fmt.Sprintf("/ws/signaling?user_id=%d&meeting_id=%d&peer_id=%s", userID, meetingID, peerID)

	// 生成测试Token
	token := suite.generateTestToken(userID)

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	return conn, err
}

// generateTestToken 生成测试Token
func (suite *ConcurrentStressTestSuite) generateTestToken(userID uint) string {
	// 简化的Token生成，实际应该使用JWT
	return fmt.Sprintf("test_token_user_%d", userID)
}

// updateMetrics 更新指标
func (suite *ConcurrentStressTestSuite) updateMetrics(latency int64) {
	suite.metrics.mu.Lock()
	defer suite.metrics.mu.Unlock()

	suite.metrics.TotalLatency += latency
	if latency > suite.metrics.MaxLatency {
		suite.metrics.MaxLatency = latency
	}
	if latency < suite.metrics.MinLatency {
		suite.metrics.MinLatency = latency
	}
}

// printMetrics 打印测试指标
func (suite *ConcurrentStressTestSuite) printMetrics() {
	suite.metrics.EndTime = time.Now()
	duration := suite.metrics.EndTime.Sub(suite.metrics.StartTime)

	fmt.Println("\n" + "="*80)
	fmt.Println("📊 信令服务并发压力测试报告")
	fmt.Println("="*80)
	fmt.Printf("测试时长: %v\n", duration)
	fmt.Printf("总连接数: %d\n", suite.metrics.TotalConnections)
	fmt.Printf("成功连接: %d (%.2f%%)\n", suite.metrics.SuccessfulConnections,
		float64(suite.metrics.SuccessfulConnections)/float64(suite.metrics.TotalConnections)*100)
	fmt.Printf("失败连接: %d (%.2f%%)\n", suite.metrics.FailedConnections,
		float64(suite.metrics.FailedConnections)/float64(suite.metrics.TotalConnections)*100)
	fmt.Printf("发送消息: %d\n", suite.metrics.MessagesSent)
	fmt.Printf("接收消息: %d\n", suite.metrics.MessagesReceived)
	fmt.Printf("消息失败: %d\n", suite.metrics.MessagesFailed)
	fmt.Printf("连接错误: %d\n", suite.metrics.ConnectionErrors)
	fmt.Printf("消息错误: %d\n", suite.metrics.MessageErrors)

	if suite.metrics.MessagesReceived > 0 {
		avgLatency := float64(suite.metrics.TotalLatency) / float64(suite.metrics.MessagesReceived) / 1000.0
		fmt.Printf("平均延迟: %.2f ms\n", avgLatency)
		fmt.Printf("最大延迟: %.2f ms\n", float64(suite.metrics.MaxLatency)/1000.0)
		fmt.Printf("最小延迟: %.2f ms\n", float64(suite.metrics.MinLatency)/1000.0)
	}

	if duration.Seconds() > 0 {
		throughput := float64(suite.metrics.MessagesReceived) / duration.Seconds()
		fmt.Printf("消息吞吐量: %.2f msg/s\n", throughput)
		connPerSec := float64(suite.metrics.SuccessfulConnections) / duration.Seconds()
		fmt.Printf("连接速率: %.2f conn/s\n", connPerSec)
	}

	fmt.Println("="*80)
}

// TestConcurrentConnections 测试并发连接
func (suite *ConcurrentStressTestSuite) TestConcurrentConnections() {
	testCases := []struct {
		name            string
		numConnections  int
		meetingID       uint
		expectedSuccess float64 // 期望成功率
	}{
		{"小规模并发-10连接", 10, 1, 0.95},
		{"中规模并发-50连接", 50, 1, 0.90},
		{"大规模并发-100连接", 100, 1, 0.85},
		{"超大规模并发-200连接", 200, 1, 0.80},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.runConcurrentConnectionTest(tc.numConnections, tc.meetingID, tc.expectedSuccess)
		})
	}
}

// runConcurrentConnectionTest 运行并发连接测试
func (suite *ConcurrentStressTestSuite) runConcurrentConnectionTest(numConnections int, meetingID uint, expectedSuccess float64) {
	// 重置指标
	suite.metrics = &StressTestMetrics{
		MinLatency: int64(^uint64(0) >> 1),
		StartTime:  time.Now(),
	}

	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, 0, numConnections)
	var connMutex sync.Mutex

	// 并发建立连接
	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			atomic.AddInt64(&suite.metrics.TotalConnections, 1)

			conn, err := suite.connectWebSocket(uint(userID+1), meetingID, fmt.Sprintf("peer_%d", userID))
			if err != nil {
				atomic.AddInt64(&suite.metrics.FailedConnections, 1)
				atomic.AddInt64(&suite.metrics.ConnectionErrors, 1)
				logger.Error(fmt.Sprintf("Connection failed for user %d: %v", userID, err))
				return
			}

			atomic.AddInt64(&suite.metrics.SuccessfulConnections, 1)

			connMutex.Lock()
			connections = append(connections, conn)
			connMutex.Unlock()
		}(i)
	}

	wg.Wait()

	// 等待连接稳定
	time.Sleep(500 * time.Millisecond)

	// 验证连接数
	actualSuccess := float64(suite.metrics.SuccessfulConnections) / float64(suite.metrics.TotalConnections)
	suite.GreaterOrEqual(actualSuccess, expectedSuccess,
		fmt.Sprintf("连接成功率 %.2f%% 低于预期 %.2f%%", actualSuccess*100, expectedSuccess*100))

	// 清理连接
	for _, conn := range connections {
		conn.Close()
	}

	suite.printMetrics()
}

// TestConcurrentMessaging 测试并发消息发送
func (suite *ConcurrentStressTestSuite) TestConcurrentMessaging() {
	testCases := []struct {
		name           string
		numClients     int
		messagesPerClient int
		meetingID      uint
	}{
		{"小规模消息-10客户端x10消息", 10, 10, 1},
		{"中规模消息-20客户端x20消息", 20, 20, 1},
		{"大规模消息-50客户端x10消息", 50, 10, 1},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.runConcurrentMessagingTest(tc.numClients, tc.messagesPerClient, tc.meetingID)
		})
	}
}

// runConcurrentMessagingTest 运行并发消息测试
func (suite *ConcurrentStressTestSuite) runConcurrentMessagingTest(numClients, messagesPerClient int, meetingID uint) {
	// 重置指标
	suite.metrics = &StressTestMetrics{
		MinLatency: int64(^uint64(0) >> 1),
		StartTime:  time.Now(),
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 建立连接
	connections := make([]*websocket.Conn, numClients)
	for i := 0; i < numClients; i++ {
		conn, err := suite.connectWebSocket(uint(i+1), meetingID, fmt.Sprintf("peer_%d", i))
		if err != nil {
			suite.T().Logf("Failed to connect client %d: %v", i, err)
			continue
		}
		connections[i] = conn
		atomic.AddInt64(&suite.metrics.SuccessfulConnections, 1)
	}

	// 等待连接稳定
	time.Sleep(500 * time.Millisecond)

	// 并发发送消息
	for i, conn := range connections {
		if conn == nil {
			continue
		}

		wg.Add(1)
		go func(clientID int, c *websocket.Conn) {
			defer wg.Done()

			// 启动接收协程
			receiveDone := make(chan struct{})
			go func() {
				defer close(receiveDone)
				for {
					select {
					case <-ctx.Done():
						return
					default:
						c.SetReadDeadline(time.Now().Add(5 * time.Second))
						_, _, err := c.ReadMessage()
						if err != nil {
							return
						}
						atomic.AddInt64(&suite.metrics.MessagesReceived, 1)
					}
				}
			}()

			// 发送消息
			for j := 0; j < messagesPerClient; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					startTime := time.Now()

					message := models.WebSocketMessage{
						ID:        fmt.Sprintf("msg_%d_%d", clientID, j),
						Type:      models.MessageTypeChat,
						MeetingID: meetingID,
						Payload: models.ChatMessage{
							Content:  fmt.Sprintf("Test message %d from client %d", j, clientID),
							UserID:   uint(clientID + 1),
							Username: fmt.Sprintf("stress_user_%d", clientID+1),
						},
						Timestamp: time.Now(),
					}

					data, err := json.Marshal(message)
					if err != nil {
						atomic.AddInt64(&suite.metrics.MessagesFailed, 1)
						continue
					}

					err = c.WriteMessage(websocket.TextMessage, data)
					if err != nil {
						atomic.AddInt64(&suite.metrics.MessagesFailed, 1)
						atomic.AddInt64(&suite.metrics.MessageErrors, 1)
						continue
					}

					atomic.AddInt64(&suite.metrics.MessagesSent, 1)
					latency := time.Since(startTime).Microseconds()
					suite.updateMetrics(latency)

					// 控制发送速率
					time.Sleep(10 * time.Millisecond)
				}
			}

			<-receiveDone
		}(i, conn)
	}

	wg.Wait()

	// 清理连接
	for _, conn := range connections {
		if conn != nil {
			conn.Close()
		}
	}

	suite.printMetrics()

	// 验证消息发送成功率
	expectedMessages := int64(numClients * messagesPerClient)
	successRate := float64(suite.metrics.MessagesSent) / float64(expectedMessages)
	suite.GreaterOrEqual(successRate, 0.80, "消息发送成功率应该大于80%")
}

// TestStressWithReconnection 测试带重连的压力场景
func (suite *ConcurrentStressTestSuite) TestStressWithReconnection() {
	suite.metrics = &StressTestMetrics{
		MinLatency: int64(^uint64(0) >> 1),
		StartTime:  time.Now(),
	}

	numClients := 30
	meetingID := uint(1)
	duration := 10 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup

	// 模拟客户端不断连接、断开、重连
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			reconnectCount := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					atomic.AddInt64(&suite.metrics.TotalConnections, 1)

					conn, err := suite.connectWebSocket(uint(clientID+1), meetingID, fmt.Sprintf("peer_%d_%d", clientID, reconnectCount))
					if err != nil {
						atomic.AddInt64(&suite.metrics.FailedConnections, 1)
						time.Sleep(100 * time.Millisecond)
						continue
					}

					atomic.AddInt64(&suite.metrics.SuccessfulConnections, 1)

					// 保持连接一段时间
					time.Sleep(time.Duration(500+clientID*10) * time.Millisecond)

					// 发送几条消息
					for j := 0; j < 3; j++ {
						message := models.WebSocketMessage{
							ID:        fmt.Sprintf("reconnect_msg_%d_%d_%d", clientID, reconnectCount, j),
							Type:      models.MessageTypeChat,
							MeetingID: meetingID,
							Payload: models.ChatMessage{
								Content:  fmt.Sprintf("Reconnect test %d", j),
								UserID:   uint(clientID + 1),
								Username: fmt.Sprintf("stress_user_%d", clientID+1),
							},
							Timestamp: time.Now(),
						}

						data, _ := json.Marshal(message)
						err := conn.WriteMessage(websocket.TextMessage, data)
						if err == nil {
							atomic.AddInt64(&suite.metrics.MessagesSent, 1)
						} else {
							atomic.AddInt64(&suite.metrics.MessagesFailed, 1)
						}
						time.Sleep(50 * time.Millisecond)
					}

					conn.Close()
					reconnectCount++

					// 等待一段时间再重连
					time.Sleep(200 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()
	suite.printMetrics()

	// 验证重连功能
	suite.Greater(suite.metrics.SuccessfulConnections, int64(numClients), "应该有多次成功重连")
}

// TestMixedScenario 测试混合场景
func (suite *ConcurrentStressTestSuite) TestMixedScenario() {
	suite.metrics = &StressTestMetrics{
		MinLatency: int64(^uint64(0) >> 1),
		StartTime:  time.Now(),
	}

	numClients := 50
	meetingID := uint(1)
	duration := 15 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	var wg sync.WaitGroup

	// 场景1: 长连接客户端 (30%)
	longLivedClients := int(float64(numClients) * 0.3)
	for i := 0; i < longLivedClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			suite.simulateLongLivedClient(ctx, clientID, meetingID)
		}(i)
	}

	// 场景2: 频繁重连客户端 (30%)
	reconnectClients := int(float64(numClients) * 0.3)
	for i := longLivedClients; i < longLivedClients+reconnectClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			suite.simulateReconnectClient(ctx, clientID, meetingID)
		}(i)
	}

	// 场景3: 高频消息客户端 (40%)
	highFreqClients := numClients - longLivedClients - reconnectClients
	for i := longLivedClients + reconnectClients; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			suite.simulateHighFrequencyClient(ctx, clientID, meetingID)
		}(i)
	}

	wg.Wait()
	suite.printMetrics()

	// 验证混合场景
	suite.Greater(suite.metrics.SuccessfulConnections, int64(numClients), "应该有成功的连接")
	suite.Greater(suite.metrics.MessagesSent, int64(0), "应该有消息发送")
}

// simulateLongLivedClient 模拟长连接客户端
func (suite *ConcurrentStressTestSuite) simulateLongLivedClient(ctx context.Context, clientID int, meetingID uint) {
	atomic.AddInt64(&suite.metrics.TotalConnections, 1)

	conn, err := suite.connectWebSocket(uint(clientID+1), meetingID, fmt.Sprintf("long_peer_%d", clientID))
	if err != nil {
		atomic.AddInt64(&suite.metrics.FailedConnections, 1)
		return
	}
	defer conn.Close()

	atomic.AddInt64(&suite.metrics.SuccessfulConnections, 1)

	// 定期发送心跳
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			message := models.WebSocketMessage{
				ID:        fmt.Sprintf("ping_%d_%d", clientID, time.Now().Unix()),
				Type:      models.MessageTypePing,
				MeetingID: meetingID,
				Timestamp: time.Now(),
			}

			data, _ := json.Marshal(message)
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err == nil {
				atomic.AddInt64(&suite.metrics.MessagesSent, 1)
			} else {
				atomic.AddInt64(&suite.metrics.MessagesFailed, 1)
				return
			}
		}
	}
}

// simulateReconnectClient 模拟频繁重连客户端
func (suite *ConcurrentStressTestSuite) simulateReconnectClient(ctx context.Context, clientID int, meetingID uint) {
	reconnectCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			atomic.AddInt64(&suite.metrics.TotalConnections, 1)

			conn, err := suite.connectWebSocket(uint(clientID+1), meetingID, fmt.Sprintf("reconnect_peer_%d_%d", clientID, reconnectCount))
			if err != nil {
				atomic.AddInt64(&suite.metrics.FailedConnections, 1)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			atomic.AddInt64(&suite.metrics.SuccessfulConnections, 1)

			// 短暂保持连接
			time.Sleep(1 * time.Second)
			conn.Close()

			reconnectCount++
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// simulateHighFrequencyClient 模拟高频消息客户端
func (suite *ConcurrentStressTestSuite) simulateHighFrequencyClient(ctx context.Context, clientID int, meetingID uint) {
	atomic.AddInt64(&suite.metrics.TotalConnections, 1)

	conn, err := suite.connectWebSocket(uint(clientID+1), meetingID, fmt.Sprintf("highfreq_peer_%d", clientID))
	if err != nil {
		atomic.AddInt64(&suite.metrics.FailedConnections, 1)
		return
	}
	defer conn.Close()

	atomic.AddInt64(&suite.metrics.SuccessfulConnections, 1)

	msgCount := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			message := models.WebSocketMessage{
				ID:        fmt.Sprintf("highfreq_msg_%d_%d", clientID, msgCount),
				Type:      models.MessageTypeChat,
				MeetingID: meetingID,
				Payload: models.ChatMessage{
					Content:  fmt.Sprintf("High frequency message %d", msgCount),
					UserID:   uint(clientID + 1),
					Username: fmt.Sprintf("stress_user_%d", clientID+1),
				},
				Timestamp: time.Now(),
			}

			data, _ := json.Marshal(message)
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err == nil {
				atomic.AddInt64(&suite.metrics.MessagesSent, 1)
			} else {
				atomic.AddInt64(&suite.metrics.MessagesFailed, 1)
				return
			}
			msgCount++
		}
	}
}

// TestConcurrentStressTestSuite 运行并发压力测试套件
func TestConcurrentStressTestSuite(t *testing.T) {
	suite.Run(t, new(ConcurrentStressTestSuite))
}

