#!/usr/bin/env python3
"""
HandoffEngine 단위 테스트

대상: core/system/handoff.py
커버리지: work_lock, filesystem_cache, asset_registry
외부 의존성 없음 (API 키 불필요)
"""

import sys
import pytest
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.handoff import HandoffEngine


@pytest.fixture
def engine():
    return HandoffEngine()


@pytest.fixture(autouse=True)
def cleanup_lock(engine):
    """테스트 전후 work_lock 정리"""
    engine.release_work_lock("TEST")
    yield
    engine.release_work_lock("TEST")


# ─── Work Lock ────────────────────────────────────────────────

def test_acquire_and_release_lock(engine):
    assert engine.acquire_work_lock("TEST", "단위 테스트", ["tests/"])
    assert engine.check_work_lock()["locked"] is True
    assert engine.check_work_lock()["agent"] == "TEST"

    engine.release_work_lock("TEST")
    assert engine.check_work_lock()["locked"] is False


def test_double_lock_rejected(engine):
    engine.acquire_work_lock("TEST", "첫 번째 잠금", ["tests/"])
    result = engine.acquire_work_lock("OTHER", "두 번째 잠금", ["tests/"])
    assert result is False


def test_release_by_wrong_agent_fails(engine):
    engine.acquire_work_lock("TEST", "잠금", ["tests/"])
    result = engine.release_work_lock("WRONG_AGENT")
    assert result is False
    assert engine.check_work_lock()["locked"] is True


# ─── Filesystem Cache ─────────────────────────────────────────

def test_cache_contains_known_paths(engine):
    cache = engine.update_filesystem_cache(force=True)
    assert len(cache["folders"]) > 0
    assert len(cache["files"]) > 0
    assert engine.check_path_exists("directives/THE_ORIGIN.md")
    assert engine.check_path_exists("directives/SYSTEM.md")


def test_cache_excludes_nonexistent(engine):
    engine.update_filesystem_cache(force=True)
    assert not engine.check_path_exists("nonexistent/fake_file.md")


# ─── Asset Registry ───────────────────────────────────────────

def test_register_and_retrieve_asset(engine):
    asset_id = engine.register_asset(
        path="tests/test_asset_temp.json",
        asset_type="test",
        source="unit_test",
        metadata={"automated": True}
    )
    assert asset_id is not None
    assert asset_id.startswith("AST-")


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
