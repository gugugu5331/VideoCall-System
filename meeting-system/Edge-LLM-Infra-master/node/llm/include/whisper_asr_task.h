/*
 * Whisper ASR Task (支持中英文)
 */
#pragma once

#include "base_task.h"
#include <onnxruntime_cxx_api.h>
#include <mutex>
#include <unordered_map>
#include <vector>

namespace AIInference {

class WhisperASRTask : public BaseTask {
private:
    std::unique_ptr<Ort::Env> env_;
    std::unique_ptr<Ort::Session> encoder_session_;
    std::unique_ptr<Ort::Session> decoder_session_;
    std::unique_ptr<Ort::SessionOptions> session_options_;
    std::mutex inference_mutex_;

    std::vector<const char*> encoder_input_names_;
    std::vector<const char*> encoder_output_names_;
    std::vector<int64_t> encoder_input_shape_;

    std::vector<const char*> decoder_input_names_;
    std::vector<const char*> decoder_output_names_;
    std::vector<int64_t> decoder_input_shape_;

    bool model_loaded_;
    std::string encoder_path_;
    std::string decoder_path_;
    
    // Whisper 配置
    int n_mels_;
    int mel_length_;
    int n_audio_ctx_;
    int n_audio_state_;
    int n_vocab_;
    
    // Tokenizer
    std::unordered_map<int, std::string> id2token_;
    int sot_token_;  // start of transcript
    int eot_token_;  // end of transcript
    int transcribe_token_;
    int zh_token_;  // 中文
    int en_token_;  // 英文
    int no_timestamps_token_;

public:
    WhisperASRTask(const std::string &workid);
    ~WhisperASRTask() override;

    int load_model(const nlohmann::json &config_body) override;
    void inference(const std::string &msg) override;
    void start() override;
    void stop() override;

private:
    std::vector<float> preprocess_audio(const std::string &audio_data);
    std::vector<float> compute_mel_spectrogram(const std::vector<float> &audio);
    std::string postprocess_output(const std::vector<float> &encoder_output);
    std::string greedy_decode(const std::vector<float> &encoder_output, const std::vector<int64_t> &encoder_output_shape);
    std::vector<int64_t> decode_with_decoder(
        const std::vector<int64_t> &input_tokens,
        const std::vector<float> &encoder_output,
        const std::vector<int64_t> &encoder_output_shape,
        int max_length = 100
    );
    std::string tokens_to_text(const std::vector<int64_t> &tokens);
    void load_vocabulary();
    void load_config();
};

} // namespace AIInference

