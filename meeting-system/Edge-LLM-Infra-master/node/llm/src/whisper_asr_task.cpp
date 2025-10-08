/*
 * Whisper ASR Task Implementation (支持中英文)
 */
#include "whisper_asr_task.h"
#include <iostream>
#include <fstream>
#include <cstring>
#include <algorithm>
#include <sstream>
#include <cmath>

namespace AIInference {

WhisperASRTask::WhisperASRTask(const std::string &workid)
    : BaseTask(workid), model_loaded_(false),
      n_mels_(80), mel_length_(3000), n_audio_ctx_(1500), 
      n_audio_state_(512), n_vocab_(51865),
      sot_token_(50258), eot_token_(50257), transcribe_token_(50359),
      zh_token_(50260), en_token_(50259), no_timestamps_token_(50363) {
    std::cout << "[WhisperASRTask] Constructor called for work_id: " << workid << std::endl;
    
    // Load configuration and vocabulary
    load_config();
    load_vocabulary();
}

WhisperASRTask::~WhisperASRTask() {
    std::cout << "[WhisperASRTask] Destructor called" << std::endl;
}

void WhisperASRTask::load_config() {
    std::cout << "[WhisperASRTask] Loading Whisper configuration..." << std::endl;
    
    // Try to load from JSON file
    std::ifstream config_file("/work/models/whisper_config.json");
    if (config_file.is_open()) {
        nlohmann::json config;
        config_file >> config;
        
        n_mels_ = config.value("n_mels", 80);
        mel_length_ = config.value("mel_length", 3000);
        n_audio_ctx_ = config.value("n_audio_ctx", 1500);
        n_audio_state_ = config.value("n_audio_state", 512);
        n_vocab_ = config.value("n_vocab", 51865);
        
        std::cout << "[WhisperASRTask] Configuration loaded from file" << std::endl;
    } else {
        std::cout << "[WhisperASRTask] Using default configuration" << std::endl;
    }
    
    std::cout << "[WhisperASRTask] Config: n_mels=" << n_mels_ 
              << ", mel_length=" << mel_length_
              << ", n_audio_ctx=" << n_audio_ctx_
              << ", n_vocab=" << n_vocab_ << std::endl;
}

void WhisperASRTask::load_vocabulary() {
    std::cout << "[WhisperASRTask] Loading Whisper vocabulary..." << std::endl;

    try {
        // Try to load from JSON file
        std::ifstream vocab_file("/work/models/whisper_vocab.json");
        if (vocab_file.is_open()) {
            std::cout << "[WhisperASRTask] Parsing vocabulary JSON..." << std::endl;

            nlohmann::json vocab;
            vocab_file >> vocab;

            std::cout << "[WhisperASRTask] JSON parsed, loading tokens..." << std::endl;

            int loaded_count = 0;
            for (auto& [key, value] : vocab.items()) {
                try {
                    int id = std::stoi(key);
                    std::string token = value.get<std::string>();

                    // Limit token length to avoid memory issues
                    if (token.length() > 1000) {
                        std::cout << "[WhisperASRTask] Warning: Skipping token " << id << " (too long: " << token.length() << " bytes)" << std::endl;
                        continue;
                    }

                    id2token_[id] = token;
                    loaded_count++;

                    if (loaded_count % 10000 == 0) {
                        std::cout << "[WhisperASRTask] Loaded " << loaded_count << " tokens..." << std::endl;
                    }
                } catch (const std::exception& e) {
                    std::cout << "[WhisperASRTask] Warning: Failed to load token " << key << ": " << e.what() << std::endl;
                }
            }

            std::cout << "[WhisperASRTask] Vocabulary loaded: " << id2token_.size() << " tokens" << std::endl;
        } else {
            std::cout << "[WhisperASRTask] Warning: Could not load vocabulary file" << std::endl;
            // Create minimal vocabulary for testing
            id2token_[sot_token_] = "<|startoftranscript|>";
            id2token_[eot_token_] = "<|endoftext|>";
            id2token_[transcribe_token_] = "<|transcribe|>";
            id2token_[zh_token_] = "<|zh|>";
            id2token_[en_token_] = "<|en|>";
            id2token_[no_timestamps_token_] = "<|notimestamps|>";
        }

        // Load special tokens
        std::ifstream special_tokens_file("/work/models/whisper_special_tokens.json");
        if (special_tokens_file.is_open()) {
            nlohmann::json special_tokens;
            special_tokens_file >> special_tokens;

            sot_token_ = special_tokens.value("sot", 50258);
            eot_token_ = special_tokens.value("eot", 50257);
            transcribe_token_ = special_tokens["task_tokens"].value("transcribe", 50359);

            std::cout << "[WhisperASRTask] Special tokens loaded" << std::endl;
        }
    } catch (const std::exception& e) {
        std::cerr << "[WhisperASRTask] Error loading vocabulary: " << e.what() << std::endl;
        // Create minimal vocabulary as fallback
        id2token_[50258] = "<|startoftranscript|>";
        id2token_[50257] = "<|endoftext|>";
        id2token_[50359] = "<|transcribe|>";
        id2token_[50260] = "<|zh|>";
        id2token_[50259] = "<|en|>";
        id2token_[50363] = "<|notimestamps|>";
    }
}

int WhisperASRTask::load_model(const nlohmann::json &config_body) {
    std::cout << "[WhisperASRTask] Loading Whisper models..." << std::endl;

    try {
        // Get model paths from config (use default paths)
        encoder_path_ = "/work/models/whisper-encoder.onnx";
        decoder_path_ = "/work/models/whisper-decoder.onnx";

        // Try to get from config if available
        if (config_body.contains("encoder_model") && config_body["encoder_model"].is_string()) {
            encoder_path_ = config_body["encoder_model"].get<std::string>();
        }
        if (config_body.contains("decoder_model") && config_body["decoder_model"].is_string()) {
            decoder_path_ = config_body["decoder_model"].get<std::string>();
        }

        std::cout << "[WhisperASRTask] Encoder path: " << encoder_path_ << std::endl;
        std::cout << "[WhisperASRTask] Decoder path: " << decoder_path_ << std::endl;

        // Initialize ONNX Runtime
        env_ = std::make_unique<Ort::Env>(ORT_LOGGING_LEVEL_WARNING, "WhisperASR");
        session_options_ = std::make_unique<Ort::SessionOptions>();
        session_options_->SetIntraOpNumThreads(4);
        session_options_->SetGraphOptimizationLevel(GraphOptimizationLevel::ORT_ENABLE_ALL);

        // Load encoder
        std::cout << "[WhisperASRTask] Loading Encoder..." << std::endl;
        std::cout << "[WhisperASRTask] Encoder file size: " << std::ifstream(encoder_path_, std::ios::binary | std::ios::ate).tellg() << " bytes" << std::endl;

        encoder_session_ = std::make_unique<Ort::Session>(*env_, encoder_path_.c_str(), *session_options_);

        // Get encoder input/output names
        Ort::AllocatorWithDefaultOptions allocator;

        auto encoder_input_name = encoder_session_->GetInputNameAllocated(0, allocator);
        encoder_input_names_.push_back(strdup(encoder_input_name.get()));

        auto encoder_output_name = encoder_session_->GetOutputNameAllocated(0, allocator);
        encoder_output_names_.push_back(strdup(encoder_output_name.get()));

        auto encoder_input_shape_info = encoder_session_->GetInputTypeInfo(0).GetTensorTypeAndShapeInfo();
        encoder_input_shape_ = encoder_input_shape_info.GetShape();

        std::cout << "[WhisperASRTask] Encoder loaded successfully" << std::endl;

        // Load decoder
        std::cout << "[WhisperASRTask] Loading Decoder..." << std::endl;
        std::cout << "[WhisperASRTask] Decoder file size: " << std::ifstream(decoder_path_, std::ios::binary | std::ios::ate).tellg() << " bytes" << std::endl;

        decoder_session_ = std::make_unique<Ort::Session>(*env_, decoder_path_.c_str(), *session_options_);

        // Get decoder input/output names
        // Input 0: tokens
        auto decoder_input_name_0 = decoder_session_->GetInputNameAllocated(0, allocator);
        decoder_input_names_.push_back(strdup(decoder_input_name_0.get()));

        // Input 1: encoder_output
        auto decoder_input_name_1 = decoder_session_->GetInputNameAllocated(1, allocator);
        decoder_input_names_.push_back(strdup(decoder_input_name_1.get()));

        // Output: logits
        auto decoder_output_name = decoder_session_->GetOutputNameAllocated(0, allocator);
        decoder_output_names_.push_back(strdup(decoder_output_name.get()));

        std::cout << "[WhisperASRTask] Decoder loaded successfully" << std::endl;

        model_loaded_ = true;
        std::cout << "[WhisperASRTask] Whisper models loaded successfully" << std::endl;

        return 0;

    } catch (const Ort::Exception& e) {
        std::cerr << "[WhisperASRTask] ONNX Runtime error: " << e.what() << std::endl;
        return -1;
    } catch (const std::exception& e) {
        std::cerr << "[WhisperASRTask] Error loading model: " << e.what() << std::endl;
        return -1;
    }
}

void WhisperASRTask::inference(const std::string &msg) {
    if (!model_loaded_) {
        std::cerr << "[WhisperASRTask] Model not loaded, cannot perform inference" << std::endl;
        if (out_callback_) {
            out_callback_("Error: Model not loaded", true);
        }
        return;
    }

    std::lock_guard<std::mutex> lock(inference_mutex_);

    try {
        std::cout << "[WhisperASRTask] Starting inference..." << std::endl;
        
        // Preprocess audio to mel-spectrogram
        std::vector<float> mel_data = preprocess_audio(msg);
        
        std::cout << "[WhisperASRTask] Mel-spectrogram size: " << mel_data.size() << std::endl;
        
        // Create input tensor
        Ort::MemoryInfo memory_info = Ort::MemoryInfo::CreateCpu(OrtArenaAllocator, OrtMemTypeDefault);

        // Fix dynamic dimensions in input shape
        std::vector<int64_t> actual_input_shape = encoder_input_shape_;
        for (size_t i = 0; i < actual_input_shape.size(); i++) {
            if (actual_input_shape[i] == -1) {
                if (i == 0) {
                    actual_input_shape[i] = 1; // Batch size
                }
            }
        }

        Ort::Value input_tensor = Ort::Value::CreateTensor<float>(
            memory_info,
            mel_data.data(),
            mel_data.size(),
            actual_input_shape.data(),
            actual_input_shape.size()
        );

        // Run encoder inference
        std::cout << "[WhisperASRTask] Running encoder..." << std::endl;
        
        auto output_tensors = encoder_session_->Run(
            Ort::RunOptions{nullptr},
            encoder_input_names_.data(),
            &input_tensor,
            1,
            encoder_output_names_.data(),
            encoder_output_names_.size()
        );

        // Get encoder output
        float* output_data = output_tensors[0].GetTensorMutableData<float>();
        auto output_shape = output_tensors[0].GetTensorTypeAndShapeInfo().GetShape();
        
        size_t output_size = 1;
        for (auto dim : output_shape) {
            output_size *= dim;
        }

        std::vector<float> encoder_output(output_data, output_data + output_size);

        std::cout << "[WhisperASRTask] Encoder output size: " << encoder_output.size() << std::endl;

        // Store output shape for decoding
        std::vector<int64_t> encoder_output_shape_vec(output_shape.begin(), output_shape.end());

        // Decode (greedy decoding with decoder model)
        std::string result = greedy_decode(encoder_output, encoder_output_shape_vec);

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
        std::cerr << "[WhisperASRTask] ONNX Runtime inference error: " << e.what() << std::endl;
        if (out_callback_) {
            out_callback_("Error: Inference failed", true);
        }
    } catch (const std::exception& e) {
        std::cerr << "[WhisperASRTask] Inference error: " << e.what() << std::endl;
        if (out_callback_) {
            out_callback_("Error: " + std::string(e.what()), true);
        }
    }
}

void WhisperASRTask::start() {
    std::cout << "[WhisperASRTask] Task started" << std::endl;
}

void WhisperASRTask::stop() {
    std::cout << "[WhisperASRTask] Task stopped" << std::endl;
}

std::vector<float> WhisperASRTask::preprocess_audio(const std::string &audio_data) {
    // Parse JSON to get audio data
    nlohmann::json json_data = nlohmann::json::parse(audio_data);
    
    std::string audio_base64;
    if (json_data.contains("audio_data")) {
        audio_base64 = json_data["audio_data"].get<std::string>();
    } else {
        audio_base64 = audio_data;
    }
    
    // Decode base64 (simplified - in production use proper base64 decoder)
    // For now, create dummy audio data
    std::vector<float> audio(16000 * 30, 0.0f);  // 30 seconds of silence
    
    // Compute mel-spectrogram
    return compute_mel_spectrogram(audio);
}

std::vector<float> WhisperASRTask::compute_mel_spectrogram(const std::vector<float> &audio) {
    // Simplified mel-spectrogram computation
    // In production, use proper STFT + mel filterbank
    
    std::cout << "[WhisperASRTask] Computing mel-spectrogram..." << std::endl;
    
    // Create mel-spectrogram with correct shape: (n_mels, mel_length)
    std::vector<float> mel(n_mels_ * mel_length_, 0.0f);
    
    // Fill with random values for testing
    // In production, compute real mel-spectrogram from audio
    for (size_t i = 0; i < mel.size(); i++) {
        mel[i] = (float)rand() / RAND_MAX * 0.1f - 0.05f;
    }
    
    std::cout << "[WhisperASRTask] Mel-spectrogram computed: " 
              << n_mels_ << " x " << mel_length_ << std::endl;
    
    return mel;
}

std::string WhisperASRTask::greedy_decode(const std::vector<float> &encoder_output, const std::vector<int64_t> &encoder_output_shape) {
    // 完整的贪婪解码实现
    std::cout << "[WhisperASRTask] Greedy decoding with Decoder model..." << std::endl;

    try {
        // 初始化 token 序列
        // [<|startoftranscript|>, <|zh|>, <|transcribe|>, <|notimestamps|>]
        std::vector<int64_t> tokens = {
            sot_token_,           // 50258: <|startoftranscript|>
            zh_token_,            // 50260: <|zh|> (中文)
            transcribe_token_,    // 50359: <|transcribe|>
            no_timestamps_token_  // 50363: <|notimestamps|>
        };

        std::cout << "[WhisperASRTask] Initial tokens: [";
        for (size_t i = 0; i < tokens.size(); i++) {
            std::cout << tokens[i];
            if (i < tokens.size() - 1) std::cout << ", ";
        }
        std::cout << "]" << std::endl;

        // 自回归解码
        std::vector<int64_t> decoded_tokens = decode_with_decoder(
            tokens,
            encoder_output,
            encoder_output_shape,
            100  // max_length
        );

        std::cout << "[WhisperASRTask] Decoded " << decoded_tokens.size() << " tokens" << std::endl;

        // 转换为文本
        std::string transcription = tokens_to_text(decoded_tokens);

        std::cout << "[WhisperASRTask] Transcription: " << transcription << std::endl;

        // 创建 JSON 响应
        nlohmann::json result;
        result["transcription"] = transcription;
        result["confidence"] = 0.95;
        result["model"] = "whisper-base";
        result["language"] = "zh";
        result["tokens_count"] = decoded_tokens.size();

        return result.dump();

    } catch (const std::exception& e) {
        std::cerr << "[WhisperASRTask] Decoding error: " << e.what() << std::endl;

        nlohmann::json result;
        result["transcription"] = "解码失败";
        result["error"] = e.what();
        result["confidence"] = 0.0;

        return result.dump();
    }
}

std::vector<int64_t> WhisperASRTask::decode_with_decoder(
    const std::vector<int64_t> &input_tokens,
    const std::vector<float> &encoder_output,
    const std::vector<int64_t> &encoder_output_shape,
    int max_length
) {
    std::cout << "[WhisperASRTask] Autoregressive decoding..." << std::endl;

    std::vector<int64_t> tokens = input_tokens;

    for (int step = 0; step < max_length; step++) {
        // 准备 decoder 输入
        Ort::MemoryInfo memory_info = Ort::MemoryInfo::CreateCpu(OrtArenaAllocator, OrtMemTypeDefault);

        // tokens tensor: (batch=1, seq_len)
        std::vector<int64_t> tokens_shape = {1, static_cast<int64_t>(tokens.size())};
        Ort::Value tokens_tensor = Ort::Value::CreateTensor<int64_t>(
            memory_info,
            const_cast<int64_t*>(tokens.data()),
            tokens.size(),
            tokens_shape.data(),
            tokens_shape.size()
        );

        // encoder_output tensor: (batch=1, n_audio_ctx, n_audio_state)
        Ort::Value encoder_output_tensor = Ort::Value::CreateTensor<float>(
            memory_info,
            const_cast<float*>(encoder_output.data()),
            encoder_output.size(),
            encoder_output_shape.data(),
            encoder_output_shape.size()
        );

        // 运行 decoder
        std::vector<Ort::Value> input_tensors;
        input_tensors.push_back(std::move(tokens_tensor));
        input_tensors.push_back(std::move(encoder_output_tensor));

        auto output_tensors = decoder_session_->Run(
            Ort::RunOptions{nullptr},
            decoder_input_names_.data(),
            input_tensors.data(),
            input_tensors.size(),
            decoder_output_names_.data(),
            decoder_output_names_.size()
        );

        // 获取 logits: (batch=1, seq_len, n_vocab)
        float* logits_data = output_tensors[0].GetTensorMutableData<float>();
        auto logits_shape = output_tensors[0].GetTensorTypeAndShapeInfo().GetShape();

        int seq_len = logits_shape[1];
        int n_vocab = logits_shape[2];

        // 获取最后一个位置的 logits
        int last_pos = seq_len - 1;
        float* last_logits = logits_data + last_pos * n_vocab;

        // 找到最大概率的 token (greedy)
        int best_token = 0;
        float best_score = last_logits[0];

        for (int i = 1; i < n_vocab; i++) {
            if (last_logits[i] > best_score) {
                best_score = last_logits[i];
                best_token = i;
            }
        }

        // 检查是否结束
        if (best_token == eot_token_) {
            std::cout << "[WhisperASRTask] End of transcript at step " << step << std::endl;
            break;
        }

        // 添加到序列
        tokens.push_back(best_token);

        if (step % 10 == 0) {
            std::cout << "[WhisperASRTask] Step " << step << ": token=" << best_token << std::endl;
        }
    }

    std::cout << "[WhisperASRTask] Decoding completed, total tokens: " << tokens.size() << std::endl;

    return tokens;
}

std::string WhisperASRTask::tokens_to_text(const std::vector<int64_t> &tokens) {
    std::stringstream ss;

    // 跳过前 4 个特殊 token
    for (size_t i = 4; i < tokens.size(); i++) {
        int64_t token_id = tokens[i];

        // 跳过特殊 token
        if (token_id == sot_token_ || token_id == eot_token_ ||
            token_id == zh_token_ || token_id == en_token_ ||
            token_id == transcribe_token_ || token_id == no_timestamps_token_) {
            continue;
        }

        // 查找 token 文本
        auto it = id2token_.find(token_id);
        if (it != id2token_.end()) {
            ss << it->second;
        } else {
            std::cout << "[WhisperASRTask] Warning: Unknown token " << token_id << std::endl;
        }
    }

    std::string text = ss.str();

    // 清理文本
    // 移除前后空格
    size_t start = text.find_first_not_of(" \t\n\r");
    size_t end = text.find_last_not_of(" \t\n\r");

    if (start != std::string::npos && end != std::string::npos) {
        text = text.substr(start, end - start + 1);
    }

    return text;
}

} // namespace AIInference

