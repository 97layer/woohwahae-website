"""Git auto-commit for directive changes"""
import subprocess
from pathlib import Path
from datetime import datetime

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent

def git_commit_directive(file_path, message_type="update"):
    """Auto-commit directive changes"""
    try:
        file_name = Path(file_path).name

        # Add file
        subprocess.run(['git', 'add', file_path], cwd=PROJECT_ROOT, check=True)

        # Commit with standardized message
        commit_msg = f"directive: {message_type} {file_name}\n\n🤖 Auto-committed by system\n{datetime.now()}"
        subprocess.run(['git', 'commit', '-m', commit_msg], cwd=PROJECT_ROOT, check=True)

        return True
    except:
        return False

if __name__ == "__main__":
    import sys
    if len(sys.argv) > 1:
        result = git_commit_directive(sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else "update")
        print("✓ Committed" if result else "✗ Failed")
