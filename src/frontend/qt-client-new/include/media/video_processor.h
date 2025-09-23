#pragma once

#include "core/common.h"

namespace VideoCallSystem {

class FilterEngine;
class OpenGLRenderer;

struct FaceInfo {
    QRect boundingBox;
    QList<QPointF> landmarks;
    float confidence = 0.0f;
    int trackingId = -1;
    bool isValid = false;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["bounding_box"] = QJsonArray{boundingBox.x(), boundingBox.y(), boundingBox.width(), boundingBox.height()};
        obj["confidence"] = confidence;
        obj["tracking_id"] = trackingId;
        obj["is_valid"] = isValid;
        
        QJsonArray landmarksArray;
        for (const auto& point : landmarks) {
            landmarksArray.append(QJsonArray{point.x(), point.y()});
        }
        obj["landmarks"] = landmarksArray;
        
        return obj;
    }
    
    static FaceInfo fromJson(const QJsonObject& obj) {
        FaceInfo info;
        QJsonArray bbox = obj["bounding_box"].toArray();
        if (bbox.size() == 4) {
            info.boundingBox = QRect(bbox[0].toInt(), bbox[1].toInt(), bbox[2].toInt(), bbox[3].toInt());
        }
        info.confidence = obj["confidence"].toDouble();
        info.trackingId = obj["tracking_id"].toInt();
        info.isValid = obj["is_valid"].toBool();
        
        QJsonArray landmarksArray = obj["landmarks"].toArray();
        for (const auto& pointArray : landmarksArray) {
            QJsonArray point = pointArray.toArray();
            if (point.size() == 2) {
                info.landmarks.append(QPointF(point[0].toDouble(), point[1].toDouble()));
            }
        }
        
        return info;
    }
};

struct ProcessingStats {
    double fps = 0.0;
    double processingTime = 0.0;
    double filterTime = 0.0;
    double renderTime = 0.0;
    int droppedFrames = 0;
    int processedFrames = 0;
    QDateTime lastUpdate;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["fps"] = fps;
        obj["processing_time"] = processingTime;
        obj["filter_time"] = filterTime;
        obj["render_time"] = renderTime;
        obj["dropped_frames"] = droppedFrames;
        obj["processed_frames"] = processedFrames;
        obj["last_update"] = lastUpdate.toString(Qt::ISODate);
        return obj;
    }
};

class VideoProcessor : public QObject
{
    Q_OBJECT

public:
    explicit VideoProcessor(QObject* parent = nullptr);
    ~VideoProcessor();

    // 初始化和清理
    bool initialize();
    void cleanup();
    bool isInitialized() const { return initialized_; }

    // 处理控制
    void startProcessing();
    void stopProcessing();
    void pauseProcessing();
    void resumeProcessing();
    bool isProcessing() const { return processing_; }
    bool isPaused() const { return paused_; }

    // 输入源管理
    void setInputSource(QObject* source); // QCamera, QVideoSink等
    QObject* inputSource() const { return inputSource_; }
    void setInputFrame(const QVideoFrame& frame);
    void setInputImage(const QImage& image);

    // 输出管理
    QVideoFrame getOutputFrame() const;
    QImage getOutputImage() const;
    void setOutputSink(QObject* sink);

    // 滤镜管理
    void setFilterEngine(FilterEngine* engine);
    FilterEngine* filterEngine() const { return filterEngine_; }
    void setFilterParams(const FilterParams& params);
    FilterParams getFilterParams() const { return currentFilterParams_; }

    // 渲染器管理
    void setOpenGLRenderer(OpenGLRenderer* renderer);
    OpenGLRenderer* openGLRenderer() const { return openGLRenderer_; }

    // 面部检测
    void enableFaceDetection(bool enable);
    bool isFaceDetectionEnabled() const { return faceDetectionEnabled_; }
    QList<FaceInfo> getDetectedFaces() const { return detectedFaces_; }
    void setFaceDetectionModel(const QString& modelPath);

    // 背景处理
    void enableBackgroundReplacement(bool enable);
    bool isBackgroundReplacementEnabled() const { return backgroundReplacementEnabled_; }
    void setBackgroundImage(const QImage& background);
    void setBackgroundVideo(const QString& videoPath);
    void setBackgroundBlur(float intensity);

    // 贴纸和特效
    void addSticker(const QString& name, const QImage& sticker, const QPointF& position);
    void removeSticker(const QString& name);
    void updateStickerPosition(const QString& name, const QPointF& position);
    QStringList getStickerNames() const;

    // 3D效果
    void enable3DEffects(bool enable);
    bool is3DEffectsEnabled() const { return effects3DEnabled_; }
    void set3DModel(const QString& modelPath);
    void update3DModelTransform(const QMatrix4x4& transform);

    // 录制功能
    void startRecording(const QString& outputPath, const QString& codec = "H264");
    void stopRecording();
    bool isRecording() const { return recording_; }
    QString recordingPath() const { return recordingPath_; }

    // 截图功能
    QImage takeScreenshot();
    bool saveScreenshot(const QString& filePath);

    // 性能监控
    ProcessingStats getProcessingStats() const { return processingStats_; }
    void resetProcessingStats();
    void enablePerformanceMonitoring(bool enable);

    // 配置管理
    void setProcessingResolution(const QSize& resolution);
    QSize getProcessingResolution() const { return processingResolution_; }
    
    void setTargetFPS(int fps);
    int getTargetFPS() const { return targetFPS_; }
    
