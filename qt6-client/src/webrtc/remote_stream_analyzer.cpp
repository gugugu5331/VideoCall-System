#include "webrtc/remote_stream_analyzer.h"
#include "webrtc/media_stream.h"
#include "services/ai_service.h"
#include "utils/logger.h"
#include <QDataStream>

RemoteStreamAnalyzer::RemoteStreamAnalyzer(int remoteUserId, AIService *aiService, QObject *parent)
    : QObject(parent)
    , m_remoteUserId(remoteUserId)
    , m_aiService(aiService)
    , m_stream(nullptr)
    , m_isAnalyzing(false)
    , m_videoAnalysisInterval(5000)  // 5 seconds
    , m_videoDownscaleSize(640, 360)
    , m_audioBufferDuration(3000)    // 3 seconds
    , m_audioTargetSampleRate(16000)
    , m_audioSourceSampleRate(48000)
    , m_audioChannels(1)
    , m_audioBitsPerSample(16)
    , m_deepfakeEnabled(true)
    , m_asrEnabled(true)
    , m_emotionEnabled(true)
{
    // 创建视频分析定时器
    m_videoAnalysisTimer = new QTimer(this);
    m_videoAnalysisTimer->setInterval(m_videoAnalysisInterval);
    connect(m_videoAnalysisTimer, &QTimer::timeout, this, &RemoteStreamAnalyzer::onVideoAnalysisTimeout);

    // 创建音频缓冲定时器
    m_audioBufferTimer = new QTimer(this);
    m_audioBufferTimer->setInterval(m_audioBufferDuration);
    connect(m_audioBufferTimer, &QTimer::timeout, this, &RemoteStreamAnalyzer::onAudioBufferTimeout);

    LOG_INFO(QString("RemoteStreamAnalyzer created for user: %1").arg(m_remoteUserId));
}

RemoteStreamAnalyzer::~RemoteStreamAnalyzer()
{
    stopAnalysis();
    detachFromStream();
    LOG_INFO(QString("RemoteStreamAnalyzer destroyed for user: %1").arg(m_remoteUserId));
}

void RemoteStreamAnalyzer::attachToStream(MediaStream *stream)
{
    if (m_stream == stream) {
        return;
    }

    // 断开旧流
    detachFromStream();

    m_stream = stream;

    if (m_stream) {
        // 连接视频帧信号
        connect(m_stream, &MediaStream::videoFrameReady,
                this, &RemoteStreamAnalyzer::onVideoFrameReady);

        // 连接音频数据信号
        connect(m_stream, &MediaStream::audioDataReady,
                this, &RemoteStreamAnalyzer::onAudioDataReady);

        LOG_INFO(QString("Attached to stream for user: %1").arg(m_remoteUserId));
    }
}

void RemoteStreamAnalyzer::detachFromStream()
{
    if (m_stream) {
        disconnect(m_stream, nullptr, this, nullptr);
        m_stream = nullptr;
        LOG_INFO(QString("Detached from stream for user: %1").arg(m_remoteUserId));
    }
}

void RemoteStreamAnalyzer::startAnalysis()
{
    if (m_isAnalyzing) {
        return;
    }

    m_isAnalyzing = true;

    // 清空缓冲区
    m_videoFrameBuffer.clear();
    m_audioDataBuffer.clear();

    // 启动定时器
    m_videoAnalysisTimer->start();
    m_audioBufferTimer->start();

    emit analysisStarted();
    LOG_INFO(QString("AI analysis started for user: %1").arg(m_remoteUserId));
}

void RemoteStreamAnalyzer::stopAnalysis()
{
    if (!m_isAnalyzing) {
        return;
    }

    m_isAnalyzing = false;

    // 停止定时器
    m_videoAnalysisTimer->stop();
    m_audioBufferTimer->stop();

    // 清空缓冲区
    m_videoFrameBuffer.clear();
    m_audioDataBuffer.clear();

    emit analysisStopped();
    LOG_INFO(QString("AI analysis stopped for user: %1").arg(m_remoteUserId));
}

