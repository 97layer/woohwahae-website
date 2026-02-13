import sys
from pathlib import Path

# Add project root to sys.path
project_root = Path(__file__).resolve().parent.parent.parent
if str(project_root) not in sys.path:
    sys.path.append(str(project_root))

from libs.google_workspace import GoogleWorkspace

def migrate_folder_name():
    gw = GoogleWorkspace()
    service = gw.get_service('drive', 'v3')
    
    # Potential old names to check
    old_names = ["97layer", "97layerOS_New", "97layerOS_New_Sync"]
    new_name = "97layerOS"
    
    folder_entry = None
    
    print(f"Searching for folders: {old_names}...")
    
    for old_name in old_names:
        results = service.files().list(
            q=f"name = '{old_name}' and mimeType = 'application/vnd.google-apps.folder' and trashed = false",
            fields="files(id, name)"
        ).execute()
        
        folders = results.get('files', [])
        if folders:
            folder_entry = folders[0]
            print(f"Found existing folder: '{old_name}' (ID: {folder_entry['id']})")
            break
            
    if not folder_entry:
        # Check if new folder already exists
        results_new = service.files().list(
            q=f"name = '{new_name}' and mimeType = 'application/vnd.google-apps.folder' and trashed = false",
            fields="files(id, name)"
        ).execute()
        
        if results_new.get('files'):
            print(f"Folder '{new_name}' already exists. No action needed.")
        else:
            print(f"None of the candidate folders found. Please run sync_to_drive.py to create '{new_name}'.")
        return

    # Rename the folder
    folder_id = folder_entry['id']
    old_name_found = folder_entry['name']
    
    if old_name_found == new_name:
         print(f"Folder is already named '{new_name}'.")
         return

    print(f"Renaming '{old_name_found}' to '{new_name}'...")
    
    try:
        file = service.files().update(
            fileId=folder_id,
            body={'name': new_name},
            fields='id, name'
        ).execute()
        print(f"Success! Folder renamed to: {file.get('name')}")
    except Exception as e:
        print(f"Error renaming folder: {e}")

if __name__ == "__main__":
    migrate_folder_name()
