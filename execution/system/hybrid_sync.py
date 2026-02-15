#!/usr/bin/env python3
"""
하이브리드 동기화 시스템
맥북(로컬) ↔ Google Drive ↔ Google Cloud VM

역할:
1. 맥북이 작업 중: 로컬 → Drive 자동 업로드
2. 맥북이 없을 때: VM이 Drive에서 작업 pull
3. 양방향 동기화로 항상 최신 상태 유지
"""

import os
import json
import subprocess
from pathlib import Path
from datetime import datetime, timedelta
from typing import Literal

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
SYNC_STATE_FILE = PROJECT_ROOT / "knowledge" / "system" / "sync_state.json"

NodeType = Literal["macbook", "gcp_vm"]

class HybridSync:
    """하이브리드 동기화 관리자"""

    def __init__(self):
        self.location = self._detect_location()
        self.sync_state = self._load_sync_state()

    def _detect_location(self) -> str:
        """현재 실행 환경 감지"""
        # GCP VM 감지 (방법 1: 파일 존재)
        if Path("/etc/google_compute_engine").exists():
            return "GCP_VM"

        # GCP VM 감지 (방법 2: 메타데이터 서버)
        try:
            result = subprocess.run(
                ["curl", "-s", "-H", "Metadata-Flavor: Google",
                 "http://metadata.google.internal/computeMetadata/v1/instance/id"],
                capture_output=True,
                timeout=2
            )
            if result.returncode == 0 and result.stdout:
                return "GCP_VM"
        except:
            pass

        # Docker/Podman 감지
        if Path("/.dockerenv").exists() or Path("/run/.containerenv").exists():
            # 호스트 확인
            hostname = subprocess.check_output(["hostname"], text=True).strip()
            if "97layer" in hostname.lower() or "layer97" in hostname.lower():
                return "LOCAL_CONTAINER"
            else:
                return "CLOUD_CONTAINER"

        # 맥북 로컬
        return "LOCAL_MAC"

    def _load_sync_state(self) -> dict:
        """동기화 상태 로드"""
        if SYNC_STATE_FILE.exists():
            with open(SYNC_STATE_FILE, 'r') as f:
                return json.load(f)
        # 초기 상태 (Handshake 프로토콜 포함)
        return {
            "last_sync": None,
            "location": self.location,
            "pending_changes": [],
            # Handshake 필드
            "active_node": "macbook",
            "last_heartbeat": datetime.now().isoformat(),
            "pending_handover": False,
            "node_history": [],
            "health": {
                "macbook": "unknown",
                "gcp_vm": "unknown"
            }
        }

    def _save_sync_state(self):
        """동기화 상태 저장"""
        self.sync_state["last_sync"] = datetime.now().isoformat()
        self.sync_state["location"] = self.location

        SYNC_STATE_FILE.parent.mkdir(parents=True, exist_ok=True)
        with open(SYNC_STATE_FILE, 'w') as f:
            json.dump(self.sync_state, f, indent=2)

    def sync_to_drive(self, paths: list = None):
        """Google Drive로 업로드"""
        if paths is None:
            paths = ["knowledge/", "directives/", "execution/", "libs/"]

        print(f"🔼 [{self.location}] Google Drive로 동기화 중...")

        for path in paths:
            source = PROJECT_ROOT / path
            if not source.exists():
                continue

            # rclone 또는 rsync 사용 (구현에 따라)
            # 여기서는 Google Drive File Stream 사용 가정
            dest = Path("/Users/97layer/내 드라이브/97layerOS_Sync") / path

            try:
                if source.is_dir():
                    subprocess.run([
                        "rsync", "-av", "--delete",
                        "--exclude", "__pycache__",
                        "--exclude", ".DS_Store",
                        "--exclude", "*.pyc",
                        str(source) + "/",
                        str(dest) + "/"
                    ], check=True)
                else:
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    subprocess.run(["cp", str(source), str(dest)], check=True)

                print(f"  ✅ {path}")
            except Exception as e:
                print(f"  ❌ {path}: {e}")

        self._save_sync_state()
        print(f"✅ 동기화 완료: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

    def sync_from_drive(self, paths: list = None):
        """Google Drive에서 다운로드"""
        if paths is None:
            paths = ["knowledge/", "directives/", "execution/", "libs/"]

        print(f"🔽 [{self.location}] Google Drive에서 동기화 중...")

        for path in paths:
            source = Path("/Users/97layer/내 드라이브/97layerOS_Sync") / path
            if not source.exists():
                continue

            dest = PROJECT_ROOT / path

            try:
                if source.is_dir():
                    subprocess.run([
                        "rsync", "-av",
                        "--exclude", "__pycache__",
                        "--exclude", ".DS_Store",
                        "--exclude", "*.pyc",
                        str(source) + "/",
                        str(dest) + "/"
                    ], check=True)
                else:
                    dest.parent.mkdir(parents=True, exist_ok=True)
                    subprocess.run(["cp", str(source), str(dest)], check=True)

                print(f"  ✅ {path}")
            except Exception as e:
                print(f"  ❌ {path}: {e}")

        self._save_sync_state()
        print(f"✅ 동기화 완료: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

    # ===========================================
    # Handshake 프로토콜 (주권 확인)
    # ===========================================

    def claim_ownership(self, node: NodeType, timeout_minutes: int = 10) -> bool:
        """
        주권 요청 (The Handshake)

        Args:
            node: 요청 노드 ("macbook" 또는 "gcp_vm")
            timeout_minutes: 타임아웃 시간 (분)

        Returns:
            True: 주권 획득 (활성 모드)
            False: 주권 없음 (관찰 모드)
        """
        # Handshake 필드가 없으면 초기화
        if "active_node" not in self.sync_state:
            self.sync_state["active_node"] = "macbook"
            self.sync_state["last_heartbeat"] = datetime.now().isoformat()
            self.sync_state["pending_handover"] = False
            self.sync_state["node_history"] = []
            self.sync_state["health"] = {"macbook": "unknown", "gcp_vm": "unknown"}

        last_heartbeat = datetime.fromisoformat(self.sync_state["last_heartbeat"])
        time_since_heartbeat = datetime.now() - last_heartbeat
        current_owner = self.sync_state["active_node"]

        # Case 1: 이미 본인이 주권 보유
        if current_owner == node:
            self._update_heartbeat(node)
            return True

        # Case 2: 타임아웃 발생 (10분 무응답)
        if time_since_heartbeat > timedelta(minutes=timeout_minutes):
            print(f"[Handshake] {current_owner} 타임아웃 ({time_since_heartbeat}) → {node}로 주권 이관")
            self._transfer_ownership(current_owner, node)
            return True

        # Case 3: 타 노드 활성 → 관찰 모드
        print(f"[Handshake] {current_owner} 활성 중 (마지막 heartbeat: {time_since_heartbeat}초 전) → {node}는 관찰 모드")
        return False

    def _update_heartbeat(self, node: NodeType):
        """Heartbeat 갱신"""
        self.sync_state["last_heartbeat"] = datetime.now().isoformat()
        self.sync_state["health"][node] = "online"
        self._save_sync_state()

    def _transfer_ownership(self, from_node: NodeType, to_node: NodeType):
        """주권 이관"""
        self.sync_state["active_node"] = to_node
        self.sync_state["last_heartbeat"] = datetime.now().isoformat()
        self.sync_state["health"][from_node] = "offline"
        self.sync_state["health"][to_node] = "online"
        self.sync_state["node_history"].append({
            "from": from_node,
            "to": to_node,
            "timestamp": datetime.now().isoformat(),
            "reason": "timeout"
        })
        self._save_sync_state()
        print(f"✅ 주권 이관: {from_node} → {to_node}")

    def get_node_type(self) -> NodeType:
        """현재 노드 타입 반환"""
        if self.location in ["LOCAL_MAC", "LOCAL_CONTAINER"]:
            return "macbook"
        elif self.location == "GCP_VM":
            return "gcp_vm"
        else:
            return "macbook"  # 기본값

    def auto_sync(self):
        """자동 동기화 (환경에 따라)"""
        print(f"🔄 자동 동기화 시작 (환경: {self.location})")

        if self.location in ["LOCAL_MAC", "LOCAL_CONTAINER"]:
            # 맥북: 로컬 → Drive 업로드
            self.sync_to_drive()
        elif self.location == "GCP_VM":
            # GCP VM: Drive에서 pull 후 작업, 다시 push
            self.sync_from_drive()
            # 작업 수행
            self.sync_to_drive()
        elif self.location == "CLOUD_CONTAINER":
            # Cloud Run: 읽기 전용
            self.sync_from_drive()

if __name__ == "__main__":
    import sys

    sync = HybridSync()
    print(f"📍 현재 위치: {sync.location}")

    if len(sys.argv) > 1:
        command = sys.argv[1]
        if command == "push":
            sync.sync_to_drive()
        elif command == "pull":
            sync.sync_from_drive()
        elif command == "auto":
            sync.auto_sync()
        else:
            print("Usage: hybrid_sync.py [push|pull|auto]")
    else:
        sync.auto_sync()