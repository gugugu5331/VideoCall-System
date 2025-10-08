# 智能视频会议平台 - API参考文档

**版本**: v1.0.0 (稳定版)  
**基础URL**: `http://gateway:8000`  
**协议**: HTTP/HTTPS + WebSocket  
**认证**: JWT Bearer Token  
**更新日期**: 2025-10-02

---

## 📌 核心说明

### 1. 访问方式
- **所有API必须通过网关访问**: `http://gateway:8000/api/v1/*`
- **不允许直接访问微服务**: 客户端只能与网关通信

### 2. 认证方式
```http
Authorization: Bearer <your_jwt_token>
```

### 3. 响应格式
```json
{
  "code": 200,
  "message": "Success",
  "data": {},
  "timestamp": "2025-10-02T10:00:00Z",
  "request_id": "abc123"
}
```

### 4. API稳定性承诺
- ✅ 接口路径不会变更
- ✅ 请求/响应格式向后兼容
- ✅ 内部实现变化不影响API
- ✅ 新增字段不影响现有功能

---

## 📋 完整API列表

### 🔐 认证服务 (`/api/v1/auth`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/register` | 用户注册 | ❌ | 5/分钟 |
| POST | `/login` | 用户登录 | ❌ | 5/分钟 |
| POST | `/refresh` | 刷新Token | ✅ | 10/分钟 |
| POST | `/forgot-password` | 忘记密码 | ❌ | 3/小时 |
| POST | `/reset-password` | 重置密码 | ❌ | 5/小时 |

### 👤 用户服务 (`/api/v1/users`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| GET | `/profile` | 获取用户资料 | ✅ | 100/分钟 |
| PUT | `/profile` | 更新用户资料 | ✅ | 50/分钟 |
| POST | `/change-password` | 修改密码 | ✅ | 10/小时 |
| POST | `/upload-avatar` | 上传头像 | ✅ | 10/小时 |
| DELETE | `/account` | 删除账户 | ✅ | 1/天 |

### 👥 用户管理 (`/api/v1/admin/users`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| GET | `/` | 获取用户列表 | 🔑 | 50/分钟 |
| GET | `/:id` | 获取指定用户 | 🔑 | 100/分钟 |
| PUT | `/:id` | 更新用户 | 🔑 | 50/分钟 |
| DELETE | `/:id` | 删除用户 | 🔑 | 20/分钟 |
| POST | `/:id/ban` | 封禁用户 | 🔑 | 20/分钟 |
| POST | `/:id/unban` | 解封用户 | 🔑 | 20/分钟 |

🔑 = 需要管理员权限

### 🎥 会议服务 (`/api/v1/meetings`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/` | 创建会议 | ✅ | 50/分钟 |
| GET | `/` | 获取会议列表 | ✅ | 100/分钟 |
| GET | `/:id` | 获取会议详情 | ✅ | 100/分钟 |
| PUT | `/:id` | 更新会议 | ✅ | 50/分钟 |
| DELETE | `/:id` | 删除会议 | ✅ | 50/分钟 |
| POST | `/:id/start` | 开始会议 | ✅ | 50/分钟 |
| POST | `/:id/end` | 结束会议 | ✅ | 50/分钟 |
| POST | `/:id/join` | 加入会议 | ✅ | 100/分钟 |
| POST | `/:id/leave` | 离开会议 | ✅ | 100/分钟 |
| GET | `/:id/participants` | 获取参与者 | ✅ | 100/分钟 |
| POST | `/:id/participants` | 添加参与者 | ✅ | 50/分钟 |
| DELETE | `/:id/participants/:user_id` | 移除参与者 | ✅ | 50/分钟 |
| PUT | `/:id/participants/:user_id/role` | 更新角色 | ✅ | 50/分钟 |
| POST | `/:id/recording/start` | 开始录制 | ✅ | 20/分钟 |
| POST | `/:id/recording/stop` | 停止录制 | ✅ | 20/分钟 |
| GET | `/:id/recordings` | 获取录制列表 | ✅ | 50/分钟 |
| GET | `/:id/messages` | 获取聊天消息 | ✅ | 100/分钟 |
| POST | `/:id/messages` | 发送聊天消息 | ✅ | 100/分钟 |

