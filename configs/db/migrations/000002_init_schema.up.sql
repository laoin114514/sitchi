BEGIN;

-- 000002: 扩展模块归属与模块成员关系（不初始化业务数据）

-- 1) modules 增加 owner 字段（模块负责人）
ALTER TABLE modules
ADD COLUMN IF NOT EXISTS owner_user_id INTEGER;

-- owner 外键（用户删除后置空）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_modules_owner_user'
    ) THEN
        ALTER TABLE modules
        ADD CONSTRAINT fk_modules_owner_user
        FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- owner 查询索引
CREATE INDEX IF NOT EXISTS idx_modules_owner_user_id ON modules(owner_user_id);

-- 2) 模块成员表：记录“模块内有哪些用户”
CREATE TABLE IF NOT EXISTS module_users (
    id SERIAL PRIMARY KEY,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description VARCHAR(200),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 仅未删除记录唯一（软删后允许复用）
CREATE UNIQUE INDEX IF NOT EXISTS uq_module_users_module_user_active
ON module_users(module_id, user_id)
WHERE is_deleted = FALSE;

-- 常用查询索引
CREATE INDEX IF NOT EXISTS idx_module_users_module_id ON module_users(module_id);
CREATE INDEX IF NOT EXISTS idx_module_users_user_id ON module_users(user_id);

COMMIT;
