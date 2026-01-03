package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// AIInferenceService AI 推理服务
type AIInferenceService struct {
	edgeLLMClient *EdgeLLMClient
	config        *config.Config
	sessions      *SessionManager
}

type ModelWarmupStatus struct {
	ModelName string `json:"model_name"`
	WorkID    string `json:"work_id,omitempty"`
	Ready     bool   `json:"ready"`
	Error     string `json:"error,omitempty"`
}

// ASRRequest ASR 请求
type ASRRequest struct {
	AudioData  string `json:"audio_data"`  // Base64 编码的音频数据
	Format     string `json:"format"`      // 音频格式（如 wav, mp3）
	SampleRate int    `json:"sample_rate"` // 采样率
	Language   string `json:"language"`    // 语言（可选）
	MeetingID  uint   `json:"meeting_id,omitempty"`
}

// ASRResponse ASR 响应
type ASRResponse struct {
	Text       string  `json:"text"`        // 识别的文本
	Confidence float64 `json:"confidence"`  // 置信度
	Language   string  `json:"language"`    // 检测到的语言
	Duration   float64 `json:"duration_ms"` // 处理时间（毫秒）
}

// EmotionRequest 情感检测请求
type EmotionRequest struct {
	Text      string `json:"text"` // 要分析的文本
	MeetingID uint   `json:"meeting_id,omitempty"`
}

// EmotionResponse 情感检测响应
type EmotionResponse struct {
	Emotion    string             `json:"emotion"`    // 主要情感
	Confidence float64            `json:"confidence"` // 置信度
	Emotions   map[string]float64 `json:"emotions"`   // 所有情感及其分数
	Duration   float64            `json:"duration_ms"`
}

// SynthesisDetectionRequest 深度伪造检测请求
type SynthesisDetectionRequest struct {
	AudioData  string `json:"audio_data"`  // Base64 编码的音频数据
	Format     string `json:"format"`      // 音频格式
	SampleRate int    `json:"sample_rate"` // 采样率
	MeetingID  uint   `json:"meeting_id,omitempty"`
}

// SynthesisDetectionResponse 深度伪造检测响应
type SynthesisDetectionResponse struct {
	IsSynthetic bool    `json:"is_synthetic"` // 是否为合成音频
	Confidence  float64 `json:"confidence"`   // 置信度
	Score       float64 `json:"score"`        // 合成分数（0-1）
	Duration    float64 `json:"duration_ms"`
}

// NewAIInferenceService 创建 AI 推理服务
func NewAIInferenceService(cfg *config.Config) *AIInferenceService {
	// 从配置中获取 unit-manager 地址
	host := "localhost"
	port := 19001
	timeout := 30 * time.Second

	if cfg != nil && cfg.ZMQ.UnitManagerHost != "" {
		host = cfg.ZMQ.UnitManagerHost
	}
	if cfg != nil && cfg.ZMQ.UnitManagerPort > 0 {
		port = cfg.ZMQ.UnitManagerPort
	}
	if cfg != nil && cfg.ZMQ.Timeout > 0 {
		timeout = time.Duration(cfg.ZMQ.Timeout) * time.Second
	}

	edgeLLMClient := NewEdgeLLMClient(host, port, timeout)
	sessionManager := NewSessionManager(edgeLLMClient, 15*time.Minute, 1*time.Minute)

	return &AIInferenceService{
		edgeLLMClient: edgeLLMClient,
		config:        cfg,
		sessions:      sessionManager,
	}
}

func (s *AIInferenceService) Close() {
	if s == nil {
		return
	}
	if s.sessions != nil {
		s.sessions.Close()
	}
}

