# 客户端与服务端外部通信设计方案

## 📋 目录

1. [系统架构概览](#系统架构概览)
2. [通信协议栈](#通信协议栈)
3. [客户端类型与接入方式](#客户端类型与接入方式)
4. [API接口设计](#api接口设计)
5. [WebSocket信令通信](#websocket信令通信)
6. [WebRTC媒体通信](#webrtc媒体通信)
7. [认证与授权](#认证与授权)
8. [消息格式规范](#消息格式规范)
9. [错误处理](#错误处理)
10. [性能优化](#性能优化)

---

## 系统架构概览

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端层                                  │
├─────────────────────────────────────────────────────────────────┤
│  Qt6桌面客户端  │  Web浏览器客户端  │  移动端客户端(React Native) │
└────────┬────────┴────────┬──────────┴────────┬──────────────────┘
         │                 │                   │
         │    HTTP/HTTPS   │   WebSocket       │   WebRTC
         │                 │                   │
┌────────┴─────────────────┴───────────────────┴──────────────────┐
│                      Nginx API网关                                │
│  - 负载均衡  - SSL终止  - 限流  - 路由转发  - WebSocket代理      │
└────────┬─────────────────┬───────────────────┬──────────────────┘
         │                 │                   │
    ┌────┴────┐      ┌─────┴─────┐      ┌─────┴─────┐
    │用户服务  │      │会议服务    │      │信令服务    │
    │:8080    │      │:8082      │      │:8081      │
    └─────────┘      └───────────┘      └───────────┘
         │                 │                   │
    ┌────┴────┐      ┌─────┴─────┐      ┌─────┴─────┐
    │媒体服务  │      │AI服务     │      │通知服务    │
    │:8083    │      │:8084      │      │:8085      │
    └─────────┘      └───────────┘      └───────────┘
         │                 │
    ┌────┴─────────────────┴────┐
    │   内部gRPC服务间通信       │
    └────────────────────────────┘
```

### 通信层次

1. **外部通信层**（客户端 ↔ 服务端）
   - HTTP/HTTPS RESTful API
   - WebSocket 实时信令
   - WebRTC 音视频流

2. **网关层**（Nginx）
   - 请求路由与转发
   - 负载均衡
   - SSL/TLS终止
   - 限流与安全防护

3. **内部通信层**（微服务间）
   - gRPC 高性能RPC调用
   - ZMQ 与AI推理层通信
   - Redis 消息队列

---

## 通信协议栈

### 1. HTTP/HTTPS RESTful API

**用途**: 业务逻辑操作、资源管理

**特点**:
- 无状态请求
- 标准HTTP方法（GET, POST, PUT, DELETE）
- JSON数据格式
- JWT Token认证

**适用场景**:
- 用户注册/登录
- 会议创建/管理
- 文件上传/下载
- 配置查询

### 2. WebSocket 信令通信

**用途**: 实时双向通信、WebRTC信令交换

**特点**:
- 全双工通信
- 低延迟
- 持久连接
- 心跳保活

**适用场景**:
- WebRTC Offer/Answer交换
- ICE候选交换
- 实时聊天消息
- 用户状态通知
- 媒体控制信令

### 3. WebRTC 媒体流

**用途**: 点对点音视频传输

**特点**:
- P2P或SFU架构
- UDP传输
- 低延迟
- 自适应码率

**适用场景**:
- 音频通话
- 视频通话
- 屏幕共享
- 文件传输（DataChannel）

---

## 客户端类型与接入方式

### 1. Qt6 桌面客户端

**技术栈**:
- Qt6 + QML
- Qt Network (HTTP/WebSocket)
- Qt WebEngine (WebRTC)

**接入方式**:
```cpp
// HTTP请求
QNetworkAccessManager *manager = new QNetworkAccessManager();
QNetworkRequest request(QUrl("https://api.meeting.com/api/v1/users/login"));
request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
request.setRawHeader("Authorization", "Bearer " + token.toUtf8());

// WebSocket连接
QWebSocket *socket = new QWebSocket();
socket->open(QUrl("wss://api.meeting.com/ws/signaling?token=" + token));

// WebRTC (通过Qt WebEngine)
QWebEngineView *webView = new QWebEngineView();
```

### 2. Web 浏览器客户端

**技术栈**:
- HTML5 + JavaScript
- Fetch API / Axios (HTTP)
- WebSocket API
- WebRTC API

**接入方式**:
```javascript
// HTTP请求
const response = await fetch('https://api.meeting.com/api/v1/users/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({ username, password })
});

// WebSocket连接
const ws = new WebSocket('wss://api.meeting.com/ws/signaling?token=' + token);
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  handleSignalingMessage(message);
};

// WebRTC
const peerConnection = new RTCPeerConnection(config);
```

### 3. 移动端客户端 (React Native)

**技术栈**:
- React Native
- Axios (HTTP)
- react-native-webrtc
- WebSocket

**接入方式**:
```javascript
// HTTP请求
import axios from 'axios';
const response = await axios.post('https://api.meeting.com/api/v1/users/login', {
  username, password
}, {
  headers: { 'Authorization': `Bearer ${token}` }
});

// WebSocket
import { WebSocket } from 'react-native';
const ws = new WebSocket('wss://api.meeting.com/ws/signaling?token=' + token);

// WebRTC
import { RTCPeerConnection } from 'react-native-webrtc';
const pc = new RTCPeerConnection(config);
```

---

## API接口设计

### 基础URL

- **生产环境**: `https://api.meeting.com`
- **开发环境**: `http://localhost:80`

### 接口版本

当前版本: `v1`

所有API路径前缀: `/api/v1`

### 1. 用户服务 API

#### 1.1 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "user123",
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "张三"
}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "user123",
    "email": "user@example.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 1.2 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "user123",
  "password": "SecurePass123!"
}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "user123",
    "email": "user@example.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

#### 1.3 获取用户信息
```http
GET /api/v1/users/profile
Authorization: Bearer {token}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "user123",
    "email": "user@example.com",
    "full_name": "张三",
    "avatar_url": "https://cdn.meeting.com/avatars/user123.jpg",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

### 2. 会议服务 API

#### 2.1 创建会议
```http
POST /api/v1/meetings
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "项目讨论会",
  "description": "讨论Q1项目进度",
  "start_time": "2025-10-01T14:00:00Z",
  "duration": 60,
  "max_participants": 10,
  "is_public": false,
  "settings": {
    "enable_recording": true,
    "enable_ai_analysis": true,
    "require_password": true,
    "password": "meeting123"
  }
}

Response 201:
{
  "code": 201,
  "message": "success",
  "data": {
    "meeting_id": 100,
    "title": "项目讨论会",
    "meeting_code": "ABC-DEF-GHI",
    "join_url": "https://meeting.com/join/ABC-DEF-GHI",
    "host_id": 1,
    "status": "scheduled",
    "created_at": "2025-10-01T10:00:00Z"
  }
}
```

#### 2.2 加入会议
```http
POST /api/v1/meetings/{meeting_id}/join
Authorization: Bearer {token}
Content-Type: application/json

{
  "password": "meeting123"
}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "meeting_id": 100,
    "participant_id": 50,
    "websocket_url": "wss://api.meeting.com/ws/signaling",
    "ice_servers": [
      {
        "urls": "stun:stun.l.google.com:19302"
      },
      {
        "urls": "turn:turn.meeting.com:3478",
        "username": "user123",
        "credential": "temp_credential"
      }
    ],
    "session_token": "session_token_here"
  }
}
```

#### 2.3 获取会议列表
```http
GET /api/v1/meetings?status=active&page=1&page_size=20
Authorization: Bearer {token}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "meetings": [
      {
        "meeting_id": 100,
        "title": "项目讨论会",
        "host_name": "张三",
        "start_time": "2025-10-01T14:00:00Z",
        "status": "active",
        "participant_count": 5
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### 3. 媒体服务 API

#### 3.1 上传文件
```http
POST /api/v1/media/upload
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: [binary data]
meeting_id: 100
file_type: document

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "file_id": "file_abc123",
    "file_name": "document.pdf",
    "file_url": "https://cdn.meeting.com/files/file_abc123.pdf",
    "file_size": 1024000,
    "uploaded_at": "2025-10-01T10:00:00Z"
  }
}
```

### 4. AI服务 API

#### 4.1 语音识别
```http
POST /api/v1/ai/speech/recognition
Authorization: Bearer {token}
Content-Type: application/json

{
  "audio_data": "base64_encoded_audio_data",
  "language": "zh",
  "format": "wav",
  "sample_rate": 16000
}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "text": "这是识别出的文本内容",
    "confidence": 0.95,
    "language": "zh",
    "duration": 5.2
  }
}
```

---

## WebSocket信令通信

### 连接建立

#### 连接URL
```
wss://api.meeting.com/ws/signaling?token={jwt_token}&meeting_id={meeting_id}&user_id={user_id}&peer_id={peer_id}
```

#### 查询参数
- `token`: JWT认证令牌（必需）
- `meeting_id`: 会议ID（必需）
- `user_id`: 用户ID（必需）
- `peer_id`: WebRTC Peer ID（必需）

#### 连接示例
```javascript
const token = localStorage.getItem('auth_token');
const meetingId = 100;
const userId = 1;
const peerId = generatePeerId(); // 生成唯一的Peer ID

const wsUrl = `wss://api.meeting.com/ws/signaling?token=${token}&meeting_id=${meetingId}&user_id=${userId}&peer_id=${peerId}`;
const ws = new WebSocket(wsUrl);

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  handleMessage(message);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket closed');
};
```

### 消息格式

所有WebSocket消息使用统一的JSON格式：

```json
{
  "id": "msg_unique_id",
  "type": 1,
  "from_user_id": 1,
  "to_user_id": 2,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "peer_id": "peer_xyz789",
  "payload": {},
  "timestamp": "2025-10-01T10:00:00Z"
}
```

### 消息类型

| 类型值 | 类型名称 | 说明 |
|-------|---------|------|
| 1 | offer | WebRTC Offer |
| 2 | answer | WebRTC Answer |
| 3 | ice-candidate | ICE候选 |
| 4 | join-room | 加入房间 |
| 5 | leave-room | 离开房间 |
| 6 | user-joined | 用户加入通知 |
| 7 | user-left | 用户离开通知 |
| 8 | chat | 聊天消息 |
| 9 | screen-share | 屏幕共享 |
| 10 | media-control | 媒体控制 |
| 11 | ping | 心跳 |
| 12 | pong | 心跳响应 |
| 13 | error | 错误消息 |
| 14 | room-info | 房间信息 |

### 消息示例

#### 1. 加入房间 (join-room)

**客户端发送**:
```json
{
  "id": "msg_001",
  "type": 4,
  "from_user_id": 1,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "peer_id": "peer_xyz789",
  "payload": {
    "meeting_id": 100,
    "user_id": 1,
    "peer_id": "peer_xyz789"
  },
  "timestamp": "2025-10-01T10:00:00Z"
}
```

**服务端响应**:
```json
{
  "id": "msg_002",
  "type": 14,
  "from_user_id": 0,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "peer_id": "peer_xyz789",
  "payload": {
    "meeting_id": 100,
    "participant_count": 3,
    "session_id": "session_abc123",
    "peer_id": "peer_xyz789",
    "ice_servers": [
      {
        "urls": "stun:stun.l.google.com:19302"
      }
    ],
    "participants": [
      {
        "user_id": 2,
        "username": "user2",
        "session_id": "session_def456",
        "peer_id": "peer_abc123",
        "joined_at": "2025-10-01T09:50:00Z",
        "last_active_at": "2025-10-01T09:59:00Z",
        "is_self": false
      }
    ]
  },
  "timestamp": "2025-10-01T10:00:01Z"
}
```

#### 2. WebRTC Offer

**客户端A发送给客户端B**:
```json
{
  "id": "msg_003",
  "type": 1,
  "from_user_id": 1,
  "to_user_id": 2,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "peer_id": "peer_xyz789",
  "payload": {
    "sdp": "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...",
    "type": "offer"
  },
  "timestamp": "2025-10-01T10:00:02Z"
}
```

#### 3. WebRTC Answer

**客户端B响应给客户端A**:
```json
{
  "id": "msg_004",
  "type": 2,
  "from_user_id": 2,
  "to_user_id": 1,
  "meeting_id": 100,
  "session_id": "session_def456",
  "peer_id": "peer_abc123",
  "payload": {
    "sdp": "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n...",
    "type": "answer"
  },
  "timestamp": "2025-10-01T10:00:03Z"
}
```

#### 4. ICE候选

```json
{
  "id": "msg_005",
  "type": 3,
  "from_user_id": 1,
  "to_user_id": 2,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "peer_id": "peer_xyz789",
  "payload": {
    "candidate": "candidate:1 1 UDP 2130706431 192.168.1.100 54321 typ host",
    "sdpMid": "0",
    "sdpMLineIndex": 0
  },
  "timestamp": "2025-10-01T10:00:04Z"
}
```

#### 5. 聊天消息

```json
{
  "id": "msg_006",
  "type": 8,
  "from_user_id": 1,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "payload": {
    "content": "大家好！",
    "user_id": 1,
    "username": "张三",
    "meeting_id": 100
  },
  "timestamp": "2025-10-01T10:00:05Z"
}
```

#### 6. 媒体控制

```json
{
  "id": "msg_007",
  "type": 10,
  "from_user_id": 1,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "peer_id": "peer_xyz789",
  "payload": {
    "action": "mute",
    "media_type": "audio",
    "user_id": 1,
    "peer_id": "peer_xyz789"
  },
  "timestamp": "2025-10-01T10:00:06Z"
}
```

#### 7. 心跳 (ping/pong)

**客户端发送**:
```json
{
  "id": "msg_008",
  "type": 11,
  "from_user_id": 1,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "payload": {},
  "timestamp": "2025-10-01T10:00:07Z"
}
```

**服务端响应**:
```json
{
  "id": "msg_009",
  "type": 12,
  "from_user_id": 0,
  "meeting_id": 100,
  "session_id": "session_abc123",
  "payload": {},
  "timestamp": "2025-10-01T10:00:07Z"
}
```

### 心跳机制

- **客户端**: 每30秒发送一次ping消息
- **服务端**: 收到ping后立即返回pong
- **超时检测**: 如果60秒内未收到任何消息，服务端将断开连接
- **重连机制**: 客户端检测到断开后，应在3秒后尝试重连

```javascript
// 客户端心跳实现
let heartbeatInterval;
let lastPongTime = Date.now();

function startHeartbeat(ws) {
  heartbeatInterval = setInterval(() => {
    if (Date.now() - lastPongTime > 60000) {
      console.error('Heartbeat timeout, reconnecting...');
      ws.close();
      reconnect();
      return;
    }

    ws.send(JSON.stringify({
      id: generateMessageId(),
      type: 11, // ping
      from_user_id: currentUserId,
      meeting_id: currentMeetingId,
      session_id: sessionId,
      payload: {},
      timestamp: new Date().toISOString()
    }));
  }, 30000);
}

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 12) { // pong
    lastPongTime = Date.now();
  }
  // ... 处理其他消息
};
```

---

## WebRTC媒体通信

### SFU架构

本系统采用SFU（Selective Forwarding Unit）架构，服务端负责转发媒体流，不进行转码。

```
客户端A ──────┐
              │
客户端B ──────┤──> SFU服务器 ──┬──> 客户端A
              │                ├──> 客户端B
客户端C ──────┘                └──> 客户端C
```

### WebRTC连接流程

#### 1. 初始化PeerConnection

```javascript
// ICE服务器配置（从加入会议API获取）
const iceServers = [
  { urls: 'stun:stun.l.google.com:19302' },
  {
    urls: 'turn:turn.meeting.com:3478',
    username: 'user123',
    credential: 'temp_credential'
  }
];

// 创建PeerConnection
const peerConnection = new RTCPeerConnection({
  iceServers: iceServers,
  iceTransportPolicy: 'all',
  bundlePolicy: 'max-bundle',
  rtcpMuxPolicy: 'require'
});

// 监听ICE候选
peerConnection.onicecandidate = (event) => {
  if (event.candidate) {
    // 通过WebSocket发送ICE候选给对方
    sendSignalingMessage({
      type: 3, // ice-candidate
      to_user_id: remoteUserId,
      payload: {
        candidate: event.candidate.candidate,
        sdpMid: event.candidate.sdpMid,
        sdpMLineIndex: event.candidate.sdpMLineIndex
      }
    });
  }
};

// 监听远程流
peerConnection.ontrack = (event) => {
  const remoteVideo = document.getElementById('remote-video-' + remoteUserId);
  remoteVideo.srcObject = event.streams[0];
};
```

#### 2. 添加本地媒体流

```javascript
// 获取本地媒体流
const localStream = await navigator.mediaDevices.getUserMedia({
  audio: {
    echoCancellation: true,
    noiseSuppression: true,
    autoGainControl: true
  },
  video: {
    width: { ideal: 1280 },
    height: { ideal: 720 },
    frameRate: { ideal: 30 }
  }
});

// 显示本地视频
const localVideo = document.getElementById('local-video');
localVideo.srcObject = localStream;

// 添加轨道到PeerConnection
localStream.getTracks().forEach(track => {
  peerConnection.addTrack(track, localStream);
});
```

#### 3. 创建并发送Offer

```javascript
// 创建Offer
const offer = await peerConnection.createOffer({
  offerToReceiveAudio: true,
  offerToReceiveVideo: true
});

// 设置本地描述
await peerConnection.setLocalDescription(offer);

// 通过WebSocket发送Offer
sendSignalingMessage({
  type: 1, // offer
  to_user_id: remoteUserId,
  payload: {
    sdp: offer.sdp,
    type: 'offer'
  }
});
```

#### 4. 接收Offer并发送Answer

```javascript
// 接收到Offer
async function handleOffer(message) {
  const offer = message.payload;

  // 设置远程描述
  await peerConnection.setRemoteDescription(
    new RTCSessionDescription({
      type: 'offer',
      sdp: offer.sdp
    })
  );

  // 创建Answer
  const answer = await peerConnection.createAnswer();

  // 设置本地描述
  await peerConnection.setLocalDescription(answer);

  // 发送Answer
  sendSignalingMessage({
    type: 2, // answer
    to_user_id: message.from_user_id,
    payload: {
      sdp: answer.sdp,
      type: 'answer'
    }
  });
}
```

#### 5. 接收Answer

```javascript
async function handleAnswer(message) {
  const answer = message.payload;

  // 设置远程描述
  await peerConnection.setRemoteDescription(
    new RTCSessionDescription({
      type: 'answer',
      sdp: answer.sdp
    })
  );
}
```

#### 6. 处理ICE候选

```javascript
async function handleIceCandidate(message) {
  const candidate = message.payload;

  await peerConnection.addIceCandidate(
    new RTCIceCandidate({
      candidate: candidate.candidate,
      sdpMid: candidate.sdpMid,
      sdpMLineIndex: candidate.sdpMLineIndex
    })
  );
}
```

### 屏幕共享

```javascript
// 开始屏幕共享
async function startScreenShare() {
  try {
    const screenStream = await navigator.mediaDevices.getDisplayMedia({
      video: {
        cursor: 'always',
        displaySurface: 'monitor'
      },
      audio: false
    });

    // 替换视频轨道
    const videoTrack = screenStream.getVideoTracks()[0];
    const sender = peerConnection.getSenders().find(s =>
      s.track && s.track.kind === 'video'
    );

    if (sender) {
      await sender.replaceTrack(videoTrack);
    }

    // 监听屏幕共享停止
    videoTrack.onended = () => {
      stopScreenShare();
    };

    // 通知其他用户
    sendSignalingMessage({
      type: 9, // screen-share
      payload: {
        action: 'start',
        user_id: currentUserId
      }
    });
  } catch (error) {
    console.error('Screen share error:', error);
  }
}

// 停止屏幕共享
async function stopScreenShare() {
  // 恢复摄像头视频
  const cameraStream = await navigator.mediaDevices.getUserMedia({
    video: true
  });

  const videoTrack = cameraStream.getVideoTracks()[0];
  const sender = peerConnection.getSenders().find(s =>
    s.track && s.track.kind === 'video'
  );

  if (sender) {
    await sender.replaceTrack(videoTrack);
  }

  // 通知其他用户
  sendSignalingMessage({
    type: 9, // screen-share
    payload: {
      action: 'stop',
      user_id: currentUserId
    }
  });
}
```

### 媒体控制

```javascript
// 静音/取消静音
function toggleAudio(muted) {
  localStream.getAudioTracks().forEach(track => {
    track.enabled = !muted;
  });

  // 通知其他用户
  sendSignalingMessage({
    type: 10, // media-control
    payload: {
      action: muted ? 'mute' : 'unmute',
      media_type: 'audio',
      user_id: currentUserId,
      peer_id: peerId
    }
  });
}

// 开启/关闭视频
function toggleVideo(enabled) {
  localStream.getVideoTracks().forEach(track => {
    track.enabled = enabled;
  });

  // 通知其他用户
  sendSignalingMessage({
    type: 10, // media-control
    payload: {
      action: enabled ? 'video_on' : 'video_off',
      media_type: 'video',
      user_id: currentUserId,
      peer_id: peerId
    }
  });
}
```

---

## 认证与授权

### JWT Token认证

#### Token结构

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": 1,
    "username": "user123",
    "email": "user@example.com",
    "iat": 1696147200,
    "exp": 1696233600,
    "nbf": 1696147200,
    "iss": "meeting-system",
    "sub": "1"
  },
  "signature": "..."
}
```

#### Token使用

**HTTP请求**:
```http
GET /api/v1/users/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**WebSocket连接**:
```
wss://api.meeting.com/ws/signaling?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Token刷新

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

Response 200:
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "new_access_token",
    "refresh_token": "new_refresh_token",
    "expires_in": 86400
  }
}
```

### 权限控制

#### 会议权限

| 角色 | 权限 |
|-----|------|
| 主持人 | 创建/结束会议、踢出参与者、静音所有人、录制控制 |
| 联席主持人 | 静音参与者、管理屏幕共享 |
| 普通参与者 | 发言、共享屏幕（需授权）、聊天 |
| 观众 | 仅观看、聊天 |

#### 权限验证流程

```javascript
// 客户端请求操作
async function kickParticipant(participantId) {
  try {
    const response = await fetch(`/api/v1/meetings/${meetingId}/participants/${participantId}/kick`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    if (response.status === 403) {
      alert('您没有权限执行此操作');
      return;
    }

    const result = await response.json();
    console.log('Participant kicked:', result);
  } catch (error) {
    console.error('Error:', error);
  }
}
```

---

## 消息格式规范

### 统一响应格式

所有HTTP API响应使用统一格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {},
  "timestamp": "2025-10-01T10:00:00Z"
}
```

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "user123"
  },
  "timestamp": "2025-10-01T10:00:00Z"
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "Invalid request",
  "error": {
    "type": "ValidationError",
    "details": "Username is required"
  },
  "timestamp": "2025-10-01T10:00:00Z"
}
```

### 分页响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20,
    "total_pages": 5
  },
  "timestamp": "2025-10-01T10:00:00Z"
}
```

