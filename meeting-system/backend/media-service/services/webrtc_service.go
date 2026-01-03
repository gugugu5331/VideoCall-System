package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"meeting-system/media-service/models"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// WebRTCService WebRTC服务
type WebRTCService struct {
	config         *config.Config
	mediaService   *MediaService
	mediaProcessor *MediaProcessor
	api            *webrtc.API
	rooms          map[string]*Room
	roomsMux       sync.RWMutex
	peers          map[string]*Peer
	peersMux       sync.RWMutex
}

// Room WebRTC房间
type Room struct {
	ID          string
	MeetingID   string
	Peers       map[string]*Peer
	PeersMux    sync.RWMutex
	Tracks      map[string]*ForwardedTrack
	TracksMux   sync.RWMutex
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

	localICECandidates         []webrtc.ICECandidateInit
	localICEGatheringCompleted bool
	localICECandidatesMux      sync.Mutex

	negotiationMux             sync.Mutex
	pendingOffer               *webrtc.SessionDescription
	isNegotiating              bool
	needsRenegotiationAfterAck bool
	subscribedExistingTracks   bool
}

// ForwardedTrack 房间内转发的媒体轨道（一个 remote track -> 一个本地 track，多 PeerConnection 绑定）
type ForwardedTrack struct {
	Key         string
	SenderPeer  string
	RemoteTrack *webrtc.TrackRemote
	// RemoteSSRC 是发布者向 SFU 发送该轨道的 SSRC（订阅者侧的 SSRC 会被 TrackLocalStaticRTP 重写）
	RemoteSSRC uint32
	LocalTrack  *webrtc.TrackLocalStaticRTP
	// subscriberPeerID -> sender(本地 rtp sender)
	// 用于在发布者离线/轨道结束时，能从订阅者 PeerConnection 中 RemoveTrack 并触发 renegotiation，
	// 否则浏览器端会一直保留“僵尸轨道/僵尸窗口”，并造成资源泄漏与卡顿。
	SubscriberSenders map[string]*webrtc.RTPSender
	CreatedAt   time.Time
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
		s.handleICECandidate(peer, candidate)
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

	finalAnswer := peerConnection.LocalDescription()
	if finalAnswer == nil {
		finalAnswer = &answer
	}

	// 保存到数据库
	go s.savePeerToDB(peer, finalAnswer.SDP)

	logger.Info(fmt.Sprintf("Created Answer for peer %s in room %s", peerID, roomID))

	return finalAnswer, peerID, nil
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

func (p *Peer) enqueueLocalICECandidate(candidate webrtc.ICECandidateInit) {
	p.localICECandidatesMux.Lock()
	p.localICECandidates = append(p.localICECandidates, candidate)
	p.localICECandidatesMux.Unlock()
}

func (p *Peer) markLocalICEGatheringCompleted() {
	p.localICECandidatesMux.Lock()
	p.localICEGatheringCompleted = true
	p.localICECandidatesMux.Unlock()
}

func (p *Peer) drainLocalICECandidates() ([]webrtc.ICECandidateInit, bool) {
	p.localICECandidatesMux.Lock()
	candidates := make([]webrtc.ICECandidateInit, len(p.localICECandidates))
	copy(candidates, p.localICECandidates)
	p.localICECandidates = nil
	complete := p.localICEGatheringCompleted
	p.localICECandidatesMux.Unlock()
	return candidates, complete
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
			Tracks:    make(map[string]*ForwardedTrack),
			CreatedAt: time.Now(),
		}
		s.rooms[roomID] = room
	}

	logger.Info(fmt.Sprintf("User %s joined room %s", userID, roomID))
	return room, nil
}

