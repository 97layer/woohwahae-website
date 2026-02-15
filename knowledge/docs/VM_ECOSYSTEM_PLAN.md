# 97layerOS VM Ecosystem - True Multi-Agent Architecture

> **목표**: LLM 연극놀이 → 진짜 독립 Agent들의 VM 생태계

## 🎯 핵심 개념

### 현재 문제 (Level 1: LLM 연극놀이)
- 하나의 Python 프로세스에서 여러 Agent "흉내"
- LLM이 role prompt만 바꿔서 연기
- 실제 도구 사용 없음 (텍스트만 생성)
- 사용자 없으면 멈춤

### 목표 (Level 2-3: VM 생태계)
- 각 Agent가 독립 Podman 컨테이너
- 실제 도구 사용 (Stable Diffusion, FFmpeg, Playwright)
- 비동기 메시지 큐로 통신
- 자율 순환 (사용자 없이도 작동)

---

## 🏗️ 아키텍처

### 1. 컨테이너 구성

```
97layerOS VM Ecosystem:

┌─────────────────────────────────────────────────────────────┐
│ Host (macOS or GCP VM)                                      │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ SA Container │  │ AD Container │  │ CE Container │      │
│  │              │  │              │  │              │      │
│  │ • Playwright │  │ • Stable     │  │ • Claude API │      │
│  │ • BeautifulS │  │   Diffusion  │  │ • Markdown   │      │
│  │ • Trend API  │  │ • DALL-E     │  │   Parser     │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │              │
│         └──────────────────┼──────────────────┘              │
│                            │                                 │
│                    ┌───────▼─────────┐                       │
│                    │  Message Queue  │                       │
│                    │  (.infra/queue) │                       │
│                    └───────┬─────────┘                       │
│                            │                                 │
│         ┌──────────────────┼──────────────────┐              │
│         │                  │                  │              │
│  ┌──────▼───────┐  ┌──────▼───────┐  ┌──────▼───────┐      │
│  │ CD Container │  │ Ralph Loop   │  │ Orchestrator │      │
│  │              │  │              │  │              │      │
│  │ • Git commit │  │ • Quality    │  │ • Scheduler  │      │
│  │ • Telegram   │  │   Check      │  │ • Event Bus  │      │
│  │   Bot        │  │ • STAP       │  │ • Health Mon │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                              │
│  Shared Volumes:                                            │
│  • /knowledge (read-write)                                  │
│  • /queue (read-write)                                      │
│  • .infra/logs (write-only)                                 │
└─────────────────────────────────────────────────────────────┘
```

### 2. Message Queue (File-based)

```
.infra/queue/
├── events/           # 이벤트 발행 (JSON)
│   ├── trend_detected_20260216_090001.json
│   ├── signal_ready_20260216_090112.json
│   ├── visual_ready_20260216_090345.json
│   └── draft_ready_20260216_090521.json
│
├── tasks/            # 작업 대기열
│   ├── pending/
│   ├── processing/
│   └── completed/
│
└── locks/            # Agent 작업 잠금
    └── sa_agent.lock
```

**메시지 포맷**:
```json
{
  "event_id": "trend_detected_20260216_090001",
  "type": "trend_detected",
  "timestamp": "2026-02-16T09:00:01Z",
  "payload": {
    "source": "youtube",
    "keyword": "slow living",
    "trend_score": 87
  },
  "next_agent": "strategy-analyst"
}
```

### 3. Agent 독립 스크립트

