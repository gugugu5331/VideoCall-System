#include "services/media_service.h"
#include "utils/logger.h"

MediaService::MediaService(ApiClient *apiClient, QObject *parent)
    : QObject(parent)
    , m_apiClient(apiClient)
{
}

MediaService::~MediaService()
{
}

void MediaService::uploadFile(const QString &filePath, int userId, int meetingId)
{
    LOG_INFO("Uploading file: " + filePath);

    emit uploadStarted(filePath);

    m_apiClient->uploadMedia(filePath, userId, meetingId,
        [this, filePath](const ApiResponse &response) {
            if (response.isSuccess()) {
                QString fileUrl = response.data["file_url"].toString();
                LOG_INFO("File uploaded successfully: " + fileUrl);
                emit uploadFinished(fileUrl);
            } else {
                LOG_ERROR("File upload failed: " + response.message);
                emit uploadFailed(response.message);
            }
        },
        [this](qint64 sent, qint64 total) {
            emit uploadProgress(sent, total);
        });
}

void MediaService::downloadFile(const QString &fileId, const QString &savePath)
{
    LOG_INFO("Downloading file: " + fileId);

    emit downloadStarted(fileId);

    int mediaId = fileId.toInt();
    m_apiClient->downloadMedia(mediaId, savePath,
        [this, savePath](const ApiResponse &response) {
            if (response.isSuccess()) {
                LOG_INFO("File downloaded successfully: " + savePath);
                emit downloadFinished(savePath);
            } else {
                LOG_ERROR("File download failed: " + response.message);
                emit downloadFailed(response.message);
            }
        },
        [this](qint64 received, qint64 total) {
            emit downloadProgress(received, total);
        });
}

void MediaService::startRecording(int meetingId)
{
    LOG_INFO("Starting recording for meeting: " + QString::number(meetingId));

    m_apiClient->startRecording(meetingId,
        [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                LOG_INFO("Recording started successfully");
                emit recordingStarted();
            } else {
                LOG_ERROR("Failed to start recording: " + response.message);
                emit recordingError(response.message);
            }
        });
}

void MediaService::stopRecording(int meetingId)
{
    LOG_INFO("Stopping recording for meeting: " + QString::number(meetingId));

    m_apiClient->stopRecording(meetingId,
        [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                LOG_INFO("Recording stopped successfully");
                emit recordingStopped();
            } else {
                LOG_ERROR("Failed to stop recording: " + response.message);
                emit recordingError(response.message);
            }
        });
}

void MediaService::getRecordings(int meetingId)
{
    LOG_INFO("Getting recordings for meeting: " + QString::number(meetingId));

    m_apiClient->getRecordings(meetingId,
        [this](const ApiResponse &response) {
            if (response.isSuccess()) {
                // TODO: Parse and emit recordings list
                LOG_INFO("Recordings retrieved successfully");
            } else {
                LOG_ERROR("Failed to get recordings: " + response.message);
                emit recordingError(response.message);
            }
        });
}

