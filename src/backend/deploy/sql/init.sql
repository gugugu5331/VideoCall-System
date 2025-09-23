-- 视频会议系统数据库初始化脚本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS video_conference;
USE video_conference;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    last_login_at TIMESTAMP
);

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 会议表
CREATE TABLE IF NOT EXISTS meetings (
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 会议表索引
CREATE INDEX IF NOT EXISTS idx_meetings_creator ON meetings(creator_id);
CREATE INDEX IF NOT EXISTS idx_meetings_start_time ON meetings(start_time);
CREATE INDEX IF NOT EXISTS idx_meetings_status ON meetings(status);
CREATE INDEX IF NOT EXISTS idx_meetings_join_code ON meetings(join_code);
CREATE INDEX IF NOT EXISTS idx_meetings_deleted_at ON meetings(deleted_at);

-- 会议参与者表
CREATE TABLE IF NOT EXISTS meeting_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    meeting_id UUID NOT NULL REFERENCES meetings(id),
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'participant', -- host, moderator, participant
    join_time TIMESTAMP,
    leave_time TIMESTAMP,
    status VARCHAR(20) DEFAULT 'invited', -- invited, joined, left, kicked
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(meeting_id, user_id)
);

-- 会议参与者表索引
CREATE INDEX IF NOT EXISTS idx_participants_meeting ON meeting_participants(meeting_id);
CREATE INDEX IF NOT EXISTS idx_participants_user ON meeting_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_participants_status ON meeting_participants(status);
CREATE INDEX IF NOT EXISTS idx_participants_deleted_at ON meeting_participants(deleted_at);

