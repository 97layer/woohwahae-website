#!/usr/bin/env python3
"""Clear Telegram webhook to use polling mode"""

import requests
import os

BOT_TOKEN = os.getenv("TELEGRAM_BOT_TOKEN")

# Delete webhook
url = f"https://api.telegram.org/bot{BOT_TOKEN}/deleteWebhook"
response = requests.post(url, json={"drop_pending_updates": True})

print("Deleting webhook...")
result = response.json()
if result.get("ok"):
    print("✓ Webhook deleted successfully")
    print(f"  Result: {result['result']}")
else:
    print("✗ Failed to delete webhook:", result)

# Get webhook info
url = f"https://api.telegram.org/bot{BOT_TOKEN}/getWebhookInfo"
response = requests.get(url)
info = response.json()

print("\nCurrent webhook status:")
if info.get("ok"):
    webhook = info["result"]
    print(f"  URL: {webhook.get('url', 'None')}")
    print(f"  Has custom certificate: {webhook.get('has_custom_certificate', False)}")
    print(f"  Pending update count: {webhook.get('pending_update_count', 0)}")
else:
    print("  Error getting webhook info")