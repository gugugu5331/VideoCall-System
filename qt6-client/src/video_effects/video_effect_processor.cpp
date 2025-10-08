#include "video_effects/video_effect_processor.h"
#include <QDebug>
#include <QFile>
#include <opencv2/imgproc.hpp>
#include <opencv2/objdetect.hpp>
#include <opencv2/dnn.hpp>
#include <opencv2/video.hpp>

VideoEffectProcessor::VideoEffectProcessor(QObject *parent)
    : QObject(parent)
    , m_beautyEnabled(false)
    , m_virtualBackgroundEnabled(false)
    , m_stickerEnabled(false)
    , m_beautyLevel(50)
    , m_whitenLevel(30)
    , m_backgroundMode(BackgroundMode::None)
    , m_hasBackgroundImage(false)
    , m_segmentationModelLoaded(false)
    , m_frameCount(0)
{
    // Initialize sticker overlay
    m_stickerOverlay = std::make_unique<StickerOverlay>(this);
    // Load face detection cascade
    QString cascadePath = ":/models/haarcascade_frontalface_default.xml";
    if (QFile::exists(cascadePath)) {
        if (!m_faceCascade.load(cascadePath.toStdString())) {
            qWarning() << "Failed to load face cascade from resources, trying system path";
            // Try system OpenCV data path
            std::string systemPath = cv::samples::findFile("haarcascade_frontalface_default.xml");
            if (!m_faceCascade.load(systemPath)) {
                qWarning() << "Failed to load face cascade classifier";
            }
        }
    }

    // Initialize background subtractor
    m_backgroundSubtractor = cv::createBackgroundSubtractorMOG2(500, 16, true);

    // Try to load segmentation model (optional)
    // You can download models from: https://github.com/opencv/opencv/wiki/TensorFlow-Object-Detection-API
    QString modelPath = ":/models/frozen_inference_graph.pb";
    QString configPath = ":/models/ssd_mobilenet_v2_coco.pbtxt";
    
    if (QFile::exists(modelPath) && QFile::exists(configPath)) {
        try {
            m_segmentationNet = cv::dnn::readNetFromTensorflow(
                modelPath.toStdString(),
                configPath.toStdString()
            );
            m_segmentationModelLoaded = true;
            qDebug() << "Segmentation model loaded successfully";
        } catch (const cv::Exception &e) {
            qWarning() << "Failed to load segmentation model:" << e.what();
        }
    }
}

VideoEffectProcessor::~VideoEffectProcessor()
{
}

void VideoEffectProcessor::setBeautyEnabled(bool enabled)
{
    if (m_beautyEnabled != enabled) {
        m_beautyEnabled = enabled;
        emit beautyEnabledChanged();
    }
}

void VideoEffectProcessor::setVirtualBackgroundEnabled(bool enabled)
{
    if (m_virtualBackgroundEnabled != enabled) {
        m_virtualBackgroundEnabled = enabled;
        emit virtualBackgroundEnabledChanged();
    }
}

void VideoEffectProcessor::setBeautyLevel(int level)
{
    level = qBound(0, level, 100);
    if (m_beautyLevel != level) {
        m_beautyLevel = level;
        emit beautyLevelChanged();
    }
}

void VideoEffectProcessor::setWhitenLevel(int level)
{
    level = qBound(0, level, 100);
    if (m_whitenLevel != level) {
        m_whitenLevel = level;
        emit whitenLevelChanged();
    }
}

void VideoEffectProcessor::setBackgroundMode(BackgroundMode mode)
{
    if (m_backgroundMode != mode) {
        m_backgroundMode = mode;
        emit backgroundModeChanged();
    }
}

void VideoEffectProcessor::setStickerEnabled(bool enabled)
{
    if (m_stickerEnabled != enabled) {
        m_stickerEnabled = enabled;
        m_stickerOverlay->setEnabled(enabled);
        emit stickerEnabledChanged();
    }
}

bool VideoEffectProcessor::setBackgroundImage(const QString &imagePath)
{
    try {
        cv::Mat image = cv::imread(imagePath.toStdString());
        if (image.empty()) {
            qWarning() << "Failed to load background image:" << imagePath;
            return false;
        }
        
        m_backgroundImage = image;
        m_hasBackgroundImage = true;
        qDebug() << "Background image loaded:" << imagePath;
        return true;
    } catch (const cv::Exception &e) {
        qWarning() << "OpenCV error loading background:" << e.what();
        emit processingError(QString("Failed to load background: %1").arg(e.what()));
        return false;
    }
}

void VideoEffectProcessor::clearBackgroundImage()
{
    m_backgroundImage.release();
    m_hasBackgroundImage = false;
}

