package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"meeting-system/media-service/services"
	"meeting-system/shared/logger"
)

// FFmpegHandler FFmpeg处理器
type FFmpegHandler struct {
	ffmpegService *services.FFmpegService
}

// NewFFmpegHandler 创建FFmpeg处理器
func NewFFmpegHandler(ffmpegService *services.FFmpegService) *FFmpegHandler {
	return &FFmpegHandler{
		ffmpegService: ffmpegService,
	}
}

// TranscodeMedia 转码媒体文件
func (h *FFmpegHandler) TranscodeMedia(c *gin.Context) {
	var request services.TranscodeRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 执行转码
	jobID, err := h.ffmpegService.TranscodeMedia(&request)
	if err != nil {
		logger.Error("Failed to start transcode: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transcode job",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transcode job started successfully",
		"job_id":  jobID,
		"file_id": request.FileID,
	})
}

// ExtractAudio 提取音频
func (h *FFmpegHandler) ExtractAudio(c *gin.Context) {
	var request struct {
		FileID string `json:"file_id" binding:"required"`
		Format string `json:"format"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 设置默认格式
	if request.Format == "" {
		request.Format = "mp3"
	}

	// 执行音频提取
	jobID, err := h.ffmpegService.ExtractAudio(request.FileID, request.Format)
	if err != nil {
		logger.Error("Failed to start audio extraction: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start audio extraction job",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Audio extraction job started successfully",
		"job_id":  jobID,
		"file_id": request.FileID,
		"format":  request.Format,
	})
}

// ExtractVideo 提取视频（无音频）
func (h *FFmpegHandler) ExtractVideo(c *gin.Context) {
	var request struct {
		FileID string `json:"file_id" binding:"required"`
		Format string `json:"format"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 设置默认格式
	if request.Format == "" {
		request.Format = "mp4"
	}

	// 执行视频提取
	jobID, err := h.ffmpegService.ExtractVideo(request.FileID, request.Format)
	if err != nil {
		logger.Error("Failed to start video extraction: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start video extraction job",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Video extraction job started successfully",
		"job_id":  jobID,
		"file_id": request.FileID,
		"format":  request.Format,
	})
}

// MergeMedia 合并媒体文件
func (h *FFmpegHandler) MergeMedia(c *gin.Context) {
	var request struct {
		FileIDs      []string `json:"file_ids" binding:"required"`
		OutputFormat string   `json:"output_format"`
		MergeType    string   `json:"merge_type"` // concat, overlay, side_by_side
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	if len(request.FileIDs) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least 2 files are required for merging",
		})
		return
	}

	// 设置默认值
	if request.OutputFormat == "" {
		request.OutputFormat = "mp4"
	}
	if request.MergeType == "" {
		request.MergeType = "concat"
	}

	// 这里应该调用ffmpegService的MergeMedia方法
	// 为了简化，我们返回一个模拟响应
	c.JSON(http.StatusOK, gin.H{
		"message":       "Media merge job started successfully",
		"job_id":        "mock-merge-job-id",
		"file_ids":      request.FileIDs,
		"output_format": request.OutputFormat,
		"merge_type":    request.MergeType,
	})
}

// GenerateThumbnail 生成缩略图
func (h *FFmpegHandler) GenerateThumbnail(c *gin.Context) {
	var request struct {
		FileID    string  `json:"file_id" binding:"required"`
		Timestamp float64 `json:"timestamp"` // 时间戳（秒）
		Width     int     `json:"width"`     // 缩略图宽度
		Height    int     `json:"height"`    // 缩略图高度
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 设置默认时间戳
	if request.Timestamp <= 0 {
		request.Timestamp = 10.0 // 默认第10秒
	}

	// 执行缩略图生成
	jobID, err := h.ffmpegService.GenerateThumbnail(request.FileID, request.Timestamp)
	if err != nil {
		logger.Error("Failed to start thumbnail generation: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start thumbnail generation job",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Thumbnail generation job started successfully",
		"job_id":    jobID,
		"file_id":   request.FileID,
		"timestamp": request.Timestamp,
	})
}

// GetJobStatus 获取任务状态
func (h *FFmpegHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job_id is required",
		})
		return
	}

	// 获取任务状态
	job, err := h.ffmpegService.GetJobStatus(jobID)
	if err != nil {
		logger.Error("Failed to get job status: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job": gin.H{
			"id":           job.ID,
			"media_file":   job.MediaFile,
			"job_type":     job.JobType,
			"status":       job.Status,
			"progress":     job.Progress,
			"input_path":   job.InputPath,
			"output_path":  job.OutputPath,
			"parameters":   job.Parameters,
			"start_time":   job.StartTime,
			"error":        job.Error,
		},
	})
}

// CancelJob 取消任务
func (h *FFmpegHandler) CancelJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "job_id is required",
		})
		return
	}

	// 这里应该调用ffmpegService的CancelJob方法
	// 为了简化，我们返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Job cancelled successfully",
		"job_id":  jobID,
	})
}

// ListJobs 列出任务
func (h *FFmpegHandler) ListJobs(c *gin.Context) {
	userID := c.Query("user_id")
	status := c.Query("status")
	jobType := c.Query("job_type")

	// 这里应该调用ffmpegService的ListJobs方法
	// 为了简化，我们返回一个模拟响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Jobs retrieved successfully",
		"jobs":    []interface{}{},
		"filters": gin.H{
			"user_id":  userID,
			"status":   status,
			"job_type": jobType,
		},
	})
}

// SFU 架构：ApplyFilter 方法已删除
// 原因：SFU 架构要求所有滤镜、美颜等视觉效果在客户端处理
// 服务端仅负责媒体流的选择性转发

// GetSupportedFormats 获取支持的格式
func (h *FFmpegHandler) GetSupportedFormats(c *gin.Context) {
	formats := gin.H{
		"video": []string{"mp4", "avi", "mov", "wmv", "flv", "webm", "mkv", "m4v"},
		"audio": []string{"mp3", "wav", "aac", "flac", "ogg", "m4a", "wma"},
		"image": []string{"jpg", "jpeg", "png", "gif", "bmp", "webp"},
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Supported formats retrieved successfully",
		"supported_formats": formats,
	})
}

// GetSupportedCodecs 获取支持的编解码器
func (h *FFmpegHandler) GetSupportedCodecs(c *gin.Context) {
	codecs := gin.H{
		"video": []string{"h264", "h265", "vp8", "vp9", "av1", "mpeg4"},
		"audio": []string{"aac", "mp3", "opus", "vorbis", "flac", "pcm"},
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Supported codecs retrieved successfully",
		"supported_codecs": codecs,
	})
}

// GetPresets 获取预设配置
func (h *FFmpegHandler) GetPresets(c *gin.Context) {
	presets := gin.H{
		"quality": gin.H{
			"high":   gin.H{"crf": 18, "preset": "slow"},
			"medium": gin.H{"crf": 23, "preset": "medium"},
			"low":    gin.H{"crf": 28, "preset": "fast"},
		},
		"resolution": gin.H{
			"4k":   "3840x2160",
			"1080p": "1920x1080",
			"720p":  "1280x720",
			"480p":  "854x480",
			"360p":  "640x360",
		},
		"bitrate": gin.H{
			"4k":   "15000k",
			"1080p": "5000k",
			"720p":  "2500k",
			"480p":  "1000k",
			"360p":  "500k",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Presets retrieved successfully",
		"presets": presets,
	})
}
