#include "network/http_client.h"
#include "utils/logger.h"
#include <QTimer>
#include <QHttpMultiPart>
#include <QFile>
#include <QFileInfo>

HttpClient::HttpClient(QObject *parent)
    : QObject(parent)
    , m_timeout(30000)
{
    m_manager = new QNetworkAccessManager(this);
}

HttpClient::~HttpClient()
{
}

void HttpClient::get(const QString &url,
                    std::function<void(const QJsonObject&)> onSuccess,
                    std::function<void(const QString&)> onError)
{
    sendRequest("GET", url, QJsonObject(), onSuccess, onError);
}

void HttpClient::post(const QString &url,
                     const QJsonObject &data,
                     std::function<void(const QJsonObject&)> onSuccess,
                     std::function<void(const QString&)> onError)
{
    sendRequest("POST", url, data, onSuccess, onError);
}

void HttpClient::put(const QString &url,
                    const QJsonObject &data,
                    std::function<void(const QJsonObject&)> onSuccess,
                    std::function<void(const QString&)> onError)
{
    sendRequest("PUT", url, data, onSuccess, onError);
}

void HttpClient::del(const QString &url,
                    std::function<void(const QJsonObject&)> onSuccess,
                    std::function<void(const QString&)> onError)
{
    sendRequest("DELETE", url, QJsonObject(), onSuccess, onError);
}

void HttpClient::upload(const QString &url,
                       const QString &filePath,
                       const QVariantMap &formData,
                       std::function<void(const QJsonObject&)> onSuccess,
                       std::function<void(const QString&)> onError,
                       std::function<void(qint64, qint64)> onProgress)
{
    QFile *file = new QFile(filePath);
    if (!file->open(QIODevice::ReadOnly)) {
        if (onError) {
            onError("Failed to open file: " + filePath);
        }
        delete file;
        return;
    }

    QHttpMultiPart *multiPart = new QHttpMultiPart(QHttpMultiPart::FormDataType);

    // Add file part
    QHttpPart filePart;
    filePart.setHeader(QNetworkRequest::ContentTypeHeader, QVariant("application/octet-stream"));
    filePart.setHeader(QNetworkRequest::ContentDispositionHeader,
                      QVariant("form-data; name=\"file\"; filename=\"" + QFileInfo(filePath).fileName() + "\""));
    filePart.setBodyDevice(file);
    file->setParent(multiPart);
    multiPart->append(filePart);

    // Add form data
    for (auto it = formData.begin(); it != formData.end(); ++it) {
        QHttpPart textPart;
        textPart.setHeader(QNetworkRequest::ContentDispositionHeader,
                          QVariant("form-data; name=\"" + it.key() + "\""));
        textPart.setBody(it.value().toString().toUtf8());
        multiPart->append(textPart);
    }

    QNetworkRequest request = createRequest(url);
    QNetworkReply *reply = m_manager->post(request, multiPart);
    multiPart->setParent(reply);

    // Progress tracking
    if (onProgress) {
        connect(reply, &QNetworkReply::uploadProgress, this, [onProgress](qint64 sent, qint64 total) {
            onProgress(sent, total);
        });
    }

    handleReply(reply, onSuccess, onError);
}

void HttpClient::setAuthToken(const QString &token)
{
    m_authToken = token;
}

void HttpClient::setCsrfToken(const QString &token)
{
    m_csrfToken = token;
}

void HttpClient::setTimeout(int milliseconds)
{
    m_timeout = milliseconds;
}

void HttpClient::sendRequest(const QString &method,
                            const QString &url,
                            const QJsonObject &data,
                            std::function<void(const QJsonObject&)> onSuccess,
                            std::function<void(const QString&)> onError)
{
    emit requestStarted(url);
    LOG_DEBUG("HTTP " + method + " " + url);

    QNetworkRequest request = createRequest(url);
    QNetworkReply *reply = nullptr;

    if (method == "GET") {
        reply = m_manager->get(request);
    } else if (method == "POST") {
        request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
        QByteArray jsonData = QJsonDocument(data).toJson();
        reply = m_manager->post(request, jsonData);
    } else if (method == "PUT") {
        request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
        QByteArray jsonData = QJsonDocument(data).toJson();
        reply = m_manager->put(request, jsonData);
    } else if (method == "DELETE") {
        reply = m_manager->deleteResource(request);
    }

    if (reply) {
        handleReply(reply, onSuccess, onError);
    }
}

QNetworkRequest HttpClient::createRequest(const QString &url)
{
    QUrl qurl(url);
    QNetworkRequest request(qurl);

    // Set headers
    request.setHeader(QNetworkRequest::UserAgentHeader, "MeetingSystemClient/1.0");

    // Set auth token if available
    if (!m_authToken.isEmpty()) {
        request.setRawHeader("Authorization", ("Bearer " + m_authToken).toUtf8());
    }

    // Set CSRF token if available
    if (!m_csrfToken.isEmpty()) {
        request.setRawHeader("X-CSRF-Token", m_csrfToken.toUtf8());
    }

    // Set timeout
    request.setTransferTimeout(m_timeout);

    return request;
}

void HttpClient::handleReply(QNetworkReply *reply,
                            std::function<void(const QJsonObject&)> onSuccess,
                            std::function<void(const QString&)> onError)
{
    connect(reply, &QNetworkReply::finished, this, [this, reply, onSuccess, onError]() {
        QString url = reply->url().toString();
        emit requestFinished(url);
        
        if (reply->error() == QNetworkReply::NoError) {
            QByteArray responseData = reply->readAll();
            QJsonDocument doc = QJsonDocument::fromJson(responseData);
            
            if (doc.isObject()) {
                QJsonObject response = doc.object();
                LOG_DEBUG("HTTP Response: " + QString::number(reply->attribute(QNetworkRequest::HttpStatusCodeAttribute).toInt()));
                
                if (onSuccess) {
                    onSuccess(response);
                }
            } else {
                QString error = "Invalid JSON response";
                LOG_ERROR(error);
                emit requestError(url, error);
                if (onError) {
                    onError(error);
                }
            }
        } else {
            QString error = reply->errorString();
            LOG_ERROR("HTTP Error: " + error);
            emit requestError(url, error);
            if (onError) {
                onError(error);
            }
        }
        
        reply->deleteLater();
    });
}

