import os
import shutil
import hashlib
from pathlib import Path

def get_hash(file_path):
    hash_md5 = hashlib.md5()
    with open(file_path, "rb") as f:
        for chunk in iter(lambda: f.read(4096), b""):
            hash_md5.update(chunk)
    return hash_md5.hexdigest()

def cleanup_duplicates(root_dir):
    root = Path(root_dir)
    redundant_files = list(root.rglob("* (1)*"))
    
    print(f"--- Starting cleanup in {root_dir} ---")
    
    for dup_path in redundant_files:
        # (1)을 제거한 원본 파일 이름 추정
        original_name = dup_path.name.replace(" (1)", "")
        original_path = dup_path.parent / original_name
        
        if original_path.exists():
            dup_hash = get_hash(dup_path)
            orig_hash = get_hash(original_path)
            
            if dup_hash == orig_hash:
                print(f"[MATCH] Content identical. Deleting redundant: {dup_path}")
                dup_path.unlink()
            else:
                # 해시가 다를 경우 더 최신 파일을 유지하는 로직 (안전을 위해 수동 확인 권장이나 여기서는 일단 보존)
                print(f"[DIFF] Content varies. Keeping both for safety: {dup_path}")
        else:
            # 원본이 없고 (1)만 있는 경우 이름을 원본으로 변경
            print(f"[RENAME] Original missing. Renaming {dup_path} to {original_path}")
            dup_path.rename(original_path)

def optimize_root(root_dir):
    root = Path(root_dir)
    archive_dir = root / "archive" / "tests"
    archive_dir.mkdir(parents=True, exist_ok=True)
    
    # VISION_CLOUDSYNC_TEST_*.md 이동
    for test_file in root.glob("VISION_CLOUDSYNC_TEST_*.md"):
        print(f"[MOVE] Archiving test file: {test_file.name}")
        shutil.move(str(test_file), str(archive_dir / test_file.name))
        
    # .env.txt 처리
    env_txt = root / ".env.txt"
    if env_txt.exists():
        print(f"[INFO] Found .env.txt. Please verify if it's needed before deletion.")

if __name__ == "__main__":
    PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
    cleanup_duplicates(PROJECT_ROOT)
    optimize_root(PROJECT_ROOT)
    print("--- Cleanup Process Finished ---")
