#!/usr/bin/env python3
"""
GCP Management HTTP Server
Port 8888에서 실행되며 다음 기능 제공:
- GET /memory : chat_memory 조회
- POST /restart : telegram_daemon 재시작
- GET /status : 시스템 상태 확인
"""
from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import subprocess
import os
from pathlib import Path
from datetime import datetime

class ManagementHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/memory':
            # 기존 chat_memory 제공
            memory_file = Path.home() / '97layerOS' / 'knowledge' / 'chat_memory' / '7565534667.json'
            if memory_file.exists():
                with open(memory_file, encoding='utf-8') as f:
                    data = f.read()
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(data.encode())
            else:
                self.send_response(404)
                self.end_headers()

        elif self.path == '/status':
            # 시스템 상태 확인
            try:
                # telegram_daemon 프로세스 확인
                result = subprocess.run(
                    ['ps', 'aux'],
                    capture_output=True,
                    text=True
                )
                telegram_running = 'telegram_daemon.py' in result.stdout

                status = {
                    "timestamp": datetime.now().isoformat(),
                    "telegram_daemon": "running" if telegram_running else "stopped",
                    "hostname": os.uname().nodename
                }

                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(status).encode())
            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())

        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        if self.path == '/restart':
            # telegram_daemon 재시작 (구버전)
            try:
                os.chdir(Path.home() / '97layerOS')

                # 1. 기존 프로세스 종료
                subprocess.run(['pkill', '-f', 'telegram_daemon.py'], check=False)

                # 2. 재시작
                subprocess.Popen(
                    ['nohup', 'python3', 'execution/telegram_daemon.py'],
                    stdout=open('/tmp/telegram_daemon.log', 'w'),
                    stderr=subprocess.STDOUT
                )

                # 3. 잠시 대기
                import time
                time.sleep(2)

                # 4. 확인
                result = subprocess.run(
                    ['ps', 'aux'],
                    capture_output=True,
                    text=True
                )
                running = 'telegram_daemon.py' in result.stdout

                response = {
                    "status": "success" if running else "failed",
                    "message": "Telegram daemon restarted" if running else "Failed to start daemon",
                    "timestamp": datetime.now().isoformat()
                }

                self.send_response(200 if running else 500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())

            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())

        elif self.path == '/restart_async':
            # async_telegram_daemon 재시작 (멀티모달 버전)
            try:
                os.chdir(Path.home() / '97layerOS')

                # 1. 기존 프로세스 종료
                subprocess.run(['pkill', '-f', 'async_telegram_daemon.py'], check=False)

                # 2. 재시작
                subprocess.Popen(
                    ['nohup', 'python3', 'execution/async_telegram_daemon.py'],
                    stdout=open('/tmp/async_telegram_daemon.log', 'w'),
                    stderr=subprocess.STDOUT
                )

                # 3. 잠시 대기
                import time
                time.sleep(3)  # 비동기 시스템이라 조금 더 대기

                # 4. 확인
                result = subprocess.run(
                    ['ps', 'aux'],
                    capture_output=True,
                    text=True
                )
                running = 'async_telegram_daemon.py' in result.stdout

                response = {
                    "status": "success" if running else "failed",
                    "message": "Async Telegram Multimodal Bot restarted" if running else "Failed to start",
                    "timestamp": datetime.now().isoformat(),
                    "multimodal": True
                }

                self.send_response(200 if running else 500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())

            except Exception as e:
                self.send_response(500)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())

        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        # 로그 출력 (필요시)
        pass

def main():
    server = HTTPServer(('0.0.0.0', 8888), ManagementHandler)
    print(f"[{datetime.now()}] 🚀 GCP Management Server started on port 8888")
    print("  - GET  /memory        : Chat memory")
    print("  - POST /restart       : Restart telegram daemon (legacy)")
    print("  - POST /restart_async : Restart async multimodal bot")
    print("  - GET  /status        : System status")
    server.serve_forever()

if __name__ == '__main__':
    main()
