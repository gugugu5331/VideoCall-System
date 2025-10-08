package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"meeting-system/media-service/models"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// RecordingService 录制服务
type RecordingService struct {
	config          *config.Config
	mediaService    *MediaService
	ffmpegService   *FFmpegService
	signalingClient *SignalingClient
	recordings      map[string]*ActiveRecording
	recordingsMux   sync.RWMutex
	stopCh          chan struct{}
}

// ActiveRecording 活跃录制
type ActiveRecording struct {
	Recording    *models.Recording
	Status       string
	StartTime    time.Time
	OutputPath   string
	Process      *RecordingProcess
	Participants map[string]*ParticipantStream
}

// RecordingProcess 录制进程
type RecordingProcess struct {
	PID       int
	Command   string
	IsRunning bool
}

// ParticipantStream 参与者流
type ParticipantStream struct {
	UserID     string
	StreamType string // audio, video, screen
	Quality    string
	IsActive   bool
	LastUpdate time.Time
}

// StartRecordingRequest 开始录制请求
type StartRecordingRequest struct {
	MeetingID    string   `json:"meeting_id" binding:"required"`
	RoomID       string   `json:"room_id" binding:"required"`
	UserID       string   `json:"user_id" binding:"required"`
	Title        string   `json:"title" binding:"required"`
	Quality      string   `json:"quality"` // 720p, 1080p, 4k
	Format       string   `json:"format"`  // mp4, webm, mkv
	Participants []string `json:"participants"`
}

// NewRecordingService 创建录制服务
func NewRecordingService(config *config.Config, mediaService *MediaService, ffmpegService *FFmpegService, signalingClient *SignalingClient) *RecordingService {
	return &RecordingService{
		config:          config,
		mediaService:    mediaService,
		ffmpegService:   ffmpegService,
		signalingClient: signalingClient,
		recordings:      make(map[string]*ActiveRecording),
		stopCh:          make(chan struct{}),
	}
}

// Initialize 初始化录制服务
func (s *RecordingService) Initialize() error {
	// 创建录制目录
	recordingDir := "/tmp/recordings"
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return fmt.Errorf("failed to create recording directory: %w", err)
	}

	// 启动清理任务
	go s.startCleanupTask()

	// 恢复未完成的录制
	go s.recoverActiveRecordings()

	logger.Info("Recording service initialized successfully")
	return nil
}

// Stop 停止录制服务
func (s *RecordingService) Stop() {
	close(s.stopCh)

	// 停止所有活跃录制
	s.recordingsMux.Lock()
	for recordingID := range s.recordings {
		s.stopRecordingInternal(recordingID)
	}
	s.recordingsMux.Unlock()

	logger.Info("Recording service stopped")
}

