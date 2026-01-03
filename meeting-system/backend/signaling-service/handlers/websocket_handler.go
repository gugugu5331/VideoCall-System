package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
	"meeting-system/shared/middleware"
	"meeting-system/shared/models"
	"meeting-system/shared/response"
	"meeting-system/signaling-service/services"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	signalingService *services.SignalingService
	upgrader         websocket.Upgrader
	clients          map[string]*Client // sessionID -> Client
	rooms            map[uint]*Room     // meetingID -> Room
	mutex            sync.RWMutex
	pingTicker       *time.Ticker
}

// Client WebSocket客户端
type Client struct {
	ID           string
	UserID       uint
	MeetingID    uint
	PeerID       string
	Username     string
	Conn         *websocket.Conn
	Send         chan []byte
	PrioritySend chan []byte
	Handler      *WebSocketHandler
	LastPing     time.Time
	JoinedAt     time.Time
	mutex        sync.Mutex
}

// Room 会议房间
type Room struct {
	ID           uint
	Clients      map[string]*Client // sessionID -> Client
	CreatedAt    time.Time
	LastActivity time.Time
	AILive       models.AILiveStatusMessage // 会议内 AI Live 共享状态（由信令服务协调）
	mutex        sync.RWMutex
}

var (
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("forbidden")
)

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler(signalingService *services.SignalingService) *WebSocketHandler {
	cfg := config.GlobalConfig

	handler := &WebSocketHandler{
		signalingService: signalingService,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.WebSocket.ReadBufferSize,
			WriteBufferSize: cfg.WebSocket.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return cfg.WebSocket.CheckOrigin
			},
		},
		clients: make(map[string]*Client),
		rooms:   make(map[uint]*Room),
	}

	// 启动心跳检查
	handler.startHeartbeat()

	// 启动房间清理
	handler.startRoomCleanup()

	return handler
}

func (h *WebSocketHandler) authorizeConnection(c *gin.Context, userID, meetingID uint) error {
	authHeader := c.GetHeader("Authorization")
	tokenString := ""
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fmt.Errorf("%w: invalid authorization header", errUnauthorized)
		}
		tokenString = strings.TrimSpace(parts[1])
	} else {
		// 浏览器 WebSocket API 无法自定义 Authorization Header，因此允许通过 Query 或 Cookie 传递 Token。
		// 生产环境建议使用一次性 WS Ticket 或 HttpOnly Cookie，以避免 Token 出现在 URL 中。
		tokenString = strings.TrimSpace(c.Query("token"))
		if tokenString == "" {
			tokenString = strings.TrimSpace(c.Query("access_token"))
		}
		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
			tokenString = strings.TrimSpace(tokenString[7:])
		}
		if tokenString == "" {
			if cookieToken, err := c.Cookie("access_token"); err == nil {
				tokenString = strings.TrimSpace(cookieToken)
			}
		}
		if tokenString == "" {
			return fmt.Errorf("%w: missing access token", errUnauthorized)
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, &middleware.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.JWT.Secret), nil
	})
	if err != nil || !token.Valid {
		return fmt.Errorf("%w: invalid token", errUnauthorized)
	}

	claims, ok := token.Claims.(*middleware.JWTClaims)
	if !ok {
		return fmt.Errorf("%w: invalid token claims", errUnauthorized)
	}
	if claims.UserID != userID {
		return fmt.Errorf("%w: token user mismatch", errForbidden)
	}

	if err := h.signalingService.ValidateUserAccess(userID, meetingID); err != nil {
		return fmt.Errorf("%w: %v", errForbidden, err)
	}

	return nil
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 获取用户ID和会议ID
	userIDStr := c.Query("user_id")
	meetingIDStr := c.Query("meeting_id")
	peerID := c.Query("peer_id")

	if userIDStr == "" || meetingIDStr == "" || peerID == "" {
		response.Error(c, http.StatusBadRequest, "Missing required parameters")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user_id")
		return
	}

	meetingID, err := strconv.ParseUint(meetingIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid meeting_id")
		return
	}

	if err := h.authorizeConnection(c, uint(userID), uint(meetingID)); err != nil {
		status := http.StatusForbidden
		if errors.Is(err, errUnauthorized) {
			status = http.StatusUnauthorized
		}
		response.Error(c, status, err.Error())
		return
	}

	username := fmt.Sprintf("user_%d", userID)
	if userInfo, err := h.signalingService.GetUserInfo(uint(userID)); err != nil {
		logger.Warn("Failed to load user info for room snapshot",
			logger.Uint("user_id", uint(userID)),
			logger.Err(err))
	} else {
		username = userInfo.Username
	}

	// 升级为WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade to WebSocket", logger.Err(err))
		return
	}

	// 创建客户端
	sessionID := fmt.Sprintf("%d_%d_%s_%d", userID, meetingID, peerID, time.Now().UnixNano())
	now := time.Now()
	client := &Client{
		ID:           sessionID,
		UserID:       uint(userID),
		MeetingID:    uint(meetingID),
		PeerID:       peerID,
		Username:     username,
		Conn:         conn,
		Send:         make(chan []byte, 2048),
		PrioritySend: make(chan []byte, 128),
		Handler:      h,
		LastPing:     now,
		JoinedAt:     now,
	}

	// 注册客户端
	h.registerClient(client)

	// 启动客户端处理协程
	go client.writePump()
	go client.readPump()

	logger.Info(fmt.Sprintf("WebSocket client connected: %s", sessionID))
}