// WarmupMeeting pre-creates (setup) inference sessions for a meeting so that subsequent requests reuse them.
// If models is empty, it will warm up ASR/Emotion/Synthesis based on current config.
func (s *AIInferenceService) WarmupMeeting(ctx context.Context, meetingID uint, models []string) (map[string]ModelWarmupStatus, error) {
	meetingKey := "global"
	if meetingID != 0 {
		meetingKey = fmt.Sprintf("%d", meetingID)
	}

	modelNames := map[string]string{
		"asr":       "asr-model",
		"emotion":   "emotion-model",
		"synthesis": "synthesis-model",
	}
	if s.config != nil {
		if s.config.AI.Models.ASR.ModelName != "" {
			modelNames["asr"] = s.config.AI.Models.ASR.ModelName
		}
		if s.config.AI.Models.Emotion.ModelName != "" {
			modelNames["emotion"] = s.config.AI.Models.Emotion.ModelName
		}
		if s.config.AI.Models.Synthesis.ModelName != "" {
			modelNames["synthesis"] = s.config.AI.Models.Synthesis.ModelName
		}
	}

	if len(models) == 0 {
		models = []string{"asr", "emotion", "synthesis"}
	}

	statuses := make(map[string]ModelWarmupStatus, len(models))
	for _, key := range models {
		modelName, ok := modelNames[key]
		if !ok {
			statuses[key] = ModelWarmupStatus{
				Ready: false,
				Error: "unknown model key",
			}
			continue
		}

		session, release, err := s.sessions.Acquire(ctx, meetingKey, modelName)
		if err != nil {
			statuses[key] = ModelWarmupStatus{
				ModelName: modelName,
				Ready:     false,
				Error:     err.Error(),
			}
			continue
		}
		workID := session.WorkID
		release()

		statuses[key] = ModelWarmupStatus{
			ModelName: modelName,
			WorkID:    workID,
			Ready:     true,
		}
	}

	return statuses, nil
}

