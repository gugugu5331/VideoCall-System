#include "network/websocket_client.h"
#include "utils/logger.h"
#include <QJsonDocument>
#include <QUuid>

WebSocketClient::WebSocketClient(QObject *parent)
    : QObject(parent)
    , m_isConnected(false)
    , m_reconnectAttempts(0)
    , m_maxReconnectAttempts(5)
    , m_lastPongTime(0)
{
    m_socket = new QWebSocket();
    m_reconnectTimer = new QTimer(this);
    m_heartbeatTimer = new QTimer(this);

    QObject::connect(m_socket, &QWebSocket::connected, this, &WebSocketClient::onConnected);
    QObject::connect(m_socket, &QWebSocket::disconnected, this, &WebSocketClient::onDisconnected);
    QObject::connect(m_socket, &QWebSocket::textMessageReceived, this, &WebSocketClient::onTextMessageReceived);
    QObject::connect(m_socket, QOverload<QAbstractSocket::SocketError>::of(&QWebSocket::error),
            this, &WebSocketClient::onError);

    QObject::connect(m_reconnectTimer, &QTimer::timeout, this, &WebSocketClient::reconnect);
    QObject::connect(m_heartbeatTimer, &QTimer::timeout, this, &WebSocketClient::onHeartbeatTimeout);
}

WebSocketClient::~WebSocketClient()
{
    disconnect();
    m_socket->deleteLater();
}

void WebSocketClient::connect(const QString &url, const QString &token,
                             int meetingId, int userId, const QString &peerId)
{
    m_url = url;
    m_token = token;
    m_meetingId = meetingId;
    m_userId = userId;
    m_peerId = peerId;

    // 根据后端API文档，WebSocket连接需要：
    // 1. Authorization头携带Bearer token
    // 2. 查询参数包含user_id, meeting_id, peer_id
    // 将token添加到URL参数中（临时方案，后续可以改为使用子协议）
    QString fullUrl = QString("%1?user_id=%2&meeting_id=%3&peer_id=%4&token=%5")
                        .arg(url).arg(userId).arg(meetingId).arg(peerId).arg(token);

    LOG_INFO("Connecting to WebSocket: " + url);
    m_socket->open(QUrl(fullUrl));
}

void WebSocketClient::disconnect()
{
    m_reconnectTimer->stop();
    m_heartbeatTimer->stop();

    if (m_socket->state() == QAbstractSocket::ConnectedState) {
        LOG_INFO("Disconnecting from WebSocket");
        m_socket->close();
    }

    m_isConnected = false;
}

bool WebSocketClient::isConnected() const
{
    return m_isConnected && m_socket->state() == QAbstractSocket::ConnectedState;
}

void WebSocketClient::sendMessage(const QJsonObject &message)
{
    if (!isConnected()) {
        LOG_WARNING("WebSocket not connected, cannot send message");
        return;
    }

    QString jsonString = QJsonDocument(message).toJson(QJsonDocument::Compact);
    m_socket->sendTextMessage(jsonString);

    LOG_DEBUG("WebSocket sent: " + jsonString);
}

void WebSocketClient::sendSignalingMessage(SignalingMessageType type,
                                          const QJsonObject &payload,
                                          int toUserId)
{
    QJsonObject message;
    message["type"] = static_cast<int>(type);
    message["from_peer_id"] = m_peerId;
    message["meeting_id"] = m_meetingId;
    message["user_id"] = m_userId;
    message["payload"] = payload;
    message["timestamp"] = QDateTime::currentDateTime().toString(Qt::ISODate);

    if (toUserId > 0) {
        message["to_user_id"] = toUserId;
    }

    sendMessage(message);
}

void WebSocketClient::startHeartbeat(int intervalMs)
{
    m_heartbeatTimer->start(intervalMs);
}

void WebSocketClient::stopHeartbeat()
{
    m_heartbeatTimer->stop();
}

void WebSocketClient::onConnected()
{
    LOG_INFO("WebSocket connected");
    m_isConnected = true;
    m_reconnectAttempts = 0;
    m_reconnectTimer->stop();
    m_lastPongTime = QDateTime::currentMSecsSinceEpoch();

    // Start heartbeat (default 30 seconds)
    startHeartbeat(30000);

    emit connected();
}

void WebSocketClient::onDisconnected()
{
    LOG_WARNING("WebSocket disconnected");
    m_isConnected = false;
    m_heartbeatTimer->stop();

    emit disconnected();

    // Attempt reconnection
    if (m_reconnectAttempts < m_maxReconnectAttempts) {
        m_reconnectTimer->start(5000); // Retry after 5 seconds
    }
}

void WebSocketClient::onError(QAbstractSocket::SocketError socketError)
{
    QString errorString = m_socket->errorString();
    LOG_ERROR("WebSocket error: " + errorString);
    emit error(errorString);
}

