package triton

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"meeting-system/ai-inference-service/runtime"
	"meeting-system/ai-inference-service/runtime/audio"
)

type Model struct {
	spec    runtime.ModelSpec
	client  *Client
	whisper *whisperAssets
	labels  []string
}

func newModel(spec runtime.ModelSpec, client *Client) (*Model, error) {
	model := &Model{
		spec:   spec,
		client: client,
	}
	if spec.Task == runtime.TaskASR {
		assets, err := loadWhisperAssets(spec.ConfigPath, spec.SpecialTokensPath, spec.TokenizerPath)
		if err != nil {
			return nil, err
		}
		model.whisper = assets
	}
	if spec.Task == runtime.TaskEmotion {
		labels, err := loadLabels(spec.LabelsPath)
		if err != nil {
			return nil, err
		}
		model.labels = labels
	}
	return model, nil
}

func (m *Model) Spec() runtime.ModelSpec {
	return m.spec
}

func (m *Model) Infer(ctx context.Context, req runtime.InferenceRequest) (*runtime.InferenceResult, error) {
	switch m.spec.Task {
	case runtime.TaskASR:
		return m.inferWhisper(ctx, req)
	case runtime.TaskEmotion:
		return m.inferEmotion(ctx, req)
	case runtime.TaskSynthesis:
		return m.inferSynthesis(ctx, req)
	default:
		return nil, runtime.ErrInferenceNotImplemented
	}
}

func (m *Model) Close() error {
	return nil
}

func (m *Model) inferWhisper(ctx context.Context, req runtime.InferenceRequest) (*runtime.InferenceResult, error) {
	if len(req.AudioFloat32) == 0 {
		return nil, fmt.Errorf("audio data is required for ASR")
	}
	if m.client == nil {
		return nil, fmt.Errorf("triton client not initialized")
	}
	inputName := strings.TrimSpace(m.spec.InputName)
	if inputName == "" {
		inputName = "mel"
	}
	outputNames := m.spec.OutputNames
	if len(outputNames) == 0 {
		outputNames = []string{"encoder_output"}
	}
	encoderModel := strings.TrimSpace(m.spec.Name)
	if encoderModel == "" {
		return nil, fmt.Errorf("encoder model name is required")
	}
	decoderModel := strings.TrimSpace(m.spec.DecoderPath)
	if decoderModel == "" {
		return nil, fmt.Errorf("decoder model name is required for whisper")
	}

	melCfg := audio.DefaultMelConfig(req.SampleRate)
	if m.whisper != nil {
		if m.whisper.config.NMels > 0 {
			melCfg.NMels = m.whisper.config.NMels
		}
	}
	targetFrames := 0
	if m.whisper != nil && m.whisper.config.MelLength > 0 {
		targetFrames = m.whisper.config.MelLength
	}

	mel, frames, err := audio.ComputeLogMelSpectrogram(req.AudioFloat32, melCfg, targetFrames)
	if err != nil {
		return nil, err
	}

	inputs := []inferenceInput{
		{
			Name:     inputName,
			Datatype: "FP32",
			Shape:    []int64{1, int64(melCfg.NMels), int64(frames)},
			Data:     mel,
		},
	}
	outputs, err := m.client.Infer(ctx, encoderModel, inputs, outputNames)
	if err != nil {
		return nil, err
	}
	encoderTensor, err := selectOutput(outputs, outputNames)
	if err != nil {
		return nil, err
	}
	encoderData := toFloat32Slice(encoderTensor.Data)
	if len(encoderData) == 0 {
		return nil, fmt.Errorf("encoder output empty")
	}

	language := ""
	if req.Params != nil {
		language = req.Params["language"]
	}

	tokens, text, err := m.decodeWhisper(ctx, decoderModel, encoderData, encoderTensor.Shape, language)
	if err != nil {
		return nil, err
	}

	result := &runtime.InferenceResult{
		Outputs: map[string]interface{}{
			"text":       text,
			"tokens":     tokens,
			"language":   language,
			"confidence": 0.0,
		},
	}
	return result, nil
}

