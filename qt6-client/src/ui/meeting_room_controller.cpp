#include "ui/meeting_room_controller.h"
#include "application.h"
#include "services/meeting_service.h"
#include "webrtc/media_stream.h"
#include "utils/logger.h"
#include <QTimer>

MeetingRoomController::MeetingRoomController(QObject *parent)
    : QObject(parent)
    , m_meetingService(nullptr)
    , m_meetingId(0)
    , m_isInMeeting(false)
    , m_audioEnabled(true)
    , m_videoEnabled(true)
    , m_isScreenSharing(false)
    , m_isHost(false)
    , m_unreadMessageCount(0)
    , m_durationTimer(nullptr)
{
    m_meetingService = Application::instance()->meetingService();
    setupConnections();

    // 创建会议时长定时器
    m_durationTimer = new QTimer(this);
    connect(m_durationTimer, &QTimer::timeout, this, &MeetingRoomController::updateMeetingDuration);
}

MeetingRoomController::~MeetingRoomController()
{
    if (m_durationTimer) {
        m_durationTimer->stop();
    }
}

// ==================== 会议操作 ====================

void MeetingRoomController::joinMeeting(int meetingId, const QString &password)
{
    LOG_INFO(QString("Joining meeting: %1").arg(meetingId));

    m_meetingId = meetingId;
    emit meetingIdChanged();

    m_meetingService->joinMeeting(meetingId, password);
}

void MeetingRoomController::leaveMeeting()
{
    LOG_INFO("Leaving meeting");
    m_meetingService->leaveMeeting();
}

void MeetingRoomController::startMeeting()
{
    LOG_INFO("Starting meeting");
    m_meetingService->startMeeting(m_meetingId);
}

void MeetingRoomController::endMeeting()
{
    LOG_INFO("Ending meeting");
    m_meetingService->endMeeting(m_meetingId);
}

// ==================== 媒体控制 ====================

void MeetingRoomController::toggleAudio()
{
    m_meetingService->toggleAudio();
    LOG_INFO(QString("Audio toggled"));
}

void MeetingRoomController::toggleVideo()
{
    m_meetingService->toggleVideo();
    LOG_INFO(QString("Video toggled"));
}

void MeetingRoomController::toggleScreenShare()
{
    if (m_isScreenSharing) {
        stopScreenShare();
    } else {
        startScreenShare();
    }
}

void MeetingRoomController::startScreenShare()
{
    m_meetingService->startScreenShare();
    LOG_INFO("Screen share started");
}

void MeetingRoomController::stopScreenShare()
{
    m_meetingService->stopScreenShare();
    LOG_INFO("Screen share stopped");
}

// ==================== 参与者管理 ====================

void MeetingRoomController::muteParticipant(int userId)
{
    LOG_INFO(QString("Muting participant: %1").arg(userId));
    m_meetingService->muteParticipant(userId, true);
}

void MeetingRoomController::kickParticipant(int userId)
{
    LOG_INFO(QString("Kicking participant: %1").arg(userId));
    m_meetingService->kickParticipant(userId);
}

void MeetingRoomController::makeHost(int userId)
{
    LOG_INFO(QString("Making participant host: %1").arg(userId));
    m_meetingService->updateParticipantRole(m_meetingId, userId, "host");
}

void MeetingRoomController::updateParticipantRole(int userId, const QString &role)
{
    LOG_INFO(QString("Updating participant %1 role to: %2").arg(userId).arg(role));
    m_meetingService->updateParticipantRole(m_meetingId, userId, role);
}

// ==================== 聊天功能 ====================

void MeetingRoomController::sendChatMessage(const QString &message)
{
    if (message.isEmpty()) {
        return;
    }

    m_meetingService->sendChatMessage(message);
    LOG_DEBUG("Chat message sent: " + message);
}

void MeetingRoomController::clearUnreadMessages()
{
    m_unreadMessageCount = 0;
    emit unreadMessageCountChanged();
}

// ==================== 视频流管理 ====================

MediaStream* MeetingRoomController::getLocalStream() const
{
    return m_meetingService->getLocalStream();
}

MediaStream* MeetingRoomController::getRemoteStream(int userId) const
{
    // 这里需要从WebRTCManager获取远程流
    // 暂时返回nullptr
    return nullptr;
}

