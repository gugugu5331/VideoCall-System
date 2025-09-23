/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "detection_utils.h"
#include <fstream>
#include <sstream>
#include <iomanip>
#include <random>
#include <algorithm>
#include <sys/stat.h>
#include "json.hpp"

using namespace StackFlows;

// Static member definitions
const std::vector<std::string> DetectionUtils::image_extensions_ = {
    ".jpg", ".jpeg", ".png", ".bmp", ".tiff", ".tif", ".webp"
};

const std::vector<std::string> DetectionUtils::video_extensions_ = {
    ".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm", ".m4v"
};

const std::vector<std::string> DetectionUtils::audio_extensions_ = {
    ".wav", ".mp3", ".flac", ".ogg", ".aac", ".m4a", ".wma"
};

bool DetectionUtils::file_exists(const std::string& path) {
    struct stat buffer;
    return (stat(path.c_str(), &buffer) == 0);
}

std::string DetectionUtils::get_file_extension(const std::string& path) {
    size_t dot_pos = path.find_last_of('.');
    if (dot_pos != std::string::npos) {
        std::string ext = path.substr(dot_pos);
        std::transform(ext.begin(), ext.end(), ext.begin(), ::tolower);
        return ext;
    }
    return "";
}

bool DetectionUtils::is_image_file(const std::string& path) {
    std::string ext = get_file_extension(path);
    return std::find(image_extensions_.begin(), image_extensions_.end(), ext) != image_extensions_.end();
}

bool DetectionUtils::is_video_file(const std::string& path) {
    std::string ext = get_file_extension(path);
    return std::find(video_extensions_.begin(), video_extensions_.end(), ext) != video_extensions_.end();
}

bool DetectionUtils::is_audio_file(const std::string& path) {
    std::string ext = get_file_extension(path);
    return std::find(audio_extensions_.begin(), audio_extensions_.end(), ext) != audio_extensions_.end();
}

cv::Mat DetectionUtils::resize_image(const cv::Mat& image, const cv::Size& target_size) {
    cv::Mat resized;
    cv::resize(image, resized, target_size);
    return resized;
}

cv::Mat DetectionUtils::normalize_image(const cv::Mat& image) {
    cv::Mat normalized;
    image.convertTo(normalized, CV_32F, 1.0/255.0);
    return normalized;
}

std::vector<float> DetectionUtils::mat_to_vector(const cv::Mat& mat) {
    std::vector<float> vec;
    
    if (mat.isContinuous()) {
        vec.assign((float*)mat.data, (float*)mat.data + mat.total() * mat.channels());
    } else {
        for (int i = 0; i < mat.rows; ++i) {
            vec.insert(vec.end(), mat.ptr<float>(i), mat.ptr<float>(i) + mat.cols * mat.channels());
        }
    }
    
    return vec;
}

cv::Mat DetectionUtils::vector_to_mat(const std::vector<float>& vec, const cv::Size& size, int type) {
    cv::Mat mat(size, type);
    std::memcpy(mat.data, vec.data(), vec.size() * sizeof(float));
    return mat;
}

int DetectionUtils::get_video_frame_count(const std::string& video_path) {
    cv::VideoCapture cap(video_path);
    if (!cap.isOpened()) {
        return -1;
    }
    
    int frame_count = static_cast<int>(cap.get(cv::CAP_PROP_FRAME_COUNT));
    cap.release();
    return frame_count;
}

double DetectionUtils::get_video_fps(const std::string& video_path) {
    cv::VideoCapture cap(video_path);
    if (!cap.isOpened()) {
        return -1.0;
    }
    
    double fps = cap.get(cv::CAP_PROP_FPS);
    cap.release();
    return fps;
}

cv::Mat DetectionUtils::extract_frame(const std::string& video_path, int frame_number) {
    cv::VideoCapture cap(video_path);
    cv::Mat frame;
    
    if (!cap.isOpened()) {
        return frame;
    }
    
    cap.set(cv::CAP_PROP_POS_FRAMES, frame_number);
    cap.read(frame);
    cap.release();
    
    return frame;
}

