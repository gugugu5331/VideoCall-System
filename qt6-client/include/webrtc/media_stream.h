#ifndef MEDIA_STREAM_H
#define MEDIA_STREAM_H

#include <QObject>
#include <QString>
#include <QVideoFrame>
#include <QAudioFormat>
#include <memory>

// Forward declarations
class QCamera;
class QAudioInput;
class QAudioOutput;
class QMediaCaptureSession;
class QVideoSink;
class QAudioSink;
class QAudioSource;
class QScreen;

/**
 * @brief MediaStream类 - 管理音视频媒体流
 *
 * 使用Qt Multimedia实现本地媒体捕获和渲染
 */
class MediaStream : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool hasAudio READ hasAudio NOTIFY hasAudioChanged)
    Q_PROPERTY(bool hasVideo READ hasVideo NOTIFY hasVideoChanged)
    Q_PROPERTY(bool audioEnabled READ audioEnabled WRITE setAudioEnabled NOTIFY audioEnabledChanged)
    Q_PROPERTY(bool videoEnabled READ videoEnabled WRITE setVideoEnabled NOTIFY videoEnabledChanged)
    Q_PROPERTY(bool isScreenShare READ isScreenShare NOTIFY isScreenShareChanged)
    Q_PROPERTY(QString streamId READ streamId CONSTANT)

public:
    explicit MediaStream(const QString &streamId = QString(), QObject *parent = nullptr);
    ~MediaStream();

    // Stream properties
    QString streamId() const { return m_streamId; }
    bool hasAudio() const { return m_hasAudio; }
    bool hasVideo() const { return m_hasVideo; }
    bool audioEnabled() const { return m_audioEnabled; }
    bool videoEnabled() const { return m_videoEnabled; }
    bool isScreenShare() const { return m_isScreenShare; }
    bool isLocal() const { return m_isLocal; }

    // Capture control (for local stream)
    Q_INVOKABLE bool startCapture(bool audio = true, bool video = true);
    Q_INVOKABLE void stopCapture();
    Q_INVOKABLE bool startScreenShare(int screenIndex = 0);
    Q_INVOKABLE void stopScreenShare();

    // Media control
    Q_INVOKABLE void setAudioEnabled(bool enabled);
    Q_INVOKABLE void setVideoEnabled(bool enabled);

    // Device management
    Q_INVOKABLE QStringList getAudioInputDevices() const;
    Q_INVOKABLE QStringList getVideoInputDevices() const;
    Q_INVOKABLE bool setAudioInputDevice(const QString &deviceName);
    Q_INVOKABLE bool setVideoInputDevice(const QString &deviceName);

    // Video sink access (for QML VideoOutput)
    QVideoSink* videoSink() const { return m_videoSink; }

    // Audio format
    QAudioFormat audioFormat() const { return m_audioFormat; }

signals:
    void hasAudioChanged();
    void hasVideoChanged();
    void audioEnabledChanged();
    void videoEnabledChanged();
    void isScreenShareChanged();

    void videoFrameReady(const QVideoFrame &frame);
    void audioDataReady(const QByteArray &data);

    void captureStarted();
    void captureStopped();
    void error(const QString &errorString);

private slots:
    void onVideoFrameChanged(const QVideoFrame &frame);
    void onCameraError();
    void onScreenCaptureTimeout();

private:
    void setupAudioFormat();
    void startScreenCapture();
    void stopScreenCapture();

private:
    QString m_streamId;
    bool m_isLocal;
    bool m_hasAudio;
    bool m_hasVideo;
    bool m_audioEnabled;
    bool m_videoEnabled;
    bool m_isScreenShare;

    // Qt Multimedia components
    QCamera *m_camera;
    QAudioInput *m_audioInput;
    QMediaCaptureSession *m_captureSession;
    QVideoSink *m_videoSink;
    QAudioSource *m_audioSource;
    QAudioFormat m_audioFormat;

    // Screen capture
    QScreen *m_screen;
    QTimer *m_screenCaptureTimer;
    int m_screenIndex;
};

#endif // MEDIA_STREAM_H

