#!/usr/bin/env python3
"""
채팅 메모리 실시간 모니터링
5시 7분 이후 대화만 표시
"""

import json
import time
from datetime import datetime
from pathlib import Path

CHAT_MEMORY_FILE = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "7565534667.json"
TARGET_TIME = datetime(2026, 2, 14, 17, 7, 0)  # 5시 7분

def monitor_chat():
    """채팅 모니터링"""
    print(f"📊 채팅 모니터링 시작 (5시 7분 이후 대화)")
    print("=" * 60)

    last_count = 0

    while True:
        try:
            if CHAT_MEMORY_FILE.exists():
                with open(CHAT_MEMORY_FILE, 'r', encoding='utf-8') as f:
                    messages = json.load(f)

                # 5시 7분 이후 메시지만 필터링
                recent_messages = []
                for msg in messages:
                    try:
                        msg_time = datetime.fromisoformat(msg['timestamp'].replace('Z', '+00:00'))
                        if msg_time > TARGET_TIME:
                            recent_messages.append(msg)
                    except:
                        continue

                # 새 메시지가 있으면 표시
                if len(recent_messages) > last_count:
                    for msg in recent_messages[last_count:]:
                        timestamp = msg['timestamp'][:19].replace('T', ' ')
                        role = "👤 User" if msg['role'] == 'user' else "🤖 Bot"
                        content = msg['content'][:100] + "..." if len(msg['content']) > 100 else msg['content']

                        print(f"\n[{timestamp}] {role}")
                        print(f"  {content}")
                        print("-" * 60)

                    last_count = len(recent_messages)

                # 상태 표시
                print(f"\r📊 모니터링 중... (5시 7분 이후 메시지: {len(recent_messages)}개)", end="", flush=True)

            time.sleep(2)  # 2초마다 체크

        except KeyboardInterrupt:
            print("\n\n모니터링 종료")
            break
        except Exception as e:
            print(f"\n오류: {e}")
            time.sleep(5)

if __name__ == "__main__":
    monitor_chat()