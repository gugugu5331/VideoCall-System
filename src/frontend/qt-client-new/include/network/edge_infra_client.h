#pragma once

#include "core/common.h"

#ifdef HAS_ZMQ
#include <zmq.h>
#endif

namespace VideoCallSystem {

struct InferenceRequest {
    QString requestId;
    QString modelName;
    QString taskType;
    QByteArray inputData;
    QVariantMap parameters;
    int priority = 0;
    int timeout = 30000; // 30秒
    QDateTime timestamp;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["request_id"] = requestId;
        obj["model_name"] = modelName;
        obj["task_type"] = taskType;
        obj["input_data"] = QString::fromLatin1(inputData.toBase64());
        obj["priority"] = priority;
        obj["timeout"] = timeout;
        obj["timestamp"] = timestamp.toString(Qt::ISODate);
        
        QJsonObject paramObj;
        for (auto it = parameters.begin(); it != parameters.end(); ++it) {
            paramObj[it.key()] = QJsonValue::fromVariant(it.value());
        }
        obj["parameters"] = paramObj;
        
        return obj;
    }
    
    static InferenceRequest fromJson(const QJsonObject& obj) {
        InferenceRequest request;
        request.requestId = obj["request_id"].toString();
        request.modelName = obj["model_name"].toString();
        request.taskType = obj["task_type"].toString();
        request.inputData = QByteArray::fromBase64(obj["input_data"].toString().toLatin1());
        request.priority = obj["priority"].toInt();
        request.timeout = obj["timeout"].toInt();
        request.timestamp = QDateTime::fromString(obj["timestamp"].toString(), Qt::ISODate);
        
        QJsonObject paramObj = obj["parameters"].toObject();
        for (auto it = paramObj.begin(); it != paramObj.end(); ++it) {
            request.parameters[it.key()] = it.value().toVariant();
        }
        
        return request;
    }
};

struct InferenceResult {
    QString requestId;
    QString modelName;
    QString taskType;
    QByteArray outputData;
    QVariantMap results;
    bool success = false;
    QString errorMessage;
    double processingTime = 0.0;
    QDateTime timestamp;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["request_id"] = requestId;
        obj["model_name"] = modelName;
        obj["task_type"] = taskType;
        obj["output_data"] = QString::fromLatin1(outputData.toBase64());
        obj["success"] = success;
        obj["error_message"] = errorMessage;
        obj["processing_time"] = processingTime;
        obj["timestamp"] = timestamp.toString(Qt::ISODate);
        
        QJsonObject resultObj;
        for (auto it = results.begin(); it != results.end(); ++it) {
            resultObj[it.key()] = QJsonValue::fromVariant(it.value());
        }
        obj["results"] = resultObj;
        
        return obj;
    }
    
    static InferenceResult fromJson(const QJsonObject& obj) {
        InferenceResult result;
        result.requestId = obj["request_id"].toString();
        result.modelName = obj["model_name"].toString();
        result.taskType = obj["task_type"].toString();
        result.outputData = QByteArray::fromBase64(obj["output_data"].toString().toLatin1());
        result.success = obj["success"].toBool();
        result.errorMessage = obj["error_message"].toString();
        result.processingTime = obj["processing_time"].toDouble();
        result.timestamp = QDateTime::fromString(obj["timestamp"].toString(), Qt::ISODate);
        
        QJsonObject resultObj = obj["results"].toObject();
        for (auto it = resultObj.begin(); it != resultObj.end(); ++it) {
            result.results[it.key()] = it.value().toVariant();
        }
        
        return result;
    }
};

struct ModelStatus {
    QString modelName;
    QString status; // "loaded", "loading", "unloaded", "error"
    QString version;
    QString framework; // "tensorflow", "pytorch", "onnx"
    QVariantMap metadata;
    double memoryUsage = 0.0; // MB
    double gpuUsage = 0.0; // %
    int activeRequests = 0;
    QDateTime lastUsed;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["model_name"] = modelName;
        obj["status"] = status;
        obj["version"] = version;
        obj["framework"] = framework;
        obj["memory_usage"] = memoryUsage;
        obj["gpu_usage"] = gpuUsage;
        obj["active_requests"] = activeRequests;
        obj["last_used"] = lastUsed.toString(Qt::ISODate);
        
