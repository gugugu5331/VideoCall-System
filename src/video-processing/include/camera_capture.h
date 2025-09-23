#pragma once

#include "common.h"

namespace VideoProcessing {

class CameraCapture {
public:
    CameraCapture();
    ~CameraCapture();

    // 初始化摄像头
    bool Initialize(int camera_id = 0, int width = VIDEO_WIDTH, int height = VIDEO_HEIGHT);
    
    // 释放资源
    void Release();
    
    // 捕获帧
    bool CaptureFrame(cv::Mat& frame);
    
    // 设置摄像头参数
    bool SetProperty(int property_id, double value);
    double GetProperty(int property_id);
    
    // 获取摄像头信息
    bool IsOpened() const { return capture_.isOpened(); }
    int GetWidth() const { return width_; }
    int GetHeight() const { return height_; }
    double GetFPS() const { return fps_; }
    
    // 设置帧率
    void SetFPS(double fps);
    
    // 获取支持的分辨率
    std::vector<cv::Size> GetSupportedResolutions();
    
    // 设置分辨率
    bool SetResolution(int width, int height);
    
    // 自动曝光和白平衡
    void SetAutoExposure(bool enable);
    void SetAutoWhiteBalance(bool enable);
    
    // 手动调节参数
    void SetBrightness(double brightness);
    void SetContrast(double contrast);
    void SetSaturation(double saturation);
    void SetExposure(double exposure);
    
private:
    cv::VideoCapture capture_;
    int camera_id_;
    int width_;
    int height_;
    double fps_;
    bool initialized_;
    
    // 性能监控
    std::chrono::high_resolution_clock::time_point last_frame_time_;
    double actual_fps_;
    int frame_count_;
    
    // 参数缓存
    std::map<int, double> property_cache_;
    
    // 初始化默认参数
    void InitializeDefaultProperties();
    
    // 验证摄像头参数
    bool ValidateProperties();
};

} // namespace VideoProcessing
