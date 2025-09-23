#include "video_processor.h"
#include <iostream>
#include <signal.h>

using namespace VideoProcessing;

// 全局变量用于信号处理
VideoProcessor* g_processor = nullptr;

// 信号处理函数
void SignalHandler(int signal) {
    std::cout << "\nReceived signal " << signal << ", shutting down gracefully..." << std::endl;
    if (g_processor) {
        g_processor->Stop();
    }
}

// 打印使用说明
void PrintUsage(const char* program_name) {
    std::cout << "Usage: " << program_name << " [options]" << std::endl;
    std::cout << "Options:" << std::endl;
    std::cout << "  -h, --help              Show this help message" << std::endl;
    std::cout << "  -c, --camera <id>       Camera device ID (default: 0)" << std::endl;
    std::cout << "  -w, --width <width>     Window width (default: 1280)" << std::endl;
    std::cout << "  -h, --height <height>   Window height (default: 720)" << std::endl;
    std::cout << "  -f, --fullscreen        Start in fullscreen mode" << std::endl;
    std::cout << "  -n, --no-ui             Hide UI controls" << std::endl;
    std::cout << "  --fps <fps>             Target FPS (default: 30)" << std::endl;
    std::cout << "  --no-vsync              Disable VSync" << std::endl;
    std::cout << "  --msaa <samples>        MSAA samples (default: 4)" << std::endl;
    std::cout << std::endl;
    std::cout << "Controls:" << std::endl;
    std::cout << "  ESC                     Exit application" << std::endl;
    std::cout << "  SPACE                   Take screenshot" << std::endl;
    std::cout << "  R                       Start/stop recording" << std::endl;
    std::cout << "  F                       Toggle fullscreen" << std::endl;
    std::cout << "  U                       Toggle UI visibility" << std::endl;
    std::cout << "  1-9                     Apply different filters" << std::endl;
    std::cout << "  0                       Remove all filters" << std::endl;
    std::cout << "  M                       Toggle mirror mode" << std::endl;
    std::cout << "  D                       Toggle face detection" << std::endl;
    std::cout << "  B                       Toggle beauty mode" << std::endl;
    std::cout << "  C                       Toggle cartoon mode" << std::endl;
    std::cout << "  S                       Toggle sketch mode" << std::endl;
}

// 解析命令行参数
struct CommandLineArgs {
    int camera_id = 0;
    int window_width = WINDOW_WIDTH;
    int window_height = WINDOW_HEIGHT;
    bool fullscreen = false;
    bool show_ui = true;
    int target_fps = 30;
    bool vsync = true;
    int msaa_samples = 4;
    bool show_help = false;
};

CommandLineArgs ParseCommandLine(int argc, char* argv[]) {
    CommandLineArgs args;
    
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        
        if (arg == "-h" || arg == "--help") {
            args.show_help = true;
        } else if (arg == "-c" || arg == "--camera") {
            if (i + 1 < argc) {
                args.camera_id = std::atoi(argv[++i]);
            }
        } else if (arg == "-w" || arg == "--width") {
            if (i + 1 < argc) {
                args.window_width = std::atoi(argv[++i]);
            }
        } else if (arg == "--height") {
            if (i + 1 < argc) {
                args.window_height = std::atoi(argv[++i]);
            }
        } else if (arg == "-f" || arg == "--fullscreen") {
            args.fullscreen = true;
        } else if (arg == "-n" || arg == "--no-ui") {
            args.show_ui = false;
        } else if (arg == "--fps") {
            if (i + 1 < argc) {
                args.target_fps = std::atoi(argv[++i]);
            }
        } else if (arg == "--no-vsync") {
            args.vsync = false;
        } else if (arg == "--msaa") {
            if (i + 1 < argc) {
                args.msaa_samples = std::atoi(argv[++i]);
            }
        }
    }
    
    return args;
}

