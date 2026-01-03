package services

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"meeting-system/shared/logger"
)

// EdgeLLMClient Edge-LLM-Infra 客户端
type EdgeLLMClient struct {
	host    string
	port    int
	timeout time.Duration
	mu      sync.Mutex
}

// EdgeLLMRequest Edge-LLM-Infra 请求格式
type EdgeLLMRequest struct {
	RequestID string                 `json:"request_id"`
	WorkID    string                 `json:"work_id"`
	Action    string                 `json:"action"`
	Object    string                 `json:"object,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// EdgeLLMResponse Edge-LLM-Infra 响应格式
// 注意：Data 和 Error 字段使用 json.RawMessage 以支持灵活的响应格式
// Edge-LLM-Infra 可能返回字符串或对象，我们需要在接收后再解析
type EdgeLLMResponse struct {
	RequestID string          `json:"request_id"`
	WorkID    string          `json:"work_id"`
	Action    string          `json:"action"`
	Object    string          `json:"object,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Error     json.RawMessage `json:"error,omitempty"`
}

// InferenceSession 推理会话
type InferenceSession struct {
	WorkID     string
	ModelType  string
	Conn       net.Conn
	Reader     *bufio.Reader
	CreatedAt  time.Time
	LastUsedAt time.Time
}

// NewEdgeLLMClient 创建 Edge-LLM-Infra 客户端
func NewEdgeLLMClient(host string, port int, timeout time.Duration) *EdgeLLMClient {
	return &EdgeLLMClient{
		host:    host,
		port:    port,
		timeout: timeout,
	}
}

// createConnection 创建 TCP 连接
func (c *EdgeLLMClient) createConnection(ctx context.Context) (net.Conn, *bufio.Reader, error) {
	// 创建带超时的连接
	dialer := &net.Dialer{
		Timeout: c.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", c.host, c.port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to unit-manager: %w", err)
	}

	// unit-manager 的 TCP 服务端对“消息边界”不健壮：可能把多次 Write 合并成一次 onMessage，
	// 或把一次大 Write 拆成多次 onMessage，从而触发 simdjson 的 "json format error"。
	// 这里尽量让每次 Write 尽快发出，降低合并概率（不能完全保证）。
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
	}

	// 设置读写超时
	if err := conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to set connection deadline: %w", err)
	}

	reader := bufio.NewReader(conn)
	return conn, reader, nil
}

// sendRequest 发送请求
func (c *EdgeLLMClient) sendRequest(conn net.Conn, req *EdgeLLMRequest) error {
	// 序列化请求
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求（添加换行符）
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	logger.Debug(fmt.Sprintf("Sent request: %s", string(data)))
	return nil
}

// receiveResponse 接收响应
func (c *EdgeLLMClient) receiveResponse(reader *bufio.Reader) (*EdgeLLMResponse, error) {
	// 读取响应（直到换行符）
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 详细日志：输出完整的原始响应
	logger.Info(fmt.Sprintf("📥 RAW RESPONSE: %s", line))
	logger.Debug(fmt.Sprintf("Response length: %d bytes", len(line)))

	// 解析响应
	var resp EdgeLLMResponse
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to unmarshal response. Raw response: %s", line))
		logger.Error(fmt.Sprintf("Unmarshal error: %v", err))
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 输出解析后的响应结构
	logger.Info(fmt.Sprintf("📦 PARSED RESPONSE: request_id=%s, work_id=%s, action=%s, object=%s",
		resp.RequestID, resp.WorkID, resp.Action, resp.Object))

	if resp.Data != nil && len(resp.Data) > 0 {
		logger.Info(fmt.Sprintf("📊 DATA FIELD: %s", string(resp.Data)))
	}

	if resp.Error != nil && len(resp.Error) > 0 {
		logger.Warn(fmt.Sprintf("⚠️ ERROR FIELD: %s", string(resp.Error)))
	}

	// 检查错误字段
	if resp.Error != nil && len(resp.Error) > 0 {
		// 解析 Error 字段
		errorData, err := parseErrorField(&resp)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to parse error field: %v", err))
			// 即使解析失败，也返回原始错误信息
			return nil, fmt.Errorf("edge-llm error: failed to parse error field")
		}

		// 检查错误代码
		if errorData != nil {
			if code, ok := errorData["code"].(float64); ok && code != 0 {
				message := "unknown error"
				if msg, ok := errorData["message"].(string); ok {
					message = msg
				}
				logger.Error(fmt.Sprintf("❌ Edge-LLM Error: code=%d, message=%s", int(code), message))
				return nil, fmt.Errorf("edge-llm error (code %d): %s", int(code), message)
			}
		}
	}

	return &resp, nil
}

