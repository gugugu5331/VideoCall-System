/*
 * Synthesis Detection Task Implementation
 */
#include "synthesis_task.h"
#include <iostream>
#include <fstream>
#include <cstring>
#include <algorithm>
#include <cmath>

namespace AIInference {

SynthesisTask::SynthesisTask(const std::string &workid) 
    : BaseTask(workid), model_loaded_(false) {
}

SynthesisTask::~SynthesisTask() {
    stop();
}

int SynthesisTask::load_model(const nlohmann::json &config_body) {
    if (parse_config(config_body)) {
        std::cerr << "Failed to parse config for Synthesis task" << std::endl;
        return -1;
    }

    try {
        // Initialize ONNX Runtime environment
        env_ = std::make_unique<Ort::Env>(ORT_LOGGING_LEVEL_WARNING, "SynthesisTask");
        session_options_ = std::make_unique<Ort::SessionOptions>();
        session_options_->SetIntraOpNumThreads(4);
        session_options_->SetGraphOptimizationLevel(GraphOptimizationLevel::ORT_ENABLE_ALL);

        // Construct model path based on model name
        model_path_ = "/work/models/" + model_ + ".onnx";
        
        std::cout << "Loading Synthesis Detection model from: " << model_path_ << std::endl;

        // Check if model file exists
        std::ifstream model_file(model_path_);
        if (!model_file.good()) {
            std::cerr << "Model file not found: " << model_path_ << std::endl;
            return -1;
        }

        // Create ONNX Runtime session
        session_ = std::make_unique<Ort::Session>(*env_, model_path_.c_str(), *session_options_);

        // Get input/output information
        Ort::AllocatorWithDefaultOptions allocator;
        
        // Input names
        size_t num_input_nodes = session_->GetInputCount();
        for (size_t i = 0; i < num_input_nodes; i++) {
            auto input_name = session_->GetInputNameAllocated(i, allocator);
            input_names_.push_back(strdup(input_name.get()));
        }

        // Output names
        size_t num_output_nodes = session_->GetOutputCount();
        for (size_t i = 0; i < num_output_nodes; i++) {
            auto output_name = session_->GetOutputNameAllocated(i, allocator);
            output_names_.push_back(strdup(output_name.get()));
        }

        // Get input shape
        auto type_info = session_->GetInputTypeInfo(0);
        auto tensor_info = type_info.GetTensorTypeAndShapeInfo();
        input_shape_ = tensor_info.GetShape();

        model_loaded_ = true;
        std::cout << "Synthesis Detection model loaded successfully" << std::endl;
        return 0;

    } catch (const Ort::Exception& e) {
        std::cerr << "ONNX Runtime error: " << e.what() << std::endl;
        return -1;
    } catch (const std::exception& e) {
        std::cerr << "Error loading Synthesis Detection model: " << e.what() << std::endl;
        return -1;
    }
}

void SynthesisTask::inference(const std::string &msg) {
    if (!model_loaded_) {
        std::cerr << "Model not loaded, cannot perform inference" << std::endl;
        if (out_callback_) {
            out_callback_("Error: Model not loaded", true);
        }
        return;
    }

    std::lock_guard<std::mutex> lock(inference_mutex_);

    try {
        // Preprocess audio data
        std::vector<float> input_data = preprocess_audio(msg);

        // Create input tensor
        Ort::MemoryInfo memory_info = Ort::MemoryInfo::CreateCpu(OrtArenaAllocator, OrtMemTypeDefault);

        // Fix dynamic dimensions in input shape
        std::vector<int64_t> actual_input_shape = input_shape_;
        for (size_t i = 0; i < actual_input_shape.size(); i++) {
            if (actual_input_shape[i] == -1) {
                if (i == 0) {
                    actual_input_shape[i] = 1; // Batch size
                } else if (i == 2) {
                    // Calculate n_frames from input data size
                    actual_input_shape[i] = input_data.size() / (actual_input_shape[0] * actual_input_shape[1]);
                } else {
                    actual_input_shape[i] = 1; // Default to 1 for other dynamic dimensions
                }
            }
        }

        Ort::Value input_tensor = Ort::Value::CreateTensor<float>(
            memory_info,
            input_data.data(),
            input_data.size(),
            actual_input_shape.data(),
            actual_input_shape.size()
        );

        // Run inference
        auto output_tensors = session_->Run(
            Ort::RunOptions{nullptr},
            input_names_.data(),
            &input_tensor,
            1,
            output_names_.data(),
            output_names_.size()
        );

        // Get output data
        float* output_data = output_tensors[0].GetTensorMutableData<float>();
        auto output_shape = output_tensors[0].GetTensorTypeAndShapeInfo().GetShape();
        
        size_t output_size = 1;
        for (auto dim : output_shape) {
            output_size *= dim;
        }

        std::vector<float> output_vec(output_data, output_data + output_size);
        
        // Postprocess output
        std::string result = postprocess_output(output_vec);

        // Send result via callback
        if (out_callback_) {
            if (enstream_) {
                out_callback_(result, false);
                out_callback_("", true);
            } else {
                out_callback_(result, true);
            }
        }

    } catch (const Ort::Exception& e) {
        std::cerr << "ONNX Runtime inference error: " << e.what() << std::endl;
        if (out_callback_) {
            out_callback_("Error: Inference failed", true);
        }
    } catch (const std::exception& e) {
        std::cerr << "Inference error: " << e.what() << std::endl;
        if (out_callback_) {
            out_callback_("Error: " + std::string(e.what()), true);
        }
    }
}

void SynthesisTask::start() {
    std::cout << "Synthesis Detection task started for work_id: " << work_id_ << std::endl;
}

void SynthesisTask::stop() {
    std::cout << "Synthesis Detection task stopped for work_id: " << work_id_ << std::endl;
    model_loaded_ = false;
    
    // Free input/output names
    for (auto name : input_names_) {
        free(const_cast<char*>(name));
    }
    input_names_.clear();
    
    for (auto name : output_names_) {
        free(const_cast<char*>(name));
    }
    output_names_.clear();
}

std::vector<float> SynthesisTask::preprocess_audio(const std::string &audio_data) {
    // Simple preprocessing: convert audio data to float array
    // In a real implementation, this would involve:
    // 1. Decoding audio format (WAV, MP3, etc.)
    // 2. Resampling to model's expected sample rate
    // 3. Feature extraction (spectral features, etc.)
    
    std::vector<float> processed_data;
    
    // For now, create dummy data matching expected input shape
    size_t expected_size = 1;
    for (auto dim : input_shape_) {
        if (dim > 0) {
            expected_size *= dim;
        }
    }
    
    processed_data.resize(expected_size, 0.0f);
    
    // Simple conversion: treat input as raw float data
    if (!audio_data.empty()) {
        size_t copy_size = std::min(audio_data.size() / sizeof(float), expected_size);
        std::memcpy(processed_data.data(), audio_data.data(), copy_size * sizeof(float));
    }
    
    return processed_data;
}

std::string SynthesisTask::postprocess_output(const std::vector<float> &output) {
    // Binary classification: real vs synthetic
    // Apply sigmoid to get probability
    
    if (output.empty()) {
        return "No detection result available";
    }
    
    // Get the first output value (assuming binary classification)
    float raw_score = output[0];
    
    // Apply sigmoid function
    float probability = 1.0f / (1.0f + std::exp(-raw_score));
    
    // Determine if synthetic (threshold at 0.5)
    bool is_synthetic = probability > 0.5f;
    float confidence = is_synthetic ? probability : (1.0f - probability);
    
    // Create JSON response
    nlohmann::json result;
    result["is_synthetic"] = is_synthetic;
    result["is_real"] = !is_synthetic;
    result["confidence"] = confidence;
    result["probability_synthetic"] = probability;
    result["probability_real"] = 1.0f - probability;
    result["model"] = model_;
    
    return result.dump();
}

} // namespace AIInference

