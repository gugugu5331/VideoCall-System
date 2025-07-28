#include "../include/integration_service.h"
#include <iostream>
#include <fstream>
#include <chrono>
#include <thread>

using namespace integration_service;

// 示例回调函数
void detectionCallback(const IntegratedDetectionResult& result) {
    std::cout << "=== 检测结果 ===" << std::endl;
    std::cout << "是否为伪造: " << (result.is_fake ? "是" : "否") << std::endl;
    std::cout << "整体置信度: " << result.overall_confidence << std::endl;
    std::cout << "风险评分: " << result.overall_risk_score << std::endl;
    std::cout << "处理时间: " << result.total_processing_time_ms << "ms" << std::endl;
    std::cout << "压缩比: " << result.compression_ratio << std::endl;
    std::cout << "检测摘要: " << result.detection_summary << std::endl;
    std::cout << "==================" << std::endl;
}

void progressCallback(int progress, const std::string& status) {
    std::cout << "进度: " << progress << "% - " << status << std::endl;
}

// 生成测试数据
std::vector<uint8_t> generateTestVideoData(int width, int height) {
    std::vector<uint8_t> data(width * height * 3); // RGB格式
    for (size_t i = 0; i < data.size(); ++i) {
        data[i] = static_cast<uint8_t>(rand() % 256);
    }
    return data;
}

std::vector<uint8_t> generateTestAudioData(int sample_rate, int channels, float duration_seconds) {
    int num_samples = static_cast<int>(sample_rate * channels * duration_seconds);
    std::vector<uint8_t> data(num_samples * sizeof(float));
    
    float* float_data = reinterpret_cast<float*>(data.data());
    for (int i = 0; i < num_samples; ++i) {
        float_data[i] = static_cast<float>(rand()) / RAND_MAX * 2.0f - 1.0f;
    }
    
    return data;
}

