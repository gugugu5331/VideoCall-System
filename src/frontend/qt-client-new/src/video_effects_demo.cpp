#include <QApplication>
#include <QMainWindow>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QGridLayout>
#include <QSplitter>
#include <QGroupBox>
#include <QLabel>
#include <QPushButton>
#include <QComboBox>
#include <QSlider>
#include <QCheckBox>
#include <QProgressBar>
#include <QTimer>
#include <QCamera>
#include <QVideoWidget>
#include <QCameraViewfinder>
#include <QDebug>
#include <QMessageBox>
#include <QFileDialog>
#include <QStandardPaths>

#include "media/video_effects_processor.h"
#include "ui/video_effects_panel.h"

using namespace VideoCallSystem;

/**
 * @brief 视频特效演示主窗口
 */
class VideoEffectsDemoWindow : public QMainWindow
{
    Q_OBJECT

public:
    VideoEffectsDemoWindow(QWidget* parent = nullptr)
        : QMainWindow(parent)
        , camera_(nullptr)
        , videoWidget_(nullptr)
        , effectsProcessor_(nullptr)
        , effectsPanel_(nullptr)
        , quickEffectsBar_(nullptr)
        , isRecording_(false)
    {
        setupUI();
        setupCamera();
        setupEffectsProcessor();
        connectSignals();
        
        setWindowTitle("智能视频会议特效演示 - VideoCall System");
        setMinimumSize(1200, 800);
        resize(1400, 900);
        
        // 加载默认贴纸
        loadDefaultStickers();
        
        qDebug() << "VideoEffectsDemoWindow initialized";
    }
    
    ~VideoEffectsDemoWindow()
    {
        cleanup();
    }

private slots:
    void onStartCameraClicked()
    {
        if (camera_ && camera_->state() != QCamera::ActiveState) {
            camera_->start();
            startCameraButton_->setText("停止摄像头");
            statusLabel_->setText("摄像头已启动");
            
            if (effectsProcessor_) {
                effectsProcessor_->enableRealTimeProcessing(true);
            }
        } else if (camera_) {
            camera_->stop();
            startCameraButton_->setText("启动摄像头");
            statusLabel_->setText("摄像头已停止");
            
            if (effectsProcessor_) {
                effectsProcessor_->enableRealTimeProcessing(false);
            }
        }
    }
    
    void onRecordClicked()
    {
        if (!isRecording_) {
            QString fileName = QFileDialog::getSaveFileName(
                this, 
                "保存录制视频", 
                QStandardPaths::writableLocation(QStandardPaths::MoviesLocation) + "/video_effects_demo.mp4",
                "视频文件 (*.mp4 *.avi)"
            );
            
            if (!fileName.isEmpty() && effectsProcessor_) {
                effectsProcessor_->startRecording(fileName);
                isRecording_ = true;
                recordButton_->setText("停止录制");
                statusLabel_->setText("正在录制...");
            }
        } else {
            if (effectsProcessor_) {
                effectsProcessor_->stopRecording();
            }
            isRecording_ = false;
            recordButton_->setText("开始录制");
            statusLabel_->setText("录制已停止");
        }
    }
    
    void onScreenshotClicked()
    {
        if (effectsProcessor_) {
            QString fileName = QFileDialog::getSaveFileName(
                this,
                "保存截图",
                QStandardPaths::writableLocation(QStandardPaths::PicturesLocation) + "/video_effects_screenshot.png",
                "图片文件 (*.png *.jpg *.jpeg)"
            );
            
            if (!fileName.isEmpty()) {
                QImage screenshot = effectsProcessor_->takeScreenshot();
                if (screenshot.save(fileName)) {
                    statusLabel_->setText("截图已保存: " + fileName);
                } else {
                    statusLabel_->setText("截图保存失败");
                }
            }
        }
    }
    
    void onEffectsToggled(bool enabled)
    {
        if (effectsProcessor_) {
            effectsProcessor_->enableRealTimeProcessing(enabled);
            statusLabel_->setText(enabled ? "特效已启用" : "特效已禁用");
        }
    }
    