// registerClient 注册客户端
func (h *WebSocketHandler) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 添加到客户端列表
	h.clients[client.ID] = client

	// 添加到房间
	room, exists := h.rooms[client.MeetingID]
	if !exists {
		room = &Room{
			ID:           client.MeetingID,
			Clients:      make(map[string]*Client),
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		h.rooms[client.MeetingID] = room
	}

	room.mutex.Lock()
	room.Clients[client.ID] = client
	room.LastActivity = time.Now()
	room.mutex.Unlock()

	// 创建信令会话
	if err := h.signalingService.CreateSession(client.ID, client.UserID, client.MeetingID, client.PeerID); err != nil {
		logger.Error("Failed to create signaling session", logger.Err(err))
	}
}

// unregisterClient 注销客户端
func (h *WebSocketHandler) unregisterClient(client *Client) {
	h.mutex.Lock()
	if _, exists := h.clients[client.ID]; exists {
		delete(h.clients, client.ID)
	}

	var shouldCleanup bool
	var shouldBroadcastAILive bool
	if room, exists := h.rooms[client.MeetingID]; exists {
		room.mutex.Lock()
		delete(room.Clients, client.ID)
		room.LastActivity = time.Now()
		if room.AILive.LeaderSessionID == client.ID {
			room.AILive = models.AILiveStatusMessage{
				Enabled:   false,
				UpdatedAt: time.Now(),
			}
			shouldBroadcastAILive = true
		}
		shouldCleanup = len(room.Clients) == 0
		room.mutex.Unlock()
		if shouldCleanup {
			delete(h.rooms, client.MeetingID)
		}
	}
	h.mutex.Unlock()

	// 关闭发送通道
	close(client.Send)
	close(client.PrioritySend)

	// 更新会话状态
	if err := h.signalingService.DisconnectSession(client.ID); err != nil {
		logger.Error("Failed to disconnect signaling session", logger.Err(err))
	}

	// 通知其他用户有用户离开
	h.broadcastUserLeft(client)
	if shouldBroadcastAILive && !shouldCleanup {
		h.broadcastAILiveStatus(client.MeetingID)
	}

	logger.Info(fmt.Sprintf("WebSocket client disconnected: %s", client.ID))
}

