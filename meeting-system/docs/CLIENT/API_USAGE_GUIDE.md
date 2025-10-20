# 客户端API调用指南

## 📋 目录

1. [API基础](#api基础)
2. [认证流程](#认证流程)
3. [用户管理](#用户管理)
4. [会议管理](#会议管理)
5. [WebSocket信令](#websocket信令)
6. [媒体服务](#媒体服务)
7. [AI服务](#ai服务)
8. [错误处理](#错误处理)
9. [代码示例](#代码示例)

---

## API基础

### 基础URL

- **开发环境**: `http://localhost`
- **生产环境**: `https://api.meeting.com`

### 请求格式

所有API请求使用JSON格式：

```http
POST /api/v1/auth/login
Content-Type: application/json
Authorization: Bearer <token>  (需要认证的接口)

{
  "username": "user123",
  "password": "password123"
}
```

### 响应格式

成功响应：
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    // 响应数据
  },
  "timestamp": "2025-10-02T10:00:00Z"
}
```

错误响应：
```json
{
  "code": 400,
  "message": "Invalid request",
  "error": "Detailed error message",
  "timestamp": "2025-10-02T10:00:00Z",
  "request_id": "abc123"
}
```

---

## 认证流程

### 1. 用户注册

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123!",
  "full_name": "测试用户"
}
```

**响应**:
```json
{
  "code": 201,
  "message": "User registered successfully",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "测试用户",
    "created_at": "2025-10-02T10:00:00Z"
  }
}
```

### 2. 用户登录

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "SecurePass123!"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "user_id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "full_name": "测试用户"
    }
  }
}
```

### 3. 刷新Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 4. 登出

```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

---

## 用户管理

### 1. 获取用户资料

```http
GET /api/v1/users/profile
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "测试用户",
    "avatar_url": "https://example.com/avatar.jpg",
    "created_at": "2025-10-02T10:00:00Z",
    "updated_at": "2025-10-02T10:00:00Z"
  }
}
```

### 2. 更新用户资料

```http
PUT /api/v1/users/profile
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "full_name": "新名字",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

### 3. 修改密码

```http
PUT /api/v1/users/password
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "old_password": "OldPass123!",
  "new_password": "NewPass123!"
}
```

---

## 会议管理

### 1. 创建会议

```http
POST /api/v1/meetings
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "title": "团队周会",
  "description": "讨论本周工作进展",
  "start_time": "2025-10-03T10:00:00Z",
  "end_time": "2025-10-03T11:00:00Z",
  "max_participants": 10,
  "meeting_type": "video",
  "is_recording_enabled": true,
  "password": "123456"
}
```

**响应**:
```json
{
  "code": 201,
  "data": {
    "meeting_id": 100,
    "title": "团队周会",
    "meeting_code": "ABC-DEF-GHI",
    "host_id": 1,
    "start_time": "2025-10-03T10:00:00Z",
    "end_time": "2025-10-03T11:00:00Z",
    "status": "scheduled",
    "join_url": "http://localhost/meeting/100"
  }
}
```

### 2. 获取会议列表

```http
GET /api/v1/meetings?status=scheduled&page=1&page_size=10
Authorization: Bearer <access_token>
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "meetings": [
      {
        "meeting_id": 100,
        "title": "团队周会",
        "start_time": "2025-10-03T10:00:00Z",
        "status": "scheduled",
        "participants_count": 0
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

### 3. 获取会议详情

```http
GET /api/v1/meetings/100
Authorization: Bearer <access_token>
```

### 4. 加入会议

```http
POST /api/v1/meetings/100/join
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "password": "123456"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "meeting_id": 100,
    "participant_id": 1,
    "peer_id": "peer_abc123",
    "signaling_url": "ws://localhost/ws/signaling?meeting_id=100&user_id=1&peer_id=peer_abc123",
    "ice_servers": [
      {
        "urls": "stun:stun.l.google.com:19302"
      }
    ]
  }
}
```

### 5. 离开会议

```http
POST /api/v1/meetings/100/leave
Authorization: Bearer <access_token>
```

### 6. 结束会议

```http
POST /api/v1/meetings/100/end
Authorization: Bearer <access_token>
```

---

## WebSocket信令

### 1. 建立连接

```javascript
// 连接URL格式
const wsUrl = `ws://localhost/ws/signaling?meeting_id=${meetingId}&user_id=${userId}&peer_id=${peerId}`;
const ws = new WebSocket(wsUrl);

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  handleSignalingMessage(message);
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket closed');
};
```

### 2. 发送信令消息

#### 加入房间
```javascript
ws.send(JSON.stringify({
  type: 'join',
  meeting_id: 100,
  user_id: 1,
  peer_id: 'peer_abc123'
}));
```

#### 发送Offer
```javascript
ws.send(JSON.stringify({
  type: 'offer',
  from_peer_id: 'peer_abc123',
  to_peer_id: 'peer_def456',
  sdp: offerSdp
}));
```

#### 发送Answer
```javascript
ws.send(JSON.stringify({
  type: 'answer',
  from_peer_id: 'peer_abc123',
  to_peer_id: 'peer_def456',
  sdp: answerSdp
}));
```

#### 发送ICE候选
```javascript
ws.send(JSON.stringify({
  type: 'ice-candidate',
  from_peer_id: 'peer_abc123',
  to_peer_id: 'peer_def456',
  candidate: iceCandidate
}));
```

#### 发送聊天消息
```javascript
ws.send(JSON.stringify({
  type: 'chat',
  meeting_id: 100,
  user_id: 1,
  message: 'Hello everyone!'
}));
```

### 3. 接收信令消息

```javascript
function handleSignalingMessage(message) {
  switch (message.type) {
    case 'user-joined':
      console.log('User joined:', message.user_id);
      break;
      
    case 'user-left':
      console.log('User left:', message.user_id);
      break;
      
    case 'offer':
      handleOffer(message.from_peer_id, message.sdp);
      break;
      
    case 'answer':
      handleAnswer(message.from_peer_id, message.sdp);
      break;
      
    case 'ice-candidate':
      handleIceCandidate(message.from_peer_id, message.candidate);
      break;
      
    case 'chat':
      displayChatMessage(message.user_id, message.message);
      break;
      
    case 'error':
      console.error('Signaling error:', message.error);
      break;
  }
}
```

---

## 媒体服务

### 1. 上传文件

```http
POST /api/v1/media/upload
Authorization: Bearer <access_token>
Content-Type: multipart/form-data

file: <binary data>
meeting_id: 100
file_type: document
```

### 2. 下载文件

```http
GET /api/v1/media/files/123/download
Authorization: Bearer <access_token>
```

### 3. 获取文件列表

```http
GET /api/v1/media/files?meeting_id=100
Authorization: Bearer <access_token>
```

---

## AI服务

### 1. 语音识别

```http
POST /api/v1/ai/speech/recognize
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "audio_data": "base64_encoded_audio",
  "audio_format": "pcm",
  "sample_rate": 16000,
  "language": "zh"
}
```

### 2. 情绪检测

```http
POST /api/v1/ai/emotion/detect
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "audio_data": "base64_encoded_audio",
  "audio_format": "wav",
  "sample_rate": 16000
}
```

### 3. 音频增强

```http
POST /api/v1/ai/audio/enhance
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "audio_data": "base64_encoded_audio",
  "audio_format": "wav",
  "enhancement_type": "denoise"
}
```

---

## 错误处理

### HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 资源创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或Token无效 |
| 403 | 无权限访问 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

### 业务错误码

| 错误码 | 说明 |
|--------|------|
| 1001 | 用户名已存在 |
| 1002 | 邮箱已存在 |
| 1003 | 用户不存在 |
| 1004 | 密码错误 |
| 2001 | 会议不存在 |
| 2002 | 会议已结束 |
| 2003 | 会议人数已满 |
| 2004 | 会议密码错误 |
| 3001 | 文件上传失败 |
| 3002 | 文件不存在 |
| 4001 | AI服务不可用 |
| 4002 | AI处理失败 |

---

## 代码示例

### Qt6 C++ 示例

```cpp
// API客户端类
class APIClient : public QObject {
    Q_OBJECT
    
public:
    APIClient(const QString& baseUrl, QObject* parent = nullptr)
        : QObject(parent), m_baseUrl(baseUrl) {
        m_manager = new QNetworkAccessManager(this);
    }
    
