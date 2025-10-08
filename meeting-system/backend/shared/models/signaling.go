package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// SignalingSession 信令会话模型
type SignalingSession struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	SessionID      string         `json:"session_id" gorm:"uniqueIndex;size:64;not null"` // WebSocket会话ID
	UserID         uint           `json:"user_id" gorm:"not null"`
	MeetingID      uint           `json:"meeting_id" gorm:"not null"`
	PeerID         string         `json:"peer_id" gorm:"size:64;not null"` // WebRTC Peer ID
	Status         SessionStatus  `json:"status" gorm:"default:1"`
	JoinedAt       time.Time      `json:"joined_at" gorm:"not null"`
	LastPingAt     *time.Time     `json:"last_ping_at"`
	DisconnectedAt *time.Time     `json:"disconnected_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系（禁用外键约束以避免迁移时的循环依赖）
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Meeting Meeting `json:"meeting,omitempty" gorm:"foreignKey:MeetingID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// SessionStatus 会话状态
type SessionStatus int

const (
	SessionStatusConnecting   SessionStatus = 1 // 连接中
	SessionStatusConnected    SessionStatus = 2 // 已连接
	SessionStatusOffering     SessionStatus = 3 // 发送Offer中
	SessionStatusAnswering    SessionStatus = 4 // 发送Answer中
	SessionStatusStable       SessionStatus = 5 // 连接稳定
	SessionStatusDisconnected SessionStatus = 6 // 已断开
	SessionStatusFailed       SessionStatus = 7 // 连接失败
)

// SignalingMessage 信令消息模型（用于持久化重要消息）
type SignalingMessage struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	MessageID   string         `json:"message_id" gorm:"uniqueIndex;size:64;not null"`
	SessionID   string         `json:"session_id" gorm:"size:64;not null;index"`
	FromUserID  uint           `json:"from_user_id" gorm:"not null"`
	ToUserID    *uint          `json:"to_user_id"` // null表示广播消息
	MeetingID   uint           `json:"meeting_id" gorm:"not null;index"`
	MessageType MessageType    `json:"message_type" gorm:"not null"`
	Payload     string         `json:"payload" gorm:"type:text"` // JSON格式的消息内容
	Status      MessageStatus  `json:"status" gorm:"default:1"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系（禁用外键约束以避免迁移时的循环依赖）
	FromUser User    `json:"from_user,omitempty" gorm:"foreignKey:FromUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ToUser   *User   `json:"to_user,omitempty" gorm:"foreignKey:ToUserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Meeting  Meeting `json:"meeting,omitempty" gorm:"foreignKey:MeetingID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// MessageType 消息类型
type MessageType int

const (
	MessageTypeOffer        MessageType = 1  // WebRTC Offer
	MessageTypeAnswer       MessageType = 2  // WebRTC Answer
	MessageTypeICECandidate MessageType = 3  // ICE候选
	MessageTypeJoinRoom     MessageType = 4  // 加入房间
	MessageTypeLeaveRoom    MessageType = 5  // 离开房间
	MessageTypeUserJoined   MessageType = 6  // 用户加入通知
	MessageTypeUserLeft     MessageType = 7  // 用户离开通知
	MessageTypeChat         MessageType = 8  // 聊天消息
	MessageTypeScreenShare  MessageType = 9  // 屏幕共享
	MessageTypeMediaControl MessageType = 10 // 媒体控制（静音/取消静音等）
	MessageTypePing         MessageType = 11 // 心跳
	MessageTypePong         MessageType = 12 // 心跳响应
	MessageTypeError        MessageType = 13 // 错误消息
	MessageTypeRoomInfo     MessageType = 14 // 房间信息/加入确认
)

// MessageStatus 消息状态
type MessageStatus int

const (
	MessageStatusPending   MessageStatus = 1 // 待发送
	MessageStatusSent      MessageStatus = 2 // 已发送
	MessageStatusDelivered MessageStatus = 3 // 已送达
	MessageStatusFailed    MessageStatus = 4 // 发送失败
)

// WebSocketMessage WebSocket消息结构（用于实时通信）
type WebSocketMessage struct {
	ID         string      `json:"id"`                   // 消息ID
	Type       MessageType `json:"type"`                 // 消息类型
	FromUserID uint        `json:"from_user_id"`         // 发送者用户ID
	ToUserID   *uint       `json:"to_user_id,omitempty"` // 接收者用户ID（可选）
	MeetingID  uint        `json:"meeting_id"`           // 会议ID
	SessionID  string      `json:"session_id"`           // 会话ID
	PeerID     string      `json:"peer_id,omitempty"`    // Peer ID
	Payload    interface{} `json:"payload"`              // 消息内容
	Timestamp  time.Time   `json:"timestamp"`            // 时间戳
}

// WebRTCOffer WebRTC Offer消息
type WebRTCOffer struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"` // "offer"
}

