#include "webrtc/webrtc_peer_connection.h"
#include "utils/logger.h"

#include <api/peer_connection_interface.h>
#include <api/create_peerconnection_factory.h>
#include <api/audio_codecs/builtin_audio_decoder_factory.h>
#include <api/audio_codecs/builtin_audio_encoder_factory.h>
#include <api/video_codecs/builtin_video_decoder_factory.h>
#include <api/video_codecs/builtin_video_encoder_factory.h>
#include <rtc_base/ssl_adapter.h>
#include <rtc_base/thread.h>

// ============================================================================
// CreateSDPObserver 实现
// ============================================================================

CreateSDPObserver::CreateSDPObserver(WebRTCPeerConnection* parent)
    : m_parent(parent)
{
}

void CreateSDPObserver::OnSuccess(webrtc::SessionDescriptionInterface* desc)
{
    LOG_INFO("SDP created successfully");
    
    // 设置本地描述
    m_parent->m_peerConnection->SetLocalDescription(
        m_parent->m_setSDPObserver.get(), desc);
    
    // 序列化SDP
    std::string sdp_str;
    desc->ToString(&sdp_str);
    
    QString type = QString::fromStdString(desc->type());
    QString sdp = QString::fromStdString(sdp_str);
    
    // 发送信号
    emit m_parent->localDescriptionCreated(type, sdp);
}

void CreateSDPObserver::OnFailure(webrtc::RTCError error)
{
    LOG_ERROR(QString("Failed to create SDP: %1").arg(error.message()));
    emit m_parent->errorOccurred(QString::fromStdString(error.message()));
}

// ============================================================================
// SetSDPObserver 实现
// ============================================================================

SetSDPObserver::SetSDPObserver(WebRTCPeerConnection* parent)
    : m_parent(parent)
{
}

void SetSDPObserver::OnSuccess()
{
    LOG_INFO("SDP set successfully");
    emit m_parent->remoteDescriptionSet();
}

void SetSDPObserver::OnFailure(webrtc::RTCError error)
{
    LOG_ERROR(QString("Failed to set SDP: %1").arg(error.message()));
    emit m_parent->errorOccurred(QString::fromStdString(error.message()));
}

// ============================================================================
// PCObserver 实现
// ============================================================================

PCObserver::PCObserver(WebRTCPeerConnection* parent)
    : m_parent(parent)
{
}

void PCObserver::OnIceCandidate(const webrtc::IceCandidateInterface* candidate)
{
    LOG_INFO("ICE candidate generated");
    
    // 序列化候选信息
    std::string sdp_mid = candidate->sdp_mid();
    int sdp_mline_index = candidate->sdp_mline_index();
    std::string sdp;
    candidate->ToString(&sdp);
    
    QVariantMap candidateMap;
    candidateMap["sdpMid"] = QString::fromStdString(sdp_mid);
    candidateMap["sdpMLineIndex"] = sdp_mline_index;
    candidateMap["candidate"] = QString::fromStdString(sdp);
    
    emit m_parent->iceCandidateGenerated(candidateMap);
}

void PCObserver::OnIceConnectionChange(webrtc::PeerConnectionInterface::IceConnectionState new_state)
{
    QString state = m_parent->iceConnectionStateToString(new_state);
    LOG_INFO(QString("ICE connection state changed: %1").arg(state));
    emit m_parent->iceConnectionStateChanged(state);
}

void PCObserver::OnSignalingChange(webrtc::PeerConnectionInterface::SignalingState new_state)
{
    QString state = m_parent->signalingStateToString(new_state);
    LOG_INFO(QString("Signaling state changed: %1").arg(state));
    emit m_parent->signalingStateChanged(state);
}

void PCObserver::OnAddStream(rtc::scoped_refptr<webrtc::MediaStreamInterface> stream)
{
    LOG_INFO(QString("Remote stream added: %1").arg(QString::fromStdString(stream->id())));
    m_parent->m_remoteStream = stream;
    emit m_parent->remoteStreamAdded(QString::fromStdString(stream->id()));
}

void PCObserver::OnRemoveStream(rtc::scoped_refptr<webrtc::MediaStreamInterface> stream)
{
    LOG_INFO(QString("Remote stream removed: %1").arg(QString::fromStdString(stream->id())));
    emit m_parent->remoteStreamRemoved(QString::fromStdString(stream->id()));
}

