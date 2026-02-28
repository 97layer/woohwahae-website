#!/usr/bin/env python3
"""
팀 모드 워커 브리핑 생성기.
리더 메모리(state.md)에서 태스크에 필요한 최소 컨텍스트만 추출해
워커에게 전달할 프롬프트를 만든다.
"""
import json
import sys
from pathlib import Path

ROOT = Path(__file__).parents[2]

CONTEXT_SECTIONS = {
    "infra": ["인프라 핵심", "실행 명령", "VM 서비스"],
    "pipeline": ["파이프라인", "P3", "P4"],
    "web": ["웹", "Cloudflare", "website", "build"],
    "content": ["콘텐츠 전략", "에세이", "어조"],
    "design": ["디자인", "시각", "CSS"],
    "agent": ["에이전트", "SA", "CE", "AD", "Gardener"],
}


def load_state() -> str:
    state_path = ROOT / "knowledge" / "agent_hub" / "state.md"
    if not state_path.exists():
        return ""
    return state_path.read_text(encoding="utf-8")


def extract_relevant_sections(state: str, context_keys: list[str]) -> str:
    if not context_keys:
        # 기본값: 현재 상태 + 인프라 섹션만
        context_keys = ["infra"]

    keywords = []
    for key in context_keys:
        keywords.extend(CONTEXT_SECTIONS.get(key, [key]))

    lines = state.split("\n")
    relevant_blocks = []
    capture = False
    current_block = []
    heading_level = 0

    for line in lines:
        # 헤딩 감지
        if line.startswith("#"):
            if current_block and capture:
                relevant_blocks.append("\n".join(current_block))
            current_block = [line]
            heading_level = len(line) - len(line.lstrip("#"))
            capture = any(kw.lower() in line.lower() for kw in keywords)
        else:
            current_block.append(line)

    if current_block and capture:
        relevant_blocks.append("\n".join(current_block))

    return "\n\n".join(relevant_blocks) if relevant_blocks else ""


def generate_worker_briefing(task: dict, state: str) -> str:
    context_keys = task.get("context_keys", ["infra"])
    relevant_context = extract_relevant_sections(state, context_keys)

    files_section = ""
    if task.get("files_to_read"):
        files_section = "\n## 읽어야 할 파일\n" + "\n".join(
            f"- {f}" for f in task["files_to_read"]
        )

    constraints_section = ""
    if task.get("constraints"):
        constraints_section = "\n## 제약 조건\n" + "\n".join(
            f"- {c}" for c in task["constraints"]
        )

    briefing = f"""# 워커 브리핑 — {task['id']}: {task['title']}

## 목표
{task['objective']}

## 관련 컨텍스트 (필요 부분만 추출됨)
{relevant_context if relevant_context else "없음 (state.md 미로드)"}
{files_section}
{constraints_section}

## 출력 형식
{task.get('output_format', '완료된 결과물 + 한 줄 요약')}

## 완료 후 반드시 포함할 것
- result_summary: 한 줄 결과 요약
- learnings: 리더에게 전달할 인사이트 목록 (없으면 빈 배열)
- changed_files: 수정/생성된 파일 경로 목록
"""
    return briefing.strip()


def update_task_result(task_id: str, result_summary: str, learnings: list[str]) -> None:
    queue_path = ROOT / "knowledge" / "system" / "team_queue.json"
    if not queue_path.exists():
        print("team_queue.json not found", file=sys.stderr)
        return

    with open(queue_path, encoding="utf-8") as f:
        queue = json.load(f)

    from datetime import datetime

    for task in queue["tasks"]:
        if task["id"] == task_id:
            task["status"] = "completed"
            task["completed_at"] = datetime.now().isoformat()
            task["result_summary"] = result_summary
            task["learnings"] = learnings
            queue["completed"].append(task_id)
            break

    # 리더 메모리 체크포인트 기록
    queue["leader_memory_checkpoints"].append({
        "task_id": task_id,
        "timestamp": datetime.now().isoformat(),
        "learnings": learnings,
    })

    with open(queue_path, "w", encoding="utf-8") as f:
        json.dump(queue, f, ensure_ascii=False, indent=2)

    print("task %s marked complete", task_id)


def print_next_task() -> None:
    queue_path = ROOT / "knowledge" / "system" / "team_queue.json"
    if not queue_path.exists():
        print("no queue")
        return

    with open(queue_path, encoding="utf-8") as f:
        queue = json.load(f)

    pending = [t for t in queue["tasks"] if t["status"] == "pending"]
    if not pending:
        print("ALL_DONE")
        return

    # priority 기준 정렬
    pending.sort(key=lambda t: t.get("priority", 99))
    next_task = pending[0]

    state = load_state()
    briefing = generate_worker_briefing(next_task, state)
    print(briefing)


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else "next"

    if cmd == "next":
        print_next_task()
    elif cmd == "complete" and len(sys.argv) >= 3:
        task_id = sys.argv[2]
        result = sys.argv[3] if len(sys.argv) > 3 else ""
        learnings = sys.argv[4:] if len(sys.argv) > 4 else []
        update_task_result(task_id, result, learnings)
    elif cmd == "briefing" and len(sys.argv) >= 3:
        task_id = sys.argv[2]
        with open(ROOT / "knowledge" / "system" / "team_queue.json", encoding="utf-8") as f:
            queue = json.load(f)
        task = next((t for t in queue["tasks"] if t["id"] == task_id), None)
        if task:
            state = load_state()
            print(generate_worker_briefing(task, state))
        else:
            print(f"task {task_id} not found", file=sys.stderr)
    else:
        print(f"usage: {sys.argv[0]} [next|complete <id> <summary> [learnings...]|briefing <id>]")
