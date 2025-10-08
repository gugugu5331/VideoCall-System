package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"meeting-system/media-service/models"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// WebRTCService WebRTC服务
type WebRTCService struct {
	config          *config.Config
	mediaService    *MediaService
	mediaProcessor  *MediaProcessor
	api             *webrtc.API
	rooms           map[string]*Room
	roomsMux        sync.RWMutex
	peers           map[string]*Peer
	peersMux        sync.RWMutex
}

// Room WebRTC房间
type Room struct {
	ID          string
	MeetingID   string
	Peers       map[string]*Peer
	PeersMux    sync.RWMutex
	CreatedAt   time.Time
	IsRecording bool
	RecordingID string
}

// Peer WebRTC对等连接
type Peer struct {
	ID           string
	UserID       string
	RoomID       string
	Connection   *webrtc.PeerConnection
	DataChannel  *webrtc.DataChannel
	MediaType    string // audio, video, screen
	Status       string // connecting, connected, disconnected
	CreatedAt    time.Time
	LastActivity time.Time
}

// WebRTCMessage WebRTC消息
type WebRTCMessage struct {
	Type   string      `json:"type"`
	PeerID string      `json:"peer_id"`
	RoomID string      `json:"room_id"`
	UserID string      `json:"user_id"`
	Data   interface{} `json:"data"`
}

// OfferData SDP Offer数据
type OfferData struct {
	SDP       string `json:"sdp"`
	MediaType string `json:"media_type"`
}

// AnswerData SDP Answer数据
type AnswerData struct {
	SDP string `json:"sdp"`
}

// ICECandidateData ICE候选数据
type ICECandidateData struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdp_mid"`
	SDPMLineIndex int    `json:"sdp_mline_index"`
}

// NewWebRTCService 创建WebRTC服务
func NewWebRTCService(config *config.Config, mediaService *MediaService, mediaProcessor *MediaProcessor) *WebRTCService {
	return &WebRTCService{
		config:         config,
		mediaService:   mediaService,
		mediaProcessor: mediaProcessor,
		rooms:          make(map[string]*Room),
		peers:          make(map[string]*Peer),
	}
}

// Initialize 初始化WebRTC服务
func (s *WebRTCService) Initialize() error {
	// 创建WebRTC API配置
	mediaEngine := &webrtc.MediaEngine{}

	// 注册编解码器
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return fmt.Errorf("failed to register codecs: %w", err)
	}

	// 创建API实例
	s.api = webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// 启动清理任务
	go s.startCleanupTask()

	logger.Info("WebRTC service initialized successfully")
	return nil
}

// Stop 停止WebRTC服务
func (s *WebRTCService) Stop() {
	// 关闭所有连接
	s.peersMux.Lock()
	for _, peer := range s.peers {
		if peer.Connection != nil {
			peer.Connection.Close()
		}
	}
	s.peersMux.Unlock()

	logger.Info("WebRTC service stopped")
}

// SFU 架构：CreateOffer 方法已删除
// 原因：在标准SFU架构中，客户端创建Offer，SFU创建Answer
// 这样SFU才能接收客户端的媒体流并进行转发
//
// 正确流程：
// 1. 客户端创建PeerConnection和Offer
// 2. 客户端通过信令服务发送Offer到SFU
// 3. SFU调用CreateAnswer()创建Answer
// 4. SFU通过信令服务返回Answer给客户端
// 5. 连接建立后，SFU转发RTP包

// CreateAnswer 创建SDP Answer（响应客户端的Offer）
func (s *WebRTCService) CreateAnswer(roomID, userID string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, string, error) {
	// 创建对等连接
	peerConnection, err := s.createPeerConnection()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create peer connection: %w", err)
	}

	// 生成Peer ID
	peerID := uuid.New().String()

	// 创建Peer对象
	peer := &Peer{
		ID:           peerID,
		UserID:       userID,
		RoomID:       roomID,
		Connection:   peerConnection,
		MediaType:    "video", // 从Offer中推断
		Status:       "connecting",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
	}

	// 设置连接状态回调
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		s.handleConnectionStateChange(peerID, state)
	})

	// 设置轨道处理回调（接收客户端的媒体流）
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		s.handleTrack(peerID, track, receiver)
	})

	// 设置ICE候选回调
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		s.handleICECandidate(peerID, candidate)
	})

	// 设置远程描述（客户端的Offer）
	if err := peerConnection.SetRemoteDescription(*offer); err != nil {
		peerConnection.Close()
		return nil, "", fmt.Errorf("failed to set remote description: %w", err)
	}

	// 创建Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		peerConnection.Close()
		return nil, "", fmt.Errorf("failed to create answer: %w", err)
	}

	// 设置本地描述
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		peerConnection.Close()
		return nil, "", fmt.Errorf("failed to set local description: %w", err)
	}

	// 保存Peer
	s.peersMux.Lock()
	s.peers[peerID] = peer
	s.peersMux.Unlock()

	// 将Peer添加到房间
	s.addPeerToRoom(roomID, peer)

	// 保存到数据库
	go s.savePeerToDB(peer, answer.SDP)

	logger.Info(fmt.Sprintf("Created Answer for peer %s in room %s", peerID, roomID))

	return &answer, peerID, nil
}

