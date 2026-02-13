# Filename: sync_to_drive.py
import sys
from pathlib import Path
from execution.ops.sync_drive import DriveSync

def main():
    sync = DriveSync()
    print("Connecting to Google Drive...")
    
    # Try to find or create the target folder
    target_folder_name = "97layerOS"
    folders = sync.gw.search_files(f"name = '{target_folder_name}' and mimeType = 'application/vnd.google-apps.folder'")
    
    if folders:
        folder_id = folders[0]['id']
        print(f"Found existing folder: {target_folder_name} ({folder_id})")
    else:
        folder_id = sync.gw.create_folder(target_folder_name)
        print(f"Created new folder: {target_folder_name} ({folder_id})")
    
    print("Starting sync (excluding venv, __pycache__, .git)...")
    sync.push_to_drive(folder_id)
    print("Sync completed successfully.")

if __name__ == "__main__":
    main()