// LeaveRoom 离开房间
func (s *WebRTCService) LeaveRoom(roomID, userID string) error {
	s.roomsMux.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMux.RUnlock()
	if !exists || room == nil {
		return fmt.Errorf("room not found: %s", roomID)
	}

	// 收集该用户在房间内的所有 peer_id，再逐个清理（不要在锁内做 Close/RemoveTrack）。
	room.PeersMux.RLock()
	peerIDs := make([]string, 0, 4)
	for peerID, peer := range room.Peers {
		if peer == nil {
			continue
		}
		if peer.UserID == userID {
			peerIDs = append(peerIDs, peerID)
		}
	}
	room.PeersMux.RUnlock()

	for _, peerID := range peerIDs {
		s.cleanupPeer(peerID, "leave_room")
	}

	logger.Info(fmt.Sprintf("User %s left room %s (peers=%d)", userID, roomID, len(peerIDs)))
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
	iceServers := make([]webrtc.ICEServer, 0, 4)
	if s.config != nil && len(s.config.WebRTC.ICEServers) > 0 {
		for _, srv := range s.config.WebRTC.ICEServers {
			if len(srv.URLs) == 0 {
				continue
			}
			iceServers = append(iceServers, webrtc.ICEServer{
				URLs:       srv.URLs,
				Username:   srv.Username,
				Credential: srv.Credential,
			})
		}
	}

	if len(iceServers) == 0 {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{"stun:stun.l.google.com:19302"},
		})
	}

	config := webrtc.Configuration{
		ICEServers: iceServers,
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
	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()

	if !exists {
		return
	}

	switch state {
	case webrtc.PeerConnectionStateConnected:
		peer.Status = "connected"
		logger.Info(fmt.Sprintf("Peer %s connected", peerID))

		peer.negotiationMux.Lock()
		shouldSubscribe := !peer.subscribedExistingTracks
		if shouldSubscribe {
			peer.subscribedExistingTracks = true
		}
		peer.negotiationMux.Unlock()

		if shouldSubscribe {
			// 等到初始 ICE/DTLS 成功（connected）后再订阅已有轨道并触发 renegotiation，
			// 避免在首轮 Answer 还未被客户端应用时覆盖本地描述导致 ICE 失败。
			s.subscribePeerToExistingTracks(peer)
		}
	case webrtc.PeerConnectionStateDisconnected, webrtc.PeerConnectionStateFailed, webrtc.PeerConnectionStateClosed:
		peer.Status = "disconnected"
		logger.Info(fmt.Sprintf("Peer %s disconnected", peerID))

		// 清理：移除 peers/forwarded tracks/订阅关系，触发 renegotiation 让浏览器端尽快结束轨道。
		go s.cleanupPeer(peerID, "pc_state_"+state.String())
	}

	peer.LastActivity = time.Now()

	// 更新数据库状态
	go s.updatePeerStatus(peerID, peer.Status)
}

// handleICECandidate 处理ICE候选
func (s *WebRTCService) handleICECandidate(peer *Peer, candidate *webrtc.ICECandidate) {
	if candidate == nil {
		peer.markLocalICEGatheringCompleted()
		return
	}

	peer.enqueueLocalICECandidate(candidate.ToJSON())
}

func (s *WebRTCService) DrainLocalICECandidates(peerID string) ([]webrtc.ICECandidateInit, bool, error) {
	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()
	if !exists {
		return nil, false, fmt.Errorf("peer not found: %s", peerID)
	}

	candidates, complete := peer.drainLocalICECandidates()
	return candidates, complete, nil
}

func (s *WebRTCService) RequestRenegotiation(peerID string) {
	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()
	if !exists || peer.Connection == nil {
		return
	}

	peer.negotiationMux.Lock()
	if peer.isNegotiating {
		peer.needsRenegotiationAfterAck = true
		peer.negotiationMux.Unlock()
		return
	}
	if peer.pendingOffer != nil {
		peer.needsRenegotiationAfterAck = true
		peer.negotiationMux.Unlock()
		return
	}
	peer.isNegotiating = true
	peer.negotiationMux.Unlock()

	go func(p *Peer) {
		offer, err := p.Connection.CreateOffer(nil)
		if err == nil {
			err = p.Connection.SetLocalDescription(offer)
		}

		p.negotiationMux.Lock()
		p.isNegotiating = false
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create renegotiation offer for peer %s: %v", p.ID, err))
			p.negotiationMux.Unlock()
			return
		}
		local := p.Connection.LocalDescription()
		if local != nil {
			copied := *local
			p.pendingOffer = &copied
		} else {
			copied := offer
			p.pendingOffer = &copied
		}
		p.negotiationMux.Unlock()
	}(peer)
}

