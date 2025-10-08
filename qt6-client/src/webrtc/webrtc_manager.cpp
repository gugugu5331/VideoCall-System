#include "webrtc/webrtc_manager.h"
#include "webrtc/peer_connection.h"
#include "webrtc/media_stream.h"
#include "webrtc/remote_stream_analyzer.h"
#include "network/websocket_client.h"
#include "services/ai_service.h"
#include "utils/logger.h"

#include <QMediaDevices>
#include <QAudioDevice>
#include <QCameraDevice>

WebRTCManager::WebRTCManager(WebSocketClient *wsClient, QObject *parent)
    : QObject(parent)
    , m_wsClient(wsClient)
    , m_aiService(nullptr)
    , m_localStream(nullptr)
    , m_audioEnabled(true)
    , m_videoEnabled(true)
    , m_isScreenSharing(false)
    , m_initialized(false)
{
    LOG_INFO("WebRTCManager created");
}

WebRTCManager::~WebRTCManager()
{
    stopLocalMedia();
    closeAllPeerConnections();
    m_streamAnalyzers.clear();
    LOG_INFO("WebRTCManager destroyed");
}

void WebRTCManager::setAIService(AIService *aiService)
{
    m_aiService = aiService;
    LOG_INFO("AIService set for WebRTCManager");
}

bool WebRTCManager::initialize(const QVariantMap &config)
{
    if (m_initialized) {
        LOG_WARNING("WebRTCManager already initialized");
        return true;
    }

    LOG_INFO("Initializing WebRTCManager");

    m_config = config;

    // 默认ICE服务器配置
    if (!m_config.contains("iceServers")) {
        QVariantList iceServers;
        QVariantMap stunServer;
        stunServer["urls"] = "stun:stun.l.google.com:19302";
        iceServers.append(stunServer);
        m_config["iceServers"] = iceServers;
    }

    m_initialized = true;
    LOG_INFO("WebRTCManager initialized successfully");

    return true;
}

bool WebRTCManager::startLocalMedia(bool audio, bool video)
{
    if (m_localStream) {
        LOG_WARNING("Local media already started");
        return true;
    }

    LOG_INFO(QString("Starting local media (audio: %1, video: %2)")
             .arg(audio).arg(video));

    try {
        // 创建本地媒体流
        QString streamId = "local_" + QString::number(QDateTime::currentMSecsSinceEpoch());
        m_localStream = std::make_unique<MediaStream>(streamId, this);

        // 启动捕获
        if (!m_localStream->startCapture(audio, video)) {
            LOG_ERROR("Failed to start media capture");
            m_localStream.reset();
            emit error("Failed to start media capture");
            return false;
        }

        m_audioEnabled = audio;
        m_videoEnabled = video;

        LOG_INFO("Local media started successfully");
        emit localStreamReady(m_localStream.get());
        emit audioEnabledChanged();
        emit videoEnabledChanged();

        return true;

    } catch (const std::exception &e) {
        LOG_ERROR(QString("Failed to start local media: %1").arg(e.what()));
        emit error(QString("Failed to start local media: %1").arg(e.what()));
        return false;
    }
}

void WebRTCManager::stopLocalMedia()
{
    if (!m_localStream) {
        return;
    }

    LOG_INFO("Stopping local media");

    m_localStream->stopCapture();
    m_localStream.reset();

    m_audioEnabled = false;
    m_videoEnabled = false;
    m_isScreenSharing = false;

    emit localStreamStopped();
    emit audioEnabledChanged();
    emit videoEnabledChanged();
    emit isScreenSharingChanged();

    LOG_INFO("Local media stopped");
}

