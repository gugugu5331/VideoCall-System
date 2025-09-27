#pragma once

#include "core/common.h"
#include "media/video_effects_processor.h"

QT_BEGIN_NAMESPACE
class QSlider;
class QComboBox;
class QPushButton;
class QLabel;
class QGridLayout;
class QVBoxLayout;
class QHBoxLayout;
class QGroupBox;
class QCheckBox;
class QSpinBox;
class QProgressBar;
class QListWidget;
class QTabWidget;
QT_END_NAMESPACE

namespace VideoCallSystem {

/**
 * @brief 视频特效控制面板
 * 
 * 提供用户友好的界面来控制滤镜、贴图、背景等视频特效
 */
class VideoEffectsPanel : public QWidget
{
    Q_OBJECT

public:
    explicit VideoEffectsPanel(QWidget* parent = nullptr);
    ~VideoEffectsPanel();

    // 设置视频特效处理器
    void setVideoEffectsProcessor(VideoEffectsProcessor* processor);
    VideoEffectsProcessor* videoEffectsProcessor() const { return effectsProcessor_; }

    // 面板控制
    void showPanel(bool show = true);
    void hidePanel() { showPanel(false); }
    bool isPanelVisible() const;
    
    // 预设管理
    void loadPresets();
    void saveCurrentAsPreset(const QString& name);
    void deletePreset(const QString& name);

public slots:
    // 滤镜控制槽
    void onFilterChanged(int filterIndex);
    void onFilterIntensityChanged(int intensity);
    void onFilterPresetClicked();
    
    // 贴图控制槽
    void onStickerChanged(int stickerIndex);
    void onLoadStickerClicked();
    void onRemoveStickerClicked();
    void onStickerPositionChanged();
    
    // 背景控制槽
    void onBackgroundToggled(bool enabled);
    void onBackgroundImageClicked();
    void onBackgroundBlurChanged(int intensity);
    void onBackgroundRemoveClicked();
    
    // 面部检测槽
    void onFaceDetectionToggled(bool enabled);
    void onFaceDetectionSensitivityChanged(int sensitivity);
    
    // 性能控制槽
    void onPerformanceSettingsChanged();
    void onGPUAccelerationToggled(bool enabled);
    void onProcessingResolutionChanged();
    
    // 预设控制槽
    void onPresetSelected(const QString& presetName);
    void onSavePresetClicked();
    void onDeletePresetClicked();
    void onResetToDefaultClicked();

signals:
    // 用户操作信号
    void filterChangeRequested(VideoProcessing::FilterType filterType);
    void filterIntensityChangeRequested(float intensity);
    void stickerChangeRequested(const QString& stickerName);
    void backgroundChangeRequested(bool enabled);
    void faceDetectionChangeRequested(bool enabled);
    
    // 面板状态信号
    void panelShown();
    void panelHidden();
    void settingsChanged();

private slots:
    // 内部更新槽
    void updatePerformanceDisplay();
    void updateFilterPreview();
    void updateStickerPreview();
    void updateBackgroundPreview();
    
    // 特效处理器信号响应
    void onEffectsProcessorFilterChanged(VideoProcessing::FilterType filterType);
    void onEffectsProcessorStickerChanged(const QString& stickerName);
    void onEffectsProcessorBackgroundChanged(bool enabled);
    void onEffectsProcessorPerformanceUpdated(const VideoEffectsProcessor::PerformanceMetrics& metrics);

private:
    // UI初始化
    void setupUI();
    void setupFilterControls();
    void setupStickerControls();
    void setupBackgroundControls();
    void setupFaceDetectionControls();
    void setupPerformanceControls();
    void setupPresetControls();
    
    // 样式设置
    void setupStyles();
    void applyModernStyle();
    
    // 数据更新
    void updateFilterList();
    void updateStickerList();
    void updatePresetList();
    void updatePerformanceMetrics();
    
    // 预览功能
    void generateFilterPreview(VideoProcessing::FilterType filterType);
    void generateStickerPreview(const QString& stickerName);
    void generateBackgroundPreview();
    