QVideoFrame VideoEffectProcessor::processFrame(const QVideoFrame &inputFrame)
{
    if (!inputFrame.isValid()) {
        return inputFrame;
    }

    try {
        // Convert QVideoFrame to cv::Mat
        cv::Mat mat = qVideoFrameToMat(inputFrame);
        if (mat.empty()) {
            return inputFrame;
        }

        // Apply effects
        cv::Mat processed = mat.clone();

        // Apply virtual background first (if enabled)
        if (m_virtualBackgroundEnabled && m_backgroundMode != BackgroundMode::None) {
            processed = applyVirtualBackground(processed);
        }

        // Apply beauty filter (if enabled)
        if (m_beautyEnabled) {
            processed = applyBeautyFilter(processed);
        }

        // Apply stickers (if enabled)
        if (m_stickerEnabled) {
            processed = m_stickerOverlay->applyStickers(processed, m_lastDetectedFaces);
        }

        // Convert back to QVideoFrame
        return matToQVideoFrame(processed, inputFrame);

    } catch (const cv::Exception &e) {
        qWarning() << "OpenCV error processing frame:" << e.what();
        emit processingError(QString("Processing error: %1").arg(e.what()));
        return inputFrame;
    }
}

QImage VideoEffectProcessor::processImage(const QImage &inputImage)
{
    if (inputImage.isNull()) {
        return inputImage;
    }

    try {
        // Convert QImage to cv::Mat
        cv::Mat mat = qImageToMat(inputImage);
        if (mat.empty()) {
            return inputImage;
        }

        // Apply effects
        cv::Mat processed = mat.clone();

        // Apply virtual background first (if enabled)
        if (m_virtualBackgroundEnabled && m_backgroundMode != BackgroundMode::None) {
            processed = applyVirtualBackground(processed);
        }

        // Apply beauty filter (if enabled)
        if (m_beautyEnabled) {
            processed = applyBeautyFilter(processed);
        }

        // Apply stickers (if enabled)
        if (m_stickerEnabled) {
            processed = m_stickerOverlay->applyStickers(processed, m_lastDetectedFaces);
        }

        // Convert back to QImage
        return matToQImage(processed);

    } catch (const cv::Exception &e) {
        qWarning() << "OpenCV error processing image:" << e.what();
        emit processingError(QString("Processing error: %1").arg(e.what()));
        return inputImage;
    }
}

// ============================================================================
// OpenCV Conversion Functions
// ============================================================================

cv::Mat VideoEffectProcessor::qImageToMat(const QImage &image)
{
    QImage img = image.convertToFormat(QImage::Format_RGB888);
    cv::Mat mat(img.height(), img.width(), CV_8UC3, 
                const_cast<uchar*>(img.bits()), img.bytesPerLine());
    return mat.clone();
}

QImage VideoEffectProcessor::matToQImage(const cv::Mat &mat)
{
    if (mat.type() == CV_8UC3) {
        QImage image(mat.data, mat.cols, mat.rows, mat.step, QImage::Format_RGB888);
        return image.copy();
    } else if (mat.type() == CV_8UC4) {
        QImage image(mat.data, mat.cols, mat.rows, mat.step, QImage::Format_RGBA8888);
        return image.copy();
    } else if (mat.type() == CV_8UC1) {
        QImage image(mat.data, mat.cols, mat.rows, mat.step, QImage::Format_Grayscale8);
        return image.copy();
    }
    return QImage();
}

cv::Mat VideoEffectProcessor::qVideoFrameToMat(const QVideoFrame &frame)
{
    QVideoFrame f = frame;
    if (!f.map(QVideoFrame::ReadOnly)) {
        return cv::Mat();
    }

    QImage image = f.toImage();
    f.unmap();

    return qImageToMat(image);
}

QVideoFrame VideoEffectProcessor::matToQVideoFrame(const cv::Mat &mat, const QVideoFrame &originalFrame)
{
    QImage image = matToQImage(mat);
    if (image.isNull()) {
        return originalFrame;
    }

    // Create QVideoFrame from QImage
    QVideoFrameFormat format(image.size(), QVideoFrameFormat::Format_ARGB8888);
    QVideoFrame frame(format);

    if (frame.map(QVideoFrame::WriteOnly)) {
        // Copy image data to video frame
        memcpy(frame.bits(0), image.constBits(), image.sizeInBytes());
        frame.unmap();
    }

    return frame;
}

// ============================================================================
// Beauty Filter Functions
// ============================================================================

