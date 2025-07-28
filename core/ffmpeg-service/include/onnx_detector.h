#pragma once

#include <string>
#include <vector>
#include <memory>
#include <functional>
#include <thread>
#include <mutex>
#include <atomic>
#include <unordered_map>

// ONNX Runtime
#include <onnxruntime_cxx_api.h>

// OpenCV for image processing
#include <opencv2/opencv.hpp>

namespace onnx_detector {

// 检测类型
enum class DetectionType {
    VOICE_SPOOFING,    // 语音伪造检测
    VIDEO_DEEPFAKE,    // 视频深度伪造检测
    FACE_SWAP,         // 换脸检测
    AUDIO_ARTIFACT,    // 音频伪影检测
    VIDEO_ARTIFACT     // 视频伪影检测
};

// 检测结果
struct DetectionResult {
    bool is_fake;
    float confidence;
    float risk_score;
    std::vector<float> feature_vector;
    std::unordered_map<std::string, float> detailed_scores;
    int64_t processing_time_ms;
    std::string model_version;
    
    DetectionResult() : is_fake(false), confidence(0.0f), risk_score(0.0f), 
                       processing_time_ms(0) {}
};

// 模型配置
struct ModelConfig {
    std::string model_path;
    std::string model_name;
    std::vector<int64_t> input_shape;
    std::vector<int64_t> output_shape;
    std::string input_name;
    std::string output_name;
    float confidence_threshold;
    float risk_threshold;
    bool enable_gpu;
    int gpu_device_id;
    int num_threads;
    bool enable_optimization;
    
    ModelConfig() : confidence_threshold(0.8f), risk_threshold(0.7f),
                   enable_gpu(false), gpu_device_id(0), num_threads(4),
                   enable_optimization(true) {}
};

// 预处理参数
struct PreprocessingParams {
    int target_width;
    int target_height;
    float mean_r;
    float mean_g;
    float mean_b;
    float std_r;
    float std_g;
    float std_b;
    bool normalize;
    bool resize;
    bool crop;
    
    PreprocessingParams() : target_width(224), target_height(224),
                           mean_r(0.485f), mean_g(0.456f), mean_b(0.406f),
                           std_r(0.229f), std_g(0.224f), std_b(0.225f),
                           normalize(true), resize(true), crop(false) {}
};

// 回调函数类型
using DetectionCallback = std::function<void(const DetectionResult&)>;

// ONNX检测器主类
class ONNXDetector {
public:
    ONNXDetector();
    ~ONNXDetector();
    
    // 初始化和清理
    bool initialize(const std::string& model_path, const ModelConfig& config = ModelConfig{});
    void cleanup();
    
    // 检测接口
    DetectionResult detectVoiceSpoofing(const std::vector<uint8_t>& audio_data,
                                       int sample_rate, int channels);
    
    DetectionResult detectVideoDeepfake(const std::vector<uint8_t>& video_data,
                                       int width, int height, int fps = 30);
    
    DetectionResult detectFaceSwap(const std::vector<uint8_t>& video_data,
                                  int width, int height, int fps = 30);
    
    DetectionResult detectAudioArtifact(const std::vector<uint8_t>& audio_data,
                                       int sample_rate, int channels);
    
    DetectionResult detectVideoArtifact(const std::vector<uint8_t>& video_data,
                                       int width, int height, int fps = 30);
    
    // 批量检测
    std::vector<DetectionResult> batchDetect(const std::vector<std::vector<uint8_t>>& data_batch,
                                            DetectionType type);
    
    // 实时检测
    void startRealTimeDetection(DetectionCallback callback = nullptr);
    void stopRealTimeDetection();
    
    // 模型管理
    bool loadModel(const std::string& model_path, const ModelConfig& config = ModelConfig{});
    bool reloadModel();
    bool switchModel(const std::string& model_path, const ModelConfig& config = ModelConfig{});
    
    // 参数设置
    void setModelConfig(const ModelConfig& config);
    void setPreprocessingParams(const PreprocessingParams& params);
    void setDetectionCallback(DetectionCallback callback);
    
    // 状态查询
    bool isInitialized() const { return initialized_; }
    bool isProcessing() const { return processing_; }
    ModelConfig getCurrentConfig() const { return current_config_; }
    std::string getModelVersion() const { return model_version_; }

private:
    // 内部处理函数
    bool initializeSession();
    bool initializeProviders();
    void cleanupSession();
    
