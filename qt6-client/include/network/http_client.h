#ifndef HTTP_CLIENT_H
#define HTTP_CLIENT_H

#include <QObject>
#include <QNetworkAccessManager>
#include <QNetworkRequest>
#include <QNetworkReply>
#include <QJsonObject>
#include <QJsonDocument>
#include <functional>

class HttpClient : public QObject
{
    Q_OBJECT

public:
    explicit HttpClient(QObject *parent = nullptr);
    ~HttpClient();

    // HTTP methods
    void get(const QString &url, 
             std::function<void(const QJsonObject&)> onSuccess,
             std::function<void(const QString&)> onError = nullptr);

    void post(const QString &url, 
              const QJsonObject &data,
              std::function<void(const QJsonObject&)> onSuccess,
              std::function<void(const QString&)> onError = nullptr);

    void put(const QString &url, 
             const QJsonObject &data,
             std::function<void(const QJsonObject&)> onSuccess,
             std::function<void(const QString&)> onError = nullptr);

    void del(const QString &url,
             std::function<void(const QJsonObject&)> onSuccess,
             std::function<void(const QString&)> onError = nullptr);

    // Upload file
    void upload(const QString &url,
                const QString &filePath,
                const QVariantMap &formData,
                std::function<void(const QJsonObject&)> onSuccess,
                std::function<void(const QString&)> onError = nullptr,
                std::function<void(qint64, qint64)> onProgress = nullptr);

    // Set authorization token
    void setAuthToken(const QString &token);

    // Set CSRF token
    void setCsrfToken(const QString &token);

    // Set timeout
    void setTimeout(int milliseconds);

signals:
    void requestStarted(const QString &url);
    void requestFinished(const QString &url);
    void requestError(const QString &url, const QString &error);

private:
    void sendRequest(const QString &method,
                    const QString &url,
                    const QJsonObject &data,
                    std::function<void(const QJsonObject&)> onSuccess,
                    std::function<void(const QString&)> onError);

    QNetworkRequest createRequest(const QString &url);
    void handleReply(QNetworkReply *reply,
                    std::function<void(const QJsonObject&)> onSuccess,
                    std::function<void(const QString&)> onError);

private:
    QNetworkAccessManager *m_manager;
    QString m_authToken;
    QString m_csrfToken;
    int m_timeout;
};

#endif // HTTP_CLIENT_H

