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

#ifdef USE_TENSORFLOW
#include <tensorflow/cc/client/client_session.h>
#include <tensorflow/cc/ops/standard_ops.h>
#include <tensorflow/core/framework/tensor.h>
#endif

namespace StackFlows {

struct DetectionResult {
    bool is_fake;
    float confidence;
    std::vector<cv::Rect> faces;
    std::string details;
};

class FaceSwapDetector {
public:
    FaceSwapDetector();
    virtual ~FaceSwapDetector();

    // Initialize the detector with model path
    bool initialize(const std::string& model_path);

    // Detect face swap in image
    DetectionResult detect_image(const cv::Mat& image);

    // Detect face swap in video
    DetectionResult detect_video(const std::string& video_path);

    // Check if detector is ready
    bool is_ready() const { return model_loaded_; }

private:
    // Face detection
    std::vector<cv::Rect> detect_faces(const cv::Mat& image);
    
    // Preprocess face for model input
    cv::Mat preprocess_face(const cv::Mat& image, const cv::Rect& face_rect);
    
    // Model inference
    float predict_face_swap(const cv::Mat& face_image);
    
    // Load model (TensorFlow or fallback)
    bool load_model(const std::string& model_path);
    
    // Create dummy model for testing
    void create_dummy_model();

private:
    cv::CascadeClassifier face_cascade_;
    bool model_loaded_;
    std::string model_path_;
    
#ifdef USE_TENSORFLOW
    std::unique_ptr<tensorflow::Session> tf_session_;
    std::string input_layer_name_;
    std::string output_layer_name_;
#endif
    
    // Model parameters
    cv::Size input_size_;
    float detection_threshold_;
};

} // namespace StackFlows
