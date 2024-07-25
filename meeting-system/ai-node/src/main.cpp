/*
 * SPDX-FileCopyrightText: 2024 Meeting System
 * SPDX-License-Identifier: MIT
 */

#include "MeetingAINode.h"
#include <iostream>
#include <fstream>
#include <signal.h>
#include <unistd.h>
#include <glog/glog.h>
#include "json.hpp"

using json = nlohmann::json;
using namespace MeetingAI;

// 全局变量
std::unique_ptr<MeetingAINode> g_ai_node;
volatile bool g_running = true;

// 信号处理函数
void signalHandler(int signal) {
    LOG(INFO) << "Received signal: " << signal;
    g_running = false;
    
    if (g_ai_node) {
        g_ai_node.reset();
    }
}

// 加载配置文件
json loadConfig(const std::string& config_path) {
    std::ifstream config_file(config_path);
    if (!config_file.is_open()) {
        LOG(ERROR) << "Failed to open config file: " << config_path;
        return json{};
    }
    
    json config;
    try {
        config_file >> config;
        LOG(INFO) << "Config loaded from: " << config_path;
    } catch (const std::exception& e) {
        LOG(ERROR) << "Failed to parse config file: " << e.what();
        return json{};
    }
    
    return config;
}

// 创建默认配置
json createDefaultConfig() {
    json config;
    
    // 基本配置
    config["unit_name"] = "meeting_ai_node_001";
    config["max_workers"] = 4;
    config["max_queue_size"] = 1000;
    config["model_base_path"] = "./models/";
    
    // 日志配置
    config["log"]["level"] = "INFO";
    config["log"]["log_dir"] = "./logs/";
    config["log"]["max_log_size"] = 100; // MB
    
    // 性能监控配置
    config["monitoring"]["enable"] = true;
    config["monitoring"]["report_interval"] = 60; // 秒
    
    // 模型配置
    config["models"]["speech_recognition"]["enabled"] = true;
    config["models"]["speech_recognition"]["model_path"] = "./models/speech_recognition.model";
    config["models"]["speech_recognition"]["max_audio_length"] = 30; // 秒
    
    config["models"]["emotion_detection"]["enabled"] = true;
    config["models"]["emotion_detection"]["model_path"] = "./models/emotion_detection.model";
    config["models"]["emotion_detection"]["max_image_size"] = 1920 * 1080;
    
    config["models"]["audio_denoising"]["enabled"] = true;
    config["models"]["audio_denoising"]["model_path"] = "./models/audio_denoising.model";
    config["models"]["audio_denoising"]["noise_threshold"] = 0.3;
    
    config["models"]["video_enhancement"]["enabled"] = true;
    config["models"]["video_enhancement"]["model_path"] = "./models/video_enhancement.model";
    config["models"]["video_enhancement"]["max_resolution"] = "1920x1080";
    
    return config;
}

// 保存配置文件
bool saveConfig(const json& config, const std::string& config_path) {
    std::ofstream config_file(config_path);
    if (!config_file.is_open()) {
        LOG(ERROR) << "Failed to create config file: " << config_path;
        return false;
    }
    
    try {
        config_file << config.dump(4);
        LOG(INFO) << "Config saved to: " << config_path;
        return true;
    } catch (const std::exception& e) {
        LOG(ERROR) << "Failed to save config file: " << e.what();
        return false;
    }
}

// 初始化日志系统
void initLogging(const json& config) {
    // 设置日志目录
    std::string log_dir = "./logs/";
    if (config.contains("log") && config["log"].contains("log_dir")) {
        log_dir = config["log"]["log_dir"];
    }
    
    // 创建日志目录
    system(("mkdir -p " + log_dir).c_str());
    
    // 初始化glog
    google::InitGoogleLogging("meeting-ai-node");
    google::SetLogDestination(google::GLOG_INFO, (log_dir + "info.log").c_str());
    google::SetLogDestination(google::GLOG_WARNING, (log_dir + "warning.log").c_str());
    google::SetLogDestination(google::GLOG_ERROR, (log_dir + "error.log").c_str());
    google::SetLogDestination(google::GLOG_FATAL, (log_dir + "fatal.log").c_str());
    
    // 设置日志级别
    if (config.contains("log") && config["log"].contains("level")) {
        std::string level = config["log"]["level"];
        if (level == "DEBUG") {
            FLAGS_minloglevel = 0;
        } else if (level == "INFO") {
            FLAGS_minloglevel = 0;
        } else if (level == "WARNING") {
            FLAGS_minloglevel = 1;
        } else if (level == "ERROR") {
            FLAGS_minloglevel = 2;
        }
    }
    
    // 设置日志文件大小限制
    FLAGS_max_log_size = 100; // MB
    FLAGS_stop_logging_if_full_disk = true;
    
    LOG(INFO) << "Logging system initialized";
}

