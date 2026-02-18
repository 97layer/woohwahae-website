#!/usr/bin/env python3
"""
97layerOS Chief Editor (CE) Agent
Phase 6.3: NotebookLM 브랜드 가이드 쿼리 연동

Role:
- Content synthesis and editorial direction
- Transform insights + visuals into cohesive narratives
- Brand voice consistency — NotebookLM RAG 기반 실시간 참조
- Final content output (copy, captions, articles)

LLM: Gemini 2.5 Pro (Free tier)
Brand Reference: NotebookLM MCP (쿠키 인증 필요, 없으면 fallback)
Queue: Autonomous task claiming via AgentWatcher
Output: Finalized content ready for CD approval

Author: 97layerOS Technical Director
Updated: 2026-02-16 (Phase 6.3 — NotebookLM 브랜드 연동)
"""

import os
import sys
import json
import logging
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

# Project setup
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.agent_watcher import AgentWatcher
from core.system.queue_manager import Task

try:
    import google.genai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False

logger = logging.getLogger(__name__)

# 브랜드 보이스 fallback (NotebookLM 연결 불가 시 사용)
_BRAND_VOICE_FALLBACK = """
97layer 브랜드 보이스 (WOOHWAHAE 슬로우 라이프 아틀리에):
- 톤: 사색적, 느리고 깊은 호흡, 과도한 흥분 없음
- 금지어: 혁신, 가속, 트렌드, 최신, 최고, 압도적
- 허용어: 본질, 느림, 깊이, 일상, 사유, 여백, 단단함
- 문장 구조: 짧고 명료. 단문 위주. 50-100자 이내.
- 결말 선호: 질문으로 끝내기 (해답 제시가 아닌 탐색 권유)
- 시제: 현재형 중심. "~이다" 보다 "~일 수 있다"
- 인칭: 2인칭(당신) 지양, 보편적 1인칭("우리는", "나는")
"""


