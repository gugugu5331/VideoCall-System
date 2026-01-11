package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"meeting-system/shared/logger"
)

// LocalEvent 本地事件
type LocalEvent struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp int64                  `json:"timestamp"`
	Source    string                 `json:"source"`
}

// LocalEventHandler 本地事件处理函数
type LocalEventHandler func(ctx context.Context, event *LocalEvent) error

// LocalEventBus 本地事件总线（基于Go Channel，用于单个服务内的高性能事件分发）
type LocalEventBus struct {
	handlers     map[string][]LocalEventHandler
	handlerMutex sync.RWMutex
	eventChan    chan *LocalEvent
	bufferSize   int
	workers      int
	stopCh       chan struct{}
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc

	// 统计信息
	stats struct {
		sync.RWMutex
		totalEvents     int64
		processedEvents int64
		failedEvents    int64
		droppedEvents   int64
	}
}

// NewLocalEventBus 创建本地事件总线
func NewLocalEventBus(bufferSize, workers int) *LocalEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &LocalEventBus{
		handlers:   make(map[string][]LocalEventHandler),
		eventChan:  make(chan *LocalEvent, bufferSize),
		bufferSize: bufferSize,
		workers:    workers,
		stopCh:     make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// On 注册事件处理器
func (bus *LocalEventBus) On(eventType string, handler LocalEventHandler) {
	bus.handlerMutex.Lock()
	defer bus.handlerMutex.Unlock()

	if _, exists := bus.handlers[eventType]; !exists {
		bus.handlers[eventType] = make([]LocalEventHandler, 0)
	}
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)

	logger.Debug(fmt.Sprintf("Registered local event handler for: %s", eventType))
}

// Emit 触发事件（非阻塞）
func (bus *LocalEventBus) Emit(eventType string, payload map[string]interface{}, source string) error {
	event := &LocalEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
		Source:    source,
	}

	bus.stats.Lock()
	bus.stats.totalEvents++
	bus.stats.Unlock()

	select {
	case bus.eventChan <- event:
		return nil
	default:
		// 队列已满，丢弃事件
		bus.stats.Lock()
		bus.stats.droppedEvents++
		bus.stats.Unlock()
		return fmt.Errorf("event queue full, event dropped")
	}
}

// EmitSync 触发事件（同步，阻塞直到事件被放入队列）
func (bus *LocalEventBus) EmitSync(ctx context.Context, eventType string, payload map[string]interface{}, source string) error {
	event := &LocalEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
		Source:    source,
	}

	bus.stats.Lock()
	bus.stats.totalEvents++
	bus.stats.Unlock()

	select {
	case bus.eventChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-bus.ctx.Done():
		return fmt.Errorf("event bus stopped")
	}
}

// Start 启动事件总线
func (bus *LocalEventBus) Start() {
	logger.Info(fmt.Sprintf("Starting local event bus with %d workers", bus.workers))

	for i := 0; i < bus.workers; i++ {
		bus.wg.Add(1)
		go bus.worker(i)
	}
}

// Stop 停止事件总线
func (bus *LocalEventBus) Stop() {
	logger.Info("Stopping local event bus...")
	bus.cancel()
	close(bus.stopCh)

	// 等待所有事件处理完成
	bus.wg.Wait()

	// 关闭事件通道
	close(bus.eventChan)

	logger.Info("Local event bus stopped")
}

// worker 工作协程
func (bus *LocalEventBus) worker(id int) {
	defer bus.wg.Done()
	logger.Debug(fmt.Sprintf("Local event bus worker %d started", id))

	for {
		select {
		case <-bus.stopCh:
			logger.Debug(fmt.Sprintf("Local event bus worker %d stopped", id))
			return
		case <-bus.ctx.Done():
			logger.Debug(fmt.Sprintf("Local event bus worker %d context cancelled", id))
			return
		case event, ok := <-bus.eventChan:
			if !ok {
				logger.Debug(fmt.Sprintf("Local event bus worker %d channel closed", id))
				return
			}
			bus.processEvent(event)
		}
	}
}

// processEvent 处理事件
func (bus *LocalEventBus) processEvent(event *LocalEvent) {
	startTime := time.Now()

	// 获取事件处理器
	bus.handlerMutex.RLock()
	handlers, exists := bus.handlers[event.Type]
	bus.handlerMutex.RUnlock()

	if !exists || len(handlers) == 0 {
		logger.Debug(fmt.Sprintf("No handlers for event type: %s", event.Type))
		return
	}

	// 并发执行所有处理器
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h LocalEventHandler) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Error(fmt.Sprintf("Event handler panic: %v", r))
					bus.stats.Lock()
					bus.stats.failedEvents++
					bus.stats.Unlock()
				}
			}()

			ctx, cancel := context.WithTimeout(bus.ctx, 10*time.Second)
			defer cancel()

			if err := h(ctx, event); err != nil {
				logger.Error(fmt.Sprintf("Event handler error for %s: %v", event.Type, err))
				bus.stats.Lock()
				bus.stats.failedEvents++
				bus.stats.Unlock()
			}
		}(handler)
	}

	wg.Wait()

	bus.stats.Lock()
	bus.stats.processedEvents++
	bus.stats.Unlock()

	duration := time.Since(startTime)
	if duration > 100*time.Millisecond {
		logger.Warn(fmt.Sprintf("Event %s processing took %v", event.Type, duration))
	}
}

