import time
import json
import os
import sys
from datetime import datetime
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent
TASK_FILE = BASE_DIR / "task_status.json"

# 에이전트 이름 → 키 매핑 (모듈 레벨 상수)
AGENT_KEY_MAP = {
    "Strategy_Analyst": "SA",
    "Creative_Director": "CD",
    "Technical_Director": "TD",
    "Chief_Editor": "CE",
    "Art_Director": "AD",
}

# libs/ 모듈 접근을 위해 프로젝트 루트를 sys.path에 추가
if str(BASE_DIR) not in sys.path:
    sys.path.insert(0, str(BASE_DIR))

# Lazy-load AI to avoid circular deps
_ai = None
_router = None
_telegram_token = None

def _get_ai():
    global _ai
    if _ai is None:
        import sys
        sys.path.insert(0, str(BASE_DIR))
        from libs.ai_engine import AIEngine
        from libs.core_config import AI_MODEL_CONFIG
        _ai = AIEngine(AI_MODEL_CONFIG)
    return _ai

def _get_syncer():
    from execution.system.sync_status import SystemSynchronizer
    return SystemSynchronizer(agent_name="Technical_Director")

def _get_router():
    global _router
    if _router is None:
        from libs.agent_router import AgentRouter
        _router = AgentRouter(_get_ai())
    return _router

def _get_token():
    global _telegram_token
    if _telegram_token is None:
        from libs.core_config import TELEGRAM_CONFIG
        _telegram_token = TELEGRAM_CONFIG["BOT_TOKEN"]
    return _telegram_token

def _broadcast_to_telegram(text: str):
    """Broadcasting utility using Notifier."""
    try:
        from libs.notifier import Notifier
        notifier = Notifier()
        notifier.broadcast(text)
    except Exception as e:
        print(f"[Telegram Error] {e}")

def _check_rituals(status: dict):
    """Checks RITUALS_CONFIG and triggers tasks if conditions met."""
    from libs.core_config import RITUALS_CONFIG
    
    now = datetime.now()
    current_hour = now.hour
    current_weekday = now.weekday()  # 0=Monday, 6=Sunday
    today_str = now.strftime("%Y-%m-%d")
    
    # Ensure 'rituals_log' exists in status
    if "rituals_log" not in status:
        status["rituals_log"] = {}

    for ritual_name, config in RITUALS_CONFIG.items():
        # Check if already done today
        last_run = status["rituals_log"].get(ritual_name)
        if last_run == today_str:
            continue
            
        # Check Time Conditions
        trigger_hour = config.get("trigger_hour")
        trigger_weekday = config.get("trigger_weekday")
        
        # Hour check (Trigger if current hour matches)
        if trigger_hour is not None and current_hour != trigger_hour:
            continue
            
        # Weekday check (If specified)
        if trigger_weekday is not None and current_weekday != trigger_weekday:
            continue
            
        # Trigger Condition Met!
        print(f"[{datetime.now()}] [Ritual] Triggering {ritual_name}...")
        
        new_task = {
            "id": f"ritual_{ritual_name}_{int(time.time())}",
            "type": config.get("task_type", "GENERAL"),
            "agent": config.get("agent", "System"),
            "instruction": config.get("instruction", ""),
            "council": config.get("council", False)
        }
        
        status.setdefault("pending_tasks", []).append(new_task)
        status["rituals_log"][ritual_name] = today_str
        
        # Notify via Telegram about Ritual Start
        _broadcast_to_telegram(f"🕯️ [Ritual Started] {ritual_name}\n{config.get('instruction')[:50]}...")

def _get_chat_ids() -> list:
    """task_status에서 알림 받을 chat_id 목록 조회"""
    status = _load_status()
    return status.get("telegram_chat_ids", [])

def _load_status() -> dict:
    if not TASK_FILE.exists():
        return {"pending_tasks": [], "completed_tasks": []}
    with open(TASK_FILE, 'r', encoding='utf-8') as f:
        return json.load(f)