void RemoteStreamAnalyzer::onVideoFrameReady(const QVideoFrame &frame)
{
    if (!m_isAnalyzing || !frame.isValid()) {
        return;
    }

    // 添加到缓冲区
    m_videoFrameBuffer.push_back(frame);

    // 限制缓冲区大小（最多保留最近30帧）
    if (m_videoFrameBuffer.size() > 30) {
        m_videoFrameBuffer.erase(m_videoFrameBuffer.begin());
    }
}

void RemoteStreamAnalyzer::onAudioDataReady(const QByteArray &data)
{
    if (!m_isAnalyzing || data.isEmpty()) {
        return;
    }

    // 累积音频数据
    m_audioDataBuffer.append(data);

    // 计算目标字节数（3秒的音频）
    int bytesPerSecond = m_audioSourceSampleRate * m_audioChannels * (m_audioBitsPerSample / 8);
    int targetBytes = bytesPerSecond * (m_audioBufferDuration / 1000);

    // 如果累积够了，立即分析
    if (m_audioDataBuffer.size() >= targetBytes) {
        analyzeAudioData();
    }
}

void RemoteStreamAnalyzer::onVideoAnalysisTimeout()
{
    analyzeVideoFrames();
}

void RemoteStreamAnalyzer::onAudioBufferTimeout()
{
    // 定时器触发时，如果有数据就分析
    if (!m_audioDataBuffer.isEmpty()) {
        analyzeAudioData();
    }
}

void RemoteStreamAnalyzer::analyzeVideoFrames()
{
    if (m_videoFrameBuffer.empty() || !m_deepfakeEnabled) {
        return;
    }

    LOG_DEBUG(QString("Analyzing video frames for user: %1 (buffer size: %2)")
              .arg(m_remoteUserId).arg(m_videoFrameBuffer.size()));

    // 取最新的一帧
    QVideoFrame frame = m_videoFrameBuffer.back();

    // 提取视频帧数据
    QByteArray videoData = extractVideoFrameData(frame);

    if (videoData.isEmpty()) {
        LOG_WARNING(QString("Failed to extract video frame data for user: %1").arg(m_remoteUserId));
        return;
    }

    // 清空缓冲区
    m_videoFrameBuffer.clear();

    // 调用AI服务进行深度伪造检测
    m_aiService->detectDeepfake(videoData, m_remoteUserId);

    LOG_DEBUG(QString("Sent video data for deepfake detection (user: %1, size: %2 bytes)")
              .arg(m_remoteUserId).arg(videoData.size()));
}

void RemoteStreamAnalyzer::analyzeAudioData()
{
    if (m_audioDataBuffer.isEmpty()) {
        return;
    }

    LOG_DEBUG(QString("Analyzing audio data for user: %1 (buffer size: %2 bytes)")
              .arg(m_remoteUserId).arg(m_audioDataBuffer.size()));

    // 重采样音频（如果需要）
    QByteArray audioData = m_audioDataBuffer;
    if (m_audioSourceSampleRate != m_audioTargetSampleRate) {
        audioData = resampleAudio(audioData, m_audioSourceSampleRate, m_audioTargetSampleRate);
    }

    // 转换为WAV格式
    QByteArray wavData = convertToWAV(audioData, m_audioTargetSampleRate, m_audioChannels, m_audioBitsPerSample);

    if (wavData.isEmpty()) {
        LOG_WARNING(QString("Failed to convert audio data to WAV for user: %1").arg(m_remoteUserId));
        m_audioDataBuffer.clear();
        return;
    }

    // 清空缓冲区
    m_audioDataBuffer.clear();

    // 调用AI服务
    if (m_asrEnabled) {
        m_aiService->recognizeSpeech(wavData, m_remoteUserId, "zh");
        LOG_DEBUG(QString("Sent audio data for ASR (user: %1, size: %2 bytes)")
                  .arg(m_remoteUserId).arg(wavData.size()));
    }

    if (m_emotionEnabled) {
        m_aiService->recognizeEmotion(wavData, m_remoteUserId);
        LOG_DEBUG(QString("Sent audio data for emotion detection (user: %1, size: %2 bytes)")
                  .arg(m_remoteUserId).arg(wavData.size()));
    }
}

