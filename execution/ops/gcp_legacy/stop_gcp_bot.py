#!/usr/bin/env python3
"""
GCP 봇 원격 중지 - HTTP 요청으로
"""

import urllib.request

# GCP 관리 서버로 재시작 요청 (재시작하면 봇이 죽음)
try:
    url = "http://35.184.30.182:8888/restart"
    req = urllib.request.Request(url, method='POST')

    try:
        with urllib.request.urlopen(req, timeout=5) as response:
            print("✅ GCP 봇 재시작 요청 전송")
    except:
        print("GCP 서버 응답 없음 (이미 중지되었을 수 있음)")

except Exception as e:
    print(f"연결 실패: {e}")

# 텔레그램 API로 직접 업데이트 클리어
import json
TOKEN = "8271602365:AAGQwvDfmLv11_CShkeTMSQvnAkDYbDiTxA"
url = f"https://api.telegram.org/bot{TOKEN}/getUpdates?offset=-1"

with urllib.request.urlopen(url) as response:
    result = json.loads(response.read())
    print(f"텔레그램 큐 상태: {result['ok']}")

print("\n이제 Mac에서만 봇 실행 가능합니다!")