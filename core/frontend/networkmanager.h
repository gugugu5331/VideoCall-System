#ifndef NETWORKMANAGER_H
#define NETWORKMANAGER_H

#include <QObject>
#include <QNetworkAccessManager>
#include <QNetworkReply>
#include <QWebSocket>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QTimer>
#include <QMutex>
#include <QQueue>

class NetworkManager : public QObject
{
    Q_OBJECT

public:
    explicit NetworkManager(QObject *parent = nullptr);
    ~NetworkManager();

    // 连接管理
    void connectToServer(const QString &serverUrl, const QString &token);
    void disconnectFromServer();
    bool isConnected() const { return m_isConnected; }

    // API请求
    void login(const QString &username, const QString &password);
    void registerUser(const QString &username, const QString &email, const QString &password);
    void getUserProfile();
    void updateUserProfile(const QJsonObject &profile);
    void getCallHistory();
    void startCall(const QString &remoteUser);
    void endCall(const QString &callId);
    void getSecurityStatus(const QString &callId);

    // WebSocket消息
    void sendSignalingMessage(const QJsonObject &message);
    void sendHeartbeat();

signals:
    // 连接状态
    void connected();
    void disconnected();
    void connectionError(const QString &error);

    // 认证事件
    void loginSuccess(const QJsonObject &userInfo);
    void loginFailed(const QString &error);
    void registerSuccess(const QJsonObject &userInfo);
    void registerFailed(const QString &error);

    // 通话事件
    void incomingCall(const QString &callId, const QString &caller);
    void callAccepted(const QString &callId);
    void callRejected(const QString &callId);
    void callEnded(const QString &callId);

    // 安全事件
    void securityAlert(const QString &callId, const QString &alertType, double riskScore);

    // 数据更新
    void userProfileUpdated(const QJsonObject &profile);
    void callHistoryUpdated(const QJsonArray &history);
    void securityStatusUpdated(const QString &callId, const QJsonObject &status);

private slots:
    void onWebSocketConnected();
    void onWebSocketDisconnected();
    void onWebSocketError(QAbstractSocket::SocketError error);
    void onWebSocketMessageReceived(const QString &message);
    void onNetworkReplyFinished();
    void onHeartbeatTimer();

private:
    void setupWebSocket();
    void setupNetworkManager();
    void handleApiResponse(QNetworkReply *reply);
    void handleWebSocketMessage(const QJsonObject &message);
    void sendApiRequest(const QString &endpoint, const QJsonObject &data, const QString &method = "POST");
    QString getAuthHeader() const;

private:
    QNetworkAccessManager *m_networkManager;
    QWebSocket *m_webSocket;
    QTimer *m_heartbeatTimer;
    QMutex m_mutex;

    QString m_serverUrl;
    QString m_authToken;
    bool m_isConnected;
    QQueue<QNetworkReply*> m_pendingRequests;
};

#endif // NETWORKMANAGER_H 