#### SA Agent (agents/sa/agent_sa.py)
```python
#!/usr/bin/env python3
"""
Strategy Analyst Agent - 독립 실행 컨테이너
역할: 웹 크롤링, 트렌드 분석
"""

import asyncio
from pathlib import Path
from playwright.async_api import async_playwright

QUEUE_DIR = Path("/workspace/queue")
KNOWLEDGE_DIR = Path("/workspace/knowledge")

class StrategyAnalyst:
    async def watch_queue(self):
        """큐를 감시하고 trend_detected 이벤트 처리"""
        while True:
            events = list(QUEUE_DIR.glob("events/trend_detected_*.json"))
            for event_file in events:
                await self.process_trend(event_file)
                event_file.rename(QUEUE_DIR / "events" / "processed" / event_file.name)
            await asyncio.sleep(5)

    async def process_trend(self, event_file):
        """실제 웹 크롤링 수행"""
        # 1. Playwright로 YouTube 크롤링
        async with async_playwright() as p:
            browser = await p.chromium.launch()
            page = await browser.new_page()
            await page.goto("https://youtube.com/...")
            # ... 실제 크롤링 로직

        # 2. 결과 저장
        result_file = KNOWLEDGE_DIR / "signals" / "trend_analysis.md"
        result_file.write_text(analysis_result)

        # 3. 다음 이벤트 발행
        next_event = {
            "event_id": f"signal_ready_{timestamp}",
            "type": "signal_ready",
            "payload": {"signal_file": str(result_file)}
        }
        (QUEUE_DIR / "events" / f"signal_ready_{timestamp}.json").write_text(
            json.dumps(next_event, indent=2)
        )

if __name__ == "__main__":
    sa = StrategyAnalyst()
    asyncio.run(sa.watch_queue())
```

#### AD Agent (agents/ad/agent_ad.py)
```python
#!/usr/bin/env python3
"""
Art Director Agent - 독립 실행 컨테이너
역할: 이미지 생성 (Stable Diffusion)
"""

import asyncio
from pathlib import Path
import requests
import os

QUEUE_DIR = Path("/workspace/queue")
KNOWLEDGE_DIR = Path("/workspace/knowledge")
STABILITY_API_KEY = os.getenv("STABILITY_API_KEY")

class ArtDirector:
    async def watch_queue(self):
        """signal_ready 이벤트 감시"""
        while True:
            events = list(QUEUE_DIR.glob("events/signal_ready_*.json"))
            for event_file in events:
                await self.generate_visual(event_file)
                event_file.rename(QUEUE_DIR / "events" / "processed" / event_file.name)
            await asyncio.sleep(5)

    async def generate_visual(self, event_file):
        """Stable Diffusion으로 실제 이미지 생성"""
        # 1. Signal 읽기
        event = json.loads(event_file.read_text())
        signal_file = Path(event["payload"]["signal_file"])
        signal_content = signal_file.read_text()

        # 2. Stable Diffusion API 호출
        response = requests.post(
            "https://api.stability.ai/v1/generation/stable-diffusion-xl-1024-v1-0/text-to-image",
            headers={"Authorization": f"Bearer {STABILITY_API_KEY}"},
            json={
                "text_prompts": [{"text": self._extract_visual_prompt(signal_content)}],
                "cfg_scale": 7,
                "height": 1024,
                "width": 1024,
                "samples": 1
            }
        )

        # 3. 이미지 저장
        image_data = response.json()["artifacts"][0]["base64"]
        image_file = KNOWLEDGE_DIR / "assets" / "images" / f"visual_{timestamp}.png"
        image_file.write_bytes(base64.b64decode(image_data))

        # 4. 다음 이벤트 발행
        next_event = {
            "event_id": f"visual_ready_{timestamp}",
            "type": "visual_ready",
            "payload": {
                "signal_file": str(signal_file),
                "image_file": str(image_file)
            }
        }
        (QUEUE_DIR / "events" / f"visual_ready_{timestamp}.json").write_text(
            json.dumps(next_event, indent=2)
        )

if __name__ == "__main__":
    ad = ArtDirector()
    asyncio.run(ad.watch_queue())
```

---

## 📦 Podman Compose 설정

### docker-compose.yml