// HandleAnswer 处理SDP Answer（已废弃 - SFU不应接收Answer）
// 注意：在纯SFU架构中，这个方法不应该被使用
// SFU创建Answer发送给客户端，而不是接收Answer
func (s *WebRTCService) HandleAnswer(peerID string, answer *webrtc.SessionDescription) error {
	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()

	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	// 设置远程描述
	if err := peer.Connection.SetRemoteDescription(*answer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	// 更新数据库
	go s.updatePeerInDB(peerID, answer.SDP)

	peer.LastActivity = time.Now()
	return nil
}

// HandleICECandidate 处理ICE候选
func (s *WebRTCService) HandleICECandidate(peerID string, candidate *webrtc.ICECandidateInit) error {
	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()

	if !exists {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	// 添加ICE候选
	if err := peer.Connection.AddICECandidate(*candidate); err != nil {
		return fmt.Errorf("failed to add ICE candidate: %w", err)
	}

	peer.LastActivity = time.Now()
	return nil
}

// JoinRoom 加入房间
func (s *WebRTCService) JoinRoom(roomID, userID string) (*Room, error) {
	s.roomsMux.Lock()
	defer s.roomsMux.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		// 创建新房间
		room = &Room{
			ID:        roomID,
			Peers:     make(map[string]*Peer),
			CreatedAt: time.Now(),
		}
		s.rooms[roomID] = room
	}

	logger.Info(fmt.Sprintf("User %s joined room %s", userID, roomID))
	return room, nil
}

// LeaveRoom 离开房间
func (s *WebRTCService) LeaveRoom(roomID, userID string) error {
	s.roomsMux.Lock()
	defer s.roomsMux.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	// 移除用户的所有Peer连接
	room.PeersMux.Lock()
	for peerID, peer := range room.Peers {
		if peer.UserID == userID {
			// 关闭连接
			if peer.Connection != nil {
				peer.Connection.Close()
			}

			// 从房间中移除
			delete(room.Peers, peerID)

			// 从全局Peer映射中移除
			s.peersMux.Lock()
			delete(s.peers, peerID)
			s.peersMux.Unlock()

			// 更新数据库状态
			go s.updatePeerStatus(peerID, "disconnected")
		}
	}
	room.PeersMux.Unlock()

	// 如果房间为空，删除房间
	if len(room.Peers) == 0 {
		delete(s.rooms, roomID)
	}

	logger.Info(fmt.Sprintf("User %s left room %s", userID, roomID))
	return nil
}

// GetRoomPeers 获取房间中的对等连接
func (s *WebRTCService) GetRoomPeers(roomID string) ([]*Peer, error) {
	s.roomsMux.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMux.RUnlock()

	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	room.PeersMux.RLock()
	defer room.PeersMux.RUnlock()

	peers := make([]*Peer, 0, len(room.Peers))
	for _, peer := range room.Peers {
		peers = append(peers, peer)
	}

	return peers, nil
}

// createPeerConnection 创建对等连接
func (s *WebRTCService) createPeerConnection() (*webrtc.PeerConnection, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	return s.api.NewPeerConnection(config)
}

// addMediaTracks 添加媒体轨道
func (s *WebRTCService) addMediaTracks(pc *webrtc.PeerConnection, mediaType string) error {
	switch mediaType {
	case "audio":
		// 添加音频轨道
		if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
	case "video":
		// 添加视频轨道
		if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	case "screen":
		// 添加屏幕共享轨道
		if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	default:
		// 默认添加音频和视频轨道
		if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
		if _, err := pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}

	return nil
}

// handleConnectionStateChange 处理连接状态变化
func (s *WebRTCService) handleConnectionStateChange(peerID string, state webrtc.PeerConnectionState) {
	s.peersMux.Lock()
	peer, exists := s.peers[peerID]
	s.peersMux.Unlock()

	if !exists {
		return
	}

	switch state {
	case webrtc.PeerConnectionStateConnected:
		peer.Status = "connected"
		logger.Info(fmt.Sprintf("Peer %s connected", peerID))
	case webrtc.PeerConnectionStateDisconnected, webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed:
		peer.Status = "disconnected"
		logger.Info(fmt.Sprintf("Peer %s disconnected", peerID))

		// 从房间中移除
		s.removePeerFromRoom(peer.RoomID, peerID)
	}

	peer.LastActivity = time.Now()

	// 更新数据库状态
	go s.updatePeerStatus(peerID, peer.Status)
}

// handleICECandidate 处理ICE候选
func (s *WebRTCService) handleICECandidate(peerID string, candidate *webrtc.ICECandidate) {
	if candidate == nil {
		return
	}

	// 这里可以将ICE候选发送给客户端
	logger.Debug(fmt.Sprintf("ICE candidate for peer %s: %s", peerID, candidate.String()))
}

// addPeerToRoom 将Peer添加到房间
func (s *WebRTCService) addPeerToRoom(roomID string, peer *Peer) {
	s.roomsMux.Lock()
	defer s.roomsMux.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		room = &Room{
			ID:        roomID,
			Peers:     make(map[string]*Peer),
			CreatedAt: time.Now(),
		}
		s.rooms[roomID] = room
	}

	room.PeersMux.Lock()
	room.Peers[peer.ID] = peer
	room.PeersMux.Unlock()
}

