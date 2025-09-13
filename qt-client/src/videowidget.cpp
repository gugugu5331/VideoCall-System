#include "videowidget.h"
#include <QtWidgets/QApplication>
#include <QtWidgets/QMenu>
#include <QtGui/QContextMenuEvent>
#include <QtCore/QDebug>

VideoWidget::VideoWidget(const QString &title, QWidget *parent)
    : QWidget(parent)
    , m_mainLayout(nullptr)
    , m_videoWidget(nullptr)
    , m_overlayWidget(nullptr)
    , m_overlayLayout(nullptr)
    , m_titleLabel(nullptr)
    , m_controlsWidget(nullptr)
    , m_controlsLayout(nullptr)
    , m_muteButton(nullptr)
    , m_fullscreenButton(nullptr)
    , m_audioLevelWidget(nullptr)
    , m_audioLevelTimer(nullptr)
    , m_currentAudioLevel(0.0)
    , m_title(title)
    , m_isMuted(false)
    , m_showControls(true)
    , m_hasVideo(false)
    , m_camera(nullptr)
    , m_mediaPlayer(nullptr)
    , m_captureSession(nullptr)
{
    setupUI();
    setTitle(title);
    
    // 设置样式
    m_styleSheet = R"(
        VideoWidget {
            background-color: #2d2d2d;
            border: 2px solid #444;
            border-radius: 8px;
        }
        VideoWidget:hover {
            border-color: #4facfe;
        }
        QLabel {
            color: white;
            font-weight: bold;
            background-color: rgba(0, 0, 0, 0.7);
            padding: 4px 8px;
            border-radius: 4px;
        }
        QPushButton {
            background-color: rgba(0, 0, 0, 0.7);
            color: white;
            border: none;
            border-radius: 4px;
            padding: 4px 8px;
            min-width: 24px;
            min-height: 24px;
        }
        QPushButton:hover {
            background-color: rgba(0, 0, 0, 0.9);
        }
        QPushButton:pressed {
            background-color: rgba(255, 255, 255, 0.2);
        }
    )";
    setStyleSheet(m_styleSheet);
    
    // 音频电平更新定时器
    m_audioLevelTimer = new QTimer(this);
    connect(m_audioLevelTimer, &QTimer::timeout, this, &VideoWidget::updateAudioLevel);
    m_audioLevelTimer->start(100); // 10fps更新
}

VideoWidget::~VideoWidget()
{
    if (m_captureSession) {
        m_captureSession->deleteLater();
    }
}

void VideoWidget::setupUI()
{
    // 主布局
    m_mainLayout = new QVBoxLayout(this);
    m_mainLayout->setContentsMargins(0, 0, 0, 0);
    m_mainLayout->setSpacing(0);
    
    // 视频窗口
    m_videoWidget = new QVideoWidget;
    m_videoWidget->setAspectRatioMode(Qt::KeepAspectRatioByExpanding);
    m_videoWidget->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
    
    // 覆盖层
    m_overlayWidget = new QWidget;
    m_overlayWidget->setAttribute(Qt::WA_TransparentForMouseEvents, false);
    m_overlayWidget->setStyleSheet("background-color: transparent;");
    
    m_overlayLayout = new QHBoxLayout(m_overlayWidget);
    m_overlayLayout->setContentsMargins(8, 8, 8, 8);
    
    // 标题标签
    m_titleLabel = new QLabel;
    m_titleLabel->setAlignment(Qt::AlignLeft | Qt::AlignTop);
    m_overlayLayout->addWidget(m_titleLabel, 0, Qt::AlignLeft | Qt::AlignTop);
    
    // 弹性空间
    m_overlayLayout->addStretch();
    
    // 控制按钮
    m_controlsWidget = new QWidget;
    m_controlsLayout = new QHBoxLayout(m_controlsWidget);
    m_controlsLayout->setContentsMargins(0, 0, 0, 0);
    m_controlsLayout->setSpacing(4);
    
    m_muteButton = new QPushButton("🔊");
    m_muteButton->setToolTip("静音/取消静音");
    connect(m_muteButton, &QPushButton::clicked, this, &VideoWidget::toggleMute);
    m_controlsLayout->addWidget(m_muteButton);
    
    m_fullscreenButton = new QPushButton("⛶");
    m_fullscreenButton->setToolTip("全屏");
    connect(m_fullscreenButton, &QPushButton::clicked, this, &VideoWidget::fullscreenRequested);
    m_controlsLayout->addWidget(m_fullscreenButton);
    
    m_overlayLayout->addWidget(m_controlsWidget, 0, Qt::AlignRight | Qt::AlignTop);
    
    // 音频电平指示器
    m_audioLevelWidget = new QWidget;
    m_audioLevelWidget->setFixedSize(4, 60);
    m_audioLevelWidget->setStyleSheet("background-color: transparent;");
    m_overlayLayout->addWidget(m_audioLevelWidget, 0, Qt::AlignRight | Qt::AlignBottom);
    
    // 添加到主布局
    m_mainLayout->addWidget(m_videoWidget);
    
    // 设置覆盖层位置
    m_overlayWidget->setParent(this);
    m_overlayWidget->raise();
}

void VideoWidget::setVideoSource(QObject *source)
{
    if (!source) {
        m_hasVideo = false;
        update();
        return;
    }
    
    // 尝试设置不同类型的视频源
    if (auto *camera = qobject_cast<QCamera*>(source)) {
        setCamera(camera);
    } else if (auto *player = qobject_cast<QMediaPlayer*>(source)) {
        setMediaPlayer(player);
    }
    
    m_hasVideo = true;
    update();
}

