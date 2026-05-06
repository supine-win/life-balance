-- LifeRecorder 初始数据库迁移
-- 创建所有表结构

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(64) DEFAULT '',
    avatar VARCHAR(512) DEFAULT '',
    status TINYINT DEFAULT 1,
    last_login DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255) DEFAULT '',
    sort_order INT DEFAULT 0,
    status TINYINT DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    resource VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL,
    description VARCHAR(255) DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);

-- 用户-角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

-- 角色-权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(36) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- 刷新令牌表
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token VARCHAR(512) NOT NULL UNIQUE,
    device_info VARCHAR(255) DEFAULT '',
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_expires ON refresh_tokens(expires_at);

-- 事件表
CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    event_time DATETIME NOT NULL,
    end_time DATETIME,
    location VARCHAR(255) DEFAULT '',
    latitude DECIMAL(10, 7),
    longitude DECIMAL(10, 7),
    mood VARCHAR(32) DEFAULT '',
    tags JSON,
    participants JSON,
    category VARCHAR(64) DEFAULT '',
    color VARCHAR(16) DEFAULT '',
    priority TINYINT DEFAULT 0,
    is_confirmed BOOLEAN DEFAULT FALSE,
    source VARCHAR(32) DEFAULT 'text',
    source_data JSON,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_events_user_time ON events(user_id, event_time DESC);
CREATE INDEX IF NOT EXISTS idx_events_user_category ON events(user_id, category);
CREATE INDEX IF NOT EXISTS idx_events_confirmed ON events(user_id, is_confirmed);
CREATE INDEX IF NOT EXISTS idx_events_created ON events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_deleted_at ON events(deleted_at);

