# Qt6客户端与后端通信架构设计

## 📋 目录

1. [架构概览](#架构概览)
2. [通信层设计](#通信层设计)
3. [API映射](#api映射)
4. [数据模型](#数据模型)
5. [错误处理](#错误处理)
6. [状态管理](#状态管理)
7. [实现计划](#实现计划)

---

## 架构概览

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      QML UI Layer                        │
│  (LoginPage, MeetingRoom, MainWindow, AIPanel)          │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                  Service Layer                           │
│  ┌──────────────┬──────────────┬──────────────────┐    │
│  │ AuthService  │MeetingService│  AIService       │    │
│  │              │              │  MediaService    │    │
│  └──────────────┴──────────────┴──────────────────┘    │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                 Network Layer                            │
│  ┌──────────────┬──────────────┬──────────────────┐    │
│  │  ApiClient   │ HttpClient   │ WebSocketClient  │    │
│  └──────────────┴──────────────┴──────────────────┘    │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│              Backend Gateway (Port 8000)                 │
│  ┌──────────────────────────────────────────────────┐  │
│  │  /api/v1/auth    /api/v1/meetings               │  │
│  │  /api/v1/users   /api/v1/media                  │  │
│  │  /api/v1/speech  /ws/signaling                  │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### 通信协议

1. **HTTP/HTTPS**: RESTful API调用
   - 认证、用户管理、会议管理、媒体上传等
   - 使用JWT Bearer Token认证

2. **WebSocket**: 实时信令通信
   - WebRTC信令交换（Offer/Answer/ICE）
   - 实时聊天消息
   - 参与者状态更新
   - 媒体控制信令

---

## 通信层设计

### 1. HttpClient (底层HTTP客户端)

**职责**: 封装QNetworkAccessManager，提供基础HTTP请求功能

**功能**:
- GET/POST/PUT/DELETE请求
- 文件上传（multipart/form-data）
- 请求超时控制
- 自动添加Authorization头
- 请求重试机制

**关键方法**:
```cpp
void get(const QString &url, callback, errorCallback);
void post(const QString &url, const QJsonObject &data, callback, errorCallback);
void put(const QString &url, const QJsonObject &data, callback, errorCallback);
void del(const QString &url, callback, errorCallback);
void upload(const QString &url, const QString &filePath, formData, callback, progressCallback);
void setAuthToken(const QString &token);
void setTimeout(int milliseconds);
```

### 2. ApiClient (API封装层)

**职责**: 封装所有后端API端点，提供类型安全的接口

**API分组**:

#### 2.1 认证API
```cpp
// POST /api/v1/auth/register
void registerUser(username, email, password, nickname, callback);

// POST /api/v1/auth/login
void login(username, password, callback);

// POST /api/v1/auth/refresh
void refreshToken(refreshToken, callback);

// POST /api/v1/auth/forgot-password
void forgotPassword(email, callback);

// POST /api/v1/auth/reset-password
void resetPassword(token, newPassword, callback);
```

#### 2.2 用户API
```cpp
// GET /api/v1/users/profile
void getUserProfile(callback);

// PUT /api/v1/users/profile
void updateUserProfile(nickname, email, avatarUrl, callback);

// POST /api/v1/users/change-password
void changePassword(oldPassword, newPassword, callback);

// POST /api/v1/users/upload-avatar
void uploadAvatar(filePath, callback, progressCallback);

// DELETE /api/v1/users/account
void deleteAccount(callback);
```

#### 2.3 会议API
```cpp
// POST /api/v1/meetings
void createMeeting(title, description, startTime, endTime, maxParticipants, 
                   meetingType, password, settings, callback);

// GET /api/v1/meetings?page=1&page_size=10&status=scheduled
void getMeetingList(page, pageSize, status, keyword, callback);

// GET /api/v1/meetings/:id
void getMeetingInfo(meetingId, callback);

// PUT /api/v1/meetings/:id
void updateMeeting(meetingId, updateData, callback);

// DELETE /api/v1/meetings/:id
void deleteMeeting(meetingId, callback);

// POST /api/v1/meetings/:id/start
void startMeeting(meetingId, callback);

// POST /api/v1/meetings/:id/end
void endMeeting(meetingId, callback);

// POST /api/v1/meetings/:id/join
void joinMeeting(meetingId, password, callback);

// POST /api/v1/meetings/:id/leave
void leaveMeeting(meetingId, callback);

// GET /api/v1/meetings/:id/participants
void getParticipants(meetingId, callback);

// POST /api/v1/meetings/:id/participants
void addParticipant(meetingId, userId, role, callback);

// DELETE /api/v1/meetings/:id/participants/:user_id
void removeParticipant(meetingId, userId, callback);

// PUT /api/v1/meetings/:id/participants/:user_id/role
void updateParticipantRole(meetingId, userId, role, callback);

// POST /api/v1/meetings/:id/recording/start
void startRecording(meetingId, callback);

// POST /api/v1/meetings/:id/recording/stop
void stopRecording(meetingId, callback);

// GET /api/v1/meetings/:id/recordings
void getRecordings(meetingId, callback);

// GET /api/v1/meetings/:id/messages
void getChatMessages(meetingId, page, pageSize, callback);

// POST /api/v1/meetings/:id/messages
void sendChatMessage(meetingId, content, callback);
```

#### 2.4 我的会议API
```cpp
// GET /api/v1/my/meetings
void getMyMeetings(callback);

// GET /api/v1/my/meetings/upcoming
void getUpcomingMeetings(callback);

// GET /api/v1/my/meetings/history
void getMeetingHistory(callback);
```

#### 2.5 媒体API
```cpp
// POST /api/v1/media/upload
void uploadMedia(filePath, meetingId, fileType, callback, progressCallback);

// GET /api/v1/media/download/:id
void downloadMedia(mediaId, savePath, callback, progressCallback);

// GET /api/v1/media
void getMediaList(meetingId, callback);

// GET /api/v1/media/info/:id
void getMediaInfo(mediaId, callback);

// DELETE /api/v1/media/:id
void deleteMedia(mediaId, callback);

// POST /api/v1/media/process
void processMedia(mediaId, processType, params, callback);
```

#### 2.6 AI服务API
```cpp
// POST /api/v1/speech/recognition
void speechRecognition(audioData, audioFormat, sampleRate, language, callback);

// POST /api/v1/speech/emotion
void emotionDetection(audioData, audioFormat, sampleRate, callback);

// POST /api/v1/speech/synthesis-detection
void synthesisDetection(audioData, callback);

// POST /api/v1/audio/denoising
void audioDenoising(audioData, callback);

// POST /api/v1/video/enhancement
void videoEnhancement(videoData, enhancementType, callback);
```

#### 2.7 信令服务API
```cpp
// GET /api/v1/sessions/:session_id
void getSessionInfo(sessionId, callback);

// GET /api/v1/sessions/room/:meeting_id
void getRoomSessions(meetingId, callback);

// GET /api/v1/messages/history/:meeting_id
void getMessageHistory(meetingId, callback);

// GET /api/v1/stats/overview
void getStatsOverview(callback);

// GET /api/v1/stats/rooms
void getRoomStats(callback);
```

### 3. WebSocketClient (WebSocket信令客户端)

**职责**: 管理WebSocket连接，处理实时信令消息

**连接URL格式**:
```
ws://gateway:8000/ws/signaling?meeting_id={id}&user_id={id}&peer_id={peer_id}
```

**消息类型**:
```cpp
enum class SignalingMessageType {
    Join = 1,           // 加入房间
    Leave = 2,          // 离开房间
    Offer = 3,          // WebRTC Offer
    Answer = 4,         // WebRTC Answer
    IceCandidate = 5,   // ICE候选
    Chat = 6,           // 聊天消息
    MediaState = 7,     // 媒体状态（静音/取消静音）
    ScreenShare = 8,    // 屏幕共享
    UserJoined = 9,     // 用户加入通知
    UserLeft = 10,      // 用户离开通知
    RoomInfo = 11,      // 房间信息
    Error = 12,         // 错误消息
    Ping = 13,          // 心跳ping
    Pong = 14           // 心跳pong
};
```

**发送消息格式**:
```json
{
  "type": "offer",
  "from_peer_id": "peer_abc123",
  "to_peer_id": "peer_def456",
  "meeting_id": 100,
  "user_id": 1,
  "payload": {
    "sdp": "v=0\r\no=- ..."
  },
  "timestamp": "2025-10-02T10:00:00Z"
}
```

**接收消息格式**:
```json
{
  "type": "user-joined",
  "user_id": 2,
  "peer_id": "peer_xyz789",
  "username": "newuser",
  "timestamp": "2025-10-02T10:01:00Z"
}
```

**关键功能**:
- 自动重连机制（断线后重连）
- 心跳保活（每30秒发送ping）
- 消息队列（连接断开时缓存消息）
- 消息确认机制

---

## API映射

### 响应格式标准化

**成功响应**:
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    // 业务数据
  },
  "timestamp": "2025-10-02T10:00:00Z"
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "Invalid request",
  "error": "Detailed error message",
  "timestamp": "2025-10-02T10:00:00Z",
  "request_id": "abc123"
}
```

### ApiResponse结构

```cpp
struct ApiResponse {
    int code;                    // HTTP状态码
    QString message;             // 消息描述
    QJsonObject data;            // 响应数据
    QString error;               // 错误详情
    QString timestamp;           // 时间戳
    QString requestId;           // 请求ID（用于追踪）
    
    bool isSuccess() const { 
        return code >= 200 && code < 300; 
    }
    
    bool isClientError() const {
        return code >= 400 && code < 500;
    }
    
    bool isServerError() const {
        return code >= 500;
    }
};
```

---

## 数据模型

### 1. User Model
```cpp
class User : public QObject {
    Q_OBJECT
    Q_PROPERTY(int userId READ userId NOTIFY userIdChanged)
    Q_PROPERTY(QString username READ username NOTIFY usernameChanged)
    Q_PROPERTY(QString email READ email NOTIFY emailChanged)
    Q_PROPERTY(QString nickname READ nickname NOTIFY nicknameChanged)
    Q_PROPERTY(QString avatarUrl READ avatarUrl NOTIFY avatarUrlChanged)
    Q_PROPERTY(QString status READ status NOTIFY statusChanged)
    
public:
    static User* fromJson(const QJsonObject &json);
    QJsonObject toJson() const;
};
```

### 2. Meeting Model
```cpp
class Meeting : public QObject {
    Q_OBJECT
    Q_PROPERTY(int meetingId READ meetingId)
    Q_PROPERTY(QString title READ title)
    Q_PROPERTY(QString description READ description)
    Q_PROPERTY(QDateTime startTime READ startTime)
    Q_PROPERTY(QDateTime endTime READ endTime)
    Q_PROPERTY(int maxParticipants READ maxParticipants)
    Q_PROPERTY(QString meetingType READ meetingType)
    Q_PROPERTY(QString status READ status)
    Q_PROPERTY(int creatorId READ creatorId)
    Q_PROPERTY(QJsonObject settings READ settings)
    
public:
    enum Status {
        Scheduled,
        Ongoing,
        Ended,
        Cancelled
    };
    
    static Meeting* fromJson(const QJsonObject &json);
    QJsonObject toJson() const;
};
```

### 3. Participant Model
```cpp
class Participant : public QObject {
    Q_OBJECT
    Q_PROPERTY(int participantId READ participantId)
    Q_PROPERTY(int userId READ userId)
    Q_PROPERTY(QString username READ username)
    Q_PROPERTY(QString nickname READ nickname)
    Q_PROPERTY(QString role READ role)
    Q_PROPERTY(QString status READ status)
    Q_PROPERTY(bool audioEnabled READ audioEnabled)
    Q_PROPERTY(bool videoEnabled READ videoEnabled)
    Q_PROPERTY(QDateTime joinedAt READ joinedAt)
    
public:
    enum Role {
        Host,
        Moderator,
        Participant
    };
    
    static Participant* fromJson(const QJsonObject &json);
};
```

### 4. Message Model
```cpp
class Message : public QObject {
    Q_OBJECT
    Q_PROPERTY(int messageId READ messageId)
    Q_PROPERTY(int userId READ userId)
    Q_PROPERTY(QString username READ username)
    Q_PROPERTY(QString content READ content)
    Q_PROPERTY(QDateTime timestamp READ timestamp)
    Q_PROPERTY(QString messageType READ messageType)
    
public:
    enum Type {
        Text,
        System,
        File
    };
    
    static Message* fromJson(const QJsonObject &json);
};
```

---

## 错误处理

### HTTP状态码处理

| 状态码 | 说明 | 客户端处理 |
|--------|------|-----------|
| 200 | 成功 | 正常处理响应数据 |
| 201 | 创建成功 | 正常处理响应数据 |
| 400 | 请求参数错误 | 显示错误提示，检查输入 |
| 401 | 未认证/Token无效 | 跳转登录页，清除Token |
| 403 | 无权限 | 显示权限不足提示 |
| 404 | 资源不存在 | 显示资源不存在提示 |
| 429 | 请求过于频繁 | 显示限流提示，延迟重试 |
| 500 | 服务器错误 | 显示服务器错误，建议稍后重试 |
| 503 | 服务不可用 | 显示服务维护提示 |

### 业务错误码

| 错误码 | 说明 | 处理方式 |
|--------|------|---------|
| 1001 | 用户名已存在 | 提示用户更换用户名 |
| 1002 | 邮箱已存在 | 提示用户更换邮箱 |
| 1003 | 用户不存在 | 提示用户不存在 |
| 1004 | 密码错误 | 提示密码错误 |
| 2001 | 会议不存在 | 提示会议不存在或已删除 |
| 2002 | 会议已结束 | 提示会议已结束 |
| 2003 | 会议人数已满 | 提示会议人数已满 |
| 2004 | 会议密码错误 | 提示密码错误 |
| 3001 | 文件上传失败 | 提示上传失败，重试 |
| 3002 | 文件不存在 | 提示文件不存在 |
| 4001 | AI服务不可用 | 提示AI服务暂时不可用 |
| 4002 | AI处理失败 | 提示处理失败 |

### 错误处理策略

```cpp
class ErrorHandler {
public:
    static void handleApiError(const ApiResponse &response) {
        if (response.code == 401) {
            // Token过期，跳转登录
            emit authenticationRequired();
        } else if (response.code == 429) {
            // 限流，延迟重试
            QTimer::singleShot(5000, []() {
                // 重试逻辑
            });
        } else if (response.code >= 500) {
            // 服务器错误
            showError("服务器错误，请稍后重试");
        } else {
            // 其他错误
            showError(response.message);
        }
    }
};
```

---

## 状态管理

### Token管理

```cpp
class TokenManager {
public:
    void setToken(const QString &accessToken, const QString &refreshToken, int expiresIn);
    QString getAccessToken() const;
    QString getRefreshToken() const;
    bool isTokenExpired() const;
    void refreshTokenIfNeeded(std::function<void(bool)> callback);
    void clearTokens();
    
private:
    QString m_accessToken;
    QString m_refreshToken;
    QDateTime m_expiresAt;
};
```

### 连接状态管理

```cpp
enum class ConnectionState {
    Disconnected,
    Connecting,
    Connected,
    Reconnecting,
    Error
};

class ConnectionManager : public QObject {
    Q_OBJECT
    Q_PROPERTY(ConnectionState state READ state NOTIFY stateChanged)
    
signals:
    void stateChanged(ConnectionState state);
    void connected();
    void disconnected();
    void error(const QString &error);
};
```

---

## 实现计划

### Phase 1: 完善网络层 (优先级: 高)
- [x] HttpClient基础实现
- [ ] 完善ApiClient所有API端点
- [ ] 添加请求重试机制
- [ ] 添加请求队列管理
- [ ] 实现Token自动刷新

### Phase 2: 完善WebSocket层 (优先级: 高)
- [x] WebSocketClient基础实现
- [ ] 完善消息类型处理
- [ ] 实现自动重连机制
- [ ] 实现心跳保活
- [ ] 实现消息队列

### Phase 3: 完善Service层 (优先级: 中)
- [x] AuthService基础实现
- [x] MeetingService基础实现
- [ ] 完善MediaService
- [ ] 完善AIService
- [ ] 添加状态管理

### Phase 4: 数据模型完善 (优先级: 中)
- [x] User Model
- [x] Meeting Model
- [x] Participant Model
- [x] Message Model
- [ ] Recording Model
- [ ] MediaFile Model

### Phase 5: 错误处理与日志 (优先级: 中)
- [ ] 统一错误处理机制
- [ ] 日志记录系统
- [ ] 错误上报机制

### Phase 6: 性能优化 (优先级: 低)
- [ ] 请求缓存
- [ ] 数据预加载
- [ ] 连接池管理

---

## 配置管理

### config.json更新

```json
{
  "api": {
    "base_url": "http://localhost:8000",
    "ws_url": "ws://localhost:8000/ws/signaling",
    "timeout": 30000,
    "retry_count": 3,
    "retry_delay": 1000
  },
  "auth": {
    "token_refresh_threshold": 300,
    "auto_refresh": true
  },
  "websocket": {
    "heartbeat_interval": 30000,
    "reconnect_interval": 5000,
    "max_reconnect_attempts": 5
  }
}
```

---

## 总结

本设计文档定义了Qt6客户端与后端服务器的完整通信架构，包括：

1. ✅ 清晰的分层架构（UI → Service → Network → Backend）
2. ✅ 完整的API映射（78个HTTP端点 + WebSocket）
3. ✅ 标准化的数据模型
4. ✅ 完善的错误处理机制
5. ✅ 可靠的状态管理
6. ✅ 详细的实现计划

**下一步行动**:
1. 完善ApiClient，实现所有API端点
2. 增强WebSocketClient的可靠性
3. 完善Service层的业务逻辑
4. 实现完整的错误处理
5. 添加单元测试

