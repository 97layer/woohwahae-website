#!/usr/bin/env python3
"""
plan_dispatch_classifier.py

Task complexity classifier for plan_dispatch auto gating.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from typing import Dict, List


HIGH_KEYWORD_PATTERNS = (
    r"리팩토링",
    r"아키텍처",
    r"구조\s*변경",
    r"대규모",
    r"하네스",
    r"파이프라인",
    r"통합",
    r"연동",
    r"마이그레이션",
    r"다중\s*파일",
    r"여러\s*파일",
    r"전면",
    r"재설계",
    r"전수조사",
    r"hardening",
    r"migration",
    r"refactor",
    r"pipeline",
    r"architecture",
    r"rollout",
    r"toolchain",
)

MEDIUM_KEYWORD_PATTERNS = (
    r"분석",
    r"진단",
    r"점검",
    r"감사",
    r"개선",
    r"업그레이드",
    r"안전화",
    r"체계",
    r"스킬",
    r"툴",
    r"도구",
    r"검증",
    r"호환성",
    r"테스트",
    r"리스크",
    r"우선순위",
    r"mvp",
    r"audit",
    r"upgrade",
    r"improve",
    r"compatibility",
)

ACTION_TOKEN_PATTERN = re.compile(
    r"(분석|진단|점검|개선|업그레이드|구현|수정|검증|테스트|설계|정리|"
    r"리팩토링|마이그레이션|audit|upgrade|implement|refactor)",
    flags=re.IGNORECASE,
)

THRESHOLD_MAP = {"simple": 0, "medium": 1, "high": 2}


def _normalize_task(task: str) -> str:
    return re.sub(r"\s+", " ", (task or "").strip().lower())


def _match_any(patterns: tuple[str, ...], text: str) -> bool:
    for pat in patterns:
        if re.search(pat, text, flags=re.IGNORECASE):
            return True
    return False


def classify_task(task: str, min_complexity: str = "medium") -> Dict[str, object]:
    task = (task or "").strip()
    normalized = _normalize_task(task)
    min_complexity = (min_complexity or "medium").strip().lower()

    score = 0
    signals: List[str] = []

    if len(task) >= 60:
        score += 1
        signals.append("len>=60")
    if len(task) >= 120:
        score += 1
        signals.append("len>=120")
    if len(task) >= 220:
        score += 1
        signals.append("len>=220")

    if _match_any(HIGH_KEYWORD_PATTERNS, normalized):
        score += 3
        signals.append("high_keyword")
    if _match_any(MEDIUM_KEYWORD_PATTERNS, normalized):
        score += 2
        signals.append("medium_keyword")

    if "그리고" in task or "," in task or " + " in task:
        score += 1
        signals.append("multi_clause")
    if "/" in task:
        score += 1
        signals.append("slash_separator")

    action_tokens = ACTION_TOKEN_PATTERN.findall(normalized)
    if len(action_tokens) >= 2:
        score += 1
        signals.append("multi_action_tokens")

    if score >= 5:
        complexity = "high"
        level = 2
    elif score >= 2:
        complexity = "medium"
        level = 1
    else:
        complexity = "simple"
        level = 0

    threshold = THRESHOLD_MAP.get(min_complexity, THRESHOLD_MAP["medium"])
    allowed = level >= threshold

    return {
        "complexity": complexity,
        "score": score,
        "level": level,
        "threshold": threshold,
        "allowed": allowed,
        "signals": signals,
    }


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Plan dispatch task complexity classifier")
    parser.add_argument("--task", required=True, help="Task text")
    parser.add_argument("--min-complexity", default="medium", help="simple|medium|high")
    parser.add_argument("--json", action="store_true", help="Print JSON payload (default)")
    return parser


def main() -> int:
    parser = _build_parser()
    args = parser.parse_args()
    payload = classify_task(args.task, args.min_complexity)
    print(json.dumps(payload, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