// StartRecording 开始录制
func (s *RecordingService) StartRecording(request *StartRecordingRequest) (*models.Recording, error) {
	// 生成录制ID
	recordingID := uuid.New().String()

	// 设置默认值
	if request.Quality == "" {
		request.Quality = "720p"
	}
	if request.Format == "" {
		request.Format = "mp4"
	}

	// 创建录制记录
	recording := &models.Recording{
		RecordingID:  recordingID,
		MeetingID:    request.MeetingID,
		RoomID:       request.RoomID,
		UserID:       request.UserID,
		Title:        request.Title,
		Status:       "recording",
		StartTime:    time.Now(),
		Quality:      request.Quality,
		Format:       request.Format,
		Participants: request.Participants,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	if err := s.mediaService.db.Create(recording).Error; err != nil {
		return nil, fmt.Errorf("failed to save recording to database: %w", err)
	}

	// 生成输出路径
	outputPath := s.generateOutputPath(recording)

	// 创建活跃录制
	activeRecording := &ActiveRecording{
		Recording:    recording,
		Status:       "recording",
		StartTime:    time.Now(),
		OutputPath:   outputPath,
		Participants: make(map[string]*ParticipantStream),
	}

	// 初始化参与者流
	for _, userID := range request.Participants {
		activeRecording.Participants[userID] = &ParticipantStream{
			UserID:     userID,
			StreamType: "video", // 默认视频流
			Quality:    request.Quality,
			IsActive:   true,
			LastUpdate: time.Now(),
		}
	}

	// 保存活跃录制
	s.recordingsMux.Lock()
	s.recordings[recordingID] = activeRecording
	s.recordingsMux.Unlock()

	// 启动录制进程
	go s.startRecordingProcess(activeRecording)

	logger.Info(fmt.Sprintf("Recording started: %s", recordingID))
	return recording, nil
}

// StopRecording 停止录制
func (s *RecordingService) StopRecording(recordingID string) error {
	s.recordingsMux.Lock()
	if _, exists := s.recordings[recordingID]; !exists {
		s.recordingsMux.Unlock()
		return fmt.Errorf("recording not found: %s", recordingID)
	}
	s.recordingsMux.Unlock()

	return s.stopRecordingInternal(recordingID)
}

// GetRecordingStatus 获取录制状态
func (s *RecordingService) GetRecordingStatus(recordingID string) (*models.Recording, error) {
	// 先从活跃录制中查找
	s.recordingsMux.RLock()
	if activeRecording, exists := s.recordings[recordingID]; exists {
		s.recordingsMux.RUnlock()
		return activeRecording.Recording, nil
	}
	s.recordingsMux.RUnlock()

	// 从数据库中查找
	var recording models.Recording
	if err := s.mediaService.db.Where("recording_id = ?", recordingID).First(&recording).Error; err != nil {
		return nil, err
	}

	return &recording, nil
}

// ListRecordings 列出录制记录
func (s *RecordingService) ListRecordings(userID string, limit, offset int) ([]*models.Recording, int64, error) {
	var recordings []*models.Recording
	var total int64

	query := s.mediaService.db.Model(&models.Recording{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取记录
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&recordings).Error; err != nil {
		return nil, 0, err
	}

	return recordings, total, nil
}

// DownloadRecording 下载录制文件
func (s *RecordingService) DownloadRecording(recordingID string) (io.Reader, *models.Recording, error) {
	// 获取录制记录
	recording, err := s.GetRecordingStatus(recordingID)
	if err != nil {
		return nil, nil, err
	}

	if recording.Status != "completed" {
		return nil, nil, fmt.Errorf("recording is not completed yet")
	}

	// 没有配置存储时直接返回错误，避免空指针
	if s.mediaService.storage == nil {
		return nil, nil, fmt.Errorf("storage service not available")
	}

	// 从存储中获取文件
	reader, err := s.mediaService.storage.GetFile("recordings", recording.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get recording file: %w", err)
	}

	return reader, recording, nil
}

// DeleteRecording 删除录制
func (s *RecordingService) DeleteRecording(recordingID string) error {
	// 获取录制记录
	recording, err := s.GetRecordingStatus(recordingID)
	if err != nil {
		return err
	}

	// 如果正在录制，先停止
	if recording.Status == "recording" {
		if err := s.StopRecording(recordingID); err != nil {
			logger.Error(fmt.Sprintf("Failed to stop recording before deletion: %v", err))
		}
	}

	// 从存储中删除文件
	if s.mediaService.storage != nil {
		if recording.FilePath != "" {
			if err := s.mediaService.storage.DeleteFile("recordings", recording.FilePath); err != nil {
				logger.Error(fmt.Sprintf("Failed to delete recording file: %v", err))
			}
		}

		// 删除缩略图
		if recording.ThumbnailPath != "" {
			if err := s.mediaService.storage.DeleteFile("recordings", recording.ThumbnailPath); err != nil {
				logger.Error(fmt.Sprintf("Failed to delete recording thumbnail: %v", err))
			}
		}
	} else {
		logger.Warn("Storage service not configured; skipping recording artifact cleanup",
			logger.String("recording_id", recordingID))
	}

	// 从数据库中删除
	if err := s.mediaService.db.Delete(recording).Error; err != nil {
		return fmt.Errorf("failed to delete recording from database: %w", err)
	}

	logger.Info(fmt.Sprintf("Recording deleted: %s", recordingID))
	return nil
}

// startRecordingProcess 启动录制进程
func (s *RecordingService) startRecordingProcess(activeRecording *ActiveRecording) {
	recording := activeRecording.Recording

	// 更新状态为处理中
	s.updateRecordingStatus(recording.RecordingID, "processing")

	// 模拟录制过程（实际应该启动FFmpeg进程来录制WebRTC流）
	// 这里我们创建一个模拟的录制进程
	process := &RecordingProcess{
		PID:       os.Getpid(), // 模拟进程ID
		Command:   fmt.Sprintf("ffmpeg -f webrtc -i room:%s -c:v libx264 -c:a aac %s", recording.RoomID, activeRecording.OutputPath),
		IsRunning: true,
	}

	activeRecording.Process = process

	// 模拟录制时间（实际应该持续到停止录制）
	time.Sleep(5 * time.Second)

	// 模拟录制完成
	s.finishRecording(activeRecording)
}

// stopRecordingInternal 内部停止录制
func (s *RecordingService) stopRecordingInternal(recordingID string) error {
	s.recordingsMux.Lock()
	activeRecording, exists := s.recordings[recordingID]
	if !exists {
		s.recordingsMux.Unlock()
		return fmt.Errorf("active recording not found: %s", recordingID)
	}

	// 停止录制进程
	if activeRecording.Process != nil && activeRecording.Process.IsRunning {
		activeRecording.Process.IsRunning = false
		// 实际应该杀死FFmpeg进程
		// syscall.Kill(activeRecording.Process.PID, syscall.SIGTERM)
	}

	// 从活跃录制中移除
	delete(s.recordings, recordingID)
	s.recordingsMux.Unlock()

	// 完成录制处理
	go s.finishRecording(activeRecording)

	return nil
}

// finishRecording 完成录制
func (s *RecordingService) finishRecording(activeRecording *ActiveRecording) {
	recording := activeRecording.Recording
	endTime := time.Now()
	duration := endTime.Sub(activeRecording.StartTime).Seconds()

	if s.mediaService.storage == nil {
		logger.Warn("Storage service not configured; skipping recording upload",
			logger.String("recording_id", recording.RecordingID))

		updates := map[string]interface{}{
			"status":     "completed",
			"end_time":   &endTime,
			"duration":   duration,
			"file_size":  0,
			"file_path":  "",
			"updated_at": time.Now(),
		}

		if err := s.mediaService.db.Model(recording).Updates(updates).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to update recording in database: %v", err))
		}

		return
	}

	// 检查输出文件是否存在
	if _, err := os.Stat(activeRecording.OutputPath); os.IsNotExist(err) {
		// 创建一个模拟文件（实际应该是真实的录制文件）
		file, err := os.Create(activeRecording.OutputPath)
		if err != nil {
			s.updateRecordingStatus(recording.RecordingID, "failed")
			return
		}
		file.WriteString("Mock recording content")
		file.Close()
	}

	// 获取文件大小
	stat, err := os.Stat(activeRecording.OutputPath)
	if err != nil {
		s.updateRecordingStatus(recording.RecordingID, "failed")
		return
	}

	// 上传到存储
	file, err := os.Open(activeRecording.OutputPath)
	if err != nil {
		s.updateRecordingStatus(recording.RecordingID, "failed")
		return
	}
	defer file.Close()

	storagePath := fmt.Sprintf("recordings/%s/%s.%s", recording.UserID, recording.RecordingID, recording.Format)
	if err := s.mediaService.storage.UploadFile("recordings", storagePath, file, stat.Size(), "video/mp4"); err != nil {
		logger.Error(fmt.Sprintf("Failed to upload recording: %v", err))
		s.updateRecordingStatus(recording.RecordingID, "failed")
		return
	}

	// 更新数据库记录
	updates := map[string]interface{}{
		"status":     "completed",
		"end_time":   &endTime,
		"duration":   duration,
		"file_size":  stat.Size(),
		"file_path":  storagePath,
		"updated_at": time.Now(),
	}

	if err := s.mediaService.db.Model(recording).Updates(updates).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update recording in database: %v", err))
	}

	// 清理临时文件
	os.Remove(activeRecording.OutputPath)

	// 生成缩略图
	go s.generateRecordingThumbnail(recording.RecordingID)

	logger.Info(fmt.Sprintf("Recording completed: %s", recording.RecordingID))
}

