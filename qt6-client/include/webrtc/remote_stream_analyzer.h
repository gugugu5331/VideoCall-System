#ifndef REMOTE_STREAM_ANALYZER_H
#define REMOTE_STREAM_ANALYZER_H

#include <QObject>
#include <QVideoFrame>
#include <QAudioBuffer>
#include <QTimer>
#include <QByteArray>
#include <QImage>
#include <QBuffer>
#include <QMap>
#include <vector>

// Forward declarations
class AIService;
class MediaStream;

/**
 * @brief RemoteStreamAnalyzer - 远程流AI分析器
 * 
 * 负责从远程用户的MediaStream中提取音视频数据，
 * 并定时发送给AI服务进行分析
 */
class RemoteStreamAnalyzer : public QObject
{
    Q_OBJECT

public:
    explicit RemoteStreamAnalyzer(int remoteUserId, AIService *aiService, QObject *parent = nullptr);
    ~RemoteStreamAnalyzer();

    int remoteUserId() const { return m_remoteUserId; }

    // 连接到远程流
    void attachToStream(MediaStream *stream);
    void detachFromStream();

    // 控制AI分析
    void startAnalysis();
    void stopAnalysis();
    bool isAnalyzing() const { return m_isAnalyzing; }

    // 配置
    void setVideoAnalysisInterval(int ms) { m_videoAnalysisInterval = ms; }
    void setAudioBufferDuration(int ms) { m_audioBufferDuration = ms; }
    void setVideoDownscaleSize(const QSize &size) { m_videoDownscaleSize = size; }
    void setAudioSampleRate(int rate) { m_audioTargetSampleRate = rate; }

    // 启用/禁用特定AI功能
    void setDeepfakeDetectionEnabled(bool enabled) { m_deepfakeEnabled = enabled; }
    void setAsrEnabled(bool enabled) { m_asrEnabled = enabled; }
    void setEmotionDetectionEnabled(bool enabled) { m_emotionEnabled = enabled; }

signals:
    void analysisStarted();
    void analysisStopped();
    void error(const QString &errorMsg);

private slots:
    void onVideoFrameReady(const QVideoFrame &frame);
    void onAudioDataReady(const QByteArray &data);
    void onVideoAnalysisTimeout();
    void onAudioBufferTimeout();

private:
    // 数据提取和编码
    QByteArray extractVideoFrameData(const QVideoFrame &frame);
    QByteArray convertToWAV(const QByteArray &pcmData, int sampleRate, int channels, int bitsPerSample);
    QImage downscaleImage(const QImage &image, const QSize &targetSize);
    QByteArray resampleAudio(const QByteArray &audioData, int fromRate, int toRate);

    // AI分析触发
    void analyzeVideoFrames();
    void analyzeAudioData();

private:
    int m_remoteUserId;
    AIService *m_aiService;
    MediaStream *m_stream;
    bool m_isAnalyzing;

    // 视频帧缓冲
    std::vector<QVideoFrame> m_videoFrameBuffer;
    QTimer *m_videoAnalysisTimer;
    int m_videoAnalysisInterval;  // 默认 5000ms
    QSize m_videoDownscaleSize;   // 默认 640x360

    // 音频数据缓冲
    QByteArray m_audioDataBuffer;
    QTimer *m_audioBufferTimer;
    int m_audioBufferDuration;    // 默认 3000ms
    int m_audioTargetSampleRate;  // 默认 16000Hz
    int m_audioSourceSampleRate;  // 源采样率
    int m_audioChannels;          // 声道数
    int m_audioBitsPerSample;     // 位深度

    // AI功能开关
    bool m_deepfakeEnabled;
    bool m_asrEnabled;
    bool m_emotionEnabled;
};

#endif // REMOTE_STREAM_ANALYZER_H

