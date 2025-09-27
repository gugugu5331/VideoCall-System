#include "video_processor.h"
#include "filter_manager.h"
#include "face_detector.h"
#include "opengl_renderer.h"
#include "texture_manager.h"
#include <opencv2/opencv.hpp>
#include <iostream>
#include <chrono>

VideoProcessor::VideoProcessor() 
    : initialized_(false)
    , processing_(false)
    , filter_manager_(nullptr)
    , face_detector_(nullptr)
    , opengl_renderer_(nullptr)
    , texture_manager_(nullptr)
    , current_frame_id_(0)
    , fps_counter_(0)
    , last_fps_time_(std::chrono::steady_clock::now())
{
}

VideoProcessor::~VideoProcessor() {
    cleanup();
}

bool VideoProcessor::initialize() {
    if (initialized_) {
        return true;
    }

    try {
        // 初始化OpenCV
        std::cout << "Initializing OpenCV..." << std::endl;
        std::cout << "OpenCV version: " << CV_VERSION << std::endl;

        // 初始化组件
        filter_manager_ = std::make_unique<FilterManager>();
        if (!filter_manager_->initialize()) {
            std::cerr << "Failed to initialize FilterManager" << std::endl;
            return false;
        }

        face_detector_ = std::make_unique<FaceDetector>();
        if (!face_detector_->initialize()) {
            std::cerr << "Failed to initialize FaceDetector" << std::endl;
            return false;
        }

        opengl_renderer_ = std::make_unique<OpenGLRenderer>();
        if (!opengl_renderer_->initialize()) {
            std::cerr << "Failed to initialize OpenGLRenderer" << std::endl;
            return false;
        }

        texture_manager_ = std::make_unique<TextureManager>();
        if (!texture_manager_->initialize()) {
            std::cerr << "Failed to initialize TextureManager" << std::endl;
            return false;
        }

        initialized_ = true;
        std::cout << "VideoProcessor initialized successfully" << std::endl;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Exception during initialization: " << e.what() << std::endl;
        return false;
    }
}

void VideoProcessor::cleanup() {
    processing_ = false;
    
    if (texture_manager_) {
        texture_manager_->cleanup();
        texture_manager_.reset();
    }
    
    if (opengl_renderer_) {
        opengl_renderer_->cleanup();
        opengl_renderer_.reset();
    }
    
    if (face_detector_) {
        face_detector_->cleanup();
        face_detector_.reset();
    }
    
    if (filter_manager_) {
        filter_manager_->cleanup();
        filter_manager_.reset();
    }
    
    initialized_ = false;
    std::cout << "VideoProcessor cleaned up" << std::endl;
}

bool VideoProcessor::processFrame(const cv::Mat& input_frame, cv::Mat& output_frame) {
    if (!initialized_ || input_frame.empty()) {
        return false;
    }

    auto start_time = std::chrono::high_resolution_clock::now();

    try {
        // 复制输入帧
        cv::Mat working_frame = input_frame.clone();
        
        // 1. 人脸检测
        std::vector<FaceInfo> faces;
        if (face_detector_) {
            faces = face_detector_->detectFaces(working_frame);
        }

        // 2. 应用滤镜
        if (filter_manager_) {
            filter_manager_->applyFilters(working_frame, faces);
        }

        // 3. 应用贴图和特效
        if (texture_manager_ && !faces.empty()) {
            texture_manager_->applyTextures(working_frame, faces);
        }

        // 4. OpenGL后处理（如果需要）
        if (opengl_renderer_) {
            opengl_renderer_->renderFrame(working_frame);
        }

        // 输出处理后的帧
        output_frame = working_frame.clone();

        // 更新性能统计
        updatePerformanceStats(start_time);
        
        current_frame_id_++;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Error processing frame: " << e.what() << std::endl;
        return false;
    }
}

void VideoProcessor::setFilterType(FilterType type) {
    if (filter_manager_) {
        filter_manager_->setActiveFilter(type);
    }
}

