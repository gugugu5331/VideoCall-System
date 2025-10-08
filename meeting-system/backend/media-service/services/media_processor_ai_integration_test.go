package services

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"meeting-system/shared/config"
)

func TestMediaProcessor_AIIntegration_SharedPlayback(t *testing.T) {
	audioPath := filepath.Join("..", "test_video", "20250602_215504.mp3")
	videoPath := filepath.Join("..", "test_video", "20250928_165500.mp4")

	audioBytes, err := os.ReadFile(audioPath)
	require.NoError(t, err)
	require.NotEmpty(t, audioBytes)

	videoBytes, err := os.ReadFile(videoPath)
	require.NoError(t, err)
	require.NotEmpty(t, videoBytes)

	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var resp AIResponse
		switch r.URL.Path {
		case "/api/v1/speech/recognition":
			resp = AIResponse{Code: 0, Message: "ok", Data: map[string]interface{}{"text": "transcript"}}
		case "/api/v1/video/enhancement":
			resp = AIResponse{Code: 0, Message: "ok", Data: map[string]interface{}{"quality": "enhanced"}}
		default:
			resp = AIResponse{Code: 1, Message: "unknown endpoint"}
			w.WriteHeader(http.StatusNotFound)
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer aiServer.Close()

	parsedURL, err := url.Parse(aiServer.URL)
	require.NoError(t, err)
	host, portStr, err := net.SplitHostPort(parsedURL.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	cfg := &config.Config{
		Services: config.ServicesConfig{
			AIService: config.ServiceConfig{Host: host, Port: port, Timeout: 10 * time.Second},
		},
	}

	aiClient := NewAIClient(cfg)
	processor := NewMediaProcessor(cfg, aiClient, nil)
	defer close(processor.processingQueue)

	done := make(chan map[string]*AIResponse, 1)
	roomResults := sync.Map{}

	roomID := "room-shared"
	streamID := uuid.New().String()

	task := &ProcessingTask{
		StreamID: streamID,
		UserID:   "user-uploader",
		RoomID:   roomID,
		AudioData: &AudioData{
			Data:       audioBytes,
			Format:     "mp3",
			SampleRate: 44100,
			Channels:   2,
			Duration:   5000,
		},
		VideoData: &VideoData{
			Data:     videoBytes,
			Format:   "mp4",
			Width:    1280,
			Height:   720,
			FPS:      30,
			Duration: 5000,
		},
		Tasks: []string{"speech_recognition", "video_enhancement"},
		Callback: func(results map[string]*AIResponse, err error) {
			require.NoError(t, err)
			roomResults.Store(roomID, results)
			done <- results
		},
		CreatedAt: time.Now(),
	}

	processor.processingQueue <- task

	var results map[string]*AIResponse
	select {
	case results = <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("AI processing did not complete in time")
	}

	require.Contains(t, results, "speech_recognition")
	require.Contains(t, results, "video_enhancement")
	require.Equal(t, "transcript", results["speech_recognition"].Data["text"])
	require.Equal(t, "enhanced", results["video_enhancement"].Data["quality"])

	value, ok := roomResults.Load(roomID)
	require.True(t, ok)
	shared := value.(map[string]*AIResponse)

	for _, member := range []string{"member-alpha", "member-beta"} {
		require.Equal(t, "transcript", shared["speech_recognition"].Data["text"], "member %s should receive transcript", member)
		require.Equal(t, "enhanced", shared["video_enhancement"].Data["quality"], "member %s should receive enhanced quality metadata", member)
	}
}
