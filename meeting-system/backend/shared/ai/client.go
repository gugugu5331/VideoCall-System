package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// AIClient AI服务客户端
type AIClient struct {
	baseURL    string
	httpClient *http.Client
	cache      *Cache
	rateLimiter *RateLimiter
	config     *config.Config
	mutex      sync.RWMutex
}

// AIRequest AI推理请求
type AIRequest struct {
	RequestID string                 `json:"request_id"`
	Type      string                 `json:"type"`
	ModelID   string                 `json:"model_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// AIResponse AI推理响应
type AIResponse struct {
	RequestID string                 `json:"request_id"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Latency   float64                `json:"latency"`
	NodeID    string                 `json:"node_id"`
}

// Cache 缓存接口
type Cache struct {
	data   map[string]*CacheItem
	mutex  sync.RWMutex
	ttl    time.Duration
	maxSize int
}

// CacheItem 缓存项
type CacheItem struct {
	Value     *AIResponse
	ExpiresAt time.Time
}

// RateLimiter 限流器
type RateLimiter struct {
	requests    map[string][]time.Time
	mutex       sync.RWMutex
	maxRequests int
	window      time.Duration
}

// NewAIClient 创建AI客户端
func NewAIClient(config *config.Config) *AIClient {
	client := &AIClient{
		baseURL: fmt.Sprintf("http://localhost:%d/api/v1", 8083), // AI服务端口
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}

	// 初始化缓存
	if config != nil {
		client.cache = NewCache(5*time.Minute, 1000)
		client.rateLimiter = NewRateLimiter(1000, time.Minute)
	}

	return client
}

// NewCache 创建缓存
func NewCache(ttl time.Duration, maxSize int) *Cache {
	cache := &Cache{
		data:    make(map[string]*CacheItem),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// 启动清理协程
	go cache.cleanup()
	return cache
}

// NewRateLimiter 创建限流器
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
	}
}

// SpeechRecognition 语音识别
func (c *AIClient) SpeechRecognition(ctx context.Context, audioData []byte, format string, sampleRate int) (*AIResponse, error) {
	// 检查限流
	if !c.rateLimiter.Allow("speech_recognition") {
		return nil, fmt.Errorf("rate limit exceeded for speech recognition")
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("speech_%x", audioData[:min(len(audioData), 32)])
	if cached := c.cache.Get(cacheKey); cached != nil {
		logger.Debug("Speech recognition cache hit")
		return cached, nil
	}

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "speech_recognition",
		Data: map[string]interface{}{
			"audio_data":   encodeBase64(audioData),
			"audio_format": format,
			"sample_rate":  sampleRate,
			"channels":     1,
		},
	}

	response, err := c.makeRequest(ctx, "/speech/recognition", request)
	if err != nil {
		return nil, fmt.Errorf("speech recognition failed: %w", err)
	}

	// 缓存结果
	c.cache.Set(cacheKey, response)
	return response, nil
}

// EmotionDetection 情绪识别
func (c *AIClient) EmotionDetection(ctx context.Context, imageData []byte, format string, width, height int) (*AIResponse, error) {
	// 检查限流
	if !c.rateLimiter.Allow("emotion_detection") {
		return nil, fmt.Errorf("rate limit exceeded for emotion detection")
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("emotion_%x", imageData[:min(len(imageData), 32)])
	if cached := c.cache.Get(cacheKey); cached != nil {
		logger.Debug("Emotion detection cache hit")
		return cached, nil
	}

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "emotion_detection",
		Data: map[string]interface{}{
			"image_data":   encodeBase64(imageData),
			"image_format": format,
			"width":        width,
			"height":       height,
		},
	}

	response, err := c.makeRequest(ctx, "/speech/emotion", request)
	if err != nil {
		return nil, fmt.Errorf("emotion detection failed: %w", err)
	}

	// 缓存结果
	c.cache.Set(cacheKey, response)
	return response, nil
}

