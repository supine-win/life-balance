"""
LifeRecorder AI 服务 - 核心配置
"""
import os
from typing import Optional


class Settings:
    """应用配置"""

    # 服务配置
    APP_NAME: str = "LifeRecorder AI Service"
    VERSION: str = "1.0.0"
    DEBUG: bool = os.getenv("DEBUG", "false").lower() == "true"
    HOST: str = os.getenv("AI_HOST", "0.0.0.0")
    PORT: int = int(os.getenv("AI_PORT", "8081"))

    # 网关连接
    GATEWAY_URL: str = os.getenv("GATEWAY_URL", "http://localhost:8080")

    # 文件上传临时目录
    UPLOAD_DIR: str = os.getenv("UPLOAD_DIR", "/tmp/life-recorder/uploads")
    MAX_UPLOAD_SIZE: int = 100 * 1024 * 1024  # 100MB

    # ASR 配置
    ASR_MODEL: str = os.getenv("ASR_MODEL", "base")  # tiny/base/small/medium/large
    ASR_DEVICE: str = os.getenv("ASR_DEVICE", "cpu")

    # LLM 配置（事件提取）
    LLM_API_URL: str = os.getenv("LLM_API_URL", "")
    LLM_API_KEY: str = os.getenv("LLM_API_KEY", "")
    LLM_MODEL: str = os.getenv("LLM_MODEL", "gpt-4o-mini")

    # 视频生成 API
    VIDEO_API_URL: str = os.getenv("VIDEO_API_URL", "")
    VIDEO_API_KEY: str = os.getenv("VIDEO_API_KEY", "")

    # GPS 逆地理编码
    GEOCODING_API: str = os.getenv("GEOCODING_API", "nominatim")  # nominatim/amap/baidu
    AMAP_API_KEY: str = os.getenv("AMAP_API_KEY", "")
    BAIDU_API_KEY: str = os.getenv("BAIDU_API_KEY", "")

    # ffmpeg 路径
    FFMPEG_PATH: str = os.getenv("FFMPEG_PATH", "ffmpeg")
    FFPROBE_PATH: str = os.getenv("FFPROBE_PATH", "ffprobe")

    # Redis（可选，用于任务队列）
    REDIS_URL: Optional[str] = os.getenv("REDIS_URL")


settings = Settings()
