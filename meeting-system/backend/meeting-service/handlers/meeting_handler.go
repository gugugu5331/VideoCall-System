package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"meeting-system/shared/logger"
	"meeting-system/shared/models"
	"meeting-system/meeting-service/services"
)

type MeetingHandler struct {
	meetingService *services.MeetingService
}

func NewMeetingHandler(meetingService *services.MeetingService) *MeetingHandler {
	return &MeetingHandler{
		meetingService: meetingService,
	}
}

// CreateMeeting 创建会议
func (h *MeetingHandler) CreateMeeting(c *gin.Context) {
	var req models.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	req.CreatorID = userID.(uint)

	// 验证时间
	if req.StartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time cannot be in the past"})
		return
	}

	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}

	meeting, err := h.meetingService.CreateMeeting(&req)
	if err != nil {
		logger.Error("Failed to create meeting", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meeting"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Meeting created successfully",
		"data":    meeting,
	})
}

// GetMeeting 获取会议信息
func (h *MeetingHandler) GetMeeting(c *gin.Context) {
	meetingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	userID, _ := c.Get("user_id")

	meeting, err := h.meetingService.GetMeeting(uint(meetingID), userID.(uint))
	if err != nil {
		if err.Error() == "meeting not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Meeting not found"})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		logger.Error("Failed to get meeting", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get meeting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": meeting,
	})
}

// UpdateMeeting 更新会议
func (h *MeetingHandler) UpdateMeeting(c *gin.Context) {
	meetingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	var req models.UpdateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	meeting, err := h.meetingService.UpdateMeeting(uint(meetingID), userID.(uint), &req)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}
		logger.Error("Failed to update meeting", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update meeting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Meeting updated successfully",
		"data":    meeting,
	})
}

// DeleteMeeting 删除会议
func (h *MeetingHandler) DeleteMeeting(c *gin.Context) {
	meetingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	userID, _ := c.Get("user_id")

	err = h.meetingService.DeleteMeeting(uint(meetingID), userID.(uint))
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}
		logger.Error("Failed to delete meeting", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete meeting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Meeting deleted successfully",
	})
}

// JoinMeeting 加入会议
func (h *MeetingHandler) JoinMeeting(c *gin.Context) {
	meetingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")

	response, err := h.meetingService.JoinMeeting(uint(meetingID), userID.(uint), req.Password)
	if err != nil {
		if err.Error() == "meeting is not available" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Meeting is not available"})
			return
		}
		if err.Error() == "invalid password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		logger.Error("Failed to join meeting", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join meeting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Joined meeting successfully",
		"data":    response,
	})
}

// LeaveMeeting 离开会议
func (h *MeetingHandler) LeaveMeeting(c *gin.Context) {
	meetingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	userID, _ := c.Get("user_id")

	err = h.meetingService.LeaveMeeting(uint(meetingID), userID.(uint))
	if err != nil {
		logger.Error("Failed to leave meeting", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave meeting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Left meeting successfully",
	})
}

// GetParticipants 获取会议参与者
func (h *MeetingHandler) GetParticipants(c *gin.Context) {
	meetingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meeting ID"})
		return
	}

	userID, _ := c.Get("user_id")

	participants, err := h.meetingService.GetParticipants(uint(meetingID), userID.(uint))
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		logger.Error("Failed to get participants", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get participants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": participants,
	})
}

// ListMeetings 获取会议列表
func (h *MeetingHandler) ListMeetings(c *gin.Context) {
	// TODO: 实现会议列表查询
	c.JSON(http.StatusOK, gin.H{
		"message": "List meetings - TODO",
		"data":    []interface{}{},
	})
}

// StartMeeting 开始会议
func (h *MeetingHandler) StartMeeting(c *gin.Context) {
	// TODO: 实现开始会议逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Start meeting - TODO",
	})
}