void WebRTCManager::createPeerConnection(int remoteUserId)
{
    if (m_peerConnections.find(remoteUserId) != m_peerConnections.end()) {
        LOG_WARNING(QString("PeerConnection already exists for user: %1").arg(remoteUserId));
        return;
    }

    LOG_INFO(QString("Creating PeerConnection for user: %1").arg(remoteUserId));

    try {
        // 创建PeerConnection
        auto pc = std::make_unique<PeerConnection>(remoteUserId, this);

        // 初始化
        if (!pc->initialize(m_config)) {
            LOG_ERROR(QString("Failed to initialize PeerConnection for user: %1").arg(remoteUserId));
            emit error(QString("Failed to create peer connection for user %1").arg(remoteUserId));
            return;
        }

        // 添加本地流
        if (m_localStream) {
            pc->addLocalStream(m_localStream.get());
        }

        // 设置信号连接
        setupPeerConnection(pc.get(), remoteUserId);

        // 保存PeerConnection
        m_peerConnections[remoteUserId] = std::move(pc);

        emit peerConnectionCreated(remoteUserId);
        emit peerConnectionCountChanged();

        LOG_INFO(QString("PeerConnection created for user: %1").arg(remoteUserId));

    } catch (const std::exception &e) {
        LOG_ERROR(QString("Failed to create PeerConnection: %1").arg(e.what()));
        emit error(QString("Failed to create peer connection: %1").arg(e.what()));
    }
}

void WebRTCManager::closePeerConnection(int remoteUserId)
{
    if (m_peerConnections.find(remoteUserId) == m_peerConnections.end()) {
        LOG_WARNING(QString("PeerConnection not found for user: %1").arg(remoteUserId));
        return;
    }

    LOG_INFO(QString("Closing PeerConnection for user: %1").arg(remoteUserId));

    cleanupPeerConnection(remoteUserId);
    m_peerConnections.erase(remoteUserId);

    emit peerConnectionClosed(remoteUserId);
    emit peerConnectionCountChanged();

    LOG_INFO(QString("PeerConnection closed for user: %1").arg(remoteUserId));
}

void WebRTCManager::closeAllPeerConnections()
{
    LOG_INFO("Closing all PeerConnections");

    std::vector<int> userIds;
    for (const auto& pair : m_peerConnections) {
        userIds.push_back(pair.first);
    }
    for (int userId : userIds) {
        closePeerConnection(userId);
    }

    LOG_INFO("All PeerConnections closed");
}

bool WebRTCManager::hasPeerConnection(int remoteUserId) const
{
    return m_peerConnections.find(remoteUserId) != m_peerConnections.end();
}

void WebRTCManager::setAudioEnabled(bool enabled)
{
    if (m_audioEnabled == enabled) {
        return;
    }

    LOG_INFO(QString("Setting audio %1").arg(enabled ? "enabled" : "disabled"));

    m_audioEnabled = enabled;

    if (m_localStream) {
        m_localStream->setAudioEnabled(enabled);
    }

    emit audioEnabledChanged();
}

void WebRTCManager::setVideoEnabled(bool enabled)
{
    if (m_videoEnabled == enabled) {
        return;
    }

    LOG_INFO(QString("Setting video %1").arg(enabled ? "enabled" : "disabled"));

    m_videoEnabled = enabled;

    if (m_localStream) {
        m_localStream->setVideoEnabled(enabled);
    }

    emit videoEnabledChanged();
}

void WebRTCManager::toggleAudio()
{
    setAudioEnabled(!m_audioEnabled);
}

void WebRTCManager::toggleVideo()
{
    setVideoEnabled(!m_videoEnabled);
}

bool WebRTCManager::startScreenShare(int screenIndex)
{
    if (m_isScreenSharing) {
        LOG_WARNING("Screen sharing already active");
        return true;
    }

    if (!m_localStream) {
        LOG_ERROR("No local stream available for screen sharing");
        emit error("No local stream available");
        return false;
    }

    LOG_INFO(QString("Starting screen share (screen: %1)").arg(screenIndex));

    if (!m_localStream->startScreenShare(screenIndex)) {
        LOG_ERROR("Failed to start screen share");
        emit error("Failed to start screen share");
        return false;
    }

    m_isScreenSharing = true;
    emit isScreenSharingChanged();

    LOG_INFO("Screen share started successfully");
    return true;
}

