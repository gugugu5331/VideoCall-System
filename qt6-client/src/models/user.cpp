#include "models/user.h"

User::User(QObject *parent)
    : QObject(parent)
    , m_userId(0)
    , m_status("offline")
{
}

User::~User()
{
}

void User::setUserId(int id)
{
    if (m_userId != id) {
        m_userId = id;
        emit userIdChanged();
    }
}

void User::setUsername(const QString &username)
{
    if (m_username != username) {
        m_username = username;
        emit usernameChanged();
    }
}

void User::setEmail(const QString &email)
{
    if (m_email != email) {
        m_email = email;
        emit emailChanged();
    }
}

void User::setFullName(const QString &fullName)
{
    if (m_fullName != fullName) {
        m_fullName = fullName;
        emit fullNameChanged();
    }
}

void User::setAvatarUrl(const QString &url)
{
    if (m_avatarUrl != url) {
        m_avatarUrl = url;
        emit avatarUrlChanged();
    }
}

void User::setStatus(const QString &status)
{
    if (m_status != status) {
        m_status = status;
        emit statusChanged();
    }
}

void User::setCreatedAt(const QDateTime &dateTime)
{
    m_createdAt = dateTime;
}

QJsonObject User::toJson() const
{
    QJsonObject obj;
    obj["user_id"] = m_userId;
    obj["username"] = m_username;
    obj["email"] = m_email;
    obj["full_name"] = m_fullName;
    obj["avatar_url"] = m_avatarUrl;
    obj["status"] = m_status;
    obj["created_at"] = m_createdAt.toString(Qt::ISODate);
    return obj;
}

void User::fromJson(const QJsonObject &json)
{
    setUserId(json["user_id"].toInt());
    setUsername(json["username"].toString());
    setEmail(json["email"].toString());
    setFullName(json["full_name"].toString());
    setAvatarUrl(json["avatar_url"].toString());
    setStatus(json["status"].toString("offline"));
    
    QString createdAtStr = json["created_at"].toString();
    if (!createdAtStr.isEmpty()) {
        setCreatedAt(QDateTime::fromString(createdAtStr, Qt::ISODate));
    }
}

