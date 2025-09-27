#include "face_detector.h"
#include <opencv2/opencv.hpp>
#include <opencv2/objdetect.hpp>
#include <opencv2/imgproc.hpp>
#include <iostream>
#include <filesystem>

FaceDetector::FaceDetector()
    : initialized_(false)
    , enabled_(true)
    , detection_interval_(1)
    , frame_counter_(0)
    , confidence_threshold_(0.5f)
{
}

FaceDetector::~FaceDetector() {
    cleanup();
}

bool FaceDetector::initialize() {
    if (initialized_) {
        return true;
    }

    try {
        // 尝试加载Haar级联分类器
        std::vector<std::string> cascade_paths = {
            "/usr/share/opencv4/haarcascades/haarcascade_frontalface_alt.xml",
            "/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_alt.xml",
            "haarcascade_frontalface_alt.xml",
            "../assets/haarcascade_frontalface_alt.xml"
        };

        bool cascade_loaded = false;
        for (const auto& path : cascade_paths) {
            if (std::filesystem::exists(path)) {
                if (face_cascade_.load(path)) {
                    cascade_loaded = true;
                    std::cout << "Loaded face cascade from: " << path << std::endl;
                    break;
                }
            }
        }

        if (!cascade_loaded) {
            std::cerr << "Warning: Could not load face cascade classifier" << std::endl;
            std::cerr << "Face detection will be disabled" << std::endl;
            enabled_ = false;
        }

        // 尝试初始化DNN人脸检测器（更准确）
        initializeDNNDetector();

        initialized_ = true;
        std::cout << "FaceDetector initialized successfully" << std::endl;
        return true;

    } catch (const std::exception& e) {
        std::cerr << "Error initializing FaceDetector: " << e.what() << std::endl;
        return false;
    }
}

void FaceDetector::cleanup() {
    face_cascade_ = cv::CascadeClassifier();
    dnn_net_ = cv::dnn::Net();
    cached_faces_.clear();
    initialized_ = false;
}

std::vector<FaceInfo> FaceDetector::detectFaces(const cv::Mat& frame) {
    if (!initialized_ || !enabled_ || frame.empty()) {
        return {};
    }

    // 检测间隔控制
    frame_counter_++;
    if (frame_counter_ % detection_interval_ != 0) {
        return cached_faces_; // 返回缓存的结果
    }

    std::vector<FaceInfo> faces;

    try {
        // 优先使用DNN检测器
        if (!dnn_net_.empty()) {
            faces = detectFacesWithDNN(frame);
        } else if (!face_cascade_.empty()) {
            faces = detectFacesWithHaar(frame);
        }

        // 缓存检测结果
        cached_faces_ = faces;

        // 更新跟踪ID
        updateTrackingIds(faces);

    } catch (const std::exception& e) {
        std::cerr << "Error in face detection: " << e.what() << std::endl;
    }

    return faces;
}

std::vector<FaceInfo> FaceDetector::detectFacesWithHaar(const cv::Mat& frame) {
    std::vector<FaceInfo> faces;
    std::vector<cv::Rect> face_rects;

    // 转换为灰度图
    cv::Mat gray;
    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
    cv::equalizeHist(gray, gray);

    // 检测人脸
    face_cascade_.detectMultiScale(gray, face_rects, 1.1, 3, 0, cv::Size(30, 30));

    // 转换为FaceInfo格式
    for (size_t i = 0; i < face_rects.size(); i++) {
        FaceInfo face_info;
        face_info.bounding_box = face_rects[i];
        face_info.confidence = 0.8f; // Haar检测器没有置信度，使用固定值
        face_info.tracking_id = static_cast<int>(i);
        face_info.landmarks = detectLandmarks(gray, face_rects[i]);
        faces.push_back(face_info);
    }

    return faces;
}

std::vector<FaceInfo> FaceDetector::detectFacesWithDNN(const cv::Mat& frame) {
    std::vector<FaceInfo> faces;

    try {
        // 准备输入
        cv::Mat blob;
        cv::dnn::blobFromImage(frame, blob, 1.0, cv::Size(300, 300), cv::Scalar(104, 117, 123));
        dnn_net_.setInput(blob);

        // 前向传播
        cv::Mat detection = dnn_net_.forward();
        cv::Mat detection_mat(detection.size[2], detection.size[3], CV_32F, detection.ptr<float>());

        // 解析检测结果
        for (int i = 0; i < detection_mat.rows; i++) {
            float confidence = detection_mat.at<float>(i, 2);
            
            if (confidence > confidence_threshold_) {
                int x1 = static_cast<int>(detection_mat.at<float>(i, 3) * frame.cols);
                int y1 = static_cast<int>(detection_mat.at<float>(i, 4) * frame.rows);
                int x2 = static_cast<int>(detection_mat.at<float>(i, 5) * frame.cols);
                int y2 = static_cast<int>(detection_mat.at<float>(i, 6) * frame.rows);

                cv::Rect face_rect(x1, y1, x2 - x1, y2 - y1);
                
                // 确保边界框在图像范围内
                face_rect &= cv::Rect(0, 0, frame.cols, frame.rows);
                
                if (face_rect.width > 0 && face_rect.height > 0) {
                    FaceInfo face_info;
                    face_info.bounding_box = face_rect;
                    face_info.confidence = confidence;
                    face_info.tracking_id = i;
                    
                    // 检测关键点
                    cv::Mat gray;
                    cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
                    face_info.landmarks = detectLandmarks(gray, face_rect);
                    
                    faces.push_back(face_info);
                }
            }
        }

    } catch (const std::exception& e) {
        std::cerr << "Error in DNN face detection: " << e.what() << std::endl;
    }

    return faces;
}

