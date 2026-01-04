package runtime

import "errors"

// TaskType identifies a model task within the AI runtime.
type TaskType string

const (
	TaskASR       TaskType = "asr"
	TaskEmotion   TaskType = "emotion"
	TaskSynthesis TaskType = "synthesis"
)

var (
	ErrModelNotLoaded          = errors.New("model not loaded")
	ErrInferenceNotImplemented = errors.New("inference not implemented")
)

// EngineConfig defines runtime-level settings for the inference engine.
type EngineConfig struct {
	Backend         string
	Providers       []string
	DeviceID        int
	LibraryPath     string
	IntraOpThreads  int
	InterOpThreads  int
	EnableFP16      bool
	EnableTensorRT  bool
	TritonEndpoint  string
	TritonTimeoutMs int
}

// ModelSpec defines how to load a model for a given task.
type ModelSpec struct {
	Name               string
	Task               TaskType
	Path               string
	InputName          string
	OutputNames        []string
	SampleRate         int
	Channels           int
	InputType          string
	DecoderPath        string
	DecoderInputNames  []string
	DecoderOutputNames []string
	TokenizerPath      string
	SpecialTokensPath  string
	ConfigPath         string
	LabelsPath         string
}

// InferenceRequest represents the input payload for a model.
type InferenceRequest struct {
	Task         TaskType
	AudioPCM     []byte
	AudioFloat32 []float32
	SampleRate   int
	Channels     int
	Text         string
	Params       map[string]string
}

// InferenceResult wraps model outputs for downstream parsing.
type InferenceResult struct {
	Outputs map[string]interface{}
	Raw     interface{}
}