### 📅 我的会议 (`/api/v1/my`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| GET | `/meetings` | 我的会议 | ✅ | 100/分钟 |
| GET | `/meetings/upcoming` | 即将开始 | ✅ | 100/分钟 |
| GET | `/meetings/history` | 会议历史 | ✅ | 100/分钟 |

### 📡 信令服务

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| WS | `/ws/signaling` | WebSocket连接 | ✅ | 无限制 |
| GET | `/api/v1/sessions/:session_id` | 获取会话 | ✅ | 100/分钟 |
| GET | `/api/v1/sessions/room/:meeting_id` | 房间会话 | ✅ | 100/分钟 |
| GET | `/api/v1/messages/history/:meeting_id` | 消息历史 | ✅ | 100/分钟 |
| GET | `/api/v1/stats/overview` | 统计概览 | ✅ | 50/分钟 |
| GET | `/api/v1/stats/rooms` | 房间统计 | ✅ | 50/分钟 |

### 📁 媒体服务 (`/api/v1/media`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/upload` | 上传文件 | ✅ | 5/分钟 |
| GET | `/download/:id` | 下载文件 | ✅ | 50/分钟 |
| GET | `/` | 获取列表 | ✅ | 50/分钟 |
| POST | `/process` | 处理文件 | ✅ | 20/分钟 |
| GET | `/info/:id` | 获取信息 | ✅ | 100/分钟 |
| DELETE | `/:id` | 删除文件 | ✅ | 50/分钟 |

### 🎬 WebRTC服务 (`/api/v1/webrtc`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| GET | `/room/:roomId/peers` | 对等端列表 | ✅ | 100/分钟 |
| GET | `/room/:roomId/stats` | 房间统计 | ✅ | 50/分钟 |
| POST | `/peer/:peerId/media` | 更新媒体 | ✅ | 50/分钟 |

### 🎞️ FFmpeg服务 (`/api/v1/ffmpeg`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/transcode` | 转码 | ✅ | 10/分钟 |
| POST | `/extract-audio` | 提取音频 | ✅ | 10/分钟 |
| POST | `/extract-video` | 提取视频 | ✅ | 10/分钟 |
| POST | `/merge` | 合并媒体 | ✅ | 10/分钟 |
| POST | `/thumbnail` | 生成缩略图 | ✅ | 20/分钟 |
| GET | `/job/:id/status` | 任务状态 | ✅ | 100/分钟 |

### 📹 录制服务 (`/api/v1/recording`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/start` | 开始录制 | ✅ | 20/分钟 |
| POST | `/stop` | 停止录制 | ✅ | 20/分钟 |
| GET | `/:id` | 录制信息 | ✅ | 100/分钟 |
| GET | `/list` | 录制列表 | ✅ | 50/分钟 |
| DELETE | `/:id` | 删除录制 | ✅ | 20/分钟 |

### 📺 流媒体服务 (`/api/v1/streaming`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/start` | 开始推流 | ✅ | 10/分钟 |
| POST | `/stop` | 停止推流 | ✅ | 10/分钟 |
| GET | `/:id/status` | 推流状态 | ✅ | 50/分钟 |

### 🤖 AI服务 - 语音 (`/api/v1/speech`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/recognition` | 语音识别 | ✅ | 10/分钟 |
| POST | `/emotion` | 情绪检测 | ✅ | 10/分钟 |
| POST | `/synthesis-detection` | 合成检测 | ✅ | 10/分钟 |

### 🎵 AI服务 - 音视频增强

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| POST | `/api/v1/audio/denoising` | 音频降噪 | ✅ | 10/分钟 |
| POST | `/api/v1/video/enhancement` | 视频增强 | ✅ | 10/分钟 |

### 🧠 AI服务 - 模型管理 (`/api/v1/models`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| GET | `/` | 模型列表 | ✅ | 50/分钟 |
| POST | `/:model_id/load` | 加载模型 | ✅ | 10/分钟 |
| DELETE | `/:model_id/unload` | 卸载模型 | ✅ | 10/分钟 |
| GET | `/:model_id/status` | 模型状态 | ✅ | 50/分钟 |

### 🖥️ AI服务 - 节点管理 (`/api/v1/nodes`)