func (s *WebRTCService) GetPendingOffer(peerID string) (*webrtc.SessionDescription, error) {
	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()
	if !exists {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}

	peer.negotiationMux.Lock()
	defer peer.negotiationMux.Unlock()
	if peer.pendingOffer == nil {
		return nil, nil
	}
	copied := *peer.pendingOffer
	return &copied, nil
}

func (s *WebRTCService) HandleRenegotiationAnswer(peerID string, answer *webrtc.SessionDescription) error {
	if err := s.HandleAnswer(peerID, answer); err != nil {
		return err
	}

	s.peersMux.RLock()
	peer, exists := s.peers[peerID]
	s.peersMux.RUnlock()
	if !exists {
		return nil
	}
	roomID := peer.RoomID

	var shouldRenegotiate bool
	peer.negotiationMux.Lock()
	peer.pendingOffer = nil
	shouldRenegotiate = peer.needsRenegotiationAfterAck
	peer.needsRenegotiationAfterAck = false
	peer.negotiationMux.Unlock()

	if shouldRenegotiate {
		s.RequestRenegotiation(peerID)
	}

	// renegotiation 完成后，主动请求房间内所有视频发布者发送关键帧，
	// 避免新订阅者长时间黑屏（等待自然关键帧或用户切换屏幕共享才出现画面）。
	go s.requestRoomKeyFrames(roomID, peerID)

	return nil
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
			Tracks:    make(map[string]*ForwardedTrack),
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

func (s *WebRTCService) cleanupPeer(peerID string, reason string) {
	s.peersMux.Lock()
	peer, exists := s.peers[peerID]
	if exists {
		delete(s.peers, peerID)
	}
	s.peersMux.Unlock()

	if !exists || peer == nil {
		return
	}

	roomID := peer.RoomID

	// 先从房间 peers 中移除，避免后续订阅/转发继续向该 peer AddTrack。
	s.roomsMux.RLock()
	room := s.rooms[roomID]
	s.roomsMux.RUnlock()

	if room != nil {
		room.PeersMux.Lock()
		delete(room.Peers, peerID)
		roomEmpty := len(room.Peers) == 0
		room.PeersMux.Unlock()

		// 1) 取消该 peer 作为订阅者在所有 ForwardedTrack 上的绑定（避免后续 RemoveTrack 时引用失效 sender）
		room.TracksMux.Lock()
		for _, ft := range room.Tracks {
			if ft == nil || ft.SubscriberSenders == nil {
				continue
			}
			delete(ft.SubscriberSenders, peerID)
		}
		room.TracksMux.Unlock()

		// 2) 下线该 peer 发布的轨道，并对订阅者执行 RemoveTrack + renegotiation
		s.unpublishPeerTracks(roomID, peerID, reason)

		// 房间已空则删除（避免 map 膨胀）
		if roomEmpty {
			s.roomsMux.Lock()
			// double-check
			if current, ok := s.rooms[roomID]; ok && current != nil {
				current.PeersMux.RLock()
				empty := len(current.Peers) == 0
				current.PeersMux.RUnlock()
				if empty {
					delete(s.rooms, roomID)
				}
			}
			s.roomsMux.Unlock()
		}
	}

	// 关闭连接（会触发更多状态回调；由于已从 peers map 删除，所以回调会直接 return）
	if peer.Connection != nil {
		_ = peer.Connection.Close()
	}

	peer.Status = "disconnected"
	peer.LastActivity = time.Now()
	go s.updatePeerStatus(peerID, peer.Status)

	logger.Info(fmt.Sprintf("Cleaned up peer %s (room=%s, reason=%s)", peerID, roomID, reason))
}

func (s *WebRTCService) unpublishPeerTracks(roomID string, senderPeerID string, reason string) {
	if roomID == "" || senderPeerID == "" {
		return
	}

	s.roomsMux.RLock()
	room := s.rooms[roomID]
	s.roomsMux.RUnlock()
	if room == nil {
		return
	}

	room.TracksMux.RLock()
	keys := make([]string, 0, len(room.Tracks))
	for key, ft := range room.Tracks {
		if ft == nil {
			continue
		}
		if ft.SenderPeer == senderPeerID {
			keys = append(keys, key)
		}
	}
	room.TracksMux.RUnlock()

	for _, key := range keys {
		s.removeForwardedTrack(roomID, key, reason)
	}
}

func (s *WebRTCService) removeForwardedTrack(roomID string, trackKey string, reason string) {
	if roomID == "" || trackKey == "" {
		return
	}

	s.roomsMux.RLock()
	room := s.rooms[roomID]
	s.roomsMux.RUnlock()
	if room == nil {
		return
	}

	var subscribers map[string]*webrtc.RTPSender
	var senderPeerID string

	room.TracksMux.Lock()
	ft, ok := room.Tracks[trackKey]
	if !ok || ft == nil {
		room.TracksMux.Unlock()
		return
	}
	delete(room.Tracks, trackKey)
	senderPeerID = ft.SenderPeer
	if ft.SubscriberSenders != nil {
		subscribers = make(map[string]*webrtc.RTPSender, len(ft.SubscriberSenders))
		for pid, sender := range ft.SubscriberSenders {
			subscribers[pid] = sender
		}
	}
	// 避免悬挂引用
	ft.SubscriberSenders = nil
	room.TracksMux.Unlock()

	if len(subscribers) == 0 {
		logger.Info(fmt.Sprintf("Removed forwarded track %s (room=%s, sender=%s, reason=%s, subscribers=0)", trackKey, roomID, senderPeerID, reason))
		return
	}

	for subPeerID, sender := range subscribers {
		if subPeerID == "" || sender == nil {
			continue
		}

		s.peersMux.RLock()
		subPeer := s.peers[subPeerID]
		s.peersMux.RUnlock()
		if subPeer == nil || subPeer.Connection == nil {
			continue
		}

		if err := subPeer.Connection.RemoveTrack(sender); err != nil {
			logger.Debug(fmt.Sprintf("RemoveTrack failed (sub_peer=%s, track=%s, sender_peer=%s): %v", subPeerID, trackKey, senderPeerID, err))
		}
		s.RequestRenegotiation(subPeerID)
	}

	logger.Info(fmt.Sprintf("Removed forwarded track %s (room=%s, sender=%s, reason=%s, subscribers=%d)", trackKey, roomID, senderPeerID, reason, len(subscribers)))
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
	// 更高频率地清理长时间无活动的 Peer，避免“关闭页面后仍占用资源/卡顿”。
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupInactivePeers()
	}
}

