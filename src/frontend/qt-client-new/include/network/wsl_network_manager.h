#ifndef WSL_NETWORK_MANAGER_H
#define WSL_NETWORK_MANAGER_H

#include <QObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QTimer>
#include <QJsonObject>
#include <QJsonDocument>
#include <QUrl>
#include <QProcess>
#include <QMutex>
#include <QCache>
#include <memory>

/**
 * @brief WSL网络管理器
 * 
 * 专门用于管理Windows Qt客户端与WSL后端服务的网络通信
 * 支持自动WSL IP检测、服务发现、连接管理等功能
 */
class WSLNetworkManager : public QObject
{
    Q_OBJECT

public:
    enum class ServiceType {
        UserService,
        MeetingService,
        SignalingService,
        MediaService,
        AIDetectionService,
        NotificationService,
        RecordService,
        SmartEditingService,
        Gateway
    };

    enum class ConnectionStatus {
        Disconnected,
        Connecting,
        Connected,
        Error
    };

    struct ServiceEndpoint {
        QString name;
        QString endpoint;
        QString healthCheck;
        bool isAvailable = false;
        qint64 lastCheck = 0;
        int responseTime = -1;
    };

    explicit WSLNetworkManager(QObject *parent = nullptr);
    ~WSLNetworkManager();

    // 初始化和配置
    bool initialize();
    bool loadConfiguration(const QString &configPath = "config/network_config.json");
    
    // WSL连接管理
    QString detectWSLIP();
    bool isWSLAvailable();
    void setWSLIP(const QString &ip);
    QString getWSLIP() const;
    
    // 服务管理
    QString getServiceUrl(ServiceType service) const;
    QString getServiceEndpoint(ServiceType service) const;
    bool isServiceAvailable(ServiceType service) const;
    void checkServiceHealth(ServiceType service);
    void checkAllServicesHealth();
    
    // 网络请求
    QNetworkReply* get(const QString &endpoint, const QJsonObject &params = QJsonObject());
    QNetworkReply* post(const QString &endpoint, const QJsonObject &data);
    QNetworkReply* put(const QString &endpoint, const QJsonObject &data);
    QNetworkReply* deleteResource(const QString &endpoint);
    
    // 文件上传
    QNetworkReply* uploadFile(const QString &endpoint, const QString &filePath, 
                             const QJsonObject &metadata = QJsonObject());
    
    // WebSocket连接
    QString getWebSocketUrl(ServiceType service = ServiceType::SignalingService) const;
    
    // 连接状态
    ConnectionStatus getConnectionStatus() const;
    bool isConnected() const;
    
    // 配置访问
    QJsonObject getConfiguration() const;
    void updateConfiguration(const QJsonObject &config);
    
    // 缓存管理
    void clearCache();
    void setCacheEnabled(bool enabled);
    
    // 错误处理
    QString getLastError() const;
    void clearLastError();

public slots:
    void connectToBackend();
    void disconnectFromBackend();
    void refreshConnection();
    void startHealthChecking();
    void stopHealthChecking();

signals:
    void connectionStatusChanged(ConnectionStatus status);
    void serviceAvailabilityChanged(ServiceType service, bool available);
    void wslIPDetected(const QString &ip);
    void healthCheckCompleted(ServiceType service, bool healthy, int responseTime);
    void networkError(const QString &error);
    void requestCompleted(QNetworkReply *reply);
    void uploadProgress(qint64 bytesSent, qint64 bytesTotal);

private slots:
    void onWSLIPDetectionFinished(int exitCode, QProcess::ExitStatus exitStatus);
    void onHealthCheckFinished();
    void onNetworkReplyFinished();
    void onHealthCheckTimer();

private:
    // 内部方法
    void setupNetworkManager();
    void setupHealthCheckTimer();
    QString buildUrl(const QString &endpoint) const;
    QNetworkRequest createRequest(const QString &url) const;
    void handleNetworkError(QNetworkReply *reply);
    void updateServiceStatus(ServiceType service, bool available, int responseTime = -1);
    QString serviceTypeToString(ServiceType service) const;
    ServiceType stringToServiceType(const QString &serviceStr) const;
    void cacheResponse(const QString &key, const QByteArray &data);
    QByteArray getCachedResponse(const QString &key) const;
    
    // 成员变量
    QNetworkAccessManager *m_networkManager;
    QTimer *m_healthCheckTimer;
    QProcess *m_wslProcess;
    QMutex m_mutex;
    QCache<QString, QByteArray> m_responseCache;
    
    // 配置和状态
    QJsonObject m_configuration;
    QString m_wslIP;
    QString m_baseUrl;
    QString m_apiBaseUrl;
    QString m_websocketUrl;
    ConnectionStatus m_connectionStatus;
    QString m_lastError;
    
    // 服务状态
    QHash<ServiceType, ServiceEndpoint> m_services;
    
    // 设置
    bool m_cacheEnabled;
    int m_healthCheckInterval;
    int m_requestTimeout;
    int m_retryCount;
    int m_retryDelay;
    
    // 常量
    static const int DEFAULT_HEALTH_CHECK_INTERVAL = 30000; // 30秒
    static const int DEFAULT_REQUEST_TIMEOUT = 30000;       // 30秒
    static const int DEFAULT_RETRY_COUNT = 3;
    static const int DEFAULT_RETRY_DELAY = 1000;            // 1秒
    static const int WSL_IP_CACHE_DURATION = 300000;        // 5分钟
};

// 便利宏定义
#define WSL_NETWORK() WSLNetworkManager::instance()

// 全局实例访问
class WSLNetworkManagerSingleton
{
public:
    static WSLNetworkManager* instance();
    static void cleanup();

private:
    static WSLNetworkManager* s_instance;
    static QMutex s_mutex;
};

#endif // WSL_NETWORK_MANAGER_H