void VideoProcessor::setFilterIntensity(float intensity) {
    if (filter_manager_) {
        filter_manager_->setFilterIntensity(intensity);
    }
}

void VideoProcessor::addSticker(const std::string& sticker_path, StickerType type) {
    if (texture_manager_) {
        texture_manager_->loadSticker(sticker_path, type);
    }
}

void VideoProcessor::removeSticker(StickerType type) {
    if (texture_manager_) {
        texture_manager_->removeSticker(type);
    }
}

void VideoProcessor::enableFaceDetection(bool enable) {
    if (face_detector_) {
        face_detector_->setEnabled(enable);
    }
}

void VideoProcessor::setFaceDetectionModel(const std::string& model_path) {
    if (face_detector_) {
        face_detector_->loadModel(model_path);
    }
}

ProcessingStats VideoProcessor::getStats() const {
    ProcessingStats stats;
    stats.fps = current_fps_;
    stats.frame_count = current_frame_id_;
    stats.avg_processing_time = avg_processing_time_;
    stats.total_processing_time = total_processing_time_;
    return stats;
}

void VideoProcessor::resetStats() {
    current_frame_id_ = 0;
    fps_counter_ = 0;
    current_fps_ = 0.0f;
    avg_processing_time_ = 0.0f;
    total_processing_time_ = 0.0f;
    last_fps_time_ = std::chrono::steady_clock::now();
}

void VideoProcessor::updatePerformanceStats(const std::chrono::high_resolution_clock::time_point& start_time) {
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::microseconds>(end_time - start_time);
    
    float processing_time_ms = duration.count() / 1000.0f;
    total_processing_time_ += processing_time_ms;
    avg_processing_time_ = total_processing_time_ / current_frame_id_;
    
    // 更新FPS
    fps_counter_++;
    auto now = std::chrono::steady_clock::now();
    auto fps_duration = std::chrono::duration_cast<std::chrono::milliseconds>(now - last_fps_time_);
    
    if (fps_duration.count() >= 1000) { // 每秒更新一次FPS
        current_fps_ = fps_counter_ * 1000.0f / fps_duration.count();
        fps_counter_ = 0;
        last_fps_time_ = now;
    }
}

std::vector<std::string> VideoProcessor::getAvailableFilters() const {
    if (filter_manager_) {
        return filter_manager_->getAvailableFilters();
    }
    return {};
}

std::vector<std::string> VideoProcessor::getAvailableStickers() const {
    if (texture_manager_) {
        return texture_manager_->getAvailableStickers();
    }
    return {};
}

bool VideoProcessor::saveFilterPreset(const std::string& name, const FilterConfig& config) {
    if (filter_manager_) {
        return filter_manager_->savePreset(name, config);
    }
    return false;
}

bool VideoProcessor::loadFilterPreset(const std::string& name) {
    if (filter_manager_) {
        return filter_manager_->loadPreset(name);
    }
    return false;
}

void VideoProcessor::setProcessingMode(ProcessingMode mode) {
    processing_mode_ = mode;
    
    // 根据模式调整处理参数
    switch (mode) {
        case ProcessingMode::PERFORMANCE:
            // 性能优先模式
            if (face_detector_) {
                face_detector_->setDetectionInterval(3); // 每3帧检测一次
            }
            break;
            
        case ProcessingMode::QUALITY:
            // 质量优先模式
            if (face_detector_) {
                face_detector_->setDetectionInterval(1); // 每帧都检测
            }
            break;
            
        case ProcessingMode::BALANCED:
        default:
            // 平衡模式
            if (face_detector_) {
                face_detector_->setDetectionInterval(2); // 每2帧检测一次
            }
            break;
    }
}

bool VideoProcessor::isProcessing() const {
    return processing_;
}

void VideoProcessor::startProcessing() {
    processing_ = true;
}

void VideoProcessor::stopProcessing() {
    processing_ = false;
}