void WebRTCManager::stopScreenShare()
{
    if (!m_isScreenSharing) {
        return;
    }

    LOG_INFO("Stopping screen share");

    if (m_localStream) {
        m_localStream->stopScreenShare();
    }

    m_isScreenSharing = false;
    emit isScreenSharingChanged();

    LOG_INFO("Screen share stopped");
}

void WebRTCManager::createOffer(int remoteUserId)
{
    PeerConnection *pc = getPeerConnection(remoteUserId);
    if (!pc) {
        LOG_ERROR(QString("PeerConnection not found for user: %1").arg(remoteUserId));
        emit error(QString("Peer connection not found for user %1").arg(remoteUserId));
        return;
    }

    LOG_INFO(QString("Creating offer for user: %1").arg(remoteUserId));

    QString sdp = pc->createOffer();

    if (sdp.isEmpty()) {
        LOG_ERROR(QString("Failed to create offer for user: %1").arg(remoteUserId));
        emit error(QString("Failed to create offer for user %1").arg(remoteUserId));
        return;
    }

    emit offerCreated(remoteUserId, sdp);
    LOG_INFO(QString("Offer created for user: %1").arg(remoteUserId));
}

void WebRTCManager::handleOffer(int remoteUserId, const QString &sdp)
{
    LOG_INFO(QString("Handling offer from user: %1").arg(remoteUserId));

    // 如果PeerConnection不存在，先创建
    if (!hasPeerConnection(remoteUserId)) {
        createPeerConnection(remoteUserId);
    }

    PeerConnection *pc = getPeerConnection(remoteUserId);
    if (!pc) {
        LOG_ERROR(QString("Failed to get PeerConnection for user: %1").arg(remoteUserId));
        emit error(QString("Failed to handle offer from user %1").arg(remoteUserId));
        return;
    }

    // 设置远程描述
    pc->setRemoteDescription(sdp, "offer");

    // 创建应答
    QString answerSdp = pc->createAnswer(sdp);

    if (answerSdp.isEmpty()) {
        LOG_ERROR(QString("Failed to create answer for user: %1").arg(remoteUserId));
        emit error(QString("Failed to create answer for user %1").arg(remoteUserId));
        return;
    }

    emit answerCreated(remoteUserId, answerSdp);
    LOG_INFO(QString("Answer created for user: %1").arg(remoteUserId));
}

void WebRTCManager::handleAnswer(int remoteUserId, const QString &sdp)
{
    LOG_INFO(QString("Handling answer from user: %1").arg(remoteUserId));

    PeerConnection *pc = getPeerConnection(remoteUserId);
    if (!pc) {
        LOG_ERROR(QString("PeerConnection not found for user: %1").arg(remoteUserId));
        emit error(QString("Peer connection not found for user %1").arg(remoteUserId));
        return;
    }

    pc->setRemoteDescription(sdp, "answer");
    LOG_INFO(QString("Answer processed for user: %1").arg(remoteUserId));
}

void WebRTCManager::handleIceCandidate(int remoteUserId, const QString &candidate,
                                       const QString &sdpMid, int sdpMLineIndex)
{
    LOG_INFO(QString("Handling ICE candidate from user: %1").arg(remoteUserId));

    PeerConnection *pc = getPeerConnection(remoteUserId);
    if (!pc) {
        LOG_ERROR(QString("PeerConnection not found for user: %1").arg(remoteUserId));
        emit error(QString("Peer connection not found for user %1").arg(remoteUserId));
        return;
    }

    pc->addIceCandidate(candidate, sdpMid, sdpMLineIndex);
    LOG_DEBUG(QString("ICE candidate added for user: %1").arg(remoteUserId));
}

