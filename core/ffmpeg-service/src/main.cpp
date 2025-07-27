#include <iostream>
#include <signal.h>
#include <thread>
#include <chrono>
#include <memory>

#include "ffmpeg_processor.h"
#include "utils.h"

using namespace ffmpeg_detection;

// 全局变量
std::unique_ptr<FFmpegProcessor> g_processor;
std::atomic<bool> g_running(true);

// 信号处理函数
void signal_handler(int signal) {
    LOG_INFO("收到信号 %d，正在关闭服务...", signal);
    g_running = false;
    
    if (g_processor) {
        g_processor->stop_realtime_processing();
    }
}

// 帧处理回调
void on_frame_processed(const FrameData& frame_data) {
    LOG_DEBUG("处理帧: 类型=%s, 大小=%dx%d, 时间戳=%lld", 
              frame_data.frame_type.c_str(), 
              frame_data.width, 
              frame_data.height, 
              frame_data.timestamp);
}

// 检测结果回调
void on_detection_result(const ProcessingResult& result) {
    if (result.is_fake) {
        LOG_WARNING("检测到伪造内容! 置信度: %.2f%%, 类型: %s, 处理时间: %lldms", 
                   result.confidence * 100, 
                   result.detection_type.c_str(), 
                   result.processing_time_ms);
    } else {
        LOG_INFO("内容正常. 置信度: %.2f%%, 处理时间: %lldms", 
                result.confidence * 100, 
                result.processing_time_ms);
    }
}

// 打印使用说明
void print_usage(const char* program_name) {
    std::cout << "FFmpeg 伪造检测服务\n";
    std::cout << "用法: " << program_name << " [选项]\n\n";
    std::cout << "选项:\n";
    std::cout << "  -i, --input <url/file>     输入流或文件路径\n";
    std::cout << "  -m, --model <path>         模型文件路径\n";
    std::cout << "  -c, --config <file>        配置文件路径\n";
    std::cout << "  -o, --output <file>        输出日志文件\n";
    std::cout << "  -v, --verbose              详细输出\n";
    std::cout << "  -h, --help                 显示此帮助信息\n\n";
    std::cout << "示例:\n";
    std::cout << "  " << program_name << " -i rtsp://192.168.1.100:554/stream -m models/detection.onnx\n";
    std::cout << "  " << program_name << " -i video.mp4 -m models/detection.onnx -c config.json\n";
}

// 解析命令行参数
struct CommandLineArgs {
    std::string input_url;
    std::string model_path;
    std::string config_file;
    std::string output_log;
    bool verbose = false;
};

CommandLineArgs parse_arguments(int argc, char* argv[]) {
    CommandLineArgs args;
    
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        
        if (arg == "-h" || arg == "--help") {
            print_usage(argv[0]);
            exit(0);
        } else if (arg == "-i" || arg == "--input") {
            if (i + 1 < argc) {
                args.input_url = argv[++i];
            }
        } else if (arg == "-m" || arg == "--model") {
            if (i + 1 < argc) {
                args.model_path = argv[++i];
            }
        } else if (arg == "-c" || arg == "--config") {
            if (i + 1 < argc) {
                args.config_file = argv[++i];
            }
        } else if (arg == "-o" || arg == "--output") {
            if (i + 1 < argc) {
                args.output_log = argv[++i];
            }
        } else if (arg == "-v" || arg == "--verbose") {
            args.verbose = true;
        }
    }
    
    return args;
}

// 加载配置
CompressionConfig load_config(const std::string& config_file) {
    CompressionConfig config;
    
    if (!FileUtils::file_exists(config_file)) {
        LOG_WARNING("配置文件不存在: %s，使用默认配置", config_file.c_str());
        return config;
    }
    
    std::unordered_map<std::string, std::string> config_map;
    if (ConfigUtils::load_config(config_file, config_map)) {
        config.target_width = ConfigUtils::get_config_value_int(config_map, "target_width", config.target_width);
        config.target_height = ConfigUtils::get_config_value_int(config_map, "target_height", config.target_height);
        config.target_fps = ConfigUtils::get_config_value_int(config_map, "target_fps", config.target_fps);
        config.video_bitrate = ConfigUtils::get_config_value_int(config_map, "video_bitrate", config.video_bitrate);
        config.audio_bitrate = ConfigUtils::get_config_value_int(config_map, "audio_bitrate", config.audio_bitrate);
        config.video_codec = ConfigUtils::get_config_value(config_map, "video_codec", config.video_codec);
        config.audio_codec = ConfigUtils::get_config_value(config_map, "audio_codec", config.audio_codec);
        config.quality = ConfigUtils::get_config_value_int(config_map, "quality", config.quality);
        
        LOG_INFO("从配置文件加载配置: %s", config_file.c_str());
    }
    
    return config;
}