// removePeerFromRoom 从房间中移除Peer
func (s *WebRTCService) removePeerFromRoom(roomID, peerID string) {
	s.roomsMux.Lock()
	defer s.roomsMux.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return
	}

	room.PeersMux.Lock()
	delete(room.Peers, peerID)
	room.PeersMux.Unlock()

	// 如果房间为空，删除房间
	if len(room.Peers) == 0 {
		delete(s.rooms, roomID)
	}
}

// savePeerToDB 保存Peer到数据库
func (s *WebRTCService) savePeerToDB(peer *Peer, sdpOffer string) {
	// 如果mediaService为nil（测试环境），跳过数据库操作
	if s.mediaService == nil || s.mediaService.db == nil {
		return
	}

	webrtcPeer := &models.WebRTCPeer{
		PeerID:      peer.ID,
		RoomID:      peer.RoomID,
		UserID:      peer.UserID,
		Status:      peer.Status,
		PeerType:    "publisher", // 默认为发布者
		MediaType:   peer.MediaType,
		SDPOffer:    sdpOffer,
		ConnectedAt: peer.CreatedAt,
		CreatedAt:   peer.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	if err := s.mediaService.db.Create(webrtcPeer).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to save peer to database: %v", err))
	}
}

// updatePeerInDB 更新数据库中的Peer信息
func (s *WebRTCService) updatePeerInDB(peerID, sdpAnswer string) {
	// 如果mediaService为nil（测试环境），跳过数据库操作
	if s.mediaService == nil || s.mediaService.db == nil {
		return
	}

	updates := map[string]interface{}{
		"sdp_answer": sdpAnswer,
		"updated_at": time.Now(),
	}

	if err := s.mediaService.db.Model(&models.WebRTCPeer{}).Where("peer_id = ?", peerID).Updates(updates).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update peer in database: %v", err))
	}
}

// updatePeerStatus 更新Peer状态
func (s *WebRTCService) updatePeerStatus(peerID, status string) {
	// 如果mediaService为nil（测试环境），跳过数据库操作
	if s.mediaService == nil || s.mediaService.db == nil {
		return
	}

	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if err := s.mediaService.db.Model(&models.WebRTCPeer{}).Where("peer_id = ?", peerID).Updates(updates).Error; err != nil {
		logger.Error(fmt.Sprintf("Failed to update peer status: %v", err))
	}
}

// startCleanupTask 启动清理任务
func (s *WebRTCService) startCleanupTask() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupInactivePeers()
	}
}

// cleanupInactivePeers 清理不活跃的Peer连接
func (s *WebRTCService) cleanupInactivePeers() {
	now := time.Now()
	inactiveThreshold := 10 * time.Minute

	s.peersMux.Lock()
	for peerID, peer := range s.peers {
		if now.Sub(peer.LastActivity) > inactiveThreshold {
			// 关闭连接
			if peer.Connection != nil {
				peer.Connection.Close()
			}

			// 从映射中移除
			delete(s.peers, peerID)

			// 从房间中移除
			s.removePeerFromRoom(peer.RoomID, peerID)

			// 更新数据库状态
			go s.updatePeerStatus(peerID, "disconnected")

			logger.Info(fmt.Sprintf("Cleaned up inactive peer: %s", peerID))
		}
	}
	s.peersMux.Unlock()
}