```yaml
version: '3.8'

services:
  # Strategy Analyst: 웹 크롤링
  strategy-analyst:
    build:
      context: ./agents/sa
      dockerfile: Dockerfile
    volumes:
      - ./knowledge:/workspace/knowledge
      - ./.infra/queue:/workspace/queue
      - ./.infra/logs:/workspace/logs
    environment:
      - AGENT_NAME=strategy-analyst
      - LOG_LEVEL=INFO
    restart: unless-stopped
    command: python agent_sa.py --watch-queue

  # Art Director: 이미지 생성
  art-director:
    build:
      context: ./agents/ad
      dockerfile: Dockerfile
    volumes:
      - ./knowledge:/workspace/knowledge
      - ./.infra/queue:/workspace/queue
      - ./.infra/logs:/workspace/logs
    environment:
      - AGENT_NAME=art-director
      - STABILITY_API_KEY=${STABILITY_API_KEY}
      - LOG_LEVEL=INFO
    restart: unless-stopped
    command: python agent_ad.py --watch-queue
    depends_on:
      - strategy-analyst

  # Chief Editor: 콘텐츠 작성
  chief-editor:
    build:
      context: ./agents/ce
      dockerfile: Dockerfile
    volumes:
      - ./knowledge:/workspace/knowledge
      - ./.infra/queue:/workspace/queue
      - ./.infra/logs:/workspace/logs
    environment:
      - AGENT_NAME=chief-editor
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - LOG_LEVEL=INFO
    restart: unless-stopped
    command: python agent_ce.py --watch-queue
    depends_on:
      - art-director

  # Creative Director: 최종 승인 및 배포
  creative-director:
    build:
      context: ./agents/cd
      dockerfile: Dockerfile
    volumes:
      - ./knowledge:/workspace/knowledge
      - ./.infra/queue:/workspace/queue
      - ./.infra/logs:/workspace/logs
      - ~/.ssh:/root/.ssh:ro  # Git push용
    environment:
      - AGENT_NAME=creative-director
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - LOG_LEVEL=INFO
    restart: unless-stopped
    command: python agent_cd.py --watch-queue
    depends_on:
      - chief-editor

  # Orchestrator: 전체 조율 및 스케줄링
  orchestrator:
    build:
      context: ./core/orchestrator
      dockerfile: Dockerfile
    volumes:
      - ./knowledge:/workspace/knowledge
      - ./.infra/queue:/workspace/queue
      - ./.infra/logs:/workspace/logs
    environment:
      - AGENT_NAME=orchestrator
      - LOG_LEVEL=INFO
    restart: unless-stopped
    command: python orchestrator.py --schedule
    depends_on:
      - strategy-analyst
      - art-director
      - chief-editor
      - creative-director

  # Ralph Loop: 품질 검증
  ralph-loop:
    build:
      context: ./core/ralph
      dockerfile: Dockerfile
    volumes:
      - ./knowledge:/workspace/knowledge
      - ./.infra/queue:/workspace/queue
      - ./.infra/logs:/workspace/logs
    environment:
      - AGENT_NAME=ralph-loop
      - LOG_LEVEL=INFO
    restart: unless-stopped
    command: python ralph_loop.py --watch-queue

volumes:
  knowledge:
  queue:
  logs:
```

### Agent Dockerfile 예시 (agents/sa/Dockerfile)

```dockerfile
FROM python:3.11-slim

# Playwright 의존성
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Python 패키지
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
RUN playwright install chromium
RUN playwright install-deps chromium

# Agent 스크립트
COPY agent_sa.py .

CMD ["python", "agent_sa.py", "--watch-queue"]
```

---

## 🔄 자율 순환 시나리오

### 시나리오 1: 트렌드 자동 포착 → 콘텐츠 생성

```
[09:00] Orchestrator Container
└─> APScheduler: morning_trend_check()
    └─> queue/events/trend_detected.json 발행

[09:00:05] SA Container
├─ trend_detected.json 감지
├─ Playwright로 YouTube/Twitter 크롤링
├─ 결과: knowledge/signals/trend_20260216.md
└─> queue/events/signal_ready.json 발행

[09:03:20] AD Container
├─ signal_ready.json 감지
├─ Stable Diffusion API → 이미지 생성
├─ 결과: knowledge/assets/images/trend_visual.png
└─> queue/events/visual_ready.json 발행

[09:05:10] CE Container
├─ visual_ready.json 감지
├─ Claude API → 블로그 포스트 작성
├─ 결과: knowledge/content/blog_draft.md
└─> queue/events/draft_ready.json 발행

[09:06:30] Ralph Loop Container
├─ draft_ready.json 감지
├─ STAP 품질 검증 (80/100 통과)
└─> queue/events/quality_approved.json 발행

[09:07:00] CD Container
├─ quality_approved.json 감지
├─ Git commit + push
└─> Telegram Bot 알림: "✅ 새 콘텐츠 발행!"
```

