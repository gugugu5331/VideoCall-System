package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"meeting-system/shared/logger"
)

// Task 任务结构
type Task struct {
	ID          string
	Type        string
	Priority    MessagePriority
	Payload     map[string]interface{}
	Handler     TaskHandler
	Timeout     time.Duration
	MaxRetries  int
	RetryCount  int
	CreatedAt   time.Time
	ScheduledAt time.Time // 计划执行时间
	Status      TaskStatus
	Error       error
}

// SchedulerTaskStatus 调度器任务状态
type SchedulerTaskStatus int

const (
	SchedulerTaskStatusPending   SchedulerTaskStatus = 0
	SchedulerTaskStatusRunning   SchedulerTaskStatus = 1
	SchedulerTaskStatusCompleted SchedulerTaskStatus = 2
	SchedulerTaskStatusFailed    SchedulerTaskStatus = 3
	SchedulerTaskStatusCancelled SchedulerTaskStatus = 4
)

// TaskHandler 任务处理函数
type TaskHandler func(ctx context.Context, task *Task) error

// TaskResult 任务结果
type TaskResult struct {
	TaskID    string
	Status    TaskStatus
	Error     error
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// TaskScheduler 任务调度器
type TaskScheduler struct {
	// 任务队列（按优先级）
	criticalQueue chan *Task
	highQueue     chan *Task
	normalQueue   chan *Task
	lowQueue      chan *Task
	
	// 延迟任务队列
	delayedTasks     []*Task
	delayedTaskMutex sync.RWMutex
	
	// 任务跟踪
	activeTasks   map[string]*Task
	taskMutex     sync.RWMutex
	
	// 结果通道
	resultChan chan *TaskResult
	
	// 工作协程
	workers int
	stopCh  chan struct{}
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	
	// 统计信息
	stats struct {
		sync.RWMutex
		totalTasks      int64
		completedTasks  int64
		failedTasks     int64
		cancelledTasks  int64
		retriedTasks    int64
	}
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(bufferSize, workers int) *TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskScheduler{
		criticalQueue: make(chan *Task, bufferSize/4),
		highQueue:     make(chan *Task, bufferSize/4),
		normalQueue:   make(chan *Task, bufferSize/2),
		lowQueue:      make(chan *Task, bufferSize/4),
		delayedTasks:  make([]*Task, 0),
		activeTasks:   make(map[string]*Task),
		resultChan:    make(chan *TaskResult, bufferSize),
		workers:       workers,
		stopCh:        make(chan struct{}),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// SubmitTask 提交任务
func (s *TaskScheduler) SubmitTask(task *Task) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.Timeout == 0 {
		task.Timeout = 30 * time.Second
	}
	
	task.Status = TaskStatusPending
	
	s.stats.Lock()
	s.stats.totalTasks++
	s.stats.Unlock()
	
	// 如果是延迟任务
	if !task.ScheduledAt.IsZero() && task.ScheduledAt.After(time.Now()) {
		s.delayedTaskMutex.Lock()
		s.delayedTasks = append(s.delayedTasks, task)
		s.delayedTaskMutex.Unlock()
		logger.Debug(fmt.Sprintf("Task %s scheduled for %v", task.ID, task.ScheduledAt))
		return nil
	}
	
	// 根据优先级选择队列
	var targetQueue chan *Task
	switch task.Priority {
	case PriorityCritical:
		targetQueue = s.criticalQueue
	case PriorityHigh:
		targetQueue = s.highQueue
	case PriorityNormal:
		targetQueue = s.normalQueue
	case PriorityLow:
		targetQueue = s.lowQueue
	default:
		targetQueue = s.normalQueue
	}
	
	select {
	case targetQueue <- task:
		logger.Debug(fmt.Sprintf("Task %s submitted to queue", task.ID))
		return nil
	default:
		return fmt.Errorf("task queue full")
	}
}

// Start 启动任务调度器
func (s *TaskScheduler) Start() {
	logger.Info(fmt.Sprintf("Starting task scheduler with %d workers", s.workers))
	
	// 启动工作协程
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	
	// 启动延迟任务调度协程
	s.wg.Add(1)
	go s.delayedTaskScheduler()
	
	// 启动结果处理协程
	s.wg.Add(1)
	go s.resultProcessor()
}

// Stop 停止任务调度器
func (s *TaskScheduler) Stop() {
	logger.Info("Stopping task scheduler...")
	s.cancel()
	close(s.stopCh)
	s.wg.Wait()
	close(s.resultChan)
	logger.Info("Task scheduler stopped")
}

// worker 工作协程
func (s *TaskScheduler) worker(id int) {
	defer s.wg.Done()
	logger.Debug(fmt.Sprintf("Task scheduler worker %d started", id))
	
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.ctx.Done():
			return
		case task := <-s.criticalQueue:
			s.executeTask(task)
		case task := <-s.highQueue:
			s.executeTask(task)
		case task := <-s.normalQueue:
			s.executeTask(task)
		case task := <-s.lowQueue:
			s.executeTask(task)
		}
	}
}

