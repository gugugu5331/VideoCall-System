package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// MediaProcessor 媒体数据处理器
type MediaProcessor struct {
	config          *config.Config
	aiClient        *AIClient
	aiGRPCClient    *AIGRPCClient // 新增：gRPC 客户端
	ffmpegService   *FFmpegService
	activeStreams   map[string]*StreamProcessor
	streamsMux      sync.RWMutex
	processingQueue chan *ProcessingTask
	workers         int
	useGRPC         bool // 是否使用 gRPC
	aiGRPCAddr      string
}

// StreamProcessor 流处理器
type StreamProcessor struct {
	StreamID       string
	UserID         string
	RoomID         string
	AudioTrack     *webrtc.TrackRemote
	VideoTrack     *webrtc.TrackRemote
	AudioBuffer    *CircularBuffer
	VideoBuffer    *CircularBuffer
	LastProcessed  time.Time
	IsActive       bool
	AITasks        []string           // 需要执行的AI任务
	AIStreamClient *AudioStreamClient // gRPC 流客户端
	Sequence       int32              // 音频片段序列号
}

// ProcessingTask 处理任务
type ProcessingTask struct {
	StreamID  string
	UserID    string
	RoomID    string
	AudioData *AudioData
	VideoData *VideoData
	Tasks     []string
	Callback  func(results map[string]*AIResponse, err error)
	CreatedAt time.Time
}

// CircularBuffer 循环缓冲区
type CircularBuffer struct {
	buffer   []byte
	size     int
	head     int
	tail     int
	count    int
	mutex    sync.RWMutex
	maxSize  int
	duration time.Duration // 缓冲区时长
}

