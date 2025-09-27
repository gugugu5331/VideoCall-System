#pragma once

#include "core/common.h"
#include "media/video_processor.h"

// 集成独立的视频处理模块
#include "../../../../video-processing/include/video_processor.h"
#include "../../../../video-processing/include/filter_manager.h"
#include "../../../../video-processing/include/face_detector.h"
#include "../../../../video-processing/include/texture_manager.h"

namespace VideoCallSystem {

/**
 * @brief 视频特效处理器 - 集成滤镜、贴图、人脸检测到WebRTC视频流
 * 
 * 这个类作为独立视频处理模块和Qt WebRTC系统之间的桥梁
 */
class VideoEffectsProcessor : public QObject
{
    Q_OBJECT

public:
    explicit VideoEffectsProcessor(QObject* parent = nullptr);
    ~VideoEffectsProcessor();

    // 初始化和清理
    bool initialize();
    void cleanup();
    bool isInitialized() const { return initialized_; }

    // 主要处理接口 - WebRTC集成点
    QVideoFrame processFrame(const QVideoFrame& inputFrame);
    QImage processImage(const QImage& inputImage);
    
    // 实时处理控制
    void enableRealTimeProcessing(bool enable);
    bool isRealTimeProcessingEnabled() const { return realTimeProcessing_; }
    
    // 滤镜控制
    void setFilter(VideoProcessing::FilterType filterType);
    VideoProcessing::FilterType getCurrentFilter() const;
    void setFilterIntensity(float intensity);
    float getFilterIntensity() const;
    
    // 预设滤镜快速切换
    void applyBeautyFilter(float intensity = 0.7f);
    void applyCartoonFilter(float intensity = 0.8f);
    void applyVintageFilter(float intensity = 0.6f);
    void applySketchFilter(float intensity = 0.9f);
    void clearAllFilters();
    
    // 贴图管理
    bool loadSticker(const QString& name, const QString& filePath);
    void setActiveSticker(const QString& name);
    void removeSticker(const QString& name);
    void removeAllStickers();
    QStringList getAvailableStickers() const;
    QString getActiveSticker() const { return activeSticker_; }
    
    // 面部检测和跟踪
    void enableFaceDetection(bool enable);
    bool isFaceDetectionEnabled() const { return faceDetectionEnabled_; }
    QList<FaceInfo> getDetectedFaces() const;
    void setFaceDetectionSensitivity(float sensitivity);
    
    // 背景处理
    void enableBackgroundReplacement(bool enable);
    bool isBackgroundReplacementEnabled() const { return backgroundReplacement_; }
    void setBackgroundImage(const QString& imagePath);
    void setBackgroundBlur(float intensity);
    void removeBackground();
    
    // 3D特效
    void enable3DEffects(bool enable);
    bool is3DEffectsEnabled() const { return effects3D_; }
    void load3DModel(const QString& modelPath);
    void update3DTransform(const QMatrix4x4& transform);
    
    // 性能优化
    void setProcessingResolution(const QSize& resolution);
    QSize getProcessingResolution() const { return processingResolution_; }
    void setTargetFPS(int fps);
    int getTargetFPS() const { return targetFPS_; }
    void enableGPUAcceleration(bool enable);
    bool isGPUAccelerationEnabled() const { return gpuAcceleration_; }
    
    // 性能监控
    struct PerformanceMetrics {
        double averageFPS = 0.0;
        double processingTimeMs = 0.0;
        double filterTimeMs = 0.0;
        double faceDetectionTimeMs = 0.0;
        double renderTimeMs = 0.0;
        int droppedFrames = 0;
        int totalFrames = 0;
        QDateTime lastUpdate;
    };
    
    PerformanceMetrics getPerformanceMetrics() const { return performanceMetrics_; }
    void resetPerformanceMetrics();
    void enablePerformanceMonitoring(bool enable);
    
    // 预设配置
    struct EffectsPreset {
        QString name;
        VideoProcessing::FilterType filterType;
        float filterIntensity;
        QString stickerName;
        bool faceDetection;
        bool backgroundBlur;
        float backgroundBlurIntensity;
        QJsonObject customParams;
    };
    
    void savePreset(const QString& name, const EffectsPreset& preset);
    void loadPreset(const QString& name);
    void deletePreset(const QString& name);
    QStringList getAvailablePresets() const;
    