| 方法 | 端点 | 说明 | 认证 | 限流 |
|------|------|------|------|------|
| GET | `/` | 节点列表 | ✅ | 50/分钟 |
| GET | `/:node_id/status` | 节点状态 | ✅ | 50/分钟 |
| POST | `/:node_id/health-check` | 健康检查 | ✅ | 20/分钟 |

---

## 📊 数据模型

### User
```typescript
{
  user_id: number
  username: string
  email: string
  nickname: string
  avatar_url: string
  status: 'active' | 'banned' | 'deleted'
  created_at: string  // ISO8601
  updated_at: string  // ISO8601
}
```

### Meeting
```typescript
{
  meeting_id: number
  title: string
  description: string
  start_time: string  // ISO8601
  end_time: string    // ISO8601
  max_participants: number
  meeting_type: 'video' | 'audio'
  status: 'scheduled' | 'ongoing' | 'ended' | 'cancelled'
  creator_id: number
  settings: {
    enable_recording: boolean
    enable_chat: boolean
    enable_screen_share: boolean
    enable_waiting_room: boolean
    mute_on_join: boolean
  }
  created_at: string  // ISO8601
}
```

### Participant
```typescript
{
  participant_id: number
  meeting_id: number
  user_id: number
  username: string
  nickname: string
  role: 'host' | 'moderator' | 'participant'
  status: 'online' | 'offline' | 'waiting'
  joined_at: string  // ISO8601
}
```

### MediaFile
```typescript
{
  file_id: string
  user_id: number
  meeting_id: number
  file_name: string
  file_url: string
  file_type: 'video' | 'audio' | 'image' | 'document'
  mime_type: string
  size: number  // bytes
  duration?: number  // seconds
  created_at: string  // ISO8601
}
```

---

## ⚠️ 错误码

### HTTP状态码
- `200` - 成功
- `201` - 创建成功
- `400` - 请求错误
- `401` - 未认证
- `403` - 无权限
- `404` - 不存在
- `429` - 请求过多
- `500` - 服务器错误

### 业务错误码
- `1001` - 用户名已存在
- `1002` - 邮箱已存在
- `1003` - 用户不存在
- `1004` - 密码错误
- `1005` - Token过期
- `2001` - 会议不存在
- `2002` - 会议已结束
- `2003` - 会议已满
- `2004` - 密码错误
- `3001` - 上传失败
- `3002` - 文件不存在
- `4001` - AI服务不可用

---

## 🔌 WebSocket协议

### 连接
```
wss://gateway:8000/ws/signaling?meeting_id=1&user_id=1
```

### 消息类型
- `join` - 加入房间
- `leave` - 离开房间
- `offer` - WebRTC Offer
- `answer` - WebRTC Answer
- `ice-candidate` - ICE候选
- `chat` - 聊天消息
- `media-state` - 媒体状态
- `user-joined` - 用户加入通知
- `user-left` - 用户离开通知
- `error` - 错误消息

### 消息格式
```json
{
  "type": "string",
  "from": "number",
  "to": "number (optional)",
  "meeting_id": "number",
  "data": {}
}
```

---

## 📝 使用示例

### 1. 完整会议流程
```bash
# 1. 注册
POST /api/v1/auth/register

# 2. 登录获取Token
POST /api/v1/auth/login

# 3. 创建会议
POST /api/v1/meetings

# 4. 加入会议
POST /api/v1/meetings/1/join

# 5. 建立WebSocket连接
WS /ws/signaling?meeting_id=1&user_id=1

# 6. 离开会议
POST /api/v1/meetings/1/leave
```

### 2. 文件上传
```bash
POST /api/v1/media/upload
Content-Type: multipart/form-data

file: <binary>
user_id: 1
meeting_id: 1
```

### 3. AI语音识别
```bash
POST /api/v1/speech/recognition
{
  "audio_url": "https://example.com/audio.wav",
  "language": "zh-CN"
}
```

---

## 🔒 安全建议

1. ✅ 使用HTTPS
2. ✅ 安全存储Token
3. ✅ 验证所有输入
4. ✅ 遵守限流规则
5. ✅ 处理所有错误

---

**文档版本**: v1.0.0 (稳定)  
**API稳定性**: 向后兼容保证  
**最后更新**: 2025-10-02


