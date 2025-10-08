-- 添加用户角色字段迁移脚本
-- 执行时间: 2025-01-XX

-- 添加 role 字段（如果不存在）
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE users ADD COLUMN role INTEGER DEFAULT 1;
        COMMENT ON COLUMN users.role IS '用户角色: 0=访客, 1=普通用户, 2=版主, 3=管理员, 4=超级管理员';
    END IF;
END $$;

-- 创建索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- 更新现有用户的角色为普通用户（如果为NULL）
UPDATE users SET role = 1 WHERE role IS NULL;

-- 设置 role 字段为 NOT NULL
ALTER TABLE users ALTER COLUMN role SET NOT NULL;

-- 提交
COMMIT;

