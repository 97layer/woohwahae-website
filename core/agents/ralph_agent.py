#!/usr/bin/env python3
"""
97layerOS Ralph Loop Agent
Phase 6.2: STAP Validation with Gemini 2.5 Flash

Role:
- Quality validation (STAP: Stop, Task, Assess, Process)
- Rapid iteration feedback
- Prevent perfectionism paralysis
- Enforce minimum quality bar

LLM: Gemini 2.5 Flash (Fast, Free)
Queue: Autonomous task claiming via AgentWatcher
Output: pass/revise/archive decisions

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os, sys, json
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.agent_watcher import AgentWatcher
from core.system.queue_manager import Task

try:
    import google.generativeai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    GEMINI_AVAILABLE = False


class RalphLoop:
    """Ralph Loop Agent - STAP Validation"""

    def __init__(self, agent_id: str = "ralph-worker-1", api_key: Optional[str] = None):
        self.agent_id = agent_id
        self.agent_type = "Ralph"
        
        if not GEMINI_AVAILABLE:
            raise ImportError("google-generativeai required")
        
        api_key = api_key or os.getenv('GOOGLE_API_KEY')
        if not api_key:
            raise ValueError("GOOGLE_API_KEY not found")
        
        genai.configure(api_key=api_key)
        self.model = genai.GenerativeModel('gemini-2.5-flash')
        
        print(f"✅ {self.agent_id}: Ralph Loop initialized (Gemini 2.5 Flash)")

    def validate_stap(self, asset_data: Dict[str, Any]) -> Dict[str, Any]:
        """STAP validation (Stop, Task, Assess, Process)"""
        signal_id = asset_data.get('signal_id', 'unknown')
        print(f"🔄 {self.agent_id}: STAP validation for {signal_id}...")

        prompt = f"""You are Ralph Loop validator for 97layer. Use STAP framework to validate work:

**Asset to Validate:**
{json.dumps(asset_data, indent=2)[:500]}...

**STAP Framework:**
- Stop: Is this worth doing? Does it align with values?
- Task: Is it clear what needs to be done?
- Assess: Is quality sufficient (not perfect)?
- Process: Can we ship this?

Provide validation in JSON:
{{
  "decision": "pass|revise|archive",
  "quality_score": <0-100>,
  "stap_check": {{
    "stop": true|false,
    "task": true|false,
    "assess": true|false,
    "process": true|false
  }},
  "feedback": "Brief feedback (2 sentences)"
}}

**Philosophy:**
- Good enough is good enough
- Progress over perfection
- Ship and iterate

Return ONLY valid JSON.
"""

        try:
            response = self.model.generate_content(prompt)
            text = response.text
            
            if '```json' in text:
                json_start = text.find('```json') + 7
                json_end = text.find('```', json_start)
                json_text = text[json_start:json_end].strip()
            else:
                json_text = text.strip()
            
            validation = json.loads(json_text)
            validation.update({
                'signal_id': signal_id,
                'validated_by': self.agent_id,
                'validated_at': datetime.now().isoformat(),
                'model': 'gemini-2.5-flash'
            })
            
            print(f"✅ {self.agent_id}: {validation['decision'].upper()} (score: {validation.get('quality_score', 0)})")
            return validation
            
        except Exception as e:
            print(f"❌ {self.agent_id}: Validation failed: {e}")
            return {'signal_id': signal_id, 'error': str(e), 'status': 'failed'}

    def process_task(self, task: Task) -> Dict[str, Any]:
        task_type = task.task_type
        payload = task.payload
        
        print(f"📋 {self.agent_id}: Processing task {task.task_id} ({task_type})")
        
        if task_type == 'validate_stap':
            asset_data = payload.get('asset', {})
            result = self.validate_stap(asset_data)
            return {'status': 'completed', 'task_id': task.task_id, 'result': result}
        else:
            return {'status': 'failed', 'error': f"Unknown task type: {task_type}"}

    def start_watching(self, interval: int = 5):
        watcher = AgentWatcher(agent_type=self.agent_type, agent_id=self.agent_id)
        print(f"👁️  {self.agent_id}: Starting autonomous operation...")
        print(f"   LLM: Gemini 2.5 Flash (Free, Fast)")
        print(f"   Tasks: validate_stap")
        print()
        watcher.watch(callback=self.process_task, interval=interval)


if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--agent-id', default='ralph-worker-1')
    parser.add_argument('--interval', type=int, default=5)
    parser.add_argument('--test', action='store_true')
    args = parser.parse_args()
    
    agent = RalphLoop(agent_id=args.agent_id)
    
    if args.test:
        print("\n🧪 Test Mode: STAP Validation\n" + "="*50)
        test_asset = {
            'signal_id': 'test_001',
            'type': 'content',
            'headline': 'Tending Our Digital Garden',
            'quality_indicators': ['coherent', 'on-brand', 'actionable']
        }
        
        result = agent.validate_stap(test_asset)
        print(f"\n🔄 Ralph Decision:")
        print(f"   Decision: {result.get('decision', 'N/A').upper()}")
        print(f"   Quality: {result.get('quality_score', 0)}/100")
        print(f"   STAP: {result.get('stap_check', {})}")
        print("\n✅ Test complete!")
    else:
        agent.start_watching(interval=args.interval)
