package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// User 用户结构
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// Meeting 会议结构
type Meeting struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	Status      string    `json:"status"`
	Creator     User      `json:"creator"`
}

// DetectionResult 检测结果
type DetectionResult struct {
	ID         string  `json:"id"`
	IsFake     bool    `json:"is_fake"`
	Confidence float64 `json:"confidence"`
	Type       string  `json:"detection_type"`
	Timestamp  time.Time `json:"timestamp"`
}

// 模拟数据
var userList = []User{
	{ID: "1", Username: "admin", Email: "admin@example.com", FullName: "系统管理员"},
	{ID: "2", Username: "user1", Email: "user1@example.com", FullName: "用户一"},
	{ID: "3", Username: "user2", Email: "user2@example.com", FullName: "用户二"},
}

var meetings = []Meeting{
	{
		ID:          "1",
		Title:       "项目讨论会议",
		Description: "讨论项目进展和下一步计划",
		StartTime:   time.Now().Add(time.Hour),
		Status:      "scheduled",
		Creator:     userList[0],
	},
	{
		ID:          "2",
		Title:       "技术分享会",
		Description: "分享最新的技术趋势",
		StartTime:   time.Now().Add(2 * time.Hour),
		Status:      "scheduled",
		Creator:     userList[1],
	},
}

var detectionResults = []DetectionResult{
	{
		ID:         "1",
		IsFake:     false,
		Confidence: 0.95,
		Type:       "face_swap",
		Timestamp:  time.Now().Add(-time.Hour),
	},
	{
		ID:         "2",
		IsFake:     true,
		Confidence: 0.87,
		Type:       "voice_synthesis",
		Timestamp:  time.Now().Add(-30 * time.Minute),
	},
}

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket客户端
type WebSocketClient struct {
	ID       string
	UserID   string
	Username string
	MeetingID string
	Conn     *websocket.Conn
	Send     chan map[string]interface{}
	Hub      *SignalingHub
}

// 信令Hub
type SignalingHub struct {
	clients     map[*WebSocketClient]bool
	meetings    map[string]map[*WebSocketClient]bool
	presenters  map[string]*WebSocketClient // 每个会议的主讲人
	register    chan *WebSocketClient
	unregister  chan *WebSocketClient
	broadcast   chan map[string]interface{}
	chatManager *ChatManager // 聊天管理器
}

// 全局信令Hub
var signalingHub = &SignalingHub{
	clients:    make(map[*WebSocketClient]bool),
	meetings:   make(map[string]map[*WebSocketClient]bool),
	presenters: make(map[string]*WebSocketClient),
	register:   make(chan *WebSocketClient),
	unregister: make(chan *WebSocketClient),
	broadcast:  make(chan map[string]interface{}),
}

// 初始化聊天管理器
func initChatManager() {
	signalingHub.chatManager = NewChatManager(signalingHub)
}

// 启动信令Hub
func init() {
	initChatManager()
	go signalingHub.run()
}

