import os
import shutil
import zipfile
from datetime import datetime
from pathlib import Path

# 동적 경로 설정 (포드맨 호환)
PROJECT_ROOT = Path(__file__).resolve().parent.parent


# 설정
SOURCE_DIR = str(PROJECT_ROOT)
PRIMARY_DEST_DIR = "/Users/97layer/내 드라이브/97layerOS_Snapshots"
FALLBACK_DEST_DIR = f"{PROJECT_ROOT}_Backups"
TMP_DEST_DIR = "/tmp/97layerOS_Snapshots"
TEMP_WORK_DIR = "/tmp/97layer_snapshot_work"
SHADOW_DIR = "/tmp/97layer_shadow_copy"

# 초강력 필터링: 지능 중심 백업을 위해 노이즈 폴더 제거
EXCLUDE_DIRS = {
    "node_modules", ".local_node", ".mcp-source", "dist", "build",
    ".git", "__pycache__", ".venv", "venv", "venv_clean", "venv_new", "venv_old",
    ".antigravity", ".gemini", ".vscode", ".idea", ".DS_Store", ".tmp"
}

# Shadow Copy 대상 (OS/클라우드 잠금 회피용)
SHADOW_FILES = ["token.json", "credentials.json"]

def create_snapshot():
    """정제된 자산과 Shadow Copy를 활용한 무중단 스냅샷 생성"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    zip_filename = f"97layerOS_Intelligence_{timestamp}.zip"
    
    # 작업 디렉토리 초기화
    for d in [TEMP_WORK_DIR, SHADOW_DIR]:
        if not os.path.exists(d):
            os.makedirs(d, exist_ok=True)
            
    temp_zip_path = os.path.join(TEMP_WORK_DIR, zip_filename)
    
    print(f"[{datetime.now()}] Shadow Copy 및 초정밀 스냅샷 생성 시작...")
    
    # 1. Shadow Copy 실행 (권한 에러 우회)
    for file in SHADOW_FILES:
        src = os.path.join(SOURCE_DIR, file)
        dst = os.path.join(SHADOW_DIR, file)
        try:
            if os.path.exists(src):
                shutil.copy2(src, dst)
                print(f"[INFO] Shadow Copy 생성 완료: {file}")
        except Exception as e:
            print(f"[WARN] Shadow Copy 실패: {file} - {e}")

    try:
        file_count = 0
        with zipfile.ZipFile(temp_zip_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
            for root, dirs, files in os.walk(SOURCE_DIR):
                # 제외 디렉토리 필터링
                dirs[:] = [d for d in dirs if d not in EXCLUDE_DIRS and not d.endswith("venv")]
                
                for file in files:
                    if file in EXCLUDE_DIRS or file.endswith(".pyc") or any(v in root for v in ["venv", "node_modules"]):
                        continue
                        
                    src_path = os.path.join(root, file)
                    rel_path = os.path.relpath(src_path, SOURCE_DIR)
                    
                    # 보안 파일은 Shadow Copy본을 사용
                    if file in SHADOW_FILES:
                        src_path = os.path.join(SHADOW_DIR, file)
                    
                    # .env 보안 처리 (제외)
                    if file == ".env":
                        continue
                    
                    try:
                        zipf.write(src_path, rel_path)
                        file_count += 1
                    except OSError as e:
                        print(f"[WARN] 접근 스킵: {rel_path} - {e}")

        print(f"[{datetime.now()}] 압축 완료 ({file_count} files). 용량: {os.path.getsize(temp_zip_path) / 1024 / 1024:.2f} MB")
        
        # 전송 시도
        for dest_dir in [PRIMARY_DEST_DIR, FALLBACK_DEST_DIR, TMP_DEST_DIR]:
            try:
                if not os.path.exists(dest_dir):
                    os.makedirs(dest_dir, exist_ok=True)
                
                final_path = os.path.join(dest_dir, zip_filename)
                shutil.copy2(temp_zip_path, final_path)
                print(f"[{datetime.now()}] 백업 전송 성공: {final_path}")
                
                if os.path.exists(temp_zip_path):
                    os.remove(temp_zip_path)
                return True
            except Exception as e:
                print(f"[WARN] {dest_dir} 전송 실패, 다음 시도... ({e})")
                
        return False
        
    except Exception as e:
        print(f"[ERROR] 스냅샷 아카이빙 실패: {e}")
        if os.path.exists(temp_zip_path):
            os.remove(temp_zip_path)
        return False

if __name__ == "__main__":
    create_snapshot()
