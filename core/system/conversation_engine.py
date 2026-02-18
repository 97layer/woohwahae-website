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
from datetime import datetime

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

            # 자가발전: Cortex에서 통합 관리하므로 여기서는 스킵 (중복 방지)
            # self._update_long_term_memory(message, answer)

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

    def _update_long_term_memory(self, user_message: str, assistant_answer: str):
        """
        대화 완료 후 long_term_memory.json 자동 업데이트 — 자가발전 핵심 고리

        Gemini로 대화에서 핵심 개념/패턴을 추출해 누적 저장.
        대화가 쌓일수록 비서의 개인화 품질이 향상된다.
        """
        if not self.model:
            return

        lm_path = self.knowledge_dir / 'long_term_memory.json'

        # 현재 메모리 로드
        try:
            if lm_path.exists():
                data = json.loads(lm_path.read_text(encoding='utf-8'))
            else:
                data = {
                    'metadata': {'created_at': datetime.now().isoformat(), 'total_entries': 0},
                    'experiences': [],
                    'concepts': {},
                    'error_patterns': []
                }
        except Exception:
            return

        # Gemini로 개념 추출
        extract_prompt = f"""다음 대화에서 97layer 개인의 관심사, 패턴, 핵심 개념을 추출하라.

사용자: {user_message[:300]}
비서: {assistant_answer[:300]}

JSON으로만 응답:
{{
  "concepts": ["개념1", "개념2"],
  "summary": "한 문장 요약",
  "category": "브랜드/개인/기술/비즈니스/라이프스타일 중 하나"
}}"""

        try:
            resp = self.client.models.generate_content(
                model=self._model_name,
                contents=[extract_prompt]
            )
            import re
            json_match = re.search(r'\{.*\}', resp.text, re.DOTALL)
            if not json_match:
                return
            extracted = json.loads(json_match.group())
        except Exception:
            return

        # concepts 카운트 누적 (빈도 = 중요도)
        for concept in extracted.get('concepts', []):
            concept = concept.strip()
            if concept:
                data['concepts'][concept] = data['concepts'].get(concept, 0) + 1

        # experiences 추가
        data['experiences'].append({
            'summary': extracted.get('summary', '')[:100],
            'category': extracted.get('category', 'unknown'),
            'timestamp': datetime.now().isoformat()[:16]
        })

        # 최근 100개만 유지
        if len(data['experiences']) > 100:
            data['experiences'] = data['experiences'][-100:]

        data['metadata']['total_entries'] = len(data['experiences'])
        data['metadata']['last_updated'] = datetime.now().isoformat()[:16]

        # 저장
        lm_path.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding='utf-8')
        logger.debug(f"Memory updated: +{len(extracted.get('concepts', []))} concepts")

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

        prompt = f"""너는 97layer(순호)의 개인 비서 겸 콘텐츠 실행 파트너야. 이름은 97layerOS.

순호에 대해 알고 있는 것:
{self.brand_philosophy[:1500]}

최근 순호가 기록한 것들:
{context.get('recent_signals', '없음')}

지금까지 쌓인 기억:
{context.get('long_term_memory', '없음')}

{knowledge_text}

이전 대화:
{history_text if history_text else "없음"}

순호: {message}

행동 지침:
- 순호의 말이 **실행 지시**면 (예: "~해줘", "~뽑아줘", "~만들어줘", "~화해줘") → 즉시 실행해서 결과물을 줘. 설명만 하지 마.
- 순호의 말이 **RAG/검색 요청**이면 → 위의 지식 베이스와 참고 자료에서 실제로 관련 내용을 찾아서 뽑아줘. "없습니다"보다 있는 것부터 보여줘.
- 순호의 말이 **콘텐츠 생성 요청**이면 → 우리 브랜드 철학(woohwahae, 슬로우라이프, 조용한 지능)을 반영해서 바로 초안을 써줘.
- 순호의 말이 **질문/대화**면 → 자연스럽게 대화해.
- 비서답게, 친근하되 존중하는 톤. 감탄사 없이. 결과물 먼저, 설명은 짧게.

97layerOS:"""

        return prompt

    def _search_knowledge(self, query: str) -> str:
        """
        로컬 지식 검색 (즉시 응답, <1초)
        NotebookLM은 저장소로만 사용 — 대화 중 실시간 쿼리 안 함 (12초+ 지연)

        검색 순서:
        1. signals/ JSON (순호가 보낸 인사이트 + SA 분석 결과)
        2. corpus/ entries + index (군집 테마 + 분석 결과)
        3. long_term_memory.json (누적 개념 및 경험)
        4. directives/ 마크다운 (브랜드/아이덴티티)
        """
        try:
            keywords = [w.lower() for w in query.split() if len(w) > 1]
            if not keywords:
                return ""
            results = []

            # 1. signals/ JSON — 순호 인사이트 + SA 분석 결과 (전체 검색)
            signals_dir = self.knowledge_dir / 'signals'
            if signals_dir.exists():
                signal_files = sorted(
                    signals_dir.glob('**/*.json'),
                    key=lambda f: f.stat().st_mtime,
                    reverse=True  # 최신순
                )
                matched = []
                for sf in signal_files:  # 전체 검색 (제한 없음)
                    try:
                        data = json.loads(sf.read_text(encoding='utf-8'))
                        content = data.get('content', '') + ' ' + json.dumps(
                            data.get('analysis', {}), ensure_ascii=False
                        )
                        if any(kw in content.lower() for kw in keywords):
                            summary = data.get('analysis', {}).get('summary', '')
                            score = data.get('analysis', {}).get('strategic_score', '')
                            text = data.get('content', '')[:200]
                            entry = f"[signal/{sf.stem}]"
                            if score:
                                entry += f" score:{score}"
                            if summary:
                                entry += f"\n요약: {summary}"
                            else:
                                entry += f"\n{text}"
                            matched.append(entry)
                            if len(matched) >= 8:  # 최대 8개 (5→8)
                                break
                    except Exception:
                        pass
                if matched:
                    results.append("**[저장된 인사이트]**\n" + "\n\n".join(matched))

            # 2. corpus/ — 군집 테마 + 분석 entries
            corpus_dir = self.knowledge_dir / 'corpus'
            if corpus_dir.exists():
                # 2a. index.json — 테마 군집 매칭
                corpus_index = corpus_dir / 'index.json'
                if corpus_index.exists():
                    try:
                        ci = json.loads(corpus_index.read_text(encoding='utf-8'))
                        matched_themes = []
                        for theme, cluster in ci.get('clusters', {}).items():
                            if any(kw in theme.lower() for kw in keywords):
                                cnt = len(cluster.get('entry_ids', []))
                                matched_themes.append(f"테마 '{theme}': {cnt}개 신호 축적")
                        if matched_themes:
                            results.append("**[Corpus 군집]**\n" + "\n".join(matched_themes))
                    except Exception:
                        pass

                # 2b. entries/ — SA 분석 결과 풀텍스트 검색
                entries_dir = corpus_dir / 'entries'
                if entries_dir.exists():
                    matched_entries = []
                    for ef in sorted(entries_dir.glob('*.json'), key=lambda f: f.stat().st_mtime, reverse=True):
                        try:
                            data = json.loads(ef.read_text(encoding='utf-8'))
                            searchable = json.dumps(data, ensure_ascii=False)
                            if any(kw in searchable.lower() for kw in keywords):
                                summary = data.get('summary', '') or data.get('analysis', {}).get('summary', '')
                                themes = data.get('themes', data.get('analysis', {}).get('themes', []))
                                entry = f"[corpus/{ef.stem}]"
                                if themes:
                                    entry += f" 테마: {', '.join(str(t) for t in themes[:3])}"
                                if summary:
                                    entry += f"\n{summary[:150]}"
                                matched_entries.append(entry)
                                if len(matched_entries) >= 5:
                                    break
                        except Exception:
                            pass
                    if matched_entries:
                        results.append("**[Corpus 분석]**\n" + "\n\n".join(matched_entries))

            # 3. long_term_memory — 누적 개념/패턴
            lm_path = self.knowledge_dir / 'long_term_memory.json'
            if lm_path.exists():
                try:
                    lm = json.loads(lm_path.read_text(encoding='utf-8'))
                    concepts = lm.get('concepts', {})
                    matched_concepts = [
                        f"{k}({v}회)" for k, v in concepts.items()
                        if any(kw in k.lower() for kw in keywords)
                    ]
                    experiences = lm.get('experiences', [])
                    matched_exp = []
                    for exp in reversed(experiences):  # 전체 경험 검색 (50→전체)
                        s = exp.get('summary', '')
                        if any(kw in s.lower() for kw in keywords):
                            matched_exp.append(f"- {s[:120]}")
                            if len(matched_exp) >= 5:  # 3→5개
                                break
                    if matched_concepts or matched_exp:
                        lm_text = ""
                        if matched_concepts:
                            lm_text += "개념: " + ", ".join(matched_concepts[:8]) + "\n"
                        if matched_exp:
                            lm_text += "\n".join(matched_exp)
                        results.append(f"**[장기 기억]**\n{lm_text}")
                except Exception:
                    pass

            # 4. directives/ 마크다운 — 브랜드/아이덴티티
            for md_file in sorted(self.directives_dir.glob('**/*.md')):
                try:
                    content = md_file.read_text(encoding='utf-8')
                    if any(kw in content.lower() for kw in keywords):
                        results.append(f"**[{md_file.name}]**\n{content[:400]}...")
                        if len(results) >= 6:  # 5→6
                            break
                except Exception:
                    pass

            return "\n\n".join(results) if results else ""

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