type streamWrapper struct {
	Index  int
	Delta  string
	Finish bool
}

func parseStreamWrapper(data map[string]interface{}) (streamWrapper, bool) {
	if data == nil {
		return streamWrapper{}, false
	}

	finish, ok := data["finish"].(bool)
	if !ok {
		return streamWrapper{}, false
	}

	delta, _ := data["delta"].(string)
	index := -1
	if idx, ok := data["index"].(float64); ok {
		index = int(idx)
	}

	return streamWrapper{
		Index:  index,
		Delta:  delta,
		Finish: finish,
	}, true
}

// receiveData reads responses and, if the response is a stream wrapper, drains until finish=true and
// concatenates all delta fragments (to avoid leaving trailing finish frames in the connection buffer).
func (c *EdgeLLMClient) receiveData(ctx context.Context, reader *bufio.Reader) (map[string]interface{}, error) {
	var deltaBuilder strings.Builder
	lastIndex := -1
	seenStream := false

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := c.receiveResponse(reader)
		if err != nil {
			return nil, err
		}

		data, err := parseDataField(resp)
		if err != nil {
			return nil, err
		}

		wrapper, ok := parseStreamWrapper(data)
		if !ok {
			// Non-stream response; return as-is.
			return data, nil
		}

		seenStream = true
		if wrapper.Index >= 0 {
			lastIndex = wrapper.Index
		}
		if wrapper.Delta != "" {
			deltaBuilder.WriteString(wrapper.Delta)
		}
		if wrapper.Finish {
			break
		}
	}

	if !seenStream {
		return nil, nil
	}

	// Return a single synthesized stream wrapper.
	return map[string]interface{}{
		"index":  lastIndex,
		"delta":  deltaBuilder.String(),
		"finish": true,
	}, nil
}

// Setup 设置推理任务
func (c *EdgeLLMClient) Setup(ctx context.Context, modelType string) (*InferenceSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Info(fmt.Sprintf("🔧 Setting up inference session for model: %s", modelType))

	// 创建连接
	logger.Debug(fmt.Sprintf("Creating connection to %s:%d...", c.host, c.port))
	startTime := time.Now()
	conn, reader, err := c.createConnection(ctx)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to create connection: %v", err))
		return nil, err
	}
	logger.Info(fmt.Sprintf("✅ Connection established in %v", time.Since(startTime)))

	// 构建 setup 请求
	setupReq := &EdgeLLMRequest{
		RequestID: generateRequestID(),
		WorkID:    "llm", // 必须是 "llm"，匹配 unit name
		Action:    "setup",
		Object:    "llm.setup",
		Data: map[string]interface{}{
			"model":           modelType,
			"response_format": "llm.utf-8.stream",
			"input":           "llm.utf-8.stream",
			"enoutput":        true,
		},
	}

	logger.Info(fmt.Sprintf("📤 Sending setup request: request_id=%s, work_id=%s, model=%s",
		setupReq.RequestID, setupReq.WorkID, modelType))

	// 发送 setup 请求
	if err := c.sendRequest(conn, setupReq); err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to send setup request: %v", err))
		conn.Close()
		return nil, err
	}

	// 接收 setup 响应
	logger.Info("⏳ Waiting for setup response...")
	receiveStartTime := time.Now()
	setupResp, err := c.receiveResponse(reader)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to receive setup response after %v: %v", time.Since(receiveStartTime), err))
		conn.Close()
		return nil, err
	}
	logger.Info(fmt.Sprintf("✅ Received setup response in %v", time.Since(receiveStartTime)))

	// 获取 work_id
	workID := setupResp.WorkID
	if workID == "" {
		logger.Error("❌ Setup response missing work_id")
		conn.Close()
		return nil, fmt.Errorf("setup response missing work_id")
	}

	logger.Info(fmt.Sprintf("✅ Setup successful, work_id: %s (total time: %v)", workID, time.Since(startTime)))

	// 创建会话
	session := &InferenceSession{
		WorkID:     workID,
		ModelType:  modelType,
		Conn:       conn,
		Reader:     reader,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	return session, nil
}

