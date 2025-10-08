package services

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "meeting-system/shared/grpc"
	"meeting-system/shared/logger"
	"meeting-system/shared/tracing"
)

// AIGRPCClient AI 服务 gRPC 客户端
type AIGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.AIServiceClient

	// 流式连接管理
	activeStreams map[string]*AudioStreamClient
	streamMutex   sync.RWMutex
}

// AudioStreamClient 音频流客户端
type AudioStreamClient struct {
	StreamID   string
	Stream     pb.AIService_StreamAudioProcessingClient
	ResultChan chan *pb.AIStreamResult
	ErrorChan  chan error
	Done       chan struct{}
	Mutex      sync.Mutex
}

// NewAIGRPCClient 创建 AI gRPC 客户端
func NewAIGRPCClient(aiServiceAddr string) (*AIGRPCClient, error) {
	// 连接到 AI 服务
	conn, err := grpc.Dial(
		aiServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(tracing.StreamClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AI service: %w", err)
	}

	client := pb.NewAIServiceClient(conn)

	return &AIGRPCClient{
		conn:          conn,
		client:        client,
		activeStreams: make(map[string]*AudioStreamClient),
	}, nil
}

// Close 关闭客户端
func (c *AIGRPCClient) Close() error {
	// 关闭所有活动流
	c.streamMutex.Lock()
	for _, stream := range c.activeStreams {
		close(stream.Done)
	}
	c.streamMutex.Unlock()

	return c.conn.Close()
}

// CheckConnectivity 检查与 AI gRPC 服务的连接
func (c *AIGRPCClient) CheckConnectivity(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("ai grpc client not initialized")
	}

	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := c.client.GetAIAnalysis(checkCtx, &pb.GetAIAnalysisRequest{TaskId: "health_check"}); err != nil {
		return fmt.Errorf("ai grpc health check failed: %w", err)
	}

	return nil
}

// ProcessAudioData 一元 RPC：批量处理音频数据
func (c *AIGRPCClient) ProcessAudioData(ctx context.Context, audioData *AudioData, tasks []string) (*AIResponse, error) {
	// 创建 span
	span, ctx := opentracing.StartSpanFromContext(ctx, "AIGRPCClient.ProcessAudioData")
	defer span.Finish()

	ext.Component.Set(span, "media-service")
	span.SetTag("data_size", len(audioData.Data))
	span.SetTag("tasks", tasks)

	req := &pb.ProcessAudioDataRequest{
		AudioData:  audioData.Data,
		Format:     audioData.Format,
		SampleRate: int32(audioData.SampleRate),
		Channels:   int32(audioData.Channels),
		Tasks:      tasks,
		Duration:   int32(audioData.Duration),
	}

	resp, err := c.client.ProcessAudioData(ctx, req)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("error.message", err.Error())
		return nil, fmt.Errorf("failed to process audio data: %w", err)
	}

	// 转换响应
	return convertGRPCResponse(resp), nil
}

// StartAudioStream 启动音频流处理
func (c *AIGRPCClient) StartAudioStream(ctx context.Context, streamID string, tasks []string) (*AudioStreamClient, error) {
	// 创建 span
	span, ctx := opentracing.StartSpanFromContext(ctx, "AIGRPCClient.StartAudioStream")
	defer span.Finish()

	ext.Component.Set(span, "media-service")
	span.SetTag("stream_id", streamID)
	span.SetTag("tasks", tasks)

	// 创建流
	stream, err := c.client.StreamAudioProcessing(ctx)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("error.message", err.Error())
		return nil, fmt.Errorf("failed to start audio stream: %w", err)
	}

	streamClient := &AudioStreamClient{
		StreamID:   streamID,
		Stream:     stream,
		ResultChan: make(chan *pb.AIStreamResult, 100),
		ErrorChan:  make(chan error, 10),
		Done:       make(chan struct{}),
	}

	// 启动接收 goroutine
	go c.receiveResults(streamClient)

	// 保存流
	c.streamMutex.Lock()
	c.activeStreams[streamID] = streamClient
	c.streamMutex.Unlock()

	logger.Info("Started audio stream",
		logger.String("stream_id", streamID),
		logger.Int("tasks", len(tasks)))

	return streamClient, nil
}

// SendAudioChunk 发送音频片段
func (c *AIGRPCClient) SendAudioChunk(streamID string, sequence int32, audioData []byte, format string, sampleRate, channels int, tasks []string, isFinal bool) error {
	c.streamMutex.RLock()
	streamClient, exists := c.activeStreams[streamID]
	c.streamMutex.RUnlock()

	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	chunk := &pb.AudioChunk{
		Data:       audioData,
		Sequence:   sequence,
		StreamId:   streamID,
		Format:     format,
		SampleRate: int32(sampleRate),
		Channels:   int32(channels),
		Tasks:      tasks,
		IsFinal:    isFinal,
	}

	streamClient.Mutex.Lock()
	err := streamClient.Stream.Send(chunk)
	streamClient.Mutex.Unlock()

	if err != nil {
		logger.Error("Failed to send audio chunk",
			logger.String("stream_id", streamID),
			logger.Err(err))
		return err
	}

	logger.Debug("Sent audio chunk",
		logger.String("stream_id", streamID),
		logger.Int32("sequence", sequence),
		logger.Int("size", len(audioData)),
		logger.Bool("is_final", isFinal))

	return nil
}

// CloseAudioStream 关闭音频流
func (c *AIGRPCClient) CloseAudioStream(streamID string) error {
	c.streamMutex.Lock()
	streamClient, exists := c.activeStreams[streamID]
	if exists {
		delete(c.activeStreams, streamID)
	}
	c.streamMutex.Unlock()

	if !exists {
		return fmt.Errorf("stream not found: %s", streamID)
	}

	streamClient.Mutex.Lock()
	err := streamClient.Stream.CloseSend()
	streamClient.Mutex.Unlock()

	close(streamClient.Done)

	logger.Info("Closed audio stream", logger.String("stream_id", streamID))

	return err
}

// receiveResults 接收流式结果
func (c *AIGRPCClient) receiveResults(streamClient *AudioStreamClient) {
	for {
		select {
		case <-streamClient.Done:
			return
		default:
			result, err := streamClient.Stream.Recv()
			if err == io.EOF {
				logger.Info("Stream ended", logger.String("stream_id", streamClient.StreamID))
				return
			}
			if err != nil {
				logger.Error("Failed to receive result",
					logger.String("stream_id", streamClient.StreamID),
					logger.Err(err))
				select {
				case streamClient.ErrorChan <- err:
				case <-streamClient.Done:
					return
				}
				return
			}

			// 发送结果到通道
			select {
			case streamClient.ResultChan <- result:
				logger.Debug("Received AI result",
					logger.String("stream_id", streamClient.StreamID),
					logger.String("result_type", result.ResultType),
					logger.Int32("sequence", result.Sequence))
			case <-streamClient.Done:
				return
			}
		}
	}
}

// 辅助函数

func convertGRPCResponse(resp *pb.ProcessAudioDataResponse) *AIResponse {
	results := make(map[string]interface{})
	for task, result := range resp.Results {
		results[task] = map[string]interface{}{
			"result_type": result.ResultType,
			"result_data": result.ResultData,
			"confidence":  result.Confidence,
		}
	}

	return &AIResponse{
		Code:    0,
		Message: resp.Status,
		Data:    results,
	}
}