// WebRTCAnswer WebRTC Answer消息
type WebRTCAnswer struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"` // "answer"
}

// ICECandidate ICE候选消息
type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid"`
	SDPMLineIndex int    `json:"sdpMLineIndex"`
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	MeetingID uint   `json:"meeting_id"`
	UserID    uint   `json:"user_id"`
	PeerID    string `json:"peer_id"`
}

// LeaveRoomRequest 离开房间请求
type LeaveRoomRequest struct {
	MeetingID uint   `json:"meeting_id"`
	UserID    uint   `json:"user_id"`
	PeerID    string `json:"peer_id"`
}

// UserJoinedNotification 用户加入通知
type UserJoinedNotification struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	PeerID    string `json:"peer_id"`
	MeetingID uint   `json:"meeting_id"`
}

// UserLeftNotification 用户离开通知
type UserLeftNotification struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	PeerID    string `json:"peer_id"`
	MeetingID uint   `json:"meeting_id"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Content   string `json:"content"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	MeetingID uint   `json:"meeting_id"`
}

// RoomInfoMessage 房间信息
type RoomInfoMessage struct {
	MeetingID        uint              `json:"meeting_id"`
	ParticipantCount int               `json:"participant_count"`
	SessionID        string            `json:"session_id"`
	PeerID           string            `json:"peer_id"`
	IceServers       []RoomICEServer   `json:"ice_servers"`
	Participants     []RoomParticipant `json:"participants"`
}

// RoomParticipant 房间参与者快照
type RoomParticipant struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	SessionID    string    `json:"session_id"`
	PeerID       string    `json:"peer_id"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsSelf       bool      `json:"is_self"`
}

// RoomICEServer ICE服务器信息
type RoomICEServer struct {
	URLs       string `json:"urls"`
	Username   string `json:"username,omitempty"`
	Credential string `json:"credential,omitempty"`
}

// MediaControlMessage 媒体控制消息
type MediaControlMessage struct {
	Action    string `json:"action"`     // "mute", "unmute", "video_on", "video_off"
	MediaType string `json:"media_type"` // "audio", "video"
	UserID    uint   `json:"user_id"`
	PeerID    string `json:"peer_id"`
}

// ErrorMessage 错误消息
type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ToJSON 将WebSocketMessage转换为JSON字符串
func (m *WebSocketMessage) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析WebSocketMessage
func (m *WebSocketMessage) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), m)
}

// GetMessageTypeName 获取消息类型名称
func (mt MessageType) String() string {
	switch mt {
	case MessageTypeOffer:
		return "offer"
	case MessageTypeAnswer:
		return "answer"
	case MessageTypeICECandidate:
		return "ice-candidate"
	case MessageTypeJoinRoom:
		return "join-room"
	case MessageTypeLeaveRoom:
		return "leave-room"
	case MessageTypeUserJoined:
		return "user-joined"
	case MessageTypeUserLeft:
		return "user-left"
	case MessageTypeChat:
		return "chat"
	case MessageTypeScreenShare:
		return "screen-share"
	case MessageTypeMediaControl:
		return "media-control"
	case MessageTypePing:
		return "ping"
	case MessageTypePong:
		return "pong"
	case MessageTypeError:
		return "error"
	case MessageTypeRoomInfo:
		return "room-info"
	default:
		return "unknown"
	}
}

// GetSessionStatusName 获取会话状态名称
func (ss SessionStatus) String() string {
	switch ss {
	case SessionStatusConnecting:
		return "connecting"
	case SessionStatusConnected:
		return "connected"
	case SessionStatusOffering:
		return "offering"
	case SessionStatusAnswering:
		return "answering"
	case SessionStatusStable:
		return "stable"
	case SessionStatusDisconnected:
		return "disconnected"
	case SessionStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}
