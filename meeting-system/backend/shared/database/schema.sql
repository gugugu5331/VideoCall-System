-- 会议系统数据库Schema
-- PostgreSQL 数据库初始化脚本

-- 创建数据库（如果不存在）
-- CREATE DATABASE meeting_system;

-- 使用数据库
-- \c meeting_system;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(50),
    avatar_url VARCHAR(255),
    phone VARCHAR(20),
    role INTEGER DEFAULT 1, -- 0:访客, 1:普通用户, 2:版主, 3:管理员, 4:超级管理员
    status INTEGER DEFAULT 1, -- 0:未激活, 1:激活, 2:禁用
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 会议表
CREATE TABLE IF NOT EXISTS meetings (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    creator_id INTEGER NOT NULL REFERENCES users(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    max_participants INTEGER DEFAULT 100,
    status INTEGER DEFAULT 1, -- 1:已安排, 2:进行中, 3:已结束, 4:已取消
    meeting_type INTEGER DEFAULT 1, -- 1:公开, 2:私人
    password VARCHAR(50),
    recording_url VARCHAR(500),
    settings JSONB, -- 会议设置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 创建会议表索引
CREATE INDEX IF NOT EXISTS idx_meetings_creator_id ON meetings(creator_id);
CREATE INDEX IF NOT EXISTS idx_meetings_status ON meetings(status);
CREATE INDEX IF NOT EXISTS idx_meetings_start_time ON meetings(start_time);
CREATE INDEX IF NOT EXISTS idx_meetings_end_time ON meetings(end_time);
CREATE INDEX IF NOT EXISTS idx_meetings_deleted_at ON meetings(deleted_at);

-- 会议参与者表
CREATE TABLE IF NOT EXISTS meeting_participants (
    id SERIAL PRIMARY KEY,
    meeting_id INTEGER NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    role INTEGER DEFAULT 1, -- 1:参与者, 2:主持人, 3:演示者
    status INTEGER DEFAULT 1, -- 1:已邀请, 2:已加入, 3:已离开, 4:已拒绝
    joined_at TIMESTAMP,
    left_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(meeting_id, user_id)
);

-- 创建会议参与者表索引
CREATE INDEX IF NOT EXISTS idx_meeting_participants_meeting_id ON meeting_participants(meeting_id);
CREATE INDEX IF NOT EXISTS idx_meeting_participants_user_id ON meeting_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_meeting_participants_status ON meeting_participants(status);
CREATE INDEX IF NOT EXISTS idx_meeting_participants_deleted_at ON meeting_participants(deleted_at);

-- 会议房间表（WebRTC房间信息）
CREATE TABLE IF NOT EXISTS meeting_rooms (
    id SERIAL PRIMARY KEY,
    meeting_id INTEGER NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    room_id VARCHAR(100) UNIQUE NOT NULL,
    sfu_node VARCHAR(100), -- SFU节点地址
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, closed
    participant_count INTEGER DEFAULT 0,
    max_bitrate INTEGER DEFAULT 1000000, -- 最大码率
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建会议房间表索引
CREATE INDEX IF NOT EXISTS idx_meeting_rooms_meeting_id ON meeting_rooms(meeting_id);
CREATE INDEX IF NOT EXISTS idx_meeting_rooms_room_id ON meeting_rooms(room_id);
CREATE INDEX IF NOT EXISTS idx_meeting_rooms_status ON meeting_rooms(status);

-- 媒体流表
CREATE TABLE IF NOT EXISTS media_streams (
    id SERIAL PRIMARY KEY,
    room_id VARCHAR(100) NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    stream_id VARCHAR(100) NOT NULL,
    stream_type INTEGER NOT NULL, -- 1:音频, 2:视频, 3:屏幕共享
    codec VARCHAR(50),
    bitrate INTEGER,
    resolution VARCHAR(20), -- 如: 1920x1080
    status INTEGER DEFAULT 1, -- 1:活跃, 2:暂停, 3:停止
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建媒体流表索引
CREATE INDEX IF NOT EXISTS idx_media_streams_room_id ON media_streams(room_id);
CREATE INDEX IF NOT EXISTS idx_media_streams_user_id ON media_streams(user_id);
CREATE INDEX IF NOT EXISTS idx_media_streams_stream_type ON media_streams(stream_type);
CREATE INDEX IF NOT EXISTS idx_media_streams_status ON media_streams(status);

-- 会议录制表
CREATE TABLE IF NOT EXISTS meeting_recordings (
    id SERIAL PRIMARY KEY,
    meeting_id INTEGER NOT NULL REFERENCES meetings(id),
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT,
    duration INTEGER, -- 录制时长（秒）
    format VARCHAR(20), -- 文件格式
    status INTEGER DEFAULT 1, -- 1:录制中, 2:已完成, 3:失败
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建会议录制表索引
CREATE INDEX IF NOT EXISTS idx_meeting_recordings_meeting_id ON meeting_recordings(meeting_id);
CREATE INDEX IF NOT EXISTS idx_meeting_recordings_status ON meeting_recordings(status);

-- 信令会话表（WebSocket 会话/房间状态）
CREATE TABLE IF NOT EXISTS signaling_sessions (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(64) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    meeting_id INTEGER NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    peer_id VARCHAR(64) NOT NULL,
    status INTEGER DEFAULT 1, -- 1:连接中, 2:已连接, 3:Offering, 4:Answering, 5:Stable, 6:断开, 7:失败
    joined_at TIMESTAMP NOT NULL,
    last_ping_at TIMESTAMP,
    disconnected_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_signaling_sessions_session_id ON signaling_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_signaling_sessions_user_id ON signaling_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_signaling_sessions_meeting_id ON signaling_sessions(meeting_id);
CREATE INDEX IF NOT EXISTS idx_signaling_sessions_status ON signaling_sessions(status);
CREATE INDEX IF NOT EXISTS idx_signaling_sessions_deleted_at ON signaling_sessions(deleted_at);

-- 信令消息表（用于持久化部分信令/聊天消息）
CREATE TABLE IF NOT EXISTS signaling_messages (
    id SERIAL PRIMARY KEY,
    message_id VARCHAR(64) UNIQUE NOT NULL,
    session_id VARCHAR(64) NOT NULL,
    from_user_id INTEGER NOT NULL REFERENCES users(id),
    to_user_id INTEGER REFERENCES users(id),
    meeting_id INTEGER NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    message_type INTEGER NOT NULL,
    payload TEXT,
    status INTEGER DEFAULT 1, -- 1:待发送, 2:已发送, 3:已送达, 4:失败
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_signaling_messages_message_id ON signaling_messages(message_id);
CREATE INDEX IF NOT EXISTS idx_signaling_messages_session_id ON signaling_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_signaling_messages_meeting_id ON signaling_messages(meeting_id);
CREATE INDEX IF NOT EXISTS idx_signaling_messages_from_user_id ON signaling_messages(from_user_id);
CREATE INDEX IF NOT EXISTS idx_signaling_messages_to_user_id ON signaling_messages(to_user_id);
CREATE INDEX IF NOT EXISTS idx_signaling_messages_deleted_at ON signaling_messages(deleted_at);

-- AI分析任务表
CREATE TABLE IF NOT EXISTS ai_tasks (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(100) UNIQUE NOT NULL,
    meeting_id INTEGER REFERENCES meetings(id),
    user_id INTEGER REFERENCES users(id),
    task_type VARCHAR(50) NOT NULL, -- speech_recognition, emotion_detection, etc.
    input_data TEXT, -- 输入数据（JSON格式）
    output_data TEXT, -- 输出结果（JSON格式）
    status INTEGER DEFAULT 1, -- 1:待处理, 2:处理中, 3:已完成, 4:失败
    priority INTEGER DEFAULT 5, -- 优先级 1-10
    retry_count INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建AI任务表索引
CREATE INDEX IF NOT EXISTS idx_ai_tasks_task_id ON ai_tasks(task_id);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_meeting_id ON ai_tasks(meeting_id);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_task_type ON ai_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_status ON ai_tasks(status);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_priority ON ai_tasks(priority);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    description TEXT,
    config_type VARCHAR(20) DEFAULT 'string', -- string, number, boolean, json
    is_public BOOLEAN DEFAULT FALSE, -- 是否对外公开
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建系统配置表索引
CREATE INDEX IF NOT EXISTS idx_system_configs_key ON system_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_system_configs_public ON system_configs(is_public);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    operation VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50), -- user, meeting, etc.
    resource_id INTEGER,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建操作日志表索引
CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_operation ON operation_logs(operation);
CREATE INDEX IF NOT EXISTS idx_operation_logs_resource ON operation_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON operation_logs(created_at);

-- 插入默认系统配置
INSERT INTO system_configs (config_key, config_value, description, config_type, is_public) VALUES
('max_meeting_duration', '480', '最大会议时长（分钟）', 'number', TRUE),
('max_participants_per_meeting', '100', '每个会议最大参与者数', 'number', TRUE),
('enable_recording', 'true', '是否启用会议录制', 'boolean', TRUE),
('enable_ai_features', 'true', '是否启用AI功能', 'boolean', TRUE),
('default_video_quality', '720p', '默认视频质量', 'string', TRUE),
('max_file_upload_size', '100', '最大文件上传大小（MB）', 'number', TRUE)
ON CONFLICT (config_key) DO NOTHING;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表创建更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_meetings_updated_at BEFORE UPDATE ON meetings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_meeting_participants_updated_at BEFORE UPDATE ON meeting_participants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_meeting_rooms_updated_at BEFORE UPDATE ON meeting_rooms FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_media_streams_updated_at BEFORE UPDATE ON media_streams FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_meeting_recordings_updated_at BEFORE UPDATE ON meeting_recordings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ai_tasks_updated_at BEFORE UPDATE ON ai_tasks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建视图：活跃会议统计
CREATE OR REPLACE VIEW active_meetings_stats AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_meetings,
    COUNT(CASE WHEN status = 2 THEN 1 END) as ongoing_meetings,
    COUNT(CASE WHEN status = 3 THEN 1 END) as completed_meetings,
    AVG(EXTRACT(EPOCH FROM (end_time - start_time))/60) as avg_duration_minutes
FROM meetings 
WHERE deleted_at IS NULL
GROUP BY DATE(created_at)
ORDER BY date DESC;
