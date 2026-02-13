# Filename: execution/telegram_daemon.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered & Portable Mode)

import os
import sys
import time
import urllib.request
import json
import logging
import subprocess
import re
import socket
from datetime import datetime
from pathlib import Path
from typing import Optional, Dict, Any

# Path Setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.append(str(PROJECT_ROOT))

# Lib Imports (Dynamic loading to handle environment variations)
try:
    from libs.ai_engine import AIEngine
    from libs.memory_manager import MemoryManager
    from libs.agent_router import AgentRouter
    from libs.gardener import Gardener
    from execution.system.log_error import ErrorLogger
    from execution.system.sync_status import SystemSynchronizer
    from libs.core_config import TELEGRAM_CONFIG, AI_MODEL_CONFIG
except ImportError as e:
    print(f"[CRITICAL] Import failed: {e}")
    sys.exit(1)

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Config
TOKEN = TELEGRAM_CONFIG["BOT_TOKEN"]
BASE_URL = f"https://api.telegram.org/bot{TOKEN}"

from libs.notifier import Notifier
notifier = Notifier()

# Global Placeholders (Initialized in main)
ai = None
memory = None
agent_router = None
gardener = None
error_logger = None

def _register_chat_id(chat_id: int):
    return notifier.register_chat_id(chat_id)

def send_message(chat_id: int, text: str):
    return notifier.send_message(chat_id, text)

