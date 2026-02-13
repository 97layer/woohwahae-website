# Filename: libs/code_executor.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import subprocess
import logging
import os
from pathlib import Path
from typing import Tuple

logger = logging.getLogger(__name__)

class CodeExecutor:
    """Executes code in the designated virtual environment."""
    
    def __init__(self, workspace: str, venv_python: str = "/tmp/venv_97layer/bin/python3"):
        self.workspace = Path(workspace)
        self.venv_python = venv_python

    def execute_python(self, code: str) -> Tuple[bool, str]:
        """Executes Python code and returns (success, output)."""
        tmp_file = self.workspace / ".tmp_exec.py"
        with open(tmp_file, "w") as f:
            f.write(code)
            
        try:
            result = subprocess.run(
                [self.venv_python, str(tmp_file)],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=str(self.workspace)
            )
            
            output = result.stdout if result.returncode == 0 else result.stderr
            return (result.returncode == 0, output)
        except Exception as e:
            return (False, str(e))
        finally:
            if tmp_file.exists():
                os.remove(tmp_file)