// cleanupInactivePeers 清理不活跃的Peer连接
func (s *WebRTCService) cleanupInactivePeers() {
	now := time.Now()
	inactiveThreshold := 2 * time.Minute

	s.peersMux.RLock()
	peerIDs := make([]string, 0, len(s.peers))
	for peerID, peer := range s.peers {
		if peer == nil {
			peerIDs = append(peerIDs, peerID)
			continue
		}
		if peer.Connection != nil && peer.Connection.ConnectionState() == webrtc.PeerConnectionStateConnected {
			continue
		}
		if now.Sub(peer.LastActivity) > inactiveThreshold {
			peerIDs = append(peerIDs, peerID)
		}
	}
	s.peersMux.RUnlock()

	for _, peerID := range peerIDs {
		s.cleanupPeer(peerID, "inactive_timeout")
	}
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
	if s.mediaProcessor != nil && len(aiTasks) > 0 {
		// 重要：TrackRemote 只能被一个 goroutine ReadRTP()。
		// SFU 转发已经会读取 RTP；AI 侧只接收转发线程分发的 payload。
		streamID := fmt.Sprintf("%s_%s", peerID, track.ID())
		if err := s.mediaProcessor.RegisterStream(streamID, peer.UserID, peer.RoomID, nil, nil, aiTasks); err != nil {
			logger.Error(fmt.Sprintf("Failed to register stream: %v", err))
		} else {
			logger.Info(fmt.Sprintf("Stream registered for AI processing: %s", streamID))
		}
	}

	// 转发轨道到房间内其他用户（SFU模式）
	s.forwardTrackToRoom(peer.RoomID, peerID, track)
}

