#!/usr/bin/env python3
"""
Ritual Module — L4 Service Layer

아틀리에 고객 프로필 + 방문 기록 + 재방문 리듬 관리.
스키마: knowledge/system/schemas/ritual_client.schema.json

사용:
    from core.modules.ritual import RitualModule
    rm = RitualModule()
    rm.create_client("김순호", hair_type="건성 세모", rhythm="느린")
    rm.add_visit("client_0001", service="커트+컬러", satisfaction=5)
"""

import json
import logging
from datetime import datetime, date
from pathlib import Path
from typing import Dict, List, Optional

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
CLIENTS_DIR = PROJECT_ROOT / "knowledge" / "clients"


class RitualModule:
    """아틀리에 고객 관리 (L4 Service Layer)"""

    def __init__(self):
        CLIENTS_DIR.mkdir(parents=True, exist_ok=True)

    # ─── CRUD ─────────────────────────────────────────────────

    def create_client(
        self,
        name: str,
        hair_type: str = "",
        preference_notes: str = "",
        rhythm: str = "보통",
    ) -> Dict:
        """신규 고객 생성. client_id 자동 부여."""
        existing = self.list_clients()
        next_num = len(existing) + 1
        client_id = "client_%04d" % next_num

        client = {
            "client_id": client_id,
            "name": name,
            "rhythm": rhythm,
            "hair_type": hair_type,
            "preference_notes": preference_notes,
            "visits": [],
            "last_visit": None,
            "created_at": datetime.now().isoformat(),
        }

        self._save_client(client)
        logger.info("고객 생성: %s (%s)", client_id, name)
        return client

    def get_client(self, client_id: str) -> Optional[Dict]:
        """client_id로 고객 조회."""
        filepath = CLIENTS_DIR / ("%s.json" % client_id)
        if not filepath.exists():
            return None
        try:
            return json.loads(filepath.read_text(encoding="utf-8"))
        except (json.JSONDecodeError, IOError) as e:
            logger.warning("고객 파일 읽기 실패 (%s): %s", client_id, e)
            return None

    def find_client(self, name: str) -> Optional[Dict]:
        """이름으로 고객 검색 (부분 일치)."""
        for client in self.list_clients():
            if name in client.get("name", ""):
                return client
        return None

    def list_clients(self) -> List[Dict]:
        """전체 고객 목록 반환."""
        clients = []
        for filepath in sorted(CLIENTS_DIR.glob("client_*.json")):
            try:
                data = json.loads(filepath.read_text(encoding="utf-8"))
                clients.append(data)
            except (json.JSONDecodeError, IOError):
                continue
        return clients

    def update_client(self, client_id: str, **kwargs) -> Optional[Dict]:
        """고객 정보 수정. 허용 필드: name, rhythm, hair_type, preference_notes."""
        client = self.get_client(client_id)
        if not client:
            logger.warning("고객 없음: %s", client_id)
            return None

        allowed = {"name", "rhythm", "hair_type", "preference_notes"}
        for key, value in kwargs.items():
            if key in allowed:
                client[key] = value

        self._save_client(client)
        logger.info("고객 수정: %s", client_id)
        return client

    # ─── 방문 기록 ────────────────────────────────────────────

    def add_visit(
        self,
        client_id: str,
        service: str,
        notes: str = "",
        satisfaction: Optional[int] = None,
        visit_date: Optional[str] = None,
    ) -> Optional[Dict]:
        """방문 기록 추가."""
        client = self.get_client(client_id)
        if not client:
            logger.warning("고객 없음: %s", client_id)
            return None

        today = visit_date or date.today().isoformat()

        visit = {
            "date": today,
            "service": service,
            "notes": notes,
        }
        if satisfaction is not None:
            visit["satisfaction"] = max(1, min(5, satisfaction))

        client.setdefault("visits", []).append(visit)
        client["last_visit"] = today

        # 리듬 자동 계산 (3회 이상 방문 시)
        if len(client["visits"]) >= 3:
            client["rhythm"] = self._calculate_rhythm(client["visits"])

        self._save_client(client)
        logger.info("방문 기록: %s → %s (%s)", client_id, service, today)
        return visit

    def get_visit_history(self, client_id: str) -> List[Dict]:
        """방문 이력 반환."""
        client = self.get_client(client_id)
        if not client:
            return []
        return client.get("visits", [])

    # ─── 리듬 분석 ────────────────────────────────────────────

    def _calculate_rhythm(self, visits: List[Dict]) -> str:
        """방문 간격 평균으로 리듬 자동 판별."""
        if len(visits) < 2:
            return "보통"

        dates = []
        for v in visits:
            try:
                dates.append(date.fromisoformat(v["date"]))
            except (ValueError, KeyError):
                continue

        if len(dates) < 2:
            return "보통"

        dates.sort()
        intervals = [
            (dates[i + 1] - dates[i]).days for i in range(len(dates) - 1)
        ]
        avg_interval = sum(intervals) / len(intervals)

        if avg_interval <= 30:
            return "빠른"
        elif avg_interval <= 60:
            return "보통"
        else:
            return "느린"

    def check_revisit_due(self, client_id: str, threshold_days: int = 45) -> Dict:
        """재방문 시기 확인."""
        client = self.get_client(client_id)
        if not client or not client.get("last_visit"):
            return {"due": False, "message": "방문 기록 없음"}

        try:
            last = date.fromisoformat(client["last_visit"])
        except ValueError:
            return {"due": False, "message": "날짜 파싱 실패"}

        days_since = (date.today() - last).days

        # 리듬별 임계값 조정
        rhythm_thresholds = {"빠른": 35, "보통": 50, "느린": 70}
        actual_threshold = rhythm_thresholds.get(
            client.get("rhythm", "보통"), threshold_days
        )

        return {
            "due": days_since >= actual_threshold,
            "days_since": days_since,
            "threshold": actual_threshold,
            "rhythm": client.get("rhythm", "보통"),
            "client_name": client.get("name", ""),
        }

    def get_due_clients(self) -> List[Dict]:
        """재방문 시기가 된 모든 고객 반환."""
        due_list = []
        for client in self.list_clients():
            result = self.check_revisit_due(client["client_id"])
            if result.get("due"):
                due_list.append(result)
        return due_list

    # ─── 통계 ─────────────────────────────────────────────────

    def get_stats(self) -> Dict:
        """Ritual 모듈 통계."""
        clients = self.list_clients()
        total_visits = sum(len(c.get("visits", [])) for c in clients)
        due = self.get_due_clients()

        return {
            "total_clients": len(clients),
            "total_visits": total_visits,
            "due_for_revisit": len(due),
            "rhythm_distribution": self._rhythm_distribution(clients),
        }

    def _rhythm_distribution(self, clients: List[Dict]) -> Dict:
        dist = {"빠른": 0, "보통": 0, "느린": 0}
        for c in clients:
            rhythm = c.get("rhythm", "보통")
            dist[rhythm] = dist.get(rhythm, 0) + 1
        return dist

    # ─── 내부 유틸 ────────────────────────────────────────────

    def _save_client(self, client: Dict) -> None:
        filepath = CLIENTS_DIR / ("%s.json" % client["client_id"])
        filepath.write_text(
            json.dumps(client, ensure_ascii=False, indent=2), encoding="utf-8"
        )