int DetectionUtils::get_audio_sample_rate(const std::string& audio_path) {
    // This would require audio library integration
    // For now, return a default value
    return 16000;
}

double DetectionUtils::get_audio_duration(const std::string& audio_path) {
    // This would require audio library integration
    // For now, return a default value
    return 0.0;
}

std::string DetectionUtils::generate_uuid() {
    static std::random_device rd;
    static std::mt19937 gen(rd());
    static std::uniform_int_distribution<> dis(0, 15);
    static std::uniform_int_distribution<> dis2(8, 11);
    
    std::stringstream ss;
    int i;
    ss << std::hex;
    for (i = 0; i < 8; i++) {
        ss << dis(gen);
    }
    ss << "-";
    for (i = 0; i < 4; i++) {
        ss << dis(gen);
    }
    ss << "-4";
    for (i = 0; i < 3; i++) {
        ss << dis(gen);
    }
    ss << "-";
    ss << dis2(gen);
    for (i = 0; i < 3; i++) {
        ss << dis(gen);
    }
    ss << "-";
    for (i = 0; i < 12; i++) {
        ss << dis(gen);
    }
    return ss.str();
}

std::string DetectionUtils::get_timestamp_string() {
    auto now = std::chrono::system_clock::now();
    auto time_t = std::chrono::system_clock::to_time_t(now);
    
    std::stringstream ss;
    ss << std::put_time(std::localtime(&time_t), "%Y-%m-%d %H:%M:%S");
    return ss.str();
}

std::string DetectionUtils::format_confidence(float confidence) {
    std::stringstream ss;
    ss << std::fixed << std::setprecision(3) << confidence;
    return ss.str();
}

std::string DetectionUtils::create_detection_response(bool is_fake, float confidence, const std::string& details) {
    nlohmann::json response;
    response["is_fake"] = is_fake;
    response["confidence"] = confidence;
    response["details"] = details;
    response["timestamp"] = get_timestamp_string();
    return response.dump();
}

std::string DetectionUtils::create_error_response(const std::string& error_message) {
    nlohmann::json response;
    response["error"] = error_message;
    response["timestamp"] = get_timestamp_string();
    return response.dump();
}

std::string DetectionUtils::create_task_status_response(const std::string& task_id, 
                                                       const std::string& status,
                                                       const std::string& result) {
    nlohmann::json response;
    response["task_id"] = task_id;
    response["status"] = status;
    response["timestamp"] = get_timestamp_string();
    
    if (!result.empty()) {
        try {
            // Try to parse result as JSON
            nlohmann::json result_json = nlohmann::json::parse(result);
            response["result"] = result_json;
        } catch (...) {
            // If parsing fails, store as string
            response["result"] = result;
        }
    }
    
    return response.dump();
}

std::chrono::high_resolution_clock::time_point DetectionUtils::get_current_time() {
    return std::chrono::high_resolution_clock::now();
}

double DetectionUtils::get_elapsed_time_ms(const std::chrono::high_resolution_clock::time_point& start_time) {
    auto end_time = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::microseconds>(end_time - start_time);
    return duration.count() / 1000.0;
}

bool DetectionUtils::validate_image_format(const cv::Mat& image) {
    return !image.empty() && (image.type() == CV_8UC3 || image.type() == CV_8UC1 || image.type() == CV_32FC3);
}

bool DetectionUtils::validate_confidence_score(float confidence) {
    return confidence >= 0.0f && confidence <= 1.0f;
}

bool DetectionUtils::validate_file_size(const std::string& path, size_t max_size_mb) {
    struct stat stat_buf;
    if (stat(path.c_str(), &stat_buf) != 0) {
        return false;
    }
    
    size_t file_size_mb = stat_buf.st_size / (1024 * 1024);
    return file_size_mb <= max_size_mb;
}
