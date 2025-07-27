package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UUID         uuid.UUID `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();uniqueIndex"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	FullName     string    `json:"full_name"`
	AvatarURL    string    `json:"avatar_url"`
	Phone        string    `json:"phone"`
	Status       string    `json:"status" gorm:"default:'active'"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	CallsAsCaller []Call `json:"calls_as_caller" gorm:"foreignKey:CallerID"`
	CallsAsCallee []Call `json:"calls_as_callee" gorm:"foreignKey:CalleeID"`
	Sessions      []UserSession `json:"sessions" gorm:"foreignKey:UserID"`
}

// Call 通话记录模型
type Call struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UUID         uuid.UUID `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();uniqueIndex"`
	CallerID     *uint     `json:"caller_id"`
	CalleeID     *uint     `json:"callee_id"`
	CallerUUID   *uuid.UUID `json:"caller_uuid" gorm:"type:uuid"`
	CalleeUUID   *uuid.UUID `json:"callee_uuid" gorm:"type:uuid"`
	StartTime    *time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	Duration     *int       `json:"duration"` // 通话时长（秒）
	CallType     string     `json:"call_type" gorm:"default:'video'"`
	Status       string     `json:"status" gorm:"default:'initiated'"`
	RoomID       string     `json:"room_id"`
	RecordingURL string     `json:"recording_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联关系
	Caller           *User                `json:"caller" gorm:"foreignKey:CallerID"`
	Callee           *User                `json:"callee" gorm:"foreignKey:CalleeID"`
	SecurityDetections []SecurityDetection `json:"security_detections" gorm:"foreignKey:CallID"`
}

// SecurityDetection 安全检测记录模型
type SecurityDetection struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UUID          uuid.UUID `json:"uuid" gorm:"type:uuid;default:uuid_generate_v4();uniqueIndex"`
	CallID        uint      `json:"call_id"`
	CallUUID      uuid.UUID `json:"call_uuid" gorm:"type:uuid"`
	DetectionType string    `json:"detection_type" gorm:"not null"`
	RiskScore     float64   `json:"risk_score" gorm:"not null"`
	Confidence    float64   `json:"confidence" gorm:"not null"`
	DetectionTime time.Time `json:"detection_time"`
	Details       JSON      `json:"details" gorm:"type:jsonb"`
	ModelVersion  string    `json:"model_version"`
	CreatedAt     time.Time `json:"created_at"`

	// 关联关系
	Call *Call `json:"call" gorm:"foreignKey:CallID"`
}

// UserSession 用户会话模型
type UserSession struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id"`
	SessionToken string    `json:"session_token" gorm:"uniqueIndex;not null"`
	RefreshToken string    `json:"refresh_token" gorm:"uniqueIndex;not null"`
	ExpiresAt    time.Time `json:"expires_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`

	// 关联关系
	User *User `json:"user" gorm:"foreignKey:UserID"`
}

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ConfigKey   string    `json:"config_key" gorm:"uniqueIndex;not null"`
	ConfigValue string    `json:"config_value"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ModelVersion 模型版本管理
type ModelVersion struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ModelName string    `json:"model_name" gorm:"not null"`
	Version   string    `json:"version" gorm:"not null"`
	ModelPath string    `json:"model_path" gorm:"not null"`
	Accuracy  *float64  `json:"accuracy"`
	IsActive  bool      `json:"is_active" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
}

// JSON 自定义JSON类型
type JSON map[string]interface{}

// BeforeCreate GORM钩子：创建前设置UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == uuid.Nil {
		u.UUID = uuid.New()
	}
	return nil
}

func (c *Call) BeforeCreate(tx *gorm.DB) error {
	if c.UUID == uuid.Nil {
		c.UUID = uuid.New()
	}
	return nil
}

func (sd *SecurityDetection) BeforeCreate(tx *gorm.DB) error {
	if sd.UUID == uuid.Nil {
		sd.UUID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Call) TableName() string {
	return "calls"
}

func (SecurityDetection) TableName() string {
	return "security_detections"
}

func (UserSession) TableName() string {
	return "user_sessions"
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

func (ModelVersion) TableName() string {
	return "model_versions"
} 