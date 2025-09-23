/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include <opencv2/opencv.hpp>
#include <string>
#include <vector>
#include <chrono>

namespace StackFlows {

class DetectionUtils {
public:
    // File utilities
    static bool file_exists(const std::string& path);
    static std::string get_file_extension(const std::string& path);
    static bool is_image_file(const std::string& path);
    static bool is_video_file(const std::string& path);
    static bool is_audio_file(const std::string& path);
    
    // Image utilities
    static cv::Mat resize_image(const cv::Mat& image, const cv::Size& target_size);
    static cv::Mat normalize_image(const cv::Mat& image);
    static std::vector<float> mat_to_vector(const cv::Mat& mat);
    static cv::Mat vector_to_mat(const std::vector<float>& vec, const cv::Size& size, int type);
    
    // Video utilities
    static int get_video_frame_count(const std::string& video_path);
    static double get_video_fps(const std::string& video_path);
    static cv::Mat extract_frame(const std::string& video_path, int frame_number);
    
    // Audio utilities
    static int get_audio_sample_rate(const std::string& audio_path);
    static double get_audio_duration(const std::string& audio_path);
    
    // String utilities
    static std::string generate_uuid();
    static std::string get_timestamp_string();
    static std::string format_confidence(float confidence);
    
    // JSON utilities
    static std::string create_detection_response(bool is_fake, float confidence, 
                                               const std::string& details = "");
    static std::string create_error_response(const std::string& error_message);
    static std::string create_task_status_response(const std::string& task_id, 
                                                  const std::string& status,
                                                  const std::string& result = "");
    
    // Performance utilities
    static std::chrono::high_resolution_clock::time_point get_current_time();
    static double get_elapsed_time_ms(const std::chrono::high_resolution_clock::time_point& start_time);
    
    // Validation utilities
    static bool validate_image_format(const cv::Mat& image);
    static bool validate_confidence_score(float confidence);
    static bool validate_file_size(const std::string& path, size_t max_size_mb = 100);

private:
    // Private constructor - utility class
    DetectionUtils() = delete;
    
    // Supported file extensions
    static const std::vector<std::string> image_extensions_;
    static const std::vector<std::string> video_extensions_;
    static const std::vector<std::string> audio_extensions_;
};

} // namespace StackFlows
