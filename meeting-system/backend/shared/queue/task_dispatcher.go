package queue

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeSpeechRecognition  TaskType = "speech_recognition"
	TaskTypeEmotionDetection   TaskType = "emotion_detection"
	TaskTypeSynthesisDetection TaskType = "synthesis_detection"
	TaskTypeAudioDenoising     TaskType = "audio_denoising"
	TaskTypeVideoEnhancement   TaskType = "video_enhancement"
	TaskTypeTextToSpeech       TaskType = "text_to_speech"

	TaskTypeMediaProcessing TaskType = "media_processing"
	TaskTypeStreamControl   TaskType = "stream_control"
	TaskTypeRecording       TaskType = "recording"

	TaskTypeMeetingControl TaskType = "meeting_control"
	TaskTypeUserManagement TaskType = "user_management"

	TaskTypeSignaling         TaskType = "signaling"
	TaskTypeWebRTCNegotiation TaskType = "webrtc_negotiation"
)

// TaskStatus 描述任务执行状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskRequest 描述待执行的任务
type TaskRequest struct {
	TaskID     string                 `json:"task_id"`
	Type       TaskType               `json:"type"`
	Priority   MessagePriority        `json:"priority"`
	UserID     uint                   `json:"user_id"`
	MeetingID  uint                   `json:"meeting_id,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	Payload    map[string]interface{} `json:"payload"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timeout    int64                  `json:"timeout"`
	MaxRetries int                    `json:"max_retries"`
	CreatedAt  int64                  `json:"created_at"`
}

