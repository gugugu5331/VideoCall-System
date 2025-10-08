#ifndef STICKER_OVERLAY_H
#define STICKER_OVERLAY_H

#include <QObject>
#include <QImage>
#include <QPoint>
#include <QSize>
#include <QRect>
#include <QMap>
#include <QString>
#include <opencv2/opencv.hpp>
#include <vector>
#include <memory>

/**
 * @brief Sticker类 - 单个贴图
 */
class Sticker
{
public:
    enum class AnchorType {
        Fixed,          // 固定位置
        Face,           // 跟随人脸
        LeftEye,        // 左眼
        RightEye,       // 右眼
        Nose,           // 鼻子
        Mouth           // 嘴巴
    };

    Sticker();
    Sticker(const QString &imagePath, AnchorType anchor = AnchorType::Fixed);
    
    // 基本属性
    QString id() const { return m_id; }
    QString name() const { return m_name; }
    QString imagePath() const { return m_imagePath; }
    cv::Mat image() const { return m_image; }
    cv::Mat alphaMask() const { return m_alphaMask; }
    
    // 位置和变换
    AnchorType anchorType() const { return m_anchorType; }
    QPoint position() const { return m_position; }
    QSize size() const { return m_size; }
    float scale() const { return m_scale; }
    float rotation() const { return m_rotation; }
    float opacity() const { return m_opacity; }
    
    // Setters
    void setId(const QString &id) { m_id = id; }
    void setName(const QString &name) { m_name = name; }
    void setAnchorType(AnchorType type) { m_anchorType = type; }
    void setPosition(const QPoint &pos) { m_position = pos; }
    void setSize(const QSize &size) { m_size = size; }
    void setScale(float scale) { m_scale = qBound(0.1f, scale, 5.0f); }
    void setRotation(float rotation) { m_rotation = rotation; }
    void setOpacity(float opacity) { m_opacity = qBound(0.0f, opacity, 1.0f); }
    
    // 加载贴图
    bool loadImage(const QString &imagePath);
    
    // 计算实际渲染位置（考虑锚点）
    QRect calculateRenderRect(const cv::Rect &faceRect = cv::Rect()) const;
    
    // 是否有效
    bool isValid() const { return !m_image.empty(); }

private:
    QString m_id;
    QString m_name;
    QString m_imagePath;
    cv::Mat m_image;        // RGBA格式
    cv::Mat m_alphaMask;    // Alpha通道
    
    AnchorType m_anchorType;
    QPoint m_position;      // 相对位置或绝对位置
    QSize m_size;           // 贴图尺寸
    float m_scale;          // 缩放比例
    float m_rotation;       // 旋转角度（度）
    float m_opacity;        // 不透明度 0.0-1.0
};

/**
 * @brief StickerOverlay类 - 贴图叠加处理器
 */
class StickerOverlay : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool enabled READ enabled WRITE setEnabled NOTIFY enabledChanged)
    Q_PROPERTY(bool faceTrackingEnabled READ faceTrackingEnabled WRITE setFaceTrackingEnabled NOTIFY faceTrackingEnabledChanged)

public:
    explicit StickerOverlay(QObject *parent = nullptr);
    ~StickerOverlay();

    // Properties
    bool enabled() const { return m_enabled; }
    bool faceTrackingEnabled() const { return m_faceTrackingEnabled; }

    // Setters
    void setEnabled(bool enabled);
    void setFaceTrackingEnabled(bool enabled);

    // 贴图管理
    Q_INVOKABLE QString addSticker(const QString &imagePath, int anchorType = 0);
    Q_INVOKABLE bool removeSticker(const QString &stickerId);
    Q_INVOKABLE void clearStickers();
    Q_INVOKABLE int stickerCount() const { return m_stickers.size(); }
    
    // 贴图属性修改
    Q_INVOKABLE bool setStickerPosition(const QString &stickerId, const QPoint &pos);
    Q_INVOKABLE bool setStickerScale(const QString &stickerId, float scale);
    Q_INVOKABLE bool setStickerRotation(const QString &stickerId, float rotation);
    Q_INVOKABLE bool setStickerOpacity(const QString &stickerId, float opacity);
    
    // 预设贴图
    Q_INVOKABLE QStringList getPresetStickers() const;
    Q_INVOKABLE QString addPresetSticker(const QString &presetName, int anchorType = 0);
    
    // 应用贴图到图像
    cv::Mat applyStickers(const cv::Mat &input, const std::vector<cv::Rect> &faces = std::vector<cv::Rect>());

signals:
    void enabledChanged();
    void faceTrackingEnabledChanged();
    void stickerAdded(const QString &stickerId);
    void stickerRemoved(const QString &stickerId);
    void processingError(const QString &error);

private:
    // 渲染单个贴图
    void renderSticker(cv::Mat &target, const Sticker &sticker, const cv::Rect &faceRect = cv::Rect());
    
    // Alpha混合
    void alphaBlend(cv::Mat &target, const cv::Mat &overlay, const cv::Mat &mask, const QRect &rect);
    
    // 查找贴图
    Sticker* findSticker(const QString &stickerId);
    
    // 生成唯一ID
    QString generateStickerId();

private:
    bool m_enabled;
    bool m_faceTrackingEnabled;
    
    std::vector<std::unique_ptr<Sticker>> m_stickers;
    int m_stickerIdCounter;
    
    // 预设贴图路径
    QMap<QString, QString> m_presetStickers;
    
    void initializePresets();
};

#endif // STICKER_OVERLAY_H

