
import os
import json
import logging
from http.server import HTTPServer, SimpleHTTPRequestHandler
from pathlib import Path
from datetime import datetime

# Path Configuration
PROJECT_ROOT = Path(__file__).resolve().parent.parent
STATUS_FILE = PROJECT_ROOT / "knowledge" / "system_state.json"
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
                if STATUS_FILE.exists():
                    with open(STATUS_FILE, 'r', encoding='utf-8') as f:
                        status_data = json.load(f)
                else:
                    status_data = {"status": "INITIALIZING", "agents": {}}
                
                # Check System Heartbeat (Guardian Logic)
                # If last update > 5 min ago, mark as stale
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
                # Get latest council log
                files = sorted(COUNCIL_DIR.glob("*.md"), key=os.path.getmtime, reverse=True)
                chat_content = ""
                chat_title = "No Active Council"
                
                if files:
                    latest = files[0]
                    content = latest.read_text(encoding='utf-8')
                    # Parse simplified chat structure (Naive parsing)
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
                            current_speaker = "Creative_Director (Decision)"
                            current_text = ""
                        else:
                            current_text += line + "\n"
                    
                    if current_speaker:
                        messages.append({"role": current_speaker, "text": current_text.strip()})
                        
                    chat_content = messages
                    chat_title = latest.name
                
                self.wfile.write(json.dumps({
                    "title": chat_title,
                    "messages": chat_content
                }).encode('utf-8'))
            except Exception as e:
                self.wfile.write(json.dumps({"error": str(e)}).encode('utf-8'))
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
