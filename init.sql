
-- 模块表
CREATE TABLE modules (
    id SERIAL PRIMARY KEY,
    module_code VARCHAR(100) UNIQUE NOT NULL,
    module_name VARCHAR(100) NOT NULL,
    description VARCHAR(200),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户表
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_code VARCHAR(50) NOT NULL,
    user_name VARCHAR(50),
    description VARCHAR(200),
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    password VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id, user_code),
    UNIQUE(id, module_id)
);

-- 角色表
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    role_code VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id, role_code),
    UNIQUE(id, module_id)
);

-- 权限表
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    perm_code VARCHAR(50) NOT NULL,
    perm_name VARCHAR(50) NOT NULL,
    description VARCHAR(200),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id, perm_code),
    UNIQUE(id, module_id)
);

-- 资源表
CREATE TABLE resources (
    id SERIAL PRIMARY KEY,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    res_code VARCHAR(100) NOT NULL,
    res_name VARCHAR(100) NOT NULL,
    res_type VARCHAR(20) NOT NULL,
    parent_id INTEGER DEFAULT NULL, 
    path VARCHAR(200),
    description VARCHAR(200),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id, res_code),
    UNIQUE(id, module_id),
    FOREIGN KEY (parent_id, module_id) REFERENCES resources(id, module_id) ON DELETE SET NULL,
    
);


-- 用户-角色关联表
CREATE TABLE user_roles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id,user_id, role_id),
    FOREIGN KEY (user_id, module_id) REFERENCES users(id, module_id) ON DELETE CASCADE,
    FOREIGN KEY (role_id, module_id) REFERENCES roles(id, module_id) ON DELETE CASCADE
);

-- 角色-权限-资源关联表
CREATE TABLE role_permission_resources (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL,
    perm_id INTEGER NOT NULL,
    res_id INTEGER NOT NULL,
    module_id INTEGER NOT NULL REFERENCES modules(id) ON DELETE CASCADE,    
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module_id,role_id, perm_id, res_id),
    FOREIGN KEY (role_id, module_id) REFERENCES roles(id, module_id) ON DELETE CASCADE,
    FOREIGN KEY (perm_id, module_id) REFERENCES permissions(id, module_id) ON DELETE CASCADE,
    FOREIGN KEY (res_id, module_id) REFERENCES resources(id, module_id) ON DELETE CASCADE
);

-- 用户最终权限表
CREATE TABLE user_perm_res (
    module_id INT NOT NULL,
    user_id   INT NOT NULL,
    perm_id   INT NOT NULL,
    res_id    INT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (module_id, user_id, perm_id, res_id),
    FOREIGN KEY (user_id, module_id) REFERENCES users(id, module_id) ON DELETE CASCADE,
    FOREIGN KEY (perm_id, module_id) REFERENCES permissions(id, module_id) ON DELETE CASCADE,
    FOREIGN KEY (res_id,  module_id) REFERENCES resources(id, module_id) ON DELETE CASCADE
);


-- resources 树查询：按父节点找子节点
CREATE INDEX IF NOT EXISTS idx_resources_module_parent ON resources(module_id, parent_id);

-- user_roles：既要按 user 找角色，也要在 role 变更时按 role 找用户
CREATE INDEX IF NOT EXISTS idx_user_roles_module_user ON user_roles(module_id, user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_module_role ON user_roles(module_id, role_id);

-- role_permission_resources：按 role 找权限资源
CREATE INDEX IF NOT EXISTS idx_rpr_module_role ON role_permission_resources(module_id, role_id);

-- user_perm_res：常见按 user 拉取所有权限（PK 可用于 exists 校验；该索引利于按 user 扫描）
CREATE INDEX IF NOT EXISTS idx_upr_module_user ON user_perm_res(module_id, user_id);
