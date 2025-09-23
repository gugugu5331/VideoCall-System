#pragma once

#include "core/common.h"

namespace VideoCallSystem {

struct DetectionRequest {
    QString detectionId;
    DetectionType type;
    QString callId;
    QByteArray audioData;
    QByteArray videoData;
    QVariantMap metadata;
    QDateTime timestamp;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["detection_id"] = detectionId;
        obj["detection_type"] = detectionTypeToString(type);
        obj["call_id"] = callId;
        obj["audio_data"] = QString::fromLatin1(audioData.toBase64());
        obj["video_data"] = QString::fromLatin1(videoData.toBase64());
        obj["timestamp"] = timestamp.toString(Qt::ISODate);
        
        QJsonObject metaObj;
        for (auto it = metadata.begin(); it != metadata.end(); ++it) {
            metaObj[it.key()] = QJsonValue::fromVariant(it.value());
        }
        obj["metadata"] = metaObj;
        
        return obj;
    }
    
    static DetectionRequest fromJson(const QJsonObject& obj) {
        DetectionRequest request;
        request.detectionId = obj["detection_id"].toString();
        request.type = stringToDetectionType(obj["detection_type"].toString());
        request.callId = obj["call_id"].toString();
        request.audioData = QByteArray::fromBase64(obj["audio_data"].toString().toLatin1());
        request.videoData = QByteArray::fromBase64(obj["video_data"].toString().toLatin1());
        request.timestamp = QDateTime::fromString(obj["timestamp"].toString(), Qt::ISODate);
        
        QJsonObject metaObj = obj["metadata"].toObject();
        for (auto it = metaObj.begin(); it != metaObj.end(); ++it) {
            request.metadata[it.key()] = it.value().toVariant();
        }
        
        return request;
    }

private:
    static QString detectionTypeToString(DetectionType type) {
        switch (type) {
            case DetectionType::FaceSwap: return "face_swap";
            case DetectionType::VoiceSynthesis: return "voice_synthesis";
            case DetectionType::ContentAnalysis: return "content_analysis";
            default: return "unknown";
        }
    }
    
    static DetectionType stringToDetectionType(const QString& str) {
        if (str == "face_swap") return DetectionType::FaceSwap;
        if (str == "voice_synthesis") return DetectionType::VoiceSynthesis;
        if (str == "content_analysis") return DetectionType::ContentAnalysis;
        return DetectionType::None;
    }
};

struct ModelInfo {
    QString name;
    QString version;
    QString type;
    QString description;
    bool isLoaded = false;
    QDateTime lastUpdated;
    QVariantMap parameters;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["name"] = name;
        obj["version"] = version;
        obj["type"] = type;
        obj["description"] = description;
        obj["is_loaded"] = isLoaded;
        obj["last_updated"] = lastUpdated.toString(Qt::ISODate);
        
        QJsonObject paramObj;
        for (auto it = parameters.begin(); it != parameters.end(); ++it) {
            paramObj[it.key()] = QJsonValue::fromVariant(it.value());
        }
        obj["parameters"] = paramObj;
        
        return obj;
    }
    
    static ModelInfo fromJson(const QJsonObject& obj) {
        ModelInfo info;
        info.name = obj["name"].toString();
        info.version = obj["version"].toString();
        info.type = obj["type"].toString();
        info.description = obj["description"].toString();
        info.isLoaded = obj["is_loaded"].toBool();
        info.lastUpdated = QDateTime::fromString(obj["last_updated"].toString(), Qt::ISODate);
        
        QJsonObject paramObj = obj["parameters"].toObject();
        for (auto it = paramObj.begin(); it != paramObj.end(); ++it) {
            info.parameters[it.key()] = it.value().toVariant();
        }
        
        return info;
    }
};

class AIDetectionClient : public QObject
{
    Q_OBJECT

public:
    explicit AIDetectionClient(QObject* parent = nullptr);
    ~AIDetectionClient();

    // 初始化和连接
    bool initialize(const QString& serverUrl);
    void cleanup();
    bool isInitialized() const { return initialized_; }

    // 连接管理
    void connectToServer();
    void disconnectFromServer();
    bool isConnected() const { return connected_; }
    QString serverUrl() const { return serverUrl_; }

    // 检测请求
    QString submitDetection(const DetectionRequest& request);
    void cancelDetection(const QString& detectionId);
    DetectionResult getDetectionResult(const QString& detectionId);
    QList<DetectionResult> getDetectionHistory(int limit = 100);

    // 异步检测
    void submitDetectionAsync(const DetectionRequest& request);
    void getDetectionResultAsync(const QString& detectionId);

    // 批量检测
    QStringList submitBatchDetection(const QList<DetectionRequest>& requests);
    QList<DetectionResult> getBatchDetectionResults(const QStringList& detectionIds);

