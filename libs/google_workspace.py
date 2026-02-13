# Filename: libs/google_workspace.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import os
import logging
from typing import Optional, List, Dict, Any
from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

logger = logging.getLogger(__name__)

SCOPES = [
    'https://www.googleapis.com/auth/drive',
    'https://www.googleapis.com/auth/documents.readonly'
]

class GoogleWorkspace:
    """Manages Google Drive and Docs API interactions."""
    
    def __init__(self, token_path: str = 'token.json'):
        self.token_path = token_path
        self.creds = self._load_credentials()

    def _load_credentials(self) -> Optional[Credentials]:
        """Loads or refreshes OAuth 2.0 credentials."""
        creds = None
        if os.path.exists(self.token_path):
            creds = Credentials.from_authorized_user_file(self.token_path, SCOPES)
        
        if not creds or not creds.valid:
            if creds and creds.expired and creds.refresh_token:
                creds.refresh(Request())
            else:
                # Try credentials.json if CLIENT_CONFIG is missing
                try:
                    if os.path.exists('credentials.json'):
                        flow = InstalledAppFlow.from_client_secrets_file('credentials.json', SCOPES)
                        creds = flow.run_local_server(port=0)
                    else:
                        from libs.google_client_config import CLIENT_CONFIG
                        flow = InstalledAppFlow.from_client_config(CLIENT_CONFIG, SCOPES)
                        creds = flow.run_local_server(port=0)
                    
                    if creds:
                        with open(self.token_path, 'w') as token:
                            token.write(creds.to_json())
                except Exception as e:
                    logger.error(f"Failed to authenticate: {e}")
                    return None
        return creds

    def get_service(self, service_name: str, version: str):
        """Returns a Google API service instance."""
        if not self.creds: return None
        return build(service_name, version, credentials=self.creds)

    def search_files(self, query: str) -> List[Dict[str, Any]]:
        """Search files in Google Drive."""
        service = self.get_service('drive', 'v3')
        if not service: return []
        
        results = service.files().list(
            q=query, 
            fields="nextPageToken, files(id, name, mimeType, md5Checksum)"
        ).execute()
        return results.get('files', [])

    def upload_file(self, local_path: str, parent_id: Optional[str] = None) -> Optional[str]:
        """Uploads a file to Google Drive."""
        service = self.get_service('drive', 'v3')
        if not service: return None
        
        file_metadata = {'name': os.path.basename(local_path)}
        if parent_id:
            file_metadata['parents'] = [parent_id]
            
        media = MediaFileUpload(local_path, resumable=True)
        file = service.files().create(body=file_metadata, media_body=media, fields='id').execute()
        return file.get('id')

    def create_folder(self, title: str, parent_id: Optional[str] = None) -> Optional[str]:
        """Creates a new folder in Google Drive."""
        service = self.get_service('drive', 'v3')
        if not service: return None
        
        file_metadata = {
            'name': title,
            'mimeType': 'application/vnd.google-apps.folder'
        }
        if parent_id:
            file_metadata['parents'] = [parent_id]
        
        file = service.files().create(body=file_metadata, fields='id').execute()
        return file.get('id')

    def read_doc_text(self, doc_id: str) -> str:
        """Reads text content from a Google Doc."""
        service = self.get_service('docs', 'v1')
        if not service: return ""
        
        doc = service.documents().get(documentId=doc_id).execute()
        content = doc.get('body').get('content')
        
        text = ""
        for element in content:
            if 'paragraph' in element:
                elements = element.get('paragraph').get('elements')
                for el in elements:
                    if 'textRun' in el:
                        text += el.get('textRun').get('content')
        return text
