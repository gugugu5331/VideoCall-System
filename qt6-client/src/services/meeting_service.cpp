#include "services/meeting_service.h"
#include "webrtc/webrtc_manager.h"
#include "webrtc/media_stream.h"
#include "utils/logger.h"
#include <QJsonArray>

MeetingService::MeetingService(ApiClient *apiClient, WebSocketClient *wsClient, QObject *parent)
    : QObject(parent)
    , m_apiClient(apiClient)
    , m_wsClient(wsClient)
    , m_webrtcManager(nullptr)
    , m_audioEnabled(true)
    , m_videoEnabled(true)
    , m_isScreenSharing(false)
{
    m_currentMeeting = std::make_unique<Meeting>();

    // 创建WebRTC管理器
    m_webrtcManager = new WebRTCManager(wsClient, this);

    setupWebSocketConnections();
    setupWebRTCConnections();

    LOG_INFO("MeetingService created with WebRTC support");
}

MeetingService::~MeetingService()
{
    if (m_webrtcManager) {
        m_webrtcManager->stopLocalMedia();
        m_webrtcManager->closeAllPeerConnections();
    }
    LOG_INFO("MeetingService destroyed");
}

void MeetingService::createMeeting(const QString &title, const QString &description,
                                  const QDateTime &startTime, const QDateTime &endTime,
                                  int maxParticipants, const QString &meetingType,
                                  const QString &password, const QJsonObject &settings)
{
    LOG_INFO("Creating meeting: " + title);

    m_apiClient->createMeeting(title, description, startTime, endTime, maxParticipants,
                              meetingType, password, settings,
        [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                QJsonObject meetingData = response.data;
                m_currentMeeting = std::make_unique<Meeting>();
                // TODO: Implement Meeting::fromJson

                LOG_INFO("Meeting created successfully");
                emit meetingCreated(m_currentMeeting.get());
            } else {
                LOG_ERROR("Failed to create meeting: " + response.message);
                emit meetingError(response.message);
            }
        });
}

