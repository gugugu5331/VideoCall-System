package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"meeting-system/ai-inference-service/runtime"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// AIInferenceService AI 推理服务
// NOTE: Edge-LLM-Infra 已移除，当前通过 Triton/TensorRT 提供推理能力。
type AIInferenceService struct {
	config *config.Config
	models *ModelManager
}

type ModelWarmupStatus struct {
	ModelName string `json:"model_name"`
	ModelPath string `json:"model_path,omitempty"`
	Ready     bool   `json:"ready"`
	Error     string `json:"error,omitempty"`
}

// ASRRequest ASR 请求
type ASRRequest struct {
	AudioData  string `json:"audio_data"`  // Base64 编码的音频数据
	Format     string `json:"format"`      // 音频格式（wav/pcm）
	SampleRate int    `json:"sample_rate"` // 采样率
	Channels   int    `json:"channels"`    // 声道数（可选）
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
	Text       string `json:"text"`        // 要分析的文本
	AudioData  string `json:"audio_data"`  // Base64 编码的音频数据（可选）
	Format     string `json:"format"`      // 音频格式（wav/pcm）
	SampleRate int    `json:"sample_rate"` // 采样率
	Channels   int    `json:"channels"`    // 声道数（可选）
	MeetingID  uint   `json:"meeting_id,omitempty"`
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
	Channels   int    `json:"channels"`    // 声道数（可选）
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
	ctx := context.Background()
	models, err := NewModelManager(ctx, cfg)
	if err != nil {
		logger.Warn("AI model manager initialization failed: " + err.Error())
	}

	return &AIInferenceService{
		config: cfg,
		models: models,
	}
}

func (s *AIInferenceService) Close() {
	if s == nil || s.models == nil {
		return
	}
	s.models.Close()
}

// WarmupMeeting 预热模型（当前仅校验模型是否加载）
func (s *AIInferenceService) WarmupMeeting(ctx context.Context, meetingID uint, models []string) (map[string]ModelWarmupStatus, error) {
	if s == nil || s.models == nil {
		return nil, fmt.Errorf("model manager not initialized")
	}

	modelMap := map[string]runtime.TaskType{
		"asr":       runtime.TaskASR,
		"emotion":   runtime.TaskEmotion,
		"synthesis": runtime.TaskSynthesis,
	}

	if len(models) == 0 {
		models = []string{"asr", "emotion", "synthesis"}
	}

	statuses := make(map[string]ModelWarmupStatus, len(models))
	for _, key := range models {
		task, ok := modelMap[key]
		if !ok {
			statuses[key] = ModelWarmupStatus{Ready: false, Error: "unknown model key"}
			continue
		}

		model, spec, loaded := s.models.GetModel(task)
		if !loaded || model == nil {
			statuses[key] = ModelWarmupStatus{
				ModelName: spec.Name,
				ModelPath: spec.Path,
				Ready:     false,
				Error:     "model not loaded",
			}
			continue
		}

		statuses[key] = ModelWarmupStatus{
			ModelName: spec.Name,
			ModelPath: spec.Path,
			Ready:     true,
		}
	}

	return statuses, nil
}

// SpeechRecognition 语音识别（HTTP）
func (s *AIInferenceService) SpeechRecognition(ctx context.Context, req *ASRRequest) (*ASRResponse, error) {
	if req.AudioData == "" {
		return nil, fmt.Errorf("audio_data is required")
	}

	format := req.Format
	if format == "" {
		format = "wav"
	}

	if req.SampleRate == 0 {
		req.SampleRate = 16000
	}
	if req.Channels == 0 {
		req.Channels = 1
	}

	audioBytes, err := base64.StdEncoding.DecodeString(req.AudioData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 audio data: %w", err)
	}

	pcmData, sampleRate, channels, err := normalizeAudioPayload(audioBytes, format, req.SampleRate, req.Channels)
	if err != nil {
		return nil, err
	}

	return s.SpeechRecognitionPCM(ctx, pcmData, sampleRate, channels, format, req.Language)
}

