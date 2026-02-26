#!/usr/bin/env python3
"""
Growth Module — L5 Business Layer

월별 수익/콘텐츠/서비스 지표 수집 + 추세 분석.
스키마: knowledge/system/schemas/growth_metrics.schema.json

사용:
    from core.system.growth import GrowthModule
    gm = GrowthModule()
    gm.record_revenue("2026-02", atelier=1800000, products=250000)
    gm.auto_count_content("2026-02")
    gm.generate_monthly_report("2026-02")
"""

import json
import logging
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional

logger = logging.getLogger(__name__)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
GROWTH_DIR = PROJECT_ROOT / "knowledge" / "reports" / "growth"
SIGNALS_DIR = PROJECT_ROOT / "knowledge" / "signals"
CORPUS_DIR = PROJECT_ROOT / "knowledge" / "corpus" / "entries"
CLIENTS_DIR = PROJECT_ROOT / "knowledge" / "clients"
ARCHIVE_DIR = PROJECT_ROOT / "website" / "archive"


class GrowthModule:
    """월별 성장 지표 관리 (L5 Business Layer)"""

    def __init__(self):
        GROWTH_DIR.mkdir(parents=True, exist_ok=True)

    # ─── 월별 지표 CRUD ──────────────────────────────────────

    def get_month(self, period: str) -> Dict:
        """월별 지표 조회. 없으면 빈 템플릿 반환."""
        filepath = GROWTH_DIR / ("growth_%s.json" % period.replace("-", ""))
        if filepath.exists():
            try:
                return json.loads(filepath.read_text(encoding="utf-8"))
            except (json.JSONDecodeError, IOError):
                pass

        return self._empty_template(period)

    def save_month(self, data: Dict) -> str:
        """월별 지표 저장."""
        period = data.get("period", "")
        if not period:
            raise ValueError("period 필드 필수")

        data["recorded_at"] = datetime.now().isoformat()
        filename = "growth_%s.json" % period.replace("-", "")
        filepath = GROWTH_DIR / filename

        filepath.write_text(
            json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8"
        )
        logger.info("성장 지표 저장: %s", filename)
        return str(filepath)

    def list_periods(self) -> List[str]:
        """기록된 월 목록 반환."""
        periods = []
        for f in sorted(GROWTH_DIR.glob("growth_*.json")):
            name = f.stem.replace("growth_", "")
            if len(name) == 6:
                periods.append("%s-%s" % (name[:4], name[4:]))
        return periods

    # ─── 수익 기록 ────────────────────────────────────────────

    def record_revenue(
        self,
        period: str,
        atelier: int = 0,
        consulting: int = 0,
        products: int = 0,
    ) -> Dict:
        """수익 수동 입력."""
        data = self.get_month(period)
        revenue = data.setdefault("revenue", {})

        if atelier:
            revenue["atelier"] = atelier
        if consulting:
            revenue["consulting"] = consulting
        if products:
            revenue["products"] = products

        revenue["total"] = (
            revenue.get("atelier", 0)
            + revenue.get("consulting", 0)
            + revenue.get("products", 0)
        )

        self.save_month(data)
        logger.info("수익 기록: %s → 총 %s원", period, revenue["total"])
        return data

    # ─── 자동 집계 ────────────────────────────────────────────

    def auto_count_content(self, period: str) -> Dict:
        """시스템 데이터에서 콘텐츠 지표 자동 집계."""
        data = self.get_month(period)
        content = data.setdefault("content", {})

        year_month = period.replace("-", "")

        # 신호 수집 수 (해당 월)
        signals_count = 0
        if SIGNALS_DIR.exists():
            for f in SIGNALS_DIR.glob("*.json"):
                if year_month[:6] in f.name:
                    signals_count += 1
        content["signals_captured"] = signals_count

        # 발행 에세이 수 (archive/ 디렉토리 기준)
        essays_count = 0
        if ARCHIVE_DIR.exists():
            for d in ARCHIVE_DIR.iterdir():
                if d.is_dir() and d.name.startswith("issue-"):
                    essays_count += 1
        content["essays_published"] = essays_count

        # 군집 수 (corpus entries 기준)
        clusters = set()
        if CORPUS_DIR.exists():
            for f in CORPUS_DIR.glob("*.json"):
                try:
                    entry = json.loads(f.read_text(encoding="utf-8"))
                    cluster = entry.get("cluster_id") or entry.get("metadata", {}).get("cluster_id")
                    if cluster:
                        clusters.add(cluster)
                except (json.JSONDecodeError, IOError):
                    continue
        content["clusters_ripe"] = len(clusters)

        self.save_month(data)
        logger.info(
            "콘텐츠 자동 집계: %s → 신호 %d, 에세이 %d, 군집 %d",
            period, signals_count, essays_count, len(clusters),
        )
        return data

    def auto_count_service(self, period: str) -> Dict:
        """Ritual Module 데이터에서 서비스 지표 자동 집계."""
        data = self.get_month(period)
        service = data.setdefault("service", {})

        year_month = period  # "2026-02"
        total_visits = 0
        client_ids_visited = set()

        if CLIENTS_DIR.exists():
            for f in CLIENTS_DIR.glob("client_*.json"):
                try:
                    client = json.loads(f.read_text(encoding="utf-8"))
                    for visit in client.get("visits", []):
                        vdate = visit.get("date", "")
                        if vdate.startswith(year_month):
                            total_visits += 1
                            client_ids_visited.add(client["client_id"])
                except (json.JSONDecodeError, IOError):
                    continue

        # 신규 vs 재방문 판별
        new_clients = 0
        returning = 0
        for cid in client_ids_visited:
            client = self._load_client(cid)
            if client:
                month_visits = [
                    v for v in client.get("visits", [])
                    if v.get("date", "").startswith(year_month)
                ]
                prior_visits = [
                    v for v in client.get("visits", [])
                    if not v.get("date", "").startswith(year_month)
                ]
                if prior_visits:
                    returning += 1
                else:
                    new_clients += 1

        service["total_visits"] = total_visits
        service["new_clients"] = new_clients
        service["returning_clients"] = returning

        self.save_month(data)
        logger.info(
            "서비스 자동 집계: %s → 방문 %d, 신규 %d, 재방문 %d",
            period, total_visits, new_clients, returning,
        )
        return data

    # ─── 추세 분석 ────────────────────────────────────────────

    def get_trend(self, months: int = 3) -> List[Dict]:
        """최근 N개월 추세 반환."""
        periods = self.list_periods()
        recent = periods[-months:] if len(periods) >= months else periods

        trend = []
        for period in recent:
            data = self.get_month(period)
            trend.append({
                "period": period,
                "revenue_total": data.get("revenue", {}).get("total", 0),
                "signals": data.get("content", {}).get("signals_captured", 0),
                "essays": data.get("content", {}).get("essays_published", 0),
                "visits": data.get("service", {}).get("total_visits", 0),
            })

        return trend

    # ─── 월간 리포트 ──────────────────────────────────────────

    def generate_monthly_report(self, period: str) -> str:
        """월간 성장 리포트 마크다운 생성."""
        # 자동 집계 먼저 실행
        self.auto_count_content(period)
        self.auto_count_service(period)

        data = self.get_month(period)
        revenue = data.get("revenue", {})
        content = data.get("content", {})
        service = data.get("service", {})

        report = """# Growth Report — %s

## 수익
| 항목 | 금액 |
|------|------|
| 아틀리에 | %s원 |
| 컨설팅 | %s원 |
| 제품 | %s원 |
| **합계** | **%s원** |

## 콘텐츠
- 신호 수집: %d건
- 발행 에세이: %d개
- 성숙 군집: %d개

## 서비스
- 총 방문: %d건
- 신규 고객: %d명
- 재방문 고객: %d명

---
*자동 생성: %s*
""" % (
            period,
            "{:,}".format(revenue.get("atelier", 0)),
            "{:,}".format(revenue.get("consulting", 0)),
            "{:,}".format(revenue.get("products", 0)),
            "{:,}".format(revenue.get("total", 0)),
            content.get("signals_captured", 0),
            content.get("essays_published", 0),
            content.get("clusters_ripe", 0),
            service.get("total_visits", 0),
            service.get("new_clients", 0),
            service.get("returning_clients", 0),
            datetime.now().isoformat(),
        )

        # MANIFEST 검증 적용 (growth_YYYYMM.json이 정확한 형식)
        from core.system.filesystem_validator import safe_write
        report_path = PROJECT_ROOT / "knowledge" / "reports" / "growth" / ("growth_%s.json" % period.replace("-", ""))
        safe_write(report_path, report, agent_id="Growth")
        logger.info("월간 리포트 생성: %s", report_path.name)
        return str(report_path)

    # ─── 내부 유틸 ────────────────────────────────────────────

    def _empty_template(self, period: str) -> Dict:
        return {
            "period": period,
            "revenue": {"atelier": 0, "consulting": 0, "products": 0, "total": 0},
            "content": {"essays_published": 0, "signals_captured": 0, "clusters_ripe": 0, "instagram_posts": 0},
            "service": {"total_visits": 0, "new_clients": 0, "returning_clients": 0},
            "recorded_at": None,
        }

    def _load_client(self, client_id: str) -> Optional[Dict]:
        filepath = CLIENTS_DIR / ("%s.json" % client_id)
        if filepath.exists():
            try:
                return json.loads(filepath.read_text(encoding="utf-8"))
            except (json.JSONDecodeError, IOError):
                pass
        return None