// 打印系统信息
void print_system_info() {
    LOG_INFO("=== 系统信息 ===");
    LOG_INFO("CPU 核心数: %d", ThreadUtils::get_cpu_count());
    LOG_INFO("可用内存: %s", StringUtils::format_bytes(MemoryUtils::get_available_memory_mb() * 1024 * 1024).c_str());
    LOG_INFO("当前内存使用: %s", StringUtils::format_bytes(MemoryUtils::get_current_memory_usage_mb() * 1024 * 1024).c_str());
    
    // 检查FFmpeg版本
    LOG_INFO("FFmpeg 版本: %s", av_version_info());
    
    // 检查ONNX Runtime
    LOG_INFO("ONNX Runtime 版本: %s", ORT_API_VERSION);
}

// 主函数
int main(int argc, char* argv[]) {
    // 解析命令行参数
    auto args = parse_arguments(argc, argv);
    
    // 设置日志
    if (!args.output_log.empty()) {
        Logger::get_instance().set_output_file(args.output_log);
    }
    
    if (args.verbose) {
        Logger::get_instance().set_level(LogLevel::DEBUG);
    }
    
    LOG_INFO("=== FFmpeg 伪造检测服务启动 ===");
    print_system_info();
    
    // 验证参数
    if (args.input_url.empty()) {
        LOG_ERROR("未指定输入源，请使用 -i 参数");
        print_usage(argv[0]);
        return 1;
    }
    
    if (args.model_path.empty()) {
        LOG_ERROR("未指定模型文件，请使用 -m 参数");
        print_usage(argv[0]);
        return 1;
    }
    
    if (!FileUtils::file_exists(args.model_path)) {
        LOG_ERROR("模型文件不存在: %s", args.model_path.c_str());
        return 1;
    }
    
    // 设置信号处理
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    
    try {
        // 创建处理器
        g_processor = std::make_unique<FFmpegProcessor>();
        
        // 加载配置
        CompressionConfig config;
        if (!args.config_file.empty()) {
            config = load_config(args.config_file);
        }
        
        // 初始化处理器
        LOG_INFO("正在初始化处理器...");
        if (!g_processor->initialize(args.model_path, config)) {
            LOG_ERROR("处理器初始化失败");
            return 1;
        }
        
        // 设置回调
        g_processor->set_frame_callback(on_frame_processed);
        g_processor->set_result_callback(on_detection_result);
        
        LOG_INFO("处理器初始化成功");
        LOG_INFO("开始处理输入: %s", args.input_url.c_str());
        
        // 开始实时处理
        if (!g_processor->start_realtime_processing(args.input_url)) {
            LOG_ERROR("启动实时处理失败");
            return 1;
        }
        
        // 主循环
        while (g_running) {
            std::this_thread::sleep_for(std::chrono::seconds(1));
            
            // 打印统计信息
            static int counter = 0;
            if (++counter % 30 == 0) { // 每30秒打印一次
                auto stats = g_processor->get_statistics();
                LOG_INFO("统计信息 - 处理帧数: %lld, 检测到伪造: %lld, 平均处理时间: %.2fms, 压缩比: %.2f", 
                        stats.frames_processed,
                        stats.fake_detections,
                        stats.average_processing_time_ms,
                        stats.compression_ratio);
                
                // 打印内存使用情况
                MemoryUtils::print_memory_info();
            }
        }
        
        LOG_INFO("正在停止服务...");
        g_processor->stop_realtime_processing();
        
        // 打印最终统计信息
        auto final_stats = g_processor->get_statistics();
        LOG_INFO("=== 最终统计信息 ===");
        LOG_INFO("总处理帧数: %lld", final_stats.frames_processed);
        LOG_INFO("检测到伪造帧数: %lld", final_stats.fake_detections);
        LOG_INFO("平均处理时间: %.2fms", final_stats.average_processing_time_ms);
        LOG_INFO("平均压缩比: %.2f", final_stats.compression_ratio);
        
    } catch (const std::exception& e) {
        LOG_ERROR("发生异常: %s", e.what());
        return 1;
    }
    
    LOG_INFO("服务已停止");
    return 0;
} 