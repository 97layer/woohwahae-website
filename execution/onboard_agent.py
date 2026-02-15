#!/usr/bin/env python3
"""
신규 에이전트 온보딩 자동화

Usage:
  python3 execution/onboard_agent.py --role SA
  python3 execution/onboard_agent.py --role NEW_AGENT

Output:
  - 필수 읽기 순서 출력
  - Directive 파일 위치 안내
  - 상태 파일 위치 확인
  - 브랜드 보호 파일 명시
"""

import sys
from pathlib import Path

ROLES = {
    "SA": {
        "name": "Strategy Analyst",
        "file": "directives/agents/strategy_analyst.md",
        "must_read": ["cycle_protocol.md", "anti_algorithm_protocol.md"],
        "multimodal": True
    },
    "AD": {
        "name": "Art Director",
        "file": "directives/agents/art_director.md",
        "must_read": ["visual_identity_guide.md", "aesop_benchmark.md"],
        "multimodal": True
    },
    "CE": {
        "name": "Chief Editor",
        "file": "directives/agents/chief_editor.md",
        "must_read": ["imperfect_publish_protocol.md", "communication_protocol.md"],
        "multimodal": True
    },
    "CD": {
        "name": "Creative Director",
        "file": "directives/agents/creative_director.md",
        "must_read": ["brand_constitution.md", "97layer_identity.md", "woohwahae_identity.md"],
        "multimodal": True
    },
    "TD": {
        "name": "Technical Director",
        "file": "directives/agents/technical_director.md",
        "must_read": ["cycle_protocol.md", "daemon_workflow.md", "sync_protocol.md"],
        "multimodal": True
    }
}

def onboard(role_code: str):
    """에이전트 온보딩 프로세스"""
    if role_code not in ROLES:
        print(f"❌ Unknown role: {role_code}")
        print(f"📋 Available roles: {', '.join(ROLES.keys())}")
        return

    r = ROLES[role_code]
    print(f"🚀 Onboarding: {r['name']} ({role_code})\n")

    print("=" * 60)
    print("📖 REQUIRED READING ORDER (위반 시 시스템 파편화)")
    print("=" * 60)

    print("\n▶ Phase 1: System Constitution")
    print("  1. CLAUDE.md (3-Layer Architecture)")
    print("  2. directives/directive_lifecycle.md ⭐ (Core Constitution)")
    print("  3. directives/system_handshake.md (Handover Protocol)")

    print("\n▶ Phase 2: Identity")
    print("  4. directives/97layer_identity.md ⭐ (Foundation)")
    print(f"  5. {r['file']} (Your Role)")

    print(f"\n▶ Phase 3: Role-Specific Directives")
    for i, doc in enumerate(r['must_read'], 1):
        path = f"directives/{doc}"
        exists = "✅" if Path(path).exists() else "❌"
        print(f"  {5+i}. {exists} {doc}")

    print("\n" + "=" * 60)
    print("🔒 READ-ONLY DIRECTIVES (절대 수정 금지)")
    print("=" * 60)
    print("  - woohwahae_identity.md 🔒")
    print("  - brand_constitution.md 🔒")
    print("  - 97layer_identity.md 🔒")
    print("\n  ⚠️  이유: 브랜드 정체성은 AI가 최적화할 대상이 아님")

    print("\n" + "=" * 60)
    print("📂 STATE FILES")
    print("=" * 60)
    print("  - knowledge/system_state.json (실시간 상태)")
    print("  - knowledge/system/task_status.json (작업 진행)")
    print("  - knowledge/agent_hub/synapse_bridge.json (협업 상태)")

    if r.get('multimodal'):
        print("\n" + "=" * 60)
        print("⚡ MULTIMODAL SYSTEM")
        print("=" * 60)
        print("  - Read: docs/milestones/ASYNC_MULTIMODAL_IMPLEMENTATION.md")
        print("  - Core: libs/async_agent_hub.py")
        print("  - Pipeline: execution/async_five_agent_multimodal.py")
        print("  - Performance: 2.5x productivity (11s parallel)")

    print("\n" + "=" * 60)
    print("🌱 GARDENER SYSTEM")
    print("=" * 60)
    print("  - Location: libs/gardener.py")
    print("  - Role: Pattern detection, Directive promotion")
    print("  - Rule: 3회 반복 시 Knowledge→Directive 승격")

    print("\n✅ Onboarding complete. Start with Step 1!\n")

if __name__ == "__main__":
    role = "SA"  # Default
    if len(sys.argv) > 2 and sys.argv[1] == "--role":
        role = sys.argv[2]
    elif len(sys.argv) > 1:
        role = sys.argv[1]

    onboard(role)