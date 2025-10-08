# 智能视频会议平台 - 完整API文档

**版本**: v1.0.0  
**基础URL**: `http://gateway:8000`  
**协议**: HTTP/HTTPS + WebSocket  
**认证方式**: JWT Bearer Token  
**文档更新**: 2025-10-02

---

## 📌 重要说明

1. **所有API请求必须通过网关层** (`http://gateway:8000`)
2. **认证**: 除公开接口外，所有API需要在请求头中携带JWT Token
   ```
   Authorization: Bearer <your_jwt_token>
   ```
3. **限流**: 每个端点都有独立的限流规则，超出限制将返回429错误
4. **响应格式**: 所有响应均为JSON格式
5. **时间格式**: 使用ISO8601格式 (例: `2025-10-02T10:00:00Z`)

---

## 📋 API端点总览

### 1. 认证服务 (`/api/v1/auth`)
- `POST /register` - 用户注册
- `POST /login` - 用户登录
- `POST /refresh` - 刷新Token
- `POST /forgot-password` - 忘记密码
- `POST /reset-password` - 重置密码

### 2. 用户服务 (`/api/v1/users`)
- `GET /profile` - 获取用户资料
- `PUT /profile` - 更新用户资料
- `POST /change-password` - 修改密码
- `POST /upload-avatar` - 上传头像
- `DELETE /account` - 删除账户

### 3. 用户管理（管理员） (`/api/v1/admin/users`)
- `GET /` - 获取用户列表
- `GET /:id` - 获取指定用户
- `PUT /:id` - 更新用户
- `DELETE /:id` - 删除用户
- `POST /:id/ban` - 封禁用户
- `POST /:id/unban` - 解封用户

### 4. 会议服务 (`/api/v1/meetings`)
- `POST /` - 创建会议
- `GET /` - 获取会议列表
- `GET /:id` - 获取会议详情
- `PUT /:id` - 更新会议
- `DELETE /:id` - 删除会议
- `POST /:id/start` - 开始会议
- `POST /:id/end` - 结束会议
- `POST /:id/join` - 加入会议
- `POST /:id/leave` - 离开会议
- `GET /:id/participants` - 获取参与者列表
- `POST /:id/participants` - 添加参与者
- `DELETE /:id/participants/:user_id` - 移除参与者
- `PUT /:id/participants/:user_id/role` - 更新参与者角色
- `POST /:id/recording/start` - 开始录制
- `POST /:id/recording/stop` - 停止录制
- `GET /:id/recordings` - 获取录制列表
- `GET /:id/messages` - 获取聊天消息
- `POST /:id/messages` - 发送聊天消息

### 5. 我的会议 (`/api/v1/my`)
- `GET /meetings` - 获取我的会议
- `GET /meetings/upcoming` - 获取即将开始的会议
- `GET /meetings/history` - 获取会议历史

### 6. 信令服务
- `WS /ws/signaling` - WebSocket信令连接
- `GET /api/v1/sessions/:session_id` - 获取会话信息
- `GET /api/v1/sessions/room/:meeting_id` - 获取房间会话列表
- `GET /api/v1/messages/history/:meeting_id` - 获取消息历史
- `GET /api/v1/stats/overview` - 获取统计概览
- `GET /api/v1/stats/rooms` - 获取房间统计

### 7. 媒体服务 (`/api/v1/media`)
- `POST /upload` - 上传媒体文件
- `GET /download/:id` - 下载媒体文件
- `GET /` - 获取媒体列表
- `POST /process` - 处理媒体文件
- `GET /info/:id` - 获取媒体信息
- `DELETE /:id` - 删除媒体文件

### 8. WebRTC服务 (`/api/v1/webrtc`)
- `GET /room/:roomId/peers` - 获取房间对等端列表
- `GET /room/:roomId/stats` - 获取房间统计
- `POST /peer/:peerId/media` - 更新对等端媒体

### 9. FFmpeg服务 (`/api/v1/ffmpeg`)
- `POST /transcode` - 转码媒体
- `POST /extract-audio` - 提取音频
- `POST /extract-video` - 提取视频
- `POST /merge` - 合并媒体
- `POST /thumbnail` - 生成缩略图
- `GET /job/:id/status` - 获取任务状态

