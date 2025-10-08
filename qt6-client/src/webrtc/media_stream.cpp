#include "webrtc/media_stream.h"
#include "utils/logger.h"

#include <QCamera>
#include <QAudioInput>
#include <QAudioOutput>
#include <QMediaCaptureSession>
#include <QVideoSink>
#include <QAudioSource>
#include <QAudioSink>
#include <QMediaDevices>
#include <QScreen>
#include <QGuiApplication>
#include <QTimer>
#include <QPixmap>
#include <QUuid>

MediaStream::MediaStream(const QString &streamId, QObject *parent)
    : QObject(parent)
    , m_streamId(streamId.isEmpty() ? QUuid::createUuid().toString(QUuid::WithoutBraces) : streamId)
    , m_isLocal(true)
    , m_hasAudio(false)
    , m_hasVideo(false)
    , m_audioEnabled(true)
    , m_videoEnabled(true)
    , m_isScreenShare(false)
    , m_camera(nullptr)
    , m_audioInput(nullptr)
    , m_captureSession(nullptr)
    , m_videoSink(nullptr)
    , m_audioSource(nullptr)
    , m_screen(nullptr)
    , m_screenCaptureTimer(nullptr)
    , m_screenIndex(0)
{
    LOG_INFO(QString("MediaStream created: %1").arg(m_streamId));
    setupAudioFormat();
}

MediaStream::~MediaStream()
{
    stopCapture();
    stopScreenShare();
    LOG_INFO(QString("MediaStream destroyed: %1").arg(m_streamId));
}

void MediaStream::setupAudioFormat()
{
    // 设置音频格式：16位，48kHz，立体声
    m_audioFormat.setSampleRate(48000);
    m_audioFormat.setChannelCount(2);
    m_audioFormat.setSampleFormat(QAudioFormat::Int16);
}

bool MediaStream::startCapture(bool audio, bool video)
{
    LOG_INFO(QString("Starting capture - audio: %1, video: %2").arg(audio).arg(video));

    if (m_captureSession) {
        LOG_WARNING("Capture session already exists");
        return false;
    }

    try {
        // 创建捕获会话
        m_captureSession = new QMediaCaptureSession(this);

        // 启动视频捕获
        if (video) {
            QCameraDevice cameraDevice = QMediaDevices::defaultVideoInput();
            if (cameraDevice.isNull()) {
                LOG_ERROR("No video input device found");
                delete m_captureSession;
                m_captureSession = nullptr;
                return false;
            }

            m_camera = new QCamera(cameraDevice, this);
            m_captureSession->setCamera(m_camera);

            // 创建视频接收器
            m_videoSink = new QVideoSink(this);
            m_captureSession->setVideoOutput(m_videoSink);

            // 连接视频帧信号
            connect(m_videoSink, &QVideoSink::videoFrameChanged,
                    this, &MediaStream::onVideoFrameChanged);

            // 连接错误信号
            connect(m_camera, &QCamera::errorOccurred,
                    this, &MediaStream::onCameraError);

            // 启动摄像头
            m_camera->start();

            m_hasVideo = true;
            emit hasVideoChanged();
            LOG_INFO("Video capture started");
        }

        // 启动音频捕获
        if (audio) {
            QAudioDevice audioDevice = QMediaDevices::defaultAudioInput();
            if (audioDevice.isNull()) {
                LOG_ERROR("No audio input device found");
                if (m_camera) {
                    m_camera->stop();
                }
                delete m_captureSession;
                m_captureSession = nullptr;
                return false;
            }

            m_audioInput = new QAudioInput(audioDevice, this);
            m_captureSession->setAudioInput(m_audioInput);

            // 创建音频源用于读取数据
            m_audioSource = new QAudioSource(audioDevice, m_audioFormat, this);

            // 启动音频源并连接数据读取
            QIODevice *audioIO = m_audioSource->start();
            if (audioIO) {
                connect(audioIO, &QIODevice::readyRead, this, [this, audioIO]() {
                    QByteArray audioData = audioIO->readAll();
                    if (!audioData.isEmpty()) {
                        emit audioDataReady(audioData);
                    }
                });
            }

            m_hasAudio = true;
            emit hasAudioChanged();
            LOG_INFO("Audio capture started");
        }

        emit captureStarted();
        return true;

    } catch (const std::exception &e) {
        LOG_ERROR(QString("Failed to start capture: %1").arg(e.what()));
        emit error(QString("Failed to start capture: %1").arg(e.what()));
        return false;
    }
}

void MediaStream::stopCapture()
{
    LOG_INFO("Stopping capture");

    if (m_camera) {
        m_camera->stop();
        m_camera->deleteLater();
        m_camera = nullptr;
    }

    if (m_audioInput) {
        m_audioInput->deleteLater();
        m_audioInput = nullptr;
    }

    if (m_audioSource) {
        m_audioSource->stop();
        m_audioSource->deleteLater();
        m_audioSource = nullptr;
    }

    if (m_videoSink) {
        m_videoSink->deleteLater();
        m_videoSink = nullptr;
    }

    if (m_captureSession) {
        m_captureSession->deleteLater();
        m_captureSession = nullptr;
    }

    if (m_hasAudio) {
        m_hasAudio = false;
        emit hasAudioChanged();
    }

    if (m_hasVideo) {
        m_hasVideo = false;
        emit hasVideoChanged();
    }

    emit captureStopped();
}

