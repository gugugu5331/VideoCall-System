# API设计文档

## 概述

本文档定义了视频会议系统的RESTful API和gRPC接口设计。

## 认证

所有API请求都需要在Header中包含JWT令牌：
```
Authorization: Bearer <jwt_token>
```

## 用户服务 API

### 用户注册
```http
POST /api/v1/users/register
Content-Type: application/json

{
  "username": "string",
  "email": "string",
  "password": "string",
  "full_name": "string"
}
```

### 用户登录
```http
POST /api/v1/users/login
Content-Type: application/json

{
  "email": "string",
  "password": "string"
}

Response:
{
  "token": "string",
  "user": {
    "id": "string",
    "username": "string",
    "email": "string",
    "full_name": "string"
  }
}
```

### 获取用户信息
```http
GET /api/v1/users/profile
Authorization: Bearer <token>

Response:
{
  "id": "string",
  "username": "string",
  "email": "string",
  "full_name": "string",
  "avatar_url": "string",
  "created_at": "timestamp"
}
```

## 会议服务 API

### 创建会议
```http
POST /api/v1/meetings
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "string",
  "description": "string",
  "start_time": "timestamp",
  "duration": "integer",
  "max_participants": "integer",
  "is_public": "boolean"
}

Response:
{
  "id": "string",
  "title": "string",
  "meeting_url": "string",
  "join_code": "string"
}
```

### 加入会议
```http
POST /api/v1/meetings/{meeting_id}/join
Authorization: Bearer <token>

Response:
{
  "signaling_url": "string",
  "ice_servers": [
    {
      "urls": ["string"],
      "username": "string",
      "credential": "string"
    }
  ]
}
```

### 获取会议列表
```http
GET /api/v1/meetings?page=1&limit=20
Authorization: Bearer <token>

Response:
{
  "meetings": [
    {
      "id": "string",
      "title": "string",
      "start_time": "timestamp",
      "status": "string",
      "participant_count": "integer"
    }
  ],
  "total": "integer",
  "page": "integer",
  "limit": "integer"
}
```

## 检测服务 API

### 提交检测任务
```http
POST /api/v1/detection/analyze
Content-Type: multipart/form-data
Authorization: Bearer <token>

{
  "file": "binary",
  "type": "video|audio|image",
  "meeting_id": "string"
}

Response:
{
  "task_id": "string",
  "status": "pending"
}
```

### 获取检测结果
```http
GET /api/v1/detection/results/{task_id}
Authorization: Bearer <token>

Response:
{
  "task_id": "string",
  "status": "completed|processing|failed",
  "result": {
    "is_fake": "boolean",
    "confidence": "float",
    "details": {
      "face_swap_probability": "float",
      "voice_synthesis_probability": "float",
      "manipulation_regions": [
        {
          "x": "integer",
          "y": "integer",
          "width": "integer",
          "height": "integer"
        }
      ]
    }
  },
  "created_at": "timestamp",
  "completed_at": "timestamp"
}
```

## 记录服务 API

### 获取通讯记录
```http
GET /api/v1/records/communications?meeting_id={id}&page=1&limit=50
Authorization: Bearer <token>

Response:
{
  "records": [
    {
      "id": "string",
      "meeting_id": "string",
      "user_id": "string",
      "message": "string",
      "type": "text|audio|video",
      "timestamp": "timestamp"
    }
  ],
  "total": "integer"
}
```

### 获取会议记录
```http
GET /api/v1/records/meetings/{meeting_id}
Authorization: Bearer <token>

Response:
{
  "meeting_id": "string",
  "title": "string",
  "start_time": "timestamp",
  "end_time": "timestamp",
  "participants": [
    {
      "user_id": "string",
      "username": "string",
      "join_time": "timestamp",
      "leave_time": "timestamp"
    }
  ],
  "recording_url": "string",
  "detection_summary": {
    "total_detections": "integer",
    "fake_detections": "integer",
    "suspicious_activities": [
      {
        "user_id": "string",
        "timestamp": "timestamp",
        "type": "string",
        "confidence": "float"
      }
    ]
  }
}
```

## WebSocket 信令协议

### 连接
```
ws://domain/signaling/{meeting_id}?token={jwt_token}
```

### 消息格式
```json
{
  "type": "offer|answer|ice-candidate|join|leave|detection-alert",
  "data": {},
  "timestamp": "timestamp",
  "from": "user_id",
  "to": "user_id"
}
```

### WebRTC信令消息

#### Offer
```json
{
  "type": "offer",
  "data": {
    "sdp": "string",
    "type": "offer"
  },
  "to": "user_id"
}
```

#### Answer
```json
{
  "type": "answer",
  "data": {
    "sdp": "string",
    "type": "answer"
  },
  "to": "user_id"
}
```

#### ICE Candidate
```json
{
  "type": "ice-candidate",
  "data": {
    "candidate": "string",
    "sdpMLineIndex": "integer",
    "sdpMid": "string"
  },
  "to": "user_id"
}
```

### 检测告警消息
```json
{
  "type": "detection-alert",
  "data": {
    "user_id": "string",
    "detection_type": "face_swap|voice_synthesis",
    "confidence": "float",
    "timestamp": "timestamp"
  }
}
```

## gRPC 服务接口

### 用户服务
```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}
```

### 会议服务
```protobuf
service MeetingService {
  rpc CreateMeeting(CreateMeetingRequest) returns (Meeting);
  rpc GetMeeting(GetMeetingRequest) returns (Meeting);
  rpc UpdateMeetingStatus(UpdateMeetingStatusRequest) returns (Meeting);
}
```

### 检测服务
```protobuf
service DetectionService {
  rpc AnalyzeMedia(stream AnalyzeMediaRequest) returns (stream AnalyzeMediaResponse);
  rpc GetDetectionResult(GetDetectionResultRequest) returns (DetectionResult);
}
```

## 错误码定义

| 错误码 | 描述 |
|--------|------|
| 1000 | 成功 |
| 4001 | 未授权 |
| 4003 | 禁止访问 |
| 4004 | 资源不存在 |
| 4009 | 冲突 |
| 4022 | 参数验证失败 |
| 5000 | 内部服务器错误 |
| 5003 | 服务不可用 |

## 限流策略

- 用户注册: 5次/小时/IP
- 用户登录: 10次/分钟/IP
- API调用: 1000次/小时/用户
- 文件上传: 100MB/次，10次/小时/用户