QVariantMap MeetingRoomController::getConnectionStatistics(int userId) const
{
    return m_meetingService->getConnectionStatistics(userId);
}


// ==================== MeetingService信号处理 ====================

void MeetingRoomController::onMeetingJoined()
{
    LOG_INFO("Meeting joined successfully");

    m_isInMeeting = true;
    emit isInMeetingChanged();
    emit meetingJoined();

    // 启动会议时长计时器
    m_meetingStartTime = QDateTime::currentDateTime();
    m_durationTimer->start(1000); // 每秒更新一次

    // 更新参与者列表
    updateParticipantsList();
}

void MeetingRoomController::onMeetingLeft()
{
    LOG_INFO("Meeting left");

    m_isInMeeting = false;
    emit isInMeetingChanged();
    emit meetingLeft();

    // 停止计时器
    m_durationTimer->stop();
    m_meetingDuration = "00:00:00";
    emit meetingDurationChanged();

    // 清空数据
    m_participants.clear();
    emit participantsChanged();
    emit participantCountChanged();

    m_chatMessages.clear();
    emit chatMessagesChanged();

    m_unreadMessageCount = 0;
    emit unreadMessageCountChanged();
}

void MeetingRoomController::onMeetingError(const QString &error)
{
    LOG_ERROR(QString("Meeting error: %1").arg(error));
    emit meetingError(error);
}

void MeetingRoomController::onAudioEnabledChanged()
{
    bool enabled = m_meetingService->audioEnabled();
    if (m_audioEnabled != enabled) {
        m_audioEnabled = enabled;
        emit audioEnabledChanged();
        emit audioToggled(enabled);
        LOG_INFO(QString("Audio %1").arg(enabled ? "enabled" : "disabled"));
    }
}

void MeetingRoomController::onVideoEnabledChanged()
{
    bool enabled = m_meetingService->videoEnabled();
    if (m_videoEnabled != enabled) {
        m_videoEnabled = enabled;
        emit videoEnabledChanged();
        emit videoToggled(enabled);
        LOG_INFO(QString("Video %1").arg(enabled ? "enabled" : "disabled"));
    }
}

void MeetingRoomController::onScreenSharingChanged()
{
    bool sharing = m_meetingService->isScreenSharing();
    if (m_isScreenSharing != sharing) {
        m_isScreenSharing = sharing;
        emit isScreenSharingChanged();

        if (sharing) {
            emit screenShareStarted();
        } else {
            emit screenShareStopped();
        }

        LOG_INFO(QString("Screen sharing %1").arg(sharing ? "started" : "stopped"));
    }
}

void MeetingRoomController::onParticipantJoined(int userId, const QString &username)
{
    LOG_INFO(QString("Participant joined: %1 (%2)").arg(username).arg(userId));

    emit participantJoined(userId, username);
    updateParticipantsList();
}

void MeetingRoomController::onParticipantLeft(int userId)
{
    LOG_INFO(QString("Participant left: %1").arg(userId));

    emit participantLeft(userId);
    updateParticipantsList();
}

void MeetingRoomController::onParticipantsListUpdated()
{
    updateParticipantsList();
}

void MeetingRoomController::onChatMessageReceived(int fromUserId, const QString &username, const QString &content)
{
    LOG_DEBUG(QString("Chat message from %1: %2").arg(username, content));

    // 添加到消息列表
    QVariantMap message;
    message["fromUserId"] = fromUserId;
    message["fromUsername"] = username;
    message["content"] = content;
    message["timestamp"] = QDateTime::currentDateTime();

    m_chatMessages.append(message);
    emit chatMessagesChanged();

    // 增加未读消息计数
    m_unreadMessageCount++;
    emit unreadMessageCountChanged();

    emit chatMessageReceived(fromUserId, username, content);
}

void MeetingRoomController::onLocalStreamReady(MediaStream *stream)
{
    LOG_INFO("Local stream ready");
    emit localStreamReady(stream);
}

void MeetingRoomController::onRemoteStreamAdded(int userId, MediaStream *stream)
{
    LOG_INFO(QString("Remote stream added for user: %1").arg(userId));
    emit remoteStreamAdded(userId, stream);
}

