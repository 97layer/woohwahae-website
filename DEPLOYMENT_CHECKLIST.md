# 97layerOS Deployment Checklist

> **목적**: 순차적 배포 체크리스트
> **날짜**: 2026-02-16
> **목표**: GCP VM에 Multi-Agent System 배포

---

## ✅ 사전 준비 (완료)

- [x] Clean Architecture refactoring (Phase 5)
- [x] Queue infrastructure (Phase 6.1)
- [x] 5 Independent agents (Phase 6.2)
- [x] Multi-agent workflow test
- [x] Deployment scripts 작성
- [x] requirements.txt 정리

---

## 📋 배포 단계

### 1️⃣ 로컬 환경 점검

**현재 상태**:
```bash
✅ Podman 5.7.1 설치됨
✅ 97layer-workspace 컨테이너 실행 중
✅ .env 파일 존재 (API keys 확인됨)
   - TELEGRAM_BOT_TOKEN: ✅
   - GOOGLE_API_KEY: ✅
   - GEMINI_API_KEY: ✅
   - ANTHROPIC_API_KEY: ⚠️ (검증 필요)
```

**체크리스트**:
- [ ] .env 파일 백업
- [ ] NotebookLM credentials 백업 (~/.notebooklm-mcp-cli/)
- [ ] Git 최신 commit 확인
- [ ] 로컬 테스트 마지막 확인

---

### 2️⃣ GCP VM 준비

**VM 정보** (사용자 제공 필요):
```
VM Name: ?
Zone: ?
IP Address: ?
SSH Access: GCP Console browser SSH 또는 gcloud
```

**체크리스트**:
- [ ] GCP VM 실행 중 확인
- [ ] VM SSH 접속 테스트
- [ ] VM 디스크 용량 확인 (최소 10GB 여유)
- [ ] VM 메모리 확인 (e2-micro = 1GB)

---

### 3️⃣ 코드 배포

**방법 선택**:
- [ ] **Option A**: Git clone (추천, GitHub/GitLab repo 있으면)
- [ ] **Option B**: tar.gz 업로드 (repo 없으면)
- [ ] **Option C**: rsync (gcloud CLI 있으면)

**진행**:
```bash
# VM에서 실행
cd ~
mkdir -p 97layerOS
cd 97layerOS

# Option A: Git clone
git clone <REPO_URL> .

# Option B: tar.gz 업로드 (로컬에서 먼저)
# 로컬: tar -czf 97layer-deploy.tar.gz core/ directives/ knowledge/ requirements.txt deployment/
# VM: tar -xzf 97layer-deploy.tar.gz
```

**체크리스트**:
- [ ] 코드 업로드 완료
- [ ] 폴더 구조 확인 (core/, knowledge/, deployment/)
- [ ] 파일 권한 확인

---

### 4️⃣ Python 환경 설정

```bash
# VM에서 실행
cd ~/97layerOS

# Python 3.11 설치
sudo apt update
sudo apt install -y python3.11 python3.11-venv python3-pip git

# Virtual environment
python3.11 -m venv .venv
source .venv/bin/activate

# Dependencies
pip install --upgrade pip
pip install -r requirements.txt

# 검증
python3 -c "from core.agents.sa_agent import StrategyAnalyst; print('✅ Import OK')"
```

**체크리스트**:
- [ ] Python 3.11 설치됨
- [ ] Virtual environment 생성됨
- [ ] requirements.txt 설치 완료 (5-10분 소요)
- [ ] Import 테스트 통과

---

### 5️⃣ 환경 변수 설정

```bash
# VM에서 실행
cd ~/97layerOS
nano .env
```

**필수 환경 변수**:
```bash
TELEGRAM_BOT_TOKEN=<your_token>
TELEGRAM_CHAT_ID=<your_chat_id>
GOOGLE_API_KEY=<your_gemini_key>
ANTHROPIC_API_KEY=<your_claude_key>
TZ=Asia/Seoul
```

**체크리스트**:
- [ ] .env 파일 생성됨
- [ ] 모든 API keys 설정됨
- [ ] 파일 권한 600 (chmod 600 .env)

---

### 6️⃣ NotebookLM Credentials (선택사항)

```bash
# 로컬에서 credentials 압축
cd ~
tar -czf notebooklm-creds.tar.gz .notebooklm-mcp-cli/

# GCP Console에서 업로드 후 VM에서
cd ~
tar -xzf notebooklm-creds.tar.gz

# 확인
ls -la ~/.notebooklm-mcp-cli/
```

