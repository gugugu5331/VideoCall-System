#include "ui/video_effects_controller.h"
#include <QDebug>
#include <QDir>
#include <QStandardPaths>

VideoEffectsController::VideoEffectsController(QObject *parent)
    : QObject(parent)
    , m_processor(std::make_unique<VideoEffectProcessor>(this))
    , m_processing(false)
{
    // Connect processor signals
    connect(m_processor.get(), &VideoEffectProcessor::processingError,
            this, &VideoEffectsController::onProcessingError);

    // Initialize presets
    initializePresets();

    qDebug() << "VideoEffectsController initialized";
}

VideoEffectsController::~VideoEffectsController()
{
}

// ============================================================================
// Beauty Properties
// ============================================================================

bool VideoEffectsController::beautyEnabled() const
{
    return m_processor->beautyEnabled();
}

int VideoEffectsController::beautyLevel() const
{
    return m_processor->beautyLevel();
}

int VideoEffectsController::whitenLevel() const
{
    return m_processor->whitenLevel();
}

void VideoEffectsController::setBeautyEnabled(bool enabled)
{
    if (m_processor->beautyEnabled() != enabled) {
        m_processor->setBeautyEnabled(enabled);
        emit beautyEnabledChanged();
        qDebug() << "Beauty enabled:" << enabled;
    }
}

void VideoEffectsController::setBeautyLevel(int level)
{
    if (m_processor->beautyLevel() != level) {
        m_processor->setBeautyLevel(level);
        emit beautyLevelChanged();
        qDebug() << "Beauty level:" << level;
    }
}

void VideoEffectsController::setWhitenLevel(int level)
{
    if (m_processor->whitenLevel() != level) {
        m_processor->setWhitenLevel(level);
        emit whitenLevelChanged();
        qDebug() << "Whiten level:" << level;
    }
}

// ============================================================================
// Virtual Background Properties
// ============================================================================

bool VideoEffectsController::virtualBackgroundEnabled() const
{
    return m_processor->virtualBackgroundEnabled();
}

int VideoEffectsController::backgroundMode() const
{
    return static_cast<int>(m_processor->backgroundMode());
}

void VideoEffectsController::setVirtualBackgroundEnabled(bool enabled)
{
    if (m_processor->virtualBackgroundEnabled() != enabled) {
        m_processor->setVirtualBackgroundEnabled(enabled);
        emit virtualBackgroundEnabledChanged();
        qDebug() << "Virtual background enabled:" << enabled;
    }
}

void VideoEffectsController::setBackgroundMode(int mode)
{
    auto bgMode = static_cast<VideoEffectProcessor::BackgroundMode>(mode);
    if (m_processor->backgroundMode() != bgMode) {
        m_processor->setBackgroundMode(bgMode);
        emit backgroundModeChanged();
        qDebug() << "Background mode:" << mode;
    }
}

// ============================================================================
// Sticker Properties
// ============================================================================

bool VideoEffectsController::stickerEnabled() const
{
    return m_processor->stickerEnabled();
}

int VideoEffectsController::stickerCount() const
{
    return m_processor->stickerOverlay()->stickerCount();
}

void VideoEffectsController::setStickerEnabled(bool enabled)
{
    if (m_processor->stickerEnabled() != enabled) {
        m_processor->setStickerEnabled(enabled);
        emit stickerEnabledChanged();
        qDebug() << "Sticker enabled:" << enabled;
    }
}

// ============================================================================
// Background Image Management
// ============================================================================

bool VideoEffectsController::loadBackgroundImage(const QString &imagePath)
{
    if (m_processor->setBackgroundImage(imagePath)) {
        m_backgroundImagePath = imagePath;
        emit backgroundImagePathChanged();
        qDebug() << "Background image loaded:" << imagePath;
        return true;
    }
    return false;
}

void VideoEffectsController::clearBackgroundImage()
{
    m_processor->clearBackgroundImage();
    m_backgroundImagePath.clear();
    emit backgroundImagePathChanged();
    qDebug() << "Background image cleared";
}

