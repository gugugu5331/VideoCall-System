#ifndef MEDIA_SERVICE_H
#define MEDIA_SERVICE_H

#include <QObject>
#include <QCamera>
#include <QAudioInput>
#include <QAudioOutput>
#include <QMediaRecorder>
#include <QVideoFrame>
#include <QAudioBuffer>
#include <QMediaDevices>
#include <QVideoSink>
#include <QAudioSink>
#include <QAudioSource>
#include <QTimer>
#include <QThread>
#include <QMutex>
#include <QQueue>
#include <memory>

extern "C" {
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libswscale/swscale.h>
#include <libswresample/swresample.h>
}

/**
 * @brief 媒体格式枚举
 */
enum class MediaFormat {
    H264,
    H265,
    VP8,
    VP9,
    AAC,
    OPUS,
    PCM
};

/**
 * @brief 视频编码参数
 */
struct VideoEncodeParams {
    MediaFormat codec = MediaFormat::H264;
    int width = 1280;
    int height = 720;
    int fps = 30;
    int bitrate = 2000000; // 2Mbps
    int keyFrameInterval = 30;
    QString preset = "medium";
    QString profile = "main";
};

/**
 * @brief 音频编码参数
 */
struct AudioEncodeParams {
    MediaFormat codec = MediaFormat::OPUS;
    int sampleRate = 48000;
    int channels = 2;
    int bitrate = 128000; // 128kbps
    int frameSize = 960;
};

/**
 * @brief 媒体统计信息
 */
struct MediaStats {
    // 视频统计
    int videoFramesEncoded = 0;
    int videoFramesDecoded = 0;
    int videoFramesDropped = 0;
    double videoEncodeFps = 0.0;
    double videoDecodeFps = 0.0;
    int videoBitrate = 0;
    
    // 音频统计
    int audioFramesEncoded = 0;
    int audioFramesDecoded = 0;
    int audioFramesDropped = 0;
    int audioBitrate = 0;
    
    // 网络统计
    qint64 bytesSent = 0;
    qint64 bytesReceived = 0;
    double packetLossRate = 0.0;
    int rtt = 0; // Round Trip Time
};

/**
 * @brief FFmpeg编码器类
 */
class FFmpegEncoder : public QObject
{
    Q_OBJECT

public:
    explicit FFmpegEncoder(QObject *parent = nullptr);
    ~FFmpegEncoder();

    bool initializeVideo(const VideoEncodeParams &params);
    bool initializeAudio(const AudioEncodeParams &params);
    void cleanup();

    QByteArray encodeVideoFrame(const QVideoFrame &frame);
    QByteArray encodeAudioFrame(const QAudioBuffer &buffer);

signals:
    void encodedDataReady(const QByteArray &data, bool isVideo);
    void error(const QString &message);

private:
    // FFmpeg上下文
    AVCodecContext *m_videoCodecContext = nullptr;
    AVCodecContext *m_audioCodecContext = nullptr;
    AVFrame *m_videoFrame = nullptr;
    AVFrame *m_audioFrame = nullptr;
    AVPacket *m_packet = nullptr;
    SwsContext *m_swsContext = nullptr;
    SwrContext *m_swrContext = nullptr;
    
    // 参数
    VideoEncodeParams m_videoParams;
    AudioEncodeParams m_audioParams;
    
    // 帧计数
    int64_t m_videoFrameCount = 0;
    int64_t m_audioFrameCount = 0;
};

/**
 * @brief FFmpeg解码器类
 */
class FFmpegDecoder : public QObject
{
    Q_OBJECT

public:
    explicit FFmpegDecoder(QObject *parent = nullptr);
    ~FFmpegDecoder();

    bool initialize();
    void cleanup();

    bool decodeVideoData(const QByteArray &data);
    bool decodeAudioData(const QByteArray &data);

signals:
    void videoFrameReady(const QVideoFrame &frame);
    void audioBufferReady(const QAudioBuffer &buffer);
    void error(const QString &message);

private:
    // FFmpeg上下文
    AVCodecContext *m_videoCodecContext = nullptr;
    AVCodecContext *m_audioCodecContext = nullptr;
    AVFrame *m_frame = nullptr;
    AVPacket *m_packet = nullptr;
    SwsContext *m_swsContext = nullptr;
    SwrContext *m_swrContext = nullptr;
    
    // 解析器
    AVCodecParserContext *m_videoParser = nullptr;
    AVCodecParserContext *m_audioParser = nullptr;
};

/**
 * @brief 媒体服务类
 * 
 * 负责音视频采集、编码、解码、播放等功能
 */
class MediaService : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool cameraEnabled READ isCameraEnabled WRITE setCameraEnabled NOTIFY cameraEnabledChanged)
    Q_PROPERTY(bool microphoneEnabled READ isMicrophoneEnabled WRITE setMicrophoneEnabled NOTIFY microphoneEnabledChanged)
    Q_PROPERTY(bool speakerEnabled READ isSpeakerEnabled WRITE setSpeakerEnabled NOTIFY speakerEnabledChanged)
    Q_PROPERTY(QStringList availableCameras READ availableCameras NOTIFY availableCamerasChanged)
    Q_PROPERTY(QStringList availableMicrophones READ availableMicrophones NOTIFY availableMicrophonesChanged)
    Q_PROPERTY(QStringList availableSpeakers READ availableSpeakers NOTIFY availableSpeakersChanged)

