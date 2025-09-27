#include <opencv2/opencv.hpp>
#include <opencv2/imgproc.hpp>
#include <opencv2/objdetect.hpp>
#include <iostream>
#include <vector>
#include <chrono>

// 简化的滤镜类型
enum class SimpleFilterType {
    NONE = 0,
    BLUR,
    SHARPEN,
    EDGE_DETECTION,
    SEPIA,
    GRAYSCALE,
    BEAUTY,
    CARTOON
};

// 简化的滤镜管理器
class SimpleFilterManager {
public:
    static void applyFilter(cv::Mat& frame, SimpleFilterType filter_type, float intensity = 1.0f) {
        switch (filter_type) {
            case SimpleFilterType::BLUR:
                applyBlur(frame, intensity);
                break;
            case SimpleFilterType::SHARPEN:
                applySharpen(frame, intensity);
                break;
            case SimpleFilterType::EDGE_DETECTION:
                applyEdgeDetection(frame, intensity);
                break;
            case SimpleFilterType::SEPIA:
                applySepia(frame, intensity);
                break;
            case SimpleFilterType::GRAYSCALE:
                applyGrayscale(frame, intensity);
                break;
            case SimpleFilterType::BEAUTY:
                applyBeauty(frame, intensity);
                break;
            case SimpleFilterType::CARTOON:
                applyCartoon(frame, intensity);
                break;
            default:
                break;
        }
    }

private:
    static void applyBlur(cv::Mat& frame, float intensity) {
        int kernel_size = static_cast<int>(5 + intensity * 10);
        if (kernel_size % 2 == 0) kernel_size++;
        cv::GaussianBlur(frame, frame, cv::Size(kernel_size, kernel_size), 0);
    }

    static void applySharpen(cv::Mat& frame, float intensity) {
        cv::Mat kernel = (cv::Mat_<float>(3, 3) << 
            0, -1 * intensity, 0,
            -1 * intensity, 1 + 4 * intensity, -1 * intensity,
            0, -1 * intensity, 0);
        cv::filter2D(frame, frame, -1, kernel);
    }

    static void applyEdgeDetection(cv::Mat& frame, float intensity) {
        cv::Mat gray, edges;
        cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
        cv::Canny(gray, edges, 50 * intensity, 150 * intensity);
        cv::cvtColor(edges, edges, cv::COLOR_GRAY2BGR);
        cv::addWeighted(frame, 1.0f - intensity, edges, intensity, 0, frame);
    }

    static void applySepia(cv::Mat& frame, float intensity) {
        cv::Mat sepia;
        cv::transform(frame, sepia, cv::Matx34f(
            0.272, 0.534, 0.131, 0,
            0.349, 0.686, 0.168, 0,
            0.393, 0.769, 0.189, 0
        ));
        cv::addWeighted(frame, 1.0f - intensity, sepia, intensity, 0, frame);
    }

    static void applyGrayscale(cv::Mat& frame, float intensity) {
        cv::Mat gray;
        cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
        cv::cvtColor(gray, gray, cv::COLOR_GRAY2BGR);
        cv::addWeighted(frame, 1.0f - intensity, gray, intensity, 0, frame);
    }

    static void applyBeauty(cv::Mat& frame, float intensity) {
        cv::Mat beauty;
        cv::bilateralFilter(frame, beauty, 15, 50, 50);
        beauty.convertTo(beauty, -1, 1.0, 10 * intensity);
        cv::addWeighted(frame, 1.0f - intensity * 0.7f, beauty, intensity * 0.7f, 0, frame);
    }

    static void applyCartoon(cv::Mat& frame, float intensity) {
        cv::Mat gray, edges, cartoon;
        cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
        cv::adaptiveThreshold(gray, edges, 255, cv::ADAPTIVE_THRESH_MEAN_C, cv::THRESH_BINARY, 7, 7);
        cv::bilateralFilter(frame, cartoon, 15, 50, 50);
        cv::cvtColor(edges, edges, cv::COLOR_GRAY2BGR);
        cv::bitwise_and(cartoon, edges, cartoon);
        cv::addWeighted(frame, 1.0f - intensity, cartoon, intensity, 0, frame);
    }
};

// 简化的人脸检测器
class SimpleFaceDetector {
private:
    cv::CascadeClassifier face_cascade_;
    bool initialized_;

public:
    SimpleFaceDetector() : initialized_(false) {}

    bool initialize() {
        // 尝试加载Haar级联分类器
        std::vector<std::string> cascade_paths = {
            "haarcascade_frontalface_alt.xml",
            "/usr/share/opencv4/haarcascades/haarcascade_frontalface_alt.xml",
            "/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_alt.xml",
            "C:/opencv/sources/data/haarcascades/haarcascade_frontalface_alt.xml"
        };

        for (const auto& path : cascade_paths) {
            if (face_cascade_.load(path)) {
                initialized_ = true;
                std::cout << "Loaded face cascade from: " << path << std::endl;
                return true;
            }
        }

        std::cout << "Warning: Could not load face cascade classifier" << std::endl;
        std::cout << "Face detection will be disabled" << std::endl;
        return false;
    }

    std::vector<cv::Rect> detectFaces(const cv::Mat& frame) {
        std::vector<cv::Rect> faces;
        if (!initialized_ || frame.empty()) {
            return faces;
        }

        cv::Mat gray;
        cv::cvtColor(frame, gray, cv::COLOR_BGR2GRAY);
        cv::equalizeHist(gray, gray);

        face_cascade_.detectMultiScale(gray, faces, 1.1, 3, 0, cv::Size(30, 30));
        return faces;
    }

