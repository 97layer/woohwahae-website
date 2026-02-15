#!/usr/bin/env python3
"""
Phase 1 통합 테스트
세션 연속성 + 멀티에이전트 병렬 + 자산 추적 전체 워크플로우 검증

Author: 97layerOS Technical Director
"""

import sys
import asyncio
from pathlib import Path
from datetime import datetime

PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.handoff import HandoffEngine
from core.agents.asset_manager import AssetManager


def test_1_handoff_onboard():
    """1. Handoff Engine - Onboarding"""
    print("\n" + "="*70)
    print("TEST 1: Handoff Engine - Onboarding")
    print("="*70)

    engine = HandoffEngine()
    state = engine.onboard()

    assert state is not None, "❌ Onboard failed"
    assert "content" in state or "status" in state, "❌ Invalid state format"

    print("✅ TEST 1 PASSED: Context restored successfully")
    return engine


def test_2_work_lock(engine):
    """2. Work Lock Mechanism"""
    print("\n" + "="*70)
    print("TEST 2: Work Lock Mechanism")
    print("="*70)

    # Acquire lock
    assert engine.acquire_work_lock("TEST_TD", "Integration Test", ["knowledge/system/"]), \
        "❌ Failed to acquire lock"

    # Check lock status
    lock_status = engine.check_work_lock()
    assert lock_status['locked'] == True, "❌ Lock not active"
    assert lock_status['agent'] == "TEST_TD", "❌ Wrong agent in lock"

    print("✅ TEST 2 PASSED: Work lock acquired and verified")

    # Release lock
    engine.release_work_lock("TEST_TD")
    lock_status = engine.check_work_lock()
    assert lock_status['locked'] == False, "❌ Lock not released"

    print("✅ TEST 2 COMPLETED: Work lock released")


def test_3_filesystem_cache(engine):
    """3. Filesystem Cache"""
    print("\n" + "="*70)
    print("TEST 3: Filesystem Cache")
    print("="*70)

    cache = engine.update_filesystem_cache(force=True)

    assert cache is not None, "❌ Cache update failed"
    assert len(cache['folders']) > 0, "❌ No folders in cache"
    assert len(cache['files']) > 0, "❌ No files in cache"

    print(f"   Cached: {len(cache['folders'])} folders, {len(cache['files'])} files")

    # Check known paths (files, not top-level folders)
    assert engine.check_path_exists("directives/README.md"), "❌ directives/README.md not in cache"

    print("✅ TEST 3 PASSED: Filesystem cache working")


def test_4_asset_registration():
    """4. Asset Manager - Registration"""
    print("\n" + "="*70)
    print("TEST 4: Asset Manager - Registration")
    print("="*70)

    manager = AssetManager()

    # Register test asset
    test_path = f"tests/test_asset_{datetime.now().strftime('%Y%m%d_%H%M%S')}.md"
    asset_id = manager.register_asset(
        path=test_path,
        asset_type="insight",
        source="test",
        metadata={"test": True}
    )

    assert asset_id is not None, "❌ Asset registration failed"
    assert asset_id.startswith("AST-"), "❌ Invalid asset ID format"

    print(f"✅ TEST 4 PASSED: Asset registered with ID: {asset_id}")

    # Verify retrieval
    asset = manager.get_asset(asset_id)
    assert asset is not None, "❌ Cannot retrieve asset"
    assert asset['type'] == "insight", "❌ Wrong asset type"

    print(f"✅ TEST 4 COMPLETED: Asset retrieval working")

    return manager, asset_id


def test_5_asset_lifecycle(manager, asset_id):
    """5. Asset Lifecycle Management"""
    print("\n" + "="*70)
    print("TEST 5: Asset Lifecycle Management")
    print("="*70)

    # Update status: captured → analyzed
    success = manager.update_asset_status(
        asset_id=asset_id,
        new_status="analyzed",
        updated_by="TEST_SA",
        quality_score=75.0
    )

    assert success == True, "❌ Status update failed"

    # Verify update
    asset = manager.get_asset(asset_id)
    assert asset['status'] == "analyzed", "❌ Status not updated"
    assert asset['quality_score'] == 75.0, "❌ Quality score not updated"
    assert len(asset['lifecycle']) == 2, "❌ Lifecycle not recorded"

    print("✅ TEST 5 PASSED: Asset lifecycle tracking working")


def test_6_handoff_save(engine):
    """6. Handoff Engine - Session Save"""
    print("\n" + "="*70)
    print("TEST 6: Handoff Engine - Session Save")
    print("="*70)

    summary = """Phase 1 통합 테스트 완료
모든 컴포넌트 정상 작동 확인"""

    next_steps = [
        "Phase 1 Git 커밋",
        "Phase 2 시작: Telegram Executive Secretary",
        "Ralph Loop 통합"
    ]

    success = engine.handoff("TEST_TD", summary, next_steps)

    assert success == True, "❌ Handoff save failed"

    print("✅ TEST 6 PASSED: Session state saved to INTELLIGENCE_QUANTA.md")


def run_integration_tests():
    """통합 테스트 실행"""
    print("\n" + "🧪"*35)
    print("PHASE 1 INTEGRATION TEST SUITE")
    print("세션 연속성 + 멀티에이전트 병렬 + 자산 추적 워크플로우")
    print("🧪"*35)

    try:
        # Test 1: Onboard
        engine = test_1_handoff_onboard()

        # Test 2: Work Lock
        test_2_work_lock(engine)

        # Test 3: Filesystem Cache
        test_3_filesystem_cache(engine)

        # Test 4: Asset Registration
        manager, asset_id = test_4_asset_registration()

        # Test 5: Asset Lifecycle
        test_5_asset_lifecycle(manager, asset_id)

        # Test 6: Session Save
        test_6_handoff_save(engine)

        # Summary
        print("\n" + "="*70)
        print("🎉 ALL INTEGRATION TESTS PASSED")
        print("="*70)
        print("\n✅ Phase 1 인프라 검증 완료:")
        print("   - Session Handoff: ✅")
        print("   - Work Locking: ✅")
        print("   - Filesystem Cache: ✅")
        print("   - Asset Registry: ✅")
        print("   - Lifecycle Tracking: ✅")
        print("\n📦 Phase 1 완료 - Git 커밋 준비 완료")
        print("="*70 + "\n")

        return True

    except AssertionError as e:
        print(f"\n❌ TEST FAILED: {e}")
        return False
    except Exception as e:
        print(f"\n❌ UNEXPECTED ERROR: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = run_integration_tests()
    sys.exit(0 if success else 1)
