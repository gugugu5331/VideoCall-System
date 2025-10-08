#ifndef MESSAGE_H
#define MESSAGE_H

#include <QObject>
#include <QString>
#include <QDateTime>
#include <QJsonObject>

enum class MessageType {
    Text,
    Image,
    File,
    System
};

class Message : public QObject
{
    Q_OBJECT
    Q_PROPERTY(QString messageId READ messageId WRITE setMessageId NOTIFY messageIdChanged)
    Q_PROPERTY(int fromUserId READ fromUserId WRITE setFromUserId NOTIFY fromUserIdChanged)
    Q_PROPERTY(QString fromUsername READ fromUsername WRITE setFromUsername NOTIFY fromUsernameChanged)
    Q_PROPERTY(QString content READ content WRITE setContent NOTIFY contentChanged)
    Q_PROPERTY(QString timestamp READ timestamp NOTIFY timestampChanged)
    Q_PROPERTY(QDateTime dateTime READ dateTime WRITE setTimestamp NOTIFY timestampChanged)

public:
    explicit Message(QObject *parent = nullptr);
    ~Message();

    // Getters
    QString messageId() const { return m_messageId; }
    int fromUserId() const { return m_fromUserId; }
    QString fromUsername() const { return m_fromUsername; }
    QString content() const { return m_content; }
    MessageType type() const { return m_type; }
    QString timestamp() const { return m_timestamp.toString("hh:mm:ss"); }
    QDateTime dateTime() const { return m_timestamp; }

    // Setters
    void setMessageId(const QString &id);
    void setFromUserId(int userId);
    void setFromUsername(const QString &username);
    void setContent(const QString &content);
    void setType(MessageType type);
    void setTimestamp(const QDateTime &timestamp);

    // Serialization
    QJsonObject toJson() const;
    void fromJson(const QJsonObject &json);

signals:
    void messageIdChanged();
    void fromUserIdChanged();
    void fromUsernameChanged();
    void contentChanged();
    void timestampChanged();

private:
    QString m_messageId;
    int m_fromUserId;
    QString m_fromUsername;
    QString m_content;
    MessageType m_type;
    QDateTime m_timestamp;
};

#endif // MESSAGE_H

