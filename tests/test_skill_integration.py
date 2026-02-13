#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Test script for skills integration
"""

import sys
from pathlib import Path

# Add project root to path
BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(BASE_DIR))

def test_skill_engine_loading():
    """Test that SkillEngine loads skills correctly"""
    from libs.skill_engine import SkillEngine

    print("[Test] Initializing SkillEngine...")
    engine = SkillEngine()

    print(f"[Test] Loaded {len(engine.registry)} skills")
    for skill_id, skill in engine.registry.items():
        print(f"  - {skill_id}: {skill.name} (v{skill.version})")

    assert len(engine.registry) > 0, "No skills loaded"
    print("[Test] SkillEngine loading: PASSED")
    return engine

def test_skill_detection():
    """Test URL pattern detection"""
    from libs.skill_engine import SkillEngine

    engine = SkillEngine()

    test_cases = [
        ("https://youtube.com/watch?v=abc123", "skill-001"),
        ("https://youtu.be/abc123", "skill-001"),
        ("https://www.instagram.com/p/abc123/", "instagram_content_curator"),
        ("just some random text", None),
    ]

    print("[Test] Testing skill detection...")
    for text, expected in test_cases:
        result = engine.detect_skill_from_input(text)
        print(f"  Input: {text[:50]}")
        print(f"  Expected: {expected}, Got: {result}")
        assert result == expected, f"Detection failed for {text}"

    print("[Test] Skill detection: PASSED")

def test_agent_router_integration():
    """Test AgentRouter with SkillEngine"""
    from libs.agent_router import AgentRouter

    print("[Test] Initializing AgentRouter with skills...")
    router = AgentRouter()

    if router.skill_engine:
        print(f"[Test] SkillEngine integrated: {len(router.skill_engine.registry)} skills available")
    else:
        print("[Test] WARNING: SkillEngine not initialized in AgentRouter")

    # Test routing with YouTube URL
    youtube_url = "https://youtube.com/watch?v=test123"
    route = router.route(youtube_url)
    print(f"[Test] Route for YouTube URL: {route}")
    assert route.startswith("SKILL:"), "YouTube URL should route to skill"

    # Test normal text routing
    normal_text = "시스템 아키텍처 설계"
    route = router.route(normal_text)
    print(f"[Test] Route for normal text: {route}")
    assert not route.startswith("SKILL:"), "Normal text should route to agent"

    print("[Test] AgentRouter integration: PASSED")

def test_uip_execution():
    """Test UIP skill execution with a real YouTube URL"""
    from libs.skill_engine import SkillEngine

    engine = SkillEngine()

    # Use a known stable YouTube video
    test_url = "https://youtube.com/watch?v=dQw4w9WgXcQ"

    print(f"[Test] Executing UIP skill with URL: {test_url}")
    result = engine.execute_skill("skill-001", {"url": test_url})

    print(f"[Test] Result status: {result.get('status')}")
    print(f"[Test] Message: {result.get('message')}")

    if result.get("status") == "success":
        print(f"[Test] Output file: {result.get('output_file')}")
        print(f"[Test] Signal ID: {result.get('signal_id')}")
        print("[Test] UIP execution: PASSED")
    else:
        print(f"[Test] UIP execution: FAILED - {result.get('message')}")
        if "traceback" in result:
            print(result["traceback"])

def main():
    print("=" * 60)
    print("97layerOS Skills Integration Test Suite")
    print("=" * 60)
    print()

    try:
        test_skill_engine_loading()
        print()

        test_skill_detection()
        print()

        test_agent_router_integration()
        print()

        # Comment out UIP execution test by default (requires network)
        # Uncomment to test with a real YouTube URL
        # test_uip_execution()
        # print()

        print("=" * 60)
        print("All tests completed successfully")
        print("=" * 60)

    except Exception as e:
        print(f"\n[ERROR] Test failed: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == "__main__":
    main()
