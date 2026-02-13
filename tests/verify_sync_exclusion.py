import sys
import unittest
from pathlib import Path

# Add project root to sys.path
PROJECT_ROOT = Path(__file__).resolve().parent.parent
sys.path.append(str(PROJECT_ROOT))

from execution.ops.sync_drive import DriveSync

class TestDriveSyncExclusion(unittest.TestCase):
    def setUp(self):
        self.syncer = DriveSync(str(PROJECT_ROOT))
        print(f"\n[Test] Loaded exclusion patterns: {len(self.syncer.exclude_patterns)}")

    def test_venv_exclusion(self):
        # These should be EXCLUDED
        excluded_paths = [
            ".venv/bin/python",
            ".venv/lib/site-packages/nothing.py",
            ".venv_runtime/bin/python",
            "token.json",
            ".env",
            ".git/config",
            ".DS_Store",
            "__pycache__/something.pyc"
        ]
        
        for path in excluded_paths:
            with self.subTest(path=path):
                is_excluded = self.syncer._should_exclude(path)
                print(f"[Check] '{path}' -> Excluded? {is_excluded}")
                self.assertTrue(is_excluded, f"Failed to exclude: {path}")

    def test_knowledge_inclusion(self):
        # These should be INCLUDED
        included_paths = [
            "knowledge/assets/briefing/report.md",
            "libs/core_config.py",
            "execution/technical_daemon.py",
            "directives/agents/creative_director.md"
        ]
        
        for path in included_paths:
            with self.subTest(path=path):
                is_excluded = self.syncer._should_exclude(path)
                print(f"[Check] '{path}' -> Excluded? {is_excluded}")
                self.assertFalse(is_excluded, f"Falsely excluded: {path}")

if __name__ == '__main__':
    unittest.main()
