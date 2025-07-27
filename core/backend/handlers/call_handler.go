package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"videocall-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebRTC信令消息类型
const (
	MessageTypeOffer        = "offer"
	MessageTypeAnswer       = "answer"
	MessageTypeICECandidate = "ice_candidate"
	MessageTypeJoin         = "join"
	MessageTypeLeave        = "leave"
	MessageTypeError        = "error"
)

// WebRTC信令消息结构
type SignalingMessage struct {
	Type      string      `json:"type"`
	CallID    string      `json:"call_id"`
	UserID    string      `json:"user_id"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// 通话房间结构
type CallRoom struct {
	ID          string                     `json:"id"`
	CallType    string                     `json:"call_type"`
	Status      string                     `json:"status"`
	StartTime   time.Time                  `json:"start_time"`
	Users       map[string]*CallUser       `json:"users"`
	Connections map[string]*websocket.Conn `json:"-"`
	mutex       sync.RWMutex               `json:"-"`
}

// 通话用户结构
type CallUser struct {
	ID       string `json:"id"`
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Role     string `json:"role"` // caller or callee
}

// 全局通话房间管理器
var (
	callRooms  = make(map[string]*CallRoom)
	roomsMutex sync.RWMutex
)

// StartCallRequest 开始通话请求
type StartCallRequest struct {
	CalleeID       string `json:"callee_id"`       // 被叫用户UUID
	CalleeUsername string `json:"callee_username"` // 被叫用户名
	CallType       string `json:"call_type" binding:"required,oneof=audio video"`
}

// StartCall 开始通话
func StartCall(c *gin.Context) {
	var req StartCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// 获取当前用户信息
	userID, _ := c.Get("user_id")
	userUUID, _ := c.Get("user_uuid")

	userIDUint := userID.(uint)
	userUUIDStr := userUUID.(string)

	// 解析调用者UUID
	callerUUIDParsed, err := uuid.Parse(userUUIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid caller UUID",
		})
		return
	}

	// 查找被叫用户
	var callee models.User
	if err := DB.Where("username = ?", req.CalleeUsername).First(&callee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Callee not found",
		})
		return
	}

	// 检查是否是自己给自己打电话
	if callee.UUID.String() == userUUIDStr {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot call yourself",
		})
		return
	}

	// 创建通话记录
	startTime := time.Now()
	call := models.Call{
		CallerID:   &userIDUint,
		CalleeID:   &callee.ID,
		CallerUUID: &callerUUIDParsed,
		CalleeUUID: &callee.UUID,
		CallType:   req.CallType,
		Status:     "initiated",
		StartTime:  &startTime,
		RoomID:     uuid.New().String(),
	}

	if err := DB.Create(&call).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create call",
		})
		return
	}

	// 通知被叫方有新通话
	notifyCallee(call, callee)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Call initiated successfully",
		"call": gin.H{
			"id":              call.ID,
			"uuid":            call.UUID,
			"room_id":         call.RoomID,
			"caller_username": userUUIDStr,
			"callee_username": callee.Username,
			"call_type":       call.CallType,
			"status":          call.Status,
			"start_time":      call.StartTime,
		},
	})
}

// notifyCallee 通知被叫方有新通话
func notifyCallee(call models.Call, callee models.User) {
	log.Printf("通知被叫方 %s (UUID: %s) 有新通话: CallID=%d, CallUUID=%s",
		callee.Username, callee.UUID.String(), call.ID, call.UUID)

	// 发送实时通知
	callData := gin.H{
		"call_id":         call.ID,
		"call_uuid":       call.UUID.String(),
		"caller_username": "caller", // 这里应该从数据库获取主叫用户名
		"call_type":       call.CallType,
		"timestamp":       time.Now().Unix(),
	}

	sendNotificationToUser(callee.UUID.String(), "incoming_call", callData)
}

// EndCall 结束通话
type EndCallRequest struct {
	CallID   uint   `json:"call_id"`
	CallUUID string `json:"call_uuid"`
}

func EndCall(c *gin.Context) {
	var req EndCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	var call models.Call
	var err error

	// 优先使用UUID查找，如果没有则使用ID
	if req.CallUUID != "" {
		err = DB.Where("uuid = ?", req.CallUUID).First(&call).Error
	} else if req.CallID != 0 {
		err = DB.First(&call, req.CallID).Error
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either call_id or call_uuid is required",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Call not found",
		})
		return
	}

	// 检查权限
	userID, _ := c.Get("user_id")
	if call.CallerID != nil && *call.CallerID != userID.(uint) &&
		call.CalleeID != nil && *call.CalleeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to end this call",
		})
		return
	}

	// 更新通话状态
	now := time.Now()
	duration := int(now.Sub(*call.StartTime).Seconds())

	updates := map[string]interface{}{
		"status":   "ended",
		"end_time": &now,
		"duration": &duration,
	}

	if err := DB.Model(&call).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to end call",
		})
		return
	}

	// 清理通话房间
	roomsMutex.Lock()
	if room, exists := callRooms[call.UUID.String()]; exists {
		// 通知所有用户通话结束
		room.mutex.Lock()
		for userID, conn := range room.Connections {
			if conn != nil {
				message := SignalingMessage{
					Type:      MessageTypeLeave,
					CallID:    room.ID,
					UserID:    userID,
					Timestamp: time.Now().Unix(),
				}
				conn.WriteJSON(message)
			}
		}
		room.mutex.Unlock()
		delete(callRooms, call.UUID.String())
	}
	roomsMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": "Call ended successfully",
		"call": gin.H{
			"id":       call.ID,
			"duration": duration,
			"status":   "ended",
		},
	})
}

// GetCallHistory 获取通话历史
func GetCallHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var calls []models.Call
	var total int64

	// 获取总数
	DB.Model(&models.Call{}).Where("caller_id = ? OR callee_id = ?", userID, userID).Count(&total)

	// 获取通话记录
	if err := DB.Preload("Caller").Preload("Callee").
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&calls).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get call history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"calls": calls,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetCallDetails 获取通话详情
func GetCallDetails(c *gin.Context) {
	callID := c.Param("id")
	userID, _ := c.Get("user_id")

	var call models.Call
	var err error

	// 尝试通过UUID查找
	if err = DB.Preload("Caller").Preload("Callee").
		Where("uuid = ?", callID).First(&call).Error; err != nil {
		// 如果UUID查找失败，尝试通过数字ID查找
		// 先检查callID是否为数字
		if _, parseErr := strconv.ParseUint(callID, 10, 64); parseErr == nil {
			// 是数字，使用First查找
			if err = DB.Preload("Caller").Preload("Callee").
				First(&call, callID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Call not found",
				})
				return
			}
		} else {
			// 不是数字，直接返回未找到
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Call not found",
			})
			return
		}
	}

	// 检查权限
	if call.CallerID != nil && *call.CallerID != userID.(uint) &&
		call.CalleeID != nil && *call.CalleeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to view this call",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"call": call,
	})
}

// WebSocketHandler WebSocket处理器 - 真正的信令服务器
func WebSocketHandler(c *gin.Context) {
	callID := c.Param("callId")

	// 从查询参数获取用户ID
	userID := c.Query("user_id")
	log.Printf("WebSocket连接请求 - CallID: %s, UserID from query: %s", callID, userID)

	// 如果查询参数中没有用户ID，尝试从数据库查找通话信息
	if userID == "" {
		var call models.Call
		if err := DB.Where("uuid = ?", callID).First(&call).Error; err == nil {
			log.Printf("从数据库找到通话记录: CallerUUID=%v, CalleeUUID=%v", call.CallerUUID, call.CalleeUUID)
			// 从通话记录中获取用户信息
			if call.CallerUUID != nil {
				userID = call.CallerUUID.String()
				log.Printf("使用主叫用户ID: %s", userID)
			} else if call.CalleeUUID != nil {
				userID = call.CalleeUUID.String()
				log.Printf("使用被叫用户ID: %s", userID)
			}
		} else {
			log.Printf("数据库中没有找到通话记录: %v", err)
		}
	}

	// 如果还是没有用户ID，尝试从JWT token中获取用户信息
	if userID == "" {
		if userUUID, exists := c.Get("user_uuid"); exists && userUUID != nil {
			if userIDStr, ok := userUUID.(string); ok {
				userID = userIDStr
				log.Printf("从JWT token获取用户ID: %s", userID)
			}
		}
	}

	// 如果还是没有用户ID，尝试从请求头获取
	if userID == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("Found Authorization header, but token parsing not implemented yet")
		}
	}

	// 如果还是没有用户ID，使用默认值（用于测试）
	if userID == "" {
		userID = "test-user-" + callID + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
		log.Printf("Using default user ID for WebSocket: %s", userID)
	}

	log.Printf("最终使用的用户ID: %s", userID)

	// 升级HTTP连接为WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境中应该更严格
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// 查找或创建通话房间
	roomsMutex.Lock()
	room, exists := callRooms[callID]
	if !exists {
		// 如果房间不存在，从数据库查找通话信息
		var call models.Call
		if err := DB.Where("uuid = ?", callID).First(&call).Error; err == nil {
			log.Printf("创建房间，从数据库获取通话信息")
			room = &CallRoom{
				ID:          callID,
				CallType:    call.CallType,
				Status:      call.Status,
				StartTime:   time.Now(),
				Users:       make(map[string]*CallUser),
				Connections: make(map[string]*websocket.Conn),
			}

			// 添加主叫用户
			if call.CallerUUID != nil {
				room.Users[call.CallerUUID.String()] = &CallUser{
					ID:       strconv.FormatUint(uint64(*call.CallerID), 10),
					UUID:     call.CallerUUID.String(),
					Username: "caller", // 这里应该从数据库获取用户名
					Role:     "caller",
				}
				log.Printf("添加主叫用户到房间: %s", call.CallerUUID.String())
			}

			// 添加被叫用户
			if call.CalleeUUID != nil {
				room.Users[call.CalleeUUID.String()] = &CallUser{
					ID:       strconv.FormatUint(uint64(*call.CalleeID), 10),
					UUID:     call.CalleeUUID.String(),
					Username: "callee", // 这里应该从数据库获取用户名
					Role:     "callee",
				}
				log.Printf("添加被叫用户到房间: %s", call.CalleeUUID.String())
			}
		} else {
			// 如果数据库中没有找到通话记录，创建临时房间
			log.Printf("数据库中没有找到通话记录，创建临时房间")
			room = &CallRoom{
				ID:          callID,
				CallType:    "video",
				Status:      "active",
				StartTime:   time.Now(),
				Users:       make(map[string]*CallUser),
				Connections: make(map[string]*websocket.Conn),
			}
		}
		callRooms[callID] = room
	}
	roomsMutex.Unlock()

	// 添加用户到房间
	room.mutex.Lock()
	userExists := false
	if _, exists := room.Users[userID]; exists {
		userExists = true
		log.Printf("用户 %s 已存在于房间中", userID)
	} else {
		// 根据房间中的用户数量分配角色
		role := "participant"
		if len(room.Users) == 0 {
			role = "caller"
		} else if len(room.Users) == 1 {
			role = "callee"
		}

		room.Users[userID] = &CallUser{
			ID:       userID,
			UUID:     userID,
			Username: "user", // 这里应该从数据库获取用户名
			Role:     role,
		}
		log.Printf("用户 %s 加入房间，角色: %s", userID, role)
	}
	room.Connections[userID] = conn
	room.mutex.Unlock()

	// 发送连接成功消息
	conn.WriteJSON(SignalingMessage{
		Type:      "connection",
		CallID:    callID,
		UserID:    userID,
		Timestamp: time.Now().Unix(),
		Data: gin.H{
			"message": "WebSocket connected successfully",
			"room":    room,
		},
	})

	// 通知其他用户有新用户加入（只有新用户才发送通知）
	if !userExists {
		room.mutex.RLock()
		log.Printf("房间 %s 中的用户数量: %d", callID, len(room.Connections))
		log.Printf("当前用户ID: %s", userID)
		log.Printf("房间中的所有用户: %v", getRoomUserIDs(room))

		notificationSent := false
		for uid, userConn := range room.Connections {
			log.Printf("检查用户: %s, 是否等于当前用户: %v", uid, uid != userID)
			if uid != userID && userConn != nil {
				log.Printf("发送join消息给用户: %s", uid)
				err := userConn.WriteJSON(SignalingMessage{
					Type:      MessageTypeJoin,
					CallID:    callID,
					UserID:    userID,
					Timestamp: time.Now().Unix(),
					Data: gin.H{
						"user": room.Users[userID],
					},
				})
				if err != nil {
					log.Printf("发送join消息给用户 %s 失败: %v", uid, err)
				} else {
					log.Printf("成功发送join消息给用户: %s", uid)
					notificationSent = true
				}
			}
		}

		if !notificationSent {
			log.Printf("警告: 没有发送任何join通知消息")
		}
		room.mutex.RUnlock()
	}

	// 消息处理循环
	for {
		// 读取消息
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// 解析消息
		var message SignalingMessage
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// 处理不同类型的消息
		switch message.Type {
		case MessageTypeOffer:
			handleOffer(room, userID, message)
		case MessageTypeAnswer:
			handleAnswer(room, userID, message)
		case MessageTypeICECandidate:
			handleICECandidate(room, userID, message)
		case MessageTypeJoin:
			handleJoin(room, userID, message)
		case MessageTypeLeave:
			handleLeave(room, userID, message)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}

	// 用户断开连接，清理资源
	room.mutex.Lock()
	delete(room.Connections, userID)
	delete(room.Users, userID)

	// 通知其他用户该用户离开
	for _, userConn := range room.Connections {
		if userConn != nil {
			userConn.WriteJSON(SignalingMessage{
				Type:      MessageTypeLeave,
				CallID:    callID,
				UserID:    userID,
				Timestamp: time.Now().Unix(),
			})
		}
	}

	// 如果房间为空，删除房间
	if len(room.Users) == 0 {
		roomsMutex.Lock()
		delete(callRooms, callID)
		roomsMutex.Unlock()
	}
	room.mutex.Unlock()
}

// 处理Offer消息
func handleOffer(room *CallRoom, senderID string, message SignalingMessage) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	// 转发Offer给其他用户
	for _, conn := range room.Connections {
		if conn != nil {
			conn.WriteJSON(SignalingMessage{
				Type:      MessageTypeOffer,
				CallID:    room.ID,
				UserID:    senderID,
				Timestamp: time.Now().Unix(),
				Data:      message.Data,
			})
		}
	}
}

// 处理Answer消息
func handleAnswer(room *CallRoom, senderID string, message SignalingMessage) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	// 转发Answer给其他用户
	for _, conn := range room.Connections {
		if conn != nil {
			conn.WriteJSON(SignalingMessage{
				Type:      MessageTypeAnswer,
				CallID:    room.ID,
				UserID:    senderID,
				Timestamp: time.Now().Unix(),
				Data:      message.Data,
			})
		}
	}
}

// 处理ICE候选消息
func handleICECandidate(room *CallRoom, senderID string, message SignalingMessage) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	// 转发ICE候选给其他用户
	for _, conn := range room.Connections {
		if conn != nil {
			conn.WriteJSON(SignalingMessage{
				Type:      MessageTypeICECandidate,
				CallID:    room.ID,
				UserID:    senderID,
				Timestamp: time.Now().Unix(),
				Data:      message.Data,
			})
		}
	}
}

// 处理加入消息
func handleJoin(room *CallRoom, senderID string, message SignalingMessage) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	// 通知其他用户有新用户加入
	for _, conn := range room.Connections {
		if conn != nil {
			conn.WriteJSON(SignalingMessage{
				Type:      MessageTypeJoin,
				CallID:    room.ID,
				UserID:    senderID,
				Timestamp: time.Now().Unix(),
				Data:      message.Data,
			})
		}
	}
}

// 处理离开消息
func handleLeave(room *CallRoom, senderID string, message SignalingMessage) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	// 通知其他用户有用户离开
	for _, conn := range room.Connections {
		if conn != nil {
			conn.WriteJSON(SignalingMessage{
				Type:      MessageTypeLeave,
				CallID:    room.ID,
				UserID:    senderID,
				Timestamp: time.Now().Unix(),
			})
		}
	}
}

// getRoomUserIDs 获取房间中所有用户的ID列表
func getRoomUserIDs(room *CallRoom) []string {
	userIDs := make([]string, 0, len(room.Users))
	for uid := range room.Users {
		userIDs = append(userIDs, uid)
	}
	return userIDs
}

// GetActiveCalls 获取活跃通话列表
func GetActiveCalls(c *gin.Context) {
	userID, _ := c.Get("user_uuid")
	var userIDStr string
	if userID != nil {
		if userIDStrTemp, ok := userID.(string); ok {
			userIDStr = userIDStrTemp
		}
	}

	roomsMutex.RLock()
	var activeCalls []gin.H
	for _, room := range callRooms {
		room.mutex.RLock()
		if _, isParticipant := room.Users[userIDStr]; isParticipant {
			activeCalls = append(activeCalls, gin.H{
				"id":         room.ID,
				"call_type":  room.CallType,
				"status":     room.Status,
				"start_time": room.StartTime,
				"users":      room.Users,
			})
		}
		room.mutex.RUnlock()
	}
	roomsMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"active_calls": activeCalls,
	})
}

// 通知WebSocket处理器
func NotificationWebSocketHandler(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		log.Printf("通知WebSocket连接缺少用户ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
		return
	}

	log.Printf("用户 %s 连接通知WebSocket", userID)

	// 升级HTTP连接为WebSocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("通知WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	// 添加到通知连接管理器
	addNotificationConnection(userID, conn)

	// 发送连接成功消息
	conn.WriteJSON(gin.H{
		"type":    "connection",
		"message": "Notification WebSocket connected successfully",
		"user_id": userID,
	})

	// 消息处理循环
	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("通知WebSocket读取错误: %v", err)
			break
		}

		var message map[string]interface{}
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("解析通知消息失败: %v", err)
			continue
		}

		// 处理订阅消息
		if messageType, ok := message["type"].(string); ok && messageType == "subscribe" {
			log.Printf("用户 %s 订阅通知事件", userID)
		}
	}

	// 移除通知连接
	removeNotificationConnection(userID)
}

// 通知连接管理器
var (
	notificationConnections = make(map[string]*websocket.Conn)
	notificationMutex       sync.RWMutex
)

// 添加通知连接
func addNotificationConnection(userID string, conn *websocket.Conn) {
	notificationMutex.Lock()
	defer notificationMutex.Unlock()

	// 关闭旧连接
	if oldConn, exists := notificationConnections[userID]; exists {
		oldConn.Close()
	}

	notificationConnections[userID] = conn
	log.Printf("用户 %s 的通知连接已添加", userID)
}

// 移除通知连接
func removeNotificationConnection(userID string) {
	notificationMutex.Lock()
	defer notificationMutex.Unlock()

	if conn, exists := notificationConnections[userID]; exists {
		conn.Close()
		delete(notificationConnections, userID)
		log.Printf("用户 %s 的通知连接已移除", userID)
	}
}

// 发送通知给用户
func sendNotificationToUser(userID string, notificationType string, data interface{}) {
	notificationMutex.RLock()
	conn, exists := notificationConnections[userID]
	notificationMutex.RUnlock()

	if !exists {
		log.Printf("用户 %s 的通知连接不存在", userID)
		return
	}

	message := gin.H{
		"type":      notificationType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	err := conn.WriteJSON(message)
	if err != nil {
		log.Printf("发送通知给用户 %s 失败: %v", userID, err)
		// 移除失效的连接
		removeNotificationConnection(userID)
	} else {
		log.Printf("成功发送 %s 通知给用户 %s", notificationType, userID)
	}
}