cv::Mat VideoEffectProcessor::applyBeautyFilter(const cv::Mat &input)
{
    cv::Mat result = input.clone();

    // Detect faces
    std::vector<cv::Rect> faces;
    if (!detectFaces(input, faces)) {
        // If no faces detected, apply light smoothing to entire image
        if (m_beautyLevel > 0) {
            result = applySkinSmoothing(result, m_beautyLevel / 2);
        }
        m_lastDetectedFaces.clear();
        return result;
    }

    // Cache detected faces for sticker overlay
    m_lastDetectedFaces = faces;

    // Create skin mask
    cv::Mat skinMask = createSkinMask(input, faces);

    // Apply skin smoothing
    if (m_beautyLevel > 0) {
        result = applySkinSmoothing(result, m_beautyLevel);
    }

    // Apply whitening
    if (m_whitenLevel > 0) {
        result = applyWhitening(result, m_whitenLevel);
    }

    return result;
}

cv::Mat VideoEffectProcessor::applySkinSmoothing(const cv::Mat &input, int level)
{
    if (level <= 0) return input;

    cv::Mat result;
    
    // Bilateral filter for edge-preserving smoothing
    int d = 5 + (level / 10);  // 5-15
    double sigmaColor = 20 + (level * 0.8);  // 20-100
    double sigmaSpace = 20 + (level * 0.8);  // 20-100
    
    cv::bilateralFilter(input, result, d, sigmaColor, sigmaSpace);
    
    // Blend with original based on level
    float alpha = level / 100.0f * 0.7f;  // Max 70% effect
    cv::addWeighted(result, alpha, input, 1.0f - alpha, 0, result);
    
    return result;
}

cv::Mat VideoEffectProcessor::applyWhitening(const cv::Mat &input, int level)
{
    if (level <= 0) return input;

    cv::Mat result;
    float brightness = level / 100.0f * 30.0f;  // Max +30 brightness
    
    input.convertTo(result, -1, 1.0, brightness);
    
    return result;
}

bool VideoEffectProcessor::detectFaces(const cv::Mat &input, std::vector<cv::Rect> &faces)
{
    if (m_faceCascade.empty()) {
        return false;
    }

    cv::Mat gray;
    cv::cvtColor(input, gray, cv::COLOR_BGR2GRAY);
    cv::equalizeHist(gray, gray);

    m_faceCascade.detectMultiScale(gray, faces, 1.1, 3, 0, cv::Size(30, 30));

    return !faces.empty();
}

cv::Mat VideoEffectProcessor::createSkinMask(const cv::Mat &input, const std::vector<cv::Rect> &faces)
{
    cv::Mat mask = cv::Mat::zeros(input.size(), CV_8UC1);
    
    // Simple skin color detection in YCrCb color space
    cv::Mat ycrcb;
    cv::cvtColor(input, ycrcb, cv::COLOR_BGR2YCrCb);
    
    // Skin color range in YCrCb
    cv::Scalar lower(0, 133, 77);
    cv::Scalar upper(255, 173, 127);
    
    cv::inRange(ycrcb, lower, upper, mask);
    
    // Morphological operations to clean up mask
    cv::Mat kernel = cv::getStructuringElement(cv::MORPH_ELLIPSE, cv::Size(5, 5));
    cv::morphologyEx(mask, mask, cv::MORPH_CLOSE, kernel);
    cv::morphologyEx(mask, mask, cv::MORPH_OPEN, kernel);
    
    return mask;
}

cv::Mat VideoEffectProcessor::applyFaceSlimming(const cv::Mat &input)
{
    // TODO: Implement face slimming using facial landmarks
    // This requires a facial landmark detector (e.g., dlib)
    return input;
}

// ============================================================================
// Virtual Background Functions
// ============================================================================

cv::Mat VideoEffectProcessor::applyVirtualBackground(const cv::Mat &input)
{
    switch (m_backgroundMode) {
        case BackgroundMode::Blur:
            return applyBackgroundBlur(input);
        case BackgroundMode::Replace:
            return applyBackgroundReplace(input);
        case BackgroundMode::GreenScreen:
            // Green screen is handled by color keying
            return applyBackgroundReplace(input);
        default:
            return input;
    }
}

