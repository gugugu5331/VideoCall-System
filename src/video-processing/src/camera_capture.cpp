#include "camera_capture.h"

namespace VideoProcessing {

CameraCapture::CameraCapture() 
    : camera_id_(-1), width_(VIDEO_WIDTH), height_(VIDEO_HEIGHT), 
      fps_(VIDEO_FPS), initialized_(false), actual_fps_(0.0), frame_count_(0) {
}

CameraCapture::~CameraCapture() {
    Release();
}

bool CameraCapture::Initialize(int camera_id, int width, int height) {
    if (initialized_) {
        Release();
    }
    
    camera_id_ = camera_id;
    width_ = width;
    height_ = height;
    
    // 尝试打开摄像头
    capture_.open(camera_id_);
    if (!capture_.isOpened()) {
        std::cerr << "Failed to open camera " << camera_id_ << std::endl;
        return false;
    }
    
    // 设置分辨率
    capture_.set(cv::CAP_PROP_FRAME_WIDTH, width_);
    capture_.set(cv::CAP_PROP_FRAME_HEIGHT, height_);
    capture_.set(cv::CAP_PROP_FPS, fps_);
    
    // 获取实际设置的参数
    width_ = static_cast<int>(capture_.get(cv::CAP_PROP_FRAME_WIDTH));
    height_ = static_cast<int>(capture_.get(cv::CAP_PROP_FRAME_HEIGHT));
    fps_ = capture_.get(cv::CAP_PROP_FPS);
    
    std::cout << "Camera initialized: " << width_ << "x" << height_ 
              << " @ " << fps_ << " FPS" << std::endl;
    
    // 初始化默认参数
    InitializeDefaultProperties();
    
    initialized_ = true;
    last_frame_time_ = std::chrono::high_resolution_clock::now();
    
    return true;
}

void CameraCapture::Release() {
    if (capture_.isOpened()) {
        capture_.release();
    }
    initialized_ = false;
    property_cache_.clear();
}

bool CameraCapture::CaptureFrame(cv::Mat& frame) {
    if (!initialized_ || !capture_.isOpened()) {
        return false;
    }
    
    bool success = capture_.read(frame);
    if (success && !frame.empty()) {
        // 更新FPS统计
        auto current_time = std::chrono::high_resolution_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(
            current_time - last_frame_time_).count();
        
        if (duration > 0) {
            actual_fps_ = 1000.0 / duration;
        }
        last_frame_time_ = current_time;
        frame_count_++;
        
        // 确保图像格式正确
        if (frame.channels() == 4) {
            cv::cvtColor(frame, frame, cv::COLOR_BGRA2BGR);
        } else if (frame.channels() == 1) {
            cv::cvtColor(frame, frame, cv::COLOR_GRAY2BGR);
        }
    }
    
    return success;
}

bool CameraCapture::SetProperty(int property_id, double value) {
    if (!initialized_) {
        return false;
    }
    
    bool success = capture_.set(property_id, value);
    if (success) {
        property_cache_[property_id] = value;
    }
    
    return success;
}

double CameraCapture::GetProperty(int property_id) {
    if (!initialized_) {
        return -1.0;
    }
    
    // 先检查缓存
    auto it = property_cache_.find(property_id);
    if (it != property_cache_.end()) {
        return it->second;
    }
    
    // 从摄像头获取
    double value = capture_.get(property_id);
    property_cache_[property_id] = value;
    
    return value;
}

void CameraCapture::SetFPS(double fps) {
    fps_ = fps;
    if (initialized_) {
        SetProperty(cv::CAP_PROP_FPS, fps);
    }
}

std::vector<cv::Size> CameraCapture::GetSupportedResolutions() {
    std::vector<cv::Size> resolutions;
    
    // 常见分辨率列表
    std::vector<cv::Size> test_resolutions = {
        cv::Size(320, 240),   // QVGA
        cv::Size(640, 480),   // VGA
        cv::Size(800, 600),   // SVGA
        cv::Size(1024, 768),  // XGA
        cv::Size(1280, 720),  // HD
        cv::Size(1280, 960),  // SXGA-
        cv::Size(1600, 1200), // UXGA
        cv::Size(1920, 1080), // Full HD
        cv::Size(2560, 1440), // QHD
        cv::Size(3840, 2160)  // 4K
    };
    
    if (!initialized_) {
        return resolutions;
    }
    
    // 保存当前设置
    int current_width = static_cast<int>(capture_.get(cv::CAP_PROP_FRAME_WIDTH));
    int current_height = static_cast<int>(capture_.get(cv::CAP_PROP_FRAME_HEIGHT));
    
    // 测试每个分辨率
    for (const auto& size : test_resolutions) {
        capture_.set(cv::CAP_PROP_FRAME_WIDTH, size.width);
        capture_.set(cv::CAP_PROP_FRAME_HEIGHT, size.height);
        
        int actual_width = static_cast<int>(capture_.get(cv::CAP_PROP_FRAME_WIDTH));
        int actual_height = static_cast<int>(capture_.get(cv::CAP_PROP_FRAME_HEIGHT));
        
        if (actual_width == size.width && actual_height == size.height) {
            resolutions.push_back(size);
        }
    }
    
    // 恢复原始设置
    capture_.set(cv::CAP_PROP_FRAME_WIDTH, current_width);
    capture_.set(cv::CAP_PROP_FRAME_HEIGHT, current_height);
    
    return resolutions;
}

bool CameraCapture::SetResolution(int width, int height) {
    if (!initialized_) {
        return false;
    }
    
    bool success = true;
    success &= SetProperty(cv::CAP_PROP_FRAME_WIDTH, width);
    success &= SetProperty(cv::CAP_PROP_FRAME_HEIGHT, height);
    
    if (success) {
        width_ = width;
        height_ = height;
    }
    
    return success;
}

void CameraCapture::SetAutoExposure(bool enable) {
    if (enable) {
        SetProperty(cv::CAP_PROP_AUTO_EXPOSURE, 0.75); // 自动曝光
    } else {
        SetProperty(cv::CAP_PROP_AUTO_EXPOSURE, 0.25); // 手动曝光
    }
}

void CameraCapture::SetAutoWhiteBalance(bool enable) {
    SetProperty(cv::CAP_PROP_AUTO_WB, enable ? 1.0 : 0.0);
}

void CameraCapture::SetBrightness(double brightness) {
    SetProperty(cv::CAP_PROP_BRIGHTNESS, brightness);
}

void CameraCapture::SetContrast(double contrast) {
    SetProperty(cv::CAP_PROP_CONTRAST, contrast);
}

void CameraCapture::SetSaturation(double saturation) {
    SetProperty(cv::CAP_PROP_SATURATION, saturation);
}

void CameraCapture::SetExposure(double exposure) {
    SetProperty(cv::CAP_PROP_EXPOSURE, exposure);
}

void CameraCapture::InitializeDefaultProperties() {
    // 设置默认参数
    SetProperty(cv::CAP_PROP_BUFFERSIZE, 1); // 减少延迟
    SetAutoExposure(true);
    SetAutoWhiteBalance(true);
    
    // 获取参数范围并设置合理默认值
    SetBrightness(0.5);
    SetContrast(0.5);
    SetSaturation(0.5);
}

bool CameraCapture::ValidateProperties() {
    if (!initialized_) {
        return false;
    }
    
    // 验证关键参数
    int width = static_cast<int>(GetProperty(cv::CAP_PROP_FRAME_WIDTH));
    int height = static_cast<int>(GetProperty(cv::CAP_PROP_FRAME_HEIGHT));
    double fps = GetProperty(cv::CAP_PROP_FPS);
    
    bool valid = (width > 0 && height > 0 && fps > 0);
    
    if (!valid) {
        std::cerr << "Camera properties validation failed: " 
                  << width << "x" << height << " @ " << fps << " FPS" << std::endl;
    }
    
    return valid;
}

} // namespace VideoProcessing
