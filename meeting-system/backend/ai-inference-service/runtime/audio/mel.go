package audio

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

// MelConfig controls log-mel spectrogram extraction.
type MelConfig struct {
	SampleRate int
	NMels      int
	NFFT       int
	HopLength  int
	WinLength  int
	FMin       float64
	FMax       float64
}

// DefaultMelConfig returns a standard 80-bin config for Whisper-style models.
func DefaultMelConfig(sampleRate int) MelConfig {
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	return MelConfig{
		SampleRate: sampleRate,
		NMels:      80,
		NFFT:       400,
		HopLength:  160,
		WinLength:  400,
		FMin:       0,
		FMax:       float64(sampleRate) / 2,
	}
}

var melFilterCache sync.Map

// ComputeLogMelSpectrogram converts audio samples into log-mel features.
func ComputeLogMelSpectrogram(samples []float32, cfg MelConfig, targetFrames int) ([]float32, int, error) {
	if len(samples) == 0 {
		return nil, 0, errors.New("empty audio for mel spectrogram")
	}
	if cfg.SampleRate <= 0 {
		return nil, 0, fmt.Errorf("invalid sample rate: %d", cfg.SampleRate)
	}
	if cfg.NMels <= 0 {
		cfg.NMels = 80
	}
	if cfg.NFFT <= 0 {
		cfg.NFFT = 400
	}
	if cfg.WinLength <= 0 {
		cfg.WinLength = cfg.NFFT
	}
	if cfg.HopLength <= 0 {
		cfg.HopLength = cfg.WinLength / 4
	}
	if cfg.FMax <= 0 {
		cfg.FMax = float64(cfg.SampleRate) / 2
	}

	floatSamples := make([]float64, len(samples))
	for i, v := range samples {
		floatSamples[i] = float64(v)
	}

	frames := frameCount(len(floatSamples), cfg.WinLength, cfg.HopLength)
	paddedLen := cfg.WinLength + (frames-1)*cfg.HopLength
	if paddedLen < 0 {
		paddedLen = 0
	}
	padded := make([]float64, paddedLen)
	copy(padded, floatSamples)

	melFilters := getMelFilterBank(cfg)
	nFreqs := cfg.NFFT/2 + 1
	if len(melFilters) == 0 || len(melFilters[0]) != nFreqs {
		return nil, 0, fmt.Errorf("mel filter bank dimension mismatch")
	}

	mel := make([]float32, cfg.NMels*frames)
	windowVals := window.Hann(cfg.WinLength)
	fftBuffer := make([]float64, cfg.NFFT)

	for frameIdx := 0; frameIdx < frames; frameIdx++ {
		start := frameIdx * cfg.HopLength
		for i := 0; i < cfg.NFFT; i++ {
			fftBuffer[i] = 0
		}
		for i := 0; i < cfg.WinLength; i++ {
			fftBuffer[i] = padded[start+i] * windowVals[i]
		}
		fftResult := fft.FFTReal(fftBuffer)

		power := make([]float64, nFreqs)
		for i := 0; i < nFreqs; i++ {
			re := real(fftResult[i])
			im := imag(fftResult[i])
			power[i] = re*re + im*im
		}

		for m := 0; m < cfg.NMels; m++ {
			energy := 0.0
			filter := melFilters[m]
			for k := 0; k < nFreqs; k++ {
				energy += filter[k] * power[k]
			}
			if energy < 1e-10 {
				energy = 1e-10
			}
			logEnergy := math.Log10(energy)
			mel[m*frames+frameIdx] = float32(logEnergy)
		}
	}

	mel, frames = padOrTrimMel(mel, cfg.NMels, frames, targetFrames)
	return mel, frames, nil
}

func frameCount(samples, winLength, hop int) int {
	if samples <= winLength {
		return 1
	}
	return 1 + int(math.Ceil(float64(samples-winLength)/float64(hop)))
}

func padOrTrimMel(mel []float32, nMels, frames, targetFrames int) ([]float32, int) {
	if targetFrames <= 0 || frames == targetFrames {
		return mel, frames
	}
	out := make([]float32, nMels*targetFrames)
	maxFrames := frames
	if maxFrames > targetFrames {
		maxFrames = targetFrames
	}
	for m := 0; m < nMels; m++ {
		copy(out[m*targetFrames:m*targetFrames+maxFrames], mel[m*frames:m*frames+maxFrames])
	}
	return out, targetFrames
}

func getMelFilterBank(cfg MelConfig) [][]float64 {
	key := fmt.Sprintf("%d-%d-%d-%d-%f-%f", cfg.SampleRate, cfg.NMels, cfg.NFFT, cfg.WinLength, cfg.FMin, cfg.FMax)
	if cached, ok := melFilterCache.Load(key); ok {
		return cached.([][]float64)
	}

	fMin := cfg.FMin
	fMax := cfg.FMax
	if fMin < 0 {
		fMin = 0
	}
	if fMax > float64(cfg.SampleRate)/2 {
		fMax = float64(cfg.SampleRate) / 2
	}

	melMin := hzToMel(fMin)
	melMax := hzToMel(fMax)
	melPoints := linspace(melMin, melMax, cfg.NMels+2)
	bin := make([]int, cfg.NMels+2)
	for i, m := range melPoints {
		hz := melToHz(m)
		b := int(math.Floor((float64(cfg.NFFT)+1)*hz/float64(cfg.SampleRate)))
		if b < 0 {
			b = 0
		}
		maxBin := cfg.NFFT/2 + 1
		if b > maxBin {
			b = maxBin
		}
		bin[i] = b
	}

	filters := make([][]float64, cfg.NMels)
	for m := 0; m < cfg.NMels; m++ {
		filters[m] = make([]float64, cfg.NFFT/2+1)
		left := bin[m]
		center := bin[m+1]
		right := bin[m+2]
		if center == left {
			center++
		}
		if right == center {
			right++
		}
		for k := left; k < center && k < len(filters[m]); k++ {
			filters[m][k] = float64(k-left) / float64(center-left)
		}
		for k := center; k < right && k < len(filters[m]); k++ {
			filters[m][k] = float64(right-k) / float64(right-center)
		}
	}

	melFilterCache.Store(key, filters)
	return filters
}

func linspace(start, end float64, count int) []float64 {
	if count <= 1 {
		return []float64{start}
	}
	step := (end - start) / float64(count-1)
	vals := make([]float64, count)
	for i := 0; i < count; i++ {
		vals[i] = start + step*float64(i)
	}
	return vals
}

func hzToMel(f float64) float64 {
	return 2595 * math.Log10(1+f/700)
}

func melToHz(m float64) float64 {
	return 700 * (math.Pow(10, m/2595) - 1)
}
