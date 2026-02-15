
import os
import sys
import json
import time
import logging
import urllib.request
import urllib.parse
from pathlib import Path
from datetime import datetime, timedelta

# Project Root Setup
PROJECT_ROOT = Path(__file__).resolve().parent.parent
if str(PROJECT_ROOT) not in sys.path:
    sys.path.append(str(PROJECT_ROOT))

from libs.ai_engine import AIEngine
from libs.core_config import AI_MODEL_CONFIG

# Logging Setup
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("GardenerDigest")

# Constants
INBOX_DIR = PROJECT_ROOT / "knowledge" / "inbox"
ARCHIVE_DIR = PROJECT_ROOT / "knowledge" / "archive"
DAILY_BRIEF_DIR = PROJECT_ROOT / "knowledge" / "daily_briefs"

# Ensure directories exist
INBOX_DIR.mkdir(parents=True, exist_ok=True)
ARCHIVE_DIR.mkdir(parents=True, exist_ok=True)
DAILY_BRIEF_DIR.mkdir(parents=True, exist_ok=True)

def load_telegram_config():
    """Load Telegram config from .env or .env.txt"""
    token = os.getenv("TELEGRAM_BOT_TOKEN")
    chat_id = os.getenv("TELEGRAM_CHAT_ID")
    
    if not token or not chat_id:
        # Fallback manual load
        env_files = [PROJECT_ROOT / ".env", PROJECT_ROOT / ".env.txt", PROJECT_ROOT / "env.txt"]
        for env in env_files:
            if env.exists():
                try:
                    with open(env, "r", encoding="utf-8") as f:
                        for line in f:
                            if line.startswith("TELEGRAM_BOT_TOKEN="):
                                token = line.split("=", 1)[1].strip().strip('"').strip("'")
                            elif line.startswith("TELEGRAM_CHAT_ID="):
                                chat_id = line.split("=", 1)[1].strip().strip('"').strip("'")
                except Exception:
                    pass
    return token, chat_id

def send_telegram_message(token, chat_id, text):
    """Send message via Telegram API (urllib)"""
    if not token or not chat_id:
        logger.error("Telegram credentials missing.")
        return

    url = f"https://api.telegram.org/bot{token}/sendMessage"
    payload = {
        "chat_id": chat_id,
        "text": text,
        "parse_mode": "Markdown"
    }
    
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"})
    
    try:
        with urllib.request.urlopen(req) as response:
            logger.info("Daily Digest sent successfully.")
    except Exception as e:
        logger.error(f"Failed to send Telegram message: {e}")

def run_daily_digest():
    """Main execution flow"""
    logger.info("Starting Daily Gardener Digest...")
    
    # 1. Collect signals (All files in inbox)
    # In a real scenario, we might filter by date, but for now we process everything in inbox
    files = list(INBOX_DIR.glob("*.md"))
    if not files:
        logger.info("No signals in inbox to process.")
        return

    logger.info(f"Found {len(files)} signals in inbox.")
    
    combined_content = ""
    for file in files:
        try:
            with open(file, "r", encoding="utf-8") as f:
                content = f.read()
                combined_content += f"\n\n--- Source: {file.name} ---\n{content}"
        except Exception as e:
            logger.error(f"Error reading {file.name}: {e}")

    # 2. AI Summarization
    engine = AIEngine(AI_MODEL_CONFIG)
    
    prompt = f"""
    당신은 97LAYER의 지능형 정원사(The Gardener)입니다.
    아래는 'Inbox'에 수집된 원시 신호들(Raw Signals)입니다.
    이 내용들을 방금 수확한 작물처럼 다루어, 가장 본질적이고 가치 있는 통찰(Insight)만 남기도록 정제하십시오.
    
    [요구사항]
    1. **Essence Only**: 불필요한 잡음(Noise)은 모두 제거하고 핵심만 남깁니다.
    2. **Strategic Link**: 각 정보가 97LAYER의 철학(Minimalism, Essentialism)과 어떻게 연결되는지 한 줄로 평론하십시오.
    3. **Tone**: 차분하고 지적인 어조 (The Narrator 페르소나).
    4. **Format**: 마크다운 리스트 또는 짧은 문단.
    
    [Raw Signals start]
    {combined_content[:30000]} 
    [Raw Signals end]
    
    출력 형식:
    # 🌿 Daily Essence Report ({datetime.now().strftime('%Y-%m-%d')})
    
    ## 1. Key Insights
    (요약 내용)
    
    ## 2. Actionable Items
    (실행 제안)
    """
    
    try:
        digest = engine.generate_response(prompt)
    except Exception as e:
        logger.error(f"AI Generation failed: {e}")
        return

    # 3. Save Digest
    digest_filename = f"digest_{datetime.now().strftime('%Y%m%d')}.md"
    digest_path = DAILY_BRIEF_DIR / digest_filename
    with open(digest_path, "w", encoding="utf-8") as f:
        f.write(digest)
    
    logger.info(f"Digest saved to {digest_path}")

    # 4. Send to Telegram
    token, chat_id = load_telegram_config()
    header = f"🌿 **Good Morning.**\nHere is your Daily Essence for {datetime.now().strftime('%Y-%m-%d')}.\n\n"
    
    # Telegram has msg length limit (4096). Truncate if needed or just send summary.
    send_content = header + digest
    if len(send_content) > 4000:
        send_content = send_content[:4000] + "\n...(Truncated)"
        
    send_telegram_message(token, chat_id, send_content)
    
    # 5. Archive Processed Files (Optional - Moving to Archive)
    # For safety, we currently just log. Or implementation plan says "Archive Policy" to be reviewed.
    # Moving them to archive/inbox_processed_{date} to keep inbox clean?
    # Let's CREATE a subfolder in archive
    archive_sub = ARCHIVE_DIR / f"processed_{datetime.now().strftime('%Y%m%d')}"
    archive_sub.mkdir(exist_ok=True)
    
    for file in files:
        try:
            target = archive_sub / file.name
            file.rename(target)
            logger.info(f"Archived: {file.name}")
        except Exception as e:
            logger.error(f"Failed to archive {file.name}: {e}")

if __name__ == "__main__":
    run_daily_digest()
