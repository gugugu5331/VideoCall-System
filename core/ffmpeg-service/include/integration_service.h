#pragma once

#include "ffmpeg_processor.h"
#include "onnx_detector.h"
#include <memory>
#include <string>
#include <functional>
#include <thread>
#include <mutex>
#include <atomic>
#include <queue>

namespace integration_service {

// 集成检测类型
enum class IntegratedDetectionType {
    REAL_TIME_VIDEO,      // 实时视频检测
    REAL_TIME_AUDIO,      // 实时音频检测
    BATCH_VIDEO,          // 批量视频检测
    BATCH_AUDIO,          // 批量音频检测
    HYBRID_DETECTION      // 混合检测（音视频同时）
};

// 集成检测结果
struct IntegratedDetectionResult {
    bool is_fake;
    float overall_confidence;
    float overall_risk_score;
    
    // 视频检测结果
    onnx_detector::DetectionResult video_result;
    
    // 音频检测结果
    onnx_detector::DetectionResult audio_result;
    
    // 压缩信息
    ffmpeg_service::ProcessingResult compression_result;
    
    // 性能指标
    int64_t total_processing_time_ms;
    float compression_ratio;
    int64_t frame_count;
    
    // 详细信息
    std::unordered_map<std::string, float> detailed_metrics;
    std::string detection_summary;
    
    IntegratedDetectionResult() : is_fake(false), overall_confidence(0.0f), 
                                overall_risk_score(0.0f), total_processing_time_ms(0),
                                compression_ratio(1.0f), frame_count(0) {}
};

// 集成服务配置
struct IntegrationConfig {
    // FFmpeg配置
    ffmpeg_service::EncodingParams ffmpeg_params;
    
    // ONNX检测配置
    onnx_detector::ModelConfig video_model_config;
    onnx_detector::ModelConfig audio_model_config;
    onnx_detector::PreprocessingParams preprocessing_params;
    
    // 集成配置
    float video_weight = 0.6f;      // 视频检测权重
    float audio_weight = 0.4f;      // 音频检测权重
    float confidence_threshold = 0.8f;
    float risk_threshold = 0.7f;
    
    // 性能配置
    int max_batch_size = 10;
    int processing_threads = 4;
    bool enable_compression = true;
    bool enable_real_time = true;
    
    // 缓存配置
    bool enable_feature_cache = true;
    size_t cache_size = 1000;
    int cache_ttl_seconds = 3600;
};

// 回调函数类型
using IntegratedDetectionCallback = std::function<void(const IntegratedDetectionResult&)>;
using ProgressCallback = std::function<void(int progress_percent, const std::string& status)>;

// 集成服务主类
class IntegrationService {
public:
    IntegrationService();
    ~IntegrationService();
    
    // 初始化和清理
    bool initialize(const IntegrationConfig& config = IntegrationConfig{});
    void cleanup();
    
    // 实时检测接口
    bool startRealTimeDetection(IntegratedDetectionType type,
                               IntegratedDetectionCallback callback = nullptr);
    void stopRealTimeDetection();
    
    // 批量检测接口
    std::vector<IntegratedDetectionResult> batchDetectVideo(
        const std::vector<std::vector<uint8_t>>& video_batch,
        ProgressCallback progress_callback = nullptr);
    
    std::vector<IntegratedDetectionResult> batchDetectAudio(
        const std::vector<std::vector<uint8_t>>& audio_batch,
        ProgressCallback progress_callback = nullptr);
    
    std::vector<IntegratedDetectionResult> batchDetectHybrid(
        const std::vector<std::pair<std::vector<uint8_t>, std::vector<uint8_t>>>& media_batch,
        ProgressCallback progress_callback = nullptr);
    
    // 单次检测接口
    IntegratedDetectionResult detectVideo(const std::vector<uint8_t>& video_data,
                                        int width, int height, int fps = 30);
    
    IntegratedDetectionResult detectAudio(const std::vector<uint8_t>& audio_data,
                                        int sample_rate, int channels);
    
    IntegratedDetectionResult detectHybrid(const std::vector<uint8_t>& video_data,
                                          const std::vector<uint8_t>& audio_data,
                                          int width, int height, int fps = 30,
                                          int sample_rate = 44100, int channels = 2);
    
    // 流式检测接口
    bool startStreamingDetection(const std::string& stream_url,
                                IntegratedDetectionCallback callback = nullptr);
    void stopStreamingDetection();
    
    // 模型管理
    bool loadVideoModel(const std::string& model_path, 
                       const onnx_detector::ModelConfig& config = onnx_detector::ModelConfig{});
    bool loadAudioModel(const std::string& model_path,
                       const onnx_detector::ModelConfig& config = onnx_detector::ModelConfig{});
    bool reloadModels();
    
