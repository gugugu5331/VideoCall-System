/*
 * Emotion Detection Task Implementation
 */
#include "emotion_task.h"
#include <iostream>
#include <fstream>
#include <cstring>
#include <algorithm>
#include <cmath>

namespace AIInference {

EmotionTask::EmotionTask(const std::string &workid) 
    : BaseTask(workid), model_loaded_(false) {
    // Initialize emotion labels (common emotion categories)
    emotion_labels_ = {"anger", "disgust", "fear", "joy", "neutral", "sadness", "surprise"};
}

EmotionTask::~EmotionTask() {
    stop();
}

int EmotionTask::load_model(const nlohmann::json &config_body) {
    if (parse_config(config_body)) {
        std::cerr << "Failed to parse config for Emotion task" << std::endl;
        return -1;
    }

    try {
        // Initialize ONNX Runtime environment
        env_ = std::make_unique<Ort::Env>(ORT_LOGGING_LEVEL_WARNING, "EmotionTask");
        session_options_ = std::make_unique<Ort::SessionOptions>();
        session_options_->SetIntraOpNumThreads(4);
        session_options_->SetGraphOptimizationLevel(GraphOptimizationLevel::ORT_ENABLE_ALL);

        // Construct model path based on model name
        model_path_ = "/work/models/" + model_ + ".onnx";
        
        std::cout << "Loading Emotion model from: " << model_path_ << std::endl;

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
        std::cout << "Emotion model loaded successfully" << std::endl;
        return 0;

    } catch (const Ort::Exception& e) {
        std::cerr << "ONNX Runtime error: " << e.what() << std::endl;
        return -1;
    } catch (const std::exception& e) {
        std::cerr << "Error loading Emotion model: " << e.what() << std::endl;
        return -1;
    }
}

void EmotionTask::inference(const std::string &msg) {
    if (!model_loaded_) {
        std::cerr << "Model not loaded, cannot perform inference" << std::endl;
        if (out_callback_) {
            out_callback_("Error: Model not loaded", true);
        }
        return;
    }

    std::lock_guard<std::mutex> lock(inference_mutex_);

    try {
        // Preprocess text data
        std::vector<float> input_data = preprocess_text(msg);

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
        auto output_info = output_tensors[0].GetTensorTypeAndShapeInfo();
        size_t output_size = output_info.GetElementCount();

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

void EmotionTask::start() {
    std::cout << "Emotion task started for work_id: " << work_id_ << std::endl;
}

void EmotionTask::stop() {
    std::cout << "Emotion task stopped for work_id: " << work_id_ << std::endl;
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

std::vector<float> EmotionTask::preprocess_text(const std::string &text) {
    // Simple preprocessing: convert text to token IDs and embeddings
    // In a real implementation, this would involve:
    // 1. Tokenization using appropriate tokenizer
    // 2. Converting tokens to IDs
    // 3. Padding/truncating to fixed length
    // 4. Creating attention masks
    
    std::vector<float> processed_data;
    
    // For now, create dummy data matching expected input shape
    size_t expected_size = 1;
    for (auto dim : input_shape_) {
        if (dim > 0) {
            expected_size *= dim;
        }
    }
    
    processed_data.resize(expected_size, 0.0f);
    
    // Simple hash-based encoding for demonstration
    if (!text.empty()) {
        std::hash<std::string> hasher;
        size_t hash_val = hasher(text);
        for (size_t i = 0; i < std::min(expected_size, size_t(128)); i++) {
            processed_data[i] = static_cast<float>((hash_val >> (i % 32)) & 0xFF) / 255.0f;
        }
    }
    
    return processed_data;
}

std::string EmotionTask::postprocess_output(const std::vector<float> &output) {
    // Apply softmax and find the emotion with highest probability
    
    if (output.empty()) {
        nlohmann::json result;
        result["emotion"] = "unknown";
        result["confidence"] = 0.0;
        result["model"] = model_;
        result["all_emotions"] = nlohmann::json::object();
        result["error"] = "empty output";
        return result.dump();
    }
    
    // Apply softmax
    std::vector<float> probabilities;
    float sum = 0.0f;
    for (float val : output) {
        float exp_val = std::exp(val);
        probabilities.push_back(exp_val);
        sum += exp_val;
    }
    
    for (float& prob : probabilities) {
        prob /= sum;
    }
    
    // Find emotion with highest probability
    auto max_it = std::max_element(probabilities.begin(), probabilities.end());
    int max_idx = std::distance(probabilities.begin(), max_it);
    float confidence = *max_it;
    
    std::string detected_emotion = "unknown";
    if (max_idx < emotion_labels_.size()) {
        detected_emotion = emotion_labels_[max_idx];
    }
    
    // Create JSON response
    nlohmann::json result;
    result["emotion"] = detected_emotion;
    result["confidence"] = confidence;
    result["model"] = model_;
    
    // Add all emotion probabilities
    nlohmann::json all_emotions;
    for (size_t i = 0; i < std::min(probabilities.size(), emotion_labels_.size()); i++) {
        all_emotions[emotion_labels_[i]] = probabilities[i];
    }
    result["all_emotions"] = all_emotions;
    
    return result.dump();
}

} // namespace AIInference
