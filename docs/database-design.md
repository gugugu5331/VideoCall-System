# 数据库设计文档

## 概述

系统采用多数据库架构：
- PostgreSQL: 存储结构化数据（用户、会议、检测结果等）
- MongoDB: 存储非结构化数据（日志、记录等）
- Redis: 缓存和会话存储

## PostgreSQL 数据库设计

### 用户表 (users)
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
```

### 会议表 (meetings)
```sql
CREATE TABLE meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration INTEGER, -- 预计时长(分钟)
    max_participants INTEGER DEFAULT 50,
    is_public BOOLEAN DEFAULT false,
    join_code VARCHAR(20) UNIQUE,
    status VARCHAR(20) DEFAULT 'scheduled', -- scheduled, active, ended, cancelled
    recording_enabled BOOLEAN DEFAULT true,
    detection_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_meetings_creator ON meetings(creator_id);
CREATE INDEX idx_meetings_start_time ON meetings(start_time);
CREATE INDEX idx_meetings_status ON meetings(status);
CREATE INDEX idx_meetings_join_code ON meetings(join_code);
```

### 会议参与者表 (meeting_participants)
```sql
CREATE TABLE meeting_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID NOT NULL REFERENCES meetings(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'participant', -- host, moderator, participant
    join_time TIMESTAMP,
    leave_time TIMESTAMP,
    status VARCHAR(20) DEFAULT 'invited', -- invited, joined, left, kicked
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(meeting_id, user_id)
);

CREATE INDEX idx_participants_meeting ON meeting_participants(meeting_id);
CREATE INDEX idx_participants_user ON meeting_participants(user_id);
CREATE INDEX idx_participants_status ON meeting_participants(status);
```

### 检测任务表 (detection_tasks)
```sql
CREATE TABLE detection_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID REFERENCES meetings(id),
    user_id UUID NOT NULL REFERENCES users(id),
    file_path VARCHAR(500) NOT NULL,
    file_type VARCHAR(20) NOT NULL, -- video, audio, image
    file_size BIGINT,
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    priority INTEGER DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX idx_detection_tasks_status ON detection_tasks(status);
CREATE INDEX idx_detection_tasks_meeting ON detection_tasks(meeting_id);
CREATE INDEX idx_detection_tasks_user ON detection_tasks(user_id);
CREATE INDEX idx_detection_tasks_priority ON detection_tasks(priority);
```

### 检测结果表 (detection_results)
```sql
CREATE TABLE detection_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES detection_tasks(id),
    is_fake BOOLEAN NOT NULL,
    confidence FLOAT NOT NULL,
    detection_type VARCHAR(50) NOT NULL, -- face_swap, voice_synthesis, deepfake
    model_version VARCHAR(50),
    processing_time INTEGER, -- 处理时间(毫秒)
    details JSONB, -- 详细检测信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_detection_results_task ON detection_results(task_id);
CREATE INDEX idx_detection_results_fake ON detection_results(is_fake);
CREATE INDEX idx_detection_results_confidence ON detection_results(confidence);
CREATE INDEX idx_detection_results_type ON detection_results(detection_type);
```

### 系统配置表 (system_configs)
```sql
CREATE TABLE system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    category VARCHAR(50),
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_configs_key ON system_configs(key);
CREATE INDEX idx_configs_category ON system_configs(category);
```

## MongoDB 集合设计

### 通讯记录集合 (communications)
```javascript
{
  _id: ObjectId,
  meeting_id: String,
  user_id: String,
  username: String,
  message_type: String, // text, audio, video, file, system
  content: {
    text: String,
    file_url: String,
    file_name: String,
    file_size: Number,
    duration: Number // 音视频时长
  },
  timestamp: Date,
  metadata: {
    client_ip: String,
    user_agent: String,
    device_info: Object
  }
}