**체크리스트**:
- [ ] Credentials 복사 완료
- [ ] nlm notebook list 작동 확인 (선택사항)

---

### 7️⃣ Foreground 테스트

```bash
# VM에서 실행
cd ~/97layerOS
source .venv/bin/activate

# Telegram bot 테스트
python3 core/daemons/telegram_secretary.py
```

**Telegram에서 테스트**:
- [ ] `/status` 명령 → 응답 확인
- [ ] Bot이 메시지 받음
- [ ] 로그에 에러 없음

**성공하면**: `Ctrl+C`로 종료, 다음 단계 진행

---

### 8️⃣ Systemd Service 설정

```bash
# VM에서 실행
cd ~/97layerOS

# Service 파일 준비
sed "s/USERNAME_PLACEHOLDER/$(whoami)/g" deployment/97layer-telegram.service > /tmp/97layer-telegram.service

# Service 설치
sudo mv /tmp/97layer-telegram.service /etc/systemd/system/97layer-telegram.service

# Systemd 설정
sudo systemctl daemon-reload
sudo systemctl enable 97layer-telegram
sudo systemctl start 97layer-telegram

# 상태 확인
sudo systemctl status 97layer-telegram
```

**체크리스트**:
- [ ] Service 파일 설치됨
- [ ] Service 실행 중 (`active (running)`)
- [ ] 로그에 에러 없음

---

### 9️⃣ 모니터링

```bash
# 실시간 로그
journalctl -u 97layer-telegram -f

# 또는 파일 로그
tail -f ~/97layerOS/logs/telegram.log

# 메모리 확인
free -h
ps aux | grep telegram_secretary
```

**체크리스트**:
- [ ] 로그 정상 출력
- [ ] 메모리 사용량 < 200MB
- [ ] Bot이 메시지 응답함

---

### 🔟 Multi-Agent 통합 (다음 단계)

**현재 상태**: Telegram bot만 배포됨
**다음 작업**: Multi-agent를 Telegram bot에 통합

```python
# core/daemons/telegram_secretary.py에 추가
from core.agents.sa_agent import StrategyAnalyst
from core.agents.ad_agent import ArtDirector
from core.agents.ce_agent import ChiefEditor
from core.agents.ralph_agent import RalphLoop

# /analyze 명령에서 multi-agent workflow 실행
```

**체크리스트**:
- [ ] SA → AD → CE → Ralph 순차 실행 통합
- [ ] Telegram으로 진행 상황 알림
- [ ] 최종 결과물 전송

---

## 🚨 트러블슈팅

### Bot이 응답 안 함
```bash
# Service 상태
sudo systemctl status 97layer-telegram

# 로그 확인
journalctl -u 97layer-telegram -n 50

# Import 테스트
cd ~/97layerOS
source .venv/bin/activate
python3 -c "from core.daemons.telegram_secretary import TelegramSecretary"
```

### 메모리 부족
```bash
# 메모리 확인
free -h

# 프로세스 확인
ps aux --sort=-%mem | head -10

# Service 재시작
sudo systemctl restart 97layer-telegram
```

### API 에러
- TELEGRAM_BOT_TOKEN: BotFather에서 재발급
- GOOGLE_API_KEY: Google AI Studio에서 확인
- ANTHROPIC_API_KEY: Anthropic Console에서 확인

---

## 📊 성공 기준

✅ **최소 요구사항**:
- [ ] Telegram bot 24/7 실행 중
- [ ] `/status` 명령 응답함
- [ ] 메모리 사용 < 200MB
- [ ] 로그에 critical error 없음

✅ **이상적 상태**:
- [ ] Multi-agent workflow 통합됨
- [ ] `/analyze` 명령으로 전체 파이프라인 실행
- [ ] NotebookLM 연동 작동
- [ ] 비용 $10/month 이내

---

## 📝 배포 후 작업

1. **1-2일 모니터링**:
   - 메모리 사용량 추이
   - API 호출 횟수 (Claude 특히)
   - Bot 안정성

2. **Multi-Agent 통합**:
   - Telegram bot에 SA → AD → CE → Ralph 통합
   - Queue-based 실행으로 전환 (Phase 6.3)

3. **Phase 6.3 진행**:
   - Podman Compose orchestration
   - Container-based agents
   - APScheduler 자동화

---

> **슬로우 라이프**: 한 단계씩, 천천히, 확실하게. 문제 생기면 롤백 후 재시도.