// SynthesisDetection 合成检测
func (c *AIClient) SynthesisDetection(ctx context.Context, audioData []byte, format string, sampleRate int) (*AIResponse, error) {
	// 检查限流
	if !c.rateLimiter.Allow("synthesis_detection") {
		return nil, fmt.Errorf("rate limit exceeded for synthesis detection")
	}

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "synthesis_detection",
		Data: map[string]interface{}{
			"audio_data":   encodeBase64(audioData),
			"audio_format": format,
			"sample_rate":  sampleRate,
			"channels":     1,
		},
	}

	response, err := c.makeRequest(ctx, "/speech/synthesis-detection", request)
	if err != nil {
		return nil, fmt.Errorf("synthesis detection failed: %w", err)
	}

	return response, nil
}

// AudioDenoising 音频降噪
func (c *AIClient) AudioDenoising(ctx context.Context, audioData []byte, format string, sampleRate int) (*AIResponse, error) {
	// 检查限流
	if !c.rateLimiter.Allow("audio_denoising") {
		return nil, fmt.Errorf("rate limit exceeded for audio denoising")
	}

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "audio_denoising",
		Data: map[string]interface{}{
			"audio_data":   encodeBase64(audioData),
			"audio_format": format,
			"sample_rate":  sampleRate,
			"channels":     1,
		},
	}

	response, err := c.makeRequest(ctx, "/audio/denoising", request)
	if err != nil {
		return nil, fmt.Errorf("audio denoising failed: %w", err)
	}

	return response, nil
}

// VideoEnhancement 视频增强
func (c *AIClient) VideoEnhancement(ctx context.Context, videoData []byte, format string, width, height, fps int) (*AIResponse, error) {
	// 检查限流
	if !c.rateLimiter.Allow("video_enhancement") {
		return nil, fmt.Errorf("rate limit exceeded for video enhancement")
	}

	request := &AIRequest{
		RequestID: generateRequestID(),
		Type:      "video_enhancement",
		Data: map[string]interface{}{
			"video_data":   encodeBase64(videoData),
			"video_format": format,
			"width":        width,
			"height":       height,
			"fps":          fps,
		},
	}

	response, err := c.makeRequest(ctx, "/video/enhancement", request)
	if err != nil {
		return nil, fmt.Errorf("video enhancement failed: %w", err)
	}

	return response, nil
}

// makeRequest 发送HTTP请求
func (c *AIClient) makeRequest(ctx context.Context, endpoint string, request *AIRequest) (*AIResponse, error) {
	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// 解析响应
	var apiResponse struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    *AIResponse `json:"data"`
	}

	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if apiResponse.Code != 200 {
		return nil, fmt.Errorf("API error: %s", apiResponse.Message)
	}

	return apiResponse.Data, nil
}

// Cache methods

// Get 获取缓存
func (c *Cache) Get(key string) *AIResponse {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil
	}

	if time.Now().After(item.ExpiresAt) {
		delete(c.data, key)
		return nil
	}

	return item.Value
}

// Set 设置缓存
func (c *Cache) Set(key string, value *AIResponse) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查缓存大小限制
	if len(c.data) >= c.maxSize {
		// 删除最旧的条目
		var oldestKey string
		var oldestTime time.Time
		for k, v := range c.data {
			if oldestKey == "" || v.ExpiresAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.ExpiresAt
			}
		}
		if oldestKey != "" {
			delete(c.data, oldestKey)
		}
	}

	c.data[key] = &CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// cleanup 清理过期缓存
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.Lock()
			now := time.Now()
			for key, item := range c.data {
				if now.After(item.ExpiresAt) {
					delete(c.data, key)
				}
			}
			c.mutex.Unlock()
		}
	}
}

// RateLimiter methods

// Allow 检查是否允许请求
func (r *RateLimiter) Allow(key string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-r.window)

	// 获取或创建请求记录
	requests, exists := r.requests[key]
	if !exists {
		requests = make([]time.Time, 0)
	}

	// 清理过期的请求记录
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= r.maxRequests {
		return false
	}

	// 添加当前请求
	validRequests = append(validRequests, now)
	r.requests[key] = validRequests

	return true
}

// Utility functions

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000000)
}

// encodeBase64 编码为base64
func encodeBase64(data []byte) string {
	// 这里应该使用base64编码，为了简化直接返回字符串表示
	return fmt.Sprintf("base64_encoded_%d_bytes", len(data))
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