# ─── Singleton ────────────────────────────────────────────────

_ritual_instance = None


def get_ritual_module() -> RitualModule:
    global _ritual_instance
    if _ritual_instance is None:
        _ritual_instance = RitualModule()
    return _ritual_instance


# ─── CLI ──────────────────────────────────────────────────────

if __name__ == "__main__":
    import sys

    logging.basicConfig(level=logging.INFO)
    rm = RitualModule()

    if len(sys.argv) < 2:
        print("Usage:")
        print("  python ritual.py create <name> [hair_type] [rhythm]")
        print("  python ritual.py visit <client_id> <service> [satisfaction]")
        print("  python ritual.py list")
        print("  python ritual.py due")
        print("  python ritual.py stats")
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "create":
        name = sys.argv[2] if len(sys.argv) > 2 else "Unknown"
        hair_type = sys.argv[3] if len(sys.argv) > 3 else ""
        rhythm = sys.argv[4] if len(sys.argv) > 4 else "보통"
        client = rm.create_client(name, hair_type=hair_type, rhythm=rhythm)
        print(json.dumps(client, ensure_ascii=False, indent=2))

    elif cmd == "visit":
        cid = sys.argv[2]
        service = sys.argv[3] if len(sys.argv) > 3 else "커트"
        sat = int(sys.argv[4]) if len(sys.argv) > 4 else None
        visit = rm.add_visit(cid, service=service, satisfaction=sat)
        print(json.dumps(visit, ensure_ascii=False, indent=2))

    elif cmd == "list":
        for c in rm.list_clients():
            visits = len(c.get("visits", []))
            print("%s | %s | 방문 %d회 | 리듬: %s" % (
                c["client_id"], c["name"], visits, c.get("rhythm", "보통")
            ))

    elif cmd == "due":
        due = rm.get_due_clients()
        if not due:
            print("재방문 대상 없음")
        for d in due:
            print("%s | %d일 경과 | 임계값: %d일" % (
                d["client_name"], d["days_since"], d["threshold"]
            ))

    elif cmd == "stats":
        stats = rm.get_stats()
        print(json.dumps(stats, ensure_ascii=False, indent=2))