// ProcessingResult 处理结果
type ProcessingResult struct {
	StreamID  string                 `json:"stream_id"`
	UserID    string                 `json:"user_id"`
	RoomID    string                 `json:"room_id"`
	Results   map[string]*AIResponse `json:"results"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// EndpointStatus 外部端点状态
type EndpointStatus struct {
	Available bool   `json:"available"`
	Address   string `json:"address,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ConnectivityStatus AI服务连通性状态
type ConnectivityStatus struct {
	Timestamp      time.Time      `json:"timestamp"`
	OverallHealthy bool           `json:"overall_healthy"`
	GRPC           EndpointStatus `json:"grpc"`
	HTTP           EndpointStatus `json:"http"`
}

// NewMediaProcessor 创建媒体处理器
func NewMediaProcessor(config *config.Config, aiClient *AIClient, ffmpegService *FFmpegService) *MediaProcessor {
	processor := &MediaProcessor{
		config:          config,
		aiClient:        aiClient,
		ffmpegService:   ffmpegService,
		activeStreams:   make(map[string]*StreamProcessor),
		processingQueue: make(chan *ProcessingTask, 1000),
		workers:         4,     // 默认4个工作协程
		useGRPC:         false, // 默认使用 HTTP，可通过环境变量切换
	}

	// 尝试初始化 gRPC 客户端
	aiHost := config.Services.AIService.Host
	if aiHost == "" {
		aiHost = "ai-service"
	}
	aiGRPCPort := config.Services.AIService.GrpcPort
	if aiGRPCPort == 0 {
		aiHTTPPort := config.Services.AIService.Port
		if aiHTTPPort == 0 {
			aiHTTPPort = 8084
		}
		aiGRPCPort = aiHTTPPort + 1000
	}
	aiServiceAddr := fmt.Sprintf("%s:%d", aiHost, aiGRPCPort)
	processor.aiGRPCAddr = aiServiceAddr
	grpcClient, err := NewAIGRPCClient(aiServiceAddr)
	if err != nil {
		logger.Warn("Failed to initialize AI gRPC client, falling back to HTTP",
			logger.Err(err),
			logger.String("address", aiServiceAddr))
	} else {
		processor.aiGRPCClient = grpcClient
		processor.aiGRPCAddr = aiServiceAddr
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := grpcClient.CheckConnectivity(ctx); err != nil {
			logger.Warn("AI gRPC health check failed, continuing with HTTP fallback",
				logger.Err(err),
				logger.String("address", aiServiceAddr))
		} else {
			processor.useGRPC = true
			logger.Info("AI gRPC client initialized successfully",
				logger.String("address", aiServiceAddr))
		}
	}

	// 启动工作协程
	for i := 0; i < processor.workers; i++ {
		go processor.worker()
	}

	return processor
}

// CheckAIConnectivity 检查与 AI 服务的连通性
func (p *MediaProcessor) CheckAIConnectivity(ctx context.Context) ConnectivityStatus {
	status := ConnectivityStatus{
		Timestamp: time.Now(),
		GRPC: EndpointStatus{
			Address: p.aiGRPCAddr,
		},
	}

	if p.aiClient != nil {
		status.HTTP.Address = p.aiClient.BaseURL()
		if err := p.aiClient.HealthCheck(ctx); err != nil {
			status.HTTP.Error = err.Error()
		} else {
			status.HTTP.Available = true
		}
	} else {
		status.HTTP.Error = "http client not initialized"
	}

	if p.aiGRPCClient != nil {
		if err := p.aiGRPCClient.CheckConnectivity(ctx); err != nil {
			status.GRPC.Error = err.Error()
		} else {
			status.GRPC.Available = true
		}
	} else {
		status.GRPC.Error = "grpc client not initialized"
	}

	status.OverallHealthy = status.GRPC.Available || status.HTTP.Available
	return status
}

// RegisterStream 注册音视频流
func (p *MediaProcessor) RegisterStream(streamID, userID, roomID string, audioTrack, videoTrack *webrtc.TrackRemote, aiTasks []string) error {
	p.streamsMux.Lock()
	defer p.streamsMux.Unlock()

	logger.Info(fmt.Sprintf("Registering stream: %s for user %s in room %s", streamID, userID, roomID))

	// 创建缓冲区
	audioBuffer := NewCircularBuffer(1024*1024, 5*time.Second)   // 5秒音频缓冲
	videoBuffer := NewCircularBuffer(5*1024*1024, 2*time.Second) // 2秒视频缓冲

	streamProcessor := &StreamProcessor{
		StreamID:      streamID,
		UserID:        userID,
		RoomID:        roomID,
		AudioTrack:    audioTrack,
		VideoTrack:    videoTrack,
		AudioBuffer:   audioBuffer,
		VideoBuffer:   videoBuffer,
		LastProcessed: time.Now(),
		IsActive:      true,
		AITasks:       aiTasks,
		Sequence:      0,
	}

	// 如果使用 gRPC，启动流式连接
	if p.useGRPC && p.aiGRPCClient != nil && len(aiTasks) > 0 {
		streamClient, err := p.aiGRPCClient.StartAudioStream(context.Background(), streamID, aiTasks)
		if err != nil {
			logger.Error("Failed to start AI audio stream",
				logger.String("stream_id", streamID),
				logger.Err(err))
		} else {
			streamProcessor.AIStreamClient = streamClient
			// 启动结果接收 goroutine
			go p.receiveAIResults(streamProcessor)
			logger.Info("Started AI audio stream",
				logger.String("stream_id", streamID))
		}
	}

	p.activeStreams[streamID] = streamProcessor

	// 启动音频数据收集
	if audioTrack != nil {
		go p.collectAudioData(streamProcessor)
	}

	// 启动视频数据收集
	if videoTrack != nil {
		go p.collectVideoData(streamProcessor)
	}

	// 启动定期处理（如果不使用 gRPC 流式）
	if !p.useGRPC || streamProcessor.AIStreamClient == nil {
		go p.scheduleProcessing(streamProcessor)
	} else {
		// 使用 gRPC 流式，启动实时处理
		go p.streamProcessing(streamProcessor)
	}

	return nil
}

// UnregisterStream 注销音视频流
func (p *MediaProcessor) UnregisterStream(streamID string) error {
	p.streamsMux.Lock()
	defer p.streamsMux.Unlock()

	if stream, exists := p.activeStreams[streamID]; exists {
		stream.IsActive = false
		delete(p.activeStreams, streamID)
		logger.Info(fmt.Sprintf("Unregistered stream: %s", streamID))
	}

	return nil
}

// collectAudioData 收集音频数据
func (p *MediaProcessor) collectAudioData(stream *StreamProcessor) {
	logger.Debug(fmt.Sprintf("Starting audio collection for stream: %s", stream.StreamID))

	for stream.IsActive {
		// 读取音频RTP包
		rtpPacket, _, err := stream.AudioTrack.ReadRTP()
		if err != nil {
			if stream.IsActive {
				logger.Error(fmt.Sprintf("Failed to read audio RTP: %v", err))
			}
			break
		}

		// 将RTP包数据写入缓冲区
		stream.AudioBuffer.Write(rtpPacket.Payload)
	}

	logger.Debug(fmt.Sprintf("Audio collection stopped for stream: %s", stream.StreamID))
}

// collectVideoData 收集视频数据
func (p *MediaProcessor) collectVideoData(stream *StreamProcessor) {
	logger.Debug(fmt.Sprintf("Starting video collection for stream: %s", stream.StreamID))

	for stream.IsActive {
		// 读取视频RTP包
		rtpPacket, _, err := stream.VideoTrack.ReadRTP()
		if err != nil {
			if stream.IsActive {
				logger.Error(fmt.Sprintf("Failed to read video RTP: %v", err))
			}
			break
		}

		// 将RTP包数据写入缓冲区
		stream.VideoBuffer.Write(rtpPacket.Payload)
	}

	logger.Debug(fmt.Sprintf("Video collection stopped for stream: %s", stream.StreamID))
}

// scheduleProcessing 定期处理调度
func (p *MediaProcessor) scheduleProcessing(stream *StreamProcessor) {
	ticker := time.NewTicker(3 * time.Second) // 每3秒处理一次
	defer ticker.Stop()

	for stream.IsActive {
		select {
		case <-ticker.C:
			p.processStream(stream)
		}
	}
}

// processStream 处理流数据
func (p *MediaProcessor) processStream(stream *StreamProcessor) {
	if time.Since(stream.LastProcessed) < 2*time.Second {
		return // 避免过于频繁的处理
	}

	// 提取音频数据（SFU 模式：不进行格式转换）
	var audioData *AudioData
	if stream.AudioBuffer.Count() > 0 {
		audioBytes := stream.AudioBuffer.ReadAll()
		if len(audioBytes) > 0 {
			// SFU 架构：直接使用原始音频数据，不进行格式转换
			// AI 服务应该能够处理原始 PCM 或 Opus 格式
			audioData = &AudioData{
				Data:       audioBytes,
				Format:     "pcm", // 原始 PCM 格式
				SampleRate: 48000, // WebRTC 默认采样率
				Channels:   2,     // 立体声
				Duration:   3000,  // 3秒
			}
		}
	}

	// 提取视频数据（SFU 模式：不进行格式转换）
	var videoData *VideoData
	if stream.VideoBuffer.Count() > 0 {
		videoBytes := stream.VideoBuffer.ReadAll()
		if len(videoBytes) > 0 {
			// SFU 架构：直接使用原始视频数据，不进行格式转换
			// AI 服务应该能够处理原始 H264/VP8 格式
			videoData = &VideoData{
				Data:     videoBytes,
				Format:   "h264", // 原始 H264 格式
				Width:    640,
				Height:   480,
				FPS:      30,
				Duration: 2000, // 2秒
			}
		}
	}

	// 如果有数据，创建处理任务
	if audioData != nil || videoData != nil {
		task := &ProcessingTask{
			StreamID:  stream.StreamID,
			UserID:    stream.UserID,
			RoomID:    stream.RoomID,
			AudioData: audioData,
			VideoData: videoData,
			Tasks:     stream.AITasks,
			Callback:  p.handleProcessingResult,
			CreatedAt: time.Now(),
		}

		// 提交到处理队列
		select {
		case p.processingQueue <- task:
			stream.LastProcessed = time.Now()
		default:
			logger.Warn("Processing queue is full, dropping task")
		}
	}
}

// worker 工作协程
func (p *MediaProcessor) worker() {
	for task := range p.processingQueue {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// 执行AI处理
		results, err := p.aiClient.BatchProcessing(ctx, task.AudioData, task.VideoData, task.Tasks)

		// 调用回调函数
		if task.Callback != nil {
			task.Callback(results, err)
		}

		cancel()
	}
}

// handleProcessingResult 处理AI推理结果
func (p *MediaProcessor) handleProcessingResult(results map[string]*AIResponse, err error) {
	if err != nil {
		logger.Error(fmt.Sprintf("AI processing failed: %v", err))
		return
	}

	// 处理结果
	for taskType, response := range results {
		logger.Info(fmt.Sprintf("AI task %s completed with code: %d", taskType, response.Code))

		// 根据不同的AI任务类型处理结果
		switch taskType {
		case "speech_recognition":
			p.handleSpeechRecognitionResult(response)
		case "emotion_detection":
			p.handleEmotionDetectionResult(response)
		case "synthesis_detection":
			p.handleSynthesisDetectionResult(response)
		case "audio_denoising":
			p.handleAudioDenoisingResult(response)
		case "video_enhancement":
			p.handleVideoEnhancementResult(response)
		}
	}
}

// handleSpeechRecognitionResult 处理语音识别结果
func (p *MediaProcessor) handleSpeechRecognitionResult(response *AIResponse) {
	if data, ok := response.Data["text"].(string); ok {
		logger.Info(fmt.Sprintf("Speech recognition result: %s", data))
		// TODO: 将识别结果发送给会议服务或存储
	}
}

// handleEmotionDetectionResult 处理情绪检测结果
func (p *MediaProcessor) handleEmotionDetectionResult(response *AIResponse) {
	if data, ok := response.Data["emotion"].(string); ok {
		logger.Info(fmt.Sprintf("Emotion detection result: %s", data))
		// TODO: 将情绪分析结果发送给会议服务
	}
}

// handleSynthesisDetectionResult 处理合成检测结果
func (p *MediaProcessor) handleSynthesisDetectionResult(response *AIResponse) {
	if isSynthetic, ok := response.Data["is_synthetic"].(bool); ok {
		if isSynthetic {
			logger.Warn("Synthetic content detected!")
			// TODO: 触发安全警报
		}
	}
}

// handleAudioDenoisingResult 处理音频降噪结果
func (p *MediaProcessor) handleAudioDenoisingResult(response *AIResponse) {
	logger.Info("Audio denoising completed")
	// TODO: 应用降噪后的音频
}

// handleVideoEnhancementResult 处理视频增强结果
func (p *MediaProcessor) handleVideoEnhancementResult(response *AIResponse) {
	logger.Info("Video enhancement completed")
	// TODO: 应用增强后的视频
}

// GetStreamStatus 获取流状态
func (p *MediaProcessor) GetStreamStatus(streamID string) (*StreamProcessor, bool) {
	p.streamsMux.RLock()
	defer p.streamsMux.RUnlock()

	stream, exists := p.activeStreams[streamID]
	return stream, exists
}

// GetAllStreams 获取所有活跃流
func (p *MediaProcessor) GetAllStreams() map[string]*StreamProcessor {
	p.streamsMux.RLock()
	defer p.streamsMux.RUnlock()

	streams := make(map[string]*StreamProcessor)
	for id, stream := range p.activeStreams {
		streams[id] = stream
	}

	return streams
}

// NewCircularBuffer 创建循环缓冲区
func NewCircularBuffer(maxSize int, duration time.Duration) *CircularBuffer {
	return &CircularBuffer{
		buffer:   make([]byte, maxSize),
		maxSize:  maxSize,
		duration: duration,
	}
}

// Write 写入数据
func (cb *CircularBuffer) Write(data []byte) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	for _, b := range data {
		cb.buffer[cb.head] = b
		cb.head = (cb.head + 1) % cb.maxSize

		if cb.count < cb.maxSize {
			cb.count++
		} else {
			cb.tail = (cb.tail + 1) % cb.maxSize
		}
	}
}