QVariantMap WebRTCManager::getStatistics(int remoteUserId) const
{
    PeerConnection *pc = const_cast<WebRTCManager*>(this)->getPeerConnection(remoteUserId);
    if (!pc) {
        return QVariantMap();
    }

    auto stats = pc->getStatistics();

    QVariantMap result;
    result["bytesSent"] = qulonglong(stats.bytesSent);
    result["bytesReceived"] = qulonglong(stats.bytesReceived);
    result["packetsSent"] = qulonglong(stats.packetsSent);
    result["packetsReceived"] = qulonglong(stats.packetsReceived);
    result["packetsLost"] = qulonglong(stats.packetsLost);
    result["currentRoundTripTime"] = stats.currentRoundTripTime;

    return result;
}

QVariantMap WebRTCManager::getAllStatistics() const
{
    QVariantMap result;

    for (auto it = m_peerConnections.cbegin(); it != m_peerConnections.cend(); ++it) {
        int userId = it->first;
        result[QString::number(userId)] = getStatistics(userId);
    }

    return result;
}

QStringList WebRTCManager::getAudioInputDevices() const
{
    QStringList devices;

    QList<QAudioDevice> audioDevices = QMediaDevices::audioInputs();
    for (const QAudioDevice &device : audioDevices) {
        devices.append(device.description());
    }

    return devices;
}

QStringList WebRTCManager::getVideoInputDevices() const
{
    QStringList devices;

    QList<QCameraDevice> cameraDevices = QMediaDevices::videoInputs();
    for (const QCameraDevice &device : cameraDevices) {
        devices.append(device.description());
    }

    return devices;
}

bool WebRTCManager::setAudioInputDevice(const QString &deviceName)
{
    if (!m_localStream) {
        LOG_ERROR("No local stream available");
        return false;
    }

    LOG_INFO(QString("Setting audio input device: %1").arg(deviceName));

    return m_localStream->setAudioInputDevice(deviceName);
}

bool WebRTCManager::setVideoInputDevice(const QString &deviceName)
{
    if (!m_localStream) {
        LOG_ERROR("No local stream available");
        return false;
    }

    LOG_INFO(QString("Setting video input device: %1").arg(deviceName));

    return m_localStream->setVideoInputDevice(deviceName);
}

// 私有槽函数
void WebRTCManager::onPeerConnectionStateChanged(const QString &state)
{
    PeerConnection *pc = qobject_cast<PeerConnection*>(sender());
    if (!pc) {
        return;
    }

    int userId = pc->remoteUserId();
    LOG_INFO(QString("PeerConnection state changed for user %1: %2").arg(userId).arg(state));

    emit connectionStateChanged(userId, state);
}

void WebRTCManager::onPeerConnectionIceStateChanged(const QString &state)
{
    PeerConnection *pc = qobject_cast<PeerConnection*>(sender());
    if (!pc) {
        return;
    }

    int userId = pc->remoteUserId();
    LOG_INFO(QString("ICE connection state changed for user %1: %2").arg(userId).arg(state));

    emit iceConnectionStateChanged(userId, state);
}

void WebRTCManager::onPeerConnectionError(const QString &errorMsg)
{
    PeerConnection *pc = qobject_cast<PeerConnection*>(sender());
    if (!pc) {
        return;
    }

    int userId = pc->remoteUserId();
    LOG_ERROR(QString("PeerConnection error for user %1: %2").arg(userId).arg(errorMsg));

    emit error(QString("Peer connection error for user %1: %2").arg(userId).arg(errorMsg));
}

void WebRTCManager::onPeerConnectionIceCandidate(const QString &candidate,
                                                 const QString &sdpMid,
                                                 int sdpMLineIndex)
{
    PeerConnection *pc = qobject_cast<PeerConnection*>(sender());
    if (!pc) {
        return;
    }

    int userId = pc->remoteUserId();
    LOG_DEBUG(QString("ICE candidate generated for user: %1").arg(userId));

    emit iceCandidateGenerated(userId, candidate, sdpMid, sdpMLineIndex);
}

