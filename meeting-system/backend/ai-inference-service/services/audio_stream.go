package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	pb "meeting-system/shared/grpc"
	"meeting-system/shared/logger"
)

// AudioStreamSession aggregates audio chunks for streaming inference.
type AudioStreamSession struct {
	streamID   string
	tasks      []string
	format     string
	sampleRate int
	channels   int
	buffer     []byte
	aiService  *AIInferenceService
	streamCfg  streamConfig
	lastFlush  time.Time
}

type streamConfig struct {
	flushInterval time.Duration
	maxBufferMs   int
}

// NewAudioStreamSession initializes a stream session.
func NewAudioStreamSession(ai *AIInferenceService, streamID string, tasks []string, format string, sampleRate, channels int) *AudioStreamSession {
	cfg := streamConfig{}
	if ai != nil && ai.config != nil {
		cfg.maxBufferMs = ai.config.AI.Streaming.MaxBufferMs
		if ai.config.AI.Streaming.FlushIntervalMs > 0 {
			cfg.flushInterval = time.Duration(ai.config.AI.Streaming.FlushIntervalMs) * time.Millisecond
		}
	}

	if format == "" {
		format = "pcm"
	}
	if sampleRate == 0 {
		sampleRate = 48000
	}
	if channels == 0 {
		channels = 1
	}

	return &AudioStreamSession{
		streamID:   streamID,
		tasks:      tasks,
		format:     format,
		sampleRate: sampleRate,
		channels:   channels,
		buffer:     make([]byte, 0, 1024*64),
		aiService:  ai,
		streamCfg:  cfg,
		lastFlush:  time.Now(),
	}
}

// Append adds a chunk to the buffer and returns results if a flush is triggered.
func (s *AudioStreamSession) Append(ctx context.Context, chunk *pb.AudioChunk) ([]*pb.AIStreamResult, error) {
	if chunk == nil {
		return nil, fmt.Errorf("audio chunk is nil")
	}
	if s == nil {
		return nil, fmt.Errorf("stream session not initialized")
	}

	if chunk.Format != "" {
		s.format = chunk.Format
	}
	if chunk.SampleRate > 0 {
		s.sampleRate = int(chunk.SampleRate)
	}
	if chunk.Channels > 0 {
		s.channels = int(chunk.Channels)
	}
	if len(chunk.Tasks) > 0 {
		s.tasks = chunk.Tasks
	}

	s.buffer = append(s.buffer, chunk.Data...)
	s.ensureBufferLimit()

	if chunk.IsFinal {
		return s.flush(ctx, chunk.Sequence, true)
	}

	if s.shouldFlush() {
		results, err := s.flush(ctx, chunk.Sequence, false)
		if err != nil {
			return nil, err
		}
		s.lastFlush = time.Now()
		return results, nil
	}

	return nil, nil
}

func (s *AudioStreamSession) ensureBufferLimit() {
	if s.streamCfg.maxBufferMs <= 0 || s.sampleRate == 0 || s.channels == 0 {
		return
	}

	bytesPerSecond := s.sampleRate * s.channels * 2
	maxBytes := int(float64(bytesPerSecond) * (float64(s.streamCfg.maxBufferMs) / 1000.0))
	if maxBytes <= 0 {
		return
	}
	if len(s.buffer) <= maxBytes {
		return
	}

	// Drop oldest data to cap buffer size.
	drop := len(s.buffer) - maxBytes
	if drop >= len(s.buffer) {
		s.buffer = s.buffer[:0]
		return
	}
	s.buffer = append([]byte{}, s.buffer[drop:]...)
	logger.Warn("Audio stream buffer trimmed",
		logger.String("stream_id", s.streamID),
		logger.Int("dropped_bytes", drop))
}

func (s *AudioStreamSession) flush(ctx context.Context, sequence int32, isFinal bool) ([]*pb.AIStreamResult, error) {
	if s.aiService == nil {
		return nil, fmt.Errorf("ai service not available")
	}

	results := make([]*pb.AIStreamResult, 0)
	streamID := s.streamID
	if streamID == "" {
		streamID = "stream"
	}

	tasks := normalizeTasks(s.tasks)
	if len(tasks) == 0 {
		tasks = []string{"speech_recognition"}
	}

	pcmData := append([]byte{}, s.buffer...)
	sampleRate := s.sampleRate
	channels := s.channels
	if strings.ToLower(strings.TrimSpace(s.format)) != "pcm" {
		decoded, sr, ch, err := normalizeAudioPayload(pcmData, s.format, sampleRate, channels)
		if err != nil {
			return nil, err
		}
		pcmData = decoded
		sampleRate = sr
		channels = ch
	}

	for _, task := range tasks {
		switch task {
		case "speech_recognition", "asr":
			resp, err := s.aiService.SpeechRecognitionPCM(ctx, pcmData, sampleRate, channels, s.format, "")
			if err != nil {
				return nil, err
			}
			results = append(results, streamResult(streamID, sequence, "speech_recognition", resp, resp.Confidence, isFinal))
		case "emotion_detection", "emotion":
			resp, err := s.aiService.EmotionDetectionPCM(ctx, pcmData, sampleRate, channels, s.format)
			if err != nil {
				return nil, err
			}
			results = append(results, streamResult(streamID, sequence, "emotion_detection", resp, resp.Confidence, isFinal))
		case "synthesis_detection", "synthesis":
			resp, err := s.aiService.SynthesisDetectionPCM(ctx, pcmData, sampleRate, channels, s.format)
			if err != nil {
				return nil, err
			}
			results = append(results, streamResult(streamID, sequence, "synthesis_detection", resp, resp.Confidence, isFinal))
		}
	}

	return results, nil
}

func (s *AudioStreamSession) shouldFlush() bool {
	if s.streamCfg.flushInterval > 0 {
		if time.Since(s.lastFlush) >= s.streamCfg.flushInterval {
			return true
		}
	}
	if s.streamCfg.maxBufferMs > 0 && s.sampleRate > 0 && s.channels > 0 {
		limit := int64(s.sampleRate*s.channels*2) * int64(s.streamCfg.maxBufferMs) / 1000
		if limit > 0 && int64(len(s.buffer)) >= limit {
			return true
		}
	}
	return false
}

func normalizeTasks(tasks []string) []string {
	out := make([]string, 0, len(tasks))
	seen := make(map[string]struct{})
	for _, task := range tasks {
		name := strings.ToLower(strings.TrimSpace(task))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func streamResult(streamID string, sequence int32, resultType string, payload interface{}, confidence float64, isFinal bool) *pb.AIStreamResult {
	encoded := "{}"
	if payload != nil {
		if data, err := json.Marshal(payload); err == nil {
			encoded = string(data)
		}
	}

	return &pb.AIStreamResult{
		StreamId:   streamID,
		Sequence:   sequence,
		ResultType: resultType,
		ResultData: encoded,
		Confidence: confidence,
		IsFinal:    isFinal,
		Timestamp:  timestamppb.Now(),
	}
}