// executeTask 执行任务
func (s *TaskScheduler) executeTask(task *Task) {
	startTime := time.Now()
	task.Status = TaskStatusProcessing
	
	// 记录活跃任务
	s.taskMutex.Lock()
	s.activeTasks[task.ID] = task
	s.taskMutex.Unlock()
	
	defer func() {
		// 移除活跃任务
		s.taskMutex.Lock()
		delete(s.activeTasks, task.ID)
		s.taskMutex.Unlock()
	}()
	
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()
	
	// 执行任务
	err := task.Handler(ctx, task)
	
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	
	result := &TaskResult{
		TaskID:    task.ID,
		Duration:  duration,
		StartTime: startTime,
		EndTime:   endTime,
	}
	
	if err != nil {
		task.Error = err
		
		// 重试逻辑
		if task.MaxRetries > 0 && task.RetryCount < task.MaxRetries {
			task.RetryCount++
			task.Status = TaskStatusPending
			
			s.stats.Lock()
			s.stats.retriedTasks++
			s.stats.Unlock()
			
			logger.Warn(fmt.Sprintf("Task %s failed, retrying (%d/%d): %v", 
				task.ID, task.RetryCount, task.MaxRetries, err))
			
			// 延迟重试
			task.ScheduledAt = time.Now().Add(time.Duration(task.RetryCount) * time.Second)
			s.delayedTaskMutex.Lock()
			s.delayedTasks = append(s.delayedTasks, task)
			s.delayedTaskMutex.Unlock()
			
			result.Status = TaskStatusPending
		} else {
			task.Status = TaskStatusFailed
			result.Status = TaskStatusFailed
			result.Error = err
			
			s.stats.Lock()
			s.stats.failedTasks++
			s.stats.Unlock()
			
			logger.Error(fmt.Sprintf("Task %s failed: %v (took %v)", task.ID, err, duration))
		}
	} else {
		task.Status = TaskStatusCompleted
		result.Status = TaskStatusCompleted
		
		s.stats.Lock()
		s.stats.completedTasks++
		s.stats.Unlock()
		
		logger.Debug(fmt.Sprintf("Task %s completed (took %v)", task.ID, duration))
	}
	
	// 发送结果
	select {
	case s.resultChan <- result:
	default:
		logger.Warn("Result channel full, dropping result")
	}
}

// delayedTaskScheduler 延迟任务调度器
func (s *TaskScheduler) delayedTaskScheduler() {
	defer s.wg.Done()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkDelayedTasks()
		}
	}
}

// checkDelayedTasks 检查延迟任务
func (s *TaskScheduler) checkDelayedTasks() {
	s.delayedTaskMutex.Lock()
	defer s.delayedTaskMutex.Unlock()
	
	now := time.Now()
	remainingTasks := make([]*Task, 0)
	
	for _, task := range s.delayedTasks {
		if task.ScheduledAt.Before(now) || task.ScheduledAt.Equal(now) {
			// 任务到期，提交到队列
			if err := s.SubmitTask(task); err != nil {
				logger.Error(fmt.Sprintf("Failed to submit delayed task %s: %v", task.ID, err))
				remainingTasks = append(remainingTasks, task)
			}
		} else {
			remainingTasks = append(remainingTasks, task)
		}
	}
	
	s.delayedTasks = remainingTasks
}

// resultProcessor 结果处理器
func (s *TaskScheduler) resultProcessor() {
	defer s.wg.Done()
	
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.ctx.Done():
			return
		case result, ok := <-s.resultChan:
			if !ok {
				return
			}
			// 这里可以添加结果持久化、通知等逻辑
			logger.Debug(fmt.Sprintf("Task result: %s - %v", result.TaskID, result.Status))
		}
	}
}

// CancelTask 取消任务
func (s *TaskScheduler) CancelTask(taskID string) error {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()
	
	task, exists := s.activeTasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	task.Status = TaskStatusCancelled
	
	s.stats.Lock()
	s.stats.cancelledTasks++
	s.stats.Unlock()
	
	logger.Info(fmt.Sprintf("Task %s cancelled", taskID))
	return nil
}

// GetTaskStatus 获取任务状态
func (s *TaskScheduler) GetTaskStatus(taskID string) (TaskStatus, error) {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()
	
	task, exists := s.activeTasks[taskID]
	if !exists {
		return TaskStatusPending, fmt.Errorf("task not found: %s", taskID)
	}
	
	return task.Status, nil
}

// GetStats 获取统计信息
func (s *TaskScheduler) GetStats() map[string]int64 {
	s.stats.RLock()
	defer s.stats.RUnlock()
	
	s.taskMutex.RLock()
	activeTasks := int64(len(s.activeTasks))
	s.taskMutex.RUnlock()
	
	s.delayedTaskMutex.RLock()
	delayedTasks := int64(len(s.delayedTasks))
	s.delayedTaskMutex.RUnlock()
	
	return map[string]int64{
		"total_tasks":     s.stats.totalTasks,
		"completed_tasks": s.stats.completedTasks,
		"failed_tasks":    s.stats.failedTasks,
		"cancelled_tasks": s.stats.cancelledTasks,
		"retried_tasks":   s.stats.retriedTasks,
		"active_tasks":    activeTasks,
		"delayed_tasks":   delayedTasks,
	}
}

