#pragma once

#include "common.h"
#include "camera_capture.h"
#include "opengl_renderer.h"
#include "filter_manager.h"
#include "face_detector.h"
#include "texture_manager.h"
#include "shader_manager.h"

namespace VideoProcessing {

class VideoProcessor {
public:
    VideoProcessor();
    ~VideoProcessor();

    // 初始化视频处理器
    bool Initialize(int window_width = WINDOW_WIDTH, int window_height = WINDOW_HEIGHT);
    
    // 释放资源
    void Release();
    
    // 主循环
    void Run();
    void Stop();
    bool IsRunning() const { return running_; }
    
    // 摄像头控制
    bool StartCamera(int camera_id = 0);
    void StopCamera();
    bool IsCameraActive() const { return camera_active_; }
    
    // 滤镜控制
    void SetFilter(FilterType filter_type);
    FilterType GetCurrentFilter() const;
    void SetFilterParams(const EffectParams& params);
    EffectParams GetFilterParams() const;
    
    // 贴图控制
    bool LoadSticker(const std::string& name, const std::string& file_path);
    void SetActiveSticker(const std::string& name);
    void RemoveSticker();
    bool HasActiveSticker() const { return !active_sticker_.empty(); }
    
    // 背景替换
    bool LoadBackground(const std::string& name, const std::string& file_path);
    void SetActiveBackground(const std::string& name);
    void RemoveBackground();
    bool HasActiveBackground() const { return !active_background_.empty(); }
    
    // 面部检测控制
    void EnableFaceDetection(bool enable);
    bool IsFaceDetectionEnabled() const { return face_detection_enabled_; }
    std::vector<FaceInfo> GetDetectedFaces() const { return detected_faces_; }
    
    // 实时效果
    void EnableBeautyMode(bool enable, float intensity = 0.5f);
    void EnableCartoonMode(bool enable, float intensity = 0.8f);
    void EnableSketchMode(bool enable, float intensity = 0.9f);
    
    // 录制功能
    bool StartRecording(const std::string& output_path, int fps = 30);
    void StopRecording();
    bool IsRecording() const { return recording_; }
    
    // 截图功能
    bool TakeScreenshot(const std::string& file_path);
    
    // 性能监控
    PerformanceStats GetPerformanceStats() const { return performance_stats_; }
    void EnablePerformanceMonitoring(bool enable) { performance_monitoring_ = enable; }
    
    // 设置和配置
    struct Settings {
        bool show_fps = true;
        bool show_face_detection = true;
        bool show_landmarks = false;
        bool mirror_mode = false;
        bool fullscreen = false;
        float ui_scale = 1.0f;
        int target_fps = 30;
        bool vsync = true;
        int msaa_samples = 4;
    };
    
    void SetSettings(const Settings& settings);
    Settings GetSettings() const { return settings_; }
    
    // UI控制
    void ShowUI(bool show) { show_ui_ = show; }
    bool IsUIVisible() const { return show_ui_; }
    
    // 事件处理
    void SetKeyCallback(std::function<void(int, int, int, int)> callback) { key_callback_ = callback; }
    void SetMouseCallback(std::function<void(double, double)> callback) { mouse_callback_ = callback; }
    void SetMouseButtonCallback(std::function<void(int, int, int)> callback) { mouse_button_callback_ = callback; }
    
    // 获取组件
    CameraCapture& GetCamera() { return camera_; }
    OpenGLRenderer& GetRenderer() { return renderer_; }
    FilterManager& GetFilterManager() { return filter_manager_; }
    FaceDetector& GetFaceDetector() { return face_detector_; }

private:
    // 核心组件
    CameraCapture camera_;
    OpenGLRenderer renderer_;
    FilterManager filter_manager_;
    FaceDetector face_detector_;
    
    // 状态变量
    bool initialized_;
    bool running_;
    bool camera_active_;
    bool face_detection_enabled_;
    bool recording_;
    bool show_ui_;
    bool performance_monitoring_;
    
    // 当前设置
    Settings settings_;
    EffectParams filter_params_;
    std::string active_sticker_;
    std::string active_background_;
    
    // 检测结果
    std::vector<FaceInfo> detected_faces_;
    
    // 性能统计
    PerformanceStats performance_stats_;
    std::chrono::high_resolution_clock::time_point last_frame_time_;
    int frame_count_;
    
    // 录制相关
    cv::VideoWriter video_writer_;
    std::string recording_path_;
    
    // 事件回调
    std::function<void(int, int, int, int)> key_callback_;
    std::function<void(double, double)> mouse_callback_;
    std::function<void(int, int, int)> mouse_button_callback_;
    
    // 内部处理函数
    void ProcessFrame();
    void UpdatePerformanceStats();
    void RenderUI();
    void HandleInput();
    
    // 图像处理流水线
    cv::Mat ApplyImageProcessing(const cv::Mat& input);
    cv::Mat ApplyFaceEffects(const cv::Mat& input, const std::vector<FaceInfo>& faces);
    cv::Mat ApplyBackgroundEffects(const cv::Mat& input);
    
    // UI渲染
    void RenderMainUI();
    void RenderFilterControls();
    void RenderCameraControls();
    void RenderPerformanceInfo();
    void RenderFaceDetectionInfo();
    
    // 工具函数
    void LoadDefaultAssets();
    void SetupDefaultShaders();
    bool ValidateInitialization();
    
    // 错误处理
    void HandleError(const std::string& error_message);
    void LogPerformanceWarning(const std::string& warning);
    
    // 资源管理
    void CleanupResources();
    void OptimizePerformance();
    
    // 静态回调函数
    static void GLFWKeyCallback(GLFWwindow* window, int key, int scancode, int action, int mods);
    static void GLFWMouseCallback(GLFWwindow* window, double xpos, double ypos);
    static void GLFWMouseButtonCallback(GLFWwindow* window, int button, int action, int mods);
};

} // namespace VideoProcessing
