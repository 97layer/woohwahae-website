#!/usr/bin/env python3
"""
GCP 실시간 메모리 동기화 스크립트
30초마다 또는 변경 감지 시 Mac으로 전송

개선사항:
- 5분 → 30초 주기로 단축
- 변경 감지 시에만 전송 (해시 비교)
- 실패 시 재시도 로직
- 압축 전송 지원
"""

import json
import sys
import time
import hashlib
import gzip
from pathlib import Path
from datetime import datetime
from typing import Optional, Dict, Any

try:
    import requests
except ImportError:
    print("❌ requests 모듈 없음. pip install requests 실행 필요")
    sys.exit(1)

# 설정
MAC_SERVER = "http://192.168.0.8:9876"  # Mac 로컬 IP
CHAT_MEMORY_FILE = Path.home() / "97layerOS" / "knowledge" / "chat_memory" / "7565534667.json"
SYNC_INTERVAL = 30  # 30초 주기
MAX_RETRIES = 3
RETRY_DELAY = 5

class RealtimeMemorySync:
    """실시간 메모리 동기화"""

    def __init__(self):
        self.last_hash: Optional[str] = None
        self.last_sync: Optional[datetime] = None
        self.error_count = 0
        self.success_count = 0

    def calculate_hash(self, data: Dict) -> str:
        """데이터 해시 계산"""
        json_str = json.dumps(data, sort_keys=True)
        return hashlib.sha256(json_str.encode()).hexdigest()

    def has_changed(self, data: Dict) -> bool:
        """데이터 변경 여부 확인"""
        current_hash = self.calculate_hash(data)

        if self.last_hash is None:
            self.last_hash = current_hash
            return True

        changed = current_hash != self.last_hash
        if changed:
            self.last_hash = current_hash

        return changed

    def compress_data(self, data: Dict) -> bytes:
        """데이터 압축"""
        json_str = json.dumps(data, ensure_ascii=False)
        return gzip.compress(json_str.encode('utf-8'))

    def push_to_mac(self, force: bool = False) -> bool:
        """
        Mac으로 메모리 전송

        Args:
            force: True면 변경 여부와 관계없이 전송

        Returns:
            성공 여부
        """
        try:
            # chat_memory 읽기
            if not CHAT_MEMORY_FILE.exists():
                print(f"[{datetime.now()}] ⚠️ chat_memory 파일 없음")
                return False

            with open(CHAT_MEMORY_FILE, 'r', encoding='utf-8') as f:
                memory_data = json.load(f)

            # 변경 확인
            if not force and not self.has_changed(memory_data):
                return True  # 변경 없음, 성공으로 간주

            # 메타데이터 추가
            sync_payload = {
                "timestamp": datetime.now().isoformat(),
                "source": "gcp_realtime",
                "hash": self.last_hash,
                "message_count": len(memory_data),
                "data": memory_data
            }

            print(f"[{datetime.now()}] 📤 변경 감지, Mac 서버로 전송 중...")

            # 재시도 로직
            for attempt in range(MAX_RETRIES):
                try:
                    # 큰 데이터는 압축 전송
                    if len(memory_data) > 100:
                        compressed = self.compress_data(sync_payload)
                        response = requests.post(
                            f"{MAC_SERVER}/sync_memory_compressed",
                            data=compressed,
                            headers={'Content-Encoding': 'gzip'},
                            timeout=10
                        )
                    else:
                        response = requests.post(
                            f"{MAC_SERVER}/sync_memory",
                            json=sync_payload,
                            timeout=10
                        )

                    if response.status_code == 200:
                        result = response.json()
                        self.last_sync = datetime.now()
                        self.success_count += 1
                        self.error_count = 0

                        print(f"[{datetime.now()}] ✅ 전송 성공 (#{self.success_count}): {result}")
                        return True

                    else:
                        print(f"[{datetime.now()}] ⚠️ 서버 응답: {response.status_code}")

                except requests.exceptions.ConnectionError:
                    print(f"[{datetime.now()}] ⚠️ 연결 실패 (시도 {attempt + 1}/{MAX_RETRIES})")
                    if attempt < MAX_RETRIES - 1:
                        time.sleep(RETRY_DELAY)

                except requests.exceptions.Timeout:
                    print(f"[{datetime.now()}] ⚠️ 타임아웃 (시도 {attempt + 1}/{MAX_RETRIES})")
                    if attempt < MAX_RETRIES - 1:
                        time.sleep(RETRY_DELAY)

            self.error_count += 1
            return False

        except Exception as e:
            print(f"[{datetime.now()}] ❌ 오류: {e}")
            import traceback
            traceback.print_exc()
            self.error_count += 1
            return False

    def get_status(self) -> Dict[str, Any]:
        """동기화 상태 반환"""
        return {
            "last_sync": self.last_sync.isoformat() if self.last_sync else None,
            "success_count": self.success_count,
            "error_count": self.error_count,
            "last_hash": self.last_hash[:8] if self.last_hash else None
        }

    def run_daemon(self):
        """데몬 모드로 실행"""
        print(f"🚀 실시간 메모리 동기화 시작")
        print(f"   - 동기화 주기: {SYNC_INTERVAL}초")
        print(f"   - 대상 서버: {MAC_SERVER}")
        print(f"   - 메모리 파일: {CHAT_MEMORY_FILE}")

        # 초기 동기화
        self.push_to_mac(force=True)

        while True:
            try:
                time.sleep(SYNC_INTERVAL)

                # 동기화 실행
                success = self.push_to_mac()

                # 상태 출력 (10회마다)
                if self.success_count % 10 == 0 and self.success_count > 0:
                    status = self.get_status()
                    print(f"[{datetime.now()}] 📊 상태: {status}")

                # 에러가 많으면 경고
                if self.error_count > 5:
                    print(f"[{datetime.now()}] ⚠️ 연속 에러 {self.error_count}회 발생")
                    # 잠시 대기 후 재시작
                    time.sleep(60)
                    self.error_count = 0

            except KeyboardInterrupt:
                print(f"\n[{datetime.now()}] 👋 동기화 종료")
                status = self.get_status()
                print(f"최종 상태: {status}")
                break

            except Exception as e:
                print(f"[{datetime.now()}] ❌ 데몬 오류: {e}")
                time.sleep(60)


def main():
    """메인 함수"""
    sync = RealtimeMemorySync()

    # 인자 처리
    if len(sys.argv) > 1:
        if sys.argv[1] == "--once":
            # 1회만 실행
            success = sync.push_to_mac(force=True)
            sys.exit(0 if success else 1)
        elif sys.argv[1] == "--status":
            # 상태 확인
            print(json.dumps(sync.get_status(), indent=2))
            sys.exit(0)

    # 데몬 모드
    sync.run_daemon()


if __name__ == "__main__":
    main()