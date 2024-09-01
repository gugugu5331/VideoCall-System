package zmq

import (
	"context"
	"fmt"
	"time"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// ZMQClient ZeroMQ客户端，用于与Edge-LLM-Infra通信
type ZMQClient struct {
	config  config.ZMQConfig
	timeout time.Duration
}

// AIRequest AI推理请求
type AIRequest struct {
	RequestID string      `json:"request_id"`
	WorkID    string      `json:"work_id"`
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

// NewZMQClient 创建ZMQ客户端
func NewZMQClient(config config.ZMQConfig) (*ZMQClient, error) {
	client := &ZMQClient{
		config:  config,
		timeout: time.Duration(config.Timeout) * time.Second,
	}

	// TODO: 实际连接ZMQ
	logger.Info("ZMQ client created (mock mode)")

	return client, nil
}

// Close 关闭ZMQ客户端
func (c *ZMQClient) Close() {
	// TODO: 实际关闭ZMQ连接
}

// SendRequest 发送AI推理请求
func (c *ZMQClient) SendRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// TODO: 实际发送ZMQ请求
	// 暂时返回模拟响应
	response := &AIResponse{
		RequestID: request.RequestID,
		WorkID:    request.WorkID,
		Object:    request.Object,
		Data:      map[string]interface{}{"status": "mock_response"},
		Error:     nil,
	}

	return response, nil
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
	request := &AIRequest{
		RequestID: fmt.Sprintf("health_%d", time.Now().Unix()),
		WorkID:    "sys",
		Object:    "health_check",
		Data:      c.config.UnitName,
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
