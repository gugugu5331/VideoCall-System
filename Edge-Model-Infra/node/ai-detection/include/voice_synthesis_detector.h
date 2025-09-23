/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include <string>
#include <vector>
#include <memory>
#include <sndfile.h>
#include <fftw3.h>

#ifdef USE_TENSORFLOW
#include <tensorflow/cc/client/client_session.h>
#include <tensorflow/cc/ops/standard_ops.h>
#include <tensorflow/core/framework/tensor.h>
#endif

namespace StackFlows {

struct AudioDetectionResult {
    bool is_fake;
    float confidence;
    std::string details;
    std::vector<float> features;
};

class VoiceSynthesisDetector {
public:
    VoiceSynthesisDetector();
    virtual ~VoiceSynthesisDetector();

    // Initialize the detector with model path
    bool initialize(const std::string& model_path);

    // Detect voice synthesis in audio file
    AudioDetectionResult detect_audio(const std::string& audio_path);

    // Detect voice synthesis in audio data
    AudioDetectionResult detect_audio_data(const std::vector<float>& audio_data, int sample_rate);

    // Check if detector is ready
    bool is_ready() const { return model_loaded_; }

private:
    // Audio processing
    std::vector<float> load_audio_file(const std::string& audio_path, int& sample_rate);
    std::vector<float> extract_mfcc_features(const std::vector<float>& audio_data, int sample_rate);
    std::vector<float> extract_spectral_features(const std::vector<float>& audio_data, int sample_rate);
    
    // Feature extraction
    std::vector<float> compute_fft(const std::vector<float>& audio_data);
    std::vector<float> apply_mel_filter_bank(const std::vector<float>& fft_data, int sample_rate);
    
    // Model inference
    float predict_voice_synthesis(const std::vector<float>& features);
    
    // Model management
    bool load_model(const std::string& model_path);
    void create_dummy_model();

private:
    bool model_loaded_;
    std::string model_path_;
    
#ifdef USE_TENSORFLOW
    std::unique_ptr<tensorflow::Session> tf_session_;
    std::string input_layer_name_;
    std::string output_layer_name_;
#endif
    
    // Audio processing parameters
    int target_sample_rate_;
    int frame_length_;
    int hop_length_;
    int n_mfcc_;
    int n_fft_;
    float detection_threshold_;
    
    // FFTW plans for efficient FFT computation
    fftw_plan fft_plan_;
    double* fft_input_;
    fftw_complex* fft_output_;
    int fft_size_;
};

} // namespace StackFlows
