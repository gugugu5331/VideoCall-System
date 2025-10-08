#ifndef API_CLIENT_H
#define API_CLIENT_H

#include <QObject>
#include <QJsonObject>
#include <QDateTime>
#include <functional>
#include "http_client.h"

// API response structure
struct ApiResponse {
    int code;
    QString message;
    QJsonObject data;
    QString error;
    QString timestamp;
    QString requestId;

    bool isSuccess() const { return code >= 200 && code < 300; }
    bool isClientError() const { return code >= 400 && code < 500; }
    bool isServerError() const { return code >= 500; }
};

class ApiClient : public QObject
{
    Q_OBJECT

public:
    explicit ApiClient(const QString &baseUrl, QObject *parent = nullptr);
    ~ApiClient();

    // Set auth token
    void setAuthToken(const QString &token);

    // Set CSRF token
    void setCsrfToken(const QString &token);

    QString baseUrl() const { return m_baseUrl; }

    // ==================== 认证API ====================

    // GET /api/v1/csrf-token
    void getCsrfToken(std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/auth/register
    void registerUser(const QString &username, const QString &email,
                     const QString &password, const QString &nickname,
                     std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/auth/login
    void login(const QString &username, const QString &password,
              std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/auth/refresh
    void refreshToken(const QString &refreshToken,
                     std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/auth/forgot-password
    void forgotPassword(const QString &email,
                       std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/auth/reset-password
    void resetPassword(const QString &token, const QString &newPassword,
                      std::function<void(const ApiResponse&)> callback);

    // ==================== 用户API ====================

    // GET /api/v1/users/profile
    void getUserProfile(std::function<void(const ApiResponse&)> callback);

    // PUT /api/v1/users/profile
    void updateUserProfile(const QString &nickname, const QString &email,
                          const QString &avatarUrl,
                          std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/users/change-password
    void changePassword(const QString &oldPassword, const QString &newPassword,
                       std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/users/upload-avatar
    void uploadAvatar(const QString &filePath,
                     std::function<void(const ApiResponse&)> callback,
                     std::function<void(qint64, qint64)> onProgress = nullptr);

    // DELETE /api/v1/users/account
    void deleteAccount(std::function<void(const ApiResponse&)> callback);

    // ==================== 会议API ====================

    // POST /api/v1/meetings
    void createMeeting(const QString &title, const QString &description,
                      const QDateTime &startTime, const QDateTime &endTime,
                      int maxParticipants, const QString &meetingType,
                      const QString &password, const QJsonObject &settings,
                      std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/meetings
    void getMeetingList(int page, int pageSize, const QString &status,
                       const QString &keyword,
                       std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/meetings/:id
    void getMeetingInfo(int meetingId,
                       std::function<void(const ApiResponse&)> callback);

    // PUT /api/v1/meetings/:id
    void updateMeeting(int meetingId, const QJsonObject &updateData,
                      std::function<void(const ApiResponse&)> callback);

    // DELETE /api/v1/meetings/:id
    void deleteMeeting(int meetingId,
                      std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/start
    void startMeeting(int meetingId,
                     std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/end
    void endMeeting(int meetingId,
                   std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/join
    void joinMeeting(int meetingId, const QString &password,
                    std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/leave
    void leaveMeeting(int meetingId,
                     std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/meetings/:id/participants
    void getParticipants(int meetingId,
                        std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/participants
    void addParticipant(int meetingId, int userId, const QString &role,
                       std::function<void(const ApiResponse&)> callback);

    // DELETE /api/v1/meetings/:id/participants/:user_id
    void removeParticipant(int meetingId, int userId,
                          std::function<void(const ApiResponse&)> callback);

    // PUT /api/v1/meetings/:id/participants/:user_id/role
    void updateParticipantRole(int meetingId, int userId, const QString &role,
                              std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/recording/start
    void startRecording(int meetingId,
                       std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/recording/stop
    void stopRecording(int meetingId,
                      std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/meetings/:id/recordings
    void getRecordings(int meetingId,
                      std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/meetings/:id/messages
    void getChatMessages(int meetingId, int page, int pageSize,
                        std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/meetings/:id/messages
    void sendChatMessage(int meetingId, const QString &content,
                        std::function<void(const ApiResponse&)> callback);

    // ==================== 我的会议API ====================

    // GET /api/v1/my/meetings
    void getMyMeetings(std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/my/meetings/upcoming
    void getUpcomingMeetings(std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/my/meetings/history
    void getMeetingHistory(std::function<void(const ApiResponse&)> callback);

    // ==================== 媒体API ====================

    // POST /api/v1/media/upload
    void uploadMedia(const QString &filePath, int userId, int meetingId,
                    std::function<void(const ApiResponse&)> callback,
                    std::function<void(qint64, qint64)> onProgress = nullptr);

    // GET /api/v1/media/download/:id
    void downloadMedia(int mediaId, const QString &savePath,
                      std::function<void(const ApiResponse&)> callback,
                      std::function<void(qint64, qint64)> onProgress = nullptr);

    // GET /api/v1/media
    void getMediaList(int meetingId,
                     std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/media/info/:id
    void getMediaInfo(int mediaId,
                     std::function<void(const ApiResponse&)> callback);

    // DELETE /api/v1/media/:id
    void deleteMedia(int mediaId,
                    std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/media/process
    void processMedia(int mediaId, const QString &processType, const QJsonObject &params,
                     std::function<void(const ApiResponse&)> callback);

    // ==================== AI服务API ====================

    // POST /api/v1/speech/recognition
    void speechRecognition(const QByteArray &audioData, const QString &audioFormat,
                          int sampleRate, const QString &language,
                          int userId,  // 新增：远程用户ID
                          std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/speech/emotion
    void emotionDetection(const QByteArray &audioData, const QString &audioFormat,
                         int sampleRate,
                         int userId,  // 新增：远程用户ID
                         std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/speech/synthesis-detection
    void synthesisDetection(const QByteArray &videoData,
                           int userId,  // 新增：远程用户ID
                           std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/audio/denoising
    void audioDenoising(const QByteArray &audioData,
                       std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/video/enhancement
    void videoEnhancement(const QByteArray &videoData, const QString &enhancementType,
                         std::function<void(const ApiResponse&)> callback);

    // AI模型管理API
    // GET /api/v1/models
    void getAIModels(std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/models/:model_id/load
    void loadAIModel(const QString &modelId, std::function<void(const ApiResponse&)> callback);

    // DELETE /api/v1/models/:model_id/unload
    void unloadAIModel(const QString &modelId, std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/models/:model_id/status
    void getAIModelStatus(const QString &modelId, std::function<void(const ApiResponse&)> callback);

    // AI节点管理API
    // GET /api/v1/nodes
    void getAINodes(std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/nodes/:node_id/health-check
    void checkAINodeHealth(const QString &nodeId, std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/load-balancer/stats
    void getLoadBalancerStats(std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/monitoring/metrics
    void getMonitoringMetrics(std::function<void(const ApiResponse&)> callback);

    // ==================== 信令服务API ====================

    // GET /api/v1/sessions/:session_id
    void getSessionInfo(const QString &sessionId,
                       std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/sessions/room/:meeting_id
    void getRoomSessions(int meetingId,
                        std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/messages/history/:meeting_id
    void getMessageHistory(int meetingId,
                          std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/stats/overview
    void getStatsOverview(std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/stats/rooms
    void getRoomStats(std::function<void(const ApiResponse&)> callback);

    // ==================== WebRTC服务API ====================

    // GET /api/v1/webrtc/room/:roomId/peers
    void getRoomPeers(int roomId,
                     std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/webrtc/room/:roomId/stats
    void getRoomWebRTCStats(int roomId,
                           std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/webrtc/peer/:peerId/media
    void updatePeerMedia(const QString &peerId, const QJsonObject &mediaState,
                        std::function<void(const ApiResponse&)> callback);

    // ==================== 管理员API ====================

    // GET /api/v1/admin/users
    void getAdminUsers(int page, int pageSize, const QString &keyword,
                      std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/admin/users/:id/ban
    void banUser(int userId, std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/admin/meetings
    void getAdminMeetings(std::function<void(const ApiResponse&)> callback);

    // GET /api/v1/admin/meetings/stats
    void getAdminMeetingStats(std::function<void(const ApiResponse&)> callback);

    // POST /api/v1/admin/meetings/:id/force-end
    void forceEndMeeting(int meetingId, std::function<void(const ApiResponse&)> callback);

signals:
    void requestStarted(const QString &endpoint);
    void requestFinished(const QString &endpoint);
    void requestError(const QString &endpoint, const QString &error);

private:
    ApiResponse parseResponse(const QJsonObject &response);
    QString buildUrl(const QString &endpoint) const;

private:
    HttpClient *m_httpClient;
    QString m_baseUrl;
};

#endif // API_CLIENT_H

