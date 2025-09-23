package handlers

import (
	"net/http"
	"strconv"
	"time"

	"videocall-backend/models"

	"github.com/gin-gonic/gin"
)

// StartCallRequest 开始通话请求
type StartCallRequest struct {
	CalleeID string `json:"callee_id" binding:"required"`
	CallType string `json:"call_type" binding:"required,oneof=audio video"`
}

// StartCall 开始通话
func StartCall(c *gin.Context) {
	var req StartCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	callerID, _ := c.Get("user_id")
	callerIDUint := callerID.(uint)

	// 查找被叫用户
	var callee models.User
	if err := DB.Where("uuid = ?", req.CalleeID).First(&callee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Callee not found",
		})
		return
	}

	// 创建通话记录
	call := models.Call{
		CallerID:   &callerIDUint,
		CalleeID:   &callee.ID,
		CallerUUID: &callee.UUID, // 临时使用callee的UUID
		CalleeUUID: &callee.UUID,
		CallType:   req.CallType,
		Status:     "initiated",
		StartTime:  &time.Time{},
	}

	if err := DB.Create(&call).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create call",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Call initiated",
		"call": gin.H{
			"id":       call.ID,
			"uuid":     call.UUID,
			"call_type": call.CallType,
			"status":   call.Status,
			"callee": gin.H{
				"id":       callee.ID,
				"username": callee.Username,
				"full_name": callee.FullName,
			},
		},
	})
}

// EndCall 结束通话
func EndCall(c *gin.Context) {
	callID := c.Query("call_id")
	if callID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Call ID is required",
		})
		return
	}

	var call models.Call
	if err := DB.First(&call, callID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Call not found",
		})
		return
	}

	// 检查权限
	userID, _ := c.Get("user_id")
	if call.CallerID != nil && *call.CallerID != userID.(uint) && 
	   call.CalleeID != nil && *call.CalleeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to end this call",
		})
		return
	}

	// 更新通话状态
	now := time.Now()
	duration := int(now.Sub(*call.StartTime).Seconds())
	
	updates := map[string]interface{}{
		"status":    "ended",
		"end_time":  &now,
		"duration":  &duration,
	}

	if err := DB.Model(&call).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to end call",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Call ended successfully",
		"call": gin.H{
			"id":       call.ID,
			"duration": duration,
			"status":   "ended",
		},
	})
}

// GetCallHistory 获取通话历史
func GetCallHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var calls []models.Call
	var total int64

	// 获取总数
	DB.Model(&models.Call{}).Where("caller_id = ? OR callee_id = ?", userID, userID).Count(&total)

	// 获取通话记录
	if err := DB.Preload("Caller").Preload("Callee").
		Where("caller_id = ? OR callee_id = ?", userID, userID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&calls).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get call history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"calls": calls,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetCallDetails 获取通话详情
func GetCallDetails(c *gin.Context) {
	callID := c.Param("id")
	userID, _ := c.Get("user_id")

	var call models.Call
	if err := DB.Preload("Caller").Preload("Callee").
		First(&call, callID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Call not found",
		})
		return
	}

	// 检查权限
	if call.CallerID != nil && *call.CallerID != userID.(uint) && 
	   call.CalleeID != nil && *call.CalleeID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to view this call",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"call": call,
	})
}

// WebSocketHandler WebSocket处理器
func WebSocketHandler(c *gin.Context) {
	// WebSocket处理逻辑将在后续实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "WebSocket not implemented yet",
	})
} 