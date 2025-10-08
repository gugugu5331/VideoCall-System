#include "video_effects/sticker_overlay.h"
#include <QDebug>
#include <QFile>
#include <QFileInfo>
#include <QUuid>
#include <QMap>
#include <opencv2/imgproc.hpp>

// ============================================================================
// Sticker Implementation
// ============================================================================

Sticker::Sticker()
    : m_anchorType(AnchorType::Fixed)
    , m_position(0, 0)
    , m_size(100, 100)
    , m_scale(1.0f)
    , m_rotation(0.0f)
    , m_opacity(1.0f)
{
}

Sticker::Sticker(const QString &imagePath, AnchorType anchor)
    : Sticker()
{
    m_anchorType = anchor;
    loadImage(imagePath);
}

bool Sticker::loadImage(const QString &imagePath)
{
    m_imagePath = imagePath;
    
    try {
        // 加载图像（包含Alpha通道）
        m_image = cv::imread(imagePath.toStdString(), cv::IMREAD_UNCHANGED);
        
        if (m_image.empty()) {
            qWarning() << "Failed to load sticker image:" << imagePath;
            return false;
        }
        
        // 确保是RGBA格式
        if (m_image.channels() == 3) {
            // BGR -> BGRA
            cv::cvtColor(m_image, m_image, cv::COLOR_BGR2BGRA);
        } else if (m_image.channels() == 4) {
            // 已经是BGRA格式
        } else {
            qWarning() << "Unsupported image format:" << m_image.channels() << "channels";
            return false;
        }
        
        // 提取Alpha通道
        std::vector<cv::Mat> channels;
        cv::split(m_image, channels);
        m_alphaMask = channels[3];
        
        // 设置默认尺寸
        m_size = QSize(m_image.cols, m_image.rows);
        
        qDebug() << "Sticker loaded:" << imagePath << "size:" << m_size;
        return true;
        
    } catch (const cv::Exception &e) {
        qWarning() << "OpenCV error loading sticker:" << e.what();
        return false;
    }
}

QRect Sticker::calculateRenderRect(const cv::Rect &faceRect) const
{
    QRect rect;
    int width = m_size.width() * m_scale;
    int height = m_size.height() * m_scale;
    
    switch (m_anchorType) {
        case AnchorType::Fixed:
            // 固定位置
            rect = QRect(m_position.x(), m_position.y(), width, height);
            break;
            
        case AnchorType::Face:
            // 人脸中心
            if (!faceRect.empty()) {
                int centerX = faceRect.x + faceRect.width / 2;
                int centerY = faceRect.y + faceRect.height / 2;
                rect = QRect(centerX - width / 2 + m_position.x(),
                           centerY - height / 2 + m_position.y(),
                           width, height);
            } else {
                rect = QRect(m_position.x(), m_position.y(), width, height);
            }
            break;
            
        case AnchorType::LeftEye:
            // 左眼位置（人脸左上1/3处）
            if (!faceRect.empty()) {
                int eyeX = faceRect.x + faceRect.width * 0.3;
                int eyeY = faceRect.y + faceRect.height * 0.35;
                rect = QRect(eyeX - width / 2 + m_position.x(),
                           eyeY - height / 2 + m_position.y(),
                           width, height);
            } else {
                rect = QRect(m_position.x(), m_position.y(), width, height);
            }
            break;
            
        case AnchorType::RightEye:
            // 右眼位置（人脸右上1/3处）
            if (!faceRect.empty()) {
                int eyeX = faceRect.x + faceRect.width * 0.7;
                int eyeY = faceRect.y + faceRect.height * 0.35;
                rect = QRect(eyeX - width / 2 + m_position.x(),
                           eyeY - height / 2 + m_position.y(),
                           width, height);
            } else {
                rect = QRect(m_position.x(), m_position.y(), width, height);
            }
            break;
            
        case AnchorType::Nose:
            // 鼻子位置（人脸中心偏下）
            if (!faceRect.empty()) {
                int noseX = faceRect.x + faceRect.width / 2;
                int noseY = faceRect.y + faceRect.height * 0.55;
                rect = QRect(noseX - width / 2 + m_position.x(),
                           noseY - height / 2 + m_position.y(),
                           width, height);
            } else {
                rect = QRect(m_position.x(), m_position.y(), width, height);
            }
            break;
            
        case AnchorType::Mouth:
            // 嘴巴位置（人脸下1/3处）
            if (!faceRect.empty()) {
                int mouthX = faceRect.x + faceRect.width / 2;
                int mouthY = faceRect.y + faceRect.height * 0.75;
                rect = QRect(mouthX - width / 2 + m_position.x(),
                           mouthY - height / 2 + m_position.y(),
                           width, height);
            } else {
                rect = QRect(m_position.x(), m_position.y(), width, height);
            }
            break;
    }
    
    return rect;
}

// ============================================================================
// StickerOverlay Implementation
// ============================================================================

