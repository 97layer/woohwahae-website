"""
Filename: execution/system/sync_status.py
Author: 97LAYER System (Upgraded)
Date: 2026-02-13

Purpose: 
1. 상태 파일 자동 병합 (Single Source of Truth)
2. 에이전트 간 생존 신고 (Heartbeat) - NEW!
"""

import os
import json
import time
from datetime import datetime
from pathlib import Path

# 경로 설정 (사용자 환경에 맞게 자동 인식)
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
STATUS_FILE_PATH = PROJECT_ROOT / "knowledge" / "system_state.json"

class SystemSynchronizer:
    def __init__(self, agent_name="Unknown_Agent"):
        self.agent_name = agent_name
        self.state_file = STATUS_FILE_PATH
        self._ensure_file_exists()

    def _ensure_file_exists(self):
        """상태 파일이 없으면 초기화"""
        if not self.state_file.exists():
            initial_state = {
                "system_status": "INITIALIZING",
                "last_update": str(datetime.now()),
                "agents": {},  # 여기에 각 에이전트 상태가 기록됨
                "tasks": {}    # 여기에 현재 미션이 기록됨
            }
            self._save_json(initial_state)

    def _load_json(self):
        try:
            with open(self.state_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception:
            return {"agents": {}, "tasks": {}}

    def _save_json(self, data):
        try:
            self.state_file.parent.mkdir(parents=True, exist_ok=True)
            with open(self.state_file, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2, ensure_ascii=False)
        except Exception as e:
            print(f"[ERROR] Save failed: {e}")

    def report_heartbeat(self, status="IDLE", current_task=None):
        """
        [핵심 기능] 나는 살아있다. (1분마다 호출)
        """
        data = self._load_json()
        
        # 내 구역 업데이트
        data["agents"][self.agent_name] = {
            "status": status,
            "last_heartbeat": str(datetime.now()),
            "current_task": current_task or "대기 중",
            "location": "Cloud" if "daemon" in str(PROJECT_ROOT) else "Local(Mac)"
        }
        
        # 전체 타임스탬프 갱신
        data["last_update"] = str(datetime.now())
        self._save_json(data)
        # 로그는 너무 시끄러우니 생략

    def get_agent_status(self):
        """누가 살아있는지 확인"""
        data = self._load_json()
        return data.get("agents", {})

# --- 기존의 파일 병합 로직도 유지 (호환성) ---
def sync_legacy_files():
    """기존 task_status.json과 knowledge/status.json 병합"""
    # (기존 코드의 로직을 여기에 통합하거나, 위 클래스로 대체 가능. 
    # 일단 Heartbeat가 급하므로 클래스 사용을 권장함)
    pass

if __name__ == "__main__":
    # 테스트
    syncer = SystemSynchronizer(agent_name="Test_Agent_Mac")
    syncer.report_heartbeat("ACTIVE", "동기화 모듈 테스트 중")
    
    print("\n=== 현재 시스템 생존 현황 ===")
    print(json.dumps(syncer.get_agent_status(), indent=2, ensure_ascii=False))