#pragma once

#include "core/common.h"

namespace VideoCallSystem {

class SignalingClient;

struct RTCConfiguration {
    QStringList iceServers;
    QString stunServer = "stun:stun.l.google.com:19302";
    QString turnServer;
    QString turnUsername;
    QString turnPassword;
    bool enableDtls = true;
    bool enableRtpDataChannels = false;
    int maxRetransmitTime = 30000;
    int maxRetransmitAttempts = 16;
};

struct MediaConstraints {
    // 视频约束
    bool enableVideo = true;
    int videoWidth = 640;
    int videoHeight = 480;
    int videoFps = 30;
    int videoBitrate = 1000000; // 1Mbps
    QString videoCodec = "VP8";
    
    // 音频约束
    bool enableAudio = true;
    int audioSampleRate = 44100;
    int audioChannels = 2;
    int audioBitrate = 128000; // 128kbps
    QString audioCodec = "OPUS";
    
    // 其他约束
    bool enableDataChannel = true;
    bool enableScreenShare = false;
};

struct PeerConnectionStats {
    QString peerId;
    QString connectionState;
    QString iceConnectionState;
    QString iceGatheringState;
    QString signalingState;
    
    // 媒体统计
    int videoPacketsSent = 0;
    int videoPacketsReceived = 0;
    int audioPacketsSent = 0;
    int audioPacketsReceived = 0;
    
    // 带宽统计
    int videoBitrateSent = 0;
    int videoBitrateReceived = 0;
    int audioBitrateSent = 0;
    int audioBitrateReceived = 0;
    
    // 质量统计
    double videoPacketLossRate = 0.0;
    double audioPacketLossRate = 0.0;
    int roundTripTime = 0;
    int jitter = 0;
    
    QDateTime timestamp;
};

class WebRTCManager : public QObject
{
    Q_OBJECT

public:
    explicit WebRTCManager(QObject* parent = nullptr);
    ~WebRTCManager();

    // 初始化和清理
    bool initialize(const RTCConfiguration& config = RTCConfiguration{});
    void cleanup();
    bool isInitialized() const { return initialized_; }

    // 连接管理
    void setSignalingClient(SignalingClient* client);
    SignalingClient* signalingClient() const { return signalingClient_; }

    // 会议控制
    bool joinMeeting(const QString& meetingId, const QString& userId);
    void leaveMeeting();
    bool isInMeeting() const { return inMeeting_; }
    QString currentMeetingId() const { return currentMeetingId_; }

    // 对等连接管理
    bool createPeerConnection(const QString& peerId);
    void closePeerConnection(const QString& peerId);
    void closeAllPeerConnections();
    QStringList getPeerIds() const;
    bool hasPeerConnection(const QString& peerId) const;

    // 媒体流管理
    bool addLocalStream();
    void removeLocalStream();
    bool hasLocalStream() const { return hasLocalStream_; }

    // 媒体控制
    void enableVideo(bool enable);
    void enableAudio(bool enable);
    void enableScreenShare(bool enable);
    bool isVideoEnabled() const { return videoEnabled_; }
    bool isAudioEnabled() const { return audioEnabled_; }
    bool isScreenShareEnabled() const { return screenShareEnabled_; }

    // 媒体约束
    void setMediaConstraints(const MediaConstraints& constraints);
    MediaConstraints getMediaConstraints() const { return mediaConstraints_; }

    // 数据通道
    bool createDataChannel(const QString& peerId, const QString& label);
    void closeDataChannel(const QString& peerId, const QString& label);
    void sendDataChannelMessage(const QString& peerId, const QString& label, const QByteArray& data);

    // 统计信息
    PeerConnectionStats getPeerConnectionStats(const QString& peerId) const;
    QList<PeerConnectionStats> getAllPeerConnectionStats() const;
    void enableStatsCollection(bool enable, int intervalMs = 1000);

    // 配置管理
    void updateRTCConfiguration(const RTCConfiguration& config);
    RTCConfiguration getRTCConfiguration() const { return rtcConfig_; }