public:
    explicit MediaService(QObject *parent = nullptr);
    ~MediaService();

    // 属性访问器
    bool isCameraEnabled() const { return m_cameraEnabled; }
    bool isMicrophoneEnabled() const { return m_microphoneEnabled; }
    bool isSpeakerEnabled() const { return m_speakerEnabled; }
    QStringList availableCameras() const { return m_availableCameras; }
    QStringList availableMicrophones() const { return m_availableMicrophones; }
    QStringList availableSpeakers() const { return m_availableSpeakers; }

    /**
     * @brief 初始化媒体服务
     */
    bool initialize();

    /**
     * @brief 清理资源
     */
    void cleanup();

    /**
     * @brief 暂停媒体服务
     */
    void pause();

    /**
     * @brief 恢复媒体服务
     */
    void resume();

    /**
     * @brief 设置视频编码参数
     */
    void setVideoEncodeParams(const VideoEncodeParams &params);

    /**
     * @brief 设置音频编码参数
     */
    void setAudioEncodeParams(const AudioEncodeParams &params);

    /**
     * @brief 获取媒体统计信息
     */
    MediaStats getMediaStats() const;

    /**
     * @brief 获取当前视频帧
     */
    QVideoFrame getCurrentVideoFrame() const;

public slots:
    /**
     * @brief 设置摄像头启用状态
     */
    void setCameraEnabled(bool enabled);

    /**
     * @brief 设置麦克风启用状态
     */
    void setMicrophoneEnabled(bool enabled);

    /**
     * @brief 设置扬声器启用状态
     */
    void setSpeakerEnabled(bool enabled);

    /**
     * @brief 切换摄像头
     */
    void switchCamera(const QString &cameraId);

    /**
     * @brief 切换麦克风
     */
    void switchMicrophone(const QString &microphoneId);

    /**
     * @brief 切换扬声器
     */
    void switchSpeaker(const QString &speakerId);

    /**
     * @brief 开始录制
     */
    void startRecording(const QString &filePath);

    /**
     * @brief 停止录制
     */
    void stopRecording();

    /**
     * @brief 播放远程音视频数据
     */
    void playRemoteMedia(const QString &participantId, const QByteArray &data, bool isVideo);

    /**
     * @brief 更新设备列表
     */
    void updateDeviceList();

signals:
    void cameraEnabledChanged();
    void microphoneEnabledChanged();
    void speakerEnabledChanged();
    void availableCamerasChanged();
    void availableMicrophonesChanged();
    void availableSpeakersChanged();
    
    void localVideoFrameReady(const QVideoFrame &frame);
    void localAudioBufferReady(const QAudioBuffer &buffer);
    void encodedVideoDataReady(const QByteArray &data);
    void encodedAudioDataReady(const QByteArray &data);
    
    void remoteVideoFrameReady(const QString &participantId, const QVideoFrame &frame);
    void remoteAudioBufferReady(const QString &participantId, const QAudioBuffer &buffer);
    
    void recordingStarted();
    void recordingStopped();
    void recordingError(const QString &error);
    
    void mediaStatsUpdated(const MediaStats &stats);
    void error(const QString &message);

private slots:
    void onCameraFrameReady(const QVideoFrame &frame);
    void onAudioBufferReady(const QAudioBuffer &buffer);
    void onDevicesChanged();
    void onStatsTimer();

private:
    /**
     * @brief 初始化FFmpeg
     */
    bool initializeFFmpeg();

    /**
     * @brief 初始化设备
     */
    void initializeDevices();

    /**
     * @brief 更新统计信息
     */
    void updateStats();

private:
    // Qt媒体对象
    std::unique_ptr<QCamera> m_camera;
    std::unique_ptr<QAudioInput> m_audioInput;
    std::unique_ptr<QAudioOutput> m_audioOutput;
    std::unique_ptr<QMediaRecorder> m_recorder;
    std::unique_ptr<QVideoSink> m_videoSink;
    std::unique_ptr<QAudioSink> m_audioSink;
    std::unique_ptr<QAudioSource> m_audioSource;
    
    // FFmpeg编解码器
    std::unique_ptr<FFmpegEncoder> m_encoder;
    QMap<QString, std::unique_ptr<FFmpegDecoder>> m_decoders;
    
    // 状态
    bool m_cameraEnabled;
    bool m_microphoneEnabled;
    bool m_speakerEnabled;
    bool m_recording;
    bool m_initialized;
    
    // 设备列表
    QStringList m_availableCameras;
    QStringList m_availableMicrophones;
    QStringList m_availableSpeakers;
    QString m_currentCameraId;
    QString m_currentMicrophoneId;
    QString m_currentSpeakerId;
    
    // 编码参数
    VideoEncodeParams m_videoParams;
    AudioEncodeParams m_audioParams;
    
    // 统计信息
    MediaStats m_stats;
    QTimer *m_statsTimer;
    
    // 线程安全
    mutable QMutex m_mutex;
    
    // 当前帧缓存
    QVideoFrame m_currentVideoFrame;
};

#endif // MEDIA_SERVICE_H
