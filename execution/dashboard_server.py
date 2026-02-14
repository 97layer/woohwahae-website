
import os
import json
import logging
import time
import threading
from http.server import HTTPServer, SimpleHTTPRequestHandler
from pathlib import Path
from datetime import datetime

# Path Configuration
PROJECT_ROOT = Path(__file__).resolve().parent.parent
STATUS_FILE = PROJECT_ROOT / "knowledge" / "system_state.json"
SYNAPSE_FILE = PROJECT_ROOT / "knowledge" / "agent_hub" / "synapse_bridge.json"
COUNCIL_DIR = PROJECT_ROOT / "knowledge" / "council_log"
DASHBOARD_DIR = PROJECT_ROOT / "dashboard"

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(message)s')

class DashboardHandler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        # Serve from Dashboard Directory
        super().__init__(*args, directory=str(DASHBOARD_DIR), **kwargs)

    def do_GET(self):
        # Special Routes
        if self.path == '/api/status':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            try:
                # 1. Read System State
                status_data = {"status": "INITIALIZING", "agents": {}}
                if STATUS_FILE.exists():
                    with open(STATUS_FILE, 'r', encoding='utf-8') as f:
                        status_data = json.load(f)

                # 2. Merge Synapse Bridge Data (Agent Collaboration + Parallel Tasks)
                bridge_file = PROJECT_ROOT / "knowledge" / "agent_hub" / "synapse_bridge.json"
                if bridge_file.exists():
                    with open(bridge_file, 'r', encoding='utf-8') as f:
                        bridge_data = json.load(f)
                        # Merge agents from bridge
                        if "active_agents" in bridge_data:
                            for agent, info in bridge_data["active_agents"].items():
                                status_data["agents"][agent] = info

                        # Add parallel processing metrics
                        status_data["parallel_mode"] = bridge_data.get("collaboration_mode") == "Parallel"
                        status_data["performance"] = bridge_data.get("performance", {})
                        status_data["stats"] = bridge_data.get("stats", {})

                # Check System Heartbeat
                last_update_str = status_data.get("last_update")
                if last_update_str:
                    try:
                        last_ts = datetime.fromisoformat(last_update_str)
                        delta = (datetime.now() - last_ts).total_seconds()
                        status_data["is_alive"] = delta < 300
                    except:
                        status_data["is_alive"] = False

                self.wfile.write(json.dumps(status_data).encode('utf-8'))
            except Exception as e:
                self.wfile.write(json.dumps({"error": str(e)}).encode('utf-8'))
            return

        elif self.path == '/api/chat':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()

            try:
                # Case 1: Latest Council Log (Markdown)
                files = sorted(COUNCIL_DIR.glob("*.md"), key=os.path.getmtime, reverse=True)
                chat_content = []
                chat_title = "No Active Council"

                if files:
                    latest = files[0]
                    content = latest.read_text(encoding='utf-8')
                    # Parse simplified chat structure
                    lines = content.split('\n')
                    messages = []
                    current_speaker = None
                    current_text = ""

                    for line in lines:
                        if line.startswith("## 🗣️"):
                            if current_speaker:
                                messages.append({"role": current_speaker, "text": current_text.strip()})
                            current_speaker = line.replace("## 🗣️", "").strip()
                            current_text = ""
                        elif line.startswith("## 👑"):
                            if current_speaker:
                                messages.append({"role": current_speaker, "text": current_text.strip()})
                            current_speaker = "Creative_Director"
                            current_text = ""
                        else:
                            current_text += line + "\n"

                    if current_speaker:
                        messages.append({"role": current_speaker, "text": current_text.strip()})
                    chat_content = messages
                    chat_title = latest.name

                # Case 2: JSON Chat Memory Fallback
                MEMORY_FILE = PROJECT_ROOT / "knowledge" / "chat_memory" / "7565534667.json"
                if not chat_content and MEMORY_FILE.exists():
                    with open(MEMORY_FILE, 'r', encoding='utf-8') as f:
                        raw_messages = json.load(f)
                        # Take last 20 messages
                        chat_content = [
                            {"role": msg.get("role", "Unknown"), "text": msg.get("content", "")}
                            for msg in raw_messages[-20:]
                        ]
                        chat_title = "Chat History (JSON)"

                self.wfile.write(json.dumps({
                    "title": chat_title,
                    "messages": chat_content
                }).encode('utf-8'))
            except Exception as e:
                self.wfile.write(json.dumps({"error": str(e)}).encode('utf-8'))
            return

        elif self.path == '/api/stream':
            # SSE - Real-time Agent Hub Updates
            self.send_response(200)
            self.send_header('Content-type', 'text/event-stream')
            self.send_header('Cache-Control', 'no-cache')
            self.send_header('Connection', 'keep-alive')
            self.end_headers()

            try:
                last_update = None
                while True:
                    # Read Synapse Bridge for real-time agent state
                    if SYNAPSE_FILE.exists():
                        with open(SYNAPSE_FILE, 'r', encoding='utf-8') as f:
                            synapse_data = json.load(f)

                        current_update = synapse_data.get('last_update')
                        if current_update != last_update:
                            # Send SSE event
                            event_data = json.dumps(synapse_data)
                            self.wfile.write(f"data: {event_data}\n\n".encode('utf-8'))
                            self.wfile.flush()
                            last_update = current_update

                    time.sleep(1)  # Poll every second

            except (BrokenPipeError, ConnectionResetError):
                # Client disconnected
                pass
            except Exception as e:
                logging.error(f"SSE stream error: {e}")
            return

        # Default: Serve Static Files
        return super().do_GET()

def run(server_class=HTTPServer, handler_class=DashboardHandler, port=8000):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f"Starting Dashboard Server on port {port}...")
    print(f"Open http://localhost:{port}/ in your browser.")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass
    httpd.server_close()
    print("Stopping Dashboard Server.")

if __name__ == '__main__':
    run()
