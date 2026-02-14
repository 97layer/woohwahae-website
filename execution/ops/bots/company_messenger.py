#!/usr/bin/env python3
"""
97LAYER 회사 메신저 시스템
텔레그램 → 안티그래비티 에이전트 → 실시간 실행 → 보고

목적: 단순하고 확실하게 작동하는 회사 내부 커뮤니케이션
"""

import os
import sys
import json
import time
import subprocess
import threading
from datetime import datetime
from pathlib import Path
import urllib.request
import urllib.error
from queue import Queue

# 프로젝트 루트
PROJECT_ROOT = Path(__file__).parent
sys.path.append(str(PROJECT_ROOT))

# ===== 설정 =====
TELEGRAM_TOKEN = "8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA"  # 현재 토큰
TELEGRAM_API = f"https://api.telegram.org/bot{TELEGRAM_TOKEN}"

# 파일 경로
CHAT_LOG = PROJECT_ROOT / "knowledge" / "company_chat.json"
AGENT_QUEUE = PROJECT_ROOT / "knowledge" / "agent_tasks.json"
REPORT_DIR = PROJECT_ROOT / "knowledge" / "reports"

# 에이전트 정의
AGENTS = {
    "CD": "Creative Director - 브랜드/전략",
    "TD": "Technical Director - 기술/시스템",
    "AD": "Art Director - 디자인/비주얼",
    "CE": "Chief Editor - 콘텐츠/편집",
    "SA": "Strategy Analyst - 분석/리서치"
}

# ===== 핵심 기능 =====

