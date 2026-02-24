#!/usr/bin/env python3
"""
auto_quanta_update.py — QUANTA 자동 갱신 (세션 종료 시 호출)

git log 기반으로 이번 세션 변경사항을 추출하여
INTELLIGENCE_QUANTA.md의 '현재 상태' 섹션을 자동 교체.

Usage:
    python3 core/system/auto_quanta_update.py [--agent-id <id>]
"""
import argparse
import re
import subprocess
import sys
from datetime import datetime
from pathlib import Path


PROJECT_ROOT = Path(__file__).parent.parent.parent
QUANTA_PATH = PROJECT_ROOT / "knowledge" / "agent_hub" / "INTELLIGENCE_QUANTA.md"
SESSION_START_PATH = PROJECT_ROOT / "knowledge" / "system" / "session_start.txt"

SECTION_MARKER = "## 📍 현재 상태 (CURRENT STATE)"


def get_session_since() -> str:
    """세션 시작 시각 반환. 파일 없으면 120분 전 fallback."""
    if SESSION_START_PATH.exists():
        ts = SESSION_START_PATH.read_text().strip()
        if ts:
            return ts
    return "120 minutes ago"


def _run(cmd: list[str], cwd: Path = PROJECT_ROOT) -> str:
    try:
        result = subprocess.run(
            cmd, cwd=cwd, capture_output=True, text=True, timeout=10
        )
        return result.stdout.strip()
    except Exception:
        return ""


def get_session_commits() -> list[str]:
    """이번 세션 시작 이후 커밋 메시지 목록 반환."""
    since = get_session_since()
    out = _run([
        "git", "log",
        f"--since={since}",
        "--oneline",
        "--no-merges",
        "--pretty=format:%s",
    ])
    if not out:
        return []
    return [line.strip() for line in out.splitlines() if line.strip()]


def get_changed_files() -> list[str]:
    """이번 세션 시작 이후 변경된 고유 파일 목록."""
    since = get_session_since()
    out = _run([
        "git", "log",
        f"--since={since}",
        "--name-only",
        "--pretty=format:",
    ])
    files = {line.strip() for line in out.splitlines() if line.strip()}
    return sorted(files)


def get_uncommitted_files() -> list[str]:
    """미커밋 변경 파일 목록."""
    out = _run(["git", "status", "--porcelain"])
    if not out:
        return []
    files = []
    for line in out.splitlines():
        if line.strip():
            # e.g. " M core/system/foo.py" → "core/system/foo.py"
            parts = line.strip().split(None, 1)
            if len(parts) == 2:
                files.append(parts[1].strip())
    return files


def build_current_state_section(agent_id: str) -> str:
    """현재 세션 기반 '현재 상태' 섹션 생성."""
    now = datetime.now()
    timestamp = now.strftime("%Y-%m-%d %H:%M")

    commits = get_session_commits()
    uncommitted = get_uncommitted_files()

    lines = [
        SECTION_MARKER,
        "",
        f"### [{timestamp}] Auto-Update — {agent_id}",
        "",
    ]

    if commits:
        lines.append("**이번 세션 커밋**:")
        for msg in commits:
            lines.append(f"- ✅ {msg}")
        lines.append("")

    if uncommitted:
        lines.append("**미커밋 변경**:")
        for f in uncommitted:
            lines.append(f"- ⚠️  {f}")
        lines.append("")

    if not commits and not uncommitted:
        lines.append("*이번 세션 변경 없음*")
        lines.append("")

    lines.append(f"**업데이트 시간**: {now.isoformat()}")

    return "\n".join(lines)


def update_quanta(agent_id: str) -> bool:
    """QUANTA의 현재 상태 섹션을 새 내용으로 교체."""
    if not QUANTA_PATH.exists():
        print(f"ERROR: QUANTA not found at {QUANTA_PATH}", file=sys.stderr)
        return False

    content = QUANTA_PATH.read_text(encoding="utf-8")

    new_section = build_current_state_section(agent_id)

    # '현재 상태' 섹션 위치 찾기
    marker_idx = content.find(SECTION_MARKER)

    if marker_idx == -1:
        # 섹션 없으면 파일 끝에 추가
        updated = content.rstrip() + "\n\n" + new_section + "\n"
    else:
        # 섹션부터 파일 끝을 새 내용으로 교체
        prefix = content[:marker_idx].rstrip()
        updated = prefix + "\n\n" + new_section + "\n"

    # 마지막 갱신 날짜 헤더 업데이트
    today = datetime.now().strftime("%Y-%m-%d")
    updated = re.sub(
        r"(>\s*\*\*마지막 갱신\*\*:).*",
        rf"\1 {today} (auto-update by {agent_id})",
        updated,
    )

    QUANTA_PATH.write_text(updated, encoding="utf-8")
    return True


def main() -> None:
    parser = argparse.ArgumentParser(description="QUANTA 자동 갱신")
    parser.add_argument("--agent-id", default="auto-session", help="에이전트 식별자")
    args = parser.parse_args()

    since = get_session_since()
    print(f"QUANTA 자동 갱신 중 ({args.agent_id}) — since: {since}")

    commits = get_session_commits()
    uncommitted = get_uncommitted_files()
    print(f"  커밋 {len(commits)}개 / 미커밋 파일 {len(uncommitted)}개 감지")

    if update_quanta(args.agent_id):
        print("✅ INTELLIGENCE_QUANTA.md 갱신 완료")
    else:
        print("❌ QUANTA 갱신 실패", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