// Inference 执行推理
func (c *EdgeLLMClient) Inference(ctx context.Context, session *InferenceSession, inputData string) (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Info(fmt.Sprintf("🚀 Starting inference with work_id: %s, model: %s", session.WorkID, session.ModelType))
	logger.Debug(fmt.Sprintf("Input data length: %d bytes", len(inputData)))

	// 更新最后使用时间
	session.LastUsedAt = time.Now()

	// 设置连接超时
	if err := session.Conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to set connection deadline: %v", err))
		return nil, fmt.Errorf("failed to set connection deadline: %w", err)
	}
	logger.Debug(fmt.Sprintf("Connection deadline set to: %v", time.Now().Add(c.timeout)))

	// 构建 inference 请求
	inferenceReq := &EdgeLLMRequest{
		RequestID: generateRequestID(),
		WorkID:    session.WorkID,
		Action:    "inference",
		Object:    "llm.utf-8.stream",
		Data: map[string]interface{}{
			"delta":  inputData,
			"index":  0,
			"finish": true,
		},
	}

	logger.Info(fmt.Sprintf("📤 Sending inference request: request_id=%s, work_id=%s, action=%s",
		inferenceReq.RequestID, inferenceReq.WorkID, inferenceReq.Action))

	// 发送 inference 请求
	startTime := time.Now()
	if err := c.sendRequest(session.Conn, inferenceReq); err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to send inference request: %v", err))
		return nil, err
	}
	logger.Debug(fmt.Sprintf("Request sent in %v", time.Since(startTime)))

	// 接收 inference 响应
	logger.Info(fmt.Sprintf("⏳ Waiting for inference response (timeout: %v)...", c.timeout))
	receiveStartTime := time.Now()
	data, err := c.receiveData(ctx, session.Reader)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to receive inference response after %v: %v", time.Since(receiveStartTime), err))
		return nil, err
	}
	logger.Info(fmt.Sprintf("✅ Received inference response in %v", time.Since(receiveStartTime)))

	logger.Info(fmt.Sprintf("✅ Inference successful for work_id: %s (total time: %v)", session.WorkID, time.Since(startTime)))

	return data, nil
}

// InferenceDelta 执行一次推理，并解析 stream wrapper 的 delta 字段（不包含 setup/exit；用于会话复用）。
func (c *EdgeLLMClient) InferenceDelta(ctx context.Context, session *InferenceSession, inputData string) (map[string]interface{}, error) {
	result, err := c.Inference(ctx, session, inputData)
	if err != nil {
		return nil, err
	}

	if deltaStr, ok := result["delta"].(string); ok && deltaStr != "" {
		var deltaData map[string]interface{}
		if err := json.Unmarshal([]byte(deltaStr), &deltaData); err != nil {
			logger.Warn(fmt.Sprintf("Failed to parse delta field as JSON: %v", err))
			return result, nil
		}
		return deltaData, nil
	}
	return result, nil
}

