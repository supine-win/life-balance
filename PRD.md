# LifeRecorder - 个人生活记录展示平台 PRD

## 项目概述

一个全功能的个人生活记录与展示平台，支持多模态事件录入、多媒体存储、AI 增强和展示模式。前后端分离架构，支持一键部署和容器化。

**代码仓库**：https://github.com/supine-win/life-balance

## 功能需求

### 1. 用户与权限系统
- 用户 CRUD（注册、登录、信息修改、注销）
- RBAC 权限管理（角色、权限、菜单关联）
- 后台管理界面参考 Vben Admin 风格
- JWT + Refresh Token 认证
- 支持多用户独立数据空间

### 2. 配置管理
- 配置中心，支持动态更新（无需重启）
- 分层配置：系统级 / 用户级
- 配置项包括：存储策略、展示风格、AI API Key、Webhook URL 等
- 配置变更审计日志

### 3. 事件录入系统（核心功能）
- **文字录入**：表单式/对话式描述事件
- **语音录入**：通过语音描述事件（ASR 转文字）
- **对话式**：与 AI 对话，AI 提取事件信息
- **多媒体附件**：支持上传图片、视频、音频
- **确认卡片**：录入后生成结构化预览卡片，用户确认后入库
- **事件属性**：时间、地点、标签、心情、参与人、分类

### 4. 多媒体存储系统
支持多种存储后端（可插拔适配器模式）：
- **本地存储**（默认）
- **云对象存储**：阿里云 OSS、华为云 OBS、腾讯云 COS、AWS S3、Azure Blob、Google Cloud Storage
- **云盘**：WebDAV、OneDrive、iCloud、百度网盘
- 存储策略可按用户/文件类型配置
- 文件去重、缩略图生成
- **元数据智能提取**：上传照片/视频时，自动读取 EXIF/IPTC/XMP（照片）和视频元数据，提取拍摄时间、GPS 位置、设备信息、人脸标签等关键字段，生成结构化提示卡片帮助用户回忆当时的场景和事件。用户可一键采纳或编辑后入库。

### 5. 展示模式
- 可配置展示参数：标题、副标题、布局样式
- 照片/视频的淡入淡出、轮播、网格等展示效果
- 背景音乐播放（MP3）
- **AI 视频生成**：对接豆包 Seedance 等视频大模型 API，生成视频短片并自动播放
- 样式对 AI Coding 友好：使用 CSS 变量 + 模板化，方便用户用 AI 替换样式
- 支持时间线视图、日历视图、地图视图、幻灯片模式

### 6. 数据库与基础设施
支持多种可插拔后端：
- **数据库**：PostgreSQL（默认）、MySQL、SQLite、MongoDB
- **缓存**：Redis（默认）、Memcached
- **消息队列**：RabbitMQ、Kafka（用于异步任务如视频生成、文件处理）
- ORM 支持数据库切换

### 7. Webhook 与 MCP 服务
- **Webhook**：事件创建/更新/删除时触发外部 Webhook
- 可配置触发条件和 payload 格式
- 重试机制和投递日志
- **MCP Server**：提供 Model Context Protocol 服务，支持 Agent 对接
  - 工具：create_event, query_events, update_event, delete_event, search_events, generate_slideshow

### 8. 数据备份与迁移
- 一键打包全部数据 + 资源文件
- 支持选择性备份（按时间范围、标签）
- 从备份恢复功能
- 导出格式：ZIP（数据库 dump + 媒体文件）
- 支持导入导出为 JSON/CSV

### 9. 部署与运维
- Dockerfile + docker-compose.yml
- 一键部署脚本（install.sh）
- 安装初始化页面（引导设置数据库、管理员账号等）
- 健康检查端点
- 日志管理

### 10. 开发规范
- 完善的 API 文档（OpenAPI/Swagger）
- 代码注解覆盖率 > 80%
- 单元测试 + 集成测试
- GitHub Actions CI/CD
- 自动发布 Release 版本（语义化版本号）

## 技术栈

### 前端
- **框架**：Vue 3 + TypeScript
- **构建**：Vite（如 Vite Plus 可用则使用）
- **UI 库**：Ant Design Vue（后台）/ 自定义组件（展示端）
- **状态管理**：Pinia
- **路由**：Vue Router 4
- **后台模板**：参考 Vben Admin（Vue 3 版）