    // 预处理函数
    std::vector<float> preprocessAudio(const std::vector<uint8_t>& audio_data,
                                      int sample_rate, int channels);
    
    std::vector<float> preprocessVideo(const std::vector<uint8_t>& video_data,
                                      int width, int height);
    
    std::vector<float> preprocessImage(const cv::Mat& image);
    
    // 推理函数
    DetectionResult runInference(const std::vector<float>& input_data);
    std::vector<float> extractFeatures(const std::vector<float>& input_data);
    
    // 后处理函数
    DetectionResult postprocessOutput(const std::vector<float>& output_data,
                                     DetectionType type);
    
    // 成员变量
    Ort::Env env_;
    Ort::Session session_;
    Ort::SessionOptions session_options_;
    
    ModelConfig current_config_;
    PreprocessingParams preprocessing_params_;
    DetectionCallback detection_callback_;
    
    std::thread detection_thread_;
    std::mutex session_mutex_;
    std::atomic<bool> initialized_;
    std::atomic<bool> processing_;
    std::atomic<bool> should_stop_;
    
    std::string model_version_;
    std::vector<const char*> input_names_;
    std::vector<const char*> output_names_;
    
    // 缓存
    std::unordered_map<std::string, std::vector<float>> feature_cache_;
    std::mutex cache_mutex_;
};

// 音频特征提取器
class AudioFeatureExtractor {
public:
    AudioFeatureExtractor();
    ~AudioFeatureExtractor();
    
    bool initialize(int sample_rate, int channels);
    void cleanup();
    
    std::vector<float> extractMFCC(const std::vector<uint8_t>& audio_data);
    std::vector<float> extractSpectrogram(const std::vector<uint8_t>& audio_data);
    std::vector<float> extractMelSpectrogram(const std::vector<uint8_t>& audio_data);
    std::vector<float> extractLPC(const std::vector<uint8_t>& audio_data);

private:
    int sample_rate_;
    int channels_;
    bool initialized_;
};

// 视频特征提取器
class VideoFeatureExtractor {
public:
    VideoFeatureExtractor();
    ~VideoFeatureExtractor();
    
    bool initialize(int width, int height);
    void cleanup();
    
    std::vector<float> extractFacialFeatures(const std::vector<uint8_t>& video_data);
    std::vector<float> extractTemporalFeatures(const std::vector<uint8_t>& video_data);
    std::vector<float> extractArtifactFeatures(const std::vector<uint8_t>& video_data);
    std::vector<float> extractMotionFeatures(const std::vector<uint8_t>& video_data);

private:
    int width_;
    int height_;
    bool initialized_;
    
    // OpenCV相关
    cv::CascadeClassifier face_cascade_;
    cv::Ptr<cv::Feature2D> feature_detector_;
};

// 模型优化器
class ModelOptimizer {
public:
    ModelOptimizer();
    ~ModelOptimizer();
    
    bool optimizeModel(const std::string& input_model_path,
                      const std::string& output_model_path,
                      const ModelConfig& config);
    
    bool quantizeModel(const std::string& input_model_path,
                      const std::string& output_model_path,
                      const std::string& calibration_data_path);
    
    bool fuseOperations(const std::string& input_model_path,
                       const std::string& output_model_path);

private:
    bool applyGraphOptimizations(Ort::SessionOptions& options);
    bool applyExecutionProviderOptimizations(Ort::SessionOptions& options);
};

// 性能监控器
class PerformanceMonitor {
public:
    PerformanceMonitor();
    ~PerformanceMonitor();
    
    void startTimer();
    void endTimer();
    
    void recordInferenceTime(int64_t time_ms);
    void recordPreprocessingTime(int64_t time_ms);
    void recordPostprocessingTime(int64_t time_ms);
    
    double getAverageInferenceTime() const;
    double getAveragePreprocessingTime() const;
    double getAveragePostprocessingTime() const;
    
    void reset();

private:
    std::vector<int64_t> inference_times_;
    std::vector<int64_t> preprocessing_times_;
    std::vector<int64_t> postprocessing_times_;
    
    std::chrono::high_resolution_clock::time_point start_time_;
    std::mutex stats_mutex_;
};

} // namespace onnx_detector 