// InferenceDeltaWithAudioStream 执行流式推理（音频大 payload 分块发送），并解析最终结果（不包含 setup/exit；用于会话复用）。
func (c *EdgeLLMClient) InferenceDeltaWithAudioStream(ctx context.Context, session *InferenceSession, audioData string, chunkSize int, chunkDelay time.Duration) (map[string]interface{}, error) {
	logger.Info(fmt.Sprintf("🚀 Starting streaming inference (reuse session): work_id=%s, model=%s, data_size=%d bytes, chunk_size=%d bytes",
		session.WorkID, session.ModelType, len(audioData), chunkSize))

	if chunkSize <= 0 {
		return nil, fmt.Errorf("chunkSize must be positive")
	}

	totalChunks := (len(audioData) + chunkSize - 1) / chunkSize
	logger.Info(fmt.Sprintf("📦 Will send %d chunks", totalChunks))

	for i := 0; i < len(audioData); i += chunkSize {
		end := i + chunkSize
		if end > len(audioData) {
			end = len(audioData)
		}

		chunk := audioData[i:end]
		isLastChunk := end >= len(audioData)
		chunkIndex := i / chunkSize

		if err := c.sendAudioChunk(ctx, session, chunk, chunkIndex, isLastChunk); err != nil {
			return nil, fmt.Errorf("failed to send audio chunk %d: %w", chunkIndex, err)
		}

		if !isLastChunk && chunkDelay > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(chunkDelay):
			}
		}
	}

	// receiveStreamingResponse 会解析 delta 并返回真正的推理结果
	result, err := c.receiveStreamingResponse(ctx, session)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Exit 退出推理任务
func (c *EdgeLLMClient) Exit(ctx context.Context, session *InferenceSession) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Info(fmt.Sprintf("Exiting inference session with work_id: %s", session.WorkID))

	// 构建 exit 请求
	exitReq := &EdgeLLMRequest{
		RequestID: generateRequestID(),
		WorkID:    session.WorkID,
		Action:    "exit",
	}

	// NOTE: The Edge-LLM-Infra unit-manager may not send a response for `exit`.
	// Waiting for it will block requests (health checks and inference) for up to
	// the full timeout. We send the `exit` request best-effort and close the
	// connection immediately.
	if err := session.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		logger.Debug(fmt.Sprintf("Failed to set write deadline: %v", err))
	}
	if err := c.sendRequest(session.Conn, exitReq); err != nil {
		logger.Debug(fmt.Sprintf("Failed to send exit request (ignored): %v", err))
	}

	// 关闭连接
	if err := session.Conn.Close(); err != nil {
		logger.Warn(fmt.Sprintf("Failed to close connection: %v", err))
	}

	logger.Info(fmt.Sprintf("Exit successful for work_id: %s", session.WorkID))
	return nil
}

// RunInference 运行完整的推理流程（setup → inference → exit）
func (c *EdgeLLMClient) RunInference(ctx context.Context, modelType string, inputData string) (map[string]interface{}, error) {
	// Setup
	session, err := c.Setup(ctx, modelType)
	if err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}

	// 确保退出时释放资源
	defer func() {
		if err := c.Exit(ctx, session); err != nil {
			logger.Error(fmt.Sprintf("Failed to exit session: %v", err))
		}
	}()

	// Inference
	result, err := c.Inference(ctx, session, inputData)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// 解析 delta 字段（如果存在）
	// Edge-LLM-Infra 返回的数据格式：{"delta": "{...}", "finish": false, "index": 0}
	// 真正的推理结果在 delta 字段的 JSON 字符串中
	if deltaStr, ok := result["delta"].(string); ok && deltaStr != "" {
		logger.Debug(fmt.Sprintf("Parsing delta field: %s", deltaStr))

		var deltaData map[string]interface{}
		if err := json.Unmarshal([]byte(deltaStr), &deltaData); err != nil {
			logger.Warn(fmt.Sprintf("Failed to parse delta field as JSON: %v", err))
			// 如果解析失败，返回原始结果
			return result, nil
		}

		logger.Info(fmt.Sprintf("✅ Successfully parsed delta field, got %d keys", len(deltaData)))
		// 返回解析后的 delta 数据（真正的推理结果）
		return deltaData, nil
	}

	// 如果没有 delta 字段，返回原始结果
	return result, nil
}