// 设置键盘回调
void SetupKeyboardCallbacks(VideoProcessor& processor) {
    processor.SetKeyCallback([&processor](int key, int scancode, int action, int mods) {
        if (action == GLFW_PRESS) {
            switch (key) {
                case GLFW_KEY_ESCAPE:
                    processor.Stop();
                    break;
                case GLFW_KEY_SPACE:
                    processor.TakeScreenshot("screenshot.png");
                    std::cout << "Screenshot saved!" << std::endl;
                    break;
                case GLFW_KEY_R:
                    if (processor.IsRecording()) {
                        processor.StopRecording();
                        std::cout << "Recording stopped!" << std::endl;
                    } else {
                        processor.StartRecording("recording.mp4");
                        std::cout << "Recording started!" << std::endl;
                    }
                    break;
                case GLFW_KEY_F:
                    // Toggle fullscreen (implementation needed)
                    break;
                case GLFW_KEY_U:
                    processor.ShowUI(!processor.IsUIVisible());
                    break;
                case GLFW_KEY_1:
                    processor.SetFilter(FilterType::BLUR);
                    break;
                case GLFW_KEY_2:
                    processor.SetFilter(FilterType::SHARPEN);
                    break;
                case GLFW_KEY_3:
                    processor.SetFilter(FilterType::EDGE_DETECTION);
                    break;
                case GLFW_KEY_4:
                    processor.SetFilter(FilterType::SEPIA);
                    break;
                case GLFW_KEY_5:
                    processor.SetFilter(FilterType::VINTAGE);
                    break;
                case GLFW_KEY_6:
                    processor.SetFilter(FilterType::CARTOON);
                    break;
                case GLFW_KEY_7:
                    processor.SetFilter(FilterType::SKETCH);
                    break;
                case GLFW_KEY_8:
                    processor.SetFilter(FilterType::NEON);
                    break;
                case GLFW_KEY_9:
                    processor.SetFilter(FilterType::THERMAL);
                    break;
                case GLFW_KEY_0:
                    processor.SetFilter(FilterType::NONE);
                    break;
                case GLFW_KEY_M:
                    // Toggle mirror mode (implementation needed)
                    break;
                case GLFW_KEY_D:
                    processor.EnableFaceDetection(!processor.IsFaceDetectionEnabled());
                    break;
                case GLFW_KEY_B:
                    processor.EnableBeautyMode(true);
                    break;
                case GLFW_KEY_C:
                    processor.EnableCartoonMode(true);
                    break;
                case GLFW_KEY_S:
                    processor.EnableSketchMode(true);
                    break;
            }
        }
    });
}

int main(int argc, char* argv[]) {
    // 解析命令行参数
    CommandLineArgs args = ParseCommandLine(argc, argv);
    
    if (args.show_help) {
        PrintUsage(argv[0]);
        return 0;
    }
    
    // 设置信号处理
    signal(SIGINT, SignalHandler);
    signal(SIGTERM, SignalHandler);
    
    std::cout << "=== Video Processing with OpenCV + OpenGL ===" << std::endl;
    std::cout << "Initializing..." << std::endl;
    
    try {
        // 创建视频处理器
        VideoProcessor processor;
        g_processor = &processor;
        
        // 初始化
        if (!processor.Initialize(args.window_width, args.window_height)) {
            std::cerr << "Failed to initialize video processor!" << std::endl;
            return -1;
        }
        
        // 设置配置
        VideoProcessor::Settings settings;
        settings.fullscreen = args.fullscreen;
        settings.target_fps = args.target_fps;
        settings.vsync = args.vsync;
        settings.msaa_samples = args.msaa_samples;
        processor.SetSettings(settings);
        
        // 显示/隐藏UI
        processor.ShowUI(args.show_ui);
        
        // 设置键盘回调
        SetupKeyboardCallbacks(processor);
        
        // 启动摄像头
        if (!processor.StartCamera(args.camera_id)) {
            std::cerr << "Failed to start camera " << args.camera_id << "!" << std::endl;
            return -1;
        }
        
        // 启用面部检测
        processor.EnableFaceDetection(true);
        
        std::cout << "Initialization complete!" << std::endl;
        std::cout << "Camera: " << args.camera_id << std::endl;
        std::cout << "Window: " << args.window_width << "x" << args.window_height << std::endl;
        std::cout << "Press 'H' for help, 'ESC' to exit" << std::endl;
        
        // 运行主循环
        processor.Run();
        
        std::cout << "Shutting down..." << std::endl;
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return -1;
    } catch (...) {
        std::cerr << "Unknown error occurred!" << std::endl;
        return -1;
    }
    
    std::cout << "Goodbye!" << std::endl;
    return 0;
}