// 索引
db.communications.createIndex({ meeting_id: 1, timestamp: -1 })
db.communications.createIndex({ user_id: 1, timestamp: -1 })
db.communications.createIndex({ timestamp: -1 })
```

### 会议记录集合 (meeting_records)
```javascript
{
  _id: ObjectId,
  meeting_id: String,
  title: String,
  start_time: Date,
  end_time: Date,
  participants: [
    {
      user_id: String,
      username: String,
      join_time: Date,
      leave_time: Date,
      total_duration: Number,
      speaking_time: Number,
      detection_alerts: Number
    }
  ],
  recording: {
    file_url: String,
    file_size: Number,
    duration: Number,
    format: String
  },
  detection_summary: {
    total_detections: Number,
    fake_detections: Number,
    suspicious_activities: [
      {
        user_id: String,
        timestamp: Date,
        type: String,
        confidence: Number,
        details: Object
      }
    ]
  },
  statistics: {
    peak_participants: Number,
    total_messages: Number,
    total_files_shared: Number,
    network_quality: Object
  },
  created_at: Date
}

// 索引
db.meeting_records.createIndex({ meeting_id: 1 })
db.meeting_records.createIndex({ start_time: -1 })
db.meeting_records.createIndex({ "participants.user_id": 1 })
```

### 系统日志集合 (system_logs)
```javascript
{
  _id: ObjectId,
  level: String, // debug, info, warn, error, fatal
  service: String, // user-service, meeting-service, etc.
  message: String,
  details: Object,
  user_id: String,
  meeting_id: String,
  request_id: String,
  timestamp: Date,
  metadata: {
    ip_address: String,
    user_agent: String,
    endpoint: String,
    method: String,
    response_time: Number,
    status_code: Number
  }
}

// 索引
db.system_logs.createIndex({ timestamp: -1 })
db.system_logs.createIndex({ level: 1, timestamp: -1 })
db.system_logs.createIndex({ service: 1, timestamp: -1 })
db.system_logs.createIndex({ user_id: 1, timestamp: -1 })
```

### 检测日志集合 (detection_logs)
```javascript
{
  _id: ObjectId,
  task_id: String,
  meeting_id: String,
  user_id: String,
  detection_type: String,
  input_file: {
    path: String,
    size: Number,
    format: String,
    duration: Number
  },
  processing_steps: [
    {
      step: String,
      start_time: Date,
      end_time: Date,
      status: String,
      details: Object
    }
  ],
  result: {
    is_fake: Boolean,
    confidence: Number,
    model_used: String,
    processing_time: Number,
    details: Object
  },
  created_at: Date
}

// 索引
db.detection_logs.createIndex({ task_id: 1 })
db.detection_logs.createIndex({ meeting_id: 1, created_at: -1 })
db.detection_logs.createIndex({ user_id: 1, created_at: -1 })
```

## Redis 缓存设计

### 会话缓存
```
Key: session:{user_id}
Value: {
  token: String,
  expires_at: Timestamp,
  user_info: Object,
  permissions: Array
}
TTL: 24小时
```

### 会议状态缓存
```
Key: meeting:{meeting_id}:status
Value: {
  status: String,
  participants: Array,
  start_time: Timestamp,
  last_activity: Timestamp
}
TTL: 会议结束后1小时
```

### 检测结果缓存
```
Key: detection:{task_id}
Value: {
  status: String,
  result: Object,
  created_at: Timestamp
}
TTL: 7天
```

### 限流缓存
```
Key: rate_limit:{user_id}:{endpoint}
Value: {
  count: Number,
  reset_time: Timestamp
}
TTL: 1小时
```

## 数据备份策略

### PostgreSQL
- 每日全量备份
- 每小时增量备份
- WAL日志实时备份
- 保留30天备份

### MongoDB
- 每日全量备份
- Oplog实时备份
- 分片集群备份
- 保留30天备份

### Redis
- RDB快照备份(每6小时)
- AOF日志备份
- 主从复制
- 保留7天备份

## 性能优化

### 索引策略
- 为所有外键创建索引
- 为查询频繁的字段创建复合索引
- 定期分析查询性能并优化索引

### 分区策略
- 按时间分区大表(如logs表)
- 按meeting_id分区相关表
- 定期清理历史数据

### 缓存策略
- 热点数据Redis缓存
- 查询结果缓存
- 会话状态缓存
- CDN静态资源缓存
