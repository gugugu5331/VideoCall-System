#pragma once

#include "common.h"

namespace VideoProcessing {

class FilterManager {
public:
    FilterManager();
    ~FilterManager();

    // 初始化滤镜管理器
    bool Initialize();
    
    // 释放资源
    void Release();
    
    // 应用滤镜到OpenCV图像
    cv::Mat ApplyFilter(const cv::Mat& input, FilterType filter_type, const EffectParams& params);
    
    // 设置当前滤镜
    void SetCurrentFilter(FilterType filter_type);
    FilterType GetCurrentFilter() const { return current_filter_; }
    
    // 设置滤镜参数
    void SetFilterParams(const EffectParams& params);
    const EffectParams& GetFilterParams() const { return filter_params_; }
    
    // 预设滤镜效果
    void ApplyBeautyFilter(cv::Mat& image, float intensity = 0.5f);
    void ApplyVintageFilter(cv::Mat& image, float intensity = 0.7f);
    void ApplyCartoonFilter(cv::Mat& image, float intensity = 0.8f);
    void ApplySketchFilter(cv::Mat& image, float intensity = 0.9f);
    void ApplyNeonFilter(cv::Mat& image, float intensity = 0.6f);
    void ApplyThermalFilter(cv::Mat& image, float intensity = 0.8f);
    void ApplyNightVisionFilter(cv::Mat& image, float intensity = 0.7f);
    
    // 几何变换滤镜
    cv::Mat ApplyFisheyeEffect(const cv::Mat& input, float strength = 0.5f);
    cv::Mat ApplyMirrorEffect(const cv::Mat& input, bool horizontal = true);
    cv::Mat ApplyPixelateEffect(const cv::Mat& input, int pixel_size = 8);
    
    // 颜色调整
    cv::Mat AdjustBrightness(const cv::Mat& input, float brightness);
    cv::Mat AdjustContrast(const cv::Mat& input, float contrast);
    cv::Mat AdjustSaturation(const cv::Mat& input, float saturation);
    cv::Mat AdjustHue(const cv::Mat& input, float hue_shift);
    cv::Mat AdjustGamma(const cv::Mat& input, float gamma);
    cv::Mat AdjustColorBalance(const cv::Mat& input, const glm::vec3& balance);
    
    // 噪声和模糊
    cv::Mat ApplyGaussianBlur(const cv::Mat& input, float sigma);
    cv::Mat ApplyMotionBlur(const cv::Mat& input, int size, float angle);
    cv::Mat ApplyNoiseReduction(const cv::Mat& input, float strength);
    cv::Mat AddNoise(const cv::Mat& input, float intensity);
    
    // 边缘检测和锐化
    cv::Mat ApplySharpen(const cv::Mat& input, float strength);
    cv::Mat ApplyUnsharpMask(const cv::Mat& input, float amount, float radius);
    cv::Mat ApplyEdgeDetection(const cv::Mat& input, float threshold);
    cv::Mat ApplyEmboss(const cv::Mat& input, float strength);
    
    // 艺术效果
    cv::Mat ApplyOilPainting(const cv::Mat& input, int size, int dynRatio);
    cv::Mat ApplyWatercolor(const cv::Mat& input, float sigma_s, float sigma_r);
    cv::Mat ApplyPencilSketch(const cv::Mat& input, float sigma_s, float sigma_r, float shade_factor);
    
    // 获取所有可用滤镜
    std::vector<FilterType> GetAvailableFilters() const;
    
    // 滤镜预览（缩略图）
    cv::Mat GenerateFilterPreview(const cv::Mat& input, FilterType filter_type, int preview_size = 128);

private:
    FilterType current_filter_;
    EffectParams filter_params_;
    
    // 内部辅助函数
    cv::Mat ConvertToHSV(const cv::Mat& input);
    cv::Mat ConvertFromHSV(const cv::Mat& hsv);
    cv::Mat ApplyColorMap(const cv::Mat& input, int colormap);
    cv::Mat CreateLookupTable(float gamma);
    
    // 美颜相关
    cv::Mat SkinSmoothing(const cv::Mat& input, float intensity);
    cv::Mat EyeBrightening(const cv::Mat& input, const std::vector<cv::Point2f>& eye_points, float intensity);
    cv::Mat TeethWhitening(const cv::Mat& input, const std::vector<cv::Point2f>& mouth_points, float intensity);
    
    // 性能优化
    cv::Mat cached_lut_;
    bool lut_valid_;
    void InvalidateLUT();
    
    // 参数验证
    bool ValidateParams(const EffectParams& params);
    EffectParams ClampParams(const EffectParams& params);
};

} // namespace VideoProcessing