func (s *WebRTCService) subscribePeerToExistingTracks(peer *Peer) {
	if peer == nil || peer.Connection == nil {
		return
	}

	s.roomsMux.RLock()
	room, exists := s.rooms[peer.RoomID]
	s.roomsMux.RUnlock()
	if !exists || room == nil {
		return
	}

	room.TracksMux.RLock()
	tracks := make([]*ForwardedTrack, 0, len(room.Tracks))
	for _, t := range room.Tracks {
		if t == nil || t.LocalTrack == nil {
			continue
		}
		// 不把自己的轨道再回送给自己
		if t.SenderPeer == peer.ID {
			continue
		}
		tracks = append(tracks, t)
	}
	room.TracksMux.RUnlock()

	var added bool
	for _, t := range tracks {
		trackKey := t.Key

		// 预留订阅位置，避免重复 AddTrack
		room.TracksMux.Lock()
		current := room.Tracks[trackKey]
		if current == nil || current.LocalTrack == nil || current.SenderPeer == peer.ID {
			room.TracksMux.Unlock()
			continue
		}
		if current.SubscriberSenders == nil {
			current.SubscriberSenders = make(map[string]*webrtc.RTPSender)
		}
		if _, ok := current.SubscriberSenders[peer.ID]; ok {
			room.TracksMux.Unlock()
			continue
		}
		current.SubscriberSenders[peer.ID] = nil // pending
		localTrack := current.LocalTrack
		senderPeerID := current.SenderPeer
		publisherSSRC := current.RemoteSSRC
		room.TracksMux.Unlock()

		rtpSender, err := peer.Connection.AddTrack(localTrack)
		if err != nil {
			room.TracksMux.Lock()
			// 失败则撤销预留
			if cur := room.Tracks[trackKey]; cur != nil && cur.SubscriberSenders != nil {
				delete(cur.SubscriberSenders, peer.ID)
			}
			room.TracksMux.Unlock()
			logger.Error(fmt.Sprintf("Failed to subscribe existing track to peer %s: %v", peer.ID, err))
			continue
		}

		room.TracksMux.Lock()
		cur := room.Tracks[trackKey]
		if cur == nil {
			room.TracksMux.Unlock()
			// 轨道在 AddTrack 期间已被清理（发布者离线/轨道结束），撤销本次 AddTrack
			_ = peer.Connection.RemoveTrack(rtpSender)
			s.RequestRenegotiation(peer.ID)
			continue
		}
		if cur.SubscriberSenders == nil {
			cur.SubscriberSenders = make(map[string]*webrtc.RTPSender)
		}
		cur.SubscriberSenders[peer.ID] = rtpSender
		room.TracksMux.Unlock()

		go s.processRTCP(peer.ID, rtpSender, senderPeerID, publisherSSRC)
		added = true
	}

	if added {
		s.RequestRenegotiation(peer.ID)
	}
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

	if track == nil {
		return
	}

	// 统一使用 senderPeerID 作为 streamID，避免浏览器端收到 audio/video 分裂成两个 MediaStream
	// （尤其是 replaceTrack/无 msid 时，StreamID 可能为空）
	streamID := senderPeerID
	trackKey := fmt.Sprintf("%s:%s", senderPeerID, track.ID())
	aiStreamID := fmt.Sprintf("%s_%s", senderPeerID, track.ID())

	// 为每个 remote track 创建一个本地 track（可绑定到多个 PeerConnection），并启动单一 RTP 转发循环
	room.TracksMux.Lock()
	ft, ok := room.Tracks[trackKey]
	if !ok {
		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			track.Codec().RTPCodecCapability,
			track.ID(),
			streamID,
		)
		if err != nil {
			room.TracksMux.Unlock()
			logger.Error(fmt.Sprintf("Failed to create local track: %v", err))
			return
		}

		ft = &ForwardedTrack{
			Key:         trackKey,
			SenderPeer:  senderPeerID,
			RemoteTrack: track,
			RemoteSSRC:  uint32(track.SSRC()),
			LocalTrack:  localTrack,
			SubscriberSenders: make(map[string]*webrtc.RTPSender),
			CreatedAt:   time.Now(),
		}
		room.Tracks[trackKey] = ft

		go s.forwardRTP(roomID, trackKey, senderPeerID, track, localTrack, aiStreamID, track.Kind())
	}
	localTrack := ft.LocalTrack
	room.TracksMux.Unlock()

	// 绑定到房间内其它 PeerConnection（由本地 track 负责 fan-out）
	room.PeersMux.RLock()
	peers := make([]*Peer, 0, len(room.Peers))
	for peerID, peer := range room.Peers {
		if peerID == senderPeerID {
			continue
		}
		peers = append(peers, peer)
	}
	room.PeersMux.RUnlock()

	for _, peer := range peers {
		if peer == nil || peer.Connection == nil {
			continue
		}
		if peer.Status != "connected" {
			// 连接尚未完成时不做 AddTrack/renegotiation，等 connected 后统一订阅已有轨道
			continue
		}

		// 预留订阅位置，避免重复 AddTrack
		room.TracksMux.Lock()
		current := room.Tracks[trackKey]
		if current == nil || current.LocalTrack == nil {
			room.TracksMux.Unlock()
			continue
		}
		if current.SubscriberSenders == nil {
			current.SubscriberSenders = make(map[string]*webrtc.RTPSender)
		}
		if _, ok := current.SubscriberSenders[peer.ID]; ok {
			room.TracksMux.Unlock()
			continue
		}
		current.SubscriberSenders[peer.ID] = nil // pending
		room.TracksMux.Unlock()

		rtpSender, err := peer.Connection.AddTrack(localTrack)
		if err != nil {
			room.TracksMux.Lock()
			if cur := room.Tracks[trackKey]; cur != nil && cur.SubscriberSenders != nil {
				delete(cur.SubscriberSenders, peer.ID)
			}
			room.TracksMux.Unlock()
			logger.Error(fmt.Sprintf("Failed to add track to peer %s: %v", peer.ID, err))
			continue
		}

		room.TracksMux.Lock()
		cur := room.Tracks[trackKey]
		if cur == nil {
			room.TracksMux.Unlock()
			// 轨道在 AddTrack 期间已被清理（发布者离线/轨道结束），撤销本次 AddTrack
			_ = peer.Connection.RemoveTrack(rtpSender)
			s.RequestRenegotiation(peer.ID)
			continue
		}
		if cur.SubscriberSenders == nil {
			cur.SubscriberSenders = make(map[string]*webrtc.RTPSender)
		}
		cur.SubscriberSenders[peer.ID] = rtpSender
		room.TracksMux.Unlock()

		go s.processRTCP(peer.ID, rtpSender, senderPeerID, uint32(track.SSRC()))
		s.RequestRenegotiation(peer.ID)

		logger.Debug(fmt.Sprintf("Track forwarded from %s to %s", senderPeerID, peer.ID))
	}
}

