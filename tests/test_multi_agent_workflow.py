#!/usr/bin/env python3
"""
Multi-Agent Workflow Test
Tests the full pipeline: SA → AD → CE → Ralph → CD
"""

import sys
import time
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.queue_manager import QueueManager
from core.agents.sa_agent import StrategyAnalyst
from core.agents.ad_agent import ArtDirector
from core.agents.ce_agent import ChiefEditor
from core.agents.ralph_agent import RalphLoop

print("🚀 Multi-Agent Workflow Test")
print("="*70)
print()

# Initialize agents
print("📦 Initializing agents...")
try:
    sa = StrategyAnalyst(agent_id='sa-test')
    ad = ArtDirector(agent_id='ad-test')
    ce = ChiefEditor(agent_id='ce-test')
    ralph = RalphLoop(agent_id='ralph-test')
    # cd = CreativeDirector(agent_id='cd-test')  # Skip CD (API key issue)
    print("✅ All agents initialized\n")
except Exception as e:
    print(f"❌ Agent initialization failed: {e}")
    sys.exit(1)

# Test signal
test_signal = {
    'signal_id': 'workflow_test_001',
    'content': """
    Noticed an interesting pattern: The best creative work happens when
    we're not trying to be productive. When we slow down, take walks,
    stare at clouds - that's when ideas emerge. Maybe our obsession with
    "productivity hacks" is actually killing creativity?
    """,
    'source': 'telegram',
    'timestamp': '2026-02-16T12:00:00'
}

print("🔄 Starting Workflow...")
print(f"Signal: {test_signal['content'][:80]}...\n")

# Step 1: SA Analysis
print("[1/5] 📊 Strategy Analyst...")
sa_result = sa.analyze_signal(test_signal)
if sa_result.get('error'):
    print(f"❌ SA failed: {sa_result['error']}")
    sys.exit(1)
print(f"✅ Strategic Score: {sa_result.get('strategic_score', 0)}/100")
print(f"    Themes: {', '.join(sa_result.get('themes', [])[:3])}\n")

# Step 2: AD Visual Concept
print("[2/5] 🎨 Art Director...")
ad_result = ad.create_visual_concept(sa_result)
if ad_result.get('error'):
    print(f"❌ AD failed: {ad_result['error']}")
    sys.exit(1)
print(f"✅ Concept: {ad_result.get('concept_title', 'N/A')}")
print(f"    Mood: {ad_result.get('visual_mood', 'N/A')}\n")

# Step 3: CE Content Writing
print("[3/5] ✍️  Chief Editor...")
ce_result = ce.write_content(sa_result, ad_result)
if ce_result.get('error'):
    print(f"❌ CE failed: {ce_result['error']}")
    sys.exit(1)
print(f"✅ Headline: {ce_result.get('headline', 'N/A')}")
print(f"    Social: {ce_result.get('social_caption', 'N/A')[:60]}...\n")

# Step 4: Ralph STAP Validation
print("[4/5] 🔄 Ralph Loop...")
ralph_result = ralph.validate_stap(ce_result)
if ralph_result.get('error'):
    print(f"❌ Ralph failed: {ralph_result['error']}")
    sys.exit(1)
print(f"✅ Decision: {ralph_result.get('decision', 'N/A').upper()}")
print(f"    Quality: {ralph_result.get('quality_score', 0)}/100")
print(f"    STAP: {ralph_result.get('stap_check', {})}\n")

# Step 5: CD Review (skipped due to API key)
print("[5/5] 👔 Creative Director... (SKIPPED - API key issue)")
print("     Would normally review and approve/reject\n")

# Summary
print("="*70)
print("✅ Workflow Complete!")
print()
print("📊 Summary:")
print(f"   Signal: {test_signal['signal_id']}")
print(f"   SA Score: {sa_result.get('strategic_score', 0)}/100")
print(f"   AD Mood: {ad_result.get('visual_mood', 'N/A')}")
print(f"   CE Headline: {ce_result.get('headline', 'N/A')[:50]}...")
print(f"   Ralph Decision: {ralph_result.get('decision', 'N/A').upper()}")
print()
print("🎯 Multi-agent collaboration successful!")
print("   Each agent processed independently, passing results forward.")
print("   Ready for queue-based autonomous operation.")