    void setQualityLevel(int level); // 0-100
    int getQualityLevel() const { return qualityLevel_; }

    // 多线程处理
    void setThreadCount(int count);
    int getThreadCount() const { return threadCount_; }
    void enableGPUAcceleration(bool enable);
    bool isGPUAccelerationEnabled() const { return gpuAcceleration_; }

    // 错误处理
    QString lastError() const { return lastError_; }
    void clearError() { lastError_.clear(); }

public slots:
    // 处理槽函数
    void processFrame();
    void onInputFrameReady(const QVideoFrame& frame);
    void onFilterParamsChanged(const FilterParams& params);

    // 面部检测槽函数
    void onFaceDetectionResult(const QList<FaceInfo>& faces);

    // 性能监控槽函数
    void updatePerformanceStats();

signals:
    // 处理状态信号
    void processingStarted();
    void processingStopped();
    void processingPaused();
    void processingResumed();

    // 帧处理信号
    void frameProcessed(const QVideoFrame& frame);
    void imageProcessed(const QImage& image);
    void processingError(const QString& error);

    // 面部检测信号
    void facesDetected(const QList<FaceInfo>& faces);
    void faceTrackingUpdated(const QList<FaceInfo>& faces);

    // 录制信号
    void recordingStarted(const QString& outputPath);
    void recordingStopped();
    void recordingError(const QString& error);

    // 性能信号
    void performanceStatsUpdated(const ProcessingStats& stats);
    void performanceWarning(const QString& warning);

    // 配置信号
    void configurationChanged();

private slots:
    // 内部处理槽
    void onProcessingTimer();
    void onStatsTimer();

private:
    // 初始化函数
    bool initializeOpenCV();
    bool initializeFaceDetection();
    bool initializeRecording();

    // 帧处理函数
    QVideoFrame processVideoFrame(const QVideoFrame& inputFrame);
    QImage processImage(const QImage& inputImage);
    void applyFilters(cv::Mat& frame);
    void applyFaceEffects(cv::Mat& frame, const QList<FaceInfo>& faces);
    void applyBackgroundEffects(cv::Mat& frame);
    void applyStickers(cv::Mat& frame);
    void apply3DEffects(cv::Mat& frame);

    // 面部检测函数
    QList<FaceInfo> detectFaces(const cv::Mat& frame);
    void updateFaceTracking(QList<FaceInfo>& faces);
    void drawFaceInfo(cv::Mat& frame, const QList<FaceInfo>& faces);

    // 背景处理函数
    cv::Mat createBackgroundMask(const cv::Mat& frame);
    void replaceBackground(cv::Mat& frame, const cv::Mat& mask);
    void blurBackground(cv::Mat& frame, const cv::Mat& mask, float intensity);

    // 贴纸处理函数
    void renderStickers(cv::Mat& frame);
    void updateStickerPositions(const QList<FaceInfo>& faces);

    // 录制函数
    bool setupVideoWriter();
    void writeFrameToVideo(const cv::Mat& frame);

    // 性能监控函数
    void updateFPS();
    void updateProcessingTime();
    void checkPerformance();

    // 工具函数
    cv::Mat qImageToCvMat(const QImage& image);
    QImage cvMatToQImage(const cv::Mat& mat);
    QVideoFrame qImageToVideoFrame(const QImage& image);
    QImage videoFrameToQImage(const QVideoFrame& frame);

    // 错误处理
    void setError(const QString& error);

private:
    // 初始化状态
    bool initialized_;
    
    // 处理状态
    bool processing_;
    bool paused_;
    
    // 输入输出
    QObject* inputSource_;
    QObject* outputSink_;
    QVideoFrame currentInputFrame_;
    QVideoFrame currentOutputFrame_;
    
    // 处理组件
    FilterEngine* filterEngine_;
    OpenGLRenderer* openGLRenderer_;
    
    // 滤镜参数
    FilterParams currentFilterParams_;
    
    // 面部检测
    bool faceDetectionEnabled_;
    cv::CascadeClassifier faceClassifier_;
    QList<FaceInfo> detectedFaces_;
    QString faceDetectionModelPath_;
    
    // 背景处理
    bool backgroundReplacementEnabled_;
    QImage backgroundImage_;
    QString backgroundVideoPath_;
    float backgroundBlurIntensity_;
    cv::Mat backgroundMask_;
    
    // 贴纸和特效
    struct StickerInfo {
        QString name;
        QImage image;
        QPointF position;
        QSizeF size;
        bool faceTracking = false;
        int faceId = -1;
    };
    QMap<QString, StickerInfo> stickers_;
    
    // 3D效果
    bool effects3DEnabled_;
    QString model3DPath_;
    QMatrix4x4 model3DTransform_;
    
    // 录制
    bool recording_;
    QString recordingPath_;
    QString recordingCodec_;
    cv::VideoWriter videoWriter_;
    
    // 性能监控
    ProcessingStats processingStats_;
    QTimer* processingTimer_;
    QTimer* statsTimer_;
    bool performanceMonitoring_;
    std::chrono::high_resolution_clock::time_point lastFrameTime_;
    
    // 配置
    QSize processingResolution_;
    int targetFPS_;
    int qualityLevel_;
    int threadCount_;
    bool gpuAcceleration_;
    
    // 错误处理
    QString lastError_;
    
    // 互斥锁
    mutable QMutex processingMutex_;
};

} // namespace VideoCallSystem
