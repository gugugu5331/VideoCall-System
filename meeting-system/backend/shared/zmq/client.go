package zmq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pebbe/zmq4"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// ZMQClient ZeroMQ客户端，用于与Edge-LLM-Infra通信
type ZMQClient struct {
	config      config.ZMQConfig
	timeout     time.Duration
	socket      *zmq4.Socket
	context     *zmq4.Context
	mutex       sync.Mutex
	connected   bool
	heartbeatCh chan bool
	stopCh      chan bool
}

// AIRequest AI推理请求
type AIRequest struct {
	RequestID string      `json:"request_id"`
	WorkID    string      `json:"work_id"`
	Action    string      `json:"action"` // 添加 action 字段
	Object    string      `json:"object"`
	Data      interface{} `json:"data"`
}

// AIResponse AI推理响应
type AIResponse struct {
	RequestID string      `json:"request_id"`
	WorkID    string      `json:"work_id"`
	Object    string      `json:"object"`
	Data      interface{} `json:"data"`
	Error     *string     `json:"error"`
}

// SpeechRecognitionData 语音识别数据
type SpeechRecognitionData struct {
	AudioFormat string `json:"audio_format"`
	SampleRate  int    `json:"sample_rate"`
	Channels    int    `json:"channels"`
	AudioData   string `json:"audio_data"` // base64编码
}

// EmotionDetectionData 情绪识别数据
type EmotionDetectionData struct {
	ImageFormat string `json:"image_format"`
	ImageData   string `json:"image_data"` // base64编码
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// AudioDenoisingData 音频降噪数据
type AudioDenoisingData struct {
	AudioFormat string `json:"audio_format"`
	SampleRate  int    `json:"sample_rate"`
	Channels    int    `json:"channels"`
	AudioData   string `json:"audio_data"` // base64编码
}

// VideoEnhancementData 视频增强数据
type VideoEnhancementData struct {
	VideoFormat string `json:"video_format"`
	VideoData   string `json:"video_data"` // base64编码
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FPS         int    `json:"fps"`
}

// SynthesisDetectionData 合成检测数据（Deepfake检测）
type SynthesisDetectionData struct {
	AudioData   string `json:"audio_data,omitempty"`   // base64编码
	VideoData   string `json:"video_data,omitempty"`   // base64编码
	AudioFormat string `json:"audio_format,omitempty"` // wav, mp3, etc.
	VideoFormat string `json:"video_format,omitempty"` // mp4, avi, etc.
	SampleRate  int    `json:"sample_rate,omitempty"`  // 音频采样率
	Width       int    `json:"width,omitempty"`        // 视频宽度
	Height      int    `json:"height,omitempty"`       // 视频高度
	FPS         int    `json:"fps,omitempty"`          // 视频帧率
	Channels    int    `json:"channels,omitempty"`     // 音频通道数
}

// AITask 通用 AI 任务（通过 ZeroMQ 发送到 C++ Unit Manager）
type AITask struct {
	TaskID    string            `json:"task_id"`
	TaskType  string            `json:"task_type"`  // "speech_recognition", "emotion_detection", "deepfake_detection" 等
	ModelPath string            `json:"model_path"` // ONNX 模型路径，如 "/models/whisper_base.onnx"
	InputData []byte            `json:"input_data"` // 原始输入数据（如音频/视频二进制数据）
	Params    map[string]string `json:"params"`     // 额外参数（如采样率、语言等）
}

// SendTask 发送通用 AI 任务
func (c *ZMQClient) SendTask(ctx context.Context, task *AITask) (*AIResponse, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	req := &AIRequest{
		RequestID: task.TaskID,
		WorkID:    c.config.UnitName,
		Action:    "inference",
		Object:    "ai_task",
		Data:      task,
	}
	return c.SendRequest(ctx, req)
}

// NewZMQClient 创建ZMQ客户端
func NewZMQClient(config config.ZMQConfig) (*ZMQClient, error) {
	fmt.Println("[ZMQ] Creating ZMQ context...")
	// 创建ZMQ上下文
	context, err := zmq4.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create ZMQ context: %w", err)
	}
	fmt.Println("[ZMQ] ZMQ context created")

	client := &ZMQClient{
		config:      config,
		timeout:     time.Duration(config.Timeout) * time.Second,
		context:     context,
		heartbeatCh: make(chan bool, 1),
		stopCh:      make(chan bool, 1),
	}

	// 建立连接
	fmt.Printf("[ZMQ] Connecting to Edge-LLM-Infra at %s:%d...\n", config.UnitManagerHost, config.UnitManagerPort)
	if err := client.connect(); err != nil {
		fmt.Printf("[ZMQ] Connection failed: %v\n", err)
		context.Term()
		return nil, fmt.Errorf("failed to connect to ZMQ: %w", err)
	}
	fmt.Println("[ZMQ] Connection established")

	// 启动心跳检测
	fmt.Println("[ZMQ] Starting heartbeat loop...")
	go client.heartbeatLoop()

	logger.Info("ZMQ client created and connected successfully")
	return client, nil
}

