#ifndef MEETING_ROOM_CONTROLLER_H
#define MEETING_ROOM_CONTROLLER_H

#include <QObject>
#include <QVariantList>
#include <QVariantMap>
#include <QDateTime>

class MeetingService;
class MediaStream;

class MeetingRoomController : public QObject
{
    Q_OBJECT

    // 会议状态属性
    Q_PROPERTY(int meetingId READ meetingId NOTIFY meetingIdChanged)
    Q_PROPERTY(QString meetingTitle READ meetingTitle NOTIFY meetingTitleChanged)
    Q_PROPERTY(bool isInMeeting READ isInMeeting NOTIFY isInMeetingChanged)
    Q_PROPERTY(QString meetingDuration READ meetingDuration NOTIFY meetingDurationChanged)

    // 媒体状态属性
    Q_PROPERTY(bool audioEnabled READ audioEnabled NOTIFY audioEnabledChanged)
    Q_PROPERTY(bool videoEnabled READ videoEnabled NOTIFY videoEnabledChanged)
    Q_PROPERTY(bool isScreenSharing READ isScreenSharing NOTIFY isScreenSharingChanged)

    // 参与者属性
    Q_PROPERTY(QVariantList participants READ participants NOTIFY participantsChanged)
    Q_PROPERTY(int participantCount READ participantCount NOTIFY participantCountChanged)
    Q_PROPERTY(bool isHost READ isHost NOTIFY isHostChanged)

    // 聊天属性
    Q_PROPERTY(QVariantList chatMessages READ chatMessages NOTIFY chatMessagesChanged)
    Q_PROPERTY(int unreadMessageCount READ unreadMessageCount NOTIFY unreadMessageCountChanged)

public:
    explicit MeetingRoomController(QObject *parent = nullptr);
    ~MeetingRoomController();

    // 属性访问器
    int meetingId() const { return m_meetingId; }
    QString meetingTitle() const { return m_meetingTitle; }
    bool isInMeeting() const { return m_isInMeeting; }
    QString meetingDuration() const { return m_meetingDuration; }

    bool audioEnabled() const { return m_audioEnabled; }
    bool videoEnabled() const { return m_videoEnabled; }
    bool isScreenSharing() const { return m_isScreenSharing; }

    QVariantList participants() const { return m_participants; }
    int participantCount() const { return m_participants.size(); }
    bool isHost() const { return m_isHost; }

    QVariantList chatMessages() const { return m_chatMessages; }
    int unreadMessageCount() const { return m_unreadMessageCount; }

    // 会议操作
    Q_INVOKABLE void joinMeeting(int meetingId, const QString &password = "");
    Q_INVOKABLE void leaveMeeting();
    Q_INVOKABLE void startMeeting();
    Q_INVOKABLE void endMeeting();

    // 媒体控制
    Q_INVOKABLE void toggleAudio();
    Q_INVOKABLE void toggleVideo();
    Q_INVOKABLE void toggleScreenShare();
    Q_INVOKABLE void startScreenShare();
    Q_INVOKABLE void stopScreenShare();

    // 参与者管理
    Q_INVOKABLE void muteParticipant(int userId);
    Q_INVOKABLE void kickParticipant(int userId);
    Q_INVOKABLE void makeHost(int userId);
    Q_INVOKABLE void updateParticipantRole(int userId, const QString &role);

    // 聊天功能
    Q_INVOKABLE void sendChatMessage(const QString &message);
    Q_INVOKABLE void clearUnreadMessages();

    // 视频流管理
    Q_INVOKABLE MediaStream* getLocalStream() const;
    Q_INVOKABLE MediaStream* getRemoteStream(int userId) const;
    Q_INVOKABLE QVariantMap getConnectionStatistics(int userId) const;

signals:
    // 会议状态信号
    void meetingIdChanged();
    void meetingTitleChanged();
    void isInMeetingChanged();
    void meetingDurationChanged();
    void meetingJoined();
    void meetingLeft();
    void meetingStarted();
    void meetingEnded();
    void meetingError(const QString &error);

    // 媒体状态信号
    void audioEnabledChanged();
    void videoEnabledChanged();
    void isScreenSharingChanged();
    void audioToggled(bool enabled);
    void videoToggled(bool enabled);
    void screenShareStarted();
    void screenShareStopped();

    // 参与者信号
    void participantsChanged();
    void participantCountChanged();
    void isHostChanged();
    void participantJoined(int userId, const QString &username);
    void participantLeft(int userId);
    void participantUpdated(int userId);

    // 聊天信号
    void chatMessagesChanged();
    void unreadMessageCountChanged();
    void chatMessageReceived(int fromUserId, const QString &username, const QString &message);

    // WebRTC信号
    void localStreamReady(MediaStream *stream);
    void remoteStreamAdded(int userId, MediaStream *stream);
    void remoteStreamRemoved(int userId);
    void connectionStateChanged(int userId, const QString &state);

private slots:
    // MeetingService信号处理
    void onMeetingJoined();
    void onMeetingLeft();
    void onMeetingError(const QString &error);

    void onAudioEnabledChanged();
    void onVideoEnabledChanged();
    void onScreenSharingChanged();

    void onParticipantJoined(int userId, const QString &username);
    void onParticipantLeft(int userId);
    void onParticipantsListUpdated();

    void onChatMessageReceived(int fromUserId, const QString &username, const QString &content);

    void onLocalStreamReady(MediaStream *stream);
    void onRemoteStreamAdded(int userId, MediaStream *stream);
    void onRemoteStreamRemoved(int userId);
    void onConnectionStateChanged(int userId, const QString &state);

    // 内部方法
    void updateMeetingDuration();
    void setupConnections();
    void updateParticipantsList();

private:
    MeetingService *m_meetingService;

    // 会议状态
    int m_meetingId;
    QString m_meetingTitle;
    bool m_isInMeeting;
    QString m_meetingDuration;
    QDateTime m_meetingStartTime;

    // 媒体状态
    bool m_audioEnabled;
    bool m_videoEnabled;
    bool m_isScreenSharing;

    // 参与者
    QVariantList m_participants;
    bool m_isHost;

    // 聊天
    QVariantList m_chatMessages;
    int m_unreadMessageCount;

    // 定时器
    QTimer *m_durationTimer;
};

#endif // MEETING_ROOM_CONTROLLER_H

