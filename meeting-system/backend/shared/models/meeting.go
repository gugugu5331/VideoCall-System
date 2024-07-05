package models

import (
	"time"

	"gorm.io/gorm"
)

// Meeting 会议模型
type Meeting struct {
	ID              uint                 `json:"id" gorm:"primaryKey"`
	Title           string               `json:"title" gorm:"size:255;not null"`
	Description     string               `json:"description" gorm:"type:text"`
	CreatorID       uint                 `json:"creator_id" gorm:"not null"`
	StartTime       time.Time            `json:"start_time" gorm:"not null"`
	EndTime         time.Time            `json:"end_time" gorm:"not null"`
	MaxParticipants int                  `json:"max_participants" gorm:"default:100"`
	Status          MeetingStatus        `json:"status" gorm:"default:1"`
	MeetingType     MeetingType          `json:"meeting_type" gorm:"default:1"`
	Password        string               `json:"password,omitempty" gorm:"size:50"`
	RecordingURL    string               `json:"recording_url" gorm:"size:500"`
	Settings        string               `json:"settings" gorm:"type:json"` // JSON格式的会议设置
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	DeletedAt       gorm.DeletedAt       `json:"-" gorm:"index"`

	// 关联关系
	Creator      User                   `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Participants []MeetingParticipant   `json:"participants,omitempty" gorm:"foreignKey:MeetingID"`
}

// MeetingStatus 会议状态
type MeetingStatus int

const (
	MeetingStatusScheduled MeetingStatus = 1 // 已安排
	MeetingStatusStarted   MeetingStatus = 2 // 进行中
	MeetingStatusEnded     MeetingStatus = 3 // 已结束
	MeetingStatusCancelled MeetingStatus = 4 // 已取消
)

// MeetingType 会议类型
type MeetingType int

const (
	MeetingTypePublic  MeetingType = 1 // 公开会议
	MeetingTypePrivate MeetingType = 2 // 私人会议
)

// MeetingParticipant 会议参与者模型
type MeetingParticipant struct {
	ID        uint                    `json:"id" gorm:"primaryKey"`
	MeetingID uint                    `json:"meeting_id" gorm:"not null"`
	UserID    uint                    `json:"user_id" gorm:"not null"`
	Role      ParticipantRole         `json:"role" gorm:"default:1"`
	Status    ParticipantStatus       `json:"status" gorm:"default:1"`
	JoinedAt  *time.Time              `json:"joined_at"`
	LeftAt    *time.Time              `json:"left_at"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
	DeletedAt gorm.DeletedAt          `json:"-" gorm:"index"`

	// 关联关系
	Meeting Meeting `json:"meeting,omitempty" gorm:"foreignKey:MeetingID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// ParticipantRole 参与者角色
type ParticipantRole int

const (
	ParticipantRoleParticipant ParticipantRole = 1 // 普通参与者
	ParticipantRoleModerator   ParticipantRole = 2 // 主持人
	ParticipantRolePresenter   ParticipantRole = 3 // 演示者
)

// ParticipantStatus 参与者状态
type ParticipantStatus int

const (
	ParticipantStatusInvited   ParticipantStatus = 1 // 已邀请
	ParticipantStatusJoined    ParticipantStatus = 2 // 已加入
	ParticipantStatusLeft      ParticipantStatus = 3 // 已离开
	ParticipantStatusRejected  ParticipantStatus = 4 // 已拒绝
)

// TableName 指定表名
func (Meeting) TableName() string {
	return "meetings"
}

func (MeetingParticipant) TableName() string {
	return "meeting_participants"
}

// BeforeCreate 创建前钩子
func (m *Meeting) BeforeCreate(tx *gorm.DB) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return nil
}

func (mp *MeetingParticipant) BeforeCreate(tx *gorm.DB) error {
	mp.CreatedAt = time.Now()
	mp.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Meeting) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

func (mp *MeetingParticipant) BeforeUpdate(tx *gorm.DB) error {
	mp.UpdatedAt = time.Now()
	return nil
}

// IsActive 检查会议是否活跃
func (m *Meeting) IsActive() bool {
	return m.Status == MeetingStatusScheduled || m.Status == MeetingStatusStarted
}

// IsStarted 检查会议是否已开始
func (m *Meeting) IsStarted() bool {
	return m.Status == MeetingStatusStarted
}

// CanJoin 检查是否可以加入会议
func (m *Meeting) CanJoin() bool {
	now := time.Now()
	return m.IsActive() && now.After(m.StartTime.Add(-30*time.Minute)) && now.Before(m.EndTime)
}

// MeetingCreateRequest 会议创建请求
type MeetingCreateRequest struct {
	Title           string    `json:"title" binding:"required,min=1,max=255"`
	Description     string    `json:"description" binding:"max=1000"`
	StartTime       time.Time `json:"start_time" binding:"required"`
	EndTime         time.Time `json:"end_time" binding:"required"`
	MaxParticipants int       `json:"max_participants" binding:"min=1,max=1000"`
	MeetingType     MeetingType `json:"meeting_type" binding:"min=1,max=2"`
	Password        string    `json:"password" binding:"max=50"`
	Settings        string    `json:"settings"`
}

// MeetingUpdateRequest 会议更新请求
type MeetingUpdateRequest struct {
	Title           string      `json:"title" binding:"min=1,max=255"`
	Description     string      `json:"description" binding:"max=1000"`
	StartTime       *time.Time  `json:"start_time"`
	EndTime         *time.Time  `json:"end_time"`
	MaxParticipants *int        `json:"max_participants" binding:"omitempty,min=1,max=1000"`
	MeetingType     *MeetingType `json:"meeting_type" binding:"omitempty,min=1,max=2"`
	Password        *string     `json:"password" binding:"omitempty,max=50"`
	Settings        *string     `json:"settings"`
}

// MeetingJoinRequest 加入会议请求
type MeetingJoinRequest struct {
	MeetingID uint   `json:"meeting_id" binding:"required"`
	Password  string `json:"password"`
}

// MeetingListRequest 会议列表请求
type MeetingListRequest struct {
	Page     int           `form:"page" binding:"min=1"`
	PageSize int           `form:"page_size" binding:"min=1,max=100"`
	Status   MeetingStatus `form:"status"`
	Keyword  string        `form:"keyword"`
}

// ParticipantInviteRequest 邀请参与者请求
type ParticipantInviteRequest struct {
	UserIDs []uint          `json:"user_ids" binding:"required"`
	Role    ParticipantRole `json:"role" binding:"min=1,max=3"`
}

// MeetingSettings 会议设置
type MeetingSettings struct {
	EnableVideo       bool `json:"enable_video"`
	EnableAudio       bool `json:"enable_audio"`
	EnableScreenShare bool `json:"enable_screen_share"`
	EnableChat        bool `json:"enable_chat"`
	EnableRecording   bool `json:"enable_recording"`
	EnableAI          bool `json:"enable_ai"`
	MuteOnJoin        bool `json:"mute_on_join"`
	RequireApproval   bool `json:"require_approval"`
}