// connect 建立ZMQ连接
func (c *ZMQClient) connect() error {
	fmt.Println("[ZMQ] Acquiring connection lock...")
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 创建REQ socket用于与Edge-LLM-Infra通信
	fmt.Println("[ZMQ] Creating REQ socket...")
	socket, err := c.context.NewSocket(zmq4.REQ)
	if err != nil {
		return fmt.Errorf("failed to create socket: %w", err)
	}
	fmt.Println("[ZMQ] REQ socket created")

	// 设置socket选项
	fmt.Printf("[ZMQ] Setting socket timeout to %v...\n", c.timeout)
	socket.SetRcvtimeo(c.timeout)
	socket.SetSndtimeo(c.timeout)
	socket.SetLinger(1000) // 1秒linger时间

	// 连接到Edge-LLM-Infra的unit-manager
	endpoint := fmt.Sprintf("tcp://%s:%d", c.config.UnitManagerHost, c.config.UnitManagerPort)
	fmt.Printf("[ZMQ] Connecting to endpoint: %s...\n", endpoint)
	if err := socket.Connect(endpoint); err != nil {
		fmt.Printf("[ZMQ] Connect failed: %v\n", err)
		socket.Close()
		return fmt.Errorf("failed to connect to %s: %w", endpoint, err)
	}
	fmt.Printf("[ZMQ] Successfully connected to %s\n", endpoint)

	c.socket = socket
	c.connected = true
	logger.Info("Connected to Edge-LLM-Infra at " + endpoint)
	return nil
}

// reconnect 重新连接
func (c *ZMQClient) reconnect() error {
	logger.Warn("Attempting to reconnect to Edge-LLM-Infra...")

	// 关闭现有连接
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
		c.connected = false
	}

	// 重新连接
	return c.connect()
}

// Close 关闭ZMQ客户端
func (c *ZMQClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 停止心跳
	select {
	case c.stopCh <- true:
	default:
	}

	// 关闭socket
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
	}

	// 关闭context
	if c.context != nil {
		c.context.Term()
		c.context = nil
	}

	c.connected = false
	logger.Info("ZMQ client closed")
}

// heartbeatLoop 心跳检测循环
func (c *ZMQClient) heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.HealthCheck(context.Background()); err != nil {
				logger.Warn("Heartbeat failed, attempting reconnection: " + err.Error())
				if err := c.reconnect(); err != nil {
					logger.Error("Reconnection failed: " + err.Error())
				}
			}
		case <-c.stopCh:
			return
		}
	}
}

// SendRequest 发送AI推理请求
func (c *ZMQClient) SendRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected || c.socket == nil {
		return nil, fmt.Errorf("ZMQ client not connected")
	}

	// 序列化请求
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	previewLen := 200
	if len(requestData) < previewLen {
		previewLen = len(requestData)
	}
	logger.Debug(fmt.Sprintf("Sending request: %s", string(requestData)[:previewLen]))

	// 发送请求 (使用Edge-LLM-Infra的TCP/JSON协议格式)
	if _, err := c.socket.SendBytes(requestData, 0); err != nil {
		// 发送失败，尝试重连
		if reconnectErr := c.reconnect(); reconnectErr != nil {
			return nil, fmt.Errorf("send failed and reconnect failed: %v, %v", err, reconnectErr)
		}
		// 重连成功，重试发送
		if _, err := c.socket.SendBytes(requestData, 0); err != nil {
			return nil, fmt.Errorf("failed to send request after reconnect: %w", err)
		}
	}

	// 接收响应
	responseData, err := c.socket.RecvBytes(0)
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	previewLen = 200
	if len(responseData) < previewLen {
		previewLen = len(responseData)
	}
	logger.Debug(fmt.Sprintf("Received response: %s", string(responseData)[:previewLen]))

	// 反序列化响应
	var response AIResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// SpeechRecognition 语音识别
