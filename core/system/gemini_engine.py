#!/usr/bin/env python3
"""
97layerOS Gemini 2.0 Engine
Uses new google-genai SDK (not deprecated)

Capabilities:
- Text generation (conversation)
- Vision (image analysis)
- Long context (1M tokens)
- File processing
"""

import os
from pathlib import Path
from typing import Optional, List, Dict, Any
import logging

try:
    from google import genai
    from google.genai import types
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


class GeminiEngine:
    """
    Gemini 2.0 Flash engine with multimodal capabilities
    """

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Gemini Engine

        Args:
            api_key: Google API key
        """
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')

        if not GEMINI_AVAILABLE:
            logger.error("❌ google-genai package not installed")
            self.client = None
            return

        if not self.api_key:
            logger.error("❌ GOOGLE_API_KEY not found in environment")
            self.client = None
            return

        try:
            # Initialize client
            self.client = genai.Client(api_key=self.api_key)

            # Default model
            self.model_name = "gemini-2.5-flash"

            logger.info(f"✅ Gemini 2.0 Engine initialized ({self.model_name})")

        except Exception as e:
            logger.error(f"❌ Failed to initialize Gemini: {e}")
            self.client = None

    def is_available(self) -> bool:
        """Check if Gemini is available"""
        return self.client is not None

    # ========================
    # Text Generation
    # ========================

    def generate_text(
        self,
        prompt: str,
        system_instruction: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: int = 2048
    ) -> str:
        """
        Generate text response

        Args:
            prompt: User prompt
            system_instruction: System instruction
            temperature: Generation temperature (0.0-1.0)
            max_tokens: Max output tokens

        Returns:
            Generated text
        """
        if not self.client:
            return "❌ Gemini not available"

        try:
            config = types.GenerateContentConfig(
                temperature=temperature,
                max_output_tokens=max_tokens,
                system_instruction=system_instruction
            )

            response = self.client.models.generate_content(
                model=self.model_name,
                contents=prompt,
                config=config
            )

            return response.text

        except Exception as e:
            logger.error(f"Text generation failed: {e}")
            return f"❌ 생성 실패: {str(e)}"

    # ========================
    # Vision (Image Analysis)
    # ========================

    def analyze_image(
        self,
        image_path: Path,
        prompt: str = "이 이미지를 상세히 설명해주세요.",
        system_instruction: Optional[str] = None
    ) -> str:
        """
        Analyze image with Gemini Vision

        Args:
            image_path: Path to image file
            prompt: Analysis prompt
            system_instruction: System instruction

        Returns:
            Analysis result
        """
        if not self.client:
            return "❌ Gemini not available"

        if not image_path.exists():
            return f"❌ 이미지 파일 없음: {image_path}"

        try:
            # Upload file
            file = self.client.files.upload(path=str(image_path))
            logger.info(f"📤 Image uploaded: {file.name}")

            # Generate content with image
            config = types.GenerateContentConfig(
                system_instruction=system_instruction,
                temperature=0.7
            )

            response = self.client.models.generate_content(
                model=self.model_name,
                contents=[
                    types.Part.from_uri(file_uri=file.uri, mime_type=file.mime_type),
                    prompt
                ],
                config=config
            )

            # Clean up
            self.client.files.delete(name=file.name)

            return response.text

        except Exception as e:
            logger.error(f"Image analysis failed: {e}")
            return f"❌ 이미지 분석 실패: {str(e)}"

    # ========================
    # Long Context (Knowledge RAG)
    # ========================

    def query_with_context(
        self,
        question: str,
        context: str,
        system_instruction: Optional[str] = None,
        max_tokens: int = 2048
    ) -> str:
        """
        Query with long context (RAG)

        Args:
            question: User question
            context: Knowledge base context
            system_instruction: System instruction
            max_tokens: Max output tokens

        Returns:
            Answer
        """
        if not self.client:
            return "❌ Gemini not available"

        try:
            prompt = f"""**Context**:
{context}

**Question**: {question}

**Answer** (간결하게 2-3문단):"""

            config = types.GenerateContentConfig(
                temperature=0.7,
                max_output_tokens=max_tokens,
                system_instruction=system_instruction
            )

            response = self.client.models.generate_content(
                model=self.model_name,
                contents=prompt,
                config=config
            )

            return response.text

        except Exception as e:
            logger.error(f"Context query failed: {e}")
            return f"❌ 쿼리 실패: {str(e)}"

    # ========================
    # Conversation
    # ========================

    def chat(
        self,
        message: str,
        history: Optional[List[Dict[str, str]]] = None,
        system_instruction: Optional[str] = None
    ) -> str:
        """
        Chat with conversation history

        Args:
            message: User message
            history: Conversation history [{'role': 'user'/'model', 'text': '...'}]
            system_instruction: System instruction

        Returns:
            Response
        """
        if not self.client:
            return "❌ Gemini not available"

        try:
            # Build contents with history
            contents = []

            if history:
                for msg in history[-10:]:  # Last 10 messages
                    role = "user" if msg['role'] == 'user' else "model"
                    contents.append(types.Content(role=role, parts=[types.Part.from_text(msg['text'])]))

            # Add current message
            contents.append(types.Content(role="user", parts=[types.Part.from_text(message)]))

            config = types.GenerateContentConfig(
                temperature=0.8,
                system_instruction=system_instruction
            )

            response = self.client.models.generate_content(
                model=self.model_name,
                contents=contents,
                config=config
            )

            return response.text

        except Exception as e:
            logger.error(f"Chat failed: {e}")
            return f"❌ 대화 실패: {str(e)}"


# ========================
# Singleton
# ========================

_gemini_engine_instance = None


def get_gemini_engine() -> GeminiEngine:
    """Get GeminiEngine instance (singleton)"""
    global _gemini_engine_instance
    if _gemini_engine_instance is None:
        _gemini_engine_instance = GeminiEngine()
    return _gemini_engine_instance


# ========================
# CLI Testing
# ========================

def main():
    """CLI test"""
    import sys

    logging.basicConfig(level=logging.INFO)

    engine = GeminiEngine()

    print("97layerOS Gemini 2.0 Engine Test")
    print("=" * 50)
    print(f"Status: {'✅ Available' if engine.is_available() else '❌ Not available'}")
    print("=" * 50)

    if not engine.is_available():
        print("\nError: Gemini not available")
        print("Check: GOOGLE_API_KEY in .env")
        return

    if len(sys.argv) > 1:
        prompt = " ".join(sys.argv[1:])
        print(f"\nPrompt: {prompt}\n")
        response = engine.generate_text(prompt)
        print(f"Response:\n{response}\n")
    else:
        print("\nUsage: python3 gemini_engine.py <prompt>")
        print("Example: python3 gemini_engine.py '안녕하세요'")


if __name__ == "__main__":
    main()
