/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#pragma once

#include <opencv2/opencv.hpp>
#include <string>
#include <vector>
#include <memory>

namespace StackFlows {

struct AnalysisConfig {
    float video_sample_rate = 1.0f;  // frames per second to analyze
    int audio_sample_rate = 16000;   // audio sample rate
    float emotion_threshold = 0.6f;  // emotion detection threshold
    float voice_activity_threshold = 0.5f;  // voice activity threshold
    float motion_threshold = 0.3f;   // motion detection threshold
    float face_confidence_threshold = 0.9f;  // face detection confidence
};

struct EmotionResult {
    std::string dominant_emotion;
    float confidence;
    std::map<std::string, float> emotions;
};

struct MotionResult {
    float motion_intensity;
    std::vector<cv::Point2f> motion_vectors;
    bool significant_motion;
};

struct ContentAnalysisResult {
    std::vector<EmotionResult> emotions;
    std::vector<MotionResult> motion_data;
    std::vector<float> voice_activity;
    std::vector<float> scene_changes;
    std::string summary;
};

class ContentAnalyzer {
public:
    ContentAnalyzer();
    virtual ~ContentAnalyzer();

    // Initialize analyzer with configuration
    bool initialize(const AnalysisConfig& config);

    // Analyze video content
    ContentAnalysisResult analyze_video(const std::string& video_path);

    // Analyze single frame
    EmotionResult analyze_frame_emotion(const cv::Mat& frame);
    MotionResult analyze_frame_motion(const cv::Mat& frame, const cv::Mat& prev_frame);

    // Check if analyzer is ready
    bool is_ready() const { return initialized_; }

private:
    // Emotion analysis
    std::vector<cv::Rect> detect_faces(const cv::Mat& frame);
    EmotionResult classify_emotion(const cv::Mat& face_roi);
    
    // Motion analysis
    std::vector<cv::Point2f> detect_optical_flow(const cv::Mat& frame, const cv::Mat& prev_frame);
    float calculate_motion_intensity(const std::vector<cv::Point2f>& flow_vectors);
    
    // Scene analysis
    float detect_scene_change(const cv::Mat& frame, const cv::Mat& prev_frame);
    
    // Voice activity detection (placeholder for audio integration)
    std::vector<float> detect_voice_activity(const std::string& audio_path);

private:
    bool initialized_;
    AnalysisConfig config_;
    
    // OpenCV components
    cv::CascadeClassifier face_cascade_;
    cv::Ptr<cv::BackgroundSubtractor> bg_subtractor_;
    
    // Previous frame for motion analysis
    cv::Mat prev_frame_;
    bool has_prev_frame_;
    
    // Emotion classification (simple heuristic-based for now)
    std::vector<std::string> emotion_labels_;
};

} // namespace StackFlows