void PCObserver::OnDataChannel(rtc::scoped_refptr<webrtc::DataChannelInterface> data_channel)
{
    LOG_INFO("Data channel received");
    // 可以在这里处理数据通道
}

void PCObserver::OnIceGatheringChange(webrtc::PeerConnectionInterface::IceGatheringState new_state)
{
    QString state;
    switch (new_state) {
        case webrtc::PeerConnectionInterface::kIceGatheringNew:
            state = "new";
            break;
        case webrtc::PeerConnectionInterface::kIceGatheringGathering:
            state = "gathering";
            break;
        case webrtc::PeerConnectionInterface::kIceGatheringComplete:
            state = "complete";
            break;
    }
    LOG_INFO(QString("ICE gathering state changed: %1").arg(state));
    emit m_parent->iceGatheringStateChanged(state);
}

// ============================================================================
// WebRTCPeerConnection 实现
// ============================================================================

WebRTCPeerConnection::WebRTCPeerConnection(int remoteUserId, QObject *parent)
    : QObject(parent)
    , m_remoteUserId(remoteUserId)
{
    LOG_INFO(QString("WebRTCPeerConnection created for user: %1").arg(m_remoteUserId));
    
    // 初始化统计信息
    m_statistics.bytesSent = 0;
    m_statistics.bytesReceived = 0;
    m_statistics.packetsSent = 0;
    m_statistics.packetsReceived = 0;
    m_statistics.packetsLost = 0;
    m_statistics.currentRoundTripTime = 0.0;
    m_statistics.availableOutgoingBitrate = 0.0;
    m_statistics.availableIncomingBitrate = 0.0;
    
    // 初始化SSL
    rtc::InitializeSSL();
}

WebRTCPeerConnection::~WebRTCPeerConnection()
{
    close();
    LOG_INFO(QString("WebRTCPeerConnection destroyed for user: %1").arg(m_remoteUserId));
    
    // 清理SSL
    rtc::CleanupSSL();
}

bool WebRTCPeerConnection::initialize(const QVariantMap &config)
{
    LOG_INFO(QString("Initializing WebRTCPeerConnection for user: %1").arg(m_remoteUserId));
    
    // 创建PeerConnectionFactory
    if (!createPeerConnectionFactory()) {
        LOG_ERROR("Failed to create PeerConnectionFactory");
        return false;
    }
    
    // 解析ICE服务器配置
    webrtc::PeerConnectionInterface::RTCConfiguration rtc_config = parseIceServers(config);
    
    // 启用DTLS
    rtc_config.enable_dtls_srtp = true;
    
    // 设置ICE传输策略
    rtc_config.type = webrtc::PeerConnectionInterface::kAll;
    
    // 设置Bundle策略
    rtc_config.bundle_policy = webrtc::PeerConnectionInterface::kBundlePolicyMaxBundle;
    
    // 设置RTCP多路复用策略
    rtc_config.rtcp_mux_policy = webrtc::PeerConnectionInterface::kRtcpMuxPolicyRequire;
    
    // 创建观察者
    m_pcObserver = std::make_unique<PCObserver>(this);
    m_createSDPObserver = new rtc::RefCountedObject<CreateSDPObserver>(this);
    m_setSDPObserver = new rtc::RefCountedObject<SetSDPObserver>(this);
    
    // 创建PeerConnection
    webrtc::PeerConnectionDependencies dependencies(m_pcObserver.get());
    auto result = m_peerConnectionFactory->CreatePeerConnectionOrError(
        rtc_config, std::move(dependencies));
    
    if (!result.ok()) {
        LOG_ERROR(QString("Failed to create PeerConnection: %1")
            .arg(QString::fromStdString(result.error().message())));
        return false;
    }
    
    m_peerConnection = result.MoveValue();
    LOG_INFO("PeerConnection created successfully");
    
    return true;
}

