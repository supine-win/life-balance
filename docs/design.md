# LifeRecorder 系统设计文档

## 1. 技术栈选择与理由

| 层级 | 技术 | 理由 |
|------|------|------|
| 前端框架 | Vue 3 + TypeScript | 响应式、组件化、类型安全 |
| 前端构建 | Vite | 极速 HMR、原生 ESM |
| UI 库 | Ant Design Vue | 企业级组件库，后台管理成熟 |
| 状态管理 | Pinia | Vue 3 官方推荐，轻量 |
| 路由 | Vue Router 4 | Vue 3 原生路由 |
| 后端主服务 | Go (Gin) | 高并发、低内存、编译型部署简单 |
| AI 服务 | Python (FastAPI) | AI/ML 生态丰富、异步支持好 |
| ORM (Go) | GORM | Go 最流行的 ORM，多数据库支持 |
| 认证 | JWT (RS256) | 无状态认证，支持微服务 |
| 配置管理 | Viper | Go 生态标准配置库 |
| 数据库 | SQLite (默认) / PostgreSQL / MySQL | SQLite 零部署，PostgreSQL 生产级 |
| 缓存 | 内存缓存 (默认) / Redis | 单机场景无需外部依赖 |
| 容器化 | Docker + Docker Compose | 标准化部署 |
| 反向代理 | Nginx | 静态资源、SSL、负载均衡 |
| MCP Server | Go (独立进程) | 与主服务共享模型定义 |

## 2. 数据模型详细设计

### 2.1 用户与权限

#### users（用户表）
```sql
CREATE TABLE users (
    id          VARCHAR(36) PRIMARY KEY,           -- UUID 主键
    username    VARCHAR(64) NOT NULL UNIQUE,       -- 用户名
    email       VARCHAR(255) NOT NULL UNIQUE,      -- 邮箱
    password    VARCHAR(255) NOT NULL,             -- 密码哈希 (bcrypt)
    nickname    VARCHAR(64) DEFAULT '',            -- 昵称
    avatar      VARCHAR(512) DEFAULT '',           -- 头像 URL
    status      TINYINT DEFAULT 1,                 -- 状态: 1=正常 0=禁用
    last_login  DATETIME,                          -- 最后登录时间
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_users_email (email),
    INDEX idx_users_status (status)
);
```