def _get_project_context(trigger_text: str = "") -> str:
    """
    프로젝트 상태 요약 (토큰 최적화 버전)
    """
    try:
        from pathlib import Path
        import json
        status_file = PROJECT_ROOT / "task_status.json"
        status = json.loads(status_file.read_text()) if status_file.exists() else {}
        
        pending = status.get("pending_tasks", [])
        top_task = pending[0]['instruction'] if pending and 'instruction' in pending[0] else 'None'
        
        # 정적 데이터 요약 처리
        vision_summary = "1인 기업 97LAYER의 고효율 자율 운영 시스템 (97LAYER OS)"
        
        context = f"[Status] Pending: {len(pending)} | Top: {top_task} | Vision: {vision_summary}"
        
        # Deep Grounding: 특정 키워드 시에만 최소 데이터 추가
        if trigger_text:
            keywords = ["안티그래비티", "antigravity", "rituals", "텔레그램"]
            for kw in keywords:
                if kw.lower() in trigger_text.lower():
                    search_dirs = [PROJECT_ROOT / "directives"]
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
    chat_id = message['chat']['id']
    text = message.get('text', '')
    
    if not text: return
    
    logger.info(f"Received from {chat_id}: {text[:50]}...")
    
    try:
        # 1. Save to Chat Memory (Persistence)
        memory.save_chat(str(chat_id), text)

        # 2. Intelligence Capture (Raw Signal to Inbox)
        if not text.startswith("/"):
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

            # --- UIP (Unified Input Protocol) Automation ---
            youtube_match = re.search(r'(https?://(?:www\.)?youtube\.com/watch\?v=[0-9A-Za-z_-]{11}|https?://youtu\.be/[0-9A-Za-z_-]{11})', text)
            url_match = re.search(r'(https?://[^\s]+)', text) # Generic URL Capture
            
            is_youtube = False
            script_path = None

            if youtube_match:
                url = youtube_match.group(0)
                is_youtube = True
                send_message(chat_id, "YouTube 신호가 감지되었습니다. UIP 가동을 시작합니다.")
                script_path = PROJECT_ROOT / "execution" / "youtube_parser.py"
            elif url_match:
                url = url_match.group(0)
                is_youtube = False
                send_message(chat_id, "외부 웹 신호가 감지되었습니다. UIP(Web Parser) 가동을 시작합니다.")
                script_path = PROJECT_ROOT / "execution" / "web_parser.py"
            else:
                url = None

            if url:
                try:
                    # 1. Parsing
                    try:
                        # Common execution for both parsers
                        parse_proc = subprocess.run([sys.executable, str(script_path), url], capture_output=True, text=True, timeout=30)
                    except subprocess.TimeoutExpired:
                        logger.error(f"UIP Parsing timed out for {url}")
                        send_message(chat_id, "UIP 분석 시간 초과. (Timeout: 30s)")
                        return
                    
                    if parse_proc.returncode == 0:
                        parse_data = parse_proc.stdout
                        
                        if is_youtube:
                            # 2. Transformation (Ontology) for YouTube
                            transform_script = PROJECT_ROOT / "execution" / "ontology_transform.py"
                            try:
                                trans_proc = subprocess.run([sys.executable, str(transform_script)], input=parse_data, capture_output=True, text=True, timeout=30)
                                if trans_proc.returncode == 0:
                                    res_data = json.loads(trans_proc.stdout)
                                    rs_id = res_data.get("id", "Unknown")
                                    send_message(chat_id, f"자산화가 완료되었습니다: {rs_id}\n\n고밀도 분석을 위해 Gemini Web에서 트랜스크립트를 추출하여 시스템에 Fuel 해주십시오.")
                                else:
                                    logger.error(f"UIP Transformation failed: {trans_proc.stderr}")
                                    send_message(chat_id, "UIP 변환 실패.")
                            except Exception as te:
                                logger.error(f"UIP Transformation error: {te}")
                        else:
                            # Direct Summary for General Web
                            try:
                                data = json.loads(parse_data)
                                title = data.get("metadata", {}).get("title", "No Title")
                                desc = data.get("metadata", {}).get("description", "")
                                rs_id = f"web-{datetime.now().strftime('%H%M%S')}"
                                summary_text = f"자산화가 완료되었습니다: {rs_id}\n\n제목: {title}\n요약: {desc}"
                                send_message(chat_id, summary_text)
                            except json.JSONDecodeError:
                                logger.error(f"UIP Parsing Output Invalid: {parse_data}")
                                send_message(chat_id, "UIP 분석 오류: 데이터 형식 불일치.")
                    else:
                        logger.error(f"UIP Parsing failed: {parse_proc.stderr}")
                        send_message(chat_id, "UIP 분석 실패.")
                except Exception as ue:
                    logger.error(f"UIP Workflow error: {ue}")
            # -----------------------------------------------

        # 3. Command Handling
        if text.startswith("/cd"):
            agent_router.set_agent("CD")
            send_message(chat_id, "Creative Director 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/td"):
            agent_router.set_agent("TD")
            send_message(chat_id, "Technical Director 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/ad"):
            agent_router.set_agent("AD")
            send_message(chat_id, "Art Director 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/ce"):
            agent_router.set_agent("CE")
            send_message(chat_id, "Chief Editor 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/sa"):
            agent_router.set_agent("SA")
            send_message(chat_id, "Strategy Analyst 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/auto"):
            agent_router.clear_agent()
            send_message(chat_id, "자동 라우팅 모드 활성화됨.\n" + agent_router.get_status())
            return
        elif text.startswith("/status"):
            try:
                from libs.system_guardian import SystemGuardian
                guardian = SystemGuardian(str(PROJECT_ROOT))
                report = guardian.get_system_report()
                
                # Add task status summary
                status_file = PROJECT_ROOT / "task_status.json"
                if status_file.exists():
                    status_data = json.loads(status_file.read_text())
                    pending_count = len(status_data.get("pending_tasks", []))
                    last_active = status_data.get("last_active", "N/A")
                    report += f"\n\n[Task Summary]\n- Pending Tasks: {pending_count}\n- Last Heartbeat: {last_active}\n- Routing: {agent_router.get_status()}"
                
                send_message(chat_id, report)
            except Exception as e:
                send_message(chat_id, f"상태 확인 중 오류 발생: {e}")
            return
        elif text.startswith("/evolve"):
            send_message(chat_id, "The Gardener 가동... 지능 배양 중.")
            try:
                result = gardener.run_cycle()
                send_message(chat_id, f"진화 완료:\n{result}")
            except Exception as e:
                send_message(chat_id, f"진화 지연: {e}")
            return
        elif text.startswith("/start"):
            # chat_id를 task_status에 자동 등록 (자율 태스크 보고 수신)
            _register_chat_id(chat_id)
            send_message(chat_id, "97LAYER OS Online.\n\n명령어:\n/cd /td /ad /ce /sa - 에이전트 전환\n/auto - 자동 라우팅\n/status - 상태 확인\n/evolve - 시스템 진화\n/council [주제] - 위원회 소집")
            return
        elif text.startswith("/council"):
            topic = text[8:].strip()
            if not topic:
                send_message(chat_id, "사용법: /council [주제]")
                return
            send_message(chat_id, f"위원회를 소집합니다...\n주제: {topic}")
            try:
                from libs.synapse import Synapse
                synapse = Synapse(ai)
                result = synapse.council_meeting(topic)
                send_message(chat_id, f"위원회 결론:\n{result[:1000]}")
            except Exception as e:
                send_message(chat_id, f"위원회 오류: {e}")
            return

        # 4. Neural Routing (AI Response)
        agent_key = agent_router.route(text)
        
        # 1. Dynamic Context Intensity (입력 분석에 따른 가변 주입)
        # 텍스트가 길거나 특정 키워드가 포함된 경우 'Deep Grounding' 수행
        is_complex = len(text) > 50 or any(k in text for k in ["분석", "보고", "설계", "구현", "정리"])
        project_context = _get_project_context(text if is_complex else "")
        
        # 2. History Compression (역할 식별자만 남기고 압축)
        chat_history = memory.load_chat(str(chat_id), limit=3 if not is_complex else 5)
        history_text = "\n".join([f"{m['role'][0].upper()}: {m['content'][:200]}" for m in chat_history])
        
        # 3. High-Density Prompt Build
        user_prompt = f"[Reality]\n{project_context}\n\n[Log]\n{history_text}\n\n[Input]\n{text}"
        
        # 4. Hierarchical System Instruction
        agent_persona = agent_router.get_persona(agent_key)
        system_instruction = (
            f"ID: 97LAYER OS ({agent_key})\n"
            f"Directive: {agent_persona}\n\n"
            "Execution: 1.Zero-Fluff(Start immediately). 2.Cold-Logic(No apology/empathy). "
            "3.Context-Anchor(Evidence-based only). 4.Density-Control(Short for simple, Deep for complex)."
        )

        response = ai.generate_response(user_prompt, system_instruction=system_instruction)
        
        # Post-process for zero noise
        clean_response = response.replace("**", "").replace("◈", "").strip()
        
        send_message(chat_id, clean_response)
        memory.save_chat(str(chat_id), response, role="assistant")
        
    except Exception as e:
        logger.error(f"Processing error: {e}")
        error_logger.log_error("telegram_daemon.py", e, {"chat_id": chat_id, "text": text})