    // 实时检测
    void startRealtimeDetection(DetectionType type, const QString& callId);
    void stopRealtimeDetection(const QString& callId);
    void sendRealtimeData(const QString& callId, const QByteArray& audioData, const QByteArray& videoData);

    // 模型管理
    QList<ModelInfo> getAvailableModels();
    bool loadModel(const QString& modelName);
    bool unloadModel(const QString& modelName);
    ModelInfo getModelInfo(const QString& modelName);

    // 配置管理
    void setDetectionThreshold(float threshold);
    float getDetectionThreshold() const { return detectionThreshold_; }
    
    void setDetectionInterval(int intervalMs);
    int getDetectionInterval() const { return detectionInterval_; }
    
    void enableDetectionType(DetectionType type, bool enable);
    bool isDetectionTypeEnabled(DetectionType type) const;

    // 统计信息
    struct DetectionStats {
        int totalDetections = 0;
        int fakeDetections = 0;
        int realDetections = 0;
        double averageConfidence = 0.0;
        double averageProcessingTime = 0.0;
        QDateTime lastDetection;
    };
    
    DetectionStats getDetectionStats() const { return detectionStats_; }
    void resetDetectionStats();

    // 错误处理
    QString lastError() const { return lastError_; }
    void clearError() { lastError_.clear(); }

public slots:
    // 网络槽函数
    void onNetworkReplyFinished();
    void onNetworkError(QNetworkReply::NetworkError error);
    void onSslErrors(const QList<QSslError>& errors);

    // 定时器槽函数
    void onRealtimeDetectionTimer();
    void onStatsUpdateTimer();

signals:
    // 连接状态信号
    void connected();
    void disconnected();
    void connectionError(const QString& error);

    // 检测结果信号
    void detectionCompleted(const QString& detectionId, const DetectionResult& result);
    void detectionFailed(const QString& detectionId, const QString& error);
    void detectionProgress(const QString& detectionId, int progress);

    // 实时检测信号
    void realtimeDetectionStarted(const QString& callId);
    void realtimeDetectionStopped(const QString& callId);
    void realtimeDetectionResult(const QString& callId, const DetectionResult& result);

    // 模型管理信号
    void modelLoaded(const QString& modelName);
    void modelUnloaded(const QString& modelName);
    void modelError(const QString& modelName, const QString& error);

    // 统计信号
    void statsUpdated(const DetectionStats& stats);

    // 警报信号
    void detectionAlert(const DetectionResult& result);
    void highRiskDetected(const DetectionResult& result);

private slots:
    // 内部处理槽
    void processDetectionQueue();
    void updateDetectionStats();
    void checkServerHealth();

private:
    // 网络请求处理
    QNetworkReply* sendRequest(const QString& endpoint, const QJsonObject& data, const QString& method = "POST");
    QJsonObject parseResponse(QNetworkReply* reply);
    void handleApiError(const QJsonObject& response);

    // 检测处理
    void processDetectionRequest(const DetectionRequest& request);
    void processDetectionResponse(const QString& detectionId, const QJsonObject& response);
    void updateDetectionResult(const QString& detectionId, const DetectionResult& result);

    // 实时检测处理
    void processRealtimeData(const QString& callId);
    void sendRealtimeDetectionRequest(const QString& callId, const QByteArray& data);

    // 缓存管理
    void cacheDetectionResult(const DetectionResult& result);
    DetectionResult getCachedDetectionResult(const QString& detectionId);
    void clearDetectionCache();

    // 配置管理
    void loadConfiguration();
    void saveConfiguration();

    // 错误处理
    void setError(const QString& error);
    void handleNetworkError(const QString& error);

private:
    // 初始化状态
    bool initialized_;
    bool connected_;
    
    // 服务器配置
    QString serverUrl_;
    QNetworkAccessManager* networkManager_;
    
    // 检测配置
    float detectionThreshold_;
    int detectionInterval_;
    QMap<DetectionType, bool> enabledDetectionTypes_;
    
    // 检测队列和缓存
    QQueue<DetectionRequest> detectionQueue_;
    QMap<QString, DetectionResult> detectionCache_;
    QMap<QString, QNetworkReply*> pendingRequests_;
    
    // 实时检测
    QMap<QString, QTimer*> realtimeTimers_;
    QMap<QString, QByteArray> realtimeAudioBuffers_;
    QMap<QString, QByteArray> realtimeVideoBuffers_;
    
    // 模型信息
    QList<ModelInfo> availableModels_;
    QStringList loadedModels_;
    
    // 统计信息
    DetectionStats detectionStats_;
    QTimer* statsTimer_;
    
    // 定时器
    QTimer* queueTimer_;
    QTimer* healthCheckTimer_;
    
    // 错误处理
    QString lastError_;
    
    // 互斥锁
    mutable QMutex mutex_;
    
    // 配置
    QSettings* settings_;
};

} // namespace VideoCallSystem
