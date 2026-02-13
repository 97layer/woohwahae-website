#!/usr/bin/env python3
"""
97layerOS - Google OAuth Token Refresher
Automatically refreshes expired Google API tokens
"""

import json
import os
from datetime import datetime
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from google.auth.transport.requests import Request

SCOPES = [
    'https://www.googleapis.com/auth/drive',
    'https://www.googleapis.com/auth/documents.readonly'
]

CREDENTIALS_FILE = 'credentials.json'
TOKEN_FILE = 'token.json'

def log(message):
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] {message}")

def refresh_token():
    """Refresh or generate new OAuth token"""
    creds = None

    # Load existing token
    if os.path.exists(TOKEN_FILE):
        log(f"Loading existing token from {TOKEN_FILE}...")
        try:
            creds = Credentials.from_authorized_user_file(TOKEN_FILE, SCOPES)
            log("✓ Token loaded")
        except Exception as e:
            log(f"✗ Failed to load token: {e}")

    # Refresh or create new token
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            log("Token expired. Attempting refresh...")
            try:
                creds.refresh(Request())
                log("✓ Token refreshed successfully")
            except Exception as e:
                log(f"✗ Refresh failed: {e}")
                log("Creating new authorization flow...")
                creds = None

        if not creds:
            if not os.path.exists(CREDENTIALS_FILE):
                log(f"✗ Error: {CREDENTIALS_FILE} not found")
                return False

            log("Starting OAuth authorization flow...")
            log("Please authorize in your browser...")

            flow = InstalledAppFlow.from_client_secrets_file(
                CREDENTIALS_FILE, SCOPES
            )
            creds = flow.run_local_server(port=0)
            log("✓ Authorization successful")

    # Save token
    with open(TOKEN_FILE, 'w') as token:
        token.write(creds.to_json())

    log(f"✓ Token saved to {TOKEN_FILE}")

    # Display token info
    if creds.expiry:
        log(f"Token expires: {creds.expiry.isoformat()}")

    return True

def main():
    log("="*60)
    log("97layerOS - OAuth Token Refresh")
    log("="*60)

    if refresh_token():
        log("\n✓ OAuth token refresh complete")
        return 0
    else:
        log("\n✗ OAuth token refresh failed")
        return 1

if __name__ == "__main__":
    import sys
    sys.exit(main())