// SpeechRecognition 语音识别
func (s *AIInferenceService) SpeechRecognition(ctx context.Context, req *ASRRequest) (*ASRResponse, error) {
	startTime := time.Now()

	logger.Info("Starting speech recognition")

	// 验证请求
	if req.AudioData == "" {
		return nil, fmt.Errorf("audio_data is required")
	}

	// 解码 Base64 音频数据
	audioBytes, err := base64.StdEncoding.DecodeString(req.AudioData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 audio data: %w", err)
	}

	logger.Debug(fmt.Sprintf("Decoded audio data: %d bytes", len(audioBytes)))

	// 准备输入数据：包含音频格式和采样率的元数据
	// 格式：audio_format=<format>,sample_rate=<rate>
	metadata := fmt.Sprintf("audio_format=%s,sample_rate=%d", req.Format, req.SampleRate)

	modelName := "asr-model"
	if s.config != nil && s.config.AI.Models.ASR.ModelName != "" {
		modelName = s.config.AI.Models.ASR.ModelName
	}

	maxSingleDeltaLen := 2048
	streamChunkSize := 1024
	if s.config != nil {
		if s.config.AI.Request.MaxSingleDeltaLen > 0 {
			maxSingleDeltaLen = s.config.AI.Request.MaxSingleDeltaLen
		}
		if s.config.AI.Request.AudioStreamChunkSize > 0 {
			streamChunkSize = s.config.AI.Request.AudioStreamChunkSize
		}
	}
	if streamChunkSize > maxSingleDeltaLen {
		streamChunkSize = maxSingleDeltaLen
	}
	// NOTE: 远端 unit-manager 的 TCP 侧对粘包非常敏感，音频类 stream 若连续快速发送，
	// 容易触发 code=-2 "json format error"。默认加一点点间隔，允许通过配置覆盖为 0。
	streamChunkDelay := 3 * time.Millisecond
	if s.config != nil && s.config.AI.Request.AudioStreamChunkDelayMs > 0 {
		streamChunkDelay = time.Duration(s.config.AI.Request.AudioStreamChunkDelayMs) * time.Millisecond
	}

	fullData := metadata + ",data=" + req.AudioData

	meetingKey := "global"
	if req.MeetingID != 0 {
		meetingKey = fmt.Sprintf("%d", req.MeetingID)
	}

	var result map[string]interface{}

	session, release, err := s.sessions.Acquire(ctx, meetingKey, modelName)
	if err != nil {
		return nil, fmt.Errorf("speech recognition acquire session failed: %w", err)
	}
	if len(fullData) > maxSingleDeltaLen {
		logger.Info(fmt.Sprintf("Using streaming mode for audio payload: meeting=%s, model=%s, wav_bytes=%d, payload_len=%d, chunk_len=%d",
			meetingKey, modelName, len(audioBytes), len(fullData), streamChunkSize))
		result, err = s.edgeLLMClient.InferenceDeltaWithAudioStream(ctx, session, fullData, streamChunkSize, streamChunkDelay)
	} else {
		logger.Info(fmt.Sprintf("Using single-shot mode for small audio file: meeting=%s, model=%s, wav_bytes=%d",
			meetingKey, modelName, len(audioBytes)))
		result, err = s.edgeLLMClient.InferenceDelta(ctx, session, fullData)
	}
	release()

	// 若复用会话出现连接断开/超时，重建并重试一次
	if err != nil {
		logger.Warn("Speech recognition failed on reused session; invalidating and retrying once",
			logger.String("meeting_id", meetingKey),
			logger.String("model", modelName),
			logger.Err(err))
		s.sessions.Invalidate(ctx, meetingKey, modelName)

		session, release, acquireErr := s.sessions.Acquire(ctx, meetingKey, modelName)
		if acquireErr != nil {
			return nil, fmt.Errorf("speech recognition acquire retry failed: %w (original: %v)", acquireErr, err)
		}
		if len(fullData) > maxSingleDeltaLen {
			result, err = s.edgeLLMClient.InferenceDeltaWithAudioStream(ctx, session, fullData, streamChunkSize, streamChunkDelay)
		} else {
			result, err = s.edgeLLMClient.InferenceDelta(ctx, session, fullData)
		}
		release()
	}
	if err != nil {
		return nil, fmt.Errorf("speech recognition failed: %w", err)
	}

	// 记录原始结果用于调试
	logger.Debug(fmt.Sprintf("ASR raw result keys: %v", getMapKeys(result)))

	// 解析结果
	// Edge-LLM-Infra 返回的字段：
	// - transcription: string (转录文本)
	// - confidence: float (置信度)
	// - model: string (模型名称)

	text := extractString(result, "transcription", "") // 注意：字段名是 transcription，不是 text
	if text == "" {
		text = extractString(result, "text", "") // 尝试备用字段名
	}

	confidence := extractFloat(result, "confidence", 0.0)
	language := extractString(result, "language", req.Language)

	// 兼容：部分实现会把纯文本直接放在 stream wrapper 的 delta 中（而不是 JSON）
	if text == "" {
		if deltaText, ok := result["delta"].(string); ok {
			deltaText = strings.TrimSpace(deltaText)
			if deltaText != "" && deltaText != "No transcription available" {
				// Best-effort: sometimes delta is actually a JSON string but wasn't parsed upstream.
				if strings.HasPrefix(deltaText, "{") && strings.HasSuffix(deltaText, "}") {
					var deltaData map[string]interface{}
					if err := json.Unmarshal([]byte(deltaText), &deltaData); err == nil {
						text = extractString(deltaData, "transcription", "")
						if text == "" {
							text = extractString(deltaData, "text", "")
						}
						if confidence == 0.0 {
							confidence = extractFloat(deltaData, "confidence", confidence)
						}
						if language == "" || language == req.Language {
							language = extractString(deltaData, "language", language)
						}
					} else {
						text = deltaText
					}
				} else {
					text = deltaText
				}
			}
		}
	}
	if text == "" {
		logger.Warn("No valid ASR text received from Edge-LLM-Infra")
	}

	// 检查 confidence 值的合理性（应该在 0-1 之间）
	if confidence < 0.0 || confidence > 1.0 {
		logger.Warn(fmt.Sprintf("Invalid ASR confidence value: %.2f, using default value", confidence))
		confidence = 0.0
	}

	response := &ASRResponse{
		Text:       text,
		Confidence: confidence,
		Language:   language,
		Duration:   float64(time.Since(startTime).Milliseconds()),
	}

	logger.Info(fmt.Sprintf("Speech recognition completed in %.2fms (text_len=%d, confidence=%.2f)",
		response.Duration, len(response.Text), response.Confidence))

	return response, nil
}

