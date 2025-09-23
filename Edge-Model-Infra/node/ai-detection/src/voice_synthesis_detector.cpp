/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "voice_synthesis_detector.h"
#include "detection_utils.h"
#include <iostream>
#include <random>
#include <cmath>
#include <algorithm>

using namespace StackFlows;

VoiceSynthesisDetector::VoiceSynthesisDetector() 
    : model_loaded_(false), target_sample_rate_(16000), frame_length_(1024),
      hop_length_(512), n_mfcc_(13), n_fft_(2048), detection_threshold_(0.5f),
      fft_plan_(nullptr), fft_input_(nullptr), fft_output_(nullptr), fft_size_(0) {
}

VoiceSynthesisDetector::~VoiceSynthesisDetector() {
    // Cleanup FFTW resources
    if (fft_plan_) {
        fftw_destroy_plan(fft_plan_);
    }
    if (fft_input_) {
        fftw_free(fft_input_);
    }
    if (fft_output_) {
        fftw_free(fft_output_);
    }
    
#ifdef USE_TENSORFLOW
    if (tf_session_) {
        tf_session_->Close();
    }
#endif
}

bool VoiceSynthesisDetector::initialize(const std::string& model_path) {
    model_path_ = model_path;
    
    // Initialize FFTW for audio processing
    fft_size_ = n_fft_;
    fft_input_ = (double*)fftw_malloc(sizeof(double) * fft_size_);
    fft_output_ = (fftw_complex*)fftw_malloc(sizeof(fftw_complex) * (fft_size_ / 2 + 1));
    fft_plan_ = fftw_plan_dft_r2c_1d(fft_size_, fft_input_, fft_output_, FFTW_ESTIMATE);
    
    if (!fft_plan_ || !fft_input_ || !fft_output_) {
        std::cerr << "Failed to initialize FFTW" << std::endl;
        return false;
    }
    
    if (DetectionUtils::file_exists(model_path)) {
        return load_model(model_path);
    } else {
        std::cout << "Model file not found, using dummy model for testing" << std::endl;
        create_dummy_model();
        return true;
    }
}

AudioDetectionResult VoiceSynthesisDetector::detect_audio(const std::string& audio_path) {
    AudioDetectionResult result;
    result.is_fake = false;
    result.confidence = 0.0f;
    result.details = "Audio analysis failed";
    
    try {
        int sample_rate;
        std::vector<float> audio_data = load_audio_file(audio_path, sample_rate);
        
        if (audio_data.empty()) {
            result.details = "Failed to load audio file";
            return result;
        }
        
        return detect_audio_data(audio_data, sample_rate);
        
    } catch (const std::exception& e) {
        result.details = "Exception in audio detection: " + std::string(e.what());
        return result;
    }
}

AudioDetectionResult VoiceSynthesisDetector::detect_audio_data(const std::vector<float>& audio_data, int sample_rate) {
    AudioDetectionResult result;
    result.is_fake = false;
    result.confidence = 0.0f;
    result.details = "Audio analysis completed";
    
    try {
        // Extract features
        std::vector<float> mfcc_features = extract_mfcc_features(audio_data, sample_rate);
        std::vector<float> spectral_features = extract_spectral_features(audio_data, sample_rate);
        
        // Combine features
        std::vector<float> combined_features;
        combined_features.insert(combined_features.end(), mfcc_features.begin(), mfcc_features.end());
        combined_features.insert(combined_features.end(), spectral_features.begin(), spectral_features.end());
        
        result.features = combined_features;
        
        // Predict using model
        float prediction = predict_voice_synthesis(combined_features);
        
        result.is_fake = prediction > detection_threshold_;
        result.confidence = result.is_fake ? prediction : (1.0f - prediction);
        result.details = result.is_fake ? "Voice synthesis detected" : "Natural voice detected";
        
    } catch (const std::exception& e) {
        result.details = "Exception in audio analysis: " + std::string(e.what());
    }
    
    return result;
}