void MeetingService::joinMeeting(int meetingId, const QString &password)
{
    LOG_INFO(QString("Joining meeting: %1").arg(meetingId));

    m_apiClient->joinMeeting(meetingId, password, [this, meetingId](const ApiResponse &response) {
        if (response.isSuccess()) {
            QJsonObject meetingData = response.data["meeting"].toObject();
            m_currentMeeting->fromJson(meetingData);

            // Get WebSocket URL and connect
            QString wsUrl = response.data["websocket_url"].toString();
            QString token = response.data["token"].toString();
            int userId = response.data["user_id"].toInt();
            QString peerId = response.data["peer_id"].toString();

            m_wsClient->connect(wsUrl, token, meetingId, userId, peerId);

            // 初始化WebRTC
            QVariantMap config;
            m_webrtcManager->initialize(config);

            // 启动本地媒体
            if (m_webrtcManager->startLocalMedia(m_audioEnabled, m_videoEnabled)) {
                LOG_INFO("Local media started successfully");
            } else {
                LOG_WARNING("Failed to start local media");
            }

            LOG_INFO("Joined meeting successfully");
            emit meetingJoined(m_currentMeeting.get());
        } else {
            LOG_ERROR("Failed to join meeting: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::leaveMeeting()
{
    if (!m_currentMeeting || m_currentMeeting->meetingId() == 0) {
        return;
    }

    LOG_INFO("Leaving meeting");

    int meetingId = m_currentMeeting->meetingId();

    // 停止WebRTC
    if (m_webrtcManager) {
        m_webrtcManager->stopLocalMedia();
        m_webrtcManager->closeAllPeerConnections();
        LOG_INFO("WebRTC stopped");
    }

    m_apiClient->leaveMeeting(meetingId, [this](const ApiResponse &response) {
        m_wsClient->disconnect();
        m_currentMeeting = std::make_unique<Meeting>();
        m_participants.clear();

        LOG_INFO("Left meeting successfully");
        emit meetingLeft();
    });
}

void MeetingService::getMeetingList()
{
    LOG_INFO("Fetching meeting list");

    m_apiClient->getMeetingList(1, 100, "", "", [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            m_meetingList.clear();

            QJsonArray meetings = response.data["meetings"].toArray();
            for (const QJsonValue &value : meetings) {
                auto meeting = std::make_unique<Meeting>();
                meeting->fromJson(value.toObject());
                m_meetingList.push_back(std::move(meeting));
            }

            LOG_INFO(QString("Fetched %1 meetings").arg(m_meetingList.size()));
            emit meetingListUpdated();
        } else {
            LOG_ERROR("Failed to fetch meeting list: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::sendChatMessage(const QString &message)
{
    m_wsClient->sendChatMessage(message);
    LOG_DEBUG("Chat message sent: " + message);
}

void MeetingService::getParticipants(int meetingId)
{
    LOG_INFO(QString("Getting participants for meeting: %1").arg(meetingId));

    m_apiClient->getParticipants(meetingId, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            m_participants.clear();

            QJsonArray participantsArray = response.data["participants"].toArray();
            for (const QJsonValue &value : participantsArray) {
                QJsonObject pObj = value.toObject();

                auto participant = std::make_unique<Participant>();
                participant->setUserId(pObj["user_id"].toInt());
                participant->setUsername(pObj["username"].toString());
                participant->setRole(pObj["role"].toString());
                participant->setStatus(pObj["status"].toString());
                participant->setAudioEnabled(pObj["audio_enabled"].toBool(true));
                participant->setVideoEnabled(pObj["video_enabled"].toBool(true));
                participant->setJoinedAt(QDateTime::fromString(pObj["joined_at"].toString(), Qt::ISODate));

                m_participants.push_back(std::move(participant));
            }

            LOG_INFO(QString("Loaded %1 participants").arg(m_participants.size()));
            emit participantsListUpdated();
        } else {
            LOG_ERROR("Failed to get participants: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::addParticipant(int meetingId, int userId, const QString &role)
{
    LOG_INFO(QString("Adding participant %1 to meeting %2").arg(userId).arg(meetingId));

    m_apiClient->addParticipant(meetingId, userId, role, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            LOG_INFO("Participant added successfully");
            // 重新获取参与者列表
            getParticipants(m_currentMeeting->meetingId());
        } else {
            LOG_ERROR("Failed to add participant: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::removeParticipant(int meetingId, int userId)
{
    LOG_INFO(QString("Removing participant %1 from meeting %2").arg(userId).arg(meetingId));

    m_apiClient->removeParticipant(meetingId, userId, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            LOG_INFO("Participant removed successfully");
            // 重新获取参与者列表
            getParticipants(m_currentMeeting->meetingId());
        } else {
            LOG_ERROR("Failed to remove participant: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::updateParticipantRole(int meetingId, int userId, const QString &role)
{
    LOG_INFO(QString("Updating participant %1 role to %2").arg(userId).arg(role));

    m_apiClient->updateParticipantRole(meetingId, userId, role, [this, userId, role](const ApiResponse &response) {
        if (response.isSuccess()) {
            LOG_INFO("Participant role updated successfully");
            // 更新本地参与者信息
            for (auto &participant : m_participants) {
                if (participant->userId() == userId) {
                    participant->setRole(role);
                    QJsonObject data;
                    data["role"] = role;
                    emit participantUpdated(userId, data);
                    break;
                }
            }
        } else {
            LOG_ERROR("Failed to update participant role: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::kickParticipant(int userId)
{
    if (!m_currentMeeting || m_currentMeeting->meetingId() == 0) {
        LOG_WARNING("Cannot kick participant: not in a meeting");
        return;
    }

    LOG_INFO(QString("Kicking participant: %1").arg(userId));
    removeParticipant(m_currentMeeting->meetingId(), userId);
}

void MeetingService::muteParticipant(int userId, bool mute)
{
    LOG_INFO(QString("%1 participant: %2").arg(mute ? "Muting" : "Unmuting").arg(userId));

    // 通过WebSocket发送媒体控制消息
    QJsonObject payload;
    payload["user_id"] = userId;
    payload["media_type"] = "audio";
    payload["enabled"] = !mute;

    m_wsClient->sendSignalingMessage(SignalingMessageType::MediaControl, payload, userId);
}

void MeetingService::getChatMessages(int meetingId)
{
    LOG_INFO(QString("Getting chat messages for meeting: %1").arg(meetingId));

    m_apiClient->getChatMessages(meetingId, 1, 100, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            m_messages.clear();

            QJsonArray messagesArray = response.data["messages"].toArray();
            for (const QJsonValue &value : messagesArray) {
                QJsonObject msgObj = value.toObject();

                auto message = std::make_unique<Message>();
                message->setMessageId(QString::number(msgObj["message_id"].toInt()));
                message->setFromUserId(msgObj["from_user_id"].toInt());
                message->setFromUsername(msgObj["from_username"].toString());
                message->setContent(msgObj["content"].toString());
                message->setTimestamp(QDateTime::fromString(msgObj["timestamp"].toString(), Qt::ISODate));

                m_messages.push_back(std::move(message));
            }

            LOG_INFO(QString("Loaded %1 chat messages").arg(m_messages.size()));
            emit chatMessagesLoaded();
        } else {
            LOG_ERROR("Failed to get chat messages: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::getMeetingInfo(int meetingId)
{
    LOG_INFO(QString("Getting meeting info: %1").arg(meetingId));

    m_apiClient->getMeetingInfo(meetingId, [this](const ApiResponse &response) {
        if (response.isSuccess()) {
            auto meeting = std::make_unique<Meeting>();
            meeting->fromJson(response.data);

            LOG_INFO("Meeting info received");
            emit meetingInfoReceived(meeting.get());
        } else {
            LOG_ERROR("Failed to get meeting info: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::startMeeting(int meetingId)
{
    LOG_INFO(QString("Starting meeting: %1").arg(meetingId));

    m_apiClient->startMeeting(meetingId, [this, meetingId](const ApiResponse &response) {
        if (response.isSuccess()) {
            LOG_INFO("Meeting started successfully");
            if (m_currentMeeting && m_currentMeeting->meetingId() == meetingId) {
                m_currentMeeting->setStatus("ongoing");
            }
        } else {
            LOG_ERROR("Failed to start meeting: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::endMeeting(int meetingId)
{
    LOG_INFO(QString("Ending meeting: %1").arg(meetingId));

    m_apiClient->endMeeting(meetingId, [this, meetingId](const ApiResponse &response) {
        if (response.isSuccess()) {
            LOG_INFO("Meeting ended successfully");
            if (m_currentMeeting && m_currentMeeting->meetingId() == meetingId) {
                m_currentMeeting->setStatus("ended");
                leaveMeeting();
            }
        } else {
            LOG_ERROR("Failed to end meeting: " + response.message);
            emit meetingError(response.message);
        }
    });
}

void MeetingService::toggleAudio()
{
    if (m_webrtcManager) {
        m_webrtcManager->toggleAudio();
        m_audioEnabled = m_webrtcManager->audioEnabled();
        emit audioEnabledChanged();

        // 通知其他参与者
        m_wsClient->sendMediaControl("audio", m_audioEnabled);
    }
}

void MeetingService::toggleVideo()
{
    if (m_webrtcManager) {
        m_webrtcManager->toggleVideo();
        m_videoEnabled = m_webrtcManager->videoEnabled();
        emit videoEnabledChanged();

        // 通知其他参与者
        m_wsClient->sendMediaControl("video", m_videoEnabled);
    }
}

void MeetingService::startScreenShare()
{
    if (m_webrtcManager) {
        if (m_webrtcManager->startScreenShare()) {
            m_isScreenSharing = true;
            emit isScreenSharingChanged();

            // 通知其他参与者
            m_wsClient->sendScreenShareControl(true);
        }
    }
}

void MeetingService::stopScreenShare()
{
    if (m_webrtcManager) {
        m_webrtcManager->stopScreenShare();
        m_isScreenSharing = false;
        emit isScreenSharingChanged();

        // 通知其他参与者
        m_wsClient->sendScreenShareControl(false);
    }
}

void MeetingService::startLocalMedia(bool audio, bool video)
{
    if (m_webrtcManager) {
        m_webrtcManager->startLocalMedia(audio, video);
    }
}

void MeetingService::stopLocalMedia()
{
    if (m_webrtcManager) {
        m_webrtcManager->stopLocalMedia();
    }
}

MediaStream* MeetingService::getLocalStream() const
{
    if (m_webrtcManager) {
        return m_webrtcManager->getLocalStream();
    }
    return nullptr;
}

QVariantMap MeetingService::getConnectionStatistics(int userId) const
{
    if (m_webrtcManager) {
        return m_webrtcManager->getStatistics(userId);
    }
    return QVariantMap();
}

void MeetingService::setAudioEnabled(bool enabled)
{
    if (m_audioEnabled != enabled) {
        m_audioEnabled = enabled;
        if (m_webrtcManager) {
            m_webrtcManager->setAudioEnabled(enabled);
        }
        emit audioEnabledChanged();
    }
}

void MeetingService::setVideoEnabled(bool enabled)
{
    if (m_videoEnabled != enabled) {
        m_videoEnabled = enabled;
        if (m_webrtcManager) {
            m_webrtcManager->setVideoEnabled(enabled);
        }
        emit videoEnabledChanged();
    }
}

void MeetingService::setIsScreenSharing(bool sharing)
{
    if (m_isScreenSharing != sharing) {
        m_isScreenSharing = sharing;
        emit isScreenSharingChanged();
    }
}

void MeetingService::setupWebSocketConnections()
{
    // 连接WebSocket基础信号
    connect(m_wsClient, &WebSocketClient::connected, this, &MeetingService::onWebSocketConnected);
    connect(m_wsClient, &WebSocketClient::disconnected, this, &MeetingService::onWebSocketDisconnected);
    connect(m_wsClient, &WebSocketClient::signalingMessageReceived,
            this, &MeetingService::onSignalingMessageReceived);
}

void MeetingService::setupWebRTCConnections()
{
    if (!m_webrtcManager) {
        return;
    }

    // 连接WebRTC信号到MeetingService
    connect(m_webrtcManager, &WebRTCManager::localStreamReady,
            this, &MeetingService::localStreamReady);

    connect(m_webrtcManager, &WebRTCManager::localStreamStopped,
            this, &MeetingService::localStreamStopped);

    connect(m_webrtcManager, &WebRTCManager::remoteStreamAdded,
            this, &MeetingService::remoteStreamAdded);

    connect(m_webrtcManager, &WebRTCManager::remoteStreamRemoved,
            this, &MeetingService::remoteStreamRemoved);

    connect(m_webrtcManager, &WebRTCManager::offerCreated,
            this, &MeetingService::onWebRTCOfferCreated);

    connect(m_webrtcManager, &WebRTCManager::answerCreated,
            this, &MeetingService::onWebRTCAnswerCreated);

    connect(m_webrtcManager, &WebRTCManager::iceCandidateGenerated,
            this, &MeetingService::onWebRTCIceCandidateGenerated);

    connect(m_webrtcManager, &WebRTCManager::connectionStateChanged,
            this, &MeetingService::onWebRTCConnectionStateChanged);

    connect(m_webrtcManager, &WebRTCManager::error,
            this, &MeetingService::onWebRTCError);

    LOG_INFO("WebRTC connections setup completed");
}

void MeetingService::onWebSocketConnected()
{
    LOG_INFO("WebSocket connected to meeting");
}

void MeetingService::onWebSocketDisconnected()
{
    LOG_INFO("WebSocket disconnected from meeting");
}

void MeetingService::onSignalingMessageReceived(SignalingMessageType type, const QJsonObject &message)
{
    LOG_DEBUG(QString("Received signaling message type: %1").arg(static_cast<int>(type)));

    switch (type) {
        case SignalingMessageType::Offer:
            handleOffer(message);
            break;
        case SignalingMessageType::Answer:
            handleAnswer(message);
            break;
        case SignalingMessageType::IceCandidate:
            handleIceCandidate(message);
            break;
        case SignalingMessageType::JoinRoom:
            handleJoinRoom(message);
            break;
        case SignalingMessageType::LeaveRoom:
            handleLeaveRoom(message);
            break;
        case SignalingMessageType::UserJoined:
            handleUserJoined(message);
            break;
        case SignalingMessageType::UserLeft:
            handleUserLeft(message);
            break;
        case SignalingMessageType::Chat:
            handleChatMessage(message);
            break;
        case SignalingMessageType::ScreenShare:
            handleScreenShare(message);
            break;
        case SignalingMessageType::MediaControl:
            handleMediaControl(message);
            break;
        case SignalingMessageType::RoomInfo:
            handleRoomInfo(message);
            break;
        case SignalingMessageType::Error:
            handleError(message);
            break;
        default:
            LOG_WARNING(QString("Unhandled signaling message type: %1").arg(static_cast<int>(type)));
            break;
    }
}

void MeetingService::onUserJoined(int userId, const QString &username)
{
    LOG_INFO(QString("User joined: %1 (%2)").arg(username).arg(userId));

    auto participant = std::make_unique<Participant>();
    participant->setUserId(userId);
    participant->setUsername(username);
    m_participants.push_back(std::move(participant));

    emit participantJoined(userId, username);
}

void MeetingService::onUserLeft(int userId)
{
    LOG_INFO(QString("User left: %1").arg(userId));

    auto it = std::remove_if(m_participants.begin(), m_participants.end(),
        [userId](const std::unique_ptr<Participant> &p) {
            return p->userId() == userId;
        });
    m_participants.erase(it, m_participants.end());

    emit participantLeft(userId);
}

void MeetingService::onChatMessageReceived(int fromUserId, const QString &username, const QString &content)
{
    LOG_DEBUG(QString("Chat from %1: %2").arg(username, content));

    // 创建消息对象并添加到消息列表
    auto message = std::make_unique<Message>();
    message->setFromUserId(fromUserId);
    message->setFromUsername(username);
    message->setContent(content);
    message->setTimestamp(QDateTime::currentDateTime());
    m_messages.push_back(std::move(message));

    emit chatMessageReceived(fromUserId, username, content);
}

// ==================== 信令消息处理函数 ====================

void MeetingService::handleOffer(const QJsonObject &message)
{
    LOG_DEBUG("Handling WebRTC Offer");

    int fromUserId = message["from_user_id"].toInt();
    QString sdp = message["sdp"].toString();

    if (m_webrtcManager) {
        m_webrtcManager->handleOffer(fromUserId, sdp);
        LOG_INFO(QString("Received and processed Offer from user %1").arg(fromUserId));
    } else {
        LOG_ERROR("WebRTCManager not available");
    }
}

void MeetingService::handleAnswer(const QJsonObject &message)
{
    LOG_DEBUG("Handling WebRTC Answer");

    int fromUserId = message["from_user_id"].toInt();
    QString sdp = message["sdp"].toString();

    if (m_webrtcManager) {
        m_webrtcManager->handleAnswer(fromUserId, sdp);
        LOG_INFO(QString("Received and processed Answer from user %1").arg(fromUserId));
    } else {
        LOG_ERROR("WebRTCManager not available");
    }
}

void MeetingService::handleIceCandidate(const QJsonObject &message)
{
    LOG_DEBUG("Handling ICE Candidate");

    int fromUserId = message["from_user_id"].toInt();
    QString candidate = message["candidate"].toString();
    QString sdpMid = message["sdp_mid"].toString();
    int sdpMLineIndex = message["sdp_mline_index"].toInt();

    if (m_webrtcManager) {
        m_webrtcManager->handleIceCandidate(fromUserId, candidate, sdpMid, sdpMLineIndex);
        LOG_DEBUG(QString("Received and processed ICE Candidate from user %1").arg(fromUserId));
    } else {
        LOG_ERROR("WebRTCManager not available");
    }
}

void MeetingService::handleJoinRoom(const QJsonObject &message)
{
    LOG_DEBUG("Handling Join Room");

    int userId = message["user_id"].toInt();
    QString username = message["username"].toString();

    LOG_INFO(QString("User %1 (%2) joined the room").arg(username).arg(userId));
}

void MeetingService::handleLeaveRoom(const QJsonObject &message)
{
    LOG_DEBUG("Handling Leave Room");

    int userId = message["user_id"].toInt();

    LOG_INFO(QString("User %1 left the room").arg(userId));
}

void MeetingService::handleUserJoined(const QJsonObject &message)
{
    LOG_DEBUG("Handling User Joined");

    int userId = message["user_id"].toInt();
    QString username = message["username"].toString();
    QString peerId = message["peer_id"].toString();

    // 创建参与者对象
    auto participant = std::make_unique<Participant>();
    participant->setUserId(userId);
    participant->setUsername(username);
    participant->setStatus("online");
    participant->setJoinedAt(QDateTime::currentDateTime());

    m_participants.push_back(std::move(participant));

    // 创建WebRTC PeerConnection
    if (m_webrtcManager) {
        m_webrtcManager->createPeerConnection(userId);
        LOG_INFO(QString("Created PeerConnection for user %1").arg(userId));
    }

    LOG_INFO(QString("User joined: %1 (%2)").arg(username).arg(userId));
    emit participantJoined(userId, username);
    emit participantsListUpdated();
}

void MeetingService::handleUserLeft(const QJsonObject &message)
{
    LOG_DEBUG("Handling User Left");

    int userId = message["user_id"].toInt();

    // 关闭WebRTC PeerConnection
    if (m_webrtcManager) {
        m_webrtcManager->closePeerConnection(userId);
        LOG_INFO(QString("Closed PeerConnection for user %1").arg(userId));
    }

    // 从参与者列表中移除
    auto it = std::remove_if(m_participants.begin(), m_participants.end(),
        [userId](const std::unique_ptr<Participant> &p) {
            return p->userId() == userId;
        });
    m_participants.erase(it, m_participants.end());

    LOG_INFO(QString("User left: %1").arg(userId));
    emit participantLeft(userId);
    emit participantsListUpdated();
}

void MeetingService::handleChatMessage(const QJsonObject &message)
{
    LOG_DEBUG("Handling Chat Message");

    int fromUserId = message["from_user_id"].toInt();
    QString username = message["username"].toString();
    QString content = message["content"].toString();
    QString timestamp = message["timestamp"].toString();

    // 创建消息对象
    auto msg = std::make_unique<Message>();
    msg->setFromUserId(fromUserId);
    msg->setFromUsername(username);
    msg->setContent(content);
    msg->setTimestamp(QDateTime::fromString(timestamp, Qt::ISODate));

    m_messages.push_back(std::move(msg));

    LOG_DEBUG(QString("Chat from %1: %2").arg(username, content));
    emit chatMessageReceived(fromUserId, username, content);
}

void MeetingService::handleScreenShare(const QJsonObject &message)
{
    LOG_DEBUG("Handling Screen Share");

    int userId = message["user_id"].toInt();
    bool enabled = message["enabled"].toBool();

    if (enabled) {
        LOG_INFO(QString("User %1 started screen sharing").arg(userId));
        emit screenShareStarted(userId);
    } else {
        LOG_INFO(QString("User %1 stopped screen sharing").arg(userId));
        emit screenShareStopped(userId);
    }
}

void MeetingService::handleMediaControl(const QJsonObject &message)
{
    LOG_DEBUG("Handling Media Control");

    int userId = message["user_id"].toInt();
    QString mediaType = message["media_type"].toString();
    bool enabled = message["enabled"].toBool();

    // 更新参与者的媒体状态
    for (auto &participant : m_participants) {
        if (participant->userId() == userId) {
            if (mediaType == "audio") {
                participant->setAudioEnabled(enabled);
            } else if (mediaType == "video") {
                participant->setVideoEnabled(enabled);
            }

            QJsonObject data;
            data["media_type"] = mediaType;
            data["enabled"] = enabled;
            emit participantUpdated(userId, data);
            break;
        }
    }

    LOG_INFO(QString("User %1 %2 %3").arg(userId).arg(enabled ? "enabled" : "disabled").arg(mediaType));
    emit mediaControlReceived(userId, mediaType, enabled);
}

void MeetingService::handleRoomInfo(const QJsonObject &message)
{
    LOG_DEBUG("Handling Room Info");

    int meetingId = message["meeting_id"].toInt();
    int participantCount = message["participant_count"].toInt();
    QJsonArray participantsArray = message["participants"].toArray();

    // 清空并重新加载参与者列表
    m_participants.clear();

    for (const QJsonValue &value : participantsArray) {
        QJsonObject pObj = value.toObject();

        auto participant = std::make_unique<Participant>();
        participant->setUserId(pObj["user_id"].toInt());
        participant->setUsername(pObj["username"].toString());
        participant->setRole(pObj["role"].toString());
        participant->setStatus(pObj["status"].toString());
        participant->setAudioEnabled(pObj["audio_enabled"].toBool());
        participant->setVideoEnabled(pObj["video_enabled"].toBool());

        m_participants.push_back(std::move(participant));
    }

    LOG_INFO(QString("Room info received: %1 participants").arg(participantCount));
    emit participantsListUpdated();
}

void MeetingService::handleError(const QJsonObject &message)
{
    LOG_ERROR("Handling Error Message");

    QString errorMessage = message["message"].toString();
    int errorCode = message["code"].toInt();

    LOG_ERROR(QString("WebSocket error [%1]: %2").arg(errorCode).arg(errorMessage));
    emit meetingError(errorMessage);
}



// ==================== WebRTC槽函数 ====================

void MeetingService::onWebRTCOfferCreated(int remoteUserId, const QString &sdp)
{
    LOG_INFO(QString("WebRTC Offer created for user: %1").arg(remoteUserId));

    // 通过WebSocket发送Offer
    m_wsClient->sendOffer(sdp, remoteUserId);
}

void MeetingService::onWebRTCAnswerCreated(int remoteUserId, const QString &sdp)
{
    LOG_INFO(QString("WebRTC Answer created for user: %1").arg(remoteUserId));

    // 通过WebSocket发送Answer
    m_wsClient->sendAnswer(sdp, remoteUserId);
}

void MeetingService::onWebRTCIceCandidateGenerated(int remoteUserId, const QString &candidate,
                                                    const QString &sdpMid, int sdpMLineIndex)
{
    LOG_DEBUG(QString("ICE Candidate generated for user: %1").arg(remoteUserId));

    // 通过WebSocket发送ICE候选
    m_wsClient->sendIceCandidate(candidate, sdpMid, sdpMLineIndex, remoteUserId);
}

void MeetingService::onWebRTCConnectionStateChanged(int userId, const QString &state)
{
    LOG_INFO(QString("WebRTC connection state changed for user %1: %2").arg(userId).arg(state));

    // 转发连接状态变化信号
    emit connectionStateChanged(userId, state);

    // 如果连接失败，可以尝试重连
    if (state == "failed" || state == "closed") {
        LOG_WARNING(QString("Connection to user %1 %2, may need reconnection").arg(userId).arg(state));
    }
}

void MeetingService::onWebRTCError(const QString &error)
{
    LOG_ERROR(QString("WebRTC error: %1").arg(error));

    // 转发WebRTC错误
    emit webrtcError(error);
}