// generateRecordingThumbnail 生成录制缩略图
func (s *RecordingService) generateRecordingThumbnail(recordingID string) {
	if s.ffmpegService == nil {
		return
	}

	// 使用FFmpeg服务生成缩略图
	if _, err := s.ffmpegService.GenerateThumbnail(recordingID, 10.0); err != nil {
		logger.Error(fmt.Sprintf("Failed to generate recording thumbnail: %v", err))
	}
}

// updateRecordingStatus 更新录制状态
func (s *RecordingService) updateRecordingStatus(recordingID, status string) {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if err := s.mediaService.db.Model(&models.Recording{}).Where("recording_id = ?", recordingID).Updates(updates).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update recording status: %v", err))
	}
}

// generateOutputPath 生成输出路径
func (s *RecordingService) generateOutputPath(recording *models.Recording) string {
	filename := fmt.Sprintf("%s_%d.%s", recording.RecordingID, time.Now().Unix(), recording.Format)
	return filepath.Join("/tmp/recordings", filename)
}

// recoverActiveRecordings 恢复活跃录制
func (s *RecordingService) recoverActiveRecordings() {
	var recordings []models.Recording
	if err := s.mediaService.db.Where("status = ?", "recording").Find(&recordings).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to recover active recordings: %v", err))
		return
	}

	for _, recording := range recordings {
		// 检查录制是否真的还在进行
		// 如果超过一定时间没有更新，标记为失败
		if time.Since(recording.UpdatedAt) > 1*time.Hour {
			s.updateRecordingStatus(recording.RecordingID, "failed")
		}
	}
}

// startCleanupTask 启动清理任务
func (s *RecordingService) startCleanupTask() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupOldRecordings()
		case <-s.stopCh:
			return
		}
	}
}

// cleanupOldRecordings 清理旧录制
func (s *RecordingService) cleanupOldRecordings() {
	// 清理超过30天的录制文件（可配置）
	cutoff := time.Now().AddDate(0, 0, -30)

	var oldRecordings []models.Recording
	if err := s.mediaService.db.Where("created_at < ? AND status = ?", cutoff, "completed").Find(&oldRecordings).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to find old recordings: %v", err))
		return
	}

	for _, recording := range oldRecordings {
		if err := s.DeleteRecording(recording.RecordingID); err != nil {
			logger.Error(fmt.Sprintf("Failed to cleanup old recording %s: %v", recording.RecordingID, err))
		}
	}

	if len(oldRecordings) > 0 {
		logger.Info(fmt.Sprintf("Cleaned up %d old recordings", len(oldRecordings)))
	}
}