    // 录制和截图
    void startRecording(const QString& outputPath);
    void stopRecording();
    bool isRecording() const { return recording_; }
    QImage takeScreenshot();
    bool saveScreenshot(const QString& filePath);

public slots:
    // WebRTC集成槽函数
    void onVideoFrameReady(const QVideoFrame& frame);
    void onCameraFrameReady(const QVideoFrame& frame);
    
    // 配置更新槽函数
    void onFilterChanged(int filterType);
    void onFilterIntensityChanged(double intensity);
    void onStickerChanged(const QString& stickerName);
    void onFaceDetectionToggled(bool enabled);
    void onBackgroundToggled(bool enabled);

signals:
    // 处理完成信号
    void frameProcessed(const QVideoFrame& processedFrame);
    void imageProcessed(const QImage& processedImage);
    
    // 检测结果信号
    void facesDetected(const QList<FaceInfo>& faces);
    void faceTrackingUpdated(const QList<FaceInfo>& faces);
    
    // 性能信号
    void performanceUpdated(const PerformanceMetrics& metrics);
    void performanceWarning(const QString& warning);
    
    // 状态信号
    void processingStarted();
    void processingStopped();
    void filterChanged(VideoProcessing::FilterType filterType);
    void stickerChanged(const QString& stickerName);
    void backgroundChanged(bool enabled);
    
    // 错误信号
    void processingError(const QString& error);
    void initializationError(const QString& error);

private slots:
    void updatePerformanceMetrics();
    void checkPerformanceWarnings();

private:
    // 初始化函数
    bool initializeVideoProcessor();
    bool initializeOpenCV();
    bool loadDefaultAssets();
    
    // 核心处理函数
    cv::Mat processOpenCVFrame(const cv::Mat& inputFrame);
    void applyEffectsChain(cv::Mat& frame);
    void updateFaceTracking(cv::Mat& frame);
    void renderStickers(cv::Mat& frame, const std::vector<VideoProcessing::FaceInfo>& faces);
    void applyBackgroundEffects(cv::Mat& frame);
    
    // 格式转换函数
    cv::Mat qVideoFrameToCvMat(const QVideoFrame& frame);
    cv::Mat qImageToCvMat(const QImage& image);
    QVideoFrame cvMatToQVideoFrame(const cv::Mat& mat);
    QImage cvMatToQImage(const cv::Mat& mat);
    
    // 性能优化函数
    cv::Mat resizeForProcessing(const cv::Mat& input);
    cv::Mat restoreOriginalSize(const cv::Mat& processed, const cv::Size& originalSize);
    void optimizeProcessingPipeline();
    
    // 配置管理
    void loadConfiguration();
    void saveConfiguration();
    QJsonObject presetToJson(const EffectsPreset& preset);
    EffectsPreset presetFromJson(const QJsonObject& json);
    
    // 错误处理
    void handleProcessingError(const QString& error);
    void logPerformanceWarning(const QString& warning);

private:
    // 初始化状态
    bool initialized_;
    
    // 核心处理组件
    std::unique_ptr<VideoProcessing::VideoProcessor> videoProcessor_;
    std::unique_ptr<VideoProcessing::FilterManager> filterManager_;
    std::unique_ptr<VideoProcessing::FaceDetector> faceDetector_;
    std::unique_ptr<VideoProcessing::TextureManager> textureManager_;
    
    // 处理状态
    bool realTimeProcessing_;
    bool faceDetectionEnabled_;
    bool backgroundReplacement_;
    bool effects3D_;
    bool recording_;
    bool gpuAcceleration_;
    
    // 当前设置
    VideoProcessing::FilterType currentFilter_;
    float filterIntensity_;
    QString activeSticker_;
    QString backgroundImagePath_;
    float backgroundBlurIntensity_;
    
    // 性能配置
    QSize processingResolution_;
    QSize originalResolution_;
    int targetFPS_;
    
    // 性能监控
    PerformanceMetrics performanceMetrics_;
    QTimer* performanceTimer_;
    std::chrono::high_resolution_clock::time_point lastFrameTime_;
    int frameCount_;
    
    // 预设管理
    QMap<QString, EffectsPreset> presets_;
    QString configFilePath_;
    
    // 资源管理
    QMap<QString, cv::Mat> stickerTextures_;
    cv::Mat backgroundTexture_;
    
    // 线程安全
    mutable QMutex processingMutex_;
    QThread* processingThread_;
    
    // 错误处理
    QString lastError_;
};

} // namespace VideoCallSystem
