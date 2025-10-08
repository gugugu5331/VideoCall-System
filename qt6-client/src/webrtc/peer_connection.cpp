#include "webrtc/peer_connection.h"
#include "webrtc/media_stream.h"
#include "utils/logger.h"

#include <QNetworkInterface>
#include <QDateTime>
#include <QRandomGenerator>
#include <QDataStream>

PeerConnection::PeerConnection(int remoteUserId, QObject *parent)
    : QObject(parent)
    , m_remoteUserId(remoteUserId)
    , m_connectionState("new")
    , m_iceConnectionState("new")
    , m_localStream(nullptr)
    , m_rtpSocket(nullptr)
    , m_rtcpSocket(nullptr)
    , m_remoteRtpPort(0)
    , m_remoteRtcpPort(0)
    , m_stunServer("stun.l.google.com")
    , m_stunPort(19302)
    , m_publicPort(0)
    , m_rtpSequenceNumber(QRandomGenerator::global()->bounded(65536))
    , m_rtpTimestamp(0)
    , m_rtpSsrc(QRandomGenerator::global()->generate())
    , m_rtcpTimer(nullptr)
{
    LOG_INFO(QString("PeerConnection created for user: %1").arg(m_remoteUserId));

    // 初始化统计信息
    m_statistics.bytesSent = 0;
    m_statistics.bytesReceived = 0;
    m_statistics.packetsSent = 0;
    m_statistics.packetsReceived = 0;
    m_statistics.packetsLost = 0;
    m_statistics.currentRoundTripTime = 0.0;
}

PeerConnection::~PeerConnection()
{
    close();
    LOG_INFO(QString("PeerConnection destroyed for user: %1").arg(m_remoteUserId));
}

bool PeerConnection::initialize(const QVariantMap &config)
{
    LOG_INFO(QString("Initializing PeerConnection for user: %1").arg(m_remoteUserId));

    try {
        // 创建RTP套接字
        m_rtpSocket = new QUdpSocket(this);
        if (!m_rtpSocket->bind(QHostAddress::Any, 0)) {
            LOG_ERROR("Failed to bind RTP socket");
            return false;
        }

        // 创建RTCP套接字（RTP端口+1）
        m_rtcpSocket = new QUdpSocket(this);
        if (!m_rtcpSocket->bind(QHostAddress::Any, m_rtpSocket->localPort() + 1)) {
            LOG_ERROR("Failed to bind RTCP socket");
            return false;
        }

        LOG_INFO(QString("RTP socket bound to port: %1").arg(m_rtpSocket->localPort()));
        LOG_INFO(QString("RTCP socket bound to port: %1").arg(m_rtcpSocket->localPort()));

        // 连接信号
        connect(m_rtpSocket, &QUdpSocket::readyRead,
                this, &PeerConnection::onRtpDataReceived);
        connect(m_rtcpSocket, &QUdpSocket::readyRead,
                this, &PeerConnection::onRtcpDataReceived);

        // 创建RTCP定时器（每5秒发送一次）
        m_rtcpTimer = new QTimer(this);
        connect(m_rtcpTimer, &QTimer::timeout,
                this, &PeerConnection::onRtcpTimeout);
        m_rtcpTimer->start(5000);

        // 解析ICE服务器配置
        if (config.contains("iceServers")) {
            QVariantList iceServers = config["iceServers"].toList();
            if (!iceServers.isEmpty()) {
                QVariantMap firstServer = iceServers.first().toMap();
                QString urls = firstServer["urls"].toString();

                // 解析STUN服务器地址
                if (urls.startsWith("stun:")) {
                    QString stunUrl = urls.mid(5); // 移除"stun:"前缀
                    QStringList parts = stunUrl.split(':');
                    if (parts.size() >= 1) {
                        m_stunServer = parts[0];
                        if (parts.size() >= 2) {
                            m_stunPort = parts[1].toUShort();
                        }
                    }
                }
            }
        }

        // 执行STUN绑定获取公网地址
        performStunBinding();

        setConnectionState("connecting");
        setIceConnectionState("checking");

        return true;

    } catch (const std::exception &e) {
        LOG_ERROR(QString("Failed to initialize PeerConnection: %1").arg(e.what()));
        emit error(QString("Failed to initialize: %1").arg(e.what()));
        return false;
    }
}

void PeerConnection::close()
{
    LOG_INFO(QString("Closing PeerConnection for user: %1").arg(m_remoteUserId));

    if (m_rtcpTimer) {
        m_rtcpTimer->stop();
        m_rtcpTimer->deleteLater();
        m_rtcpTimer = nullptr;
    }

    if (m_rtpSocket) {
        m_rtpSocket->close();
        m_rtpSocket->deleteLater();
        m_rtpSocket = nullptr;
    }

    if (m_rtcpSocket) {
        m_rtcpSocket->close();
        m_rtcpSocket->deleteLater();
        m_rtcpSocket = nullptr;
    }

    m_localStream = nullptr;

    setConnectionState("closed");
    setIceConnectionState("closed");
}