### 后端
- **主服务**：Go (Gin/Fiber) — 高性能 API 网关和核心业务
- **AI/处理服务**：Python (FastAPI) — ASR、NLP、视频生成等 AI 能力
- **语言选择理由**：Go 处理高并发和存储适配器，Python 处理 AI 调用
- **ORM**：GORM (Go) / SQLAlchemy (Python)
- **认证**：JWT
- **配置**：Viper (Go) + Nacos/Consul（可选配置中心）

### 基础设施
- **容器**：Docker + Docker Compose
- **反向代理**：Nginx（内置配置）
- **CI/CD**：GitHub Actions

## 项目结构

```
life-recorder/
├── frontend/                # Vue 3 前端
│   ├── src/
│   │   ├── admin/          # 后台管理（Vben 风格）
│   │   ├── display/        # 展示模式
│   │   ├── event/          # 事件录入
│   │   └── shared/         # 共享组件
│   ├── Dockerfile
│   └── vite.config.ts
├── backend/
│   ├── gateway/            # Go 主服务
│   │   ├── internal/
│   │   │   ├── handler/    # HTTP handlers
│   │   │   ├── service/    # 业务逻辑
│   │   │   ├── repository/ # 数据访问
│   │   │   ├── middleware/  # 中间件
│   │   │   └── model/      # 数据模型
│   │   ├── cmd/            # 入口
│   │   └── Dockerfile
│   ├── ai-service/         # Python AI 服务
│   │   ├── routers/
│   │   ├── services/
│   │   └── Dockerfile
│   └── mcp-server/         # MCP Server
├── deploy/
│   ├── docker-compose.yml
│   ├── install.sh
│   └── nginx/
├── docs/
│   ├── api/               # API 文档
│   ├── deployment/        # 部署文档
│   └── development/       # 开发文档
├── .github/
│   └── workflows/         # CI/CD
└── README.md
```

## 数据模型

### Event（事件）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| user_id | UUID | 所属用户 |
| title | string | 事件标题 |
| description | text | 事件描述 |
| event_time | datetime | 事件发生时间 |
| location | string | 地点（可选） |
| mood | string | 心情标签 |
| tags | json | 标签数组 |
| participants | json | 参与人 |
| category | string | 分类 |
| media_ids | json | 关联媒体 ID |
| is_confirmed | bool | 是否已确认入库 |
| source | string | 录入来源（voice/text/chat） |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### Media（多媒体）
| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| user_id | UUID | 所属用户 |
| event_id | UUID | 关联事件 |
| type | string | image/video/audio |
| storage_adapter | string | 存储适配器 |
| storage_path | string | 存储路径 |
| thumbnail_path | string | 缩略图路径 |
| mime_type | string | MIME 类型 |
| size | bigint | 文件大小 |
| metadata | json | 文件元数据 |
| created_at | datetime | 创建时间 |

## API 端点概览

### 认证
- POST /api/v1/auth/register
- POST /api/v1/auth/login
- POST /api/v1/auth/refresh
- GET /api/v1/auth/me

### 事件
- GET /api/v1/events
- POST /api/v1/events
- GET /api/v1/events/:id
- PUT /api/v1/events/:id
- DELETE /api/v1/events/:id
- POST /api/v1/events/confirm/:id
- POST /api/v1/events/search

### 媒体
- POST /api/v1/media/upload
- POST /api/v1/media/upload-with-metadata （上传并返回 EXIF/元数据提取结果作为事件提示）
- GET /api/v1/media/:id
- DELETE /api/v1/media/:id
- GET /api/v1/media/:id/metadata （获取完整的元数据信息）

### AI 服务
- POST /api/v1/ai/transcribe (语音转文字)
- POST /api/v1/ai/extract-event (从对话提取事件)
- POST /api/v1/ai/generate-video (生成视频短片)

### 展示
- GET /api/v1/display/config
- PUT /api/v1/display/config
- GET /api/v1/display/slideshow

### 配置
- GET /api/v1/config
- PUT /api/v1/config
- GET /api/v1/config/:key

### 管理
- GET /api/v1/admin/users
- PUT /api/v1/admin/users/:id
- GET /api/v1/admin/roles
- GET /api/v1/admin/audit-logs

### 系统
- POST /api/v1/system/backup
- POST /api/v1/system/restore
- GET /api/v1/system/health

### Webhook
- POST /api/v1/webhooks
- GET /api/v1/webhooks
- DELETE /api/v1/webhooks/:id
- GET /api/v1/webhooks/:id/deliveries
