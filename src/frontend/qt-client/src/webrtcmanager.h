#ifndef WEBRTCMANAGER_H
#define WEBRTCMANAGER_H

#include <QtCore/QObject>
#include <QtCore/QMap>
#include <QtCore/QTimer>
#include <QtMultimedia/QCamera>
#include <QtMultimedia/QMediaCaptureSession>
#include <QtMultimedia/QMediaPlayer>
#include <QtMultimedia/QAudioInput>
#include <QtMultimedia/QAudioOutput>
#include <QtMultimedia/QMediaDevices>

class SignalingClient;

/**
 * WebRTC管理器
 * 负责音视频采集、传输和接收
 * 注意：这是一个简化版本，实际的WebRTC功能需要集成第三方库
 */
class WebRTCManager : public QObject
{
    Q_OBJECT

public:
    explicit WebRTCManager(QObject *parent = nullptr);
    ~WebRTCManager();

    // 初始化和清理
    bool initialize();
    void cleanup();

    // 会议控制
    bool joinMeeting(const QString &meetingId, const QString &userName);
    void leaveMeeting();
    bool isInMeeting() const { return m_isInMeeting; }

    // 媒体控制
    void toggleCamera(bool enabled);
    void toggleMicrophone(bool enabled);
    void toggleScreenShare(bool enabled);

    // 设备管理
    QList<QCameraDevice> getAvailableCameras() const;
    QList<QAudioDevice> getAvailableMicrophones() const;
    QList<QAudioDevice> getAvailableSpeakers() const;
    
    void setCamera(const QCameraDevice &device);
    void setMicrophone(const QAudioDevice &device);
    void setSpeaker(const QAudioDevice &device);

    // 获取媒体源
    QObject* getLocalVideoSource() const;
    QObject* getLocalAudioSource() const;

    // 状态查询
    bool isCameraEnabled() const { return m_isCameraEnabled; }
    bool isMicrophoneEnabled() const { return m_isMicrophoneEnabled; }
    bool isScreenSharing() const { return m_isScreenSharing; }

    // 设置信令客户端
    void setSignalingClient(SignalingClient *client);

signals:
    // 本地流事件
    void localStreamReady();
    void localStreamStopped();
    
    // 远程流事件
    void remoteStreamReceived(const QString &userId, QObject *stream);
    void remoteStreamRemoved(const QString &userId);
    
    // 连接事件
    void peerConnected(const QString &userId);
    void peerDisconnected(const QString &userId);
    
    // 错误事件
    void error(const QString &message);

private slots:
    void onCameraStateChanged();
    void onAudioInputStateChanged();
    void updateMediaStats();
    void simulateRemoteStream();

private:
    void setupLocalMedia();
    void cleanupLocalMedia();
    void createPeerConnection(const QString &userId);
    void removePeerConnection(const QString &userId);
    void startMediaCapture();
    void stopMediaCapture();

    // 信令客户端
    SignalingClient *m_signalingClient;

    // 会议状态
    bool m_isInMeeting;
    QString m_currentMeetingId;
    QString m_currentUserName;

    // 媒体设备
    QCamera *m_camera;
    QMediaCaptureSession *m_captureSession;
    QAudioInput *m_audioInput;
    QAudioOutput *m_audioOutput;
    
    // 媒体状态
    bool m_isCameraEnabled;
    bool m_isMicrophoneEnabled;
    bool m_isScreenSharing;

    // 设备列表
    QCameraDevice m_currentCameraDevice;
    QAudioDevice m_currentMicrophoneDevice;
    QAudioDevice m_currentSpeakerDevice;

    // 远程流管理
    QMap<QString, QMediaPlayer*> m_remoteStreams;
    
    // 统计定时器
    QTimer *m_statsTimer;
    
    // 模拟定时器（用于演示）
    QTimer *m_simulationTimer;
};

#endif // WEBRTCMANAGER_H
