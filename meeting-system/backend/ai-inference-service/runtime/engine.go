package runtime

import "context"

// Engine loads models and executes inference.
type Engine interface {
	LoadModel(ctx context.Context, spec ModelSpec) (Model, error)
	Close() error
}

// Model represents a loaded model instance.
type Model interface {
	Spec() ModelSpec
	Infer(ctx context.Context, req InferenceRequest) (*InferenceResult, error)
	Close() error
}