// broadcastUserJoined 广播用户加入消息
func (h *WebSocketHandler) broadcastUserJoined(client *Client) {
	username := client.Username
	if username == "" {
		if userInfo, err := h.signalingService.GetUserInfo(client.UserID); err != nil {
			logger.Error("Failed to get user info", logger.Err(err))
			username = fmt.Sprintf("user_%d", client.UserID)
		} else {
			username = userInfo.Username
			client.Username = username
		}
	}

	notification := models.UserJoinedNotification{
		UserID:    client.UserID,
		Username:  username,
		PeerID:    client.PeerID,
		MeetingID: client.MeetingID,
	}

	message := &models.WebSocketMessage{
		ID:         fmt.Sprintf("join_%d_%d", client.UserID, time.Now().Unix()),
		Type:       models.MessageTypeUserJoined,
		FromUserID: client.UserID,
		MeetingID:  client.MeetingID,
		SessionID:  client.ID,
		PeerID:     client.PeerID,
		Payload:    notification,
		Timestamp:  time.Now(),
	}

	h.broadcastToRoom(client.MeetingID, message, client.ID)
	h.broadcastRoomInfo(client.MeetingID)
}

// broadcastUserLeft 广播用户离开消息
func (h *WebSocketHandler) broadcastUserLeft(client *Client) {
	username := client.Username
	if username == "" {
		if userInfo, err := h.signalingService.GetUserInfo(client.UserID); err != nil {
			logger.Error("Failed to get user info", logger.Err(err))
			username = fmt.Sprintf("user_%d", client.UserID)
		} else {
			username = userInfo.Username
			client.Username = username
		}
	}

	notification := models.UserLeftNotification{
		UserID:    client.UserID,
		Username:  username,
		PeerID:    client.PeerID,
		MeetingID: client.MeetingID,
	}

	message := &models.WebSocketMessage{
		ID:         fmt.Sprintf("leave_%d_%d", client.UserID, time.Now().Unix()),
		Type:       models.MessageTypeUserLeft,
		FromUserID: client.UserID,
		MeetingID:  client.MeetingID,
		SessionID:  client.ID,
		PeerID:     client.PeerID,
		Payload:    notification,
		Timestamp:  time.Now(),
	}

	h.broadcastToRoom(client.MeetingID, message, client.ID)
	h.broadcastRoomInfo(client.MeetingID)
}

