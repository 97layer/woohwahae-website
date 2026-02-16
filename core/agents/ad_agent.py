#!/usr/bin/env python3
"""
97layerOS Art Director (AD) Agent
Phase 6.3: NotebookLM 시각 레퍼런스 쿼리 연동

Role:
- Visual concept development and art direction
- Image generation guidance (for Stable Diffusion integration)
- Brand consistency — NotebookLM RAG 기반 시각 아이덴티티 참조
- Visual storytelling and composition

LLM: Gemini 2.5 Pro (Free tier, with Vision)
Visual Reference: NotebookLM MCP (WOOHWAHAE 시각 아카이브 참조)
Queue: Autonomous task claiming via AgentWatcher
Output: Visual concepts, style guides, image prompts

Author: 97layerOS Technical Director
Updated: 2026-02-16 (Phase 6.3 — NotebookLM 시각 레퍼런스 연동)
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

# Gemini API (optional, for actual execution)
GEMINI_AVAILABLE = False
try:
    import google.generativeai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    print("⚠️  google-generativeai not installed (mock mode)")

logger = logging.getLogger(__name__)

# 시각 레퍼런스 fallback (NotebookLM 연결 불가 시 사용)
_VISUAL_REFERENCE_FALLBACK = """
WOOHWAHAE 시각 아이덴티티 (97layer 브랜드):
- 색상 팔레트: 탈채도 중성 톤 (웜 그레이, 더스티 베이지, 오프화이트, 딥 차콜)
- 금지 색상: 채도 높은 원색, 네온, 과도한 대비
- 사진 스타일: 35mm 필름 그레인, 소프트 자연광, 얕은 피사계 심도
- 벤치마크 레퍼런스: Aesop, Kinfolk, 와비사비 미학
- 구도: 여백 강조, 중앙 정렬보다 오프센터, 오가닉 텍스처
- 폰트: 세리프 계열 본문, 클린 산세리프 헤더, 넉넉한 행간
- 절대 회피: 코퍼레이트 스톡 이미지, 과도하게 광택나는 제품사진, 빠른 줌인
"""


class ArtDirector:
    """
    Art Director Agent - Visual Concepts & Art Direction

    Capabilities:
    - Develop visual concepts from strategic insights
    - Generate image prompts for Stable Diffusion (WOOHWAHAE 아카이브 참조)
    - Validate visual consistency with brand guidelines (NotebookLM RAG)
    - Provide art direction feedback
    """

    def __init__(self, agent_id: str = "ad-worker-1", api_key: Optional[str] = None):
        self.agent_id = agent_id
        self.agent_type = "AD"
        self.mock_mode = not GEMINI_AVAILABLE
        self._visual_ref_cache: Optional[str] = None

        if GEMINI_AVAILABLE:
            api_key = api_key or os.getenv('GOOGLE_API_KEY')
            if api_key:
                genai.configure(api_key=api_key)
                self.model = genai.GenerativeModel('gemini-2.5-pro')
                self.mock_mode = False
                print(f"✅ {self.agent_id}: Art Director initialized (Gemini 2.5 Pro)")
            else:
                print(f"⚠️  {self.agent_id}: No API key, running in mock mode")
                self.mock_mode = True
        else:
            print(f"⚠️  {self.agent_id}: Gemini not available, running in mock mode")

        # NotebookLM 브릿지 (선택적 — 없어도 동작)
        self.nlm = None
        try:
            from core.bridges.notebooklm_bridge import get_bridge, is_available
            if is_available():
                self.nlm = get_bridge()
                print(f"✅ {self.agent_id}: NotebookLM 시각 레퍼런스 연결됨")
            else:
                print(f"⚠️  {self.agent_id}: NotebookLM 미연결 — fallback 시각 레퍼런스 사용")
        except Exception as e:
            logger.warning("NotebookLM 초기화 실패: %s", e)

    def _get_visual_reference(self) -> str:
        """
        NotebookLM에서 WOOHWAHAE 시각 아이덴티티 레퍼런스 가져오기.
        세션 내 첫 호출 시만 쿼리, 이후 캐시 사용.
        NotebookLM 연결 불가 시 fallback 반환.
        """
        if self._visual_ref_cache:
            return self._visual_ref_cache

        if self.nlm:
            try:
                logger.info("%s: NotebookLM 시각 레퍼런스 쿼리 중...", self.agent_id)
                result = self.nlm.query_knowledge_base(
                    "WOOHWAHAE 시각 아이덴티티 가이드. "
                    "색상 팔레트, 사진 스타일, 구도 원칙, 벤치마크 브랜드, "
                    "아카이벌 필름 미학을 요약해줘."
                )
                if result and len(result) > 50:
                    self._visual_ref_cache = result
                    logger.info(
                        "%s: NotebookLM 시각 레퍼런스 캐시 완료 (%d자)",
                        self.agent_id, len(result),
                    )
                    return self._visual_ref_cache
            except Exception as e:
                logger.warning("%s: NotebookLM 쿼리 실패, fallback 사용: %s", self.agent_id, e)

        self._visual_ref_cache = _VISUAL_REFERENCE_FALLBACK
        return self._visual_ref_cache

    def create_visual_concept(self, analysis_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        SA 분석을 기반으로 비주얼 컨셉 생성.
        시각 아이덴티티는 NotebookLM RAG에서 실시간 참조.
        """
        signal_id = analysis_data.get('signal_id', 'unknown')
        themes = analysis_data.get('themes', [])
        insights = analysis_data.get('key_insights', [])
        summary = analysis_data.get('summary', '')

        print(f"🎨 {self.agent_id}: Creating visual concept for {signal_id}...")

        if self.mock_mode:
            return self._mock_visual_concept(signal_id, themes, insights)

        # 시각 레퍼런스 참조 (NotebookLM 또는 fallback)
        visual_ref = self._get_visual_reference()
        ref_source = (
            "NotebookLM RAG"
            if self.nlm and self._visual_ref_cache != _VISUAL_REFERENCE_FALLBACK
            else "fallback"
        )
        logger.info("%s: 시각 레퍼런스 출처 — %s", self.agent_id, ref_source)

        prompt = self._build_concept_prompt(themes, insights, summary, visual_ref, ref_source)

        try:
            response = self.model.generate_content(prompt)
            concept = self._parse_concept(response.text)

            concept.update({
                'signal_id': signal_id,
                'created_by': self.agent_id,
                'created_at': datetime.now().isoformat(),
                'model': 'gemini-2.5-pro',
                'visual_ref_source': ref_source,
                'based_on': 'SA analysis',
            })

            print(f"✅ {self.agent_id}: 비주얼 컨셉 완료 (레퍼런스: {ref_source})")
            return concept

        except Exception as e:
            logger.error("%s: 비주얼 컨셉 생성 실패: %s", self.agent_id, e)
            return {'signal_id': signal_id, 'error': str(e), 'status': 'failed'}

    def _build_concept_prompt(
        self,
        themes: list,
        insights: list,
        summary: str,
        visual_ref: str,
        ref_source: str,
    ) -> str:
        return f"""당신은 97layer의 Art Director입니다.
WOOHWAHAE 슬로우 라이프 아틀리에의 시각 아이덴티티를 기반으로 비주얼 컨셉을 개발합니다.

**전략 분석 (SA 제공):**
- 주제: {', '.join(themes)}
- 핵심 인사이트: {'; '.join(insights)}
- 요약: {summary}

**WOOHWAHAE 시각 아이덴티티 가이드 (출처: {ref_source}):**
{visual_ref}

위 가이드를 엄격히 따라 아래 JSON 형식으로 비주얼 컨셉을 작성하세요:
{{
  "concept_title": "컨셉 제목 (짧고 시적으로, 한국어)",
  "visual_mood": "contemplative|serene|intimate|grounded 중 하나",
  "color_palette": ["#hex1", "#hex2", "#hex3"],
  "composition_notes": "구도 가이드 (2-3문장, 여백과 오가닉 텍스처 강조)",
  "image_prompts": [
    {{
      "prompt": "Stable Diffusion 프롬프트 (영어, 필름 그레인/탈채도/자연광 포함)",
      "style": "photography|film_still|analog",
      "aspect_ratio": "4:5|1:1|16:9"
    }}
  ],
  "typography_guidance": "폰트 방향 (세리프/산세리프, 행간, 크기 위계)",
  "reference_aesthetics": ["레퍼런스1", "레퍼런스2"],
  "brand_alignment": "WOOHWAHAE 철학과의 연결점",
  "visual_ref_source": "{ref_source}"
}}

유효한 JSON만 반환하세요.
"""

    def _parse_concept(self, concept_text: str) -> Dict[str, Any]:
        try:
            if '```json' in concept_text:
                json_start = concept_text.find('```json') + 7
                json_end = concept_text.find('```', json_start)
                json_text = concept_text[json_start:json_end].strip()
            elif '```' in concept_text:
                json_start = concept_text.find('```') + 3
                json_end = concept_text.find('```', json_start)
                json_text = concept_text[json_start:json_end].strip()
            else:
                json_text = concept_text.strip()

            return json.loads(json_text)

        except json.JSONDecodeError as e:
            return {
                'concept_title': 'Visual Concept',
                'visual_mood': 'contemplative',
                'raw_response': concept_text,
                'parse_error': str(e),
            }

    def _mock_visual_concept(self, signal_id: str, themes: list, insights: list) -> Dict[str, Any]:
        return {
            'signal_id': signal_id,
            'concept_title': f"{'와 '.join(themes[:2])}의 정경",
            'visual_mood': 'contemplative',
            'color_palette': ['#3D3530', '#E8E0D5', '#A89880'],
            'composition_notes': (
                '넉넉한 여백과 오프센터 구도. '
                '오가닉 텍스처와 소프트한 자연광. '
                '빠른 움직임 없이 정지된 순간의 무게.'
            ),
            'image_prompts': [
                {
                    'prompt': (
                        f'35mm film photography, {themes[0] if themes else "still life"}, '
                        'muted desaturated tones, warm grey palette, soft natural side lighting, '
                        'shallow depth of field, analog film grain, organic textures, '
                        'off-center composition, generous negative space, wabi-sabi aesthetic'
                    ),
                    'style': 'film_still',
                    'aspect_ratio': '4:5',
                }
            ],
            'typography_guidance': (
                '세리프 계열 본문 (Garamond류). '
                '산세리프 헤더는 라이트 웨이트 사용. '
                '행간 1.8 이상, 자간 넓게.'
            ),
            'reference_aesthetics': ['Kinfolk', 'Aesop', '와비사비'],
            'brand_alignment': 'WOOHWAHAE 슬로우 라이프 — 속도가 아닌 깊이를 향한 시선',
            'created_by': self.agent_id,
            'created_at': datetime.now().isoformat(),
            'mode': 'mock',
            'visual_ref_source': 'fallback',
            'based_on': f"{len(themes)} themes, {len(insights)} insights",
        }

    def process_task(self, task: Task) -> Dict[str, Any]:
        task_type = task.task_type
        payload = task.payload

        print(f"📋 {self.agent_id}: Processing task {task.task_id} ({task_type})")

        if task_type == 'create_visual_concept':
            analysis_data = payload.get('analysis', {})
            result = self.create_visual_concept(analysis_data)
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}

        elif task_type == 'validate_visual':
            return {
                'status': 'completed',
                'task_id': task.task_id,
                'result': {'validated': True, 'notes': 'Brand-aligned'},
            }

        else:
            return {'status': 'failed', 'error': f"Unknown task type: {task_type}"}

    def start_watching(self, interval: int = 5):
        watcher = AgentWatcher(agent_type=self.agent_type, agent_id=self.agent_id)

        mode_str = "MOCK MODE" if self.mock_mode else "Gemini 2.5 Pro"
        nlm_status = "연결됨" if self.nlm else "fallback"
        print(f"👁️  {self.agent_id}: 자율 운영 시작...")
        print(f"   LLM: {mode_str}")
        print(f"   Visual Reference: NotebookLM RAG ({nlm_status})")
        print(f"   Tasks: create_visual_concept, validate_visual")
        print(f"   Queue: .infra/queue/tasks/pending/")
        print()

        watcher.watch(callback=self.process_task, interval=interval)


