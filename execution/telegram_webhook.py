"""
Telegram Webhook Server (Google Cloud Ready)
맥북 없이도 24/7 작동 가능한 Webhook 기반 텔레그램 봇

Google Cloud Run 배포 방법:
1. gcloud run deploy telegram-bot --source . --platform managed --region asia-northeast3 --allow-unauthenticated
2. 배포된 URL을 TELEGRAM_WEBHOOK_URL 환경변수로 설정
3. 자동으로 webhook 등록 및 메시지 수신

또는 Google Cloud Functions:
1. gcloud functions deploy telegram-webhook --runtime python39 --trigger-http --allow-unauthenticated
2. 배포된 URL을 TELEGRAM_WEBHOOK_URL 환경변수로 설정
"""

import os
import sys
import json
import logging
from datetime import datetime
from pathlib import Path
from flask import Flask, request, jsonify
from typing import Dict, Any

# Path Setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.append(str(PROJECT_ROOT))

# Lib Imports
try:
    from libs.ai_engine import AIEngine
    from libs.memory_manager import MemoryManager
    from libs.agent_router import AgentRouter
    from libs.gardener import Gardener
    from execution.system.log_error import ErrorLogger
    from execution.system.sync_status import SystemSynchronizer
    from libs.core_config import TELEGRAM_CONFIG, AI_MODEL_CONFIG
    from libs.notifier import Notifier
except ImportError as e:
    print(f"[CRITICAL] Import failed: {e}")
    sys.exit(1)

# Logging Setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Flask App
app = Flask(__name__)

# Config
TOKEN = TELEGRAM_CONFIG["BOT_TOKEN"]
WEBHOOK_URL = os.getenv("TELEGRAM_WEBHOOK_URL", "")

# Global instances
ai = None
memory = None
agent_router = None
gardener = None
error_logger = None
notifier = None

def init_services():
    """Initialize all services once at startup"""
    global ai, memory, agent_router, gardener, error_logger, notifier

    logger.info("Initializing services...")
    ai = AIEngine(AI_MODEL_CONFIG)
    memory = MemoryManager(str(PROJECT_ROOT))
    agent_router = AgentRouter(ai)
    gardener = Gardener(ai, memory, str(PROJECT_ROOT))
    error_logger = ErrorLogger()
    notifier = Notifier()
    logger.info("Services initialized successfully")

def setup_webhook():
    """Setup Telegram webhook"""
    if not WEBHOOK_URL:
        logger.warning("TELEGRAM_WEBHOOK_URL not set. Webhook not configured.")
        return False

    import urllib.request
    webhook_endpoint = f"{WEBHOOK_URL}/webhook"
    set_webhook_url = f"https://api.telegram.org/bot{TOKEN}/setWebhook?url={webhook_endpoint}"

    try:
        with urllib.request.urlopen(set_webhook_url) as response:
            result = json.loads(response.read().decode())
            if result.get("ok"):
                logger.info(f"✓ Webhook set successfully: {webhook_endpoint}")
                return True
            else:
                logger.error(f"✗ Webhook setup failed: {result}")
                return False
    except Exception as e:
        logger.error(f"✗ Webhook setup error: {e}")
        return False

