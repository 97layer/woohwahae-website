# Filename: libs/notifier.py
# Purpose: Shared notification utility for 97LAYER OS

import json
import urllib.request
import logging
from pathlib import Path
from typing import List, Optional

PROJECT_ROOT = Path(__file__).resolve().parent.parent
STATUS_FILE = PROJECT_ROOT / "task_status.json"

class Notifier:
    def __init__(self):
        from libs.core_config import TELEGRAM_CONFIG
        self.token = TELEGRAM_CONFIG.get("BOT_TOKEN")
        self.base_url = f"https://api.telegram.org/bot{self.token}"
        self.logger = logging.getLogger("Notifier")

    def register_chat_id(self, chat_id: int):
        """Persists chat_id to task_status.json for broadcasting with simple retry."""
        import time
        for _ in range(3):
            try:
                status = {}
                if STATUS_FILE.exists():
                    status = json.loads(STATUS_FILE.read_text())
                
                ids = status.get("telegram_chat_ids", [])
                if chat_id not in ids:
                    ids.append(chat_id)
                    status["telegram_chat_ids"] = ids
                    STATUS_FILE.write_text(json.dumps(status, indent=4, ensure_ascii=False))
                    self.logger.info(f"Chat ID {chat_id} registered.")
                return
            except Exception as e:
                self.logger.warning(f"Registration retry due to: {e}")
                time.sleep(0.5)

    def send_message(self, chat_id: int, text: str):
        """Sends a direct message to a specific chat ID."""
        if not self.token: return
        url = f"{self.base_url}/sendMessage"
        payload = json.dumps({"chat_id": chat_id, "text": text}).encode("utf-8")
        req = urllib.request.Request(url, data=payload, headers={"Content-Type": "application/json"})
        try:
            with urllib.request.urlopen(req, timeout=10) as response:
                return json.loads(response.read().decode())
        except Exception as e:
            self.logger.error(f"Telegram Send Error: {e}")
            return None

    def broadcast(self, text: str):
        """Sends a message to all registered chat IDs in task_status.json."""
        if not self.token or not STATUS_FILE.exists(): return
        try:
            status = json.loads(STATUS_FILE.read_text())
            ids = status.get("telegram_chat_ids", [])
            for cid in ids:
                self.send_message(cid, text)
        except Exception as e:
            self.logger.error(f"Broadcast Error: {e}")

    def send_message_to_admin(self, text: str):
        """Sends a message to admin (broadcasts to all registered chat IDs)."""
        self.broadcast(text)
