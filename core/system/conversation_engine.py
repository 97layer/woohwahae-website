#!/usr/bin/env python3
"""
LAYER OS Conversation Engine
순호의 두 번째 뇌 — 동료 포지션 대화
"""

import json
import logging
import os
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional

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

from core.system.notebooklm_bridge import get_bridge

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
CONTEXTS_PATH = PROJECT_ROOT / 'knowledge' / 'system' / 'conversation_contexts.json'

# 메모리 업데이트 주기 (N번째 대화마다)
MEMORY_UPDATE_EVERY = 5


class ConversationEngine:

    def __init__(self, api_key: Optional[str] = None):
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')

        if GEMINI_AVAILABLE and self.api_key:
            self.client = genai.Client(api_key=self.api_key)
            self._model_name = 'gemini-2.5-flash'
            self.model = True
            logger.info("ConversationEngine ready (LLM)")
        else:
            self.client = None
            self.model = None
            logger.warning("ConversationEngine: Gemini unavailable")

        try:
            self.notebooklm = get_bridge()
        except Exception as e:
            self.notebooklm = None
            logger.warning("NotebookLM bridge unavailable: %s", e)

        self.knowledge_dir = PROJECT_ROOT / 'knowledge'
        self.directives_dir = PROJECT_ROOT / 'directives'

        # 파일 기반 컨텍스트 (재시작 후에도 유지)
        self.user_contexts: Dict = self._load_contexts()

        # 브랜드 철학 (세션 단위 캐시)
        self._identity_cache: Optional[str] = None

    # ─── 컨텍스트 영속성 ──────────────────────────────────────────

    def _load_contexts(self) -> Dict:
        try:
            if CONTEXTS_PATH.exists():
                data = json.loads(CONTEXTS_PATH.read_text(encoding='utf-8'))
                # 각 유저의 history만 복원 (recent_signals/long_term_memory은 매 대화마다 갱신)
                return {
                    uid: {'history': ctx.get('history', []), 'last_interaction': ctx.get('last_interaction')}
                    for uid, ctx in data.items()
                }
        except Exception as e:
            logger.warning("컨텍스트 로드 실패: %s", e)
        return {}

    def _save_context(self, user_id: str) -> None:
        try:
            ctx = self.user_contexts.get(user_id, {})
            save_data = {
                uid: {'history': c.get('history', []), 'last_interaction': c.get('last_interaction')}
                for uid, c in self.user_contexts.items()
            }
            CONTEXTS_PATH.parent.mkdir(parents=True, exist_ok=True)
            CONTEXTS_PATH.write_text(json.dumps(save_data, ensure_ascii=False, indent=2), encoding='utf-8')
        except Exception as e:
            logger.warning("컨텍스트 저장 실패: %s", e)

    # ─── 메인 대화 ────────────────────────────────────────────────

    def chat(self, user_id: str, message: str, use_knowledge: bool = True) -> str:
        if not self.model:
            return "LLM 연결 안 됨."

        try:
            if user_id not in self.user_contexts:
                self.user_contexts[user_id] = {'history': [], 'last_interaction': None}

            ctx = self.user_contexts[user_id]
            ctx['recent_signals'] = self._load_recent_signals(limit=8)
            ctx['long_term_memory'] = self._load_long_term_memory()

            knowledge = self._search_knowledge(message) if use_knowledge else ""
            prompt = self._build_prompt(message, ctx, knowledge)

            response = self.client.models.generate_content(
                model=self._model_name,
                contents=[prompt]
            )
            answer = response.text.strip()

            ctx['history'].append({'user': message, 'assistant': answer})
            if len(ctx['history']) > 6:
                ctx['history'] = ctx['history'][-6:]
            ctx['last_interaction'] = datetime.now().isoformat()[:16]

            self._save_context(user_id)

            # 학습: MEMORY_UPDATE_EVERY번마다 장기 기억 갱신
            if len(ctx['history']) % MEMORY_UPDATE_EVERY == 0:
                self._update_long_term_memory(message, answer)

            return answer

        except Exception as e:
            logger.error("대화 오류: %s", e)
            return f"오류: {e}"

    # ─── 프롬프트 ────────────────────────────────────────────────

    def _build_prompt(self, message: str, context: Dict, knowledge: str) -> str:
        identity = self._get_identity()

        history_lines = ""
        for h in context.get('history', [])[-4:]:
            history_lines += f"순호: {h['user']}\nLAYER OS: {h['assistant']}\n\n"

        knowledge_block = f"\n[참고]\n{knowledge}\n" if knowledge else ""

        recent = context.get('recent_signals', '')
        memory = context.get('long_term_memory', '')

        return f"""너는 LAYER OS. 순호의 두 번째 뇌이자 동료.

[순호에 대해]
{identity}

[최근 기록된 신호]
{recent if recent else '없음'}

[장기 기억]
{memory if memory else '없음'}
{knowledge_block}
[이전 대화]
{history_lines if history_lines else '없음'}

순호: {message}

[응답 규칙 — 반드시 준수]
- 한국어. 반말. (~야/~어/~다/~지)
- 감탄사("오!", "맞아요!", "훌륭하네요") 절대 금지.
- 빈 공감("좋은 생각", "그렇군요") 절대 금지.
- 불필요한 서두/마무리 없이 바로 본론.
- 실행 요청 → 결과물 먼저, 설명은 한 줄 이하.
- 질문/대화 → 핵심만. 보통 2-3문장.
- 콘텐츠 생성 요청 → 우리 브랜드(슬로우라이프, 본질, 여백) 기준으로 바로 초안.
- 모르면 모른다고. 추측 안 함.

LAYER OS:"""

    def _get_identity(self) -> str:
        if self._identity_cache:
            return self._identity_cache
        parts = []
        for path in [self.directives_dir / 'the_origin.md', self.directives_dir / 'system.md']:
            try:
                if path.exists():
                    parts.append(path.read_text(encoding='utf-8')[:2000])
            except Exception:
                pass
        self._identity_cache = '\n\n---\n\n'.join(parts) if parts else (
            "WOOHWAHAE — 슬로우라이프 헤어 아틀리에. 본질, 여백, 조용한 지능."
        )
        return self._identity_cache

    # ─── 지식 검색 ───────────────────────────────────────────────

    def _search_knowledge(self, query: str) -> str:
        try:
            keywords = [w.lower() for w in query.split() if len(w) > 1]
            if not keywords:
                return ""
            results = []

            # 1. signals/
            signals_dir = self.knowledge_dir / 'signals'
            if signals_dir.exists():
                matched = []
                for sf in sorted(signals_dir.glob('**/*.json'), key=lambda f: f.stat().st_mtime, reverse=True):
                    try:
                        data = json.loads(sf.read_text(encoding='utf-8'))
                        content = data.get('content', '') + ' ' + json.dumps(
                            data.get('analysis', {}), ensure_ascii=False
                        )
                        if any(kw in content.lower() for kw in keywords):
                            summary = data.get('analysis', {}).get('summary', '')
                            score = data.get('analysis', {}).get('strategic_score', '')
                            entry = f"[{sf.stem}]"
                            if score:
                                entry += f" score:{score}"
                            entry += f"\n{summary or data.get('content', '')[:200]}"
                            matched.append(entry)
                            if len(matched) >= 6:
                                break
                    except Exception:
                        pass
                if matched:
                    results.append("[인사이트]\n" + "\n\n".join(matched))

            # 2. corpus/
            corpus_dir = self.knowledge_dir / 'corpus'
            if corpus_dir.exists():
                corpus_index = corpus_dir / 'index.json'
                if corpus_index.exists():
                    try:
                        ci = json.loads(corpus_index.read_text(encoding='utf-8'))
                        themes = [
                            f"'{t}': {len(c.get('entry_ids', []))}개"
                            for t, c in ci.get('clusters', {}).items()
                            if any(kw in t.lower() for kw in keywords)
                        ]
                        if themes:
                            results.append("[Corpus]\n" + "\n".join(themes))
                    except Exception:
                        pass

            # 3. long_term_memory
            lm_path = self.knowledge_dir / 'long_term_memory.json'
            if lm_path.exists():
                try:
                    lm = json.loads(lm_path.read_text(encoding='utf-8'))
                    matched_c = [f"{k}({v}회)" for k, v in lm.get('concepts', {}).items()
                                 if any(kw in k.lower() for kw in keywords)]
                    matched_e = [f"- {e.get('summary', '')[:100]}" for e in reversed(lm.get('experiences', []))
                                 if any(kw in e.get('summary', '').lower() for kw in keywords)][:4]
                    if matched_c or matched_e:
                        lm_text = ("개념: " + ", ".join(matched_c[:6]) + "\n" if matched_c else "")
                        lm_text += "\n".join(matched_e)
                        results.append(f"[기억]\n{lm_text}")
                except Exception:
                    pass

            # 4. directives/
            for md in sorted(self.directives_dir.glob('**/*.md')):
                try:
                    content = md.read_text(encoding='utf-8')
                    if any(kw in content.lower() for kw in keywords):
                        results.append(f"[{md.name}]\n{content[:400]}")
                        if len(results) >= 5:
                            break
                except Exception:
                    pass

            return "\n\n".join(results) if results else ""

        except Exception as e:
            logger.error("지식 검색 오류: %s", e)
            return ""

    # ─── 컨텍스트 로더 ───────────────────────────────────────────

    def _load_recent_signals(self, limit: int = 8) -> str:
        try:
            signals_dir = self.knowledge_dir / 'signals'
            files = sorted(signals_dir.glob('**/*.json'), key=lambda p: p.stat().st_mtime, reverse=True)[:limit]
            lines = []
            for f in files:
                try:
                    data = json.loads(f.read_text(encoding='utf-8'))
                    date = data.get('captured_at', '')[:10]
                    stype = data.get('type', '')
                    content = data.get('content', data.get('transcript', ''))[:150]
                    lines.append(f"[{date}][{stype}] {content}")
                except Exception:
                    pass
            return '\n'.join(lines)
        except Exception:
            return ""

    def _load_long_term_memory(self) -> str:
        try:
            lm_path = self.knowledge_dir / 'long_term_memory.json'
            if not lm_path.exists():
                return ""
            data = json.loads(lm_path.read_text(encoding='utf-8'))
            parts = []
            experiences = data.get('experiences', [])
            if experiences:
                recent = experiences[-4:]
                parts.append("경험: " + " | ".join(e.get('summary', '')[:60] for e in recent))
            concepts = data.get('concepts', {})
            if concepts:
                top = sorted(concepts.items(), key=lambda x: x[1], reverse=True)[:8]
                parts.append("개념: " + ", ".join(f"{k}({v})" for k, v in top))
            retro = data.get('retrospective', {})
            if retro:
                counts = retro.get('decision_counts', {})
                if counts:
                    parts.append("결정 패턴: " + str(counts))
            return '\n'.join(parts)
        except Exception:
            return ""

    # ─── 장기 기억 업데이트 ──────────────────────────────────────

    def _update_long_term_memory(self, user_message: str, assistant_answer: str) -> None:
        if not self.model:
            return

        lm_path = self.knowledge_dir / 'long_term_memory.json'
        try:
            data = json.loads(lm_path.read_text(encoding='utf-8')) if lm_path.exists() else {
                'metadata': {'created_at': datetime.now().isoformat(), 'total_entries': 0},
                'experiences': [],
                'concepts': {},
            }
        except Exception:
            return

        extract_prompt = f"""다음 대화에서 순호의 관심사, 패턴, 핵심 개념을 추출하라.

사용자: {user_message[:250]}
응답: {assistant_answer[:250]}

JSON만 출력:
{{"concepts": ["개념1", "개념2"], "summary": "한 문장 요약", "category": "브랜드/개인/기술/비즈니스/라이프스타일"}}"""

        try:
            import re
            resp = self.client.models.generate_content(model=self._model_name, contents=[extract_prompt])
            m = re.search(r'\{.*\}', resp.text, re.DOTALL)
            if not m:
                return
            extracted = json.loads(m.group())
        except Exception:
            return

        for concept in extracted.get('concepts', []):
            c = concept.strip()
            if c:
                data['concepts'][c] = data['concepts'].get(c, 0) + 1

        data['experiences'].append({
            'summary': extracted.get('summary', '')[:100],
            'category': extracted.get('category', 'unknown'),
            'timestamp': datetime.now().isoformat()[:16],
        })
        if len(data['experiences']) > 100:
            data['experiences'] = data['experiences'][-100:]

        data['metadata']['total_entries'] = len(data['experiences'])
        data['metadata']['last_updated'] = datetime.now().isoformat()[:16]

        lm_path.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding='utf-8')
        logger.debug("장기 기억 업데이트: +%d개 개념", len(extracted.get('concepts', [])))

    # ─── 유틸 ────────────────────────────────────────────────────

    def clear_context(self, user_id: str) -> None:
        if user_id in self.user_contexts:
            del self.user_contexts[user_id]
            self._save_context(user_id)
            logger.info("컨텍스트 초기화: %s", user_id)


# Singleton
_engine: Optional[ConversationEngine] = None


def get_conversation_engine() -> ConversationEngine:
    global _engine
    if _engine is None:
        _engine = ConversationEngine()
    return _engine
