package onnx

import (
	"context"
	"fmt"
	"sync"

	"meeting-system/ai-inference-service/runtime"
)

// Engine keeps a placeholder ONNX runtime implementation for optional use.
type Engine struct {
	cfg    runtime.EngineConfig
	mu     sync.RWMutex
	models map[string]*Model
}

// NewEngine creates a new placeholder engine.
func NewEngine(cfg runtime.EngineConfig) *Engine {
	return &Engine{
		cfg:    cfg,
		models: make(map[string]*Model),
	}
}

// LoadModel registers a model spec without loading ONNX Runtime.
func (e *Engine) LoadModel(ctx context.Context, spec runtime.ModelSpec) (runtime.Model, error) {
	model := &Model{spec: spec}
	e.mu.Lock()
	e.models[spec.Name] = model
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

// Model is a stub model wrapper.
type Model struct {
	spec runtime.ModelSpec
}

// Spec returns the model specification.
func (m *Model) Spec() runtime.ModelSpec {
	return m.spec
}

// Infer returns a placeholder error until ONNX Runtime bindings are enabled.
func (m *Model) Infer(ctx context.Context, req runtime.InferenceRequest) (*runtime.InferenceResult, error) {
	return nil, runtime.ErrInferenceNotImplemented
}

// Close releases model resources.
func (m *Model) Close() error {
	return nil
}

// ErrNotEnabled indicates ONNX runtime usage is disabled in this build.
var ErrNotEnabled = fmt.Errorf("onnx runtime is not enabled in this build")