func (c *ZMQClient) SpeechRecognition(ctx context.Context, requestID string, data *SpeechRecognitionData) (*AIResponse, error) {
	request := &AIRequest{
		RequestID: requestID,
		WorkID:    c.config.UnitName,
		Object:    "speech_recognition",
		Data:      data,
	}

	return c.SendRequest(ctx, request)
}

// EmotionDetection 情绪识别
func (c *ZMQClient) EmotionDetection(ctx context.Context, requestID string, data *EmotionDetectionData) (*AIResponse, error) {
	request := &AIRequest{
		RequestID: requestID,
		WorkID:    c.config.UnitName,
		Object:    "emotion_detection",
		Data:      data,
	}

	return c.SendRequest(ctx, request)
}

// AudioDenoising 音频降噪
func (c *ZMQClient) AudioDenoising(ctx context.Context, requestID string, data *AudioDenoisingData) (*AIResponse, error) {
	request := &AIRequest{
		RequestID: requestID,
		WorkID:    c.config.UnitName,
		Object:    "audio_denoising",
		Data:      data,
	}

	return c.SendRequest(ctx, request)
}

// VideoEnhancement 视频增强
func (c *ZMQClient) VideoEnhancement(ctx context.Context, requestID string, data *VideoEnhancementData) (*AIResponse, error) {
	request := &AIRequest{
		RequestID: requestID,
		WorkID:    c.config.UnitName,
		Object:    "video_enhancement",
		Data:      data,
	}

	return c.SendRequest(ctx, request)
}

// SynthesisDetection 合成检测（Deepfake检测）
func (c *ZMQClient) SynthesisDetection(ctx context.Context, requestID string, data *SynthesisDetectionData) (*AIResponse, error) {
	request := &AIRequest{
		RequestID: requestID,
		WorkID:    c.config.UnitName,
		Object:    "synthesis_detection",
		Data:      data,
	}

	return c.SendRequest(ctx, request)
}

// RegisterUnit 注册AI处理单元
func (c *ZMQClient) RegisterUnit(ctx context.Context) (*AIResponse, error) {
	request := &AIRequest{
		RequestID: fmt.Sprintf("register_%d", time.Now().Unix()),
		WorkID:    "sys",
		Object:    "register_unit",
		Data:      c.config.UnitName,
	}

	return c.SendRequest(ctx, request)
}

// ReleaseUnit 释放AI处理单元
func (c *ZMQClient) ReleaseUnit(ctx context.Context) (*AIResponse, error) {
	request := &AIRequest{
		RequestID: fmt.Sprintf("release_%d", time.Now().Unix()),
		WorkID:    "sys",
		Object:    "release_unit",
		Data:      c.config.UnitName,
	}

	return c.SendRequest(ctx, request)
}

// HealthCheck 健康检查
func (c *ZMQClient) HealthCheck(ctx context.Context) error {
	// 创建健康检查请求，使用Edge-LLM-Infra的协议格式
	request := &AIRequest{
		RequestID: fmt.Sprintf("health_%d", time.Now().Unix()),
		WorkID:    "sys",
		Object:    "health_check",
		Data: map[string]interface{}{
			"unit_name": c.config.UnitName,
			"timestamp": time.Now().Unix(),
		},
	}

	response, err := c.SendRequest(ctx, request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("health check failed: %s", *response.Error)
	}

	return nil
}

// ZMQManager ZMQ连接管理器
type ZMQManager struct {
	client *ZMQClient
	config config.ZMQConfig
}

var globalZMQManager *ZMQManager

// InitZMQ 初始化ZMQ管理器
func InitZMQ(config config.ZMQConfig) error {
	client, err := NewZMQClient(config)
	if err != nil {
		return fmt.Errorf("failed to create ZMQ client: %w", err)
	}

	globalZMQManager = &ZMQManager{
		client: client,
		config: config,
	}

	// 注册AI处理单元
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.RegisterUnit(ctx)
	if err != nil {
		logger.Warn("Failed to register AI unit: " + err.Error())
		// 注册失败不影响启动，可能是Edge-LLM-Infra未启动
	} else {
		logger.Info("AI unit registered successfully")
	}

	return nil
}

// GetZMQClient 获取ZMQ客户端
func GetZMQClient() *ZMQClient {
	if globalZMQManager == nil {
		return nil
	}
	return globalZMQManager.client
}

// CloseZMQ 关闭ZMQ连接
func CloseZMQ() {
	if globalZMQManager != nil && globalZMQManager.client != nil {
		// 释放AI处理单元
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		globalZMQManager.client.ReleaseUnit(ctx)
		globalZMQManager.client.Close()
		globalZMQManager = nil
		logger.Info("ZMQ client closed successfully")
	}
}