-- 检测任务表
CREATE TABLE IF NOT EXISTS detection_tasks (
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
    completed_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 检测任务表索引
CREATE INDEX IF NOT EXISTS idx_detection_tasks_status ON detection_tasks(status);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_meeting ON detection_tasks(meeting_id);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_user ON detection_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_priority ON detection_tasks(priority);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_created_at ON detection_tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_detection_tasks_deleted_at ON detection_tasks(deleted_at);

-- 检测结果表
CREATE TABLE IF NOT EXISTS detection_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES detection_tasks(id),
    is_fake BOOLEAN NOT NULL,
    confidence FLOAT NOT NULL,
    detection_type VARCHAR(50) NOT NULL, -- face_swap, voice_synthesis, deepfake
    model_version VARCHAR(50),
    processing_time INTEGER, -- 处理时间(毫秒)
    details JSONB, -- 详细检测信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 检测结果表索引
CREATE INDEX IF NOT EXISTS idx_detection_results_task ON detection_results(task_id);
CREATE INDEX IF NOT EXISTS idx_detection_results_fake ON detection_results(is_fake);
CREATE INDEX IF NOT EXISTS idx_detection_results_confidence ON detection_results(confidence);
CREATE INDEX IF NOT EXISTS idx_detection_results_type ON detection_results(detection_type);
CREATE INDEX IF NOT EXISTS idx_detection_results_created_at ON detection_results(created_at);
CREATE INDEX IF NOT EXISTS idx_detection_results_deleted_at ON detection_results(deleted_at);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    description TEXT,
    category VARCHAR(50),
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 系统配置表索引
CREATE INDEX IF NOT EXISTS idx_configs_key ON system_configs(key);
CREATE INDEX IF NOT EXISTS idx_configs_category ON system_configs(category);

-- 插入默认系统配置
INSERT INTO system_configs (key, value, description, category) VALUES
('max_meeting_duration', '480', '最大会议时长(分钟)', 'meeting'),
('max_participants', '100', '最大参与者数量', 'meeting'),
('recording_enabled', 'true', '是否启用录制功能', 'recording'),
('detection_enabled', 'true', '是否启用检测功能', 'detection'),
('file_upload_max_size', '104857600', '文件上传最大大小(字节)', 'upload'),
('session_timeout', '86400', '会话超时时间(秒)', 'auth'),
('password_min_length', '6', '密码最小长度', 'auth'),
('jwt_expire_time', '86400', 'JWT过期时间(秒)', 'auth')
ON CONFLICT (key) DO NOTHING;

-- 创建触发器函数：更新updated_at字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_meetings_updated_at BEFORE UPDATE ON meetings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建视图：活跃会议
CREATE OR REPLACE VIEW active_meetings AS
SELECT 
    m.*,
    u.username as creator_username,
    u.full_name as creator_full_name,
    COUNT(mp.id) as participant_count
FROM meetings m
LEFT JOIN users u ON m.creator_id = u.id
LEFT JOIN meeting_participants mp ON m.id = mp.meeting_id AND mp.status = 'joined'
WHERE m.status = 'active' AND m.deleted_at IS NULL
GROUP BY m.id, u.username, u.full_name;

-- 创建视图：用户统计
CREATE OR REPLACE VIEW user_statistics AS
SELECT 
    u.id,
    u.username,
    u.full_name,
    COUNT(DISTINCT m.id) as meetings_created,
    COUNT(DISTINCT mp.meeting_id) as meetings_participated,
    COUNT(DISTINCT dt.id) as detection_tasks,
    COUNT(DISTINCT CASE WHEN dr.is_fake = true THEN dr.id END) as fake_detections
FROM users u
LEFT JOIN meetings m ON u.id = m.creator_id AND m.deleted_at IS NULL
LEFT JOIN meeting_participants mp ON u.id = mp.user_id AND mp.deleted_at IS NULL
LEFT JOIN detection_tasks dt ON u.id = dt.user_id AND dt.deleted_at IS NULL
LEFT JOIN detection_results dr ON dt.id = dr.task_id AND dr.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.username, u.full_name;

-- 创建视图：检测统计
CREATE OR REPLACE VIEW detection_statistics AS
SELECT 
    DATE(dt.created_at) as date,
    dt.file_type,
    COUNT(*) as total_tasks,
    COUNT(CASE WHEN dt.status = 'completed' THEN 1 END) as completed_tasks,
    COUNT(CASE WHEN dt.status = 'failed' THEN 1 END) as failed_tasks,
    COUNT(CASE WHEN dr.is_fake = true THEN 1 END) as fake_detections,
    AVG(dr.confidence) as avg_confidence,
    AVG(dr.processing_time) as avg_processing_time
FROM detection_tasks dt
LEFT JOIN detection_results dr ON dt.id = dr.task_id
WHERE dt.deleted_at IS NULL
GROUP BY DATE(dt.created_at), dt.file_type
ORDER BY date DESC;

-- 创建函数：清理过期数据
CREATE OR REPLACE FUNCTION cleanup_expired_data()
RETURNS void AS $$
BEGIN
    -- 清理30天前的系统日志（在MongoDB中）
    -- 这里只是示例，实际清理需要在应用层处理
    
    -- 清理已完成的检测任务文件（保留结果）
    UPDATE detection_tasks 
    SET file_path = NULL 
    WHERE status = 'completed' 
    AND completed_at < NOW() - INTERVAL '7 days'
    AND file_path IS NOT NULL;
    
    -- 软删除90天前的已结束会议
    UPDATE meetings 
    SET deleted_at = NOW() 
    WHERE status = 'ended' 
    AND end_time < NOW() - INTERVAL '90 days'
    AND deleted_at IS NULL;
    
END;
$$ LANGUAGE plpgsql;

-- 创建定期清理任务（需要pg_cron扩展）
-- SELECT cron.schedule('cleanup-expired-data', '0 2 * * *', 'SELECT cleanup_expired_data();');

-- 创建性能监控视图
CREATE OR REPLACE VIEW performance_metrics AS
SELECT 
    'meetings' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN created_at > NOW() - INTERVAL '24 hours' THEN 1 END) as records_last_24h,
    COUNT(CASE WHEN created_at > NOW() - INTERVAL '7 days' THEN 1 END) as records_last_7d
FROM meetings WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'users' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN created_at > NOW() - INTERVAL '24 hours' THEN 1 END) as records_last_24h,
    COUNT(CASE WHEN created_at > NOW() - INTERVAL '7 days' THEN 1 END) as records_last_7d
FROM users WHERE deleted_at IS NULL
UNION ALL
SELECT 
    'detection_tasks' as table_name,
    COUNT(*) as total_records,
    COUNT(CASE WHEN created_at > NOW() - INTERVAL '24 hours' THEN 1 END) as records_last_24h,
    COUNT(CASE WHEN created_at > NOW() - INTERVAL '7 days' THEN 1 END) as records_last_7d
FROM detection_tasks WHERE deleted_at IS NULL;

-- 授权（如果需要特定用户）
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO video_conference_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO video_conference_user;
-- GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO video_conference_user;