// RunInferenceWithAudioStream 运行完整的推理流程，支持流式传输音频数据
// audioData: base64 编码的音频数据
// chunkSize: 每个数据块的大小（字节）
func (c *EdgeLLMClient) RunInferenceWithAudioStream(ctx context.Context, modelType string, audioData string, chunkSize int, chunkDelay time.Duration) (map[string]interface{}, error) {
	// Setup
	session, err := c.Setup(ctx, modelType)
	if err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}

	// 确保退出时释放资源
	defer func() {
		if err := c.Exit(ctx, session); err != nil {
			logger.Error(fmt.Sprintf("Failed to exit session: %v", err))
		}
	}()

	// 流式发送音频数据
	logger.Info(fmt.Sprintf("📡 Starting streaming audio data: total_size=%d bytes, chunk_size=%d bytes", len(audioData), chunkSize))

	totalChunks := (len(audioData) + chunkSize - 1) / chunkSize
	logger.Info(fmt.Sprintf("📦 Will send %d chunks", totalChunks))

	for i := 0; i < len(audioData); i += chunkSize {
		end := i + chunkSize
		if end > len(audioData) {
			end = len(audioData)
		}

		chunk := audioData[i:end]
		isLastChunk := end >= len(audioData)
		chunkIndex := i / chunkSize

		logger.Debug(fmt.Sprintf("📤 Sending chunk %d/%d: size=%d bytes, finish=%v", chunkIndex+1, totalChunks, len(chunk), isLastChunk))

		// 发送数据块
		if err := c.sendAudioChunk(ctx, session, chunk, chunkIndex, isLastChunk); err != nil {
			return nil, fmt.Errorf("failed to send audio chunk %d: %w", chunkIndex, err)
		}

		logger.Debug(fmt.Sprintf("✅ Chunk %d/%d sent successfully", chunkIndex+1, totalChunks))

		// unit-manager 侧 TCP 未按行切分时，多个 JSON 很容易粘包导致解析失败。
		// 在不改动 unit-manager 的情况下，适当的 chunk 间隔能显著降低概率。
		if !isLastChunk && chunkDelay > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(chunkDelay):
			}
		}
	}

	logger.Info("📡 All audio chunks sent, waiting for final response...")

	// 接收最终响应
	result, err := c.receiveStreamingResponse(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to receive streaming response: %w", err)
	}

	logger.Info("✅ Streaming inference completed successfully")

	return result, nil
}

// sendAudioChunk 发送单个音频数据块
func (c *EdgeLLMClient) sendAudioChunk(ctx context.Context, session *InferenceSession, chunk string, index int, finish bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 更新最后使用时间
	session.LastUsedAt = time.Now()

	// 设置连接超时（只对最后一块设置较长超时）
	timeout := 5 * time.Second
	if finish {
		timeout = c.timeout
	}
	if err := session.Conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("failed to set connection deadline: %w", err)
	}

	// 构建 inference 请求
	inferenceReq := &EdgeLLMRequest{
		RequestID: generateRequestID(),
		WorkID:    session.WorkID,
		Action:    "inference",
		Object:    "llm.utf-8.stream",
		Data: map[string]interface{}{
			"delta":  chunk,
			"index":  index,
			"finish": finish,
		},
	}

	// 发送请求
	if err := c.sendRequest(session.Conn, inferenceReq); err != nil {
		return fmt.Errorf("failed to send chunk: %w", err)
	}

	// Edge-LLM-Infra 不会对中间块发送响应，只在最后一块后发送最终结果
	// 所以我们不等待中间块的响应

	return nil
}

