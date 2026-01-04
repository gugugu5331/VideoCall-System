package triton

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type whisperAssets struct {
	config       whisperConfig
	special      whisperSpecialTokens
	vocab        map[int]string
	maxTokens    int
	languageHint string
}

type whisperConfig struct {
	NMels      int `json:"n_mels"`
	MelLength  int `json:"mel_length"`
	NAudioCtx  int `json:"n_audio_ctx"`
	NAudioState int `json:"n_audio_state"`
}

type whisperSpecialTokens struct {
	Sot           int            `json:"sot"`
	Eot           int            `json:"eot"`
	SotPrev       int            `json:"sot_prev"`
	NoTimestamps  int            `json:"no_timestamps"`
	LanguageTokens map[string]int `json:"language_tokens"`
	TaskTokens     map[string]int `json:"task_tokens"`
}

func loadWhisperAssets(specPath, specialPath, vocabPath string) (*whisperAssets, error) {
	assets := &whisperAssets{
		config: whisperConfig{
			NMels:     80,
			MelLength: 3000,
			NAudioCtx: 1500,
		},
		maxTokens: 128,
	}

	if specPath != "" {
		if err := readJSONFile(specPath, &assets.config); err != nil {
			return nil, fmt.Errorf("load whisper config: %w", err)
		}
	}
	if specialPath != "" {
		if err := readJSONFile(specialPath, &assets.special); err != nil {
			return nil, fmt.Errorf("load whisper special tokens: %w", err)
		}
	}
	if vocabPath != "" {
		vocab, err := loadWhisperVocab(vocabPath)
		if err != nil {
			return nil, err
		}
		assets.vocab = vocab
	}

	return assets, nil
}

func readJSONFile(path string, target interface{}) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func loadWhisperVocab(path string) (map[int]string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err == nil {
		vocab := make(map[int]string, len(raw))
		for key, val := range raw {
			idx, err := strconv.Atoi(key)
			if err != nil {
				continue
			}
			vocab[idx] = val
		}
		return vocab, nil
	}
	var alt map[string]interface{}
	if err := json.Unmarshal(data, &alt); err != nil {
		return nil, err
	}
	vocab := make(map[int]string, len(alt))
	for key, val := range alt {
		idx, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		if token, ok := val.(string); ok {
			vocab[idx] = token
		}
	}
	return vocab, nil
}

func (w *whisperAssets) initialTokens(language string) []int64 {
	if w == nil {
		return nil
	}
	tokens := make([]int64, 0, 4)
	if w.special.Sot != 0 {
		tokens = append(tokens, int64(w.special.Sot))
	}
	langToken := w.languageToken(language)
	if langToken != 0 {
		tokens = append(tokens, int64(langToken))
	}
	if taskToken, ok := w.special.TaskTokens["transcribe"]; ok {
		tokens = append(tokens, int64(taskToken))
	}
	if w.special.NoTimestamps != 0 {
		tokens = append(tokens, int64(w.special.NoTimestamps))
	}
	return tokens
}

func (w *whisperAssets) languageToken(lang string) int {
	if w == nil {
		return 0
	}
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" && w.languageHint != "" {
		lang = w.languageHint
	}
	if lang != "" {
		if token, ok := w.special.LanguageTokens[lang]; ok {
			return token
		}
	}
	if token, ok := w.special.LanguageTokens["en"]; ok {
		return token
	}
	for _, token := range w.special.LanguageTokens {
		return token
	}
	return 0
}

func (w *whisperAssets) tokensToText(tokens []int64) string {
	if w == nil || len(tokens) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for _, token := range tokens {
		if w.isSpecialToken(int(token)) {
			continue
		}
		if w.vocab == nil {
			continue
		}
		if text, ok := w.vocab[int(token)]; ok {
			if strings.HasPrefix(text, "<|") {
				continue
			}
			builder.WriteString(text)
		}
	}
	return strings.TrimSpace(builder.String())
}

func (w *whisperAssets) isSpecialToken(token int) bool {
	if w == nil {
		return false
	}
	if token == w.special.Sot || token == w.special.Eot || token == w.special.SotPrev || token == w.special.NoTimestamps {
		return true
	}
	for _, v := range w.special.LanguageTokens {
		if v == token {
			return true
		}
	}
	for _, v := range w.special.TaskTokens {
		if v == token {
			return true
		}
	}
	return false
}
