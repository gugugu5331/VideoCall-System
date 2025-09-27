#include "filter_manager.h"
#include <opencv2/opencv.hpp>
#include <opencv2/imgproc.hpp>
#include <iostream>
#include <fstream>
#include <json/json.h>

FilterManager::FilterManager() 
    : initialized_(false)
    , active_filter_(FilterType::NONE)
    , filter_intensity_(1.0f)
{
}

FilterManager::~FilterManager() {
    cleanup();
}

bool FilterManager::initialize() {
    if (initialized_) {
        return true;
    }

    try {
        // 初始化滤镜参数
        initializeFilterParams();
        
        // 加载预设配置
        loadPresets();
        
        initialized_ = true;
        std::cout << "FilterManager initialized successfully" << std::endl;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Error initializing FilterManager: " << e.what() << std::endl;
        return false;
    }
}

void FilterManager::cleanup() {
    filter_presets_.clear();
    initialized_ = false;
}

void FilterManager::applyFilters(cv::Mat& frame, const std::vector<FaceInfo>& faces) {
    if (!initialized_ || frame.empty() || active_filter_ == FilterType::NONE) {
        return;
    }

    try {
        switch (active_filter_) {
            case FilterType::BLUR:
                applyBlurFilter(frame);
                break;
            case FilterType::SHARPEN:
                applySharpenFilter(frame);
                break;
            case FilterType::VINTAGE:
                applyVintageFilter(frame);
                break;
            case FilterType::CARTOON:
                applyCartoonFilter(frame);
                break;
            case FilterType::BEAUTY:
                applyBeautyFilter(frame, faces);
                break;
            case FilterType::EDGE_DETECTION:
                applyEdgeDetectionFilter(frame);
                break;
            case FilterType::EMBOSS:
                applyEmbossFilter(frame);
                break;
            case FilterType::SEPIA:
                applySepiaFilter(frame);
                break;
            case FilterType::GRAYSCALE:
                applyGrayscaleFilter(frame);
                break;
            case FilterType::NEON:
                applyNeonFilter(frame);
                break;
            default:
                break;
        }
    } catch (const std::exception& e) {
        std::cerr << "Error applying filter: " << e.what() << std::endl;
    }
}

void FilterManager::setActiveFilter(FilterType type) {
    active_filter_ = type;
    std::cout << "Active filter set to: " << static_cast<int>(type) << std::endl;
}

void FilterManager::setFilterIntensity(float intensity) {
    filter_intensity_ = std::clamp(intensity, 0.0f, 2.0f);
}

void FilterManager::applyBlurFilter(cv::Mat& frame) {
    int kernel_size = static_cast<int>(5 + filter_intensity_ * 10);
    if (kernel_size % 2 == 0) kernel_size++;
    cv::GaussianBlur(frame, frame, cv::Size(kernel_size, kernel_size), 0);
}

void FilterManager::applySharpenFilter(cv::Mat& frame) {
    cv::Mat kernel = (cv::Mat_<float>(3, 3) << 
        0, -1 * filter_intensity_, 0,
        -1 * filter_intensity_, 1 + 4 * filter_intensity_, -1 * filter_intensity_,
        0, -1 * filter_intensity_, 0);
    cv::filter2D(frame, frame, -1, kernel);
}

void FilterManager::applyVintageFilter(cv::Mat& frame) {
    // 创建复古效果
    cv::Mat vintage;
    frame.convertTo(vintage, -1, 0.8, 20); // 降低对比度，增加亮度
    
    // 添加棕褐色调
    std::vector<cv::Mat> channels;
    cv::split(vintage, channels);
    
    // 调整颜色通道
    channels[0] *= 0.8; // 蓝色通道
    channels[1] *= 0.9; // 绿色通道
    channels[2] *= 1.1; // 红色通道
    
    cv::merge(channels, vintage);
    
    // 混合原图和效果图
    cv::addWeighted(frame, 1.0f - filter_intensity_, vintage, filter_intensity_, 0, frame);
}

void FilterManager::applyCartoonFilter(cv::Mat& frame) {
    cv::Mat gray, edges, cartoon;
    
    // 转换为灰度图
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    
    // 边缘检测
    cv::adaptiveThreshold(gray, edges, 255, cv::ADAPTIVE_THRESH_MEAN_C, cv::THRESH_BINARY, 7, 7);
    
    // 双边滤波平滑图像
    cv::bilateralFilter(frame, cartoon, 15, 50, 50);
    
    // 将边缘转换为3通道
    cv::cvtColor(edges, edges, cv::COLOR_GRAY2BGR);
    
    // 合并卡通效果和边缘
    cv::bitwise_and(cartoon, edges, cartoon);
    
    // 混合原图和卡通效果
    cv::addWeighted(frame, 1.0f - filter_intensity_, cartoon, filter_intensity_, 0, frame);
}

void FilterManager::applyBeautyFilter(cv::Mat& frame, const std::vector<FaceInfo>& faces) {
    if (faces.empty()) {
        return;
    }

    cv::Mat beauty = frame.clone();
    
    // 对每个检测到的人脸应用美颜效果
    for (const auto& face : faces) {
        cv::Rect face_rect = face.bounding_box;
        
        // 确保人脸区域在图像范围内
        face_rect &= cv::Rect(0, 0, frame.cols, frame.rows);
        if (face_rect.width <= 0 || face_rect.height <= 0) {
            continue;
        }
        
        cv::Mat face_roi = beauty(face_rect);
        
        // 磨皮效果 - 双边滤波
        cv::Mat smooth_face;
        cv::bilateralFilter(face_roi, smooth_face, 15, 50, 50);
        
        // 美白效果
        cv::Mat brightened_face;
        smooth_face.convertTo(brightened_face, -1, 1.0, 10 * filter_intensity_);
        
        // 混合效果
        cv::addWeighted(face_roi, 1.0f - filter_intensity_ * 0.7f, 
                       brightened_face, filter_intensity_ * 0.7f, 0, face_roi);
    }
    
    frame = beauty;
}