def main():
    logger.info("◈ 97LAYER Telegram Daemon Starting (urllib Portable Mode)...")
    
    # Global Instances initialization within main
    global ai, memory, agent_router, gardener, error_logger
    ai = AIEngine(AI_MODEL_CONFIG)
    memory = MemoryManager(str(PROJECT_ROOT))
    agent_router = AgentRouter(ai)
    gardener = Gardener(ai, memory, str(PROJECT_ROOT))
    error_logger = ErrorLogger()

    # Load initial offset from task_status
    status_file = PROJECT_ROOT / "task_status.json"
    initial_offset = None
    try:
        if status_file.exists():
            status = json.loads(status_file.read_text())
            initial_offset = status.get("last_telegram_update_id")
    except Exception as e:
        logger.error(f"Failed to load initial offset: {e}")

    offset = initial_offset
    last_digest_date = None
    retry_delay = 1

    while True:
        try:
            # 1. 생존 신고 (Heartbeat)
            bot_syncer = SystemSynchronizer(agent_name="Telegram_Bot_Cloud")
            bot_syncer.report_heartbeat(status="ONLINE", current_task="메시지 수신 대기")
            # --- Scheduler: Gardener Daily Digest (06:00 AM) ---
            now = datetime.now()
            current_date_str = now.strftime("%Y-%m-%d")
            
            # Check if it's 6 AM (or close to it) AND we haven't run digest today
            # Also runs once if start time is past 6 AM to catch up, but let's stick to hour check for stability
            # or just ensure it runs once per day if hour >= 6
            if now.hour == 6 and last_digest_date != current_date_str:
                logger.info(f"⏰ Triggering Daily Gardener Digest for {current_date_str}")
                digest_script = PROJECT_ROOT / "execution" / "gardener_daily_digest.py"
                try:
                    # Run in background via subprocess with timeout to prevent zombie
                    subprocess.Popen([sys.executable, str(digest_script)]) 
                    last_digest_date = current_date_str
                    logger.info("Daily Digest process launched.")
                except Exception as e:
                    logger.error(f"Failed to launch Daily Digest: {e}")

            # --- Telegram Polling ---
            url = f"{BASE_URL}/getUpdates"
            # Increased timeout to 50s and urllib timeout to 60s for stability
            if offset:
                url += f"?offset={offset}&timeout=50"
            else:
                url += "?timeout=50"
            
            try:
                with urllib.request.urlopen(url, timeout=60) as response:
                    res = json.loads(response.read().decode())
                # Reset error counter on success
                retry_delay = 1
            except (urllib.error.URLError, ConnectionResetError, TimeoutError, socket.timeout if 'socket' in globals() else Exception) as e:
                 # Reduce log noise for normal polling timeouts
                 if "timed out" in str(e).lower():
                     logger.debug("Polling timeout (normal behavior)")
                 else:
                     logger.error(f"Polling network error: {e}. Retrying in {retry_delay}s...")
                 time.sleep(retry_delay)
                 retry_delay = min(retry_delay * 2, 60)
                 continue 
            
            for update in res.get("result", []):
                offset = update["update_id"] + 1
                
                # Persist offset to task_status.json
                try:
                    status = json.loads(status_file.read_text()) if status_file.exists() else {}
                    status["last_telegram_update_id"] = offset
                    status_file.write_text(json.dumps(status, indent=2, ensure_ascii=False))
                except Exception as e:
                    logger.error(f"Failed to persist offset: {e}")

                if "message" in update:
                    process_message(update["message"])
                    
            time.sleep(1)
        except KeyboardInterrupt: break
        except Exception as e:
            logger.error(f"Main Loop unexpected error: {e}") # Improved logging message
            time.sleep(5)

if __name__ == "__main__":
    main()
