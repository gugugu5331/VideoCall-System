/*
 * ASR (Automatic Speech Recognition) Task
 */
#pragma once

#include "base_task.h"
#include <onnxruntime_cxx_api.h>
#include <mutex>

namespace AIInference {

class ASRTask : public BaseTask {
private:
    std::unique_ptr<Ort::Env> env_;
    std::unique_ptr<Ort::Session> session_;
    std::unique_ptr<Ort::SessionOptions> session_options_;
    std::mutex inference_mutex_;
    
    std::vector<const char*> input_names_;
    std::vector<const char*> output_names_;
    std::vector<int64_t> input_shape_;
    
    bool model_loaded_;
    std::string model_path_;

public:
    ASRTask(const std::string &workid);
    ~ASRTask() override;

    int load_model(const nlohmann::json &config_body) override;
    void inference(const std::string &msg) override;
    void start() override;
    void stop() override;

private:
    std::vector<float> preprocess_audio(const std::string &audio_data);
    std::string postprocess_output(const std::vector<float> &output, const std::vector<int64_t> &output_shape);
    std::string ctc_decode(const std::vector<float> &logits, const std::vector<int64_t> &output_shape);
    void load_vocabulary();

    // Vocabulary mapping
    std::unordered_map<int, std::string> id2char_;
    int vocab_size_;
    int blank_id_;  // CTC blank token ID
};

} // namespace AIInference

