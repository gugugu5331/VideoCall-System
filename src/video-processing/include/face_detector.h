#pragma once

#include "common.h"

namespace VideoProcessing {

class FaceDetector {
public:
    FaceDetector();
    ~FaceDetector();

    // 初始化面部检测器
    bool Initialize(const std::string& cascade_path = "", const std::string& landmarks_model_path = "");
    
    // 释放资源
    void Release();
    
    // 检测面部
    std::vector<FaceInfo> DetectFaces(const cv::Mat& image);
    
    // 检测面部关键点
    bool DetectLandmarks(const cv::Mat& image, FaceInfo& face_info);
    
    // 设置检测参数
    void SetScaleFactor(double scale_factor) { scale_factor_ = scale_factor; }
    void SetMinNeighbors(int min_neighbors) { min_neighbors_ = min_neighbors; }
    void SetMinSize(const cv::Size& min_size) { min_size_ = min_size; }
    void SetMaxSize(const cv::Size& max_size) { max_size_ = max_size; }
    
    // 获取检测参数
    double GetScaleFactor() const { return scale_factor_; }
    int GetMinNeighbors() const { return min_neighbors_; }
    cv::Size GetMinSize() const { return min_size_; }
    cv::Size GetMaxSize() const { return max_size_; }
    
    // 面部区域提取
    cv::Mat ExtractFaceRegion(const cv::Mat& image, const FaceInfo& face_info, float padding = 0.2f);
    
    // 面部对齐
    cv::Mat AlignFace(const cv::Mat& image, const FaceInfo& face_info, const cv::Size& output_size = cv::Size(256, 256));
    
    // 面部特征点分析
    struct FaceFeatures {
        std::vector<cv::Point2f> left_eye;
        std::vector<cv::Point2f> right_eye;
        std::vector<cv::Point2f> nose;
        std::vector<cv::Point2f> mouth;
        std::vector<cv::Point2f> jaw;
        std::vector<cv::Point2f> eyebrows;
        cv::Point2f left_eye_center;
        cv::Point2f right_eye_center;
        cv::Point2f nose_tip;
        cv::Point2f mouth_center;
        float face_angle;
        float eye_distance;
    };
    
    FaceFeatures AnalyzeFaceFeatures(const FaceInfo& face_info);
    
    // 面部姿态估计
    struct FacePose {
        cv::Vec3f rotation;    // 旋转角度 (pitch, yaw, roll)
        cv::Vec3f translation; // 平移向量
        float confidence;
        bool valid;
    };
    
    FacePose EstimateFacePose(const FaceInfo& face_info, const cv::Size& image_size);
    
    // 面部表情识别
    enum class Expression {
        NEUTRAL = 0,
        HAPPY,
        SAD,
        ANGRY,
        SURPRISED,
        DISGUSTED,
        FEARFUL
    };
    
    Expression RecognizeExpression(const cv::Mat& face_region);
    
    // 年龄和性别估计
    struct Demographics {
        int estimated_age;
        float gender_confidence; // 0.0 = female, 1.0 = male
        bool valid;
    };
    
    Demographics EstimateDemographics(const cv::Mat& face_region);
    
    // 面部质量评估
    struct FaceQuality {
        float sharpness;
        float brightness;
        float contrast;
        float symmetry;
        float frontal_score;
        float overall_score;
    };
    
    FaceQuality AssessFaceQuality(const cv::Mat& face_region, const FaceInfo& face_info);
    
    // 多帧跟踪
    void EnableTracking(bool enable) { tracking_enabled_ = enable; }
    bool IsTrackingEnabled() const { return tracking_enabled_; }
    void UpdateTracking(const std::vector<FaceInfo>& current_faces);
    
    // 性能统计
    struct DetectionStats {
        float detection_time;
        float landmarks_time;
        int faces_detected;
        float average_confidence;
    };
    
    DetectionStats GetDetectionStats() const { return stats_; }

private:
    cv::CascadeClassifier face_cascade_;
    cv::Ptr<cv::face::Facemark> facemark_;
    
    // 检测参数
    double scale_factor_;
    int min_neighbors_;
    cv::Size min_size_;
    cv::Size max_size_;
    
    // 跟踪相关
    bool tracking_enabled_;
    std::vector<cv::Ptr<cv::Tracker>> trackers_;
    std::vector<int> face_ids_;
    int next_face_id_;
    
    // 性能统计
    DetectionStats stats_;
    std::chrono::high_resolution_clock::time_point last_detection_time_;
    
    // 内部辅助函数
    bool LoadModels(const std::string& cascade_path, const std::string& landmarks_model_path);
    std::string GetDefaultCascadePath();
    std::string GetDefaultLandmarksPath();
    
    // 面部验证
    bool ValidateFace(const cv::Rect& face_rect, const cv::Size& image_size);
    float CalculateFaceConfidence(const cv::Rect& face_rect, const cv::Mat& image);
    
    // 关键点处理
    std::vector<cv::Point2f> FilterLandmarks(const std::vector<cv::Point2f>& landmarks);
    std::vector<cv::Point2f> SmoothLandmarks(const std::vector<cv::Point2f>& current, 
                                           const std::vector<cv::Point2f>& previous, 
                                           float alpha = 0.7f);
    
    // 几何计算
    float CalculateEyeDistance(const std::vector<cv::Point2f>& landmarks);
    float CalculateFaceAngle(const std::vector<cv::Point2f>& landmarks);
    cv::Point2f CalculateEyeCenter(const std::vector<cv::Point2f>& eye_points);
    
    // 3D模型点（用于姿态估计）
    std::vector<cv::Point3f> model_points_;
    cv::Mat camera_matrix_;
    cv::Mat dist_coeffs_;
    
    void InitializePoseEstimation();
};

} // namespace VideoProcessing
