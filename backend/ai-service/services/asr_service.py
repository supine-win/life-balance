"""
LifeRecorder AI 服务 - ASR 语音转文字服务
"""

from loguru import logger

from core.config import settings
from models.schemas import TranscribeResponse, TranscribeSegment


class ASRService:
    """语音识别服务"""

    def __init__(self):
        self._model = None

    def _load_model(self):
        """延迟加载 Whisper 模型"""
        if self._model is None:
            try:
                import whisper
                self._model = whisper.load_model(settings.ASR_MODEL, device=settings.ASR_DEVICE)
                logger.info(f"Whisper 模型加载完成: {settings.ASR_MODEL}")
            except ImportError:
                logger.warning("whisper 未安装，语音转文字不可用")
                raise
        return self._model

    async def transcribe(self, file_path: str, language: str = "zh") -> TranscribeResponse:
        """
        语音转文字

        Args:
            file_path: 音频文件路径
            language: 语言代码

        Returns:
            TranscribeResponse
        """
        try:
            model = self._load_model()
            result = model.transcribe(file_path, language=language)

            segments = [
                TranscribeSegment(
                    start=seg["start"],
                    end=seg["end"],
                    text=seg["text"].strip()
                )
                for seg in result.get("segments", [])
            ]

            return TranscribeResponse(
                task_id="",
                text=result["text"].strip(),
                segments=segments,
                language=language,
                duration=result.get("duration", 0.0)
            )

        except Exception as e:
            logger.error(f"语音转文字失败: {e}")
            return TranscribeResponse(
                task_id="",
                text="",
                segments=[],
                language=language,
                duration=0.0
            )


# 全局实例
asr_service = ASRService()