cv::Mat VideoEffectProcessor::applyBackgroundBlur(const cv::Mat &input)
{
    // Create person mask
    cv::Mat mask = createPersonMask(input);

    if (mask.empty()) {
        return input;
    }

    // Blur the background
    cv::Mat blurred;
    int kernelSize = 31;  // Large kernel for strong blur
    cv::GaussianBlur(input, blurred, cv::Size(kernelSize, kernelSize), 0);

    // Smooth the mask edges
    cv::Mat smoothMask;
    cv::GaussianBlur(mask, smoothMask, cv::Size(15, 15), 0);
    smoothMask.convertTo(smoothMask, CV_32F, 1.0 / 255.0);

    // Convert images to float for blending
    cv::Mat inputFloat, blurredFloat;
    input.convertTo(inputFloat, CV_32F);
    blurred.convertTo(blurredFloat, CV_32F);

    // Blend: result = input * mask + blurred * (1 - mask)
    cv::Mat result;
    std::vector<cv::Mat> channels(3);
    cv::split(inputFloat, channels);

    for (int i = 0; i < 3; i++) {
        channels[i] = channels[i].mul(smoothMask) +
                      blurredFloat.mul(cv::Scalar(1.0) - smoothMask);
    }

    cv::merge(channels, result);
    result.convertTo(result, CV_8U);

    return result;
}

cv::Mat VideoEffectProcessor::applyBackgroundReplace(const cv::Mat &input)
{
    // Create person mask
    cv::Mat mask = createPersonMask(input);

    if (mask.empty()) {
        return input;
    }

    // Get or create background
    cv::Mat background;
    if (m_hasBackgroundImage && !m_backgroundImage.empty()) {
        // Resize background to match input size
        cv::resize(m_backgroundImage, background, input.size());
    } else {
        // Create solid color background (green screen effect)
        background = cv::Mat(input.size(), input.type(), cv::Scalar(0, 255, 0));
    }

    // Smooth the mask edges
    cv::Mat smoothMask;
    cv::GaussianBlur(mask, smoothMask, cv::Size(15, 15), 0);
    smoothMask.convertTo(smoothMask, CV_32F, 1.0 / 255.0);

    // Convert images to float for blending
    cv::Mat inputFloat, backgroundFloat;
    input.convertTo(inputFloat, CV_32F);
    background.convertTo(backgroundFloat, CV_32F);

    // Blend: result = input * mask + background * (1 - mask)
    cv::Mat result;
    std::vector<cv::Mat> inputChannels(3), bgChannels(3);
    cv::split(inputFloat, inputChannels);
    cv::split(backgroundFloat, bgChannels);

    std::vector<cv::Mat> resultChannels(3);
    for (int i = 0; i < 3; i++) {
        resultChannels[i] = inputChannels[i].mul(smoothMask) +
                           bgChannels[i].mul(cv::Scalar(1.0) - smoothMask);
    }

    cv::merge(resultChannels, result);
    result.convertTo(result, CV_8U);

    return result;
}

cv::Mat VideoEffectProcessor::createPersonMask(const cv::Mat &input)
{
    cv::Mat mask;

    // Method 1: Use DNN segmentation model (if available)
    if (m_segmentationModelLoaded) {
        try {
            mask = segmentPerson(input, mask);
            if (!mask.empty()) {
                m_previousMask = mask.clone();
                return mask;
            }
        } catch (const cv::Exception &e) {
            qWarning() << "DNN segmentation failed:" << e.what();
        }
    }

    // Method 2: Use background subtraction
    if (m_frameCount > 30) {  // Need some frames to build background model
        cv::Mat fgMask;
        m_backgroundSubtractor->apply(input, fgMask, 0.01);

        // Clean up the mask
        cv::Mat kernel = cv::getStructuringElement(cv::MORPH_ELLIPSE, cv::Size(5, 5));
        cv::morphologyEx(fgMask, fgMask, cv::MORPH_CLOSE, kernel);
        cv::morphologyEx(fgMask, fgMask, cv::MORPH_OPEN, kernel);

        // Dilate to include more of the person
        cv::dilate(fgMask, fgMask, kernel, cv::Point(-1, -1), 2);

        mask = fgMask;
    } else {
        // Not enough frames yet, use previous mask or full mask
        if (!m_previousMask.empty()) {
            mask = m_previousMask.clone();
        } else {
            mask = cv::Mat::ones(input.size(), CV_8UC1) * 255;
        }
    }

    m_frameCount++;
    m_previousMask = mask.clone();

    return mask;
}

cv::Mat VideoEffectProcessor::segmentPerson(const cv::Mat &input, cv::Mat &mask)
{
    // This is a placeholder for DNN-based person segmentation
    // You would need to implement this with a proper segmentation model
    // such as DeepLabV3, U-Net, or similar

    // Example using a simple approach:
    // 1. Detect person using object detection
    // 2. Create mask from bounding box
    // 3. Refine mask using GrabCut or similar

    mask = cv::Mat();
    return mask;
}

void VideoEffectProcessor::initBackgroundSegmentation()
{
    // Initialize background segmentation model
    // This would load a pre-trained model for person segmentation
}

