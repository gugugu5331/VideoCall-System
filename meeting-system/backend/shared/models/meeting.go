package models

import (
	"time"

	"gorm.io/gorm"
)

// Meeting 会议模型
type Meeting struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Title           string         `json:"title" gorm:"size:255;not null"`
	Description     string         `json:"description" gorm:"type:text"`
	CreatorID       uint           `json:"creator_id" gorm:"not null"`
	StartTime       time.Time      `json:"start_time" gorm:"not null"`
	EndTime         time.Time      `json:"end_time" gorm:"not null"`
	MaxParticipants int            `json:"max_participants" gorm:"default:100"`
	Status          MeetingStatus  `json:"status" gorm:"default:1"`
	MeetingType     MeetingType    `json:"meeting_type" gorm:"default:1"`
	Password        string         `json:"password,omitempty" gorm:"size:50"`
	RecordingURL    string         `json:"recording_url" gorm:"size:500"`
	Settings        string         `json:"settings" gorm:"type:json"` // JSON格式的会议设置
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Creator      User                 `json:"creator,omitempty" gorm:"foreignKey:CreatorID"`
	Participants []MeetingParticipant `json:"participants,omitempty" gorm:"foreignKey:MeetingID"`
}

// MeetingStatus 会议状态
type MeetingStatus int

const (
	MeetingStatusScheduled MeetingStatus = 1 // 已安排
	MeetingStatusStarted   MeetingStatus = 2 // 进行中
	MeetingStatusOngoing   MeetingStatus = 2 // 进行中 (别名)
	MeetingStatusEnded     MeetingStatus = 3 // 已结束
	MeetingStatusCancelled MeetingStatus = 4 // 已取消
)

