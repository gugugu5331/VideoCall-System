#!/bin/bash

# 数据库初始化脚本
# 用于初始化PostgreSQL、Redis、MongoDB和MinIO

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 配置变量
POSTGRES_HOST=${POSTGRES_HOST:-localhost}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_DB=${POSTGRES_DB:-meeting_system}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-password}

REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}

MONGODB_HOST=${MONGODB_HOST:-localhost}
MONGODB_PORT=${MONGODB_PORT:-27017}
MONGODB_DB=${MONGODB_DB:-meeting_system}
MONGODB_USER=${MONGODB_USER:-admin}
MONGODB_PASSWORD=${MONGODB_PASSWORD:-password}

MINIO_HOST=${MINIO_HOST:-localhost}
MINIO_PORT=${MINIO_PORT:-9000}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}
MINIO_BUCKET=${MINIO_BUCKET:-meeting-system}

# 等待服务启动
wait_for_service() {
    local host=$1
    local port=$2
    local service_name=$3
    local max_attempts=30
    local attempt=1

    log_info "Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if nc -z $host $port 2>/dev/null; then
            log_info "$service_name is ready!"
            return 0
        fi
        
        log_warn "Attempt $attempt/$max_attempts: $service_name not ready, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "$service_name failed to start within expected time"
    return 1
}

