#pragma once

#include <iostream>
#include <vector>
#include <string>
#include <memory>
#include <map>
#include <functional>
#include <chrono>
#include <thread>
#include <mutex>
#include <atomic>

// OpenCV
#include <opencv2/opencv.hpp>
#include <opencv2/imgproc.hpp>
#include <opencv2/objdetect.hpp>
#include <opencv2/face.hpp>

// OpenGL
#include <GL/glew.h>
#include <GLFW/glfw3.h>
#include <glm/glm.hpp>
#include <glm/gtc/matrix_transform.hpp>
#include <glm/gtc/type_ptr.hpp>

// 常量定义
namespace VideoProcessing {
    // 窗口配置
    constexpr int WINDOW_WIDTH = 1280;
    constexpr int WINDOW_HEIGHT = 720;
    constexpr const char* WINDOW_TITLE = "Video Processing - OpenCV + OpenGL";

    // 视频配置
    constexpr int VIDEO_WIDTH = 640;
    constexpr int VIDEO_HEIGHT = 480;
    constexpr int VIDEO_FPS = 30;

    // 滤镜类型
    enum class FilterType {
        NONE = 0,
        BLUR,
        SHARPEN,
        EDGE_DETECTION,
        EMBOSS,
        SEPIA,
        VINTAGE,
        BEAUTY,
        CARTOON,
        SKETCH,
        NEON,
        THERMAL,
        NIGHT_VISION,
        FISHEYE,
        MIRROR,
        PIXELATE
    };

    // 贴图类型
    enum class TextureType {
        NONE = 0,
        FACE_STICKER,
        BACKGROUND,
        OVERLAY,
        PARTICLE_EFFECT,
        MASK,
        FRAME
    };

    // 渲染模式
    enum class RenderMode {
        NORMAL = 0,
        WIREFRAME,
        POINT_CLOUD,
        TEXTURED,
        LIT,
        UNLIT
    };

    // 效果强度
    struct EffectParams {
        float intensity = 1.0f;
        float brightness = 0.0f;
        float contrast = 1.0f;
        float saturation = 1.0f;
        float hue = 0.0f;
        float gamma = 1.0f;
        glm::vec3 color_balance = glm::vec3(1.0f);
        bool enabled = true;
    };

    // 面部检测结果
    struct FaceInfo {
        cv::Rect face_rect;
        std::vector<cv::Point2f> landmarks;
        float confidence = 0.0f;
        bool valid = false;
    };

    // 纹理信息
    struct TextureInfo {
        GLuint texture_id = 0;
        int width = 0;
        int height = 0;
        GLenum format = GL_RGB;
        std::string name;
        bool loaded = false;
    };

    // 着色器信息
    struct ShaderInfo {
        GLuint program_id = 0;
        std::string vertex_path;
        std::string fragment_path;
        std::map<std::string, GLint> uniforms;
        bool compiled = false;
    };

    // 性能统计
    struct PerformanceStats {
        float fps = 0.0f;
        float frame_time = 0.0f;
        float cpu_usage = 0.0f;
        float gpu_usage = 0.0f;
        size_t memory_usage = 0;
        std::chrono::high_resolution_clock::time_point last_update;
    };

    // 工具函数
    inline std::string FilterTypeToString(FilterType type) {
        switch (type) {
            case FilterType::NONE: return "None";
            case FilterType::BLUR: return "Blur";
            case FilterType::SHARPEN: return "Sharpen";
            case FilterType::EDGE_DETECTION: return "Edge Detection";
            case FilterType::EMBOSS: return "Emboss";
            case FilterType::SEPIA: return "Sepia";
            case FilterType::VINTAGE: return "Vintage";
            case FilterType::BEAUTY: return "Beauty";
            case FilterType::CARTOON: return "Cartoon";
            case FilterType::SKETCH: return "Sketch";
            case FilterType::NEON: return "Neon";
            case FilterType::THERMAL: return "Thermal";
            case FilterType::NIGHT_VISION: return "Night Vision";
            case FilterType::FISHEYE: return "Fisheye";
            case FilterType::MIRROR: return "Mirror";
            case FilterType::PIXELATE: return "Pixelate";
            default: return "Unknown";
        }
    }

    inline glm::vec3 HSVtoRGB(float h, float s, float v) {
        float c = v * s;
        float x = c * (1 - abs(fmod(h / 60.0f, 2) - 1));
        float m = v - c;
        
        glm::vec3 rgb;
        if (h >= 0 && h < 60) {
            rgb = glm::vec3(c, x, 0);
        } else if (h >= 60 && h < 120) {
            rgb = glm::vec3(x, c, 0);
        } else if (h >= 120 && h < 180) {
            rgb = glm::vec3(0, c, x);
        } else if (h >= 180 && h < 240) {
            rgb = glm::vec3(0, x, c);
        } else if (h >= 240 && h < 300) {
            rgb = glm::vec3(x, 0, c);
        } else {
            rgb = glm::vec3(c, 0, x);
        }
        
        return rgb + glm::vec3(m);
    }

    // 错误检查宏
    #define CHECK_GL_ERROR() \
        do { \
            GLenum error = glGetError(); \
            if (error != GL_NO_ERROR) { \
                std::cerr << "OpenGL Error: " << error << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            } \
        } while(0)

    #define CHECK_CV_ERROR(expr) \
        do { \
            try { \
                expr; \
            } catch (const cv::Exception& e) { \
                std::cerr << "OpenCV Error: " << e.what() << " at " << __FILE__ << ":" << __LINE__ << std::endl; \
            } \
        } while(0)
}