### 10. 录制服务 (`/api/v1/recording`)
- `POST /start` - 开始录制
- `POST /stop` - 停止录制
- `GET /:id` - 获取录制信息
- `GET /list` - 获取录制列表
- `DELETE /:id` - 删除录制

### 11. 流媒体服务 (`/api/v1/streaming`)
- `POST /start` - 开始推流
- `POST /stop` - 停止推流
- `GET /:id/status` - 获取推流状态

### 12. AI服务 - 语音 (`/api/v1/speech`)
- `POST /recognition` - 语音识别
- `POST /emotion` - 情绪检测
- `POST /synthesis-detection` - 合成检测

### 13. AI服务 - 音视频增强 (`/api/v1/audio`, `/api/v1/video`)
- `POST /audio/denoising` - 音频降噪
- `POST /video/enhancement` - 视频增强

### 14. AI服务 - 模型管理 (`/api/v1/models`)
- `GET /` - 获取模型列表
- `POST /:model_id/load` - 加载模型
- `DELETE /:model_id/unload` - 卸载模型
- `GET /:model_id/status` - 获取模型状态

### 15. AI服务 - 节点管理 (`/api/v1/nodes`)
- `GET /` - 获取节点列表
- `GET /:node_id/status` - 获取节点状态
- `POST /:node_id/health-check` - 节点健康检查

---

## 🔐 认证流程

### 1. 注册新用户
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "nickname": "Test User"
}
```

### 2. 登录获取Token
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

### 3. 使用Token访问受保护的API
```http
GET /api/v1/users/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 4. Token过期后刷新
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## 🎥 会议流程

### 1. 创建会议
```http
POST /api/v1/meetings
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Team Meeting",
  "description": "Weekly sync",
  "start_time": "2025-10-03T10:00:00Z",
  "end_time": "2025-10-03T11:00:00Z",
  "max_participants": 10,
  "meeting_type": "video",
  "settings": {
    "enable_recording": true,
    "enable_chat": true,
    "enable_screen_share": true
  }
}
```

### 2. 加入会议
```http
POST /api/v1/meetings/1/join
Authorization: Bearer <token>
Content-Type: application/json

{
  "password": "optional_password"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "meeting_id": 1,
    "participant_id": 123,
    "room_url": "wss://gateway:8000/ws/signaling?meeting_id=1&user_id=1",
    "ice_servers": [
      {"urls": "stun:stun.l.google.com:19302"}
    ]
  }
}
```

### 3. 建立WebSocket连接
```javascript
const ws = new WebSocket('wss://gateway:8000/ws/signaling?meeting_id=1&user_id=1');

ws.onopen = () => {
  // 发送加入房间消息
  ws.send(JSON.stringify({
    type: 'join',
    meeting_id: 1,
    user_id: 1
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  // 处理信令消息 (offer, answer, ice-candidate等)
};
```

### 4. 离开会议
```http
POST /api/v1/meetings/1/leave
Authorization: Bearer <token>
```

---

## 📁 媒体文件上传

```http
POST /api/v1/media/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>
user_id: 1
meeting_id: 1
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "file_id": "file_123456",
    "file_url": "https://example.com/media/file_123456.mp4",
    "size": 1024000,
    "mime_type": "video/mp4"
  }
}
```

---

## 🤖 AI服务使用

### 语音识别
```http
POST /api/v1/speech/recognition
Authorization: Bearer <token>
Content-Type: application/json

{
  "audio_url": "https://example.com/audio.wav",
  "language": "zh-CN"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "text": "这是识别的文本内容",
    "confidence": 0.95,
    "language": "zh-CN"
  }
}
```

### 情绪检测
```http
POST /api/v1/speech/emotion
Authorization: Bearer <token>
Content-Type: application/json

{
  "audio_url": "https://example.com/audio.wav"
}
```

**响应**:
```json
{
  "code": 200,
  "data": {
    "emotion": "neutral",
    "confidence": 0.88,
    "details": {
      "happy": 0.15,
      "sad": 0.05,
      "angry": 0.02,
      "neutral": 0.78
    }
  }
}
```

---