void FilterManager::applyEdgeDetectionFilter(cv::Mat& frame) {
    cv::Mat gray, edges;
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    cv::Canny(gray, edges, 50 * filter_intensity_, 150 * filter_intensity_);
    cv::cvtColor(edges, edges, cv::COLOR_GRAY2BGR);
    
    // 混合原图和边缘检测结果
    cv::addWeighted(frame, 1.0f - filter_intensity_, edges, filter_intensity_, 0, frame);
}

void FilterManager::applyEmbossFilter(cv::Mat& frame) {
    cv::Mat kernel = (cv::Mat_<float>(3, 3) << 
        -2 * filter_intensity_, -1 * filter_intensity_, 0,
        -1 * filter_intensity_, 1, 1 * filter_intensity_,
        0, 1 * filter_intensity_, 2 * filter_intensity_);
    
    cv::Mat embossed;
    cv::filter2D(frame, embossed, -1, kernel);
    embossed += cv::Scalar(128, 128, 128); // 添加灰度偏移
    
    cv::addWeighted(frame, 1.0f - filter_intensity_, embossed, filter_intensity_, 0, frame);
}

void FilterManager::applySepiaFilter(cv::Mat& frame) {
    cv::Mat sepia;
    cv::transform(frame, sepia, cv::Matx34f(
        0.272, 0.534, 0.131, 0,
        0.349, 0.686, 0.168, 0,
        0.393, 0.769, 0.189, 0
    ));
    
    cv::addWeighted(frame, 1.0f - filter_intensity_, sepia, filter_intensity_, 0, frame);
}

void FilterManager::applyGrayscaleFilter(cv::Mat& frame) {
    cv::Mat gray;
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    cv::cvtColor(gray, gray, cv::COLOR_GRAY2BGR);
    
    cv::addWeighted(frame, 1.0f - filter_intensity_, gray, filter_intensity_, 0, frame);
}

void FilterManager::applyNeonFilter(cv::Mat& frame) {
    cv::Mat neon;
    
    // 增强对比度和饱和度
    frame.convertTo(neon, -1, 1.5 * filter_intensity_, 0);
    
    // 转换到HSV色彩空间
    cv::Mat hsv;
    cv::cvtColor(neon, hsv, cv::COLOR_BGR2HSV);
    
    std::vector<cv::Mat> channels;
    cv::split(hsv, channels);
    
    // 增强饱和度
    channels[1] *= (1.0 + filter_intensity_);
    
    cv::merge(channels, hsv);
    cv::cvtColor(hsv, neon, cv::COLOR_HSV2BGR);
    
    cv::addWeighted(frame, 1.0f - filter_intensity_, neon, filter_intensity_, 0, frame);
}

void FilterManager::initializeFilterParams() {
    // 初始化默认滤镜参数
    filter_params_[FilterType::BLUR] = {1.0f, {{"kernel_size", 5}}};
    filter_params_[FilterType::SHARPEN] = {1.0f, {{"strength", 1.0}}};
    filter_params_[FilterType::VINTAGE] = {0.8f, {{"sepia_strength", 0.7}}};
    filter_params_[FilterType::CARTOON] = {0.9f, {{"edge_threshold", 7}}};
    filter_params_[FilterType::BEAUTY] = {0.7f, {{"smooth_strength", 15}}};
}

std::vector<std::string> FilterManager::getAvailableFilters() const {
    return {
        "None", "Blur", "Sharpen", "Vintage", "Cartoon", 
        "Beauty", "Edge Detection", "Emboss", "Sepia", 
        "Grayscale", "Neon"
    };
}

bool FilterManager::savePreset(const std::string& name, const FilterConfig& config) {
    filter_presets_[name] = config;
    
    // 保存到文件
    try {
        Json::Value root;
        for (const auto& preset : filter_presets_) {
            Json::Value preset_json;
            preset_json["intensity"] = preset.second.intensity;
            preset_json["filter_type"] = static_cast<int>(preset.second.type);
            
            Json::Value params_json;
            for (const auto& param : preset.second.parameters) {
                params_json[param.first] = param.second;
            }
            preset_json["parameters"] = params_json;
            
            root[preset.first] = preset_json;
        }
        
        std::ofstream file("filter_presets.json");
        file << root;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Error saving preset: " << e.what() << std::endl;
        return false;
    }
}

bool FilterManager::loadPreset(const std::string& name) {
    auto it = filter_presets_.find(name);
    if (it != filter_presets_.end()) {
        active_filter_ = it->second.type;
        filter_intensity_ = it->second.intensity;
        return true;
    }
    return false;
}

void FilterManager::loadPresets() {
    try {
        std::ifstream file("filter_presets.json");
        if (!file.is_open()) {
            return; // 文件不存在，使用默认设置
        }
        
        Json::Value root;
        file >> root;
        
        for (const auto& member : root.getMemberNames()) {
            FilterConfig config;
            config.intensity = root[member]["intensity"].asFloat();
            config.type = static_cast<FilterType>(root[member]["filter_type"].asInt());
            
            const Json::Value& params = root[member]["parameters"];
            for (const auto& param_name : params.getMemberNames()) {
                config.parameters[param_name] = params[param_name].asFloat();
            }
            
            filter_presets_[member] = config;
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Error loading presets: " << e.what() << std::endl;
    }
}
