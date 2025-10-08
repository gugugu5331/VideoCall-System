package services

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"meeting-system/media-service/models"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// MediaService 媒体服务
type MediaService struct {
	config          *config.Config
	db              *gorm.DB
	storage         StorageClient
	signalingClient *SignalingClient
	// SFU 架构：滤镜功能已移除
	// filters         map[string]*models.Filter
	// filtersMux      sync.RWMutex
}

// StorageClient 抽象的存储客户端，便于在测试场景下禁用或替换实现
type StorageClient interface {
	UploadFile(bucket, object string, reader io.Reader, size int64, contentType string) error
	GetFile(bucket, object string) (io.ReadCloser, error)
	DeleteFile(bucket, object string) error
}

// NewMediaService 创建媒体服务
func NewMediaService(config *config.Config, db *gorm.DB, signalingClient *SignalingClient) *MediaService {
	return &MediaService{
		config:          config,
		db:              db,
		storage:         nil,
		signalingClient: signalingClient,
		// SFU 架构：滤镜功能已移除
	}
}

// SetStorageClient 设置存储客户端实现
func (s *MediaService) SetStorageClient(storageClient StorageClient) {
	s.storage = storageClient
}

// Initialize 初始化媒体服务
func (s *MediaService) Initialize() error {
	// 自动迁移数据库表
	if err := s.migrateDatabase(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// SFU 架构：滤镜功能已移除，不再加载滤镜配置

	logger.Info("Media service initialized successfully (SFU mode)")
	return nil
}

// Stop 停止媒体服务
func (s *MediaService) Stop() {
	logger.Info("Media service stopped")
}

// migrateDatabase 迁移数据库
func (s *MediaService) migrateDatabase() error {
	return s.db.AutoMigrate(
		&models.MediaFile{},
		&models.ProcessingJob{},
		&models.Recording{},
		// SFU 架构：Stream 模型已移除（用于RTMP/HLS）
		&models.WebRTCPeer{},
		// SFU 架构：Filter 模型已移除
		&models.MediaStats{},
	)
}

// SFU 架构：滤镜相关函数已删除
// 原因：SFU 架构要求所有滤镜、美颜等视觉效果在客户端处理
//
// 已删除的函数：
// - loadFilters: 加载滤镜配置（已删除）
// - initializeDefaultFilters: 初始化默认滤镜（已删除）

// UploadMedia 上传媒体文件
func (s *MediaService) UploadMedia(file *multipart.FileHeader, userID, meetingID string) (*models.MediaFile, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("storage service not configured")
	}

	// 生成文件ID
	fileID := uuid.New().String()

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))

	// 确定文件类型
	fileType := s.getFileType(ext)
	if fileType == "" {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// 生成存储路径
	storagePath := fmt.Sprintf("media/%s/%s%s", userID, fileID, ext)

	// 决定内容类型，优先使用请求头中的值
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		if detected := mime.TypeByExtension(ext); detected != "" {
			contentType = detected
		} else {
			contentType = "application/octet-stream"
		}
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// 上传到MinIO
	if err := s.storage.UploadFile("media", storagePath, src, file.Size, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	// 创建媒体文件记录
	mediaFile := &models.MediaFile{
		FileID:       fileID,
		FileName:     fmt.Sprintf("%s%s", fileID, ext),
		OriginalName: file.Filename,
		FileType:     fileType,
		MimeType:     contentType,
		FileSize:     file.Size,
		StoragePath:  storagePath,
		Status:       "uploaded",
		UserID:       userID,
		MeetingID:    meetingID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	if err := s.db.Create(mediaFile).Error; err != nil {
		// 如果数据库保存失败，删除已上传的文件
		s.storage.DeleteFile("media", storagePath)
		return nil, fmt.Errorf("failed to save media file record: %w", err)
	}

	// 异步提取媒体信息
	go s.extractMediaInfo(mediaFile)

	logger.Info(fmt.Sprintf("Media file uploaded successfully: %s", fileID))
	return mediaFile, nil
}

// getFileType 根据扩展名确定文件类型
func (s *MediaService) getFileType(ext string) string {
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".m4v"}
	audioExts := []string{".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a", ".wma"}
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	documentExts := []string{".txt", ".pdf", ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx", ".csv", ".md"}

	for _, videoExt := range videoExts {
		if ext == videoExt {
			return "video"
		}
	}

	for _, audioExt := range audioExts {
		if ext == audioExt {
			return "audio"
		}
	}

	for _, imageExt := range imageExts {
		if ext == imageExt {
			return "image"
		}
	}

	for _, docExt := range documentExts {
		if ext == docExt {
			return "document"
		}
	}

	if strings.TrimSpace(ext) == "" {
		return "other"
	}

	return "other"
}

// extractMediaInfo 提取媒体信息
func (s *MediaService) extractMediaInfo(mediaFile *models.MediaFile) {
	// 这里应该使用FFmpeg来提取媒体信息
	// 为了简化，我们先设置一些默认值

	// 更新媒体文件信息
	updates := map[string]interface{}{
		"status":     "processed",
		"updated_at": time.Now(),
	}

	// 根据文件类型设置默认信息
	switch mediaFile.FileType {
	case "video":
		updates["width"] = 1920
		updates["height"] = 1080
		updates["duration"] = 120.0
		updates["frame_rate"] = 30.0
		updates["bitrate"] = 2000
		updates["codec"] = "h264"
	case "audio":
		updates["duration"] = 180.0
		updates["bitrate"] = 128
		updates["codec"] = "aac"
	case "image":
		updates["width"] = 1920
		updates["height"] = 1080
	}

	if err := s.db.Model(mediaFile).Updates(updates).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update media file info: %v", err))
	}

	// 生成缩略图
	if mediaFile.FileType == "video" || mediaFile.FileType == "image" {
		go s.generateThumbnail(mediaFile)
	}
}

// generateThumbnail 生成缩略图
func (s *MediaService) generateThumbnail(mediaFile *models.MediaFile) {
	// 生成缩略图路径
	thumbnailPath := fmt.Sprintf("thumbnails/%s/%s_thumb.jpg", mediaFile.UserID, mediaFile.FileID)

	// 这里应该使用FFmpeg或其他工具生成缩略图
	// 为了简化，我们先跳过实际的缩略图生成

	// 更新缩略图路径
	if err := s.db.Model(mediaFile).Update("thumbnail_path", thumbnailPath).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update thumbnail path: %v", err))
	}
}

