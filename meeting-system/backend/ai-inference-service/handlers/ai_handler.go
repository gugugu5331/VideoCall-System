package handlers

import (
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"meeting-system/ai-inference-service/services"
	"meeting-system/shared/database"
	"meeting-system/shared/logger"
	"meeting-system/shared/queue"
	"meeting-system/shared/response"
	"meeting-system/shared/zmq"
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
		response.Error(c, http.StatusInternalServerError, "Speech recognition failed: "+err.Error())
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

	// 执行情感检测
	ctx := c.Request.Context()
	result, err := h.aiService.EmotionDetection(ctx, &req)
	if err != nil {
		logger.Error("Emotion detection failed: " + err.Error())
		response.Error(c, http.StatusInternalServerError, "Emotion detection failed: "+err.Error())
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
		response.Error(c, http.StatusInternalServerError, "Synthesis detection failed: "+err.Error())
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
// 通过 ZeroMQ 将 Task 对象发送到 Edge-LLM-Infra (C++) 并返回结果
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

	// Build AITask
	task := &zmq.AITask{
		TaskID:    time.Now().Format("20060102150405.000000000"),
		TaskType:  req.TaskType,
		ModelPath: req.ModelPath,
		InputData: payload,
		Params:    req.Params,
	}

	ctx := c.Request.Context()
	client := zmq.GetZMQClient()
	if client == nil {
		response.Error(c, http.StatusServiceUnavailable, "ZMQ client not initialized")
		return
	}

	// Send task to Edge-LLM-Infra via ZeroMQ
	aiResp, err := client.SendTask(ctx, task)
	if err != nil {
		logger.Error("AI analyze via ZMQ failed: " + err.Error())
		response.Error(c, http.StatusInternalServerError, "AI analyze failed: "+err.Error())
		return
	}

	// Publish ai_events for downstream services (meeting-service to persist)
	if rdb := database.GetRedis(); rdb != nil {
		pubsub := queue.NewRedisPubSubQueue(rdb)
		_ = pubsub.Publish(ctx, "ai_events", &queue.PubSubMessage{
			Type: "" + req.TaskType + ".completed",
			Payload: map[string]interface{}{
				"task_id":    task.TaskID,
				"meeting_id": req.MeetingID,
				"user_id":    req.UserID,
				"result":     aiResp.Data,
			},
		})
	}

	response.Success(c, gin.H{
		"task_id": task.TaskID,
		"result":  aiResp.Data,
	})
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
		response.Error(c, http.StatusServiceUnavailable, "Service unhealthy: "+err.Error())
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
		"description": "AI Inference Service integrated with Edge-LLM-Infra",
		"capabilities": []string{
			"speech_recognition",
			"emotion_detection",
			"synthesis_detection",
		},
		"models": gin.H{
			"asr":       "asr-model",
			"emotion":   "emotion-model",
			"synthesis": "synthesis-model",
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
					Text: getString(emotionReq, "text"),
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
