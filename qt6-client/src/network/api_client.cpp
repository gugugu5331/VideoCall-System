#include "network/api_client.h"
#include "utils/logger.h"

ApiClient::ApiClient(const QString &baseUrl, QObject *parent)
    : QObject(parent)
    , m_baseUrl(baseUrl)
{
    m_httpClient = new HttpClient(this);
}

ApiClient::~ApiClient()
{
}

void ApiClient::setAuthToken(const QString &token)
{
    m_httpClient->setAuthToken(token);
}

void ApiClient::setCsrfToken(const QString &token)
{
    m_httpClient->setCsrfToken(token);
}

ApiResponse ApiClient::parseResponse(const QJsonObject &response)
{
    ApiResponse apiResponse;
    apiResponse.code = response["code"].toInt();
    apiResponse.message = response["message"].toString();
    apiResponse.data = response["data"].toObject();
    apiResponse.error = response["error"].toString();
    apiResponse.timestamp = response["timestamp"].toString();
    apiResponse.requestId = response["request_id"].toString();
    return apiResponse;
}

QString ApiClient::buildUrl(const QString &endpoint) const
{
    return m_baseUrl + endpoint;
}

// ==================== 认证API ====================

void ApiClient::getCsrfToken(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/csrf-token"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::registerUser(const QString &username, const QString &email,
                            const QString &password, const QString &nickname,
                            std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["username"] = username;
    data["email"] = email;
    data["password"] = password;
    data["nickname"] = nickname;

    m_httpClient->post(buildUrl("/api/v1/auth/register"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::login(const QString &username, const QString &password,
                     std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["username"] = username;
    data["password"] = password;

    m_httpClient->post(buildUrl("/api/v1/auth/login"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::refreshToken(const QString &refreshToken,
                            std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["refresh_token"] = refreshToken;

    m_httpClient->post(buildUrl("/api/v1/auth/refresh"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::forgotPassword(const QString &email,
                              std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["email"] = email;

    m_httpClient->post(buildUrl("/api/v1/auth/forgot-password"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::resetPassword(const QString &token, const QString &newPassword,
                             std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["token"] = token;
    data["new_password"] = newPassword;

    m_httpClient->post(buildUrl("/api/v1/auth/reset-password"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== 用户API ====================

void ApiClient::getUserProfile(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/users/profile"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::updateUserProfile(const QString &nickname, const QString &email,
                                 const QString &avatarUrl,
                                 std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    if (!nickname.isEmpty()) data["nickname"] = nickname;
    if (!email.isEmpty()) data["email"] = email;
    if (!avatarUrl.isEmpty()) data["avatar_url"] = avatarUrl;

    m_httpClient->put(buildUrl("/api/v1/users/profile"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::changePassword(const QString &oldPassword, const QString &newPassword,
                              std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["old_password"] = oldPassword;
    data["new_password"] = newPassword;

    m_httpClient->post(buildUrl("/api/v1/users/change-password"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::uploadAvatar(const QString &filePath,
                            std::function<void(const ApiResponse&)> callback,
                            std::function<void(qint64, qint64)> onProgress)
{
    QVariantMap formData;

    m_httpClient->upload(buildUrl("/api/v1/users/upload-avatar"), filePath, formData,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        },
        onProgress);
}

void ApiClient::deleteAccount(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->del(buildUrl("/api/v1/users/account"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== 会议API ====================

void ApiClient::createMeeting(const QString &title, const QString &description,
                             const QDateTime &startTime, const QDateTime &endTime,
                             int maxParticipants, const QString &meetingType,
                             const QString &password, const QJsonObject &settings,
                             std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["title"] = title;
    data["description"] = description;
    data["start_time"] = startTime.toString(Qt::ISODate);
    data["end_time"] = endTime.toString(Qt::ISODate);
    data["max_participants"] = maxParticipants;
    data["meeting_type"] = meetingType;
    if (!password.isEmpty()) data["password"] = password;
    if (!settings.isEmpty()) data["settings"] = settings;

    m_httpClient->post(buildUrl("/api/v1/meetings"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getMeetingList(int page, int pageSize, const QString &status,
                              const QString &keyword,
                              std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings?page=%2&page_size=%3")
                      .arg(m_baseUrl).arg(page).arg(pageSize);
    if (!status.isEmpty()) url += "&status=" + status;
    if (!keyword.isEmpty()) url += "&keyword=" + keyword;

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getMeetingInfo(int meetingId,
                              std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2").arg(m_baseUrl).arg(meetingId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::updateMeeting(int meetingId, const QJsonObject &updateData,
                             std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2").arg(m_baseUrl).arg(meetingId);

    m_httpClient->put(url, updateData,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::deleteMeeting(int meetingId,
                             std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2").arg(m_baseUrl).arg(meetingId);

    m_httpClient->del(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::startMeeting(int meetingId,
                            std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/start").arg(m_baseUrl).arg(meetingId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::endMeeting(int meetingId,
                          std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/end").arg(m_baseUrl).arg(meetingId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::joinMeeting(int meetingId, const QString &password,
                           std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/join").arg(m_baseUrl).arg(meetingId);

    QJsonObject data;
    if (!password.isEmpty()) {
        data["password"] = password;
    }

    m_httpClient->post(url, data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::leaveMeeting(int meetingId,
                            std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/leave").arg(m_baseUrl).arg(meetingId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getParticipants(int meetingId,
                               std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/participants").arg(m_baseUrl).arg(meetingId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::addParticipant(int meetingId, int userId, const QString &role,
                              std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/participants").arg(m_baseUrl).arg(meetingId);

    QJsonObject data;
    data["user_id"] = userId;
    if (!role.isEmpty()) {
        data["role"] = role;
    }

    m_httpClient->post(url, data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}



void ApiClient::removeParticipant(int meetingId, int userId,
                                 std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/participants/%3")
                      .arg(m_baseUrl).arg(meetingId).arg(userId);

    m_httpClient->del(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::updateParticipantRole(int meetingId, int userId, const QString &role,
                                     std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/participants/%3/role")
                      .arg(m_baseUrl).arg(meetingId).arg(userId);

    QJsonObject data;
    data["role"] = role;

    m_httpClient->put(url, data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::startRecording(int meetingId,
                               std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/recording/start")
                      .arg(m_baseUrl).arg(meetingId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::stopRecording(int meetingId,
                              std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/recording/stop")
                      .arg(m_baseUrl).arg(meetingId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getRecordings(int meetingId,
                             std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/recordings")
                      .arg(m_baseUrl).arg(meetingId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getChatMessages(int meetingId, int page, int pageSize,
                               std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/messages?page=%3&page_size=%4")
                      .arg(m_baseUrl).arg(meetingId).arg(page).arg(pageSize);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::sendChatMessage(int meetingId, const QString &content,
                               std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/meetings/%2/messages")
                      .arg(m_baseUrl).arg(meetingId);

    QJsonObject data;
    data["content"] = content;

    m_httpClient->post(url, data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== 我的会议API ====================

void ApiClient::getMyMeetings(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/my/meetings"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getUpcomingMeetings(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/my/meetings/upcoming"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getMeetingHistory(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/my/meetings/history"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== 媒体API ====================

void ApiClient::uploadMedia(const QString &filePath, int userId, int meetingId,
                           std::function<void(const ApiResponse&)> callback,
                           std::function<void(qint64, qint64)> onProgress)
{
    QVariantMap formData;
    formData["user_id"] = userId;
    formData["meeting_id"] = meetingId;

    m_httpClient->upload(buildUrl("/api/v1/media/upload"), filePath, formData,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        },
        onProgress);
}

void ApiClient::downloadMedia(int mediaId, const QString &savePath,
                             std::function<void(const ApiResponse&)> callback,
                             std::function<void(qint64, qint64)> onProgress)
{
    QString url = QString("%1/api/v1/media/download/%2").arg(m_baseUrl).arg(mediaId);

    // TODO: Implement file download with progress tracking
    // For now, just call the callback with an error
    ApiResponse apiResponse;
    apiResponse.code = 501;
    apiResponse.message = "Download not implemented yet";
    callback(apiResponse);
}

void ApiClient::getMediaList(int meetingId,
                            std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/media?meeting_id=%2").arg(m_baseUrl).arg(meetingId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getMediaInfo(int mediaId,
                            std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/media/info/%2").arg(m_baseUrl).arg(mediaId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::deleteMedia(int mediaId,
                           std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/media/%2").arg(m_baseUrl).arg(mediaId);

    m_httpClient->del(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::processMedia(int mediaId, const QString &processType, const QJsonObject &params,
                            std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["media_id"] = mediaId;
    data["process_type"] = processType;
    data["params"] = params;

    m_httpClient->post(buildUrl("/api/v1/media/process"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== AI服务API ====================

void ApiClient::speechRecognition(const QByteArray &audioData, const QString &audioFormat,
                                 int sampleRate, const QString &language,
                                 int userId,
                                 std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["user_id"] = userId;  // 添加用户ID
    data["audio_data"] = QString(audioData.toBase64());
    data["audio_format"] = audioFormat;
    data["sample_rate"] = sampleRate;
    data["language"] = language;

    m_httpClient->post(buildUrl("/api/v1/speech/recognition"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::emotionDetection(const QByteArray &audioData, const QString &audioFormat,
                                int sampleRate,
                                int userId,
                                std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["user_id"] = userId;  // 添加用户ID
    data["audio_data"] = QString(audioData.toBase64());
    data["audio_format"] = audioFormat;
    data["sample_rate"] = sampleRate;

    m_httpClient->post(buildUrl("/api/v1/speech/emotion"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}


void ApiClient::synthesisDetection(const QByteArray &videoData,
                                  int userId,
                                  std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["user_id"] = userId;  // 添加用户ID
    data["video_data"] = QString(videoData.toBase64());  // 修正：应该是video_data

    m_httpClient->post(buildUrl("/api/v1/speech/synthesis-detection"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::audioDenoising(const QByteArray &audioData,
                              std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["audio_data"] = QString(audioData.toBase64());

    m_httpClient->post(buildUrl("/api/v1/audio/denoising"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::videoEnhancement(const QByteArray &videoData, const QString &enhancementType,
                                std::function<void(const ApiResponse&)> callback)
{
    QJsonObject data;
    data["video_data"] = QString(videoData.toBase64());
    data["enhancement_type"] = enhancementType;

    m_httpClient->post(buildUrl("/api/v1/video/enhancement"), data,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// AI模型管理API
void ApiClient::getAIModels(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/models"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::loadAIModel(const QString &modelId, std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/models/%2/load").arg(m_baseUrl).arg(modelId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::unloadAIModel(const QString &modelId, std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/models/%2/unload").arg(m_baseUrl).arg(modelId);

    m_httpClient->del(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getAIModelStatus(const QString &modelId, std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/models/%2/status").arg(m_baseUrl).arg(modelId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getAINodes(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/nodes"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::checkAINodeHealth(const QString &nodeId, std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/nodes/%2/health-check").arg(m_baseUrl).arg(nodeId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getLoadBalancerStats(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/load-balancer/stats"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getMonitoringMetrics(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/monitoring/metrics"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== 信令服务API ====================

void ApiClient::getSessionInfo(const QString &sessionId,
                               std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/sessions/%2").arg(m_baseUrl).arg(sessionId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getRoomSessions(int meetingId,
                               std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/sessions/room/%2").arg(m_baseUrl).arg(meetingId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getMessageHistory(int meetingId,
                                  std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/messages/history/%2").arg(m_baseUrl).arg(meetingId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getStatsOverview(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/stats/overview"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getRoomStats(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/stats/rooms"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== WebRTC服务API ====================

void ApiClient::getRoomPeers(int roomId,
                            std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/webrtc/room/%2/peers").arg(m_baseUrl).arg(roomId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getRoomWebRTCStats(int roomId,
                                   std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/webrtc/room/%2/stats").arg(m_baseUrl).arg(roomId);

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::updatePeerMedia(const QString &peerId, const QJsonObject &mediaState,
                               std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/webrtc/peer/%2/media").arg(m_baseUrl).arg(peerId);

    m_httpClient->post(url, mediaState,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

// ==================== 管理员API ====================

void ApiClient::getAdminUsers(int page, int pageSize, const QString &keyword,
                             std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/admin/users?page=%2&page_size=%3")
                      .arg(m_baseUrl).arg(page).arg(pageSize);
    if (!keyword.isEmpty()) {
        url += "&keyword=" + keyword;
    }

    m_httpClient->get(url,
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::banUser(int userId, std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/admin/users/%2/ban").arg(m_baseUrl).arg(userId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getAdminMeetings(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/admin/meetings"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::getAdminMeetingStats(std::function<void(const ApiResponse&)> callback)
{
    m_httpClient->get(buildUrl("/api/v1/admin/meetings/stats"),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

void ApiClient::forceEndMeeting(int meetingId, std::function<void(const ApiResponse&)> callback)
{
    QString url = QString("%1/api/v1/admin/meetings/%2/force-end").arg(m_baseUrl).arg(meetingId);

    m_httpClient->post(url, QJsonObject(),
        [this, callback](const QJsonObject &response) {
            callback(parseResponse(response));
        },
        [callback](const QString &error) {
            ApiResponse apiResponse;
            apiResponse.code = 500;
            apiResponse.message = error;
            callback(apiResponse);
        });
}