def _save_status(status: dict):
    status["last_active"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    with open(TASK_FILE, 'w', encoding='utf-8') as f:
        json.dump(status, f, indent=4, ensure_ascii=False)

def _handle_consolidation(task: dict) -> str:
    """
    Handles the NIGHTLY_CONSOLIDATION task.
    Aggregates all raw signals from the last 24h and generates a pattern update.
    """
    try:
        raw_signals_dir = BASE_DIR / "knowledge" / "raw_signals"
        today_str = datetime.now().strftime("%Y-%m-%d")
        
        # Find files modified today
        recent_signals = []
        for f in raw_signals_dir.glob("*.md"):
            if f.stat().st_mtime > time.time() - 86400:  # Last 24 hours
                with open(f, "r", encoding="utf-8") as rf:
                    content = rf.read()
                    recent_signals.append(f"FileName: {f.name}\n{content[:2000]}") # Truncate per file
        
        if not recent_signals:
            return "No new signals found in the last 24 hours."

        aggregated_content = "\n---\n".join(recent_signals)
        
        instruction = task.get("instruction", "")
        prompt = f"""
        {instruction}

        [Collected Raw Signals (Last 24h)]
        {aggregated_content}
        
        [Output Logic]
        1. Summarize key themes.
        2. Identify recurring patterns.
        3. Suggest strategic actions for tomorrow.
        """
        
        ai = _get_ai()
        result = ai.generate_response(prompt)
        
        # Save to patterns
        pattern_file = BASE_DIR / "knowledge" / "patterns" / f"daily_insight_{today_str}.md"
        pattern_file.parent.mkdir(parents=True, exist_ok=True)
        
        with open(pattern_file, "w", encoding="utf-8") as pf:
            pf.write(f"# Daily Insight ({today_str})\n\n{result}")
            
        return f"Consolidation Complete. Saved to {pattern_file}"

    except Exception as e:
        return f"Consolidation Failed: {e}"

def _get_shared_memory() -> str:
    """
    최근 수집된 지식 패턴(Patterns)에서 핵심 요약을 추출
    """
    try:
        patterns_dir = BASE_DIR / "knowledge" / "patterns"
        if not patterns_dir.exists():
            return "No shared patterns available."
        
        recent_patterns = sorted(patterns_dir.glob("*.md"), key=os.path.getmtime, reverse=True)[:3]
        summaries = []
        for f in recent_patterns:
            with open(f, "r", encoding="utf-8") as pf:
                # 첫 500자 혹은 요약 섹션 추출
                content = pf.read()
                summaries.append(f"[{f.name}] {content[:500]}")
        
        return "\n".join(summaries)
    except Exception as e:
        return f"Shared memory retrieval failed: {e}"

def _get_project_context() -> str:
    """
    현재 프로젝트의 실제 상태(Task, Vision, Knowledge 등)를 수집하여 지침 주입용 텍스트 생성
    """
    try:
        from pathlib import Path
        import json
        
        status = _load_status()
        pending = status.get("pending_tasks", [])
        completed = status.get("completed_tasks", [])
        
        vision_path = BASE_DIR / "VISION.md"
        vision = ""
        if vision_path.exists():
            with open(vision_path, "r", encoding="utf-8") as f:
                vision = f.read()[:500]
        
        assets_dir = BASE_DIR / "knowledge" / "assets"
        recent_assets = []
        if assets_dir.exists():
            for f in sorted(assets_dir.rglob("*.md"), key=os.path.getmtime, reverse=True)[:5]:
                recent_assets.append(f.name)

        shared_memory = _get_shared_memory()

        context = f"""
        [Current Project Reality]
        - Vision Summary: {vision}
        - Pending Tasks: {len(pending)} (First: {pending[0]['instruction'] if pending else 'None'})
        - Recently Completed: {completed[-3:] if completed else 'None'}
        - Recent Knowledge Assets: {', '.join(recent_assets)}
        
        [Shared Memory / Recent Patterns]
        {shared_memory}
        """
        return context
    except Exception as e:
        return f"Context collection failed: {e}"

def execute_agent(task: dict) -> str:
    """
    실제 에이전트 LLM 호출
    """
    # Special Handler for Skill Execution
    if task.get("type") == "SKILL":
        return _handle_skill_execution(task)
    # Special Handler for Consolidation
    elif task.get("type") == "CONSOLIDATION":
        return _handle_consolidation(task)
    elif task.get("type") == "AUTONOMOUS_DEV":
        res = _handle_autonomous_dev(task)
        _broadcast_to_telegram(f"🤖 자율 시스템 진화 보고:\n\n{res}")
        return res
    elif task.get("type") == "DIAGNOSTIC":
        res = _handle_diagnostic(task)
        if "❌" in res or "⚠️" in res:
            _broadcast_to_telegram(f"🛡️ 가디언 긴급 점검 리포트:\n\n{res}")
        return res
    elif task.get("type") == "PUBLISH_CHECK":
        res = _handle_publish_check(task)
        if res and ("자동 폐기" in res or "CD 결정 필요" in res):
            _broadcast_to_telegram(f"⏰ [72h Rule]\n\n{res}")
        return res
    elif task.get("type") == "INSTAGRAM_PUBLISH":
        res = _handle_instagram_publish(task)
        if res and "발행 완료" in res:
            _broadcast_to_telegram(f"📸 [Instagram 발행]\n\n{res}")
        return res
    elif task.get("type") == "INSIGHT":
        # Insight tasks are handled by standard execute_agent flow but generate a proactive report
        pass

    agent_name = task.get("agent", "CD")
    agent_key = AGENT_KEY_MAP.get(agent_name, "CD")
    instruction = task.get("instruction", "")

    router = _get_router()
    system_prompt = router.build_system_prompt(agent_key)
    ai = _get_ai()

    # Context Grounding (Hallucination 방지)
    project_context = _get_project_context()
    grounded_instruction = f"""
    {project_context}

    [Instruction]
    {instruction}

    위의 [Current Project Reality]를 참고하여 사실에 기반한 답변을 하십시오.
    만약 현재 프로젝트와 관계없는 내용(예: Athena, Hermes 등 가공의 이름)을 지어내지 마십시오.
    """

    print(f"[{datetime.now()}] [{agent_key}] 태스크 실행: {instruction[:60]}...")

    result = ai.generate_response(
        prompt=grounded_instruction,
        system_instruction=system_prompt
    )

    print(f"[{datetime.now()}] [{agent_key}] 완료.")
    return result

def _handle_autonomous_dev(task: dict) -> str:
    """Runs the autonomous_developer.py script."""
    try:
        import subprocess
        script_path = BASE_DIR / "execution" / "autonomous_developer.py"
        result = subprocess.run([sys.executable, str(script_path)], capture_output=True, text=True)
        if result.returncode == 0:
            return f"Autonomous Development Cycle Complete.\n{result.stdout}"
        else:
            return f"Autonomous Development Failed: {result.stderr}"
    except Exception as e:
        return f"Autonomous Development Error: {e}"

def _handle_diagnostic(task: dict) -> str:
    """Runs common diagnostics via SystemGuardian."""
    try:
        from libs.system_guardian import SystemGuardian
        guardian = SystemGuardian(str(BASE_DIR))
        return guardian.get_system_report()
    except Exception as e:
        return f"Diagnostic Failed: {e}"

def _handle_publish_check(task: dict) -> str:
    """Runs auto_publisher to check 72h rule."""
    try:
        import sys
        script_path = BASE_DIR / "execution" / "auto_publisher.py"

        # Import and run directly for better control
        sys.path.insert(0, str(BASE_DIR / "execution"))
        from auto_publisher import AutoPublisher

        publisher = AutoPublisher()
        violations = publisher.check_72h_rule()

        if not violations:
            return ""

        # Process violations
        result_lines = []
        for v in violations:
            if v["status"] == "violation":
                # 76h+ auto discard
                publisher.auto_discard(v["path"])
                result_lines.append(f"🚨 자동 폐기: {v['file']} ({v['elapsed_hours']}h)")
            else:
                # 72-76h warning
                result_lines.append(f"⚠️ CD 결정 필요: {v['file']} ({v['elapsed_hours']}h)")

        # Generate CD notification
        notification = publisher.notify_cd(violations)
        return "\n".join(result_lines) + "\n\n" + notification

    except Exception as e:
        return f"Publish Check Failed: {e}"

def _handle_instagram_publish(task: dict) -> str:
    """Runs instagram_publisher to publish scheduled content."""
    try:
        import subprocess
        script_path = BASE_DIR / "execution" / "instagram_publisher.py"

        result = subprocess.run(
            [sys.executable, str(script_path)],
            capture_output=True,
            text=True,
            timeout=300  # 5 min timeout
        )

        if result.returncode == 0:
            return f"Instagram 발행 프로세스 완료.\n{result.stdout}"
        else:
            return f"Instagram 발행 실패: {result.stderr}"

    except subprocess.TimeoutExpired:
        return "Instagram 발행 타임아웃 (5분 초과)"
    except Exception as e:
        return f"Instagram Publish Failed: {e}"

def _handle_skill_execution(task: dict) -> str:
    """Executes a skill based on task specification."""
    try:
        from libs.skill_engine import SkillEngine
        skill_engine = SkillEngine()

        skill_id = task.get("skill_id")
        context = task.get("context", {})

        if not skill_id:
            return "Skill execution failed: No skill_id provided"

        print(f"[{datetime.now()}] [SKILL] Executing: {skill_id}")
        result = skill_engine.execute_skill(skill_id, context)

        if result.get("status") == "success":
            output_msg = f"Skill [{skill_id}] executed successfully.\n"
            if result.get("output_file"):
                output_msg += f"Output: {result['output_file']}"
            print(f"[{datetime.now()}] [SKILL] Success: {skill_id}")
            return output_msg
        else:
            error_msg = f"Skill [{skill_id}] failed: {result.get('message')}"
            print(f"[{datetime.now()}] [SKILL] Failed: {error_msg}")
            return error_msg

    except Exception as e:
        return f"Skill Execution Error: {e}"


def council_on_task(task: dict) -> str:
    """
    복잡한 태스크는 Synapse council_meeting으로 처리
    """
    from libs.synapse import Synapse
    synapse = Synapse(_get_ai())
    topic = task.get("instruction", "")
    print(f"[{datetime.now()}] [COUNCIL] 다중 에이전트 토론 시작: {topic[:60]}...")
    return synapse.council_meeting(topic)

def check_system_entropy():
    """시스템 상태 점검 및 태스크 실행"""
    try:
        # Heartbeat for Dashboard
        syncer = _get_syncer()
        syncer.report_heartbeat(status="ACTIVE", current_task="시스템 상태 점검 및 태스크 스캔")

        # -1. 72h Rule Check (Every Loop - Non-blocking)
        try:
            _handle_publish_check({})
        except Exception as pc_e:
            print(f"[Publish Check Error] {pc_e}")

        # 0. System Guardian: Self-Diagnostic
        try:
            from libs.system_guardian import SystemGuardian
            guardian = SystemGuardian(str(BASE_DIR))
            health_report = guardian.get_system_report()
            
            # If any critical daemon is down, notify or log error
            if "❌" in health_report:
                print(f"[Guardian Alert] {health_report}")
                # We can choose to notify user or attempt restart here
                # For now, let's just log it and potentially send a mini-alert
            
            # Update health in status
            status = _load_status()
            status["system_health"] = health_report
            _save_status(status)
        except Exception as sg_e:
            print(f"[Guardian Error] {sg_e}")

        # 0.5. Ingestion Loop (Gatekeeper) - Picking up new insights FIRST
        try:
            from execution import ingest_gatekeeper
            ingest_gatekeeper.process_inbox()
        except Exception as ig_e:
            print(f"[Gatekeeper Error] {ig_e}")

        # 1. Load Status AFTER ingestion to pick up circular actions
        status = _load_status()
        
        # Update heartbeat even if no tasks
        status["last_active"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        _save_status(status)

        # Check for Rituals (Scheduled Tasks)
        _check_rituals(status)
        _save_status(status)

        # 대기 태스크 실행
        pending = status.get("pending_tasks", [])
        if not pending:
            print(f"[{datetime.now()}] [Standby] 대기 중인 태스크 없음.")
            return

        task = pending[0]
        use_council = task.get("council", False)

        # 실제 LLM 실행
        if use_council:
            result = council_on_task(task)
        else:
            result = execute_agent(task)
            _archive_result(task, result)

        # 태스크 완료 처리
        status["pending_tasks"].pop(0)
        status.setdefault("completed_tasks", []).append(f"{task['id']}_done")
        if use_council:
            status["last_council_result"] = result[:500]
        status["last_active"] = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        _save_status(status)

        # 텔레그램 보고 (고지능 브리핑 포맷)
        report_prompt = f"""
        97LAYER의 오케스트레이터로서 다음 수행 결과를 사용자에게 보고하십시오.
        
        [수험 결과]
        에이전트: {task.get('agent')}
        타입: {task.get('type')}
        내용: {result[:1200]}
        
        [지침]
        1. '타입:', '담당:', '◈', '[ ]' 와 같은 기계적 또는 상징적인 머리말을 절대 사용하지 마십시오.
        2. 답변은 즉시 자연스러운 문장(Narrative)으로 시작하십시오.
        3. 냉철하고 권위 있는 어조로 핵심만 전달하되, 사실에 기반하여 보고하십시오.
        4. 이 결과가 프로젝트의 흐름상 어떤 의미를 갖는지 1문장으로 해석을 덧붙이십시오.
        5. 볼드(**)를 사용하지 마십시오.
        """
        briefing = _get_ai().generate_response(report_prompt)
        
        _broadcast_to_telegram(briefing)

        print(f"[{datetime.now()}] [Done] 태스크 완료 및 보고 전송.")

    except Exception as e:
        print(f"[{datetime.now()}] [Error] {e}")

def _archive_result(task: dict, content: str):
    """Saves the result of an autonomous task as a markdown asset."""
    try:
        task_type = task.get("type", "GENERAL").lower()
        date_str = datetime.now().strftime("%Y%m%d")
        timestamp = int(time.time())
        
        # Determine path based on type
        save_dir = BASE_DIR / "knowledge" / "assets" / task_type
        save_dir.mkdir(parents=True, exist_ok=True)
        
        filename = f"{task_type}_{date_str}_{timestamp}.md"
        filepath = save_dir / filename
        
        # Frontmatter + Content
        file_content = f"""---
id: {task.get('id')}
date: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}
type: {task.get('type')}
agent: {task.get('agent')}
---

# {task.get('type')} Report

{content}
"""
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(file_content)
            
        print(f"[{datetime.now()}] [Archive] Saved asset to {filepath}")
        
    except Exception as e:
        print(f"[{datetime.now()}] [Archive Error] {e}")

def main_loop():
    print(f"[{datetime.now()}] === 97LAYER Technical Daemon (LLM Connected) Started ===")
    check_system_entropy()
    # Initial run for testing
    while True:
        try:
            time.sleep(600)  # 10분 대기
            check_system_entropy()
        except KeyboardInterrupt:
            print("Daemon Stopped.")
            break
        except Exception as e:
            print(f"Loop Error: {e}")
            time.sleep(60)

if __name__ == "__main__":
    main_loop()
