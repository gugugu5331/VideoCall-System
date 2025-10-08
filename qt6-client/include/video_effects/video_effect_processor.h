#ifndef VIDEO_EFFECT_PROCESSOR_H
#define VIDEO_EFFECT_PROCESSOR_H

#include <QObject>
#include <QImage>
#include <QVideoFrame>
#include <memory>
#include <opencv2/opencv.hpp>
#include "video_effects/sticker_overlay.h"

/**
 * @brief VideoEffectProcessor类 - 视频效果处理器
 * 
 * 使用OpenCV实现美颜、虚拟背景等视频效果
 */
class VideoEffectProcessor : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool beautyEnabled READ beautyEnabled WRITE setBeautyEnabled NOTIFY beautyEnabledChanged)
    Q_PROPERTY(bool virtualBackgroundEnabled READ virtualBackgroundEnabled WRITE setVirtualBackgroundEnabled NOTIFY virtualBackgroundEnabledChanged)
    Q_PROPERTY(int beautyLevel READ beautyLevel WRITE setBeautyLevel NOTIFY beautyLevelChanged)
    Q_PROPERTY(int whitenLevel READ whitenLevel WRITE setWhitenLevel NOTIFY whitenLevelChanged)
    Q_PROPERTY(BackgroundMode backgroundMode READ backgroundMode WRITE setBackgroundMode NOTIFY backgroundModeChanged)
    Q_PROPERTY(bool stickerEnabled READ stickerEnabled WRITE setStickerEnabled NOTIFY stickerEnabledChanged)

public:
    enum BackgroundMode {
        None = 0,           // 无背景效果
        Blur = 1,           // 背景模糊
        Replace = 2,        // 背景替换
        GreenScreen = 3     // 绿幕
    };
    Q_ENUM(BackgroundMode)

    explicit VideoEffectProcessor(QObject *parent = nullptr);
    ~VideoEffectProcessor();

    // Properties
    bool beautyEnabled() const { return m_beautyEnabled; }
    bool virtualBackgroundEnabled() const { return m_virtualBackgroundEnabled; }
    int beautyLevel() const { return m_beautyLevel; }
    int whitenLevel() const { return m_whitenLevel; }
    BackgroundMode backgroundMode() const { return m_backgroundMode; }
    bool stickerEnabled() const { return m_stickerEnabled; }

    // Setters
    void setBeautyEnabled(bool enabled);
    void setVirtualBackgroundEnabled(bool enabled);
    void setBeautyLevel(int level);  // 0-100
    void setWhitenLevel(int level);  // 0-100
    void setBackgroundMode(BackgroundMode mode);
    void setStickerEnabled(bool enabled);

    // Background image
    Q_INVOKABLE bool setBackgroundImage(const QString &imagePath);
    Q_INVOKABLE void clearBackgroundImage();

    // Process video frame
    Q_INVOKABLE QVideoFrame processFrame(const QVideoFrame &inputFrame);
    Q_INVOKABLE QImage processImage(const QImage &inputImage);

    // Sticker management
    Q_INVOKABLE StickerOverlay* stickerOverlay() { return m_stickerOverlay.get(); }

signals:
    void beautyEnabledChanged();
    void virtualBackgroundEnabledChanged();
    void beautyLevelChanged();
    void whitenLevelChanged();
    void backgroundModeChanged();
    void stickerEnabledChanged();
    void processingError(const QString &error);

private:
    // OpenCV processing functions
    cv::Mat qImageToMat(const QImage &image);
    QImage matToQImage(const cv::Mat &mat);
    cv::Mat qVideoFrameToMat(const QVideoFrame &frame);
    QVideoFrame matToQVideoFrame(const cv::Mat &mat, const QVideoFrame &originalFrame);

    // Beauty filters
    cv::Mat applyBeautyFilter(const cv::Mat &input);
    cv::Mat applySkinSmoothing(const cv::Mat &input, int level);
    cv::Mat applyWhitening(const cv::Mat &input, int level);
    cv::Mat applyFaceSlimming(const cv::Mat &input);

    // Face detection
    bool detectFaces(const cv::Mat &input, std::vector<cv::Rect> &faces);
    cv::Mat createSkinMask(const cv::Mat &input, const std::vector<cv::Rect> &faces);

    // Virtual background
    cv::Mat applyVirtualBackground(const cv::Mat &input);
    cv::Mat applyBackgroundBlur(const cv::Mat &input);
    cv::Mat applyBackgroundReplace(const cv::Mat &input);
    cv::Mat segmentPerson(const cv::Mat &input, cv::Mat &mask);

    // Background segmentation
    void initBackgroundSegmentation();
    cv::Mat createPersonMask(const cv::Mat &input);

private:
    // Effect settings
    bool m_beautyEnabled;
    bool m_virtualBackgroundEnabled;
    bool m_stickerEnabled;
    int m_beautyLevel;      // 0-100
    int m_whitenLevel;      // 0-100
    BackgroundMode m_backgroundMode;

    // Background image
    cv::Mat m_backgroundImage;
    bool m_hasBackgroundImage;

    // OpenCV models
    cv::CascadeClassifier m_faceCascade;
    cv::Ptr<cv::BackgroundSubtractor> m_backgroundSubtractor;
    
    // DNN models for person segmentation
    cv::dnn::Net m_segmentationNet;
    bool m_segmentationModelLoaded;

    // Cache
    cv::Mat m_previousFrame;
    cv::Mat m_previousMask;
    int m_frameCount;

    // Sticker overlay
    std::unique_ptr<StickerOverlay> m_stickerOverlay;
    std::vector<cv::Rect> m_lastDetectedFaces;
};

#endif // VIDEO_EFFECT_PROCESSOR_H