    // 错误处理
    QString lastError() const { return lastError_; }
    void clearError() { lastError_.clear(); }

public slots:
    // 信令处理
    void handleSignalingMessage(const QString& peerId, const QJsonObject& message);
    void handleIceCandidate(const QString& peerId, const QJsonObject& candidate);
    void handleOffer(const QString& peerId, const QJsonObject& offer);
    void handleAnswer(const QString& peerId, const QJsonObject& answer);

    // 媒体控制槽
    void toggleVideo();
    void toggleAudio();
    void toggleScreenShare();

    // 统计更新
    void updateStats();

signals:
    // 连接状态信号
    void initialized();
    void meetingJoined(const QString& meetingId);
    void meetingLeft();
    void peerConnected(const QString& peerId);
    void peerDisconnected(const QString& peerId);

    // 媒体流信号
    void localStreamAdded();
    void localStreamRemoved();
    void remoteStreamAdded(const QString& peerId, QObject* stream);
    void remoteStreamRemoved(const QString& peerId);

    // 媒体状态信号
    void videoStateChanged(bool enabled);
    void audioStateChanged(bool enabled);
    void screenShareStateChanged(bool enabled);

    // 数据通道信号
    void dataChannelOpened(const QString& peerId, const QString& label);
    void dataChannelClosed(const QString& peerId, const QString& label);
    void dataChannelMessageReceived(const QString& peerId, const QString& label, const QByteArray& data);

    // 统计信号
    void statsUpdated(const QList<PeerConnectionStats>& stats);

    // 错误信号
    void error(const QString& error);
    void peerConnectionError(const QString& peerId, const QString& error);

private slots:
    // 内部信号处理
    void onSignalingConnected();
    void onSignalingDisconnected();
    void onSignalingError(const QString& error);

private:
    // 内部结构
    struct PeerConnection {
        QString peerId;
        QObject* connection; // 实际的WebRTC PeerConnection对象
        QObject* localStream;
        QObject* remoteStream;
        QMap<QString, QObject*> dataChannels;
        PeerConnectionStats stats;
        bool isInitiator;
        QDateTime createdAt;
    };

    // 初始化函数
    bool initializeWebRTC();
    bool setupMediaDevices();
    void setupSignalingHandlers();

    // 对等连接处理
    PeerConnection* findPeerConnection(const QString& peerId);
    bool setupPeerConnection(PeerConnection* peer);
    void cleanupPeerConnection(PeerConnection* peer);

    // 媒体流处理
    QObject* createLocalStream();
    void configureMediaStream(QObject* stream);
    void handleRemoteStream(const QString& peerId, QObject* stream);

    // 信令处理
    void sendSignalingMessage(const QString& peerId, const QJsonObject& message);
    void createOffer(const QString& peerId);
    void createAnswer(const QString& peerId, const QJsonObject& offer);
    void setRemoteDescription(const QString& peerId, const QJsonObject& description);
    void addIceCandidate(const QString& peerId, const QJsonObject& candidate);

    // 统计收集
    void collectStats();
    void updatePeerStats(PeerConnection* peer);

    // 错误处理
    void setError(const QString& error);
    void handlePeerConnectionError(const QString& peerId, const QString& error);

private:
    // 初始化状态
    bool initialized_;
    
    // 配置
    RTCConfiguration rtcConfig_;
    MediaConstraints mediaConstraints_;
    
    // 信令客户端
    SignalingClient* signalingClient_;
    
    // 会议状态
    bool inMeeting_;
    QString currentMeetingId_;
    QString currentUserId_;
    
    // 对等连接
    QMap<QString, std::unique_ptr<PeerConnection>> peerConnections_;
    
    // 本地媒体流
    QObject* localStream_;
    bool hasLocalStream_;
    bool videoEnabled_;
    bool audioEnabled_;
    bool screenShareEnabled_;
    
    // 统计收集
    QTimer* statsTimer_;
    bool statsEnabled_;
    
    // 错误处理
    QString lastError_;
    
    // 互斥锁
    mutable QMutex mutex_;
};

} // namespace VideoCallSystem
