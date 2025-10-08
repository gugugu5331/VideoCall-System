# 智能视频会议平台 API 文档

**版本**: v1.0.0  
**基础URL**: `http://gateway:8000`  
**协议**: HTTP/HTTPS + WebSocket  
**认证方式**: JWT Bearer Token

---

## 📋 目录

1. [认证与授权](#1-认证与授权)
2. [用户服务](#2-用户服务)
3. [会议服务](#3-会议服务)
4. [信令服务](#4-信令服务)
5. [媒体服务](#5-媒体服务)
6. [AI服务](#6-ai服务)
7. [数据模型](#7-数据模型)
8. [错误码](#8-错误码)
9. [限流规则](#9-限流规则)

---

## 1. 认证与授权

### 1.1 用户注册

**端点**: `POST /api/v1/auth/register`  
**认证**: 不需要  
**限流**: 5次/分钟

**请求体**:
```json
{
  "username": "string (必需, 3-50字符)",
  "email": "string (必需, 有效邮箱)",
  "password": "string (必需, 6-100字符)",
  "nickname": "string (可选, 最多50字符)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "User registered successfully",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "Test User",
    "created_at": "2025-10-02T10:00:00Z"
  }
}
```

---

### 1.2 用户登录

**端点**: `POST /api/v1/auth/login`  
**认证**: 不需要  
**限流**: 5次/分钟

**请求体**:
```json
{
  "username": "string (必需)",
  "password": "string (必需)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "user_id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "nickname": "Test User"
    }
  }
}
```

---

### 1.3 刷新Token

**端点**: `POST /api/v1/auth/refresh`  
**认证**: 需要 Refresh Token  
**限流**: 10次/分钟

**请求体**:
```json
{
  "refresh_token": "string (必需)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

---

### 1.4 忘记密码

**端点**: `POST /api/v1/auth/forgot-password`  
**认证**: 不需要  
**限流**: 3次/小时

**请求体**:
```json
{
  "email": "string (必需)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Password reset email sent"
}
```

---

### 1.5 重置密码

**端点**: `POST /api/v1/auth/reset-password`  
**认证**: 不需要  
**限流**: 5次/小时

**请求体**:
```json
{
  "token": "string (必需, 重置令牌)",
  "new_password": "string (必需, 6-100字符)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Password reset successfully"
}
```

---

## 2. 用户服务

### 2.1 获取用户资料

**端点**: `GET /api/v1/users/profile`  
**认证**: 需要 JWT Token  
**限流**: 100次/分钟

**请求头**:
```
Authorization: Bearer <token>
```

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "Test User",
    "avatar_url": "https://example.com/avatar.jpg",
    "status": "active",
    "created_at": "2025-10-02T10:00:00Z",
    "updated_at": "2025-10-02T10:00:00Z"
  }
}
```

---

### 2.2 更新用户资料

**端点**: `PUT /api/v1/users/profile`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**请求体**:
```json
{
  "nickname": "string (可选, 最多50字符)",
  "email": "string (可选, 有效邮箱)",
  "avatar_url": "string (可选, 有效URL)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Profile updated successfully",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "newemail@example.com",
    "nickname": "New Nickname",
    "avatar_url": "https://example.com/new-avatar.jpg"
  }
}
```

---

### 2.3 修改密码

**端点**: `POST /api/v1/users/change-password`  
**认证**: 需要 JWT Token  
**限流**: 10次/小时

**请求体**:
```json
{
  "old_password": "string (必需)",
  "new_password": "string (必需, 6-100字符)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Password changed successfully"
}
```

---

### 2.4 上传头像

**端点**: `POST /api/v1/users/upload-avatar`  
**认证**: 需要 JWT Token  
**限流**: 10次/小时  
**Content-Type**: `multipart/form-data`

**请求体**:
```
file: <binary> (必需, 图片文件, 最大5MB)
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Avatar uploaded successfully",
  "data": {
    "avatar_url": "https://example.com/avatars/user123.jpg"
  }
}
```

---

### 2.5 删除账户

**端点**: `DELETE /api/v1/users/account`  
**认证**: 需要 JWT Token  
**限流**: 1次/天

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "Account deleted successfully"
}
```

---

### 2.6 获取用户列表（管理员）

**端点**: `GET /api/v1/admin/users`  
**认证**: 需要 JWT Token + 管理员权限  
**限流**: 50次/分钟

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20, 最大: 100)
- `keyword`: 搜索关键词 (可选)
- `status`: 用户状态 (可选: active, banned, deleted)

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "users": [
      {
        "user_id": 1,
        "username": "testuser",
        "email": "test@example.com",
        "nickname": "Test User",
        "status": "active",
        "created_at": "2025-10-02T10:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 2.7 获取指定用户（管理员）

**端点**: `GET /api/v1/admin/users/:id`  
**认证**: 需要 JWT Token + 管理员权限  
**限流**: 100次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "nickname": "Test User",
    "avatar_url": "https://example.com/avatar.jpg",
    "status": "active",
    "created_at": "2025-10-02T10:00:00Z",
    "updated_at": "2025-10-02T10:00:00Z",
    "last_login": "2025-10-02T12:00:00Z"
  }
}
```

---

### 2.8 更新用户（管理员）

**端点**: `PUT /api/v1/admin/users/:id`  
**认证**: 需要 JWT Token + 管理员权限  
**限流**: 50次/分钟

**请求体**:
```json
{
  "nickname": "string (可选)",
  "email": "string (可选)",
  "status": "string (可选: active, banned)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "User updated successfully"
}
```

---

### 2.9 删除用户（管理员）

**端点**: `DELETE /api/v1/admin/users/:id`  
**认证**: 需要 JWT Token + 管理员权限  
**限流**: 20次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "User deleted successfully"
}
```

---

### 2.10 封禁用户（管理员）

**端点**: `POST /api/v1/admin/users/:id/ban`  
**认证**: 需要 JWT Token + 管理员权限  
**限流**: 20次/分钟

**请求体**:
```json
{
  "reason": "string (可选)",
  "duration": "number (可选, 封禁时长，单位：小时)"
}
```

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "User banned successfully"
}
```

---

### 2.11 解封用户（管理员）

**端点**: `POST /api/v1/admin/users/:id/unban`  
**认证**: 需要 JWT Token + 管理员权限  
**限流**: 20次/分钟

**响应** (200 OK):
```json
{
  "code": 200,
  "message": "User unbanned successfully"
}
```

---

## 3. 会议服务

### 3.1 创建会议

**端点**: `POST /api/v1/meetings`  
**认证**: 需要 JWT Token  
**限流**: 50次/分钟

**请求体**:
```json
{
  "title": "string (必需, 1-100字符)",
  "description": "string (可选, 最多500字符)",
  "start_time": "string (必需, ISO8601格式)",
  "end_time": "string (必需, ISO8601格式)",
  "max_participants": "number (必需, 1-1000)",
  "meeting_type": "string (必需, video|audio)",
  "password": "string (可选, 最多50字符)",
  "settings": {
    "enable_recording": "boolean (可选, 默认false)",
    "enable_chat": "boolean (可选, 默认true)",
    "enable_screen_share": "boolean (可选, 默认true)",
    "enable_waiting_room": "boolean (可选, 默认false)",
    "mute_on_join": "boolean (可选, 默认false)"
  }
}
```

**响应** (201 Created):
```json
{
  "code": 201,
  "message": "Meeting created successfully",
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
    "created_at": "2025-10-02T14:00:00Z"
  }
}
```

---


