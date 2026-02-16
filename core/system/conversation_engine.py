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
    import google.genai as genai
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
            self.client = genai.Client(api_key=self.api_key)
            self._model_name = 'gemini-2.5-flash'
            self.model = True  # flag: LLM available
            logger.info("✅ Conversation Engine (LLM-powered)")
        else:
            self.client = None
            self.model = None
            logger.info("⚠️ Gemini unavailable — LLM disabled")

        # Always initialize NotebookLM bridge (used in _search_knowledge)
        try:
            self.notebooklm = get_bridge()
            logger.info("✅ NotebookLM Bridge standby")
        except Exception as e:
            self.notebooklm = None
            logger.warning("⚠️ NotebookLM Bridge unavailable: %s", e)

        # Knowledge base
        self.knowledge_dir = PROJECT_ROOT / 'knowledge'
        self.signals_dir = self.knowledge_dir / 'signals'
        self.docs_dir = self.knowledge_dir / 'docs'
        self.directives_dir = PROJECT_ROOT / 'directives'

        # User conversation contexts
        self.user_contexts = {}

        # 브랜드 철학: IDENTITY.md에서 동적 로드
        self.brand_philosophy = self._load_identity()

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

            # 로컬 컨텍스트 주입 (매 대화마다 최신 상태 반영)
            context['recent_signals'] = self._load_recent_signals(limit=10)
            context['long_term_memory'] = self._load_long_term_memory()

            # 지식 베이스 검색
            relevant_knowledge = ""
            if use_knowledge:
                relevant_knowledge = self._search_knowledge(message)

            # 프롬프트 구성
            prompt = self._build_prompt(message, context, relevant_knowledge)

            # LLM 호출
            response = self.client.models.generate_content(
                model=self._model_name,
                contents=[prompt]
            )
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

    def _load_identity(self) -> str:
        """IDENTITY.md + SYSTEM.md에서 브랜드 철학 로드"""
        sections = []
        for path in [
            self.directives_dir / 'IDENTITY.md',
            self.directives_dir / 'system' / 'SYSTEM.md',
        ]:
            try:
                if path.exists():
                    content = path.read_text(encoding='utf-8')
                    sections.append(content[:3000])  # 각 최대 3000자
            except Exception:
                pass
        if sections:
            return '\n\n---\n\n'.join(sections)
        # fallback
        return """97layer 브랜드 철학: 본질만 남기고 소음을 제거한다.
톤: 냉철하고 건조하며 실용적인 한국어 존댓말. 감탄사·미사여구 배제."""

    def _load_recent_signals(self, limit: int = 10) -> str:
        """최근 저장된 signals 로드 — 내가 뭘 기록했는지 비서가 알 수 있도록"""
        try:
            files = sorted(
                self.signals_dir.glob('**/*.json'),
                key=lambda p: p.stat().st_mtime,
                reverse=True
            )[:limit]
            if not files:
                return ""
            lines = []
            for f in files:
                try:
                    data = json.loads(f.read_text(encoding='utf-8'))
                    captured = data.get('captured_at', '')[:10]
                    content = data.get('content', data.get('transcript', ''))[:200]
                    signal_type = data.get('type', 'unknown')
                    lines.append(f"[{captured}][{signal_type}] {content}")
                except Exception:
                    pass
            return '\n'.join(lines)
        except Exception:
            return ""

    def _load_long_term_memory(self) -> str:
        """long_term_memory.json 로드"""
        try:
            lm_path = self.knowledge_dir / 'long_term_memory.json'
            if not lm_path.exists():
                return ""
            data = json.loads(lm_path.read_text(encoding='utf-8'))
            experiences = data.get('experiences', [])
            concepts = data.get('concepts', {})
            parts = []
            if experiences:
                recent = experiences[-5:]
                parts.append("경험: " + " | ".join(
                    e.get('summary', str(e))[:80] for e in recent
                ))
            if concepts:
                top = list(concepts.items())[:10]
                parts.append("개념: " + ", ".join(f"{k}({v})" for k, v in top))
            return '\n'.join(parts)
        except Exception:
            return ""

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

        prompt = f"""당신은 97layer(순호)의 전용 AI 비서 — 97layerOS입니다.

=== 브랜드 정체성 & 운영 원칙 ===
{self.brand_philosophy}

=== 최근 기록된 신호 (내가 저장한 인사이트/메모) ===
{context.get('recent_signals', '없음')}

=== 장기 기억 ===
{context.get('long_term_memory', '없음')}

{knowledge_text}

=== 최근 대화 ===
{history_text if history_text else "없음"}

=== 현재 질문 ===
{message}

지침:
1. 본론부터. 인사·감탄사·미사여구 없이.
2. 위의 신호·기억·지식을 실제로 참조해서 개인화된 답변을 구성하라.
3. 모르는 것은 추측 말고 "데이터 없음"을 명시 후 대안 제시.
4. 수석 오케스트레이터 스탠스 유지 — 전략적 조언 우선.
5. 마크다운(리스트, 볼드) 활용해 가독성 극대화."""

        return prompt

    def _search_knowledge(self, query: str) -> str:
        try:
            # NotebookLM Deep RAG 우선 사용
            if self.notebooklm and self.notebooklm.authenticated:
                logger.info(f"Deep RAG searching for: {query[:30]}...")
                answer = self.notebooklm.query_knowledge_base(query)
                if answer and "no information found" not in answer.lower():
                    return f"**지식 베이스 응답**:\n{answer}\n"

            # Fallback: 로컬 마크다운 검색 (docs + agent_hub + directives)
            relevant_docs = []
            keywords = [w.lower() for w in query.split() if len(w) > 2]

            search_dirs = [
                self.docs_dir,
                self.knowledge_dir / 'agent_hub',
                self.directives_dir,
            ]
            for search_dir in search_dirs:
                if not search_dir.exists():
                    continue
                for md_file in sorted(search_dir.glob('**/*.md')):
                    try:
                        content = md_file.read_text(encoding='utf-8')
                        if any(kw in content.lower() for kw in keywords):
                            preview = content[:500]
                            relevant_docs.append(f"**[{md_file.name}]**:\n{preview}...\n")
                            if len(relevant_docs) >= 3:
                                break
                    except Exception:
                        pass
                if len(relevant_docs) >= 3:
                    break

            if relevant_docs:
                return '\n\n'.join(relevant_docs)
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