// forwardRTP 转发RTP包（单读 TrackRemote -> fan-out 到各 PeerConnection，并可选分发给 AI 缓冲区）
func (s *WebRTCService) forwardRTP(roomID string, trackKey string, senderPeerID string, remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP, aiStreamID string, kind webrtc.RTPCodecType) {
	defer func() {
		if s.mediaProcessor != nil && aiStreamID != "" {
			_ = s.mediaProcessor.UnregisterStream(aiStreamID)
		}
		// 轨道结束后，必须从房间 tracks 中移除并对订阅者执行 RemoveTrack + renegotiation，
		// 否则浏览器端会一直保留“僵尸窗口/僵尸轨道”，并累积造成卡顿。
		if roomID != "" && trackKey != "" {
			s.removeForwardedTrack(roomID, trackKey, "rtp_end")
		}
	}()

	lastTouch := time.Now()
	for {
		// 读取RTP包
		rtpPacket, _, err := remoteTrack.ReadRTP()
		if err != nil {
			logger.Debug(fmt.Sprintf("RTP read error: %v", err))
			return
		}

		// 仅做节流更新，避免每包都加锁
		if senderPeerID != "" && time.Since(lastTouch) > 2*time.Second {
			lastTouch = time.Now()
			s.peersMux.RLock()
			peer := s.peers[senderPeerID]
			s.peersMux.RUnlock()
			if peer != nil {
				peer.LastActivity = lastTouch
			}
		}

		if s.mediaProcessor != nil && aiStreamID != "" {
			s.mediaProcessor.IngestRTPPayload(aiStreamID, kind, rtpPacket.Payload)
		}

		// 写入到本地轨道
		if err := localTrack.WriteRTP(rtpPacket); err != nil {
			// 注意：TrackLocalStaticRTP 可能会在某个 PeerConnection 写入失败时返回 error，
			// 但仍然会继续向其它 PeerConnection 写入；这里不应停止整个转发循环。
			logger.Warn(fmt.Sprintf("RTP write warning: %v", err))
		}
	}
}