class ChiefEditor:
    """
    Chief Editor Agent - Content Synthesis & Editorial Direction

    Capabilities:
    - Synthesize SA analysis + AD visuals into cohesive content
    - Write copy in 97layer brand voice (NotebookLM RAG 참조)
    - Create social media captions, articles, newsletters
    - Editorial QA before CD review
    """

    def __init__(self, agent_id: str = "ce-worker-1", api_key: Optional[str] = None):
        self.agent_id = agent_id
        self.agent_type = "CE"
        self._brand_voice_cache: Optional[str] = None

        if not GEMINI_AVAILABLE:
            raise ImportError("google-generativeai required")

        api_key = api_key or os.getenv('GOOGLE_API_KEY')
        if not api_key:
            raise ValueError("GOOGLE_API_KEY not found")

        self.client = genai.Client(api_key=api_key)
        self._model_name = 'gemini-2.5-pro'

        # NotebookLM 브릿지 (선택적 — 없어도 동작)
        self.nlm = None
        try:
            from core.bridges.notebooklm_bridge import get_bridge, is_available
            if is_available():
                self.nlm = get_bridge()
                print(f"✅ {self.agent_id}: NotebookLM 브랜드 RAG 연결됨")
            else:
                print(f"⚠️  {self.agent_id}: NotebookLM 미연결 — fallback 브랜드 보이스 사용")
        except Exception as e:
            logger.warning("NotebookLM 초기화 실패: %s", e)

        print(f"Ray: 준비됨.")

    def _get_brand_voice(self) -> str:
        """
        NotebookLM에서 97layer 브랜드 보이스 참조 가져오기.
        세션 내 첫 호출 시만 쿼리, 이후 캐시 사용.
        NotebookLM 연결 불가 시 fallback 반환.
        """
        if self._brand_voice_cache:
            return self._brand_voice_cache

        if self.nlm:
            try:
                logger.info("%s: NotebookLM 브랜드 보이스 쿼리 중...", self.agent_id)
                result = self.nlm.query_knowledge_base(
                    "97layer 브랜드 보이스와 WOOHWAHAE 톤앤매너 가이드. "
                    "금지어, 허용어, 문장 스타일, 인칭, 시제 규칙을 요약해줘."
                )
                if result and len(result) > 50:
                    self._brand_voice_cache = result
                    logger.info("%s: NotebookLM 브랜드 보이스 캐시 완료 (%d자)", self.agent_id, len(result))
                    return self._brand_voice_cache
            except Exception as e:
                logger.warning("%s: NotebookLM 쿼리 실패, fallback 사용: %s", self.agent_id, e)

        self._brand_voice_cache = _BRAND_VOICE_FALLBACK
        return self._brand_voice_cache

    def write_content(self, analysis: Dict[str, Any], visual_concept: Dict[str, Any],
                      retry_count: int = 0, feedback: str = "", previous_output: Dict = None) -> Dict[str, Any]:
        """
        SA 분석 + AD 비주얼 컨셉을 기반으로 콘텐츠 작성.
        인스타그램 패키지 + 아카이브 에세이 이중 포맷 생성.

        Args:
            analysis: SA strategic analysis
            visual_concept: AD visual concept
            retry_count: 재작업 횟수 (CD 거절 또는 Ralph 점수 미달)
            feedback: 이전 거절 피드백 (재작업 시)
            previous_output: 이전 결과물 (재작업 참고용)

        Returns:
            {
              instagram_caption, hashtags, archive_essay,
              headline, tone, ...
            }
        """
        signal_id = analysis.get('signal_id', 'unknown')
        print(f"Ray: {signal_id} 초안 작업." + (f" (재작업 {retry_count}회차)" if retry_count > 0 else ""))

        # 브랜드 보이스 참조 (NotebookLM 또는 fallback)
        brand_voice = self._get_brand_voice()
        brand_source = "NotebookLM RAG" if self.nlm and self._brand_voice_cache != _BRAND_VOICE_FALLBACK else "fallback"
        logger.info("%s: 브랜드 보이스 출처 — %s", self.agent_id, brand_source)

        # 재작업 컨텍스트
        retry_context = ""
        if retry_count > 0 and feedback:
            retry_context = f"""
**이전 피드백 (반드시 반영):**
{feedback}

"""
        if previous_output:
            retry_context += f"""**이전 출력 (개선 필요):**
- 인스타 캡션: {previous_output.get('instagram_caption', 'N/A')[:100]}
- 에세이 일부: {str(previous_output.get('archive_essay', 'N/A'))[:200]}

"""

        prompt = f"""당신은 97layer의 Chief Editor Ray입니다.
WOOHWAHAE 슬로우 라이프 아틀리에의 브랜드 목소리로 콘텐츠를 작성합니다.

**전략 분석 (SA 제공):**
- 주제: {', '.join(analysis.get('themes', []))}
- 핵심 인사이트: {'; '.join(analysis.get('key_insights', []))}
- 요약: {analysis.get('summary', '')}

**비주얼 컨셉 (AD 제공):**
- 제목: {visual_concept.get('concept_title', '(없음)')}
- 무드: {visual_concept.get('visual_mood', '(없음)')}
- 브랜드 정렬: {visual_concept.get('brand_alignment', '(없음)')}

**97layer 브랜드 보이스 가이드:**
{brand_voice}

{retry_context}
위 가이드를 철저히 따라 **두 가지 포맷**으로 콘텐츠를 작성하세요:

1. **Instagram 패키지**: 발행 즉시 사용 가능한 형태
2. **Archive Essay**: Notion/블로그용 롱폼 에세이

아래 JSON 형식으로 반환하세요:
{{
  "instagram_caption": "인스타그램 캡션 (한국어, 150자 이내, 브랜드 톤 철저히 준수, 이모지 최소화)",
  "hashtags": "#woohwahae #slowlife #아카이브 (관련 한국어 해시태그 5-8개)",
  "archive_essay": "아카이브 에세이 (한국어, 500-800자, 사색적 롱폼, 단락 구분 포함, 느리고 깊은 톤)",
  "headline": "헤드라인 (10-20자)",
  "tone": "contemplative|reflective|grounded 중 하나",
  "brand_voice_source": "{brand_source}"
}}

필수 준수 사항:
- instagram_caption: 반드시 150자 이내. 직관적이고 핵심만.
- hashtags: #woohwahae 반드시 포함
- archive_essay: 반드시 500자 이상. 질문으로 마무리 권장.
- 금지어: 혁신, 트렌드, 최신, 혁명적, 압도적

유효한 JSON만 반환하세요.
"""

        try:
            response = self.client.models.generate_content(
                model=self._model_name,
                contents=[prompt]
            )
            content_text = response.text

            # JSON 파싱
            if '```json' in content_text:
                json_start = content_text.find('```json') + 7
                json_end = content_text.find('```', json_start)
                json_text = content_text[json_start:json_end].strip()
            elif '```' in content_text:
                json_start = content_text.find('```') + 3
                json_end = content_text.find('```', json_start)
                json_text = content_text[json_start:json_end].strip()
            else:
                json_text = content_text.strip()

            content = json.loads(json_text)

            content.update({
                'signal_id': signal_id,
                'written_by': self.agent_id,
                'written_at': datetime.now().isoformat(),
                'model': self._model_name,
                'brand_voice_source': brand_source,
                'retry_count': retry_count,
                'status': 'draft_for_cd',
            })

            caption_len = len(content.get('instagram_caption', ''))
            essay_len = len(content.get('archive_essay', ''))
            print(f"Ray: 초안 완료. 캡션 {caption_len}자, 에세이 {essay_len}자.")
            return content

        except Exception as e:
            logger.error("%s: 콘텐츠 생성 실패: %s", self.agent_id, e)
            return {'signal_id': signal_id, 'error': str(e), 'status': 'failed'}

    def process_task(self, task: Task) -> Dict[str, Any]:
        task_type = task.task_type
        payload = task.payload

        print(f"Ray: {task.task_id} ({task_type})")

        if task_type == 'write_content':
            # Orchestrator에서 오는 새 payload 구조 지원
            # payload에 sa_result가 있으면 Orchestrator 경유
            sa_result = payload.get('sa_result', payload.get('analysis', {}))
            visual = payload.get('visual_concept', payload.get('ad_result', {}).get('visual_concept', {}))

            # 재작업 파라미터
            retry_count = payload.get('retry_count', 0)
            feedback = payload.get('feedback', payload.get('cd_feedback', ''))
            previous_output = payload.get('previous_output', None)

            result = self.write_content(
                analysis=sa_result,
                visual_concept=visual,
                retry_count=retry_count,
                feedback=feedback,
                previous_output=previous_output
            )
            # SA 전략 점수를 result에 포함 (Ralph 채점용)
            result['sa_strategic_score'] = sa_result.get('strategic_score', 0)
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}

        elif task_type == 'write_corpus_essay':
            # Gardener가 트리거한 corpus 기반 에세이 작성
            # 단일 신호가 아닌 군집 전체 RAG → Magazine B 스타일 롱폼
            result = self._write_corpus_essay(payload)
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}

        else:
            return {'status': 'failed', 'error': f"Unknown task type: {task_type}"}

    def _write_corpus_essay(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """
        Corpus 군집 기반 원소스 멀티유즈 콘텐츠 생성.

        하나의 Gemini 호출로 5개 포맷 동시 생성:
          1. archive_essay     — 롱폼 에세이 (800~1200자) → woohwahae.kr/archive/
          2. instagram_caption — 캡션 (150자 이내) → 인스타그램 단일 피드
          3. carousel_slides   — 3~5장 텍스트 슬라이드 → 인스타그램 캐러셀
          4. telegram_summary  — 3줄 요약 → 봇 푸시 알림
          5. pull_quote        — 핵심 1문장 → 웹사이트 히어로/about 인용구

        원칙: 포맷이 다를 뿐 동일한 본질에서 파생. 재가공 아닌 파생.
        """
        theme = payload.get("theme", "")
        rag_context = payload.get("rag_context", [])
        entry_count = payload.get("entry_count", 0)
        instruction = payload.get("instruction", "")

        # RAG 컨텍스트 직렬화
        context_text = ""
        for i, entry in enumerate(rag_context, 1):
            context_text += f"\n[{i}] {entry.get('captured_at', '')[:10]} | {entry.get('signal_type', '')}\n"
            context_text += f"요약: {entry.get('summary', '')}\n"
            insights = entry.get('key_insights', [])
            if insights:
                context_text += f"인사이트: {' / '.join(str(x) for x in insights[:3])}\n"

        prompt = f"""너는 WOOHWAHAE의 편집장이다.

주제: {theme}
신호 수: {entry_count}개

아래는 이 주제와 관련해 시간을 두고 쌓인 신호들의 요약이다:
{context_text}

이 신호들의 흐름에서 본질을 읽어내고, 아래 5개 포맷을 동시에 만들어라.
모두 같은 본질에서 파생된다. 재가공이 아닌 파생이다.

공통 규칙:
- 한국어
- 이모지 완전 금지
- 볼드, 헤더 사용 금지
- WOOHWAHAE 톤: 절제, 사색, 여백. 감탄사 없음.

응답 형식 (JSON):
{{
  "essay_title": "제목 (10자 이내, 명사형)",
  "pull_quote": "이 글 전체를 관통하는 핵심 문장 1개 (30자 이내). 웹사이트 히어로에 써도 될 만큼 밀도 있게.",
  "archive_essay": "롱폼 에세이. 800~1200자. 도입(관찰) → 전개(맥락) → 마무리(열린 질문 또는 여백). 단락 사이 빈 줄.",
  "instagram_caption": "인스타그램 캡션. 150자 이내. 에세이의 핵심을 압축. 마지막 줄은 여백을 주는 한 문장.",
  "carousel_slides": [
    "슬라이드 1: 도입 문장 (30자 이내)",
    "슬라이드 2: 핵심 관찰 (30자 이내)",
    "슬라이드 3: 맥락 또는 역설 (30자 이내)",
    "슬라이드 4: 마무리 또는 질문 (30자 이내)"
  ],
  "telegram_summary": "봇 푸시 알림용 3줄 요약. 각 줄 40자 이내. 첫 줄: 제목. 둘째 줄: 핵심. 셋째 줄: 링크 유도.",
  "theme": "{theme}",
  "entry_count": {entry_count}
}}

JSON만 출력."""

        try:
            import google.genai as genai
            import os, re
            client = genai.Client(api_key=os.getenv('GOOGLE_API_KEY'))
            response = client.models.generate_content(
                model='gemini-2.5-pro',
                contents=[prompt]
            )
            text = response.text.strip()
            match = re.search(r'\{.*\}', text, re.DOTALL)
            if match:
                result = json.loads(match.group())
                formats = [k for k in ['archive_essay', 'instagram_caption', 'carousel_slides',
                                        'telegram_summary', 'pull_quote'] if k in result]
                print(f"Ray: 원소스 멀티유즈 완료 — {theme} | 포맷: {', '.join(formats)}")
                return result
            else:
                return {
                    "archive_essay": text,
                    "essay_title": theme,
                    "theme": theme,
                    "entry_count": entry_count
                }
        except Exception as e:
            print(f"Ray: corpus 에세이 실패 — {e}")
            return {"error": str(e), "theme": theme}

    def start_watching(self, interval: int = 5):
        watcher = AgentWatcher(agent_type=self.agent_type, agent_id=self.agent_id)
        nlm_status = "연결됨" if self.nlm else "fallback"
        print(f"Ray: 큐 감시 시작.")
        print(f"   LLM: Gemini 2.5 Pro")
        print(f"   Brand Voice: NotebookLM RAG ({nlm_status})")
        print(f"   Tasks: write_content")
        print()
        watcher.watch(callback=self.process_task, interval=interval)


if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser(description='97layerOS Chief Editor Agent')
    parser.add_argument('--agent-id', default='ce-worker-1')
    parser.add_argument('--interval', type=int, default=5)
    parser.add_argument('--test', action='store_true')
    args = parser.parse_args()

    agent = ChiefEditor(agent_id=args.agent_id)

    if args.test:
        print("\n🧪 Test Mode: Content Writing\n" + "=" * 50)
        test_analysis = {
            'signal_id': 'test_001',
            'themes': ['AI와 창작', '느린 삶'],
            'key_insights': ['AI는 반복을 제거하고 창작에 집중하게 한다', '속도보다 깊이가 더 오래 남는다'],
            'summary': 'AI는 슬로우 라이프를 가능하게 하는 도구다',
        }
        test_visual = {
            'concept_title': '디지털 정원',
            'visual_mood': 'contemplative',
            'brand_alignment': '여백과 느림의 미학',
        }

        result = agent.write_content(test_analysis, test_visual)
        print(f"\n✍️  콘텐츠 초안:")
        print(f"   헤드라인: {result.get('headline', 'N/A')}")
        print(f"   캡션: {result.get('social_caption', 'N/A')}")
        print(f"   브랜드 보이스 출처: {result.get('brand_voice_source', 'N/A')}")
        print("\n✅ 테스트 완료!")
    else:
        agent.start_watching(interval=args.interval)
