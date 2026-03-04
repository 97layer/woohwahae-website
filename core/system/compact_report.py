#!/usr/bin/env python3
"""
compact_report.py — 간단 요약형 작업 리포트 생성기

입력: 요약, 개선/다음 단계
출력: knowledge/reports/daily/compact_<TS>.md 에 짧은 리포트 저장

예)
  python3 core/system/compact_report.py --agent-id codex \\
      --summary "아카이브 스타일 수정 완료" \\
      --next-step "CF Pages 배포 확인" \\
      --improvement-from-next
"""

from __future__ import annotations

import argparse
from datetime import datetime
from pathlib import Path
from typing import List

PROJECT_ROOT = Path(__file__).resolve().parents[2]
REPORT_DIR = PROJECT_ROOT / "knowledge" / "reports" / "daily"


def _lines(text: str) -> List[str]:
    return [ln.strip() for ln in text.splitlines() if ln.strip()]


def format_block(title: str, items: List[str]) -> str:
    if not items:
        return f"- {title}: (없음)"
    bullets = "\n".join(f"  - {item}" for item in items)
    return f"- {title}:\n{bullets}"


def build_report(ts: datetime, agent_id: str, summary: str, improvements: List[str], next_steps: List[str]) -> str:
    lines: List[str] = []
    header = ts.strftime("%Y-%m-%d %H:%M")
    lines.append(f"## {header} — {agent_id}")
    lines.append(format_block("완료", _lines(summary)))
    lines.append(format_block("개선", improvements))
    lines.append(format_block("다음", next_steps))
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description="Compact work report generator")
    parser.add_argument("--agent-id", required=True, help="에이전트 ID (예: codex)")
    parser.add_argument("--summary", required=True, help="완료 작업 요약(멀티라인 허용)")
    parser.add_argument("--next-step", action="append", default=[], help="다음 단계 (여러 번 지정 가능)")
    parser.add_argument("--improvement", action="append", default=[], help="개선/보완 사항 (여러 번 지정 가능)")
    parser.add_argument(
        "--improvement-from-next",
        action="store_true",
        help="개선 리스트가 비었을 때 next-step을 개선 항목으로도 사용",
    )
    parser.add_argument(
        "--out",
        help="출력 파일 경로 (기본: knowledge/reports/daily/compact_<TS>.md)",
    )
    parser.add_argument(
        "--quiet",
        action="store_true",
        help="경로만 출력하고 본문은 표시하지 않음",
    )
    args = parser.parse_args()

    ts = datetime.now()

    improvements = list(args.improvement)
    if not improvements and args.improvement_from_next:
        improvements = list(args.next_step)

    report_text = build_report(ts, args.agent_id, args.summary, improvements, args.next_step)

    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    # 파일명 충돌 방지를 위해 밀리초 + 4자리 난수 접미사를 사용한다.
    suffix = ts.strftime('%f')[:3]  # milliseconds
    try:
        import random
        rand = f"{random.randint(0, 9999):04d}"
    except Exception:
        rand = "0000"
    fname = f"compact_{ts.strftime('%Y%m%d_%H%M%S')}_{suffix}_{rand}.md"
    out_path = Path(args.out) if args.out else REPORT_DIR / fname
    out_path.write_text(report_text, encoding="utf-8")

    if args.quiet:
        print(out_path)
    else:
        print(f"✅ Compact report saved: {out_path}")
        print(report_text.strip())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
