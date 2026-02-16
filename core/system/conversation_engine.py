#!/usr/bin/env python3
"""
97layerOS Conversation Engine
Gemini LLM을 사용한 자연스러운 대화

Features:
- Gemini LLM 연결
- 지식 베이스 기반 응답 (RAG)
- 대화 컨텍스트 유지
- 97layer 브랜드 철학 반영

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import json
import logging
from pathlib import Path
from typing import Dict, List, Optional

# Load .env
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

try:
    import google.generativeai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

from core.bridges.notebooklm_bridge import get_bridge

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


class ConversationEngine:
    """
    Gemini LLM 기반 대화 엔진

    지식 베이스를 참고하여 자연스럽게 응답
    """

    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Conversation Engine

        Args:
            api_key: Gemini API key
        """
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')

        # Initialize Gemini
        if GEMINI_AVAILABLE and self.api_key:
            genai.configure(api_key=self.api_key)
            self.model = genai.GenerativeModel('gemini-1.5-flash')
            logger.info("✅ Conversation Engine (LLM-powered)")
        else:
            # Initialize NotebookLM
            self.notebooklm = get_bridge()
            logger.info(f"✅ NotebookLM Bridge {'ready' if self.notebooklm.authenticated else 'standby'}")

        # Knowledge base
        self.knowledge_dir = PROJECT_ROOT / 'knowledge'
        self.signals_dir = self.knowledge_dir / 'signals'
        self.docs_dir = self.knowledge_dir / 'docs'

        # User conversation contexts
        self.user_contexts = {}

        # 97layer 브랜드 철학
        self.brand_philosophy = """
97layer 브랜드 철학:
- **본질 (Essence)**: 핵심만 남기고 불필요한 것은 제거
- **절제 (Restraint)**: 과도함을 피하고 균형 추구
- **자기긍정 (Self-affirmation)**: 자신만의 가치 발견

톤앤매너:
- 냉철하고 고지능적인 사고 기반의 정제된 언어
- 가식적인 공감과 불필요한 수식어(감탄사 등) 완전 배제
- 구조적 필연성을 도출하는 오케스트레이터의 위엄 유지
- 한국어 존댓말을 사용하되 지극히 실용적이고 건조한 문체
"""

    def chat(self, user_id: str, message: str, use_knowledge: bool = True) -> str:
        """
        사용자와 대화

        Args:
            user_id: 사용자 ID
            message: 사용자 메시지
            use_knowledge: 지식 베이스 사용 여부

        Returns:
            응답 메시지
        """
        if not self.model:
            return "죄송합니다. 대화 기능을 사용할 수 없습니다."

        try:
            # 사용자 컨텍스트 가져오기
            if user_id not in self.user_contexts:
                self.user_contexts[user_id] = {
                    'history': [],
                    'last_interaction': None
                }

            context = self.user_contexts[user_id]

            # 지식 베이스 검색 (간단한 버전)
            relevant_knowledge = ""
            if use_knowledge:
                relevant_knowledge = self._search_knowledge(message)

            # 프롬프트 구성
            prompt = self._build_prompt(message, context, relevant_knowledge)

            # LLM 호출
            response = self.model.generate_content(prompt)
            answer = response.text.strip()

            # 컨텍스트 업데이트
            context['history'].append({
                'user': message,
                'assistant': answer
            })

            # 최근 5개만 유지
            if len(context['history']) > 5:
                context['history'] = context['history'][-5:]

            return answer

        except Exception as e:
            logger.error(f"Conversation error: {e}")
            return f"죄송합니다. 응답 생성 중 오류가 발생했습니다: {str(e)}"

    def _build_prompt(self, message: str, context: Dict, knowledge: str) -> str:
        """프롬프트 구성"""

        # 대화 히스토리
        history_text = ""
        if context['history']:
            recent = context['history'][-3:]  # 최근 3개
            for h in recent:
                history_text += f"사용자: {h['user']}\n조수: {h['assistant']}\n\n"

        # 지식 베이스
        knowledge_text = ""
        if knowledge:
            knowledge_text = f"""
**참고 자료** (우리 지식 베이스):
{knowledge}
"""

        prompt = f"""당신은 97layer의 AI 비서입니다.

{self.brand_philosophy}

**역할**:
- 97layer 팀의 업무를 지원하는 친근한 AI 비서
- 질문에 답하고, 인사이트를 제공하며, 작업을 돕습니다
- 우리의 지식 베이스를 참고하여 정확한 답변을 제공합니다

{knowledge_text}

**최근 대화**:
{history_text if history_text else "없음"}

**현재 질문**:
사용자: {message}

**지침**:
1. 불필요한 인사나 미사여구를 생략하고 본론부터 간결하게 답하십시오.
2. 지식 베이스(Deep RAG)의 데이터를 논리적으로 연결하여 통찰력 있는 결론을 도출하십시오.
3. 모르는 데이터에 대해 추측하지 말고, 데이터 부재를 명확히 밝히고 대안을 제시하십시오.
4. 97layer의 수석 오케스트레이터로서 사령관에게 전략적 조언을 건네는 스탠스를 유지하십시오.
5. 마크다운 구조(리스트, 볼드 등)를 사용하여 가독성을 극대화하십시오.

조수:[(냉철하고 정제된 응답 시작)]"""

        return prompt

    def _search_knowledge(self, query: str) -> str:
        try:
            # NotebookLM Deep RAG 우선 사용
            if self.notebooklm.authenticated:
                logger.info(f"Deep RAG searching for: {query[:30]}...")
                answer = self.notebooklm.query_knowledge_base(query)
                if answer and "no information found" not in answer.lower():
                    return f"**지식 베이스 응답**:\n{answer}\n"

            # Fallback: 로컬 마크다운 검색
            relevant_docs = []
            keywords = [w.lower() for w in query.split() if len(w) > 2]

            if self.docs_dir.exists():
                for md_file in self.docs_dir.glob('**/*.md'):
                    try:
                        with open(md_file, 'r', encoding='utf-8') as f:
                            content = f.read()
                            if any(kw in content.lower() for kw in keywords):
                                preview = content[:300]
                                relevant_docs.append(f"**{md_file.name}**:\n{preview}...\n")
                                if len(relevant_docs) >= 1: break
                    except Exception: pass

            if relevant_docs:
                return ''.join(relevant_docs)
            return ""

        except Exception as e:
            logger.error(f"Knowledge search error: {e}")
            return ""

    def _get_recent_signals(self, limit: int = 5) -> List[Dict]:
        """최근 신호 가져오기"""
        signals = []

        try:
            # JSON 파일들 수집
            json_files = sorted(
                self.signals_dir.glob('**/*.json'),
                key=lambda p: p.stat().st_mtime,
                reverse=True
            )[:limit]

            for file in json_files:
                try:
                    with open(file, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        signals.append(data)
                except Exception as e:
                    logger.error(f"Error reading {file}: {e}")

        except Exception as e:
            logger.error(f"Error getting signals: {e}")

        return signals

    def clear_context(self, user_id: str):
        """사용자 컨텍스트 초기화"""
        if user_id in self.user_contexts:
            del self.user_contexts[user_id]
            logger.info(f"Cleared context for user {user_id}")


# Singleton instance
_conversation_engine = None


def get_conversation_engine() -> ConversationEngine:
    """Get ConversationEngine instance (singleton)"""
    global _conversation_engine
    if _conversation_engine is None:
        _conversation_engine = ConversationEngine()
    return _conversation_engine


def main():
    """Test conversation engine"""
    import sys

    logging.basicConfig(level=logging.INFO)

    engine = ConversationEngine()

    print("\n" + "="*60)
    print("97layer AI 비서 - 대화 테스트")
    print("="*60)
    print("('exit' 입력시 종료)\n")

    user_id = "test_user"

    while True:
        try:
            user_input = input("사용자: ")
            if user_input.lower() in ['exit', 'quit', '종료']:
                break

            response = engine.chat(user_id, user_input)
            print(f"비서: {response}\n")

        except KeyboardInterrupt:
            break
        except EOFError:
            break

    print("\n대화를 종료합니다.")


if __name__ == "__main__":
    main()
