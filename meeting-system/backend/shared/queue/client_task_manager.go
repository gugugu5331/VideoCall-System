package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ClientTaskType 描述客户端任务类型
type ClientTaskType string

const (
	ClientTaskAISpeechRecognition  ClientTaskType = "ai.speech_recognition"
	ClientTaskAIEmotionDetection   ClientTaskType = "ai.emotion_detection"
	ClientTaskAISynthesisDetection ClientTaskType = "ai.synthesis_detection"
	ClientTaskAIAudioDenoising     ClientTaskType = "ai.audio_denoising"
	ClientTaskAIVideoEnhancement   ClientTaskType = "ai.video_enhancement"

	ClientTaskMediaStart       ClientTaskType = "media.start"
	ClientTaskMediaStop        ClientTaskType = "media.stop"
	ClientTaskMediaRecord      ClientTaskType = "media.record"
	ClientTaskMediaScreenShare ClientTaskType = "media.screen_share"

	ClientTaskMeetingJoin     ClientTaskType = "meeting.join"
	ClientTaskMeetingLeave    ClientTaskType = "meeting.leave"
	ClientTaskMeetingMute     ClientTaskType = "meeting.mute"
	ClientTaskMeetingUnmute   ClientTaskType = "meeting.unmute"
	ClientTaskMeetingKickUser ClientTaskType = "meeting.kick_user"

	ClientTaskChatSend    ClientTaskType = "chat.send"
	ClientTaskChatHistory ClientTaskType = "chat.history"
)

// ClientTaskRequest 客户端任务请求
type ClientTaskRequest struct {
	RequestID string                 `json:"request_id"`
	Type      ClientTaskType         `json:"type"`
	UserID    uint                   `json:"user_id"`
	MeetingID uint                   `json:"meeting_id,omitempty"`
	SessionID string                 `json:"session_id"`
	Data      map[string]interface{} `json:"data"`
	Priority  string                 `json:"priority,omitempty"`
	Timeout   int                    `json:"timeout,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// ClientTaskResponse 客户端任务响应
type ClientTaskResponse struct {
	RequestID   string                 `json:"request_id"`
	Status      string                 `json:"status"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ProcessedAt int64                  `json:"processed_at"`
}

// ClientSession 保存基本会话信息
type ClientSession struct {
	SessionID   string
	UserID      uint
	MeetingID   uint
	ConnectedAt time.Time
	LastActive  time.Time
}

// ClientTaskManager 轻量级客户端任务管理器（无消息队列）
type ClientTaskManager struct {
	dispatcher *TaskDispatcher

	sessions      map[string]*ClientSession
	sessionsMutex sync.RWMutex

	stats struct {
		sync.Mutex
		totalRequests uint64
		totalSuccess  uint64
		totalFailed   uint64
	}
}

// NewClientTaskManager 构造函数
func NewClientTaskManager(dispatcher *TaskDispatcher) *ClientTaskManager {
	return &ClientTaskManager{
		dispatcher: dispatcher,
		sessions:   make(map[string]*ClientSession),
	}
}

// RegisterSession 注册客户端会话
func (ctm *ClientTaskManager) RegisterSession(sessionID string, userID, meetingID uint) {
	ctm.sessionsMutex.Lock()
	defer ctm.sessionsMutex.Unlock()
	ctm.sessions[sessionID] = &ClientSession{
		SessionID:   sessionID,
		UserID:      userID,
		MeetingID:   meetingID,
		ConnectedAt: time.Now(),
		LastActive:  time.Now(),
	}
}

// UnregisterSession 注销会话
func (ctm *ClientTaskManager) UnregisterSession(sessionID string) {
	ctm.sessionsMutex.Lock()
	defer ctm.sessionsMutex.Unlock()
	delete(ctm.sessions, sessionID)
}