// broadcastToRoom 向房间广播消息
func (h *WebSocketHandler) broadcastToRoom(meetingID uint, message *models.WebSocketMessage, excludeSessionID string) {
	h.mutex.RLock()
	room, exists := h.rooms[meetingID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal message", logger.Err(err))
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for sessionID, client := range room.Clients {
		if sessionID != excludeSessionID {
			if !client.enqueue(data, fmt.Sprintf("broadcast:%d", message.Type)) {
				logger.Warn("Broadcast send failed; unregistering slow client",
					logger.String("target_session", client.ID),
					logger.Uint("target_user", client.UserID),
					logger.Uint("meeting_id", meetingID),
					logger.Int("message_type", int(message.Type)),
				)
				go h.unregisterClient(client)
			}
		}
	}
}

func (h *WebSocketHandler) broadcastRoomInfo(meetingID uint) {
	h.mutex.RLock()
	room, exists := h.rooms[meetingID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	room.mutex.RLock()
	clients := make([]*Client, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, client)
	}
	room.mutex.RUnlock()

	for _, client := range clients {
		client.sendRoomInfo()
	}
}

// sendToClient 向特定客户端发送消息
func (h *WebSocketHandler) sendToClient(sessionID string, message *models.WebSocketMessage) {
	h.mutex.RLock()
	client, exists := h.clients[sessionID]
	h.mutex.RUnlock()

	if !exists {
		logger.Warn(fmt.Sprintf("Client not found: %s", sessionID))
		return
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal message", logger.Err(err))
		return
	}

	if !client.enqueue(data, fmt.Sprintf("direct:%d", message.Type)) {
		logger.Warn("Direct send failed; unregistering slow client",
			logger.String("target_session", client.ID),
			logger.Uint("target_user", client.UserID),
			logger.Uint("meeting_id", client.MeetingID),
			logger.Int("message_type", int(message.Type)),
		)
		go h.unregisterClient(client)
	}
}

// startHeartbeat 启动心跳检查
func (h *WebSocketHandler) startHeartbeat() {
	cfg := config.GlobalConfig
	h.pingTicker = time.NewTicker(time.Duration(cfg.WebSocket.PingPeriod) * time.Second)

	go func() {
		for range h.pingTicker.C {
			h.checkHeartbeat()
		}
	}()
}

// checkHeartbeat 检查心跳
func (h *WebSocketHandler) checkHeartbeat() {
	cfg := config.GlobalConfig
	timeout := time.Duration(cfg.WebSocket.PongWait) * time.Second

	h.mutex.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	h.mutex.RUnlock()

	for _, client := range clients {
		if time.Since(client.LastPing) > timeout {
			logger.Warn(fmt.Sprintf("Client heartbeat timeout: %s", client.ID))
			go h.unregisterClient(client)
		}
	}
}

// startRoomCleanup 启动房间清理
func (h *WebSocketHandler) startRoomCleanup() {
	cfg := config.GlobalConfig
	ticker := time.NewTicker(time.Duration(cfg.Signaling.Room.CleanupInterval) * time.Second)

	go func() {
		for range ticker.C {
			h.cleanupRooms()
		}
	}()
}

// cleanupRooms 清理非活跃房间
func (h *WebSocketHandler) cleanupRooms() {
	cfg := config.GlobalConfig
	timeout := time.Duration(cfg.Signaling.Room.InactiveTimeout) * time.Second

	h.mutex.Lock()
	defer h.mutex.Unlock()

	for meetingID, room := range h.rooms {
		room.mutex.RLock()
		inactive := time.Since(room.LastActivity) > timeout
		clientCount := len(room.Clients)
		room.mutex.RUnlock()

		if inactive && clientCount == 0 {
			delete(h.rooms, meetingID)
			logger.Info(fmt.Sprintf("Cleaned up inactive room: %d", meetingID))
		}
	}
}

// GetRoomStats 获取房间统计信息
func (h *WebSocketHandler) GetRoomStats() map[uint]int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	stats := make(map[uint]int)
	for meetingID, room := range h.rooms {
		room.mutex.RLock()
		uniqueUsers := make(map[uint]struct{}, len(room.Clients))
		for _, client := range room.Clients {
			uniqueUsers[client.UserID] = struct{}{}
		}
		stats[meetingID] = len(uniqueUsers)
		room.mutex.RUnlock()
	}

	return stats
}

func (h *WebSocketHandler) collectRoomParticipants(meetingID uint) []models.RoomParticipant {
	h.mutex.RLock()
	room, exists := h.rooms[meetingID]
	h.mutex.RUnlock()
	if !exists {
		return []models.RoomParticipant{}
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	participants := make([]models.RoomParticipant, 0, len(room.Clients))
	for _, client := range room.Clients {
		participants = append(participants, models.RoomParticipant{
			UserID:       client.UserID,
			Username:     client.Username,
			SessionID:    client.ID,
			PeerID:       client.PeerID,
			JoinedAt:     client.JoinedAt,
			LastActiveAt: client.LastPing,
		})
	}

	return participants
}

// GetClientCount 获取客户端总数
func (h *WebSocketHandler) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// Stop 停止WebSocket处理器
func (h *WebSocketHandler) Stop() {
	if h.pingTicker != nil {
		h.pingTicker.Stop()
	}

	// 关闭所有客户端连接
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, client := range h.clients {
		client.Conn.Close()
	}
}

// readPump 读取消息泵
func (c *Client) readPump() {
	defer func() {
		c.Handler.unregisterClient(c)
		c.Conn.Close()
	}()

	cfg := config.GlobalConfig
	c.Conn.SetReadLimit(int64(cfg.WebSocket.MaxMessageSize * 1024))
	c.Conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.WebSocket.PongWait) * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.LastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.WebSocket.PongWait) * time.Second))
		return nil
	})

	for {
		msgType, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			if closeErr, ok := err.(*websocket.CloseError); ok {
				logger.Info("WebSocket closed",
					logger.String("session", c.ID),
					logger.Uint("user_id", c.UserID),
					logger.Int("code", closeErr.Code),
					logger.String("text", closeErr.Text),
				)
			} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("Unexpected WebSocket close", logger.Err(err),
					logger.String("session", c.ID),
					logger.Uint("user_id", c.UserID),
				)
			} else {
				logger.Warn("WebSocket read ended", logger.Err(err),
					logger.String("session", c.ID),
					logger.Uint("user_id", c.UserID),
				)
			}
			break
		}

		// 只处理文本消息
		if msgType != websocket.TextMessage {
			continue
		}

		// 解析消息
		var message models.WebSocketMessage
		if err := json.Unmarshal(messageData, &message); err != nil {
			preview := string(messageData)
			if len(preview) > 256 {
				preview = preview[:256] + "..."
			}
			logger.Error("Failed to unmarshal WebSocket message",
				logger.Err(err),
				logger.String("session", c.ID),
				logger.Uint("user_id", c.UserID),
				logger.String("payload_preview", preview),
			)
			c.sendError("Invalid message format", err.Error())
			continue
		}

		// 设置消息元数据
		message.FromUserID = c.UserID
		message.MeetingID = c.MeetingID
		message.SessionID = c.ID
		message.Timestamp = time.Now()

		// 处理消息
		c.handleMessage(&message)
	}
}

