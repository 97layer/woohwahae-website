#!/usr/bin/env python3
"""
97layerOS Agent Event Stream Server
Flask SSE로 에이전트 이벤트를 브라우저에 실시간 전송

Author: 97layerOS + Claude Code
Port: 8082 (dashboard_server.py가 8081 사용 중)
"""

import json
import time
import os
from pathlib import Path
from flask import Flask, Response, render_template_string, send_from_directory
from flask_cors import CORS
import threading
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

# Project root
PROJECT_ROOT = Path(__file__).parent.parent.parent
LOG_DIR = PROJECT_ROOT / ".infra" / "logs"

# Flask app
app = Flask(__name__)
CORS(app)  # CORS 허용 (localhost:8081 ← 8082)

# SSE 구독자 관리
subscribers = []


class EventFileHandler(FileSystemEventHandler):
    """
    _events.jsonl 파일 변경 감지 → SSE 브로드캐스트
    """

    def on_modified(self, event):
        if event.is_directory:
            return

        # _events.jsonl 파일만 감시
        if not event.src_path.endswith("_events.jsonl"):
            return

        # 마지막 라인 읽기
        try:
            with open(event.src_path, "r", encoding="utf-8") as f:
                lines = f.readlines()
                if lines:
                    last_line = lines[-1].strip()
                    event_data = json.loads(last_line)
                    broadcast_event(event_data)
        except Exception as e:
            print(f"[Stream Server] 파일 읽기 실패: {e}", flush=True)


def broadcast_event(event_data):
    """
    모든 SSE 구독자에게 이벤트 전송

    Args:
        event_data: dict 형태의 이벤트 데이터
    """
    # 구독자 전체에게 전송
    for sub_queue in subscribers:
        sub_queue.append(event_data)


def start_file_watcher():
    """
    watchdog로 .infra/logs/ 디렉토리 감시
    """
    event_handler = EventFileHandler()
    observer = Observer()
    observer.schedule(event_handler, str(LOG_DIR), recursive=False)
    observer.start()
    print(f"[Stream Server] 파일 감시 시작: {LOG_DIR}", flush=True)
    return observer


# ─── SSE Endpoint ───

@app.route("/events")
def events():
    """
    SSE 스트림 엔드포인트

    브라우저에서 EventSource로 연결:
    const source = new EventSource('http://localhost:8082/events');
    source.onmessage = (e) => { const data = JSON.parse(e.data); ... };
    """

    def event_stream():
        # 이 구독자 전용 큐
        queue = []
        subscribers.append(queue)

        try:
            while True:
                if queue:
                    # 큐에 이벤트가 있으면 전송
                    event_data = queue.pop(0)
                    yield f"data: {json.dumps(event_data, ensure_ascii=False)}\n\n"
                else:
                    # 큐가 비었으면 heartbeat (30초마다)
                    yield ": heartbeat\n\n"
                time.sleep(0.5)
        except GeneratorExit:
            # 연결 종료 시 구독자 제거
            subscribers.remove(queue)

    return Response(event_stream(), mimetype="text/event-stream")


@app.route("/agents/status")
def agents_status():
    """
    전체 에이전트 최근 상태 조회 (REST API)

    Returns:
        JSON: {agent_id: {last_action, last_target, timestamp}}
    """
    status = {}

    # .infra/logs/*_events.jsonl 순회
    for log_file in LOG_DIR.glob("*_events.jsonl"):
        agent_id = log_file.stem.replace("_events", "")

        try:
            with open(log_file, "r", encoding="utf-8") as f:
                lines = f.readlines()
                if lines:
                    last_event = json.loads(lines[-1].strip())
                    status[agent_id] = {
                        "last_action": last_event.get("action"),
                        "last_target": last_event.get("target"),
                        "timestamp": last_event.get("timestamp")
                    }
        except Exception:
            continue

    return json.dumps(status, ensure_ascii=False, indent=2)


@app.route("/")
def index():
    """
    테스트 페이지 (간단한 SSE 수신 확인용)
    """
    html = """
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Agent Event Stream Test</title>
    <style>
        body { font-family: 'IBM Plex Mono', monospace; background: #1a1a1a; color: #aaa; padding: 20px; }
        h1 { color: #fff; }
        #events { background: #222; padding: 15px; border: 1px solid #333; max-height: 600px; overflow-y: auto; }
        .event { margin: 5px 0; font-size: 12px; }
        .agent { color: #5ad; font-weight: bold; }
        .action { color: #5d5; }
        .target { color: #da5; }
    </style>
</head>
<body>
    <h1>Agent Event Stream Test</h1>
    <div id="events"></div>

    <script>
        const eventsDiv = document.getElementById('events');
        const source = new EventSource('/events');

        source.onmessage = (e) => {
            const data = JSON.parse(e.data);
            const eventEl = document.createElement('div');
            eventEl.className = 'event';
            eventEl.innerHTML =
                `<span class="agent">${data.agent}</span> → ` +
                `<span class="action">${data.action}</span> ` +
                `<span class="target">${data.target || ''}</span> ` +
                `<span style="color:#666">(${data.timestamp})</span>`;
            eventsDiv.appendChild(eventEl);
            eventsDiv.scrollTop = eventsDiv.scrollHeight;
        };

        source.onerror = (e) => {
            console.error('SSE error:', e);
        };
    </script>
</body>
</html>
    """
    return render_template_string(html)


def main():
    """
    서버 시작
    """
    # 로그 디렉토리 생성
    LOG_DIR.mkdir(parents=True, exist_ok=True)

    # 파일 감시 시작
    observer = start_file_watcher()

    try:
        print("[Stream Server] Starting on http://localhost:8082", flush=True)
        print("[Stream Server] SSE endpoint: http://localhost:8082/events", flush=True)
        print("[Stream Server] Status API: http://localhost:8082/agents/status", flush=True)
        app.run(host="0.0.0.0", port=8082, debug=False, threaded=True)
    except KeyboardInterrupt:
        observer.stop()
        print("\n[Stream Server] Stopped", flush=True)
    finally:
        observer.join()


if __name__ == "__main__":
    main()