// SubmitTask 将客户端请求同步转发给 TaskDispatcher
func (ctm *ClientTaskManager) SubmitTask(ctx context.Context, request *ClientTaskRequest) (*ClientTaskResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	ctm.sessionsMutex.RLock()
	session, exists := ctm.sessions[request.SessionID]
	ctm.sessionsMutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("invalid session: %s", request.SessionID)
	}

	if request.RequestID == "" {
		request.RequestID = generateMessageID()
	}
	if request.Timestamp == 0 {
		request.Timestamp = time.Now().Unix()
	}

	taskRequest := &TaskRequest{
		TaskID:    request.RequestID,
		Type:      ctm.mapTaskType(request.Type),
		Priority:  PriorityNormal,
		UserID:    session.UserID,
		MeetingID: session.MeetingID,
		SessionID: request.SessionID,
		Payload:   request.Data,
		CreatedAt: time.Now().Unix(),
	}

	ctm.stats.Lock()
	ctm.stats.totalRequests++
	ctm.stats.Unlock()

	response, err := ctm.dispatcher.DispatchTask(ctx, taskRequest)
	if err != nil {
		ctm.stats.Lock()
		ctm.stats.totalFailed++
		ctm.stats.Unlock()
		return &ClientTaskResponse{
			RequestID:   request.RequestID,
			Status:      "error",
			Error:       err.Error(),
			ProcessedAt: time.Now().Unix(),
		}, err
	}

	ctm.stats.Lock()
	ctm.stats.totalSuccess++
	ctm.stats.Unlock()

	ctm.sessionsMutex.Lock()
	session.LastActive = time.Now()
	ctm.sessionsMutex.Unlock()

	return &ClientTaskResponse{
		RequestID:   request.RequestID,
		Status:      "success",
		Data:        response.Result,
		ProcessedAt: response.ProcessedAt,
	}, nil
}

// SubmitTaskAsync 同步执行 SubmitTask 后异步回调
func (ctm *ClientTaskManager) SubmitTaskAsync(ctx context.Context, request *ClientTaskRequest, callback func(*ClientTaskResponse)) error {
	resp, err := ctm.SubmitTask(ctx, request)
	if callback != nil {
		go callback(resp)
	}
	return err
}

// Start/Stop 保留空实现
func (ctm *ClientTaskManager) Start() error { return nil }
func (ctm *ClientTaskManager) Stop()        {}

// GetStats 返回简单统计数据
func (ctm *ClientTaskManager) GetStats() map[string]uint64 {
	ctm.stats.Lock()
	defer ctm.stats.Unlock()
	return map[string]uint64{
		"total_requests": ctm.stats.totalRequests,
		"total_success":  ctm.stats.totalSuccess,
		"total_failed":   ctm.stats.totalFailed,
	}
}

// mapTaskType 将客户端任务类型映射到调度器任务类型
func (ctm *ClientTaskManager) mapTaskType(taskType ClientTaskType) TaskType {
	switch taskType {
	case ClientTaskAISpeechRecognition:
		return TaskTypeSpeechRecognition
	case ClientTaskAIEmotionDetection:
		return TaskTypeEmotionDetection
	case ClientTaskAISynthesisDetection:
		return TaskTypeSynthesisDetection
	case ClientTaskAIAudioDenoising:
		return TaskTypeAudioDenoising
	case ClientTaskAIVideoEnhancement:
		return TaskTypeVideoEnhancement
	case ClientTaskMediaStart, ClientTaskMediaStop, ClientTaskMediaRecord, ClientTaskMediaScreenShare:
		return TaskTypeMediaProcessing
	case ClientTaskMeetingJoin, ClientTaskMeetingLeave, ClientTaskMeetingMute, ClientTaskMeetingUnmute, ClientTaskMeetingKickUser:
		return TaskTypeMeetingControl
	case ClientTaskChatSend, ClientTaskChatHistory:
		return TaskTypeSignaling
	default:
		return TaskTypeSignaling
	}
}