// writeBytes 写入文本消息
func (c *Client) writeBytes(message []byte) error {
	cfg := config.GlobalConfig
	c.Conn.SetWriteDeadline(time.Now().Add(time.Duration(cfg.WebSocket.WriteWait) * time.Second))
	if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
		logger.Error("Failed to write WebSocket message", logger.Err(err),
			logger.String("session", c.ID),
			logger.Uint("user_id", c.UserID),
		)
		return err
	}
	return nil
}

func (c *Client) writeCloseFrame() {
	cfg := config.GlobalConfig
	c.Conn.SetWriteDeadline(time.Now().Add(time.Duration(cfg.WebSocket.WriteWait) * time.Second))
	if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
		logger.Error("Failed to write WebSocket close message", logger.Err(err),
			logger.String("session", c.ID),
			logger.Uint("user_id", c.UserID),
		)
	}
}

func (c *Client) enqueue(data []byte, context string) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Attempted to send on closed channel",
				logger.String("session", c.ID),
				logger.Uint("user_id", c.UserID),
				logger.String("context", context),
				logger.Any("panic", r))
			ok = false
		}
	}()

	select {
	case c.Send <- data:
		return true
	default:
		logger.Warn("Send buffer full",
			logger.String("session", c.ID),
			logger.Uint("user_id", c.UserID),
			logger.String("context", context),
		)
		return false
	}
}

func (c *Client) enqueuePriority(data []byte, context string) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Attempted to send on closed priority channel",
				logger.String("session", c.ID),
				logger.Uint("user_id", c.UserID),
				logger.String("context", context),
				logger.Any("panic", r))
			ok = false
		}
	}()

	timeout := time.Duration(config.GlobalConfig.WebSocket.WriteWait) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case c.PrioritySend <- data:
		return true
	case <-timer.C:
		logger.Error("Priority send timed out",
			logger.String("session", c.ID),
			logger.Uint("user_id", c.UserID),
			logger.String("context", context),
		)
		return false
	}
}

