#ifndef PARTICIPANT_H
#define PARTICIPANT_H

#include <QObject>
#include <QString>
#include <QDateTime>
#include <QJsonObject>

class Participant : public QObject
{
    Q_OBJECT
    Q_PROPERTY(int userId READ userId WRITE setUserId NOTIFY userIdChanged)
    Q_PROPERTY(QString username READ username WRITE setUsername NOTIFY usernameChanged)
    Q_PROPERTY(QString peerId READ peerId WRITE setPeerId NOTIFY peerIdChanged)
    Q_PROPERTY(QString role READ role WRITE setRole NOTIFY roleChanged)
    Q_PROPERTY(QString status READ status WRITE setStatus NOTIFY statusChanged)
    Q_PROPERTY(bool audioEnabled READ audioEnabled WRITE setAudioEnabled NOTIFY audioEnabledChanged)
    Q_PROPERTY(bool videoEnabled READ videoEnabled WRITE setVideoEnabled NOTIFY videoEnabledChanged)
    Q_PROPERTY(bool isScreenSharing READ isScreenSharing WRITE setIsScreenSharing NOTIFY isScreenSharingChanged)
    Q_PROPERTY(bool isSelf READ isSelf WRITE setIsSelf NOTIFY isSelfChanged)

public:
    explicit Participant(QObject *parent = nullptr);
    ~Participant();

    // Getters
    int userId() const { return m_userId; }
    QString username() const { return m_username; }
    QString peerId() const { return m_peerId; }
    QString sessionId() const { return m_sessionId; }
    QString role() const { return m_role; }
    QString status() const { return m_status; }
    bool audioEnabled() const { return m_audioEnabled; }
    bool videoEnabled() const { return m_videoEnabled; }
    bool isScreenSharing() const { return m_isScreenSharing; }
    bool isSelf() const { return m_isSelf; }
    QDateTime joinedAt() const { return m_joinedAt; }

    // Setters
    void setUserId(int id);
    void setUsername(const QString &username);
    void setPeerId(const QString &peerId);
    void setSessionId(const QString &sessionId);
    void setRole(const QString &role);
    void setStatus(const QString &status);
    void setAudioEnabled(bool enabled);
    void setVideoEnabled(bool enabled);
    void setIsScreenSharing(bool sharing);
    void setIsSelf(bool isSelf);
    void setJoinedAt(const QDateTime &time);

    // Serialization
    QJsonObject toJson() const;
    void fromJson(const QJsonObject &json);

signals:
    void userIdChanged();
    void usernameChanged();
    void peerIdChanged();
    void roleChanged();
    void statusChanged();
    void audioEnabledChanged();
    void videoEnabledChanged();
    void isScreenSharingChanged();
    void isSelfChanged();

private:
    int m_userId;
    QString m_username;
    QString m_peerId;
    QString m_sessionId;
    QString m_role;
    QString m_status;
    bool m_audioEnabled;
    bool m_videoEnabled;
    bool m_isScreenSharing;
    bool m_isSelf;
    QDateTime m_joinedAt;
};

#endif // PARTICIPANT_H