// EmotionDetection 情感检测
func (s *AIInferenceService) EmotionDetection(ctx context.Context, req *EmotionRequest) (*EmotionResponse, error) {
	startTime := time.Now()

	logger.Info("Starting emotion detection")

	// 验证请求
	if req.Text == "" {
		return nil, fmt.Errorf("text is required")
	}

	// 准备输入数据
	inputData := req.Text

	// 调用 Edge-LLM-Infra
	modelName := "emotion-model"
	if s.config != nil && s.config.AI.Models.Emotion.ModelName != "" {
		modelName = s.config.AI.Models.Emotion.ModelName
	}
	meetingKey := "global"
	if req.MeetingID != 0 {
		meetingKey = fmt.Sprintf("%d", req.MeetingID)
	}
	session, release, err := s.sessions.Acquire(ctx, meetingKey, modelName)
	if err != nil {
		return nil, fmt.Errorf("emotion detection acquire session failed: %w", err)
	}
	result, err := s.edgeLLMClient.InferenceDelta(ctx, session, inputData)
	release()
	if err != nil {
		logger.Warn("Emotion detection failed on reused session; invalidating and retrying once",
			logger.String("meeting_id", meetingKey),
			logger.String("model", modelName),
			logger.Err(err))
		s.sessions.Invalidate(ctx, meetingKey, modelName)

		session, release, acquireErr := s.sessions.Acquire(ctx, meetingKey, modelName)
		if acquireErr != nil {
			return nil, fmt.Errorf("emotion detection acquire retry failed: %w (original: %v)", acquireErr, err)
		}
		result, err = s.edgeLLMClient.InferenceDelta(ctx, session, inputData)
		release()
	}
	if err != nil {
		return nil, fmt.Errorf("emotion detection failed: %w", err)
	}

	// 记录原始结果用于调试
	logger.Debug(fmt.Sprintf("Emotion raw result keys: %v", getMapKeys(result)))

	// 解析结果
	// Edge-LLM-Infra 返回的字段：
	// - emotion: string (主要情感)
	// - confidence: float (置信度)
	// - all_emotions: map (所有情感及其分数)
	// - model: string (模型名称)

	emotions := make(map[string]float64)

	// 尝试解析 all_emotions 字段
	if emotionsData, ok := result["all_emotions"].(map[string]interface{}); ok {
		for k, v := range emotionsData {
			if score, ok := v.(float64); ok {
				emotions[k] = score
			}
		}
		logger.Debug(fmt.Sprintf("Parsed %d emotions from all_emotions field", len(emotions)))
	} else if emotionsData, ok := result["emotions"].(map[string]interface{}); ok {
		// 尝试备用字段名
		for k, v := range emotionsData {
			if score, ok := v.(float64); ok {
				emotions[k] = score
			}
		}
		logger.Debug(fmt.Sprintf("Parsed %d emotions from emotions field", len(emotions)))
	} else {
		// 默认情感分数
		logger.Warn("No valid emotions data received from Edge-LLM-Infra, using default values")
		emotions = map[string]float64{
			"happy":   0.3,
			"sad":     0.1,
			"angry":   0.05,
			"neutral": 0.55,
		}
	}

	emotion := extractString(result, "emotion", "")
	confidence := extractFloat(result, "confidence", 0.0)

	// 如果没有获取到真实数据，使用默认值
	if emotion == "" {
		logger.Warn("No valid emotion received from Edge-LLM-Infra, using default value")
		emotion = "neutral"
	}

	// 检查 confidence 值的合理性（应该在 0-1 之间）
	if confidence < 0.0 || confidence > 1.0 {
		logger.Warn(fmt.Sprintf("Invalid emotion confidence value: %.2f, using default value", confidence))
		confidence = 0.0
	}

	response := &EmotionResponse{
		Emotion:    emotion,
		Confidence: confidence,
		Emotions:   emotions,
		Duration:   float64(time.Since(startTime).Milliseconds()),
	}

	logger.Info(fmt.Sprintf("Emotion detection completed in %.2fms (emotion=%s, confidence=%.2f)",
		response.Duration, response.Emotion, response.Confidence))

	return response, nil
}

