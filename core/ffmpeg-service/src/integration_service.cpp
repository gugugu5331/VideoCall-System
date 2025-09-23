#include "../include/integration_service.h"
#include <iostream>
#include <fstream>
#include <chrono>
#include <algorithm>
#include <sstream>

namespace integration_service {

// IntegrationService 实现
IntegrationService::IntegrationService() 
    : initialized_(false), processing_(false), should_stop_(false), 
      performance_monitoring_enabled_(false) {
}

IntegrationService::~IntegrationService() {
    cleanup();
}

bool IntegrationService::initialize(const IntegrationConfig& config) {
    if (initialized_) {
        return true;
    }
    
    try {
        current_config_ = config;
        
        if (!initializeComponents()) {
            return false;
        }
        
        initialized_ = true;
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Integration service initialization error: " << e.what() << std::endl;
        return false;
    }
}

void IntegrationService::cleanup() {
    if (processing_) {
        stopRealTimeDetection();
        stopStreamingDetection();
    }
    
    cleanupComponents();
    initialized_ = false;
}

bool IntegrationService::startRealTimeDetection(IntegratedDetectionType type,
                                               IntegratedDetectionCallback callback) {
    if (!initialized_ || processing_) {
        return false;
    }
    
    try {
        detection_callback_ = callback;
        processing_ = true;
        should_stop_ = false;
        
        real_time_thread_ = std::thread(&IntegrationService::realTimeProcessingLoop, this);
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Real-time detection start error: " << e.what() << std::endl;
        return false;
    }
}

void IntegrationService::stopRealTimeDetection() {
    if (!processing_) {
        return;
    }
    
    should_stop_ = true;
    
    if (real_time_thread_.joinable()) {
        real_time_thread_.join();
    }
    
    processing_ = false;
}

std::vector<IntegratedDetectionResult> IntegrationService::batchDetectVideo(
    const std::vector<std::vector<uint8_t>>& video_batch,
    ProgressCallback progress_callback) {
    std::vector<IntegratedDetectionResult> results;
    
    if (!initialized_) {
        return results;
    }
    
    try {
        size_t total_batches = video_batch.size();
        
        for (size_t i = 0; i < total_batches; ++i) {
            if (should_stop_) {
                break;
            }
            
            // 更新进度
            if (progress_callback) {
                int progress = static_cast<int>((i * 100) / total_batches);
                progress_callback(progress, "Processing video batch " + std::to_string(i + 1));
            }
            
            // 处理单个视频
            IntegratedDetectionResult result = detectVideo(video_batch[i], 1280, 720, 30);
            results.push_back(result);
        }
        
        if (progress_callback) {
            progress_callback(100, "Video batch processing completed");
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Video batch detection error: " << e.what() << std::endl;
    }
    
    return results;
}

std::vector<IntegratedDetectionResult> IntegrationService::batchDetectAudio(
    const std::vector<std::vector<uint8_t>>& audio_batch,
    ProgressCallback progress_callback) {
    std::vector<IntegratedDetectionResult> results;
    
    if (!initialized_) {
        return results;
    }
    
    try {
        size_t total_batches = audio_batch.size();
        
        for (size_t i = 0; i < total_batches; ++i) {
            if (should_stop_) {
                break;
            }
            
            // 更新进度
            if (progress_callback) {
                int progress = static_cast<int>((i * 100) / total_batches);
                progress_callback(progress, "Processing audio batch " + std::to_string(i + 1));
            }
            
            // 处理单个音频
            IntegratedDetectionResult result = detectAudio(audio_batch[i], 44100, 2);
            results.push_back(result);
        }
        
        if (progress_callback) {
            progress_callback(100, "Audio batch processing completed");
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Audio batch detection error: " << e.what() << std::endl;
    }
    
    return results;
}

std::vector<IntegratedDetectionResult> IntegrationService::batchDetectHybrid(
    const std::vector<std::pair<std::vector<uint8_t>, std::vector<uint8_t>>>& media_batch,
    ProgressCallback progress_callback) {
    std::vector<IntegratedDetectionResult> results;
    
    if (!initialized_) {
        return results;
    }
    
    try {
        size_t total_batches = media_batch.size();
        
        for (size_t i = 0; i < total_batches; ++i) {
            if (should_stop_) {
                break;
            }
            
            // 更新进度
            if (progress_callback) {
                int progress = static_cast<int>((i * 100) / total_batches);
                progress_callback(progress, "Processing hybrid batch " + std::to_string(i + 1));
            }
            
            // 处理音视频对
            const auto& media_pair = media_batch[i];
            IntegratedDetectionResult result = detectHybrid(
                media_pair.first, media_pair.second, 1280, 720, 30, 44100, 2);
            results.push_back(result);
        }
        
        if (progress_callback) {
            progress_callback(100, "Hybrid batch processing completed");
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Hybrid batch detection error: " << e.what() << std::endl;
    }
    
    return results;
}

IntegratedDetectionResult IntegrationService::detectVideo(const std::vector<uint8_t>& video_data,
                                                        int width, int height, int fps) {
    return processVideoDetection(video_data, width, height, fps);
}

IntegratedDetectionResult IntegrationService::detectAudio(const std::vector<uint8_t>& audio_data,
                                                        int sample_rate, int channels) {
    return processAudioDetection(audio_data, sample_rate, channels);
}

IntegratedDetectionResult IntegrationService::detectHybrid(const std::vector<uint8_t>& video_data,
                                                          const std::vector<uint8_t>& audio_data,
                                                          int width, int height, int fps,
                                                          int sample_rate, int channels) {
    IntegratedDetectionResult result;
    
    if (!initialized_) {
        result.detection_summary = "Service not initialized";
        return result;
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    try {
        // 并行处理视频和音频
        auto video_result = processVideoDetection(video_data, width, height, fps);
        auto audio_result = processAudioDetection(audio_data, sample_rate, channels);
        
        // 合并结果
        result = combineResults(video_result.video_result, audio_result.audio_result, 
                              video_result.compression_result);
        
        auto end_time = std::chrono::high_resolution_clock::now();
        result.total_processing_time_ms = std::chrono::duration_cast<std::chrono::milliseconds>(
            end_time - start_time).count();
        
        // 生成检测摘要
        std::stringstream summary;
        summary << "Hybrid detection completed in " << result.total_processing_time_ms << "ms. ";
        summary << "Video confidence: " << video_result.video_result.confidence << ", ";
        summary << "Audio confidence: " << audio_result.audio_result.confidence << ", ";
        summary << "Overall risk: " << result.overall_risk_score;
        result.detection_summary = summary.str();
        
    } catch (const std::exception& e) {
        result.detection_summary = "Error: " + std::string(e.what());
    }
    
    return result;
}

bool IntegrationService::startStreamingDetection(const std::string& stream_url,
                                                IntegratedDetectionCallback callback) {
    if (!initialized_ || processing_) {
        return false;
    }
    
    try {
        detection_callback_ = callback;
        processing_ = true;
        should_stop_ = false;
        
        streaming_thread_ = std::thread(&IntegrationService::streamingProcessingLoop, this);
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Streaming detection start error: " << e.what() << std::endl;
        return false;
    }
}

void IntegrationService::stopStreamingDetection() {
    if (!processing_) {
        return;
    }
    
    should_stop_ = true;
    
    if (streaming_thread_.joinable()) {
        streaming_thread_.join();
    }
    
    processing_ = false;
}

bool IntegrationService::loadVideoModel(const std::string& model_path, 
                                       const onnx_detector::ModelConfig& config) {
    if (!video_detector_) {
        return false;
    }
    
    return video_detector_->loadModel(model_path, config);
}

bool IntegrationService::loadAudioModel(const std::string& model_path,
                                       const onnx_detector::ModelConfig& config) {
    if (!audio_detector_) {
        return false;
    }
    
    return audio_detector_->loadModel(model_path, config);
}

bool IntegrationService::reloadModels() {
    bool success = true;
    
    if (video_detector_) {
        success &= video_detector_->reloadModel();
    }
    
    if (audio_detector_) {
        success &= audio_detector_->reloadModel();
    }
    
    return success;
}

void IntegrationService::setIntegrationConfig(const IntegrationConfig& config) {
    current_config_ = config;
    
    if (ffmpeg_processor_) {
        ffmpeg_processor_->setEncodingParams(config.ffmpeg_params);
    }
    
    if (video_detector_) {
        video_detector_->setModelConfig(config.video_model_config);
        video_detector_->setPreprocessingParams(config.preprocessing_params);
    }
    
    if (audio_detector_) {
        audio_detector_->setModelConfig(config.audio_model_config);
        audio_detector_->setPreprocessingParams(config.preprocessing_params);
    }
}

void IntegrationService::setFFmpegParams(const ffmpeg_service::EncodingParams& params) {
    current_config_.ffmpeg_params = params;
    
    if (ffmpeg_processor_) {
        ffmpeg_processor_->setEncodingParams(params);
    }
}

void IntegrationService::setVideoModelConfig(const onnx_detector::ModelConfig& config) {
    current_config_.video_model_config = config;
    
    if (video_detector_) {
        video_detector_->setModelConfig(config);
    }
}

void IntegrationService::setAudioModelConfig(const onnx_detector::ModelConfig& config) {
    current_config_.audio_model_config = config;
    
    if (audio_detector_) {
        audio_detector_->setModelConfig(config);
    }
}

void IntegrationService::enablePerformanceMonitoring(bool enable) {
    performance_monitoring_enabled_ = enable;
}

void IntegrationService::getPerformanceStats(std::unordered_map<std::string, double>& stats) {
    if (!performance_monitor_) {
        return;
    }
    
    stats["avg_inference_time"] = performance_monitor_->getAverageInferenceTime();
    stats["avg_preprocessing_time"] = performance_monitor_->getAveragePreprocessingTime();
    stats["avg_postprocessing_time"] = performance_monitor_->getAveragePostprocessingTime();
}

void IntegrationService::resetPerformanceStats() {
    if (performance_monitor_) {
        performance_monitor_->reset();
    }
}

std::string IntegrationService::getServiceStatus() const {
    std::lock_guard<std::mutex> lock(status_mutex_);
    return current_status_;
}

bool IntegrationService::initializeComponents() {
    try {
        // 初始化FFmpeg处理器
        ffmpeg_processor_ = std::make_unique<ffmpeg_service::FFmpegProcessor>();
        if (!ffmpeg_processor_->initialize(current_config_.ffmpeg_params)) {
            return false;
        }
        
        // 初始化视频检测器
        video_detector_ = std::make_unique<onnx_detector::ONNXDetector>();
        if (!video_detector_->initialize("", current_config_.video_model_config)) {
            return false;
        }
        
        // 初始化音频检测器
        audio_detector_ = std::make_unique<onnx_detector::ONNXDetector>();
        if (!audio_detector_->initialize("", current_config_.audio_model_config)) {
            return false;
        }
        
        // 初始化性能监控器
        performance_monitor_ = std::make_unique<onnx_detector::PerformanceMonitor>();
        
        // 启动缓存清理线程
        cache_cleanup_thread_ = std::thread([this]() {
            while (!should_stop_) {
                std::this_thread::sleep_for(std::chrono::seconds(60)); // 每分钟清理一次
                cleanupExpiredCache();
            }
        });
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Component initialization error: " << e.what() << std::endl;
        return false;
    }
}

void IntegrationService::cleanupComponents() {
    if (cache_cleanup_thread_.joinable()) {
        cache_cleanup_thread_.join();
    }
    
    if (ffmpeg_processor_) {
        ffmpeg_processor_->cleanup();
    }
    
    if (video_detector_) {
        video_detector_->cleanup();
    }
    
    if (audio_detector_) {
        audio_detector_->cleanup();
    }
}

IntegratedDetectionResult IntegrationService::processVideoDetection(const std::vector<uint8_t>& video_data,
                                                                   int width, int height, int fps) {
    IntegratedDetectionResult result;
    
    if (!initialized_) {
        result.detection_summary = "Service not initialized";
        return result;
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    try {
        // 1. 压缩视频数据
        ffmpeg_service::ProcessingResult compression_result;
        if (current_config_.enable_compression && ffmpeg_processor_) {
            compression_result = ffmpeg_processor_->compressVideo(video_data, current_config_.ffmpeg_params);
        } else {
            compression_result.processed_data = video_data;
            compression_result.success = true;
            compression_result.compression_ratio = 1.0f;
        }
        
        // 2. 视频伪造检测
        onnx_detector::DetectionResult video_result;
        if (video_detector_ && compression_result.success) {
            video_result = video_detector_->detectVideoDeepfake(
                compression_result.processed_data, width, height, fps);
        }
        
        // 3. 组合结果
        result.video_result = video_result;
        result.compression_result = compression_result;
        result.overall_confidence = video_result.confidence;
        result.overall_risk_score = video_result.risk_score;
        result.is_fake = video_result.is_fake;
        result.compression_ratio = compression_result.compression_ratio;
        result.frame_count = 1;
        
        auto end_time = std::chrono::high_resolution_clock::now();
        result.total_processing_time_ms = std::chrono::duration_cast<std::chrono::milliseconds>(
            end_time - start_time).count();
        
        // 4. 生成检测摘要
        std::stringstream summary;
        summary << "Video detection completed in " << result.total_processing_time_ms << "ms. ";
        summary << "Confidence: " << video_result.confidence << ", ";
        summary << "Risk score: " << video_result.risk_score << ", ";
        summary << "Compression ratio: " << compression_result.compression_ratio;
        result.detection_summary = summary.str();
        
        // 5. 记录性能统计
        if (performance_monitoring_enabled_ && performance_monitor_) {
            performance_monitor_->recordInferenceTime(video_result.processing_time_ms);
        }
        
    } catch (const std::exception& e) {
        result.detection_summary = "Error: " + std::string(e.what());
    }
    
    return result;
}

IntegratedDetectionResult IntegrationService::processAudioDetection(const std::vector<uint8_t>& audio_data,
                                                                   int sample_rate, int channels) {
    IntegratedDetectionResult result;
    
    if (!initialized_) {
        result.detection_summary = "Service not initialized";
        return result;
    }
    
    auto start_time = std::chrono::high_resolution_clock::now();
    
    try {
        // 1. 压缩音频数据
        ffmpeg_service::ProcessingResult compression_result;
        if (current_config_.enable_compression && ffmpeg_processor_) {
            compression_result = ffmpeg_processor_->compressAudio(audio_data, current_config_.ffmpeg_params);
        } else {
            compression_result.processed_data = audio_data;
            compression_result.success = true;
            compression_result.compression_ratio = 1.0f;
        }
        
        // 2. 音频伪造检测
        onnx_detector::DetectionResult audio_result;
        if (audio_detector_ && compression_result.success) {
            audio_result = audio_detector_->detectVoiceSpoofing(
                compression_result.processed_data, sample_rate, channels);
        }
        
        // 3. 组合结果
        result.audio_result = audio_result;
        result.compression_result = compression_result;
        result.overall_confidence = audio_result.confidence;
        result.overall_risk_score = audio_result.risk_score;
        result.is_fake = audio_result.is_fake;
        result.compression_ratio = compression_result.compression_ratio;
        result.frame_count = 1;
        
        auto end_time = std::chrono::high_resolution_clock::now();
        result.total_processing_time_ms = std::chrono::duration_cast<std::chrono::milliseconds>(
            end_time - start_time).count();
        
        // 4. 生成检测摘要
        std::stringstream summary;
        summary << "Audio detection completed in " << result.total_processing_time_ms << "ms. ";
        summary << "Confidence: " << audio_result.confidence << ", ";
        summary << "Risk score: " << audio_result.risk_score << ", ";
        summary << "Compression ratio: " << compression_result.compression_ratio;
        result.detection_summary = summary.str();
        
        // 5. 记录性能统计
        if (performance_monitoring_enabled_ && performance_monitor_) {
            performance_monitor_->recordInferenceTime(audio_result.processing_time_ms);
        }
        
    } catch (const std::exception& e) {
        result.detection_summary = "Error: " + std::string(e.what());
    }
    
    return result;
}

IntegratedDetectionResult IntegrationService::combineResults(
    const onnx_detector::DetectionResult& video_result,
    const onnx_detector::DetectionResult& audio_result,
    const ffmpeg_service::ProcessingResult& compression_result) {
    IntegratedDetectionResult result;
    
    try {
        // 加权组合视频和音频结果
        result.video_result = video_result;
        result.audio_result = audio_result;
        result.compression_result = compression_result;
        
        result.overall_confidence = current_config_.video_weight * video_result.confidence +
                                  current_config_.audio_weight * audio_result.confidence;
        
        result.overall_risk_score = current_config_.video_weight * video_result.risk_score +
                                  current_config_.audio_weight * audio_result.risk_score;
        
        // 判断是否为伪造
        result.is_fake = result.overall_confidence > current_config_.confidence_threshold ||
                        result.overall_risk_score > current_config_.risk_threshold;
        
        result.compression_ratio = compression_result.compression_ratio;
        result.frame_count = 1;
        
        // 合并详细指标
        result.detailed_metrics["video_confidence"] = video_result.confidence;
        result.detailed_metrics["audio_confidence"] = audio_result.confidence;
        result.detailed_metrics["video_risk_score"] = video_result.risk_score;
        result.detailed_metrics["audio_risk_score"] = audio_result.risk_score;
        result.detailed_metrics["compression_ratio"] = compression_result.compression_ratio;
        result.detailed_metrics["overall_confidence"] = result.overall_confidence;
        result.detailed_metrics["overall_risk_score"] = result.overall_risk_score;
        
    } catch (const std::exception& e) {
        result.detection_summary = "Error combining results: " + std::string(e.what());
    }
    
    return result;
}

void IntegrationService::realTimeProcessingLoop() {
    while (!should_stop_) {
        try {
            // 实时处理逻辑
            // 这里可以从队列中获取帧数据进行处理
            
            std::this_thread::sleep_for(std::chrono::milliseconds(33)); // 30 FPS
            
        } catch (const std::exception& e) {
            std::cerr << "Real-time processing error: " << e.what() << std::endl;
        }
    }
}

void IntegrationService::handleRealTimeVideoFrame(const ffmpeg_service::MediaFrame& frame) {
    if (!detection_callback_) {
        return;
    }
    
    try {
        // 处理实时视频帧
        IntegratedDetectionResult result = detectVideo(frame.data, frame.width, frame.height, 30);
        
        if (detection_callback_) {
            detection_callback_(result);
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Real-time video frame processing error: " << e.what() << std::endl;
    }
}

void IntegrationService::handleRealTimeAudioFrame(const ffmpeg_service::MediaFrame& frame) {
    if (!detection_callback_) {
        return;
    }
    
    try {
        // 处理实时音频帧
        IntegratedDetectionResult result = detectAudio(frame.data, frame.sample_rate, frame.channels);
        
        if (detection_callback_) {
            detection_callback_(result);
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Real-time audio frame processing error: " << e.what() << std::endl;
    }
}

void IntegrationService::streamingProcessingLoop() {
    while (!should_stop_) {
        try {
            // 流式处理逻辑
            // 这里可以从流中读取数据进行处理
            
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
            
        } catch (const std::exception& e) {
            std::cerr << "Streaming processing error: " << e.what() << std::endl;
        }
    }
}

void IntegrationService::updateFeatureCache(const std::string& key, const std::vector<float>& features) {
    if (!current_config_.enable_feature_cache) {
        return;
    }
    
    std::lock_guard<std::mutex> lock(cache_mutex_);
    
    auto now = std::chrono::system_clock::now().time_since_epoch().count();
    
    // 检查缓存大小
    if (feature_cache_.size() >= current_config_.cache_size) {
        // 移除最旧的条目
        if (!cache_keys_.empty()) {
            std::string oldest_key = cache_keys_.front();
            cache_keys_.pop();
            feature_cache_.erase(oldest_key);
        }
    }
    
    // 添加新条目
    feature_cache_[key] = std::make_pair(features, now);
    cache_keys_.push(key);
}

bool IntegrationService::getCachedFeatures(const std::string& key, std::vector<float>& features) {
    if (!current_config_.enable_feature_cache) {
        return false;
    }
    
    std::lock_guard<std::mutex> lock(cache_mutex_);
    
    auto it = feature_cache_.find(key);
    if (it != feature_cache_.end()) {
        features = it->second.first;
        return true;
    }
    
    return false;
}

void IntegrationService::cleanupExpiredCache() {
    if (!current_config_.enable_feature_cache) {
        return;
    }
    
    std::lock_guard<std::mutex> lock(cache_mutex_);
    
    auto now = std::chrono::system_clock::now().time_since_epoch().count();
    auto ttl_ns = static_cast<int64_t>(current_config_.cache_ttl_seconds) * 1000000000LL;
    
    std::vector<std::string> expired_keys;
    
    for (const auto& entry : feature_cache_) {
        if (now - entry.second.second > ttl_ns) {
            expired_keys.push_back(entry.first);
        }
    }
    
    for (const auto& key : expired_keys) {
        feature_cache_.erase(key);
    }
}

// ServiceManager 实现
ServiceManager::ServiceManager() : service_running_(false) {
}

ServiceManager::~ServiceManager() {
    stopService();
}

bool ServiceManager::startService(const IntegrationConfig& config) {
    if (service_running_) {
        return true;
    }
    
    try {
        std::lock_guard<std::mutex> lock(service_mutex_);
        
        integration_service_ = std::make_unique<IntegrationService>();
        config_ = config;
        
        if (!integration_service_->initialize(config)) {
            return false;
        }
        
        service_running_ = true;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Service start error: " << e.what() << std::endl;
        return false;
    }
}

void ServiceManager::stopService() {
    if (!service_running_) {
        return;
    }
    
    try {
        std::lock_guard<std::mutex> lock(service_mutex_);
        
        if (integration_service_) {
            integration_service_->cleanup();
        }
        
        service_running_ = false;
        
    } catch (const std::exception& e) {
        std::cerr << "Service stop error: " << e.what() << std::endl;
    }
}

void ServiceManager::restartService() {
    IntegrationConfig current_config = config_;
    stopService();
    startService(current_config);
}

std::string ServiceManager::getServiceStatus() const {
    if (!service_running_) {
        return "Stopped";
    }
    
    if (!integration_service_) {
        return "Error";
    }
    
    return integration_service_->getServiceStatus();
}

bool ServiceManager::loadConfigFromFile(const std::string& config_file) {
    try {
        // 这里实现从文件加载配置的逻辑
        // 可以使用JSON、YAML等格式
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Config load error: " << e.what() << std::endl;
        return false;
    }
}

bool ServiceManager::saveConfigToFile(const std::string& config_file) {
    try {
        // 这里实现保存配置到文件的逻辑
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Config save error: " << e.what() << std::endl;
        return false;
    }
}

bool ServiceManager::performHealthCheck() {
    if (!service_running_ || !integration_service_) {
        return false;
    }
    
    try {
        // 执行健康检查
        auto component_status = getComponentStatus();
        
        for (const auto& status : component_status) {
            if (!status.second) {
                return false;
            }
        }
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Health check error: " << e.what() << std::endl;
        return false;
    }
}

std::unordered_map<std::string, bool> ServiceManager::getComponentStatus() {
    std::unordered_map<std::string, bool> status;
    
    if (!integration_service_) {
        status["integration_service"] = false;
        status["ffmpeg_processor"] = false;
        status["video_detector"] = false;
        status["audio_detector"] = false;
        return status;
    }
    
    status["integration_service"] = integration_service_->isInitialized();
    status["ffmpeg_processor"] = true; // 需要从FFmpeg处理器获取状态
    status["video_detector"] = true;   // 需要从视频检测器获取状态
    status["audio_detector"] = true;   // 需要从音频检测器获取状态
    
    return status;
}

// 工具函数实现
namespace utils {

bool validateConfig(const IntegrationConfig& config) {
    try {
        // 验证FFmpeg参数
        if (config.ffmpeg_params.video_bitrate <= 0 || config.ffmpeg_params.audio_bitrate <= 0) {
            return false;
        }
        
        if (config.ffmpeg_params.video_width <= 0 || config.ffmpeg_params.video_height <= 0) {
            return false;
        }
        
        // 验证检测参数
        if (config.confidence_threshold < 0.0f || config.confidence_threshold > 1.0f) {
            return false;
        }
        
        if (config.risk_threshold < 0.0f || config.risk_threshold > 1.0f) {
            return false;
        }
        
        // 验证权重
        if (config.video_weight < 0.0f || config.audio_weight < 0.0f) {
            return false;
        }
        
        if (std::abs(config.video_weight + config.audio_weight - 1.0f) > 0.01f) {
            return false;
        }
        
        return true;
    } catch (const std::exception& e) {
        std::cerr << "Config validation error: " << e.what() << std::endl;
        return false;
    }
}

void optimizeForPlatform(IntegrationConfig& config) {
    try {
        // 根据平台优化配置
        // 例如：检测CPU核心数、GPU可用性等
        
        // 设置线程数
        int num_cores = std::thread::hardware_concurrency();
        if (num_cores > 0) {
            config.processing_threads = std::min(num_cores, 8); // 最多8个线程
        }
        
        // 检测GPU可用性
        // 这里可以添加GPU检测逻辑
        
    } catch (const std::exception& e) {
        std::cerr << "Platform optimization error: " << e.what() << std::endl;
    }
}

void logDetectionResult(const IntegratedDetectionResult& result) {
    try {
        std::cout << "Detection Result:" << std::endl;
        std::cout << "  Is Fake: " << (result.is_fake ? "Yes" : "No") << std::endl;
        std::cout << "  Overall Confidence: " << result.overall_confidence << std::endl;
        std::cout << "  Overall Risk Score: " << result.overall_risk_score << std::endl;
        std::cout << "  Processing Time: " << result.total_processing_time_ms << "ms" << std::endl;
        std::cout << "  Compression Ratio: " << result.compression_ratio << std::endl;
        std::cout << "  Summary: " << result.detection_summary << std::endl;
    } catch (const std::exception& e) {
        std::cerr << "Logging error: " << e.what() << std::endl;
    }
}

void logPerformanceStats(const std::unordered_map<std::string, double>& stats) {
    try {
        std::cout << "Performance Statistics:" << std::endl;
        for (const auto& stat : stats) {
            std::cout << "  " << stat.first << ": " << stat.second << std::endl;
        }
    } catch (const std::exception& e) {
        std::cerr << "Performance logging error: " << e.what() << std::endl;
    }
}

std::vector<uint8_t> convertVideoFormat(const std::vector<uint8_t>& data,
                                       int width, int height,
                                       ffmpeg_service::EncodingParams& params) {
    std::vector<uint8_t> converted_data;
    
    try {
        // 这里实现视频格式转换逻辑
        // 可以使用FFmpeg进行转换
        
        converted_data = data; // 临时实现
        
    } catch (const std::exception& e) {
        std::cerr << "Video format conversion error: " << e.what() << std::endl;
    }
    
    return converted_data;
}

std::vector<uint8_t> convertAudioFormat(const std::vector<uint8_t>& data,
                                       int sample_rate, int channels,
                                       ffmpeg_service::EncodingParams& params) {
    std::vector<uint8_t> converted_data;
    
    try {
        // 这里实现音频格式转换逻辑
        // 可以使用FFmpeg进行转换
        
        converted_data = data; // 临时实现
        
    } catch (const std::exception& e) {
        std::cerr << "Audio format conversion error: " << e.what() << std::endl;
    }
    
    return converted_data;
}

} // namespace utils

} // namespace integration_service 