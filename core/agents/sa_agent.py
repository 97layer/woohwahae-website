#!/usr/bin/env python3
"""
97layerOS Strategy Analyst (SA) Agent
Phase 6.2: Independent agent with Gemini Flash API

Role:
- Signal analysis and strategic insight extraction
- First responder to incoming signals
- Pattern recognition across signals
- Strategic recommendations

LLM: Gemini 1.5 Flash (Free tier, fast)
Queue: Autonomous task claiming via AgentWatcher
Output: Structured analysis → passed to AD/CE

Author: 97layerOS Technical Director
Created: 2026-02-16
"""

import os
import sys
import json
import time
from pathlib import Path
from typing import Dict, Any, Optional
from datetime import datetime

# Project setup
PROJECT_ROOT = Path(__file__).parent.parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.agent_watcher import AgentWatcher
from core.system.queue_manager import Task

# Gemini API imports
try:
    import google.generativeai as genai
except ImportError:
    print("⚠️  google-generativeai not installed. Run: pip install google-generativeai")
    sys.exit(1)


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

        # Initialize Gemini API
        api_key = api_key or os.getenv('GOOGLE_API_KEY')
        if not api_key:
            raise ValueError("GOOGLE_API_KEY not found in environment")

        genai.configure(api_key=api_key)

        # Use Gemini 2.5 Flash (fast, free)
        self.model = genai.GenerativeModel('gemini-2.5-flash')

        print(f"✅ {self.agent_id}: Strategy Analyst initialized (Gemini 2.5 Flash)")

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

        print(f"🔍 {self.agent_id}: Analyzing signal {signal_id}...")

        # Construct prompt for strategic analysis
        prompt = self._build_analysis_prompt(content, source)

        try:
            # Call Gemini API
            response = self.model.generate_content(prompt)
            analysis_text = response.text

            # Parse structured response
            analysis = self._parse_analysis(analysis_text)

            # Add metadata
            analysis.update({
                'signal_id': signal_id,
                'analyzed_by': self.agent_id,
                'analyzed_at': datetime.now().isoformat(),
                'model': 'gemini-1.5-flash',
                'source': source
            })

            print(f"✅ {self.agent_id}: Analysis complete (score: {analysis.get('strategic_score', 0)})")

            return analysis

        except Exception as e:
            print(f"❌ {self.agent_id}: Analysis failed: {e}")
            return {
                'signal_id': signal_id,
                'error': str(e),
                'status': 'failed'
            }

    def _build_analysis_prompt(self, content: str, source: str) -> str:
        """Build analysis prompt for Gemini"""
        return f"""You are a Strategy Analyst for 97layer, a creative collective focused on slow living and meaningful work.

Analyze the following signal and provide strategic insights:

**Signal Content:**
{content}

**Source:** {source}

Provide your analysis in the following JSON format:
{{
  "strategic_score": <0-100, how strategic/valuable is this signal>,
  "category": "<trend|insight|opportunity|question|noise>",
  "themes": ["theme1", "theme2", "theme3"],
  "key_insights": [
    "Insight 1 (one sentence)",
    "Insight 2 (one sentence)",
    "Insight 3 (one sentence)"
  ],
  "action_items": [
    "Suggested action 1",
    "Suggested action 2"
  ],
  "recommended_agents": ["AD", "CE", "CD"],
  "summary": "One-sentence summary of the signal"
}}

**Guidelines:**
- Strategic score 80+: High-value, actionable insights
- Strategic score 50-80: Interesting patterns, worth exploring
- Strategic score <50: Low signal, noise, or needs clarification
- Themes: Identify recurring patterns or topics
- Action items: Concrete next steps
- Recommended agents: Which agents should handle this (AD for visual, CE for content, CD for big decisions)

Return ONLY valid JSON, no additional text.
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

        print(f"👁️  {self.agent_id}: Starting autonomous operation...")
        print(f"   LLM: Gemini 1.5 Flash (Free)")
        print(f"   Tasks: analyze_signal, batch_analyze")
        print(f"   Queue: .infra/queue/tasks/pending/")
        print()

        # Start watching (blocking)
        watcher.watch(
            callback=self.process_task,
            interval=interval
        )


# ================== Standalone Execution ==================

if __name__ == '__main__':
    import argparse

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
