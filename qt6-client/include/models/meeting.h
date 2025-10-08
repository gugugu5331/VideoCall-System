#ifndef MEETING_H
#define MEETING_H

#include <QObject>
#include <QString>
#include <QDateTime>
#include <QJsonObject>
#include <QList>

class Meeting : public QObject
{
    Q_OBJECT
    Q_PROPERTY(int meetingId READ meetingId WRITE setMeetingId NOTIFY meetingIdChanged)
    Q_PROPERTY(QString title READ title WRITE setTitle NOTIFY titleChanged)
    Q_PROPERTY(QString description READ description WRITE setDescription NOTIFY descriptionChanged)
    Q_PROPERTY(QString meetingCode READ meetingCode WRITE setMeetingCode NOTIFY meetingCodeChanged)
    Q_PROPERTY(QString status READ status WRITE setStatus NOTIFY statusChanged)
    Q_PROPERTY(int hostId READ hostId WRITE setHostId NOTIFY hostIdChanged)
    Q_PROPERTY(int participantCount READ participantCount WRITE setParticipantCount NOTIFY participantCountChanged)

public:
    explicit Meeting(QObject *parent = nullptr);
    ~Meeting();

    // Getters
    int meetingId() const { return m_meetingId; }
    QString title() const { return m_title; }
    QString description() const { return m_description; }
    QString meetingCode() const { return m_meetingCode; }
    QString status() const { return m_status; }
    int hostId() const { return m_hostId; }
    int participantCount() const { return m_participantCount; }
    int maxParticipants() const { return m_maxParticipants; }
    QDateTime startTime() const { return m_startTime; }
    int duration() const { return m_duration; }
    bool isPublic() const { return m_isPublic; }
    QJsonObject settings() const { return m_settings; }

    // Setters
    void setMeetingId(int id);
    void setTitle(const QString &title);
    void setDescription(const QString &description);
    void setMeetingCode(const QString &code);
    void setStatus(const QString &status);
    void setHostId(int hostId);
    void setParticipantCount(int count);
    void setMaxParticipants(int max);
    void setStartTime(const QDateTime &time);
    void setDuration(int minutes);
    void setIsPublic(bool isPublic);
    void setSettings(const QJsonObject &settings);

    // Serialization
    QJsonObject toJson() const;
    void fromJson(const QJsonObject &json);

signals:
    void meetingIdChanged();
    void titleChanged();
    void descriptionChanged();
    void meetingCodeChanged();
    void statusChanged();
    void hostIdChanged();
    void participantCountChanged();

private:
    int m_meetingId;
    QString m_title;
    QString m_description;
    QString m_meetingCode;
    QString m_status;
    int m_hostId;
    int m_participantCount;
    int m_maxParticipants;
    QDateTime m_startTime;
    int m_duration;
    bool m_isPublic;
    QJsonObject m_settings;
};

#endif // MEETING_H

