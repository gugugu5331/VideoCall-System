#ifndef USER_H
#define USER_H

#include <QObject>
#include <QString>
#include <QDateTime>
#include <QJsonObject>

class User : public QObject
{
    Q_OBJECT
    Q_PROPERTY(int userId READ userId WRITE setUserId NOTIFY userIdChanged)
    Q_PROPERTY(QString username READ username WRITE setUsername NOTIFY usernameChanged)
    Q_PROPERTY(QString email READ email WRITE setEmail NOTIFY emailChanged)
    Q_PROPERTY(QString fullName READ fullName WRITE setFullName NOTIFY fullNameChanged)
    Q_PROPERTY(QString avatarUrl READ avatarUrl WRITE setAvatarUrl NOTIFY avatarUrlChanged)
    Q_PROPERTY(QString status READ status WRITE setStatus NOTIFY statusChanged)

public:
    explicit User(QObject *parent = nullptr);
    ~User();

    // Getters
    int userId() const { return m_userId; }
    QString username() const { return m_username; }
    QString email() const { return m_email; }
    QString fullName() const { return m_fullName; }
    QString avatarUrl() const { return m_avatarUrl; }
    QString status() const { return m_status; }
    QDateTime createdAt() const { return m_createdAt; }

    // Setters
    void setUserId(int id);
    void setUsername(const QString &username);
    void setEmail(const QString &email);
    void setFullName(const QString &fullName);
    void setAvatarUrl(const QString &url);
    void setStatus(const QString &status);
    void setCreatedAt(const QDateTime &dateTime);

    // Serialization
    QJsonObject toJson() const;
    void fromJson(const QJsonObject &json);

signals:
    void userIdChanged();
    void usernameChanged();
    void emailChanged();
    void fullNameChanged();
    void avatarUrlChanged();
    void statusChanged();

private:
    int m_userId;
    QString m_username;
    QString m_email;
    QString m_fullName;
    QString m_avatarUrl;
    QString m_status;
    QDateTime m_createdAt;
};

#endif // USER_H