---

## 错误处理

### HTTP状态码

| 状态码 | 说明 | 示例 |
|-------|------|------|
| 200 | 成功 | 请求成功处理 |
| 201 | 创建成功 | 资源创建成功 |
| 400 | 请求错误 | 参数验证失败 |
| 401 | 未认证 | Token无效或过期 |
| 403 | 无权限 | 没有操作权限 |
| 404 | 未找到 | 资源不存在 |
| 409 | 冲突 | 资源已存在 |
| 429 | 请求过多 | 触发限流 |
| 500 | 服务器错误 | 内部错误 |
| 503 | 服务不可用 | 服务维护中 |

### 错误码定义

| 错误码 | 说明 |
|-------|------|
| 10001 | 用户名或密码错误 |
| 10002 | 用户不存在 |
| 10003 | 用户已存在 |
| 10004 | Token无效 |
| 10005 | Token过期 |
| 20001 | 会议不存在 |
| 20002 | 会议已结束 |
| 20003 | 会议人数已满 |
| 20004 | 会议密码错误 |
| 30001 | 文件上传失败 |
| 30002 | 文件格式不支持 |
| 30003 | 文件大小超限 |
| 40001 | AI服务不可用 |
| 40002 | AI处理失败 |

### 客户端错误处理

```javascript
async function apiRequest(url, options) {
  try {
    const response = await fetch(url, options);
    const result = await response.json();

    if (response.ok) {
      return result.data;
    }

    // 处理错误
    switch (response.status) {
      case 401:
        // Token过期，刷新Token
        await refreshToken();
        return apiRequest(url, options); // 重试

      case 403:
        alert('您没有权限执行此操作');
        break;

      case 404:
        alert('请求的资源不存在');
        break;

      case 429:
        alert('请求过于频繁，请稍后再试');
        break;

      case 500:
        alert('服务器错误，请稍后再试');
        break;

      default:
        alert(result.message || '请求失败');
    }

    throw new Error(result.message);
  } catch (error) {
    console.error('API request error:', error);
    throw error;
  }
}
```