std::vector<float> VoiceSynthesisDetector::load_audio_file(const std::string& audio_path, int& sample_rate) {
    std::vector<float> audio_data;
    
    SF_INFO sf_info;
    memset(&sf_info, 0, sizeof(sf_info));
    
    SNDFILE* file = sf_open(audio_path.c_str(), SFM_READ, &sf_info);
    if (!file) {
        std::cerr << "Failed to open audio file: " << audio_path << std::endl;
        return audio_data;
    }
    
    sample_rate = sf_info.samplerate;
    int channels = sf_info.channels;
    sf_count_t frames = sf_info.frames;
    
    // Read audio data
    std::vector<float> buffer(frames * channels);
    sf_count_t read_frames = sf_readf_float(file, buffer.data(), frames);
    sf_close(file);
    
    // Convert to mono if stereo
    if (channels == 1) {
        audio_data = buffer;
    } else {
        audio_data.reserve(read_frames);
        for (sf_count_t i = 0; i < read_frames; ++i) {
            float mono_sample = 0.0f;
            for (int c = 0; c < channels; ++c) {
                mono_sample += buffer[i * channels + c];
            }
            audio_data.push_back(mono_sample / channels);
        }
    }
    
    // Resample if necessary (simplified - just decimation/interpolation)
    if (sample_rate != target_sample_rate_) {
        float ratio = static_cast<float>(target_sample_rate_) / sample_rate;
        std::vector<float> resampled;
        resampled.reserve(static_cast<size_t>(audio_data.size() * ratio));
        
        for (size_t i = 0; i < audio_data.size(); i += static_cast<size_t>(1.0f / ratio)) {
            if (i < audio_data.size()) {
                resampled.push_back(audio_data[i]);
            }
        }
        
        audio_data = resampled;
        sample_rate = target_sample_rate_;
    }
    
    return audio_data;
}

std::vector<float> VoiceSynthesisDetector::extract_mfcc_features(const std::vector<float>& audio_data, int sample_rate) {
    std::vector<float> mfcc_features;
    
    // Simplified MFCC extraction
    // In a real implementation, you would use a proper MFCC library
    
    int num_frames = (audio_data.size() - frame_length_) / hop_length_ + 1;
    mfcc_features.reserve(num_frames * n_mfcc_);
    
    for (int frame = 0; frame < num_frames; ++frame) {
        int start_idx = frame * hop_length_;
        
        // Extract frame
        std::vector<float> frame_data(audio_data.begin() + start_idx, 
                                     audio_data.begin() + start_idx + frame_length_);
        
        // Apply window (Hamming)
        for (size_t i = 0; i < frame_data.size(); ++i) {
            float window = 0.54f - 0.46f * cos(2.0f * M_PI * i / (frame_data.size() - 1));
            frame_data[i] *= window;
        }
        
        // Compute FFT
        std::vector<float> fft_result = compute_fft(frame_data);
        
        // Apply mel filter bank
        std::vector<float> mel_features = apply_mel_filter_bank(fft_result, sample_rate);
        
        // Take DCT (simplified - just take first n_mfcc_ coefficients)
        for (int i = 0; i < n_mfcc_ && i < mel_features.size(); ++i) {
            mfcc_features.push_back(mel_features[i]);
        }
    }
    
    return mfcc_features;
}

std::vector<float> VoiceSynthesisDetector::extract_spectral_features(const std::vector<float>& audio_data, int sample_rate) {
    std::vector<float> spectral_features;
    
    // Extract simple spectral features: spectral centroid, rolloff, etc.
    int num_frames = (audio_data.size() - frame_length_) / hop_length_ + 1;
    
    for (int frame = 0; frame < num_frames; ++frame) {
        int start_idx = frame * hop_length_;
        
        std::vector<float> frame_data(audio_data.begin() + start_idx, 
                                     audio_data.begin() + start_idx + frame_length_);
        
        // Compute FFT
        std::vector<float> fft_result = compute_fft(frame_data);
        
        // Spectral centroid
        float centroid = 0.0f;
        float magnitude_sum = 0.0f;
        for (size_t i = 0; i < fft_result.size(); ++i) {
            float freq = static_cast<float>(i) * sample_rate / (2 * fft_result.size());
            centroid += freq * fft_result[i];
            magnitude_sum += fft_result[i];
        }
        centroid = magnitude_sum > 0 ? centroid / magnitude_sum : 0.0f;
        spectral_features.push_back(centroid);
        
        // Spectral rolloff (95% of energy)
        float energy_threshold = 0.95f * magnitude_sum;
        float cumulative_energy = 0.0f;
        float rolloff = 0.0f;
        for (size_t i = 0; i < fft_result.size(); ++i) {
            cumulative_energy += fft_result[i];
            if (cumulative_energy >= energy_threshold) {
                rolloff = static_cast<float>(i) * sample_rate / (2 * fft_result.size());
                break;
            }
        }
        spectral_features.push_back(rolloff);
    }
    
    return spectral_features;
}

std::vector<float> VoiceSynthesisDetector::compute_fft(const std::vector<float>& audio_data) {
    std::vector<float> magnitude_spectrum;
    
    if (!fft_plan_ || audio_data.size() > fft_size_) {
        return magnitude_spectrum;
    }
    
    // Copy data to FFTW input buffer
    for (size_t i = 0; i < audio_data.size() && i < fft_size_; ++i) {
        fft_input_[i] = audio_data[i];
    }
    
    // Zero-pad if necessary
    for (size_t i = audio_data.size(); i < fft_size_; ++i) {
        fft_input_[i] = 0.0;
    }
    
    // Execute FFT
    fftw_execute(fft_plan_);
    
    // Compute magnitude spectrum
    magnitude_spectrum.reserve(fft_size_ / 2 + 1);
    for (int i = 0; i < fft_size_ / 2 + 1; ++i) {
        float real = fft_output_[i][0];
        float imag = fft_output_[i][1];
        magnitude_spectrum.push_back(sqrt(real * real + imag * imag));
    }
    
    return magnitude_spectrum;
}

