#ifndef SECURITYMANAGER_H
#define SECURITYMANAGER_H

#include <QObject>
#include <QThread>
#include <QMutex>
#include <QTimer>
#include <QImage>
#include <QAudioBuffer>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>

class SecurityManager : public QObject
{
    Q_OBJECT

public:
    explicit SecurityManager(QObject *parent = nullptr);
    ~SecurityManager();

    // 安全检测控制
    void startDetection();
    void stopDetection();
    void enableFaceDetection(bool enabled);
    void enableVoiceDetection(bool enabled);
    void enableVideoDetection(bool enabled);

    // 数据处理
    void processVideoFrame(const QImage &frame);
    void processAudioFrame(const QAudioBuffer &buffer);
    void processVideoStream(const QByteArray &data);

    // 配置
    void setDetectionThreshold(double threshold);
    void setDetectionInterval(int interval);
    void setModelPath(const QString &path);

signals:
    // 检测结果
    void faceDetectionResult(const QRect &faceRect, double confidence, bool isReal);
    void voiceDetectionResult(bool isSpoofed, double confidence, const QString &details);
    void videoDetectionResult(bool isDeepfake, double confidence, const QString &details);
    
    // 综合安全状态
    void securityAlert(const QString &alertType, double riskScore, const QString &details);
    void securityStatusChanged(const QString &status);
    
    // 检测状态
    void detectionStarted();
    void detectionStopped();
    void detectionError(const QString &error);

private slots:
    void onDetectionTimer();
    void onApiResponse(QNetworkReply *reply);

private:
    // 检测算法
    void detectFace(const QImage &frame);
    void detectVoiceSpoofing(const QAudioBuffer &buffer);
    void detectVideoDeepfake(const QImage &frame);
    void analyzeSecurityPatterns();
    
    // 工具方法
    QImage preprocessFrame(const QImage &frame);
    QByteArray preprocessAudio(const QAudioBuffer &buffer);
    double calculateRiskScore();
    QString generateSecurityReport();

private:
    QThread *m_detectionThread;
    QTimer *m_detectionTimer;
    QMutex m_detectionMutex;
    QNetworkAccessManager *m_networkManager;

    // 检测状态
    bool m_isDetecting;
    bool m_faceDetectionEnabled;
    bool m_voiceDetectionEnabled;
    bool m_videoDetectionEnabled;

    // 配置参数
    double m_detectionThreshold;
    int m_detectionInterval;
    QString m_modelPath;

    // 检测结果缓存
    QList<QRect> m_detectedFaces;
    QList<double> m_voiceConfidenceScores;
    QList<double> m_videoConfidenceScores;
    
    // 统计数据
    int m_totalFramesProcessed;
    int m_suspiciousFramesDetected;
    double m_averageRiskScore;
    QDateTime m_lastDetectionTime;
};

#endif // SECURITYMANAGER_H 