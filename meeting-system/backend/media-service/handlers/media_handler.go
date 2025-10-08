package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"meeting-system/media-service/services"
	"meeting-system/shared/logger"
)

// MediaHandler 媒体处理器
type MediaHandler struct {
	mediaService *services.MediaService
}

// NewMediaHandler 创建媒体处理器
func NewMediaHandler(mediaService *services.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// UploadMedia 上传媒体文件
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	// 获取用户ID和会议ID
	userID := c.PostForm("user_id")
	meetingID := c.PostForm("meeting_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		logger.Error("Failed to get uploaded file: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get uploaded file",
		})
		return
	}

	// 检查文件大小（限制为100MB）
	maxSize := int64(100 * 1024 * 1024) // 100MB
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File size exceeds 100MB limit",
		})
		return
	}

	// 上传文件
	mediaFile, err := h.mediaService.UploadMedia(file, userID, meetingID)
	if err != nil {
		logger.Error("Failed to upload media: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload media file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Media file uploaded successfully",
		"media_file": mediaFile,
	})
}

// DownloadMedia 下载媒体文件
func (h *MediaHandler) DownloadMedia(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_id is required",
		})
		return
	}

	// 下载文件
	reader, mediaFile, err := h.mediaService.DownloadMedia(fileID)
	if err != nil {
		logger.Error("Failed to download media: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Media file not found",
		})
		return
	}

	// 设置响应头
	c.Header("Content-Disposition", "attachment; filename="+mediaFile.OriginalName)
	c.Header("Content-Type", mediaFile.MimeType)
	c.Header("Content-Length", strconv.FormatInt(mediaFile.FileSize, 10))

	// 流式传输文件
	c.DataFromReader(http.StatusOK, mediaFile.FileSize, mediaFile.MimeType, reader, nil)
}

// ProcessMedia 处理媒体文件
func (h *MediaHandler) ProcessMedia(c *gin.Context) {
	var request struct {
		FileID    string `json:"file_id" binding:"required"`
		Operation string `json:"operation" binding:"required"` // transcode, extract_audio, extract_video, thumbnail
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 获取媒体文件
	mediaFile, err := h.mediaService.GetMediaFile(request.FileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Media file not found",
		})
		return
	}

	// 根据操作类型处理
	var jobID string
	switch request.Operation {
	case "transcode":
		// 转码操作需要通过FFmpeg服务处理
		c.JSON(http.StatusOK, gin.H{
			"message": "Transcode operation should be handled by FFmpeg service",
			"file_id": request.FileID,
		})
		return
	case "thumbnail":
		// 生成缩略图
		c.JSON(http.StatusOK, gin.H{
			"message": "Thumbnail generation should be handled by FFmpeg service",
			"file_id": request.FileID,
		})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported operation: " + request.Operation,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Processing started",
		"job_id":  jobID,
		"file_id": mediaFile.FileID,
	})
}

// GetMediaInfo 获取媒体文件信息
func (h *MediaHandler) GetMediaInfo(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_id is required",
		})
		return
	}

	// 获取媒体文件信息
	mediaFile, err := h.mediaService.GetMediaFile(fileID)
	if err != nil {
		logger.Error("Failed to get media info: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Media file not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"media_file": mediaFile,
	})
}

// DeleteMedia 删除媒体文件
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_id is required",
		})
		return
	}

	// 删除媒体文件
	if err := h.mediaService.DeleteMedia(fileID); err != nil {
		logger.Error("Failed to delete media: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete media file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Media file deleted successfully",
	})
}

// ListMedia 列出媒体文件
func (h *MediaHandler) ListMedia(c *gin.Context) {
	userID := c.Query("user_id")
	meetingID := c.Query("meeting_id")
	fileType := c.Query("file_type")
	
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize

	// 这里应该调用mediaService的ListMedia方法
	// 为了简化，我们返回一个模拟响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Media list retrieved successfully",
		"data": gin.H{
			"media_files": []interface{}{},
			"total":       0,
			"page":        page,
			"page_size":   pageSize,
		},
		"filters": gin.H{
			"user_id":   userID,
			"meeting_id": meetingID,
			"file_type": fileType,
		},
		"offset": offset,
	})
}

// GetMediaStats 获取媒体统计信息
func (h *MediaHandler) GetMediaStats(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_id is required",
		})
		return
	}

	// 这里应该调用mediaService的GetMediaStats方法
	// 为了简化，我们返回一个模拟响应
	c.JSON(http.StatusOK, gin.H{
		"file_id": fileID,
		"stats": gin.H{
			"view_count":     0,
			"download_count": 0,
			"share_count":    0,
			"last_viewed_at": nil,
		},
	})
}

// UpdateMediaMetadata 更新媒体元数据
func (h *MediaHandler) UpdateMediaMetadata(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_id is required",
		})
		return
	}

	var request struct {
		Title       string            `json:"title"`
		Description string            `json:"description"`
		Tags        []string          `json:"tags"`
		Properties  map[string]string `json:"properties"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 获取媒体文件
	mediaFile, err := h.mediaService.GetMediaFile(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Media file not found",
		})
		return
	}

	// 这里应该调用mediaService的UpdateMediaMetadata方法
	// 为了简化，我们返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":    "Media metadata updated successfully",
		"file_id":    fileID,
		"media_file": mediaFile,
		"metadata": gin.H{
			"title":       request.Title,
			"description": request.Description,
			"tags":        request.Tags,
			"properties":  request.Properties,
		},
	})
}

// SearchMedia 搜索媒体文件
func (h *MediaHandler) SearchMedia(c *gin.Context) {
	query := c.Query("q")
	userID := c.Query("user_id")
	fileType := c.Query("file_type")
	
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	// 这里应该调用mediaService的SearchMedia方法
	// 为了简化，我们返回一个模拟响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Media search completed",
		"query":   query,
		"data": gin.H{
			"media_files": []interface{}{},
			"total":       0,
			"page":        page,
			"page_size":   pageSize,
		},
		"filters": gin.H{
			"user_id":   userID,
			"file_type": fileType,
		},
	})
}

// GetMediaThumbnail 获取媒体缩略图
func (h *MediaHandler) GetMediaThumbnail(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file_id is required",
		})
		return
	}

	// 获取媒体文件信息
	mediaFile, err := h.mediaService.GetMediaFile(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Media file not found",
		})
		return
	}

	if mediaFile.ThumbnailPath == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thumbnail not available",
		})
		return
	}

	// 这里应该从存储中获取缩略图文件
	// 为了简化，我们返回缩略图信息
	c.JSON(http.StatusOK, gin.H{
		"file_id":        fileID,
		"thumbnail_path": mediaFile.ThumbnailPath,
		"message":        "Thumbnail available",
	})
}
