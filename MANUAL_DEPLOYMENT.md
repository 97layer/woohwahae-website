# 97layerOS Manual Deployment (Without gcloud CLI)

> **목적**: gcloud CLI 없이 GCP VM에 수동 배포
> **대상**: 브라우저 SSH 또는 직접 VM 접속
> **소요 시간**: 20-30분

---

## 📋 사전 준비

### 1. GCP Console에서 VM 확인
1. https://console.cloud.google.com/compute/instances
2. 97layer-vm (또는 사용 중인 VM) 확인
3. "SSH" 버튼 클릭 → 브라우저 SSH 창 열림

### 2. 로컬 파일 준비
```bash
# 로컬 맥북에서 실행
cd ~/97layerOS

# .env 파일 확인
cat .env
# 필요한 환경 변수들이 있는지 확인:
# - TELEGRAM_BOT_TOKEN
# - ANTHROPIC_API_KEY (Claude)
# - GOOGLE_API_KEY (Gemini)
```

---

## 🚀 Step 1: VM에 프로젝트 폴더 생성

```bash
# GCP VM SSH 창에서 실행
cd ~
mkdir -p 97layerOS
cd 97layerOS
```

---

## 📤 Step 2: 코드 업로드 (2가지 방법)

### 방법 A: Git clone (추천)
```bash
# VM에서 실행
cd ~/97layerOS

# Git 설치 (없으면)
sudo apt update && sudo apt install -y git

# 코드 clone (GitHub/GitLab repo가 있다면)
git clone https://github.com/YOUR_USERNAME/97layerOS.git .
```

### 방법 B: 파일 직접 업로드 (Git repo 없으면)
1. GCP Console → VM 인스턴스 → "SSH" 옆 ⋮ 메뉴
2. "파일 업로드" 클릭
3. 로컬 97layerOS 폴더에서 파일들 선택:
   - `core/` 폴더 전체
   - `directives/` 폴더 전체
   - `knowledge/` 폴더 전체 (용량 주의)
   - `requirements.txt`
   - `deployment/` 폴더

4. 업로드된 파일들을 ~/97layerOS로 이동:
```bash
mv ~/core ~/97layerOS/
mv ~/directives ~/97layerOS/
mv ~/knowledge ~/97layerOS/
mv ~/requirements.txt ~/97layerOS/
mv ~/deployment ~/97layerOS/
```

### 방법 C: tar.gz 압축 후 업로드
```bash
# 로컬 맥북에서
cd ~/97layerOS
tar -czf 97layer-deploy.tar.gz \
  core/ directives/ knowledge/ \
  requirements.txt deployment/ \
  --exclude='__pycache__' --exclude='.venv'

# GCP Console에서 97layer-deploy.tar.gz 업로드

# VM에서 압축 해제
cd ~
tar -xzf 97layer-deploy.tar.gz -C ~/97layerOS/
```

---

## 🐍 Step 3: Python 환경 설정

```bash
# VM에서 실행
cd ~/97layerOS

# Python 3.11 설치
sudo apt update
sudo apt install -y python3.11 python3.11-venv python3-pip

# Virtual environment 생성
python3.11 -m venv .venv
source .venv/bin/activate

# Dependencies 설치
pip install --upgrade pip
pip install -r requirements.txt

# 설치 확인 (5-10분 소요)
python3 -c "from core.daemons.telegram_secretary import TelegramSecretary; print('✅ Import OK')"
```

---

## 🔐 Step 4: 환경 변수 설정

### 방법 A: .env 파일 직접 생성
```bash
# VM에서 실행
cd ~/97layerOS
nano .env
```

다음 내용 붙여넣기 (로컬 .env에서 복사):
```bash
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here
ANTHROPIC_API_KEY=your_claude_api_key_here
GOOGLE_API_KEY=your_gemini_api_key_here
```

저장: `Ctrl+X` → `Y` → `Enter`

### 방법 B: .env 파일 업로드
1. GCP Console에서 .env 파일 업로드
2. VM에서 이동:
```bash
mv ~/.env ~/97layerOS/.env
chmod 600 ~/97layerOS/.env
```

---

## 📋 Step 5: NotebookLM Credentials 복사