---

## 性能优化

### 1. 连接优化

#### HTTP/2支持
- 多路复用
- 头部压缩
- 服务器推送

#### Keep-Alive
```http
Connection: keep-alive
Keep-Alive: timeout=60, max=1000
```

### 2. 数据压缩

#### Gzip压缩
```http
Accept-Encoding: gzip, deflate, br
Content-Encoding: gzip
```

### 3. 缓存策略

#### 静态资源缓存
```http
Cache-Control: public, max-age=31536000, immutable
```

#### API响应缓存
```http
Cache-Control: private, max-age=60
ETag: "abc123"
```

### 4. WebSocket优化

- 消息批处理
- 二进制传输（Protobuf）
- 消息压缩

### 5. WebRTC优化

#### 自适应码率
```javascript
const sender = peerConnection.getSenders().find(s => s.track.kind === 'video');
const parameters = sender.getParameters();

if (!parameters.encodings) {
  parameters.encodings = [{}];
}

// 设置最大码率
parameters.encodings[0].maxBitrate = 1000000; // 1 Mbps

await sender.setParameters(parameters);
```

#### 网络质量监控
```javascript
setInterval(async () => {
  const stats = await peerConnection.getStats();
  stats.forEach(report => {
    if (report.type === 'inbound-rtp' && report.kind === 'video') {
      console.log('Packets lost:', report.packetsLost);
      console.log('Jitter:', report.jitter);
      console.log('Bitrate:', report.bytesReceived * 8 / report.timestamp);
    }
  });
}, 1000);
```