void PeerConnection::addLocalStream(MediaStream *stream)
{
    if (!stream) {
        LOG_WARNING("Attempted to add null stream");
        return;
    }

    m_localStream = stream;
    LOG_INFO(QString("Local stream added to PeerConnection: %1").arg(stream->streamId()));

    // 连接媒体流信号
    connect(stream, &MediaStream::videoFrameReady,
            this, &PeerConnection::onLocalVideoFrame);
    connect(stream, &MediaStream::audioDataReady,
            this, &PeerConnection::onLocalAudioData);
}

void PeerConnection::removeLocalStream()
{
    if (m_localStream) {
        disconnect(m_localStream, nullptr, this, nullptr);
        m_localStream = nullptr;
        LOG_INFO("Local stream removed from PeerConnection");
    }
}

QString PeerConnection::createOffer()
{
    LOG_INFO(QString("Creating offer for user: %1").arg(m_remoteUserId));

    m_localSdp = generateSdp("offer");

    // 生成ICE候选
    QString localIp = getLocalIpAddress();
    quint16 rtpPort = m_rtpSocket ? m_rtpSocket->localPort() : 0;

    QString candidate = QString("candidate:1 1 UDP 2130706431 %1 %2 typ host")
                        .arg(localIp)
                        .arg(rtpPort);
    m_localCandidates.append(candidate);

    // 发送ICE候选
    emit iceCandidate(candidate, "0", 0);

    return m_localSdp;
}

QString PeerConnection::createAnswer(const QString &offerSdp)
{
    LOG_INFO(QString("Creating answer for user: %1").arg(m_remoteUserId));

    // 解析offer SDP
    if (!parseSdp(offerSdp)) {
        LOG_ERROR("Failed to parse offer SDP");
        emit error("Failed to parse offer SDP");
        return QString();
    }

    m_remoteSdp = offerSdp;
    m_localSdp = generateSdp("answer");

    // 生成ICE候选
    QString localIp = getLocalIpAddress();
    quint16 rtpPort = m_rtpSocket ? m_rtpSocket->localPort() : 0;

    QString candidate = QString("candidate:1 1 UDP 2130706431 %1 %2 typ host")
                        .arg(localIp)
                        .arg(rtpPort);
    m_localCandidates.append(candidate);

    // 发送ICE候选
    emit iceCandidate(candidate, "0", 0);

    setConnectionState("connected");
    setIceConnectionState("connected");

    return m_localSdp;
}

void PeerConnection::setRemoteDescription(const QString &sdp, const QString &type)
{
    LOG_INFO(QString("Setting remote description (%1) for user: %2").arg(type).arg(m_remoteUserId));

    m_remoteSdp = sdp;

    if (!parseSdp(sdp)) {
        LOG_ERROR("Failed to parse remote SDP");
        emit error("Failed to parse remote SDP");
        return;
    }

    if (type == "answer") {
        setConnectionState("connected");
        setIceConnectionState("connected");
    }
}

void PeerConnection::addIceCandidate(const QString &candidate, const QString &sdpMid, int sdpMLineIndex)
{
    LOG_INFO(QString("Adding ICE candidate for user: %1").arg(m_remoteUserId));

    m_remoteCandidates.append(candidate);

    // 解析候选地址
    // 格式: "candidate:1 1 UDP 2130706431 192.168.1.100 50000 typ host"
    QStringList parts = candidate.split(' ');
    if (parts.size() >= 6) {
        m_remoteAddress = QHostAddress(parts[4]);
        m_remoteRtpPort = parts[5].toUShort();
        m_remoteRtcpPort = m_remoteRtpPort + 1;

        LOG_INFO(QString("Remote address: %1:%2").arg(m_remoteAddress.toString()).arg(m_remoteRtpPort));

        setIceConnectionState("connected");
    }
}

QString PeerConnection::generateSdp(const QString &type)
{
    QString sdp;

    // Session description
    sdp += "v=0\r\n";
    sdp += QString("o=- %1 %2 IN IP4 %3\r\n")
           .arg(QDateTime::currentMSecsSinceEpoch())
           .arg(QDateTime::currentMSecsSinceEpoch())
           .arg(getLocalIpAddress());
    sdp += "s=Qt6 Meeting Session\r\n";
    sdp += "t=0 0\r\n";

    // Audio media description
    quint16 rtpPort = m_rtpSocket ? m_rtpSocket->localPort() : 0;
    sdp += QString("m=audio %1 RTP/AVP 0 8\r\n").arg(rtpPort);
    sdp += "c=IN IP4 " + getLocalIpAddress() + "\r\n";
    sdp += "a=rtpmap:0 PCMU/8000\r\n";
    sdp += "a=rtpmap:8 PCMA/8000\r\n";
    sdp += QString("a=%1\r\n").arg(type == "offer" ? "sendrecv" : "sendrecv");

    // Video media description
    sdp += QString("m=video %1 RTP/AVP 96\r\n").arg(rtpPort + 2);
    sdp += "c=IN IP4 " + getLocalIpAddress() + "\r\n";
    sdp += "a=rtpmap:96 H264/90000\r\n";
    sdp += "a=fmtp:96 profile-level-id=42e01f\r\n";
    sdp += QString("a=%1\r\n").arg(type == "offer" ? "sendrecv" : "sendrecv");

    return sdp;
}

