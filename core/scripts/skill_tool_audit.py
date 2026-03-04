#!/usr/bin/env python3
"""
skill_tool_audit.py

Audit skill metadata and baseline tooling health for the local runtime.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Optional, Tuple


PROJECT_ROOT = Path(__file__).resolve().parents[2]
CONTRACT_PATH = PROJECT_ROOT / "knowledge" / "system" / "agent_runtime_contract.json"
REQUIRED_FRONTMATTER_KEYS = ("name", "description", "version", "updated")


@dataclass
class CheckResult:
    name: str
    status: str  # pass, warn, fail
    detail: str


def run_command(cmd: List[str]) -> Tuple[int, str, str]:
    proc = subprocess.run(
        cmd,
        cwd=str(PROJECT_ROOT),
        capture_output=True,
        text=True,
    )
    return proc.returncode, proc.stdout.strip(), proc.stderr.strip()


def load_contract() -> Dict[str, object]:
    if not CONTRACT_PATH.exists():
        return {}
    try:
        raw = json.loads(CONTRACT_PATH.read_text(encoding="utf-8"))
    except Exception:
        return {}
    if not isinstance(raw, dict):
        return {}
    return raw


def parse_frontmatter(path: Path) -> Optional[Dict[str, str]]:
    text = path.read_text(encoding="utf-8", errors="ignore")
    lines = text.splitlines()
    if not lines or lines[0].strip() != "---":
        return None

    frontmatter: Dict[str, str] = {}
    for line in lines[1:]:
        stripped = line.strip()
        if stripped == "---":
            return frontmatter
        if not stripped or ":" not in stripped:
            continue
        key, value = stripped.split(":", 1)
        frontmatter[key.strip()] = value.strip()
    return None


def check_external_required_skills(contract: Dict[str, object]) -> CheckResult:
    required = contract.get("required_skills", [])
    if not isinstance(required, list) or not required:
        return CheckResult("external-required-skills", "warn", "required_skills missing in runtime contract")

    missing = []
    for raw_path in required:
        skill_path = Path(str(raw_path))
        if not skill_path.is_file():
            missing.append(str(skill_path))

    if missing:
        return CheckResult("external-required-skills", "fail", f"missing: {', '.join(missing)}")
    return CheckResult("external-required-skills", "pass", f"checked={len(required)}")


def check_core_skill_frontmatter() -> CheckResult:
    skill_files = sorted((PROJECT_ROOT / "core" / "skills").glob("*/SKILL.md"))
    if not skill_files:
        return CheckResult("core-skill-frontmatter", "fail", "no core skill files found")

    invalid = []
    warnings = []
    semver_pat = re.compile(r"^\d+\.\d+\.\d+$")

    for skill_file in skill_files:
        frontmatter = parse_frontmatter(skill_file)
        rel = str(skill_file.relative_to(PROJECT_ROOT))
        if frontmatter is None:
            invalid.append(f"{rel}: missing yaml frontmatter")
            continue

        missing_keys = [k for k in REQUIRED_FRONTMATTER_KEYS if not frontmatter.get(k)]
        if missing_keys:
            invalid.append(f"{rel}: missing keys={','.join(missing_keys)}")
            continue

        expected_name = skill_file.parent.name
        current_name = frontmatter.get("name", "").strip()
        if current_name and current_name != expected_name:
            warnings.append(f"{rel}: name '{current_name}' != dir '{expected_name}'")

        version = frontmatter.get("version", "").strip()
        if version and not semver_pat.match(version):
            warnings.append(f"{rel}: non-semver version '{version}'")

    if invalid:
        return CheckResult("core-skill-frontmatter", "fail", " | ".join(invalid[:4]))
    if warnings:
        return CheckResult("core-skill-frontmatter", "warn", " | ".join(warnings[:4]))
    return CheckResult("core-skill-frontmatter", "pass", f"checked={len(skill_files)}")


def check_core_tool_files(contract: Dict[str, object]) -> CheckResult:
    required: List[str] = []

    contract_files = contract.get("required_files", [])
    if isinstance(contract_files, list):
        for rel in contract_files:
            rel_text = str(rel)
            if rel_text.startswith("core/scripts/") or rel_text.startswith("core/system/"):
                required.append(rel_text)

    for key in ("plan_dispatch_file", "plan_council_file"):
        value = contract.get(key)
        if value:
            required.append(str(value))

    required.extend(
        [
            "core/scripts/session_bootstrap.sh",
            "core/scripts/harness_doctor.py",
            "core/scripts/structure_audit.py",
            "core/system/evidence_guard.py",
        ]
    )

    deduped = sorted(set(required))
    missing = []
    non_exec = []

    for rel in deduped:
        path = PROJECT_ROOT / rel
        if not path.exists():
            missing.append(rel)
            continue
        if path.suffix == ".sh" and not os.access(path, os.X_OK):
            non_exec.append(rel)

    if missing:
        return CheckResult("core-tool-files", "fail", f"missing: {', '.join(missing)}")
    if non_exec:
        return CheckResult("core-tool-files", "warn", f"non-executable shell scripts: {', '.join(non_exec)}")
    return CheckResult("core-tool-files", "pass", f"checked={len(deduped)}")


def check_plan_dispatch_smoke_behavior() -> CheckResult:
    cases = (
        (
            "simple",
            "ok",
            False,
            {"simple"},
        ),
        (
            "nontrivial",
            "스킬 및 기본 툴 체계를 분석하고 업그레이드 항목을 구현한다",
            True,
            {"medium", "high"},
        ),
    )

    for label, task, expected_executed, allowed_complexity in cases:
        code, out, err = run_command(
            ["bash", "core/scripts/plan_dispatch.sh", task, "--auto", "--smoke"]
        )
        if code != 0:
            msg = (err or out or "plan dispatch smoke failed").splitlines()
            tail = msg[-1] if msg else "plan dispatch smoke failed"
            return CheckResult("plan-dispatch-smoke", "fail", f"{label}: {tail[:180]}")

        try:
            payload = json.loads(out)
        except Exception as exc:  # noqa: BLE001
            return CheckResult("plan-dispatch-smoke", "fail", f"{label}: invalid json ({exc})")

        dispatcher = payload.get("dispatcher")
        if not isinstance(dispatcher, dict):
            return CheckResult("plan-dispatch-smoke", "fail", f"{label}: missing dispatcher")

        executed = dispatcher.get("executed")
        complexity = str(dispatcher.get("complexity", "")).strip().lower()
        if bool(executed) != expected_executed:
            return CheckResult(
                "plan-dispatch-smoke",
                "fail",
                f"{label}: executed={executed} expected={expected_executed}",
            )
        if complexity not in allowed_complexity:
            return CheckResult(
                "plan-dispatch-smoke",
                "fail",
                f"{label}: complexity={complexity} expected one of {sorted(allowed_complexity)}",
            )

    return CheckResult("plan-dispatch-smoke", "pass", "simple and nontrivial auto paths verified")


def compute_score(results: List[CheckResult]) -> int:
    if not results:
        return 0
    points = 0.0
    for result in results:
        if result.status == "pass":
            points += 2
        elif result.status == "warn":
            points += 1
    return int(round((points / (len(results) * 2.0)) * 100))


def summarize(results: List[CheckResult]) -> str:
    passes = sum(1 for r in results if r.status == "pass")
    warns = sum(1 for r in results if r.status == "warn")
    fails = sum(1 for r in results if r.status == "fail")
    return f"pass={passes} warn={warns} fail={fails}"


def overall_status(results: List[CheckResult]) -> str:
    if any(r.status == "fail" for r in results):
        return "fail"
    if any(r.status == "warn" for r in results):
        return "warn"
    return "pass"


def build_payload(results: List[CheckResult]) -> Dict[str, object]:
    score = compute_score(results)
    overall = overall_status(results)
    return {
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "overall": overall,
        "score": score,
        "summary": summarize(results),
        "checks": [asdict(r) for r in results],
    }


def print_human(payload: Dict[str, object]) -> None:
    checks = payload.get("checks", [])
    icon = {"pass": "PASS", "warn": "WARN", "fail": "FAIL"}
    print("=== Skill Tool Audit ===")
    for check in checks:
        if not isinstance(check, dict):
            continue
        status = str(check.get("status", "")).lower()
        name = str(check.get("name", "unknown"))
        detail = str(check.get("detail", ""))
        print(f"{icon.get(status, 'INFO'):>4} {name:<24} {detail}")
    print(f"Score: {payload.get('score')}")
    print(f"Overall: {str(payload.get('overall', '')).upper()}")
    print(f"Summary: {payload.get('summary')}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Audit skill metadata and baseline tools")
    parser.add_argument("--json", action="store_true", help="Print JSON payload")
    args = parser.parse_args()

    contract = load_contract()
    results = [
        check_external_required_skills(contract),
        check_core_skill_frontmatter(),
        check_core_tool_files(contract),
        check_plan_dispatch_smoke_behavior(),
    ]
    payload = build_payload(results)

    if args.json:
        print(json.dumps(payload, ensure_ascii=False))
    else:
        print_human(payload)

    return 1 if payload.get("overall") == "fail" else 0


if __name__ == "__main__":
    sys.exit(main())
