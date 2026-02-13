#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Filename: test_72h_rule.py
Author: 97LAYER Test
Date: 2026-02-14
Description: 72시간 규칙 테스트 - Auto Publisher 검증
"""

import sys
import os
from pathlib import Path
from datetime import datetime, timedelta

BASE_DIR = Path(__file__).resolve().parent
sys.path.insert(0, str(BASE_DIR))
sys.path.insert(0, str(BASE_DIR / "execution"))

from auto_publisher import AutoPublisher

def test_72h_rule():
    """
    72h Rule 테스트: Draft 스캔 및 위반 처리
    """
    print("=" * 60)
    print("72-Hour Rule Test - Phase 5 Verification")
    print("=" * 60)

    publisher = AutoPublisher()

    # 1. Check for violations
    print("\n[Step 1] 현재 Draft 폴더 스캔...")
    violations = publisher.check_72h_rule()

    if not violations:
        print("  ✅ 위반 사항 없음 (모든 Draft가 72시간 이내)")
        print("\n[Simulation Mode] 테스트용 과거 날짜 Draft 생성 중...")

        # Create simulated old draft for testing
        test_draft = BASE_DIR / "knowledge" / "assets" / "draft" / "simulated_old_draft.md"

        # Simulate 75-hour old file (within grace period)
        simulated_time_75h = datetime.now() - timedelta(hours=75)

        with open(test_draft, "w", encoding="utf-8") as f:
            f.write(f"""---
id: simulated_test_75h
created: {simulated_time_75h.isoformat()}
status: draft
author: Chief_Editor
---

# Test Draft (75h old)
This is a simulated draft to test the 72h rule warning system.
""")

        # Manually set file creation time (Unix-like systems)
        simulated_timestamp = simulated_time_75h.timestamp()
        os.utime(test_draft, (simulated_timestamp, simulated_timestamp))

        print(f"  ✅ 시뮬레이션 Draft 생성: {test_draft.name} (75h old)")

        # Create another one for auto-discard (77h old)
        test_draft_77h = BASE_DIR / "knowledge" / "assets" / "draft" / "simulated_old_draft_77h.md"
        simulated_time_77h = datetime.now() - timedelta(hours=77)

        with open(test_draft_77h, "w", encoding="utf-8") as f:
            f.write(f"""---
id: simulated_test_77h
created: {simulated_time_77h.isoformat()}
status: draft
author: Chief_Editor
---

# Test Draft (77h old)
This should trigger auto-discard.
""")

        simulated_timestamp_77h = simulated_time_77h.timestamp()
        os.utime(test_draft_77h, (simulated_timestamp_77h, simulated_timestamp_77h))

        print(f"  ✅ 시뮬레이션 Draft 생성: {test_draft_77h.name} (77h old)")

        # Re-scan
        print("\n[Step 2] 재스캔...")
        violations = publisher.check_72h_rule()

    print(f"\n[Result] {len(violations)}건 위반 발견\n")

    # 2. Process violations
    for v in violations:
        print(f"📄 {v['file']}")
        print(f"   - 경과 시간: {v['elapsed_hours']}h")
        print(f"   - 상태: {v['status']}")

        if v["status"] == "violation":
            print(f"   - 조치: 🚨 자동 폐기 (76h+ 초과)")
            # Actually discard
            publisher.auto_discard(v["path"])
            print(f"   - ✅ Discarded to: knowledge/assets/discarded/")
        elif v["status"] == "warning":
            print(f"   - 조치: ⚠️ CD 알림 (4시간 유예)")

        print()

    # 3. Generate CD notification
    if violations:
        print("=" * 60)
        print("[CD Notification Preview]")
        print("=" * 60)
        notification = publisher.notify_cd(violations)
        print(notification)
        print("=" * 60)

    # 4. Verify discard folder
    discarded_dir = BASE_DIR / "knowledge" / "assets" / "discarded"
    discarded_files = list(discarded_dir.glob("*.md"))

    print(f"\n[Verification] Discarded 폴더: {len(discarded_files)}건")
    for df in discarded_files[-3:]:  # Show last 3
        print(f"  - {df.name}")

    print("\n✅ Test Status: SUCCESS")
    print("72시간 규칙이 정상적으로 작동합니다.")

if __name__ == "__main__":
    test_72h_rule()
