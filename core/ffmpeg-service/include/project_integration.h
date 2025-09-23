#pragma once

#include "integration_service.h"
#include <string>
#include <memory>
#include <functional>

namespace project_integration {

// 与Python AI服务的集成接口
class PythonAIServiceIntegration {
public:
    PythonAIServiceIntegration();
    ~PythonAIServiceIntegration();
    
    // 初始化集成
    bool initialize(const std::string& config_file = "");
    
    // 检测接口（与Python AI服务兼容）
    struct DetectionRequest {
        std::string detection_id;
        std::string detection_type;  // "voice_spoofing", "video_deepfake", "face_swap"
        std::vector<uint8_t> audio_data;
        std::vector<uint8_t> video_data;
        std::unordered_map<std::string, std::string> metadata;
    };
    
    struct DetectionResponse {
        std::string detection_id;
        std::string detection_type;
        float risk_score;
        float confidence;
        std::string status;  // "completed", "failed", "processing"
        std::unordered_map<std::string, std::string> details;
        int64_t processing_time_ms;
    };
    
    // 检测方法
    DetectionResponse detect(const DetectionRequest& request);
    
    // 批量检测
    std::vector<DetectionResponse> batchDetect(const std::vector<DetectionRequest>& requests);
    
    // 实时检测
    bool startRealTimeDetection(std::function<void(const DetectionResponse&)> callback);
    void stopRealTimeDetection();
    
    // 状态查询
    bool isInitialized() const;
    std::string getStatus() const;

private:
    std::unique_ptr<integration_service::IntegrationService> service_;
    std::unique_ptr<integration_service::ServiceManager> service_manager_;
    bool initialized_;
};

// 与Go后端的集成接口
class GoBackendIntegration {
public:
    GoBackendIntegration();
    ~GoBackendIntegration();
    
    // 初始化集成
    bool initialize(const std::string& config_file = "");
    
    // 检测结果结构（与Go后端兼容）
    struct GoDetectionResult {
        bool is_fake;
        float confidence;
        float risk_score;
        std::string detection_type;
        std::string details;
        int64_t processing_time_ms;
        float compression_ratio;
    };
    
    // 检测方法
    GoDetectionResult detectVideo(const std::vector<uint8_t>& video_data, 
                                 int width, int height, int fps = 30);
    
    GoDetectionResult detectAudio(const std::vector<uint8_t>& audio_data,
                                 int sample_rate, int channels);
    
    GoDetectionResult detectHybrid(const std::vector<uint8_t>& video_data,
                                  const std::vector<uint8_t>& audio_data,
                                  int width, int height, int fps = 30,
                                  int sample_rate = 44100, int channels = 2);
    
    // 批量检测
    std::vector<GoDetectionResult> batchDetectVideo(const std::vector<std::vector<uint8_t>>& video_batch);
    std::vector<GoDetectionResult> batchDetectAudio(const std::vector<std::vector<uint8_t>>& audio_batch);
    
    // 性能监控
    struct PerformanceStats {
        double avg_inference_time;
        double avg_preprocessing_time;
        double avg_postprocessing_time;
        int total_detections;
        double success_rate;
    };
    
    PerformanceStats getPerformanceStats();
    void resetPerformanceStats();
    
    // 配置管理
    bool loadConfig(const std::string& config_file);
    bool saveConfig(const std::string& config_file);
    
    // 健康检查
    bool performHealthCheck();
    std::unordered_map<std::string, bool> getComponentStatus();

private:
    std::unique_ptr<integration_service::IntegrationService> service_;
    std::unique_ptr<integration_service::ServiceManager> service_manager_;
    bool initialized_;
};

// 与WebRTC的集成接口
class WebRTCIntegration {
public:
    WebRTCIntegration();
    ~WebRTCIntegration();
    
    // 初始化集成
    bool initialize(const std::string& config_file = "");
    
    // WebRTC媒体流处理
    struct MediaStream {
        std::vector<uint8_t> video_data;
        std::vector<uint8_t> audio_data;
        int video_width;
        int video_height;
        int video_fps;
        int audio_sample_rate;
        int audio_channels;
        int64_t timestamp;
    };
    
    // 实时流检测
    bool startStreamDetection(std::function<void(const integration_service::IntegratedDetectionResult&)> callback);
    void stopStreamDetection();
    
    // 处理媒体流
    integration_service::IntegratedDetectionResult processMediaStream(const MediaStream& stream);
    
    // 流式检测配置
    struct StreamConfig {
        int detection_interval_ms;  // 检测间隔
        bool enable_video_detection;
        bool enable_audio_detection;
        float confidence_threshold;
        float risk_threshold;
        bool enable_compression;
    };
    
    void setStreamConfig(const StreamConfig& config);
    StreamConfig getStreamConfig() const;
    
    // 状态查询
    bool isStreaming() const;
    std::string getStreamStatus() const;

private:
    std::unique_ptr<integration_service::IntegrationService> service_;
    StreamConfig stream_config_;
    bool streaming_;
    bool initialized_;
};

// 与Docker容器的集成接口
class DockerIntegration {
public:
    DockerIntegration();
    ~DockerIntegration();
    
    // 初始化Docker集成
    bool initialize(const std::string& docker_config_file = "");
    
    // Docker服务管理
    bool startService();
    void stopService();
    bool restartService();
    
    // 服务状态
    bool isServiceRunning() const;
    std::string getServiceStatus() const;
    
