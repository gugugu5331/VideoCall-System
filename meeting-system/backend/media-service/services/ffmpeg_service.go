package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"meeting-system/media-service/models"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// FFmpegService FFmpeg处理服务
type FFmpegService struct {
	config          *config.Config
	mediaService    *MediaService
	signalingClient *SignalingClient
	jobs            map[string]*ProcessingJob
	jobsMux         sync.RWMutex
	workerPool      chan struct{}
	stopCh          chan struct{}
}

// ProcessingJob 处理任务
type ProcessingJob struct {
	ID         string
	MediaFile  *models.MediaFile
	JobType    string
	Status     string
	Progress   float64
	InputPath  string
	OutputPath string
	Parameters models.JobParameters
	Command    *exec.Cmd
	StartTime  time.Time
	Error      string
}

// TranscodeRequest 转码请求
type TranscodeRequest struct {
	FileID     string `json:"file_id" binding:"required"`
	Format     string `json:"format" binding:"required"`
	Quality    string `json:"quality"`
	Resolution string `json:"resolution"`
	Bitrate    string `json:"bitrate"`
	FrameRate  string `json:"frame_rate"`
}

// FilterRequest 滤镜请求
type FilterRequest struct {
	FileID       string   `json:"file_id" binding:"required"`
	Filters      []string `json:"filters" binding:"required"`
	OutputFormat string   `json:"output_format"`
}

// NewFFmpegService 创建FFmpeg服务
func NewFFmpegService(config *config.Config, mediaService *MediaService, signalingClient *SignalingClient) *FFmpegService {
	return &FFmpegService{
		config:          config,
		mediaService:    mediaService,
		signalingClient: signalingClient,
		jobs:            make(map[string]*ProcessingJob),
		workerPool:      make(chan struct{}, 5), // 最多5个并发任务
		stopCh:          make(chan struct{}),
	}
}

// Initialize 初始化FFmpeg服务
func (s *FFmpegService) Initialize() error {
	// 检查FFmpeg是否可用
	if err := s.checkFFmpegAvailability(); err != nil {
		return fmt.Errorf("FFmpeg not available: %w", err)
	}

	// 启动工作池
	for i := 0; i < cap(s.workerPool); i++ {
		s.workerPool <- struct{}{}
	}

	// 启动清理任务
	go s.startCleanupTask()

	logger.Info("FFmpeg service initialized successfully")
	return nil
}

// Stop 停止FFmpeg服务
func (s *FFmpegService) Stop() {
	close(s.stopCh)

	// 停止所有正在运行的任务
	s.jobsMux.Lock()
	for _, job := range s.jobs {
		if job.Command != nil && job.Command.Process != nil {
			job.Command.Process.Kill()
		}
	}
	s.jobsMux.Unlock()

	logger.Info("FFmpeg service stopped")
}

// checkFFmpegAvailability 检查FFmpeg可用性
func (s *FFmpegService) checkFFmpegAvailability() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg command not found: %w", err)
	}
	return nil
}

// TranscodeMedia 转码媒体文件
// SFU 架构：此功能已禁用，因为 SFU 不应进行服务端转码
// 转码应该在客户端完成，或者使用专门的转码服务（非实时）
func (s *FFmpegService) TranscodeMedia(request *TranscodeRequest) (string, error) {
	return "", fmt.Errorf("transcode功能已禁用：SFU架构不支持服务端实时转码，请在客户端进行格式转换或使用离线转码服务")
}

// ExtractAudio 提取音频
// SFU 架构：此功能已禁用，因为 SFU 不应进行服务端媒体处理
func (s *FFmpegService) ExtractAudio(fileID, format string) (string, error) {
	return "", fmt.Errorf("音频提取功能已禁用：SFU架构不支持服务端媒体处理，请在客户端进行音频提取")
}

// ExtractVideo 提取视频（无音频）
// SFU 架构：此功能已禁用，因为 SFU 不应进行服务端媒体处理
func (s *FFmpegService) ExtractVideo(fileID, format string) (string, error) {
	return "", fmt.Errorf("视频提取功能已禁用：SFU架构不支持服务端媒体处理，请在客户端进行视频提取")
}