// SpeechRecognitionPCM 语音识别（PCM 输入，用于 gRPC 流）
func (s *AIInferenceService) SpeechRecognitionPCM(ctx context.Context, pcmData []byte, sampleRate, channels int, format, language string) (*ASRResponse, error) {
	startTime := time.Now()

	if s == nil || s.models == nil {
		return nil, fmt.Errorf("model manager not initialized")
	}

	_, spec, loaded := s.models.GetModel(runtime.TaskASR)
	if !loaded {
		return nil, runtime.ErrModelNotLoaded
	}

	targetRate := sampleRate
	if spec.SampleRate > 0 {
		targetRate = spec.SampleRate
	}
	targetChannels := channels
	if spec.Channels > 0 {
		targetChannels = spec.Channels
	}

	audioFloat, err := prepareAudioPCM(pcmData, sampleRate, channels, targetRate, targetChannels)
	if err != nil {
		return nil, err
	}

	params := map[string]string{}
	if language != "" {
		params["language"] = language
	}
	result, _, err := s.models.Infer(ctx, runtime.TaskASR, runtime.InferenceRequest{
		Task:         runtime.TaskASR,
		AudioPCM:     pcmData,
		AudioFloat32: audioFloat,
		SampleRate:   targetRate,
		Channels:     targetChannels,
		Params:       params,
	})
	if err != nil {
		return nil, err
	}

	text := ""
	confidence := 0.0
	lang := language
	if result != nil && result.Outputs != nil {
		text = extractString(result.Outputs, "text", "")
		if text == "" {
			text = extractString(result.Outputs, "transcription", "")
		}
		confidence = extractFloat(result.Outputs, "confidence", confidence)
		if lang == "" {
			lang = extractString(result.Outputs, "language", lang)
		}
	}

	response := &ASRResponse{
		Text:       text,
		Confidence: confidence,
		Language:   lang,
		Duration:   float64(time.Since(startTime).Milliseconds()),
	}

	return response, nil
}

// EmotionDetection 情感检测
func (s *AIInferenceService) EmotionDetection(ctx context.Context, req *EmotionRequest) (*EmotionResponse, error) {
	startTime := time.Now()

	if req.AudioData != "" {
		format := req.Format
		if format == "" {
			format = "wav"
		}
		if req.SampleRate == 0 {
			req.SampleRate = 16000
		}
		if req.Channels == 0 {
			req.Channels = 1
		}
		audioBytes, err := base64.StdEncoding.DecodeString(req.AudioData)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 audio data: %w", err)
		}
		pcmData, sampleRate, channels, err := normalizeAudioPayload(audioBytes, format, req.SampleRate, req.Channels)
		if err != nil {
			return nil, err
		}
		return s.EmotionDetectionPCM(ctx, pcmData, sampleRate, channels, format)
	}

	if req.Text == "" {
		return nil, fmt.Errorf("text is required")
	}

	if s == nil || s.models == nil {
		return nil, fmt.Errorf("model manager not initialized")
	}

	result, _, err := s.models.Infer(ctx, runtime.TaskEmotion, runtime.InferenceRequest{
		Task: runtime.TaskEmotion,
		Text: req.Text,
	})
	if err != nil {
		return nil, err
	}

	emotions := make(map[string]float64)
	emotion := ""
	confidence := 0.0
	if result != nil && result.Outputs != nil {
		if emotionsData, ok := result.Outputs["emotions"].(map[string]interface{}); ok {
			for k, v := range emotionsData {
				if score, ok := v.(float64); ok {
					emotions[k] = score
				}
			}
		}
		if emotionsData, ok := result.Outputs["all_emotions"].(map[string]interface{}); ok {
			for k, v := range emotionsData {
				if score, ok := v.(float64); ok {
					emotions[k] = score
				}
			}
		}
		emotion = extractString(result.Outputs, "emotion", emotion)
		confidence = extractFloat(result.Outputs, "confidence", confidence)
	}

	response := &EmotionResponse{
		Emotion:    emotion,
		Confidence: confidence,
		Emotions:   emotions,
		Duration:   float64(time.Since(startTime).Milliseconds()),
	}

	return response, nil
}

