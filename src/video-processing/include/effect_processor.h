#pragma once

#include "common.h"
#include <opencv2/opencv.hpp>
#include <vector>
#include <string>
#include <functional>
#include <unordered_map>

// 粒子结构
struct Particle {
    cv::Point2f position;
    cv::Point2f velocity;
    float life;
    float max_life;
    float size;
    cv::Scalar color;
};

// 自定义特效函数类型
using EffectFunction = std::function<void(cv::Mat&, const std::vector<FaceInfo>&, float)>;

class EffectProcessor {
public:
    EffectProcessor();
    ~EffectProcessor();

    // 初始化和清理
    bool initialize();
    void cleanup();

    // 主要处理函数
    void processEffects(cv::Mat& frame, const std::vector<FaceInfo>& faces);

    // 特效控制
    void setParticleCount(int count);
    void setAnimationSpeed(float speed);

    // 自定义特效管理
    void addCustomEffect(const std::string& name, EffectFunction effect_func);
    void removeCustomEffect(const std::string& name);
    void applyCustomEffect(cv::Mat& frame, const std::vector<FaceInfo>& faces, const std::string& effect_name);

    // 获取可用特效列表
    std::vector<std::string> getAvailableEffects() const;

private:
    bool initialized_;
    
    // 粒子系统
    std::vector<Particle> particles_;
    int particle_count_;
    
    // 动画参数
    float animation_time_;
    float animation_speed_;
    
    // 自定义特效
    std::unordered_map<std::string, EffectFunction> custom_effects_;

    // 特效处理函数
    void applyParticleEffects(cv::Mat& frame, const FaceInfo& face);
    void applyAnimatedStickers(cv::Mat& frame, const FaceInfo& face);
    void applyFaceDistortion(cv::Mat& frame, const FaceInfo& face);
    void applyScreenEffects(cv::Mat& frame);

    // 粒子系统函数
    void initializeParticleSystem();
    void updateParticles(const FaceInfo& face);
    void drawParticle(cv::Mat& frame, const Particle& particle);
};