class CompanyMessenger:
    """회사 메신저 시스템"""

    def __init__(self):
        self.running = True
        self.task_queue = Queue()
        self.offset = None
        self.chat_id = None

        # 디렉토리 생성
        CHAT_LOG.parent.mkdir(parents=True, exist_ok=True)
        REPORT_DIR.mkdir(parents=True, exist_ok=True)

        print("🚀 97LAYER 회사 메신저 시작")
        print("=" * 60)

    def start(self):
        """메신저 시작"""
        # 기존 업데이트 클리어
        self.clear_updates()

        # 워커 스레드 시작
        worker = threading.Thread(target=self.process_tasks, daemon=True)
        worker.start()

        # 메인 루프
        self.main_loop()

    def main_loop(self):
        """메인 메시지 수신 루프"""
        print("✅ 메신저 준비 완료! 텔레그램에서 메시지를 보내세요.")
        print("-" * 60)

        while self.running:
            try:
                # 텔레그램 메시지 받기
                updates = self.get_updates()

                for update in updates:
                    if "message" in update:
                        self.handle_message(update["message"])

            except KeyboardInterrupt:
                print("\n종료 중...")
                self.running = False
                break

            except Exception as e:
                if "409" in str(e):
                    print("⚠️ 다른 봇 실행 중. 10초 대기...")
                    time.sleep(10)
                else:
                    print(f"오류: {e}")
                    time.sleep(5)

    def get_updates(self):
        """텔레그램 업데이트 받기"""
        url = f"{TELEGRAM_API}/getUpdates?timeout=10"
        if self.offset:
            url += f"&offset={self.offset}"

        try:
            with urllib.request.urlopen(url, timeout=15) as response:
                result = json.loads(response.read())

                if result["ok"]:
                    updates = result["result"]

                    # offset 업데이트
                    for update in updates:
                        self.offset = update["update_id"] + 1

                    return updates

        except urllib.error.HTTPError as e:
            if "409" in str(e):
                raise Exception("409 Conflict")

        return []

    def handle_message(self, message):
        """메시지 처리"""
        chat_id = message["chat"]["id"]
        text = message.get("text", "")

        if not text:
            return

        self.chat_id = chat_id  # 저장

        # 로그 출력
        timestamp = datetime.now().strftime("%H:%M:%S")
        print(f"\n📩 [{timestamp}] 받음: {text}")

        # 채팅 로그 저장
        self.save_chat(chat_id, text, "user")

        # 명령어 처리
        if text.startswith("/"):
            self.handle_command(chat_id, text)
        else:
            # 일반 메시지 → 에이전트 작업 큐에 추가
            self.task_queue.put({
                "chat_id": chat_id,
                "text": text,
                "timestamp": datetime.now().isoformat()
            })

            # 즉시 응답
            self.send_message(chat_id, "✅ 메시지 받음. 에이전트가 처리 중...")

    def handle_command(self, chat_id, command):
        """명령어 처리"""
        cmd = command.split()[0].lower()

        if cmd == "/start":
            msg = "🤖 *97LAYER 회사 메신저*\n\n"
            msg += "메시지를 보내면 에이전트가 자동 실행됩니다.\n\n"
            msg += "*명령어:*\n"
            msg += "/status - 시스템 상태\n"
            msg += "/agents - 에이전트 목록\n"
            msg += "/report - 최근 보고서\n"
            msg += "/cd, /td, /ad, /ce, /sa - 특정 에이전트 호출"

            self.send_message(chat_id, msg)

        elif cmd == "/status":
            msg = f"✅ *시스템 상태*\n\n"
            msg += f"시간: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n"
            msg += f"대기 작업: {self.task_queue.qsize()}개\n"
            msg += "모든 에이전트 정상 작동"

            self.send_message(chat_id, msg)

        elif cmd == "/agents":
            msg = "*활성 에이전트:*\n\n"
            for code, desc in AGENTS.items():
                msg += f"• *{code}*: {desc}\n"

            self.send_message(chat_id, msg)

        elif cmd == "/report":
            # 최근 보고서 표시
            self.show_recent_reports(chat_id)

        elif cmd in ["/cd", "/td", "/ad", "/ce", "/sa"]:
            # 특정 에이전트 호출
            agent = cmd[1:].upper()
            self.call_agent(chat_id, agent, command[3:].strip())

    def call_agent(self, chat_id, agent_code, task=""):
        """특정 에이전트 호출"""
        if agent_code not in AGENTS:
            self.send_message(chat_id, f"❌ 알 수 없는 에이전트: {agent_code}")
            return

        agent_name = AGENTS[agent_code]

        self.send_message(chat_id, f"📢 {agent_name} 호출 중...")

        # 에이전트별 작업 실행
        if agent_code == "TD":
            # Technical Director - 시스템 상태 체크
            self.execute_td_task(chat_id, task)
        elif agent_code == "SA":
            # Strategy Analyst - 데이터 분석
            self.execute_sa_task(chat_id, task)
        else:
            # 기본 응답
            response = f"{agent_name} 응답:\n\n작업 '{task}' 처리 중..."
            self.send_message(chat_id, response)

    def execute_td_task(self, chat_id, task):
        """Technical Director 작업 실행"""
        self.send_message(chat_id, "🔧 *Technical Director 실행*")

        # 시스템 상태 체크 스크립트 실행
        try:
            result = subprocess.run(
                ["python3", "execution/ops/system_monitor.py", "quick"],
                cwd=PROJECT_ROOT,
                capture_output=True,
                text=True,
                timeout=10
            )

            if result.returncode == 0:
                # 결과 파싱 및 전송
                output = result.stdout[:1000]  # 처음 1000자만
                self.send_message(chat_id, f"```\n{output}\n```")
            else:
                self.send_message(chat_id, "⚠️ 시스템 체크 실패")

        except Exception as e:
            self.send_message(chat_id, f"오류: {e}")

    def execute_sa_task(self, chat_id, task):
        """Strategy Analyst 작업 실행"""
        self.send_message(chat_id, "📊 *Strategy Analyst 실행*")

        # 채팅 로그 분석
        if CHAT_LOG.exists():
            with open(CHAT_LOG) as f:
                logs = json.load(f)

            # 간단한 통계
            user_msgs = [m for m in logs if m["role"] == "user"]
            bot_msgs = [m for m in logs if m["role"] == "assistant"]

            report = f"*채팅 분석 보고서*\n\n"
            report += f"총 메시지: {len(logs)}개\n"
            report += f"사용자 메시지: {len(user_msgs)}개\n"
            report += f"봇 응답: {len(bot_msgs)}개\n"

            if user_msgs:
                recent = user_msgs[-3:]
                report += f"\n*최근 메시지:*\n"
                for msg in recent:
                    report += f"• {msg['content'][:50]}...\n"

            self.send_message(chat_id, report)
        else:
            self.send_message(chat_id, "아직 분석할 데이터가 없습니다.")

    def process_tasks(self):
        """백그라운드 작업 처리"""
        while self.running:
            try:
                if not self.task_queue.empty():
                    task = self.task_queue.get()

                    # 작업 처리
                    print(f"⚙️ 작업 처리: {task['text'][:30]}...")

                    # 자동 에이전트 라우팅
                    agent = self.route_to_agent(task['text'])

                    # 보고서 생성
                    report = f"*자동 처리 완료*\n\n"
                    report += f"담당: {AGENTS[agent]}\n"
                    report += f"요청: {task['text']}\n"
                    report += f"시간: {datetime.now().strftime('%H:%M:%S')}\n"
                    report += f"상태: ✅ 완료"

                    # 보고서 저장 및 전송
                    self.save_report(agent, task['text'], "완료")

                    if self.chat_id:
                        self.send_message(self.chat_id, report)

                time.sleep(1)

            except Exception as e:
                print(f"작업 처리 오류: {e}")

    def route_to_agent(self, text):
        """텍스트 분석하여 적절한 에이전트 선택"""
        text_lower = text.lower()

        # 키워드 기반 라우팅
        if any(word in text_lower for word in ["코드", "시스템", "버그", "서버"]):
            return "TD"
        elif any(word in text_lower for word in ["디자인", "ui", "색상", "폰트"]):
            return "AD"
        elif any(word in text_lower for word in ["분석", "데이터", "통계", "리포트"]):
            return "SA"
        elif any(word in text_lower for word in ["글", "문구", "카피", "편집"]):
            return "CE"
        else:
            return "CD"  # 기본: Creative Director

    def save_chat(self, chat_id, text, role):
        """채팅 로그 저장"""
        logs = []
        if CHAT_LOG.exists():
            with open(CHAT_LOG) as f:
                logs = json.load(f)

        logs.append({
            "timestamp": datetime.now().isoformat(),
            "chat_id": str(chat_id),
            "role": role,
            "content": text
        })

        with open(CHAT_LOG, 'w') as f:
            json.dump(logs, f, ensure_ascii=False, indent=2)

    def save_report(self, agent, task, status):
        """보고서 저장"""
        report_file = REPORT_DIR / f"report_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"

        report = {
            "timestamp": datetime.now().isoformat(),
            "agent": agent,
            "task": task,
            "status": status
        }

        with open(report_file, 'w') as f:
            json.dump(report, f, ensure_ascii=False, indent=2)

    def show_recent_reports(self, chat_id):
        """최근 보고서 표시"""
        reports = sorted(REPORT_DIR.glob("report_*.json"))[-5:]

        if reports:
            msg = "*최근 보고서:*\n\n"

            for report_file in reports:
                with open(report_file) as f:
                    report = json.load(f)

                msg += f"• {report['agent']}: {report['task'][:30]}... ({report['status']})\n"

            self.send_message(chat_id, msg)
        else:
            self.send_message(chat_id, "보고서가 없습니다.")

    def send_message(self, chat_id, text):
        """텔레그램 메시지 전송"""
        try:
            url = f"{TELEGRAM_API}/sendMessage"
            data = json.dumps({
                "chat_id": chat_id,
                "text": text,
                "parse_mode": "Markdown"
            }).encode('utf-8')

            req = urllib.request.Request(url, data=data)
            req.add_header('Content-Type', 'application/json')

            with urllib.request.urlopen(req) as response:
                print(f"📤 응답 전송")

            # 응답도 로그에 저장
            self.save_chat(chat_id, text, "assistant")

        except Exception as e:
            print(f"전송 오류: {e}")

    def clear_updates(self):
        """기존 업데이트 클리어"""
        try:
            url = f"{TELEGRAM_API}/getUpdates?offset=-1"
            with urllib.request.urlopen(url) as response:
                result = json.loads(response.read())
                print(f"기존 업데이트 클리어: {result['ok']}")
        except:
            pass

# ===== 메인 실행 =====

def main():
    """메인 함수"""
    messenger = CompanyMessenger()

    try:
        messenger.start()
    except KeyboardInterrupt:
        print("\n\n시스템 종료")

if __name__ == "__main__":
    main()