// GetMediaFile 获取媒体文件信息
func (s *MediaService) GetMediaFile(fileID string) (*models.MediaFile, error) {
	var mediaFile models.MediaFile
	if err := s.db.Where("file_id = ?", fileID).First(&mediaFile).Error; err != nil {
		return nil, err
	}
	return &mediaFile, nil
}

// DownloadMedia 下载媒体文件
func (s *MediaService) DownloadMedia(fileID string) (io.Reader, *models.MediaFile, error) {
	// 获取媒体文件信息
	mediaFile, err := s.GetMediaFile(fileID)
	if err != nil {
		return nil, nil, err
	}

	if s.storage == nil {
		return nil, nil, fmt.Errorf("storage service not configured")
	}

	// 从存储中获取文件
	reader, err := s.storage.GetFile("media", mediaFile.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file from storage: %w", err)
	}

	// 更新下载统计
	go s.updateDownloadStats(fileID)

	return reader, mediaFile, nil
}

// updateDownloadStats 更新下载统计
func (s *MediaService) updateDownloadStats(fileID string) {
	var stats models.MediaStats
	if err := s.db.Where("media_file_id = ?", fileID).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的统计记录
			stats = models.MediaStats{
				MediaFileID:    fileID,
				DownloadCount:  1,
				LastDownloadAt: &[]time.Time{time.Now()}[0],
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			s.db.Create(&stats)
		}
	} else {
		// 更新现有统计
		now := time.Now()
		s.db.Model(&stats).Updates(map[string]interface{}{
			"download_count":   gorm.Expr("download_count + 1"),
			"last_download_at": &now,
			"updated_at":       now,
		})
	}
}

// DeleteMedia 删除媒体文件
func (s *MediaService) DeleteMedia(fileID string) error {
	// 获取媒体文件信息
	mediaFile, err := s.GetMediaFile(fileID)
	if err != nil {
		return err
	}

	if s.storage != nil {
		// 从存储中删除文件
		if err := s.storage.DeleteFile("media", mediaFile.StoragePath); err != nil {
			logger.Error(fmt.Sprintf("Failed to delete file from storage: %v", err))
		}

		// 删除缩略图
		if mediaFile.ThumbnailPath != "" {
			if err := s.storage.DeleteFile("media", mediaFile.ThumbnailPath); err != nil {
				logger.Error(fmt.Sprintf("Failed to delete thumbnail from storage: %v", err))
			}
		}
	} else {
		logger.Warn("Storage service not configured; skipping delete operations",
			logger.String("file_id", fileID))
	}

	// 从数据库中删除记录
	if err := s.db.Delete(mediaFile).Error; err != nil {
		return fmt.Errorf("failed to delete media file record: %w", err)
	}

	// 删除统计记录
	s.db.Where("media_file_id = ?", fileID).Delete(&models.MediaStats{})

	logger.Info(fmt.Sprintf("Media file deleted successfully: %s", fileID))
	return nil
}

// SFU 架构：滤镜相关函数已删除
// 已删除的函数：
// - GetFilter: 获取滤镜（已删除）
// - ListFilters: 列出所有可用滤镜（已删除）
//
// 替代方案：
// 如需滤镜配置功能，应在用户服务中实现，通过信令传递给客户端
