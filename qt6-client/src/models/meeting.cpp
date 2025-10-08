#include "models/meeting.h"

Meeting::Meeting(QObject *parent)
    : QObject(parent)
    , m_meetingId(0)
    , m_hostId(0)
    , m_participantCount(0)
    , m_maxParticipants(10)
    , m_duration(60)
    , m_isPublic(false)
    , m_status("scheduled")
{
}

Meeting::~Meeting()
{
}

void Meeting::setMeetingId(int id)
{
    if (m_meetingId != id) {
        m_meetingId = id;
        emit meetingIdChanged();
    }
}

void Meeting::setTitle(const QString &title)
{
    if (m_title != title) {
        m_title = title;
        emit titleChanged();
    }
}

void Meeting::setDescription(const QString &description)
{
    if (m_description != description) {
        m_description = description;
        emit descriptionChanged();
    }
}

void Meeting::setMeetingCode(const QString &code)
{
    if (m_meetingCode != code) {
        m_meetingCode = code;
        emit meetingCodeChanged();
    }
}

void Meeting::setStatus(const QString &status)
{
    if (m_status != status) {
        m_status = status;
        emit statusChanged();
    }
}

void Meeting::setHostId(int hostId)
{
    if (m_hostId != hostId) {
        m_hostId = hostId;
        emit hostIdChanged();
    }
}

void Meeting::setParticipantCount(int count)
{
    if (m_participantCount != count) {
        m_participantCount = count;
        emit participantCountChanged();
    }
}

void Meeting::setMaxParticipants(int max)
{
    m_maxParticipants = max;
}

void Meeting::setStartTime(const QDateTime &time)
{
    m_startTime = time;
}

void Meeting::setDuration(int minutes)
{
    m_duration = minutes;
}

void Meeting::setIsPublic(bool isPublic)
{
    m_isPublic = isPublic;
}

void Meeting::setSettings(const QJsonObject &settings)
{
    m_settings = settings;
}

QJsonObject Meeting::toJson() const
{
    QJsonObject obj;
    obj["meeting_id"] = m_meetingId;
    obj["title"] = m_title;
    obj["description"] = m_description;
    obj["meeting_code"] = m_meetingCode;
    obj["status"] = m_status;
    obj["host_id"] = m_hostId;
    obj["participant_count"] = m_participantCount;
    obj["max_participants"] = m_maxParticipants;
    obj["start_time"] = m_startTime.toString(Qt::ISODate);
    obj["duration"] = m_duration;
    obj["is_public"] = m_isPublic;
    obj["settings"] = m_settings;
    return obj;
}

void Meeting::fromJson(const QJsonObject &json)
{
    setMeetingId(json["meeting_id"].toInt());
    setTitle(json["title"].toString());
    setDescription(json["description"].toString());
    setMeetingCode(json["meeting_code"].toString());
    setStatus(json["status"].toString("scheduled"));
    setHostId(json["host_id"].toInt());
    setParticipantCount(json["participant_count"].toInt());
    setMaxParticipants(json["max_participants"].toInt(10));
    setDuration(json["duration"].toInt(60));
    setIsPublic(json["is_public"].toBool(false));
    setSettings(json["settings"].toObject());
    
    QString startTimeStr = json["start_time"].toString();
    if (!startTimeStr.isEmpty()) {
        setStartTime(QDateTime::fromString(startTimeStr, Qt::ISODate));
    }
}