// receiveStreamingResponse 接收流式传输的最终响应
func (c *EdgeLLMClient) receiveStreamingResponse(ctx context.Context, session *InferenceSession) (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Info("⏳ Waiting for final streaming response...")

	// 接收最终响应（并 drain 到 finish=true，避免残留帧影响后续复用会话）
	data, err := c.receiveData(ctx, session.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to receive final response: %w", err)
	}

	// 解析 delta 字段（如果存在）
	if deltaStr, ok := data["delta"].(string); ok && deltaStr != "" {
		logger.Debug(fmt.Sprintf("Parsing delta field: %s", deltaStr))

		var deltaData map[string]interface{}
		if err := json.Unmarshal([]byte(deltaStr), &deltaData); err != nil {
			logger.Warn(fmt.Sprintf("Failed to parse delta field as JSON: %v", err))
			return data, nil
		}

		logger.Info(fmt.Sprintf("✅ Successfully parsed delta field, got %d keys", len(deltaData)))
		return deltaData, nil
	}

	return data, nil
}

// generateRequestID 生成请求 ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000000)
}

// parseDataField 解析 Data 字段
// Edge-LLM-Infra 可能返回以下格式：
// 1. 直接的 JSON 对象: {"key": "value"}
// 2. JSON 字符串: "{\"key\": \"value\"}"
// 此函数会尝试两种方式解析
func parseDataField(resp *EdgeLLMResponse) (map[string]interface{}, error) {
	if resp.Data == nil || len(resp.Data) == 0 {
		return nil, nil
	}

	var data map[string]interface{}

	// 尝试 1: 直接解析为 map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err == nil {
		logger.Debug("Successfully parsed Data field as JSON object")
		return data, nil
	}

	// 尝试 2: 先解析为字符串，再解析字符串内容
	var dataStr string
	if err := json.Unmarshal(resp.Data, &dataStr); err != nil {
		// 记录原始数据以便调试
		logger.Error(fmt.Sprintf("Failed to parse Data field. Raw data: %s", string(resp.Data)))
		return nil, fmt.Errorf("failed to parse data field: cannot unmarshal as object or string: %w", err)
	}

	// 解析字符串内容为 JSON
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		logger.Error(fmt.Sprintf("Failed to parse Data string content. String: %s", dataStr))
		return nil, fmt.Errorf("failed to parse data string content: %w", err)
	}

	logger.Debug("Successfully parsed Data field as JSON string")
	return data, nil
}

// parseErrorField 解析 Error 字段
// 与 parseDataField 类似，支持对象和字符串两种格式
func parseErrorField(resp *EdgeLLMResponse) (map[string]interface{}, error) {
	if resp.Error == nil || len(resp.Error) == 0 {
		return nil, nil
	}

	var errorData map[string]interface{}

	// 尝试 1: 直接解析为 map[string]interface{}
	if err := json.Unmarshal(resp.Error, &errorData); err == nil {
		logger.Debug("Successfully parsed Error field as JSON object")
		return errorData, nil
	}

	// 尝试 2: 先解析为字符串，再解析字符串内容
	var errorStr string
	if err := json.Unmarshal(resp.Error, &errorStr); err != nil {
		// 记录原始数据以便调试
		logger.Error(fmt.Sprintf("Failed to parse Error field. Raw data: %s", string(resp.Error)))
		return nil, fmt.Errorf("failed to parse error field: cannot unmarshal as object or string: %w", err)
	}

	// 解析字符串内容为 JSON
	if err := json.Unmarshal([]byte(errorStr), &errorData); err != nil {
		logger.Error(fmt.Sprintf("Failed to parse Error string content. String: %s", errorStr))
		return nil, fmt.Errorf("failed to parse error string content: %w", err)
	}

	logger.Debug("Successfully parsed Error field as JSON string")
	return errorData, nil
}
