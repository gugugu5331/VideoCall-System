#include "ui/ai_panel_controller.h"
#include "application.h"
#include "services/ai_service.h"
#include "services/meeting_service.h"
#include "utils/logger.h"

AIPanelController::AIPanelController(QObject *parent)
    : QObject(parent)
    , m_aiService(nullptr)
    , m_detectionEnabled(false)
    , m_asrEnabled(false)
    , m_emotionEnabled(false)
    , m_maxResultsPerUser(20)  // 默认每个用户最多保留20条结果
{
    m_aiService = Application::instance()->aiService();

    // Connect to AI service signals
    connect(m_aiService, &AIService::deepfakeDetected, this, &AIPanelController::onDeepfakeDetected);
    connect(m_aiService, &AIService::speechRecognized, this, &AIPanelController::onSpeechRecognized);
    connect(m_aiService, &AIService::emotionRecognized, this, &AIPanelController::onEmotionRecognized);
}

AIPanelController::~AIPanelController()
{
}

void AIPanelController::enableDetection(bool enabled)
{
    if (m_detectionEnabled == enabled) {
        return;
    }

    m_aiService->setDetectionEnabled(enabled);

    if (enabled) {
        m_aiService->startContinuousDetection(5000); // 5 seconds interval
    } else {
        m_aiService->stopContinuousDetection();
    }

    m_detectionEnabled = enabled;
    emit detectionEnabledChanged();

    LOG_INFO(QString("Deepfake detection %1").arg(enabled ? "enabled" : "disabled"));
}

void AIPanelController::enableASR(bool enabled)
{
    if (m_asrEnabled == enabled) {
        return;
    }

    m_aiService->setAsrEnabled(enabled);

    m_asrEnabled = enabled;
    emit asrEnabledChanged();

    LOG_INFO(QString("ASR %1").arg(enabled ? "enabled" : "disabled"));
}

void AIPanelController::enableEmotion(bool enabled)
{
    if (m_emotionEnabled == enabled) {
        return;
    }

    m_aiService->setEmotionEnabled(enabled);

    m_emotionEnabled = enabled;
    emit emotionEnabledChanged();

    LOG_INFO(QString("Emotion recognition %1").arg(enabled ? "enabled" : "disabled"));
}

void AIPanelController::clearResults()
{
    m_detectionResults.clear();
    emit detectionResultsChanged();

    m_emotionResults.clear();
    emit emotionResultsChanged();

    m_asrResults.clear();
    emit asrResultsChanged();

    LOG_INFO("AI results cleared");
}

void AIPanelController::onDeepfakeDetected(const DeepfakeDetectionResult &result)
{
    QVariantMap resultMap;
    resultMap["userId"] = result.userId;
    resultMap["username"] = getUsernameById(result.userId);
    resultMap["isReal"] = result.isReal;
    resultMap["confidence"] = result.confidence;
    resultMap["timestamp"] = result.timestamp;

    // 添加到扁平化列表
    m_detectionResults.append(resultMap);

    // 按用户ID存储（每个用户只保留最新结果）
    m_detectionResultsByUser[result.userId] = resultMap;

    // Keep only last 10 results in flat list
    if (m_detectionResults.size() > 10) {
        m_detectionResults.removeFirst();
    }

    emit detectionResultsChanged();

    LOG_INFO(QString("Deepfake detection result added for user %1: %2 (confidence: %3)")
             .arg(result.userId)
             .arg(result.isReal ? "Real" : "Synthetic")
             .arg(result.confidence));
}

void AIPanelController::onSpeechRecognized(const ASRResult &result)
{
    QVariantMap resultMap;
    resultMap["userId"] = result.userId;
    resultMap["username"] = getUsernameById(result.userId);
    resultMap["text"] = result.text;
    resultMap["confidence"] = result.confidence;
    resultMap["timestamp"] = result.timestamp;

    // 添加到扁平化列表
    m_asrResults.append(resultMap);

    // 按用户ID分组存储
    if (!m_asrResultsByUser.contains(result.userId)) {
        m_asrResultsByUser[result.userId] = QVariantList();
    }
    m_asrResultsByUser[result.userId].append(resultMap);

    // 限制每个用户的结果数量
    if (m_asrResultsByUser[result.userId].size() > m_maxResultsPerUser) {
        m_asrResultsByUser[result.userId].removeFirst();
    }

    // Keep only last 20 results in flat list
    if (m_asrResults.size() > 20) {
        m_asrResults.removeFirst();
    }

    emit asrResultsChanged();

    LOG_INFO(QString("ASR result added for user %1: %2 (confidence: %3)")
             .arg(result.userId)
             .arg(result.text)
             .arg(result.confidence));
}

void AIPanelController::onEmotionRecognized(const EmotionRecognitionResult &result)
{
    QVariantMap resultMap;
    resultMap["userId"] = result.userId;
    resultMap["username"] = getUsernameById(result.userId);
    resultMap["emotion"] = result.emotion;
    resultMap["confidence"] = result.confidence;
    resultMap["timestamp"] = result.timestamp;

    // Convert emotions map
    QVariantMap emotionsMap;
    for (auto it = result.emotions.begin(); it != result.emotions.end(); ++it) {
        emotionsMap[it.key()] = it.value();
    }
    resultMap["emotions"] = emotionsMap;

    // 添加到扁平化列表
    m_emotionResults.append(resultMap);

    // 按用户ID存储（每个用户只保留最新结果）
    m_emotionResultsByUser[result.userId] = resultMap;

    // Keep only last 10 results in flat list
    if (m_emotionResults.size() > 10) {
        m_emotionResults.removeFirst();
    }

    emit emotionResultsChanged();

    LOG_INFO(QString("Emotion result added for user %1: %2 (confidence: %3)")
             .arg(result.userId)
             .arg(result.emotion)
             .arg(result.confidence));
}

QVariantMap AIPanelController::getDetectionResultForUser(int userId) const
{
    return m_detectionResultsByUser.value(userId, QVariantMap());
}

QVariantMap AIPanelController::getEmotionResultForUser(int userId) const
{
    return m_emotionResultsByUser.value(userId, QVariantMap());
}

QVariantList AIPanelController::getAsrResultsForUser(int userId) const
{
    return m_asrResultsByUser.value(userId, QVariantList());
}

QString AIPanelController::getUsernameById(int userId) const
{
    // 从 MeetingService 获取用户名
    MeetingService *meetingService = Application::instance()->meetingService();
    if (!meetingService) {
        return QString("用户%1").arg(userId);
    }

    // 获取参与者列表
    const auto &participants = meetingService->participants();
    for (const auto &participant : participants) {
        if (participant && participant->userId() == userId) {
            return participant->username();
        }
    }

    return QString("用户%1").arg(userId);
}

