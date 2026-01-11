//go:build !zmq

package zmq

import (
	"context"
	"errors"

	"meeting-system/shared/config"
)

// 在未启用 zmq 构建标签或缺少 libzmq 时的占位实现。
var errZMQUnavailable = errors.New("zmq not built (missing libzmq or build tag)")

type ZMQClient struct{}

func NewZMQClient(_ config.ZMQConfig) (*ZMQClient, error) { return nil, errZMQUnavailable }
func (c *ZMQClient) Close()                               {}
func (c *ZMQClient) RegisterUnit(ctx context.Context) (*AIResponse, error) {
	return nil, errZMQUnavailable
}
func (c *ZMQClient) ReleaseUnit(ctx context.Context) error { return errZMQUnavailable }
func (c *ZMQClient) SendRequest(ctx context.Context, req *AIRequest) (*AIResponse, error) {
	return nil, errZMQUnavailable
}
func (c *ZMQClient) HealthCheck(ctx context.Context) error { return errZMQUnavailable }

type ZMQManager struct {
	client *ZMQClient
	config config.ZMQConfig
}

var globalZMQManager *ZMQManager

func InitZMQ(_ config.ZMQConfig) error { return errZMQUnavailable }
func GetZMQClient() *ZMQClient         { return nil }
func CloseZMQ()                        {}

// Minimal request/response structs to satisfy references.
type AIRequest struct {
	RequestID string      `json:"request_id"`
	WorkID    string      `json:"work_id"`
	Action    string      `json:"action"`
	Object    string      `json:"object"`
	Data      interface{} `json:"data"`
}

type AIResponse struct {
	RequestID string      `json:"request_id"`
	WorkID    string      `json:"work_id"`
	Object    string      `json:"object"`
	Data      interface{} `json:"data"`
	Error     *string     `json:"error"`
}
