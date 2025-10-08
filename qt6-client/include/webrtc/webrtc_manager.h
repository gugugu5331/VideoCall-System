#ifndef WEBRTC_MANAGER_H
#define WEBRTC_MANAGER_H

#include <QObject>
#include <QMap>
#include <QVariantMap>
#include <memory>
#include <map>

// Forward declarations
class PeerConnection;
class MediaStream;
class WebSocketClient;
class AIService;
class RemoteStreamAnalyzer;

/**
 * @brief WebRTCManager类 - WebRTC管理器
 *
 * 管理本地媒体流和多个对等连接
 * 协调信令消息和媒体传输
 */
class WebRTCManager : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool audioEnabled READ audioEnabled NOTIFY audioEnabledChanged)
    Q_PROPERTY(bool videoEnabled READ videoEnabled NOTIFY videoEnabledChanged)
    Q_PROPERTY(bool isScreenSharing READ isScreenSharing NOTIFY isScreenSharingChanged)
    Q_PROPERTY(int peerConnectionCount READ peerConnectionCount NOTIFY peerConnectionCountChanged)

public:
    explicit WebRTCManager(WebSocketClient *wsClient, QObject *parent = nullptr);
    ~WebRTCManager();

    // Initialization
    Q_INVOKABLE bool initialize(const QVariantMap &config = QVariantMap());

    // AI Service integration
    void setAIService(AIService *aiService);
    AIService* aiService() const { return m_aiService; }

    // Media stream management
    Q_INVOKABLE bool startLocalMedia(bool audio = true, bool video = true);
    Q_INVOKABLE void stopLocalMedia();
    Q_INVOKABLE MediaStream* getLocalStream() const { return m_localStream.get(); }

    // Peer connection management
    Q_INVOKABLE void createPeerConnection(int remoteUserId);
    Q_INVOKABLE void closePeerConnection(int remoteUserId);
    Q_INVOKABLE void closeAllPeerConnections();
    Q_INVOKABLE bool hasPeerConnection(int remoteUserId) const;
    Q_INVOKABLE int peerConnectionCount() const { return m_peerConnections.size(); }

    // Media control
    Q_INVOKABLE void setAudioEnabled(bool enabled);
    Q_INVOKABLE void setVideoEnabled(bool enabled);
    Q_INVOKABLE void toggleAudio();
    Q_INVOKABLE void toggleVideo();
    Q_INVOKABLE bool startScreenShare(int screenIndex = 0);
    Q_INVOKABLE void stopScreenShare();

    // Properties
    bool audioEnabled() const { return m_audioEnabled; }
    bool videoEnabled() const { return m_videoEnabled; }
    bool isScreenSharing() const { return m_isScreenSharing; }

    // Signaling
    Q_INVOKABLE void createOffer(int remoteUserId);
    Q_INVOKABLE void handleOffer(int remoteUserId, const QString &sdp);
    Q_INVOKABLE void handleAnswer(int remoteUserId, const QString &sdp);
    Q_INVOKABLE void handleIceCandidate(int remoteUserId, const QString &candidate,
                                       const QString &sdpMid, int sdpMLineIndex);

    // Statistics
    Q_INVOKABLE QVariantMap getStatistics(int remoteUserId) const;
    Q_INVOKABLE QVariantMap getAllStatistics() const;

    // Device management
    Q_INVOKABLE QStringList getAudioInputDevices() const;
    Q_INVOKABLE QStringList getVideoInputDevices() const;
    Q_INVOKABLE bool setAudioInputDevice(const QString &deviceName);
    Q_INVOKABLE bool setVideoInputDevice(const QString &deviceName);

signals:
    // Media stream signals
    void localStreamReady(MediaStream *stream);
    void localStreamStopped();
    void remoteStreamAdded(int userId, MediaStream *stream);
    void remoteStreamRemoved(int userId);

    // Signaling signals
    void offerCreated(int remoteUserId, const QString &sdp);
    void answerCreated(int remoteUserId, const QString &sdp);
    void iceCandidateGenerated(int remoteUserId, const QString &candidate,
                              const QString &sdpMid, int sdpMLineIndex);

    // Connection signals
    void peerConnectionCreated(int userId);
    void peerConnectionClosed(int userId);
    void connectionStateChanged(int userId, const QString &state);
    void iceConnectionStateChanged(int userId, const QString &state);

    // Property change signals
    void audioEnabledChanged();
    void videoEnabledChanged();
    void isScreenSharingChanged();
    void peerConnectionCountChanged();

    // Error signals
    void error(const QString &error);

private slots:
    void onPeerConnectionStateChanged(const QString &state);
    void onPeerConnectionIceStateChanged(const QString &state);
    void onPeerConnectionError(const QString &error);
    void onPeerConnectionIceCandidate(const QString &candidate, const QString &sdpMid, int sdpMLineIndex);

private:
    PeerConnection* getPeerConnection(int remoteUserId);
    void setupPeerConnection(PeerConnection *pc, int remoteUserId);
    void setupAIAnalysisForRemoteStream(int remoteUserId, MediaStream *stream);
    void cleanupPeerConnection(int remoteUserId);

private:
    WebSocketClient *m_wsClient;
    AIService *m_aiService;
    std::unique_ptr<MediaStream> m_localStream;
    std::map<int, std::unique_ptr<PeerConnection>> m_peerConnections;
    std::map<int, std::unique_ptr<RemoteStreamAnalyzer>> m_streamAnalyzers;

    bool m_audioEnabled;
    bool m_videoEnabled;
    bool m_isScreenSharing;
    bool m_initialized;

    QVariantMap m_config;
};

#endif // WEBRTC_MANAGER_H