void MeetingRoomController::onRemoteStreamRemoved(int userId)
{
    LOG_INFO(QString("Remote stream removed for user: %1").arg(userId));
    emit remoteStreamRemoved(userId);
}

void MeetingRoomController::onConnectionStateChanged(int userId, const QString &state)
{
    LOG_INFO(QString("Connection state changed for user %1: %2").arg(userId).arg(state));
    emit connectionStateChanged(userId, state);
}

// ==================== 内部方法 ====================

void MeetingRoomController::updateMeetingDuration()
{
    if (!m_isInMeeting) {
        return;
    }

    qint64 seconds = m_meetingStartTime.secsTo(QDateTime::currentDateTime());
    int hours = seconds / 3600;
    int minutes = (seconds % 3600) / 60;
    int secs = seconds % 60;

    m_meetingDuration = QString("%1:%2:%3")
        .arg(hours, 2, 10, QChar('0'))
        .arg(minutes, 2, 10, QChar('0'))
        .arg(secs, 2, 10, QChar('0'));

    emit meetingDurationChanged();
}

void MeetingRoomController::setupConnections()
{
    if (!m_meetingService) {
        return;
    }

    // 会议状态信号
    connect(m_meetingService, &MeetingService::meetingJoined,
            this, &MeetingRoomController::onMeetingJoined);
    connect(m_meetingService, &MeetingService::meetingLeft,
            this, &MeetingRoomController::onMeetingLeft);
    connect(m_meetingService, &MeetingService::meetingError,
            this, &MeetingRoomController::onMeetingError);

    // 媒体状态信号
    connect(m_meetingService, &MeetingService::audioEnabledChanged,
            this, &MeetingRoomController::onAudioEnabledChanged);
    connect(m_meetingService, &MeetingService::videoEnabledChanged,
            this, &MeetingRoomController::onVideoEnabledChanged);
    connect(m_meetingService, &MeetingService::isScreenSharingChanged,
            this, &MeetingRoomController::onScreenSharingChanged);

    // 参与者信号
    connect(m_meetingService, &MeetingService::participantJoined,
            this, &MeetingRoomController::onParticipantJoined);
    connect(m_meetingService, &MeetingService::participantLeft,
            this, &MeetingRoomController::onParticipantLeft);
    connect(m_meetingService, &MeetingService::participantsListUpdated,
            this, &MeetingRoomController::onParticipantsListUpdated);

    // 聊天信号
    connect(m_meetingService, &MeetingService::chatMessageReceived,
            this, &MeetingRoomController::onChatMessageReceived);

    // WebRTC信号
    connect(m_meetingService, &MeetingService::localStreamReady,
            this, &MeetingRoomController::onLocalStreamReady);
    connect(m_meetingService, &MeetingService::remoteStreamAdded,
            this, &MeetingRoomController::onRemoteStreamAdded);
    connect(m_meetingService, &MeetingService::remoteStreamRemoved,
            this, &MeetingRoomController::onRemoteStreamRemoved);
    connect(m_meetingService, &MeetingService::connectionStateChanged,
            this, &MeetingRoomController::onConnectionStateChanged);
}

void MeetingRoomController::updateParticipantsList()
{
    m_participants.clear();

    // 从MeetingService获取参与者列表
    const auto &participants = m_meetingService->participants();

    for (const auto &participant : participants) {
        QVariantMap participantMap;
        participantMap["userId"] = participant->userId();
        participantMap["username"] = participant->username();
        participantMap["role"] = participant->role();
        participantMap["status"] = participant->status();
        participantMap["audioEnabled"] = participant->audioEnabled();
        participantMap["videoEnabled"] = participant->videoEnabled();
        participantMap["isScreenSharing"] = participant->isScreenSharing();
        participantMap["joinedAt"] = participant->joinedAt().toString("yyyy-MM-dd hh:mm:ss");
        participantMap["networkQuality"] = 2; // 默认网络质量，后续可以从统计信息获取

        m_participants.append(participantMap);
    }

    emit participantsChanged();
    emit participantCountChanged();

    LOG_DEBUG(QString("Participants list updated: %1 participants").arg(m_participants.size()));
}

