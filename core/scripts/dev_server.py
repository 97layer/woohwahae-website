import http.server
import socketserver
import os

PORT = 9700
DIRECTORY = "website"

class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DIRECTORY, **kwargs)

if __name__ == "__main__":
    # Change to project root
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    os.chdir(project_root)
    
    if not os.path.exists(DIRECTORY):
        print(f"❌ Error: Directory '{DIRECTORY}' not found.")
        exit(1)

    socketserver.TCPServer.allow_reuse_address = True
    with socketserver.TCPServer(("", PORT), Handler) as httpd:
        print(f"🚀 WOOHWAHAE Dev Server running at: http://localhost:{PORT}")
        print(f"📁 Serving from: {os.path.abspath(DIRECTORY)}")
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n🛑 Server stopped.")
            httpd.server_close()
