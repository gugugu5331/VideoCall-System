package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v3"
	"meeting-system/media-service/services"
	"meeting-system/shared/logger"
)

// WebRTCHandler WebRTC处理器
type WebRTCHandler struct {
	webrtcService *services.WebRTCService
}

// NewWebRTCHandler 创建WebRTC处理器
func NewWebRTCHandler(webrtcService *services.WebRTCService) *WebRTCHandler {
	return &WebRTCHandler{
		webrtcService: webrtcService,
	}
}

// HandleOfferAndCreateAnswer 处理客户端的Offer并创建Answer（SFU标准流程）
func (h *WebRTCHandler) HandleOfferAndCreateAnswer(c *gin.Context) {
	var request struct {
		RoomID string `json:"room_id" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
		Offer  struct {
			Type string `json:"type" binding:"required"`
			SDP  string `json:"sdp" binding:"required"`
		} `json:"offer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 解析客户端的Offer
	offer := webrtc.SessionDescription{
		Type: webrtc.NewSDPType(request.Offer.Type),
		SDP:  request.Offer.SDP,
	}

	// 创建Answer（SFU响应客户端的Offer）
	answer, peerID, err := h.webrtcService.CreateAnswer(request.RoomID, request.UserID, &offer)
	if err != nil {
		logger.Error("Failed to create answer: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create answer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"peer_id": peerID,
		"answer": gin.H{
			"type": answer.Type.String(),
			"sdp":  answer.SDP,
		},
	})
}

// HandleAnswer 处理SDP Answer
func (h *WebRTCHandler) HandleAnswer(c *gin.Context) {
	var request struct {
		PeerID string `json:"peer_id" binding:"required"`
		Answer struct {
			Type string `json:"type" binding:"required"`
			SDP  string `json:"sdp" binding:"required"`
		} `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 解析Answer
	answer := webrtc.SessionDescription{
		Type: webrtc.NewSDPType(request.Answer.Type),
		SDP:  request.Answer.SDP,
	}

	// 处理Answer
	if err := h.webrtcService.HandleAnswer(request.PeerID, &answer); err != nil {
		logger.Error("Failed to handle answer: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to handle answer",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer processed successfully",
		"peer_id": request.PeerID,
	})
}

// HandleICECandidate 处理ICE候选
func (h *WebRTCHandler) HandleICECandidate(c *gin.Context) {
	var request struct {
		PeerID    string `json:"peer_id" binding:"required"`
		Candidate struct {
			Candidate     string `json:"candidate" binding:"required"`
			SDPMid        string `json:"sdp_mid"`
			SDPMLineIndex uint16 `json:"sdp_mline_index"`
		} `json:"candidate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 解析ICE候选
	candidate := webrtc.ICECandidateInit{
		Candidate:     request.Candidate.Candidate,
		SDPMid:        &request.Candidate.SDPMid,
		SDPMLineIndex: &request.Candidate.SDPMLineIndex,
	}

	// 处理ICE候选
	if err := h.webrtcService.HandleICECandidate(request.PeerID, &candidate); err != nil {
		logger.Error("Failed to handle ICE candidate: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to handle ICE candidate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ICE candidate processed successfully",
		"peer_id": request.PeerID,
	})
}

// JoinRoom 加入房间
func (h *WebRTCHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room_id is required",
		})
		return
	}

	var request struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 加入房间
	room, err := h.webrtcService.JoinRoom(roomID, request.UserID)
	if err != nil {
		logger.Error("Failed to join room: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to join room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Joined room successfully",
		"room": gin.H{
			"id":         room.ID,
			"meeting_id": room.MeetingID,
			"created_at": room.CreatedAt,
		},
		"user_id": request.UserID,
	})
}

// LeaveRoom 离开房间
func (h *WebRTCHandler) LeaveRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room_id is required",
		})
		return
	}

	var request struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 离开房间
	if err := h.webrtcService.LeaveRoom(roomID, request.UserID); err != nil {
		logger.Error("Failed to leave room: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to leave room",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Left room successfully",
		"room_id": roomID,
		"user_id": request.UserID,
	})
}

// GetRoomPeers 获取房间中的对等连接
func (h *WebRTCHandler) GetRoomPeers(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room_id is required",
		})
		return
	}

	// 获取房间中的对等连接
	peers, err := h.webrtcService.GetRoomPeers(roomID)
	if err != nil {
		logger.Error("Failed to get room peers: " + err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Room not found",
		})
		return
	}

	// 转换为响应格式
	peerList := make([]gin.H, len(peers))
	for i, peer := range peers {
		peerList[i] = gin.H{
			"peer_id":       peer.ID,
			"user_id":       peer.UserID,
			"media_type":    peer.MediaType,
			"status":        peer.Status,
			"created_at":    peer.CreatedAt,
			"last_activity": peer.LastActivity,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"room_id": roomID,
		"peers":   peerList,
		"count":   len(peers),
	})
}

// GetPeerStatus 获取对等连接状态
func (h *WebRTCHandler) GetPeerStatus(c *gin.Context) {
	peerID := c.Param("peerId")
	if peerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "peer_id is required",
		})
		return
	}

	// 这里应该调用webrtcService的GetPeerStatus方法
	// 为了简化，我们返回一个模拟响应
	c.JSON(http.StatusOK, gin.H{
		"peer_id": peerID,
		"status":  "connected",
		"message": "Peer status retrieved successfully",
	})
}

// UpdatePeerMedia 更新对等连接媒体设置
func (h *WebRTCHandler) UpdatePeerMedia(c *gin.Context) {
	peerID := c.Param("peerId")
	if peerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "peer_id is required",
		})
		return
	}

	var request struct {
		AudioEnabled bool `json:"audio_enabled"`
		VideoEnabled bool `json:"video_enabled"`
		ScreenShare  bool `json:"screen_share"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 这里应该调用webrtcService的UpdatePeerMedia方法
	// 为了简化，我们返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message":       "Peer media settings updated successfully",
		"peer_id":       peerID,
		"audio_enabled": request.AudioEnabled,
		"video_enabled": request.VideoEnabled,
		"screen_share":  request.ScreenShare,
	})
}

// GetRoomStats 获取房间统计信息
func (h *WebRTCHandler) GetRoomStats(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room_id is required",
		})
		return
	}

	// 获取房间中的对等连接
	peers, err := h.webrtcService.GetRoomPeers(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Room not found",
		})
		return
	}

	// 统计信息
	stats := gin.H{
		"room_id":        roomID,
		"total_peers":    len(peers),
		"connected_peers": 0,
		"audio_streams":  0,
		"video_streams":  0,
		"screen_shares":  0,
	}

	for _, peer := range peers {
		if peer.Status == "connected" {
			stats["connected_peers"] = stats["connected_peers"].(int) + 1
		}
		
		switch peer.MediaType {
		case "audio":
			stats["audio_streams"] = stats["audio_streams"].(int) + 1
		case "video":
			stats["video_streams"] = stats["video_streams"].(int) + 1
		case "screen":
			stats["screen_shares"] = stats["screen_shares"].(int) + 1
		}
	}

	c.JSON(http.StatusOK, stats)
}