// 运行信令Hub
func (h *SignalingHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if h.meetings[client.MeetingID] == nil {
				h.meetings[client.MeetingID] = make(map[*WebSocketClient]bool)
			}

			// 先向新用户发送现有用户列表
			h.sendExistingUsersToNewUser(client)

			// 然后将新用户添加到会议
			h.meetings[client.MeetingID][client] = true

			log.Printf("用户 %s 加入会议 %s，当前会议人数: %d", client.Username, client.MeetingID, len(h.meetings[client.MeetingID]))

			// 通知其他用户有新用户加入
			h.notifyUserJoined(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if meetingClients, exists := h.meetings[client.MeetingID]; exists {
					delete(meetingClients, client)
					if len(meetingClients) == 0 {
						delete(h.meetings, client.MeetingID)
						// 清理主讲人信息
						delete(h.presenters, client.MeetingID)
					}
				}

				// 如果离开的用户是主讲人，清理主讲人状态
				h.removePresenter(client.MeetingID, client)

				close(client.Send)

				log.Printf("用户 %s 离开会议 %s", client.Username, client.MeetingID)

				// 通知其他用户
				h.notifyUserLeft(client)
			}

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// 向新用户发送现有用户列表
func (h *SignalingHub) sendExistingUsersToNewUser(newClient *WebSocketClient) {
	if meetingClients, exists := h.meetings[newClient.MeetingID]; exists {
		log.Printf("向新用户 %s 发送现有用户列表，现有用户数: %d", newClient.Username, len(meetingClients))

		for existingClient := range meetingClients {
			if existingClient.UserID != newClient.UserID {
				// 向新用户发送现有用户信息
				userJoinedMessage := map[string]interface{}{
					"type": "user-joined",
					"data": map[string]interface{}{
						"id":   existingClient.UserID,
						"name": existingClient.Username,
					},
				}

				if newClient.safeSend(userJoinedMessage) {
					log.Printf("向新用户 %s 发送现有用户 %s 的信息", newClient.Username, existingClient.Username)
				} else {
					log.Printf("向新用户发送现有用户信息失败: %s -> %s", existingClient.Username, newClient.Username)
				}
			}
		}
	} else {
		log.Printf("会议 %s 不存在或为空，新用户 %s 是第一个加入的用户", newClient.MeetingID, newClient.Username)
	}
}

// 通知用户加入
func (h *SignalingHub) notifyUserJoined(client *WebSocketClient) {
	message := map[string]interface{}{
		"type": "user-joined",
		"data": map[string]interface{}{
			"id":   client.UserID,
			"name": client.Username,
		},
	}

	log.Printf("通知其他用户 %s 加入会议 %s", client.Username, client.MeetingID)
	h.broadcastToMeeting(client.MeetingID, message, client.UserID)
}

// 通知用户离开
func (h *SignalingHub) notifyUserLeft(client *WebSocketClient) {
	message := map[string]interface{}{
		"type": "user-left",
		"data": map[string]interface{}{
			"id":       client.UserID,
			"name":     client.Username,
			"meeting_id": client.MeetingID,
		},
	}

	h.broadcastToMeeting(client.MeetingID, message, client.UserID)
}

// 设置主讲人
func (h *SignalingHub) setPresenter(meetingID string, client *WebSocketClient) {
	// 检查是否已有主讲人
	if currentPresenter, exists := h.presenters[meetingID]; exists && currentPresenter != nil {
		// 通知当前主讲人失去主讲权限
		presenterRemovedMessage := map[string]interface{}{
			"type": "presenter-removed",
			"data": map[string]interface{}{
				"meeting_id": meetingID,
				"message":    "您已不再是主讲人",
			},
		}

		if currentPresenter.safeSend(presenterRemovedMessage) {
			log.Printf("通知用户 %s 失去主讲权限", currentPresenter.Username)
		} else {
			log.Printf("通知用户失去主讲权限失败: %s", currentPresenter.Username)
		}
	}

	// 设置新主讲人
	h.presenters[meetingID] = client
	log.Printf("用户 %s 成为会议 %s 的主讲人", client.Username, meetingID)

	// 通知新主讲人获得权限
	presenterSetMessage := map[string]interface{}{
		"type": "presenter-set",
		"data": map[string]interface{}{
			"meeting_id": meetingID,
			"message":    "您现在是主讲人",
		},
	}

	if client.safeSend(presenterSetMessage) {
		log.Printf("通知用户 %s 获得主讲权限", client.Username)
	} else {
		log.Printf("通知用户获得主讲权限失败: %s", client.Username)
	}

	// 通知会议中的其他用户主讲人变更
	presenterChangedMessage := map[string]interface{}{
		"type": "presenter-changed",
		"data": map[string]interface{}{
			"meeting_id":    meetingID,
			"presenter_id":  client.UserID,
			"presenter_name": client.Username,
		},
	}

	h.broadcastToMeeting(meetingID, presenterChangedMessage, client.UserID)
}

// 移除主讲人
func (h *SignalingHub) removePresenter(meetingID string, client *WebSocketClient) {
	if currentPresenter, exists := h.presenters[meetingID]; exists && currentPresenter == client {
		delete(h.presenters, meetingID)
		log.Printf("移除会议 %s 的主讲人 %s", meetingID, client.Username)

		// 通知会议中的所有用户主讲人已移除
		presenterRemovedMessage := map[string]interface{}{
			"type": "presenter-removed",
			"data": map[string]interface{}{
				"meeting_id": meetingID,
				"message":    "主讲人已离开",
			},
		}

		h.broadcastToMeeting(meetingID, presenterRemovedMessage, "")
	}
}

// 广播消息
func (h *SignalingHub) broadcastMessage(message map[string]interface{}) {
	messageType, _ := message["type"].(string)
	to, hasTo := message["to"].(string)
	from, _ := message["from"].(string)

	if hasTo && to != "" {
		// 点对点消息
		for client := range h.clients {
			if client.UserID == to {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
				break
			}
		}
	} else {
		// 广播消息
		meetingID, _ := message["meeting_id"].(string)
		if meetingID != "" {
			h.broadcastToMeeting(meetingID, message, from)
		}
	}

	log.Printf("广播消息: %s", messageType)
}

// 广播到会议
func (h *SignalingHub) broadcastToMeeting(meetingID string, message map[string]interface{}, excludeUserID string) {
	if meetingClients, exists := h.meetings[meetingID]; exists {
		for client := range meetingClients {
			if client.UserID != excludeUserID {
				if !client.safeSend(message) {
					log.Printf("向用户 %s 发送消息失败", client.Username)
				}
			}
		}
	}
}

// 检查客户端连接状态
func (c *WebSocketClient) isConnected() bool {
	return c.Conn != nil && c.Send != nil
}

// 安全发送消息
func (c *WebSocketClient) safeSend(message map[string]interface{}) bool {
	if !c.isConnected() {
		return false
	}

	select {
	case c.Send <- message:
		return true
	case <-time.After(100 * time.Millisecond):
		log.Printf("向用户 %s 发送消息超时", c.Username)
		return false
	}
}

func main() {
	// 设置Gin为发布模式
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "video-conference-demo",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	// 主页
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "欢迎使用视频会议系统演示版",
			"features": []string{
				"用户管理",
				"会议管理", 
				"伪造音视频检测",
				"实时通信",
			},
			"endpoints": gin.H{
				"users":     "/api/v1/users",
				"meetings":  "/api/v1/meetings",
				"detection": "/api/v1/detection",
				"signaling": "/signaling",
			},
		})
	})

	// API路由组
	api := router.Group("/api/v1")
	{
		// 用户相关API
		usersAPI := api.Group("/users")
		{
			usersAPI.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"users": userList,
					"total": len(userList),
				})
			})

			usersAPI.POST("/login", func(c *gin.Context) {
				var loginReq struct {
					Email    string `json:"email"`
					Password string `json:"password"`
				}

				if err := c.ShouldBindJSON(&loginReq); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
					return
				}

				// 模拟登录验证
				for _, user := range userList {
					if user.Email == loginReq.Email {
						c.JSON(http.StatusOK, gin.H{
							"message": "登录成功",
							"token":   "demo-jwt-token-" + user.ID,
							"user":    user,
						})
						return
					}
				}

				c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码错误"})
			})

			usersAPI.POST("/register", func(c *gin.Context) {
				var registerReq struct {
					Username string `json:"username"`
					Email    string `json:"email"`
					Password string `json:"password"`
					FullName string `json:"full_name"`
				}

				if err := c.ShouldBindJSON(&registerReq); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
					return
				}

				// 模拟用户注册
				newUser := User{
					ID:       fmt.Sprintf("%d", len(userList)+1),
					Username: registerReq.Username,
					Email:    registerReq.Email,
					FullName: registerReq.FullName,
				}

				userList = append(userList, newUser)

				c.JSON(http.StatusCreated, gin.H{
					"message": "注册成功",
					"user":    newUser,
				})
			})
		}

		// 会议相关API
		meetingsAPI := api.Group("/meetings")
		{
			meetingsAPI.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"meetings": meetings,
					"total":    len(meetings),
				})
			})

			meetingsAPI.POST("", func(c *gin.Context) {
				var meetingReq struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					StartTime   string `json:"start_time"`
				}

				if err := c.ShouldBindJSON(&meetingReq); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
					return
				}

				startTime, _ := time.Parse(time.RFC3339, meetingReq.StartTime)
				newMeeting := Meeting{
					ID:          fmt.Sprintf("%d", len(meetings)+1),
					Title:       meetingReq.Title,
					Description: meetingReq.Description,
					StartTime:   startTime,
					Status:      "scheduled",
					Creator:     userList[0], // 默认创建者
				}

				meetings = append(meetings, newMeeting)

				c.JSON(http.StatusCreated, gin.H{
					"message": "会议创建成功",
					"meeting": newMeeting,
				})
			})

			meetingsAPI.GET("/:id", func(c *gin.Context) {
				id := c.Param("id")
				for _, meeting := range meetings {
					if meeting.ID == id {
						c.JSON(http.StatusOK, gin.H{"meeting": meeting})
						return
					}
				}
				c.JSON(http.StatusNotFound, gin.H{"error": "会议不存在"})
			})
		}

		// 检测相关API
		detection := api.Group("/detection")
		{
			detection.GET("/results", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"results": detectionResults,
					"total":   len(detectionResults),
				})
			})

			detection.POST("/analyze", func(c *gin.Context) {
				// 模拟文件上传和检测
				file, header, err := c.Request.FormFile("file")
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "没有上传文件"})
					return
				}
				defer file.Close()

				// 模拟检测过程
				time.Sleep(2 * time.Second)

				result := DetectionResult{
					ID:         fmt.Sprintf("%d", len(detectionResults)+1),
					IsFake:     false, // 随机结果
					Confidence: 0.92,
					Type:       "face_swap",
					Timestamp:  time.Now(),
				}

				detectionResults = append(detectionResults, result)

				c.JSON(http.StatusOK, gin.H{
					"message":  "检测完成",
					"filename": header.Filename,
					"result":   result,
				})
			})

			// 检测报告API
			detection.POST("/report", func(c *gin.Context) {
				var report map[string]interface{}
				if err := c.ShouldBindJSON(&report); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   "Invalid request format",
					})
					return
				}

				// 记录检测报告
				log.Printf("收到检测报告: %+v", report)

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Detection report received",
					"id":      time.Now().Unix(),
				})
			})
		}

		// 系统统计API
		api.GET("/stats", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"system_stats": gin.H{
					"total_users":       len(userList),
					"total_meetings":    len(meetings),
					"total_detections":  len(detectionResults),
					"fake_detections":   1, // 模拟数据
					"system_uptime":     "2h 30m",
					"active_meetings":   1,
					"online_users":      3,
				},
			})
		})
	}

	// WebSocket信令服务
	router.GET("/signaling", func(c *gin.Context) {
		handleWebSocketSignaling(c)
	})

	// 聊天相关API
	router.GET("/api/v1/chat/:meetingId/history", getChatHistory)
	router.GET("/api/v1/chat/:meetingId/stats", getChatStats)
	router.POST("/api/v1/chat/:meetingId/export", exportChatHistory)
	router.DELETE("/api/v1/chat/:meetingId/clear", clearChatHistory)

	// 启动服务器
	port := "8081"
	fmt.Printf("\n🚀 视频会议系统演示版启动成功!\n")
	fmt.Printf("📍 服务地址: http://localhost:%s\n", port)
	fmt.Printf("📖 API文档: http://localhost:%s/api/v1\n", port)
	fmt.Printf("🔍 健康检查: http://localhost:%s/health\n", port)
	fmt.Printf("💬 WebSocket: ws://localhost:%s/signaling\n\n", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}