def _get_project_context(trigger_text: str = "") -> str:
    """프로젝트 상태 요약 (토큰 최적화 버전)"""
    try:
        # Task Status
        status_file = PROJECT_ROOT / "task_status.json"
        status = json.loads(status_file.read_text()) if status_file.exists() else {}
        pending = status.get("pending_tasks", [])
        top_task = pending[0]['instruction'] if pending and 'instruction' in pending[0] else 'None'

        # System State (에이전트 상태)
        system_state_file = PROJECT_ROOT / "knowledge" / "system_state.json"
        system_state = json.loads(system_state_file.read_text()) if system_state_file.exists() else {}
        agents_status = system_state.get("agents", {})
        active_agents = [name for name, info in agents_status.items() if info.get("status") == "ACTIVE"]

        # Sync State (맥북/VM 주권)
        sync_state_file = PROJECT_ROOT / "knowledge" / "system" / "sync_state.json"
        sync_state = json.loads(sync_state_file.read_text()) if sync_state_file.exists() else {}
        active_node = sync_state.get("active_node", "unknown")

        vision_summary = "1인 기업 97LAYER의 고효율 자율 운영 시스템 (97LAYER OS)"

        context = f"""[System Status]
- Pending Tasks: {len(pending)}
- Active Node: {active_node}
- Active Agents: {", ".join(active_agents[:3])} ({len(active_agents)} total)
- Last Update: {system_state.get("last_update", "N/A")}

[Top Task]
{top_task}

[Vision]
{vision_summary}"""

        # Deep Grounding: 특정 키워드 시에만 최소 데이터 추가
        if trigger_text:
            keywords = ["안티그래비티", "antigravity", "rituals", "텔레그램", "진단", "diagnostic"]
            for kw in keywords:
                if kw.lower() in trigger_text.lower():
                    search_dirs = [PROJECT_ROOT / "directives", PROJECT_ROOT / "knowledge" / "reports"]
                    for s_dir in search_dirs:
                        if not s_dir.exists(): continue
                        matches = list(s_dir.rglob(f"*{kw}*"))
                        if matches:
                            try:
                                context += f"\n[Ref:{matches[0].name}] {matches[0].read_text(encoding='utf-8')[:300]}"
                                break
                            except: pass
        return context
    except Exception as e:
        return f"Context Error: {e}"

