package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"meeting-system/shared/config"
)

// RemoteAIClient 调用远端 HTTP AI 服务的客户端
type RemoteAIClient struct {
	baseURL string
	client  *http.Client
	token   string
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func NewRemoteAIClient(cfg config.AIHTTPConfig) *RemoteAIClient {
	base := strings.TrimRight(cfg.Endpoint, "/")
	if base == "" {
		return nil
	}
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
	}

	return &RemoteAIClient{
		baseURL: base,
		token:   cfg.Token,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

func (c *RemoteAIClient) SpeechRecognition(ctx context.Context, req *ASRRequest) (*ASRResponse, error) {
	var resp ASRResponse
	if err := c.doPOST(ctx, "/api/v1/ai/asr", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RemoteAIClient) EmotionDetection(ctx context.Context, req *EmotionRequest) (*EmotionResponse, error) {
	var resp EmotionResponse
	if err := c.doPOST(ctx, "/api/v1/ai/emotion", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RemoteAIClient) SynthesisDetection(ctx context.Context, req *SynthesisDetectionRequest) (*SynthesisDetectionResponse, error) {
	var resp SynthesisDetectionResponse
	if err := c.doPOST(ctx, "/api/v1/ai/synthesis", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *RemoteAIClient) Health(ctx context.Context) error {
	if err := c.doGET(ctx, "/api/v1/ai/health", nil); err == nil {
		return nil
	}
	// 兼容部分服务仅暴露 /health
	return c.doGET(ctx, "/health", nil)
}

func (c *RemoteAIClient) doPOST(ctx context.Context, path string, payload interface{}, out interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.doRequest(req, out)
}

func (c *RemoteAIClient) doGET(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request failed: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.doRequest(req, out)
}

func (c *RemoteAIClient) doRequest(req *http.Request, out interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	// 以 code==0 或 HTTP 200/201 視為成功
	if resp.StatusCode >= 400 || (apiResp.Code != 0 && apiResp.Code != http.StatusOK && apiResp.Code != http.StatusCreated) {
		msg := apiResp.Message
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("remote ai error: %s (code=%d, http=%d)", msg, apiResp.Code, resp.StatusCode)
	}

	if out == nil || len(apiResp.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(apiResp.Data, out); err != nil {
		return fmt.Errorf("decode data failed: %w", err)
	}
	return nil
}