// processRTCP 处理RTCP包
func (s *WebRTCService) processRTCP(subscriberPeerID string, rtpSender *webrtc.RTPSender, senderPeerID string, publisherSSRC uint32) {
	rtcpBuf := make([]byte, 1500)
	for {
		n, _, err := rtpSender.Read(rtcpBuf)
		if err != nil {
			return
		}

		if subscriberPeerID != "" {
			s.peersMux.RLock()
			subPeer := s.peers[subscriberPeerID]
			s.peersMux.RUnlock()
			if subPeer != nil {
				subPeer.LastActivity = time.Now()
			}
		}

		pkts, err := rtcp.Unmarshal(rtcpBuf[:n])
		if err != nil {
			continue
		}
		for _, pkt := range pkts {
			switch p := pkt.(type) {
			case *rtcp.PictureLossIndication:
				// TrackLocalStaticRTP 会为每个订阅者重写 SSRC，RTCP PLI 的 MediaSSRC 不能直接转发给发布者端。
				// 必须使用发布者 remote track 的 SSRC 才能触发正确的关键帧请求。
				s.sendPLI(senderPeerID, publisherSSRC)
			case *rtcp.FullIntraRequest:
				// FIR 更常见于视频流；这里统一用 PLI 触发关键帧即可
				_ = p
				s.sendPLI(senderPeerID, publisherSSRC)
			}
		}
	}
}

func (s *WebRTCService) sendPLI(senderPeerID string, mediaSSRC uint32) {
	if senderPeerID == "" || mediaSSRC == 0 {
		return
	}

	s.peersMux.RLock()
	peer, exists := s.peers[senderPeerID]
	s.peersMux.RUnlock()
	if !exists || peer == nil || peer.Connection == nil {
		return
	}

	if err := peer.Connection.WriteRTCP([]rtcp.Packet{
		&rtcp.PictureLossIndication{MediaSSRC: mediaSSRC},
	}); err != nil {
		logger.Debug(fmt.Sprintf("Failed to send PLI to peer %s: %v", senderPeerID, err))
	}
}

func (s *WebRTCService) requestRoomKeyFrames(roomID, excludePeerID string) {
	if roomID == "" {
		return
	}

	s.roomsMux.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMux.RUnlock()
	if !exists || room == nil {
		return
	}

	room.TracksMux.RLock()
	tracks := make([]*ForwardedTrack, 0, len(room.Tracks))
	for _, t := range room.Tracks {
		tracks = append(tracks, t)
	}
	room.TracksMux.RUnlock()

	for _, t := range tracks {
		if t == nil {
			continue
		}
		if excludePeerID != "" && t.SenderPeer == excludePeerID {
			continue
		}
		if t.RemoteTrack == nil || t.RemoteTrack.Kind() != webrtc.RTPCodecTypeVideo {
			continue
		}
		s.sendPLI(t.SenderPeer, t.RemoteSSRC)
	}
}
