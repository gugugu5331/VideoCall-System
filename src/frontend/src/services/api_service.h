#ifndef API_SERVICE_H
#define API_SERVICE_H

#include <QObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QJsonObject>
#include <QJsonDocument>
#include <QUrl>
#include <QTimer>
#include <functional>

/**
 * @brief HTTP方法枚举
 */
enum class HttpMethod {
    GET,
    POST,
    PUT,
    DELETE,
    PATCH
};

/**
 * @brief API响应结构
 */
struct ApiResponse {
    bool success;
    int statusCode;
    QJsonObject data;
    QString error;
    
    ApiResponse() : success(false), statusCode(0) {}
    ApiResponse(bool s, int code, const QJsonObject& d = {}, const QString& e = "")
        : success(s), statusCode(code), data(d), error(e) {}
};

/**
 * @brief API服务类
 * 
 * 负责与后端API的通信，包括HTTP请求、认证、错误处理等
 */
class ApiService : public QObject
{
    Q_OBJECT
    Q_PROPERTY(QString baseUrl READ baseUrl WRITE setBaseUrl NOTIFY baseUrlChanged)
    Q_PROPERTY(QString authToken READ authToken WRITE setAuthToken NOTIFY authTokenChanged)
    Q_PROPERTY(bool isOnline READ isOnline NOTIFY onlineStatusChanged)

public:
    explicit ApiService(QObject *parent = nullptr);
    ~ApiService();

    // 属性访问器
    QString baseUrl() const { return m_baseUrl; }
    void setBaseUrl(const QString &url);

    QString authToken() const { return m_authToken; }
    void setAuthToken(const QString &token);

    bool isOnline() const { return m_isOnline; }

    /**
     * @brief 发送HTTP请求
     * @param method HTTP方法
     * @param endpoint API端点
     * @param data 请求数据
     * @param callback 回调函数
     */
    void request(HttpMethod method, 
                const QString &endpoint, 
                const QJsonObject &data = {},
                std::function<void(const ApiResponse&)> callback = nullptr);

    /**
     * @brief 上传文件
     * @param endpoint API端点
     * @param filePath 文件路径
     * @param fieldName 字段名
     * @param additionalData 附加数据
     * @param callback 回调函数
     */
    void uploadFile(const QString &endpoint,
                   const QString &filePath,
                   const QString &fieldName = "file",
                   const QJsonObject &additionalData = {},
                   std::function<void(const ApiResponse&)> callback = nullptr);

    /**
     * @brief 下载文件
     * @param url 文件URL
     * @param savePath 保存路径
     * @param callback 回调函数
     */
    void downloadFile(const QString &url,
                     const QString &savePath,
                     std::function<void(bool, const QString&)> callback = nullptr);

    // 便捷方法
    void get(const QString &endpoint, std::function<void(const ApiResponse&)> callback = nullptr);
    void post(const QString &endpoint, const QJsonObject &data, std::function<void(const ApiResponse&)> callback = nullptr);
    void put(const QString &endpoint, const QJsonObject &data, std::function<void(const ApiResponse&)> callback = nullptr);
    void del(const QString &endpoint, std::function<void(const ApiResponse&)> callback = nullptr);

public slots:
    /**
     * @brief 取消所有请求
     */
    void cancelAllRequests();

    /**
     * @brief 检查网络连接状态
     */
    void checkNetworkStatus();

signals:
    void baseUrlChanged();
    void authTokenChanged();
    void onlineStatusChanged();
    void requestStarted();
    void requestFinished();
    void networkError(const QString &error);

private slots:
    void onNetworkReplyFinished();
    void onNetworkError(QNetworkReply::NetworkError error);
    void onSslErrors(const QList<QSslError> &errors);
    void onNetworkStatusChanged();

private:
    /**
     * @brief 创建网络请求
     */
    QNetworkRequest createRequest(const QString &endpoint);

    /**
     * @brief 处理网络响应
     */
    ApiResponse processResponse(QNetworkReply *reply);

    /**
     * @brief 解析JSON响应
     */
    QJsonObject parseJsonResponse(const QByteArray &data);

    /**
     * @brief 处理HTTP错误
     */
    QString getHttpErrorString(int statusCode);

    /**
     * @brief 重试请求
     */
    void retryRequest(QNetworkReply *reply);

private:
    QNetworkAccessManager *m_networkManager;
    QString m_baseUrl;
    QString m_authToken;
    bool m_isOnline;
    
    // 请求管理
    QList<QNetworkReply*> m_activeRequests;
    QMap<QNetworkReply*, std::function<void(const ApiResponse&)>> m_callbacks;
    QMap<QNetworkReply*, int> m_retryCount;
    
    // 网络状态检查
    QTimer *m_networkStatusTimer;
    
    // 配置
    int m_maxRetries;
    int m_timeoutMs;
    
    static const int DEFAULT_TIMEOUT = 30000; // 30秒
    static const int MAX_RETRIES = 3;
};

#endif // API_SERVICE_H