bool WebRTCPeerConnection::createPeerConnectionFactory()
{
    // 创建线程
    rtc::Thread* signaling_thread = rtc::Thread::Current();
    rtc::Thread* worker_thread = rtc::Thread::Create().release();
    worker_thread->Start();
    
    // 创建工厂
    m_peerConnectionFactory = webrtc::CreatePeerConnectionFactory(
        worker_thread,
        worker_thread,
        signaling_thread,
        nullptr,  // default ADM
        webrtc::CreateBuiltinAudioEncoderFactory(),
        webrtc::CreateBuiltinAudioDecoderFactory(),
        webrtc::CreateBuiltinVideoEncoderFactory(),
        webrtc::CreateBuiltinVideoDecoderFactory(),
        nullptr,  // audio mixer
        nullptr   // audio processing
    );
    
    return m_peerConnectionFactory != nullptr;
}

webrtc::PeerConnectionInterface::RTCConfiguration 
WebRTCPeerConnection::parseIceServers(const QVariantMap &config)
{
    webrtc::PeerConnectionInterface::RTCConfiguration rtc_config;
    
    if (config.contains("iceServers")) {
        QVariantList iceServers = config["iceServers"].toList();
        
        for (const QVariant &serverVar : iceServers) {
            QVariantMap serverMap = serverVar.toMap();
            QString urls = serverMap["urls"].toString();
            
            webrtc::PeerConnectionInterface::IceServer ice_server;
            ice_server.uri = urls.toStdString();
            
            // TURN服务器需要认证
            if (urls.startsWith("turn:")) {
                if (serverMap.contains("username")) {
                    ice_server.username = serverMap["username"].toString().toStdString();
                }
                if (serverMap.contains("credential")) {
                    ice_server.password = serverMap["credential"].toString().toStdString();
                }
            }
            
            rtc_config.servers.push_back(ice_server);
        }
    }
    
    return rtc_config;
}

void WebRTCPeerConnection::createOffer()
{
    if (!m_peerConnection) {
        LOG_ERROR("PeerConnection not initialized");
        return;
    }

    LOG_INFO("Creating offer");

    // 设置Offer选项
    webrtc::PeerConnectionInterface::RTCOfferAnswerOptions options;
    options.offer_to_receive_audio = true;
    options.offer_to_receive_video = true;

    m_peerConnection->CreateOffer(m_createSDPObserver.get(), options);
}

void WebRTCPeerConnection::createAnswer()
{
    if (!m_peerConnection) {
        LOG_ERROR("PeerConnection not initialized");
        return;
    }

    LOG_INFO("Creating answer");

    // 设置Answer选项
    webrtc::PeerConnectionInterface::RTCOfferAnswerOptions options;

    m_peerConnection->CreateAnswer(m_createSDPObserver.get(), options);
}

void WebRTCPeerConnection::setRemoteDescription(const QString &type, const QString &sdp)
{
    if (!m_peerConnection) {
        LOG_ERROR("PeerConnection not initialized");
        return;
    }

    LOG_INFO(QString("Setting remote description: %1").arg(type));

    // 创建SessionDescription
    webrtc::SdpType sdp_type;
    if (type == "offer") {
        sdp_type = webrtc::SdpType::kOffer;
    } else if (type == "answer") {
        sdp_type = webrtc::SdpType::kAnswer;
    } else {
        LOG_ERROR(QString("Unknown SDP type: %1").arg(type));
        return;
    }

    webrtc::SdpParseError error;
    std::unique_ptr<webrtc::SessionDescriptionInterface> session_description =
        webrtc::CreateSessionDescription(sdp_type, sdp.toStdString(), &error);

    if (!session_description) {
        LOG_ERROR(QString("Failed to parse SDP: %1").arg(QString::fromStdString(error.description)));
        emit errorOccurred(QString::fromStdString(error.description));
        return;
    }

    m_peerConnection->SetRemoteDescription(
        m_setSDPObserver.get(), session_description.release());
}

