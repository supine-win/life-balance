"""
LifeRecorder AI 服务 - Pydantic 模型
"""
from datetime import datetime
from typing import Any, Dict, List, Optional

from pydantic import BaseModel, Field


# ==================== 通用 ====================

class TaskResponse(BaseModel):
    """AI 任务响应"""
    task_id: str
    status: str = "processing"
    progress: int = 0
    message: str = ""


# ==================== ASR 语音转文字 ====================

class TranscribeRequest(BaseModel):
    """语音转文字请求"""
    language: str = "zh"
    task: str = "transcribe"  # transcribe / translate


class TranscribeSegment(BaseModel):
    """转写片段"""
    start: float
    end: float
    text: str


class TranscribeResponse(BaseModel):
    """转写响应"""
    task_id: str
    text: str
    segments: List[TranscribeSegment] = []
    language: str = "zh"
    duration: float = 0.0


# ==================== 事件提取 ====================

class ChatMessage(BaseModel):
    """对话消息"""
    role: str = "user"
    content: str


class ExtractEventRequest(BaseModel):
    """事件提取请求"""
    text: str
    conversation_history: List[ChatMessage] = []
    language: str = "zh"


class ExtractedEvent(BaseModel):
    """提取的事件信息"""
    title: str = ""
    description: str = ""
    suggested_time: Optional[str] = None
    location: str = ""
    tags: List[str] = []
    participants: List[str] = []
    category: str = ""
    mood: str = ""
    confidence: float = 0.0


class ExtractEventResponse(BaseModel):
    """事件提取响应"""
    task_id: str
    extracted: ExtractedEvent
    follow_up_questions: List[str] = []


# ==================== 视频生成 ====================

class GenerateVideoRequest(BaseModel):
    """视频生成请求"""
    event_ids: List[str] = []
    media_ids: List[str] = []
    style: str = "cinematic"  # cinematic/documentary/slideshow
    duration: int = 30
    resolution: str = "1080p"  # 720p/1080p/4k
    config: Dict[str, Any] = {}


class GenerateVideoResponse(BaseModel):
    """视频生成响应"""
    task_id: str
    status: str = "processing"
    estimated_duration: int = 120  # 预估等待秒数


# ==================== 对话式录入 ====================

class ChatRequest(BaseModel):
    """对话请求"""
    message: str
    conversation_id: Optional[str] = None
    context: Dict[str, Any] = {}


class ChatResponse(BaseModel):
    """对话响应"""
    type: str = "text"  # text/extracted/done
    content: str = ""
    data: Optional[Dict[str, Any]] = None


# ==================== 媒体元数据 ====================

class MediaMetadataRequest(BaseModel):
    """媒体元数据提取请求"""
    media_id: str
    file_path: str
    original_name: str


class GPSCoordinate(BaseModel):
    """GPS 坐标"""
    latitude: float
    longitude: float
    altitude: Optional[float] = None


class EXIFData(BaseModel):
    """EXIF 数据"""
    datetime_original: Optional[str] = None
    camera_make: Optional[str] = None
    camera_model: Optional[str] = None
    lens_model: Optional[str] = None
    focal_length: Optional[float] = None
    aperture: Optional[float] = None
    shutter_speed: Optional[str] = None
    iso: Optional[int] = None
    gps: Optional[GPSCoordinate] = None
    width: Optional[int] = None
    height: Optional[int] = None


class VideoMetadata(BaseModel):
    """视频元数据"""
    creation_time: Optional[str] = None
    duration: Optional[float] = None
    width: Optional[int] = None
    height: Optional[int] = None
    video_codec: Optional[str] = None
    audio_codec: Optional[str] = None
    frame_rate: Optional[float] = None
    bit_rate: Optional[int] = None
    gps: Optional[GPSCoordinate] = None


class MetadataSuggestion(BaseModel):
    """元数据建议"""
    event_time: Optional[str] = None
    location: Optional[str] = None
    tags: List[str] = []
    category: str = ""
    hint_text: str = ""
    device_text: str = ""
    scene_text: str = ""
    confidence: Dict[str, float] = {}


class MetadataExtractionResult(BaseModel):
    """元数据提取结果"""
    media_id: str
    shot_time: Optional[datetime] = None
    shot_time_src: str = ""
    location_name: str = ""
    location_src: str = ""
    camera_make: str = ""
    camera_model: str = ""
    lens_model: str = ""
    focal_length: Optional[float] = None
    aperture: Optional[float] = None
    shutter_speed: str = ""
    iso: Optional[int] = None
    video_codec: str = ""
    audio_codec: str = ""
    frame_rate: Optional[float] = None
    bitrate: Optional[int] = None
    iptc_keywords: List[str] = []
    iptc_caption: str = ""
    filename_pattern: str = ""
    raw_exif: Dict[str, Any] = {}
    suggestions: MetadataSuggestion = MetadataSuggestion()
    extract_status: str = "completed"
    extract_error: str = ""