// 性能监控线程
void performanceMonitorThread(const json& config) {
    if (!config.contains("monitoring") || !config["monitoring"]["enable"]) {
        return;
    }
    
    int report_interval = 60;
    if (config["monitoring"].contains("report_interval")) {
        report_interval = config["monitoring"]["report_interval"];
    }
    
    PerformanceMonitor monitor;
    
    while (g_running) {
        std::this_thread::sleep_for(std::chrono::seconds(report_interval));
        
        if (g_ai_node && g_running) {
            monitor.reportMetrics(*g_ai_node);
        }
    }
}

// 打印使用说明
void printUsage(const char* program_name) {
    std::cout << "Usage: " << program_name << " [options]\n"
              << "Options:\n"
              << "  -c, --config <file>    Configuration file path (default: ./config/ai_node_config.json)\n"
              << "  -h, --help            Show this help message\n"
              << "  -v, --version         Show version information\n"
              << "  --create-config       Create default configuration file\n"
              << std::endl;
}

// 打印版本信息
void printVersion() {
    std::cout << "Meeting AI Node v1.0.0\n"
              << "Built with Edge-LLM-Infra integration\n"
              << "Copyright (c) 2024 Meeting System\n"
              << std::endl;
}

int main(int argc, char* argv[]) {
    std::string config_path = "./config/ai_node_config.json";
    bool create_config = false;
    
    // 解析命令行参数
    for (int i = 1; i < argc; ++i) {
        std::string arg = argv[i];
        
        if (arg == "-h" || arg == "--help") {
            printUsage(argv[0]);
            return 0;
        } else if (arg == "-v" || arg == "--version") {
            printVersion();
            return 0;
        } else if (arg == "--create-config") {
            create_config = true;
        } else if ((arg == "-c" || arg == "--config") && i + 1 < argc) {
            config_path = argv[++i];
        } else {
            std::cerr << "Unknown option: " << arg << std::endl;
            printUsage(argv[0]);
            return 1;
        }
    }
    
    // 创建默认配置文件
    if (create_config) {
        json default_config = createDefaultConfig();
        if (saveConfig(default_config, config_path)) {
            std::cout << "Default configuration created: " << config_path << std::endl;
            return 0;
        } else {
            std::cerr << "Failed to create configuration file" << std::endl;
            return 1;
        }
    }
    
    // 加载配置
    json config = loadConfig(config_path);
    if (config.empty()) {
        std::cerr << "Failed to load configuration, creating default config..." << std::endl;
        config = createDefaultConfig();
        saveConfig(config, config_path);
    }
    
    // 初始化日志系统
    initLogging(config);
    
    // 注册信号处理器
    signal(SIGINT, signalHandler);
    signal(SIGTERM, signalHandler);
    
    LOG(INFO) << "Starting Meeting AI Node...";
    
    try {
        // 获取单元名称
        std::string unit_name = "meeting_ai_node_001";
        if (config.contains("unit_name")) {
            unit_name = config["unit_name"];
        }
        
        // 创建AI节点
        g_ai_node = std::make_unique<MeetingAINode>(unit_name);
        
        // 设置配置
        if (config.contains("max_workers")) {
            g_ai_node->setMaxWorkers(config["max_workers"]);
        }
        if (config.contains("max_queue_size")) {
            g_ai_node->setMaxQueueSize(config["max_queue_size"]);
        }
        if (config.contains("model_base_path")) {
            g_ai_node->setModelBasePath(config["model_base_path"]);
        }
        
        // 初始化AI节点
        if (g_ai_node->setup("", "", config.dump()) != 0) {
            LOG(ERROR) << "Failed to setup AI node";
            return 1;
        }
        
        // 启动性能监控线程
        std::thread monitor_thread(performanceMonitorThread, config);
        
        LOG(INFO) << "Meeting AI Node started successfully";
        
        // 主循环
        while (g_running) {
            std::this_thread::sleep_for(std::chrono::seconds(1));
        }
        
        // 等待监控线程结束
        if (monitor_thread.joinable()) {
            monitor_thread.join();
        }
        
        LOG(INFO) << "Meeting AI Node stopped";
        
    } catch (const std::exception& e) {
        LOG(ERROR) << "Exception in main: " << e.what();
        return 1;
    }
    
    // 清理glog
    google::ShutdownGoogleLogging();
    
    return 0;
}
