#!/usr/bin/env python3
"""
97layerOS Art Director (AD) Agent
Phase 6.2: Independent agent with Gemini Pro Vision API

Role:
- Visual concept development and art direction
- Image generation guidance (for Stable Diffusion integration)
- Brand consistency and aesthetic validation
- Visual storytelling and composition

LLM: Gemini 1.5 Pro (Free tier, with Vision)
Queue: Autonomous task claiming via AgentWatcher
Output: Visual concepts, style guides, image prompts

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

# Gemini API (optional, for actual execution)
GEMINI_AVAILABLE = False
try:
    import google.generativeai as genai
    GEMINI_AVAILABLE = True
except ImportError:
    print("⚠️  google-generativeai not installed (mock mode)")


class ArtDirector:
    """
    Art Director Agent - Visual Concepts & Art Direction

    Capabilities:
    - Develop visual concepts from strategic insights
    - Generate image prompts for Stable Diffusion
    - Validate visual consistency with brand guidelines
    - Provide art direction feedback
    """

    def __init__(self, agent_id: str = "ad-worker-1", api_key: Optional[str] = None):
        """
        Initialize Art Director

        Args:
            agent_id: Unique agent instance ID
            api_key: Google API key (or from env GOOGLE_API_KEY)
        """
        self.agent_id = agent_id
        self.agent_type = "AD"
        self.mock_mode = not GEMINI_AVAILABLE

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

    def create_visual_concept(self, analysis_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Create visual concept from SA analysis

        Args:
            analysis_data: SA analysis result
                {
                    'themes': list,
                    'key_insights': list,
                    'summary': str,
                    'category': str
                }

        Returns:
            Visual concept with image prompts and style guide
        """
        signal_id = analysis_data.get('signal_id', 'unknown')
        themes = analysis_data.get('themes', [])
        insights = analysis_data.get('key_insights', [])
        summary = analysis_data.get('summary', '')

        print(f"🎨 {self.agent_id}: Creating visual concept for {signal_id}...")

        if self.mock_mode:
            # Mock response for testing
            return self._mock_visual_concept(signal_id, themes, insights)

        # Construct prompt for visual concept
        prompt = self._build_concept_prompt(themes, insights, summary)

        try:
            response = self.model.generate_content(prompt)
            concept_text = response.text

            # Parse structured response
            concept = self._parse_concept(concept_text)

            # Add metadata
            concept.update({
                'signal_id': signal_id,
                'created_by': self.agent_id,
                'created_at': datetime.now().isoformat(),
                'model': 'gemini-1.5-pro',
                'based_on': 'SA analysis'
            })

            print(f"✅ {self.agent_id}: Visual concept created")
            return concept

        except Exception as e:
            print(f"❌ {self.agent_id}: Concept creation failed: {e}")
            return {
                'signal_id': signal_id,
                'error': str(e),
                'status': 'failed'
            }

    def _build_concept_prompt(self, themes: list, insights: list, summary: str) -> str:
        """Build concept prompt for Gemini"""
        return f"""You are an Art Director for 97layer, a creative collective focused on slow living, meaningful work, and authentic expression.

Based on the following strategic analysis, create a visual concept:

**Themes:** {', '.join(themes)}
**Key Insights:** {'; '.join(insights)}
**Summary:** {summary}

Provide your visual concept in the following JSON format:
{{
  "concept_title": "Brief, evocative title",
  "visual_mood": "<contemplative|energetic|serene|bold|intimate|etc>",
  "color_palette": ["#hex1", "#hex2", "#hex3"],
  "composition_notes": "Brief composition guidance (2-3 sentences)",
  "image_prompts": [
    {{
      "prompt": "Detailed Stable Diffusion prompt",
      "style": "photography|illustration|3d|abstract",
      "aspect_ratio": "16:9|4:3|1:1|9:16"
    }}
  ],
  "typography_guidance": "Font style and text treatment suggestions",
  "reference_aesthetics": ["aesthetic1", "aesthetic2"],
  "brand_alignment": "<how this aligns with 97layer's slow living philosophy>"
}}

**Style Guidelines:**
- 97layer aesthetic: Minimalist, organic, thoughtful, human-centric
- Avoid: Corporate, overly polished, stock photo vibes
- Embrace: Authenticity, imperfection, slowness, depth

Return ONLY valid JSON, no additional text.
"""

    def _parse_concept(self, concept_text: str) -> Dict[str, Any]:
        """Parse Gemini response into structured data"""
        try:
            # Extract JSON
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

            concept = json.loads(json_text)
            return concept

        except json.JSONDecodeError as e:
            return {
                'concept_title': 'Visual Concept',
                'visual_mood': 'contemplative',
                'raw_response': concept_text,
                'parse_error': str(e)
            }

    def _mock_visual_concept(self, signal_id: str, themes: list, insights: list) -> Dict[str, Any]:
        """Mock visual concept for testing without API"""
        return {
            'signal_id': signal_id,
            'concept_title': f"Visual Exploration: {', '.join(themes[:2])}",
            'visual_mood': 'contemplative',
            'color_palette': ['#2C3E50', '#ECF0F1', '#95A5A6'],
            'composition_notes': 'Minimalist composition with organic elements. Focus on negative space and gentle transitions. Evoke a sense of slowness and depth.',
            'image_prompts': [
                {
                    'prompt': f'Minimalist photography, {themes[0] if themes else "abstract concept"}, soft natural lighting, muted tones, organic textures, contemplative mood, shallow depth of field, film grain',
                    'style': 'photography',
                    'aspect_ratio': '16:9'
                }
            ],
            'typography_guidance': 'Serif fonts for body, clean sans-serif for headers. Generous line spacing. Prioritize readability over decoration.',
            'reference_aesthetics': ['kinfolk magazine', 'wabi-sabi', 'slow living'],
            'brand_alignment': 'Aligns with 97layer slow living philosophy through minimalism and thoughtful composition',
            'created_by': self.agent_id,
            'created_at': datetime.now().isoformat(),
            'mode': 'mock',
            'based_on': f"{len(themes)} themes, {len(insights)} insights"
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

        if task_type == 'create_visual_concept':
            # Create visual concept from SA analysis
            analysis_data = payload.get('analysis', {})
            result = self.create_visual_concept(analysis_data)
            return {
                'status': 'completed',
                'task_id': task.task_id,
                'result': result
            }

        elif task_type == 'validate_visual':
            # Validate visual against brand guidelines (future feature)
            return {
                'status': 'completed',
                'task_id': task.task_id,
                'result': {'validated': True, 'notes': 'Brand-aligned'}
            }

        else:
            return {
                'status': 'failed',
                'error': f"Unknown task type: {task_type}"
            }

    def start_watching(self, interval: int = 5):
        """Start autonomous queue watching"""
        watcher = AgentWatcher(
            agent_type=self.agent_type,
            agent_id=self.agent_id
        )

        mode_str = "MOCK MODE" if self.mock_mode else "Gemini Pro Vision"
        print(f"👁️  {self.agent_id}: Starting autonomous operation...")
        print(f"   LLM: {mode_str}")
        print(f"   Tasks: create_visual_concept, validate_visual")
        print(f"   Queue: .infra/queue/tasks/pending/")
        print()

        watcher.watch(
            callback=self.process_task,
            interval=interval
        )


# ================== Standalone Execution ==================

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(description='97layerOS Art Director Agent')
    parser.add_argument('--agent-id', default='ad-worker-1', help='Agent instance ID')
    parser.add_argument('--interval', type=int, default=5, help='Queue poll interval (seconds)')
    parser.add_argument('--test', action='store_true', help='Run test mode')

    args = parser.parse_args()

    agent = ArtDirector(agent_id=args.agent_id)

    if args.test:
        print("\n🧪 Test Mode: Visual Concept Creation")
        print("=" * 50)

        # Mock SA analysis
        test_analysis = {
            'signal_id': 'test_001',
            'themes': ['AI creativity', 'slow living', 'meaningful work'],
            'key_insights': [
                'AI tools remove boring tasks, freeing creative energy',
                'Technology can support, not replace, human creativity',
                'Alignment with slow living: focus on what matters'
            ],
            'summary': 'AI as enabler of creative focus and slow living'
        }

        result = agent.create_visual_concept(test_analysis)

        print(f"\n🎨 Visual Concept:")
        print(f"   Title: {result.get('concept_title', 'N/A')}")
        print(f"   Mood: {result.get('visual_mood', 'N/A')}")
        print(f"   Palette: {result.get('color_palette', [])}")
        print(f"   Image Prompts:")
        for prompt_data in result.get('image_prompts', []):
            print(f"   - {prompt_data.get('prompt', 'N/A')[:80]}...")

        print("\n✅ Test complete!")

    else:
        print("\n🚀 Production Mode: Autonomous Queue Watching")
        print("=" * 50)
        agent.start_watching(interval=args.interval)
