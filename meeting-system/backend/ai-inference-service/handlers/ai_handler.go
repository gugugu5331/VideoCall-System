package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"meeting-system/ai-inference-service/runtime"
	"meeting-system/ai-inference-service/services"
	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/queue"
	"meeting-system/shared/response"
)

// AIHandler AI 推理处理器
type AIHandler struct {
	aiService *services.AIInferenceService
}

// NewAIHandler 创建 AI 推理处理器
func NewAIHandler(aiService *services.AIInferenceService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// SpeechRecognition 语音识别接口
// @Summary 语音识别
// @Description 将音频转换为文本
// @Tags AI
// @Accept json
// @Produce json
// @Param request body services.ASRRequest true "ASR 请求"
// @Success 200 {object} response.Response{data=services.ASRResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/ai/asr [post]
func (h *AIHandler) SpeechRecognition(c *gin.Context) {
	var req services.ASRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid ASR request: " + err.Error())
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// 设置默认值
	if req.Format == "" {
		req.Format = "wav"
	}
	if req.SampleRate == 0 {
		req.SampleRate = 16000
	}

	// 执行语音识别
	ctx := c.Request.Context()
	result, err := h.aiService.SpeechRecognition(ctx, &req)
	if err != nil {
		logger.Error("Speech recognition failed: " + err.Error())
		respondAIError(c, err, "Speech recognition failed")
		return
	}

	response.Success(c, result)
}

// EmotionDetection 情感检测接口
// @Summary 情感检测
// @Description 分析文本的情感倾向
// @Tags AI
// @Accept json
// @Produce json
// @Param request body services.EmotionRequest true "情感检测请求"
// @Success 200 {object} response.Response{data=services.EmotionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/ai/emotion [post]
func (h *AIHandler) EmotionDetection(c *gin.Context) {
	var req services.EmotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid emotion detection request: " + err.Error())
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if req.AudioData != "" {
		if req.Format == "" {
			req.Format = "wav"
		}
		if req.SampleRate == 0 {
			req.SampleRate = 16000
		}
		if req.Channels == 0 {
			req.Channels = 1
		}
	}

	// 执行情感检测
	ctx := c.Request.Context()
	result, err := h.aiService.EmotionDetection(ctx, &req)
	if err != nil {
		logger.Error("Emotion detection failed: " + err.Error())
		respondAIError(c, err, "Emotion detection failed")
		return
	}

	response.Success(c, result)
}

// SynthesisDetection 深度伪造检测接口
// @Summary 深度伪造检测
// @Description 检测音频是否为 AI 合成
// @Tags AI
// @Accept json
// @Produce json
// @Param request body services.SynthesisDetectionRequest true "深度伪造检测请求"
// @Success 200 {object} response.Response{data=services.SynthesisDetectionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/ai/synthesis [post]
func (h *AIHandler) SynthesisDetection(c *gin.Context) {
	var req services.SynthesisDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid synthesis detection request: " + err.Error())
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// 设置默认值
	if req.Format == "" {
		req.Format = "wav"
	}
	if req.SampleRate == 0 {
		req.SampleRate = 16000
	}

	// 执行深度伪造检测
	ctx := c.Request.Context()
	result, err := h.aiService.SynthesisDetection(ctx, &req)
	if err != nil {
		logger.Error("Synthesis detection failed: " + err.Error())
		respondAIError(c, err, "Synthesis detection failed")
		return
	}
	response.Success(c, result)
}

// SetupMeeting 预热/初始化会议的 AI 会话（一次 setup，后续复用）
// @Summary 预热会议 AI 会话
// @Description 为指定 meeting_id 预先 setup ASR/Emotion/Synthesis 会话，避免首次推理卡顿
// @Tags AI
// @Accept json
// @Produce json
// @Param request body object false "可选: {meeting_id, models}"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/ai/setup [post]
func (h *AIHandler) SetupMeeting(c *gin.Context) {
	var req struct {
		MeetingID uint     `json:"meeting_id,omitempty"`
		Models    []string `json:"models,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		logger.Warn("Invalid setup request: " + err.Error())
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	statuses, err := h.aiService.WarmupMeeting(ctx, req.MeetingID, req.Models)
	if err != nil {
		logger.Error("AI meeting warmup failed: " + err.Error())
		response.Error(c, http.StatusInternalServerError, "AI meeting warmup failed: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"meeting_id": req.MeetingID,
		"models":     statuses,
		"timestamp":  time.Now().Unix(),
	})
}

// Analyze 通用 AI 推理接口（客户端直连 AI 服务）
// 当前仅支持 asr/emotion/synthesis 基础任务。
func (h *AIHandler) Analyze(c *gin.Context) {
	var req struct {
		TaskType  string            `json:"task_type"`
		ModelPath string            `json:"model_path"`
		InputData string            `json:"input_data"` // base64
		Params    map[string]string `json:"params"`
		MeetingID string            `json:"meeting_id"`
		UserID    string            `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid analyze request: " + err.Error())
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// decode base64 input
	var payload []byte
	if req.InputData != "" {
		b, err := base64.StdEncoding.DecodeString(req.InputData)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid input_data (base64): "+err.Error())
			return
		}
		payload = b
	}

	taskID := time.Now().Format("20060102150405.000000000")
	ctx := c.Request.Context()

	taskType := strings.ToLower(strings.TrimSpace(req.TaskType))
	switch taskType {
	case "asr", "speech_recognition":
		sampleRate := parseIntParam(req.Params, "sample_rate", 16000)
		channels := parseIntParam(req.Params, "channels", 1)
		format := ""
		language := ""
		if req.Params != nil {
			format = strings.TrimSpace(req.Params["format"])
			language = req.Params["language"]
		}
		if format == "" {
			format = "pcm"
		}
		resp, err := h.aiService.SpeechRecognition(ctx, &services.ASRRequest{
			AudioData:  req.InputData,
			Format:     format,
			SampleRate: sampleRate,
			Channels:   channels,
			Language:   language,
		})
		if err != nil {
			respondAIError(c, err, "AI analyze failed")
			return
		}
		publishAIEvent(ctx, taskType, taskID, req.MeetingID, req.UserID, resp)
		response.Success(c, gin.H{"task_id": taskID, "result": resp})
	case "emotion", "emotion_detection":
		text := ""
		format := ""
		sampleRate := parseIntParam(req.Params, "sample_rate", 16000)
		channels := parseIntParam(req.Params, "channels", 1)
		if req.Params != nil {
			text = strings.TrimSpace(req.Params["text"])
			format = strings.TrimSpace(req.Params["format"])
		}
		if format != "" && req.InputData != "" {
			if sampleRate == 0 {
				sampleRate = 16000
			}
			if channels == 0 {
				channels = 1
			}
			resp, err := h.aiService.EmotionDetection(ctx, &services.EmotionRequest{
				AudioData:  req.InputData,
				Format:     format,
				SampleRate: sampleRate,
				Channels:   channels,
			})
			if err != nil {
				respondAIError(c, err, "AI analyze failed")
				return
			}
			publishAIEvent(ctx, taskType, taskID, req.MeetingID, req.UserID, resp)
			response.Success(c, gin.H{"task_id": taskID, "result": resp})
			return
		}
		if text == "" && len(payload) > 0 {
			text = strings.TrimSpace(string(payload))
		}
		if text == "" {
			response.Error(c, http.StatusBadRequest, "text is required for emotion detection")
			return
		}
		resp, err := h.aiService.EmotionDetection(ctx, &services.EmotionRequest{Text: text})
		if err != nil {
			respondAIError(c, err, "AI analyze failed")
			return
		}
		publishAIEvent(ctx, taskType, taskID, req.MeetingID, req.UserID, resp)
		response.Success(c, gin.H{"task_id": taskID, "result": resp})
	case "synthesis", "synthesis_detection":
		sampleRate := parseIntParam(req.Params, "sample_rate", 16000)
		channels := parseIntParam(req.Params, "channels", 1)
		format := ""
		if req.Params != nil {
			format = strings.TrimSpace(req.Params["format"])
		}
		if format == "" {
			format = "pcm"
		}
		resp, err := h.aiService.SynthesisDetection(ctx, &services.SynthesisDetectionRequest{
			AudioData:  req.InputData,
			Format:     format,
			SampleRate: sampleRate,
			Channels:   channels,
		})
		if err != nil {
			respondAIError(c, err, "AI analyze failed")
			return
		}
		publishAIEvent(ctx, taskType, taskID, req.MeetingID, req.UserID, resp)
		response.Success(c, gin.H{"task_id": taskID, "result": resp})
	default:
		response.Error(c, http.StatusBadRequest, "unsupported task_type: "+req.TaskType)
	}
}

// HealthCheck 健康检查接口
// @Summary 健康检查
// @Description 检查 AI 推理服务是否正常
// @Tags System
// @Produce json
// @Success 200 {object} response.Response
// @Failure 503 {object} response.Response
// @Router /api/v1/ai/health [get]
func (h *AIHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.aiService.HealthCheck(ctx); err != nil {
		logger.Error("Health check failed: " + err.Error())
		respondAIError(c, err, "Service unhealthy")
		return
	}

	response.Success(c, gin.H{
		"status":    "healthy",
		"service":   "ai-inference-service",
		"timestamp": time.Now().Unix(),
	})
}

// GetServiceInfo 获取服务信息
// @Summary 获取服务信息
// @Description 获取 AI 推理服务的详细信息
// @Tags System
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/ai/info [get]
func (h *AIHandler) GetServiceInfo(c *gin.Context) {
	info := gin.H{
		"service":     "ai-inference-service",
		"version":     "1.0.0",
		"description": "AI Inference Service powered by Triton + TensorRT",
		"capabilities": []string{
			"speech_recognition",
			"emotion_detection",
			"synthesis_detection",
		},
		"models": gin.H{
			"asr":       "whisper",
			"emotion":   "emotion",
			"synthesis": "synthesis",
		},
		"timestamp": time.Now().Unix(),
	}

	response.Success(c, info)
}

// BatchInference 批量推理接口
// @Summary 批量推理
// @Description 批量执行多个 AI 推理任务
// @Tags AI
// @Accept json
// @Produce json
// @Param request body BatchInferenceRequest true "批量推理请求"
// @Success 200 {object} response.Response{data=BatchInferenceResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/ai/batch [post]
func (h *AIHandler) BatchInference(c *gin.Context) {
	var req BatchInferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid batch inference request: " + err.Error())
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	results := make([]BatchInferenceResult, 0, len(req.Tasks))

	// 执行每个任务
	for _, task := range req.Tasks {
		result := BatchInferenceResult{
			TaskID: task.TaskID,
			Type:   task.Type,
		}

		switch task.Type {
		case "asr":
			if asrReq, ok := task.Data.(map[string]interface{}); ok {
				req := &services.ASRRequest{
					AudioData:  getString(asrReq, "audio_data"),
					Format:     getString(asrReq, "format"),
					SampleRate: getInt(asrReq, "sample_rate"),
					Channels:   getInt(asrReq, "channels"),
					Language:   getString(asrReq, "language"),
				}
				resp, err := h.aiService.SpeechRecognition(ctx, req)
				if err != nil {
					result.Error = err.Error()
				} else {
					result.Result = resp
				}
			}

		case "emotion":
			if emotionReq, ok := task.Data.(map[string]interface{}); ok {
				req := &services.EmotionRequest{
					Text:       getString(emotionReq, "text"),
					AudioData:  getString(emotionReq, "audio_data"),
					Format:     getString(emotionReq, "format"),
					SampleRate: getInt(emotionReq, "sample_rate"),
					Channels:   getInt(emotionReq, "channels"),
				}
				if req.AudioData != "" {
					if req.Format == "" {
						req.Format = "wav"
					}
					if req.SampleRate == 0 {
						req.SampleRate = 16000
					}
					if req.Channels == 0 {
						req.Channels = 1
					}
				}
				resp, err := h.aiService.EmotionDetection(ctx, req)
				if err != nil {
					result.Error = err.Error()
				} else {
					result.Result = resp
				}
			}

		case "synthesis":
			if synthesisReq, ok := task.Data.(map[string]interface{}); ok {
				req := &services.SynthesisDetectionRequest{
					AudioData:  getString(synthesisReq, "audio_data"),
					Format:     getString(synthesisReq, "format"),
					SampleRate: getInt(synthesisReq, "sample_rate"),
					Channels:   getInt(synthesisReq, "channels"),
				}
				resp, err := h.aiService.SynthesisDetection(ctx, req)
				if err != nil {
					result.Error = err.Error()
				} else {
					result.Result = resp
				}
			}

		default:
			result.Error = "unknown task type: " + task.Type
		}

		results = append(results, result)
	}

	response.Success(c, BatchInferenceResponse{
		Results: results,
		Total:   len(results),
	})
}

// BatchInferenceRequest 批量推理请求
type BatchInferenceRequest struct {
	Tasks []BatchTask `json:"tasks"`
}

// BatchTask 批量任务
type BatchTask struct {
	TaskID string      `json:"task_id"`
	Type   string      `json:"type"` // asr, emotion, synthesis
	Data   interface{} `json:"data"`
}

// BatchInferenceResponse 批量推理响应
type BatchInferenceResponse struct {
	Results []BatchInferenceResult `json:"results"`
	Total   int                    `json:"total"`
}

// BatchInferenceResult 批量推理结果
type BatchInferenceResult struct {
	TaskID string      `json:"task_id"`
	Type   string      `json:"type"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func respondAIError(c *gin.Context, err error, fallback string) {
	if err == nil {
		response.Error(c, http.StatusInternalServerError, fallback)
		return
	}
	if errors.Is(err, runtime.ErrInferenceNotImplemented) {
		response.Error(c, http.StatusNotImplemented, fallback+": "+err.Error())
		return
	}
	if errors.Is(err, runtime.ErrModelNotLoaded) {
		response.Error(c, http.StatusServiceUnavailable, fallback+": "+err.Error())
		return
	}
	response.Error(c, http.StatusInternalServerError, fallback+": "+err.Error())
}

func parseIntParam(params map[string]string, key string, defaultValue int) int {
	if params == nil {
		return defaultValue
	}
	val := strings.TrimSpace(params[key])
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func publishAIEvent(ctx context.Context, taskType, taskID, meetingID, userID string, result interface{}) {
	if rdb := database.GetRedis(); rdb != nil {
		pubsub := queue.NewRedisPubSubQueue(rdb)
		_ = pubsub.Publish(ctx, "ai_events", &queue.PubSubMessage{
			Type: taskType + ".completed",
			Payload: map[string]interface{}{
				"task_id":    taskID,
				"meeting_id": meetingID,
				"user_id":    userID,
				"result":     result,
			},
		})
	}
}
