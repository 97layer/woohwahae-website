# Filename: execution/gdrive_import.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import os
import logging
from typing import Optional
from libs.google_workspace import GoogleWorkspace

logger = logging.getLogger(__name__)

class GDriveImporter:
    def __init__(self, gw: Optional[GoogleWorkspace] = None):
        self.gw = gw or GoogleWorkspace()

    def import_to_inbox(self, local_path: str) -> Optional[str]:
        """Uploads a local file to the Google Drive Inbox folder."""
        # Search for Inbox folder
        folders = self.gw.search_files("name = 'Inbox' and mimeType = 'application/vnd.google-apps.folder'")
        inbox_id = folders[0]['id'] if folders else None
        
        if not inbox_id:
            inbox_id = self.gw.create_folder("Inbox")
            
        return self.gw.upload_file(local_path, parent_id=inbox_id)
