#include "effect_processor.h"
#include <opencv2/opencv.hpp>
#include <opencv2/imgproc.hpp>
#include <iostream>
#include <random>

EffectProcessor::EffectProcessor()
    : initialized_(false)
    , particle_count_(100)
    , animation_speed_(1.0f)
{
}

EffectProcessor::~EffectProcessor() {
    cleanup();
}

bool EffectProcessor::initialize() {
    if (initialized_) {
        return true;
    }

    try {
        // 初始化粒子系统
        initializeParticleSystem();
        
        // 初始化动画参数
        animation_time_ = 0.0f;
        
        initialized_ = true;
        std::cout << "EffectProcessor initialized successfully" << std::endl;
        return true;
        
    } catch (const std::exception& e) {
        std::cerr << "Error initializing EffectProcessor: " << e.what() << std::endl;
        return false;
    }
}

void EffectProcessor::cleanup() {
    particles_.clear();
    initialized_ = false;
}

void EffectProcessor::processEffects(cv::Mat& frame, const std::vector<FaceInfo>& faces) {
    if (!initialized_ || frame.empty()) {
        return;
    }

    try {
        // 更新动画时间
        animation_time_ += 0.016f * animation_speed_; // 假设60FPS

        // 应用各种特效
        for (const auto& face : faces) {
            applyParticleEffects(frame, face);
            applyAnimatedStickers(frame, face);
            applyFaceDistortion(frame, face);
        }

        // 应用全局特效
        applyScreenEffects(frame);

    } catch (const std::exception& e) {
        std::cerr << "Error processing effects: " << e.what() << std::endl;
    }
}

void EffectProcessor::applyParticleEffects(cv::Mat& frame, const FaceInfo& face) {
    // 更新粒子位置
    updateParticles(face);

    // 渲染粒子
    for (const auto& particle : particles_) {
        if (particle.life > 0.0f) {
            drawParticle(frame, particle);
        }
    }
}

void EffectProcessor::applyAnimatedStickers(cv::Mat& frame, const FaceInfo& face) {
    // 计算动画参数
    float scale_factor = 1.0f + 0.1f * std::sin(animation_time_ * 2.0f);
    float rotation_angle = std::sin(animation_time_) * 5.0f; // ±5度摆动

    // 应用动画变换到贴纸
    if (!face.landmarks.empty()) {
        cv::Point2f center = face.landmarks[0]; // 假设第一个关键点是中心
        
        // 绘制动画贴纸（简单的圆形示例）
        int radius = static_cast<int>(face.bounding_box.width * 0.1f * scale_factor);
        cv::Scalar color(0, 255, 255); // 黄色
        
        cv::circle(frame, cv::Point(static_cast<int>(center.x), static_cast<int>(center.y)), 
                  radius, color, -1);
    }
}

void EffectProcessor::applyFaceDistortion(cv::Mat& frame, const FaceInfo& face) {
    // 简单的面部变形效果
    cv::Rect face_rect = face.bounding_box;
    
    // 确保区域在图像范围内
    face_rect &= cv::Rect(0, 0, frame.cols, frame.rows);
    if (face_rect.width <= 0 || face_rect.height <= 0) {
        return;
    }

    cv::Mat face_roi = frame(face_rect);
    cv::Mat distorted_face;

    // 应用轻微的鱼眼效果
    cv::Point2f center(face_roi.cols / 2.0f, face_roi.rows / 2.0f);
    float max_radius = std::min(face_roi.cols, face_roi.rows) / 2.0f;
    
    cv::Mat map_x(face_roi.size(), CV_32FC1);
    cv::Mat map_y(face_roi.size(), CV_32FC1);

    for (int y = 0; y < face_roi.rows; y++) {
        for (int x = 0; x < face_roi.cols; x++) {
            float dx = x - center.x;
            float dy = y - center.y;
            float distance = std::sqrt(dx * dx + dy * dy);
            
            if (distance < max_radius) {
                float distortion_factor = 1.0f + 0.2f * std::sin(animation_time_) * 
                                         (1.0f - distance / max_radius);
                
                map_x.at<float>(y, x) = center.x + dx * distortion_factor;
                map_y.at<float>(y, x) = center.y + dy * distortion_factor;
            } else {
                map_x.at<float>(y, x) = x;
                map_y.at<float>(y, x) = y;
            }
        }
    }

    cv::remap(face_roi, distorted_face, map_x, map_y, cv::INTER_LINEAR);
    distorted_face.copyTo(face_roi);
}

void EffectProcessor::applyScreenEffects(cv::Mat& frame) {
    // 应用全屏特效，如闪烁、色彩变化等
    
    // 示例：周期性的色调变化
    float hue_shift = std::sin(animation_time_ * 0.5f) * 10.0f;
    
    if (std::abs(hue_shift) > 1.0f) {
        cv::Mat hsv;
        cv::cvtColor(frame, hsv, cv::COLOR_BGR2HSV);
        
        std::vector<cv::Mat> channels;
        cv::split(hsv, channels);
        
        // 调整色调通道
        channels[0] += hue_shift;
        
        cv::merge(channels, hsv);
        cv::cvtColor(hsv, frame, cv::COLOR_HSV2BGR);
    }
}

