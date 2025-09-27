#include "media/video_effects_processor.h"
#include <QTimer>
#include <QThread>
#include <QMutexLocker>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QStandardPaths>
#include <QDir>
#include <QDebug>
#include <chrono>

namespace VideoCallSystem {

VideoEffectsProcessor::VideoEffectsProcessor(QObject* parent)
    : QObject(parent)
    , initialized_(false)
    , realTimeProcessing_(true)
    , faceDetectionEnabled_(true)
    , backgroundReplacement_(false)
    , effects3D_(false)
    , recording_(false)
    , gpuAcceleration_(true)
    , currentFilter_(VideoProcessing::FilterType::NONE)
    , filterIntensity_(0.5f)
    , backgroundBlurIntensity_(0.7f)
    , processingResolution_(640, 480)
    , originalResolution_(1920, 1080)
    , targetFPS_(30)
    , frameCount_(0)
    , processingThread_(nullptr)
{
    // 初始化性能计时器
    performanceTimer_ = new QTimer(this);
    connect(performanceTimer_, &QTimer::timeout, this, &VideoEffectsProcessor::updatePerformanceMetrics);
    performanceTimer_->start(1000); // 每秒更新一次性能指标
    
    // 设置配置文件路径
    QString configDir = QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation);
    QDir().mkpath(configDir);
    configFilePath_ = configDir + "/video_effects_config.json";
    
    // 加载配置
    loadConfiguration();
}

VideoEffectsProcessor::~VideoEffectsProcessor()
{
    cleanup();
}

bool VideoEffectsProcessor::initialize()
{
    if (initialized_) {
        return true;
    }
    
    qDebug() << "Initializing VideoEffectsProcessor...";
    
    try {
        // 初始化视频处理器
        if (!initializeVideoProcessor()) {
            handleProcessingError("Failed to initialize video processor");
            return false;
        }
        
        // 初始化OpenCV
        if (!initializeOpenCV()) {
            handleProcessingError("Failed to initialize OpenCV");
            return false;
        }
        
        // 加载默认资源
        if (!loadDefaultAssets()) {
            qWarning() << "Failed to load some default assets, continuing...";
        }
        
        // 创建处理线程
        processingThread_ = new QThread(this);
        processingThread_->start();
        
        initialized_ = true;
        qDebug() << "VideoEffectsProcessor initialized successfully";
        
        emit processingStarted();
        return true;
        
    } catch (const std::exception& e) {
        handleProcessingError(QString("Initialization exception: %1").arg(e.what()));
        return false;
    }
}

void VideoEffectsProcessor::cleanup()
{
    if (!initialized_) {
        return;
    }
    
    qDebug() << "Cleaning up VideoEffectsProcessor...";
    
    // 停止录制
    if (recording_) {
        stopRecording();
    }
    
    // 停止性能监控
    if (performanceTimer_) {
        performanceTimer_->stop();
    }
    
    // 清理处理线程
    if (processingThread_) {
        processingThread_->quit();
        processingThread_->wait(3000);
        processingThread_ = nullptr;
    }
    
    // 清理视频处理组件
    videoProcessor_.reset();
    filterManager_.reset();
    faceDetector_.reset();
    textureManager_.reset();
    
    // 保存配置
    saveConfiguration();
    
    initialized_ = false;
    emit processingStopped();
    
    qDebug() << "VideoEffectsProcessor cleanup completed";
}

QVideoFrame VideoEffectsProcessor::processFrame(const QVideoFrame& inputFrame)
{
    if (!initialized_ || !realTimeProcessing_) {
        return inputFrame;
    }
    
    QMutexLocker locker(&processingMutex_);
    
    auto startTime = std::chrono::high_resolution_clock::now();
    
    try {
        // 转换为OpenCV格式
        cv::Mat cvFrame = qVideoFrameToCvMat(inputFrame);
        if (cvFrame.empty()) {
            handleProcessingError("Failed to convert QVideoFrame to cv::Mat");
            return inputFrame;
        }
        
        // 记录原始尺寸
        originalResolution_ = QSize(cvFrame.cols, cvFrame.rows);
        
        // 处理帧
        cv::Mat processedFrame = processOpenCVFrame(cvFrame);
        
        // 转换回QVideoFrame
        QVideoFrame outputFrame = cvMatToQVideoFrame(processedFrame);
        
        // 更新性能统计
        auto endTime = std::chrono::high_resolution_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::microseconds>(endTime - startTime);
        performanceMetrics_.processingTimeMs = duration.count() / 1000.0;
        
        frameCount_++;
        performanceMetrics_.totalFrames++;
        
        emit frameProcessed(outputFrame);
        return outputFrame;
        
    } catch (const std::exception& e) {
        handleProcessingError(QString("Frame processing exception: %1").arg(e.what()));
        performanceMetrics_.droppedFrames++;
        return inputFrame;
    }
}

