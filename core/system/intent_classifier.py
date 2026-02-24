#!/usr/bin/env python3
"""
97layerOS Intent Classifier
대화 vs 인사이트 지능형 분류

Features:
- Gemini를 사용한 지능형 분류
- 질문 vs 명령 vs 인사이트 구분
- 컨텍스트 기반 판단

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import logging
from typing import Dict, Optional
from pathlib import Path

# Load .env
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

try:
    import google.genai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

logger = logging.getLogger(__name__)


class IntentClassifier:
    """
    사용자 메시지의 의도를 지능적으로 분류

    분류:
    - conversation: 대화 (질문, 잡담, 명령)
    - insight: 저장할 인사이트 (아이디어, 메모, 생각)
    - command: 시스템 명령
    """

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Intent Classifier

        Args:
            api_key: Gemini API key
        """
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')

        # Initialize Gemini
        if GEMINI_AVAILABLE and self.api_key:
            self.client = genai.Client(api_key=self.api_key)
            self._model_name = 'gemini-2.5-flash'
            self.use_ai = True
            logger.info("✅ Intent Classifier (AI-powered)")
        else:
            self.client = None
            self.use_ai = False
            logger.warning("⚠️  Intent Classifier (rule-based fallback)")

    def classify(self, text: str, user_context: Optional[Dict] = None) -> Dict:
        """
        메시지 분류

        Returns:
            {
                'intent': 'conversation' | 'insight' | 'command',
                'confidence': float,
                'reasoning': str,
                'suggested_action': str
            }
        """
        if self.use_ai:
            return self._classify_with_ai(text, user_context)
        else:
            return self._classify_with_rules(text)

    def _classify_with_ai(self, text: str, user_context: Optional[Dict] = None) -> Dict:
        """Gemini를 사용한 지능형 분류"""
        try:
            prompt = f"""너는 97layer(순호)의 AI 비서다. 순호가 텔레그램으로 메시지를 보냈다.
이 메시지가 어떤 의도인지 분류해라.

메시지: "{text}"

**분류 기준 (의심스러우면 conversation으로)**:

1. **conversation** (대화 — 기본값):
   - 질문, 요청 (예: "이거 어떻게 생각해?", "분석해줘", "찾아줘")
   - 감정 표현, 반응 (예: "좋네", "별로다", "신기하다")
   - 불만, 피드백 (예: "왜 이렇게 돼?", "아무것도 없는데?", "이상하다")
   - 짧은 대화 (10단어 이하는 거의 항상 conversation)
   - 시스템/봇에 대한 반응 (예: "기록해준다면서", "왜 저장 안 해?")
   - 확인/확답 (예: "맞아", "응", "알겠어")

2. **insight** (저장할 인사이트 — 명확할 때만):
   - 순호가 의도적으로 기록하려는 아이디어/생각임이 명확한 경우
   - 예: "브랜드 컨셉: 여백과 침묵", "릭오웬스처럼 불편함을 미학으로"
   - **반드시**: 구체적인 내용이 있고, 단순 반응/대화가 아닌 경우
   - **절대 아닌 경우**: 짧은 감탄/불만/질문, 시스템 반응, 확인 메시지

3. **command** (시스템 명령):
   - /로 시작하는 명령어만

**핵심 규칙**: 불확실하면 무조건 conversation. insight 오분류가 conversation 오분류보다 훨씬 나쁘다.

**응답 형식** (JSON):
{{
  "intent": "conversation 또는 insight 또는 command",
  "confidence": 0.0 ~ 1.0,
  "reasoning": "이유 한 문장",
  "suggested_action": "어떻게 처리할지"
}}

JSON만 출력하세요.
"""

            response = self.client.models.generate_content(
                model=self._model_name,
                contents=[prompt]
            )
            text_response = response.text.strip()

            # JSON 파싱
            import json
            import re

            # JSON 추출 (```json 태그 제거)
            json_match = re.search(r'\{.*\}', text_response, re.DOTALL)
            if json_match:
                result = json.loads(json_match.group())
                return result
            else:
                # 파싱 실패 시 fallback
                logger.warning("AI response parsing failed, using fallback")
                return self._classify_with_rules(text)

        except Exception as e:
            logger.error("AI classification error: %s", e)
            return self._classify_with_rules(text)

    def _classify_with_rules(self, text: str) -> Dict:
        """규칙 기반 분류 (fallback)"""
        text_lower = text.lower()

        # 명령어
        if text.startswith('/'):
            return {
                'intent': 'command',
                'confidence': 1.0,
                'reasoning': 'Starts with /',
                'suggested_action': 'Process as command'
            }

        # 질문 키워드
        question_keywords = ['?', '뭐', '무엇', '어떻게', '왜', '언제', '어디', '누구',
                            'what', 'how', 'why', 'when', 'where', 'who']
        if any(kw in text_lower for kw in question_keywords):
            return {
                'intent': 'conversation',
                'confidence': 0.8,
                'reasoning': 'Contains question keywords',
                'suggested_action': 'Respond with knowledge base'
            }

        # 인사/감사
        greeting_keywords = ['안녕', '고마', '감사', '좋아', 'hi', 'hello', 'thanks', 'thank you']
        if any(kw in text_lower for kw in greeting_keywords):
            return {
                'intent': 'conversation',
                'confidence': 0.9,
                'reasoning': 'Greeting or gratitude',
                'suggested_action': 'Respond politely'
            }

        # 인사이트 키워드
        insight_keywords = ['아이디어', '컨셉', '기획', '제안', '참고', '메모', '생각', '통찰',
                          '트렌드', '인용', 'idea', 'concept', 'note']
        if any(kw in text_lower for kw in insight_keywords):
            return {
                'intent': 'insight',
                'confidence': 0.7,
                'reasoning': 'Contains insight keywords',
                'suggested_action': 'Save to knowledge base'
            }

        # 길이 기반 판단
        if len(text) > 50:
            # 긴 텍스트는 인사이트일 가능성
            return {
                'intent': 'insight',
                'confidence': 0.6,
                'reasoning': 'Long text (likely insight)',
                'suggested_action': 'Save to knowledge base'
            }
        else:
            # 짧은 텍스트는 대화일 가능성
            return {
                'intent': 'conversation',
                'confidence': 0.5,
                'reasoning': 'Short text (likely conversation)',
                'suggested_action': 'Respond naturally'
            }


