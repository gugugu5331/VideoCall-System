#pragma once

#include <memory>
#include <string>
#include <vector>
#include <chrono>
#include <mutex>

// ONNX Runtime
#include <onnxruntime_cxx_api.h>

namespace ffmpeg_detection {

// 检测类型
enum class DetectionType {
    FACE_FORGERY,      // 人脸伪造
    DEEPFAKE,          // Deepfake
    FACE_SWAP,         // 换脸
    AUDIO_FORGERY,     // 音频伪造
    LIP_SYNC,          // 唇同步
    GENERAL_FAKE       // 通用伪造
};

// 检测结果
struct DetectionResult {
    bool is_fake;
    float confidence;
    DetectionType type;
    std::string details;
    int64_t processing_time_ms;
    std::vector<float> raw_scores;
};

// 模型配置
struct ModelConfig {
    std::string model_path;
    int input_width = 224;
    int input_height = 224;
    int input_channels = 3;
    float mean[3] = {0.485f, 0.456f, 0.406f};
    float std[3] = {0.229f, 0.224f, 0.225f};
    bool use_gpu = false;
    int gpu_device_id = 0;
    int num_threads = 4;
    float confidence_threshold = 0.5f;
};

class DetectionEngine {
public:
    DetectionEngine();
    ~DetectionEngine();

    // 初始化
    bool initialize(const ModelConfig& config);
    
    // 检测函数
    DetectionResult detect_video_frame(const std::vector<uint8_t>& frame_data, 
                                      int width, int height, int channels);
    DetectionResult detect_audio_frame(const std::vector<float>& audio_data, 
                                      int sample_rate, int channels);
    DetectionResult detect_combined(const std::vector<uint8_t>& video_data,
                                   const std::vector<float>& audio_data,
                                   int video_width, int video_height,
                                   int audio_sample_rate);
    
    // 批量检测
    std::vector<DetectionResult> detect_batch(const std::vector<std::vector<uint8_t>>& frames);
    
    // 获取模型信息
    std::string get_model_info() const;
    bool is_initialized() const;
    
    // 预热模型
    void warmup();

private:
    // 内部处理函数
    bool load_model();
    bool setup_session();
    std::vector<float> preprocess_video(const std::vector<uint8_t>& frame_data,
                                       int width, int height, int channels);
    std::vector<float> preprocess_audio(const std::vector<float>& audio_data,
                                       int sample_rate, int channels);
    DetectionResult postprocess_output(const std::vector<float>& output);
    std::string detection_type_to_string(DetectionType type);

    // ONNX Runtime 相关
    Ort::Env env_;
    Ort::Session session_;
    Ort::SessionOptions session_options_;
    
    // 模型信息
    ModelConfig config_;
    std::vector<const char*> input_names_;
    std::vector<const char*> output_names_;
    std::vector<std::vector<int64_t>> input_shapes_;
    std::vector<std::vector<int64_t>> output_shapes_;
    
    // 状态
    bool is_initialized_;
    mutable std::mutex mutex_;
    
    // 统计信息
    int64_t total_inferences_;
    double total_processing_time_ms_;
};

} // namespace ffmpeg_detection 