package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"meeting-system/media-service/services"
	"meeting-system/shared/response"
)

// AIHandler AI处理Handler
type AIHandler struct {
	mediaProcessor *services.MediaProcessor
	aiClient       *services.AIClient
}

// NewAIHandler 创建AI Handler
func NewAIHandler(mediaProcessor *services.MediaProcessor, aiClient *services.AIClient) *AIHandler {
	return &AIHandler{
		mediaProcessor: mediaProcessor,
		aiClient:       aiClient,
	}
}

// CheckAIConnectivity 检测与 AI 服务之间的通信状态
func (h *AIHandler) CheckAIConnectivity(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	status := h.mediaProcessor.CheckAIConnectivity(ctx)
	response.Success(c, gin.H{
		"timestamp":       status.Timestamp,
		"overall_healthy": status.OverallHealthy,
		"grpc":            status.GRPC,
		"http":            status.HTTP,
	})
}

// GetStreamStatus 获取流处理状态
func (h *AIHandler) GetStreamStatus(c *gin.Context) {
	streamID := c.Param("stream_id")
	if streamID == "" {
		response.Error(c, http.StatusBadRequest, "Stream ID is required")
		return
	}

	stream, exists := h.mediaProcessor.GetStreamStatus(streamID)
	if !exists {
		response.Error(c, http.StatusNotFound, "Stream not found")
		return
	}

	response.Success(c, gin.H{
		"stream_id":      stream.StreamID,
		"user_id":        stream.UserID,
		"room_id":        stream.RoomID,
		"is_active":      stream.IsActive,
		"last_processed": stream.LastProcessed,
		"ai_tasks":       stream.AITasks,
		"audio_buffer":   stream.AudioBuffer.Count(),
		"video_buffer":   stream.VideoBuffer.Count(),
	})
}

// ListActiveStreams 列出所有活跃流
func (h *AIHandler) ListActiveStreams(c *gin.Context) {
	streams := h.mediaProcessor.GetAllStreams()

	streamList := make([]gin.H, 0, len(streams))
	for _, stream := range streams {
		streamList = append(streamList, gin.H{
			"stream_id":      stream.StreamID,
			"user_id":        stream.UserID,
			"room_id":        stream.RoomID,
			"is_active":      stream.IsActive,
			"last_processed": stream.LastProcessed,
			"ai_tasks":       stream.AITasks,
		})
	}

	response.Success(c, gin.H{
		"streams": streamList,
		"count":   len(streamList),
	})
}

// SFU 架构：服务端AI处理方法已删除
// 原因：SFU架构中，客户端应直接调用AI服务接口，服务端仅负责RTP转发
// 已删除的方法：
// - ProcessAudioRequest: 服务端音频AI处理（已删除）
// - ProcessVideoRequest: 服务端视频AI处理（已删除）
// - ProcessMultimodalRequest: 服务端多模态AI处理（已删除）
//
// 替代方案：
// - 客户端直接调用AI服务的HTTP/gRPC接口
// - 媒体服务仅提供流状态查询和管理功能

// SFU 架构：流AI管理方法已删除
// 原因：客户端应直接管理AI处理，服务端不应主动控制
// 已删除的方法：
// - EnableStreamAI: 启用流AI处理（已删除）
// - DisableStreamAI: 禁用流AI处理（已删除）

// GetAIProcessingStats 获取AI处理统计信息
func (h *AIHandler) GetAIProcessingStats(c *gin.Context) {
	streams := h.mediaProcessor.GetAllStreams()

	stats := gin.H{
		"total_streams":  len(streams),
		"active_streams": 0,
		"total_tasks":    0,
		"tasks_by_type":  make(map[string]int),
	}

	for _, stream := range streams {
		if stream.IsActive {
			stats["active_streams"] = stats["active_streams"].(int) + 1
		}
		stats["total_tasks"] = stats["total_tasks"].(int) + len(stream.AITasks)

		for _, task := range stream.AITasks {
			count := stats["tasks_by_type"].(map[string]int)[task]
			stats["tasks_by_type"].(map[string]int)[task] = count + 1
		}
	}

	response.Success(c, stats)
}
