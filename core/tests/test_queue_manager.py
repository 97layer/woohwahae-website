#!/usr/bin/env python3
"""
QueueManager 단위 테스트

대상: core/system/queue_manager.py
커버리지: create_task, claim_task, complete_task, fail_task, get_pending_tasks
임시 디렉토리 사용 — 실제 큐 오염 없음
"""

import sys
import time
import pytest
import tempfile
from pathlib import Path

PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

from core.system.queue_manager import QueueManager, TaskStatus


@pytest.fixture
def queue(tmp_path):
    """임시 디렉토리 기반 격리된 QueueManager"""
    return QueueManager(queue_root=tmp_path / "queue")


# ─── Task 생성 ────────────────────────────────────────────────

def test_create_task_returns_id(queue):
    task_id = queue.create_task("SA", "analyze_signal", {"signal_id": "test_001"})
    assert task_id is not None
    assert "SA" in task_id
    assert "analyze_signal" in task_id


def test_created_task_in_pending(queue):
    task_id = queue.create_task("CE", "write_essay", {"corpus_id": "corp_001"})
    pending = queue.get_pending_tasks()
    ids = [t.task_id for t in pending]
    assert task_id in ids


# ─── Task claim ───────────────────────────────────────────────

def test_claim_task_moves_to_processing(queue):
    task_id = queue.create_task("SA", "analyze_signal", {"signal_id": "test_002"})
    task = queue.claim_task("sa-agent-01", task_id)

    assert task is not None
    assert task.status == TaskStatus.PROCESSING.value  # JSON 역직렬화 시 문자열
    assert task.claimed_by == "sa-agent-01"


def test_double_claim_rejected(queue):
    task_id = queue.create_task("SA", "analyze_signal", {"signal_id": "test_003"})
    queue.claim_task("sa-agent-01", task_id)
    result = queue.claim_task("sa-agent-02", task_id)
    assert result is None


def test_claimed_task_not_in_pending(queue):
    task_id = queue.create_task("AD", "validate_design", {"content_id": "test_004"})
    queue.claim_task("ad-agent-01", task_id)
    pending_ids = [t.task_id for t in queue.get_pending_tasks()]
    assert task_id not in pending_ids


# ─── Task 완료 / 실패 ─────────────────────────────────────────

def test_complete_task(queue):
    task_id = queue.create_task("CE", "write_essay", {"corpus_id": "corp_002"})
    queue.claim_task("ce-agent-01", task_id)
    success = queue.complete_task(task_id, {"essay_title": "슬로우라이프", "published": True})
    assert success is True


def test_fail_task(queue):
    task_id = queue.create_task("SA", "analyze_signal", {"signal_id": "bad_signal"})
    queue.claim_task("sa-agent-01", task_id)
    success = queue.fail_task(task_id, "Gemini API timeout")
    assert success is True


# ─── 필터링 ───────────────────────────────────────────────────

def test_get_pending_by_agent_type(queue):
    # task_id가 밀리초 타임스탬프 기반 — 충돌 방지
    queue.create_task("SA", "analyze_signal", {"id": "s1"})
    time.sleep(0.002)
    queue.create_task("SA", "analyze_signal", {"id": "s2"})
    time.sleep(0.002)
    queue.create_task("CE", "write_essay", {"id": "c1"})

    sa_tasks = queue.get_pending_tasks(agent_type="SA")
    assert len(sa_tasks) == 2
    assert all(t.agent_type == "SA" for t in sa_tasks)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
