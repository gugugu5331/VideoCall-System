package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// AIClient AI服务客户端
type AIClient struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
}

// AIRequest AI推理请求
type AIRequest struct {
	RequestID string                 `json:"request_id,omitempty"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
}

// AIResponse AI推理响应
type AIResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// AudioData 音频数据结构
type AudioData struct {
	Data       []byte `json:"-"`      // 原始音频数据
	Format     string `json:"format"` // 音频格式 (wav, mp3, etc.)
	SampleRate int    `json:"sample_rate"`
	Channels   int    `json:"channels"`
	Duration   int    `json:"duration"` // 持续时间(毫秒)
}

// VideoData 视频数据结构
type VideoData struct {
	Data     []byte `json:"-"`      // 原始视频数据
	Format   string `json:"format"` // 视频格式 (mp4, avi, etc.)
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FPS      int    `json:"fps"`
	Duration int    `json:"duration"` // 持续时间(毫秒)
}

// NewAIClient 创建AI客户端
func NewAIClient(config *config.Config) *AIClient {
	host := config.Services.AIService.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := config.Services.AIService.Port
	if port == 0 {
		port = 8084
	}

	baseURL := fmt.Sprintf("http://%s:%d", host, port)

	return &AIClient{
		config:  config,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SpeechRecognition 语音识别
func (c *AIClient) SpeechRecognition(ctx context.Context, audioData *AudioData) (*AIResponse, error) {
	logger.Info("Calling AI service for speech recognition")

	// 将音频数据编码为base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData.Data)

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "speech_recognition",
		Data: map[string]interface{}{
			"audio_data":   audioBase64,
			"audio_format": audioData.Format,
			"sample_rate":  audioData.SampleRate,
			"channels":     audioData.Channels,
			"duration":     audioData.Duration,
		},
	}

	return c.sendRequest(ctx, "/api/v1/speech/recognition", request)
}

// EmotionDetection 情绪检测
func (c *AIClient) EmotionDetection(ctx context.Context, audioData *AudioData) (*AIResponse, error) {
	logger.Info("Calling AI service for emotion detection")

	audioBase64 := base64.StdEncoding.EncodeToString(audioData.Data)

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "emotion_detection",
		Data: map[string]interface{}{
			"audio_data":   audioBase64,
			"audio_format": audioData.Format,
			"sample_rate":  audioData.SampleRate,
			"channels":     audioData.Channels,
			"duration":     audioData.Duration,
		},
	}

	return c.sendRequest(ctx, "/api/v1/speech/emotion", request)
}

// SynthesisDetection 合成检测（Deepfake检测）
func (c *AIClient) SynthesisDetection(ctx context.Context, audioData *AudioData, videoData *VideoData) (*AIResponse, error) {
	logger.Info("Calling AI service for synthesis detection")

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "synthesis_detection",
		Data:      make(map[string]interface{}),
	}

	// 添加音频数据（如果有）
	if audioData != nil {
		audioBase64 := base64.StdEncoding.EncodeToString(audioData.Data)
		request.Data["audio_data"] = audioBase64
		request.Data["audio_format"] = audioData.Format
		request.Data["sample_rate"] = audioData.SampleRate
		request.Data["channels"] = audioData.Channels
	}

	// 添加视频数据（如果有）
	if videoData != nil {
		videoBase64 := base64.StdEncoding.EncodeToString(videoData.Data)
		request.Data["video_data"] = videoBase64
		request.Data["video_format"] = videoData.Format
		request.Data["width"] = videoData.Width
		request.Data["height"] = videoData.Height
		request.Data["fps"] = videoData.FPS
	}

	return c.sendRequest(ctx, "/api/v1/speech/synthesis-detection", request)
}

// AudioDenoising 音频降噪
func (c *AIClient) AudioDenoising(ctx context.Context, audioData *AudioData) (*AIResponse, error) {
	logger.Info("Calling AI service for audio denoising")

	audioBase64 := base64.StdEncoding.EncodeToString(audioData.Data)

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "audio_denoising",
		Data: map[string]interface{}{
			"audio_data":   audioBase64,
			"audio_format": audioData.Format,
			"sample_rate":  audioData.SampleRate,
			"channels":     audioData.Channels,
			"duration":     audioData.Duration,
		},
	}

	return c.sendRequest(ctx, "/api/v1/audio/denoising", request)
}

// VideoEnhancement 视频增强
func (c *AIClient) VideoEnhancement(ctx context.Context, videoData *VideoData) (*AIResponse, error) {
	logger.Info("Calling AI service for video enhancement")

	videoBase64 := base64.StdEncoding.EncodeToString(videoData.Data)

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "video_enhancement",
		Data: map[string]interface{}{
			"video_data":   videoBase64,
			"video_format": videoData.Format,
			"width":        videoData.Width,
			"height":       videoData.Height,
			"fps":          videoData.FPS,
			"duration":     videoData.Duration,
		},
	}

	return c.sendRequest(ctx, "/api/v1/video/enhancement", request)
}

// HealthCheck 检查 HTTP AI 服务健康状况
func (c *AIClient) HealthCheck(ctx context.Context) error {
	if c == nil || c.httpClient == nil {
		return fmt.Errorf("ai http client not initialized")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/health", c.baseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to build health check request: %w", err)
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("ai http health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ai http health check unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BaseURL 返回 AI 服务的基础地址
func (c *AIClient) BaseURL() string {
	if c == nil {
		return ""
	}
	return c.baseURL
}

// BatchProcessing 批量处理音视频数据
func (c *AIClient) BatchProcessing(ctx context.Context, audioData *AudioData, videoData *VideoData, tasks []string) (map[string]*AIResponse, error) {
	logger.Info(fmt.Sprintf("Batch processing with tasks: %v", tasks))

	results := make(map[string]*AIResponse)
	errors := make([]error, 0)

	// 并发执行多个AI任务
	type taskResult struct {
		task     string
		response *AIResponse
		err      error
	}

	resultChan := make(chan taskResult, len(tasks))

	for _, task := range tasks {
		go func(taskType string) {
			var response *AIResponse
			var err error

			switch taskType {
			case "speech_recognition":
				if audioData != nil {
					response, err = c.SpeechRecognition(ctx, audioData)
				} else {
					err = fmt.Errorf("audio data required for speech recognition")
				}
			case "emotion_detection":
				if audioData != nil {
					response, err = c.EmotionDetection(ctx, audioData)
				} else {
					err = fmt.Errorf("audio data required for emotion detection")
				}
			case "synthesis_detection":
				response, err = c.SynthesisDetection(ctx, audioData, videoData)
			case "audio_denoising":
				if audioData != nil {
					response, err = c.AudioDenoising(ctx, audioData)
				} else {
					err = fmt.Errorf("audio data required for audio denoising")
				}
			case "video_enhancement":
				if videoData != nil {
					response, err = c.VideoEnhancement(ctx, videoData)
				} else {
					err = fmt.Errorf("video data required for video enhancement")
				}
			default:
				err = fmt.Errorf("unknown task type: %s", taskType)
			}

			resultChan <- taskResult{task: taskType, response: response, err: err}
		}(task)
	}

	// 收集结果
	for i := 0; i < len(tasks); i++ {
		result := <-resultChan
		if result.err != nil {
			errors = append(errors, fmt.Errorf("task %s failed: %v", result.task, result.err))
			logger.Error(fmt.Sprintf("Batch task %s failed: %v", result.task, result.err))
		} else {
			results[result.task] = result.response
		}
	}

	if len(errors) > 0 {
		logger.Warn(fmt.Sprintf("Batch processing completed with %d errors", len(errors)))
	}

	return results, nil
}

// sendRequest 发送HTTP请求到AI服务
func (c *AIClient) sendRequest(ctx context.Context, endpoint string, request *AIRequest) (*AIResponse, error) {
	url := c.baseURL + endpoint

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var aiResponse AIResponse
	if err := json.Unmarshal(responseBody, &aiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &aiResponse, fmt.Errorf("AI service returned error: %d - %s", resp.StatusCode, aiResponse.Message)
	}

	logger.Debug(fmt.Sprintf("AI service request successful: %s", endpoint))
	return &aiResponse, nil
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("media_ai_%d", time.Now().UnixNano())
}
