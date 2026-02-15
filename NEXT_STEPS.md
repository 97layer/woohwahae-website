# 97layerOS - 다음 단계 (Next Steps)

**현재 상태**: Phase 2 완료 (100%) - 모든 코드 준비 완료 ✅

---

## 🚀 즉시 실행 가능 (오늘 바로)

### 1. Telegram Bot 실행

```bash
# .env에 토큰 추가 (처음 1회만)
echo "TELEGRAM_BOT_TOKEN=your_telegram_bot_token" >> .env

# Bot 실행
./start_telegram.sh
```

**사용 가능한 명령어** (10개):
- `/start` - 비서 소개 및 도움말
- `/status` - 시스템 현재 상태
- `/report` - 오늘의 작업 보고
- `/analyze` - 마지막 신호 멀티에이전트 분석
- `/signal <텍스트>` - 새 신호 수동 입력
- `/morning` - 아침 브리핑 (09:00 권장)
- `/evening` - 저녁 리포트 (21:00 권장)
- `/search <검색어>` - 과거 지식 베이스 검색
- `/memo <메모>` - 빠른 메모 저장
- `/sync` - 클라우드 동기화 (수동)

**자동 기능**:
- 텍스트 메시지 → 자동 신호 포착
- 이미지 전송 → 비주얼 신호 포착
- 링크 공유 → 웹 콘텐츠 분석 (예정)

### 2. 실시간 모니터링 대시보드

```bash
# 새 터미널 창에서 실행
./start_monitor.sh

# 사용자 지정 갱신 주기 (3초마다)
./start_monitor.sh 3
```

**모니터링 항목**:
- 🔒 Work Lock: 현재 작업 중인 에이전트
- 📦 Asset Manager: 자산 통계 (상태별 분포)
- 🔄 Ralph Loop: 품질 검증 통계 (통과율)
- 📅 Daily Routine: 오늘 브리핑/리포트 완료 여부
- 📝 Recent Changes: 최근 5개 파일 수정 내역
- 🔀 Git: 브랜치, 변경사항, 최근 커밋

---

## ☁️ Phase 3.1: Google Drive 인증 설정 (30분)

### 목적
- 세션 연속성 클라우드 백업
- Telegram `/search`, `/sync` 작동
- NotebookLM 자동 소스 공급
- 모델 교체 시에도 맥락 보존

### 설정 방법

#### 1. Google Cloud Console에서 Service Account 생성