    void drawFaces(cv::Mat& frame, const std::vector<cv::Rect>& faces) {
        for (const auto& face : faces) {
            cv::rectangle(frame, face, cv::Scalar(0, 255, 0), 2);
            
            // 添加简单的贴纸效果（圆圈代表眼镜）
            cv::Point center(face.x + face.width / 2, face.y + face.height / 3);
            int radius = face.width / 8;
            cv::circle(frame, cv::Point(center.x - face.width / 4, center.y), radius, cv::Scalar(255, 255, 0), 2);
            cv::circle(frame, cv::Point(center.x + face.width / 4, center.y), radius, cv::Scalar(255, 255, 0), 2);
            cv::line(frame, cv::Point(center.x - face.width / 8, center.y), 
                    cv::Point(center.x + face.width / 8, center.y), cv::Scalar(255, 255, 0), 2);
        }
    }
};

int main() {
    std::cout << "=== 简化视频处理系统测试 ===" << std::endl;
    std::cout << "OpenCV 版本: " << CV_VERSION << std::endl;

    // 初始化摄像头
    cv::VideoCapture cap(0);
    if (!cap.isOpened()) {
        std::cerr << "错误: 无法打开摄像头" << std::endl;
        return -1;
    }

    // 设置摄像头参数
    cap.set(cv::CAP_PROP_FRAME_WIDTH, 640);
    cap.set(cv::CAP_PROP_FRAME_HEIGHT, 480);
    cap.set(cv::CAP_PROP_FPS, 30);

    // 初始化人脸检测器
    SimpleFaceDetector face_detector;
    bool face_detection_enabled = face_detector.initialize();

    // 当前滤镜设置
    SimpleFilterType current_filter = SimpleFilterType::NONE;
    float filter_intensity = 1.0f;
    bool show_faces = false;

    std::cout << "系统初始化完成!" << std::endl;
    std::cout << "控制键:" << std::endl;
    std::cout << "  ESC - 退出" << std::endl;
    std::cout << "  SPACE - 截图" << std::endl;
    std::cout << "  1 - 模糊滤镜" << std::endl;
    std::cout << "  2 - 锐化滤镜" << std::endl;
    std::cout << "  3 - 边缘检测" << std::endl;
    std::cout << "  4 - 复古滤镜" << std::endl;
    std::cout << "  5 - 灰度滤镜" << std::endl;
    std::cout << "  6 - 美颜滤镜" << std::endl;
    std::cout << "  7 - 卡通滤镜" << std::endl;
    std::cout << "  0 - 移除滤镜" << std::endl;
    std::cout << "  F - 切换人脸检测" << std::endl;
    std::cout << "  + - 增加滤镜强度" << std::endl;
    std::cout << "  - - 减少滤镜强度" << std::endl;

    cv::Mat frame;
    int frame_count = 0;
    auto start_time = std::chrono::steady_clock::now();

    while (true) {
        cap >> frame;
        if (frame.empty()) {
            continue;
        }

        // 应用滤镜
        if (current_filter != SimpleFilterType::NONE) {
            SimpleFilterManager::applyFilter(frame, current_filter, filter_intensity);
        }

        // 人脸检测和贴纸
        if (face_detection_enabled && show_faces) {
            auto faces = face_detector.detectFaces(frame);
            face_detector.drawFaces(frame, faces);
        }

        // 显示信息
        std::string info = "Filter: " + std::to_string(static_cast<int>(current_filter)) + 
                          " | Intensity: " + std::to_string(filter_intensity);
        cv::putText(frame, info, cv::Point(10, 30), cv::FONT_HERSHEY_SIMPLEX, 0.7, cv::Scalar(0, 255, 0), 2);

        // 显示FPS
        frame_count++;
        if (frame_count % 30 == 0) {
            auto current_time = std::chrono::steady_clock::now();
            auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(current_time - start_time);
            float fps = 30000.0f / duration.count();
            std::cout << "FPS: " << fps << " | 滤镜: " << static_cast<int>(current_filter) 
                     << " | 强度: " << filter_intensity << std::endl;
            start_time = current_time;
        }

        cv::imshow("简化视频处理系统", frame);

        // 处理键盘输入
        char key = cv::waitKey(1) & 0xFF;
        if (key == 27) { // ESC
            break;
        } else if (key == ' ') { // SPACE - 截图
            std::string filename = "screenshot_" + std::to_string(frame_count) + ".jpg";
            cv::imwrite(filename, frame);
            std::cout << "截图已保存: " << filename << std::endl;
        } else if (key >= '0' && key <= '7') { // 数字键 - 滤镜
            current_filter = static_cast<SimpleFilterType>(key - '0');
            std::cout << "应用滤镜: " << static_cast<int>(current_filter) << std::endl;
        } else if (key == 'f' || key == 'F') { // 切换人脸检测
            show_faces = !show_faces;
            std::cout << "人脸检测: " << (show_faces ? "开启" : "关闭") << std::endl;
        } else if (key == '+' || key == '=') { // 增加强度
            filter_intensity = std::min(2.0f, filter_intensity + 0.1f);
            std::cout << "滤镜强度: " << filter_intensity << std::endl;
        } else if (key == '-') { // 减少强度
            filter_intensity = std::max(0.1f, filter_intensity - 0.1f);
            std::cout << "滤镜强度: " << filter_intensity << std::endl;
        }
    }

    cap.release();
    cv::destroyAllWindows();

    std::cout << "视频处理系统已退出" << std::endl;
    return 0;
}
