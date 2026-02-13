import os
import re
import time
from datetime import datetime
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

# 설정
SOURCE_DIR = "/Users/97layer/97layerOS"
INDEX_HTML = os.path.join(SOURCE_DIR, "archive/index.html")

def extract_metadata_from_md(file_path):
    """마크다운 파일에서 제목과 요약을 추출"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # 제목 추출 (# 제목)
        title_match = re.search(r'^#\s+(.+)$', content, re.MULTILINE)
        title = title_match.group(1).split(':')[0].strip() if title_match else os.path.basename(file_path)
        
        # 요약 추출 (전략적 통찰 섹션 또는 첫 문단)
        summary_match = re.search(r'##\s+1\.\s+전략적\s+정돈.+?\-\s+정서적\s+안식과\s+감각적\s+환대', content, re.DOTALL)
        if not summary_match:
            # 대안: 첫 번째 문단 추출
            paragraphs = [p.strip() for p in content.split('\n\n') if p.strip() and not p.startswith('#')]
            summary = paragraphs[0][:150] + "..." if paragraphs else "No description available."
        else:
            summary = "2026년 미니멀리즘은 단순한 '삭제'가 아닌 '정서적 안식'과 '감각적 환대'로 나아갑니다. 우리는 인위적인 개입을 배제하고, 사용자가 본질적인 평온을 마주할 수 있도록 배경으로 물러나는 기술을 지향합니다."
            
        return title, summary
    except Exception as e:
        print(f"[ERROR] Failed to parse MD: {e}")
        return os.path.basename(file_path), "Failed to extract summary."

def inject_to_html(title, summary, date_str, category="Archive"):
    """HTML 파일의 <main> 섹션 최상단에 새로운 항목 삽입"""
    try:
        with open(INDEX_HTML, 'r', encoding='utf-8') as f:
            html = f.read()
        
        # 중복 체크
        if f'"{title}"' in html or f'>{title}<' in html:
            print(f"[SKIP] Item already exists in archive: {title}")
            return False

        # 삽입할 HTML 조각
        new_item = f"""
            <article class="archive-item">
                <div class="item-meta">{date_str} / {category}</div>
                <h2 class="item-title">{title}</h2>
                <p class="item-excerpt">
                    {summary}
                </p>
            </article>"""
        
        marker = '<main class="archive-list" id="archive-list">'
        if marker in html:
            updated_html = html.replace(marker, marker + new_item)
            
            with open(INDEX_HTML, 'w', encoding='utf-8') as f:
                f.write(updated_html)
            print(f"[SUCCESS] Injected new entry to archive: {title}")
            return True
        else:
            print(f"[ERROR] Could not find marker in HTML: {marker}")
            return False
            
    except Exception as e:
        print(f"[ERROR] Failed to inject HTML: {e}")
        return False

class BlueprintHandler(FileSystemEventHandler):
    """Final_Omni_Blueprint.md 파일 생성을 감시"""
    def on_created(self, event):
        if not event.is_directory and event.src_path.endswith("Final_Omni_Blueprint.md"):
            self.process_file(event.src_path)
            
    def on_modified(self, event):
        if not event.is_directory and event.src_path.endswith("Final_Omni_Blueprint.md"):
            self.process_file(event.src_path)

    def process_file(self, file_path):
        print(f"[{datetime.now()}] Detected Blueprint: {file_path}")
        time.sleep(1)
        title, summary = extract_metadata_from_md(file_path)
        date_str = datetime.now().strftime("%Y. %m. %d")
        inject_to_html(title, summary, date_str)

def run_archive_daemon():
    print(f"[{datetime.now()}] Archive Daemon 가동 시작 (Monitoring Blueprints)...")
    
    # 초기 스캔
    for root, dirs, files in os.walk(SOURCE_DIR):
        for file in files:
            if file.endswith("Final_Omni_Blueprint.md"):
                src_path = os.path.join(root, file)
                title, summary = extract_metadata_from_md(src_path)
                date_str = datetime.fromtimestamp(os.path.getmtime(src_path)).strftime("%Y. %m. %d")
                inject_to_html(title, summary, date_str)

    event_handler = BlueprintHandler()
    observer = Observer()
    observer.schedule(event_handler, SOURCE_DIR, recursive=True)
    observer.start()
    
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()

if __name__ == "__main__":
    run_archive_daemon()
