#!/usr/bin/env python3
"""
Mac 실시간 동기화 수신 서버 (개선 버전)
GCP에서 실시간으로 전송되는 메모리를 수신하고 에이전트들에게 알림

개선사항:
- 압축 데이터 지원
- 에이전트 알림 시스템 통합
- 변경 사항 감지 및 diff 생성
- 통계 및 모니터링
"""

import json
import gzip
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from datetime import datetime
from typing import Dict, Any, Optional
import threading

# 프로젝트 루트 설정
PROJECT_ROOT = Path.home() / "97layerOS"
sys.path.append(str(PROJECT_ROOT))

# AgentNotifier 임포트
try:
    from libs.agent_notifier import get_notifier
    notifier = get_notifier(str(PROJECT_ROOT))
except ImportError:
    print("⚠️ AgentNotifier를 로드할 수 없음. 알림 기능 비활성화")
    notifier = None

PORT = 9876
CHAT_MEMORY_FILE = PROJECT_ROOT / "knowledge" / "chat_memory" / "7565534667.json"

class RealtimeSyncStats:
    """동기화 통계 관리"""

    def __init__(self):
        self.total_syncs = 0
        self.successful_syncs = 0
        self.failed_syncs = 0
        self.total_messages = 0
        self.last_sync_time: Optional[datetime] = None
        self.start_time = datetime.now()

    def record_sync(self, success: bool, message_count: int = 0):
        """동기화 기록"""
        self.total_syncs += 1
        if success:
            self.successful_syncs += 1
            self.total_messages += message_count
            self.last_sync_time = datetime.now()
        else:
            self.failed_syncs += 1

    def get_stats(self) -> Dict[str, Any]:
        """통계 반환"""
        uptime = (datetime.now() - self.start_time).total_seconds()
        return {
            "uptime_seconds": uptime,
            "total_syncs": self.total_syncs,
            "successful_syncs": self.successful_syncs,
            "failed_syncs": self.failed_syncs,
            "success_rate": (self.successful_syncs / max(1, self.total_syncs)) * 100,
            "total_messages": self.total_messages,
            "last_sync": self.last_sync_time.isoformat() if self.last_sync_time else None
        }


# 전역 통계 인스턴스
stats = RealtimeSyncStats()