// writePump 写入消息泵
func (c *Client) writePump() {
	cfg := config.GlobalConfig
	ticker := time.NewTicker(time.Duration(cfg.WebSocket.PingPeriod) * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		// 优先发送高优先级消息，确保关键响应（如join ack）不被普通广播挤掉
		priorityDrained := false
		for !priorityDrained {
			select {
			case message, ok := <-c.PrioritySend:
				if !ok {
					c.writeCloseFrame()
					return
				}

				if err := c.writeBytes(message); err != nil {
					return
				}
				// 继续尝试提取剩余的优先消息
				continue
			default:
				priorityDrained = true
			}
		}

		select {
		case message, ok := <-c.PrioritySend:
			if !ok {
				c.writeCloseFrame()
				return
			}
			if err := c.writeBytes(message); err != nil {
				return
			}
		case message, ok := <-c.Send:
			if !ok {
				c.writeCloseFrame()
				return
			}
			if err := c.writeBytes(message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(time.Duration(cfg.WebSocket.WriteWait) * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("Failed to write WebSocket ping message", logger.Err(err),
					logger.String("session", c.ID),
					logger.Uint("user_id", c.UserID),
				)
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(message *models.WebSocketMessage) {
	logger.Info("Received message", logger.Uint("meeting_id", message.MeetingID), logger.Uint("from", message.FromUserID), logger.Int("type", int(message.Type)))
	switch message.Type {
	case models.MessageTypeOffer:
		c.handleOffer(message)
	case models.MessageTypeAnswer:
		c.handleAnswer(message)
	case models.MessageTypeICECandidate:
		c.handleICECandidate(message)
	case models.MessageTypeJoinRoom:
		c.handleJoinRoom(message)
	case models.MessageTypeLeaveRoom:
		c.handleLeaveRoom(message)
	case models.MessageTypeChat:
		c.handleChat(message)
	case models.MessageTypeMediaControl:
		c.handleMediaControl(message)
	case models.MessageTypePing:
		c.handlePing(message)
	case models.MessageTypeAILiveClaim:
		c.handleAILiveClaim(message)
	case models.MessageTypeAILiveResult:
		c.handleAILiveResult(message)
	default:
		logger.Warn(fmt.Sprintf("Unknown message type: %d", message.Type))
		c.sendError("Unknown message type", fmt.Sprintf("Type: %d", message.Type))
	}
}

// handleOffer 处理WebRTC Offer
func (c *Client) handleOffer(message *models.WebSocketMessage) {
	// 验证目标用户
	if message.ToUserID == nil {
		c.sendError("Missing target user", "ToUserID is required for offer")
		return
	}

	// 转发给目标用户
	c.Handler.forwardToUser(*message.ToUserID, message)

	// 记录消息
	if err := c.Handler.signalingService.SaveMessage(message); err != nil {
		logger.Error("Failed to save offer message", logger.Err(err))
	}
}

// handleAnswer 处理WebRTC Answer
func (c *Client) handleAnswer(message *models.WebSocketMessage) {
	// 验证目标用户
	if message.ToUserID == nil {
		c.sendError("Missing target user", "ToUserID is required for answer")
		return
	}

	// 转发给目标用户
	c.Handler.forwardToUser(*message.ToUserID, message)

	// 记录消息
	if err := c.Handler.signalingService.SaveMessage(message); err != nil {
		logger.Error("Failed to save answer message", logger.Err(err))
	}
}

// handleICECandidate 处理ICE候选
func (c *Client) handleICECandidate(message *models.WebSocketMessage) {
	// 验证目标用户
	if message.ToUserID == nil {
		c.sendError("Missing target user", "ToUserID is required for ICE candidate")
		return
	}

	// 转发给目标用户
	c.Handler.forwardToUser(*message.ToUserID, message)
}

// handleJoinRoom 处理加入房间
func (c *Client) handleJoinRoom(message *models.WebSocketMessage) {
	// 更新会话状态
	if err := c.Handler.signalingService.UpdateSessionStatus(c.ID, models.SessionStatusConnected); err != nil {
		logger.Error("Failed to update session status", logger.Err(err))
	}

	// 发送房间信息
	logger.Info(fmt.Sprintf("Sending room info to session %s", c.ID))
	c.sendRoomInfo()

	// 通知其他用户新成员加入
	c.Handler.broadcastUserJoined(c)
}

// handleLeaveRoom 处理离开房间
func (c *Client) handleLeaveRoom(message *models.WebSocketMessage) {
	// 注销客户端
	go c.Handler.unregisterClient(c)
}

// handleChat 处理聊天消息
func (c *Client) handleChat(message *models.WebSocketMessage) {
	// 广播聊天消息到房间
	c.Handler.broadcastToRoom(c.MeetingID, message, "")

	// 记录消息
	if err := c.Handler.signalingService.SaveMessage(message); err != nil {
		logger.Error("Failed to save chat message", logger.Err(err))
	}
}

// handleMediaControl 处理媒体控制
func (c *Client) handleMediaControl(message *models.WebSocketMessage) {
	// 广播媒体控制消息到房间
	c.Handler.broadcastToRoom(c.MeetingID, message, c.ID)
}

// handlePing 处理心跳
func (c *Client) handlePing(message *models.WebSocketMessage) {
	c.LastPing = time.Now()

	// 发送Pong响应
	pongMessage := &models.WebSocketMessage{
		ID:         fmt.Sprintf("pong_%d", time.Now().Unix()),
		Type:       models.MessageTypePong,
		FromUserID: c.UserID,
		MeetingID:  c.MeetingID,
		SessionID:  c.ID,
		Timestamp:  time.Now(),
	}

	data, _ := json.Marshal(pongMessage)
	if !c.enqueuePriority(data, "pong") {
		c.enqueue(data, "pong_fallback")
	}
}

func (c *Client) handleAILiveClaim(message *models.WebSocketMessage) {
	enable := true
	if payload, ok := message.Payload.(map[string]interface{}); ok {
		if v, ok := payload["enable"].(bool); ok {
			enable = v
		}
	}

	h := c.Handler
	h.mutex.RLock()
	room := h.rooms[c.MeetingID]
	h.mutex.RUnlock()
	if room == nil {
		c.sendError("Room not found", "meeting room not available")
		return
	}

	now := time.Now()
	room.mutex.Lock()
	if enable {
		// 申请成为领导者：只有在无人占用或自己已是领导者时允许
		if room.AILive.LeaderSessionID == "" || room.AILive.LeaderSessionID == c.ID {
			room.AILive.Enabled = true
			room.AILive.LeaderUserID = c.UserID
			room.AILive.LeaderSessionID = c.ID
			room.AILive.LeaderUsername = c.Username
			room.AILive.UpdatedAt = now
		}
	} else {
		// 释放：仅领导者有效
		if room.AILive.LeaderSessionID == c.ID {
			room.AILive = models.AILiveStatusMessage{
				Enabled:   false,
				UpdatedAt: now,
			}
		}
	}
	room.mutex.Unlock()

	h.broadcastAILiveStatus(c.MeetingID)
}

func (c *Client) handleAILiveResult(message *models.WebSocketMessage) {
	h := c.Handler
	h.mutex.RLock()
	room := h.rooms[c.MeetingID]
	h.mutex.RUnlock()
	if room == nil {
		return
	}

	room.mutex.RLock()
	leaderSession := room.AILive.LeaderSessionID
	room.mutex.RUnlock()
	if leaderSession != c.ID {
		c.sendError("AI Live denied", "only AI Live leader can broadcast results")
		return
	}

	h.broadcastToRoom(c.MeetingID, message, "")
}

func (h *WebSocketHandler) broadcastAILiveStatus(meetingID uint) {
	status := h.getAILiveStatus(meetingID)
	if status == nil {
		return
	}

	msg := &models.WebSocketMessage{
		ID:         fmt.Sprintf("ai_live_status_%d", time.Now().UnixNano()),
		Type:       models.MessageTypeAILiveStatus,
		FromUserID: 0,
		MeetingID:  meetingID,
		Payload:    status,
		Timestamp:  time.Now(),
	}

	h.broadcastToRoom(meetingID, msg, "")
}

func (h *WebSocketHandler) getAILiveStatus(meetingID uint) *models.AILiveStatusMessage {
	h.mutex.RLock()
	room := h.rooms[meetingID]
	h.mutex.RUnlock()
	if room == nil {
		return nil
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	status := room.AILive
	if status.UpdatedAt.IsZero() {
		status.UpdatedAt = time.Now()
	}
	return &status
}

// sendError 发送错误消息
func (c *Client) sendError(message, details string) {
	errorMsg := &models.WebSocketMessage{
		ID:         fmt.Sprintf("error_%d", time.Now().Unix()),
		Type:       models.MessageTypeError,
		FromUserID: 0, // 系统消息
		MeetingID:  c.MeetingID,
		SessionID:  c.ID,
		Payload: models.ErrorMessage{
			Code:    400,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(errorMsg)
	if !c.enqueuePriority(data, "error") {
		c.enqueue(data, "error_fallback")
	}
}

// sendRoomInfo 发送房间信息
func (c *Client) sendRoomInfo() {
	participants := c.Handler.collectRoomParticipants(c.MeetingID)
	participantSnapshot := make([]models.RoomParticipant, 0, len(participants))
	uniqueUsers := make(map[uint]struct{}, len(participants))
	for _, participant := range participants {
		if participant.UserID != 0 {
			uniqueUsers[participant.UserID] = struct{}{}
		}
		participant.IsSelf = participant.SessionID == c.ID
		participantSnapshot = append(participantSnapshot, participant)
	}

	participantCount := len(uniqueUsers)

	iceServers := make([]models.RoomICEServer, 0, len(config.GlobalConfig.Signaling.ICEServers))
	for _, srv := range config.GlobalConfig.Signaling.ICEServers {
		iceServers = append(iceServers, models.RoomICEServer{
			URLs:       srv.URLs,
			Username:   srv.Username,
			Credential: srv.Credential,
		})
	}

	roomInfo := models.RoomInfoMessage{
		MeetingID:        c.MeetingID,
		ParticipantCount: participantCount,
		SessionID:        c.ID,
		PeerID:           c.PeerID,
		IceServers:       iceServers,
		Participants:     participantSnapshot,
		AILive:           c.Handler.getAILiveStatus(c.MeetingID),
	}

	logger.Debug("Room info payload", logger.Uint("meeting_id", c.MeetingID), logger.Int("participants", participantCount))

	message := &models.WebSocketMessage{
		ID:         fmt.Sprintf("room_info_%d", time.Now().UnixNano()),
		Type:       models.MessageTypeRoomInfo,
		FromUserID: 0, // 系统消息
		MeetingID:  c.MeetingID,
		SessionID:  c.ID,
		PeerID:     c.PeerID,
		Payload:    roomInfo,
		Timestamp:  time.Now(),
	}

	data, _ := json.Marshal(message)
	if c.enqueuePriority(data, "room_info") {
		logger.Debug("Enqueued room info ack",
			logger.String("session", c.ID),
			logger.Uint("user_id", c.UserID),
			logger.Uint("meeting_id", c.MeetingID),
		)
		return
	}

	logger.Error("Priority channel saturated for room info",
		logger.String("session", c.ID),
		logger.Uint("user_id", c.UserID),
		logger.Uint("meeting_id", c.MeetingID),
	)
	if !c.enqueue(data, "room_info_fallback") {
		logger.Error("Failed to enqueue room info on fallback channel",
			logger.String("session", c.ID),
			logger.Uint("user_id", c.UserID),
			logger.Uint("meeting_id", c.MeetingID),
		)
	}
}

// forwardToUser 转发消息给指定用户
func (h *WebSocketHandler) forwardToUser(userID uint, message *models.WebSocketMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 查找目标用户的所有会话
	for _, client := range h.clients {
		if client.UserID == userID && client.MeetingID == message.MeetingID {
			data, err := json.Marshal(message)
			if err != nil {
				logger.Error("Failed to marshal message for forwarding", logger.Err(err))
				continue
			}

			if !client.enqueue(data, fmt.Sprintf("forward:%d", message.Type)) {
				logger.Warn("Forward send failed; unregistering slow client",
					logger.String("target_session", client.ID),
					logger.Uint("target_user", client.UserID),
					logger.Uint("meeting_id", client.MeetingID),
					logger.Int("message_type", int(message.Type)),
				)
				go h.unregisterClient(client)
			}
		}
	}
}
