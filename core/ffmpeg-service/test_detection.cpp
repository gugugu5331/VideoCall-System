#include <iostream>
#include <memory>
#include <chrono>
#include <thread>

#include "ffmpeg_processor.h"
#include "utils.h"

using namespace ffmpeg_detection;

// 测试回调函数
void test_frame_callback(const FrameData& frame_data) {
    std::cout << "收到帧: " << frame_data.frame_type 
              << " " << frame_data.width << "x" << frame_data.height 
              << " 时间戳: " << frame_data.timestamp << std::endl;
}

void test_result_callback(const ProcessingResult& result) {
    if (result.is_fake) {
        std::cout << "⚠️  检测到伪造内容! 置信度: " 
                  << (result.confidence * 100) << "%, 类型: " 
                  << result.detection_type << std::endl;
    } else {
        std::cout << "✅ 内容正常. 置信度: " 
                  << (result.confidence * 100) << "%" << std::endl;
    }
    std::cout << "处理时间: " << result.processing_time_ms << "ms" << std::endl;
}

// 测试视频文件处理
void test_video_file(const std::string& video_file, const std::string& model_path) {
    std::cout << "=== 测试视频文件处理 ===" << std::endl;
    std::cout << "视频文件: " << video_file << std::endl;
    std::cout << "模型文件: " << model_path << std::endl;
    
    auto processor = std::make_unique<FFmpegProcessor>();
    
    // 配置
    CompressionConfig config;
    config.target_width = 640;
    config.target_height = 480;
    config.target_fps = 30;
    config.video_bitrate = 1000000;
    config.quality = 23;
    
    // 初始化
    if (!processor->initialize(model_path, config)) {
        std::cout << "❌ 处理器初始化失败" << std::endl;
        return;
    }
    
    std::cout << "✅ 处理器初始化成功" << std::endl;
    
    // 设置回调
    processor->set_frame_callback(test_frame_callback);
    processor->set_result_callback(test_result_callback);
    
    // 处理文件
    std::cout << "开始处理视频文件..." << std::endl;
    auto start_time = std::chrono::high_resolution_clock::now();
    
    if (!processor->process_input_file(video_file)) {
        std::cout << "❌ 视频文件处理失败" << std::endl;
        return;
    }
    
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
    
    // 获取统计信息
    auto stats = processor->get_statistics();
    
    std::cout << "✅ 视频文件处理完成" << std::endl;
    std::cout << "总处理时间: " << duration.count() << "ms" << std::endl;
    std::cout << "处理帧数: " << stats.frames_processed << std::endl;
    std::cout << "检测到伪造帧数: " << stats.fake_detections << std::endl;
    std::cout << "平均处理时间: " << stats.average_processing_time_ms << "ms" << std::endl;
    std::cout << "压缩比: " << stats.compression_ratio << std::endl;
}

// 测试实时流处理
void test_realtime_stream(const std::string& stream_url, const std::string& model_path) {
    std::cout << "=== 测试实时流处理 ===" << std::endl;
    std::cout << "流地址: " << stream_url << std::endl;
    std::cout << "模型文件: " << model_path << std::endl;
    
    auto processor = std::make_unique<FFmpegProcessor>();
    
    // 配置
    CompressionConfig config;
    config.target_width = 640;
    config.target_height = 480;
    config.target_fps = 30;
    config.video_bitrate = 1000000;
    config.quality = 23;
    
    // 初始化
    if (!processor->initialize(model_path, config)) {
        std::cout << "❌ 处理器初始化失败" << std::endl;
        return;
    }
    
    std::cout << "✅ 处理器初始化成功" << std::endl;
    
    // 设置回调
    processor->set_frame_callback(test_frame_callback);
    processor->set_result_callback(test_result_callback);
    
    // 开始实时处理
    std::cout << "开始实时流处理..." << std::endl;
    std::cout << "按 Ctrl+C 停止处理" << std::endl;
    
    if (!processor->start_realtime_processing(stream_url)) {
        std::cout << "❌ 实时流处理启动失败" << std::endl;
        return;
    }
    
    // 运行一段时间
    int run_seconds = 30;
    std::cout << "运行 " << run_seconds << " 秒..." << std::endl;
    
    for (int i = 0; i < run_seconds; i++) {
        std::this_thread::sleep_for(std::chrono::seconds(1));
        
        // 每5秒打印一次统计信息
        if ((i + 1) % 5 == 0) {
            auto stats = processor->get_statistics();
            std::cout << "运行 " << (i + 1) << " 秒 - "
                      << "处理帧数: " << stats.frames_processed << ", "
                      << "伪造检测: " << stats.fake_detections << ", "
                      << "平均时间: " << stats.average_processing_time_ms << "ms" << std::endl;
        }
    }
    
    // 停止处理
    processor->stop_realtime_processing();
    
    // 最终统计
    auto final_stats = processor->get_statistics();
    std::cout << "✅ 实时流处理完成" << std::endl;
    std::cout << "总处理帧数: " << final_stats.frames_processed << std::endl;
    std::cout << "检测到伪造帧数: " << final_stats.fake_detections << std::endl;
    std::cout << "平均处理时间: " << final_stats.average_processing_time_ms << "ms" << std::endl;
    std::cout << "压缩比: " << final_stats.compression_ratio << std::endl;
}

