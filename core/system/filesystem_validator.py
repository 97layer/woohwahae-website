#!/usr/bin/env python3
"""
Filesystem Validator — system.md §10 기반 파일 쓰기 사전 검증

Purpose:
- Python 에이전트의 모든 파일 쓰기를 system.md §10 규칙에 따라 검증
- 위반 시 PermissionError 발생 → 파일 생성 차단
- 성공 시 filesystem_cache.json 자동 갱신

Difference from filesystem_guard.py:
- guard.py: 사후 감시 daemon (15초마다 스캔 → 격리)
- validator.py: 사전 차단 API (쓰기 전 검증 → 차단)

Usage:
    from core.system.filesystem_validator import safe_write
    safe_write(path, content, agent_id="SA")

Author: THE ORIGIN Agent
Created: 2026-02-26
"""

from pathlib import Path
import json
import re
import sys
from datetime import datetime
from typing import Tuple, Optional, Union

PROJECT_ROOT = Path(__file__).parent.parent.parent
MANIFEST = PROJECT_ROOT / "directives" / "system.md"  # §10 Filesystem Placement
CACHE = PROJECT_ROOT / "knowledge" / "system" / "filesystem_cache.json"

# system.md §10 기반 명명 규칙
ALLOWED_PATTERNS = {
    "knowledge/signals/": [
        r"^(text|url|messenger|voice)_\d{8}_\d{6}\.json$",
        r"^youtube_[A-Za-z0-9_-]{11}_\d{8}_\d{6}\.json$",  # youtube with video ID
        r"^url_content_\d{8}_\d{6}_[a-f0-9]{10}\.json$",
    ],
    "knowledge/reports/": [
        r"^morning_\d{8}\.md$",
        r"^evening_\d{8}\.md$",
        r"^audit_\d{8}\.md$",
    ],
    "knowledge/corpus/entries/": [
        r"^entry_.*\.json$",
    ],
    "knowledge/clients/": [
        r"^client_\d{4}\.json$",
    ],
    "knowledge/system/schemas/": [
        r"^[a-z_]+\.schema\.json$",
    ],
    "knowledge/reports/growth/": [
        r"^growth_\d{6}\.json$",
    ],
    "knowledge/reports/daily/": [
        r"^(morning|evening)_\d{8}\.json$",
        r"^weekly_\d{4}W\d{2}\.json$",
    ],
    "website/archive/": [
        r"^(essay|magazine|lookbook)-\d{3}-.+/index\.html$",
        r"^index\.(html|json)$",
    ],
}

# 금지 패턴
FORBIDDEN_NAMES = [
    "SESSION_SUMMARY_", "WAKEUP_REPORT", "DEPLOY_", "NEXT_STEPS",
    "temp_", "untitled_", "무제", "Untitled"
]

# 루트 허용 파일
ROOT_ALLOWED = ["CLAUDE.md", "README.md", ".gitignore", ".env", ".ai_rules"]


def validate_write(path: Path) -> Tuple[bool, str]:
    """
    파일 쓰기 전 system.md §10 규칙 검증

    Returns:
        (허용 여부, 거부 사유 or "")
    """
    try:
        rel_path = path.relative_to(PROJECT_ROOT)
    except ValueError:
        return False, f"프로젝트 외부 경로: {path}"

    rel_str = str(rel_path)

    # 1. 금지 패턴 체크
    for forbidden in FORBIDDEN_NAMES:
        if forbidden in path.name:
            return False, f"금지 파일명 패턴: {forbidden}"

    # 2. 루트 .md 생성 제한
    if path.parent == PROJECT_ROOT:
        if path.suffix == ".md" and path.name not in ROOT_ALLOWED:
            return False, "루트에 .md 생성 금지 (CLAUDE.md, README.md 외)"
        # 루트 허용 파일은 통과
        if path.name in ROOT_ALLOWED:
            return True, ""

    # 3. website/ 예외 파일
    if rel_str == "website/archive/index.json":
        return True, ""
    if rel_str.startswith("website/") and path.suffix == ".md":
        if path.name != "README.md":
            return False, "website/ 내 .md 파일 생성 금지 (system.md §10)"

    # 4. 경로별 명명 규칙 검증
    matched_dir = None
    for allowed_dir in ALLOWED_PATTERNS.keys():
        if rel_str.startswith(allowed_dir):
            # 가장 긴 매칭 경로 선택 (reports/daily/ vs reports/)
            if matched_dir is None or len(allowed_dir) > len(matched_dir):
                matched_dir = allowed_dir

    if matched_dir:
        patterns = ALLOWED_PATTERNS[matched_dir]
        # 패턴 중 하나라도 매칭되면 통과
        for pattern in patterns:
            # 전체 경로 검증 (website/archive 같은 복잡한 구조)
            if "/" in pattern:
                check_path = rel_str.replace(matched_dir, "", 1)
                if re.match(pattern, check_path):
                    return True, ""
            # 파일명만 검증
            else:
                if re.match(pattern, path.name):
                    return True, ""

        # 이 디렉토리인데 패턴 불일치
        return False, (
            f"명명 규칙 위반: {matched_dir}\n"
            f"허용 패턴: {patterns}\n"
            f"시도한 파일명: {path.name}"
        )

    # 5. MANIFEST에 명시되지 않은 경로는 knowledge/ 외는 허용
    if rel_str.startswith("knowledge/"):
        # knowledge/ 하위는 엄격 적용
        # 단, 일부 허용 경로 추가 (agent_hub, system, docs 등)
        allowed_knowledge = [
            "knowledge/agent_hub/",
            "knowledge/system/",
            "knowledge/docs/",
            "knowledge/brands/",
            "knowledge/service/",
            "knowledge/corpus/",  # index.json, entries/
        ]
        for allowed in allowed_knowledge:
            if rel_str.startswith(allowed):
                return True, ""

        # 위 패턴에 없는 knowledge 하위 경로는 거부
        return False, f"system.md §10에 정의되지 않은 knowledge 경로: {rel_str}"

    # core/, directives/, .claude/ 등은 통과
    return True, ""


