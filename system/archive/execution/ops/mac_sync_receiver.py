#!/usr/bin/env python3
"""
Mac에서 실행되는 동기화 수신 서버
GCP가 HTTP POST로 chat_memory를 전송하면 자동으로 적용
"""
import json
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from datetime import datetime

PORT = 9876
CHAT_MEMORY_FILE = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "7565534667.json"

class SyncHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == "/sync_memory":
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)

            try:
                # JSON 파싱
                new_memory = json.loads(post_data.decode('utf-8'))

                # 백업
                backup = CHAT_MEMORY_FILE.with_suffix('.json.backup')
                if CHAT_MEMORY_FILE.exists():
                    CHAT_MEMORY_FILE.rename(backup)

                # 저장
                with open(CHAT_MEMORY_FILE, 'w', encoding='utf-8') as f:
                    json.dump(new_memory, f, ensure_ascii=False, indent=4)

                print(f"[{datetime.now()}] ✅ GCP에서 chat_memory 수신: {len(new_memory)}개 메시지")

                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(json.dumps({"status": "success", "count": len(new_memory)}).encode())

            except Exception as e:
                print(f"[{datetime.now()}] ❌ 오류: {e}")
                self.send_response(500)
                self.end_headers()
                self.wfile.write(json.dumps({"status": "error", "message": str(e)}).encode())
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, format, *args):
        # 로그 출력
        print(f"[{datetime.now()}] {format % args}")

def main():
    server = HTTPServer(('0.0.0.0', PORT), SyncHandler)
    print(f"🚀 Mac 동기화 수신 서버 시작: http://0.0.0.0:{PORT}")
    print(f"   GCP가 이 서버로 chat_memory를 전송하면 자동으로 적용됩니다.")
    server.serve_forever()

if __name__ == "__main__":
    main()