void VideoWidget::setCamera(QCamera *camera)
{
    if (m_camera == camera) return;
    
    m_camera = camera;
    
    if (!m_captureSession) {
        m_captureSession = new QMediaCaptureSession(this);
    }
    
    m_captureSession->setCamera(camera);
    m_captureSession->setVideoOutput(m_videoWidget);
    
    if (camera) {
        camera->start();
        m_hasVideo = true;
    } else {
        m_hasVideo = false;
    }
    
    update();
}

void VideoWidget::setMediaPlayer(QMediaPlayer *player)
{
    if (m_mediaPlayer == player) return;
    
    m_mediaPlayer = player;
    
    if (player) {
        player->setVideoOutput(m_videoWidget);
        m_hasVideo = true;
    } else {
        m_hasVideo = false;
    }
    
    update();
}

void VideoWidget::setMuted(bool muted)
{
    if (m_isMuted == muted) return;
    
    m_isMuted = muted;
    m_muteButton->setText(muted ? "🔇" : "🔊");
    m_muteButton->setToolTip(muted ? "取消静音" : "静音");
    
    update();
}

void VideoWidget::setTitle(const QString &title)
{
    m_title = title;
    m_titleLabel->setText(title);
    setToolTip(title);
}

void VideoWidget::showControls(bool show)
{
    m_showControls = show;
    updateControlsVisibility();
}

void VideoWidget::setAudioLevel(double level)
{
    m_currentAudioLevel = qBound(0.0, level, 1.0);
    m_audioLevelWidget->update();
}

void VideoWidget::paintEvent(QPaintEvent *event)
{
    QWidget::paintEvent(event);
    
    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing);
    
    // 如果没有视频，绘制占位符
    if (!m_hasVideo) {
        drawNoVideoPlaceholder(painter);
    }
    
    // 绘制音频电平指示器
    if (m_currentAudioLevel > 0.01) {
        drawAudioLevelIndicator(painter);
    }
}

void VideoWidget::mouseDoubleClickEvent(QMouseEvent *event)
{
    Q_UNUSED(event)
    emit fullscreenRequested();
}

void VideoWidget::contextMenuEvent(QContextMenuEvent *event)
{
    QMenu contextMenu(this);
    
    QAction *muteAction = contextMenu.addAction(m_isMuted ? "取消静音" : "静音");
    connect(muteAction, &QAction::triggered, this, &VideoWidget::toggleMute);
    
    contextMenu.addSeparator();
    
    QAction *fullscreenAction = contextMenu.addAction("全屏");
    connect(fullscreenAction, &QAction::triggered, this, &VideoWidget::fullscreenRequested);
    
    contextMenu.exec(event->globalPos());
}

void VideoWidget::resizeEvent(QResizeEvent *event)
{
    QWidget::resizeEvent(event);
    
    // 调整覆盖层大小
    if (m_overlayWidget) {
        m_overlayWidget->resize(size());
    }
}

void VideoWidget::toggleMute()
{
    setMuted(!m_isMuted);
    emit muteToggled(m_isMuted);
}

void VideoWidget::updateAudioLevel()
{
    // 模拟音频电平变化
    if (m_hasVideo && !m_isMuted) {
        // 这里应该从实际音频源获取电平
        // 现在使用随机值模拟
        static double lastLevel = 0.0;
        double targetLevel = (qrand() % 100) / 100.0 * 0.8;
        lastLevel = lastLevel * 0.7 + targetLevel * 0.3; // 平滑过渡
        setAudioLevel(lastLevel);
    } else {
        setAudioLevel(0.0);
    }
}

void VideoWidget::onVideoFrameChanged()
{
    // 视频帧变化处理
    update();
}

void VideoWidget::updateControlsVisibility()
{
    m_controlsWidget->setVisible(m_showControls);
}

void VideoWidget::drawAudioLevelIndicator(QPainter &painter)
{
    if (!m_audioLevelWidget) return;
    
    QRect levelRect = m_audioLevelWidget->geometry();
    levelRect.moveTopRight(QPoint(width() - 10, 10));
    
    // 绘制背景
    painter.fillRect(levelRect, QColor(0, 0, 0, 100));
    
    // 绘制电平条
    int levelHeight = static_cast<int>(levelRect.height() * m_currentAudioLevel);
    QRect activeRect = levelRect;
    activeRect.setTop(levelRect.bottom() - levelHeight);
    
    // 根据电平设置颜色
    QColor levelColor;
    if (m_currentAudioLevel < 0.3) {
        levelColor = QColor(0, 255, 0); // 绿色
    } else if (m_currentAudioLevel < 0.7) {
        levelColor = QColor(255, 255, 0); // 黄色
    } else {
        levelColor = QColor(255, 0, 0); // 红色
    }
    
    painter.fillRect(activeRect, levelColor);
}

void VideoWidget::drawNoVideoPlaceholder(QPainter &painter)
{
    QRect rect = this->rect();
    
    // 绘制背景
    painter.fillRect(rect, QColor(45, 45, 45));
    
    // 绘制占位符图标
    painter.setPen(QColor(150, 150, 150));
    painter.setFont(QFont("Arial", 48));
    
    QString placeholderText = "📹";
    QFontMetrics fm(painter.font());
    QRect textRect = fm.boundingRect(placeholderText);
    
    painter.drawText(rect.center() - textRect.center(), placeholderText);
    
    // 绘制提示文本
    painter.setFont(QFont("Arial", 12));
    painter.drawText(rect.adjusted(10, 10, -10, -10), 
                    Qt::AlignCenter | Qt::TextWordWrap, 
                    m_title.isEmpty() ? "无视频" : m_title);
}
