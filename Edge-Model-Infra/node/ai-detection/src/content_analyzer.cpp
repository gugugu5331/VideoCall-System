/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "content_analyzer.h"
#include "detection_utils.h"
#include <iostream>
#include <algorithm>
#include <random>

using namespace StackFlows;

ContentAnalyzer::ContentAnalyzer() 
    : initialized_(false), has_prev_frame_(false) {
    
    emotion_labels_ = {
        "neutral", "happy", "sad", "angry", "surprised", "fear", "disgust"
    };
}

ContentAnalyzer::~ContentAnalyzer() {
}

bool ContentAnalyzer::initialize(const AnalysisConfig& config) {
    config_ = config;
    
    // Initialize face cascade classifier
    std::string cascade_path = cv::data::haarcascades + "haarcascade_frontalface_default.xml";
    if (!face_cascade_.load(cascade_path)) {
        std::cerr << "Warning: Could not load face cascade classifier" << std::endl;
        return false;
    }
    
    // Initialize background subtractor for motion detection
    bg_subtractor_ = cv::createBackgroundSubtractorMOG2();
    
    initialized_ = true;
    std::cout << "ContentAnalyzer initialized successfully" << std::endl;
    return true;
}

ContentAnalysisResult ContentAnalyzer::analyze_video(const std::string& video_path) {
    ContentAnalysisResult result;
    
    if (!initialized_) {
        result.summary = "Analyzer not initialized";
        return result;
    }
    
    cv::VideoCapture cap(video_path);
    if (!cap.isOpened()) {
        result.summary = "Failed to open video";
        return result;
    }
    
    double fps = cap.get(cv::CAP_PROP_FPS);
    int total_frames = static_cast<int>(cap.get(cv::CAP_PROP_FRAME_COUNT));
    int sample_interval = static_cast<int>(fps / config_.video_sample_rate);
    
    cv::Mat frame, prev_frame;
    int frame_count = 0;
    int analyzed_frames = 0;
    
    std::cout << "Analyzing video: " << video_path << " (FPS: " << fps << ", Frames: " << total_frames << ")" << std::endl;
    
    while (cap.read(frame)) {
        // Sample frames based on configured rate
        if (frame_count % sample_interval == 0) {
            float timestamp = static_cast<float>(frame_count) / fps;
            
            // Emotion analysis
            EmotionResult emotion = analyze_frame_emotion(frame);
            if (!emotion.dominant_emotion.empty()) {
                result.emotions.push_back(emotion);
            }
            
            // Motion analysis (if we have a previous frame)
            if (has_prev_frame_) {
                MotionResult motion = analyze_frame_motion(frame, prev_frame_);
                result.motion_data.push_back(motion);
                
                // Scene change detection
                float scene_change = detect_scene_change(frame, prev_frame_);
                if (scene_change > 0.5f) {
                    result.scene_changes.push_back(timestamp);
                }
            }
            
            prev_frame_ = frame.clone();
            has_prev_frame_ = true;
            analyzed_frames++;
        }
        
        frame_count++;
        
        // Limit analysis for performance
        if (analyzed_frames > 1000) {
            break;
        }
    }
    
    cap.release();
    
    // Voice activity detection (placeholder)
    result.voice_activity = detect_voice_activity(video_path);
    
    // Generate summary
    std::stringstream summary;
    summary << "Analyzed " << analyzed_frames << " frames. ";
    summary << "Found " << result.emotions.size() << " emotion segments, ";
    summary << result.motion_data.size() << " motion segments, ";
    summary << result.scene_changes.size() << " scene changes.";
    result.summary = summary.str();
    
    return result;
}

EmotionResult ContentAnalyzer::analyze_frame_emotion(const cv::Mat& frame) {
    EmotionResult result;
    result.dominant_emotion = "neutral";
    result.confidence = 0.0f;
    
    // Detect faces
    std::vector<cv::Rect> faces = detect_faces(frame);
    
    if (faces.empty()) {
        return result;
    }
    
    // Analyze the largest face
    cv::Rect largest_face = *std::max_element(faces.begin(), faces.end(),
        [](const cv::Rect& a, const cv::Rect& b) {
            return a.area() < b.area();
        });
    
    cv::Mat face_roi = frame(largest_face);
    result = classify_emotion(face_roi);
    
    return result;
}

MotionResult ContentAnalyzer::analyze_frame_motion(const cv::Mat& frame, const cv::Mat& prev_frame) {
    MotionResult result;
    result.motion_intensity = 0.0f;
    result.significant_motion = false;
    
    // Detect optical flow
    std::vector<cv::Point2f> flow_vectors = detect_optical_flow(frame, prev_frame);
    result.motion_vectors = flow_vectors;
    
    // Calculate motion intensity
    result.motion_intensity = calculate_motion_intensity(flow_vectors);
    result.significant_motion = result.motion_intensity > config_.motion_threshold;
    
    return result;
}