StickerOverlay::StickerOverlay(QObject *parent)
    : QObject(parent)
    , m_enabled(false)
    , m_faceTrackingEnabled(true)
    , m_stickerIdCounter(0)
{
    initializePresets();
}

StickerOverlay::~StickerOverlay()
{
}

void StickerOverlay::setEnabled(bool enabled)
{
    if (m_enabled != enabled) {
        m_enabled = enabled;
        emit enabledChanged();
        qDebug() << "Sticker overlay enabled:" << enabled;
    }
}

void StickerOverlay::setFaceTrackingEnabled(bool enabled)
{
    if (m_faceTrackingEnabled != enabled) {
        m_faceTrackingEnabled = enabled;
        emit faceTrackingEnabledChanged();
        qDebug() << "Face tracking enabled:" << enabled;
    }
}

QString StickerOverlay::addSticker(const QString &imagePath, int anchorType)
{
    try {
        auto sticker = std::make_unique<Sticker>(
            imagePath,
            static_cast<Sticker::AnchorType>(anchorType)
        );
        
        if (!sticker->isValid()) {
            qWarning() << "Failed to create sticker from:" << imagePath;
            return QString();
        }
        
        QString id = generateStickerId();
        sticker->setId(id);
        sticker->setName(QFileInfo(imagePath).baseName());
        
        m_stickers.push_back(std::move(sticker));
        
        emit stickerAdded(id);
        qDebug() << "Sticker added:" << id << imagePath;
        
        return id;
        
    } catch (const std::exception &e) {
        qWarning() << "Error adding sticker:" << e.what();
        emit processingError(QString("Failed to add sticker: %1").arg(e.what()));
        return QString();
    }
}

bool StickerOverlay::removeSticker(const QString &stickerId)
{
    auto it = std::remove_if(m_stickers.begin(), m_stickers.end(),
        [&stickerId](const std::unique_ptr<Sticker> &s) {
            return s->id() == stickerId;
        });
    
    if (it != m_stickers.end()) {
        m_stickers.erase(it, m_stickers.end());
        emit stickerRemoved(stickerId);
        qDebug() << "Sticker removed:" << stickerId;
        return true;
    }
    
    return false;
}

void StickerOverlay::clearStickers()
{
    m_stickers.clear();
    qDebug() << "All stickers cleared";
}

bool StickerOverlay::setStickerPosition(const QString &stickerId, const QPoint &pos)
{
    Sticker *sticker = findSticker(stickerId);
    if (sticker) {
        sticker->setPosition(pos);
        return true;
    }
    return false;
}

bool StickerOverlay::setStickerScale(const QString &stickerId, float scale)
{
    Sticker *sticker = findSticker(stickerId);
    if (sticker) {
        sticker->setScale(scale);
        return true;
    }
    return false;
}

bool StickerOverlay::setStickerRotation(const QString &stickerId, float rotation)
{
    Sticker *sticker = findSticker(stickerId);
    if (sticker) {
        sticker->setRotation(rotation);
        return true;
    }
    return false;
}

bool StickerOverlay::setStickerOpacity(const QString &stickerId, float opacity)
{
    Sticker *sticker = findSticker(stickerId);
    if (sticker) {
        sticker->setOpacity(opacity);
        return true;
    }
    return false;
}

QStringList StickerOverlay::getPresetStickers() const
{
    return m_presetStickers.keys();
}

QString StickerOverlay::addPresetSticker(const QString &presetName, int anchorType)
{
    if (m_presetStickers.contains(presetName)) {
        return addSticker(m_presetStickers[presetName], anchorType);
    }
    
    qWarning() << "Preset sticker not found:" << presetName;
    return QString();
}

cv::Mat StickerOverlay::applyStickers(const cv::Mat &input, const std::vector<cv::Rect> &faces)
{
    if (!m_enabled || m_stickers.empty()) {
        return input;
    }
    
    cv::Mat result = input.clone();
    
    try {
        for (const auto &sticker : m_stickers) {
            if (!sticker->isValid()) {
                continue;
            }
            
            // 确定使用哪个人脸（如果需要）
            cv::Rect faceRect;
            if (m_faceTrackingEnabled && !faces.empty()) {
                faceRect = faces[0];  // 使用第一个检测到的人脸
            }
            
            renderSticker(result, *sticker, faceRect);
        }
        
    } catch (const cv::Exception &e) {
        qWarning() << "OpenCV error applying stickers:" << e.what();
        emit processingError(QString("Sticker rendering error: %1").arg(e.what()));
        return input;
    }
    
    return result;
}