# Singleton instance
_classifier_instance = None


def get_intent_classifier() -> IntentClassifier:
    """Get IntentClassifier instance (singleton)"""
    global _classifier_instance
    if _classifier_instance is None:
        _classifier_instance = IntentClassifier()
    return _classifier_instance


def main():
    """Test intent classifier"""
    import sys

    logging.basicConfig(level=logging.INFO)

    classifier = IntentClassifier()

    test_messages = [
        "우리 철학은",
        "안녕하세요",
        "이거 분석해줘",
        "아이디어: 브랜드 리뉴얼을 위한 새로운 컨셉 - 미니멀리즘과 본질 추구",
        "요즘 트렌드를 보니 사람들이 과도한 정보에 지쳐있는 것 같다. 우리의 절제 철학이 더 중요해질 것.",
        "보고서 작성",
        "/status",
        "이게 뭐야?",
        "고마워"
    ]

    print("\n" + "="*60)
    print("Intent Classification Test")
    print("="*60 + "\n")

    for msg in test_messages:
        result = classifier.classify(msg)
        print(f"📝 Message: \"{msg}\"")
        print(f"   Intent: {result['intent']} (confidence: {result['confidence']})")
        print(f"   Reasoning: {result['reasoning']}")
        print(f"   Action: {result['suggested_action']}")
        print()


if __name__ == "__main__":
    main()