QByteArray RemoteStreamAnalyzer::extractVideoFrameData(const QVideoFrame &frame)
{
    if (!frame.isValid()) {
        return QByteArray();
    }

    // 映射视频帧
    QVideoFrame f = frame;
    if (!f.map(QVideoFrame::ReadOnly)) {
        LOG_ERROR("Failed to map video frame");
        return QByteArray();
    }

    // 转换为QImage
    QImage image = f.toImage();
    f.unmap();

    if (image.isNull()) {
        LOG_ERROR("Failed to convert video frame to image");
        return QByteArray();
    }

    // 降采样
    QImage resized = downscaleImage(image, m_videoDownscaleSize);

    // 转换为JPEG格式
    QByteArray data;
    QBuffer buffer(&data);
    buffer.open(QIODevice::WriteOnly);
    resized.save(&buffer, "JPEG", 85);  // 85% quality

    return data;
}

QImage RemoteStreamAnalyzer::downscaleImage(const QImage &image, const QSize &targetSize)
{
    if (image.size() == targetSize) {
        return image;
    }

    return image.scaled(targetSize, Qt::KeepAspectRatio, Qt::SmoothTransformation);
}

QByteArray RemoteStreamAnalyzer::convertToWAV(const QByteArray &pcmData, int sampleRate, int channels, int bitsPerSample)
{
    QByteArray wavData;
    QDataStream stream(&wavData, QIODevice::WriteOnly);
    stream.setByteOrder(QDataStream::LittleEndian);

    // WAV文件头
    int dataSize = pcmData.size();
    int fileSize = 36 + dataSize;

    // RIFF chunk
    stream.writeRawData("RIFF", 4);
    stream << (quint32)fileSize;
    stream.writeRawData("WAVE", 4);

    // fmt chunk
    stream.writeRawData("fmt ", 4);
    stream << (quint32)16;  // fmt chunk size
    stream << (quint16)1;   // audio format (PCM)
    stream << (quint16)channels;
    stream << (quint32)sampleRate;
    stream << (quint32)(sampleRate * channels * bitsPerSample / 8);  // byte rate
    stream << (quint16)(channels * bitsPerSample / 8);  // block align
    stream << (quint16)bitsPerSample;

    // data chunk
    stream.writeRawData("data", 4);
    stream << (quint32)dataSize;
    stream.writeRawData(pcmData.data(), dataSize);

    return wavData;
}

QByteArray RemoteStreamAnalyzer::resampleAudio(const QByteArray &audioData, int fromRate, int toRate)
{
    // 简单的线性插值重采样
    // 注意：这是一个简化实现，生产环境应该使用专业的重采样库（如libsamplerate）

    if (fromRate == toRate) {
        return audioData;
    }

    const qint16 *input = reinterpret_cast<const qint16*>(audioData.data());
    int inputSamples = audioData.size() / sizeof(qint16);

    double ratio = static_cast<double>(toRate) / fromRate;
    int outputSamples = static_cast<int>(inputSamples * ratio);

    QByteArray output;
    output.resize(outputSamples * sizeof(qint16));
    qint16 *outputPtr = reinterpret_cast<qint16*>(output.data());

    for (int i = 0; i < outputSamples; ++i) {
        double srcIndex = i / ratio;
        int index1 = static_cast<int>(srcIndex);
        int index2 = qMin(index1 + 1, inputSamples - 1);
        double frac = srcIndex - index1;

        // 线性插值
        qint16 sample1 = input[index1];
        qint16 sample2 = input[index2];
        outputPtr[i] = static_cast<qint16>(sample1 * (1.0 - frac) + sample2 * frac);
    }

    return output;
}