QImage VideoEffectsProcessor::processImage(const QImage& inputImage)
{
    if (!initialized_ || !realTimeProcessing_) {
        return inputImage;
    }
    
    QMutexLocker locker(&processingMutex_);
    
    try {
        // 转换为OpenCV格式
        cv::Mat cvFrame = qImageToCvMat(inputImage);
        if (cvFrame.empty()) {
            handleProcessingError("Failed to convert QImage to cv::Mat");
            return inputImage;
        }
        
        // 处理帧
        cv::Mat processedFrame = processOpenCVFrame(cvFrame);
        
        // 转换回QImage
        QImage outputImage = cvMatToQImage(processedFrame);
        
        emit imageProcessed(outputImage);
        return outputImage;
        
    } catch (const std::exception& e) {
        handleProcessingError(QString("Image processing exception: %1").arg(e.what()));
        return inputImage;
    }
}

void VideoEffectsProcessor::setFilter(VideoProcessing::FilterType filterType)
{
    if (currentFilter_ != filterType) {
        currentFilter_ = filterType;
        
        if (filterManager_) {
            // 这里需要调用实际的滤镜管理器API
            // filterManager_->setActiveFilter(filterType);
        }
        
        emit filterChanged(filterType);
        qDebug() << "Filter changed to:" << static_cast<int>(filterType);
    }
}

VideoProcessing::FilterType VideoEffectsProcessor::getCurrentFilter() const
{
    return currentFilter_;
}

void VideoEffectsProcessor::setFilterIntensity(float intensity)
{
    filterIntensity_ = qBound(0.0f, intensity, 1.0f);
    
    if (filterManager_) {
        // 这里需要调用实际的滤镜管理器API
        // filterManager_->setIntensity(filterIntensity_);
    }
    
    qDebug() << "Filter intensity set to:" << filterIntensity_;
}

float VideoEffectsProcessor::getFilterIntensity() const
{
    return filterIntensity_;
}

void VideoEffectsProcessor::applyBeautyFilter(float intensity)
{
    setFilter(VideoProcessing::FilterType::BEAUTY);
    setFilterIntensity(intensity);
}

void VideoEffectsProcessor::applyCartoonFilter(float intensity)
{
    setFilter(VideoProcessing::FilterType::CARTOON);
    setFilterIntensity(intensity);
}

void VideoEffectsProcessor::applyVintageFilter(float intensity)
{
    setFilter(VideoProcessing::FilterType::VINTAGE);
    setFilterIntensity(intensity);
}

void VideoEffectsProcessor::applySketchFilter(float intensity)
{
    setFilter(VideoProcessing::FilterType::SKETCH);
    setFilterIntensity(intensity);
}

void VideoEffectsProcessor::clearAllFilters()
{
    setFilter(VideoProcessing::FilterType::NONE);
    removeAllStickers();
    removeBackground();
}

bool VideoEffectsProcessor::loadSticker(const QString& name, const QString& filePath)
{
    try {
        cv::Mat stickerImage = cv::imread(filePath.toStdString(), cv::IMREAD_UNCHANGED);
        if (stickerImage.empty()) {
            handleProcessingError(QString("Failed to load sticker: %1").arg(filePath));
            return false;
        }
        
        stickerTextures_[name] = stickerImage;
        qDebug() << "Loaded sticker:" << name << "from" << filePath;
        return true;
        
    } catch (const std::exception& e) {
        handleProcessingError(QString("Exception loading sticker %1: %2").arg(name, e.what()));
        return false;
    }
}

void VideoEffectsProcessor::setActiveSticker(const QString& name)
{
    if (stickerTextures_.contains(name)) {
        activeSticker_ = name;
        emit stickerChanged(name);
        qDebug() << "Active sticker set to:" << name;
    } else {
        qWarning() << "Sticker not found:" << name;
    }
}

void VideoEffectsProcessor::removeSticker(const QString& name)
{
    stickerTextures_.remove(name);
    if (activeSticker_ == name) {
        activeSticker_.clear();
        emit stickerChanged("");
    }
    qDebug() << "Removed sticker:" << name;
}

