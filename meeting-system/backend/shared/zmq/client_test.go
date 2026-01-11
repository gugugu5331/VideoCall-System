//go:build zmq
// +build zmq

package zmq

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pebbe/zmq4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"meeting-system/shared/config"
)

// MockZMQServer 模拟ZMQ服务器用于测试
type MockZMQServer struct {
	socket  *zmq4.Socket
	context *zmq4.Context
	port    int
	stopCh  chan bool
}

// NewMockZMQServer 创建模拟ZMQ服务器
func NewMockZMQServer(port int) (*MockZMQServer, error) {
	context, err := zmq4.NewContext()
	if err != nil {
		return nil, err
	}

	socket, err := context.NewSocket(zmq4.REP)
	if err != nil {
		context.Term()
		return nil, err
	}

	endpoint := fmt.Sprintf("tcp://*:%d", port)
	if err := socket.Bind(endpoint); err != nil {
		socket.Close()
		context.Term()
		return nil, err
	}

	server := &MockZMQServer{
		socket:  socket,
		context: context,
		port:    port,
		stopCh:  make(chan bool, 1),
	}

	go server.handleRequests()
	return server, nil
}

// handleRequests 处理请求
func (s *MockZMQServer) handleRequests() {
	for {
		select {
		case <-s.stopCh:
			return
		default:
			// 设置接收超时
			s.socket.SetRcvtimeo(100 * time.Millisecond)

			data, err := s.socket.RecvBytes(0)
			if err != nil {
				continue // 超时或其他错误，继续循环
			}

			// 解析请求
			var request AIRequest
			if err := json.Unmarshal(data, &request); err != nil {
				continue
			}

			// 构建响应
			response := &AIResponse{
				RequestID: request.RequestID,
				WorkID:    request.WorkID,
				Object:    request.Object,
				Data:      map[string]interface{}{"status": "success", "message": "mock response"},
				Error:     nil,
			}

			// 特殊处理健康检查
			if request.Object == "health_check" {
				response.Data = map[string]interface{}{"status": "healthy"}
			}

			// 发送响应
			responseData, _ := json.Marshal(response)
			s.socket.SendBytes(responseData, 0)
		}
	}
}

// Close 关闭模拟服务器
func (s *MockZMQServer) Close() {
	select {
	case s.stopCh <- true:
	default:
	}

	if s.socket != nil {
		s.socket.Close()
	}
	if s.context != nil {
		s.context.Term()
	}
}

func TestNewZMQClient(t *testing.T) {
	// 启动模拟服务器
	server, err := NewMockZMQServer(15001)
	require.NoError(t, err)
	defer server.Close()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端配置
	cfg := config.ZMQConfig{
		UnitManagerHost: "localhost",
		UnitManagerPort: 15001,
		UnitName:        "test_unit",
		Timeout:         5,
	}

	// 创建客户端
	client, err := NewZMQClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	assert.True(t, client.connected)
	assert.NotNil(t, client.socket)
	assert.NotNil(t, client.context)
}

func TestZMQClient_SendRequest(t *testing.T) {
	// 启动模拟服务器
	server, err := NewMockZMQServer(15002)
	require.NoError(t, err)
	defer server.Close()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	cfg := config.ZMQConfig{
		UnitManagerHost: "localhost",
		UnitManagerPort: 15002,
		UnitName:        "test_unit",
		Timeout:         5,
	}

	client, err := NewZMQClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// 发送请求
	request := &AIRequest{
		RequestID: "test_001",
		WorkID:    "test_work",
		Object:    "test_object",
		Data:      map[string]interface{}{"test": "data"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.SendRequest(ctx, request)
	require.NoError(t, err)
	assert.Equal(t, "test_001", response.RequestID)
	assert.Equal(t, "test_work", response.WorkID)
	assert.Equal(t, "test_object", response.Object)
	assert.Nil(t, response.Error)
}

func TestZMQClient_HealthCheck(t *testing.T) {
	// 启动模拟服务器
	server, err := NewMockZMQServer(15003)
	require.NoError(t, err)
	defer server.Close()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	cfg := config.ZMQConfig{
		UnitManagerHost: "localhost",
		UnitManagerPort: 15003,
		UnitName:        "test_unit",
		Timeout:         5,
	}

	client, err := NewZMQClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// 执行健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.HealthCheck(ctx)
	assert.NoError(t, err)
}

func TestZMQClient_SpeechRecognition(t *testing.T) {
	// 启动模拟服务器
	server, err := NewMockZMQServer(15004)
	require.NoError(t, err)
	defer server.Close()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	cfg := config.ZMQConfig{
		UnitManagerHost: "localhost",
		UnitManagerPort: 15004,
		UnitName:        "test_unit",
		Timeout:         5,
	}

	client, err := NewZMQClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// 测试语音识别
	data := &SpeechRecognitionData{
		AudioFormat: "wav",
		SampleRate:  16000,
		Channels:    1,
		AudioData:   "base64_encoded_audio_data",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.SpeechRecognition(ctx, "test_speech_001", data)
	require.NoError(t, err)
	assert.Equal(t, "test_speech_001", response.RequestID)
	assert.Equal(t, "speech_recognition", response.Object)
}

func TestZMQClient_EmotionDetection(t *testing.T) {
	// 启动模拟服务器
	server, err := NewMockZMQServer(15005)
	require.NoError(t, err)
	defer server.Close()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	cfg := config.ZMQConfig{
		UnitManagerHost: "localhost",
		UnitManagerPort: 15005,
		UnitName:        "test_unit",
		Timeout:         5,
	}

	client, err := NewZMQClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// 测试情绪识别
	data := &EmotionDetectionData{
		ImageFormat: "jpg",
		ImageData:   "base64_encoded_image_data",
		Width:       640,
		Height:      480,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.EmotionDetection(ctx, "test_emotion_001", data)
	require.NoError(t, err)
	assert.Equal(t, "test_emotion_001", response.RequestID)
	assert.Equal(t, "emotion_detection", response.Object)
}

func TestZMQClient_ConnectionFailure(t *testing.T) {
	// 测试连接失败的情况
	cfg := config.ZMQConfig{
		UnitManagerHost: "192.0.2.1", // 不可达的IP地址
		UnitManagerPort: 19999,
		UnitName:        "test_unit",
		Timeout:         1, // 短超时
	}

	client, err := NewZMQClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestZMQClient_Reconnection(t *testing.T) {
	// 启动模拟服务器
	server, err := NewMockZMQServer(15006)
	require.NoError(t, err)

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建客户端
	cfg := config.ZMQConfig{
		UnitManagerHost: "localhost",
		UnitManagerPort: 15006,
		UnitName:        "test_unit",
		Timeout:         5,
	}

	client, err := NewZMQClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// 关闭服务器模拟连接断开
	server.Close()
	time.Sleep(100 * time.Millisecond)

	// 重新启动服务器
	server, err = NewMockZMQServer(15006)
	require.NoError(t, err)
	defer server.Close()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 测试重连
	err = client.reconnect()
	assert.NoError(t, err)
	assert.True(t, client.connected)
}