// GetStats 获取统计信息
func (bus *LocalEventBus) GetStats() map[string]int64 {
	bus.stats.RLock()
	defer bus.stats.RUnlock()

	return map[string]int64{
		"total_events":     bus.stats.totalEvents,
		"processed_events": bus.stats.processedEvents,
		"failed_events":    bus.stats.failedEvents,
		"dropped_events":   bus.stats.droppedEvents,
		"pending_events":   int64(len(bus.eventChan)),
		"buffer_size":      int64(bus.bufferSize),
	}
}

// PriorityLocalEventBus 支持优先级的本地事件总线
type PriorityLocalEventBus struct {
	handlers     map[string][]LocalEventHandler
	handlerMutex sync.RWMutex

	// 不同优先级的事件通道
	criticalChan chan *LocalEvent
	highChan     chan *LocalEvent
	normalChan   chan *LocalEvent
	lowChan      chan *LocalEvent

	workers int
	stopCh  chan struct{}
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewPriorityLocalEventBus 创建支持优先级的本地事件总线
func NewPriorityLocalEventBus(bufferSize, workers int) *PriorityLocalEventBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &PriorityLocalEventBus{
		handlers:     make(map[string][]LocalEventHandler),
		criticalChan: make(chan *LocalEvent, bufferSize/4),
		highChan:     make(chan *LocalEvent, bufferSize/4),
		normalChan:   make(chan *LocalEvent, bufferSize/2),
		lowChan:      make(chan *LocalEvent, bufferSize/4),
		workers:      workers,
		stopCh:       make(chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// On 注册事件处理器
func (bus *PriorityLocalEventBus) On(eventType string, handler LocalEventHandler) {
	bus.handlerMutex.Lock()
	defer bus.handlerMutex.Unlock()

	if _, exists := bus.handlers[eventType]; !exists {
		bus.handlers[eventType] = make([]LocalEventHandler, 0)
	}
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
}

// EmitWithPriority 触发带优先级的事件
func (bus *PriorityLocalEventBus) EmitWithPriority(priority MessagePriority, eventType string, payload map[string]interface{}, source string) error {
	event := &LocalEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UnixNano(),
		Source:    source,
	}

	var targetChan chan *LocalEvent
	switch priority {
	case PriorityCritical:
		targetChan = bus.criticalChan
	case PriorityHigh:
		targetChan = bus.highChan
	case PriorityNormal:
		targetChan = bus.normalChan
	case PriorityLow:
		targetChan = bus.lowChan
	default:
		targetChan = bus.normalChan
	}

	select {
	case targetChan <- event:
		return nil
	default:
		return fmt.Errorf("priority queue full, event dropped")
	}
}

// Start 启动优先级事件总线
func (bus *PriorityLocalEventBus) Start() {
	logger.Info(fmt.Sprintf("Starting priority local event bus with %d workers", bus.workers))

	for i := 0; i < bus.workers; i++ {
		bus.wg.Add(1)
		go bus.priorityWorker(i)
	}
}

// Stop 停止优先级事件总线
func (bus *PriorityLocalEventBus) Stop() {
	logger.Info("Stopping priority local event bus...")
	bus.cancel()
	close(bus.stopCh)
	bus.wg.Wait()
	logger.Info("Priority local event bus stopped")
}

// priorityWorker 优先级工作协程
func (bus *PriorityLocalEventBus) priorityWorker(id int) {
	defer bus.wg.Done()

	for {
		// 尝试非阻塞地按优先级获取事件
		if bus.tryProcess(bus.criticalChan) {
			continue
		}
		if bus.tryProcess(bus.highChan) {
			continue
		}
		if bus.tryProcess(bus.normalChan) {
			continue
		}
		if bus.tryProcess(bus.lowChan) {
			continue
		}

		// 阻塞等待下一个事件或停止信号
		select {
		case <-bus.stopCh:
			return
		case <-bus.ctx.Done():
			return
		case event := <-bus.criticalChan:
			bus.processEvent(event)
		case event := <-bus.highChan:
			bus.processEvent(event)
		case event := <-bus.normalChan:
			bus.processEvent(event)
		case event := <-bus.lowChan:
			bus.processEvent(event)
		}
	}
}

func (bus *PriorityLocalEventBus) tryProcess(ch chan *LocalEvent) bool {
	select {
	case event := <-ch:
		bus.processEvent(event)
		return true
	default:
		return false
	}
}

// processEvent 处理事件
func (bus *PriorityLocalEventBus) processEvent(event *LocalEvent) {
	bus.handlerMutex.RLock()
	handlers, exists := bus.handlers[event.Type]
	bus.handlerMutex.RUnlock()

	if !exists || len(handlers) == 0 {
		return
	}

	for _, handler := range handlers {
		ctx, cancel := context.WithTimeout(bus.ctx, 10*time.Second)
		if err := handler(ctx, event); err != nil {
			logger.Error(fmt.Sprintf("Priority event handler error: %v", err))
		}
		cancel()
	}
}