std::vector<cv::Point2f> FaceDetector::detectLandmarks(const cv::Mat& gray_frame, const cv::Rect& face_rect) {
    std::vector<cv::Point2f> landmarks;

    // 简化的关键点检测 - 基于人脸区域的几何估计
    // 在实际应用中，这里应该使用专门的关键点检测模型
    
    cv::Point2f center(face_rect.x + face_rect.width / 2.0f, face_rect.y + face_rect.height / 2.0f);
    
    // 估计主要关键点位置
    landmarks.push_back(cv::Point2f(face_rect.x + face_rect.width * 0.3f, face_rect.y + face_rect.height * 0.4f)); // 左眼
    landmarks.push_back(cv::Point2f(face_rect.x + face_rect.width * 0.7f, face_rect.y + face_rect.height * 0.4f)); // 右眼
    landmarks.push_back(cv::Point2f(center.x, face_rect.y + face_rect.height * 0.6f)); // 鼻子
    landmarks.push_back(cv::Point2f(face_rect.x + face_rect.width * 0.3f, face_rect.y + face_rect.height * 0.8f)); // 左嘴角
    landmarks.push_back(cv::Point2f(face_rect.x + face_rect.width * 0.7f, face_rect.y + face_rect.height * 0.8f)); // 右嘴角

    return landmarks;
}

void FaceDetector::updateTrackingIds(std::vector<FaceInfo>& faces) {
    // 简单的跟踪ID分配
    // 在实际应用中，这里应该实现更复杂的跟踪算法
    
    static int next_id = 0;
    
    for (auto& face : faces) {
        // 尝试匹配之前的人脸
        bool matched = false;
        for (const auto& cached_face : cached_faces_) {
            cv::Point2f current_center(face.bounding_box.x + face.bounding_box.width / 2.0f,
                                     face.bounding_box.y + face.bounding_box.height / 2.0f);
            cv::Point2f cached_center(cached_face.bounding_box.x + cached_face.bounding_box.width / 2.0f,
                                    cached_face.bounding_box.y + cached_face.bounding_box.height / 2.0f);
            
            float distance = cv::norm(current_center - cached_center);
            if (distance < 50.0f) { // 距离阈值
                face.tracking_id = cached_face.tracking_id;
                matched = true;
                break;
            }
        }
        
        if (!matched) {
            face.tracking_id = next_id++;
        }
    }
}

void FaceDetector::setEnabled(bool enabled) {
    enabled_ = enabled;
}

void FaceDetector::setDetectionInterval(int interval) {
    detection_interval_ = std::max(1, interval);
}

void FaceDetector::setConfidenceThreshold(float threshold) {
    confidence_threshold_ = std::clamp(threshold, 0.1f, 1.0f);
}

bool FaceDetector::loadModel(const std::string& model_path) {
    try {
        if (model_path.find(".caffemodel") != std::string::npos) {
            // Caffe模型
            std::string config_path = model_path;
            config_path.replace(config_path.find(".caffemodel"), 11, ".prototxt");
            dnn_net_ = cv::dnn::readNetFromCaffe(config_path, model_path);
        } else if (model_path.find(".pb") != std::string::npos) {
            // TensorFlow模型
            dnn_net_ = cv::dnn::readNetFromTensorflow(model_path);
        } else if (model_path.find(".onnx") != std::string::npos) {
            // ONNX模型
            dnn_net_ = cv::dnn::readNetFromONNX(model_path);
        }
        
        if (!dnn_net_.empty()) {
            std::cout << "Loaded DNN model from: " << model_path << std::endl;
            return true;
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Error loading model: " << e.what() << std::endl;
    }
    
    return false;
}

void FaceDetector::initializeDNNDetector() {
    // 尝试加载预训练的DNN模型
    std::vector<std::string> model_paths = {
        "../assets/opencv_face_detector_uint8.pb",
        "opencv_face_detector_uint8.pb",
        "../models/face_detection_yunet_2023mar.onnx"
    };

    for (const auto& path : model_paths) {
        if (std::filesystem::exists(path)) {
            if (loadModel(path)) {
                break;
            }
        }
    }
}

DetectionStats FaceDetector::getStats() const {
    DetectionStats stats;
    stats.total_detections = frame_counter_;
    stats.faces_detected = cached_faces_.size();
    stats.detection_rate = enabled_ ? (100.0f / detection_interval_) : 0.0f;
    stats.average_confidence = 0.0f;
    
    if (!cached_faces_.empty()) {
        float total_confidence = 0.0f;
        for (const auto& face : cached_faces_) {
            total_confidence += face.confidence;
        }
        stats.average_confidence = total_confidence / cached_faces_.size();
    }
    
    return stats;
}