bool MediaStream::startScreenShare(int screenIndex)
{
    LOG_INFO(QString("Starting screen share - screen: %1").arg(screenIndex));

    if (m_isScreenShare) {
        LOG_WARNING("Screen share already active");
        return false;
    }

    QList<QScreen*> screens = QGuiApplication::screens();
    if (screenIndex < 0 || screenIndex >= screens.size()) {
        LOG_ERROR(QString("Invalid screen index: %1").arg(screenIndex));
        return false;
    }

    m_screen = screens[screenIndex];
    m_screenIndex = screenIndex;
    m_isScreenShare = true;

    // 创建定时器用于定期捕获屏幕
    m_screenCaptureTimer = new QTimer(this);
    connect(m_screenCaptureTimer, &QTimer::timeout,
            this, &MediaStream::onScreenCaptureTimeout);

    // 30 FPS
    m_screenCaptureTimer->start(33);

    m_hasVideo = true;
    emit hasVideoChanged();
    emit isScreenShareChanged();

    LOG_INFO("Screen share started");
    return true;
}

void MediaStream::stopScreenShare()
{
    if (!m_isScreenShare) {
        return;
    }

    LOG_INFO("Stopping screen share");

    if (m_screenCaptureTimer) {
        m_screenCaptureTimer->stop();
        m_screenCaptureTimer->deleteLater();
        m_screenCaptureTimer = nullptr;
    }

    m_screen = nullptr;
    m_isScreenShare = false;
    m_hasVideo = false;

    emit hasVideoChanged();
    emit isScreenShareChanged();
}

void MediaStream::setAudioEnabled(bool enabled)
{
    if (m_audioEnabled == enabled) {
        return;
    }

    m_audioEnabled = enabled;
    LOG_INFO(QString("Audio %1 for stream: %2").arg(enabled ? "enabled" : "disabled").arg(m_streamId));

    if (m_audioInput) {
        m_audioInput->setMuted(!enabled);
    }

    emit audioEnabledChanged();
}

void MediaStream::setVideoEnabled(bool enabled)
{
    if (m_videoEnabled == enabled) {
        return;
    }

    m_videoEnabled = enabled;
    LOG_INFO(QString("Video %1 for stream: %2").arg(enabled ? "enabled" : "disabled").arg(m_streamId));

    if (m_camera) {
        if (enabled) {
            m_camera->start();
        } else {
            m_camera->stop();
        }
    }

    emit videoEnabledChanged();
}

QStringList MediaStream::getAudioInputDevices() const
{
    QStringList devices;
    const QList<QAudioDevice> audioDevices = QMediaDevices::audioInputs();

    for (const QAudioDevice &device : audioDevices) {
        devices.append(device.description());
    }

    return devices;
}

QStringList MediaStream::getVideoInputDevices() const
{
    QStringList devices;
    const QList<QCameraDevice> videoDevices = QMediaDevices::videoInputs();

    for (const QCameraDevice &device : videoDevices) {
        devices.append(device.description());
    }

    return devices;
}

bool MediaStream::setAudioInputDevice(const QString &deviceName)
{
    const QList<QAudioDevice> audioDevices = QMediaDevices::audioInputs();

    for (const QAudioDevice &device : audioDevices) {
        if (device.description() == deviceName) {
            if (m_audioInput) {
                m_audioInput->setDevice(device);
                LOG_INFO(QString("Audio input device changed to: %1").arg(deviceName));
                return true;
            }
        }
    }

    LOG_WARNING(QString("Audio input device not found: %1").arg(deviceName));
    return false;
}

bool MediaStream::setVideoInputDevice(const QString &deviceName)
{
    const QList<QCameraDevice> videoDevices = QMediaDevices::videoInputs();

    for (const QCameraDevice &device : videoDevices) {
        if (device.description() == deviceName) {
            if (m_camera) {
                m_camera->stop();
                m_camera->deleteLater();
            }

            m_camera = new QCamera(device, this);
            if (m_captureSession) {
                m_captureSession->setCamera(m_camera);
            }

            connect(m_camera, &QCamera::errorOccurred,
                    this, &MediaStream::onCameraError);

            m_camera->start();
            LOG_INFO(QString("Video input device changed to: %1").arg(deviceName));
            return true;
        }
    }

    LOG_WARNING(QString("Video input device not found: %1").arg(deviceName));
    return false;
}

void MediaStream::onVideoFrameChanged(const QVideoFrame &frame)
{
    if (!m_videoEnabled) {
        return;
    }

    emit videoFrameReady(frame);
}

void MediaStream::onCameraError()
{
    if (m_camera) {
        QString errorString = QString("Camera error: %1").arg(static_cast<int>(m_camera->error()));
        LOG_ERROR(errorString);
        emit error(errorString);
    }
}

void MediaStream::onScreenCaptureTimeout()
{
    if (!m_screen || !m_isScreenShare) {
        return;
    }

    // 捕获屏幕
    QPixmap pixmap = m_screen->grabWindow(0);
    QImage image = pixmap.toImage();

    // 转换为QVideoFrame
    QVideoFrame frame(QVideoFrameFormat(image.size(), QVideoFrameFormat::Format_ARGB8888));
    if (frame.map(QVideoFrame::WriteOnly)) {
        memcpy(frame.bits(0), image.constBits(), image.sizeInBytes());
        frame.unmap();

        emit videoFrameReady(frame);
    }
}

