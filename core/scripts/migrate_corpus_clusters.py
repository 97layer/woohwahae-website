"""
Corpus 군집 과분화 마이그레이션 스크립트 (one-time fix)
- 20개 micro-cluster → 4개 핵심 군집으로 병합
- 중복 entry_id 제거 (같은 entry가 여러 군집에 태깅된 것)
"""
import json
import logging
import sys
from datetime import datetime
from pathlib import Path

logging.basicConfig(level=logging.INFO, format="%(message)s")
logger = logging.getLogger(__name__)

ROOT = Path(__file__).parents[2]
INDEX_PATH = ROOT / "knowledge/corpus/index.json"

MERGE_MAP = {
    "슬로우라이프": [
        "균형", "성찰", "정체성", "느린 삶의 지향",
        "Slowlife", "지속가능성"
    ],
    "본질": [
        "본질 탐구", "사유의 가치", "알고리즘 비판", "사회적 피로"
    ],
    "여백": [
        "미니멀리즘", "기대와 현실"
    ],
    "시각 감도": [
        "Visual Aesthetic", "Editorial Curation",
        "Sensibility", "Composition"
    ],
}

KEEP_SOLO = ["정보 부족"]


def migrate():
    if not INDEX_PATH.exists():
        logger.info("corpus/index.json 없음 — VM에 corpus 미생성 상태. 스킵.")
        return

    with open(INDEX_PATH) as f:
        data = json.load(f)

    clusters = data.get("clusters", {})
    total_before = len(clusters)

    if total_before <= 5:
        logger.info("군집 %d개 — 이미 정리됨. 스킵.", total_before)
        return

    merged = {}

    for primary, absorb_list in MERGE_MAP.items():
        # 대표 군집 base
        base = clusters.get(primary, {
            "theme": primary,
            "entry_ids": [],
            "first_seen": datetime.now().isoformat(),
            "last_seen": datetime.now().isoformat(),
            "maturity": "accumulating",
            "essay_issued": None,
        })

        entry_ids = list(base.get("entry_ids", []))
        first_seen = base.get("first_seen", "")
        last_seen = base.get("last_seen", "")

        for absorb_name in absorb_list:
            src = clusters.get(absorb_name)
            if not src:
                continue
            for eid in src.get("entry_ids", []):
                if eid not in entry_ids:
                    entry_ids.append(eid)
            # first_seen은 가장 이른 것
            if src.get("first_seen", "") < first_seen:
                first_seen = src["first_seen"]
            if src.get("last_seen", "") > last_seen:
                last_seen = src["last_seen"]
            logger.info("  %s ← %s (%d entries)", primary, absorb_name, len(src.get("entry_ids", [])))

        merged[primary] = {
            "theme": primary,
            "entry_ids": entry_ids,
            "first_seen": first_seen,
            "last_seen": last_seen,
            "maturity": base.get("maturity", "accumulating"),
            "essay_issued": base.get("essay_issued"),
        }
        logger.info("[%s] 최종 entries: %d개", primary, len(entry_ids))

    for solo in KEEP_SOLO:
        if solo in clusters:
            merged[solo] = clusters[solo]

    data["clusters"] = merged
    data["last_updated"] = datetime.now().isoformat()

    with open(INDEX_PATH, "w") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)

    logger.info("완료: %d개 → %d개 군집", total_before, len(merged))


if __name__ == "__main__":
    migrate()
