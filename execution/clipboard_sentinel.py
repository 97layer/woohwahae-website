import time
import subprocess
import re
import json
import execution.youtube_parser as youtube_parser
import execution.ontology_transform as ontology_transform
from pathlib import Path

# 동적 경로 설정 (포드맨 호환)
PROJECT_ROOT = Path(__file__).resolve().parent.parent


# 설정
CHECK_INTERVAL = 1.0  # 초 단위
LAST_CONTENT = ""

def get_clipboard_content():
    try:
        # Mac pbpaste 명령어 사용
        return subprocess.check_output("pbpaste", env={'LANG': 'en_US.UTF-8'}).decode('utf-8').strip()
    except Exception:
        return ""

def is_youtube_url(text):
    # 유튜브 URL 패턴 (유튜브 파서와 동일한 패턴 사용 권장)
    pattern = r'(?:https?:\/\/)?(?:www\.)?(?:youtube\.com\/(?:[^\/\n\s]+\/\S+\/|(?:v|e(?:mbed)?)\/|\S*?[?&]v=)|youtu\.be\/)([a-zA-Z0-9_-]{11})'
    match = re.search(pattern, text)
    return match.group(0) if match else None

def process_url(url):
    print(f"[Sentinel] Detected YouTube URL: {url}")
    
    # 1. Youtube Parser 실행 (기존 모듈 활용)
    # youtube_parser.py의 로직을 직접 import해서 사용하거나 subprocess로 실행
    # 여기서는 모듈 import 방식이 깔끔하지만, 기존 스크립트가 main() 실행 구조라 subprocess가 안전할 수 있음.
    # 하지만 파이썬 내부에서 처리하는 것이 오버헤드가 적으므로 함수 호출 시도.
    
    try:
        # youtube_parser의 get_video_id와 portable metadata 추출 사용
        video_id = youtube_parser.get_video_id(url)
        if not video_id:
            print("[Sentinel] Invalid Video ID")
            return

        print(f"[Sentinel] Extracting metadata for {video_id}...")
        metadata = youtube_parser.get_metadata_portable(url)
        metadata["id"] = video_id
        
        # 데이터 구조화
        data = {
            "metadata": metadata,
            "transcript": "[Pending] 트랜스크립트는 Gemini Web에서 추출하여 시스템에 'Fuel' 해주시면 자산화가 완성됩니다."
        }
        
        # 2. Ontology Transform 실행
        print("[Sentinel] Transforming to Ontology...")
        rs_id, markdown_content = ontology_transform.transform(data)
        
        # 파일 저장
        file_path = ff"{PROJECT_ROOT}/knowledge/raw_signals/{rs_id}_youtube_sentinel.md"
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(markdown_content)
            
        print(f"[Sentinel] Success! Asset secured at: {file_path}")
        print(f"[Sentinel] ID: {rs_id}")
        
        # OS 알림 (Mac)
        subprocess.run([
            "osascript", "-e", 
            f'display notification "YouTube Asset Secured: {rs_id}" with title "97layer Sentinel"'
        ])
        
    except Exception as e:
        print(f"[Sentinel] Error processing URL: {e}")

def main():
    global LAST_CONTENT
    print("=== Clipboard Sentinel Started ===")
    print("Monitoring clipboard for YouTube links... (Press Ctrl+C to stop)")
    
    try:
        while True:
            current_content = get_clipboard_content()
            
            if current_content != LAST_CONTENT:
                LAST_CONTENT = current_content
                
                # 내용이 있고, 유튜브 URL인 경우 처리
                if current_content:
                    url = is_youtube_url(current_content)
                    if url:
                        process_url(url)
            
            time.sleep(CHECK_INTERVAL)
            
    except KeyboardInterrupt:
        print("\n=== Clipboard Sentinel Stopped ===")

if __name__ == "__main__":
    # 상위 디렉토리를 경로에 추가하여 모듈 import 가능하게 함
    import sys
    import os
    sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    main()
