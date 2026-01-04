package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"meeting-system/ai-inference-service/runtime"
	"meeting-system/ai-inference-service/runtime/onnx"
	"meeting-system/ai-inference-service/runtime/triton"
	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// ModelManager manages loaded AI models.
type ModelManager struct {
	engine runtime.Engine
	models map[runtime.TaskType]runtime.Model
	specs  map[runtime.TaskType]runtime.ModelSpec
}

// NewModelManager initializes the runtime engine and loads configured models.
func NewModelManager(ctx context.Context, cfg *config.Config) (*ModelManager, error) {
	engineCfg := runtime.EngineConfig{}
	backend := "triton"
	if cfg != nil {
		if cfg.AI.Runtime.Backend != "" {
			backend = cfg.AI.Runtime.Backend
		}
		engineCfg = runtime.EngineConfig{
			Backend:        backend,
			Providers:      cfg.AI.Runtime.Providers,
			DeviceID:       cfg.AI.Runtime.DeviceID,
			LibraryPath:    cfg.AI.Runtime.LibraryPath,
			IntraOpThreads: cfg.AI.Runtime.IntraOpThreads,
			InterOpThreads: cfg.AI.Runtime.InterOpThreads,
			EnableFP16:     cfg.AI.Runtime.EnableFP16,
			EnableTensorRT: cfg.AI.Runtime.EnableTensorRT,
			TritonEndpoint: cfg.AI.Runtime.Triton.Endpoint,
			TritonTimeoutMs: cfg.AI.Runtime.Triton.TimeoutMs,
		}
	}

	var engine runtime.Engine
	switch strings.ToLower(strings.TrimSpace(backend)) {
	case "triton":
		engine = triton.NewEngine(engineCfg)
	default:
		engine = onnx.NewEngine(engineCfg)
	}
	manager := &ModelManager{
		engine: engine,
		models: make(map[runtime.TaskType]runtime.Model),
		specs:  make(map[runtime.TaskType]runtime.ModelSpec),
	}

	var loadErrors []error
	allowEmptyPath := strings.EqualFold(strings.TrimSpace(backend), "triton")
	load := func(task runtime.TaskType, spec runtime.ModelSpec) {
		if spec.Path == "" && !allowEmptyPath {
			manager.specs[task] = spec
			return
		}
		if spec.Name == "" && spec.Path == "" {
			manager.specs[task] = spec
			return
		}
		model, err := engine.LoadModel(ctx, spec)
		if err != nil {
			loadErrors = append(loadErrors, err)
			logger.Warn("Failed to load model",
				logger.String("task", string(task)),
				logger.String("path", spec.Path),
				logger.Err(err))
			manager.specs[task] = spec
			return
		}
		manager.models[task] = model
		manager.specs[task] = spec
		logger.Info("Model loaded",
			logger.String("task", string(task)),
			logger.String("name", spec.Name),
			logger.String("path", spec.Path))
	}

	if cfg != nil {
		load(runtime.TaskASR, runtime.ModelSpec{
			Name:               cfg.AI.Models.ASR.ModelName,
			Task:               runtime.TaskASR,
			Path:               cfg.AI.Models.ASR.ModelPath,
			InputName:          cfg.AI.Models.ASR.InputName,
			OutputNames:        cfg.AI.Models.ASR.OutputNames,
			InputType:          cfg.AI.Models.ASR.InputType,
			DecoderPath:        cfg.AI.Models.ASR.DecoderPath,
			DecoderInputNames:  cfg.AI.Models.ASR.DecoderInputNames,
			DecoderOutputNames: cfg.AI.Models.ASR.DecoderOutputNames,
			TokenizerPath:      cfg.AI.Models.ASR.TokenizerPath,
			SpecialTokensPath:  cfg.AI.Models.ASR.SpecialTokensPath,
			ConfigPath:         cfg.AI.Models.ASR.ConfigPath,
			LabelsPath:         cfg.AI.Models.ASR.LabelsPath,
			SampleRate:         cfg.AI.Models.ASR.SampleRate,
			Channels:           cfg.AI.Models.ASR.Channels,
		})
		load(runtime.TaskEmotion, runtime.ModelSpec{
			Name:               cfg.AI.Models.Emotion.ModelName,
			Task:               runtime.TaskEmotion,
			Path:               cfg.AI.Models.Emotion.ModelPath,
			InputName:          cfg.AI.Models.Emotion.InputName,
			OutputNames:        cfg.AI.Models.Emotion.OutputNames,
			InputType:          cfg.AI.Models.Emotion.InputType,
			DecoderPath:        cfg.AI.Models.Emotion.DecoderPath,
			DecoderInputNames:  cfg.AI.Models.Emotion.DecoderInputNames,
			DecoderOutputNames: cfg.AI.Models.Emotion.DecoderOutputNames,
			TokenizerPath:      cfg.AI.Models.Emotion.TokenizerPath,
			SpecialTokensPath:  cfg.AI.Models.Emotion.SpecialTokensPath,
			ConfigPath:         cfg.AI.Models.Emotion.ConfigPath,
			LabelsPath:         cfg.AI.Models.Emotion.LabelsPath,
			SampleRate:         cfg.AI.Models.Emotion.SampleRate,
			Channels:           cfg.AI.Models.Emotion.Channels,
		})
		load(runtime.TaskSynthesis, runtime.ModelSpec{
			Name:               cfg.AI.Models.Synthesis.ModelName,
			Task:               runtime.TaskSynthesis,
			Path:               cfg.AI.Models.Synthesis.ModelPath,
			InputName:          cfg.AI.Models.Synthesis.InputName,
			OutputNames:        cfg.AI.Models.Synthesis.OutputNames,
			InputType:          cfg.AI.Models.Synthesis.InputType,
			DecoderPath:        cfg.AI.Models.Synthesis.DecoderPath,
			DecoderInputNames:  cfg.AI.Models.Synthesis.DecoderInputNames,
			DecoderOutputNames: cfg.AI.Models.Synthesis.DecoderOutputNames,
			TokenizerPath:      cfg.AI.Models.Synthesis.TokenizerPath,
			SpecialTokensPath:  cfg.AI.Models.Synthesis.SpecialTokensPath,
			ConfigPath:         cfg.AI.Models.Synthesis.ConfigPath,
			LabelsPath:         cfg.AI.Models.Synthesis.LabelsPath,
			SampleRate:         cfg.AI.Models.Synthesis.SampleRate,
			Channels:           cfg.AI.Models.Synthesis.Channels,
		})
	}

	if len(loadErrors) > 0 {
		return manager, fmt.Errorf("model load errors: %w", errors.Join(loadErrors...))
	}

	return manager, nil
}

// GetModel returns a loaded model for the given task.
func (m *ModelManager) GetModel(task runtime.TaskType) (runtime.Model, runtime.ModelSpec, bool) {
	if m == nil {
		return nil, runtime.ModelSpec{}, false
	}
	model, ok := m.models[task]
	spec, _ := m.specs[task]
	return model, spec, ok
}

// Infer executes inference for a task using the loaded model.
func (m *ModelManager) Infer(ctx context.Context, task runtime.TaskType, req runtime.InferenceRequest) (*runtime.InferenceResult, runtime.ModelSpec, error) {
	model, spec, ok := m.GetModel(task)
	if !ok || model == nil {
		return nil, spec, runtime.ErrModelNotLoaded
	}

	result, err := model.Infer(ctx, req)
	return result, spec, err
}

// Close shuts down models and engine.
func (m *ModelManager) Close() {
	if m == nil {
		return
	}
	for _, model := range m.models {
		_ = model.Close()
	}
	_ = m.engine.Close()
}
