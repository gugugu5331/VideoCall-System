#ifndef MEDIA_SERVICE_H
#define MEDIA_SERVICE_H

#include <QObject>
#include <QString>
#include "network/api_client.h"

class MediaService : public QObject
{
    Q_OBJECT

public:
    explicit MediaService(ApiClient *apiClient, QObject *parent = nullptr);
    ~MediaService();

    // File operations
    Q_INVOKABLE void uploadFile(const QString &filePath, int userId, int meetingId);
    Q_INVOKABLE void downloadFile(const QString &fileId, const QString &savePath);

    // Recording operations
    Q_INVOKABLE void startRecording(int meetingId);
    Q_INVOKABLE void stopRecording(int meetingId);
    Q_INVOKABLE void getRecordings(int meetingId);

signals:
    void uploadStarted(const QString &filePath);
    void uploadProgress(qint64 bytesSent, qint64 bytesTotal);
    void uploadFinished(const QString &fileUrl);
    void uploadFailed(const QString &error);
    
    void downloadStarted(const QString &fileId);
    void downloadProgress(qint64 bytesReceived, qint64 bytesTotal);
    void downloadFinished(const QString &savePath);
    void downloadFailed(const QString &error);
    
    void recordingStarted();
    void recordingStopped();
    void recordingError(const QString &error);

private:
    ApiClient *m_apiClient;
};

#endif // MEDIA_SERVICE_H

