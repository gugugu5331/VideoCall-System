#ifndef AI_PANEL_CONTROLLER_H
#define AI_PANEL_CONTROLLER_H

#include <QObject>
#include <QVariantList>
#include <QVariantMap>
#include <QMap>

// 前向声明
class AIService;
class MeetingService;
struct DeepfakeDetectionResult;
struct ASRResult;
struct EmotionRecognitionResult;

class AIPanelController : public QObject
{
    Q_OBJECT
    Q_PROPERTY(QVariantList detectionResults READ detectionResults NOTIFY detectionResultsChanged)
    Q_PROPERTY(QVariantList emotionResults READ emotionResults NOTIFY emotionResultsChanged)
    Q_PROPERTY(QVariantList asrResults READ asrResults NOTIFY asrResultsChanged)
    Q_PROPERTY(bool detectionEnabled READ detectionEnabled NOTIFY detectionEnabledChanged)
    Q_PROPERTY(bool asrEnabled READ asrEnabled NOTIFY asrEnabledChanged)
    Q_PROPERTY(bool emotionEnabled READ emotionEnabled NOTIFY emotionEnabledChanged)

public:
    explicit AIPanelController(QObject *parent = nullptr);
    ~AIPanelController();

    QVariantList detectionResults() const { return m_detectionResults; }
    QVariantList emotionResults() const { return m_emotionResults; }
    QVariantList asrResults() const { return m_asrResults; }

    bool detectionEnabled() const { return m_detectionEnabled; }
    bool asrEnabled() const { return m_asrEnabled; }
    bool emotionEnabled() const { return m_emotionEnabled; }

    Q_INVOKABLE void enableDetection(bool enabled);
    Q_INVOKABLE void enableASR(bool enabled);
    Q_INVOKABLE void enableEmotion(bool enabled);
    Q_INVOKABLE void clearResults();

    // 新增：按用户ID查询结果
    Q_INVOKABLE QVariantMap getDetectionResultForUser(int userId) const;
    Q_INVOKABLE QVariantMap getEmotionResultForUser(int userId) const;
    Q_INVOKABLE QVariantList getAsrResultsForUser(int userId) const;
    Q_INVOKABLE QString getUsernameById(int userId) const;

signals:
    void detectionResultsChanged();
    void emotionResultsChanged();
    void asrResultsChanged();
    void detectionEnabledChanged();
    void asrEnabledChanged();
    void emotionEnabledChanged();

private slots:
    void onDeepfakeDetected(const DeepfakeDetectionResult &result);
    void onSpeechRecognized(const ASRResult &result);
    void onEmotionRecognized(const EmotionRecognitionResult &result);

private:
    AIService *m_aiService;

    // 扁平化列表（用于ListView显示）
    QVariantList m_detectionResults;
    QVariantList m_emotionResults;
    QVariantList m_asrResults;

    // 按用户ID分组存储（用于查询）
    QMap<int, QVariantMap> m_detectionResultsByUser;
    QMap<int, QVariantMap> m_emotionResultsByUser;
    QMap<int, QVariantList> m_asrResultsByUser;

    bool m_detectionEnabled;
    bool m_asrEnabled;
    bool m_emotionEnabled;

    int m_maxResultsPerUser;  // 每个用户最多保留的结果数量
};

#endif // AI_PANEL_CONTROLLER_H

