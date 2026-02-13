# Filename: execution/ops/sync_drive.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import os
import hashlib
import logging
from pathlib import Path
from typing import Dict, List, Optional, Set

logger = logging.getLogger(__name__)

class DriveSync:
    """Intelligent synchronization between local and Google Drive."""

    def __init__(self, workspace_root: Optional[str] = None):
        if workspace_root is None:
            workspace_root = Path(__file__).resolve().parent.parent.parent
        self.workspace_root = Path(workspace_root)
        self._gw = None
        self.exclude_patterns = self._load_ignore_patterns()

    @property
    def gw(self):
        if self._gw is None:
            import sys
            if str(self.workspace_root) not in sys.path:
                sys.path.append(str(self.workspace_root))
            from libs.google_workspace import GoogleWorkspace
            self._gw = GoogleWorkspace()
        return self._gw

    def _load_ignore_patterns(self) -> Set[str]:
        """
        Load exclusion patterns from .driveignore file.

        Reads patterns from .driveignore in the workspace root. Lines starting
        with '#' are treated as comments and empty lines are ignored.

        Returns:
            Set of exclusion patterns to use for filtering files

        Notes:
            Falls back to hardcoded exclusions if .driveignore doesn't exist
        """
        driveignore_path = self.workspace_root / '.driveignore'
        patterns = set()

        try:
            with open(driveignore_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    # Skip comments and empty lines
                    if line and not line.startswith('#'):
                        patterns.add(line)
            logger.info("Loaded %d patterns from .driveignore", len(patterns))
        except FileNotFoundError:
            # Fallback to hardcoded exclusions
            logger.warning(".driveignore not found at %s, using hardcoded exclusions", driveignore_path)
            patterns = {
                'venv/',
                '.venv/',
                'env/',
                '__pycache__/',
                '*.pyc',
                '*.pyo',
                '.git/',
                '.tmp/',
                '.DS_Store',
                '.env',
                'token.json',
                'credentials.json',
            }

        return patterns

    def _should_exclude(self, file_path: str) -> bool:
        """
        Check if a file path should be excluded from sync.

        Matches the relative file path against loaded exclusion patterns.
        Supports:
        - Exact directory names (e.g., "venv/")
        - File extensions (e.g., "*.pyc")
        - Path prefixes (e.g., ".tmp")

        Args:
            file_path: Relative file path from workspace root

        Returns:
            True if the file should be excluded, False otherwise
        """
        file_path_normalized = file_path.replace('\\', '/')

        for pattern in self.exclude_patterns:
            # Handle directory patterns (ending with /)
            if pattern.endswith('/'):
                dir_pattern = pattern.rstrip('/')
                # Check if path starts with or contains this directory
                path_parts = file_path_normalized.split('/')
                if dir_pattern in path_parts:
                    return True
                if file_path_normalized.startswith(dir_pattern + '/'):
                    return True

            # Handle wildcard patterns (e.g., *.pyc)
            elif pattern.startswith('*.'):
                extension = pattern[1:]  # includes the dot
                if file_path_normalized.endswith(extension):
                    return True

            # Handle exact matches (e.g., .DS_Store)
            elif '/' not in pattern and '*' not in pattern:
                # Check if pattern matches filename or any path component
                if pattern in file_path_normalized.split('/'):
                    return True

            # Handle path prefix patterns (e.g., .tmp)
            else:
                if file_path_normalized.startswith(pattern):
                    return True
                # Also check if pattern appears as a path component
                if f"/{pattern}/" in f"/{file_path_normalized}/":
                    return True

        return False

    def compute_hash(self, filepath: Path) -> str:
        """
        Compute SHA-256 hash of a file.

        Args:
            filepath: Path to file to hash

        Returns:
            Hexadecimal hash digest string
        """
        hasher = hashlib.sha256()
        with open(filepath, 'rb') as f:
            while chunk := f.read(8192):
                hasher.update(chunk)
        return hasher.hexdigest()

    def check_diff(self, local_dir: str, folder_id: str) -> Dict[str, List[str]]:
        """
        Compare local directory with Google Drive folder.

        Scans local directory and computes hashes for all non-excluded files.

        Args:
            local_dir: Path to local directory to scan
            folder_id: Google Drive folder ID to compare against

        Returns:
            Dictionary with local file count and list of file paths
        """
        local_path = Path(local_dir)

        local_files = {}
        for f in local_path.rglob('*'):
            if f.is_file():
                rel_path = str(f.relative_to(local_path))
                if not self._should_exclude(rel_path):
                    local_files[rel_path] = self.compute_hash(f)

        logger.info("Found %d local files after exclusions", len(local_files))
        return {"local_count": len(local_files), "files": list(local_files.keys())}

    def push_to_drive(self, folder_id: str, files: Optional[List[str]] = None):
        """
        Upload files to Google Drive.

        Uploads all non-excluded files from workspace to specified Drive folder.

        Args:
            folder_id: Google Drive folder ID to upload to
            files: Optional list of specific files to upload (if None, uploads all)
        """
        from googleapiclient.http import MediaFileUpload
        service = self.gw.get_service('drive', 'v3')

        # Upload all non-excluded files in workspace
        for f in self.workspace_root.rglob('*'):
            if f.is_file():
                rel_path = str(f.relative_to(self.workspace_root))
                if not self._should_exclude(rel_path):
                    logger.info("Uploading %s...", rel_path)
                    self.gw.upload_file(str(f), parent_id=folder_id)
                else:
                    logger.debug("Skipping excluded file: %s", rel_path)
