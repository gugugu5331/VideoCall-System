package handlers

import (
	"net/http"
	"strconv"
	"time"

	"videocall-backend/models"

	"github.com/gin-gonic/gin"
)

// TriggerDetectionRequest 触发检测请求
type TriggerDetectionRequest struct {
	CallID        string `json:"call_id" binding:"required"`
	DetectionType string `json:"detection_type" binding:"required,oneof=voice_spoofing video_deepfake face_swap"`
}

// TriggerDetection 触发安全检测
func TriggerDetection(c *gin.Context) {
	var req TriggerDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	userID, _ := c.Get("user_id")

	// 查找通话记录
	var call models.Call
	if err := DB.First(&call, req.CallID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Call not found",
		})
		return
	}

	// 检查权限
	if call.CallerID != nil && *call.CallerID != userID.(uint) && 
	   call.CalleeID != nil && *call.CalleeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to access this call",
		})
		return
	}

	// 创建检测记录
	detection := models.SecurityDetection{
		CallID:        call.ID,
		CallUUID:      call.UUID,
		DetectionType: req.DetectionType,
		RiskScore:     0.0, // 初始值，将由AI服务更新
		Confidence:    0.0, // 初始值，将由AI服务更新
		DetectionTime: time.Now(),
		ModelVersion:  "v1.0.0", // 默认版本
		Details:       models.JSON{"status": "pending"},
	}

	if err := DB.Create(&detection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create detection record",
		})
		return
	}

	// TODO: 异步调用AI服务进行检测
	// 这里应该发送消息到消息队列，由AI服务处理

	c.JSON(http.StatusOK, gin.H{
		"message": "Detection triggered successfully",
		"detection": gin.H{
			"id":             detection.ID,
			"uuid":           detection.UUID,
			"detection_type": detection.DetectionType,
			"status":         "pending",
		},
	})
}

// GetDetectionStatus 获取检测状态
func GetDetectionStatus(c *gin.Context) {
	callID := c.Param("callId")
	userID, _ := c.Get("user_id")

	// 查找通话记录
	var call models.Call
	if err := DB.First(&call, callID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Call not found",
		})
		return
	}

	// 检查权限
	if call.CallerID != nil && *call.CallerID != userID.(uint) && 
	   call.CalleeID != nil && *call.CalleeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to access this call",
		})
		return
	}

	// 获取最新的检测记录
	var detections []models.SecurityDetection
	if err := DB.Where("call_id = ?", call.ID).
		Order("created_at DESC").
		Find(&detections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get detection status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"call_id":     callID,
		"detections":  detections,
		"total_count": len(detections),
	})
}

// GetDetectionHistory 获取检测历史
func GetDetectionHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var detections []models.SecurityDetection
	var total int64

	// 获取用户相关的通话ID
	var callIDs []uint
	DB.Model(&models.Call{}).
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Pluck("id", &callIDs)

	if len(callIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"detections": []interface{}{},
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": 0,
			},
		})
		return
	}

	// 获取总数
	DB.Model(&models.SecurityDetection{}).
		Where("call_id IN ?", callIDs).
		Count(&total)

	// 获取检测记录
	if err := DB.Preload("Call").
		Where("call_id IN ?", callIDs).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&detections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get detection history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"detections": detections,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
} 