    // 文件操作
    QString selectImageFile(const QString& title = "Select Image");
    QString selectVideoFile(const QString& title = "Select Video");
    bool validateImageFile(const QString& filePath);
    
    // 工具函数
    QString filterTypeToString(VideoProcessing::FilterType filterType);
    VideoProcessing::FilterType stringToFilterType(const QString& filterString);
    void showErrorMessage(const QString& message);
    void showSuccessMessage(const QString& message);

private:
    // 核心组件
    VideoEffectsProcessor* effectsProcessor_;
    
    // 主布局
    QTabWidget* tabWidget_;
    QVBoxLayout* mainLayout_;
    
    // 滤镜控制组件
    QGroupBox* filterGroup_;
    QComboBox* filterComboBox_;
    QSlider* filterIntensitySlider_;
    QLabel* filterIntensityLabel_;
    QPushButton* filterPresetButton_;
    QLabel* filterPreviewLabel_;
    
    // 贴图控制组件
    QGroupBox* stickerGroup_;
    QComboBox* stickerComboBox_;
    QPushButton* loadStickerButton_;
    QPushButton* removeStickerButton_;
    QLabel* stickerPreviewLabel_;
    QCheckBox* stickerFaceTrackingCheckBox_;
    
    // 背景控制组件
    QGroupBox* backgroundGroup_;
    QCheckBox* backgroundEnabledCheckBox_;
    QPushButton* backgroundImageButton_;
    QSlider* backgroundBlurSlider_;
    QLabel* backgroundBlurLabel_;
    QPushButton* backgroundRemoveButton_;
    QLabel* backgroundPreviewLabel_;
    
    // 面部检测控制组件
    QGroupBox* faceDetectionGroup_;
    QCheckBox* faceDetectionEnabledCheckBox_;
    QSlider* faceDetectionSensitivitySlider_;
    QLabel* faceDetectionSensitivityLabel_;
    QLabel* faceCountLabel_;
    
    // 性能控制组件
    QGroupBox* performanceGroup_;
    QCheckBox* gpuAccelerationCheckBox_;
    QComboBox* processingResolutionComboBox_;
    QSpinBox* targetFPSSpinBox_;
    QProgressBar* cpuUsageBar_;
    QProgressBar* gpuUsageBar_;
    QLabel* fpsLabel_;
    QLabel* processingTimeLabel_;
    
    // 预设控制组件
    QGroupBox* presetGroup_;
    QListWidget* presetListWidget_;
    QPushButton* savePresetButton_;
    QPushButton* deletePresetButton_;
    QPushButton* resetDefaultButton_;
    
    // 状态变量
    bool panelVisible_;
    bool updatingUI_;
    
    // 预览图像
    QPixmap filterPreviewPixmap_;
    QPixmap stickerPreviewPixmap_;
    QPixmap backgroundPreviewPixmap_;
    
    // 定时器
    QTimer* performanceUpdateTimer_;
    QTimer* previewUpdateTimer_;
    
    // 样式表
    QString modernStyleSheet_;
};

/**
 * @brief 快速特效按钮栏
 * 
 * 提供常用特效的快速访问按钮
 */
class QuickEffectsBar : public QWidget
{
    Q_OBJECT

public:
    explicit QuickEffectsBar(QWidget* parent = nullptr);
    
    void setVideoEffectsProcessor(VideoEffectsProcessor* processor);

signals:
    void quickEffectRequested(const QString& effectName);

private slots:
    void onBeautyClicked();
    void onCartoonClicked();
    void onVintageClicked();
    void onSketchClicked();
    void onClearAllClicked();

private:
    void setupUI();
    void setupStyles();

private:
    VideoEffectsProcessor* effectsProcessor_;
    
    QPushButton* beautyButton_;
    QPushButton* cartoonButton_;
    QPushButton* vintageButton_;
    QPushButton* sketchButton_;
    QPushButton* clearAllButton_;
    
    QHBoxLayout* layout_;
};

} // namespace VideoCallSystem