def process_message(message: Dict[str, Any]):
    """Process incoming Telegram message (shared logic with polling version)"""
    chat_id = message['chat']['id']
    text = message.get('text', '')

    if not text:
        return

    logger.info(f"Received from {chat_id}: {text[:50]}...")

    try:
        # 1. Save to Chat Memory
        memory.save_chat(str(chat_id), text)

        # 2. Intelligence Capture (Raw Signal to Inbox)
        if not text.startswith("/"):
            import re
            import subprocess

            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            inbox_path = PROJECT_ROOT / "knowledge" / "inbox" / f"raw_telegram_{timestamp}.md"
            inbox_path.parent.mkdir(parents=True, exist_ok=True)

            raw_content = f"""---
id: tg-{timestamp}
context: Telegram 수신 신호 (ChatID: {chat_id})
density: 2
last_curated: {datetime.now().strftime("%Y-%m-%d")}
---
# Raw Telegram Signal

{text}
"""
            with open(inbox_path, "w", encoding="utf-8") as f:
                f.write(raw_content)
            logger.info(f"Captured intelligence to inbox: {inbox_path.name}")

            # UIP (Unified Input Protocol) Automation
            youtube_match = re.search(r'(https?://(?:www\.)?youtube\.com/watch\?v=[0-9A-Za-z_-]{11}|https?://youtu\.be/[0-9A-Za-z_-]{11})', text)
            url_match = re.search(r'(https?://[^\s]+)', text)

            is_youtube = False
            script_path = None

            if youtube_match:
                url = youtube_match.group(0)
                is_youtube = True
                notifier.send_message(chat_id, "YouTube 신호가 감지되었습니다. UIP 가동을 시작합니다.")
                script_path = PROJECT_ROOT / "execution" / "youtube_parser.py"
            elif url_match:
                url = url_match.group(0)
                is_youtube = False
                notifier.send_message(chat_id, "외부 웹 신호가 감지되었습니다. UIP(Web Parser) 가동을 시작합니다.")
                script_path = PROJECT_ROOT / "execution" / "web_parser.py"
            else:
                url = None

            if url:
                try:
                    try:
                        parse_proc = subprocess.run([sys.executable, str(script_path), url], capture_output=True, text=True, timeout=30)
                    except subprocess.TimeoutExpired:
                        logger.error(f"UIP Parsing timed out for {url}")
                        notifier.send_message(chat_id, "UIP 분석 시간 초과. (Timeout: 30s)")
                        return

                    if parse_proc.returncode == 0:
                        parse_data = parse_proc.stdout

                        if is_youtube:
                            transform_script = PROJECT_ROOT / "execution" / "ontology_transform.py"
                            try:
                                trans_proc = subprocess.run([sys.executable, str(transform_script)], input=parse_data, capture_output=True, text=True, timeout=30)
                                if trans_proc.returncode == 0:
                                    res_data = json.loads(trans_proc.stdout)
                                    rs_id = res_data.get("id", "Unknown")
                                    notifier.send_message(chat_id, f"자산화가 완료되었습니다: {rs_id}\n\n고밀도 분석을 위해 Gemini Web에서 트랜스크립트를 추출하여 시스템에 Fuel 해주십시오.")
                                else:
                                    logger.error(f"UIP Transformation failed: {trans_proc.stderr}")
                                    notifier.send_message(chat_id, "UIP 변환 실패.")
                            except Exception as te:
                                logger.error(f"UIP Transformation error: {te}")
                        else:
                            try:
                                data = json.loads(parse_data)
                                title = data.get("metadata", {}).get("title", "No Title")
                                desc = data.get("metadata", {}).get("description", "")
                                rs_id = f"web-{datetime.now().strftime('%H%M%S')}"
                                summary_text = f"자산화가 완료되었습니다: {rs_id}\n\n제목: {title}\n요약: {desc}"
                                notifier.send_message(chat_id, summary_text)
                            except json.JSONDecodeError:
                                logger.error(f"UIP Parsing Output Invalid: {parse_data}")
                                notifier.send_message(chat_id, "UIP 분석 오류: 데이터 형식 불일치.")
                    else:
                        logger.error(f"UIP Parsing failed: {parse_proc.stderr}")
                        notifier.send_message(chat_id, "UIP 분석 실패.")
                except Exception as ue:
                    logger.error(f"UIP Workflow error: {ue}")

        # 3. Command Handling
        if text.startswith("/cd"):
            agent_router.set_agent("CD")
            notifier.send_message(chat_id, "Creative Director 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/td"):
            agent_router.set_agent("TD")
            notifier.send_message(chat_id, "Technical Director 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/ad"):
            agent_router.set_agent("AD")
            notifier.send_message(chat_id, "Art Director 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/ce"):
            agent_router.set_agent("CE")
            notifier.send_message(chat_id, "Chief Editor 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/sa"):
            agent_router.set_agent("SA")
            notifier.send_message(chat_id, "Strategy Analyst 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/auto"):
            agent_router.clear_agent()
            notifier.send_message(chat_id, "자동 라우팅 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/status"):
            try:
                from libs.system_guardian import SystemGuardian
                guardian = SystemGuardian(str(PROJECT_ROOT))
                report = guardian.get_system_report()

                status_file = PROJECT_ROOT / "task_status.json"
                if status_file.exists():
                    status_data = json.loads(status_file.read_text())
                    pending_count = len(status_data.get("pending_tasks", []))
                    last_active = status_data.get("last_active", "N/A")
                    report += f"\n\n[Task Summary]\n- Pending Tasks: {pending_count}\n- Last Heartbeat: {last_active}\n- Routing: {agent_router.get_status()}"

                notifier.send_message(chat_id, report)
            except Exception as e:
                notifier.send_message(chat_id, f"상태 확인 중 오류 발생: {e}")
            return
        elif text.startswith("/evolve"):
            notifier.send_message(chat_id, "The Gardener 가동... 지능 배양 중.")
            try:
                result = gardener.run_cycle()
                notifier.send_message(chat_id, f"진화 완료:\n{result}")
            except Exception as e:
                notifier.send_message(chat_id, f"진화 지연: {e}")
            return
        elif text.startswith("/start"):
            notifier.register_chat_id(chat_id)
            notifier.send_message(chat_id, "97LAYER OS Online (Webhook Mode).\n\n명령어:\n/cd /td /ad /ce /sa - 에이전트 전환\n/auto - 자동 라우팅\n/status - 상태 확인\n/evolve - 시스템 진화\n/council [주제] - 위원회 소집")
            return
        elif text.startswith("/council"):
            topic = text[8:].strip()
            if not topic:
                notifier.send_message(chat_id, "사용법: /council [주제]")
                return
            notifier.send_message(chat_id, f"위원회를 소집합니다...\n주제: {topic}")
            try:
                from libs.synapse import Synapse
                synapse = Synapse(ai)
                result = synapse.council_meeting(topic)
                notifier.send_message(chat_id, f"위원회 결론:\n{result[:1000]}")
            except Exception as e:
                notifier.send_message(chat_id, f"위원회 오류: {e}")
            return

        # 4. Neural Routing (AI Response)
        agent_key = agent_router.route(text)

        is_complex = (
            len(text) > 30 or
            any(k in text for k in [
                "분석", "보고", "설계", "구현", "정리",
                "진단", "확인", "문제", "플로우", "구조",
                "어떻게", "왜", "뭐", "상태", "현황"
            ])
        )
        project_context = _get_project_context(text if is_complex else "")

        # Increased history limit for better continuity
        chat_history = memory.load_chat(str(chat_id), limit=30 if is_complex else 15)
        history_text = "\n".join([f"{m['role'][0].upper()}: {m['content']}" for m in chat_history])

        user_prompt = f"""[System Context]
{project_context}

[Recent Conversation Log]
{history_text}

[Current User Input]
{text}
"""

        agent_persona = agent_router.get_persona(agent_key)
        system_instruction = (
            f"You are {agent_key} of 97LAYER OS. Maintain your core identity and collaborate with the user.\n\n"
            f"Core Persona:\n{agent_persona}\n\n"
            "Communication Protocol:\n"
            "- Acknowledge the context in [Log] to maintain conversation continuity.\n"
            "- Speak naturally in Korean (Professional but casual).\n"
            "- Be precise and insightful. Don't repeat yourself.\n"
            "- If the user's input refers to previous messages, explicitly reference the context found in the [Log].\n"
            "- Focus on the user's ultimate goal of building 97LAYER."
        )

        response = ai.generate_response(user_prompt, system_instruction=system_instruction)
        clean_response = response.replace("**", "").replace("◈", "").strip()

        notifier.send_message(chat_id, clean_response)
        memory.save_chat(str(chat_id), clean_response, role="assistant")

    except Exception as e:
        logger.error(f"Processing error: {e}")
        error_logger.log_error("telegram_webhook.py", e, {"chat_id": chat_id, "text": text})

