#ifndef WEBRTC_SERVICE_H
#define WEBRTC_SERVICE_H

#include <QObject>
#include <QWebSocket>
#include <QJsonObject>
#include <QJsonDocument>
#include <QTimer>
#include <QVideoFrame>
#include <QAudioBuffer>
#include <QCamera>
#include <QAudioInput>
#include <QMediaDevices>
#include <memory>

// WebRTC前向声明（需要WebRTC库）
namespace webrtc {
    class PeerConnectionInterface;
    class PeerConnectionFactoryInterface;
    class CreateSessionDescriptionObserver;
    class SetSessionDescriptionObserver;
    class DataChannelInterface;
    class MediaStreamInterface;
    class VideoTrackInterface;
    class AudioTrackInterface;
}

/**
 * @brief WebRTC连接状态
 */
enum class WebRTCState {
    Disconnected,
    Connecting,
    Connected,
    Failed,
    Closed
};

/**
 * @brief ICE连接状态
 */
enum class IceConnectionState {
    New,
    Checking,
    Connected,
    Completed,
    Failed,
    Disconnected,
    Closed
};

/**
 * @brief 媒体流类型
 */
enum class MediaStreamType {
    Audio,
    Video,
    Screen
};

/**
 * @brief 参与者信息
 */
struct Participant {
    QString id;
    QString name;
    bool audioEnabled;
    bool videoEnabled;
    bool screenSharing;
    WebRTCState connectionState;
    
    Participant() : audioEnabled(false), videoEnabled(false), 
                   screenSharing(false), connectionState(WebRTCState::Disconnected) {}
};

/**
 * @brief WebRTC服务类
 * 
 * 负责WebRTC连接管理、音视频传输、信令处理等
 */
class WebRTCService : public QObject
{
    Q_OBJECT
    Q_PROPERTY(WebRTCState connectionState READ connectionState NOTIFY connectionStateChanged)
    Q_PROPERTY(bool audioEnabled READ isAudioEnabled WRITE setAudioEnabled NOTIFY audioEnabledChanged)
    Q_PROPERTY(bool videoEnabled READ isVideoEnabled WRITE setVideoEnabled NOTIFY videoEnabledChanged)
    Q_PROPERTY(bool screenSharing READ isScreenSharing NOTIFY screenSharingChanged)
    Q_PROPERTY(QStringList availableCameras READ availableCameras NOTIFY availableCamerasChanged)
    Q_PROPERTY(QStringList availableMicrophones READ availableMicrophones NOTIFY availableMicrophonesChanged)

public:
    explicit WebRTCService(QObject *parent = nullptr);
    ~WebRTCService();

    // 属性访问器
    WebRTCState connectionState() const { return m_connectionState; }
    bool isAudioEnabled() const { return m_audioEnabled; }
    bool isVideoEnabled() const { return m_videoEnabled; }
    bool isScreenSharing() const { return m_screenSharing; }
    QStringList availableCameras() const { return m_availableCameras; }
    QStringList availableMicrophones() const { return m_availableMicrophones; }

    /**
     * @brief 初始化WebRTC
     */
    bool initialize();

    /**
     * @brief 清理资源
     */
    void cleanup();

    /**
     * @brief 连接到信令服务器
     * @param url 信令服务器URL
     * @param meetingId 会议ID
     * @param userId 用户ID
     * @param token 认证令牌
     */
    void connectToSignalingServer(const QString &url, const QString &meetingId, 
                                 const QString &userId, const QString &token);

    /**
     * @brief 断开信令服务器连接
     */
    void disconnectFromSignalingServer();

    /**
     * @brief 创建对等连接
     * @param participantId 参与者ID
     */
    void createPeerConnection(const QString &participantId);

    /**
     * @brief 关闭对等连接
     * @param participantId 参与者ID
     */
    void closePeerConnection(const QString &participantId);

    /**
     * @brief 获取参与者列表
     */
    QList<Participant> getParticipants() const;

    /**
     * @brief 获取本地视频帧
     */
    QVideoFrame getLocalVideoFrame() const;

    /**
     * @brief 获取远程视频帧
     * @param participantId 参与者ID
     */
    QVideoFrame getRemoteVideoFrame(const QString &participantId) const;

public slots:
    /**
     * @brief 设置音频启用状态
     */
    void setAudioEnabled(bool enabled);

    /**
     * @brief 设置视频启用状态
     */
    void setVideoEnabled(bool enabled);

