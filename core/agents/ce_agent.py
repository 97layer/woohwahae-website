#!/usr/bin/env python3
"""
97layerOS Chief Editor (CE) Agent
Phase 6.2: Independent agent with Gemini 2.5 Pro

Role:
- Content synthesis and editorial direction
- Transform insights + visuals into cohesive narratives
- Brand voice consistency (slow living, authentic)
- Final content output (copy, captions, articles)

LLM: Gemini 2.5 Pro (Free tier)
Queue: Autonomous task claiming via AgentWatcher
Output: Finalized content ready for CD approval

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
    import google.generativeai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False


class ChiefEditor:
    """
    Chief Editor Agent - Content Synthesis & Editorial Direction
    
    Capabilities:
    - Synthesize SA analysis + AD visuals into cohesive content
    - Write copy in 97layer brand voice (slow, authentic, human)
    - Create social media captions, articles, newsletters
    - Editorial QA before CD review
    """

    def __init__(self, agent_id: str = "ce-worker-1", api_key: Optional[str] = None):
        self.agent_id = agent_id
        self.agent_type = "CE"
        
        if not GEMINI_AVAILABLE:
            raise ImportError("google-generativeai required")
        
        api_key = api_key or os.getenv('GOOGLE_API_KEY')
        if not api_key:
            raise ValueError("GOOGLE_API_KEY not found")
        
        genai.configure(api_key=api_key)
        self.model = genai.GenerativeModel('gemini-2.5-pro')
        
        print(f"✅ {self.agent_id}: Chief Editor initialized (Gemini 2.5 Pro)")

    def write_content(self, analysis: Dict[str, Any], visual_concept: Dict[str, Any]) -> Dict[str, Any]:
        """
        Write content from SA analysis + AD visual concept
        
        Args:
            analysis: SA strategic analysis
            visual_concept: AD visual concept
        
        Returns:
            Final content draft
        """
        signal_id = analysis.get('signal_id', 'unknown')
        print(f"✍️  {self.agent_id}: Writing content for {signal_id}...")

        prompt = f"""You are Chief Editor for 97layer, a creative collective focused on slow living and meaningful work.

Based on the following inputs, write compelling content:

**Strategic Analysis:**
- Themes: {', '.join(analysis.get('themes', []))}
- Key Insights: {'; '.join(analysis.get('key_insights', []))}
- Summary: {analysis.get('summary', '')}

**Visual Concept:**
- Title: {visual_concept.get('concept_title', '')}
- Mood: {visual_concept.get('visual_mood', '')}
- Brand Alignment: {visual_concept.get('brand_alignment', '')}

Write content in JSON format:
{{
  "headline": "Compelling headline (5-10 words)",
  "subheadline": "Clarifying subheadline (10-15 words)",
  "body": "Main content (3-4 paragraphs, conversational tone)",
  "social_caption": "Instagram/Twitter caption (150 chars max)",
  "call_to_action": "What should readers do next?",
  "tags": ["tag1", "tag2", "tag3"],
  "tone": "contemplative|inspiring|thought-provoking"
}}

**97layer Brand Voice:**
- Slow, not rushed. Thoughtful, not reactive.
- Human, not corporate. Authentic, not polished.
- Deep, not surface. Meaningful, not trendy.
- Questions, not answers. Journey, not destination.

Return ONLY valid JSON.
"""

        try:
            response = self.model.generate_content(prompt)
            content_text = response.text
            
            # Parse JSON
            if '```json' in content_text:
                json_start = content_text.find('```json') + 7
                json_end = content_text.find('```', json_start)
                json_text = content_text[json_start:json_end].strip()
            else:
                json_text = content_text.strip()
            
            content = json.loads(json_text)
            
            content.update({
                'signal_id': signal_id,
                'written_by': self.agent_id,
                'written_at': datetime.now().isoformat(),
                'model': 'gemini-2.5-pro',
                'status': 'draft_for_cd'
            })
            
            print(f"✅ {self.agent_id}: Content draft complete")
            return content
            
        except Exception as e:
            print(f"❌ {self.agent_id}: Content creation failed: {e}")
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
        print(f"👁️  {self.agent_id}: Starting autonomous operation...")
        print(f"   LLM: Gemini 2.5 Pro")
        print(f"   Tasks: write_content")
        print()
        watcher.watch(callback=self.process_task, interval=interval)


if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--agent-id', default='ce-worker-1')
    parser.add_argument('--interval', type=int, default=5)
    parser.add_argument('--test', action='store_true')
    args = parser.parse_args()
    
    agent = ChiefEditor(agent_id=args.agent_id)
    
    if args.test:
        print("\n🧪 Test Mode: Content Writing\n" + "="*50)
        test_analysis = {
            'signal_id': 'test_001',
            'themes': ['AI creativity', 'slow living'],
            'key_insights': ['AI removes boring tasks', 'Focus on meaningful work'],
            'summary': 'AI enables creative focus'
        }
        test_visual = {
            'concept_title': 'Digital Garden',
            'visual_mood': 'contemplative',
            'brand_alignment': 'Aligns with slow living'
        }
        
        result = agent.write_content(test_analysis, test_visual)
        print(f"\n✍️  Content Draft:")
        print(f"   Headline: {result.get('headline', 'N/A')}")
        print(f"   Social: {result.get('social_caption', 'N/A')[:80]}...")
        print("\n✅ Test complete!")
    else:
        agent.start_watching(interval=args.interval)