bool PeerConnection::parseSdp(const QString &sdp)
{
    QStringList lines = sdp.split("\r\n", Qt::SkipEmptyParts);

    for (const QString &line : lines) {
        if (line.startsWith("m=audio")) {
            // 解析音频端口
            QStringList parts = line.split(' ');
            if (parts.size() >= 2) {
                // 端口信息在第二个字段
                // 注意：这里我们暂时不使用，因为会通过ICE候选获取
            }
        } else if (line.startsWith("m=video")) {
            // 解析视频端口
            QStringList parts = line.split(' ');
            if (parts.size() >= 2) {
                // 端口信息在第二个字段
            }
        } else if (line.startsWith("c=IN IP4")) {
            // 解析连接地址
            QStringList parts = line.split(' ');
            if (parts.size() >= 3) {
                // 地址信息在第三个字段
                // 注意：这里我们暂时不使用，因为会通过ICE候选获取
            }
        }
    }

    return true;
}

QString PeerConnection::getLocalIpAddress() const
{
    // 获取本地IP地址
    QList<QHostAddress> addresses = QNetworkInterface::allAddresses();

    for (const QHostAddress &address : addresses) {
        // 跳过回环地址和IPv6地址
        if (address.protocol() == QAbstractSocket::IPv4Protocol &&
            !address.isLoopback()) {
            return address.toString();
        }
    }

    return "127.0.0.1";
}

// RTP处理
QByteArray PeerConnection::createRtpHeader(quint8 payloadType, quint16 sequenceNumber,
                                           quint32 timestamp, quint32 ssrc)
{
    QByteArray header(12, 0);
    QDataStream stream(&header, QIODevice::WriteOnly);
    stream.setByteOrder(QDataStream::BigEndian);

    // RTP版本2, 无填充, 无扩展, 无CSRC
    quint8 byte0 = 0x80; // 10000000
    stream << byte0;

    // 标记位和负载类型
    quint8 byte1 = payloadType & 0x7F;
    stream << byte1;

    // 序列号
    stream << sequenceNumber;

    // 时间戳
    stream << timestamp;

    // SSRC
    stream << ssrc;

    return header;
}

void PeerConnection::sendRtpPacket(const QByteArray &payload, quint8 payloadType, quint32 timestamp)
{
    if (!m_rtpSocket || m_remoteAddress.isNull() || m_remoteRtpPort == 0) {
        return;
    }

    // 创建RTP头部
    QByteArray header = createRtpHeader(payloadType, m_rtpSequenceNumber, timestamp, m_rtpSsrc);

    // 组合头部和负载
    QByteArray packet = header + payload;

    // 发送数据包
    qint64 sent = m_rtpSocket->writeDatagram(packet, m_remoteAddress, m_remoteRtpPort);

    if (sent > 0) {
        m_statistics.bytesSent += sent;
        m_statistics.packetsSent++;
        m_rtpSequenceNumber++;
    }
}

void PeerConnection::processRtpPacket(const QByteArray &packet)
{
    if (packet.size() < 12) {
        return; // RTP头部至少12字节
    }

    QDataStream stream(packet);
    stream.setByteOrder(QDataStream::BigEndian);

    // 解析RTP头部
    quint8 byte0, byte1;
    quint16 sequenceNumber;
    quint32 timestamp, ssrc;

    stream >> byte0 >> byte1 >> sequenceNumber >> timestamp >> ssrc;

    quint8 version = (byte0 >> 6) & 0x03;
    quint8 payloadType = byte1 & 0x7F;

    if (version != 2) {
        return; // 只支持RTP版本2
    }

    // 提取负载数据
    QByteArray payload = packet.mid(12);

    // 更新统计信息
    m_statistics.bytesReceived += packet.size();
    m_statistics.packetsReceived++;

    // 根据负载类型处理数据
    if (payloadType == 96) {
        // 视频数据（H.264）
        // TODO: 解码视频帧
        // 暂时直接发送原始数据
        emit audioDataReceived(payload);
    } else if (payloadType == 0 || payloadType == 8) {
        // 音频数据（PCMU/PCMA）
        emit audioDataReceived(payload);
    }
}

