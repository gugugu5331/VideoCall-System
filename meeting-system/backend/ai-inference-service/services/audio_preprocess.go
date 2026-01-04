package services

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// normalizeAudioPayload converts supported formats to raw PCM bytes and returns updated metadata.
func normalizeAudioPayload(data []byte, format string, fallbackSampleRate, fallbackChannels int) ([]byte, int, int, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "wav":
		pcm, sampleRate, channels, err := decodeWAV(data)
		if err != nil {
			return nil, 0, 0, err
		}
		if sampleRate == 0 {
			sampleRate = fallbackSampleRate
		}
		if channels == 0 {
			channels = fallbackChannels
		}
		return pcm, sampleRate, channels, nil
	case "pcm":
		if fallbackSampleRate == 0 {
			fallbackSampleRate = 16000
		}
		if fallbackChannels == 0 {
			fallbackChannels = 1
		}
		return data, fallbackSampleRate, fallbackChannels, nil
	default:
		return nil, 0, 0, fmt.Errorf("unsupported audio format: %s", format)
	}
}

// prepareAudioPCM converts PCM16 bytes to float32 and resamples to target rate.
func prepareAudioPCM(pcm []byte, inRate, inChannels, outRate, outChannels int) ([]float32, error) {
	if inRate <= 0 || inChannels <= 0 {
		return nil, fmt.Errorf("invalid audio metadata: sample_rate=%d channels=%d", inRate, inChannels)
	}
	if outChannels <= 0 {
		outChannels = 1
	}

	mono, err := pcm16ToMonoFloat32(pcm, inChannels)
	if err != nil {
		return nil, err
	}

	if outRate > 0 && outRate != inRate {
		mono = resampleLinear(mono, inRate, outRate)
	}

	if outChannels != 1 {
		return nil, fmt.Errorf("only mono output supported in current streaming path")
	}

	return mono, nil
}

// pcm16ToMonoFloat32 decodes little-endian PCM16 and downmixes to mono.
func pcm16ToMonoFloat32(pcm []byte, channels int) ([]float32, error) {
	if channels <= 0 {
		return nil, fmt.Errorf("invalid channel count: %d", channels)
	}
	if len(pcm)%2 != 0 {
		return nil, fmt.Errorf("pcm16 payload must be even length")
	}

	sampleCount := len(pcm) / 2
	if sampleCount%channels != 0 {
		return nil, fmt.Errorf("pcm16 samples not divisible by channels")
	}

	frameCount := sampleCount / channels
	out := make([]float32, frameCount)
	idx := 0
	for i := 0; i < frameCount; i++ {
		sum := float32(0)
		for ch := 0; ch < channels; ch++ {
			sample := int16(binary.LittleEndian.Uint16(pcm[idx : idx+2]))
			idx += 2
			sum += float32(sample) / 32768.0
		}
		out[i] = sum / float32(channels)
	}

	return out, nil
}

// resampleLinear performs a simple linear resampling.
func resampleLinear(input []float32, inRate, outRate int) []float32 {
	if inRate == outRate || len(input) == 0 {
		return input
	}

	ratio := float64(outRate) / float64(inRate)
	outLen := int(math.Round(float64(len(input)) * ratio))
	if outLen <= 0 {
		return nil
	}

	output := make([]float32, outLen)
	for i := 0; i < outLen; i++ {
		pos := float64(i) / ratio
		idx := int(pos)
		if idx >= len(input)-1 {
			output[i] = input[len(input)-1]
			continue
		}
		frac := float32(pos - float64(idx))
		output[i] = input[idx]*(1-frac) + input[idx+1]*frac
	}

	return output
}

// decodeWAV extracts PCM16 data from a WAV payload.
func decodeWAV(data []byte) ([]byte, int, int, error) {
	if len(data) < 44 {
		return nil, 0, 0, fmt.Errorf("wav payload too short")
	}
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, 0, 0, fmt.Errorf("invalid wav header")
	}

	var (
		sampleRate    int
		channels      int
		bitsPerSample int
		pcmData       []byte
		fmtFound      bool
	)

	pos := 12
	for pos+8 <= len(data) {
		chunkID := string(data[pos : pos+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		pos += 8
		if pos+chunkSize > len(data) {
			return nil, 0, 0, fmt.Errorf("wav chunk out of bounds")
		}

		switch chunkID {
		case "fmt ":
			if chunkSize < 16 {
				return nil, 0, 0, fmt.Errorf("wav fmt chunk too short")
			}
			audioFormat := int(binary.LittleEndian.Uint16(data[pos : pos+2]))
			channels = int(binary.LittleEndian.Uint16(data[pos+2 : pos+4]))
			sampleRate = int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
			bitsPerSample = int(binary.LittleEndian.Uint16(data[pos+14 : pos+16]))
			if audioFormat != 1 {
				return nil, 0, 0, fmt.Errorf("unsupported wav format: %d", audioFormat)
			}
			fmtFound = true
		case "data":
			pcmData = data[pos : pos+chunkSize]
		}

		pos += chunkSize
		if chunkSize%2 == 1 {
			pos++
		}
	}

	if !fmtFound {
		return nil, 0, 0, fmt.Errorf("wav fmt chunk not found")
	}
	if bitsPerSample != 16 {
		return nil, 0, 0, fmt.Errorf("unsupported wav bit depth: %d", bitsPerSample)
	}
	if pcmData == nil {
		return nil, 0, 0, fmt.Errorf("wav data chunk not found")
	}

	return pcmData, sampleRate, channels, nil
}
