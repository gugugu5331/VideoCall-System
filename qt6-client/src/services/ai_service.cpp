#include "services/ai_service.h"
#include "utils/logger.h"

AIService::AIService(ApiClient *apiClient, QObject *parent)
    : QObject(parent)
    , m_apiClient(apiClient)
    , m_detectionEnabled(false)
    , m_asrEnabled(false)
    , m_emotionEnabled(false)
{
    m_detectionTimer = new QTimer(this);
    connect(m_detectionTimer, &QTimer::timeout, this, &AIService::performDetection);
}

AIService::~AIService()
{
}

void AIService::detectDeepfake(const QByteArray &videoData, int userId)
{
    LOG_DEBUG(QString("Performing synthesis detection for user: %1").arg(userId));

    // 使用synthesisDetection API进行深度伪造检测，传递userId参数
    m_apiClient->synthesisDetection(videoData, userId, [this, userId](const ApiResponse &response) {
        if (response.isSuccess()) {
            DeepfakeDetectionResult result;
            result.userId = userId;
            result.isReal = !response.data["is_synthetic"].toBool(); // is_synthetic的反义
            result.confidence = response.data["confidence"].toDouble();
            result.timestamp = QDateTime::currentDateTime();

            emit deepfakeDetected(result);

            LOG_INFO(QString("Deepfake detection completed for user %1: %2 (confidence: %3)")
                     .arg(userId)
                     .arg(result.isReal ? "Real" : "Synthetic")
                     .arg(result.confidence));
        } else {
            LOG_ERROR(QString("Synthesis detection failed for user %1: %2").arg(userId).arg(response.message));
        }
    });
}

void AIService::recognizeSpeech(const QByteArray &audioData, int userId, const QString &language)
{
    LOG_DEBUG(QString("Performing speech recognition for user: %1").arg(userId));

    // 使用speechRecognition API，参数匹配ApiClient的实际接口
    // speechRecognition(audioData, audioFormat, sampleRate, language, userId, callback)
    QString audioFormat = "wav"; // 默认格式
    int sampleRate = 16000;      // 默认采样率

    m_apiClient->speechRecognition(audioData, audioFormat, sampleRate, language, userId,
        [this, userId](const ApiResponse &response) {
            if (response.isSuccess()) {
                ASRResult result;
                result.userId = userId;
                result.text = response.data["text"].toString();
                result.confidence = response.data["confidence"].toDouble();
                result.timestamp = QDateTime::currentDateTime();

                emit speechRecognized(result);

                LOG_INFO(QString("Speech recognized for user %1: %2 (confidence: %3)")
                         .arg(userId)
                         .arg(result.text)
                         .arg(result.confidence));
            } else {
                LOG_ERROR(QString("Speech recognition failed for user %1: %2").arg(userId).arg(response.message));
            }
        });
}

void AIService::recognizeEmotion(const QByteArray &audioData, int userId)
{
    LOG_DEBUG(QString("Performing emotion detection for user: %1").arg(userId));

    // 使用emotionDetection API，参数匹配ApiClient的实际接口
    // emotionDetection(audioData, audioFormat, sampleRate, userId, callback)
    QString audioFormat = "wav"; // 默认格式
    int sampleRate = 16000;      // 默认采样率

    m_apiClient->emotionDetection(audioData, audioFormat, sampleRate, userId,
        [this, userId](const ApiResponse &response) {
            if (response.isSuccess()) {
                EmotionRecognitionResult result;
                result.userId = userId;
                result.emotion = response.data["emotion"].toString();
                result.confidence = response.data["confidence"].toDouble();
                result.timestamp = QDateTime::currentDateTime();

                // Parse emotions map if available
                if (response.data.contains("emotions")) {
                    QJsonObject emotionsObj = response.data["emotions"].toObject();
                    for (auto it = emotionsObj.begin(); it != emotionsObj.end(); ++it) {
                        result.emotions[it.key()] = it.value().toDouble();
                    }
                }

                emit emotionRecognized(result);

                LOG_INFO(QString("Emotion recognized for user %1: %2 (confidence: %3)")
                         .arg(userId)
                         .arg(result.emotion)
                         .arg(result.confidence));
            } else {
                LOG_ERROR(QString("Emotion detection failed for user %1: %2").arg(userId).arg(response.message));
            }
        });
}

void AIService::denoiseAudio(const QByteArray &audioData)
{
    LOG_DEBUG("Performing audio denoising");

    m_apiClient->audioDenoising(audioData, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            // 解码base64音频数据
            QString denoisedAudioBase64 = response.data["denoised_audio"].toString();
            QByteArray denoisedAudio = QByteArray::fromBase64(denoisedAudioBase64.toUtf8());

            emit audioDenoised(denoisedAudio);
        } else {
            LOG_ERROR("Audio denoising failed: " + response.message);
        }
    });
}

void AIService::enhanceVideo(const QByteArray &videoData, const QString &enhancementType)
{
    LOG_DEBUG("Performing video enhancement: " + enhancementType);

    m_apiClient->videoEnhancement(videoData, enhancementType, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            // 解码base64视频数据
            QString enhancedVideoBase64 = response.data["enhanced_video"].toString();
            QByteArray enhancedVideo = QByteArray::fromBase64(enhancedVideoBase64.toUtf8());

            emit videoEnhanced(enhancedVideo);
        } else {
            LOG_ERROR("Video enhancement failed: " + response.message);
        }
    });
}

void AIService::startContinuousDetection(int intervalMs)
{
    LOG_INFO(QString("Starting continuous AI detection (interval: %1ms)").arg(intervalMs));
    m_detectionTimer->start(intervalMs);
}

void AIService::stopContinuousDetection()
{
    LOG_INFO("Stopping continuous AI detection");
    m_detectionTimer->stop();
}

void AIService::setDetectionEnabled(bool enabled)
{
    if (m_detectionEnabled != enabled) {
        m_detectionEnabled = enabled;
        emit detectionEnabledChanged();
        LOG_INFO(QString("Deepfake detection %1").arg(enabled ? "enabled" : "disabled"));
    }
}

void AIService::setAsrEnabled(bool enabled)
{
    if (m_asrEnabled != enabled) {
        m_asrEnabled = enabled;
        emit asrEnabledChanged();
        LOG_INFO(QString("ASR %1").arg(enabled ? "enabled" : "disabled"));
    }
}

void AIService::setEmotionEnabled(bool enabled)
{
    if (m_emotionEnabled != enabled) {
        m_emotionEnabled = enabled;
        emit emotionEnabledChanged();
        LOG_INFO(QString("Emotion recognition %1").arg(enabled ? "enabled" : "disabled"));
    }
}

void AIService::performDetection()
{
    // TODO: Capture current video/audio frames and perform detection
    // This would integrate with WebRTC media streams
    
    // For now, this is a placeholder
    LOG_DEBUG("Performing periodic AI detection");
}

