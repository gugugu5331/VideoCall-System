# 智能视频会议平台 API 文档 - 第2部分

**版本**: v1.0.0  
**基础URL**: `http://gateway:8000`

---

## 3. 会议服务（续）

### 3.2 获取会议列表

**端点**: `GET /api/v1/meetings`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**查询参数**:
- `page`: 页码 (必需, ≥1)
- `page_size`: 每页数量 (必需, 1-100)
- `status`: 会议状态 (可选: scheduled, ongoing, ended, cancelled)
- `keyword`: 搜索关键词 (可选)

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "meetings": [
      {
        "meeting_id": 1,
        "title": "Team Meeting",
        "start_time": "2025-10-03T10:00:00Z",
        "end_time": "2025-10-03T11:00:00Z",
        "status": "scheduled",
        "participant_count": 5
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

---

### 3.3 获取会议详情

**端点**: `GET /api/v1/meetings/:id`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "meeting_id": 1,
    "title": "Team Meeting",
    "description": "Weekly team sync",
    "start_time": "2025-10-03T10:00:00Z",
    "end_time": "2025-10-03T11:00:00Z",
    "max_participants": 10,
    "meeting_type": "video",
    "status": "scheduled",
    "creator_id": 1,
    "settings": {
      "enable_recording": false,
      "enable_chat": true,
      "enable_screen_share": true
    },
    "created_at": "2025-10-02T14:00:00Z"
  }
}
```

---

### 3.4 更新会议

**端点**: `PUT /api/v1/meetings/:id`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**请求体**:
```json
{
  "title": "string (可选)",
  "description": "string (可选)",
  "start_time": "string (可选)",
  "end_time": "string (可选)",
  "max_participants": "number (可选)",
  "settings": {
    "enable_recording": "boolean (可选)",
    "enable_chat": "boolean (可选)"
  }
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Meeting updated successfully"
}
```

---

### 3.5 删除会议

**端点**: `DELETE /api/v1/meetings/:id`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Meeting deleted successfully"
}
```

---

### 3.6 开始会议

**端点**: `POST /api/v1/meetings/:id/start`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Meeting started successfully",
  "data": {
    "meeting_id": 1,
    "status": "ongoing",
    "room_url": "wss://gateway:8000/ws/signaling?meeting_id=1&user_id=1"
  }
}
```

---

### 3.7 结束会议

**端点**: `POST /api/v1/meetings/:id/end`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Meeting ended successfully"
}
```

---

### 3.8 加入会议

**端点**: `POST /api/v1/meetings/:id/join`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**请求体**:
```json
{
  "password": "string (可选, 如果会议有密码)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Joined meeting successfully",
  "data": {
    "meeting_id": 1,
    "participant_id": 123,
    "room_url": "wss://gateway:8000/ws/signaling?meeting_id=1&user_id=1",
    "ice_servers": [
      {
        "urls": "stun:stun.l.google.com:19302"
      }
    ]
  }
}
```

---

### 3.9 离开会议

**端点**: `POST /api/v1/meetings/:id/leave`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Left meeting successfully"
}
```

---

### 3.10 获取参与者列表

**端点**: `GET /api/v1/meetings/:id/participants`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "participants": [
      {
        "participant_id": 123,
        "user_id": 1,
        "username": "testuser",
        "nickname": "Test User",
        "role": "host",
        "status": "online",
        "joined_at": "2025-10-03T10:05:00Z"
      }
    ],
    "total": 5
  }
}
```

---

### 3.11 添加参与者

**端点**: `POST /api/v1/meetings/:id/participants`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**请求体**:
```json
{
  "user_id": "number (必需)",
  "role": "string (可选: host, moderator, participant, 默认participant)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Participant added successfully"
}
```

---

### 3.12 移除参与者

**端点**: `DELETE /api/v1/meetings/:id/participants/:user_id`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Participant removed successfully"
}
```

---

### 3.13 更新参与者角色

**端点**: `PUT /api/v1/meetings/:id/participants/:user_id/role`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**请求体**:
```json
{
  "role": "string (必需: host, moderator, participant)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Participant role updated successfully"
}
```

---

### 3.14 开始录制

**端点**: `POST /api/v1/meetings/:id/recording/start`  
**认证**: 需要 JWT Token  
**限流**: 20次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Recording started successfully",
  "data": {
    "recording_id": "rec_123456"
  }
}
```

---

### 3.15 停止录制

**端点**: `POST /api/v1/meetings/:id/recording/stop`  
**认证**: 需要 JWT Token  
**限流**: 20次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Recording stopped successfully"
}
```

---

### 3.16 获取录制列表

**端点**: `GET /api/v1/meetings/:id/recordings`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "recordings": [
      {
        "recording_id": "rec_123456",
        "meeting_id": 1,
        "file_url": "https://example.com/recordings/rec_123456.mp4",
        "duration": 3600,
        "size": 1024000000,
        "created_at": "2025-10-03T10:00:00Z"
      }
    ]
  }
}
```

---

### 3.17 获取聊天消息

**端点**: `GET /api/v1/meetings/:id/messages`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 50)

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "messages": [
      {
        "message_id": 1,
        "user_id": 1,
        "username": "testuser",
        "content": "Hello everyone!",
        "timestamp": "2025-10-03T10:15:00Z"
      }
    ],
    "total": 100
  }
}
```

---

### 3.18 发送聊天消息

**端点**: `POST /api/v1/meetings/:id/messages`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**请求体**:
```json
{
  "content": "string (必需, 1-1000字符)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Message sent successfully",
  "data": {
    "message_id": 1,
    "timestamp": "2025-10-03T10:15:00Z"
  }
}
```

---