# 初始化PostgreSQL
init_postgres() {
    log_info "Initializing PostgreSQL..."
    
    wait_for_service $POSTGRES_HOST $POSTGRES_PORT "PostgreSQL"
    
    # 检查数据库是否存在
    if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -lqt | cut -d \| -f 1 | grep -qw $POSTGRES_DB; then
        log_info "Database $POSTGRES_DB already exists"
    else
        log_info "Creating database $POSTGRES_DB..."
        PGPASSWORD=$POSTGRES_PASSWORD createdb -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER $POSTGRES_DB
    fi
    
    # 执行schema脚本
    log_info "Executing database schema..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -f ../backend/shared/database/schema.sql
    
    # 插入测试数据
    log_info "Inserting test data..."
    PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB << EOF
-- 插入测试用户
INSERT INTO users (username, email, password_hash, nickname, status) VALUES
('admin', 'admin@example.com', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMye.Uo0.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc', '管理员', 1),
('testuser1', 'user1@example.com', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMye.Uo0.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc', '测试用户1', 1),
('testuser2', 'user2@example.com', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMye.Uo0.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc.Vc', '测试用户2', 1)
ON CONFLICT (username) DO NOTHING;

-- 插入测试会议
INSERT INTO meetings (title, description, creator_id, start_time, end_time, status) VALUES
('测试会议1', '这是一个测试会议', 1, NOW() + INTERVAL '1 hour', NOW() + INTERVAL '2 hours', 1),
('测试会议2', '这是另一个测试会议', 2, NOW() + INTERVAL '2 hours', NOW() + INTERVAL '3 hours', 1);

-- 插入系统配置
INSERT INTO system_configs (config_key, config_value, description, config_type, is_public) VALUES
('system_name', 'Meeting System', '系统名称', 'string', true),
('version', '1.0.0', '系统版本', 'string', true),
('maintenance_mode', 'false', '维护模式', 'boolean', false)
ON CONFLICT (config_key) DO NOTHING;
EOF
    
    log_info "PostgreSQL initialization completed"
}

# 初始化Redis
init_redis() {
    log_info "Initializing Redis..."
    
    wait_for_service $REDIS_HOST $REDIS_PORT "Redis"
    
    # 测试Redis连接
    redis-cli -h $REDIS_HOST -p $REDIS_PORT ping
    
    # 设置一些初始缓存数据
    redis-cli -h $REDIS_HOST -p $REDIS_PORT << EOF
SET system:status "running"
SET system:init_time "$(date -Iseconds)"
HSET system:stats users 0 meetings 0 active_connections 0
EXPIRE system:stats 3600
EOF
    
    log_info "Redis initialization completed"
}

# 初始化MongoDB
init_mongodb() {
    log_info "Initializing MongoDB..."
    
    wait_for_service $MONGODB_HOST $MONGODB_PORT "MongoDB"
    
    # 创建数据库和集合
    mongosh --host $MONGODB_HOST:$MONGODB_PORT --username $MONGODB_USER --password $MONGODB_PASSWORD --authenticationDatabase admin << EOF
use $MONGODB_DB

// 创建聊天消息集合
db.createCollection("chat_messages")
db.chat_messages.createIndex({"meeting_id": 1, "timestamp": -1})
db.chat_messages.createIndex({"user_id": 1})

// 创建AI分析结果集合
db.createCollection("ai_analysis_results")
db.ai_analysis_results.createIndex({"meeting_id": 1, "timestamp": -1})
db.ai_analysis_results.createIndex({"analysis_type": 1})

// 创建会议事件集合
db.createCollection("meeting_events")
db.meeting_events.createIndex({"meeting_id": 1, "timestamp": -1})
db.meeting_events.createIndex({"event_type": 1})

// 插入测试数据
db.chat_messages.insertOne({
    meeting_id: "test_meeting_1",
    user_id: "1",
    username: "admin",
    message_type: "text",
    content: "欢迎使用会议系统！",
    timestamp: new Date()
})

db.ai_analysis_results.insertOne({
    meeting_id: "test_meeting_1",
    analysis_type: "system_test",
    result: {
        status: "initialized",
        message: "AI分析系统已初始化"
    },
    confidence: 1.0,
    timestamp: new Date()
})

db.meeting_events.insertOne({
    meeting_id: "test_meeting_1",
    event_type: "system_init",
    data: {
        message: "系统初始化完成"
    },
    timestamp: new Date()
})
EOF
    
    log_info "MongoDB initialization completed"
}

# 初始化MinIO
init_minio() {
    log_info "Initializing MinIO..."
    
    wait_for_service $MINIO_HOST $MINIO_PORT "MinIO"
    
    # 配置MinIO客户端
    mc alias set myminio http://$MINIO_HOST:$MINIO_PORT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY
    
    # 创建存储桶
    if mc ls myminio/$MINIO_BUCKET >/dev/null 2>&1; then
        log_info "Bucket $MINIO_BUCKET already exists"
    else
        log_info "Creating bucket $MINIO_BUCKET..."
        mc mb myminio/$MINIO_BUCKET
    fi
    
    # 设置存储桶策略
    mc policy set public myminio/$MINIO_BUCKET
    
    # 创建子目录结构
    echo "test" | mc pipe myminio/$MINIO_BUCKET/users/.keep
    echo "test" | mc pipe myminio/$MINIO_BUCKET/meetings/.keep
    echo "test" | mc pipe myminio/$MINIO_BUCKET/recordings/.keep
    echo "test" | mc pipe myminio/$MINIO_BUCKET/temp/.keep
    
    log_info "MinIO initialization completed"
}

# 验证初始化
verify_initialization() {
    log_info "Verifying initialization..."
    
    # 验证PostgreSQL
    if PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d $POSTGRES_DB -c "SELECT COUNT(*) FROM users;" >/dev/null 2>&1; then
        log_info "PostgreSQL verification passed"
    else
        log_error "PostgreSQL verification failed"
        return 1
    fi
    
    # 验证Redis
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping >/dev/null 2>&1; then
        log_info "Redis verification passed"
    else
        log_error "Redis verification failed"
        return 1
    fi
    
    # 验证MongoDB
    if mongosh --host $MONGODB_HOST:$MONGODB_PORT --username $MONGODB_USER --password $MONGODB_PASSWORD --authenticationDatabase admin --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
        log_info "MongoDB verification passed"
    else
        log_error "MongoDB verification failed"
        return 1
    fi
    
    # 验证MinIO
    if mc ls myminio/$MINIO_BUCKET >/dev/null 2>&1; then
        log_info "MinIO verification passed"
    else
        log_error "MinIO verification failed"
        return 1
    fi
    
    log_info "All services verified successfully!"
}

# 主函数
main() {
    log_info "Starting database initialization..."
    
    # 检查必要的工具
    for tool in psql redis-cli mongosh mc nc; do
        if ! command -v $tool >/dev/null 2>&1; then
            log_error "$tool is not installed"
            exit 1
        fi
    done
    
    # 初始化各个服务
    init_postgres
    init_redis
    init_mongodb
    init_minio
    
    # 验证初始化
    verify_initialization
    
    log_info "Database initialization completed successfully!"
}

# 执行主函数
main "$@"
