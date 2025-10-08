#ifndef WEBRTC_PEER_CONNECTION_H
#define WEBRTC_PEER_CONNECTION_H

#include <QObject>
#include <QVideoFrame>
#include <memory>
#include <functional>

// WebRTC前向声明
namespace webrtc {
    class PeerConnectionInterface;
    class PeerConnectionFactoryInterface;
    class DataChannelInterface;
    class MediaStreamInterface;
    class VideoTrackInterface;
    class AudioTrackInterface;
    class IceCandidateInterface;
    class SessionDescriptionInterface;
}

namespace rtc {
    template<class T> class scoped_refptr;
}

class WebRTCPeerConnection;

/**
 * @brief CreateSessionDescriptionObserver - SDP创建观察者
 */
class CreateSDPObserver : public webrtc::CreateSessionDescriptionObserver {
public:
    explicit CreateSDPObserver(WebRTCPeerConnection* parent);
    
    // 成功回调
    void OnSuccess(webrtc::SessionDescriptionInterface* desc) override;
    
    // 失败回调
    void OnFailure(webrtc::RTCError error) override;
    
private:
    WebRTCPeerConnection* m_parent;
};

/**
 * @brief SetSessionDescriptionObserver - SDP设置观察者
 */
class SetSDPObserver : public webrtc::SetSessionDescriptionObserver {
public:
    explicit SetSDPObserver(WebRTCPeerConnection* parent);
    
    void OnSuccess() override;
    void OnFailure(webrtc::RTCError error) override;
    
private:
    WebRTCPeerConnection* m_parent;
};

/**
 * @brief PeerConnectionObserver - 连接状态观察者
 */
class PCObserver : public webrtc::PeerConnectionObserver {
public:
    explicit PCObserver(WebRTCPeerConnection* parent);
    
    // ICE候选回调
    void OnIceCandidate(const webrtc::IceCandidateInterface* candidate) override;
    
    // ICE连接状态变化
    void OnIceConnectionChange(webrtc::PeerConnectionInterface::IceConnectionState new_state) override;
    
    // 信令状态变化
    void OnSignalingChange(webrtc::PeerConnectionInterface::SignalingState new_state) override;
    
    // 添加远程流
    void OnAddStream(rtc::scoped_refptr<webrtc::MediaStreamInterface> stream) override;
    
    // 移除远程流
    void OnRemoveStream(rtc::scoped_refptr<webrtc::MediaStreamInterface> stream) override;
    
    // 数据通道回调
    void OnDataChannel(rtc::scoped_refptr<webrtc::DataChannelInterface> data_channel) override;
    
    // ICE收集状态变化
    void OnIceGatheringChange(webrtc::PeerConnectionInterface::IceGatheringState new_state) override;
    
private:
    WebRTCPeerConnection* m_parent;
};

/**
 * @brief WebRTCPeerConnection - 官方libwebrtc封装类
 * 
 * 替代原有的简化PeerConnection实现，提供完整的WebRTC功能：
 * - 完整的ICE协商（STUN/TURN）
 * - DTLS-SRTP加密
 * - 多编解码器支持（H.264, VP8, VP9, Opus）
 * - 拥塞控制和QoS
 * - 与标准WebRTC客户端互操作
 */
class WebRTCPeerConnection : public QObject {
    Q_OBJECT
    
public:
    explicit WebRTCPeerConnection(int remoteUserId, QObject *parent = nullptr);
    ~WebRTCPeerConnection();
    
    /**
     * @brief 初始化连接
     * @param config ICE服务器配置
     * @return 是否成功
     */
    bool initialize(const QVariantMap &config);
    
    /**
     * @brief 创建Offer
     */
    void createOffer();
    
    /**
     * @brief 创建Answer
     */
    void createAnswer();
    
    /**
     * @brief 设置远程SDP
     * @param type "offer" 或 "answer"
     * @param sdp SDP字符串
     */
    void setRemoteDescription(const QString &type, const QString &sdp);
    
    /**
     * @brief 添加ICE候选
     * @param candidate 候选信息
     */
    void addIceCandidate(const QVariantMap &candidate);
    
    /**
     * @brief 添加本地媒体流
     * @param stream 媒体流
     */
    void addLocalStream(rtc::scoped_refptr<webrtc::MediaStreamInterface> stream);
    
    /**
     * @brief 关闭连接
     */
    void close();
    
    /**
     * @brief 获取连接统计信息
     */
    QVariantMap getStatistics() const;
    
    // 友元类声明
    friend class CreateSDPObserver;
    friend class SetSDPObserver;
    friend class PCObserver;
    
signals:
    // SDP相关信号
    void localDescriptionCreated(const QString &type, const QString &sdp);
    void remoteDescriptionSet();
    
    // ICE相关信号
    void iceCandidateGenerated(const QVariantMap &candidate);
    void iceConnectionStateChanged(const QString &state);
    void iceGatheringStateChanged(const QString &state);
    
    // 媒体流信号
    void remoteStreamAdded(const QString &streamId);
    void remoteStreamRemoved(const QString &streamId);
    void remoteVideoFrameReceived(const QVideoFrame &frame);
    void remoteAudioDataReceived(const QByteArray &data);
    
    // 连接状态信号
    void connectionStateChanged(const QString &state);
    void signalingStateChanged(const QString &state);
    
    // 错误信号
    void errorOccurred(const QString &error);
    
private:
    int m_remoteUserId;
    
    // WebRTC核心对象
    rtc::scoped_refptr<webrtc::PeerConnectionInterface> m_peerConnection;
    rtc::scoped_refptr<webrtc::PeerConnectionFactoryInterface> m_peerConnectionFactory;
    
    // 观察者对象
    std::unique_ptr<PCObserver> m_pcObserver;
    rtc::scoped_refptr<CreateSDPObserver> m_createSDPObserver;
    rtc::scoped_refptr<SetSDPObserver> m_setSDPObserver;
    
    // 媒体流
    rtc::scoped_refptr<webrtc::MediaStreamInterface> m_localStream;
    rtc::scoped_refptr<webrtc::MediaStreamInterface> m_remoteStream;
    
    // 统计信息
    struct Statistics {
        uint64_t bytesSent;
        uint64_t bytesReceived;
        uint64_t packetsSent;
        uint64_t packetsReceived;
        uint64_t packetsLost;
        double currentRoundTripTime;
        double availableOutgoingBitrate;
        double availableIncomingBitrate;
    } m_statistics;
    
    /**
     * @brief 创建PeerConnectionFactory
     */
    bool createPeerConnectionFactory();
    
    /**
     * @brief 解析ICE服务器配置
     */
    webrtc::PeerConnectionInterface::RTCConfiguration parseIceServers(const QVariantMap &config);
    
    /**
     * @brief 转换ICE连接状态为字符串
     */
    QString iceConnectionStateToString(webrtc::PeerConnectionInterface::IceConnectionState state);
    
    /**
     * @brief 转换信令状态为字符串
     */
    QString signalingStateToString(webrtc::PeerConnectionInterface::SignalingState state);
};

#endif // WEBRTC_PEER_CONNECTION_H

