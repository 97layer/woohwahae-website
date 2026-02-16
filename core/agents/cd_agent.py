#!/usr/bin/env python3
"""
97layerOS Creative Director (CD) Agent
Phase 6.2: Independent agent with Claude Sonnet 4.5

Role:
- Final strategic decisions and brand stewardship
- Quality gate: approve/reject/revise content
- Big picture thinking and long-term vision
- Budget gatekeeper ($10/month limit)

LLM: Claude Sonnet 4.5 (Paid, $10/month budget)
Queue: Autonomous task claiming via AgentWatcher
Output: Final approve/reject decisions

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import sys
import json
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

# Project setup
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.agent_watcher import AgentWatcher
from core.system.queue_manager import Task

try:
    import anthropic
    CLAUDE_AVAILABLE = True
except ImportError:
    CLAUDE_AVAILABLE = False


class CreativeDirector:
    """
    Creative Director Agent - Final Decisions & Brand Stewardship
    
    Capabilities:
    - Review and approve/reject content drafts
    - Ensure brand alignment (slow living philosophy)
    - Strategic decisions on content direction
    - Budget-conscious (limited Claude API calls)
    """

    def __init__(self, agent_id: str = "cd-worker-1", api_key: Optional[str] = None):
        self.agent_id = agent_id
        self.agent_type = "CD"
        
        if not CLAUDE_AVAILABLE:
            raise ImportError("anthropic required: pip install anthropic")
        
        api_key = api_key or os.getenv('ANTHROPIC_API_KEY')
        if not api_key:
            raise ValueError("ANTHROPIC_API_KEY not found")
        
        self.client = anthropic.Anthropic(api_key=api_key)

        # 순호의 판단 기준 + IDENTITY 로드
        self._criteria = self._load_criteria()

        print(f"CD: 준비됨. 브랜드 기준 로드 완료.")

    def _load_criteria(self) -> str:
        """CD_SUNHO.md + IDENTITY.md 브랜드 판단 기준 로드"""
        parts = []
        for path in [
            PROJECT_ROOT / 'directives' / 'agents' / 'CD_SUNHO.md',
            PROJECT_ROOT / 'directives' / 'IDENTITY.md',
        ]:
            try:
                if path.exists():
                    parts.append(path.read_text(encoding='utf-8')[:2000])
            except Exception:
                pass
        return '\n\n---\n\n'.join(parts) if parts else "Remove the Noise, Reveal the Essence."

    def review_content(self, content_draft: Dict[str, Any]) -> Dict[str, Any]:
        """
        Review content draft and make final decision
        
        Args:
            content_draft: CE content draft
        
        Returns:
            Decision: approve/revise/reject + feedback
        """
        signal_id = content_draft.get('signal_id', 'unknown')
        print(f"CD: {signal_id} 검토 중.")

        prompt = f"""다음은 97layer 브랜드 판단 기준과 IDENTITY 문서다.

{self._criteria[:2500]}

---

이 콘텐츠 초안을 검토하고 최종 결정을 내려라.
판단은 단 하나의 질문으로 귀결된다: "내가 이걸 보고 싶은가? 이게 진짜 97layer인가?"

**콘텐츠 초안:**
- 헤드라인: {content_draft.get('headline', '')}
- 서브헤드라인: {content_draft.get('subheadline', '')}
- 본문 (일부): {content_draft.get('body', '')[:300]}
- 소셜 캡션: {content_draft.get('social_caption', '')}
- 톤: {content_draft.get('tone', '')}

JSON 형식으로 결정:
{{
  "decision": "approve|revise|reject",
  "brand_score": <0-100, 97layer 정렬도>,
  "strengths": ["강점 1", "강점 2"],
  "concerns": ["우려사항 1"] 또는 [],
  "revision_notes": "수정 방향 (revise일 때만, 구체적으로)" 또는 null,
  "strategic_rationale": "결정 이유 — 한두 문장, 직접적으로"
}}

JSON만 출력.
"""

        try:
            message = self.client.messages.create(
                model="claude-sonnet-4-5",
                max_tokens=1024,
                messages=[{"role": "user", "content": prompt}]
            )
            
            response_text = message.content[0].text
            
            # Parse JSON
            if '```json' in response_text:
                json_start = response_text.find('```json') + 7
                json_end = response_text.find('```', json_start)
                json_text = response_text[json_start:json_end].strip()
            else:
                json_text = response_text.strip()
            
            decision = json.loads(json_text)
            
            decision.update({
                'signal_id': signal_id,
                'reviewed_by': self.agent_id,
                'reviewed_at': datetime.now().isoformat(),
                'model': 'claude-sonnet-4-5',
                'api_usage': {
                    'input_tokens': message.usage.input_tokens,
                    'output_tokens': message.usage.output_tokens
                }
            })
            
            d = decision.get('decision', '').upper()
            score = decision.get('brand_score', 0)
            rationale = decision.get('strategic_rationale', '')[:60]
            print(f"CD: {d}. 점수 {score}. {rationale}")
            return decision
            
        except Exception as e:
            print(f"CD: 검토 실패. {e}")
            return {'signal_id': signal_id, 'error': str(e), 'status': 'failed'}

    def process_task(self, task: Task) -> Dict[str, Any]:
        task_type = task.task_type
        payload = task.payload
        
        print(f"📋 {self.agent_id}: Processing task {task.task_id} ({task_type})")
        
        if task_type == 'review_content':
            content_draft = payload.get('content_draft', {})
            result = self.review_content(content_draft)
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}
        else:
            return {'status': 'failed', 'error': f"Unknown task type: {task_type}"}

    def start_watching(self, interval: int = 5):
        watcher = AgentWatcher(agent_type=self.agent_type, agent_id=self.agent_id)
        print(f"👁️  {self.agent_id}: Starting autonomous operation...")
        print(f"   LLM: Claude Sonnet 4.5 ($10/month budget)")
        print(f"   Tasks: review_content")
        print()
        watcher.watch(callback=self.process_task, interval=interval)


if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--agent-id', default='cd-worker-1')
    parser.add_argument('--interval', type=int, default=5)
    parser.add_argument('--test', action='store_true')
    args = parser.parse_args()
    
    agent = CreativeDirector(agent_id=args.agent_id)
    
    if args.test:
        print("\n🧪 Test Mode: Content Review\n" + "="*50)
        test_draft = {
            'signal_id': 'test_001',
            'headline': 'Tending Our Digital Garden',
            'subheadline': 'How AI can support slow, meaningful work',
            'body': 'In an age of constant acceleration, AI tools offer an unexpected gift: not speed, but depth...',
            'social_caption': 'What if AI\'s purpose was depth, not speed?',
            'tone': 'contemplative'
        }
        
        result = agent.review_content(test_draft)
        print(f"\n👔 CD Decision:")
        print(f"   Decision: {result.get('decision', 'N/A').upper()}")
        print(f"   Brand Score: {result.get('brand_score', 0)}/100")
        print(f"   Strengths: {', '.join(result.get('strengths', [])[:2])}")
        if result.get('api_usage'):
            print(f"   API Usage: {result['api_usage']['input_tokens']}in + {result['api_usage']['output_tokens']}out tokens")
        print("\n✅ Test complete!")
    else:
        agent.start_watching(interval=args.interval)
