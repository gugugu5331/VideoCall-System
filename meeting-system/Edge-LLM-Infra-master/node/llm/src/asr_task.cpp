/*
 * ASR Task Implementation
 */
#include "asr_task.h"
#include <iostream>
#include <fstream>
#include <cstring>
#include <algorithm>
#include <sstream>

namespace AIInference {

ASRTask::ASRTask(const std::string &workid)
    : BaseTask(workid), model_loaded_(false), vocab_size_(32), blank_id_(0) {
    std::cout << "[ASRTask] Constructor called for work_id: " << workid << std::endl;

    // Load vocabulary
    load_vocabulary();
}

ASRTask::~ASRTask() {
    stop();
}

int ASRTask::load_model(const nlohmann::json &config_body) {
    std::cout << "[ASRTask] load_model called" << std::endl;

    if (parse_config(config_body)) {
        std::cerr << "[ASRTask] Failed to parse config" << std::endl;
        return -1;
    }

    std::cout << "[ASRTask] Config parsed successfully, model=" << model_ << std::endl;

    try {
        std::cout << "[ASRTask] Initializing ONNX Runtime..." << std::endl;
        // Initialize ONNX Runtime environment
        env_ = std::make_unique<Ort::Env>(ORT_LOGGING_LEVEL_WARNING, "ASRTask");
        session_options_ = std::make_unique<Ort::SessionOptions>();
        session_options_->SetIntraOpNumThreads(4);
        session_options_->SetGraphOptimizationLevel(GraphOptimizationLevel::ORT_ENABLE_ALL);
        std::cout << "[ASRTask] ONNX Runtime initialized" << std::endl;

        // Construct model path based on model name
        model_path_ = "/work/models/" + model_ + ".onnx";
        
        std::cout << "Loading ASR model from: " << model_path_ << std::endl;

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
        std::cout << "ASR model loaded successfully" << std::endl;
        return 0;

    } catch (const Ort::Exception& e) {
        std::cerr << "ONNX Runtime error: " << e.what() << std::endl;
        return -1;
    } catch (const std::exception& e) {
        std::cerr << "Error loading ASR model: " << e.what() << std::endl;
        return -1;
    }
}

void ASRTask::inference(const std::string &msg) {
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
                    // input_data.size() = batch * n_mels * n_frames
                    // Assuming batch=1, n_mels=80
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
        auto output_shape = output_info.GetShape();
        size_t output_size = output_info.GetElementCount();

        std::vector<float> output_vec(output_data, output_data + output_size);

        // Store output shape for CTC decoding
        std::vector<int64_t> shape_vec(output_shape.begin(), output_shape.end());

        // Postprocess output with CTC decoding
        std::string result = postprocess_output(output_vec, shape_vec);

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

void ASRTask::start() {
    std::cout << "ASR task started for work_id: " << work_id_ << std::endl;
}

void ASRTask::stop() {
    std::cout << "ASR task stopped for work_id: " << work_id_ << std::endl;
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

std::vector<float> ASRTask::preprocess_audio(const std::string &audio_data) {
    // Simple preprocessing: convert audio data to float array
    // In a real implementation, this would involve:
    // 1. Decoding audio format (WAV, MP3, etc.)
    // 2. Resampling to model's expected sample rate
    // 3. Feature extraction (MFCC, mel-spectrogram, etc.)
    
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

std::string ASRTask::postprocess_output(const std::vector<float> &output, const std::vector<int64_t> &output_shape) {
    // CTC decoding with output shape

    nlohmann::json result;
    result["transcription"] = "";
    result["confidence"] = 0.0;
    result["model"] = model_;

    if (output.empty()) {
        result["error"] = "empty output";
        return result.dump();
    }

    std::vector<int64_t> normalized_shape;
    if (output_shape.size() == 2) {
        // Some models output (time_steps, vocab_size) without a batch dimension.
        normalized_shape = {1, output_shape[0], output_shape[1]};
    } else if (output_shape.size() >= 3) {
        normalized_shape = {output_shape[0], output_shape[1], output_shape[2]};
    } else {
        result["error"] = "invalid output shape";
        result["output_shape"] = output_shape;
        return result.dump();
    }

    // Use CTC decode
    std::string transcription = ctc_decode(output, normalized_shape);

    // Find max probability for confidence
    auto max_it = std::max_element(output.begin(), output.end());
    float confidence = *max_it;

    // Create JSON response
    result["transcription"] = transcription;
    result["confidence"] = confidence;

    return result.dump();
}

void ASRTask::load_vocabulary() {
    std::cout << "[ASRTask] Loading vocabulary..." << std::endl;

    // Wav2Vec2 vocabulary (32 characters)
    // Based on facebook/wav2vec2-base-960h
    id2char_[0] = "<pad>";   // CTC blank token
    id2char_[1] = "<s>";
    id2char_[2] = "</s>";
    id2char_[3] = "<unk>";
    id2char_[4] = "|";       // Word separator (space)
    id2char_[5] = "E";
    id2char_[6] = "T";
    id2char_[7] = "A";
    id2char_[8] = "O";
    id2char_[9] = "N";
    id2char_[10] = "I";
    id2char_[11] = "H";
    id2char_[12] = "S";
    id2char_[13] = "R";
    id2char_[14] = "D";
    id2char_[15] = "L";
    id2char_[16] = "U";
    id2char_[17] = "M";
    id2char_[18] = "W";
    id2char_[19] = "C";
    id2char_[20] = "F";
    id2char_[21] = "G";
    id2char_[22] = "Y";
    id2char_[23] = "P";
    id2char_[24] = "B";
    id2char_[25] = "V";
    id2char_[26] = "K";
    id2char_[27] = "'";
    id2char_[28] = "X";
    id2char_[29] = "J";
    id2char_[30] = "Q";
    id2char_[31] = "Z";

    vocab_size_ = 32;
    blank_id_ = 0;  // <pad> is the CTC blank token

    std::cout << "[ASRTask] Vocabulary loaded: " << vocab_size_ << " tokens" << std::endl;
}

std::string ASRTask::ctc_decode(const std::vector<float> &logits, const std::vector<int64_t> &output_shape) {
    // CTC Greedy Decoding
    // Input: logits with shape (batch, time_steps, vocab_size)
    // Output: decoded text string

    if (logits.empty() || output_shape.size() < 3) {
        std::cout << "[ASRTask] Invalid logits or output shape" << std::endl;
        return "";
    }

    int batch_size = output_shape[0];
    int time_steps = output_shape[1];
    int vocab_size = output_shape[2];

    std::cout << "[ASRTask] CTC Decode - batch: " << batch_size
              << ", time_steps: " << time_steps
              << ", vocab_size: " << vocab_size << std::endl;

    if (vocab_size != vocab_size_) {
        std::cout << "[ASRTask] Warning: vocab_size mismatch. Expected "
                  << vocab_size_ << ", got " << vocab_size << std::endl;
    }

    // Process first batch only
    std::vector<int> predicted_ids;
    predicted_ids.reserve(time_steps);

    // For each time step, find the argmax
    for (int t = 0; t < time_steps; t++) {
        int offset = t * vocab_size;

        // Find max probability
        float max_prob = logits[offset];
        int max_id = 0;

        for (int v = 1; v < vocab_size; v++) {
            if (logits[offset + v] > max_prob) {
                max_prob = logits[offset + v];
                max_id = v;
            }
        }

        predicted_ids.push_back(max_id);
    }

    // CTC collapse: remove consecutive duplicates and blank tokens
    std::vector<int> collapsed_ids;
    int prev_id = -1;

    for (int id : predicted_ids) {
        // Skip blank token and consecutive duplicates
        if (id != blank_id_ && id != prev_id) {
            collapsed_ids.push_back(id);
        }
        prev_id = id;
    }

    std::cout << "[ASRTask] CTC collapsed: " << time_steps
              << " -> " << collapsed_ids.size() << " tokens" << std::endl;

    // Convert IDs to text
    std::stringstream ss;
    for (int id : collapsed_ids) {
        auto it = id2char_.find(id);
        if (it != id2char_.end()) {
            std::string token = it->second;

            // Handle special tokens
            if (token == "|") {
                ss << " ";  // Word separator
            } else if (token == "<s>" || token == "</s>" || token == "<unk>") {
                // Skip special tokens
                continue;
            } else {
                ss << token;
            }
        } else {
            std::cout << "[ASRTask] Warning: Unknown token ID " << id << std::endl;
        }
    }

    std::string decoded_text = ss.str();

    // Trim whitespace
    size_t start = decoded_text.find_first_not_of(" \t\n\r");
    size_t end = decoded_text.find_last_not_of(" \t\n\r");

    if (start != std::string::npos && end != std::string::npos) {
        decoded_text = decoded_text.substr(start, end - start + 1);
    }

    std::cout << "[ASRTask] Decoded text: \"" << decoded_text << "\"" << std::endl;

    return decoded_text;
}

} // namespace AIInference