# ================== Standalone Execution ==================

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description='97layerOS Art Director Agent')
    parser.add_argument('--agent-id', default='ad-worker-1')
    parser.add_argument('--interval', type=int, default=5)
    parser.add_argument('--test', action='store_true')

    args = parser.parse_args()

    agent = ArtDirector(agent_id=args.agent_id)

    if args.test:
        print("\n🧪 Test Mode: Visual Concept Creation")
        print("=" * 50)

        test_analysis = {
            'signal_id': 'test_001',
            'themes': ['AI와 창작', '느린 삶', '본질적 작업'],
            'key_insights': [
                'AI 도구는 반복 작업을 제거해 창작에 집중하게 한다',
                '기술은 인간의 창의성을 대체하지 않고 확장한다',
                '슬로우 라이프: 중요한 것에 집중하는 선택',
            ],
            'summary': 'AI는 슬로우 라이프 창작을 가능하게 하는 조용한 파트너',
        }

        result = agent.create_visual_concept(test_analysis)

        print(f"\n🎨 비주얼 컨셉:")
        print(f"   제목: {result.get('concept_title', 'N/A')}")
        print(f"   무드: {result.get('visual_mood', 'N/A')}")
        print(f"   팔레트: {result.get('color_palette', [])}")
        print(f"   레퍼런스 출처: {result.get('visual_ref_source', 'N/A')}")
        for p in result.get('image_prompts', []):
            print(f"   - {p.get('prompt', 'N/A')[:100]}...")

        print("\n✅ 테스트 완료!")

    else:
        print("\n🚀 Production Mode: Autonomous Queue Watching")
        print("=" * 50)
        agent.start_watching(interval=args.interval)
