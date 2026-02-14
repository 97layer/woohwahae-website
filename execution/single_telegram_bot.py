#!/usr/bin/env python3
"""
Single Instance Telegram Bot - 409 Conflict 방지 버전
하나의 인스턴스만 실행되도록 보장
"""

import os
import sys
import json
import time
import signal
import fcntl
import urllib.request
import urllib.error
from pathlib import Path
from datetime import datetime

# 프로젝트 설정
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

from libs.core_config import TELEGRAM_CONFIG

# 로깅 설정
import logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 설정
TOKEN = TELEGRAM_CONFIG["BOT_TOKEN"]
BASE_URL = f"https://api.telegram.org/bot{TOKEN}"

# Lock 파일로 단일 인스턴스 보장
LOCK_FILE = PROJECT_ROOT / ".tmp" / "telegram_bot.lock"


class SingletonBot:
    """단일 인스턴스 텔레그램 봇"""

    def __init__(self):
        self.running = True
        self.offset = None
        self.lock_file = None

        # Lock 디렉토리 생성
        LOCK_FILE.parent.mkdir(parents=True, exist_ok=True)

    def acquire_lock(self):
        """Lock 획득 (단일 인스턴스 보장)"""
        try:
            self.lock_file = open(LOCK_FILE, 'w')
            fcntl.lockf(self.lock_file, fcntl.LOCK_EX | fcntl.LOCK_NB)
            self.lock_file.write(str(os.getpid()))
            self.lock_file.flush()
            logger.info(f"✅ Lock 획득 (PID: {os.getpid()})")
            return True
        except IOError:
            logger.error("❌ 다른 인스턴스가 이미 실행 중입니다")
            return False

    def release_lock(self):
        """Lock 해제"""
        if self.lock_file:
            fcntl.lockf(self.lock_file, fcntl.LOCK_UN)
            self.lock_file.close()
            try:
                os.remove(LOCK_FILE)
            except:
                pass
            logger.info("Lock 해제됨")

    def clear_updates(self):
        """대기 중인 업데이트 모두 클리어"""
        logger.info("대기 중인 업데이트 클리어 중...")
        try:
            # 타임아웃 0으로 즉시 가져오기
            url = f"{BASE_URL}/getUpdates?timeout=0"
            with urllib.request.urlopen(url, timeout=5) as response:
                result = json.loads(response.read())
                updates = result.get("result", [])

                if updates:
                    # 마지막 update_id로 모두 확인 처리
                    last_id = updates[-1]["update_id"] + 1
                    confirm_url = f"{BASE_URL}/getUpdates?offset={last_id}&timeout=0"
                    urllib.request.urlopen(confirm_url, timeout=5)
                    logger.info(f"✅ {len(updates)}개 업데이트 클리어")
                    self.offset = last_id
                else:
                    logger.info("대기 중인 업데이트 없음")

        except Exception as e:
            logger.error(f"업데이트 클리어 실패: {e}")

    def run(self):
        """봇 실행"""
        # Lock 획득
        if not self.acquire_lock():
            return

        # 시그널 핸들러
        signal.signal(signal.SIGINT, self.signal_handler)
        signal.signal(signal.SIGTERM, self.signal_handler)

        try:
            # 초기 업데이트 클리어
            self.clear_updates()

            logger.info("🤖 봇 시작됨! 텔레그램에서 메시지를 보내보세요.")
            logger.info("종료하려면 Ctrl+C")

            # 메인 루프
            while self.running:
                try:
                    self.poll_updates()
                except KeyboardInterrupt:
                    break
                except Exception as e:
                    logger.error(f"폴링 에러: {e}")
                    time.sleep(5)

        finally:
            self.release_lock()
            logger.info("봇 종료됨")

    def poll_updates(self):
        """업데이트 폴링"""
        try:
            # Long polling (30초 대기)
            url = f"{BASE_URL}/getUpdates?timeout=30"
            if self.offset:
                url += f"&offset={self.offset}"

            req = urllib.request.Request(url)
            with urllib.request.urlopen(req, timeout=35) as response:
                result = json.loads(response.read())

                if result.get("ok"):
                    updates = result.get("result", [])

                    for update in updates:
                        self.offset = update["update_id"] + 1
                        self.process_update(update)

        except urllib.error.HTTPError as e:
            if "409" in str(e):
                logger.error("⚠️ 409 Conflict - 다른 곳에서 봇이 실행 중입니다!")
                logger.info("GCP나 다른 터미널에서 봇을 중지하세요.")
                time.sleep(10)
            else:
                logger.error(f"HTTP 에러: {e}")
                time.sleep(5)

        except urllib.error.URLError as e:
            logger.error(f"네트워크 에러: {e}")
            time.sleep(5)

        except Exception as e:
            logger.error(f"알 수 없는 에러: {e}")
            time.sleep(5)

    def process_update(self, update):
        """업데이트 처리"""
        try:
            # 메시지 처리
            if "message" in update:
                message = update["message"]
                chat_id = message["chat"]["id"]
                text = message.get("text", "")

                logger.info(f"📩 받은 메시지: {text[:50]}")

                # 응답 생성
                if text.startswith("/start"):
                    response = "🤖 97LAYER OS 봇이 활성화되었습니다!\n\n명령어:\n/status - 상태 확인\n/help - 도움말"
                elif text.startswith("/status"):
                    response = f"✅ 봇 정상 작동 중\n시간: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
                elif text.startswith("/help"):
                    response = "사용 가능한 명령어:\n/start - 시작\n/status - 상태\n/help - 도움말"
                else:
                    response = f"메시지 받음: {text}"

                # 응답 전송
                self.send_message(chat_id, response)

        except Exception as e:
            logger.error(f"업데이트 처리 에러: {e}")

    def send_message(self, chat_id, text):
        """메시지 전송"""
        try:
            url = f"{BASE_URL}/sendMessage"
            data = json.dumps({
                "chat_id": chat_id,
                "text": text
            }).encode('utf-8')

            req = urllib.request.Request(url, data=data)
            req.add_header('Content-Type', 'application/json')

            with urllib.request.urlopen(req) as response:
                result = json.loads(response.read())
                if result.get("ok"):
                    logger.info(f"✅ 메시지 전송됨")
                else:
                    logger.error(f"메시지 전송 실패: {result}")

        except Exception as e:
            logger.error(f"메시지 전송 에러: {e}")

    def signal_handler(self, signum, frame):
        """시그널 핸들러"""
        logger.info("\n종료 신호 받음...")
        self.running = False


def main():
    """메인 함수"""
    print("=" * 60)
    print("Single Instance Telegram Bot")
    print("=" * 60)

    bot = SingletonBot()
    bot.run()


if __name__ == "__main__":
    main()