1. [Google Cloud Console](https://console.cloud.google.com/) 접속
2. 프로젝트 선택 또는 생성
3. **IAM & Admin** → **Service Accounts** 이동
4. **CREATE SERVICE ACCOUNT** 클릭
5. 이름: `97layer-gdrive-sync`
6. 역할: `Editor` 또는 `Owner`
7. **CREATE KEY** → JSON 선택
8. JSON 파일 다운로드 → `service-account-key.json`

#### 2. Google Drive API 활성화

1. **APIs & Services** → **Library** 이동
2. `Google Drive API` 검색
3. **ENABLE** 클릭

#### 3. Google Drive 폴더 ID 확인

1. Google Drive에서 NotebookLM 소스 폴더 생성
   - 이름: `97layerOS Knowledge Base` (예시)
2. 폴더 열기 → URL에서 ID 복사
   - URL 형식: `https://drive.google.com/drive/folders/{FOLDER_ID}`
   - FOLDER_ID: 복사해두기

#### 4. 로컬 설정

```bash
# 인증 파일 설치
mkdir -p credentials
mv ~/Downloads/service-account-key.json credentials/gdrive_auth.json
chmod 600 credentials/gdrive_auth.json

# .env에 폴더 ID 추가
echo "GOOGLE_DRIVE_FOLDER_ID=your_folder_id_here" >> .env
```

#### 5. 테스트

```bash
# INTELLIGENCE_QUANTA.md 업로드 테스트
python3 execution/system/gdrive_sync.py --intelligence

# 일일 리포트 업로드 테스트
python3 execution/system/gdrive_sync.py --reports

# 전체 동기화
python3 execution/system/gdrive_sync.py --all

# 검색 테스트
python3 execution/system/gdrive_sync.py --search "INTELLIGENCE"
```

#### 6. Telegram에서 확인

```
/sync                    # 수동 동기화
/search 슬로우 라이프     # 검색 테스트
/memo 테스트 메모         # 메모 + 자동 업로드
```

---

## 🐳 Phase 3.2: MCP 컨테이너 빌드 (10분)

### 목적
- Claude Desktop에서 Google Drive 직접 검색
- Container-First 원칙 완전 실현
- Node.js 네이티브 설치 없이 MCP 사용

### 빌드 방법

```bash
# 1. 컨테이너 빌드
cd execution/ops/mcp
./build_mcp_container.sh

# 2. 테스트 실행
./run_mcp_server.sh

# 3. Claude Desktop 설정
# ~/Library/Application Support/Claude/claude_desktop_config.json 편집
# claude_desktop_config.json 내용 복사

# 4. Claude Desktop 재시작
# MCP 서버가 자동으로 연결됨
```

### Claude Desktop에서 사용

```
# Claude Desktop에서 대화 시작
"INTELLIGENCE_QUANTA.md의 최신 내용을 보여줘"
"지난 주 일일 리포트를 요약해줘"
"슬로우 라이프 관련 문서를 검색해줘"
```

---

## 🔄 Phase 3.3: 자동 스케줄링 (선택사항, 1시간)

### 목적
- 매일 09:00 자동 아침 브리핑
- 매일 21:00 자동 저녁 리포트
- 일요일 21:00 자동 주간 요약

### 구현 방법

#### 1. APScheduler 설치

```bash
pip install apscheduler
```

#### 2. telegram_secretary.py에 스케줄러 추가

```python
from apscheduler.schedulers.asyncio import AsyncIOScheduler

class TelegramSecretary:
    def __init__(self, bot_token: str):
        # ... 기존 코드 ...

        # 스케줄러 초기화
        self.scheduler = AsyncIOScheduler()

        # 아침 브리핑 (매일 09:00)
        self.scheduler.add_job(
            self._auto_morning_briefing,
            'cron',
            hour=9,
            minute=0,
            id='morning_briefing'
        )

        # 저녁 리포트 (매일 21:00)
        self.scheduler.add_job(
            self._auto_evening_report,
            'cron',
            hour=21,
            minute=0,
            id='evening_report'
        )

        # 주간 요약 (일요일 21:00)
        self.scheduler.add_job(
            self._auto_weekly_summary,
            'cron',
            day_of_week='sun',
            hour=21,
            minute=0,
            id='weekly_summary'
        )

        self.scheduler.start()
        logger.info("✅ 자동 스케줄러 시작됨")

    async def _auto_morning_briefing(self):
        """자동 아침 브리핑 (모든 사용자에게)"""
        briefing = self.daily_routine.morning_briefing()
        # Telegram 메시지 전송 로직...

    async def _auto_evening_report(self):
        """자동 저녁 리포트"""
        report = self.daily_routine.evening_report()
        # Telegram 메시지 전송 로직...

    async def _auto_weekly_summary(self):
        """자동 주간 요약"""
        summary = self.daily_routine.weekly_summary()
        # Telegram 메시지 전송 로직...
```

---

## 📊 Phase 3.4: 성과 측정 대시보드 (선택사항, 2시간)

### 목적
- 월간 생산성 측정
- 품질 트렌드 분석
- 시각화된 리포트

### 구현 옵션

#### Option A: 터미널 대시보드 (Rich 라이브러리)

```python
from rich.console import Console
from rich.table import Table
from rich.live import Live

# 컬러풀한 테이블, 차트, 진행 바
```

#### Option B: 웹 대시보드 (FastAPI + Chart.js)

```python
# FastAPI로 REST API
# Chart.js로 시각화
# localhost:8000에서 접속
```

---

## 🎯 추천 진행 순서

### 오늘 바로 (필수)
1. ✅ **Telegram Bot 실행** → `./start_telegram.sh`
2. ✅ **모니터링 시작** → `./start_monitor.sh`
3. ✅ **Telegram에서 테스트** → `/start`, `/morning`, `/status`

### 이번 주 (우선순위)
1. ⏳ **Google Drive 인증** → 클라우드 동기화 활성화
2. ⏳ **MCP 컨테이너 빌드** → Claude Desktop 연동
3. ⏳ **실제 신호 처리** → 멀티에이전트 검증

### 다음 주 (선택사항)
1. ⏳ **자동 스케줄링** → APScheduler 추가
2. ⏳ **NotebookLM API** → `/ask` 명령어
3. ⏳ **성과 측정 대시보드** → 월간 통계

### 나중에 (Enhancement)
- Telegram Bot 컨테이너화
- 웹 인터페이스 (PWA)
- Slack 통합
- GitHub Actions CI/CD

---

## 📁 중요 파일 경로

### 실행 스크립트
- `./start_telegram.sh` - Telegram Bot 시작
- `./start_monitor.sh` - 모니터링 대시보드

### 설정 파일
- `.env` - 환경 변수 (TELEGRAM_BOT_TOKEN, GOOGLE_DRIVE_FOLDER_ID)
- `credentials/gdrive_auth.json` - Google Drive 인증
- `execution/ops/mcp/claude_desktop_config.json` - Claude Desktop MCP 설정

### 핵심 코드
- `execution/daemons/telegram_secretary.py` - Telegram Bot 메인
- `execution/system/daily_routine.py` - 일일 자동화
- `execution/system/gdrive_sync.py` - Google Drive 동기화
- `execution/system/monitor_dashboard.py` - 모니터링 대시보드
- `execution/system/parallel_orchestrator.py` - 멀티에이전트 협업
- `execution/system/ralph_loop.py` - 품질 검증

### 지식 베이스
- `knowledge/agent_hub/INTELLIGENCE_QUANTA.md` - 세션 연속성
- `knowledge/reports/daily/` - 일일 브리핑/리포트
- `knowledge/system/asset_registry.json` - 자산 통계
- `knowledge/system/ralph_validations.jsonl` - 품질 검증 로그

### 문서
- `execution/ops/mcp/README.md` - MCP 설정 가이드
- `NEXT_STEPS.md` - 본 파일

---

## ❓ 문제 해결

### Telegram Bot이 시작되지 않음

```bash
# 1. 토큰 확인
echo $TELEGRAM_BOT_TOKEN

# 2. Python 패키지 설치
pip install python-telegram-bot

# 3. 로그 확인
tail -f logs/telegram_secretary.log
```

### Google Drive 동기화 실패

```bash
# 1. 인증 파일 확인
ls -la credentials/gdrive_auth.json

# 2. 권한 확인
chmod 600 credentials/gdrive_auth.json

# 3. .env 확인
grep GOOGLE_DRIVE_FOLDER_ID .env

# 4. Python 패키지 설치
pip install google-auth google-api-python-client
```

### MCP 컨테이너 빌드 실패

```bash
# 1. Podman 확인
podman --version

# 2. Podman 시작
podman machine start

# 3. TMPDIR 설정
export TMPDIR=/tmp
```

---

## 💡 팁

### 분할 화면 추천
```
┌─────────────────┬─────────────────┐
│  터미널 1       │  터미널 2       │
│  모니터링       │  Telegram Bot   │
│  ./start_mon... │  ./start_tel... │
├─────────────────┴─────────────────┤
│  터미널 3                         │
│  작업 공간 (git, 테스트 등)        │
└───────────────────────────────────┘
```

### tmux 사용 (선택사항)
```bash
# 새 세션 생성
tmux new -s 97layer

# 화면 분할
Ctrl+B "  # 수평 분할
Ctrl+B %  # 수직 분할

# 이동
Ctrl+B 화살표

# 세션 나가기
Ctrl+B D

# 다시 접속
tmux attach -t 97layer
```

---

**97layerOS - Slow Life Archive System**
*Container-First. Context-Preserved. Cloud-Synced.*

마지막 업데이트: 2026-02-16
