#ifndef MEETING_SERVICE_H
#define MEETING_SERVICE_H

#include <QObject>
#include <QList>
#include <QDateTime>
#include <memory>
#include <vector>
#include "network/api_client.h"
#include "network/websocket_client.h"
#include "models/meeting.h"
#include "models/participant.h"
#include "models/message.h"

// Forward declarations
class WebRTCManager;
class MediaStream;

class MeetingService : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool audioEnabled READ audioEnabled WRITE setAudioEnabled NOTIFY audioEnabledChanged)
    Q_PROPERTY(bool videoEnabled READ videoEnabled WRITE setVideoEnabled NOTIFY videoEnabledChanged)
    Q_PROPERTY(bool isScreenSharing READ isScreenSharing WRITE setIsScreenSharing NOTIFY isScreenSharingChanged)

public:
    explicit MeetingService(ApiClient *apiClient, WebSocketClient *wsClient, QObject *parent = nullptr);
    ~MeetingService();

    // Current meeting
    Meeting* currentMeeting() const { return m_currentMeeting.get(); }
    const std::vector<std::unique_ptr<Participant>>& participants() const { return m_participants; }
    const std::vector<std::unique_ptr<Message>>& messages() const { return m_messages; }

    // Media state
    bool audioEnabled() const { return m_audioEnabled; }
    bool videoEnabled() const { return m_videoEnabled; }
    bool isScreenSharing() const { return m_isScreenSharing; }

    // Meeting operations
    Q_INVOKABLE void createMeeting(const QString &title, const QString &description,
                                   const QDateTime &startTime, const QDateTime &endTime,
                                   int maxParticipants, const QString &meetingType,
                                   const QString &password, const QJsonObject &settings);
    Q_INVOKABLE void joinMeeting(int meetingId, const QString &password = QString());
    Q_INVOKABLE void leaveMeeting();
    Q_INVOKABLE void getMeetingList();
    Q_INVOKABLE void getMeetingInfo(int meetingId);
    Q_INVOKABLE void startMeeting(int meetingId);
    Q_INVOKABLE void endMeeting(int meetingId);

    // Participant operations
    Q_INVOKABLE void getParticipants(int meetingId);
    Q_INVOKABLE void addParticipant(int meetingId, int userId, const QString &role);
    Q_INVOKABLE void removeParticipant(int meetingId, int userId);
    Q_INVOKABLE void updateParticipantRole(int meetingId, int userId, const QString &role);
    Q_INVOKABLE void kickParticipant(int userId);
    Q_INVOKABLE void muteParticipant(int userId, bool mute);

    // Chat operations
    Q_INVOKABLE void sendChatMessage(const QString &message);
    Q_INVOKABLE void getChatMessages(int meetingId);

    // Media control
    Q_INVOKABLE void toggleAudio();
    Q_INVOKABLE void toggleVideo();
    Q_INVOKABLE void startScreenShare();
    Q_INVOKABLE void stopScreenShare();

    // WebRTC operations
    Q_INVOKABLE void startLocalMedia(bool audio = true, bool video = true);
    Q_INVOKABLE void stopLocalMedia();
    Q_INVOKABLE MediaStream* getLocalStream() const;
    Q_INVOKABLE QVariantMap getConnectionStatistics(int userId) const;

    // Setters
    void setAudioEnabled(bool enabled);
    void setVideoEnabled(bool enabled);
    void setIsScreenSharing(bool sharing);

signals:
    void audioEnabledChanged();
    void videoEnabledChanged();
    void isScreenSharingChanged();

    void meetingCreated(Meeting *meeting);
    void meetingJoined(Meeting *meeting);
    void meetingLeft();
    void meetingError(const QString &error);
    void meetingListUpdated();
    void meetingInfoReceived(Meeting *meeting);

    void participantJoined(int userId, const QString &username);
    void participantLeft(int userId);
    void participantUpdated(int userId, const QJsonObject &data);
    void participantsListUpdated();

    void chatMessageReceived(int fromUserId, const QString &username, const QString &content);
    void chatMessagesLoaded();

    void screenShareStarted(int userId);
    void screenShareStopped(int userId);

    void mediaControlReceived(int userId, const QString &mediaType, bool enabled);

    // WebRTC signals
    void localStreamReady(MediaStream *stream);
    void localStreamStopped();
    void remoteStreamAdded(int userId, MediaStream *stream);
    void remoteStreamRemoved(int userId);
    void connectionStateChanged(int userId, const QString &state);
    void webrtcError(const QString &error);

private slots:
    void onWebSocketConnected();
    void onWebSocketDisconnected();
    void onSignalingMessageReceived(SignalingMessageType type, const QJsonObject &message);
    void onUserJoined(int userId, const QString &username);
    void onUserLeft(int userId);
    void onChatMessageReceived(int fromUserId, const QString &username, const QString &content);

    // WebRTC slots
    void onWebRTCOfferCreated(int remoteUserId, const QString &sdp);
    void onWebRTCAnswerCreated(int remoteUserId, const QString &sdp);
    void onWebRTCIceCandidateGenerated(int remoteUserId, const QString &candidate,
                                       const QString &sdpMid, int sdpMLineIndex);
    void onWebRTCConnectionStateChanged(int userId, const QString &state);
    void onWebRTCError(const QString &error);

private:
    void setupWebSocketConnections();
    void setupWebRTCConnections();

    void handleOffer(const QJsonObject &message);
    void handleAnswer(const QJsonObject &message);
    void handleIceCandidate(const QJsonObject &message);
    void handleJoinRoom(const QJsonObject &message);
    void handleLeaveRoom(const QJsonObject &message);
    void handleUserJoined(const QJsonObject &message);
    void handleUserLeft(const QJsonObject &message);
    void handleChatMessage(const QJsonObject &message);
    void handleScreenShare(const QJsonObject &message);
    void handleMediaControl(const QJsonObject &message);
    void handleRoomInfo(const QJsonObject &message);
    void handleError(const QJsonObject &message);

private:
    ApiClient *m_apiClient;
    WebSocketClient *m_wsClient;
    WebRTCManager *m_webrtcManager;

    std::unique_ptr<Meeting> m_currentMeeting;
    std::vector<std::unique_ptr<Meeting>> m_meetingList;
    std::vector<std::unique_ptr<Participant>> m_participants;
    std::vector<std::unique_ptr<Message>> m_messages;

    bool m_audioEnabled;
    bool m_videoEnabled;
    bool m_isScreenSharing;
};

#endif // MEETING_SERVICE_H