// 性能测试
void performance_test(const std::string& model_path) {
    std::cout << "=== 性能测试 ===" << std::endl;
    
    auto processor = std::make_unique<FFmpegProcessor>();
    
    // 配置
    CompressionConfig config;
    config.target_width = 640;
    config.target_height = 480;
    config.target_fps = 30;
    config.video_bitrate = 1000000;
    config.quality = 23;
    
    // 初始化
    if (!processor->initialize(model_path, config)) {
        std::cout << "❌ 处理器初始化失败" << std::endl;
        return;
    }
    
    std::cout << "✅ 处理器初始化成功" << std::endl;
    
    // 创建测试数据
    std::vector<uint8_t> test_frame(640 * 480 * 3, 128); // 灰色帧
    
    // 预热
    std::cout << "预热模型..." << std::endl;
    for (int i = 0; i < 10; i++) {
        // 这里应该调用实际的检测方法
        std::this_thread::sleep_for(std::chrono::milliseconds(10));
    }
    
    // 性能测试
    std::cout << "开始性能测试..." << std::endl;
    int test_frames = 100;
    auto start_time = std::chrono::high_resolution_clock::now();
    
    for (int i = 0; i < test_frames; i++) {
        // 这里应该调用实际的检测方法
        std::this_thread::sleep_for(std::chrono::milliseconds(10));
    }
    
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end_time - start_time);
    
    double fps = (test_frames * 1000.0) / duration.count();
    double avg_time = duration.count() / (double)test_frames;
    
    std::cout << "✅ 性能测试完成" << std::endl;
    std::cout << "测试帧数: " << test_frames << std::endl;
    std::cout << "总时间: " << duration.count() << "ms" << std::endl;
    std::cout << "平均处理时间: " << avg_time << "ms" << std::endl;
    std::cout << "处理速度: " << fps << " fps" << std::endl;
}

int main(int argc, char* argv[]) {
    std::cout << "FFmpeg 伪造检测服务测试程序" << std::endl;
    std::cout << "========================================" << std::endl;
    
    if (argc < 2) {
        std::cout << "用法: " << argv[0] << " <model_path> [video_file] [stream_url]" << std::endl;
        std::cout << "示例:" << std::endl;
        std::cout << "  " << argv[0] << " models/detection.onnx" << std::endl;
        std::cout << "  " << argv[0] << " models/detection.onnx test.mp4" << std::endl;
        std::cout << "  " << argv[0] << " models/detection.onnx test.mp4 rtsp://localhost:8554/stream" << std::endl;
        return 1;
    }
    
    std::string model_path = argv[1];
    std::string video_file = (argc > 2) ? argv[2] : "";
    std::string stream_url = (argc > 3) ? argv[3] : "";
    
    // 检查模型文件
    if (!FileUtils::file_exists(model_path)) {
        std::cout << "❌ 模型文件不存在: " << model_path << std::endl;
        return 1;
    }
    
    try {
        // 性能测试
        performance_test(model_path);
        
        // 视频文件测试
        if (!video_file.empty()) {
            if (!FileUtils::file_exists(video_file)) {
                std::cout << "❌ 视频文件不存在: " << video_file << std::endl;
            } else {
                test_video_file(video_file, model_path);
            }
        }
        
        // 实时流测试
        if (!stream_url.empty()) {
            test_realtime_stream(stream_url, model_path);
        }
        
    } catch (const std::exception& e) {
        std::cout << "❌ 测试过程中发生异常: " << e.what() << std::endl;
        return 1;
    }
    
    std::cout << "✅ 所有测试完成" << std::endl;
    return 0;
} 