    void onPerformanceUpdated(const VideoEffectsProcessor::PerformanceMetrics& metrics)
    {
        fpsLabel_->setText(QString("FPS: %1").arg(metrics.averageFPS, 0, 'f', 1));
        processingTimeLabel_->setText(QString("处理时间: %1ms").arg(metrics.processingTimeMs, 0, 'f', 1));
        
        // 更新性能进度条
        int cpuUsage = static_cast<int>(metrics.processingTimeMs / 33.33 * 100); // 假设30fps
        cpuUsageBar_->setValue(qMin(cpuUsage, 100));
        
        // 性能警告
        if (metrics.averageFPS < 20) {
            statusLabel_->setText("警告: FPS过低，建议降低处理质量");
            statusLabel_->setStyleSheet("color: orange;");
        } else if (metrics.droppedFrames > 10) {
            statusLabel_->setText("警告: 丢帧过多，建议优化设置");
            statusLabel_->setStyleSheet("color: red;");
        } else {
            statusLabel_->setStyleSheet("color: green;");
        }
    }
    
    void onCameraError(QCamera::Error error)
    {
        QString errorString;
        switch (error) {
        case QCamera::NoError:
            return;
        case QCamera::CameraError:
            errorString = "摄像头错误";
            break;
        case QCamera::InvalidRequestError:
            errorString = "无效请求";
            break;
        case QCamera::ServiceMissingError:
            errorString = "服务缺失";
            break;
        case QCamera::NotSupportedFeatureError:
            errorString = "功能不支持";
            break;
        }
        
        QMessageBox::warning(this, "摄像头错误", errorString);
        statusLabel_->setText("摄像头错误: " + errorString);
    }

private:
    void setupUI()
    {
        auto* centralWidget = new QWidget(this);
        setCentralWidget(centralWidget);
        
        auto* mainLayout = new QHBoxLayout(centralWidget);
        
        // 创建分割器
        auto* splitter = new QSplitter(Qt::Horizontal, this);
        mainLayout->addWidget(splitter);
        
        // 左侧：视频预览区域
        setupVideoArea(splitter);
        
        // 右侧：控制面板
        setupControlPanel(splitter);
        
        // 设置分割器比例
        splitter->setStretchFactor(0, 2); // 视频区域占2/3
        splitter->setStretchFactor(1, 1); // 控制面板占1/3
        
        // 状态栏
        statusLabel_ = new QLabel("准备就绪", this);
        statusBar()->addWidget(statusLabel_);
        
        // 性能显示
        fpsLabel_ = new QLabel("FPS: 0", this);
        processingTimeLabel_ = new QLabel("处理时间: 0ms", this);
        statusBar()->addPermanentWidget(processingTimeLabel_);
        statusBar()->addPermanentWidget(fpsLabel_);
    }
    
    void setupVideoArea(QSplitter* parent)
    {
        auto* videoWidget = new QWidget();
        auto* videoLayout = new QVBoxLayout(videoWidget);
        
        // 视频显示
        videoWidget_ = new QVideoWidget();
        videoWidget_->setMinimumSize(640, 480);
        videoLayout->addWidget(videoWidget_);
        
        // 快速特效按钮栏
        quickEffectsBar_ = new QuickEffectsBar();
        videoLayout->addWidget(quickEffectsBar_);
        
        // 控制按钮
        auto* buttonLayout = new QHBoxLayout();
        
        startCameraButton_ = new QPushButton("启动摄像头");
        recordButton_ = new QPushButton("开始录制");
        screenshotButton_ = new QPushButton("截图");
        auto* effectsToggle = new QCheckBox("启用特效");
        effectsToggle->setChecked(true);
        
        buttonLayout->addWidget(startCameraButton_);
        buttonLayout->addWidget(recordButton_);
        buttonLayout->addWidget(screenshotButton_);
        buttonLayout->addWidget(effectsToggle);
        buttonLayout->addStretch();
        
        videoLayout->addLayout(buttonLayout);
        
        // 性能监控
        auto* perfLayout = new QHBoxLayout();
        cpuUsageBar_ = new QProgressBar();
        cpuUsageBar_->setMaximum(100);
        cpuUsageBar_->setFormat("CPU: %p%");
        perfLayout->addWidget(new QLabel("性能:"));
        perfLayout->addWidget(cpuUsageBar_);
        
        videoLayout->addLayout(perfLayout);
        
        parent->addWidget(videoWidget);
        
        // 连接信号
        connect(startCameraButton_, &QPushButton::clicked, this, &VideoEffectsDemoWindow::onStartCameraClicked);
        connect(recordButton_, &QPushButton::clicked, this, &VideoEffectsDemoWindow::onRecordClicked);
        connect(screenshotButton_, &QPushButton::clicked, this, &VideoEffectsDemoWindow::onScreenshotClicked);
        connect(effectsToggle, &QCheckBox::toggled, this, &VideoEffectsDemoWindow::onEffectsToggled);
    }
    
