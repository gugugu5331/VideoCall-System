package triton

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"meeting-system/ai-inference-service/runtime"
)

// Engine executes inference via a remote Triton server.
type Engine struct {
	cfg    runtime.EngineConfig
	mu     sync.RWMutex
	client *Client
	models map[string]*Model
}

// NewEngine creates a Triton engine.
func NewEngine(cfg runtime.EngineConfig) *Engine {
	return &Engine{
		cfg:    cfg,
		models: make(map[string]*Model),
	}
}

// LoadModel registers a model spec and validates Triton readiness.
func (e *Engine) LoadModel(ctx context.Context, spec runtime.ModelSpec) (runtime.Model, error) {
	if err := e.ensureClient(); err != nil {
		return nil, err
	}

	name := resolveModelName(spec)
	if name == "" {
		return nil, fmt.Errorf("model name is required for task %s", spec.Task)
	}

	if err := e.client.ModelReady(ctx, name); err != nil {
		return nil, err
	}
	if spec.Task == runtime.TaskASR && spec.DecoderPath != "" {
		decoderName := strings.TrimSpace(spec.DecoderPath)
		if decoderName != "" {
			if err := e.client.ModelReady(ctx, decoderName); err != nil {
				return nil, err
			}
		}
	}

	model, err := newModel(spec, e.client)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	e.models[name] = model
	e.mu.Unlock()

	return model, nil
}

// Close releases runtime resources.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, model := range e.models {
		_ = model.Close()
	}
	e.models = make(map[string]*Model)
	return nil
}

func (e *Engine) ensureClient() error {
	if e.client != nil {
		return nil
	}
	endpoint := strings.TrimSpace(e.cfg.TritonEndpoint)
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}
	timeout := time.Duration(e.cfg.TritonTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	client, err := NewClient(endpoint, timeout)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

func resolveModelName(spec runtime.ModelSpec) string {
	name := strings.TrimSpace(spec.Name)
	if name != "" {
		return name
	}
	if spec.Path != "" {
		return strings.TrimSuffix(filepath.Base(spec.Path), filepath.Ext(spec.Path))
	}
	return ""
}