        QJsonObject metaObj;
        for (auto it = metadata.begin(); it != metadata.end(); ++it) {
            metaObj[it.key()] = QJsonValue::fromVariant(it.value());
        }
        obj["metadata"] = metaObj;
        
        return obj;
    }
    
    static ModelStatus fromJson(const QJsonObject& obj) {
        ModelStatus status;
        status.modelName = obj["model_name"].toString();
        status.status = obj["status"].toString();
        status.version = obj["version"].toString();
        status.framework = obj["framework"].toString();
        status.memoryUsage = obj["memory_usage"].toDouble();
        status.gpuUsage = obj["gpu_usage"].toDouble();
        status.activeRequests = obj["active_requests"].toInt();
        status.lastUsed = QDateTime::fromString(obj["last_used"].toString(), Qt::ISODate);
        
        QJsonObject metaObj = obj["metadata"].toObject();
        for (auto it = metaObj.begin(); it != metaObj.end(); ++it) {
            status.metadata[it.key()] = it.value().toVariant();
        }
        
        return status;
    }
};

struct SystemMetrics {
    double cpuUsage = 0.0; // %
    double memoryUsage = 0.0; // %
    double gpuUsage = 0.0; // %
    double diskUsage = 0.0; // %
    double networkIn = 0.0; // MB/s
    double networkOut = 0.0; // MB/s
    int activeConnections = 0;
    int queuedRequests = 0;
    QDateTime timestamp;
    
    QJsonObject toJson() const {
        QJsonObject obj;
        obj["cpu_usage"] = cpuUsage;
        obj["memory_usage"] = memoryUsage;
        obj["gpu_usage"] = gpuUsage;
        obj["disk_usage"] = diskUsage;
        obj["network_in"] = networkIn;
        obj["network_out"] = networkOut;
        obj["active_connections"] = activeConnections;
        obj["queued_requests"] = queuedRequests;
        obj["timestamp"] = timestamp.toString(Qt::ISODate);
        return obj;
    }
    
    static SystemMetrics fromJson(const QJsonObject& obj) {
        SystemMetrics metrics;
        metrics.cpuUsage = obj["cpu_usage"].toDouble();
        metrics.memoryUsage = obj["memory_usage"].toDouble();
        metrics.gpuUsage = obj["gpu_usage"].toDouble();
        metrics.diskUsage = obj["disk_usage"].toDouble();
        metrics.networkIn = obj["network_in"].toDouble();
        metrics.networkOut = obj["network_out"].toDouble();
        metrics.activeConnections = obj["active_connections"].toInt();
        metrics.queuedRequests = obj["queued_requests"].toInt();
        metrics.timestamp = QDateTime::fromString(obj["timestamp"].toString(), Qt::ISODate);
        return metrics;
    }
};

class EdgeInfraClient : public QObject
{
    Q_OBJECT

public:
    explicit EdgeInfraClient(QObject* parent = nullptr);
    ~EdgeInfraClient();

    // 初始化和连接
    bool initialize(const QString& serverAddress, int port = 9000);
    void cleanup();
    bool isInitialized() const { return initialized_; }

    // 连接管理
    void connectToServer();
    void disconnectFromServer();
    bool isConnected() const { return connected_; }
    QString serverAddress() const { return serverAddress_; }

    // 推理请求
    QString submitInference(const InferenceRequest& request);
    void cancelInference(const QString& requestId);
    InferenceResult getInferenceResult(const QString& requestId);
    QList<InferenceResult> getInferenceHistory(int limit = 100);

    // 异步推理
    void submitInferenceAsync(const InferenceRequest& request);
    void getInferenceResultAsync(const QString& requestId);

    // 批量推理
    QStringList submitBatchInference(const QList<InferenceRequest>& requests);
    QList<InferenceResult> getBatchInferenceResults(const QStringList& requestIds);

    // 模型管理
    QList<ModelStatus> getAvailableModels();
    bool loadModel(const QString& modelName, const QVariantMap& config = QVariantMap{});
    bool unloadModel(const QString& modelName);
    ModelStatus getModelStatus(const QString& modelName);
    void refreshModelStatus();

