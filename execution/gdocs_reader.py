# Filename: execution/gdocs_reader.py
# Author: 97LAYER Mercenary
# Date: 2026-02-12 (Recovered)

import logging
from typing import List, Dict, Any, Optional
from libs.google_workspace import GoogleWorkspace

logger = logging.getLogger(__name__)

def read_structural_elements(elements: List[Dict[str, Any]]) -> str:
    """Recursively extracts text from Google Docs structural elements."""
    text = ""
    for value in elements:
        if 'paragraph' in value:
            for el in value.get('paragraph').get('elements'):
                if 'textRun' in el:
                    text += el.get('textRun').get('content')
        elif 'table' in value:
            for row in value.get('table').get('tableRows'):
                for cell in row.get('tableCells'):
                    text += read_structural_elements(cell.get('content'))
        elif 'tableOfContents' in value:
            text += read_structural_elements(value.get('tableOfContents').get('content'))
    return text

class GDocsReader:
    def __init__(self, gw: Optional[GoogleWorkspace] = None):
        self.gw = gw or GoogleWorkspace()

    def get_doc_text(self, document_id: str) -> str:
        """Fetch and parse text from a Google Doc."""
        try:
            service = self.gw.get_service('docs', 'v1')
            doc = service.documents().get(documentId=document_id).execute()
            return read_structural_elements(doc.get('body').get('content'))
        except Exception as e:
            logger.error(f"Error reading doc {document_id}: {e}")
            return ""