### 시나리오 2: 자산 품질 자동 개선

```
[21:00] Orchestrator Container
└─> evening_quality_check()
    └─> Ralph Loop에게 오늘 자산 검사 요청

[21:01] Ralph Loop Container
├─ knowledge/assets/ 전체 스캔
├─ asset_001.md: 55/100 (낮음!)
└─> queue/events/refinement_needed.json 발행

[21:02] CE Container
├─ refinement_needed.json 감지
├─ asset_001.md 재작성
├─> knowledge/assets/asset_001_v2.md 저장
└─> queue/events/refinement_done.json 발행

[21:03] Ralph Loop Container
├─ refinement_done.json 감지
├─ asset_001_v2.md 재검사: 82/100 (통과!)
└─> AssetManager 상태 업데이트: refined
```

---

## 🛠️ 구현 단계

### Phase 6.1: Queue 인프라 (1-2일)
- [ ] .infra/queue/ 폴더 구조 생성
- [ ] Message format 정의 (JSON schema)
- [ ] File watcher 유틸리티 (queue_watcher.py)
- [ ] Event publisher/subscriber 기본 구현

### Phase 6.2: Agent 독립화 (2-3일)
- [ ] SA Agent 독립 스크립트 (agent_sa.py)
- [ ] AD Agent 독립 스크립트 (agent_ad.py)
- [ ] CE Agent 독립 스크립트 (agent_ce.py)
- [ ] CD Agent 독립 스크립트 (agent_cd.py)
- [ ] 각 Agent Dockerfile 작성

### Phase 6.3: Podman Compose (1일)
- [ ] docker-compose.yml 작성
- [ ] Volume 마운트 설정
- [ ] 환경 변수 관리 (.env)
- [ ] Podman Compose 테스트

### Phase 6.4: Tool Integration (3-5일)
- [ ] SA: Playwright 웹 크롤링
- [ ] AD: Stable Diffusion API 통합
- [ ] Audio Agent: ElevenLabs/TTS 통합
- [ ] Video Agent: FFmpeg 영상 편집

### Phase 6.5: Orchestrator (2일)
- [ ] APScheduler 스케줄링
- [ ] Health Monitor (Agent 상태 감시)
- [ ] Auto-recovery (실패 시 재시작)
- [ ] Dashboard (실시간 상태 표시)

---

## 💡 핵심 차이점 요약

| | Level 1 (현재) | Level 2-3 (VM 생태계) |
|---|---|---|
| **실행 방식** | 단일 Python 프로세스 | 각 Agent 독립 컨테이너 |
| **통신** | 함수 호출 (동기) | 메시지 큐 (비동기) |
| **도구 사용** | ❌ LLM만 | ✅ Stable Diffusion, FFmpeg, Playwright |
| **결과물** | Markdown 문서 | PNG, MP4, MP3 실제 파일 |
| **자율 작동** | ❌ 사용자 필요 | ✅ Scheduler + 자율 순환 |
| **확장성** | 어려움 | 쉬움 (컨테이너 추가) |
| **디버깅** | 쉬움 | 어려움 (로그 분산) |

---

## 🎯 예상 효과

1. **진정한 멀티모달**: 텍스트+이미지+오디오+비디오 실제 생성
2. **자율 순환**: 사용자 없이도 트렌드 감지 → 콘텐츠 생성
3. **확장 용이**: 새 Agent (Video, Audio) 추가가 독립적
4. **Tool 자유도**: 각 Agent가 필요한 도구 자유롭게 사용
5. **내구성**: 한 Agent 실패해도 다른 Agent는 계속 작동

---

## 📚 참고 자료

- Podman Compose: https://github.com/containers/podman-compose
- Stable Diffusion API: https://platform.stability.ai/docs
- Playwright: https://playwright.dev/python/
- APScheduler: https://apscheduler.readthedocs.io/

---

> "진짜 협업은 독립된 주체들이 비동기로 소통할 때 일어난다." — 97layerOS
