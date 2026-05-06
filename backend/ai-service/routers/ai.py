"""
LifeRecorder AI 服务 - 路由定义
"""
from fastapi import APIRouter, File, Form, HTTPException, UploadFile

from models.schemas import (
    ChatRequest, ChatResponse, ExtractEventRequest, ExtractEventResponse,
    GenerateVideoRequest, GenerateVideoResponse, MetadataExtractionResult,
    TranscribeResponse
)
from services.asr_service import asr_service
from services.extract_service import extract_service
from services.metadata_service import MetadataExtractor

router = APIRouter()

# 元数据提取器实例
metadata_extractor = MetadataExtractor()


# ==================== 语音转文字 ====================

@router.post("/transcribe", response_model=TranscribeResponse)
async def transcribe_audio(
    file: UploadFile = File(...),
    language: str = Form("zh"),
):
    """语音转文字"""
    # 保存临时文件
    import tempfile
    import os
    suffix = os.path.splitext(file.filename or ".wav")[1]
    with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name

    try:
        result = await asr_service.transcribe(tmp_path, language)
        return result
    finally:
        os.unlink(tmp_path)


# ==================== 事件提取 ====================

@router.post("/extract-event", response_model=ExtractEventResponse)
async def extract_event(request: ExtractEventRequest):
    """从文本/对话中提取事件"""
    result = await extract_service.extract_from_text(
        text=request.text,
        conversation_history=request.conversation_history,
        language=request.language,
    )
    return result


# ==================== 视频生成 ====================

@router.post("/generate-video", response_model=GenerateVideoResponse)
async def generate_video(request: GenerateVideoRequest):
    """生成视频短片（异步任务）"""
    # TODO: 实现视频生成任务队列
    return GenerateVideoResponse(
        task_id="pending",
        status="processing",
        estimated_duration=120
    )


# ==================== 对话式录入 ====================

@router.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest):
    """对话式事件录入"""
    result = await extract_service.extract_from_text(
        text=request.message,
        language="zh",
    )
    return ChatResponse(
        type="extracted",
        content="",
        data={
            "extracted": result.extracted.model_dump(),
            "follow_up_questions": result.follow_up_questions,
        }
    )


# ==================== 媒体元数据提取 ====================

@router.post("/extract-metadata", response_model=MetadataExtractionResult)
async def extract_metadata(
    media_id: str = Form(...),
    file: UploadFile = File(...),
):
    """上传并提取媒体元数据"""
    import tempfile
    import os
    suffix = os.path.splitext(file.filename or ".jpg")[1]
    with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
        content = await file.read()
        tmp.write(content)
        tmp_path = tmp.name

    try:
        result = await metadata_extractor.extract(
            media_id=media_id,
            file_path=tmp_path,
            original_name=file.filename or "unknown",
        )
        return result
    finally:
        os.unlink(tmp_path)


@router.post("/extract-metadata-by-path", response_model=MetadataExtractionResult)
async def extract_metadata_by_path(request: dict):
    """根据文件路径提取元数据（内部调用）"""
    result = await metadata_extractor.extract(
        media_id=request.get("media_id", ""),
        file_path=request.get("file_path", ""),
        original_name=request.get("original_name", ""),
    )
    return result