func (m *Model) decodeWhisper(ctx context.Context, decoderModel string, encoderData []float32, encoderShape []int64, language string) ([]int64, string, error) {
	if m.whisper == nil {
		m.whisper = &whisperAssets{config: whisperConfig{NMels: 80, MelLength: 3000, NAudioCtx: 1500}, maxTokens: 128}
	}
	initial := m.whisper.initialTokens(language)
	if len(initial) == 0 {
		initial = []int64{50258}
	}
	maxSteps := m.whisper.maxTokens
	if maxSteps <= 0 {
		maxSteps = 128
	}

	tokens := append([]int64{}, initial...)
	decoderInputs := m.spec.DecoderInputNames
	if len(decoderInputs) == 0 {
		decoderInputs = []string{"tokens", "encoder_output"}
	}
	decoderOutputs := m.spec.DecoderOutputNames
	if len(decoderOutputs) == 0 {
		decoderOutputs = []string{"logits"}
	}

	for step := 0; step < maxSteps; step++ {
		logits, shape, err := m.runDecoder(ctx, decoderModel, decoderInputs, decoderOutputs, tokens, encoderData, encoderShape)
		if err != nil {
			return nil, "", err
		}
		if len(shape) < 3 {
			return nil, "", fmt.Errorf("unexpected decoder output shape: %v", shape)
		}
		seqLen := int(shape[len(shape)-2])
		vocab := int(shape[len(shape)-1])
		if seqLen == 0 || vocab == 0 {
			return nil, "", fmt.Errorf("invalid decoder output shape: %v", shape)
		}
		start := (seqLen - 1) * vocab
		if start+vocab > len(logits) {
			return nil, "", fmt.Errorf("decoder output size mismatch")
		}
		bestToken := argmax(logits[start : start+vocab])
		if bestToken == m.whisper.special.Eot && bestToken != 0 {
			break
		}
		tokens = append(tokens, int64(bestToken))
	}

	text := m.whisper.tokensToText(tokens)
	return tokens, text, nil
}

func (m *Model) runDecoder(ctx context.Context, decoderModel string, inputNames, outputNames []string, tokens []int64, encoderData []float32, encoderShape []int64) ([]float32, []int64, error) {
	inputs := []inferenceInput{
		{
			Name:     inputNames[0],
			Datatype: "INT64",
			Shape:    []int64{1, int64(len(tokens))},
			Data:     tokens,
		},
		{
			Name:     inputNames[1],
			Datatype: "FP32",
			Shape:    encoderShape,
			Data:     encoderData,
		},
	}
	outputs, err := m.client.Infer(ctx, decoderModel, inputs, outputNames)
	if err != nil {
		return nil, nil, err
	}
	decoderTensor, err := selectOutput(outputs, outputNames)
	if err != nil {
		return nil, nil, err
	}
	return toFloat32Slice(decoderTensor.Data), decoderTensor.Shape, nil
}

func (m *Model) inferEmotion(ctx context.Context, req runtime.InferenceRequest) (*runtime.InferenceResult, error) {
	inputType := strings.ToLower(strings.TrimSpace(m.spec.InputType))
	if inputType == "text" {
		return nil, fmt.Errorf("text emotion inference is not supported by triton runtime")
	}
	if len(req.AudioFloat32) == 0 {
		return nil, fmt.Errorf("audio data is required for emotion detection")
	}
	inputName := strings.TrimSpace(m.spec.InputName)
	if inputName == "" {
		inputName = "audio_input"
	}
	outputNames := m.spec.OutputNames
	if len(outputNames) == 0 {
		outputNames = []string{"logits"}
	}
	modelName := strings.TrimSpace(m.spec.Name)
	if modelName == "" {
		return nil, fmt.Errorf("emotion model name is required")
	}

	inputs := []inferenceInput{
		{
			Name:     inputName,
			Datatype: "FP32",
			Shape:    []int64{1, int64(len(req.AudioFloat32))},
			Data:     req.AudioFloat32,
		},
	}

	outputs, err := m.client.Infer(ctx, modelName, inputs, outputNames)
	if err != nil {
		return nil, err
	}
	emotionTensor, err := selectOutput(outputs, outputNames)
	if err != nil {
		return nil, err
	}
	logits := toFloat32Slice(emotionTensor.Data)
	if len(logits) == 0 {
		return nil, fmt.Errorf("emotion output empty")
	}

	probs := softmax(logits)
	bestIdx := argmaxFloat64(probs)
	label := fmt.Sprintf("emotion_%d", bestIdx)
	if bestIdx >= 0 && bestIdx < len(m.labels) {
		label = m.labels[bestIdx]
	}
	confidence := probs[bestIdx]
	all := make(map[string]float64)
	for i, p := range probs {
		name := fmt.Sprintf("emotion_%d", i)
		if i < len(m.labels) {
			name = m.labels[i]
		}
		all[name] = p
	}

	return &runtime.InferenceResult{
		Outputs: map[string]interface{}{
			"emotion":    label,
			"confidence": confidence,
			"emotions":   all,
		},
	}, nil
}

