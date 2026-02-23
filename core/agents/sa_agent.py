#!/usr/bin/env python3
"""
97layerOS Strategy Analyst (SA) Agent
Phase 6.2: Independent agent with Gemini Flash API

Role:
- Signal analysis and strategic insight extraction
- First responder to incoming signals
- Pattern recognition across signals
- Strategic recommendations

LLM: Gemini 2.5 Flash (Free tier, fast)
Queue: Autonomous task claiming via AgentWatcher
Output: Structured analysis → passed to AD/CE

Author: 97layerOS Technical Director
Updated: 2026-02-16 (google.genai SDK 마이그레이션)
"""

import os
import sys
import json
import time
import re
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

# Project setup
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.agent_watcher import AgentWatcher
from core.system.queue_manager import Task

# Using REST API directly - no SDK needed


class StrategyAnalyst:
    """
    Strategy Analyst Agent - Signal Analysis & Strategic Insights

    Capabilities:
    - Analyze incoming signals (text, links, images)
    - Extract strategic themes and patterns
    - Identify action items and opportunities
    - Generate initial insights for other agents
    """

    def __init__(self, agent_id: str = "sa-worker-1", api_key: Optional[str] = None):
        """
        Initialize Strategy Analyst

        Args:
            agent_id: Unique agent instance ID
            api_key: Google API key (or from env GOOGLE_API_KEY)
        """
        self.agent_id = agent_id
        self.agent_type = "SA"

        # Initialize Gemini API (REST API 직접 호출)
        self.api_key = api_key or os.getenv('GOOGLE_API_KEY')
        if not self.api_key:
            raise ValueError("GOOGLE_API_KEY not found in environment")

        self._model_name = 'gemini-2.5-flash'
        self._api_url = 'https://generativelanguage.googleapis.com/v1beta/models'

        # SA 에이전트 지침 로드
        self._persona = self._load_directive()

        print(f"SA: 준비됨. (Key: ...{self.api_key[-4:]})")

    def analyze_signal(self, signal_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Analyze incoming signal and extract strategic insights

        Args:
            signal_data: Signal content and metadata
                {
                    'signal_id': str,
                    'content': str,
                    'source': str (telegram, slack, etc.),
                    'timestamp': str,
                    'attachments': list (optional)
                }

        Returns:
            Analysis result with strategic insights
        """
        signal_id = signal_data.get('signal_id', 'unknown')
        content = signal_data.get('content', '')
        source = signal_data.get('source', 'unknown')

        print(f"SA: 신호 {signal_id} 분석 시작. [API: {self._api_url}]")

        # Construct prompt for strategic analysis
        prompt = self._build_analysis_prompt(content, source)

        try:
            # Call Gemini API (REST API 직접 호출)
            import requests
            api_url = f"{self._api_url}/{self._model_name}:generateContent"
            headers = {"Content-Type": "application/json"}
            payload = {"contents": [{"parts": [{"text": prompt}]}]}

            response = requests.post(f"{api_url}?key={self.api_key}",
                                    headers=headers, json=payload, timeout=60)
            response.raise_for_status()
            analysis_text = response.json()['candidates'][0]['content']['parts'][0]['text']

            # Parse structured response
            analysis = self._parse_analysis(analysis_text)

            # Add metadata
            analysis.update({
                'signal_id': signal_id,
                'analyzed_by': self.agent_id,
                'analyzed_at': datetime.now().isoformat(),
                'model': self._model_name,
                'source': source
            })

            score = analysis.get('strategic_score', 0)
            category = analysis.get('category', '')
            themes = ', '.join(analysis.get('themes', [])[:3])
            print(f"SA: 완료. 점수 {score}. 카테고리: {category}. 테마: {themes}.")

            # 자가발전: SA 분석 완료 → long_term_memory 피드백
            try:
                self._feedback_to_memory(analysis, content)
            except Exception as mem_e:
                print(f"⚠️  Memory feedback skipped: {mem_e}")

            # 자가발전: SA 분석 완료 → NotebookLM Signal Archive 저장
            try:
                self._save_to_notebooklm(analysis, content, source)
            except Exception as nlm_e:
                print(f"⚠️  NotebookLM skipped: {nlm_e}")

            # Corpus 누적: 신호 → 지식 풀 (군집 기반 발행을 위한 축적)
            try:
                from core.system.corpus_manager import CorpusManager
                corpus = CorpusManager()
                corpus.add_entry(signal_id, analysis, signal_data)
                print(f"SA: Corpus 누적 완료 → {signal_id}")
            except Exception as corpus_e:
                print(f"⚠️  Corpus skipped: {corpus_e}")

            # 신호 파일 status → analyzed (Orchestrator 중복 투입 방지)
            try:
                import pathlib
                signal_path = signal_data.get('signal_path') or str(
                    PROJECT_ROOT / 'knowledge' / 'signals' / f"{signal_id}.json"
                )
                sp = pathlib.Path(str(signal_path))
                if sp.exists():
                    sig_json = json.loads(sp.read_text())
                    sig_json['status'] = 'analyzed'
                    sig_json['analyzed_at'] = datetime.now().isoformat()
                    sp.write_text(json.dumps(sig_json, indent=2, ensure_ascii=False))
                    print(f"SA: 신호 status → analyzed: {signal_id}")
            except Exception as upd_e:
                print(f"⚠️  Signal status update skipped: {upd_e}")

            return analysis

        except Exception as e:
            print(f"SA: 분석 실패. {e}")
            return {
                'signal_id': signal_id,
                'error': str(e),
                'status': 'failed'
            }

    def _load_directive(self) -> str:
        """SA.md 에이전트 지침 로드"""
        persona_path = PROJECT_ROOT / 'directives' / 'agents' / 'SA.md'
        try:
            if persona_path.exists():
                return persona_path.read_text(encoding='utf-8')
        except Exception:
            pass
        return "냉정하게 분석하라. 필요한 것만 말해라."

    def _feedback_to_memory(self, analysis: Dict[str, Any], original_content: str):
        """
        SA 분석 완료 후 long_term_memory.json에 인사이트 누적 — 자가발전 고리

        themes → concepts 카운트 누적
        key_insights + summary → experiences 추가
        strategic_score 80+ 만 저장 (노이즈 필터)
        """
        score = analysis.get('strategic_score', 0)
        if score < 50:
            return  # 노이즈는 기록 안 함

        lm_path = PROJECT_ROOT / 'knowledge' / 'long_term_memory.json'

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

        # themes → concepts 카운트 (빈도 = 중요도)
        for theme in analysis.get('themes', []):
            theme = theme.strip()
            if theme:
                data['concepts'][theme] = data['concepts'].get(theme, 0) + 1

        # category 도 concepts에 반영
        category = analysis.get('category', '')
        if category and category not in ('noise', 'unknown'):
            data['concepts'][f'SA:{category}'] = data['concepts'].get(f'SA:{category}', 0) + 1

        # experiences: summary + top insight 기록
        insights = analysis.get('key_insights', [])
        insight_preview = insights[0][:80] if insights else ''
        summary = analysis.get('summary', '')[:100]
        combined = f"[SA score:{score}] {summary}" + (f" | {insight_preview}" if insight_preview else '')

        data['experiences'].append({
            'summary': combined[:150],
            'category': 'SA분석',
            'timestamp': datetime.now().isoformat()[:16],
            'source': 'sa_agent',
            'score': score
        })

        # 최근 100개 유지
        if len(data['experiences']) > 100:
            data['experiences'] = data['experiences'][-100:]

        data['metadata']['total_entries'] = len(data['experiences'])
        data['metadata']['last_updated'] = datetime.now().isoformat()[:16]

        lm_path.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding='utf-8')
        print(f"📝 Memory updated: score={score}, themes={analysis.get('themes', [])[:3]}")

    def _save_to_notebooklm(self, analysis: Dict[str, Any], content: str, source: str):
        """
        SA 분석 완료 후 NotebookLM Signal Archive에 저장.
        score 60+ 만 저장 (노이즈 제외).
        """
        score = analysis.get('strategic_score', 0)
        if score < 60:
            return

        try:
            from core.bridges.notebooklm_bridge import get_bridge, is_available
            if not is_available():
                return

            bridge = get_bridge()
            bridge.add_signal_to_archive({
                'signal_id': analysis.get('signal_id', ''),
                'content': content,
                'source': source,
                'analysis': analysis,
            })
            print(f"📚 NotebookLM 저장: score={score}")
        except ImportError:
            pass  # 브릿지 없으면 조용히 스킵

    def _build_analysis_prompt(self, content: str, source: str) -> str:
        """SA의 시각으로 신호 분석 프롬프트 구성"""
        return f"""너는 SA이다. 97layer 크루의 전략 분석가.

{self._persona[:600]}

---

다음 신호를 분석하라. 97layer의 방향과 얼마나 정렬되어 있는지 냉정하게 판단한다.

**신호 내용:**
{content}

**출처:** {source}

JSON 형식으로만 응답:
{{
  "strategic_score": <0-100, 97layer 방향 정렬도 + 전략 가치>,
  "category": "<trend|insight|opportunity|question|noise>",
  "themes": ["테마1", "테마2", "테마3"],
  "key_insights": [
    "인사이트 1 (한 문장, 직접적으로)",
    "인사이트 2",
    "인사이트 3"
  ],
  "action_items": [
    "구체적인 다음 행동 1",
    "구체적인 다음 행동 2"
  ],
  "recommended_agents": ["AD", "CE", "CD"],
  "summary": "한 문장 요약 — 핵심만"
}}

**판단 기준:**
- 80+: 97layer 방향과 정렬. 행동할 가치 있음.
- 50-79: 탐색할 만함. 패턴 주시.
- 50 미만: 노이즈. 알고리즘 추종 또는 방향 불일치.
- 슬로우라이프·본질·여백·진정성 관련 신호는 높게 평가.
- 빠른 트렌드·과도한 장식·알고리즘 추종 신호는 낮게 평가.

JSON만 출력. 설명 없이.
"""

    def _parse_analysis(self, analysis_text: str) -> Dict[str, Any]:
        """Parse Gemini response into structured data"""
        try:
            # Try to extract JSON from response
            # Sometimes Gemini wraps JSON in ```json ... ```
            if '```json' in analysis_text:
                json_start = analysis_text.find('```json') + 7
                json_end = analysis_text.find('```', json_start)
                json_text = analysis_text[json_start:json_end].strip()
            elif '```' in analysis_text:
                json_start = analysis_text.find('```') + 3
                json_end = analysis_text.find('```', json_start)
                json_text = analysis_text[json_start:json_end].strip()
            else:
                json_text = analysis_text.strip()

            analysis = json.loads(json_text)

            # Validate required fields
            required_fields = ['strategic_score', 'category', 'themes', 'key_insights', 'summary']
            for field in required_fields:
                if field not in analysis:
                    analysis[field] = None

            return analysis

        except json.JSONDecodeError as e:
            print(f"⚠️  {self.agent_id}: Failed to parse JSON, using fallback")
            return {
                'strategic_score': 50,
                'category': 'insight',
                'themes': ['unknown'],
                'key_insights': [analysis_text[:200]],
                'summary': analysis_text[:100],
                'raw_response': analysis_text,
                'parse_error': str(e)
            }

    def process_task(self, task: Task) -> Dict[str, Any]:
        """
        Process task from queue (callback for AgentWatcher)

        Args:
            task: Task from queue

        Returns:
            Task result
        """
        task_type = task.task_type
        payload = task.payload

        print(f"📋 {self.agent_id}: Processing task {task.task_id} ({task_type})")

        if task_type == 'analyze_signal':
            # Analyze signal
            result = self.analyze_signal(payload)
            return {
                'status': 'completed',
                'task_id': task.task_id,
                'result': result
            }

        elif task_type == 'batch_analyze':
            # Batch analysis (multiple signals)
            signals = payload.get('signals', [])
            results = []
            for signal in signals:
                analysis = self.analyze_signal(signal)
                results.append(analysis)

            return {
                'status': 'completed',
                'task_id': task.task_id,
                'results': results,
                'count': len(results)
            }

        else:
            return {
                'status': 'failed',
                'error': f"Unknown task type: {task_type}"
            }

    def start_watching(self, interval: int = 5):
        """
        Start autonomous queue watching

        Args:
            interval: Poll interval in seconds
        """
        watcher = AgentWatcher(
            agent_type=self.agent_type,
            agent_id=self.agent_id
        )

        print(f"SA: 큐 감시 시작.")

        # Start watching (blocking)
        watcher.watch(
            callback=self.process_task,
            interval=interval
        )


# ================== Standalone Execution ==================

if __name__ == '__main__':
    import argparse
    from core.system.env_validator import validate_env
    validate_env("sa_agent")

    parser = argparse.ArgumentParser(description='97layerOS Strategy Analyst Agent')
    parser.add_argument('--agent-id', default='sa-worker-1', help='Agent instance ID')
    parser.add_argument('--interval', type=int, default=5, help='Queue poll interval (seconds)')
    parser.add_argument('--test', action='store_true', help='Run test mode (single analysis)')

    args = parser.parse_args()

    # Initialize agent
    try:
        agent = StrategyAnalyst(agent_id=args.agent_id)
    except ValueError as e:
        print(f"❌ Initialization failed: {e}")
        print("💡 Set GOOGLE_API_KEY environment variable")
        sys.exit(1)

    if args.test:
        # Test mode: Single analysis
        print("\n🧪 Test Mode: Single Signal Analysis")
        print("=" * 50)

        test_signal = {
            'signal_id': 'test_001',
            'content': """
            I've been thinking about how AI tools are changing creative work.
            Instead of replacing creativity, they're removing the boring parts -
            research, formatting, repetitive tasks. This frees us to focus on
            what matters: ideas, connections, meaningful work. Maybe this aligns
            with slow living philosophy?
            """,
            'source': 'telegram',
            'timestamp': datetime.now().isoformat()
        }

        result = agent.analyze_signal(test_signal)

        print(f"\n📊 Analysis Result:")
        print(f"   Strategic Score: {result.get('strategic_score', 'N/A')}")
        print(f"   Category: {result.get('category', 'N/A')}")
        print(f"   Themes: {', '.join(result.get('themes', []))}")
        print(f"   Summary: {result.get('summary', 'N/A')}")
        print(f"\n   Key Insights:")
        for insight in result.get('key_insights', []):
            print(f"   - {insight}")

        print("\n✅ Test complete!")

    else:
        # Production mode: Autonomous queue watching
        print("\n🚀 Production Mode: Autonomous Queue Watching")
        print("=" * 50)
        agent.start_watching(interval=args.interval)