void WebSocketClient::onTextMessageReceived(const QString &message)
{
    LOG_DEBUG("WebSocket received: " + message);

    QJsonDocument doc = QJsonDocument::fromJson(message.toUtf8());
    if (!doc.isObject()) {
        LOG_ERROR("Invalid WebSocket message format");
        return;
    }

    QJsonObject obj = doc.object();

    // Emit raw message
    emit messageReceived(obj);

    // Parse message type
    QString typeStr = obj["type"].toString();
    int typeInt = obj["type"].toInt();

    // Try to convert to SignalingMessageType
    SignalingMessageType type = messageTypeFromInt(typeInt);

    // Handle pong response
    if (type == SignalingMessageType::Pong) {
        m_lastPongTime = QDateTime::currentMSecsSinceEpoch();
        return; // Heartbeat acknowledged
    }

    // Emit signaling message
    emit signalingMessageReceived(type, obj);
}

void WebSocketClient::onHeartbeatTimeout()
{
    if (!isConnected()) {
        return;
    }

    // Send ping message
    QJsonObject payload;
    payload["timestamp"] = QDateTime::currentMSecsSinceEpoch();

    sendSignalingMessage(SignalingMessageType::Ping, payload);

    // Check if we received pong recently
    qint64 now = QDateTime::currentMSecsSinceEpoch();
    if (m_lastPongTime > 0 && (now - m_lastPongTime) > 60000) {
        // No pong received for 60 seconds, connection might be dead
        LOG_WARNING("No pong received for 60 seconds, reconnecting...");
        m_socket->close();
    }
}

void WebSocketClient::reconnect()
{
    if (m_reconnectAttempts >= m_maxReconnectAttempts) {
        LOG_ERROR("Max reconnection attempts reached");
        m_reconnectTimer->stop();
        emit error("Max reconnection attempts reached");
        return;
    }

    m_reconnectAttempts++;
    LOG_INFO(QString("Reconnection attempt %1/%2").arg(m_reconnectAttempts).arg(m_maxReconnectAttempts));

    connect(m_url, m_token, m_meetingId, m_userId, m_peerId);
}

QString WebSocketClient::generateMessageId() const
{
    return QUuid::createUuid().toString(QUuid::WithoutBraces);
}

SignalingMessageType WebSocketClient::messageTypeFromInt(int type) const
{
    switch (type) {
        case 1: return SignalingMessageType::Offer;
        case 2: return SignalingMessageType::Answer;
        case 3: return SignalingMessageType::IceCandidate;
        case 4: return SignalingMessageType::JoinRoom;
        case 5: return SignalingMessageType::LeaveRoom;
        case 6: return SignalingMessageType::UserJoined;
        case 7: return SignalingMessageType::UserLeft;
        case 8: return SignalingMessageType::Chat;
        case 9: return SignalingMessageType::ScreenShare;
        case 10: return SignalingMessageType::MediaControl;
        case 11: return SignalingMessageType::Ping;
        case 12: return SignalingMessageType::Pong;
        case 13: return SignalingMessageType::Error;
        case 14: return SignalingMessageType::RoomInfo;
        default: return SignalingMessageType::Error;
    }
}

void WebSocketClient::sendChatMessage(const QString &content, int toUserId)
{
    QJsonObject payload;
    payload["content"] = content;
    payload["timestamp"] = QDateTime::currentDateTime().toString(Qt::ISODate);

    sendSignalingMessage(SignalingMessageType::Chat, payload, toUserId);
}

void WebSocketClient::sendMediaControl(const QString &mediaType, bool enabled, int toUserId)
{
    QJsonObject payload;
    payload["media_type"] = mediaType;
    payload["enabled"] = enabled;

    sendSignalingMessage(SignalingMessageType::MediaControl, payload, toUserId);
}

void WebSocketClient::sendScreenShareControl(bool enabled, int toUserId)
{
    QJsonObject payload;
    payload["enabled"] = enabled;

    sendSignalingMessage(SignalingMessageType::ScreenShare, payload, toUserId);
}

void WebSocketClient::sendOffer(const QString &sdp, int toUserId)
{
    QJsonObject payload;
    payload["sdp"] = sdp;
    payload["type"] = "offer";

    sendSignalingMessage(SignalingMessageType::Offer, payload, toUserId);
}

void WebSocketClient::sendAnswer(const QString &sdp, int toUserId)
{
    QJsonObject payload;
    payload["sdp"] = sdp;
    payload["type"] = "answer";

    sendSignalingMessage(SignalingMessageType::Answer, payload, toUserId);
}

void WebSocketClient::sendIceCandidate(const QString &candidate, const QString &sdpMid, int sdpMLineIndex, int toUserId)
{
    QJsonObject payload;
    payload["candidate"] = candidate;
    payload["sdp_mid"] = sdpMid;
    payload["sdp_mline_index"] = sdpMLineIndex;

    sendSignalingMessage(SignalingMessageType::IceCandidate, payload, toUserId);
}