// EmotionDetectionPCM 情感检测（PCM 输入）
func (s *AIInferenceService) EmotionDetectionPCM(ctx context.Context, pcmData []byte, sampleRate, channels int, format string) (*EmotionResponse, error) {
	startTime := time.Now()

	if s == nil || s.models == nil {
		return nil, fmt.Errorf("model manager not initialized")
	}

	_, spec, loaded := s.models.GetModel(runtime.TaskEmotion)
	if !loaded {
		return nil, runtime.ErrModelNotLoaded
	}

	targetRate := sampleRate
	if spec.SampleRate > 0 {
		targetRate = spec.SampleRate
	}
	targetChannels := channels
	if spec.Channels > 0 {
		targetChannels = spec.Channels
	}

	audioFloat, err := prepareAudioPCM(pcmData, sampleRate, channels, targetRate, targetChannels)
	if err != nil {
		return nil, err
	}

	result, _, err := s.models.Infer(ctx, runtime.TaskEmotion, runtime.InferenceRequest{
		Task:         runtime.TaskEmotion,
		AudioPCM:     pcmData,
		AudioFloat32: audioFloat,
		SampleRate:   targetRate,
		Channels:     targetChannels,
	})
	if err != nil {
		return nil, err
	}

	emotions := make(map[string]float64)
	emotion := ""
	confidence := 0.0
	if result != nil && result.Outputs != nil {
		if emotionsData, ok := result.Outputs["emotions"].(map[string]interface{}); ok {
			for k, v := range emotionsData {
				if score, ok := v.(float64); ok {
					emotions[k] = score
				}
			}
		}
		if emotionsData, ok := result.Outputs["all_emotions"].(map[string]interface{}); ok {
			for k, v := range emotionsData {
				if score, ok := v.(float64); ok {
					emotions[k] = score
				}
			}
		}
		emotion = extractString(result.Outputs, "emotion", emotion)
		confidence = extractFloat(result.Outputs, "confidence", confidence)
	}

	response := &EmotionResponse{
		Emotion:    emotion,
		Confidence: confidence,
		Emotions:   emotions,
		Duration:   float64(time.Since(startTime).Milliseconds()),
	}

	return response, nil
}

// SynthesisDetection 深度伪造检测
func (s *AIInferenceService) SynthesisDetection(ctx context.Context, req *SynthesisDetectionRequest) (*SynthesisDetectionResponse, error) {
	if req.AudioData == "" {
		return nil, fmt.Errorf("audio_data is required")
	}

	format := req.Format
	if format == "" {
		format = "wav"
	}

	if req.SampleRate == 0 {
		req.SampleRate = 16000
	}
	if req.Channels == 0 {
		req.Channels = 1
	}

	audioBytes, err := base64.StdEncoding.DecodeString(req.AudioData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 audio data: %w", err)
	}

	pcmData, sampleRate, channels, err := normalizeAudioPayload(audioBytes, format, req.SampleRate, req.Channels)
	if err != nil {
		return nil, err
	}

	return s.SynthesisDetectionPCM(ctx, pcmData, sampleRate, channels, format)
}

// SynthesisDetectionPCM 深度伪造检测（PCM 输入）
func (s *AIInferenceService) SynthesisDetectionPCM(ctx context.Context, pcmData []byte, sampleRate, channels int, format string) (*SynthesisDetectionResponse, error) {
	startTime := time.Now()

	if s == nil || s.models == nil {
		return nil, fmt.Errorf("model manager not initialized")
	}

	_, spec, loaded := s.models.GetModel(runtime.TaskSynthesis)
	if !loaded {
		return nil, runtime.ErrModelNotLoaded
	}

	targetRate := sampleRate
	if spec.SampleRate > 0 {
		targetRate = spec.SampleRate
	}
	targetChannels := channels
	if spec.Channels > 0 {
		targetChannels = spec.Channels
	}

	audioFloat, err := prepareAudioPCM(pcmData, sampleRate, channels, targetRate, targetChannels)
	if err != nil {
		return nil, err
	}

	result, _, err := s.models.Infer(ctx, runtime.TaskSynthesis, runtime.InferenceRequest{
		Task:         runtime.TaskSynthesis,
		AudioPCM:     pcmData,
		AudioFloat32: audioFloat,
		SampleRate:   targetRate,
		Channels:     targetChannels,
	})
	if err != nil {
		return nil, err
	}

	isSynthetic := false
	confidence := 0.0
	score := 0.0
	if result != nil && result.Outputs != nil {
		if val, ok := result.Outputs["is_synthetic"].(bool); ok {
			isSynthetic = val
		}
		score = extractFloat(result.Outputs, "probability_synthetic", score)
		confidence = extractFloat(result.Outputs, "confidence", confidence)
		if !isSynthetic && score > 0 {
			isSynthetic = score > 0.5
		}
	}

	response := &SynthesisDetectionResponse{
		IsSynthetic: isSynthetic,
		Confidence:  confidence,
		Score:       score,
		Duration:    float64(time.Since(startTime).Milliseconds()),
	}

	return response, nil
}

// HealthCheck 健康检查
func (s *AIInferenceService) HealthCheck(ctx context.Context) error {
	if s == nil || s.models == nil {
		return fmt.Errorf("model manager not initialized")
	}

	_, _, loaded := s.models.GetModel(runtime.TaskASR)
	if !loaded {
		return runtime.ErrModelNotLoaded
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
	if val, ok := data[key].(int); ok {
		return float64(val)
	}
	return defaultValue
}
