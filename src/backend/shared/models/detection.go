package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DetectionTask 检测任务模型
type DetectionTask struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MeetingID   *uuid.UUID     `json:"meeting_id" gorm:"type:uuid"`
	Meeting     *Meeting       `json:"meeting,omitempty" gorm:"foreignKey:MeetingID"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	FilePath    string         `json:"file_path" gorm:"not null;size:500"`
	FileType    string         `json:"file_type" gorm:"not null;size:20"`
	FileSize    int64          `json:"file_size"`
	Status      string         `json:"status" gorm:"default:'pending';size:20"`
	Priority    int            `json:"priority" gorm:"default:5"`
	CreatedAt   time.Time      `json:"created_at"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Results []DetectionResult `json:"results,omitempty" gorm:"foreignKey:TaskID"`
}

// DetectionResult 检测结果模型
type DetectionResult struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TaskID         uuid.UUID      `json:"task_id" gorm:"type:uuid;not null"`
	Task           DetectionTask  `json:"task" gorm:"foreignKey:TaskID"`
	IsFake         bool           `json:"is_fake" gorm:"not null"`
	Confidence     float64        `json:"confidence" gorm:"not null"`
	DetectionType  string         `json:"detection_type" gorm:"not null;size:50"`
	ModelVersion   string         `json:"model_version" gorm:"size:50"`
	ProcessingTime int            `json:"processing_time"` // 处理时间(毫秒)
	Details        string         `json:"details" gorm:"type:jsonb"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// TaskStatus 任务状态常量
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)

// FileType 文件类型常量
const (
	FileTypeVideo = "video"
	FileTypeAudio = "audio"
	FileTypeImage = "image"
)

// DetectionType 检测类型常量
const (
	DetectionTypeFaceSwap       = "face_swap"
	DetectionTypeVoiceSynthesis = "voice_synthesis"
	DetectionTypeDeepfake       = "deepfake"
	DetectionTypeManipulation   = "manipulation"
)

// BeforeCreate GORM钩子：创建前
func (dt *DetectionTask) BeforeCreate(tx *gorm.DB) error {
	if dt.ID == uuid.Nil {
		dt.ID = uuid.New()
	}
	return nil
}

// BeforeCreate GORM钩子：创建前
func (dr *DetectionResult) BeforeCreate(tx *gorm.DB) error {
	if dr.ID == uuid.Nil {
		dr.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (DetectionTask) TableName() string {
	return "detection_tasks"
}

// TableName 指定表名
func (DetectionResult) TableName() string {
	return "detection_results"
}

// IsCompleted 检查任务是否完成
func (dt *DetectionTask) IsCompleted() bool {
	return dt.Status == TaskStatusCompleted
}

// IsFailed 检查任务是否失败
func (dt *DetectionTask) IsFailed() bool {
	return dt.Status == TaskStatusFailed
}

// DetectionTaskCreateRequest 检测任务创建请求
type DetectionTaskCreateRequest struct {
	MeetingID string `json:"meeting_id,omitempty"`
	FileType  string `json:"file_type" binding:"required,oneof=video audio image"`
	Priority  int    `json:"priority,omitempty" binding:"omitempty,min=1,max=10"`
}

// DetectionTaskResponse 检测任务响应
type DetectionTaskResponse struct {
	ID          uuid.UUID                `json:"id"`
	MeetingID   *uuid.UUID               `json:"meeting_id"`
	User        *UserResponse            `json:"user"`
	FilePath    string                   `json:"file_path,omitempty"`
	FileType    string                   `json:"file_type"`
	FileSize    int64                    `json:"file_size"`
	Status      string                   `json:"status"`
	Priority    int                      `json:"priority"`
	CreatedAt   time.Time                `json:"created_at"`
	StartedAt   *time.Time               `json:"started_at"`
	CompletedAt *time.Time               `json:"completed_at"`
	Results     []DetectionResultResponse `json:"results,omitempty"`
}

// DetectionResultResponse 检测结果响应
type DetectionResultResponse struct {
	ID             uuid.UUID              `json:"id"`
	IsFake         bool                   `json:"is_fake"`
	Confidence     float64                `json:"confidence"`
	DetectionType  string                 `json:"detection_type"`
	ModelVersion   string                 `json:"model_version"`
	ProcessingTime int                    `json:"processing_time"`
	Details        map[string]interface{} `json:"details,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// ToResponse 转换为响应格式
func (dt *DetectionTask) ToResponse() *DetectionTaskResponse {
	response := &DetectionTaskResponse{
		ID:          dt.ID,
		MeetingID:   dt.MeetingID,
		FileType:    dt.FileType,
		FileSize:    dt.FileSize,
		Status:      dt.Status,
		Priority:    dt.Priority,
		CreatedAt:   dt.CreatedAt,
		StartedAt:   dt.StartedAt,
		CompletedAt: dt.CompletedAt,
	}

	if dt.User.ID != uuid.Nil {
		response.User = dt.User.ToResponse()
	}

	// 只有在任务完成时才显示文件路径
	if dt.IsCompleted() {
		response.FilePath = dt.FilePath
	}

	// 转换检测结果
	for _, result := range dt.Results {
		response.Results = append(response.Results, *result.ToResponse())
	}

	return response
}

// ToResponse 转换为响应格式
func (dr *DetectionResult) ToResponse() *DetectionResultResponse {
	response := &DetectionResultResponse{
		ID:             dr.ID,
		IsFake:         dr.IsFake,
		Confidence:     dr.Confidence,
		DetectionType:  dr.DetectionType,
		ModelVersion:   dr.ModelVersion,
		ProcessingTime: dr.ProcessingTime,
		CreatedAt:      dr.CreatedAt,
	}

	// 解析详细信息JSON
	if dr.Details != "" {
		// 这里可以添加JSON解析逻辑
		// json.Unmarshal([]byte(dr.Details), &response.Details)
	}

	return response
}

// DetectionAlert 检测告警
type DetectionAlert struct {
	UserID        string                 `json:"user_id"`
	MeetingID     string                 `json:"meeting_id"`
	DetectionType string                 `json:"detection_type"`
	Confidence    float64                `json:"confidence"`
	Timestamp     time.Time              `json:"timestamp"`
	Details       map[string]interface{} `json:"details,omitempty"`
}
