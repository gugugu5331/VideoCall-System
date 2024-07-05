package zmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
)

// Message 标准化消息格式，兼容Edge-LLM-Infra
type Message struct {
	RequestID string      `json:"request_id"`
	WorkID    string      `json:"work_id"`
	Action    string      `json:"action"`
	Object    string      `json:"object"`
	Data      interface{} `json:"data"`
	Error     *string     `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ZMQBridge ZMQ通信桥接器，与Edge-LLM-Infra兼容
type ZMQBridge struct {
	context    *zmq4.Context
	publisher  *zmq4.Socket // PUB模式，向C++节点发布消息
	subscriber *zmq4.Socket // SUB模式，订阅C++节点消息
	requester  *zmq4.Socket // REQ模式，向unit-manager请求服务
	replier    *zmq4.Socket // REP模式，响应C++节点请求

	serviceName string
	endpoints   map[string]string
	callbacks   map[string]MessageHandler
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *logrus.Logger
}

// MessageHandler 消息处理回调函数
type MessageHandler func(*Message) (*Message, error)

// Config ZMQ桥接器配置
type Config struct {
	ServiceName     string            `json:"service_name"`
	UnitManagerURL  string            `json:"unit_manager_url"`
	PublisherURL    string            `json:"publisher_url"`
	SubscriberURL   string            `json:"subscriber_url"`
	ReplierURL      string            `json:"replier_url"`
	Endpoints       map[string]string `json:"endpoints"`
	Timeout         time.Duration     `json:"timeout"`
}

// NewZMQBridge 创建新的ZMQ桥接器
func NewZMQBridge(config *Config, logger *logrus.Logger) (*ZMQBridge, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	bridge := &ZMQBridge{
		serviceName: config.ServiceName,
		endpoints:   config.Endpoints,
		callbacks:   make(map[string]MessageHandler),
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}

	// 创建ZMQ上下文
	context, err := zmq4.NewContext()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create ZMQ context: %w", err)
	}
	bridge.context = context

	// 初始化套接字
	if err := bridge.initSockets(config); err != nil {
		bridge.Close()
		return nil, fmt.Errorf("failed to initialize sockets: %w", err)
	}

	return bridge, nil
}

// initSockets 初始化ZMQ套接字
func (b *ZMQBridge) initSockets(config *Config) error {
	var err error

	// 创建PUB套接字
	if config.PublisherURL != "" {
		b.publisher, err = b.context.NewSocket(zmq4.PUB)
		if err != nil {
			return fmt.Errorf("failed to create PUB socket: %w", err)
		}
		if err = b.publisher.Bind(config.PublisherURL); err != nil {
			return fmt.Errorf("failed to bind PUB socket: %w", err)
		}
		b.logger.Infof("PUB socket bound to %s", config.PublisherURL)
	}

	// 创建SUB套接字
	if config.SubscriberURL != "" {
		b.subscriber, err = b.context.NewSocket(zmq4.SUB)
		if err != nil {
			return fmt.Errorf("failed to create SUB socket: %w", err)
		}
		if err = b.subscriber.Connect(config.SubscriberURL); err != nil {
			return fmt.Errorf("failed to connect SUB socket: %w", err)
		}
		// 订阅所有消息
		if err = b.subscriber.SetSubscribe(""); err != nil {
			return fmt.Errorf("failed to set SUB subscription: %w", err)
		}
		b.logger.Infof("SUB socket connected to %s", config.SubscriberURL)
	}

	// 创建REQ套接字（用于与unit-manager通信）
	if config.UnitManagerURL != "" {
		b.requester, err = b.context.NewSocket(zmq4.REQ)
		if err != nil {
			return fmt.Errorf("failed to create REQ socket: %w", err)
		}
		if err = b.requester.Connect(config.UnitManagerURL); err != nil {
			return fmt.Errorf("failed to connect REQ socket: %w", err)
		}
		b.logger.Infof("REQ socket connected to %s", config.UnitManagerURL)
	}

	// 创建REP套接字（响应C++节点请求）
	if config.ReplierURL != "" {
		b.replier, err = b.context.NewSocket(zmq4.REP)
		if err != nil {
			return fmt.Errorf("failed to create REP socket: %w", err)
		}
		if err = b.replier.Bind(config.ReplierURL); err != nil {
			return fmt.Errorf("failed to bind REP socket: %w", err)
		}
		b.logger.Infof("REP socket bound to %s", config.ReplierURL)
	}

	return nil
}

// RegisterService 向unit-manager注册服务
func (b *ZMQBridge) RegisterService() error {
	if b.requester == nil {
		return fmt.Errorf("REQ socket not initialized")
	}

	serviceInfo := map[string]interface{}{
		"service_name": b.serviceName,
		"service_type": "go_service",
		"endpoints":    b.endpoints,
		"health":       "healthy",
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	msg := &Message{
		RequestID: uuid.New().String(),
		WorkID:    b.serviceName,
		Action:    "register_unit",
		Object:    "service",
		Data:      serviceInfo,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return b.SendRequest(msg)
}

// RegisterHandler 注册消息处理器
func (b *ZMQBridge) RegisterHandler(action string, handler MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.callbacks[action] = handler
	b.logger.Infof("Registered handler for action: %s", action)
}

// SendRequest 发送请求到unit-manager
func (b *ZMQBridge) SendRequest(msg *Message) error {
	if b.requester == nil {
		return fmt.Errorf("REQ socket not initialized")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发送action
	if _, err = b.requester.Send(msg.Action, zmq4.SNDMORE); err != nil {
		return fmt.Errorf("failed to send action: %w", err)
	}

	// 发送数据
	if _, err = b.requester.Send(string(data), 0); err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	// 接收响应
	response, err := b.requester.Recv(0)
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	b.logger.Debugf("Received response: %s", response)
	return nil
}

// PublishMessage 发布消息
func (b *ZMQBridge) PublishMessage(msg *Message) error {
	if b.publisher == nil {
		return fmt.Errorf("PUB socket not initialized")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err = b.publisher.Send(string(data), 0); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	b.logger.Debugf("Published message: %s", string(data))
	return nil
}

// StartListening 开始监听消息
func (b *ZMQBridge) StartListening() error {
	// 启动订阅者监听
	if b.subscriber != nil {
		go b.subscribeLoop()
	}

	// 启动响应者监听
	if b.replier != nil {
		go b.replyLoop()
	}

	return nil
}

// subscribeLoop 订阅消息循环
func (b *ZMQBridge) subscribeLoop() {
	b.logger.Info("Starting subscriber loop")
	
	for {
		select {
		case <-b.ctx.Done():
			b.logger.Info("Subscriber loop stopped")
			return
		default:
			data, err := b.subscriber.Recv(zmq4.DONTWAIT)
			if err != nil {
				if err == zmq4.EAGAIN {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				b.logger.Errorf("Failed to receive message: %v", err)
				continue
			}

			var msg Message
			if err := json.Unmarshal([]byte(data), &msg); err != nil {
				b.logger.Errorf("Failed to unmarshal message: %v", err)
				continue
			}

			b.handleMessage(&msg)
		}
	}
}

// replyLoop 响应消息循环
func (b *ZMQBridge) replyLoop() {
	b.logger.Info("Starting replier loop")
	
	for {
		select {
		case <-b.ctx.Done():
			b.logger.Info("Replier loop stopped")
			return
		default:
			data, err := b.replier.Recv(zmq4.DONTWAIT)
			if err != nil {
				if err == zmq4.EAGAIN {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				b.logger.Errorf("Failed to receive request: %v", err)
				continue
			}

			var msg Message
			if err := json.Unmarshal([]byte(data), &msg); err != nil {
				b.logger.Errorf("Failed to unmarshal request: %v", err)
				// 发送错误响应
				errorMsg := "Invalid JSON format"
				b.replier.Send(errorMsg, 0)
				continue
			}

			response := b.handleMessage(&msg)
			
			responseData, err := json.Marshal(response)
			if err != nil {
				b.logger.Errorf("Failed to marshal response: %v", err)
				errorMsg := "Internal server error"
				b.replier.Send(errorMsg, 0)
				continue
			}

			if _, err = b.replier.Send(string(responseData), 0); err != nil {
				b.logger.Errorf("Failed to send response: %v", err)
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (b *ZMQBridge) handleMessage(msg *Message) *Message {
	b.mu.RLock()
	handler, exists := b.callbacks[msg.Action]
	b.mu.RUnlock()

	if !exists {
		b.logger.Warnf("No handler found for action: %s", msg.Action)
		errorMsg := fmt.Sprintf("Unknown action: %s", msg.Action)
		return &Message{
			RequestID: msg.RequestID,
			WorkID:    b.serviceName,
			Action:    msg.Action,
			Object:    msg.Object,
			Error:     &errorMsg,
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	response, err := handler(msg)
	if err != nil {
		b.logger.Errorf("Handler error for action %s: %v", msg.Action, err)
		errorMsg := err.Error()
		return &Message{
			RequestID: msg.RequestID,
			WorkID:    b.serviceName,
			Action:    msg.Action,
			Object:    msg.Object,
			Error:     &errorMsg,
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	if response == nil {
		response = &Message{
			RequestID: msg.RequestID,
			WorkID:    b.serviceName,
			Action:    msg.Action,
			Object:    msg.Object,
			Data:      "OK",
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return response
}

// Close 关闭ZMQ桥接器
func (b *ZMQBridge) Close() error {
	b.cancel()

	var errors []error

	if b.publisher != nil {
		if err := b.publisher.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if b.subscriber != nil {
		if err := b.subscriber.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if b.requester != nil {
		if err := b.requester.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if b.replier != nil {
		if err := b.replier.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if b.context != nil {
		if err := b.context.Term(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing ZMQ bridge: %v", errors)
	}

	b.logger.Info("ZMQ bridge closed successfully")
	return nil
}