// 处理WebSocket信令
func handleWebSocketSignaling(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := &WebSocketClient{
		ID:   uuid.New().String(),
		Conn: conn,
		Send: make(chan map[string]interface{}, 1024), // 增加缓冲区大小
		Hub:  signalingHub,
	}

	// 启动goroutines
	go client.writePump()
	go client.readPump()

	// 发送欢迎消息
	welcomeMsg := map[string]interface{}{
		"type":    "welcome",
		"message": "欢迎连接到信令服务器",
		"time":    time.Now(),
	}
	client.Send <- welcomeMsg
}

// 读取消息
func (c *WebSocketClient) readPump() {
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
		var message map[string]interface{}
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket错误: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// 写入消息
func (c *WebSocketClient) writePump() {
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

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("写入消息失败: %v", err)
				// 只有在严重错误时才断开连接
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					return
				}
				// 其他错误继续尝试
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 处理消息
func (c *WebSocketClient) handleMessage(message map[string]interface{}) {
	messageType, _ := message["type"].(string)

	log.Printf("收到消息: %s from %s (客户端ID: %s)", messageType, c.UserID, c.ID)
	log.Printf("完整消息内容: %+v", message)

	switch messageType {
	case "join-meeting":
		log.Printf("处理加入会议消息...")
		c.handleJoinMeeting(message)
	case "leave-meeting":
		c.handleLeaveMeeting(message)
	case "offer", "answer", "ice-candidate":
		c.handleWebRTCSignaling(message)
	case "chat-message":
		c.handleChatMessage(message)
	case "media-state":
		c.handleMediaState(message)
	case "request-presenter":
		c.handleRequestPresenter(message)
	case "release-presenter":
		c.handleReleasePresenter(message)
	default:
		log.Printf("未知消息类型: %s", messageType)
	}
}

// 处理加入会议
func (c *WebSocketClient) handleJoinMeeting(message map[string]interface{}) {
	log.Printf("开始处理加入会议消息: %+v", message)

	data, ok := message["data"].(map[string]interface{})
	if !ok {
		log.Printf("错误: 无法解析消息数据部分")
		return
	}
	log.Printf("解析到数据: %+v", data)

	meetingID, ok := data["meetingId"].(string)
	if !ok {
		log.Printf("错误: 无法解析meetingId")
		return
	}
	log.Printf("解析到会议ID: %s", meetingID)

	user, ok := data["user"].(map[string]interface{})
	if !ok {
		log.Printf("错误: 无法解析用户信息")
		return
	}
	log.Printf("解析到用户信息: %+v", user)

	userID, ok := user["id"].(string)
	if !ok {
		log.Printf("错误: 无法解析用户ID")
		return
	}

	username, ok := user["name"].(string)
	if !ok {
		log.Printf("错误: 无法解析用户名")
		return
	}

	c.UserID = userID
	c.Username = username
	c.MeetingID = meetingID

	log.Printf("用户 %s (%s) 请求加入会议 %s", username, userID, meetingID)

	// 注册到Hub，Hub会自动处理用户同步
	c.Hub.register <- c

	log.Printf("用户 %s 注册请求已发送到Hub", username)
}

// 处理离开会议
func (c *WebSocketClient) handleLeaveMeeting(message map[string]interface{}) {
	c.Hub.unregister <- c
}

// 处理WebRTC信令
func (c *WebSocketClient) handleWebRTCSignaling(message map[string]interface{}) {
	// 添加发送者信息
	message["from"] = c.UserID
	message["userName"] = c.Username

	// 检查是否有目标用户
	to, hasTo := message["to"].(string)
	if hasTo && to != "" {
		// 点对点消息，直接发送给目标用户
		for client := range c.Hub.clients {
			if client.UserID == to && client.MeetingID == c.MeetingID {
				select {
				case client.Send <- message:
					log.Printf("转发%s消息: %s(%s) -> %s", message["type"], c.UserID, c.Username, to)
				default:
					log.Printf("发送消息失败: %s -> %s", c.UserID, to)
				}
				return
			}
		}
		log.Printf("目标用户不存在: %s", to)
	} else {
		// 广播消息到会议中的所有其他用户
		c.Hub.broadcastToMeeting(c.MeetingID, message, c.UserID)
	}
}

// 处理聊天消息
func (c *WebSocketClient) handleChatMessage(message map[string]interface{}) {
	// 使用新的聊天管理器处理消息
	if c.Hub.chatManager == nil {
		log.Printf("聊天管理器未初始化")
		return
	}

	if err := c.Hub.chatManager.HandleChatMessage(c, message); err != nil {
		log.Printf("处理聊天消息失败: %v", err)

		// 向客户端发送错误消息
		errorMsg := map[string]interface{}{
			"type": "chat-error",
			"data": map[string]interface{}{
				"error":   err.Error(),
				"message": "消息发送失败",
			},
		}

		select {
		case c.Send <- errorMsg:
		default:
			log.Printf("发送错误消息失败")
		}
	}
}

// 处理媒体状态
func (c *WebSocketClient) handleMediaState(message map[string]interface{}) {
	// 广播媒体状态变化
	broadcastMsg := map[string]interface{}{
		"type":       "media-state",
		"data":       message["data"],
		"from":       c.UserID,
		"meeting_id": c.MeetingID,
	}

	c.Hub.broadcast <- broadcastMsg
}

// 处理申请主讲人
func (c *WebSocketClient) handleRequestPresenter(message map[string]interface{}) {
	log.Printf("用户 %s 申请成为会议 %s 的主讲人", c.Username, c.MeetingID)

	// 设置为主讲人
	c.Hub.setPresenter(c.MeetingID, c)
}

// 处理释放主讲人
func (c *WebSocketClient) handleReleasePresenter(message map[string]interface{}) {
	log.Printf("用户 %s 释放会议 %s 的主讲人权限", c.Username, c.MeetingID)

	// 移除主讲人
	c.Hub.removePresenter(c.MeetingID, c)
}

// 聊天API处理函数

// getChatHistory 获取聊天历史
func getChatHistory(c *gin.Context) {
	meetingID := c.Param("meetingId")
	limitStr := c.DefaultQuery("limit", "50")

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	if signalingHub.chatManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "聊天管理器未初始化"})
		return
	}

	history := signalingHub.chatManager.GetChatHistory(meetingID, limit)

	c.JSON(http.StatusOK, gin.H{
		"meetingId": meetingID,
		"messages":  history,
		"count":     len(history),
	})
}

// getChatStats 获取聊天统计
func getChatStats(c *gin.Context) {
	meetingID := c.Param("meetingId")

	if signalingHub.chatManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "聊天管理器未初始化"})
		return
	}

	stats := signalingHub.chatManager.GetChatStats(meetingID)

	c.JSON(http.StatusOK, gin.H{
		"meetingId": meetingID,
		"stats":     stats,
	})
}

// exportChatHistory 导出聊天历史
func exportChatHistory(c *gin.Context) {
	meetingID := c.Param("meetingId")

	if signalingHub.chatManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "聊天管理器未初始化"})
		return
	}

	data, err := signalingHub.chatManager.ExportChatHistory(meetingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("chat_history_%s_%s.json", meetingID, time.Now().Format("20060102_150405"))

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", data)
}

// clearChatHistory 清除聊天历史
func clearChatHistory(c *gin.Context) {
	meetingID := c.Param("meetingId")

	if signalingHub.chatManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "聊天管理器未初始化"})
		return
	}

	signalingHub.chatManager.ClearChatHistory(meetingID)

	c.JSON(http.StatusOK, gin.H{
		"message":   "聊天历史已清除",
		"meetingId": meetingID,
	})
}
