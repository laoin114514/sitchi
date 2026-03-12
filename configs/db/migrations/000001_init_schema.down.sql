-- configs/db/migrations/000001_init_schema.down.sql
DROP INDEX IF EXISTS uq_resources_module_res_code_active;
DROP INDEX IF EXISTS uq_permissions_module_perm_code_active;
DROP INDEX IF EXISTS uq_roles_module_role_code_active;
DROP INDEX IF EXISTS uq_users_module_user_code_active;
DROP INDEX IF EXISTS uq_modules_module_code_active;

DROP TABLE IF EXISTS user_perm_res;
DROP TABLE IF EXISTS role_permission_resources;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS resources;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS modules;