// 私有辅助方法
PeerConnection* WebRTCManager::getPeerConnection(int remoteUserId)
{
    auto it = m_peerConnections.find(remoteUserId);
    if (it != m_peerConnections.end()) {
        return it->second.get();
    }
    return nullptr;
}

void WebRTCManager::setupPeerConnection(PeerConnection *pc, int remoteUserId)
{
    // 连接PeerConnection的信号
    connect(pc, &PeerConnection::connectionStateChanged,
            this, &WebRTCManager::onPeerConnectionStateChanged);

    connect(pc, &PeerConnection::iceConnectionStateChanged,
            this, &WebRTCManager::onPeerConnectionIceStateChanged);

    connect(pc, &PeerConnection::error,
            this, &WebRTCManager::onPeerConnectionError);

    connect(pc, &PeerConnection::iceCandidate,
            this, &WebRTCManager::onPeerConnectionIceCandidate);

    // 连接远程流信号
    connect(pc, &PeerConnection::remoteStreamAdded,
            this, [this, remoteUserId](MediaStream *stream) {
                LOG_INFO(QString("Remote stream added for user: %1").arg(remoteUserId));

                // 设置AI分析
                setupAIAnalysisForRemoteStream(remoteUserId, stream);

                emit remoteStreamAdded(remoteUserId, stream);
            });

    connect(pc, &PeerConnection::remoteStreamRemoved,
            this, [this, remoteUserId]() {
                LOG_INFO(QString("Remote stream removed for user: %1").arg(remoteUserId));

                // 停止并移除AI分析器
                if (m_streamAnalyzers.count(remoteUserId)) {
                    m_streamAnalyzers[remoteUserId]->stopAnalysis();
                    m_streamAnalyzers.erase(remoteUserId);
                }

                emit remoteStreamRemoved(remoteUserId);
            });
}

void WebRTCManager::setupAIAnalysisForRemoteStream(int remoteUserId, MediaStream *stream)
{
    if (!m_aiService) {
        LOG_WARNING("AIService not set, skipping AI analysis setup");
        return;
    }

    if (!stream) {
        LOG_WARNING(QString("Invalid stream for user: %1").arg(remoteUserId));
        return;
    }

    LOG_INFO(QString("Setting up AI analysis for remote user: %1").arg(remoteUserId));

    // 创建远程流分析器
    auto analyzer = std::make_unique<RemoteStreamAnalyzer>(remoteUserId, m_aiService, this);

    // 配置分析参数
    analyzer->setVideoAnalysisInterval(5000);      // 5秒分析一次视频
    analyzer->setAudioBufferDuration(3000);        // 累积3秒音频
    analyzer->setVideoDownscaleSize(QSize(640, 360)); // 降采样到360p
    analyzer->setAudioSampleRate(16000);           // 重采样到16kHz

    // 启用AI功能
    analyzer->setDeepfakeDetectionEnabled(true);
    analyzer->setAsrEnabled(true);
    analyzer->setEmotionDetectionEnabled(true);

    // 连接到远程流
    analyzer->attachToStream(stream);

    // 启动分析
    analyzer->startAnalysis();

    // 保存分析器
    m_streamAnalyzers[remoteUserId] = std::move(analyzer);

    LOG_INFO(QString("AI analysis started for remote user: %1").arg(remoteUserId));
}

void WebRTCManager::cleanupPeerConnection(int remoteUserId)
{
    auto it = m_peerConnections.find(remoteUserId);
    if (it != m_peerConnections.end()) {
        PeerConnection *pc = it->second.get();

        // 断开所有信号连接
        disconnect(pc, nullptr, this, nullptr);

        // 关闭连接
        pc->close();
    }
}