// RTCP处理
void PeerConnection::sendRtcpSenderReport()
{
    if (!m_rtcpSocket || m_remoteAddress.isNull() || m_remoteRtcpPort == 0) {
        return;
    }

    // 创建RTCP SR包（简化版本）
    QByteArray packet(28, 0);
    QDataStream stream(&packet, QIODevice::WriteOnly);
    stream.setByteOrder(QDataStream::BigEndian);

    // RTCP头部
    quint8 byte0 = 0x80; // 版本2, 无填充, 0个接收报告块
    quint8 packetType = 200; // SR
    quint16 length = 6; // (28字节 - 4) / 4

    stream << byte0 << packetType << length;
    stream << m_rtpSsrc;

    // NTP时间戳
    quint64 ntpTime = QDateTime::currentMSecsSinceEpoch();
    stream << static_cast<quint32>(ntpTime >> 32);
    stream << static_cast<quint32>(ntpTime & 0xFFFFFFFF);

    // RTP时间戳
    stream << m_rtpTimestamp;

    // 发送的包数和字节数
    stream << static_cast<quint32>(m_statistics.packetsSent);
    stream << static_cast<quint32>(m_statistics.bytesSent);

    m_rtcpSocket->writeDatagram(packet, m_remoteAddress, m_remoteRtcpPort);
}

void PeerConnection::sendRtcpReceiverReport()
{
    // TODO: 实现接收者报告
}

void PeerConnection::processRtcpPacket(const QByteArray &packet)
{
    if (packet.size() < 8) {
        return;
    }

    QDataStream stream(packet);
    stream.setByteOrder(QDataStream::BigEndian);

    quint8 byte0, packetType;
    quint16 length;

    stream >> byte0 >> packetType >> length;

    // 处理不同类型的RTCP包
    if (packetType == 200) {
        // Sender Report
        LOG_DEBUG("Received RTCP SR");
    } else if (packetType == 201) {
        // Receiver Report
        LOG_DEBUG("Received RTCP RR");
    }
}

// STUN处理
void PeerConnection::performStunBinding()
{
    // TODO: 实现STUN绑定请求
    // 暂时使用本地地址
    m_publicAddress = QHostAddress(getLocalIpAddress());
    m_publicPort = m_rtpSocket ? m_rtpSocket->localPort() : 0;

    LOG_INFO(QString("Public address: %1:%2").arg(m_publicAddress.toString()).arg(m_publicPort));
}

void PeerConnection::processStunResponse(const QByteArray &response)
{
    // TODO: 解析STUN响应
}

// 槽函数
void PeerConnection::onRtpDataReceived()
{
    while (m_rtpSocket && m_rtpSocket->hasPendingDatagrams()) {
        QByteArray datagram;
        datagram.resize(m_rtpSocket->pendingDatagramSize());

        QHostAddress sender;
        quint16 senderPort;

        m_rtpSocket->readDatagram(datagram.data(), datagram.size(), &sender, &senderPort);

        processRtpPacket(datagram);
    }
}

void PeerConnection::onRtcpDataReceived()
{
    while (m_rtcpSocket && m_rtcpSocket->hasPendingDatagrams()) {
        QByteArray datagram;
        datagram.resize(m_rtcpSocket->pendingDatagramSize());

        QHostAddress sender;
        quint16 senderPort;

        m_rtcpSocket->readDatagram(datagram.data(), datagram.size(), &sender, &senderPort);

        processRtcpPacket(datagram);
    }
}

void PeerConnection::onLocalVideoFrame(const QVideoFrame &frame)
{
    // TODO: 编码视频帧并通过RTP发送
    // 暂时跳过，因为需要H.264编码器

    // 简化：直接更新时间戳
    m_rtpTimestamp += 3000; // 假设30fps，90kHz时钟
}

void PeerConnection::onLocalAudioData(const QByteArray &data)
{
    // 发送音频数据
    sendRtpPacket(data, 0, m_rtpTimestamp); // 负载类型0 = PCMU
    m_rtpTimestamp += data.size();
}

void PeerConnection::onRtcpTimeout()
{
    // 定期发送RTCP报告
    sendRtcpSenderReport();
}

void PeerConnection::onStunTimeout()
{
    // STUN超时处理
}

// 状态管理
void PeerConnection::setConnectionState(const QString &state)
{
    if (m_connectionState != state) {
        m_connectionState = state;
        LOG_INFO(QString("Connection state changed to: %1").arg(state));
        emit connectionStateChanged(state);
    }
}

void PeerConnection::setIceConnectionState(const QString &state)
{
    if (m_iceConnectionState != state) {
        m_iceConnectionState = state;
        LOG_INFO(QString("ICE connection state changed to: %1").arg(state));
        emit iceConnectionStateChanged(state);
    }
}

