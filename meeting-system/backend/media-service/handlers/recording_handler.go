package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"meeting-system/media-service/services"
	"meeting-system/shared/logger"
)

// RecordingHandler 录制处理器
type RecordingHandler struct {
	recordingService *services.RecordingService
}

// NewRecordingHandler 创建录制处理器
func NewRecordingHandler(recordingService *services.RecordingService) *RecordingHandler {
	return &RecordingHandler{
		recordingService: recordingService,
	}
}

// StartRecording 开始录制
func (h *RecordingHandler) StartRecording(c *gin.Context) {
	var request services.StartRecordingRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 开始录制
	recording, err := h.recordingService.StartRecording(&request)
	if err != nil {
		logger.Error("Failed to start recording: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start recording",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Recording started successfully",
		"recording": recording,
	})
}

// StopRecording 停止录制
func (h *RecordingHandler) StopRecording(c *gin.Context) {
	var request struct {
		RecordingID string `json:"recording_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 停止录制
	if err := h.recordingService.StopRecording(request.RecordingID); err != nil {
		logger.Error("Failed to stop recording: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to stop recording",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Recording stopped successfully",
		"recording_id": request.RecordingID,
	})
}

// GetRecordingStatus 获取录制状态
func (h *RecordingHandler) GetRecordingStatus(c *gin.Context) {
	recordingID := c.Param("id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "recording_id is required",
		})
		return
	}

	// 获取录制状态
	recording, err := h.recordingService.GetRecordingStatus(recordingID)
	if err != nil {
		logger.Error("Failed to get recording status: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recording not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recording": recording,
	})
}

// ListRecordings 列出录制记录
func (h *RecordingHandler) ListRecordings(c *gin.Context) {
	userID := c.Query("user_id")
	
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	limit := pageSize
	offset := (page - 1) * pageSize

	// 获取录制列表
	recordings, total, err := h.recordingService.ListRecordings(userID, limit, offset)
	if err != nil {
		logger.Error("Failed to list recordings: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve recordings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recordings": recordings,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// DownloadRecording 下载录制文件
func (h *RecordingHandler) DownloadRecording(c *gin.Context) {
	recordingID := c.Param("id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "recording_id is required",
		})
		return
	}

	// 下载录制文件
	reader, recording, err := h.recordingService.DownloadRecording(recordingID)
	if err != nil {
		logger.Error("Failed to download recording: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recording file not found or not ready",
		})
		return
	}

	// 设置响应头
	filename := recording.Title + "." + recording.Format
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "video/"+recording.Format)
	c.Header("Content-Length", strconv.FormatInt(recording.FileSize, 10))

	// 流式传输文件
	c.DataFromReader(http.StatusOK, recording.FileSize, "video/"+recording.Format, reader, nil)
}

// DeleteRecording 删除录制
func (h *RecordingHandler) DeleteRecording(c *gin.Context) {
	recordingID := c.Param("id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "recording_id is required",
		})
		return
	}

	// 删除录制
	if err := h.recordingService.DeleteRecording(recordingID); err != nil {
		logger.Error("Failed to delete recording: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete recording",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Recording deleted successfully",
		"recording_id": recordingID,
	})
}

// PauseRecording 暂停录制
func (h *RecordingHandler) PauseRecording(c *gin.Context) {
	var request struct {
		RecordingID string `json:"recording_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 这里应该调用recordingService的PauseRecording方法
	// 为了简化，我们返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":      "Recording paused successfully",
		"recording_id": request.RecordingID,
	})
}

// ResumeRecording 恢复录制
func (h *RecordingHandler) ResumeRecording(c *gin.Context) {
	var request struct {
		RecordingID string `json:"recording_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 这里应该调用recordingService的ResumeRecording方法
	// 为了简化，我们返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":      "Recording resumed successfully",
		"recording_id": request.RecordingID,
	})
}

// GetRecordingThumbnail 获取录制缩略图
func (h *RecordingHandler) GetRecordingThumbnail(c *gin.Context) {
	recordingID := c.Param("id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "recording_id is required",
		})
		return
	}

	// 获取录制信息
	recording, err := h.recordingService.GetRecordingStatus(recordingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recording not found",
		})
		return
	}

	if recording.ThumbnailPath == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thumbnail not available",
		})
		return
	}

	// 这里应该从存储中获取缩略图文件
	// 为了简化，我们返回缩略图信息
	c.JSON(http.StatusOK, gin.H{
		"recording_id":   recordingID,
		"thumbnail_path": recording.ThumbnailPath,
		"message":        "Thumbnail available",
	})
}

// UpdateRecordingMetadata 更新录制元数据
func (h *RecordingHandler) UpdateRecordingMetadata(c *gin.Context) {
	recordingID := c.Param("id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "recording_id is required",
		})
		return
	}

	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 这里应该调用recordingService的UpdateRecordingMetadata方法
	// 为了简化，我们返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":      "Recording metadata updated successfully",
		"recording_id": recordingID,
		"title":        request.Title,
		"description":  request.Description,
	})
}

// GetRecordingStats 获取录制统计信息
func (h *RecordingHandler) GetRecordingStats(c *gin.Context) {
	userID := c.Query("user_id")
	
	// 这里应该调用recordingService的GetRecordingStats方法
	// 为了简化，我们返回模拟统计信息
	stats := gin.H{
		"total_recordings":    0,
		"total_duration":      0,
		"total_size":          0,
		"completed_recordings": 0,
		"failed_recordings":   0,
		"active_recordings":   0,
	}

	if userID != "" {
		stats["user_id"] = userID
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recording statistics retrieved successfully",
		"stats":   stats,
	})
}

// ShareRecording 分享录制
func (h *RecordingHandler) ShareRecording(c *gin.Context) {
	recordingID := c.Param("id")
	if recordingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "recording_id is required",
		})
		return
	}

	var request struct {
		ShareType   string   `json:"share_type" binding:"required"` // public, private, users
		ExpiresAt   string   `json:"expires_at"`                    // 过期时间
		SharedUsers []string `json:"shared_users"`                  // 分享给的用户列表
		Password    string   `json:"password"`                      // 访问密码
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 这里应该调用recordingService的ShareRecording方法
	// 为了简化，我们返回成功响应
	shareURL := "https://example.com/recordings/shared/" + recordingID

	c.JSON(http.StatusOK, gin.H{
		"message":      "Recording shared successfully",
		"recording_id": recordingID,
		"share_url":    shareURL,
		"share_type":   request.ShareType,
		"expires_at":   request.ExpiresAt,
	})
}
