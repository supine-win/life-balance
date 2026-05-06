"""
LifeRecorder AI 服务 - 主入口

提供 AI 能力：
- 语音转文字 (ASR)
- 事件提取 (NLP)
- 媒体元数据提取 (EXIF/视频)
- 视频生成
- 对话式录入
"""
import os

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger

from core.config import settings
from routers.ai import router as ai_router

# 创建 FastAPI 应用
app = FastAPI(
    title=settings.APP_NAME,
    version=settings.VERSION,
    description="LifeRecorder AI 服务 - 提供语音转文字、事件提取、媒体元数据提取等 AI 能力",
)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 注册路由
app.include_router(ai_router, prefix="/api/v1/ai", tags=["AI"])


@app.on_event("startup")
async def startup():
    """启动事件"""
    logger.info(f"{settings.APP_NAME} v{settings.VERSION} 启动中...")
    os.makedirs(settings.UPLOAD_DIR, exist_ok=True)
    logger.info("AI 服务就绪")


@app.on_event("shutdown")
async def shutdown():
    """关闭事件"""
    logger.info("AI 服务关闭")


@app.get("/health")
async def health():
    """健康检查"""
    return {"status": "healthy", "version": settings.VERSION}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=settings.DEBUG,
    )