std::vector<cv::Rect> ContentAnalyzer::detect_faces(const cv::Mat& frame) {
    std::vector<cv::Rect> faces;
    
    if (face_cascade_.empty()) {
        return faces;
    }
    
    cv::Mat gray;
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    
    face_cascade_.detectMultiScale(gray, faces, 1.1, 3, 0, cv::Size(30, 30));
    
    return faces;
}

EmotionResult ContentAnalyzer::classify_emotion(const cv::Mat& face_roi) {
    EmotionResult result;
    
    // Simple heuristic-based emotion classification
    // In a real implementation, you would use a trained emotion recognition model
    
    cv::Mat gray;
    cv::cvtColor(face_roi, gray, cv::COLOR_BGR2GRAY);
    
    // Calculate basic image statistics
    cv::Scalar mean_intensity, std_intensity;
    cv::meanStdDev(gray, mean_intensity, std_intensity);
    
    // Simple heuristic based on intensity distribution
    float mean_val = mean_intensity[0];
    float std_val = std_intensity[0];
    
    // Random emotion for demonstration (replace with actual model)
    static std::random_device rd;
    static std::mt19937 gen(rd());
    static std::uniform_int_distribution<> dis(0, emotion_labels_.size() - 1);
    
    int emotion_idx = dis(gen);
    result.dominant_emotion = emotion_labels_[emotion_idx];
    result.confidence = 0.5f + (std_val / 255.0f) * 0.5f; // Use std as confidence proxy
    
    // Fill emotion probabilities
    for (const auto& emotion : emotion_labels_) {
        if (emotion == result.dominant_emotion) {
            result.emotions[emotion] = result.confidence;
        } else {
            result.emotions[emotion] = (1.0f - result.confidence) / (emotion_labels_.size() - 1);
        }
    }
    
    return result;
}

std::vector<cv::Point2f> ContentAnalyzer::detect_optical_flow(const cv::Mat& frame, const cv::Mat& prev_frame) {
    std::vector<cv::Point2f> flow_vectors;
    
    cv::Mat gray, prev_gray;
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    cv::cvtColor(prev_frame, prev_gray, cv::COLOR_BGR2GRAY);
    
    // Detect corners in previous frame
    std::vector<cv::Point2f> prev_points;
    cv::goodFeaturesToTrack(prev_gray, prev_points, 100, 0.01, 10);
    
    if (prev_points.empty()) {
        return flow_vectors;
    }
    
    // Calculate optical flow
    std::vector<cv::Point2f> curr_points;
    std::vector<uchar> status;
    std::vector<float> errors;
    
    cv::calcOpticalFlowPyrLK(prev_gray, gray, prev_points, curr_points, status, errors);
    
    // Extract valid flow vectors
    for (size_t i = 0; i < prev_points.size(); ++i) {
        if (status[i]) {
            cv::Point2f flow = curr_points[i] - prev_points[i];
            flow_vectors.push_back(flow);
        }
    }
    
    return flow_vectors;
}

float ContentAnalyzer::calculate_motion_intensity(const std::vector<cv::Point2f>& flow_vectors) {
    if (flow_vectors.empty()) {
        return 0.0f;
    }
    
    float total_magnitude = 0.0f;
    for (const auto& flow : flow_vectors) {
        total_magnitude += cv::norm(flow);
    }
    
    return total_magnitude / flow_vectors.size();
}

float ContentAnalyzer::detect_scene_change(const cv::Mat& frame, const cv::Mat& prev_frame) {
    // Simple histogram-based scene change detection
    cv::Mat hist1, hist2;
    
    int histSize = 256;
    float range[] = {0, 256};
    const float* histRange = {range};
    
    cv::calcHist(&frame, 1, 0, cv::Mat(), hist1, 1, &histSize, &histRange);
    cv::calcHist(&prev_frame, 1, 0, cv::Mat(), hist2, 1, &histSize, &histRange);
    
    // Normalize histograms
    cv::normalize(hist1, hist1, 0, 1, cv::NORM_L1);
    cv::normalize(hist2, hist2, 0, 1, cv::NORM_L1);
    
    // Calculate correlation
    double correlation = cv::compareHist(hist1, hist2, cv::HISTCMP_CORREL);
    
    // Return scene change score (1 - correlation)
    return static_cast<float>(1.0 - correlation);
}

std::vector<float> ContentAnalyzer::detect_voice_activity(const std::string& audio_path) {
    // Placeholder for voice activity detection
    // In a real implementation, this would extract audio from video and analyze it
    std::vector<float> voice_activity;
    
    // Generate dummy voice activity data
    for (int i = 0; i < 100; ++i) {
        voice_activity.push_back(0.5f);
    }
    
    return voice_activity;
}