### 6. 限流与防护

#### 客户端限流
```javascript
// 使用防抖
const debouncedSend = debounce((message) => {
  ws.send(JSON.stringify(message));
}, 100);

// 使用节流
const throttledSend = throttle((message) => {
  ws.send(JSON.stringify(message));
}, 1000);
```

#### 服务端限流
- Nginx限流配置
- 令牌桶算法
- 滑动窗口算法

---

## 安全性

### 1. HTTPS/WSS

所有外部通信必须使用加密传输：
- HTTP → HTTPS (TLS 1.2+)
- WS → WSS (TLS 1.2+)

### 2. CORS配置

```javascript
// 服务端CORS配置
app.use(cors({
  origin: ['https://meeting.com', 'https://app.meeting.com'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE'],
  allowedHeaders: ['Content-Type', 'Authorization']
}));
```

### 3. XSS防护

- 输入验证
- 输出编码
- CSP策略

```http
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'
```

### 4. CSRF防护

- CSRF Token
- SameSite Cookie

### 5. 数据加密

- 敏感数据加密存储
- 传输层加密（TLS）
- 端到端加密（WebRTC DTLS-SRTP）

---

## 完整示例

### Web客户端完整实现

```javascript
class MeetingClient {
  constructor(apiUrl, wsUrl) {
    this.apiUrl = apiUrl;
    this.wsUrl = wsUrl;
    this.token = null;
    this.ws = null;
    this.peerConnections = new Map();
    this.localStream = null;
  }

  // 登录
  async login(username, password) {
    const response = await fetch(`${this.apiUrl}/api/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const result = await response.json();
    if (result.code === 200) {
      this.token = result.data.token;
      return result.data;
    }
    throw new Error(result.message);
  }

  // 加入会议
  async joinMeeting(meetingId, password) {
    const response = await fetch(`${this.apiUrl}/api/v1/meetings/${meetingId}/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`
      },
      body: JSON.stringify({ password })
    });

    const result = await response.json();
    if (result.code === 200) {
      await this.connectWebSocket(meetingId, result.data);
      await this.setupLocalMedia();
      return result.data;
    }
    throw new Error(result.message);
  }

  // 连接WebSocket
  async connectWebSocket(meetingId, joinData) {
    const userId = 1; // 从token解析
    const peerId = this.generatePeerId();

    const wsUrl = `${this.wsUrl}?token=${this.token}&meeting_id=${meetingId}&user_id=${userId}&peer_id=${peerId}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.handleSignalingMessage(message);
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket closed');
      this.reconnect();
    };
  }

  // 设置本地媒体
  async setupLocalMedia() {
    this.localStream = await navigator.mediaDevices.getUserMedia({
      audio: true,
      video: true
    });

    const localVideo = document.getElementById('local-video');
    localVideo.srcObject = this.localStream;
  }

  // 处理信令消息
  async handleSignalingMessage(message) {
    switch (message.type) {
      case 1: // offer
        await this.handleOffer(message);
        break;
      case 2: // answer
        await this.handleAnswer(message);
        break;
      case 3: // ice-candidate
        await this.handleIceCandidate(message);
        break;
      case 6: // user-joined
        await this.handleUserJoined(message);
        break;
      case 7: // user-left
        this.handleUserLeft(message);
        break;
      case 8: // chat
        this.handleChat(message);
        break;
      case 14: // room-info
        this.handleRoomInfo(message);
        break;
    }
  }

  // 创建PeerConnection
  createPeerConnection(remoteUserId, iceServers) {
    const pc = new RTCPeerConnection({ iceServers });

    pc.onicecandidate = (event) => {
      if (event.candidate) {
        this.sendSignalingMessage({
          type: 3,
          to_user_id: remoteUserId,
          payload: {
            candidate: event.candidate.candidate,
            sdpMid: event.candidate.sdpMid,
            sdpMLineIndex: event.candidate.sdpMLineIndex
          }
        });
      }
    };

    pc.ontrack = (event) => {
      const remoteVideo = document.getElementById(`remote-video-${remoteUserId}`);
      if (remoteVideo) {
        remoteVideo.srcObject = event.streams[0];
      }
    };

    // 添加本地流
    this.localStream.getTracks().forEach(track => {
      pc.addTrack(track, this.localStream);
    });

    this.peerConnections.set(remoteUserId, pc);
    return pc;
  }

  // 发送信令消息
  sendSignalingMessage(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        id: this.generateMessageId(),
        ...message,
        from_user_id: 1, // 当前用户ID
        meeting_id: this.currentMeetingId,
        session_id: this.sessionId,
        timestamp: new Date().toISOString()
      }));
    }
  }

  // 工具方法
  generatePeerId() {
    return 'peer_' + Math.random().toString(36).substr(2, 9);
  }

  generateMessageId() {
    return 'msg_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
  }

  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      this.sendSignalingMessage({ type: 11, payload: {} });
    }, 30000);
  }
}

// 使用示例
const client = new MeetingClient(
  'https://api.meeting.com',
  'wss://api.meeting.com/ws/signaling'
);

// 登录并加入会议
async function start() {
  try {
    await client.login('user123', 'password');
    await client.joinMeeting(100, 'meeting123');
  } catch (error) {
    console.error('Error:', error);
  }
}

start();
```

---

## 总结

本文档详细描述了智能视频会议平台的客户端与服务端外部通信方案，包括：

1. **多协议支持**: HTTP/HTTPS、WebSocket、WebRTC
2. **多客户端支持**: Qt6桌面、Web浏览器、移动端
3. **完整的API设计**: RESTful API、WebSocket信令、WebRTC媒体
4. **安全认证**: JWT Token、权限控制
5. **性能优化**: 连接复用、数据压缩、自适应码率
6. **错误处理**: 统一错误码、重试机制
7. **实时通信**: 心跳保活、断线重连

该方案确保了系统的**高可用性**、**低延迟**、**高安全性**和**良好的用户体验**。


