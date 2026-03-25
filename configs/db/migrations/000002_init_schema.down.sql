BEGIN;

-- 000002 回滚：移除 owner 字段与模块成员表

-- 1) 删除模块成员表（先删依赖对象）
DROP TABLE IF EXISTS module_users;

-- 2) 删除 modules.owner_user_id 的索引与外键，再删字段
DROP INDEX IF EXISTS idx_modules_owner_user_id;

ALTER TABLE modules
DROP CONSTRAINT IF EXISTS fk_modules_owner_user;

ALTER TABLE modules
DROP COLUMN IF EXISTS owner_user_id;

COMMIT;
