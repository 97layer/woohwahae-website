# 🚀 GCP 최종 배포 가이드 (완전 자동화)

## ✅ 준비 완료 상태

### Mac에서 생성된 파일:
1. `/tmp/97layerOS_full_deploy.tar.gz` (94MB) - 전체 시스템
2. `/tmp/deploy_on_gcp.sh` - GCP 자동 배포 스크립트
3. `/tmp/gcp_auto_download.sh` - GCP 자동 다운로드 스크립트 (참고용)

---

## 🎯 최종 배포 방법 (가장 간단)

### 단계 1: GCP 브라우저 SSH 접속

1. https://console.cloud.google.com 접속
2. **Compute Engine** → **VM instances**
3. `debian-micro-instance` 행에서 **SSH** 버튼 클릭

### 단계 2: 간단한 한 줄 명령어

GCP SSH 터미널에 다음 명령어를 **복사-붙여넣기** (클립보드에 준비됨):

```bash
curl -sL https://raw.githubusercontent.com/97layer/97layerOS/main/deploy.sh 2>/dev/null || (
cat > /tmp/quick_deploy.sh << 'EOF'
#!/bin/bash
echo "⚡ 97layerOS 빠른 배포 시작..."
cd ~
# 기존 프로세스 중지
pkill -f "technical_daemon.py" || true
pkill -f "telegram_daemon.py" || true
sleep 2
# 백업
cp 97layerOS/.env /tmp/backup_env 2>/dev/null || true
# 최신 코드만 유지 (중요 파일은 그대로)
cd 97layerOS
git pull 2>/dev/null || echo "Git pull 스킵"
# 데몬 재시작
source .venv/bin/activate || python3 -m venv .venv && source .venv/bin/activate
pip install -q google-generativeai python-dotenv requests
nohup python execution/technical_daemon.py > /tmp/technical_daemon.log 2>&1 &
nohup python execution/telegram_daemon.py > /tmp/telegram_daemon.log 2>&1 &
sleep 3
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep
echo "✅ 배포 완료!"
EOF
bash /tmp/quick_deploy.sh
)
```

---

## 🔄 대안: 수동 업로드 방식 (더 안정적)

### A. 파일 업로드

GCP SSH 창 상단:
1. **톱니바퀴 ⚙️** 아이콘 클릭
2. **"Upload file"** 선택
3. 다음 파일 업로드:
   - `/tmp/97layerOS_full_deploy.tar.gz` (94MB)
   - `/tmp/deploy_on_gcp.sh`

### B. 배포 실행

업로드 완료 후 GCP SSH에서:

```bash
bash /tmp/deploy_on_gcp.sh
```

---

## 📊 배포 후 확인

### 1. 프로세스 확인
```bash
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep
```

예상 출력:
```
skyto5339  12345  0.5  2.1  ... python execution/technical_daemon.py
skyto5339  12346  0.3  1.8  ... python execution/telegram_daemon.py
```

### 2. 로그 확인
```bash
tail -f /tmp/technical_daemon.log
tail -f /tmp/telegram_daemon.log
```

### 3. Telegram 테스트
Telegram에서 `/status` 전송 → 응답 확인

---

## 🔧 동기화 시스템 설정 (배포 후)

### GCP → Google Drive 자동 동기화

#### 방법 1: 간단 스크립트 (권장)

```bash
cd ~/97layerOS

# 동기화 스크립트 생성
cat > execution/ops/gcp_sync_simple.sh << 'EOF'
#!/bin/bash
# GCP knowledge → tar 패키지 생성
cd ~/97layerOS
timestamp=$(date +%Y%m%d_%H%M%S)
tar czf /tmp/knowledge_$timestamp.tar.gz knowledge/
echo "✅ 패키지 생성: /tmp/knowledge_$timestamp.tar.gz"
ls -lh /tmp/knowledge_$timestamp.tar.gz
EOF

chmod +x execution/ops/gcp_sync_simple.sh

# Cron 등록 (5분마다)
crontab -e
# 추가:
# */5 * * * * /home/skyto5339/97layerOS/execution/ops/gcp_sync_simple.sh >> /tmp/gcp_sync.log 2>&1
```

#### 방법 2: 수동 동기화 (필요할 때만)

GCP에서:
```bash
cd ~/97layerOS
tar czf /tmp/knowledge_latest.tar.gz knowledge/
```

GCP 브라우저 SSH:
- **톱니바퀴 → Download file**
- 경로: `/tmp/knowledge_latest.tar.gz`

Mac에서:
```bash
cd ~/내\ 드라이브\(skyto5339@gmail.com\)/97layerOS/
tar xzf ~/Downloads/knowledge_latest.tar.gz
```

---

## 🎯 완전 자동 동기화 (고급)

### 요구사항:
- GCP에 Google Drive API 인증 필요
- Mac에서 생성한 토큰 전송 필요

### 설정 방법:

1. Mac에서 인증 토큰 생성:
```bash
cd /Users/97layer/97layerOS
python3 -c "
from google_auth_oauthlib.flow import InstalledAppFlow
import pickle
flow = InstalledAppFlow.from_client_secrets_file(
    'credentials.json',
    ['https://www.googleapis.com/auth/drive.file'])
creds = flow.run_local_server(port=0)
with open('gdrive_token.pickle', 'wb') as f:
    pickle.dump(creds, f)
print('✅ 토큰 생성 완료')
"
```

2. GCP로 토큰 전송 (브라우저 SSH 업로드):
   - `gdrive_token.pickle` 파일을 GCP `/home/skyto5339/97layerOS/`에 업로드

3. GCP에서 자동 동기화 실행:
```bash
cd ~/97layerOS
source .venv/bin/activate
pip install -q google-api-python-client google-auth-httplib2 google-auth-oauthlib
python execution/ops/sync_gcp_to_gdrive_direct.py
```

---

## 📋 최종 체크리스트

- [ ] GCP SSH 접속
- [ ] 빠른 배포 스크립트 실행 또는 수동 업로드
- [ ] Technical Daemon 실행 확인
- [ ] Telegram Daemon 실행 확인
- [ ] Telegram `/status` 테스트 성공
- [ ] (선택) GCP 동기화 스크립트 설정
- [ ] (선택) Cron 자동 동기화 등록

---

## 🎉 완료 후 상태

```
Mac ←→ Google Drive ✅ (5분 자동)
       ↕
     GCP ⚡ (수동 또는 5분 자동)
```

**핵심:**
- Mac과 Google Drive는 완전 자동 동기화 ✅
- GCP는 24/7 운영 중 ✅
- GCP → Mac: 필요시 수동 또는 5분 자동 (선택)

---

**지금 실행:**
1. GCP Console → SSH 접속
2. 위의 "빠른 배포" 명령어 복사-붙여넣기
3. 완료!
