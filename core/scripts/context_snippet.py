#!/usr/bin/env python3
"""
context_snippet.py — state.md에서 관련 섹션만 추출.
Task 서브에이전트 스폰 전 최소 컨텍스트 주입용.

사용:
  python3 core/scripts/context_snippet.py infra pipeline
  python3 core/scripts/context_snippet.py --list
"""
import sys
from pathlib import Path

ROOT = Path(__file__).parents[2]

SECTIONS = {
    "infra":    ["인프라 핵심", "실행 명령", "VM 서비스", "배포"],
    "pipeline": ["파이프라인", "P3", "P4", "오케스트레이터"],
    "web":      ["웹", "Cloudflare", "website", "빌드"],
    "content":  ["콘텐츠 전략", "에세이", "어조", "신호"],
    "design":   ["디자인", "시각", "CSS", "AD"],
    "agent":    ["에이전트", "SA", "CE", "AD", "Gardener", "Ralph"],
    "state":    ["현재 상태", "CURRENT STATE", "완료", "다음 작업"],
}


def extract(keys: list[str]) -> str:
    state_path = ROOT / "knowledge" / "agent_hub" / "state.md"
    if not state_path.exists():
        return "(state.md 없음)"

    text = state_path.read_text(encoding="utf-8")
    keywords = []
    for k in keys:
        keywords.extend(SECTIONS.get(k, [k]))

    lines = text.split("\n")
    blocks, current, capturing = [], [], False

    for line in lines:
        if line.startswith("#"):
            if current and capturing:
                blocks.append("\n".join(current))
            current = [line]
            capturing = any(kw.lower() in line.lower() for kw in keywords)
        else:
            current.append(line)

    if current and capturing:
        blocks.append("\n".join(current))

    return "\n\n".join(blocks) if blocks else "(해당 섹션 없음)"


if __name__ == "__main__":
    args = sys.argv[1:]

    if not args or args[0] == "--list":
        print("사용 가능한 키:", ", ".join(SECTIONS.keys()))
        sys.exit(0)

    print(extract(args))
