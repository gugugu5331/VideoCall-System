#ifndef AI_SERVICE_H
#define AI_SERVICE_H

#include <QObject>
#include <QString>
#include <QJsonObject>
#include <QTimer>
#include "network/api_client.h"

struct DeepfakeDetectionResult {
    int userId;
    QString username;
    bool isReal;
    double confidence;
    QString videoStatus;
    QString audioStatus;
    QDateTime timestamp;
};

struct EmotionRecognitionResult {
    int userId;
    QString username;
    QString emotion;
    double confidence;
    QString engagement;
    QMap<QString, double> emotions; // 各种情绪的置信度
    QDateTime timestamp;
};

struct ASRResult {
    int userId;
    QString username;
    QString text;
    double confidence;
    QDateTime timestamp;
};

class AIService : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool detectionEnabled READ detectionEnabled WRITE setDetectionEnabled NOTIFY detectionEnabledChanged)
    Q_PROPERTY(bool asrEnabled READ asrEnabled WRITE setAsrEnabled NOTIFY asrEnabledChanged)
    Q_PROPERTY(bool emotionEnabled READ emotionEnabled WRITE setEmotionEnabled NOTIFY emotionEnabledChanged)

public:
    explicit AIService(ApiClient *apiClient, QObject *parent = nullptr);
    ~AIService();

    // Enable/disable features
    bool detectionEnabled() const { return m_detectionEnabled; }
    bool asrEnabled() const { return m_asrEnabled; }
    bool emotionEnabled() const { return m_emotionEnabled; }

    void setDetectionEnabled(bool enabled);
    void setAsrEnabled(bool enabled);
    void setEmotionEnabled(bool enabled);

    // AI operations
    Q_INVOKABLE void detectDeepfake(const QByteArray &videoData, int userId);
    Q_INVOKABLE void recognizeEmotion(const QByteArray &audioData, int userId);
    Q_INVOKABLE void recognizeSpeech(const QByteArray &audioData, int userId, const QString &language = "zh");
    Q_INVOKABLE void denoiseAudio(const QByteArray &audioData);
    Q_INVOKABLE void enhanceVideo(const QByteArray &videoData, const QString &enhancementType = "denoise");

    // Batch operations for all participants
    Q_INVOKABLE void startContinuousDetection(int intervalMs = 5000);
    Q_INVOKABLE void stopContinuousDetection();

signals:
    void detectionEnabledChanged();
    void asrEnabledChanged();
    void emotionEnabledChanged();

    void deepfakeDetected(const DeepfakeDetectionResult &result);
    void emotionRecognized(const EmotionRecognitionResult &result);
    void speechRecognized(const ASRResult &result);
    void audioDenoised(const QByteArray &denoisedAudio);
    void videoEnhanced(const QByteArray &enhancedVideo);

    void aiError(const QString &error);

private slots:
    void performDetection();

private:

private:
    ApiClient *m_apiClient;
    QTimer *m_detectionTimer;
    
    bool m_detectionEnabled;
    bool m_asrEnabled;
    bool m_emotionEnabled;
    
    QMap<int, DeepfakeDetectionResult> m_detectionResults;
    QMap<int, EmotionRecognitionResult> m_emotionResults;
    QList<ASRResult> m_asrResults;
};

Q_DECLARE_METATYPE(DeepfakeDetectionResult)
Q_DECLARE_METATYPE(EmotionRecognitionResult)
Q_DECLARE_METATYPE(ASRResult)

#endif // AI_SERVICE_H

