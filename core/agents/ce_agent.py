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

        print(f"✅ {self.agent_id}: Chief Editor initialized (Gemini 2.5 Pro)")

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

    def write_content(self, analysis: Dict[str, Any], visual_concept: Dict[str, Any]) -> Dict[str, Any]:
        """
        SA 분석 + AD 비주얼 컨셉을 기반으로 콘텐츠 작성.
        브랜드 보이스는 NotebookLM RAG에서 실시간 참조.

        Args:
            analysis: SA strategic analysis
            visual_concept: AD visual concept

        Returns:
            Final content draft
        """
        signal_id = analysis.get('signal_id', 'unknown')
        print(f"✍️  {self.agent_id}: Writing content for {signal_id}...")

        # 브랜드 보이스 참조 (NotebookLM 또는 fallback)
        brand_voice = self._get_brand_voice()
        brand_source = "NotebookLM RAG" if self.nlm and self._brand_voice_cache != _BRAND_VOICE_FALLBACK else "fallback"
        logger.info("%s: 브랜드 보이스 출처 — %s", self.agent_id, brand_source)

        prompt = f"""당신은 97layer의 Chief Editor입니다.
WOOHWAHAE 슬로우 라이프 아틀리에의 브랜드 목소리로 콘텐츠를 작성합니다.

**전략 분석 (SA 제공):**
- 주제: {', '.join(analysis.get('themes', []))}
- 핵심 인사이트: {'; '.join(analysis.get('key_insights', []))}
- 요약: {analysis.get('summary', '')}

**비주얼 컨셉 (AD 제공):**
- 제목: {visual_concept.get('concept_title', '')}
- 무드: {visual_concept.get('visual_mood', '')}
- 브랜드 정렬: {visual_concept.get('brand_alignment', '')}

**97layer 브랜드 보이스 가이드:**
{brand_voice}

위 가이드를 철저히 따라 아래 JSON 형식으로 콘텐츠를 작성하세요:
{{
  "headline": "헤드라인 (한국어, 10-20자)",
  "subheadline": "서브헤드라인 (20-35자)",
  "body": "본문 (2-3단락, 각 2-3문장, 사색적 톤)",
  "social_caption": "인스타그램 캡션 (80자 이내, 해시태그 2-3개 포함)",
  "call_to_action": "독자에게 던지는 질문 또는 초대 (질문형 권장)",
  "tags": ["태그1", "태그2", "태그3"],
  "tone": "contemplative|reflective|grounded 중 하나",
  "brand_voice_source": "{brand_source}"
}}

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
                'status': 'draft_for_cd',
            })

            print(f"✅ {self.agent_id}: 콘텐츠 초안 완료 (브랜드 보이스: {brand_source})")
            return content

        except Exception as e:
            logger.error("%s: 콘텐츠 생성 실패: %s", self.agent_id, e)
            return {'signal_id': signal_id, 'error': str(e), 'status': 'failed'}

    def process_task(self, task: Task) -> Dict[str, Any]:
        task_type = task.task_type
        payload = task.payload

        print(f"📋 {self.agent_id}: Processing task {task.task_id} ({task_type})")

        if task_type == 'write_content':
            analysis = payload.get('analysis', {})
            visual = payload.get('visual_concept', {})
            result = self.write_content(analysis, visual)
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}
        else:
            return {'status': 'failed', 'error': f"Unknown task type: {task_type}"}

    def start_watching(self, interval: int = 5):
        watcher = AgentWatcher(agent_type=self.agent_type, agent_id=self.agent_id)
        nlm_status = "연결됨" if self.nlm else "fallback"
        print(f"👁️  {self.agent_id}: 자율 운영 시작...")
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