@app.route('/webhook', methods=['POST'])
def webhook():
    """Webhook endpoint for Telegram"""
    try:
        update = request.get_json()

        # Update heartbeat
        syncer = SystemSynchronizer(agent_name="Telegram_Bot_Cloud")
        syncer.report_heartbeat(status="ONLINE", current_task="메시지 처리 중")

        if "message" in update:
            process_message(update["message"])

        return jsonify({"ok": True})
    except Exception as e:
        logger.error(f"Webhook error: {e}")
        return jsonify({"ok": False, "error": str(e)}), 500

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "service": "97LAYER Telegram Webhook"
    })

@app.route('/', methods=['GET'])
def index():
    """Root endpoint"""
    return jsonify({
        "service": "97LAYER Telegram Webhook",
        "status": "online",
        "webhook_configured": bool(WEBHOOK_URL)
    })

if __name__ == "__main__":
    logger.info("◈ 97LAYER Telegram Webhook Server Starting...")

    # Initialize services
    init_services()

    # Setup webhook
    if WEBHOOK_URL:
        setup_webhook()
    else:
        logger.warning("⚠️ TELEGRAM_WEBHOOK_URL not set. Webhook will not be configured.")
        logger.warning("   Set this environment variable to enable cloud deployment.")

    # Start Flask server
    port = int(os.getenv("PORT", 8080))
    app.run(host="0.0.0.0", port=port, debug=False)
