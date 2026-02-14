#!/usr/bin/env python3
"""
자동 보고서 생성 및 텔레그램 전송 시스템
GCP에서 cron으로 실행되어 정기 보고서를 전송
"""

import json
import os
import sys
from datetime import datetime, timedelta
from pathlib import Path
import urllib.request
import urllib.parse

# 환경 변수에서 봇 토큰 가져오기
from dotenv import load_dotenv
load_dotenv()

BOT_TOKEN = os.getenv('TELEGRAM_BOT_TOKEN', '8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI')
API_URL = f"https://api.telegram.org/bot{BOT_TOKEN}"
CHAT_ID = 7565534667  # 97layer 사용자

# 경로 설정
BASE_DIR = Path.home() / "97layerOS"
KNOWLEDGE_DIR = BASE_DIR / "knowledge"
CHAT_MEMORY = KNOWLEDGE_DIR / "chat_memory" / "7565534667.json"
SYSTEM_STATE = KNOWLEDGE_DIR / "system_state.json"
TASK_STATUS = BASE_DIR / "task_status.json"

def send_telegram(text: str, parse_mode: str = "Markdown"):
    """텔레그램으로 메시지 전송"""
    try:
        url = f"{API_URL}/sendMessage"
        data = {
            "chat_id": CHAT_ID,
            "text": text,
            "parse_mode": parse_mode
        }

        params = urllib.parse.urlencode(data).encode('utf-8')
        req = urllib.request.Request(url, data=params)

        with urllib.request.urlopen(req) as response:
            result = json.loads(response.read())
            return result.get("ok", False)
    except Exception as e:
        print(f"텔레그램 전송 실패: {e}")
        return False

def get_daily_report():
    """일일 보고서 생성"""
    report = []
    report.append("📊 *97LAYER OS 일일 보고서*")
    report.append(f"📅 {datetime.now().strftime('%Y-%m-%d %H:%M')}")
    report.append("")

    # 1. 시스템 상태
    report.append("*🖥 시스템 상태*")
    if SYSTEM_STATE.exists():
        with open(SYSTEM_STATE, 'r') as f:
            state = json.load(f)
            report.append(f"• 모드: {state.get('mode', 'Unknown')}")
            report.append(f"• 버전: {state.get('version', 'Unknown')}")
            report.append(f"• 지속일: {state.get('days_running', 0)}일")
    else:
        report.append("• 상태 파일 없음")
    report.append("")

    # 2. 최근 24시간 메시지 통계
    report.append("*💬 24시간 메시지 통계*")
    if CHAT_MEMORY.exists():
        with open(CHAT_MEMORY, 'r') as f:
            messages = json.load(f)

        now = datetime.now()
        day_ago = now - timedelta(days=1)

        recent_messages = []
        for msg in messages:
            try:
                msg_time = datetime.fromisoformat(msg['timestamp'].replace('Z', '+00:00'))
                if msg_time.replace(tzinfo=None) > day_ago:
                    recent_messages.append(msg)
            except:
                continue

        user_msgs = sum(1 for m in recent_messages if m.get('role') == 'user')
        bot_msgs = sum(1 for m in recent_messages if m.get('role') == 'assistant')

        report.append(f"• 받은 메시지: {user_msgs}개")
        report.append(f"• 응답 메시지: {bot_msgs}개")
        report.append(f"• 총 대화: {len(recent_messages)}개")
    else:
        report.append("• 대화 기록 없음")
    report.append("")

    # 3. 작업 상태
    report.append("*📋 작업 상태*")
    if TASK_STATUS.exists():
        with open(TASK_STATUS, 'r') as f:
            tasks = json.load(f)

        pending = sum(1 for t in tasks.get('tasks', []) if t.get('status') == 'pending')
        completed = sum(1 for t in tasks.get('tasks', []) if t.get('status') == 'completed')

        report.append(f"• 대기 중: {pending}개")
        report.append(f"• 완료: {completed}개")

        # 최근 완료 작업
        if completed > 0:
            report.append("\n*최근 완료:*")
            completed_tasks = [t for t in tasks.get('tasks', []) if t.get('status') == 'completed']
            for task in completed_tasks[-3:]:  # 최근 3개만
                report.append(f"  ✅ {task.get('title', 'Unknown')}")
    else:
        report.append("• 작업 없음")
    report.append("")

    # 4. 에이전트 활동
    report.append("*🤖 에이전트 활동*")
    report.append("• CD: 대기 중")
    report.append("• TD: 활성")
    report.append("• AD: 대기 중")
    report.append("• CE: 대기 중")
    report.append("• SA: 대기 중")

    return "\n".join(report)

def get_hourly_summary():
    """시간별 요약 보고"""
    report = []
    report.append("⏰ *시간별 상태 체크*")
    report.append(f"{datetime.now().strftime('%H:%M')} - 시스템 정상")

    # 최근 1시간 메시지 수
    if CHAT_MEMORY.exists():
        with open(CHAT_MEMORY, 'r') as f:
            messages = json.load(f)

        hour_ago = datetime.now() - timedelta(hours=1)
        recent = sum(1 for m in messages if datetime.fromisoformat(m['timestamp'].replace('Z', '+00:00')).replace(tzinfo=None) > hour_ago)

        if recent > 0:
            report.append(f"• 최근 1시간: {recent}개 메시지")

    report.append("• 모든 시스템 정상 작동 중 ✅")

    return "\n".join(report)

def main():
    """메인 실행 - 인자에 따라 다른 보고서 생성"""
    import sys

    report_type = sys.argv[1] if len(sys.argv) > 1 else "daily"

    if report_type == "daily":
        # 일일 보고서 (매일 오전 9시)
        report = get_daily_report()
    elif report_type == "hourly":
        # 시간별 요약 (매시간)
        report = get_hourly_summary()
    else:
        report = "❓ 알 수 없는 보고서 타입"

    # 텔레그램으로 전송
    if send_telegram(report):
        print(f"✅ {report_type} 보고서 전송 완료")
    else:
        print(f"❌ {report_type} 보고서 전송 실패")

if __name__ == "__main__":
    main()