#### roles（角色表）
```sql
CREATE TABLE roles (
    id          VARCHAR(36) PRIMARY KEY,
    name        VARCHAR(64) NOT NULL UNIQUE,       -- 角色名: admin, user, viewer
    description VARCHAR(255) DEFAULT '',
    sort_order  INT DEFAULT 0,
    status      TINYINT DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

#### permissions（权限表）
```sql
CREATE TABLE permissions (
    id          VARCHAR(36) PRIMARY KEY,
    name        VARCHAR(128) NOT NULL UNIQUE,      -- 权限标识: event:create, event:delete
    resource    VARCHAR(64) NOT NULL,              -- 资源: event, media, config
    action      VARCHAR(32) NOT NULL,              -- 动作: create, read, update, delete
    description VARCHAR(255) DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_permissions_resource (resource)
);
```

#### role_permissions（角色-权限关联表）
```sql
CREATE TABLE role_permissions (
    role_id       VARCHAR(36) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
```

#### user_roles（用户-角色关联表）
```sql
CREATE TABLE user_roles (
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
```

#### refresh_tokens（刷新令牌表）
```sql
CREATE TABLE refresh_tokens (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,
    token       VARCHAR(512) NOT NULL UNIQUE,
    device_info VARCHAR(255) DEFAULT '',
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at  DATETIME,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_tokens_user (user_id),
    INDEX idx_tokens_expires (expires_at)
);
```

### 2.2 核心业务

#### events（事件表）
```sql
CREATE TABLE events (
    id           VARCHAR(36) PRIMARY KEY,
    user_id      VARCHAR(36) NOT NULL,              -- 所属用户
    title        VARCHAR(255) NOT NULL,             -- 事件标题
    description  TEXT,                              -- 事件描述
    event_time   DATETIME NOT NULL,                 -- 事件发生时间
    end_time     DATETIME,                          -- 结束时间（可选）
    location     VARCHAR(255) DEFAULT '',           -- 地点
    latitude     DECIMAL(10, 7),                    -- 纬度
    longitude    DECIMAL(10, 7),                    -- 经度
    mood         VARCHAR(32) DEFAULT '',            -- 心情标签
    tags         JSON,                              -- 标签数组 ["旅行", "美食"]
    participants JSON,                              -- 参与人 ["张三", "李四"]
    category     VARCHAR(64) DEFAULT '',            -- 分类
    color        VARCHAR(16) DEFAULT '',            -- 展示颜色
    priority     TINYINT DEFAULT 0,                 -- 优先级 0-5
    is_confirmed BOOLEAN DEFAULT FALSE,             -- 是否已确认入库
    source       VARCHAR(32) DEFAULT 'text',        -- 录入来源: voice/text/chat/import
    source_data  JSON,                              -- 原始录入数据（如对话记录）
    is_deleted   BOOLEAN DEFAULT FALSE,             -- 软删除标记
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_events_user_time (user_id, event_time DESC),
    INDEX idx_events_user_category (user_id, category),
    INDEX idx_events_user_tags (user_id, tags),
    INDEX idx_events_confirmed (user_id, is_confirmed),
    INDEX idx_events_created (user_id, created_at DESC)
);
```

#### media（多媒体表）
```sql
CREATE TABLE media (
    id               VARCHAR(36) PRIMARY KEY,
    user_id          VARCHAR(36) NOT NULL,
    event_id         VARCHAR(36),                   -- 关联事件（可选，未关联时为空）
    type             VARCHAR(16) NOT NULL,          -- image/video/audio/document
    storage_adapter  VARCHAR(32) NOT NULL DEFAULT 'local', -- 存储适配器
    storage_path     VARCHAR(512) NOT NULL,         -- 存储路径/Key
    original_name    VARCHAR(255) DEFAULT '',       -- 原始文件名
    thumbnail_path   VARCHAR(512) DEFAULT '',       -- 缩略图路径
    mime_type        VARCHAR(128) NOT NULL,         -- MIME 类型
    size             BIGINT NOT NULL DEFAULT 0,     -- 文件大小(字节)
    width            INT,                           -- 图片/视频宽度
    height           INT,                           -- 图片/视频高度
    duration         DECIMAL(10, 2),                -- 音频/视频时长(秒)
    file_hash        VARCHAR(64) DEFAULT '',        -- 文件哈希(SHA256)，用于去重
    metadata         JSON,                          -- 额外元数据(EXIF等)
    sort_order       INT DEFAULT 0,                 -- 排序
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE SET NULL,
    INDEX idx_media_user (user_id),
    INDEX idx_media_event (event_id),
    INDEX idx_media_type (user_id, type),
    INDEX idx_media_hash (file_hash)
);
```

#### tags（标签表）
```sql
CREATE TABLE tags (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    color       VARCHAR(16) DEFAULT '',
    usage_count INT DEFAULT 0,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_tags_user_name (user_id, name)
);
```

#### categories（分类表）
```sql
CREATE TABLE categories (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,
    name        VARCHAR(64) NOT NULL,
    icon        VARCHAR(64) DEFAULT '',
    color       VARCHAR(16) DEFAULT '',
    sort_order  INT DEFAULT 0,
    parent_id   VARCHAR(36),                       -- 父分类（支持多级）
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL,
    INDEX idx_categories_user (user_id)
);
```

### 2.3 配置系统

#### configs（配置表）
```sql
CREATE TABLE configs (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36),                       -- NULL 表示系统级配置
    scope       VARCHAR(16) NOT NULL DEFAULT 'system', -- system / user
    category    VARCHAR(64) NOT NULL,              -- 配置分类: storage, display, ai, webhook
    key         VARCHAR(128) NOT NULL,             -- 配置键
    value       TEXT NOT NULL,                     -- 配置值 (JSON string)
    value_type  VARCHAR(16) DEFAULT 'string',      -- 值类型: string/number/bool/json
    description VARCHAR(255) DEFAULT '',
    is_public   BOOLEAN DEFAULT FALSE,             -- 是否可公开读取（前端展示配置）
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE INDEX idx_configs_unique (user_id, category, key),
    INDEX idx_configs_category (category)
);
```

### 2.4 展示系统

#### display_templates（展示模板表）
```sql
CREATE TABLE display_templates (
    id           VARCHAR(36) PRIMARY KEY,
    user_id      VARCHAR(36) NOT NULL,
    name         VARCHAR(128) NOT NULL,
    type         VARCHAR(32) NOT NULL,             -- timeline/calendar/slideshow/grid/map
    config       JSON NOT NULL,                    -- 模板配置（CSS变量、布局参数等）
    is_default   BOOLEAN DEFAULT FALSE,
    is_builtin   BOOLEAN DEFAULT FALSE,            -- 内置模板不可删除
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_display_templates_user (user_id, type)
);
```

#### slideshows（幻灯片表）
```sql
CREATE TABLE slideshows (
    id            VARCHAR(36) PRIMARY KEY,
    user_id       VARCHAR(36) NOT NULL,
    title         VARCHAR(255) NOT NULL,
    description   TEXT,
    template_id   VARCHAR(36),                     -- 关联展示模板
    event_ids     JSON,                            -- 包含的事件ID列表
    media_ids     JSON,                            -- 包含的媒体ID列表
    bg_music_id   VARCHAR(36),                     -- 背景音乐媒体ID
    config        JSON,                            -- 幻灯片配置
    duration      INT DEFAULT 0,                   -- 总时长(秒)
    is_published  BOOLEAN DEFAULT FALSE,
    published_url VARCHAR(255) DEFAULT '',          -- 发布后的分享URL
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### 2.5 Webhook 系统

#### webhooks（Webhook 配置表）
```sql
CREATE TABLE webhooks (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,
    name        VARCHAR(128) NOT NULL,
    url         VARCHAR(512) NOT NULL,
    secret      VARCHAR(255) DEFAULT '',           -- 签名密钥
    events      JSON NOT NULL,                     -- 触发事件类型: ["event.created","event.updated"]
    headers     JSON,                              -- 自定义请求头
    is_active   BOOLEAN DEFAULT TRUE,
    retry_count TINYINT DEFAULT 3,                 -- 重试次数
    timeout_ms  INT DEFAULT 5000,                  -- 超时时间
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### webhook_deliveries（Webhook 投递日志）
```sql
CREATE TABLE webhook_deliveries (
    id           VARCHAR(36) PRIMARY KEY,
    webhook_id   VARCHAR(36) NOT NULL,
    event_type   VARCHAR(64) NOT NULL,
    payload      JSON NOT NULL,
    request_headers JSON,
    response_status INT,
    response_body    TEXT,
    duration_ms      INT,
    success      BOOLEAN DEFAULT FALSE,
    attempts     TINYINT DEFAULT 0,
    next_retry   DATETIME,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    INDEX idx_deliveries_webhook (webhook_id, created_at DESC),
    INDEX idx_deliveries_retry (success, next_retry)
);
```

### 2.6 系统管理

#### audit_logs（审计日志表）
```sql
CREATE TABLE audit_logs (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36),
    username    VARCHAR(64),
    action      VARCHAR(64) NOT NULL,              -- 操作: create/update/delete/login
    resource    VARCHAR(64) NOT NULL,              -- 资源类型
    resource_id VARCHAR(36),                       -- 资源ID
    detail      JSON,                              -- 变更详情
    ip_address  VARCHAR(45),                       -- IP地址
    user_agent  VARCHAR(512),                      -- User-Agent
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_audit_user (user_id, created_at DESC),
    INDEX idx_audit_resource (resource, resource_id),
    INDEX idx_audit_time (created_at DESC)
);
```

#### system_settings（系统设置表）
```sql
CREATE TABLE system_settings (
    id          VARCHAR(36) PRIMARY KEY,
    key         VARCHAR(128) NOT NULL UNIQUE,
    value       TEXT NOT NULL,
    value_type  VARCHAR(16) DEFAULT 'string',
    description VARCHAR(255) DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

#### backups（备份记录表）
```sql
CREATE TABLE backups (
    id           VARCHAR(36) PRIMARY KEY,
    user_id      VARCHAR(36),                      -- NULL=系统备份
    filename     VARCHAR(255) NOT NULL,
    file_path    VARCHAR(512) NOT NULL,
    file_size    BIGINT DEFAULT 0,
    backup_type  VARCHAR(32) DEFAULT 'full',       -- full/partial/export
    scope        JSON,                             -- 备份范围（时间、标签等过滤条件）
    status       VARCHAR(16) DEFAULT 'pending',    -- pending/processing/completed/failed
    error        TEXT,
    started_at   DATETIME,
    completed_at DATETIME,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_backups_user (user_id, created_at DESC)
);
```

### 2.7 AI 服务相关

#### ai_tasks（AI 任务表）
```sql
CREATE TABLE ai_tasks (
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36) NOT NULL,
    type        VARCHAR(32) NOT NULL,              -- transcribe/extract_event/generate_video
    status      VARCHAR(16) DEFAULT 'pending',     -- pending/processing/completed/failed
    input_data  JSON NOT NULL,                     -- 输入参数
    output_data JSON,                              -- 输出结果
    error       TEXT,
    progress    TINYINT DEFAULT 0,                 -- 进度 0-100
    started_at  DATETIME,
    completed_at DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_ai_tasks_user (user_id, status, created_at DESC)
);
```

## 3. API 端点完整定义

### 3.1 通用响应格式

```json
// 成功响应
{
    "code": 0,
    "message": "success",
    "data": { ... }
}

// 错误响应
{
    "code": 40001,
    "message": "参数错误: title 不能为空",
    "data": null
}

// 分页响应
{
    "code": 0,
    "message": "success",
    "data": {
        "items": [ ... ],
        "total": 100,
        "page": 1,
        "page_size": 20
    }
}
```

### 3.2 认证模块 `/api/v1/auth`

#### POST /api/v1/auth/register
注册新用户

请求:
```json
{
    "username": "string (3-64字符)",
    "email": "string (有效邮箱)",
    "password": "string (8-128字符)",
    "nickname": "string (可选)"
}
```
响应:
```json
{
    "code": 0,
    "data": {
        "user": { "id": "...", "username": "...", "email": "..." },
        "access_token": "string",
        "refresh_token": "string",
        "expires_in": 86400
    }
}
```

#### POST /api/v1/auth/login
登录

请求:
```json
{
    "username": "string (用户名或邮箱)",
    "password": "string"
}
```
响应: 同注册

#### POST /api/v1/auth/refresh
刷新令牌

请求:
```json
{
    "refresh_token": "string"
}
```
响应:
```json
{
    "code": 0,
    "data": {
        "access_token": "string",
        "refresh_token": "string",
        "expires_in": 86400
    }
}
```

#### POST /api/v1/auth/logout
登出（撤销刷新令牌）

请求:
```json
{
    "refresh_token": "string"
}
```

#### GET /api/v1/auth/me
获取当前用户信息

响应:
```json
{
    "code": 0,
    "data": {
        "id": "...", "username": "...", "email": "...",
        "nickname": "...", "avatar": "...", "roles": ["admin"],
        "permissions": ["event:create", "event:read", ...]
    }
}
```

#### PUT /api/v1/auth/profile
更新个人资料

请求:
```json
{
    "nickname": "string",
    "avatar": "string (URL)",
    "password": "string (新密码，可选)",
    "old_password": "string (修改密码时必填)"
}
```

### 3.3 事件模块 `/api/v1/events`

#### GET /api/v1/events
查询事件列表

查询参数:
- `page`: 页码 (默认 1)
- `page_size`: 每页数量 (默认 20, 最大 100)
- `start_time`: 开始时间 (RFC3339)
- `end_time`: 结束时间 (RFC3339)
- `category`: 分类筛选
- `tags`: 标签筛选 (逗号分隔)
- `mood`: 心情筛选
- `keyword`: 关键词搜索 (标题/描述)
- `confirmed`: 是否已确认 (bool)
- `sort`: 排序字段 (event_time/created_at)
- `order`: 排序方向 (asc/desc)

响应:
```json
{
    "code": 0,
    "data": {
        "items": [
            {
                "id": "...", "title": "...", "description": "...",
                "event_time": "2024-01-01T10:00:00Z",
                "location": "...", "mood": "happy",
                "tags": ["旅行"], "participants": ["张三"],
                "category": "生活", "color": "#FF5722",
                "media": [
                    { "id": "...", "type": "image", "thumbnail_path": "...", "storage_path": "..." }
                ],
                "is_confirmed": true, "source": "text",
                "created_at": "...", "updated_at": "..."
            }
        ],
        "total": 100, "page": 1, "page_size": 20
    }
}
```

#### POST /api/v1/events
创建事件

请求:
```json
{
    "title": "string (必填)",
    "description": "string",
    "event_time": "RFC3339 datetime",
    "end_time": "RFC3339 datetime (可选)",
    "location": "string",
    "latitude": 39.9042,
    "longitude": 116.4074,
    "mood": "string",
    "tags": ["string"],
    "participants": ["string"],
    "category": "string",
    "color": "string",
    "priority": 0,
    "media_ids": ["uuid"],
    "is_confirmed": false,
    "source": "text",
    "source_data": {}
}
```

#### GET /api/v1/events/:id
获取单个事件详情

#### PUT /api/v1/events/:id
更新事件（请求体同创建）

#### DELETE /api/v1/events/:id
软删除事件

#### POST /api/v1/events/confirm/:id
确认事件入库

请求:
```json
{
    "modifications": {
        "title": "修改后的标题 (可选)",
        "description": "修改后的描述 (可选)"
    }
}
```

#### POST /api/v1/events/search
高级搜索

请求:
```json
{
    "keyword": "string",
    "filters": {
        "start_time": "RFC3339",
        "end_time": "RFC3339",
        "categories": ["string"],
        "tags": ["string"],
        "moods": ["string"],
        "has_media": true,
        "media_types": ["image", "video"]
    },
    "sort": { "field": "event_time", "order": "desc" },
    "page": 1,
    "page_size": 20
}
```

#### GET /api/v1/events/calendar/:year/:month
日历视图数据

响应:
```json
{
    "code": 0,
    "data": {
        "year": 2024, "month": 1,
        "days": {
            "1": [{ "id": "...", "title": "...", "color": "..." }],
            "15": [{ "id": "...", "title": "...", "color": "..." }]
        }
    }
}
```

### 3.4 媒体模块 `/api/v1/media`

#### POST /api/v1/media/upload
上传媒体文件 (multipart/form-data)

请求:
- `file`: 文件 (支持 image/jpeg, image/png, image/gif, image/webp, video/mp4, audio/mpeg, audio/wav, audio/ogg)
- `event_id`: 关联事件ID (可选)
- `type`: 文件类型 (可选, 自动检测)

响应:
```json
{
    "code": 0,
    "data": {
        "id": "...", "type": "image", "mime_type": "image/jpeg",
        "size": 1024000, "width": 1920, "height": 1080,
        "thumbnail_path": "/media/thumbs/xxx.jpg",
        "storage_path": "/media/xxx.jpg"
    }
}
```

#### POST /api/v1/media/batch-upload
批量上传 (multipart/form-data, 多个 file)

#### GET /api/v1/media/:id
获取媒体信息

#### GET /api/v1/media/:id/file
获取/下载媒体文件（支持 Range 请求）

#### GET /api/v1/media/:id/thumbnail
获取缩略图

#### DELETE /api/v1/media/:id
删除媒体文件（同时删除存储文件）

### 3.5 AI 服务模块 `/api/v1/ai`

#### POST /api/v1/ai/transcribe
语音转文字

请求 (multipart/form-data):
- `file`: 音频文件
- `language`: 语言代码 (可选, 默认 zh)

响应:
```json
{
    "code": 0,
    "data": {
        "task_id": "...", "text": "转写结果", "segments": [
            { "start": 0.0, "end": 2.5, "text": "..." }
        ]
    }
}
```

#### POST /api/v1/ai/extract-event
从对话/文本提取事件

请求:
```json
{
    "text": "string (对话文本)",
    "conversation_history": [
        { "role": "user", "content": "..." },
        { "role": "assistant", "content": "..." }
    ],
    "language": "zh"
}
```

响应:
```json
{
    "code": 0,
    "data": {
        "task_id": "...",
        "extracted": {
            "title": "去北京旅行",
            "description": "和朋友张三一起去了故宫和长城",
            "suggested_time": "2024-05-01T09:00:00Z",
            "location": "北京",
            "tags": ["旅行"],
            "participants": ["张三"],
            "category": "旅行",
            "mood": "happy",
            "confidence": 0.92
        },
        "follow_up_questions": ["具体去了哪些景点？"]
    }
}
```

#### POST /api/v1/ai/generate-video
生成视频短片

请求:
```json
{
    "event_ids": ["uuid"],
    "media_ids": ["uuid"],
    "style": "cinematic",
    "duration": 30,
    "resolution": "1080p",
    "config": {}
}
```

响应:
```json
{
    "code": 0,
    "data": {
        "task_id": "...", "status": "processing",
        "estimated_duration": 120
    }
}
```

#### GET /api/v1/ai/tasks/:id
查询 AI 任务状态

#### POST /api/v1/ai/chat
对话式录入（流式 SSE）

请求:
```json
{
    "message": "string",
    "conversation_id": "string (可选, 用于连续对话)",
    "context": { "current_event": { ... } }
}
```

响应: SSE 流
```
data: {"type": "text", "content": "我理解了，你"}
data: {"type": "text", "content": "去了北京旅行"}
data: {"type": "extracted", "data": { "title": "...", ... }}
data: {"type": "done"}
```

### 3.6 展示模块 `/api/v1/display`

#### GET /api/v1/display/config
获取展示配置

#### PUT /api/v1/display/config
更新展示配置

#### GET /api/v1/display/templates
获取展示模板列表

#### POST /api/v1/display/templates
创建展示模板

#### PUT /api/v1/display/templates/:id
更新展示模板

#### DELETE /api/v1/display/templates/:id
删除展示模板

#### GET /api/v1/display/slideshow
获取幻灯片列表

#### POST /api/v1/display/slideshow
创建幻灯片

#### GET /api/v1/display/slideshow/:id
获取幻灯片详情

#### PUT /api/v1/display/slideshow/:id
更新幻灯片

#### DELETE /api/v1/display/slideshow/:id
删除幻灯片

#### POST /api/v1/display/slideshow/:id/publish
发布幻灯片（生成分享链接）

### 3.7 配置模块 `/api/v1/config`

#### GET /api/v1/config
获取所有配置（按分类分组）

查询参数:
- `scope`: system / user
- `category`: 配置分类

#### PUT /api/v1/config
批量更新配置

请求:
```json
{
    "configs": [
        { "category": "storage", "key": "default_adapter", "value": "local" },
        { "category": "ai", "key": "api_key", "value": "sk-xxx" }
    ]
}
```

#### GET /api/v1/config/:category/:key
获取单个配置项

### 3.8 管理模块 `/api/v1/admin` (需要 admin 角色)

#### 用户管理
- GET /api/v1/admin/users — 用户列表
- GET /api/v1/admin/users/:id — 用户详情
- PUT /api/v1/admin/users/:id — 更新用户
- DELETE /api/v1/admin/users/:id — 删除用户
- PUT /api/v1/admin/users/:id/status — 启用/禁用用户
- PUT /api/v1/admin/users/:id/roles — 分配角色

#### 角色管理
- GET /api/v1/admin/roles — 角色列表
- POST /api/v1/admin/roles — 创建角色
- PUT /api/v1/admin/roles/:id — 更新角色
- DELETE /api/v1/admin/roles/:id — 删除角色
- PUT /api/v1/admin/roles/:id/permissions — 分配权限

#### 权限管理
- GET /api/v1/admin/permissions — 权限列表

#### 系统管理
- GET /api/v1/admin/audit-logs — 审计日志
- GET /api/v1/admin/system/info — 系统信息
- GET /api/v1/admin/system/stats — 系统统计

### 3.9 Webhook 模块 `/api/v1/webhooks`

#### POST /api/v1/webhooks
创建 Webhook

请求:
```json
{
    "name": "string",
    "url": "https://example.com/webhook",
    "secret": "string (可选)",
    "events": ["event.created", "event.updated"],
    "headers": { "X-Custom": "value" },
    "retry_count": 3,
    "timeout_ms": 5000
}
```

#### GET /api/v1/webhooks
Webhook 列表

#### PUT /api/v1/webhooks/:id
更新 Webhook

#### DELETE /api/v1/webhooks/:id
删除 Webhook

#### POST /api/v1/webhooks/:id/test
测试 Webhook

#### GET /api/v1/webhooks/:id/deliveries
投递日志

### 3.10 系统模块 `/api/v1/system`

#### GET /api/v1/system/health
健康检查

响应:
```json
{
    "status": "healthy",
    "version": "1.0.0",
    "uptime": 86400,
    "components": {
        "database": "ok",
        "storage": "ok",
        "ai_service": "ok"
    }
}
```

#### POST /api/v1/system/backup
创建备份

请求:
```json
{
    "type": "full",
    "scope": {
        "start_time": "...",
        "end_time": "...",
        "tags": ["..."],
        "include_media": true
    }
}
```

#### GET /api/v1/system/backups
备份列表

#### POST /api/v1/system/restore
从备份恢复

请求:
```json
{
    "backup_id": "uuid",
    "options": {
        "overwrite": false,
        "include_media": true
    }
}
```

#### GET /api/v1/system/backups/:id/download
下载备份文件

### 3.11 安装初始化 `/api/v1/setup`

#### GET /api/v1/setup/status
检查是否已初始化

#### POST /api/v1/setup/init
执行初始化

请求:
```json
{
    "admin": {
        "username": "admin",
        "email": "admin@example.com",
        "password": "admin123"
    },
    "database": {
        "type": "sqlite",
        "path": "./data/life-recorder.db"
    },
    "storage": {
        "adapter": "local",
        "path": "./data/uploads"
    }
}
```

## 4. 前端设计

### 4.1 路由结构

```
/                          → 重定向到 /setup 或 /events
/setup                     → 安装初始化页 (仅首次)
/login                     → 登录页
/register                  → 注册页

--- 以下需要认证 ---
/events                    → 事件列表（时间线视图）
/events/calendar           → 日历视图
/events/map                → 地图视图
/events/:id                → 事件详情

/events/create             → 创建事件
/events/create/voice       → 语音录入
/events/create/chat        → 对话式录入

/display                   → 展示模式入口
/display/slideshow/:id     → 幻灯片播放
/display/templates         → 模板管理

/media                     → 媒体库
/settings                  → 个人设置
/webhooks                  → Webhook 管理

--- 后台管理 (需要 admin 角色) ---
/admin                     → 管理后台首页
/admin/users               → 用户管理
/admin/roles               → 角色管理
/admin/permissions         → 权限管理
/admin/config              → 系统配置
/admin/audit-logs          → 审计日志
/admin/system              → 系统监控
/admin/backup              → 数据备份
```

### 4.2 组件树

```
App.vue
├── layouts/
│   ├── MainLayout.vue              # 主布局（侧边栏 + 顶栏 + 内容区）
│   ├── AdminLayout.vue             # 管理后台布局
│   ├── DisplayLayout.vue           # 展示模式布局（全屏，无导航）
│   └── SetupLayout.vue             # 安装向导布局
│
├── views/
│   ├── setup/
│   │   └── SetupWizard.vue         # 安装初始化向导
│   │
│   ├── auth/
│   │   ├── Login.vue
│   │   └── Register.vue
│   │
│   ├── events/
│   │   ├── EventList.vue           # 事件列表页
│   │   ├── EventTimeline.vue       # 时间线视图
│   │   ├── EventCalendar.vue       # 日历视图
│   │   ├── EventMap.vue            # 地图视图
│   │   ├── EventDetail.vue         # 事件详情
│   │   ├── EventForm.vue           # 事件表单（创建/编辑）
│   │   ├── VoiceInput.vue          # 语音录入
│   │   ├── ChatInput.vue           # 对话式录入
│   │   └── EventConfirmCard.vue    # 确认卡片
│   │
│   ├── display/
│   │   ├── DisplayHome.vue         # 展示首页
│   │   ├── SlideshowPlayer.vue     # 幻灯片播放器
│   │   ├── TemplateEditor.vue      # 模板/样式编辑器
│   │   ├── TemplateList.vue        # 模板列表
│   │   ├── SlideshowEditor.vue     # 幻灯片编辑器
│   │   └── BackgroundMusic.vue     # 背景音乐组件
│   │
│   ├── media/
│   │   ├── MediaLibrary.vue        # 媒体库
│   │   └── MediaUpload.vue         # 上传组件
│   │
│   ├── webhooks/
│   │   ├── WebhookList.vue
│   │   └── WebhookForm.vue
│   │
│   ├── admin/
│   │   ├── Dashboard.vue           # 管理后台首页
│   │   ├── UserList.vue
│   │   ├── UserForm.vue
│   │   ├── RoleList.vue
│   │   ├── RoleForm.vue
│   │   ├── PermissionList.vue
│   │   ├── ConfigEditor.vue        # 配置编辑器
│   │   ├── AuditLogList.vue
│   │   ├── SystemMonitor.vue
│   │   └── BackupManager.vue
│   │
│   └── settings/
│       ├── ProfileSettings.vue
│       ├── DisplaySettings.vue
│       └── StorageSettings.vue
│
├── components/
│   ├── common/
│   │   ├── AppHeader.vue
│   │   ├── AppSidebar.vue
│   │   ├── AppBreadcrumb.vue
│   │   ├── Pagination.vue
│   │   ├── EmptyState.vue
│   │   ├── LoadingSpinner.vue
│   │   └── ConfirmDialog.vue
│   │
│   ├── event/
│   │   ├── EventCard.vue           # 事件卡片
│   │   ├── EventFilter.vue         # 筛选器
│   │   ├── TagInput.vue            # 标签输入
│   │   ├── DateTimePicker.vue      # 日期时间选择
│   │   ├── MoodSelector.vue        # 心情选择器
│   │   ├── LocationPicker.vue      # 地点选择
│   │   ├── MediaGallery.vue        # 媒体画廊
│   │   └── ParticipantInput.vue    # 参与人输入
│   │
│   ├── display/
│   │   ├── TransitionEffect.vue    # 过渡效果
│   │   ├── MediaSlide.vue          # 媒体幻灯片
│   │   ├── CssVarEditor.vue        # CSS 变量编辑器
│   │   └── StylePreview.vue        # 样式预览
│   │
│   └── ai/
│       ├── ChatBubble.vue          # 对话气泡
│       ├── ChatInput.vue           # 对话输入
│       ├── VoiceRecorder.vue       # 录音组件
│       └── AiStatus.vue            # AI 任务状态
```

### 4.3 状态管理 (Pinia Stores)

```
stores/
├── auth.ts         # 用户认证状态、token 管理
├── event.ts        # 事件数据、筛选条件
├── media.ts        # 媒体库数据
├── display.ts      # 展示配置、模板
├── config.ts       # 系统配置
├── admin.ts        # 管理后台状态
├── chat.ts         # 对话式录入状态
└── app.ts          # 全局应用状态（侧边栏、主题等）
```

### 4.4 CSS 变量体系（AI Coding 友好）

```css
:root {
    /* 全局 */
    --lr-primary: #1890ff;
    --lr-secondary: #722ed1;
    --lr-bg: #ffffff;
    --lr-text: #333333;
    --lr-border-radius: 8px;
    
    /* 时间线 */
    --lr-timeline-line-color: #e8e8e8;
    --lr-timeline-dot-size: 12px;
    --lr-timeline-card-shadow: 0 2px 8px rgba(0,0,0,0.1);
    
    /* 幻灯片 */
    --lr-slide-transition: fade;       /* fade / slide / zoom / flip */
    --lr-slide-duration: 5s;
    --lr-slide-bg: #000000;
    --lr-slide-text-color: #ffffff;
    --lr-slide-font-size: 24px;
    --lr-slide-caption-bg: rgba(0,0,0,0.6);
    
    /* 日历 */
    --lr-calendar-cell-size: 80px;
    --lr-calendar-today-color: #1890ff;
    --lr-calendar-event-dot-size: 6px;
    
    /* 卡片 */
    --lr-card-padding: 16px;
    --lr-card-gap: 16px;
    --lr-card-hover-scale: 1.02;
}
```

## 5. 项目目录结构

```
life-recorder/
├── README.md
├── PRD.md
├── docs/
│   ├── design.md
│   ├── api/
│   │   └── openapi.yaml
│   ├── deployment/
│   │   └── guide.md
│   └── development/
│       └── guide.md
│
├── frontend/                         # Vue 3 前端
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── env.d.ts
│   ├── index.html
│   ├── .env
│   ├── .env.production
│   ├── public/
│   │   └── favicon.ico
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── router/
│   │   │   └── index.ts
│   │   ├── stores/
│   │   │   ├── auth.ts
│   │   │   ├── event.ts
│   │   │   ├── media.ts
│   │   │   ├── display.ts
│   │   │   ├── config.ts
│   │   │   ├── admin.ts
│   │   │   ├── chat.ts
│   │   │   └── app.ts
│   │   ├── api/
│   │   │   ├── client.ts            # axios 实例
│   │   │   ├── auth.ts
│   │   │   ├── events.ts
│   │   │   ├── media.ts
│   │   │   ├── ai.ts
│   │   │   ├── display.ts
│   │   │   ├── config.ts
│   │   │   ├── webhooks.ts
│   │   │   ├── admin.ts
│   │   │   └── system.ts
│   │   ├── layouts/
│   │   │   ├── MainLayout.vue
│   │   │   ├── AdminLayout.vue
│   │   │   ├── DisplayLayout.vue
│   │   │   └── SetupLayout.vue
│   │   ├── views/
│   │   │   ├── setup/
│   │   │   ├── auth/
│   │   │   ├── events/
│   │   │   ├── display/
│   │   │   ├── media/
│   │   │   ├── webhooks/
│   │   │   ├── admin/
│   │   │   └── settings/
│   │   ├── components/
│   │   │   ├── common/
│   │   │   ├── event/
│   │   │   ├── display/
│   │   │   └── ai/
│   │   ├── composables/
│   │   │   ├── useAuth.ts
│   │   │   ├── useEvent.ts
│   │   │   ├── useMedia.ts
│   │   │   ├── useAudioRecorder.ts
│   │   │   └── useSSE.ts
│   │   ├── utils/
│   │   │   ├── date.ts
│   │   │   ├── format.ts
│   │   │   └── storage.ts
│   │   └── styles/
│   │       ├── variables.css
│   │       ├── global.css
│   │       └── transitions.css
│   └── Dockerfile
│
├── backend/
│   ├── gateway/                      # Go 主服务
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go           # 入口
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   │   └── config.go         # 配置加载
│   │   │   ├── model/
│   │   │   │   ├── user.go
│   │   │   │   ├── event.go
│   │   │   │   ├── media.go
│   │   │   │   ├── config.go
│   │   │   │   ├── webhook.go
│   │   │   │   ├── display.go
│   │   │   │   ├── backup.go
│   │   │   │   └── audit.go
│   │   │   ├── repository/
│   │   │   │   ├── user_repo.go
│   │   │   │   ├── event_repo.go
│   │   │   │   ├── media_repo.go
│   │   │   │   ├── config_repo.go
│   │   │   │   ├── webhook_repo.go
│   │   │   │   └── display_repo.go
│   │   │   ├── service/
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── event_service.go
│   │   │   │   ├── media_service.go
│   │   │   │   ├── config_service.go
│   │   │   │   ├── webhook_service.go
│   │   │   │   ├── display_service.go
│   │   │   │   ├── backup_service.go
│   │   │   │   ├── storage/
│   │   │   │   │   ├── adapter.go       # 存储适配器接口
│   │   │   │   │   ├── local.go         # 本地存储
│   │   │   │   │   ├── s3.go            # S3 兼容存储
│   │   │   │   │   ├── oss.go           # 阿里云 OSS
│   │   │   │   │   ├── obs.go           # 华为云 OBS
│   │   │   │   │   ├── cos.go           # 腾讯云 COS
│   │   │   │   │   └── webdav.go        # WebDAV
│   │   │   │   └── cache/
│   │   │   │       ├── cache.go         # 缓存接口
│   │   │   │       ├── memory.go        # 内存缓存
│   │   │   │       └── redis.go         # Redis 缓存
│   │   │   ├── handler/
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── event_handler.go
│   │   │   │   ├── media_handler.go
│   │   │   │   ├── ai_handler.go
│   │   │   │   ├── config_handler.go
│   │   │   │   ├── webhook_handler.go
│   │   │   │   ├── display_handler.go
│   │   │   │   ├── admin_handler.go
│   │   │   │   ├── system_handler.go
│   │   │   │   ├── setup_handler.go
│   │   │   │   └── middleware/
│   │   │   │       ├── auth.go          # JWT 认证中间件
│   │   │   │       ├── rbac.go          # RBAC 权限中间件
│   │   │   │       ├── cors.go          # CORS 中间件
│   │   │   │       ├── logger.go        # 请求日志
│   │   │   │       ├── recovery.go      # panic 恢复
│   │   │   │       └── rate_limit.go    # 限流
│   │   │   └── pkg/
│   │   │       ├── jwt/
│   │   │       │   └── jwt.go
│   │   │       ├── hash/
│   │   │       │   └── hash.go
│   │   │       ├── uuid/
│   │   │       │   └── uuid.go
│   │   │       ├── logger/
│   │   │       │   └── logger.go
│   │   │       ├── response/
│   │   │       │   └── response.go
│   │   │       ├── thumbnail/
│   │   │       │   └── thumbnail.go
│   │   │       └── validator/
│   │   │           └── validator.go
│   │   ├── migrations/
│   │   │   ├── 001_init.up.sql
│   │   │   └── 001_init.down.sql
│   │   ├── configs/
│   │   │   └── config.yaml
│   │   └── Dockerfile
│   │
│   ├── ai-service/                   # Python AI 服务
│   │   ├── requirements.txt
│   │   ├── main.py
│   │   ├── core/
│   │   │   ├── config.py
│   │   │   └── logger.py
│   │   ├── routers/
│   │   │   ├── transcribe.py
│   │   │   ├── extract.py
│   │   │   ├── video.py
│   │   │   └── chat.py
│   │   ├── services/
│   │   │   ├── asr_service.py        # Whisper / FunASR
│   │   │   ├── extract_service.py    # 事件提取 (LLM)
│   │   │   ├── video_service.py      # 视频生成 (Seedance API)
│   │   │   └── chat_service.py       # 对话服务
│   │   ├── models/
│   │   │   └── schemas.py            # Pydantic 模型
│   │   └── Dockerfile
│   │
│   └── mcp-server/                   # MCP Server
│       ├── go.mod
│       ├── main.go
│       ├── server.go
│       └── tools/
│           ├── create_event.go
│           ├── query_events.go
│           ├── update_event.go
│           ├── delete_event.go
│           ├── search_events.go
│           └── generate_slideshow.go
│
├── deploy/
│   ├── docker-compose.yml
│   ├── install.sh
│   └── nginx/
│       └── default.conf
│
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml
│
└── scripts/
    └── seed.go                       # 初始化种子数据
```

## 6. 存储适配器模式设计

### 6.1 适配器接口

```go
// StorageAdapter 存储适配器接口
type StorageAdapter interface {
    // 保存文件
    Save(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
    // 获取文件
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    // 删除文件
    Delete(ctx context.Context, key string) error
    // 获取公开URL（如果支持）
    GetURL(ctx context.Context, key string) (string, error)
    // 检查文件是否存在
    Exists(ctx context.Context, key string) (bool, error)
    // 获取文件大小
    GetSize(ctx context.Context, key string) (int64, error)
}
```

### 6.2 适配器注册

```go
// 通过工厂模式注册适配器
var adapterFactories = map[string]func(cfg map[string]interface{}) (StorageAdapter, error){
    "local":   NewLocalStorage,
    "s3":      NewS3Storage,
    "oss":     NewOSSStorage,
    "obs":     NewOBSStorage,
    "cos":     NewCOSStorage,
    "webdav":  NewWebDAVStorage,
}
```

## 7. 认证流程

1. 用户登录 → 验证密码 → 生成 access_token (1h) + refresh_token (7d)
2. 请求携带 `Authorization: Bearer <access_token>`
3. access_token 过期 → 前端自动调用 /auth/refresh → 获取新 token
4. 登出 → 服务端撤销 refresh_token
5. JWT Claims 包含: user_id, username, roles, permissions

## 8. Webhook 投递流程

1. 事件发生 (创建/更新/删除)
2. 查询匹配的活跃 Webhook
3. 异步投递（使用 goroutine）
4. 构造 payload + HMAC 签名
5. 发送 HTTP POST
6. 记录投递日志
7. 失败则按指数退避重试

## 9. AI 服务调用流程

1. 前端发起 AI 请求 → Go 网关转发到 Python AI 服务
2. AI 服务创建 ai_tasks 记录
3. 异步处理（ASR/提取/视频生成）
4. 前端轮询 /ai/tasks/:id 或通过 SSE 接收进度
5. 完成后更新 ai_tasks 状态和结果

## 10. 部署架构

```
                    ┌─────────────┐
                    │   Nginx     │
                    │  (反向代理)  │
                    └──────┬──────┘
                           │
               ┌───────────┼───────────┐
               │           │           │
        ┌──────▼──────┐   │   ┌───────▼──────┐
        │  Frontend   │   │   │  Go Gateway  │
        │  (静态资源)  │   │   │  (API 服务)   │
        └─────────────┘   │   └───────┬──────┘
                          │           │
                          │   ┌───────▼──────┐
                          │   │ Python AI    │
                          │   │ (FastAPI)    │
                          │   └──────────────┘
                          │
                   ┌──────▼──────┐
                   │   SQLite /  │
                   │ PostgreSQL  │
                   └─────────────┘
```

所有服务运行在单个 Docker Compose 中，Nginx 统一入口。
单机部署时 SQLite 零外部依赖，生产环境推荐 PostgreSQL。
