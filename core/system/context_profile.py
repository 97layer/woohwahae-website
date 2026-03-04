#!/usr/bin/env python3
"""
Adaptive context profile helper.
최근 태스크 intent/키워드 기반으로 추가 섹션을 추천해 directive_loader가 동적으로 로드할 수 있게 한다.
"""

import json
import re
from pathlib import Path
from typing import List, Dict

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
PROFILE_PATH = PROJECT_ROOT / "knowledge/system/context_profile.json"

KEYWORD_SECTIONS = [
    (r"(배포|deploy|push|reset)", ["system.md:§15"]),  # 웹 락/빌드/배포
    (r"(cors|c\\W?o\\W?r\\W?s|origin)", ["practice.md:II-10", "system.md:§8"]),  # 보안/SA 관측
    (r"(디자인|시각|ui|css|레이아웃|폰트|spacing)", ["practice.md:Part I", "practice.md:I-10"]),
    (r"(카피|문구|어조|어미|에세이|콘텐츠)", ["practice.md:Part II", "practice.md:II-11"]),
    (r"(signal|신호|분석|데이터)", ["practice.md:II-10"]),
    (r"(서비스|예약|practice)", ["practice.md:Part III"]),
]


def save_profile(intents: List[str], max_items: int = 50) -> None:
    intents = intents[-max_items:]
    PROFILE_PATH.parent.mkdir(parents=True, exist_ok=True)
    PROFILE_PATH.write_text(json.dumps({"intents": intents}, ensure_ascii=False, indent=2), encoding="utf-8")


def load_profile() -> List[str]:
    if not PROFILE_PATH.exists():
        return []
    try:
        data = json.loads(PROFILE_PATH.read_text(encoding="utf-8"))
        return data.get("intents", [])
    except Exception:
        return []


def recommend_sections(intents: List[str]) -> List[str]:
    sections: List[str] = []
    for intent in intents[-5:]:  # 최근 5개만 사용
        text = intent.lower()
        for pattern, sec_list in KEYWORD_SECTIONS:
            if re.search(pattern, text):
                sections.extend(sec_list)
    # 중복 제거 순서 보존
    seen = set()
    uniq = []
    for sec in sections:
        if sec not in seen:
            seen.add(sec)
            uniq.append(sec)
    return uniq