// GenerateThumbnail 生成缩略图
func (s *FFmpegService) GenerateThumbnail(fileID string, timestamp float64) (string, error) {
	mediaFile, err := s.mediaService.GetMediaFile(fileID)
	if err != nil {
		return "", fmt.Errorf("media file not found: %w", err)
	}

	if mediaFile.FileType != "video" {
		return "", fmt.Errorf("can only generate thumbnails from video files")
	}

	jobID := uuid.New().String()
	job := &ProcessingJob{
		ID:        jobID,
		MediaFile: mediaFile,
		JobType:   "thumbnail",
		Status:    "pending",
		Parameters: models.JobParameters{
			Format: "jpg",
			CustomArgs: map[string]string{
				"timestamp": fmt.Sprintf("%.2f", timestamp),
			},
		},
		StartTime: time.Now(),
	}

	job.InputPath = s.getLocalPath(mediaFile.StoragePath)
	job.OutputPath = s.generateOutputPath(mediaFile, "jpg", "thumbnail")

	s.jobsMux.Lock()
	s.jobs[jobID] = job
	s.jobsMux.Unlock()

	go s.saveJobToDB(job)
	go s.executeGenerateThumbnail(job)

	return jobID, nil
}

// ApplyFilters 应用滤镜
// SFU 架构：此功能已禁用，滤镜处理应在客户端完成
func (s *FFmpegService) ApplyFilters(request *FilterRequest) (string, error) {
	return "", fmt.Errorf("滤镜功能已禁用：SFU架构要求所有滤镜、美颜等视觉效果在客户端处理，服务端仅负责媒体流转发")
}

// GetJobStatus 获取任务状态
func (s *FFmpegService) GetJobStatus(jobID string) (*ProcessingJob, error) {
	s.jobsMux.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// SFU 架构：以下执行函数已删除，因为违反 SFU 原则
// - executeTranscode: 服务端转码（已删除）
// - executeExtractAudio: 音频提取（已删除）
// - executeExtractVideo: 视频提取（已删除）

// executeGenerateThumbnail 执行缩略图生成
func (s *FFmpegService) executeGenerateThumbnail(job *ProcessingJob) {
	<-s.workerPool
	defer func() { s.workerPool <- struct{}{} }()

	s.updateJobStatus(job.ID, "processing", 0)

	if err := s.downloadInputFile(job); err != nil {
		s.updateJobError(job.ID, fmt.Sprintf("failed to download input file: %v", err))
		return
	}

	timestamp := job.Parameters.CustomArgs["timestamp"]
	args := []string{
		"-i", job.InputPath,
		"-ss", timestamp, // 跳转到指定时间
		"-vframes", "1", // 只提取一帧
		"-q:v", "2", // 高质量
		"-y",
		job.OutputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	job.Command = cmd

	if err := cmd.Run(); err != nil {
		s.updateJobError(job.ID, fmt.Sprintf("ffmpeg execution failed: %v", err))
		return
	}

	if err := s.uploadOutputFile(job); err != nil {
		s.updateJobError(job.ID, fmt.Sprintf("failed to upload output file: %v", err))
		return
	}

	s.cleanupTempFiles(job)
	s.updateJobStatus(job.ID, "completed", 100)
}

// SFU 架构：以下函数已删除，因为违反 SFU 原则
// - executeApplyFilters: 滤镜应用（已删除）
// - buildTranscodeArgs: 转码参数构建（已删除）

// 辅助方法
func (s *FFmpegService) getLocalPath(storagePath string) string {
	return filepath.Join("/tmp", filepath.Base(storagePath))
}

func (s *FFmpegService) generateOutputPath(mediaFile *models.MediaFile, format, suffix string) string {
	ext := "." + format
	filename := fmt.Sprintf("%s_%s_%d%s", mediaFile.FileID, suffix, time.Now().Unix(), ext)
	return filepath.Join("/tmp", filename)
}

func (s *FFmpegService) downloadInputFile(job *ProcessingJob) error {
	// 从MinIO下载文件到本地临时目录
	reader, err := s.mediaService.storage.GetFile("media", job.MediaFile.StoragePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 创建本地文件
	file, err := os.Create(job.InputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 复制数据
	_, err = file.ReadFrom(reader)
	return err
}

func (s *FFmpegService) uploadOutputFile(job *ProcessingJob) error {
	// 打开输出文件
	file, err := os.Open(job.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// 上传到MinIO
	outputStoragePath := fmt.Sprintf("processed/%s", filepath.Base(job.OutputPath))
	return s.mediaService.storage.UploadFile("media", outputStoragePath, file, stat.Size(), "application/octet-stream")
}

func (s *FFmpegService) cleanupTempFiles(job *ProcessingJob) {
	os.Remove(job.InputPath)
	os.Remove(job.OutputPath)
}

func (s *FFmpegService) updateJobStatus(jobID, status string, progress float64) {
	s.jobsMux.Lock()
	if job, exists := s.jobs[jobID]; exists {
		job.Status = status
		job.Progress = progress
	}
	s.jobsMux.Unlock()

	// 更新数据库
	go s.updateJobInDB(jobID, status, progress, "")
}

func (s *FFmpegService) updateJobError(jobID, errorMsg string) {
	s.jobsMux.Lock()
	if job, exists := s.jobs[jobID]; exists {
		job.Status = "failed"
		job.Error = errorMsg
	}
	s.jobsMux.Unlock()

	// 更新数据库
	go s.updateJobInDB(jobID, "failed", 0, errorMsg)
}

func (s *FFmpegService) saveJobToDB(job *ProcessingJob) {
	dbJob := &models.ProcessingJob{
		JobID:       job.ID,
		MediaFileID: job.MediaFile.FileID,
		JobType:     job.JobType,
		Status:      job.Status,
		Progress:    job.Progress,
		InputPath:   job.InputPath,
		OutputPath:  job.OutputPath,
		Parameters:  job.Parameters,
		StartedAt:   &job.StartTime,
		CreatedAt:   job.StartTime,
		UpdatedAt:   time.Now(),
	}

	if err := s.mediaService.db.Create(dbJob).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to save job to database: %v", err))
	}
}

func (s *FFmpegService) updateJobInDB(jobID, status string, progress float64, errorMsg string) {
	updates := map[string]interface{}{
		"status":     status,
		"progress":   progress,
		"updated_at": time.Now(),
	}

	if errorMsg != "" {
		updates["error"] = errorMsg
	}

	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	if err := s.mediaService.db.Model(&models.ProcessingJob{}).Where("job_id = ?", jobID).Updates(updates).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update job in database: %v", err))
	}
}

func (s *FFmpegService) startCleanupTask() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupCompletedJobs()
		case <-s.stopCh:
			return
		}
	}
}

