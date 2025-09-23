/*
 * SPDX-FileCopyrightText: 2024 M5Stack Technology CO LTD
 *
 * SPDX-License-Identifier: MIT
 */
#include "face_swap_detector.h"
#include "detection_utils.h"
#include <iostream>
#include <random>

using namespace StackFlows;

FaceSwapDetector::FaceSwapDetector() 
    : model_loaded_(false), input_size_(224, 224), detection_threshold_(0.5f) {
    
    // Initialize face cascade classifier
    std::string cascade_path = cv::data::haarcascades + "haarcascade_frontalface_default.xml";
    if (!face_cascade_.load(cascade_path)) {
        std::cerr << "Warning: Could not load face cascade classifier" << std::endl;
    }
}

FaceSwapDetector::~FaceSwapDetector() {
#ifdef USE_TENSORFLOW
    if (tf_session_) {
        tf_session_->Close();
    }
#endif
}

bool FaceSwapDetector::initialize(const std::string& model_path) {
    model_path_ = model_path;
    
    if (DetectionUtils::file_exists(model_path)) {
        return load_model(model_path);
    } else {
        std::cout << "Model file not found, using dummy model for testing" << std::endl;
        create_dummy_model();
        return true;
    }
}

DetectionResult FaceSwapDetector::detect_image(const cv::Mat& image) {
    DetectionResult result;
    result.is_fake = false;
    result.confidence = 0.0f;
    result.details = "No faces detected";
    
    if (image.empty()) {
        result.details = "Invalid image";
        return result;
    }
    
    // Detect faces
    std::vector<cv::Rect> faces = detect_faces(image);
    result.faces = faces;
    
    if (faces.empty()) {
        return result;
    }
    
    float max_fake_confidence = 0.0f;
    bool any_fake = false;
    
    // Analyze each detected face
    for (const auto& face_rect : faces) {
        cv::Mat face_roi = preprocess_face(image, face_rect);
        float prediction = predict_face_swap(face_roi);
        
        bool is_fake = prediction > detection_threshold_;
        float confidence = is_fake ? prediction : (1.0f - prediction);
        
        if (is_fake && confidence > max_fake_confidence) {
            max_fake_confidence = confidence;
            any_fake = true;
        }
    }
    
    result.is_fake = any_fake;
    result.confidence = max_fake_confidence;
    result.details = any_fake ? "Face swap detected" : "No face swap detected";
    
    return result;
}

DetectionResult FaceSwapDetector::detect_video(const std::string& video_path) {
    DetectionResult result;
    result.is_fake = false;
    result.confidence = 0.0f;
    result.details = "Video analysis failed";
    
    cv::VideoCapture cap(video_path);
    if (!cap.isOpened()) {
        result.details = "Failed to open video";
        return result;
    }
    
    int frame_count = 0;
    int fake_frames = 0;
    float total_confidence = 0.0f;
    int analyzed_frames = 0;
    
    cv::Mat frame;
    while (cap.read(frame)) {
        // Analyze every 30th frame to reduce computation
        if (frame_count % 30 == 0) {
            DetectionResult frame_result = detect_image(frame);
            
            if (!frame_result.faces.empty()) {
                analyzed_frames++;
                total_confidence += frame_result.confidence;
                
                if (frame_result.is_fake) {
                    fake_frames++;
                }
            }
        }
        frame_count++;
    }
    
    cap.release();
    
    if (analyzed_frames > 0) {
        float avg_confidence = total_confidence / analyzed_frames;
        float fake_ratio = static_cast<float>(fake_frames) / analyzed_frames;
        
        result.is_fake = fake_ratio > 0.3f; // If more than 30% of frames are fake
        result.confidence = avg_confidence;
        result.details = "Video analysis completed. Analyzed " + std::to_string(analyzed_frames) + " frames";
    } else {
        result.details = "No faces detected in video";
    }
    
    return result;
}