    // 登录
    void login(const QString& username, const QString& password) {
        QJsonObject json;
        json["username"] = username;
        json["password"] = password;
        
        QNetworkRequest request(QUrl(m_baseUrl + "/api/v1/auth/login"));
        request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
        
        QNetworkReply* reply = m_manager->post(request, QJsonDocument(json).toJson());
        connect(reply, &QNetworkReply::finished, this, [this, reply]() {
            if (reply->error() == QNetworkReply::NoError) {
                QJsonDocument doc = QJsonDocument::fromJson(reply->readAll());
                QJsonObject obj = doc.object();
                m_token = obj["data"].toObject()["access_token"].toString();
                emit loginSuccess();
            } else {
                emit loginFailed(reply->errorString());
            }
            reply->deleteLater();
        });
    }
    
    // 创建会议
    void createMeeting(const QString& title, const QDateTime& startTime) {
        QJsonObject json;
        json["title"] = title;
        json["start_time"] = startTime.toString(Qt::ISODate);
        json["meeting_type"] = "video";
        
        QNetworkRequest request(QUrl(m_baseUrl + "/api/v1/meetings"));
        request.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
        request.setRawHeader("Authorization", ("Bearer " + m_token).toUtf8());
        
        QNetworkReply* reply = m_manager->post(request, QJsonDocument(json).toJson());
        connect(reply, &QNetworkReply::finished, this, [this, reply]() {
            if (reply->error() == QNetworkReply::NoError) {
                QJsonDocument doc = QJsonDocument::fromJson(reply->readAll());
                emit meetingCreated(doc.object());
            } else {
                emit meetingCreationFailed(reply->errorString());
            }
            reply->deleteLater();
        });
    }
    
signals:
    void loginSuccess();
    void loginFailed(const QString& error);
    void meetingCreated(const QJsonObject& meeting);
    void meetingCreationFailed(const QString& error);
    
private:
    QNetworkAccessManager* m_manager;
    QString m_baseUrl;
    QString m_token;
};
```

### WebSocket信令示例

```cpp
// WebSocket客户端类
class SignalingClient : public QObject {
    Q_OBJECT
    
public:
    SignalingClient(const QString& url, QObject* parent = nullptr)
        : QObject(parent) {
        m_socket = new QWebSocket();
        connect(m_socket, &QWebSocket::connected, this, &SignalingClient::onConnected);
        connect(m_socket, &QWebSocket::textMessageReceived, this, &SignalingClient::onMessageReceived);
        m_socket->open(QUrl(url));
    }
    