    // 系统监控
    SystemMetrics getSystemMetrics();
    void startMetricsCollection(int intervalMs = 5000);
    void stopMetricsCollection();

    // 配置管理
    void setInferenceTimeout(int timeoutMs);
    int getInferenceTimeout() const { return inferenceTimeout_; }
    
    void setMaxConcurrentRequests(int maxRequests);
    int getMaxConcurrentRequests() const { return maxConcurrentRequests_; }
    
    void setPriority(int priority);
    int getPriority() const { return priority_; }

    // 统计信息
    struct InferenceStats {
        int totalRequests = 0;
        int successfulRequests = 0;
        int failedRequests = 0;
        double averageProcessingTime = 0.0;
        double averageQueueTime = 0.0;
        QDateTime lastRequest;
    };
    
    InferenceStats getInferenceStats() const { return inferenceStats_; }
    void resetInferenceStats();

    // 错误处理
    QString lastError() const { return lastError_; }
    void clearError() { lastError_.clear(); }

public slots:
    // ZMQ消息处理
    void onZmqMessageReceived();
    void onZmqError();

    // 定时器槽函数
    void onMetricsTimer();
    void onHeartbeatTimer();

signals:
    // 连接状态信号
    void connected();
    void disconnected();
    void connectionError(const QString& error);

    // 推理结果信号
    void inferenceCompleted(const QString& requestId, const InferenceResult& result);
    void inferenceFailed(const QString& requestId, const QString& error);
    void inferenceProgress(const QString& requestId, int progress);

    // 模型管理信号
    void modelLoaded(const QString& modelName);
    void modelUnloaded(const QString& modelName);
    void modelError(const QString& modelName, const QString& error);
    void modelStatusChanged(const QString& modelName, const ModelStatus& status);

    // 系统监控信号
    void systemMetricsUpdated(const SystemMetrics& metrics);
    void systemOverloaded(const SystemMetrics& metrics);

    // 统计信号
    void inferenceStatsUpdated(const InferenceStats& stats);

private slots:
    // 内部处理槽
    void processInferenceQueue();
    void updateInferenceStats();
    void sendHeartbeat();

private:
    // ZMQ通信处理
    bool initializeZMQ();
    void cleanupZMQ();
    bool sendZmqMessage(const QJsonObject& message);
    QJsonObject receiveZmqMessage();
    void handleZmqMessage(const QJsonObject& message);

    // 推理处理
    void processInferenceRequest(const InferenceRequest& request);
    void processInferenceResponse(const QString& requestId, const QJsonObject& response);
    void updateInferenceResult(const QString& requestId, const InferenceResult& result);

    // 模型管理处理
    void processModelStatusUpdate(const QJsonObject& update);
    void refreshModelList();

    // 系统监控处理
    void processSystemMetrics(const QJsonObject& metrics);

    // 缓存管理
    void cacheInferenceResult(const InferenceResult& result);
    InferenceResult getCachedInferenceResult(const QString& requestId);
    void clearInferenceCache();

    // 错误处理
    void setError(const QString& error);
    void handleZmqError(const QString& error);

private:
    // 初始化状态
    bool initialized_;
    bool connected_;
    
    // 服务器配置
    QString serverAddress_;
    int serverPort_;
    
#ifdef HAS_ZMQ
    // ZMQ上下文和套接字
    void* zmqContext_;
    void* zmqSocket_;
#endif
    
    // 推理配置
    int inferenceTimeout_;
    int maxConcurrentRequests_;
    int priority_;
    
    // 推理队列和缓存
    QQueue<InferenceRequest> inferenceQueue_;
    QMap<QString, InferenceResult> inferenceCache_;
    QMap<QString, QDateTime> pendingRequests_;
    
    // 模型状态
    QList<ModelStatus> availableModels_;
    QMap<QString, ModelStatus> modelStatusCache_;
    
    // 系统监控
    SystemMetrics currentMetrics_;
    QTimer* metricsTimer_;
    
    // 统计信息
    InferenceStats inferenceStats_;
    
    // 定时器
    QTimer* queueTimer_;
    QTimer* heartbeatTimer_;
    
    // 错误处理
    QString lastError_;
    
    // 互斥锁
    mutable QMutex mutex_;
    
    // 配置
    QSettings* settings_;
};

} // namespace VideoCallSystem
