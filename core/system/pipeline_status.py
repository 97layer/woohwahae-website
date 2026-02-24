#!/usr/bin/env python3
"""
pipeline_status.py — LAYER OS 파이프라인 현황 한 번에 출력

에이전트가 작업 시작 전 컨텍스트를 빠르게 파악하기 위한 도구.

Usage:
    python3 core/system/pipeline_status.py
    python3 core/system/pipeline_status.py --json
"""
import argparse
import json
import subprocess
from datetime import datetime, timezone
from pathlib import Path


PROJECT_ROOT = Path(__file__).parent.parent.parent
KNOWLEDGE = PROJECT_ROOT / "knowledge"


def _read_jsons(path: Path) -> list[dict]:
    result = []
    for f in path.glob("*.json"):
        try:
            data = json.loads(f.read_text(encoding="utf-8"))
            if isinstance(data, dict):
                result.append(data)
            elif isinstance(data, list):
                result.extend([d for d in data if isinstance(d, dict)])
        except Exception:
            pass
    return result


def signal_stats() -> dict:
    signals = _read_jsons(KNOWLEDGE / "signals")
    by_status: dict[str, int] = {}
    by_type: dict[str, int] = {}
    recent = []

    for s in signals:
        st = s.get("status", "unknown")
        tp = s.get("type", "unknown")
        by_status[st] = by_status.get(st, 0) + 1
        by_type[tp] = by_type.get(tp, 0) + 1
        if st == "captured":
            recent.append(s.get("signal_id", "?"))

    return {
        "total": len(signals),
        "by_status": by_status,
        "by_type": by_type,
        "pending": by_status.get("captured", 0),
        "pending_ids": recent[:5],
    }


def corpus_stats() -> dict:
    entries = _read_jsons(KNOWLEDGE / "corpus" / "entries")
    published = sum(1 for e in entries if e.get("status") == "published")
    return {
        "total": len(entries),
        "published": published,
        "draft": len(entries) - published,
    }


def client_stats() -> dict:
    clients = _read_jsons(KNOWLEDGE / "clients")
    now = datetime.now()
    due = []
    for c in clients:
        next_visit = c.get("next_visit_date")
        if next_visit:
            try:
                nv = datetime.fromisoformat(next_visit)
                if nv <= now:
                    due.append(c.get("name", "?"))
            except Exception:
                pass
    return {
        "total": len(clients),
        "due_count": len(due),
        "due_names": due,
    }


def growth_stats() -> dict:
    reports = sorted(
        (KNOWLEDGE / "reports" / "growth").glob("*.json"),
        key=lambda f: f.stat().st_mtime,
        reverse=True,
    )
    if not reports:
        return {}
    try:
        latest = json.loads(reports[0].read_text(encoding="utf-8"))
        return {
            "period": latest.get("period", "?"),
            "revenue": latest.get("revenue", {}).get("total", 0),
            "recorded_at": latest.get("recorded_at", "?"),
        }
    except Exception:
        return {}


def vm_status() -> dict:
    services = [
        "97layer-telegram",
        "97layer-ecosystem",
        "97layer-gardener",
        "woohwahae-backend",
    ]
    try:
        result = subprocess.run(
            ["ssh", "97layer-vm",
             f"systemctl is-active {' '.join(services)}"],
            capture_output=True, text=True, timeout=10,
        )
        lines = result.stdout.strip().splitlines()
        return {svc: (lines[i] if i < len(lines) else "unknown")
                for i, svc in enumerate(services)}
    except Exception:
        return {svc: "unreachable" for svc in services}


def quanta_age() -> str:
    quanta = PROJECT_ROOT / "knowledge" / "agent_hub" / "INTELLIGENCE_QUANTA.md"
    if not quanta.exists():
        return "없음"
    mtime = quanta.stat().st_mtime
    diff = int((datetime.now().timestamp() - mtime) / 60)
    if diff < 60:
        return f"{diff}분 전"
    return f"{diff // 60}시간 {diff % 60}분 전"


def print_status(data: dict) -> None:
    sig = data["signals"]
    corp = data["corpus"]
    cli = data["clients"]
    gr = data["growth"]
    vm = data["vm"]

    print("━━━ LAYER OS 파이프라인 현황 ━━━")
    print(f"  {datetime.now().strftime('%Y-%m-%d %H:%M')} | QUANTA: {data['quanta_age']}")
    print()

    # 신호
    pending = sig["pending"]
    print(f"📥 신호  총 {sig['total']}개  |  "
          f"대기 {pending}개  |  "
          f"분석됨 {sig['by_status'].get('analyzed', 0)}개")
    if pending > 0:
        print(f"   └ 미처리: {', '.join(sig['pending_ids'])}"
              + ("..." if pending > 5 else ""))

    # Corpus
    print(f"📚 Corpus  entries {corp['total']}개  |  발행 {corp['published']}개  |  draft {corp['draft']}개")

    # 고객
    due = cli["due_count"]
    print(f"💇 고객  {cli['total']}명  |  재방문 알림 {due}명"
          + (f"  ← {', '.join(cli['due_names'])}" if due else ""))

    # Growth
    if gr:
        print(f"📈 Growth  {gr['period']}  수익 {gr['revenue']:,}원  ({gr['recorded_at'][:10]})")

    # VM
    print()
    status_icons = {"active": "✅", "inactive": "❌", "failed": "🔴", "unreachable": "⚠️"}
    for svc, st in vm.items():
        icon = status_icons.get(st, "❓")
        short = svc.replace("97layer-", "").replace("woohwahae-", "wh-")
        print(f"  {icon} {short}: {st}")

    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")


def main() -> None:
    parser = argparse.ArgumentParser(description="LAYER OS 파이프라인 현황")
    parser.add_argument("--json", action="store_true", help="JSON 출력")
    parser.add_argument("--no-vm", action="store_true", help="VM 상태 스킵 (빠른 모드)")
    args = parser.parse_args()

    data = {
        "signals": signal_stats(),
        "corpus": corpus_stats(),
        "clients": client_stats(),
        "growth": growth_stats(),
        "vm": vm_status() if not args.no_vm else {},
        "quanta_age": quanta_age(),
        "generated_at": datetime.now().isoformat(),
    }

    if args.json:
        print(json.dumps(data, ensure_ascii=False, indent=2))
    else:
        print_status(data)


if __name__ == "__main__":
    main()
