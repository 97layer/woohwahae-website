#!/usr/bin/env python3
"""
구현 계획(implementation_plan.md) 파일을 읽어 다른 에이전트에게 전달하는 스크립트.
핵심 흐름:
1. 파일 경로 지정 및 내용 로드
2. Task 객체 생성 (task_type='process_implementation_plan')
3. queue_manager에 enqueue
4. 로그 출력
"""
import os
import sys
from pathlib import Path
import json

# 프로젝트 루트 경로 (현재 파일 위치 기준으로 두 단계 위)
PROJECT_ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(PROJECT_ROOT))

# 내부 모듈 import (이미 프로젝트에 존재한다고 가정)
from core.system.queue_manager import QueueManager, Task

# 구현 계획 파일 위치 (artifact 디렉터리 아래)
PLAN_PATH = Path(os.getenv('ANTIGRAVITY_BRAIN', str(PROJECT_ROOT / '.gemini' / 'antigravity' / 'brain' / '3edde949-b26e-4406-9c15-2dc453410775')) / 'implementation_plan.md'

def load_plan() -> str:
    if not PLAN_PATH.is_file():
        raise FileNotFoundError(f"Implementation plan not found at {PLAN_PATH}")
    return PLAN_PATH.read_text(encoding='utf-8')

def dispatch_plan():
    content = load_plan()
    task = Task(
        task_id='dispatch-plan-' + os.urandom(4).hex(),
        task_type='process_implementation_plan',
        payload={'plan_content': content}
    )
    qm = QueueManager()
    qm.enqueue(task)
    print(f"[PlanDispatcher] 구현 계획을 큐에 전송했습니다: {task.task_id}")

if __name__ == '__main__':
    dispatch_plan()