// handleTrack 处理接收到的媒体轨道
func (s *WebRTCService) handleTrack(peerID string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	logger.Info(fmt.Sprintf("Received track: %s, type: %s, codec: %s",
		track.ID(), track.Kind().String(), track.Codec().MimeType))

	// 获取Peer信息
	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()

	if !exists {
		logger.Error(fmt.Sprintf("Peer not found for track: %s", peerID))
		return
	}

	// 确定AI处理任务
	aiTasks := s.determineAITasks(peer.RoomID, track.Kind())

	// 注册到媒体处理器
	if s.mediaProcessor != nil {
		streamID := fmt.Sprintf("%s_%s", peerID, track.ID())

		var audioTrack, videoTrack *webrtc.TrackRemote
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			audioTrack = track
		} else if track.Kind() == webrtc.RTPCodecTypeVideo {
			videoTrack = track
		}

		err := s.mediaProcessor.RegisterStream(
			streamID,
			peer.UserID,
			peer.RoomID,
			audioTrack,
			videoTrack,
			aiTasks,
		)

		if err != nil {
			logger.Error(fmt.Sprintf("Failed to register stream: %v", err))
		} else {
			logger.Info(fmt.Sprintf("Stream registered for AI processing: %s", streamID))
		}
	}

	// 转发轨道到房间内其他用户（SFU模式）
	s.forwardTrackToRoom(peer.RoomID, peerID, track)
}

// determineAITasks 确定需要执行的AI任务
func (s *WebRTCService) determineAITasks(roomID string, trackKind webrtc.RTPCodecType) []string {
	// 默认AI任务配置
	tasks := []string{}

	if trackKind == webrtc.RTPCodecTypeAudio {
		// 音频相关AI任务
		tasks = append(tasks,
			"speech_recognition",  // 语音识别
			"emotion_detection",   // 情绪检测
			"synthesis_detection", // 合成检测
		)
	} else if trackKind == webrtc.RTPCodecTypeVideo {
		// 视频相关AI任务
		tasks = append(tasks,
			"synthesis_detection", // Deepfake检测
			"video_enhancement",   // 视频增强
		)
	}

	logger.Debug(fmt.Sprintf("AI tasks for room %s, track type %s: %v",
		roomID, trackKind.String(), tasks))

	return tasks
}

// forwardTrackToRoom 转发轨道到房间内其他用户
func (s *WebRTCService) forwardTrackToRoom(roomID, senderPeerID string, track *webrtc.TrackRemote) {
	s.roomsMux.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMux.RUnlock()

	if !exists {
		logger.Warn(fmt.Sprintf("Room not found for forwarding: %s", roomID))
		return
	}

	// 遍历房间内的其他Peer
	room.PeersMux.RLock()
	defer room.PeersMux.RUnlock()

	for peerID, peer := range room.Peers {
		if peerID == senderPeerID {
			continue // 不转发给发送者自己
		}

		// 创建本地轨道
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			track.Codec().RTPCodecCapability,
			track.ID(),
			track.StreamID(),
		)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create local track: %v", err))
			continue
		}

		// 添加轨道到接收者的连接
		rtpSender, err := peer.Connection.AddTrack(localTrack)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to add track to peer %s: %v", peerID, err))
			continue
		}

		// 启动RTCP处理
		go s.processRTCP(rtpSender)

		// 启动RTP转发
		go s.forwardRTP(track, localTrack)

		logger.Debug(fmt.Sprintf("Track forwarded from %s to %s", senderPeerID, peerID))
	}
}

// forwardRTP 转发RTP包
func (s *WebRTCService) forwardRTP(remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) {
	for {
		// 读取RTP包
		rtpPacket, _, err := remoteTrack.ReadRTP()
		if err != nil {
			logger.Debug(fmt.Sprintf("RTP read error: %v", err))
			return
		}

		// 写入到本地轨道
		if err := localTrack.WriteRTP(rtpPacket); err != nil {
			logger.Error(fmt.Sprintf("RTP write error: %v", err))
			return
		}
	}
}

// processRTCP 处理RTCP包
func (s *WebRTCService) processRTCP(rtpSender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, 1500)
	for {
		if _, _, err := rtpSender.Read(rtcpBuf); err != nil {
			return
		}
	}
}