// SynthesisDetection 深度伪造检测
func (s *AIInferenceService) SynthesisDetection(ctx context.Context, req *SynthesisDetectionRequest) (*SynthesisDetectionResponse, error) {
	startTime := time.Now()

	logger.Info("Starting synthesis detection")

	// 验证请求
	if req.AudioData == "" {
		return nil, fmt.Errorf("audio_data is required")
	}

	// 解码 Base64 音频数据
	audioBytes, err := base64.StdEncoding.DecodeString(req.AudioData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 audio data: %w", err)
	}

	logger.Debug(fmt.Sprintf("Decoded audio data: %d bytes", len(audioBytes)))

	// 准备输入数据：包含音频格式和采样率的元数据
	// 格式：audio_format=<format>,sample_rate=<rate>
	metadata := fmt.Sprintf("audio_format=%s,sample_rate=%d", req.Format, req.SampleRate)

	modelName := "synthesis-model"
	if s.config != nil && s.config.AI.Models.Synthesis.ModelName != "" {
		modelName = s.config.AI.Models.Synthesis.ModelName
	}

	maxSingleDeltaLen := 2048
	streamChunkSize := 1024
	if s.config != nil {
		if s.config.AI.Request.MaxSingleDeltaLen > 0 {
			maxSingleDeltaLen = s.config.AI.Request.MaxSingleDeltaLen
		}
		if s.config.AI.Request.AudioStreamChunkSize > 0 {
			streamChunkSize = s.config.AI.Request.AudioStreamChunkSize
		}
	}
	if streamChunkSize > maxSingleDeltaLen {
		streamChunkSize = maxSingleDeltaLen
	}
	// NOTE: 同 SpeechRecognition。
	streamChunkDelay := 3 * time.Millisecond
	if s.config != nil && s.config.AI.Request.AudioStreamChunkDelayMs > 0 {
		streamChunkDelay = time.Duration(s.config.AI.Request.AudioStreamChunkDelayMs) * time.Millisecond
	}

	fullData := metadata + ",data=" + req.AudioData

	meetingKey := "global"
	if req.MeetingID != 0 {
		meetingKey = fmt.Sprintf("%d", req.MeetingID)
	}

	var result map[string]interface{}
	session, release, err := s.sessions.Acquire(ctx, meetingKey, modelName)
	if err != nil {
		return nil, fmt.Errorf("synthesis detection acquire session failed: %w", err)
	}
	if len(fullData) > maxSingleDeltaLen {
		logger.Info(fmt.Sprintf("Using streaming mode for audio payload: meeting=%s, model=%s, wav_bytes=%d, payload_len=%d, chunk_len=%d",
			meetingKey, modelName, len(audioBytes), len(fullData), streamChunkSize))
		result, err = s.edgeLLMClient.InferenceDeltaWithAudioStream(ctx, session, fullData, streamChunkSize, streamChunkDelay)
	} else {
		logger.Info(fmt.Sprintf("Using single-shot mode for small audio file: meeting=%s, model=%s, wav_bytes=%d",
			meetingKey, modelName, len(audioBytes)))
		result, err = s.edgeLLMClient.InferenceDelta(ctx, session, fullData)
	}
	release()
	if err != nil {
		logger.Warn("Synthesis detection failed on reused session; invalidating and retrying once",
			logger.String("meeting_id", meetingKey),
			logger.String("model", modelName),
			logger.Err(err))
		s.sessions.Invalidate(ctx, meetingKey, modelName)

		session, release, acquireErr := s.sessions.Acquire(ctx, meetingKey, modelName)
		if acquireErr != nil {
			return nil, fmt.Errorf("synthesis detection acquire retry failed: %w (original: %v)", acquireErr, err)
		}
		if len(fullData) > maxSingleDeltaLen {
			result, err = s.edgeLLMClient.InferenceDeltaWithAudioStream(ctx, session, fullData, streamChunkSize, streamChunkDelay)
		} else {
			result, err = s.edgeLLMClient.InferenceDelta(ctx, session, fullData)
		}
		release()
	}
	if err != nil {
		return nil, fmt.Errorf("synthesis detection failed: %w", err)
	}

	// 解析结果
	// Edge-LLM-Infra 返回的字段：
	// - is_synthetic: bool (是否为合成音频)
	// - confidence: float (置信度)
	// - probability_synthetic: float (合成概率，0-1)
	// - probability_real: float (真实概率，0-1)

	// 优先使用 is_synthetic 字段
	var isSynthetic bool
	if val, ok := result["is_synthetic"].(bool); ok {
		isSynthetic = val
	} else {
		// 如果没有 is_synthetic 字段，使用 probability_synthetic
		probSynthetic := extractFloat(result, "probability_synthetic", 0.0)
		isSynthetic = probSynthetic > 0.5
	}

	// 获取置信度和分数
	confidence := extractFloat(result, "confidence", 0.0)
	probSynthetic := extractFloat(result, "probability_synthetic", 0.0)

	// 检查值的合理性（应该在 0-1 之间）
	if confidence < 0.0 || confidence > 1.0 {
		logger.Warn(fmt.Sprintf("Invalid synthesis confidence value: %.2f, using default value", confidence))
		confidence = 0.90
	}
	if probSynthetic < 0.0 || probSynthetic > 1.0 {
		logger.Warn(fmt.Sprintf("Invalid synthesis probability value: %.2f, using default value", probSynthetic))
		probSynthetic = 0.15
	}

	// 如果没有获取到真实数据，记录警告
	if confidence == 0.0 && probSynthetic == 0.0 {
		logger.Warn("No valid synthesis detection data received from Edge-LLM-Infra, using default values")
		confidence = 0.90
		probSynthetic = 0.15
	}

	response := &SynthesisDetectionResponse{
		IsSynthetic: isSynthetic,
		Confidence:  confidence,
		Score:       probSynthetic,
		Duration:    float64(time.Since(startTime).Milliseconds()),
	}

	logger.Info(fmt.Sprintf("Synthesis detection completed in %.2fms (is_synthetic=%v, confidence=%.2f, score=%.2f)",
		response.Duration, response.IsSynthetic, response.Confidence, response.Score))

	return response, nil
}

// HealthCheck 健康检查
func (s *AIInferenceService) HealthCheck(ctx context.Context) error {
	// 尝试创建一个简单的连接来验证 unit-manager 是否可达
	modelName := "asr-model"
	if s.config != nil && s.config.AI.Models.ASR.ModelName != "" {
		modelName = s.config.AI.Models.ASR.ModelName
	}
	session, err := s.edgeLLMClient.Setup(ctx, modelName)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// 立即退出
	if err := s.edgeLLMClient.Exit(ctx, session); err != nil {
		logger.Warn(fmt.Sprintf("Failed to exit health check session: %v", err))
	}

	return nil
}

// Helper functions

func extractString(data map[string]interface{}, key string, defaultValue string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultValue
}

func extractFloat(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return defaultValue
}

func getMapKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}