int main() {
    std::cout << "=== FFmpeg服务 + ONNX检测器示例程序 ===" << std::endl;
    
    try {
        // 1. 配置集成服务
        IntegrationConfig config;
        
        // FFmpeg配置
        config.ffmpeg_params.video_bitrate = 1000000;  // 1Mbps
        config.ffmpeg_params.audio_bitrate = 128000;   // 128kbps
        config.ffmpeg_params.video_width = 1280;
        config.ffmpeg_params.video_height = 720;
        config.ffmpeg_params.video_fps = 30;
        config.ffmpeg_params.audio_sample_rate = 44100;
        config.ffmpeg_params.audio_channels = 2;
        
        // ONNX检测配置
        config.video_model_config.confidence_threshold = 0.8f;
        config.video_model_config.risk_threshold = 0.7f;
        config.video_model_config.enable_gpu = false;
        config.video_model_config.num_threads = 4;
        
        config.audio_model_config.confidence_threshold = 0.8f;
        config.audio_model_config.risk_threshold = 0.7f;
        config.audio_model_config.enable_gpu = false;
        config.audio_model_config.num_threads = 4;
        
        // 集成配置
        config.video_weight = 0.6f;
        config.audio_weight = 0.4f;
        config.confidence_threshold = 0.8f;
        config.risk_threshold = 0.7f;
        config.enable_compression = true;
        config.enable_real_time = true;
        config.enable_feature_cache = true;
        
        // 2. 创建并初始化集成服务
        std::cout << "正在初始化集成服务..." << std::endl;
        IntegrationService service;
        
        if (!service.initialize(config)) {
            std::cerr << "服务初始化失败!" << std::endl;
            return -1;
        }
        
        std::cout << "服务初始化成功!" << std::endl;
        
        // 3. 单次检测示例
        std::cout << "\n=== 单次检测示例 ===" << std::endl;
        
        // 视频检测
        std::cout << "执行视频检测..." << std::endl;
        auto video_data = generateTestVideoData(1280, 720);
        auto video_result = service.detectVideo(video_data, 1280, 720, 30);
        detectionCallback(video_result);
        
        // 音频检测
        std::cout << "执行音频检测..." << std::endl;
        auto audio_data = generateTestAudioData(44100, 2, 1.0f);
        auto audio_result = service.detectAudio(audio_data, 44100, 2);
        detectionCallback(audio_result);
        
        // 混合检测
        std::cout << "执行混合检测..." << std::endl;
        auto hybrid_result = service.detectHybrid(video_data, audio_data, 1280, 720, 30, 44100, 2);
        detectionCallback(hybrid_result);
        
        // 4. 批量检测示例
        std::cout << "\n=== 批量检测示例 ===" << std::endl;
        
        std::vector<std::vector<uint8_t>> video_batch;
        std::vector<std::vector<uint8_t>> audio_batch;
        
        // 生成测试数据
        for (int i = 0; i < 5; ++i) {
            video_batch.push_back(generateTestVideoData(1280, 720));
            audio_batch.push_back(generateTestAudioData(44100, 2, 1.0f));
        }
        
        // 批量视频检测
        std::cout << "执行批量视频检测..." << std::endl;
        auto video_batch_results = service.batchDetectVideo(video_batch, progressCallback);
        std::cout << "批量视频检测完成，共处理 " << video_batch_results.size() << " 个视频" << std::endl;
        
        // 批量音频检测
        std::cout << "执行批量音频检测..." << std::endl;
        auto audio_batch_results = service.batchDetectAudio(audio_batch, progressCallback);
        std::cout << "批量音频检测完成，共处理 " << audio_batch_results.size() << " 个音频" << std::endl;
        
        // 5. 实时检测示例
        std::cout << "\n=== 实时检测示例 ===" << std::endl;
        
        std::cout << "启动实时检测..." << std::endl;
        if (service.startRealTimeDetection(IntegratedDetectionType::REAL_TIME_VIDEO, detectionCallback)) {
            std::cout << "实时检测已启动，运行5秒..." << std::endl;
            
            // 模拟实时数据输入
            for (int i = 0; i < 10; ++i) {
                auto frame_data = generateTestVideoData(1280, 720);
                // 这里应该调用service的实时处理接口
                std::this_thread::sleep_for(std::chrono::milliseconds(500));
            }
            
            service.stopRealTimeDetection();
            std::cout << "实时检测已停止" << std::endl;
        }
        
        // 6. 性能监控示例
        std::cout << "\n=== 性能监控示例 ===" << std::endl;
        
        service.enablePerformanceMonitoring(true);
        
        // 执行一些检测操作
        for (int i = 0; i < 10; ++i) {
            auto test_data = generateTestVideoData(1280, 720);
            service.detectVideo(test_data, 1280, 720, 30);
        }
        
        // 获取性能统计
        std::unordered_map<std::string, double> stats;
        service.getPerformanceStats(stats);
        
        std::cout << "性能统计:" << std::endl;
        for (const auto& stat : stats) {
            std::cout << "  " << stat.first << ": " << stat.second << std::endl;
        }
        
        // 7. 配置管理示例
        std::cout << "\n=== 配置管理示例 ===" << std::endl;
        
        // 修改配置
        auto new_config = service.getCurrentConfig();
        new_config.confidence_threshold = 0.9f;
        new_config.risk_threshold = 0.8f;
        
        service.setIntegrationConfig(new_config);
        std::cout << "配置已更新" << std::endl;
        
        // 8. 服务状态查询
        std::cout << "\n=== 服务状态 ===" << std::endl;
        std::cout << "服务已初始化: " << (service.isInitialized() ? "是" : "否") << std::endl;
        std::cout << "正在处理: " << (service.isProcessing() ? "是" : "否") << std::endl;
        std::cout << "服务状态: " << service.getServiceStatus() << std::endl;
        
        // 9. 清理资源
        std::cout << "\n正在清理资源..." << std::endl;
        service.cleanup();
        std::cout << "资源清理完成" << std::endl;
        
        std::cout << "\n=== 示例程序执行完成 ===" << std::endl;
        
    } catch (const std::exception& e) {
        std::cerr << "程序执行错误: " << e.what() << std::endl;
        return -1;
    }
    
    return 0;
} 