// String 返回状态字符串
func (s MeetingStatus) String() string {
	switch s {
	case MeetingStatusScheduled:
		return "scheduled"
	case MeetingStatusStarted:
		return "active"
	case MeetingStatusEnded:
		return "ended"
	case MeetingStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// MeetingType 会议类型
type MeetingType int

const (
	MeetingTypePublic  MeetingType = 1 // 公开会议
	MeetingTypePrivate MeetingType = 2 // 私人会议
	MeetingTypeVideo   MeetingType = 3 // 视频会议
	MeetingTypeAudio   MeetingType = 4 // 音频会议
)

// MeetingParticipant 会议参与者模型
type MeetingParticipant struct {
	ID        uint              `json:"id" gorm:"primaryKey"`
	MeetingID uint              `json:"meeting_id" gorm:"not null"`
	UserID    uint              `json:"user_id" gorm:"not null"`
	Role      ParticipantRole   `json:"role" gorm:"default:1"`
	Status    ParticipantStatus `json:"status" gorm:"default:1"`
	JoinedAt  *time.Time        `json:"joined_at"`
	LeftAt    *time.Time        `json:"left_at"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt gorm.DeletedAt    `json:"-" gorm:"index"`

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
	ParticipantRoleHost        ParticipantRole = 4 // 主办人
)

// String 返回角色字符串
func (r ParticipantRole) String() string {
	switch r {
	case ParticipantRoleParticipant:
		return "participant"
	case ParticipantRoleModerator:
		return "moderator"
	case ParticipantRolePresenter:
		return "presenter"
	case ParticipantRoleHost:
		return "host"
	default:
		return "participant"
	}
}

// ParticipantStatus 参与者状态
type ParticipantStatus int

const (
	ParticipantStatusInvited  ParticipantStatus = 1 // 已邀请
	ParticipantStatusJoined   ParticipantStatus = 2 // 已加入
	ParticipantStatusLeft     ParticipantStatus = 3 // 已离开
	ParticipantStatusRejected ParticipantStatus = 4 // 已拒绝
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
	Title           string      `json:"title" binding:"required,min=1,max=255"`
	Description     string      `json:"description" binding:"max=1000"`
	StartTime       time.Time   `json:"start_time" binding:"required"`
	EndTime         time.Time   `json:"end_time" binding:"required"`
	MaxParticipants int         `json:"max_participants" binding:"min=1,max=1000"`
	MeetingType     MeetingType `json:"meeting_type" binding:"min=1,max=2"`
	Password        string      `json:"password" binding:"max=50"`
	Settings        string      `json:"settings"`
}

// MeetingUpdateRequest 会议更新请求
type MeetingUpdateRequest struct {
	Title           string       `json:"title" binding:"min=1,max=255"`
	Description     string       `json:"description" binding:"max=1000"`
	StartTime       *time.Time   `json:"start_time"`
	EndTime         *time.Time   `json:"end_time"`
	MaxParticipants *int         `json:"max_participants" binding:"omitempty,min=1,max=1000"`
	MeetingType     *MeetingType `json:"meeting_type" binding:"omitempty,min=1,max=2"`
	Password        *string      `json:"password" binding:"omitempty,max=50"`
	Settings        *string      `json:"settings"`
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

// ===== 请求模型 =====

// CreateMeetingRequest 创建会议请求
type CreateMeetingRequest struct {
	Title           string          `json:"title" binding:"required,min=1,max=100"`
	Description     string          `json:"description" binding:"max=500"`
	StartTime       time.Time       `json:"start_time" binding:"required"`
	EndTime         time.Time       `json:"end_time" binding:"required"`
	MaxParticipants int             `json:"max_participants" binding:"min=1,max=1000"`
	MeetingType     string          `json:"meeting_type" binding:"required,oneof=video audio"`
	Password        string          `json:"password,omitempty" binding:"max=50"`
	Settings        MeetingSettings `json:"settings"`
	CreatorID       uint            `json:"creator_id"`
}

// UpdateMeetingRequest 更新会议请求
type UpdateMeetingRequest struct {
	Title           *string          `json:"title,omitempty" binding:"omitempty,min=1,max=100"`
	Description     *string          `json:"description,omitempty" binding:"omitempty,max=500"`
	StartTime       *time.Time       `json:"start_time,omitempty"`
	EndTime         *time.Time       `json:"end_time,omitempty"`
	MaxParticipants *int             `json:"max_participants,omitempty" binding:"omitempty,min=1,max=1000"`
	Settings        *MeetingSettings `json:"settings,omitempty"`
}

// ===== 响应模型 =====

// MeetingResponse 会议响应
type MeetingResponse struct {
	ID              uint                  `json:"id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	StartTime       time.Time             `json:"start_time"`
	EndTime         time.Time             `json:"end_time"`
	MaxParticipants int                   `json:"max_participants"`
	MeetingType     string                `json:"meeting_type"`
	Status          MeetingStatus         `json:"status"`
	Settings        MeetingSettings       `json:"settings"`
	CreatorID       uint                  `json:"creator_id"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	Participants    []ParticipantResponse `json:"participants,omitempty"`
}

// ParticipantResponse 参与者响应
type ParticipantResponse struct {
	ID        uint              `json:"id"`
	UserID    uint              `json:"user_id"`
	MeetingID uint              `json:"meeting_id"`
	Role      ParticipantRole   `json:"role"`
	Status    ParticipantStatus `json:"status"`
	JoinedAt  *time.Time        `json:"joined_at"`
	LeftAt    *time.Time        `json:"left_at"`
	User      *UserProfile      `json:"user,omitempty"`
}

// RoomStatus 房间状态
type RoomStatus string

const (
	RoomStatusActive   RoomStatus = "active"   // 活跃
	RoomStatusInactive RoomStatus = "inactive" // 非活跃
	RoomStatusClosed   RoomStatus = "closed"   // 已关闭
)

// MeetingRoom 会议室模型
type MeetingRoom struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	MeetingID        uint       `json:"meeting_id" gorm:"not null"`
	RoomID           string     `json:"room_id" gorm:"size:100;not null;unique"`
	SFUNode          string     `json:"sfu_node" gorm:"size:100"`
	Status           RoomStatus `json:"status" gorm:"size:20;default:'active'"`
	ParticipantCount int        `json:"participant_count" gorm:"default:0"`
	MaxBitrate       int        `json:"max_bitrate" gorm:"default:1000000"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// 关联关系
	Meeting Meeting `json:"meeting,omitempty" gorm:"foreignKey:MeetingID"`
}

// JoinMeetingResponse 加入会议响应
type JoinMeetingResponse struct {
	Meeting       MeetingResponse     `json:"meeting"`
	Participant   ParticipantResponse `json:"participant"`
	Token         string              `json:"token,omitempty"`
	MeetingID     uint                `json:"meeting_id"`
	RoomID        string              `json:"room_id"`
	ParticipantID uint                `json:"participant_id"`
	Role          ParticipantRole     `json:"role"`
	SFUNode       string              `json:"sfu_node"`
}

// MediaStream 媒体流模型
type MediaStream struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	MeetingID     uint      `json:"meeting_id" gorm:"not null"`
	ParticipantID uint      `json:"participant_id" gorm:"not null"`
	StreamType    string    `json:"stream_type" gorm:"size:20;not null"` // video, audio, screen
	StreamID      string    `json:"stream_id" gorm:"size:100;not null;unique"`
	Status        string    `json:"status" gorm:"size:20;default:'active'"`  // active, inactive, ended
	Quality       string    `json:"quality" gorm:"size:20;default:'medium'"` // low, medium, high
	Bitrate       int       `json:"bitrate" gorm:"default:0"`
	Resolution    string    `json:"resolution" gorm:"size:20"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联关系
	Meeting     Meeting            `json:"meeting,omitempty" gorm:"foreignKey:MeetingID"`
	Participant MeetingParticipant `json:"participant,omitempty" gorm:"foreignKey:ParticipantID"`
}

// MeetingRecording 会议录制模型
type MeetingRecording struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	MeetingID uint       `json:"meeting_id" gorm:"not null"`
	FileName  string     `json:"file_name" gorm:"size:255;not null"`
	FilePath  string     `json:"file_path" gorm:"size:500;not null"`
	FileSize  int64      `json:"file_size" gorm:"default:0"`
	Duration  int        `json:"duration" gorm:"default:0"` // 录制时长(秒)
	Format    string     `json:"format" gorm:"size:20;default:'mp4'"`
	Quality   string     `json:"quality" gorm:"size:20;default:'medium'"`
	Status    string     `json:"status" gorm:"size:20;default:'processing'"` // processing, completed, failed
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// 关联关系
	Meeting Meeting `json:"meeting,omitempty" gorm:"foreignKey:MeetingID"`
}
