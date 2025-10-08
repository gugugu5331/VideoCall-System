#ifndef PEER_CONNECTION_H
#define PEER_CONNECTION_H

#include <QObject>
#include <QString>
#include <QUdpSocket>
#include <QHostAddress>
#include <QTimer>
#include <QVideoFrame>
#include <memory>

// Forward declarations
class MediaStream;

/**
 * @brief PeerConnection类 - WebRTC对等连接（简化实现）
 *
 * 使用UDP套接字实现基本的RTP/RTCP传输
 * 注意：这是一个简化的实现，用于快速开发
 */
class PeerConnection : public QObject
{
    Q_OBJECT
    Q_PROPERTY(QString connectionState READ connectionState NOTIFY connectionStateChanged)
    Q_PROPERTY(QString iceConnectionState READ iceConnectionState NOTIFY iceConnectionStateChanged)

public:
    explicit PeerConnection(int remoteUserId, QObject *parent = nullptr);
    ~PeerConnection();

    // Connection management
    bool initialize(const QVariantMap &config);
    void close();

    // Remote user ID
    int remoteUserId() const { return m_remoteUserId; }

    // Media management
    void addLocalStream(MediaStream *stream);
    void removeLocalStream();

    // Signaling
    QString createOffer();
    QString createAnswer(const QString &offerSdp);
    void setRemoteDescription(const QString &sdp, const QString &type);
    void addIceCandidate(const QString &candidate, const QString &sdpMid, int sdpMLineIndex);

    // State
    QString connectionState() const { return m_connectionState; }
    QString iceConnectionState() const { return m_iceConnectionState; }

    // Statistics
    struct Statistics {
        quint64 bytesSent;
        quint64 bytesReceived;
        quint64 packetsSent;
        quint64 packetsReceived;
        quint64 packetsLost;
        double currentRoundTripTime;
    };
    Statistics getStatistics() const { return m_statistics; }

signals:
    void iceCandidate(const QString &candidate, const QString &sdpMid, int sdpMLineIndex);
    void remoteStreamAdded(MediaStream *stream);
    void remoteStreamRemoved();
    void connectionStateChanged(const QString &state);
    void iceConnectionStateChanged(const QString &state);
    void error(const QString &error);
    void videoFrameReceived(const QVideoFrame &frame);
    void audioDataReceived(const QByteArray &data);

private slots:
    void onRtpDataReceived();
    void onRtcpDataReceived();
    void onLocalVideoFrame(const QVideoFrame &frame);
    void onLocalAudioData(const QByteArray &data);
    void onRtcpTimeout();
    void onStunTimeout();

private:
    // SDP handling
    QString generateSdp(const QString &type);
    bool parseSdp(const QString &sdp);
    QString getLocalIpAddress() const;

    // RTP handling
    void sendRtpPacket(const QByteArray &payload, quint8 payloadType, quint32 timestamp);
    void processRtpPacket(const QByteArray &packet);
    QByteArray createRtpHeader(quint8 payloadType, quint16 sequenceNumber,
                               quint32 timestamp, quint32 ssrc);

    // RTCP handling
    void sendRtcpSenderReport();
    void sendRtcpReceiverReport();
    void processRtcpPacket(const QByteArray &packet);

    // STUN handling
    void performStunBinding();
    void processStunResponse(const QByteArray &response);

    // State management
    void setConnectionState(const QString &state);
    void setIceConnectionState(const QString &state);

private:
    int m_remoteUserId;
    QString m_connectionState;
    QString m_iceConnectionState;

    // Local stream
    MediaStream *m_localStream;

    // Network
    QUdpSocket *m_rtpSocket;
    QUdpSocket *m_rtcpSocket;
    QHostAddress m_remoteAddress;
    quint16 m_remoteRtpPort;
    quint16 m_remoteRtcpPort;

    // STUN
    QString m_stunServer;
    quint16 m_stunPort;
    QHostAddress m_publicAddress;
    quint16 m_publicPort;

    // RTP state
    quint16 m_rtpSequenceNumber;
    quint32 m_rtpTimestamp;
    quint32 m_rtpSsrc;

    // RTCP
    QTimer *m_rtcpTimer;
    Statistics m_statistics;

    // SDP
    QString m_localSdp;
    QString m_remoteSdp;

    // ICE candidates
    QList<QString> m_localCandidates;
    QList<QString> m_remoteCandidates;
};

#endif // PEER_CONNECTION_H