    // 配置管理
    void setIntegrationConfig(const IntegrationConfig& config);
    void setFFmpegParams(const ffmpeg_service::EncodingParams& params);
    void setVideoModelConfig(const onnx_detector::ModelConfig& config);
    void setAudioModelConfig(const onnx_detector::ModelConfig& config);
    
    // 性能监控
    void enablePerformanceMonitoring(bool enable);
    void getPerformanceStats(std::unordered_map<std::string, double>& stats);
    void resetPerformanceStats();
    
    // 状态查询
    bool isInitialized() const { return initialized_; }
    bool isProcessing() const { return processing_; }
    IntegrationConfig getCurrentConfig() const { return current_config_; }
    std::string getServiceStatus() const;

private:
    // 内部处理函数
    bool initializeComponents();
    void cleanupComponents();
    
    // 检测处理函数
    IntegratedDetectionResult processVideoDetection(const std::vector<uint8_t>& video_data,
                                                   int width, int height, int fps);
    
    IntegratedDetectionResult processAudioDetection(const std::vector<uint8_t>& audio_data,
                                                   int sample_rate, int channels);
    
    IntegratedDetectionResult combineResults(const onnx_detector::DetectionResult& video_result,
                                            const onnx_detector::DetectionResult& audio_result,
                                            const ffmpeg_service::ProcessingResult& compression_result);
    
    // 实时处理函数
    void realTimeProcessingLoop();
    void handleRealTimeVideoFrame(const ffmpeg_service::MediaFrame& frame);
    void handleRealTimeAudioFrame(const ffmpeg_service::MediaFrame& frame);
    
    // 流式处理函数
    void streamingProcessingLoop();
    
    // 缓存管理
    void updateFeatureCache(const std::string& key, const std::vector<float>& features);
    bool getCachedFeatures(const std::string& key, std::vector<float>& features);
    void cleanupExpiredCache();
    
    // 成员变量
    std::unique_ptr<ffmpeg_service::FFmpegProcessor> ffmpeg_processor_;
    std::unique_ptr<onnx_detector::ONNXDetector> video_detector_;
    std::unique_ptr<onnx_detector::ONNXDetector> audio_detector_;
    std::unique_ptr<onnx_detector::PerformanceMonitor> performance_monitor_;
    
    IntegrationConfig current_config_;
    IntegratedDetectionCallback detection_callback_;
    ProgressCallback progress_callback_;
    
    // 线程管理
    std::thread real_time_thread_;
    std::thread streaming_thread_;
    std::thread cache_cleanup_thread_;
    
    std::mutex processing_mutex_;
    std::mutex cache_mutex_;
    std::condition_variable processing_cv_;
    
    std::atomic<bool> initialized_;
    std::atomic<bool> processing_;
    std::atomic<bool> should_stop_;
    std::atomic<bool> performance_monitoring_enabled_;
    
    // 缓存
    std::unordered_map<std::string, std::pair<std::vector<float>, int64_t>> feature_cache_;
    std::queue<std::string> cache_keys_;
    
    // 状态信息
    std::string current_status_;
    std::mutex status_mutex_;
};

// 服务管理器
class ServiceManager {
public:
    ServiceManager();
    ~ServiceManager();
    
    // 服务管理
    bool startService(const IntegrationConfig& config = IntegrationConfig{});
    void stopService();
    void restartService();
    
    // 服务状态
    bool isServiceRunning() const { return service_running_; }
    std::string getServiceStatus() const;
    
    // 配置管理
    bool loadConfigFromFile(const std::string& config_file);
    bool saveConfigToFile(const std::string& config_file);
    
    // 健康检查
    bool performHealthCheck();
    std::unordered_map<std::string, bool> getComponentStatus();

private:
    std::unique_ptr<IntegrationService> integration_service_;
    IntegrationConfig config_;
    std::atomic<bool> service_running_;
    std::mutex service_mutex_;
};

// 工具函数
namespace utils {
    // 配置验证
    bool validateConfig(const IntegrationConfig& config);
    
    // 性能优化
    void optimizeForPlatform(IntegrationConfig& config);
    
    // 日志记录
    void logDetectionResult(const IntegratedDetectionResult& result);
    void logPerformanceStats(const std::unordered_map<std::string, double>& stats);
    
    // 数据转换
    std::vector<uint8_t> convertVideoFormat(const std::vector<uint8_t>& data,
                                           int width, int height,
                                           ffmpeg_service::EncodingParams& params);
    
    std::vector<uint8_t> convertAudioFormat(const std::vector<uint8_t>& data,
                                           int sample_rate, int channels,
                                           ffmpeg_service::EncodingParams& params);
}

} // namespace integration_service 