    /**
     * @brief 开始屏幕共享
     */
    void startScreenSharing();

    /**
     * @brief 停止屏幕共享
     */
    void stopScreenSharing();

    /**
     * @brief 切换摄像头
     * @param cameraId 摄像头ID
     */
    void switchCamera(const QString &cameraId);

    /**
     * @brief 切换麦克风
     * @param microphoneId 麦克风ID
     */
    void switchMicrophone(const QString &microphoneId);

    /**
     * @brief 发送数据通道消息
     * @param participantId 参与者ID
     * @param message 消息内容
     */
    void sendDataChannelMessage(const QString &participantId, const QJsonObject &message);

    /**
     * @brief 获取连接统计信息
     * @param participantId 参与者ID
     */
    void getConnectionStats(const QString &participantId);

signals:
    void connectionStateChanged();
    void audioEnabledChanged();
    void videoEnabledChanged();
    void screenSharingChanged();
    void availableCamerasChanged();
    void availableMicrophonesChanged();
    
    void participantJoined(const QString &participantId, const QString &name);
    void participantLeft(const QString &participantId);
    void participantAudioChanged(const QString &participantId, bool enabled);
    void participantVideoChanged(const QString &participantId, bool enabled);
    void participantScreenSharingChanged(const QString &participantId, bool sharing);
    
    void localVideoFrameReady(const QVideoFrame &frame);
    void remoteVideoFrameReady(const QString &participantId, const QVideoFrame &frame);
    void audioLevelChanged(const QString &participantId, float level);
    
    void dataChannelMessageReceived(const QString &participantId, const QJsonObject &message);
    void connectionStatsReceived(const QString &participantId, const QJsonObject &stats);
    
    void error(const QString &message);

private slots:
    void onSignalingConnected();
    void onSignalingDisconnected();
    void onSignalingMessageReceived(const QString &message);
    void onSignalingError(QAbstractSocket::SocketError error);
    
    void onMediaDevicesChanged();
    void onCameraError();
    void onAudioInputError();

private:
    /**
     * @brief 处理信令消息
     */
    void handleSignalingMessage(const QJsonObject &message);

    /**
     * @brief 发送信令消息
     */
    void sendSignalingMessage(const QJsonObject &message);

    /**
     * @brief 创建Offer
     */
    void createOffer(const QString &participantId);

    /**
     * @brief 创建Answer
     */
    void createAnswer(const QString &participantId, const QJsonObject &offer);

    /**
     * @brief 处理ICE候选
     */
    void handleIceCandidate(const QString &participantId, const QJsonObject &candidate);

    /**
     * @brief 初始化媒体设备
     */
    void initializeMediaDevices();

    /**
     * @brief 创建本地媒体流
     */
    void createLocalMediaStream();

    /**
     * @brief 更新可用设备列表
     */
    void updateAvailableDevices();

    /**
     * @brief 设置连接状态
     */
    void setConnectionState(WebRTCState state);

private:
    // WebRTC核心对象
    std::unique_ptr<webrtc::PeerConnectionFactoryInterface> m_peerConnectionFactory;
    QMap<QString, std::unique_ptr<webrtc::PeerConnectionInterface>> m_peerConnections;
    std::unique_ptr<webrtc::MediaStreamInterface> m_localStream;
    
    // 信令
    std::unique_ptr<QWebSocket> m_signalingSocket;
    QString m_signalingUrl;
    QString m_meetingId;
    QString m_userId;
    QString m_authToken;
    
    // 状态
    WebRTCState m_connectionState;
    bool m_audioEnabled;
    bool m_videoEnabled;
    bool m_screenSharing;
    
    // 媒体设备
    std::unique_ptr<QCamera> m_camera;
    std::unique_ptr<QAudioInput> m_audioInput;
    QStringList m_availableCameras;
    QStringList m_availableMicrophones;
    QString m_currentCameraId;
    QString m_currentMicrophoneId;
    
    // 参与者
    QMap<QString, Participant> m_participants;
    
    // 定时器
    QTimer *m_statsTimer;
    QTimer *m_reconnectTimer;
    
    // 配置
    QJsonObject m_iceServers;
    int m_maxRetries;
    int m_reconnectInterval;
};

Q_DECLARE_METATYPE(WebRTCState)
Q_DECLARE_METATYPE(IceConnectionState)
Q_DECLARE_METATYPE(MediaStreamType)

#endif // WEBRTC_SERVICE_H