def safe_write(
    path: "Union[Path, str]",
    content: str,
    agent_id: str = "unknown",
    auto_register: bool = True
) -> None:
    """
    MANIFEST 검증 + 파일 쓰기 + 자동 등록

    Args:
        path: 쓸 파일 경로 (Path 또는 str)
        content: 파일 내용
        agent_id: 호출 에이전트 ID (로깅용)
        auto_register: filesystem_cache.json 자동 갱신 여부

    Raises:
        PermissionError: system.md §10 규칙 위반 시
    """
    if isinstance(path, str):
        path = Path(path)

    # 절대 경로로 변환
    if not path.is_absolute():
        path = PROJECT_ROOT / path

    # 검증
    ok, reason = validate_write(path)
    if not ok:
        raise PermissionError(
            f"\n{'='*60}\n"
            f"[Filesystem Validator] 파일 쓰기 거부\n"
            f"{'='*60}\n"
            f"에이전트: {agent_id}\n"
            f"경로: {path}\n"
            f"사유: {reason}\n"
            f"\nsystem.md §10 참조: {MANIFEST}\n"
            f"{'='*60}"
        )

    # 쓰기
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")

    # 캐시 등록
    if auto_register:
        register_to_cache(path)

    try:
        rel = path.relative_to(PROJECT_ROOT)
        print(f"✅ [{agent_id}] {rel}")
    except ValueError:
        print(f"✅ [{agent_id}] {path}")


def register_to_cache(path: Path) -> None:
    """
    filesystem_cache.json에 파일 등록
    """
    if not CACHE.exists():
        cache = {"last_scan": datetime.now().isoformat(), "folders": [], "files": []}
    else:
        try:
            cache = json.loads(CACHE.read_text(encoding="utf-8"))
        except Exception:
            cache = {"last_scan": datetime.now().isoformat(), "folders": [], "files": []}

    try:
        rel = str(path.relative_to(PROJECT_ROOT))
    except ValueError:
        return  # 프로젝트 외부 파일은 등록 안 함

    # 폴더 등록
    try:
        folder = str(path.parent.relative_to(PROJECT_ROOT))
        if folder and folder not in cache.get("folders", []):
            cache.setdefault("folders", []).append(folder)
    except ValueError:
        pass

    # 파일 등록
    if rel not in cache.get("files", []):
        cache.setdefault("files", []).append(rel)
        cache["last_scan"] = datetime.now().isoformat()
        CACHE.write_text(json.dumps(cache, indent=2, ensure_ascii=False), encoding="utf-8")


def validate_existing_files(root: Optional[Path] = None) -> list[dict]:
    """
    기존 파일 전체 검증 (감사용)

    Returns:
        위반 파일 목록 [{"path": "...", "reason": "..."}]
    """
    if root is None:
        root = PROJECT_ROOT

    violations = []

    for pattern in ["**/*.md", "**/*.json"]:
        for path in root.glob(pattern):
            # .git, node_modules 등 제외
            exclude_paths = [".git", "node_modules", ".pytest_cache", ".venv", "images"]
            if any(ignore in path.parts for ignore in exclude_paths):
                continue

            ok, reason = validate_write(path)
            if not ok:
                violations.append({
                    "path": str(path.relative_to(PROJECT_ROOT)),
                    "reason": reason
                })

    return violations


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="system.md §10 기반 파일 검증")
    parser.add_argument("--staged", action="store_true", help="Git staged 파일만 검증 (pre-commit)")
    parser.add_argument("--all", action="store_true", help="전체 파일 검증")
    args = parser.parse_args()

    if args.staged:
        # Git staged 파일 검증
        import subprocess
        result = subprocess.run(
            ["git", "diff", "--cached", "--name-only", "--diff-filter=ACM"],
            capture_output=True, text=True, cwd=PROJECT_ROOT
        )
        files = [PROJECT_ROOT / f for f in result.stdout.strip().split("\n") if f]

        violations = []
        for path in files:
            if path.exists() and path.is_file():
                ok, reason = validate_write(path)
                if not ok:
                    violations.append({"path": str(path.relative_to(PROJECT_ROOT)), "reason": reason})

        if violations:
            print("\n🔴 system.md §10 위반 파일 발견 — 커밋 중단\n")
            for v in violations:
                print(f"  - {v['path']}")
                print(f"    사유: {v['reason']}\n")
            sys.exit(1)
        else:
            print("✅ 모든 staged 파일이 system.md §10 규칙 준수")
            sys.exit(0)

    elif args.all:
        print("[Filesystem Validator] 전체 파일 검증 시작...\n")
        violations = validate_existing_files()

        if violations:
            print(f"🔴 system.md §10 위반 파일: {len(violations)}개\n")
            for v in violations[:20]:
                print(f"  - {v['path']}")
                print(f"    사유: {v['reason']}\n")
            sys.exit(1)
        else:
            print("✅ 모든 파일이 system.md §10 규칙을 준수합니다.")
            sys.exit(0)

    else:
        parser.print_help()
