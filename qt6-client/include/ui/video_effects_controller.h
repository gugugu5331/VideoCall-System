#ifndef VIDEO_EFFECTS_CONTROLLER_H
#define VIDEO_EFFECTS_CONTROLLER_H

#include <QObject>
#include <QVideoFrame>
#include <memory>
#include "video_effects/video_effect_processor.h"

/**
 * @brief VideoEffectsController类 - 视频效果控制器
 * 
 * 管理美颜和虚拟背景效果的UI控制器
 */
class VideoEffectsController : public QObject
{
    Q_OBJECT
    
    // Beauty properties
    Q_PROPERTY(bool beautyEnabled READ beautyEnabled WRITE setBeautyEnabled NOTIFY beautyEnabledChanged)
    Q_PROPERTY(int beautyLevel READ beautyLevel WRITE setBeautyLevel NOTIFY beautyLevelChanged)
    Q_PROPERTY(int whitenLevel READ whitenLevel WRITE setWhitenLevel NOTIFY whitenLevelChanged)
    
    // Virtual background properties
    Q_PROPERTY(bool virtualBackgroundEnabled READ virtualBackgroundEnabled WRITE setVirtualBackgroundEnabled NOTIFY virtualBackgroundEnabledChanged)
    Q_PROPERTY(int backgroundMode READ backgroundMode WRITE setBackgroundMode NOTIFY backgroundModeChanged)
    Q_PROPERTY(QString backgroundImagePath READ backgroundImagePath NOTIFY backgroundImagePathChanged)

    // Sticker properties
    Q_PROPERTY(bool stickerEnabled READ stickerEnabled WRITE setStickerEnabled NOTIFY stickerEnabledChanged)
    Q_PROPERTY(int stickerCount READ stickerCount NOTIFY stickerCountChanged)
    
    // Status
    Q_PROPERTY(bool processing READ processing NOTIFY processingChanged)
    Q_PROPERTY(QString lastError READ lastError NOTIFY lastErrorChanged)

public:
    explicit VideoEffectsController(QObject *parent = nullptr);
    ~VideoEffectsController();

    // Beauty properties
    bool beautyEnabled() const;
    int beautyLevel() const;
    int whitenLevel() const;

    // Virtual background properties
    bool virtualBackgroundEnabled() const;
    int backgroundMode() const;
    QString backgroundImagePath() const { return m_backgroundImagePath; }

    // Sticker properties
    bool stickerEnabled() const;
    int stickerCount() const;

    // Status
    bool processing() const { return m_processing; }
    QString lastError() const { return m_lastError; }

    // Setters
    void setBeautyEnabled(bool enabled);
    void setBeautyLevel(int level);
    void setWhitenLevel(int level);
    void setVirtualBackgroundEnabled(bool enabled);
    void setBackgroundMode(int mode);
    void setStickerEnabled(bool enabled);

    // Background image management
    Q_INVOKABLE bool loadBackgroundImage(const QString &imagePath);
    Q_INVOKABLE void clearBackgroundImage();
    Q_INVOKABLE QStringList getPresetBackgrounds() const;

    // Frame processing
    Q_INVOKABLE QVideoFrame processVideoFrame(const QVideoFrame &inputFrame);

    // Presets
    Q_INVOKABLE void applyBeautyPreset(const QString &presetName);
    Q_INVOKABLE QStringList getBeautyPresets() const;

    // Sticker management
    Q_INVOKABLE QString addSticker(const QString &imagePath, int anchorType = 0);
    Q_INVOKABLE bool removeSticker(const QString &stickerId);
    Q_INVOKABLE void clearStickers();
    Q_INVOKABLE QStringList getPresetStickers() const;
    Q_INVOKABLE QString addPresetSticker(const QString &presetName, int anchorType = 0);
    Q_INVOKABLE bool setStickerScale(const QString &stickerId, float scale);
    Q_INVOKABLE bool setStickerOpacity(const QString &stickerId, float opacity);

signals:
    // Beauty signals
    void beautyEnabledChanged();
    void beautyLevelChanged();
    void whitenLevelChanged();

    // Virtual background signals
    void virtualBackgroundEnabledChanged();
    void backgroundModeChanged();
    void backgroundImagePathChanged();

    // Sticker signals
    void stickerEnabledChanged();
    void stickerCountChanged();

    // Status signals
    void processingChanged();
    void lastErrorChanged();
    void processingError(const QString &error);

private slots:
    void onProcessingError(const QString &error);

private:
    std::unique_ptr<VideoEffectProcessor> m_processor;
    
    QString m_backgroundImagePath;
    bool m_processing;
    QString m_lastError;

    // Preset configurations
    struct BeautyPreset {
        QString name;
        int beautyLevel;
        int whitenLevel;
    };
    QList<BeautyPreset> m_beautyPresets;

    void initializePresets();
};

#endif // VIDEO_EFFECTS_CONTROLLER_H

