"""
LifeRecorder AI 服务 - 事件提取服务
"""
import json
from typing import List

import httpx
from loguru import logger

from core.config import settings
from models.schemas import ChatMessage, ExtractEventResponse, ExtractedEvent


# 事件提取 System Prompt
EXTRACT_SYSTEM_PROMPT = """你是一个生活事件提取助手。用户会描述他们生活中的事件，你需要从中提取结构化的事件信息。

请从用户的描述中提取以下信息（如果提到的话）：
- title: 事件标题（简洁概括）
- description: 事件描述（详细）
- suggested_time: 建议的时间（如果用户提到了具体时间或相对时间如"昨天"、"上周"）
- location: 地点
- tags: 标签列表
- participants: 参与的人
- category: 分类（如：旅行、美食、运动、工作、学习、家庭、朋友等）
- mood: 心情（如：开心、平静、兴奋、疲惫等）

请以 JSON 格式返回，包含 "extracted" 和 "follow_up_questions" 两个字段。
如果信息不足，在 follow_up_questions 中提出追问。
"""


class ExtractService:
    """事件提取服务"""

    async def extract_from_text(self, text: str, conversation_history: List[ChatMessage] = None,
                                 language: str = "zh") -> ExtractEventResponse:
        """
        从文本/对话中提取事件信息

        Args:
            text: 输入文本
            conversation_history: 对话历史
            language: 语言

        Returns:
            ExtractEventResponse
        """
        if not settings.LLM_API_URL:
            # 无 LLM 配置时返回基础提取
            return self._basic_extract(text)

        try:
            messages = [{"role": "system", "content": EXTRACT_SYSTEM_PROMPT}]
            if conversation_history:
                for msg in conversation_history[-10:]:  # 最近10条
                    messages.append({"role": msg.role, "content": msg.content})
            messages.append({"role": "user", "content": text})

            async with httpx.AsyncClient() as client:
                resp = await client.post(
                    settings.LLM_API_URL,
                    headers={"Authorization": f"Bearer {settings.LLM_API_KEY}"},
                    json={
                        "model": settings.LLM_MODEL,
                        "messages": messages,
                        "temperature": 0.3,
                        "response_format": {"type": "json_object"},
                    },
                    timeout=30,
                )
                resp.raise_for_status()
                data = resp.json()

                content = data["choices"][0]["message"]["content"]
                parsed = json.loads(content)

                extracted_data = parsed.get("extracted", {})
                return ExtractEventResponse(
                    task_id="",
                    extracted=ExtractedEvent(**extracted_data),
                    follow_up_questions=parsed.get("follow_up_questions", [])
                )

        except Exception as e:
            logger.error(f"LLM 事件提取失败: {e}")
            return self._basic_extract(text)

    def _basic_extract(self, text: str) -> ExtractEventResponse:
        """基础文本提取（无 LLM 时使用）"""
        # 简单规则提取
        title = text[:50] if len(text) > 50 else text
        return ExtractEventResponse(
            task_id="",
            extracted=ExtractedEvent(
                title=title,
                description=text,
                confidence=0.3
            ),
            follow_up_questions=["能告诉我这件事发生的时间吗？", "还有什么想补充的吗？"]
        )


# 全局实例
extract_service = ExtractService()
