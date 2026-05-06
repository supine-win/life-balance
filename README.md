# LifeRecorder - 个人生活记录展示平台

> 📸 记录生活中的每一个精彩瞬间，用 AI 让回忆更生动

## 功能特性

### 🎯 核心功能
- **多模态事件录入**：文字、语音、对话式，选择你喜欢的方式
- **多媒体管理**：照片、视频、音频，支持多种云存储后端
- **智能元数据提取**：上传照片自动识别 EXIF、GPS、设备信息，生成事件建议
- **AI 增强**：语音转文字、事件信息提取、视频生成
- **丰富展示模式**：时间线、日历、地图、幻灯片，背景音乐

### 🔧 技术特性
- 前后端分离，Vue 3 + Go + Python 架构
- 支持 SQLite（零部署）/ PostgreSQL / MySQL
- 可插拔存储适配器（本地/S3/OSS/COS/WebDAV）
- JWT 认证 + RBAC 权限管理
- Docker 一键部署
- MCP Server 支持 Agent 对接
- Webhook 事件通知

## 快速开始

### Docker 一键部署（推荐）

```bash
git clone https://github.com/supine-win/life-balance.git
cd life-balance
chmod +x deploy/install.sh
./deploy/install.sh
```

启动后访问 `http://localhost`，按照安装向导完成初始化。

### 手动部署

#### 前置要求
- Docker & Docker Compose
- 或：Go 1.22+, Python 3.12+, Node.js 20+

#### 使用 Docker Compose

```bash
cd deploy
cp .env.example .env  # 编辑 .env 配置
docker compose up -d
```

#### 开发模式

```bash
# 启动 Go 网关
cd backend/gateway
go mod tidy
go run cmd/server/main.go

# 启动 Python AI 服务
cd backend/ai-service
pip install -r requirements.txt
uvicorn main:app --port 8081

# 启动前端开发服务器
cd frontend
npm install
npm run dev
```

## 项目结构

```
life-recorder/
├── frontend/              # Vue 3 前端
│   ├── src/
│   │   ├── views/        # 页面组件
│   │   ├── components/   # 通用组件
│   │   ├── stores/       # Pinia 状态管理
│   │   ├── api/          # API 客户端
│   │   └── router/       # 路由配置
│   └── Dockerfile
├── backend/
│   ├── gateway/          # Go 主服务 (Gin + GORM)
│   │   ├── cmd/server/   # 入口
│   │   ├── internal/     # 内部包
│   │   └── migrations/   # 数据库迁移
│   ├── ai-service/       # Python AI 服务 (FastAPI)
│   └── mcp-server/       # MCP Server (Go)
├── deploy/
│   ├── docker-compose.yml
│   ├── install.sh
│   └── nginx/
├── docs/
│   ├── design.md         # 系统设计文档
│   └── api/              # API 文档
└── .github/workflows/    # CI/CD
```

## 多媒体元数据智能提取

上传照片/视频时，系统自动：

1. **提取 EXIF 元数据**：拍摄时间、GPS 坐标、相机型号、镜头参数
2. **解析文件名日期**：识别 `IMG_20260501_143200.jpg` 等格式
3. **GPS 逆地理编码**：自动将坐标转换为地点名称
4. **生成建议卡片**：一键采纳时间、地点、标签建议
5. **视频元数据**：时长、分辨率、编码、关键帧提取

### 支持的文件格式
- 图片：JPG, PNG, GIF, WebP, HEIC/HEIF, TIFF
- 视频：MP4, MOV, AVI, MKV, WebM
- 音频：MP3, WAV, OGG, FLAC, AAC

## API 概览

| 模块 | 端点 | 说明 |
|------|------|------|
| 认证 | `POST /api/v1/auth/login` | 登录 |
| 认证 | `POST /api/v1/auth/register` | 注册 |
| 事件 | `GET /api/v1/events` | 事件列表 |
| 事件 | `POST /api/v1/events` | 创建事件 |
| 媒体 | `POST /api/v1/media/upload-with-metadata` | 上传并提取元数据 |
| 媒体 | `GET /api/v1/media/:id/metadata` | 获取元数据 |
| AI | `POST /api/v1/ai/transcribe` | 语音转文字 |
| AI | `POST /api/v1/ai/extract-event` | 事件提取 |
| 展示 | `GET /api/v1/display/slideshow` | 幻灯片 |
| 系统 | `GET /api/v1/system/health` | 健康检查 |

详细 API 文档参见 `docs/design.md`。

## 配置

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `LR_SERVER_PORT` | 服务端口 | 8080 |
| `LR_DATABASE_TYPE` | 数据库类型 | sqlite |
| `LR_JWT_SECRET` | JWT 密钥 | change-me |
| `LR_STORAGE_LOCAL_BASE_PATH` | 本地存储路径 | ./data/uploads |
| `GEOCODING_API` | 地理编码服务 | nominatim |
| `AMAP_API_KEY` | 高德地图 API Key | |
| `LLM_API_URL` | LLM API 地址 | |
| `LLM_API_KEY` | LLM API Key | |

## 技术栈

- **前端**：Vue 3 + TypeScript + Vite + Ant Design Vue + Pinia
- **后端**：Go (Gin + GORM) + Python (FastAPI)
- **数据库**：SQLite / PostgreSQL / MySQL
- **容器化**：Docker + Docker Compose
- **CI/CD**：GitHub Actions

## License

MIT