std::vector<float> VoiceSynthesisDetector::apply_mel_filter_bank(const std::vector<float>& fft_data, int sample_rate) {
    std::vector<float> mel_features;

    // Simplified mel filter bank (normally would use proper mel scale conversion)
    int num_filters = 26;
    int fft_size = fft_data.size();

    mel_features.reserve(num_filters);

    for (int filter = 0; filter < num_filters; ++filter) {
        float filter_sum = 0.0f;
        int start_bin = (filter * fft_size) / (num_filters + 1);
        int end_bin = ((filter + 2) * fft_size) / (num_filters + 1);

        for (int bin = start_bin; bin < end_bin && bin < fft_size; ++bin) {
            // Triangular filter
            float weight = 1.0f;
            if (bin < (start_bin + end_bin) / 2) {
                weight = static_cast<float>(bin - start_bin) / ((start_bin + end_bin) / 2 - start_bin);
            } else {
                weight = static_cast<float>(end_bin - bin) / (end_bin - (start_bin + end_bin) / 2);
            }

            filter_sum += fft_data[bin] * weight;
        }

        // Log mel energy
        mel_features.push_back(log(std::max(filter_sum, 1e-10f)));
    }

    return mel_features;
}

float VoiceSynthesisDetector::predict_voice_synthesis(const std::vector<float>& features) {
    if (!model_loaded_) {
        // Return random prediction for testing
        static std::random_device rd;
        static std::mt19937 gen(rd());
        static std::uniform_real_distribution<float> dis(0.0f, 1.0f);
        return dis(gen);
    }

#ifdef USE_TENSORFLOW
    if (tf_session_) {
        // Convert features to TensorFlow Tensor
        tensorflow::Tensor input_tensor(tensorflow::DT_FLOAT,
            tensorflow::TensorShape({1, static_cast<int>(features.size())}));

        auto input_tensor_mapped = input_tensor.tensor<float, 2>();
        for (size_t i = 0; i < features.size(); ++i) {
            input_tensor_mapped(0, i) = features[i];
        }

        // Run inference
        std::vector<tensorflow::Tensor> outputs;
        tensorflow::Status status = tf_session_->Run(
            {{input_layer_name_, input_tensor}},
            {output_layer_name_},
            {},
            &outputs
        );

        if (status.ok() && !outputs.empty()) {
            auto output_tensor = outputs[0].tensor<float, 2>();
            return output_tensor(0, 0);
        }
    }
#endif

    // Fallback: simple heuristic based on feature statistics
    if (features.empty()) {
        return 0.0f;
    }

    float mean = 0.0f;
    float variance = 0.0f;

    for (float feature : features) {
        mean += feature;
    }
    mean /= features.size();

    for (float feature : features) {
        variance += (feature - mean) * (feature - mean);
    }
    variance /= features.size();

    // Simple heuristic: unusual variance might indicate synthesis
    float normalized_variance = std::min(1.0f, variance / 100.0f);
    return normalized_variance;
}

bool VoiceSynthesisDetector::load_model(const std::string& model_path) {
#ifdef USE_TENSORFLOW
    try {
        tensorflow::SessionOptions options;
        tf_session_.reset(tensorflow::NewSession(options));

        tensorflow::GraphDef graph_def;
        tensorflow::Status status = tensorflow::ReadBinaryProto(
            tensorflow::Env::Default(), model_path, &graph_def);

        if (!status.ok()) {
            std::cerr << "Failed to load model: " << status.ToString() << std::endl;
            return false;
        }

        status = tf_session_->Create(graph_def);
        if (!status.ok()) {
            std::cerr << "Failed to create session: " << status.ToString() << std::endl;
            return false;
        }

        input_layer_name_ = "input_1";
        output_layer_name_ = "output_1";
        model_loaded_ = true;

        std::cout << "TensorFlow voice model loaded successfully" << std::endl;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Exception loading TensorFlow model: " << e.what() << std::endl;
        return false;
    }
#else
    std::cout << "TensorFlow not available, using dummy model" << std::endl;
    create_dummy_model();
    return true;
#endif
}

void VoiceSynthesisDetector::create_dummy_model() {
    model_loaded_ = true;
    std::cout << "Using dummy voice synthesis detection model" << std::endl;
}
