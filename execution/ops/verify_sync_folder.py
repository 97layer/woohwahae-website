import sys
from pathlib import Path

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.google_workspace import GoogleWorkspace

def verify_folder_name():
    gw = GoogleWorkspace()
    service = gw.get_service('drive', 'v3')
    
    target_name = "97layerOS"
    
    print(f"Verifying existence of folder: {target_name}...")
    
    results = service.files().list(
        q=f"name = '{target_name}' and mimeType = 'application/vnd.google-apps.folder' and trashed = false",
        fields="files(id, name)"
    ).execute()
    
    folders = results.get('files', [])
    
    if folders:
        folder = folders[0]
        print(f"VERIFICATION SUCCESS: Found folder '{folder['name']}' with ID: {folder['id']}")
    else:
        print(f"VERIFICATION FAILED: Folder '{target_name}' not found.")

if __name__ == "__main__":
    verify_folder_name()
