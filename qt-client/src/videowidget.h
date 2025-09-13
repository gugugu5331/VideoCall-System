#ifndef VIDEOWIDGET_H
#define VIDEOWIDGET_H

#include <QtWidgets/QWidget>
#include <QtWidgets/QVBoxLayout>
#include <QtWidgets/QHBoxLayout>
#include <QtWidgets/QLabel>
#include <QtWidgets/QPushButton>
#include <QtMultimediaWidgets/QVideoWidget>
#include <QtMultimedia/QMediaPlayer>
#include <QtMultimedia/QCamera>
#include <QtMultimedia/QMediaCaptureSession>
#include <QtCore/QTimer>
#include <QtGui/QPainter>
#include <QtGui/QPixmap>

class VideoWidget : public QWidget
{
    Q_OBJECT

public:
    explicit VideoWidget(const QString &title = QString(), QWidget *parent = nullptr);
    ~VideoWidget();

    // 设置视频源
    void setVideoSource(QObject *source);
    void setCamera(QCamera *camera);
    void setMediaPlayer(QMediaPlayer *player);
    
    // 控制功能
    void setMuted(bool muted);
    void setTitle(const QString &title);
    void showControls(bool show);
    void setAudioLevel(double level);
    
    // 状态
    bool isMuted() const { return m_isMuted; }
    QString title() const { return m_title; }

signals:
    void muteToggled(bool muted);
    void fullscreenRequested();

protected:
    void paintEvent(QPaintEvent *event) override;
    void mouseDoubleClickEvent(QMouseEvent *event) override;
    void contextMenuEvent(QContextMenuEvent *event) override;
    void resizeEvent(QResizeEvent *event) override;

private slots:
    void toggleMute();
    void updateAudioLevel();
    void onVideoFrameChanged();

private:
    void setupUI();
    void updateControlsVisibility();
    void drawAudioLevelIndicator(QPainter &painter);
    void drawNoVideoPlaceholder(QPainter &painter);
    
    // UI组件
    QVBoxLayout *m_mainLayout;
    QVideoWidget *m_videoWidget;
    QWidget *m_overlayWidget;
    QHBoxLayout *m_overlayLayout;
    QLabel *m_titleLabel;
    QWidget *m_controlsWidget;
    QHBoxLayout *m_controlsLayout;
    QPushButton *m_muteButton;
    QPushButton *m_fullscreenButton;
    
    // 音频电平指示器
    QWidget *m_audioLevelWidget;
    QTimer *m_audioLevelTimer;
    double m_currentAudioLevel;
    
    // 状态
    QString m_title;
    bool m_isMuted;
    bool m_showControls;
    bool m_hasVideo;
    
    // 视频源
    QCamera *m_camera;
    QMediaPlayer *m_mediaPlayer;
    QMediaCaptureSession *m_captureSession;
    
    // 样式
    QString m_styleSheet;
};

#endif // VIDEOWIDGET_H