// EndMeeting 结束会议
func (h *MeetingHandler) EndMeeting(c *gin.Context) {
	// TODO: 实现结束会议逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "End meeting - TODO",
	})
}

// AddParticipant 添加参与者
func (h *MeetingHandler) AddParticipant(c *gin.Context) {
	// TODO: 实现添加参与者逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Add participant - TODO",
	})
}

// RemoveParticipant 移除参与者
func (h *MeetingHandler) RemoveParticipant(c *gin.Context) {
	// TODO: 实现移除参与者逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Remove participant - TODO",
	})
}

// UpdateParticipantRole 更新参与者角色
func (h *MeetingHandler) UpdateParticipantRole(c *gin.Context) {
	// TODO: 实现更新参与者角色逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Update participant role - TODO",
	})
}

// GetMeetingRoom 获取会议室信息
func (h *MeetingHandler) GetMeetingRoom(c *gin.Context) {
	// TODO: 实现获取会议室信息逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Get meeting room - TODO",
	})
}

// CreateMeetingRoom 创建会议室
func (h *MeetingHandler) CreateMeetingRoom(c *gin.Context) {
	// TODO: 实现创建会议室逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Create meeting room - TODO",
	})
}

// CloseMeetingRoom 关闭会议室
func (h *MeetingHandler) CloseMeetingRoom(c *gin.Context) {
	// TODO: 实现关闭会议室逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Close meeting room - TODO",
	})
}

// StartRecording 开始录制
func (h *MeetingHandler) StartRecording(c *gin.Context) {
	// TODO: 实现开始录制逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Start recording - TODO",
	})
}

// StopRecording 停止录制
func (h *MeetingHandler) StopRecording(c *gin.Context) {
	// TODO: 实现停止录制逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Stop recording - TODO",
	})
}

// GetRecordings 获取录制列表
func (h *MeetingHandler) GetRecordings(c *gin.Context) {
	// TODO: 实现获取录制列表逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Get recordings - TODO",
		"data":    []interface{}{},
	})
}

// GetChatMessages 获取聊天消息
func (h *MeetingHandler) GetChatMessages(c *gin.Context) {
	// TODO: 实现获取聊天消息逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Get chat messages - TODO",
		"data":    []interface{}{},
	})
}

// SendChatMessage 发送聊天消息
func (h *MeetingHandler) SendChatMessage(c *gin.Context) {
	// TODO: 实现发送聊天消息逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Send chat message - TODO",
	})
}

// GetMyMeetings 获取我的会议
func (h *MeetingHandler) GetMyMeetings(c *gin.Context) {
	// TODO: 实现获取我的会议逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Get my meetings - TODO",
		"data":    []interface{}{},
	})
}

// GetUpcomingMeetings 获取即将开始的会议
func (h *MeetingHandler) GetUpcomingMeetings(c *gin.Context) {
	// TODO: 实现获取即将开始的会议逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Get upcoming meetings - TODO",
		"data":    []interface{}{},
	})
}

// GetMeetingHistory 获取会议历史
func (h *MeetingHandler) GetMeetingHistory(c *gin.Context) {
	// TODO: 实现获取会议历史逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Get meeting history - TODO",
		"data":    []interface{}{},
	})
}

// AdminListMeetings 管理员获取会议列表
func (h *MeetingHandler) AdminListMeetings(c *gin.Context) {
	// TODO: 实现管理员获取会议列表逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Admin list meetings - TODO",
		"data":    []interface{}{},
	})
}

// GetMeetingStats 获取会议统计
func (h *MeetingHandler) GetMeetingStats(c *gin.Context) {
	// TODO: 实现获取会议统计逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Get meeting stats - TODO",
		"data":    gin.H{},
	})
}

// ForceEndMeeting 强制结束会议
func (h *MeetingHandler) ForceEndMeeting(c *gin.Context) {
	// TODO: 实现强制结束会议逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Force end meeting - TODO",
	})
}
