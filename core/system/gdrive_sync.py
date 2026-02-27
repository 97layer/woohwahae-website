#!/usr/bin/env python3
"""
LAYER OS Google Drive Sync Utility
Purpose: Sync intelligence files (state.md, daily reports) to Google Drive
Philosophy: Cloud-backed session continuity for model-agnostic knowledge preservation

Author: LAYER OS Technical Director
Created: 2026-02-16
"""

import os
import sys
import json
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Optional

try:
    from google.oauth2 import service_account
    from googleapiclient.discovery import build
    from googleapiclient.http import MediaFileUpload, MediaIoBaseDownload
except ImportError:
    print("⚠️  Google API libraries not installed.")
    print("   Run: pip install google-auth google-auth-oauthlib google-auth-httplib2 google-api-python-client")
    sys.exit(1)

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent


class GDriveSync:
    """
    Google Drive Sync for LAYER OS Knowledge Base

    Features:
    - Upload state.md (session continuity)
    - Upload daily reports (morning/evening/weekly)
    - Search existing files
    - Create folder structure automatically
    """

    def __init__(self):
        """Initialize Google Drive API client"""
        self.credentials_path = PROJECT_ROOT / 'credentials' / 'gdrive_auth.json'
        self.folder_id = os.getenv('GOOGLE_DRIVE_FOLDER_ID')

        if not self.credentials_path.exists():
            raise FileNotFoundError(
                f"Google Drive credentials not found: {self.credentials_path}\n"
                f"Please download service account JSON from Google Cloud Console."
            )

        if not self.folder_id:
            raise ValueError(
                "GOOGLE_DRIVE_FOLDER_ID not set in .env\n"
                "Please add: GOOGLE_DRIVE_FOLDER_ID=your_folder_id"
            )

        # Authenticate
        credentials = service_account.Credentials.from_service_account_file(
            str(self.credentials_path),
            scopes=['https://www.googleapis.com/auth/drive']
        )

        self.service = build('drive', 'v3', credentials=credentials)
        print(f"✅ Google Drive API initialized (Folder: {self.folder_id})")

    def search_files(self, query: str, folder_id: Optional[str] = None) -> List[Dict]:
        """
        Search files in Google Drive

        Args:
            query: Search query (e.g., "name contains 'INTELLIGENCE'")
            folder_id: Optional folder ID to search within (defaults to main folder)

        Returns:
            List of file metadata dicts
        """
        search_folder = folder_id or self.folder_id

        # Build query
        full_query = f"'{search_folder}' in parents and {query} and trashed=false"

        try:
            results = self.service.files().list(
                q=full_query,
                fields="files(id, name, mimeType, modifiedTime, size)",
                orderBy="modifiedTime desc"
            ).execute()

            files = results.get('files', [])
            print(f"🔍 Found {len(files)} files matching: {query}")
            return files

        except Exception as e:
            print(f"❌ Search failed: {e}")
            return []

    def upload_file(self, local_path: Path, drive_folder: str = "intelligence") -> Optional[str]:
        """
        Upload file to Google Drive

        Args:
            local_path: Local file path
            drive_folder: Subfolder name (intelligence, daily_reports, content)

        Returns:
            File ID if successful, None otherwise
        """
        if not local_path.exists():
            print(f"❌ File not found: {local_path}")
            return None

        # Get or create subfolder
        subfolder_id = self._get_or_create_folder(drive_folder, self.folder_id)

        # Check if file already exists
        existing = self.search_files(
            f"name = '{local_path.name}'",
            folder_id=subfolder_id
        )

        # Prepare file metadata
        file_metadata = {
            'name': local_path.name,
            'parents': [subfolder_id]
        }

        # Prepare media upload
        media = MediaFileUpload(
            str(local_path),
            mimetype=self._get_mimetype(local_path),
            resumable=True
        )

        try:
            if existing:
                # Update existing file
                file_id = existing[0]['id']
                updated = self.service.files().update(
                    fileId=file_id,
                    media_body=media
                ).execute()
                print(f"✅ Updated: {local_path.name} (ID: {file_id})")
                return file_id
            else:
                # Create new file
                created = self.service.files().create(
                    body=file_metadata,
                    media_body=media,
                    fields='id'
                ).execute()
                file_id = created.get('id')
                print(f"✅ Uploaded: {local_path.name} (ID: {file_id})")
                return file_id

        except Exception as e:
            print(f"❌ Upload failed: {e}")
            return None

    def download_file(self, file_id: str, local_path: Path) -> bool:
        """
        Download file from Google Drive

        Args:
            file_id: Google Drive file ID
            local_path: Local destination path

        Returns:
            True if successful
        """
        try:
            request = self.service.files().get_media(fileId=file_id)

            local_path.parent.mkdir(parents=True, exist_ok=True)

            with open(local_path, 'wb') as f:
                downloader = MediaIoBaseDownload(f, request)
                done = False
                while not done:
                    status, done = downloader.next_chunk()
                    if status:
                        print(f"📥 Download progress: {int(status.progress() * 100)}%")

            print(f"✅ Downloaded: {local_path}")
            return True

        except Exception as e:
            print(f"❌ Download failed: {e}")
            return False

    def sync_intelligence_quanta(self) -> bool:
        """
        Sync state.md to Google Drive

        Returns:
            True if successful
        """
        quanta_path = PROJECT_ROOT / 'knowledge' / 'agent_hub' / 'state.md'

        if not quanta_path.exists():
            print(f"⚠️  state.md not found: {quanta_path}")
            return False

        print("📤 Syncing state.md to Google Drive...")
        file_id = self.upload_file(quanta_path, drive_folder="intelligence")
        return file_id is not None

    def sync_daily_reports(self) -> Dict[str, bool]:
        """
        Sync all daily reports to Google Drive

        Returns:
            Dict of {filename: success_status}
        """
        reports_dir = PROJECT_ROOT / 'knowledge' / 'reports' / 'daily'

        if not reports_dir.exists():
            print(f"⚠️  Daily reports directory not found: {reports_dir}")
            return {}

        results = {}

        # Get all JSON reports
        report_files = sorted(reports_dir.glob('*.json'), key=lambda p: p.stat().st_mtime, reverse=True)

        if not report_files:
            print("ℹ️  No daily reports to sync.")
            return {}

        print(f"📤 Syncing {len(report_files)} daily reports...")

        for report_file in report_files:
            file_id = self.upload_file(report_file, drive_folder="daily_reports")
            results[report_file.name] = (file_id is not None)

        success_count = sum(results.values())
        print(f"✅ Synced {success_count}/{len(report_files)} reports")

        return results

    def sync_execution_context(self) -> bool:
        """
        Sync execution_context.json to Google Drive (intelligence/ 폴더).

        Mac↔GCP 공유 상태 파일 — heartbeat.py가 갱신, GCP가 감시.
        Drive 동기화로 양쪽이 항상 최신 상태를 볼 수 있음.
        """
        ctx_path = PROJECT_ROOT / 'knowledge' / 'system' / 'execution_context.json'

        if not ctx_path.exists():
            print("⚠️  execution_context.json not found")
            return False

        print("📤 Syncing execution_context.json to Google Drive...")
        file_id = self.upload_file(ctx_path, drive_folder="intelligence")
        return file_id is not None

    def _get_or_create_folder(self, folder_name: str, parent_id: str) -> str:
        """
        Get existing folder ID or create new folder

        Args:
            folder_name: Folder name
            parent_id: Parent folder ID

        Returns:
            Folder ID
        """
        # Search for existing folder
        query = f"name = '{folder_name}' and mimeType = 'application/vnd.google-apps.folder'"
        existing = self.search_files(query, folder_id=parent_id)

        if existing:
            return existing[0]['id']

        # Create new folder
        folder_metadata = {
            'name': folder_name,
            'mimeType': 'application/vnd.google-apps.folder',
            'parents': [parent_id]
        }

        try:
            folder = self.service.files().create(
                body=folder_metadata,
                fields='id'
            ).execute()
            folder_id = folder.get('id')
            print(f"📁 Created folder: {folder_name} (ID: {folder_id})")
            return folder_id
        except Exception as e:
            print(f"❌ Failed to create folder: {e}")
            return parent_id  # Fallback to parent

    def _get_mimetype(self, path: Path) -> str:
        """Get MIME type based on file extension"""
        mime_types = {
            '.md': 'text/markdown',
            '.json': 'application/json',
            '.txt': 'text/plain',
            '.pdf': 'application/pdf',
            '.jpg': 'image/jpeg',
            '.png': 'image/png'
        }
        return mime_types.get(path.suffix.lower(), 'application/octet-stream')


def main():
    """CLI interface for manual sync"""
    import argparse

    parser = argparse.ArgumentParser(description='LAYER OS Google Drive Sync')
    parser.add_argument('--intelligence', action='store_true', help='Sync state.md')
    parser.add_argument('--reports', action='store_true', help='Sync daily reports')
    parser.add_argument('--all', action='store_true', help='Sync everything')
    parser.add_argument('--search', type=str, help='Search files by name')

    args = parser.parse_args()

    try:
        sync = GDriveSync()

        if args.search:
            results = sync.search_files(f"name contains '{args.search}'")
            print(f"\n📋 Search results for '{args.search}':")
            for file in results:
                print(f"   • {file['name']} (ID: {file['id']}, Modified: {file['modifiedTime']})")

        elif args.intelligence or args.all:
            sync.sync_intelligence_quanta()
            sync.sync_execution_context()  # execution_context.json도 함께 동기화

        elif args.reports or args.all:
            sync.sync_daily_reports()

        else:
            parser.print_help()

    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