# ─── Singleton ────────────────────────────────────────────────

_growth_instance = None


def get_growth_module() -> GrowthModule:
    global _growth_instance
    if _growth_instance is None:
        _growth_instance = GrowthModule()
    return _growth_instance


# ─── CLI ──────────────────────────────────────────────────────

if __name__ == "__main__":
    import sys

    logging.basicConfig(level=logging.INFO)
    gm = GrowthModule()

    if len(sys.argv) < 2:
        print("Usage:")
        print("  python growth.py record <YYYY-MM> [atelier] [consulting] [products]")
        print("  python growth.py count <YYYY-MM>")
        print("  python growth.py report <YYYY-MM>")
        print("  python growth.py trend [months]")
        sys.exit(1)

    cmd = sys.argv[1]

    if cmd == "record":
        period = sys.argv[2]
        atelier = int(sys.argv[3]) if len(sys.argv) > 3 else 0
        consulting = int(sys.argv[4]) if len(sys.argv) > 4 else 0
        products = int(sys.argv[5]) if len(sys.argv) > 5 else 0
        data = gm.record_revenue(period, atelier=atelier, consulting=consulting, products=products)
        print(json.dumps(data, ensure_ascii=False, indent=2))

    elif cmd == "count":
        period = sys.argv[2]
        gm.auto_count_content(period)
        gm.auto_count_service(period)
        data = gm.get_month(period)
        print(json.dumps(data, ensure_ascii=False, indent=2))

    elif cmd == "report":
        period = sys.argv[2]
        path = gm.generate_monthly_report(period)
        print("리포트 생성: %s" % path)

    elif cmd == "trend":
        months = int(sys.argv[2]) if len(sys.argv) > 2 else 3
        trend = gm.get_trend(months)
        for t in trend:
            print("%s | 수익: %s원 | 신호: %d | 에세이: %d | 방문: %d" % (
                t["period"],
                "{:,}".format(t["revenue_total"]),
                t["signals"],
                t["essays"],
                t["visits"],
            ))