    void sendJoin(int meetingId, int userId, const QString& peerId) {
        QJsonObject json;
        json["type"] = "join";
        json["meeting_id"] = meetingId;
        json["user_id"] = userId;
        json["peer_id"] = peerId;
        m_socket->sendTextMessage(QJsonDocument(json).toJson());
    }
    
    void sendOffer(const QString& fromPeer, const QString& toPeer, const QString& sdp) {
        QJsonObject json;
        json["type"] = "offer";
        json["from_peer_id"] = fromPeer;
        json["to_peer_id"] = toPeer;
        json["sdp"] = sdp;
        m_socket->sendTextMessage(QJsonDocument(json).toJson());
    }
    
private slots:
    void onConnected() {
        qDebug() << "WebSocket connected";
        emit connected();
    }
    
    void onMessageReceived(const QString& message) {
        QJsonDocument doc = QJsonDocument::fromJson(message.toUtf8());
        QJsonObject obj = doc.object();
        QString type = obj["type"].toString();
        
        if (type == "offer") {
            emit offerReceived(obj["from_peer_id"].toString(), obj["sdp"].toString());
        } else if (type == "answer") {
            emit answerReceived(obj["from_peer_id"].toString(), obj["sdp"].toString());
        }
        // ... 处理其他消息类型
    }
    
signals:
    void connected();
    void offerReceived(const QString& fromPeer, const QString& sdp);
    void answerReceived(const QString& fromPeer, const QString& sdp);
    
private:
    QWebSocket* m_socket;
};
```

---

## 完整调用流程示例

```cpp
// 1. 登录
apiClient->login("testuser", "password123");

// 2. 登录成功后创建会议
connect(apiClient, &APIClient::loginSuccess, [=]() {
    apiClient->createMeeting("团队会议", QDateTime::currentDateTime().addSecs(3600));
});

// 3. 会议创建成功后加入会议
connect(apiClient, &APIClient::meetingCreated, [=](const QJsonObject& meeting) {
    int meetingId = meeting["meeting_id"].toInt();
    apiClient->joinMeeting(meetingId);
});

// 4. 加入成功后建立WebSocket连接
connect(apiClient, &APIClient::joinedMeeting, [=](const QJsonObject& joinInfo) {
    QString wsUrl = joinInfo["signaling_url"].toString();
    signalingClient = new SignalingClient(wsUrl);
});

// 5. WebSocket连接成功后发送join消息
connect(signalingClient, &SignalingClient::connected, [=]() {
    signalingClient->sendJoin(meetingId, userId, peerId);
});
```

---

## 总结

本指南涵盖了客户端调用后端API的所有主要场景：

1. ✅ 用户认证和授权
2. ✅ 会议管理（创建、加入、离开）
3. ✅ WebSocket实时信令
4. ✅ 媒体文件管理
5. ✅ AI服务调用
6. ✅ 错误处理
7. ✅ 完整的代码示例

更多详细信息请参考：
- [API完整文档](API_REFERENCE.md)
- [WSL部署指南](WSL_DEPLOYMENT_GUIDE.md)