```bash
# 로컬 맥북에서 credentials 압축
cd ~
tar -czf notebooklm-creds.tar.gz .notebooklm-mcp-cli/

# GCP Console에서 notebooklm-creds.tar.gz 업로드

# VM에서 압축 해제
cd ~
tar -xzf notebooklm-creds.tar.gz

# 확인
ls -la ~/.notebooklm-mcp-cli/
```

---

## 🧪 Step 6: 테스트 실행

```bash
# VM에서 실행
cd ~/97layerOS
source .venv/bin/activate

# 로그 디렉토리 생성
mkdir -p logs

# Foreground 실행 (테스트)
python3 core/daemons/telegram_secretary.py
```

**다른 터미널/폰에서**:
- Telegram 봇에게 `/status` 메시지 전송
- 응답 확인

**테스트 성공하면**:
- `Ctrl+C`로 종료
- 다음 단계 진행

---

## 🔄 Step 7: Systemd Service 설정 (24/7 운영)

```bash
# VM에서 실행
cd ~/97layerOS

# Username 확인
whoami  # 예: your_username

# Service 파일 수정
sed "s/USERNAME_PLACEHOLDER/$(whoami)/g" deployment/97layer-telegram.service > /tmp/97layer-telegram.service

# Service 설치
sudo mv /tmp/97layer-telegram.service /etc/systemd/system/97layer-telegram.service

# Systemd 리로드
sudo systemctl daemon-reload

# Service 활성화 및 시작
sudo systemctl enable 97layer-telegram
sudo systemctl start 97layer-telegram

# 상태 확인
sudo systemctl status 97layer-telegram
```

**정상 작동 확인**:
```
● 97layer-telegram.service - 97layerOS Telegram Executive Secretary
   Loaded: loaded (/etc/systemd/system/97layer-telegram.service; enabled)
   Active: active (running) since ...
```

---

## 📊 Step 8: 모니터링

### 실시간 로그 확인
```bash
# Systemd logs
journalctl -u 97layer-telegram -f

# 또는 파일 로그
tail -f ~/97layerOS/logs/telegram.log
```

### 메모리 사용량 확인
```bash
free -h
ps aux | grep telegram_secretary
```

### Service 관리 명령어
```bash
# 재시작
sudo systemctl restart 97layer-telegram

# 중지
sudo systemctl stop 97layer-telegram

# 시작
sudo systemctl start 97layer-telegram

# 상태
sudo systemctl status 97layer-telegram
```

---

## ✅ 배포 완료 체크리스트

- [ ] VM에 코드 업로드 완료
- [ ] Python 3.11 + venv + dependencies 설치 완료
- [ ] .env 파일 설정 완료
- [ ] NotebookLM credentials 복사 완료
- [ ] Foreground 테스트 성공 (/status 응답 확인)
- [ ] Systemd service 실행 중
- [ ] 로그에 에러 없음
- [ ] 메모리 사용량 < 200MB

---

## 🚨 트러블슈팅

### Bot이 응답 안 함
```bash
# Service 상태 확인
sudo systemctl status 97layer-telegram

# 로그 확인
journalctl -u 97layer-telegram -n 50

# Import 테스트
cd ~/97layerOS
source .venv/bin/activate
python3 -c "from core.daemons.telegram_secretary import TelegramSecretary"
```

### 메모리 부족 (OOM)
```bash
# 메모리 확인
free -h

# 프로세스 확인
ps aux --sort=-%mem | head -10

# Service 재시작
sudo systemctl restart 97layer-telegram
```

### NotebookLM 작동 안 함
```bash
# Credentials 확인
ls -la ~/.notebooklm-mcp-cli/

# 재로그인 필요시 (로컬 맥북에서)
nlm login
# 그 후 credentials 다시 복사
```

---

## 📝 다음 단계

배포 완료 후:

1. **1-2일 모니터링**:
   - 메모리 사용량 추이
   - Claude API 호출 횟수
   - Bot 안정성

2. **기능 테스트**:
   - `/status` - 시스템 상태
   - `/report` - 일일 리포트
   - `/youtube [URL]` - NotebookLM 분석
   - `/analyze` - 자산 분석

3. **Phase 6 진행**:
   - 현재 시스템이 안정적이면
   - VM Ecosystem 구현 시작

---

> **슬로우 라이프**: 급하게 하지 말고, 각 단계 확인하며 진행. 문제 생기면 journalctl로 로그 확인.