-- 媒体表
CREATE TABLE IF NOT EXISTS media (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    event_id VARCHAR(36),
    type VARCHAR(16) NOT NULL,
    storage_adapter VARCHAR(32) NOT NULL DEFAULT 'local',
    storage_path VARCHAR(512) NOT NULL,
    original_name VARCHAR(255) DEFAULT '',
    thumbnail_path VARCHAR(512) DEFAULT '',
    mime_type VARCHAR(128) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    width INT,
    height INT,
    duration DECIMAL(10, 2),
    file_hash VARCHAR(64) DEFAULT '',
    metadata JSON,
    sort_order INT DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_media_user ON media(user_id);
CREATE INDEX IF NOT EXISTS idx_media_event ON media(event_id);
CREATE INDEX IF NOT EXISTS idx_media_type ON media(user_id, type);
CREATE INDEX IF NOT EXISTS idx_media_hash ON media(file_hash);

-- 媒体元数据表（多媒体智能提取）
CREATE TABLE IF NOT EXISTS media_metadata (
    id VARCHAR(36) PRIMARY KEY,
    media_id VARCHAR(36) NOT NULL UNIQUE,
    shot_time DATETIME,
    shot_time_src VARCHAR(32) DEFAULT '',
    gps_latitude DECIMAL(10, 7),
    gps_longitude DECIMAL(10, 7),
    gps_altitude DECIMAL(10, 2),
    location_name VARCHAR(255) DEFAULT '',
    location_src VARCHAR(32) DEFAULT '',
    camera_make VARCHAR(128) DEFAULT '',
    camera_model VARCHAR(128) DEFAULT '',
    lens_model VARCHAR(128) DEFAULT '',
    focal_length DECIMAL(8, 2),
    aperture DECIMAL(4, 2),
    shutter_speed VARCHAR(32) DEFAULT '',
    iso INT,
    video_codec VARCHAR(32) DEFAULT '',
    audio_codec VARCHAR(32) DEFAULT '',
    frame_rate DECIMAL(8, 2),
    bitrate BIGINT,
    iptc_keywords JSON,
    iptc_caption TEXT,
    filename_pattern VARCHAR(64) DEFAULT '',
    raw_exif JSON,
    suggestions JSON,
    extract_status VARCHAR(16) DEFAULT 'pending',
    extract_error TEXT,
    extracted_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 标签表
CREATE TABLE IF NOT EXISTS tags (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(64) NOT NULL,
    color VARCHAR(16) DEFAULT '',
    usage_count INT DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(64) NOT NULL,
    icon VARCHAR(64) DEFAULT '',
    color VARCHAR(16) DEFAULT '',
    sort_order INT DEFAULT 0,
    parent_id VARCHAR(36),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 配置表
CREATE TABLE IF NOT EXISTS configs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    scope VARCHAR(16) NOT NULL DEFAULT 'system',
    category VARCHAR(64) NOT NULL,
    key_name VARCHAR(128) NOT NULL,
    value TEXT NOT NULL,
    value_type VARCHAR(16) DEFAULT 'string',
    description VARCHAR(255) DEFAULT '',
    is_public BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_configs_category ON configs(category);

-- 展示模板表
CREATE TABLE IF NOT EXISTS display_templates (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(128) NOT NULL,
    type VARCHAR(32) NOT NULL,
    config JSON NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    is_builtin BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 幻灯片表
CREATE TABLE IF NOT EXISTS slideshows (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    template_id VARCHAR(36),
    event_ids JSON,
    media_ids JSON,
    bg_music_id VARCHAR(36),
    config JSON,
    duration INT DEFAULT 0,
    is_published BOOLEAN DEFAULT FALSE,
    published_url VARCHAR(255) DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Webhook 配置表
CREATE TABLE IF NOT EXISTS webhooks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(128) NOT NULL,
    url VARCHAR(512) NOT NULL,
    secret VARCHAR(255) DEFAULT '',
    events JSON NOT NULL,
    headers JSON,
    is_active BOOLEAN DEFAULT TRUE,
    retry_count TINYINT DEFAULT 3,
    timeout_ms INT DEFAULT 5000,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Webhook 投递日志表
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id VARCHAR(36) PRIMARY KEY,
    webhook_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSON NOT NULL,
    request_headers JSON,
    response_status INT,
    response_body TEXT,
    duration_ms INT,
    success BOOLEAN DEFAULT FALSE,
    attempts TINYINT DEFAULT 0,
    next_retry DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_deliveries_webhook ON webhook_deliveries(webhook_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_deliveries_retry ON webhook_deliveries(success, next_retry);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    username VARCHAR(64),
    action VARCHAR(64) NOT NULL,
    resource VARCHAR(64) NOT NULL,
    resource_id VARCHAR(36),
    detail JSON,
    ip_address VARCHAR(45),
    user_agent VARCHAR(512),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_logs(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_time ON audit_logs(created_at DESC);

-- 备份记录表
CREATE TABLE IF NOT EXISTS backups (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(512) NOT NULL,
    file_size BIGINT DEFAULT 0,
    backup_type VARCHAR(32) DEFAULT 'full',
    scope JSON,
    status VARCHAR(16) DEFAULT 'pending',
    error TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- AI 任务表
CREATE TABLE IF NOT EXISTS ai_tasks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(32) NOT NULL,
    status VARCHAR(16) DEFAULT 'pending',
    input_data JSON NOT NULL,
    output_data JSON,
    error TEXT,
    progress TINYINT DEFAULT 0,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 系统设置表
CREATE TABLE IF NOT EXISTS system_settings (
    id VARCHAR(36) PRIMARY KEY,
    key_name VARCHAR(128) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    value_type VARCHAR(16) DEFAULT 'string',
    description VARCHAR(255) DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 插入默认角色
INSERT OR IGNORE INTO roles (id, name, description, sort_order, status) VALUES
    ('role-admin', 'admin', '管理员', 0, 1),
    ('role-user', 'user', '普通用户', 1, 1),
    ('role-viewer', 'viewer', '访客', 2, 1);

-- 插入默认权限
INSERT OR IGNORE INTO permissions (id, name, resource, action, description) VALUES
    ('perm-event-create', 'event:create', 'event', 'create', '创建事件'),
    ('perm-event-read', 'event:read', 'event', 'read', '查看事件'),
    ('perm-event-update', 'event:update', 'event', 'update', '更新事件'),
    ('perm-event-delete', 'event:delete', 'event', 'delete', '删除事件'),
    ('perm-media-upload', 'media:upload', 'media', 'create', '上传媒体'),
    ('perm-media-read', 'media:read', 'media', 'read', '查看媒体'),
    ('perm-media-delete', 'media:delete', 'media', 'delete', '删除媒体'),
    ('perm-admin-users', 'admin:users', 'admin', 'manage', '用户管理'),
    ('perm-admin-config', 'admin:config', 'admin', 'config', '系统配置'),
    ('perm-admin-backup', 'admin:backup', 'admin', 'backup', '数据备份');

-- 管理员拥有所有权限
INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES
    ('role-admin', 'perm-event-create'),
    ('role-admin', 'perm-event-read'),
    ('role-admin', 'perm-event-update'),
    ('role-admin', 'perm-event-delete'),
    ('role-admin', 'perm-media-upload'),
    ('role-admin', 'perm-media-read'),
    ('role-admin', 'perm-media-delete'),
    ('role-admin', 'perm-admin-users'),
    ('role-admin', 'perm-admin-config'),
    ('role-admin', 'perm-admin-backup');

-- 普通用户拥有基本权限
INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES
    ('role-user', 'perm-event-create'),
    ('role-user', 'perm-event-read'),
    ('role-user', 'perm-event-update'),
    ('role-user', 'perm-event-delete'),
    ('role-user', 'perm-media-upload'),
    ('role-user', 'perm-media-read'),
    ('role-user', 'perm-media-delete');

-- 访客只有读取权限
INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES
    ('role-viewer', 'perm-event-read'),
    ('role-viewer', 'perm-media-read');