func (s *FFmpegService) cleanupCompletedJobs() {
	cutoff := time.Now().Add(-24 * time.Hour)

	s.jobsMux.Lock()
	for jobID, job := range s.jobs {
		if (job.Status == "completed" || job.Status == "failed") && job.StartTime.Before(cutoff) {
			delete(s.jobs, jobID)
		}
	}
	s.jobsMux.Unlock()
}

// SFU 架构说明：
// 以下函数已被删除，因为它们违反了 SFU (Selective Forwarding Unit) 架构原则。
// SFU 应该仅进行媒体流的选择性转发，不应进行任何服务端编解码、转码或格式转换。
//
// 已删除的函数：
// - ConvertAudioToWAV: 音频格式转换（违反 SFU 原则）
// - ConvertVideoToMP4: 视频格式转换（违反 SFU 原则）
// - ConvertRTPToAudio: RTP 到音频转换（违反 SFU 原则）
// - ConvertRTPToVideo: RTP 到视频转换（违反 SFU 原则）
//
// 替代方案：
// - AI 分析服务应直接处理原始 RTP/PCM/H264 数据
// - 客户端负责所有媒体编解码和格式转换
// - 服务端仅负责 RTP 包的路由和转发

// ExtractAudioFromVideo 从视频中提取音频
func (s *FFmpegService) ExtractAudioFromVideo(videoData []byte) ([]byte, error) {
	// 创建临时文件
	tempDir := os.TempDir()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("input_%s.mp4", uuid.New().String()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("output_%s.wav", uuid.New().String()))

	defer func() {
		os.Remove(inputFile)
		os.Remove(outputFile)
	}()

	// 写入视频数据
	if err := os.WriteFile(inputFile, videoData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write video file: %v", err)
	}

	// 构建FFmpeg命令
	args := []string{
		"-i", inputFile,
		"-vn",           // 不处理视频
		"-acodec", "pcm_s16le", // 音频编码器
		"-ar", "16000",  // 采样率
		"-ac", "1",      // 单声道
		"-f", "wav",
		"-y",
		outputFile,
	}

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("audio extraction failed: %v", err)
	}

	// 读取输出文件
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %v", err)
	}

	return outputData, nil
}