void EffectProcessor::initializeParticleSystem() {
    particles_.clear();
    particles_.reserve(particle_count_);

    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_real_distribution<float> pos_dist(-50.0f, 50.0f);
    std::uniform_real_distribution<float> vel_dist(-2.0f, 2.0f);
    std::uniform_real_distribution<float> life_dist(1.0f, 3.0f);

    for (int i = 0; i < particle_count_; i++) {
        Particle particle;
        particle.position = cv::Point2f(pos_dist(gen), pos_dist(gen));
        particle.velocity = cv::Point2f(vel_dist(gen), vel_dist(gen));
        particle.life = life_dist(gen);
        particle.max_life = particle.life;
        particle.size = 2.0f + pos_dist(gen) * 0.1f;
        particle.color = cv::Scalar(
            std::abs(static_cast<int>(pos_dist(gen) * 5)) % 256,
            std::abs(static_cast<int>(pos_dist(gen) * 5)) % 256,
            std::abs(static_cast<int>(pos_dist(gen) * 5)) % 256
        );
        
        particles_.push_back(particle);
    }
}

void EffectProcessor::updateParticles(const FaceInfo& face) {
    cv::Point2f face_center(
        face.bounding_box.x + face.bounding_box.width / 2.0f,
        face.bounding_box.y + face.bounding_box.height / 2.0f
    );

    for (auto& particle : particles_) {
        // 更新粒子生命周期
        particle.life -= 0.016f;
        
        if (particle.life <= 0.0f) {
            // 重新初始化粒子
            std::random_device rd;
            std::mt19937 gen(rd());
            std::uniform_real_distribution<float> pos_dist(-50.0f, 50.0f);
            std::uniform_real_distribution<float> vel_dist(-2.0f, 2.0f);
            
            particle.position = face_center + cv::Point2f(pos_dist(gen), pos_dist(gen));
            particle.velocity = cv::Point2f(vel_dist(gen), vel_dist(gen));
            particle.life = particle.max_life;
        } else {
            // 更新粒子位置
            particle.position += particle.velocity;
            
            // 添加重力效果
            particle.velocity.y += 0.1f;
            
            // 添加一些随机扰动
            std::random_device rd;
            std::mt19937 gen(rd());
            std::uniform_real_distribution<float> noise_dist(-0.1f, 0.1f);
            
            particle.velocity.x += noise_dist(gen);
            particle.velocity.y += noise_dist(gen);
        }
    }
}

void EffectProcessor::drawParticle(cv::Mat& frame, const Particle& particle) {
    if (particle.position.x < 0 || particle.position.x >= frame.cols ||
        particle.position.y < 0 || particle.position.y >= frame.rows) {
        return;
    }

    // 计算透明度基于生命周期
    float alpha = particle.life / particle.max_life;
    cv::Scalar color = particle.color * alpha;

    // 绘制粒子
    cv::Point center(static_cast<int>(particle.position.x), static_cast<int>(particle.position.y));
    int radius = static_cast<int>(particle.size * alpha);
    
    if (radius > 0) {
        cv::circle(frame, center, radius, color, -1);
        
        // 添加发光效果
        cv::circle(frame, center, radius + 1, color * 0.5, 1);
    }
}

void EffectProcessor::setParticleCount(int count) {
    particle_count_ = std::clamp(count, 10, 1000);
    initializeParticleSystem();
}

void EffectProcessor::setAnimationSpeed(float speed) {
    animation_speed_ = std::clamp(speed, 0.1f, 5.0f);
}

void EffectProcessor::addCustomEffect(const std::string& name, EffectFunction effect_func) {
    custom_effects_[name] = effect_func;
}

void EffectProcessor::removeCustomEffect(const std::string& name) {
    auto it = custom_effects_.find(name);
    if (it != custom_effects_.end()) {
        custom_effects_.erase(it);
    }
}

void EffectProcessor::applyCustomEffect(cv::Mat& frame, const std::vector<FaceInfo>& faces, const std::string& effect_name) {
    auto it = custom_effects_.find(effect_name);
    if (it != custom_effects_.end()) {
        it->second(frame, faces, animation_time_);
    }
}

std::vector<std::string> EffectProcessor::getAvailableEffects() const {
    std::vector<std::string> effects = {
        "Particles", "Animated Stickers", "Face Distortion", "Screen Effects"
    };
    
    // 添加自定义特效
    for (const auto& effect_pair : custom_effects_) {
        effects.push_back(effect_pair.first);
    }
    
    return effects;
}
