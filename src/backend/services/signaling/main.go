package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"video-conference-system/shared/auth"
	"video-conference-system/shared/config"
	"video-conference-system/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// SignalingMessage 信令消息结构
type SignalingMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	MeetingID string                 `json:"meeting_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// Client WebSocket客户端
type Client struct {
	ID        string
	UserID    string
	Username  string
	MeetingID string
	Conn      *websocket.Conn
	Send      chan SignalingMessage
	Hub       *Hub
}

// Hub 管理所有客户端连接
type Hub struct {
	// 注册的客户端
	clients map[*Client]bool

	// 按会议ID分组的客户端
	meetings map[string]map[*Client]bool

	// 按用户ID索引的客户端
	users map[string]*Client

	// 注册请求
	register chan *Client

	// 注销请求
	unregister chan *Client

	// 广播消息
	broadcast chan SignalingMessage

	// 互斥锁
	mutex sync.RWMutex
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		meetings:   make(map[string]map[*Client]bool),
		users:      make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan SignalingMessage),
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	h.users[client.UserID] = client

	// 添加到会议组
	if h.meetings[client.MeetingID] == nil {
		h.meetings[client.MeetingID] = make(map[*Client]bool)
	}
	h.meetings[client.MeetingID][client] = true

	log.Printf("Client %s joined meeting %s", client.UserID, client.MeetingID)

	// 通知其他参与者
	h.notifyParticipantJoined(client)
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		delete(h.users, client.UserID)

		// 从会议组中移除
		if meetingClients, exists := h.meetings[client.MeetingID]; exists {
			delete(meetingClients, client)
			if len(meetingClients) == 0 {
				delete(h.meetings, client.MeetingID)
			}
		}

		close(client.Send)
		log.Printf("Client %s left meeting %s", client.UserID, client.MeetingID)

		// 通知其他参与者
		h.notifyParticipantLeft(client)
	}
}

// broadcastMessage 广播消息
func (h *Hub) broadcastMessage(message SignalingMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if message.To != "" {
		// 点对点消息
		if targetClient, exists := h.users[message.To]; exists {
			select {
			case targetClient.Send <- message:
			default:
				close(targetClient.Send)
				delete(h.clients, targetClient)
				delete(h.users, targetClient.UserID)
			}
		}
	} else {
		// 广播到会议中的所有客户端
		if meetingClients, exists := h.meetings[message.MeetingID]; exists {
			for client := range meetingClients {
				if client.UserID != message.From {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, client)
						delete(h.users, client.UserID)
						delete(meetingClients, client)
					}
				}
			}
		}
	}
}

// notifyParticipantJoined 通知参与者加入
func (h *Hub) notifyParticipantJoined(client *Client) {
	message := SignalingMessage{
		Type:      "participant-joined",
		From:      client.UserID,
		MeetingID: client.MeetingID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_id":  client.UserID,
			"username": client.Username,
		},
	}

	h.broadcast <- message
}

// notifyParticipantLeft 通知参与者离开
func (h *Hub) notifyParticipantLeft(client *Client) {
	message := SignalingMessage{
		Type:      "participant-left",
		From:      client.UserID,
		MeetingID: client.MeetingID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"user_id": client.UserID,
		},
	}

	h.broadcast <- message
}

// GetMeetingParticipants 获取会议参与者
func (h *Hub) GetMeetingParticipants(meetingID string) []map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var participants []map[string]interface{}
	if meetingClients, exists := h.meetings[meetingID]; exists {
		for client := range meetingClients {
			participants = append(participants, map[string]interface{}{
				"user_id":  client.UserID,
				"username": client.Username,
			})
		}
	}

	return participants
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 连接Redis
	redis, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// 创建JWT管理器
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireTime)

	// 创建Hub
	hub := NewHub()
	go hub.Run()

	// 设置路由
	router := setupRoutes(hub, jwtManager)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("Signaling service starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(hub *Hub, jwtManager *auth.JWTManager) *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "signaling-service",
			"timestamp": time.Now(),
		})
	})

	// WebSocket端点
	router.GET("/signaling/:meeting_id", func(c *gin.Context) {
		handleWebSocket(hub, jwtManager, c)
	})

	// REST API
	api := router.Group("/api/v1")
	{
		api.GET("/meetings/:meeting_id/participants", func(c *gin.Context) {
			meetingID := c.Param("meeting_id")
			participants := hub.GetMeetingParticipants(meetingID)
			c.JSON(http.StatusOK, gin.H{
				"participants": participants,
			})
		})
	}

	return router
}

func handleWebSocket(hub *Hub, jwtManager *auth.JWTManager, c *gin.Context) {
	meetingID := c.Param("meeting_id")
	token := c.Query("token")

	// 验证JWT令牌
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// 升级到WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	// 创建客户端
	client := &Client{
		ID:        uuid.New().String(),
		UserID:    claims.UserID,
		Username:  claims.Username,
		MeetingID: meetingID,
		Conn:      conn,
		Send:      make(chan SignalingMessage, 256),
		Hub:       hub,
	}

	// 注册客户端
	client.Hub.register <- client

	// 启动goroutines
	go client.writePump()
	go client.readPump()
}

// readPump 读取WebSocket消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var message SignalingMessage
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		message.From = c.UserID
		message.MeetingID = c.MeetingID
		message.Timestamp = time.Now()

		c.Hub.broadcast <- message
	}
}

// writePump 写入WebSocket消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				log.Printf("Failed to write message: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