void VideoEffectsProcessor::removeAllStickers()
{
    stickerTextures_.clear();
    activeSticker_.clear();
    emit stickerChanged("");
    qDebug() << "Removed all stickers";
}

QStringList VideoEffectsProcessor::getAvailableStickers() const
{
    return stickerTextures_.keys();
}

void VideoEffectsProcessor::enableFaceDetection(bool enable)
{
    if (faceDetectionEnabled_ != enable) {
        faceDetectionEnabled_ = enable;
        qDebug() << "Face detection" << (enable ? "enabled" : "disabled");
    }
}

bool VideoEffectsProcessor::isFaceDetectionEnabled() const
{
    return faceDetectionEnabled_;
}

QList<FaceInfo> VideoEffectsProcessor::getDetectedFaces() const
{
    // 这里需要从实际的面部检测器获取结果
    // 暂时返回空列表
    return QList<FaceInfo>();
}

void VideoEffectsProcessor::enableBackgroundReplacement(bool enable)
{
    if (backgroundReplacement_ != enable) {
        backgroundReplacement_ = enable;
        emit backgroundChanged(enable);
        qDebug() << "Background replacement" << (enable ? "enabled" : "disabled");
    }
}

bool VideoEffectsProcessor::isBackgroundReplacementEnabled() const
{
    return backgroundReplacement_;
}

void VideoEffectsProcessor::setBackgroundImage(const QString& imagePath)
{
    try {
        backgroundTexture_ = cv::imread(imagePath.toStdString());
        if (!backgroundTexture_.empty()) {
            backgroundImagePath_ = imagePath;
            qDebug() << "Background image set to:" << imagePath;
        } else {
            handleProcessingError(QString("Failed to load background image: %1").arg(imagePath));
        }
    } catch (const std::exception& e) {
        handleProcessingError(QString("Exception loading background image: %1").arg(e.what()));
    }
}

void VideoEffectsProcessor::setBackgroundBlur(float intensity)
{
    backgroundBlurIntensity_ = qBound(0.0f, intensity, 1.0f);
    qDebug() << "Background blur intensity set to:" << backgroundBlurIntensity_;
}

void VideoEffectsProcessor::removeBackground()
{
    backgroundReplacement_ = false;
    backgroundTexture_ = cv::Mat();
    backgroundImagePath_.clear();
    emit backgroundChanged(false);
    qDebug() << "Background removed";
}

// 核心处理函数实现
cv::Mat VideoEffectsProcessor::processOpenCVFrame(const cv::Mat& inputFrame)
{
    cv::Mat processedFrame = inputFrame.clone();

    try {
        // 调整处理分辨率以提高性能
        cv::Mat resizedFrame = resizeForProcessing(processedFrame);

        // 应用效果链
        applyEffectsChain(resizedFrame);

        // 恢复原始尺寸
        cv::Mat finalFrame = restoreOriginalSize(resizedFrame, inputFrame.size());

        return finalFrame;

    } catch (const std::exception& e) {
        qWarning() << "Error in processOpenCVFrame:" << e.what();
        return inputFrame;
    }
}

void VideoEffectsProcessor::applyEffectsChain(cv::Mat& frame)
{
    auto startTime = std::chrono::high_resolution_clock::now();

    // 1. 面部检测
    std::vector<VideoProcessing::FaceInfo> faces;
    if (faceDetectionEnabled_) {
        updateFaceTracking(frame);
        // faces = faceDetector_->detectFaces(frame); // 需要实现
    }

    auto faceDetectionTime = std::chrono::high_resolution_clock::now();
    performanceMetrics_.faceDetectionTimeMs =
        std::chrono::duration_cast<std::chrono::microseconds>(faceDetectionTime - startTime).count() / 1000.0;

    // 2. 应用滤镜
    if (currentFilter_ != VideoProcessing::FilterType::NONE && filterManager_) {
        // filterManager_->applyFilter(frame, currentFilter_, filterIntensity_); // 需要实现
    }

    auto filterTime = std::chrono::high_resolution_clock::now();
    performanceMetrics_.filterTimeMs =
        std::chrono::duration_cast<std::chrono::microseconds>(filterTime - faceDetectionTime).count() / 1000.0;

    // 3. 背景处理
    if (backgroundReplacement_) {
        applyBackgroundEffects(frame);
    }

    // 4. 渲染贴纸
    if (!activeSticker_.isEmpty() && !faces.empty()) {
        renderStickers(frame, faces);
    }

    auto renderTime = std::chrono::high_resolution_clock::now();
    performanceMetrics_.renderTimeMs =
        std::chrono::duration_cast<std::chrono::microseconds>(renderTime - filterTime).count() / 1000.0;
}

