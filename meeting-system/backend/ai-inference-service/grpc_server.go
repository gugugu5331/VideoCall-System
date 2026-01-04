package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"meeting-system/ai-inference-service/services"
	"meeting-system/shared/config"
	pb "meeting-system/shared/grpc"
	"meeting-system/shared/logger"
	"meeting-system/shared/tracing"
)

type aiGRPCServer struct {
	pb.UnimplementedAIServiceServer
	aiService *services.AIInferenceService
}

func newAIGrpcServer(aiService *services.AIInferenceService) *aiGRPCServer {
	return &aiGRPCServer{aiService: aiService}
}

func (s *aiGRPCServer) ProcessAudioData(ctx context.Context, req *pb.ProcessAudioDataRequest) (*pb.ProcessAudioDataResponse, error) {
	taskID := fmt.Sprintf("ai_%d", time.Now().UnixNano())
	now := timestamppb.Now()

	audio := req.GetAudioData()
	if len(audio) == 0 {
		return &pb.ProcessAudioDataResponse{
			TaskId:  taskID,
			Status:  "error",
			Error:   "audio_data is required",
			Results: map[string]*pb.AIResult{},
		}, nil
	}

	format := req.GetFormat()
	if strings.TrimSpace(format) == "" {
		format = "pcm"
	}

	sampleRate := int(req.GetSampleRate())
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	channels := int(req.GetChannels())
	if channels <= 0 {
		channels = 1
	}

	taskList := req.GetTasks()
	if len(taskList) == 0 {
		taskList = []string{"speech_recognition"}
	}

	b64 := base64.StdEncoding.EncodeToString(audio)
	results := make(map[string]*pb.AIResult, len(taskList))

	var firstErr error

	for _, rawTask := range taskList {
		task := strings.ToLower(strings.TrimSpace(rawTask))
		if task == "" {
			continue
		}

		switch task {
		case "speech_recognition", "asr":
			resp, err := s.aiService.SpeechRecognition(ctx, &services.ASRRequest{
				AudioData:  b64,
				Format:     format,
				SampleRate: sampleRate,
				Channels:   channels,
			})
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			payload, _ := json.Marshal(resp)
			results[rawTask] = &pb.AIResult{
				ResultType: task,
				ResultData: string(payload),
				Confidence: resp.Confidence,
				CreatedAt:  now,
			}
		case "synthesis_detection", "synthesis":
			resp, err := s.aiService.SynthesisDetection(ctx, &services.SynthesisDetectionRequest{
				AudioData:  b64,
				Format:     format,
				SampleRate: sampleRate,
				Channels:   channels,
			})
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}

			payload, _ := json.Marshal(resp)
			results[rawTask] = &pb.AIResult{
				ResultType: task,
				ResultData: string(payload),
				Confidence: resp.Confidence,
				CreatedAt:  now,
			}
		case "emotion_detection", "emotion":
			resp, err := s.aiService.EmotionDetection(ctx, &services.EmotionRequest{
				AudioData:  b64,
				Format:     format,
				SampleRate: sampleRate,
				Channels:   channels,
			})
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}

			payload, _ := json.Marshal(resp)
			results[rawTask] = &pb.AIResult{
				ResultType: task,
				ResultData: string(payload),
				Confidence: resp.Confidence,
				CreatedAt:  now,
			}
		default:
			if firstErr == nil {
				firstErr = fmt.Errorf("unsupported task: %s", rawTask)
			}
		}
	}

	status := "ok"
	errText := ""
	if firstErr != nil && len(results) == 0 {
		status = "error"
		errText = firstErr.Error()
	} else if firstErr != nil {
		status = "partial"
		errText = firstErr.Error()
	}

	return &pb.ProcessAudioDataResponse{
		TaskId:  taskID,
		Status:  status,
		Error:   errText,
		Results: results,
	}, nil
}

func (s *aiGRPCServer) ProcessVideoFrame(ctx context.Context, req *pb.ProcessVideoFrameRequest) (*pb.ProcessVideoFrameResponse, error) {
	return &pb.ProcessVideoFrameResponse{
		TaskId:  fmt.Sprintf("ai_video_%d", time.Now().UnixNano()),
		Status:  "unimplemented",
		Error:   "video processing is not implemented by ai-inference-service",
		Results: map[string]*pb.AIResult{},
	}, nil
}

func (s *aiGRPCServer) StreamAudioProcessing(stream pb.AIService_StreamAudioProcessingServer) error {
	ctx := stream.Context()
	var session *services.AudioStreamSession

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if session == nil {
			session = services.NewAudioStreamSession(
				s.aiService,
				chunk.GetStreamId(),
				chunk.GetTasks(),
				chunk.GetFormat(),
				int(chunk.GetSampleRate()),
				int(chunk.GetChannels()),
			)
		}

		results, err := session.Append(ctx, chunk)
		if err != nil {
			return err
		}
		for _, result := range results {
			if err := stream.Send(result); err != nil {
				return err
			}
		}

		if chunk.GetIsFinal() {
			return nil
		}
	}
}

func (s *aiGRPCServer) StreamVideoProcessing(stream pb.AIService_StreamVideoProcessingServer) error {
	return fmt.Errorf("video streaming is not implemented by ai-inference-service")
}

func startAIGrpcServer(port int, aiService *services.AIInferenceService) (*grpc.Server, error) {
	if port <= 0 {
		return nil, fmt.Errorf("invalid grpc port: %d", port)
	}

	grpcHost := ""
	if cfg := config.GlobalConfig; cfg != nil {
		grpcHost = cfg.Server.Host
	}
	if grpcHost == "" {
		grpcHost = "0.0.0.0"
	}

	grpcAddr := fmt.Sprintf("%s:%d", grpcHost, port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(tracing.UnaryServerInterceptor()),
	)
	pb.RegisterAIServiceServer(grpcServer, newAIGrpcServer(aiService))

	go func() {
		logger.Info("AI service gRPC server starting on " + grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("Failed to start AI gRPC server: " + err.Error())
		}
	}()

	logger.Info("AI service gRPC started successfully on " + grpcAddr)
	return grpcServer, nil
}