    void setupControlPanel(QSplitter* parent)
    {
        effectsPanel_ = new VideoEffectsPanel();
        parent->addWidget(effectsPanel_);
    }
    
    void setupCamera()
    {
        // 获取默认摄像头
        auto cameras = QCameraInfo::availableCameras();
        if (cameras.isEmpty()) {
            QMessageBox::warning(this, "警告", "未找到可用的摄像头");
            return;
        }
        
        camera_ = new QCamera(cameras.first(), this);
        camera_->setViewfinder(videoWidget_);
        
        connect(camera_, QOverload<QCamera::Error>::of(&QCamera::error),
                this, &VideoEffectsDemoWindow::onCameraError);
    }
    
    void setupEffectsProcessor()
    {
        effectsProcessor_ = new VideoEffectsProcessor(this);
        
        if (!effectsProcessor_->initialize()) {
            QMessageBox::critical(this, "错误", "视频特效处理器初始化失败");
            return;
        }
        
        // 连接到控制面板
        effectsPanel_->setVideoEffectsProcessor(effectsProcessor_);
        quickEffectsBar_->setVideoEffectsProcessor(effectsProcessor_);
        
        // 连接性能监控
        connect(effectsProcessor_, &VideoEffectsProcessor::performanceUpdated,
                this, &VideoEffectsDemoWindow::onPerformanceUpdated);
    }
    
    void connectSignals()
    {
        // 这里可以添加更多信号连接
    }
    
    void loadDefaultStickers()
    {
        // 加载一些默认贴纸
        QString assetsPath = ":/assets/stickers/";
        QStringList defaultStickers = {
            "heart.png",
            "star.png", 
            "crown.png",
            "glasses.png"
        };
        
        for (const QString& sticker : defaultStickers) {
            QString fullPath = assetsPath + sticker;
            if (QFile::exists(fullPath)) {
                effectsProcessor_->loadSticker(sticker, fullPath);
            }
        }
    }
    
    void cleanup()
    {
        if (camera_) {
            camera_->stop();
        }
        
        if (effectsProcessor_) {
            effectsProcessor_->cleanup();
        }
    }

private:
    // 摄像头组件
    QCamera* camera_;
    QVideoWidget* videoWidget_;
    
    // 特效组件
    VideoEffectsProcessor* effectsProcessor_;
    VideoEffectsPanel* effectsPanel_;
    QuickEffectsBar* quickEffectsBar_;
    
    // UI组件
    QPushButton* startCameraButton_;
    QPushButton* recordButton_;
    QPushButton* screenshotButton_;
    QLabel* statusLabel_;
    QLabel* fpsLabel_;
    QLabel* processingTimeLabel_;
    QProgressBar* cpuUsageBar_;
    
    // 状态
    bool isRecording_;
};

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    
    app.setApplicationName("VideoCall System Effects Demo");
    app.setApplicationVersion("1.0.0");
    app.setOrganizationName("VideoCall System");
    
    VideoEffectsDemoWindow window;
    window.show();
    
    return app.exec();
}

#include "video_effects_demo.moc"