std::vector<cv::Rect> FaceSwapDetector::detect_faces(const cv::Mat& image) {
    std::vector<cv::Rect> faces;
    
    if (face_cascade_.empty()) {
        return faces;
    }
    
    cv::Mat gray;
    cv::cvtColor(image, gray, cv::COLOR_BGR2GRAY);
    
    face_cascade_.detectMultiScale(gray, faces, 1.1, 3, 0, cv::Size(30, 30));
    
    return faces;
}

cv::Mat FaceSwapDetector::preprocess_face(const cv::Mat& image, const cv::Rect& face_rect) {
    // Extract face region
    cv::Mat face_roi = image(face_rect);
    
    // Resize to model input size
    cv::Mat resized;
    cv::resize(face_roi, resized, input_size_);
    
    // Normalize pixel values to [0, 1]
    cv::Mat normalized;
    resized.convertTo(normalized, CV_32F, 1.0/255.0);
    
    return normalized;
}

float FaceSwapDetector::predict_face_swap(const cv::Mat& face_image) {
    if (!model_loaded_) {
        // Return random prediction for testing
        static std::random_device rd;
        static std::mt19937 gen(rd());
        static std::uniform_real_distribution<float> dis(0.0f, 1.0f);
        return dis(gen);
    }
    
#ifdef USE_TENSORFLOW
    if (tf_session_) {
        // Convert OpenCV Mat to TensorFlow Tensor
        std::vector<float> input_data = DetectionUtils::mat_to_vector(face_image);
        
        tensorflow::Tensor input_tensor(tensorflow::DT_FLOAT, 
            tensorflow::TensorShape({1, input_size_.height, input_size_.width, 3}));
        
        auto input_tensor_mapped = input_tensor.tensor<float, 4>();
        for (int i = 0; i < input_data.size(); ++i) {
            input_tensor_mapped(0, i / (input_size_.width * 3), (i / 3) % input_size_.width, i % 3) = input_data[i];
        }
        
        // Run inference
        std::vector<tensorflow::Tensor> outputs;
        tensorflow::Status status = tf_session_->Run(
            {{input_layer_name_, input_tensor}},
            {output_layer_name_},
            {},
            &outputs
        );
        
        if (status.ok() && !outputs.empty()) {
            auto output_tensor = outputs[0].tensor<float, 2>();
            return output_tensor(0, 0);
        }
    }
#endif
    
    // Fallback: simple heuristic based on image properties
    cv::Scalar mean_color = cv::mean(face_image);
    cv::Mat gray;
    cv::cvtColor(face_image, gray, cv::COLOR_BGR2GRAY);
    
    cv::Scalar mean_intensity, std_intensity;
    cv::meanStdDev(gray, mean_intensity, std_intensity);
    
    // Simple heuristic: unusual color distribution might indicate fake
    float color_variance = (mean_color[0] + mean_color[1] + mean_color[2]) / 3.0f;
    float intensity_std = std_intensity[0];
    
    return std::min(1.0f, (color_variance * intensity_std) / 10000.0f);
}

bool FaceSwapDetector::load_model(const std::string& model_path) {
#ifdef USE_TENSORFLOW
    try {
        tensorflow::SessionOptions options;
        tf_session_.reset(tensorflow::NewSession(options));
        
        tensorflow::GraphDef graph_def;
        tensorflow::Status status = tensorflow::ReadBinaryProto(
            tensorflow::Env::Default(), model_path, &graph_def);
        
        if (!status.ok()) {
            std::cerr << "Failed to load model: " << status.ToString() << std::endl;
            return false;
        }
        
        status = tf_session_->Create(graph_def);
        if (!status.ok()) {
            std::cerr << "Failed to create session: " << status.ToString() << std::endl;
            return false;
        }
        
        input_layer_name_ = "input_1";
        output_layer_name_ = "output_1";
        model_loaded_ = true;
        
        std::cout << "TensorFlow model loaded successfully" << std::endl;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Exception loading TensorFlow model: " << e.what() << std::endl;
        return false;
    }
#else
    std::cout << "TensorFlow not available, using dummy model" << std::endl;
    create_dummy_model();
    return true;
#endif
}

void FaceSwapDetector::create_dummy_model() {
    model_loaded_ = true;
    std::cout << "Using dummy face swap detection model" << std::endl;
}