// TaskResponse 为同步分发返回结果
type TaskResponse struct {
	TaskID      string                 `json:"task_id"`
	Status      TaskStatus             `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ProcessedAt int64                  `json:"processed_at"`
	Duration    int64                  `json:"duration"`
}

// TaskCallback 回调函数
type TaskCallback func(ctx context.Context, response *TaskResponse) error

// ServiceRoute 描述任务与服务的映射关系
type ServiceRoute struct {
	ServiceName string
	QueueName   string
	TaskTypes   []TaskType
}

// TaskDispatcher 现为轻量级同步调度器，保留 API 方便上层调用
// 所有任务被立即标记为完成，可以在需要时扩展为真正的执行逻辑
type TaskDispatcher struct {
	routes      map[TaskType]*ServiceRoute
	routesMutex sync.RWMutex

	taskStatus  map[string]TaskStatus
	statusMutex sync.RWMutex

	taskStore  map[string]*TaskRequest
	storeMutex sync.RWMutex

	callbacks     map[string]TaskCallback
	callbackMutex sync.RWMutex

	stats struct {
		sync.Mutex
		totalDispatched uint64
		totalCompleted  uint64
		totalFailed     uint64
	}
}

// NewTaskDispatcher 创建调度器（同步实现，不依赖外部队列）。
func NewTaskDispatcher() *TaskDispatcher {

	td := &TaskDispatcher{
		routes:     make(map[TaskType]*ServiceRoute),
		taskStatus: make(map[string]TaskStatus),
		taskStore:  make(map[string]*TaskRequest),
		callbacks:  make(map[string]TaskCallback),
	}
	td.registerDefaultRoutes()
	return td
}

// registerDefaultRoutes 保留默认路由配置
func (td *TaskDispatcher) registerDefaultRoutes() {
	td.RegisterRoute(&ServiceRoute{
		ServiceName: "ai-service",
		QueueName:   "ai_tasks",
		TaskTypes: []TaskType{
			TaskTypeSpeechRecognition,
			TaskTypeEmotionDetection,
			TaskTypeSynthesisDetection,
			TaskTypeAudioDenoising,
			TaskTypeVideoEnhancement,
			TaskTypeTextToSpeech,
		},
	})

	td.RegisterRoute(&ServiceRoute{
		ServiceName: "media-service",
		QueueName:   "media_tasks",
		TaskTypes: []TaskType{
			TaskTypeMediaProcessing,
			TaskTypeStreamControl,
			TaskTypeRecording,
		},
	})

	td.RegisterRoute(&ServiceRoute{
		ServiceName: "meeting-service",
		QueueName:   "meeting_tasks",
		TaskTypes: []TaskType{
			TaskTypeMeetingControl,
			TaskTypeUserManagement,
		},
	})

	td.RegisterRoute(&ServiceRoute{
		ServiceName: "signaling-service",
		QueueName:   "signaling_tasks",
		TaskTypes: []TaskType{
			TaskTypeSignaling,
			TaskTypeWebRTCNegotiation,
		},
	})
}

// RegisterRoute 注册自定义路由
func (td *TaskDispatcher) RegisterRoute(route *ServiceRoute) {
	if route == nil {
		return
	}
	td.routesMutex.Lock()
	defer td.routesMutex.Unlock()
	for _, t := range route.TaskTypes {
		td.routes[t] = route
	}
}

// DispatchTask 同步处理任务，直接回传完成状态
func (td *TaskDispatcher) DispatchTask(ctx context.Context, request *TaskRequest) (*TaskResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if request.TaskID == "" {
		request.TaskID = generateMessageID()
	}
	if request.CreatedAt == 0 {
		request.CreatedAt = time.Now().Unix()
	}

	td.routesMutex.RLock()
	_, exists := td.routes[request.Type]
	td.routesMutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("no route registered for task type: %s", request.Type)
	}

	td.storeMutex.Lock()
	td.taskStore[request.TaskID] = request
	td.storeMutex.Unlock()

	td.updateTaskStatus(request.TaskID, TaskStatusCompleted)

	response := &TaskResponse{
		TaskID:      request.TaskID,
		Status:      TaskStatusCompleted,
		Result:      request.Payload,
		ProcessedAt: time.Now().Unix(),
		Duration:    0,
	}

	td.stats.Lock()
	td.stats.totalDispatched++
	td.stats.totalCompleted++
	td.stats.Unlock()

	// 如果注册了回调，同步执行一次
	td.callbackMutex.RLock()
	callback := td.callbacks[request.TaskID]
	td.callbackMutex.RUnlock()
	if callback != nil {
		_ = callback(ctx, response)
	}

	return response, nil
}

// DispatchTaskAsync 以 goroutine 触发 callback，保持 API 兼容
func (td *TaskDispatcher) DispatchTaskAsync(ctx context.Context, request *TaskRequest, callback TaskCallback) error {
	response, err := td.DispatchTask(ctx, request)
	if err != nil {
		return err
	}

	if callback != nil {
		go callback(ctx, response)
	}
	return nil
}

// DispatchBatchTasks 顺序调用 DispatchTask
func (td *TaskDispatcher) DispatchBatchTasks(ctx context.Context, requests []*TaskRequest) ([]*TaskResponse, error) {
	responses := make([]*TaskResponse, 0, len(requests))
	for _, req := range requests {
		resp, err := td.DispatchTask(ctx, req)
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// GetTaskStatus 返回任务存储信息
func (td *TaskDispatcher) GetTaskStatus(taskID string) (*TaskRequest, TaskStatus, error) {
	td.storeMutex.RLock()
	req, exists := td.taskStore[taskID]
	td.storeMutex.RUnlock()
	if !exists {
		return nil, "", fmt.Errorf("task not found: %s", taskID)
	}

	td.statusMutex.RLock()
	status := td.taskStatus[taskID]
	td.statusMutex.RUnlock()
	return req, status, nil
}

// CancelTask 将任务状态置为取消
func (td *TaskDispatcher) CancelTask(_ context.Context, taskID string) error {
	td.statusMutex.Lock()
	defer td.statusMutex.Unlock()
	if _, exists := td.taskStatus[taskID]; !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	td.taskStatus[taskID] = TaskStatusCancelled
	return nil
}

// RegisterCallback 注册任务回调
func (td *TaskDispatcher) RegisterCallback(taskID string, callback TaskCallback) {
	td.callbackMutex.Lock()
	defer td.callbackMutex.Unlock()
	td.callbacks[taskID] = callback
}

// Start/Stop 保留空实现，兼容旧逻辑
func (td *TaskDispatcher) Start() error { return nil }
func (td *TaskDispatcher) Stop() error  { return nil }

// GetStats 返回简单统计
func (td *TaskDispatcher) GetStats() map[string]uint64 {
	td.stats.Lock()
	defer td.stats.Unlock()
	return map[string]uint64{
		"total_dispatched": td.stats.totalDispatched,
		"total_completed":  td.stats.totalCompleted,
		"total_failed":     td.stats.totalFailed,
	}
}

// track/update 状态的辅助函数
func (td *TaskDispatcher) updateTaskStatus(taskID string, status TaskStatus) {
	td.statusMutex.Lock()
	defer td.statusMutex.Unlock()
	td.taskStatus[taskID] = status
}

// GetRoutes 便于调试/测试
func (td *TaskDispatcher) GetRoutes() map[TaskType]*ServiceRoute {
	td.routesMutex.RLock()
	defer td.routesMutex.RUnlock()
	result := make(map[TaskType]*ServiceRoute, len(td.routes))
	for k, v := range td.routes {
		result[k] = v
	}
	return result
}

// RemoveRoute 删除路由
func (td *TaskDispatcher) RemoveRoute(taskType TaskType) {
	td.routesMutex.Lock()
	defer td.routesMutex.Unlock()
	delete(td.routes, taskType)
}