// ReadAll 读取所有数据
func (cb *CircularBuffer) ReadAll() []byte {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.count == 0 {
		return nil
	}

	result := make([]byte, cb.count)
	for i := 0; i < cb.count; i++ {
		result[i] = cb.buffer[(cb.tail+i)%cb.maxSize]
	}

	return result
}

// Count 获取缓冲区数据量
func (cb *CircularBuffer) Count() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.count
}

// Clear 清空缓冲区
func (cb *CircularBuffer) Clear() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.head = 0
	cb.tail = 0
	cb.count = 0
}

// streamProcessing gRPC 流式处理（实时发送音频片段）
func (p *MediaProcessor) streamProcessing(stream *StreamProcessor) {
	ticker := time.NewTicker(500 * time.Millisecond) // 每 500ms 发送一次
	defer ticker.Stop()

	for stream.IsActive {
		select {
		case <-ticker.C:
			// 提取音频数据
			if stream.AudioBuffer.Count() > 0 {
				audioBytes := stream.AudioBuffer.ReadAll()
				if len(audioBytes) > 0 {
					// 发送音频片段到 AI 服务
					stream.Sequence++
					err := p.aiGRPCClient.SendAudioChunk(
						stream.StreamID,
						stream.Sequence,
						audioBytes,
						"pcm", // 原始 PCM 格式
						48000, // 48kHz
						2,     // 2 channels
						stream.AITasks,
						false, // 不是最后一个片段
					)
					if err != nil {
						logger.Error("Failed to send audio chunk",
							logger.String("stream_id", stream.StreamID),
							logger.Err(err))
					}
				}
			}
		}
	}

	// 流结束时发送最后一个片段
	if stream.AIStreamClient != nil {
		if stream.AudioBuffer.Count() > 0 {
			audioBytes := stream.AudioBuffer.ReadAll()
			if len(audioBytes) > 0 {
				stream.Sequence++
				p.aiGRPCClient.SendAudioChunk(
					stream.StreamID,
					stream.Sequence,
					audioBytes,
					"pcm",
					48000,
					2,
					stream.AITasks,
					true, // 最后一个片段
				)
			}
		}

		// 关闭流
		p.aiGRPCClient.CloseAudioStream(stream.StreamID)
	}
}

// receiveAIResults 接收 AI 流式结果
func (p *MediaProcessor) receiveAIResults(stream *StreamProcessor) {
	if stream.AIStreamClient == nil {
		return
	}

	for {
		select {
		case result := <-stream.AIStreamClient.ResultChan:
			// 处理 AI 结果
			logger.Info("Received AI result",
				logger.String("stream_id", stream.StreamID),
				logger.String("result_type", result.ResultType),
				logger.Float64("confidence", result.Confidence),
				logger.Int32("sequence", result.Sequence))

			// TODO: 将结果发送到前端或存储
			// 可以通过 WebSocket 或其他方式通知客户端

		case err := <-stream.AIStreamClient.ErrorChan:
			logger.Error("AI stream error",
				logger.String("stream_id", stream.StreamID),
				logger.Err(err))
			return

		case <-stream.AIStreamClient.Done:
			logger.Info("AI stream done",
				logger.String("stream_id", stream.StreamID))
			return
		}

		if !stream.IsActive {
			return
		}
	}
}