    // 配置管理
    bool loadDockerConfig(const std::string& config_file);
    bool saveDockerConfig(const std::string& config_file);
    
    // 日志管理
    std::string getServiceLogs(int max_lines = 100);
    void clearServiceLogs();
    
    // 资源监控
    struct ResourceUsage {
        double cpu_usage_percent;
        double memory_usage_mb;
        double disk_usage_percent;
        int network_connections;
    };
    
    ResourceUsage getResourceUsage();

private:
    std::unique_ptr<integration_service::ServiceManager> service_manager_;
    std::string docker_config_;
    bool service_running_;
};

// 统一的集成管理器
class IntegrationManager {
public:
    IntegrationManager();
    ~IntegrationManager();
    
    // 初始化所有集成
    bool initialize(const std::string& config_file = "");
    
    // 获取各种集成接口
    PythonAIServiceIntegration* getPythonAIServiceIntegration();
    GoBackendIntegration* getGoBackendIntegration();
    WebRTCIntegration* getWebRTCIntegration();
    DockerIntegration* getDockerIntegration();
    
    // 统一配置管理
    bool loadGlobalConfig(const std::string& config_file);
    bool saveGlobalConfig(const std::string& config_file);
    
    // 全局状态管理
    bool isAllServicesRunning() const;
    std::unordered_map<std::string, std::string> getAllServiceStatus() const;
    
    // 全局性能监控
    struct GlobalPerformanceStats {
        std::unordered_map<std::string, double> service_stats;
        double overall_throughput;
        double average_response_time;
        int total_requests;
        double success_rate;
    };
    
    GlobalPerformanceStats getGlobalPerformanceStats();
    
    // 健康检查
    bool performGlobalHealthCheck();
    std::unordered_map<std::string, bool> getGlobalComponentStatus();
    
    // 日志管理
    void setLogLevel(const std::string& level);
    std::string getLogLevel() const;
    
    // 清理资源
    void cleanup();

private:
    std::unique_ptr<PythonAIServiceIntegration> python_ai_integration_;
    std::unique_ptr<GoBackendIntegration> go_backend_integration_;
    std::unique_ptr<WebRTCIntegration> webrtc_integration_;
    std::unique_ptr<DockerIntegration> docker_integration_;
    
    std::string global_config_;
    std::string log_level_;
    bool initialized_;
};

// 配置管理工具
namespace config_utils {
    
    // 配置文件结构
    struct GlobalConfig {
        // Python AI服务配置
        struct PythonAIConfig {
            std::string model_path;
            float confidence_threshold;
            float risk_threshold;
            bool enable_gpu;
            int num_threads;
        } python_ai;
        
        // Go后端配置
        struct GoBackendConfig {
            std::string service_url;
            int timeout_ms;
            bool enable_compression;
            int max_batch_size;
        } go_backend;
        
        // WebRTC配置
        struct WebRTCConfig {
            int detection_interval_ms;
            bool enable_video_detection;
            bool enable_audio_detection;
            float confidence_threshold;
            float risk_threshold;
        } webrtc;
        
        // Docker配置
        struct DockerConfig {
            std::string image_name;
            std::string container_name;
            int port;
            std::vector<std::string> environment_vars;
        } docker;
        
        // FFmpeg配置
        struct FFmpegConfig {
            int video_bitrate;
            int audio_bitrate;
            int video_width;
            int video_height;
            int video_fps;
            int audio_sample_rate;
            int audio_channels;
        } ffmpeg;
    };
    
    // 配置加载和保存
    bool loadConfigFromFile(const std::string& file_path, GlobalConfig& config);
    bool saveConfigToFile(const std::string& file_path, const GlobalConfig& config);
    
    // 配置验证
    bool validateConfig(const GlobalConfig& config);
    
    // 默认配置生成
    GlobalConfig generateDefaultConfig();
    
    // 配置合并
    GlobalConfig mergeConfigs(const GlobalConfig& base_config, const GlobalConfig& override_config);
}

// 日志管理工具
namespace log_utils {
    
    enum class LogLevel {
        DEBUG,
        INFO,
        WARNING,
        ERROR,
        FATAL
    };
    
    void setLogLevel(LogLevel level);
    LogLevel getLogLevel();
    
    void log(LogLevel level, const std::string& message);
    void logDebug(const std::string& message);
    void logInfo(const std::string& message);
    void logWarning(const std::string& message);
    void logError(const std::string& message);
    void logFatal(const std::string& message);
    
    // 性能日志
    void logPerformance(const std::string& operation, int64_t duration_ms);
    void logDetectionResult(const std::string& detection_type, bool is_fake, float confidence);
}

// 错误处理工具
namespace error_utils {
    
    enum class ErrorCode {
        SUCCESS = 0,
        INITIALIZATION_FAILED,
        CONFIG_LOAD_FAILED,
        SERVICE_START_FAILED,
        DETECTION_FAILED,
        INVALID_PARAMETER,
        RESOURCE_NOT_AVAILABLE,
        TIMEOUT,
        UNKNOWN_ERROR
    };
    
    struct ErrorInfo {
        ErrorCode code;
        std::string message;
        std::string details;
        std::string timestamp;
    };
    
    ErrorInfo getLastError();
    void setLastError(ErrorCode code, const std::string& message, const std::string& details = "");
    void clearLastError();
    
    std::string errorCodeToString(ErrorCode code);
    bool isError(ErrorCode code);
}

} // namespace project_integration 