package models

import (
	"time"

	"gorm.io/gorm"
)

// MediaFile 媒体文件模型
type MediaFile struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FileID      string    `json:"file_id" gorm:"uniqueIndex;not null"`
	FileName    string    `json:"file_name" gorm:"not null"`
	OriginalName string   `json:"original_name" gorm:"not null"`
	FileType    string    `json:"file_type" gorm:"not null"` // video, audio, image
	MimeType    string    `json:"mime_type" gorm:"not null"`
	FileSize    int64     `json:"file_size" gorm:"not null"`
	Duration    float64   `json:"duration"`                  // 媒体时长（秒）
	Width       int       `json:"width"`                     // 视频/图片宽度
	Height      int       `json:"height"`                    // 视频/图片高度
	Bitrate     int       `json:"bitrate"`                   // 比特率
	FrameRate   float64   `json:"frame_rate"`                // 帧率
	Codec       string    `json:"codec"`                     // 编解码器
	StoragePath string    `json:"storage_path" gorm:"not null"`
	ThumbnailPath string  `json:"thumbnail_path"`
	Status      string    `json:"status" gorm:"default:'uploaded'"` // uploaded, processing, processed, error
	UserID      string    `json:"user_id" gorm:"not null"`
	MeetingID   string    `json:"meeting_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 元数据
	Metadata MediaMetadata `json:"metadata" gorm:"embedded"`
}

// MediaMetadata 媒体元数据
type MediaMetadata struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags" gorm:"serializer:json"`
	Properties  map[string]string `json:"properties" gorm:"serializer:json"`
}

// ProcessingJob 媒体处理任务
type ProcessingJob struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	JobID       string    `json:"job_id" gorm:"uniqueIndex;not null"`
	MediaFileID string    `json:"media_file_id" gorm:"not null"`
	JobType     string    `json:"job_type" gorm:"not null"` // transcode, extract_audio, extract_video, merge, thumbnail, filter
	Status      string    `json:"status" gorm:"default:'pending'"` // pending, processing, completed, failed
	Progress    float64   `json:"progress" gorm:"default:0"`
	InputPath   string    `json:"input_path" gorm:"not null"`
	OutputPath  string    `json:"output_path"`
	Parameters  JobParameters `json:"parameters" gorm:"embedded"`
	Error       string    `json:"error"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JobParameters 任务参数
type JobParameters struct {
	Format      string            `json:"format"`       // 输出格式
	Quality     string            `json:"quality"`      // 质量设置
	Resolution  string            `json:"resolution"`   // 分辨率
	Bitrate     string            `json:"bitrate"`      // 比特率
	FrameRate   string            `json:"frame_rate"`   // 帧率
	AudioCodec  string            `json:"audio_codec"`  // 音频编解码器
	VideoCodec  string            `json:"video_codec"`  // 视频编解码器
	Filters     []string          `json:"filters" gorm:"serializer:json"`     // 滤镜列表
	CustomArgs  map[string]string `json:"custom_args" gorm:"serializer:json"` // 自定义参数
}

// Recording 录制记录
type Recording struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	RecordingID string    `json:"recording_id" gorm:"uniqueIndex;not null"`
	MeetingID   string    `json:"meeting_id" gorm:"not null"`
	RoomID      string    `json:"room_id" gorm:"not null"`
	UserID      string    `json:"user_id" gorm:"not null"`
	Title       string    `json:"title" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'recording'"` // recording, processing, completed, failed
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Duration    float64   `json:"duration"` // 录制时长（秒）
	FileSize    int64     `json:"file_size"`
	FilePath    string    `json:"file_path"`
	ThumbnailPath string  `json:"thumbnail_path"`
	Quality     string    `json:"quality" gorm:"default:'720p'"`
	Format      string    `json:"format" gorm:"default:'mp4'"`
	Participants []string `json:"participants" gorm:"serializer:json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// SFU 架构：Stream 模型已删除
// 原因：SFU不应进行服务端流媒体处理（RTMP/HLS等），仅负责WebRTC RTP转发
// 已删除的模型：
// - Stream: 流媒体记录（用于RTMP/HLS，已删除）

// WebRTCPeer WebRTC对等连接
type WebRTCPeer struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	PeerID      string    `json:"peer_id" gorm:"uniqueIndex;not null"`
	RoomID      string    `json:"room_id" gorm:"not null"`
	UserID      string    `json:"user_id" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'connecting'"` // connecting, connected, disconnected
	PeerType    string    `json:"peer_type" gorm:"not null"`          // publisher, subscriber
	MediaType   string    `json:"media_type" gorm:"not null"`         // audio, video, screen
	SDPOffer    string    `json:"sdp_offer"`
	SDPAnswer   string    `json:"sdp_answer"`
	ICECandidates []string `json:"ice_candidates" gorm:"serializer:json"`
	ConnectedAt time.Time `json:"connected_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SFU 架构：滤镜模型已删除
// 原因：SFU 架构要求所有滤镜、美颜等视觉效果在客户端处理
// 如需保留滤镜配置，应将其移至用户服务，通过信令传递参数给客户端
//
// 已删除的模型：
// - Filter: 滤镜配置（已删除）
// - FilterParameters: 滤镜参数（已删除）
//
// 替代方案：
// - 在用户服务中存储用户的滤镜偏好设置
// - 通过 WebSocket 信令将滤镜参数发送给客户端
// - 客户端本地应用滤镜效果后再发送媒体流

// MediaStats 媒体统计信息
type MediaStats struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	MediaFileID     string    `json:"media_file_id" gorm:"not null"`
	ViewCount       int64     `json:"view_count" gorm:"default:0"`
	DownloadCount   int64     `json:"download_count" gorm:"default:0"`
	ProcessingCount int64     `json:"processing_count" gorm:"default:0"`
	ShareCount      int64     `json:"share_count" gorm:"default:0"`
	LastViewedAt    *time.Time `json:"last_viewed_at"`
	LastDownloadAt  *time.Time `json:"last_download_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (MediaFile) TableName() string {
	return "media_files"
}

func (ProcessingJob) TableName() string {
	return "processing_jobs"
}

func (Recording) TableName() string {
	return "recordings"
}

// SFU 架构：Stream 表名函数已删除（流媒体模型已移除）

func (WebRTCPeer) TableName() string {
	return "webrtc_peers"
}

// SFU 架构：Filter 表名函数已删除（滤镜功能已移除）

func (MediaStats) TableName() string {
	return "media_stats"
}
