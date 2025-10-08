#ifndef WEBSOCKET_CLIENT_H
#define WEBSOCKET_CLIENT_H

#include <QObject>
#include <QWebSocket>
#include <QTimer>
#include <QJsonObject>
#include <functional>

enum class SignalingMessageType {
    Offer = 1,
    Answer = 2,
    IceCandidate = 3,
    JoinRoom = 4,
    LeaveRoom = 5,
    UserJoined = 6,
    UserLeft = 7,
    Chat = 8,
    ScreenShare = 9,
    MediaControl = 10,
    Ping = 11,
    Pong = 12,
    Error = 13,
    RoomInfo = 14
};

class WebSocketClient : public QObject
{
    Q_OBJECT

public:
    explicit WebSocketClient(QObject *parent = nullptr);
    ~WebSocketClient();

    // Connection management
    void connect(const QString &url, const QString &token, 
                int meetingId, int userId, const QString &peerId);
    void disconnect();
    bool isConnected() const;

    // Send messages
    void sendMessage(const QJsonObject &message);
    void sendSignalingMessage(SignalingMessageType type,
                             const QJsonObject &payload,
                             int toUserId = 0);

    // Convenience methods for common signaling messages
    void sendChatMessage(const QString &content, int toUserId = 0);
    void sendMediaControl(const QString &mediaType, bool enabled, int toUserId = 0);
    void sendScreenShareControl(bool enabled, int toUserId = 0);
    void sendOffer(const QString &sdp, int toUserId);
    void sendAnswer(const QString &sdp, int toUserId);
    void sendIceCandidate(const QString &candidate, const QString &sdpMid, int sdpMLineIndex, int toUserId);

    // Heartbeat
    void startHeartbeat(int intervalMs = 30000);
    void stopHeartbeat();

signals:
    void connected();
    void disconnected();
    void error(const QString &error);
    void messageReceived(const QJsonObject &message);
    void signalingMessageReceived(SignalingMessageType type, const QJsonObject &message);

private slots:
    void onConnected();
    void onDisconnected();
    void onError(QAbstractSocket::SocketError error);
    void onTextMessageReceived(const QString &message);
    void onHeartbeatTimeout();

private:
    void reconnect();
    QString generateMessageId() const;
    SignalingMessageType messageTypeFromInt(int type) const;

private:
    QWebSocket *m_socket;
    QTimer *m_heartbeatTimer;
    QTimer *m_reconnectTimer;
    
    QString m_url;
    QString m_token;
    int m_meetingId;
    int m_userId;
    QString m_peerId;
    QString m_sessionId;
    
    bool m_isConnected;
    int m_reconnectAttempts;
    int m_maxReconnectAttempts;
    qint64 m_lastPongTime;
};

#endif // WEBSOCKET_CLIENT_H