void WebRTCPeerConnection::addIceCandidate(const QVariantMap &candidate)
{
    if (!m_peerConnection) {
        LOG_ERROR("PeerConnection not initialized");
        return;
    }

    QString sdp_mid = candidate["sdpMid"].toString();
    int sdp_mline_index = candidate["sdpMLineIndex"].toInt();
    QString sdp = candidate["candidate"].toString();

    LOG_INFO(QString("Adding ICE candidate: %1").arg(sdp));

    webrtc::SdpParseError error;
    std::unique_ptr<webrtc::IceCandidateInterface> ice_candidate(
        webrtc::CreateIceCandidate(
            sdp_mid.toStdString(),
            sdp_mline_index,
            sdp.toStdString(),
            &error
        )
    );

    if (!ice_candidate) {
        LOG_ERROR(QString("Failed to parse ICE candidate: %1")
            .arg(QString::fromStdString(error.description)));
        return;
    }

    if (!m_peerConnection->AddIceCandidate(ice_candidate.get())) {
        LOG_ERROR("Failed to add ICE candidate");
    }
}

void WebRTCPeerConnection::addLocalStream(rtc::scoped_refptr<webrtc::MediaStreamInterface> stream)
{
    if (!m_peerConnection) {
        LOG_ERROR("PeerConnection not initialized");
        return;
    }

    m_localStream = stream;

    // 添加音频轨道
    auto audio_tracks = stream->GetAudioTracks();
    for (auto& track : audio_tracks) {
        m_peerConnection->AddTrack(track, {stream->id()});
        LOG_INFO(QString("Added audio track: %1").arg(QString::fromStdString(track->id())));
    }

    // 添加视频轨道
    auto video_tracks = stream->GetVideoTracks();
    for (auto& track : video_tracks) {
        m_peerConnection->AddTrack(track, {stream->id()});
        LOG_INFO(QString("Added video track: %1").arg(QString::fromStdString(track->id())));
    }
}

void WebRTCPeerConnection::close()
{
    if (m_peerConnection) {
        m_peerConnection->Close();
        m_peerConnection = nullptr;
    }

    m_localStream = nullptr;
    m_remoteStream = nullptr;
}

QVariantMap WebRTCPeerConnection::getStatistics() const
{
    QVariantMap stats;
    stats["bytesSent"] = static_cast<qulonglong>(m_statistics.bytesSent);
    stats["bytesReceived"] = static_cast<qulonglong>(m_statistics.bytesReceived);
    stats["packetsSent"] = static_cast<qulonglong>(m_statistics.packetsSent);
    stats["packetsReceived"] = static_cast<qulonglong>(m_statistics.packetsReceived);
    stats["packetsLost"] = static_cast<qulonglong>(m_statistics.packetsLost);
    stats["currentRoundTripTime"] = m_statistics.currentRoundTripTime;
    stats["availableOutgoingBitrate"] = m_statistics.availableOutgoingBitrate;
    stats["availableIncomingBitrate"] = m_statistics.availableIncomingBitrate;
    return stats;
}

QString WebRTCPeerConnection::iceConnectionStateToString(
    webrtc::PeerConnectionInterface::IceConnectionState state)
{
    switch (state) {
        case webrtc::PeerConnectionInterface::kIceConnectionNew:
            return "new";
        case webrtc::PeerConnectionInterface::kIceConnectionChecking:
            return "checking";
        case webrtc::PeerConnectionInterface::kIceConnectionConnected:
            return "connected";
        case webrtc::PeerConnectionInterface::kIceConnectionCompleted:
            return "completed";
        case webrtc::PeerConnectionInterface::kIceConnectionFailed:
            return "failed";
        case webrtc::PeerConnectionInterface::kIceConnectionDisconnected:
            return "disconnected";
        case webrtc::PeerConnectionInterface::kIceConnectionClosed:
            return "closed";
        default:
            return "unknown";
    }
}

QString WebRTCPeerConnection::signalingStateToString(
    webrtc::PeerConnectionInterface::SignalingState state)
{
    switch (state) {
        case webrtc::PeerConnectionInterface::kStable:
            return "stable";
        case webrtc::PeerConnectionInterface::kHaveLocalOffer:
            return "have-local-offer";
        case webrtc::PeerConnectionInterface::kHaveLocalPrAnswer:
            return "have-local-pranswer";
        case webrtc::PeerConnectionInterface::kHaveRemoteOffer:
            return "have-remote-offer";
        case webrtc::PeerConnectionInterface::kHaveRemotePrAnswer:
            return "have-remote-pranswer";
        case webrtc::PeerConnectionInterface::kClosed:
            return "closed";
        default:
            return "unknown";
    }
}