QStringList VideoEffectsController::getPresetBackgrounds() const
{
    QStringList backgrounds;
    
    // Add built-in backgrounds from resources
    backgrounds << ":/backgrounds/office.jpg"
                << ":/backgrounds/home.jpg"
                << ":/backgrounds/nature.jpg"
                << ":/backgrounds/abstract.jpg"
                << ":/backgrounds/gradient.jpg";
    
    // Add user backgrounds from documents folder
    QString userBackgroundsPath = QStandardPaths::writableLocation(QStandardPaths::DocumentsLocation) 
                                  + "/MeetingSystem/Backgrounds";
    QDir userDir(userBackgroundsPath);
    if (userDir.exists()) {
        QStringList filters;
        filters << "*.jpg" << "*.jpeg" << "*.png" << "*.bmp";
        QStringList userBackgrounds = userDir.entryList(filters, QDir::Files);
        for (const QString &bg : userBackgrounds) {
            backgrounds << userDir.absoluteFilePath(bg);
        }
    }
    
    return backgrounds;
}

// ============================================================================
// Frame Processing
// ============================================================================

QVideoFrame VideoEffectsController::processVideoFrame(const QVideoFrame &inputFrame)
{
    if (!inputFrame.isValid()) {
        return inputFrame;
    }

    // Check if any effects are enabled
    if (!m_processor->beautyEnabled() && !m_processor->virtualBackgroundEnabled()) {
        return inputFrame;
    }

    m_processing = true;
    emit processingChanged();

    QVideoFrame result = m_processor->processFrame(inputFrame);

    m_processing = false;
    emit processingChanged();

    return result;
}

// ============================================================================
// Presets
// ============================================================================

void VideoEffectsController::applyBeautyPreset(const QString &presetName)
{
    for (const auto &preset : m_beautyPresets) {
        if (preset.name == presetName) {
            setBeautyLevel(preset.beautyLevel);
            setWhitenLevel(preset.whitenLevel);
            setBeautyEnabled(true);
            qDebug() << "Applied beauty preset:" << presetName;
            return;
        }
    }
    qWarning() << "Beauty preset not found:" << presetName;
}

QStringList VideoEffectsController::getBeautyPresets() const
{
    QStringList presets;
    for (const auto &preset : m_beautyPresets) {
        presets << preset.name;
    }
    return presets;
}

void VideoEffectsController::initializePresets()
{
    m_beautyPresets.clear();
    
    // Natural preset
    m_beautyPresets.append({
        "自然",      // Natural
        30,          // beautyLevel
        20           // whitenLevel
    });
    
    // Light preset
    m_beautyPresets.append({
        "清新",      // Light
        50,          // beautyLevel
        30           // whitenLevel
    });
    
    // Strong preset
    m_beautyPresets.append({
        "魅力",      // Strong
        70,          // beautyLevel
        50           // whitenLevel
    });
    
    // Professional preset
    m_beautyPresets.append({
        "专业",      // Professional
        40,          // beautyLevel
        25           // whitenLevel
    });
    
    // Custom preset (user can modify)
    m_beautyPresets.append({
        "自定义",    // Custom
        50,          // beautyLevel
        30           // whitenLevel
    });
    
    qDebug() << "Initialized" << m_beautyPresets.size() << "beauty presets";
}

// ============================================================================
// Error Handling
// ============================================================================

void VideoEffectsController::onProcessingError(const QString &error)
{
    m_lastError = error;
    emit lastErrorChanged();
    emit processingError(error);
    qWarning() << "Video processing error:" << error;
}

// ============================================================================
// Sticker Management
// ============================================================================

QString VideoEffectsController::addSticker(const QString &imagePath, int anchorType)
{
    QString stickerId = m_processor->stickerOverlay()->addSticker(imagePath, anchorType);
    if (!stickerId.isEmpty()) {
        emit stickerCountChanged();
    }
    return stickerId;
}

bool VideoEffectsController::removeSticker(const QString &stickerId)
{
    bool success = m_processor->stickerOverlay()->removeSticker(stickerId);
    if (success) {
        emit stickerCountChanged();
    }
    return success;
}

void VideoEffectsController::clearStickers()
{
    m_processor->stickerOverlay()->clearStickers();
    emit stickerCountChanged();
}

QStringList VideoEffectsController::getPresetStickers() const
{
    return m_processor->stickerOverlay()->getPresetStickers();
}

QString VideoEffectsController::addPresetSticker(const QString &presetName, int anchorType)
{
    QString stickerId = m_processor->stickerOverlay()->addPresetSticker(presetName, anchorType);
    if (!stickerId.isEmpty()) {
        emit stickerCountChanged();
    }
    return stickerId;
}

bool VideoEffectsController::setStickerScale(const QString &stickerId, float scale)
{
    return m_processor->stickerOverlay()->setStickerScale(stickerId, scale);
}

bool VideoEffectsController::setStickerOpacity(const QString &stickerId, float opacity)
{
    return m_processor->stickerOverlay()->setStickerOpacity(stickerId, opacity);
}


