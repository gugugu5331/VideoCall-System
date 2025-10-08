#include "models/participant.h"

Participant::Participant(QObject *parent)
    : QObject(parent)
    , m_userId(0)
    , m_audioEnabled(true)
    , m_videoEnabled(true)
    , m_isScreenSharing(false)
    , m_isSelf(false)
{
}

Participant::~Participant()
{
}

void Participant::setUserId(int id)
{
    if (m_userId != id) {
        m_userId = id;
        emit userIdChanged();
    }
}

void Participant::setUsername(const QString &username)
{
    if (m_username != username) {
        m_username = username;
        emit usernameChanged();
    }
}

void Participant::setPeerId(const QString &peerId)
{
    if (m_peerId != peerId) {
        m_peerId = peerId;
        emit peerIdChanged();
    }
}

void Participant::setSessionId(const QString &sessionId)
{
    m_sessionId = sessionId;
}

void Participant::setRole(const QString &role)
{
    if (m_role != role) {
        m_role = role;
        emit roleChanged();
    }
}

void Participant::setStatus(const QString &status)
{
    if (m_status != status) {
        m_status = status;
        emit statusChanged();
    }
}

void Participant::setAudioEnabled(bool enabled)
{
    if (m_audioEnabled != enabled) {
        m_audioEnabled = enabled;
        emit audioEnabledChanged();
    }
}

void Participant::setVideoEnabled(bool enabled)
{
    if (m_videoEnabled != enabled) {
        m_videoEnabled = enabled;
        emit videoEnabledChanged();
    }
}

void Participant::setIsScreenSharing(bool sharing)
{
    if (m_isScreenSharing != sharing) {
        m_isScreenSharing = sharing;
        emit isScreenSharingChanged();
    }
}

void Participant::setIsSelf(bool isSelf)
{
    if (m_isSelf != isSelf) {
        m_isSelf = isSelf;
        emit isSelfChanged();
    }
}

void Participant::setJoinedAt(const QDateTime &time)
{
    m_joinedAt = time;
}

QJsonObject Participant::toJson() const
{
    QJsonObject obj;
    obj["user_id"] = m_userId;
    obj["username"] = m_username;
    obj["peer_id"] = m_peerId;
    obj["session_id"] = m_sessionId;
    obj["role"] = m_role;
    obj["status"] = m_status;
    obj["audio_enabled"] = m_audioEnabled;
    obj["video_enabled"] = m_videoEnabled;
    obj["is_screen_sharing"] = m_isScreenSharing;
    obj["is_self"] = m_isSelf;
    obj["joined_at"] = m_joinedAt.toString(Qt::ISODate);
    return obj;
}

void Participant::fromJson(const QJsonObject &json)
{
    setUserId(json["user_id"].toInt());
    setUsername(json["username"].toString());
    setPeerId(json["peer_id"].toString());
    setSessionId(json["session_id"].toString());
    setRole(json["role"].toString("participant"));
    setStatus(json["status"].toString("active"));
    setAudioEnabled(json["audio_enabled"].toBool(true));
    setVideoEnabled(json["video_enabled"].toBool(true));
    setIsScreenSharing(json["is_screen_sharing"].toBool(false));
    setIsSelf(json["is_self"].toBool(false));

    QString joinedAtStr = json["joined_at"].toString();
    if (!joinedAtStr.isEmpty()) {
        setJoinedAt(QDateTime::fromString(joinedAtStr, Qt::ISODate));
    }
}

