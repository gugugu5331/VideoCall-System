package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Meeting 会议模型
type Meeting struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Title             string         `json:"title" gorm:"not null;size:200"`
	Description       string         `json:"description" gorm:"type:text"`
	CreatorID         uuid.UUID      `json:"creator_id" gorm:"type:uuid;not null"`
	Creator           User           `json:"creator" gorm:"foreignKey:CreatorID"`
	StartTime         time.Time      `json:"start_time" gorm:"not null"`
	EndTime           *time.Time     `json:"end_time"`
	Duration          int            `json:"duration"` // 预计时长(分钟)
	MaxParticipants   int            `json:"max_participants" gorm:"default:50"`
	IsPublic          bool           `json:"is_public" gorm:"default:false"`
	JoinCode          string         `json:"join_code" gorm:"uniqueIndex;size:20"`
	Status            string         `json:"status" gorm:"default:'scheduled';size:20"`
	RecordingEnabled  bool           `json:"recording_enabled" gorm:"default:true"`
	DetectionEnabled  bool           `json:"detection_enabled" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Participants []MeetingParticipant `json:"participants" gorm:"foreignKey:MeetingID"`
}

// MeetingStatus 会议状态常量
const (
	MeetingStatusScheduled = "scheduled"
	MeetingStatusActive    = "active"
	MeetingStatusEnded     = "ended"
	MeetingStatusCancelled = "cancelled"
)

// MeetingParticipant 会议参与者模型
type MeetingParticipant struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MeetingID uuid.UUID      `json:"meeting_id" gorm:"type:uuid;not null"`
	Meeting   Meeting        `json:"meeting" gorm:"foreignKey:MeetingID"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	Role      string         `json:"role" gorm:"default:'participant';size:20"`
	JoinTime  *time.Time     `json:"join_time"`
	LeaveTime *time.Time     `json:"leave_time"`
	Status    string         `json:"status" gorm:"default:'invited';size:20"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// ParticipantRole 参与者角色常量
const (
	ParticipantRoleHost      = "host"
	ParticipantRoleModerator = "moderator"
	ParticipantRoleGuest     = "participant"
)

// ParticipantStatus 参与者状态常量
const (
	ParticipantStatusInvited = "invited"
	ParticipantStatusJoined  = "joined"
	ParticipantStatusLeft    = "left"
	ParticipantStatusKicked  = "kicked"
)

// BeforeCreate GORM钩子：创建前
func (m *Meeting) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// BeforeCreate GORM钩子：创建前
func (mp *MeetingParticipant) BeforeCreate(tx *gorm.DB) error {
	if mp.ID == uuid.Nil {
		mp.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (Meeting) TableName() string {
	return "meetings"
}

// TableName 指定表名
func (MeetingParticipant) TableName() string {
	return "meeting_participants"
}

// IsActive 检查会议是否活跃
func (m *Meeting) IsActive() bool {
	return m.Status == MeetingStatusActive
}

// CanJoin 检查是否可以加入会议
func (m *Meeting) CanJoin() bool {
	return m.Status == MeetingStatusScheduled || m.Status == MeetingStatusActive
}

// MeetingCreateRequest 会议创建请求
type MeetingCreateRequest struct {
	Title            string    `json:"title" binding:"required,min=1,max=200"`
	Description      string    `json:"description,omitempty"`
	StartTime        time.Time `json:"start_time" binding:"required"`
	Duration         int       `json:"duration" binding:"required,min=1,max=1440"` // 最长24小时
	MaxParticipants  int       `json:"max_participants,omitempty" binding:"omitempty,min=2,max=100"`
	IsPublic         bool      `json:"is_public"`
	RecordingEnabled bool      `json:"recording_enabled"`
	DetectionEnabled bool      `json:"detection_enabled"`
}

// MeetingUpdateRequest 会议更新请求
type MeetingUpdateRequest struct {
	Title            string     `json:"title,omitempty" binding:"omitempty,min=1,max=200"`
	Description      string     `json:"description,omitempty"`
	StartTime        *time.Time `json:"start_time,omitempty"`
	Duration         int        `json:"duration,omitempty" binding:"omitempty,min=1,max=1440"`
	MaxParticipants  int        `json:"max_participants,omitempty" binding:"omitempty,min=2,max=100"`
	IsPublic         *bool      `json:"is_public,omitempty"`
	RecordingEnabled *bool      `json:"recording_enabled,omitempty"`
	DetectionEnabled *bool      `json:"detection_enabled,omitempty"`
}

// MeetingJoinRequest 加入会议请求
type MeetingJoinRequest struct {
	JoinCode string `json:"join_code,omitempty"`
}

// MeetingResponse 会议响应
type MeetingResponse struct {
	ID               uuid.UUID                    `json:"id"`
	Title            string                       `json:"title"`
	Description      string                       `json:"description"`
	Creator          *UserResponse                `json:"creator"`
	StartTime        time.Time                    `json:"start_time"`
	EndTime          *time.Time                   `json:"end_time"`
	Duration         int                          `json:"duration"`
	MaxParticipants  int                          `json:"max_participants"`
	IsPublic         bool                         `json:"is_public"`
	JoinCode         string                       `json:"join_code,omitempty"`
	Status           string                       `json:"status"`
	RecordingEnabled bool                         `json:"recording_enabled"`
	DetectionEnabled bool                         `json:"detection_enabled"`
	ParticipantCount int                          `json:"participant_count"`
	Participants     []ParticipantResponse        `json:"participants,omitempty"`
	CreatedAt        time.Time                    `json:"created_at"`
}

// ParticipantResponse 参与者响应
type ParticipantResponse struct {
	ID       uuid.UUID     `json:"id"`
	User     *UserResponse `json:"user"`
	Role     string        `json:"role"`
	JoinTime *time.Time    `json:"join_time"`
	Status   string        `json:"status"`
}

// ToResponse 转换为响应格式
func (m *Meeting) ToResponse() *MeetingResponse {
	response := &MeetingResponse{
		ID:               m.ID,
		Title:            m.Title,
		Description:      m.Description,
		StartTime:        m.StartTime,
		EndTime:          m.EndTime,
		Duration:         m.Duration,
		MaxParticipants:  m.MaxParticipants,
		IsPublic:         m.IsPublic,
		Status:           m.Status,
		RecordingEnabled: m.RecordingEnabled,
		DetectionEnabled: m.DetectionEnabled,
		ParticipantCount: len(m.Participants),
		CreatedAt:        m.CreatedAt,
	}

	if m.Creator.ID != uuid.Nil {
		response.Creator = m.Creator.ToResponse()
	}

	// 只有创建者或参与者才能看到加入码
	if !m.IsPublic {
		response.JoinCode = m.JoinCode
	}

	// 转换参与者信息
	for _, p := range m.Participants {
		participant := ParticipantResponse{
			ID:       p.ID,
			Role:     p.Role,
			JoinTime: p.JoinTime,
			Status:   p.Status,
		}
		if p.User.ID != uuid.Nil {
			participant.User = p.User.ToResponse()
		}
		response.Participants = append(response.Participants, participant)
	}

	return response
}
