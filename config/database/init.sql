-- 音视频通话系统数据库初始化脚本

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url VARCHAR(255),
    phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'banned')),
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 通话记录表
CREATE TABLE IF NOT EXISTS calls (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    caller_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    callee_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    caller_uuid UUID REFERENCES users(uuid) ON DELETE SET NULL,
    callee_uuid UUID REFERENCES users(uuid) ON DELETE SET NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    duration INTEGER, -- 通话时长（秒）
    call_type VARCHAR(20) DEFAULT 'video' CHECK (call_type IN ('audio', 'video')),
    status VARCHAR(20) DEFAULT 'initiated' CHECK (status IN ('initiated', 'ringing', 'answered', 'ended', 'missed', 'rejected')),
    room_id VARCHAR(100), -- WebRTC房间ID
    recording_url VARCHAR(255), -- 录音/录像文件URL
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 安全检测记录表
CREATE TABLE IF NOT EXISTS security_detections (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    call_id INTEGER REFERENCES calls(id) ON DELETE CASCADE,
    call_uuid UUID REFERENCES calls(uuid) ON DELETE CASCADE,
    detection_type VARCHAR(20) NOT NULL CHECK (detection_type IN ('voice_spoofing', 'video_deepfake', 'face_swap')),
    risk_score DECIMAL(5,2) NOT NULL CHECK (risk_score >= 0 AND risk_score <= 100),
    confidence DECIMAL(5,2) NOT NULL CHECK (confidence >= 0 AND confidence <= 100),
    detection_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    details JSONB, -- 详细检测结果
    model_version VARCHAR(50), -- 使用的模型版本
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户会话表
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 模型版本管理表
CREATE TABLE IF NOT EXISTS model_versions (
    id SERIAL PRIMARY KEY,
    model_name VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    model_path VARCHAR(255) NOT NULL,
    accuracy DECIMAL(5,2),
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(model_name, version)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_calls_caller_id ON calls(caller_id);
CREATE INDEX IF NOT EXISTS idx_calls_callee_id ON calls(callee_id);
CREATE INDEX IF NOT EXISTS idx_calls_start_time ON calls(start_time);
CREATE INDEX IF NOT EXISTS idx_security_detections_call_id ON security_detections(call_id);
CREATE INDEX IF NOT EXISTS idx_security_detections_detection_time ON security_detections(detection_time);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要自动更新时间的表添加触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_calls_updated_at BEFORE UPDATE ON calls FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入初始系统配置
INSERT INTO system_configs (config_key, config_value, description) VALUES
('max_call_duration', '3600', '最大通话时长（秒）'),
('voice_detection_threshold', '0.7', '语音伪造检测阈值'),
('video_detection_threshold', '0.8', '视频伪造检测阈值'),
('session_timeout', '86400', '会话超时时间（秒）'),
('max_concurrent_calls', '1000', '最大并发通话数')
ON CONFLICT (config_key) DO NOTHING;

-- 插入默认模型版本
INSERT INTO model_versions (model_name, version, model_path, accuracy, is_active) VALUES
('voice_anti_spoofing', 'v1.0.0', '/app/models/voice_anti_spoofing_v1.0.0.pth', 95.5, true),
('video_deepfake_detection', 'v1.0.0', '/app/models/video_deepfake_v1.0.0.pth', 92.3, true)
ON CONFLICT (model_name, version) DO NOTHING;