func (m *Model) inferSynthesis(ctx context.Context, req runtime.InferenceRequest) (*runtime.InferenceResult, error) {
	if len(req.AudioFloat32) == 0 {
		return nil, fmt.Errorf("audio data is required for synthesis detection")
	}
	inputName := strings.TrimSpace(m.spec.InputName)
	if inputName == "" {
		inputName = "audio_input"
	}
	outputNames := m.spec.OutputNames
	if len(outputNames) == 0 {
		outputNames = []string{"synthesis_output"}
	}
	modelName := strings.TrimSpace(m.spec.Name)
	if modelName == "" {
		return nil, fmt.Errorf("synthesis model name is required")
	}

	inputType := strings.ToLower(strings.TrimSpace(m.spec.InputType))
	inputs := make([]inferenceInput, 0, 1)
	if inputType == "waveform" {
		inputs = append(inputs, inferenceInput{
			Name:     inputName,
			Datatype: "FP32",
			Shape:    []int64{1, int64(len(req.AudioFloat32))},
			Data:     req.AudioFloat32,
		})
	} else {
		melCfg := audio.DefaultMelConfig(req.SampleRate)
		mel, frames, err := audio.ComputeLogMelSpectrogram(req.AudioFloat32, melCfg, 0)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, inferenceInput{
			Name:     inputName,
			Datatype: "FP32",
			Shape:    []int64{1, int64(melCfg.NMels), int64(frames)},
			Data:     mel,
		})
	}

	outputs, err := m.client.Infer(ctx, modelName, inputs, outputNames)
	if err != nil {
		return nil, err
	}
	synthTensor, err := selectOutput(outputs, outputNames)
	if err != nil {
		return nil, err
	}
	values := toFloat32Slice(synthTensor.Data)
	if len(values) == 0 {
		return nil, fmt.Errorf("synthesis output empty")
	}

	prob := synthesisProbability(values)
	isSynthetic := prob > 0.5

	return &runtime.InferenceResult{
		Outputs: map[string]interface{}{
			"is_synthetic":        isSynthetic,
			"probability_synthetic": prob,
			"confidence":          prob,
		},
	}, nil
}

func loadLabels(path string) ([]string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	var raw interface{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	switch val := raw.(type) {
	case []interface{}:
		labels := make([]string, 0, len(val))
		for _, item := range val {
			if label, ok := item.(string); ok {
				labels = append(labels, label)
			}
		}
		return labels, nil
	case map[string]interface{}:
		keys := make([]int, 0, len(val))
		labels := make(map[int]string, len(val))
		for key, item := range val {
			idx, err := strconv.Atoi(key)
			if err != nil {
				continue
			}
			if label, ok := item.(string); ok {
				labels[idx] = label
				keys = append(keys, idx)
			}
		}
		sort.Ints(keys)
		ordered := make([]string, 0, len(keys))
		for _, idx := range keys {
			ordered = append(ordered, labels[idx])
		}
		return ordered, nil
	default:
		return nil, nil
	}
}

func selectOutput(outputs map[string]tensorResponse, outputNames []string) (tensorResponse, error) {
	for _, name := range outputNames {
		if out, ok := outputs[name]; ok {
			return out, nil
		}
	}
	for _, out := range outputs {
		return out, nil
	}
	return tensorResponse{}, fmt.Errorf("no outputs returned from triton")
}

func toFloat32Slice(data []interface{}) []float32 {
	out := make([]float32, 0, len(data))
	for _, val := range data {
		switch v := val.(type) {
		case float64:
			out = append(out, float32(v))
		case float32:
			out = append(out, v)
		case int:
			out = append(out, float32(v))
		case int64:
			out = append(out, float32(v))
		}
	}
	return out
}

func argmax(values []float32) int {
	bestIdx := 0
	bestVal := float32(math.Inf(-1))
	for i, v := range values {
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

func softmax(values []float32) []float64 {
	if len(values) == 0 {
		return nil
	}
	maxVal := float32(math.Inf(-1))
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	expSum := 0.0
	out := make([]float64, len(values))
	for i, v := range values {
		expVal := math.Exp(float64(v - maxVal))
		out[i] = expVal
		expSum += expVal
	}
	if expSum == 0 {
		return out
	}
	for i := range out {
		out[i] /= expSum
	}
	return out
}

func argmaxFloat64(values []float64) int {
	bestIdx := 0
	bestVal := math.Inf(-1)
	for i, v := range values {
		if v > bestVal {
			bestVal = v
			bestIdx = i
		}
	}
	return bestIdx
}

func synthesisProbability(values []float32) float64 {
	if len(values) == 1 {
		return sigmoid(float64(values[0]))
	}
	probs := softmax(values)
	if len(probs) > 1 {
		return probs[1]
	}
	return probs[0]
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}