void StickerOverlay::renderSticker(cv::Mat &target, const Sticker &sticker, const cv::Rect &faceRect)
{
    // 计算渲染位置
    QRect renderRect = sticker.calculateRenderRect(faceRect);
    
    // 边界检查
    if (renderRect.x() >= target.cols || renderRect.y() >= target.rows ||
        renderRect.x() + renderRect.width() <= 0 || renderRect.y() + renderRect.height() <= 0) {
        return;  // 完全在画面外
    }
    
    // 调整贴图尺寸
    cv::Mat resizedSticker, resizedMask;
    cv::resize(sticker.image(), resizedSticker, cv::Size(renderRect.width(), renderRect.height()));
    cv::resize(sticker.alphaMask(), resizedMask, cv::Size(renderRect.width(), renderRect.height()));
    
    // 应用不透明度
    if (sticker.opacity() < 1.0f) {
        resizedMask.convertTo(resizedMask, CV_32F, sticker.opacity());
        resizedMask.convertTo(resizedMask, CV_8U);
    }
    
    // Alpha混合
    alphaBlend(target, resizedSticker, resizedMask, renderRect);
}

void StickerOverlay::alphaBlend(cv::Mat &target, const cv::Mat &overlay, const cv::Mat &mask, const QRect &rect)
{
    // 裁剪到目标图像范围内
    int x1 = std::max(0, rect.x());
    int y1 = std::max(0, rect.y());
    int x2 = std::min(target.cols, rect.x() + rect.width());
    int y2 = std::min(target.rows, rect.y() + rect.height());
    
    if (x2 <= x1 || y2 <= y1) {
        return;  // 无效区域
    }
    
    // 计算overlay的ROI
    int ox1 = x1 - rect.x();
    int oy1 = y1 - rect.y();
    int ox2 = ox1 + (x2 - x1);
    int oy2 = oy1 + (y2 - y1);
    
    cv::Mat targetROI = target(cv::Rect(x1, y1, x2 - x1, y2 - y1));
    cv::Mat overlayROI = overlay(cv::Rect(ox1, oy1, ox2 - ox1, oy2 - oy1));
    cv::Mat maskROI = mask(cv::Rect(ox1, oy1, ox2 - ox1, oy2 - oy1));
    
    // 转换为浮点数进行混合
    cv::Mat targetFloat, overlayFloat, maskFloat;
    targetROI.convertTo(targetFloat, CV_32F);
    overlayROI.convertTo(overlayFloat, CV_32F);
    maskROI.convertTo(maskFloat, CV_32F, 1.0 / 255.0);
    
    // Alpha混合：result = overlay * alpha + target * (1 - alpha)
    cv::Mat result;
    std::vector<cv::Mat> resultChannels;

    if (overlayFloat.channels() == 4) {
        // BGRA格式，只使用BGR通道
        std::vector<cv::Mat> overlayChannels;
        cv::split(overlayFloat, overlayChannels);

        for (int i = 0; i < 3; i++) {
            cv::Mat targetChannel, overlayChannel;
            cv::extractChannel(targetFloat, targetChannel, i);
            overlayChannel = overlayChannels[i];

            cv::Mat blended = overlayChannel.mul(maskFloat) + targetChannel.mul(cv::Scalar(1.0) - maskFloat);
            resultChannels.push_back(blended);
        }

        cv::merge(resultChannels, result);
    } else {
        // BGR格式
        for (int i = 0; i < 3; i++) {
            cv::Mat targetChannel, overlayChannel;
            cv::extractChannel(targetFloat, targetChannel, i);
            cv::extractChannel(overlayFloat, overlayChannel, i);

            cv::Mat blended = overlayChannel.mul(maskFloat) + targetChannel.mul(cv::Scalar(1.0) - maskFloat);
            resultChannels.push_back(blended);
        }

        cv::merge(resultChannels, result);
    }

    result.convertTo(targetROI, CV_8U);
}

Sticker* StickerOverlay::findSticker(const QString &stickerId)
{
    for (auto &sticker : m_stickers) {
        if (sticker->id() == stickerId) {
            return sticker.get();
        }
    }
    return nullptr;
}

QString StickerOverlay::generateStickerId()
{
    return QString("sticker_%1").arg(++m_stickerIdCounter);
}

void StickerOverlay::initializePresets()
{
    // 表情包
    m_presetStickers.insert(QString::fromUtf8("😀 笑脸"), ":/stickers/emoji_smile.png");
    m_presetStickers.insert(QString::fromUtf8("😎 墨镜"), ":/stickers/emoji_sunglasses.png");
    m_presetStickers.insert(QString::fromUtf8("😍 爱心眼"), ":/stickers/emoji_heart_eyes.png");
    m_presetStickers.insert(QString::fromUtf8("🤔 思考"), ":/stickers/emoji_thinking.png");

    // 装饰物
    m_presetStickers.insert(QString::fromUtf8("👑 皇冠"), ":/stickers/crown.png");
    m_presetStickers.insert(QString::fromUtf8("🎩 帽子"), ":/stickers/hat.png");
    m_presetStickers.insert(QString::fromUtf8("🎀 蝴蝶结"), ":/stickers/bow.png");
    m_presetStickers.insert(QString::fromUtf8("🌟 星星"), ":/stickers/star.png");

    qDebug() << "Initialized" << m_presetStickers.size() << "preset stickers";
}