class EnhancedSyncHandler(BaseHTTPRequestHandler):
    """개선된 동기화 핸들러"""

    def do_GET(self):
        """GET 요청 처리"""
        if self.path == "/status":
            # 서버 상태 반환
            status_data = {
                "server": "running",
                "port": PORT,
                "stats": stats.get_stats(),
                "notifier": "enabled" if notifier else "disabled"
            }

            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(status_data, indent=2).encode())

        elif self.path == "/memory":
            # 현재 메모리 조회
            if CHAT_MEMORY_FILE.exists():
                with open(CHAT_MEMORY_FILE, 'r', encoding='utf-8') as f:
                    data = f.read()
                self.send_response(200)
                self.send_header('Content-type', 'application/json')
                self.end_headers()
                self.wfile.write(data.encode())
            else:
                self.send_response(404)
                self.end_headers()

        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        """POST 요청 처리"""
        if self.path == "/sync_memory":
            self._handle_sync(compressed=False)

        elif self.path == "/sync_memory_compressed":
            self._handle_sync(compressed=True)

        else:
            self.send_response(404)
            self.end_headers()

    def _handle_sync(self, compressed: bool = False):
        """동기화 처리"""
        try:
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)

            # 압축 해제
            if compressed:
                post_data = gzip.decompress(post_data)

            # JSON 파싱
            sync_data = json.loads(post_data.decode('utf-8'))

            # 페이로드 구조 확인
            if "data" in sync_data:
                new_memory = sync_data["data"]
                metadata = {
                    "timestamp": sync_data.get("timestamp"),
                    "source": sync_data.get("source", "unknown"),
                    "hash": sync_data.get("hash"),
                    "message_count": sync_data.get("message_count", 0)
                }
            else:
                # 레거시 형식 지원
                new_memory = sync_data
                metadata = {
                    "timestamp": datetime.now().isoformat(),
                    "source": "legacy",
                    "message_count": len(new_memory)
                }

            # 기존 데이터와 비교
            old_count = 0
            if CHAT_MEMORY_FILE.exists():
                with open(CHAT_MEMORY_FILE, 'r', encoding='utf-8') as f:
                    old_memory = json.load(f)
                    old_count = len(old_memory)

                # 백업
                backup = CHAT_MEMORY_FILE.with_suffix('.json.backup')
                CHAT_MEMORY_FILE.rename(backup)

            # 새 데이터 저장
            with open(CHAT_MEMORY_FILE, 'w', encoding='utf-8') as f:
                json.dump(new_memory, f, ensure_ascii=False, indent=4)

            new_count = len(new_memory)
            diff_count = new_count - old_count

            # 통계 기록
            stats.record_sync(True, new_count)

            # 에이전트들에게 알림
            if notifier and diff_count > 0:
                self._notify_agents(metadata, diff_count)

            # 응답
            response_data = {
                "status": "success",
                "old_count": old_count,
                "new_count": new_count,
                "diff": diff_count,
                "metadata": metadata
            }

            print(f"[{datetime.now()}] ✅ 동기화 완료: +{diff_count} 메시지 (총 {new_count})")

            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response_data).encode())

        except Exception as e:
            print(f"[{datetime.now()}] ❌ 동기화 오류: {e}")
            stats.record_sync(False)

            self.send_response(500)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({
                "status": "error",
                "message": str(e)
            }).encode())

    def _notify_agents(self, metadata: Dict[str, Any], diff_count: int):
        """에이전트들에게 변경 알림"""
        try:
            # 메모리 동기화 이벤트
            sync_event = {
                "type": "memory_sync",
                "source": metadata.get("source", "unknown"),
                "timestamp": metadata.get("timestamp"),
                "stats": {
                    "total_messages": metadata.get("message_count", 0),
                    "new_messages": diff_count
                }
            }

            # 모든 에이전트에게 브로드캐스트
            notifier.broadcast(sync_event, priority=4)

            # 새 메시지가 많으면 긴급 알림
            if diff_count > 10:
                urgent_event = {
                    "type": "urgent_sync",
                    "message": f"긴급: {diff_count}개의 새 메시지 수신",
                    **sync_event
                }
                notifier.broadcast(urgent_event, priority=1)

            print(f"[{datetime.now()}] 📢 {diff_count}개 변경사항을 에이전트들에게 알림")

        except Exception as e:
            print(f"[{datetime.now()}] ⚠️ 알림 실패: {e}")

    def log_message(self, format, *args):
        """로그 메시지 (간소화)"""
        # 필요한 경우에만 로깅
        if "GET /status" not in format % args:
            print(f"[{datetime.now()}] {format % args}")


class MonitoringThread(threading.Thread):
    """모니터링 스레드"""

    def __init__(self):
        super().__init__(daemon=True)

    def run(self):
        """5분마다 통계 출력"""
        import time
        while True:
            time.sleep(300)  # 5분
            print(f"[{datetime.now()}] 📊 동기화 통계:")
            print(json.dumps(stats.get_stats(), indent=2))


def main():
    """메인 함수"""
    # 서버 시작
    server = HTTPServer(('0.0.0.0', PORT), EnhancedSyncHandler)

    print(f"🚀 Mac 실시간 동기화 수신 서버 시작")
    print(f"   - 포트: {PORT}")
    print(f"   - 메모리 파일: {CHAT_MEMORY_FILE}")
    print(f"   - 에이전트 알림: {'활성화' if notifier else '비활성화'}")
    print(f"   - 상태 조회: http://localhost:{PORT}/status")

    # 모니터링 스레드 시작
    monitor = MonitoringThread()
    monitor.start()

    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print(f"\n[{datetime.now()}] 👋 서버 종료")
        print(f"최종 통계:")
        print(json.dumps(stats.get_stats(), indent=2))


if __name__ == "__main__":
    main()