void VideoEffectsProcessor::updateFaceTracking(cv::Mat& frame)
{
    // 面部检测和跟踪逻辑
    if (faceDetector_) {
        // auto detectedFaces = faceDetector_->detectFaces(frame);
        // emit facesDetected(convertToQtFaceInfo(detectedFaces));
    }
}

void VideoEffectsProcessor::renderStickers(cv::Mat& frame, const std::vector<VideoProcessing::FaceInfo>& faces)
{
    if (activeSticker_.isEmpty() || !stickerTextures_.contains(activeSticker_)) {
        return;
    }

    cv::Mat sticker = stickerTextures_[activeSticker_];

    // 为每个检测到的面部渲染贴纸
    for (const auto& face : faces) {
        try {
            // 计算贴纸位置和大小
            int stickerWidth = static_cast<int>(face.boundingBox.width * 1.2);
            int stickerHeight = static_cast<int>(face.boundingBox.height * 1.2);

            // 调整贴纸大小
            cv::Mat resizedSticker;
            cv::resize(sticker, resizedSticker, cv::Size(stickerWidth, stickerHeight));

            // 计算贴纸位置（居中在面部上方）
            int x = face.boundingBox.x - (stickerWidth - face.boundingBox.width) / 2;
            int y = face.boundingBox.y - stickerHeight / 3;

            // 确保贴纸在图像范围内
            x = std::max(0, std::min(x, frame.cols - stickerWidth));
            y = std::max(0, std::min(y, frame.rows - stickerHeight));

            // 创建ROI并混合贴纸
            cv::Rect roi(x, y, stickerWidth, stickerHeight);
            if (roi.x + roi.width <= frame.cols && roi.y + roi.height <= frame.rows) {
                cv::Mat frameROI = frame(roi);

                // 如果贴纸有alpha通道，使用alpha混合
                if (resizedSticker.channels() == 4) {
                    // Alpha混合逻辑
                    for (int i = 0; i < resizedSticker.rows; ++i) {
                        for (int j = 0; j < resizedSticker.cols; ++j) {
                            cv::Vec4b stickerPixel = resizedSticker.at<cv::Vec4b>(i, j);
                            float alpha = stickerPixel[3] / 255.0f;

                            if (alpha > 0.1f) {
                                cv::Vec3b& framePixel = frameROI.at<cv::Vec3b>(i, j);
                                framePixel[0] = static_cast<uchar>(framePixel[0] * (1 - alpha) + stickerPixel[0] * alpha);
                                framePixel[1] = static_cast<uchar>(framePixel[1] * (1 - alpha) + stickerPixel[1] * alpha);
                                framePixel[2] = static_cast<uchar>(framePixel[2] * (1 - alpha) + stickerPixel[2] * alpha);
                            }
                        }
                    }
                } else {
                    // 简单覆盖
                    cv::Mat stickerBGR;
                    cv::cvtColor(resizedSticker, stickerBGR, cv::COLOR_BGRA2BGR);
                    stickerBGR.copyTo(frameROI);
                }
            }

        } catch (const std::exception& e) {
            qWarning() << "Error rendering sticker:" << e.what();
        }
    }
}

void VideoEffectsProcessor::applyBackgroundEffects(cv::Mat& frame)
{
    try {
        if (!backgroundTexture_.empty()) {
            // 背景替换逻辑
            cv::Mat background;
            cv::resize(backgroundTexture_, background, frame.size());

            // 简单的背景替换（这里需要更复杂的分割算法）
            // 实际应用中应该使用深度学习模型进行人像分割
            cv::Mat mask;
            cv::cvtColor(frame, mask, cv::COLOR_BGR2GRAY);
            cv::threshold(mask, mask, 100, 255, cv::THRESH_BINARY);

            // 应用背景
            frame.copyTo(background, mask);
            background.copyTo(frame);
        } else if (backgroundBlurIntensity_ > 0.0f) {
            // 背景模糊
            cv::Mat blurred;
            int kernelSize = static_cast<int>(backgroundBlurIntensity_ * 50) | 1; // 确保是奇数
            cv::GaussianBlur(frame, blurred, cv::Size(kernelSize, kernelSize), 0);

            // 这里应该只对背景区域应用模糊，保持人物清晰
            // 简化版本：直接应用模糊
            blurred.copyTo(frame);
        }

    } catch (const std::exception& e) {
        qWarning() << "Error applying background effects:" << e.what();
    }
}
