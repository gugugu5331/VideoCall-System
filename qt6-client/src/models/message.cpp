#include "models/message.h"

Message::Message(QObject *parent)
    : QObject(parent)
    , m_fromUserId(0)
    , m_type(MessageType::Text)
{
    m_timestamp = QDateTime::currentDateTime();
}

Message::~Message()
{
}

void Message::setMessageId(const QString &id)
{
    if (m_messageId != id) {
        m_messageId = id;
        emit messageIdChanged();
    }
}

void Message::setFromUserId(int userId)
{
    if (m_fromUserId != userId) {
        m_fromUserId = userId;
        emit fromUserIdChanged();
    }
}

void Message::setFromUsername(const QString &username)
{
    if (m_fromUsername != username) {
        m_fromUsername = username;
        emit fromUsernameChanged();
    }
}

void Message::setContent(const QString &content)
{
    if (m_content != content) {
        m_content = content;
        emit contentChanged();
    }
}

void Message::setType(MessageType type)
{
    m_type = type;
}

void Message::setTimestamp(const QDateTime &timestamp)
{
    if (m_timestamp != timestamp) {
        m_timestamp = timestamp;
        emit timestampChanged();
    }
}

QJsonObject Message::toJson() const
{
    QJsonObject obj;
    obj["message_id"] = m_messageId;
    obj["from_user_id"] = m_fromUserId;
    obj["from_username"] = m_fromUsername;
    obj["content"] = m_content;
    obj["type"] = static_cast<int>(m_type);
    obj["timestamp"] = m_timestamp.toString(Qt::ISODate);
    return obj;
}

void Message::fromJson(const QJsonObject &json)
{
    setMessageId(json["message_id"].toString());
    setFromUserId(json["from_user_id"].toInt());
    setFromUsername(json["from_username"].toString());
    setContent(json["content"].toString());
    setType(static_cast<MessageType>(json["type"].toInt(0)));
    
    QString timestampStr = json["timestamp"].toString();
    if (!timestampStr.isEmpty()) {
        setTimestamp(QDateTime::fromString(timestampStr, Qt::ISODate));
    }
}

