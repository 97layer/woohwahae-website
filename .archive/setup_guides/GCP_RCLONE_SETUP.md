# GCP rclone 설정 가이드

## 목적
GCP 서버에서 Google Drive로 자동 동기화하여 Mac과 실시간 데이터 공유

## 1단계: GCP에 rclone 설치

GCP 브라우저 SSH에서:

```bash
# rclone 설치
curl https://rclone.org/install.sh | sudo bash

# 설치 확인
rclone version
```

## 2단계: Google Drive OAuth 인증 설정

**방법 A: 로컬 Mac에서 인증 후 토큰 복사 (권장)**

Mac에서:
```bash
# Mac에 rclone 설치 (Homebrew)
brew install rclone

# Google Drive 설정 시작
rclone config

# 설정 과정:
# n) New remote
# name> gdrive
# Storage> drive (Google Drive 선택)
# client_id> (Enter - 기본값)
# client_secret> (Enter - 기본값)
# scope> 1 (Full access)
# root_folder_id> (Enter)
# service_account_file> (Enter)
# Edit advanced config? n
# Use auto config? y (브라우저가 열리고 Google 로그인)
# Configure this as a Shared Drive? n
# Yes this is OK

# 인증 완료 후 설정 파일 확인
cat ~/.config/rclone/rclone.conf
```

생성된 `rclone.conf` 내용을 복사하여 GCP에 전송:

```bash
# Mac에서 GCP로 전송
scp -i ~/.ssh/id_ed25519_gcp ~/.config/rclone/rclone.conf skyto5339@35.184.30.182:/tmp/
```

GCP에서:
```bash
# rclone 설정 디렉토리 생성
mkdir -p ~/.config/rclone

# 설정 파일 이동
mv /tmp/rclone.conf ~/.config/rclone/

# 권한 설정
chmod 600 ~/.config/rclone/rclone.conf

# 테스트
rclone lsd gdrive:
```

**방법 B: GCP에서 직접 인증 (SSH 포트포워딩 필요)**

```bash
# Mac에서 SSH 포트포워딩
ssh -i ~/.ssh/id_ed25519_gcp -L 53682:localhost:53682 skyto5339@35.184.30.182

# GCP SSH 세션에서
rclone config
# ... 위와 동일한 설정 과정
# Use auto config? y 선택하면 Mac의 localhost:53682로 인증 페이지 열림
```

## 3단계: Google Drive에 97layerOS 폴더 확인

GCP에서:
```bash
# 97layerOS 폴더 확인
rclone lsd gdrive:

# 97layerOS 폴더가 없으면 생성
rclone mkdir gdrive:97layerOS

# 하위 폴더 생성
rclone mkdir gdrive:97layerOS/knowledge
rclone mkdir gdrive:97layerOS/directives
rclone mkdir gdrive:97layerOS/execution
rclone mkdir gdrive:97layerOS/libs
```

## 4단계: 동기화 스크립트 배포

GCP에서:
```bash
cd ~/97layerOS

# 실행 권한 부여
chmod +x execution/ops/sync_from_gcp_to_gdrive.sh

# 수동 테스트
./execution/ops/sync_from_gcp_to_gdrive.sh
```

예상 출력:
```
[2026-02-14 09:30:00] 🔄 GCP → Google Drive 동기화 시작...
[2026-02-14 09:30:05] ✅ 동기화 완료
```

## 5단계: Cron 자동 실행 설정

GCP에서:
```bash
# crontab 편집
crontab -e

# 5분마다 실행 추가
*/5 * * * * /home/skyto5339/97layerOS/execution/ops/sync_from_gcp_to_gdrive.sh >> /tmp/gdrive_sync.log 2>&1
```

저장 후 cron 확인:
```bash
crontab -l
```

## 6단계: Mac에서 Google Drive → Local 동기화 설정

Mac에서:
```bash
cd /Users/97layer/97layerOS

# 실행 권한 부여
chmod +x execution/ops/sync_from_gdrive_to_mac.sh

# 수동 테스트
./execution/ops/sync_from_gdrive_to_mac.sh
```

LaunchAgent 설정 (5분마다 자동 실행):

```bash
cat > ~/Library/LaunchAgents/com.97layer.gdrive-to-mac-sync.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.97layer.gdrive-to-mac-sync</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/97layer/97layerOS/execution/ops/sync_from_gdrive_to_mac.sh</string>
    </array>
    <key>StartInterval</key>
    <integer>300</integer>
    <key>StandardOutPath</key>
    <string>/tmp/gdrive_to_mac_sync.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/gdrive_to_mac_sync_error.log</string>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
EOF

# LaunchAgent 활성화
launchctl load ~/Library/LaunchAgents/com.97layer.gdrive-to-mac-sync.plist

# 상태 확인
launchctl list | grep gdrive-to-mac-sync
```

## 7단계: 양방향 동기화 검증

**테스트 시나리오 1: GCP → Mac**

GCP에서 테스트 파일 생성:
```bash
echo "Test from GCP at $(date)" > ~/97layerOS/knowledge/test_gcp.txt
./execution/ops/sync_from_gcp_to_gdrive.sh
```

5분 후 Mac에서 확인:
```bash
cat /Users/97layer/97layerOS/knowledge/test_gcp.txt
```

**테스트 시나리오 2: Mac → GCP**

Mac에서 테스트 파일 생성:
```bash
echo "Test from Mac at $(date)" > /Users/97layer/97layerOS/knowledge/test_mac.txt
python3 execution/ops/sync_to_gdrive.py
```

Google Drive 확인:
```bash
# Mac 또는 GCP에서
rclone cat gdrive:97layerOS/knowledge/test_mac.txt
```

GCP에서 동기화 후 확인:
```bash
# GCP: Google Drive → Local pull 필요
rclone sync gdrive:97layerOS/knowledge/ ~/97layerOS/knowledge/ --exclude ".DS_Store"
cat ~/97layerOS/knowledge/test_mac.txt
```

## 동기화 아키텍처

```
┌──────────────┐
│   Mac Local  │
│              │
│  - Technical │
│  - Telegram  │
│    (Stopped) │
└──────┬───────┘
       │
       │ sync_to_gdrive.py (5분)
       │ sync_from_gdrive_to_mac.sh (5분)
       ▼
┌──────────────────────┐
│   Google Drive       │
│   (Sync Hub)         │
│                      │
│  📁 97layerOS/       │
│    ├─ knowledge/     │◀─────┐
│    ├─ directives/    │      │
│    ├─ execution/     │      │
│    └─ libs/          │      │
└──────────────────────┘      │
                              │
                              │ rclone sync (5분)
                              │
                        ┌─────┴─────┐
                        │ GCP Server│
                        │           │
                        │ Technical │
                        │ Telegram  │
                        │ (Primary) │
                        └───────────┘
```

## 로그 확인

**GCP:**
```bash
# 동기화 로그
tail -f /tmp/gdrive_sync.log

# rclone 수동 테스트
rclone ls gdrive:97layerOS/knowledge/ | head -10
```

**Mac:**
```bash
# Mac → GDrive 로그
tail -f /tmp/sync_to_gdrive.log

# GDrive → Mac 로그
tail -f /tmp/gdrive_to_mac_sync.log
```

## 문제 해결

### rclone 인증 실패
```bash
# 설정 확인
cat ~/.config/rclone/rclone.conf

# 재인증
rclone config reconnect gdrive:
```

### 동기화 충돌
```bash
# Google Drive 상태 확인
rclone check ~/97layerOS/knowledge/ gdrive:97layerOS/knowledge/

# 수동 양방향 동기화
rclone sync ~/97layerOS/knowledge/ gdrive:97layerOS/knowledge/ --interactive
```

### 퍼미션 에러
```bash
# 스크립트 권한 확인
ls -l execution/ops/*.sh

# 실행 권한 부여
chmod +x execution/ops/*.sh
```

## 참고사항

- **동기화 주기**: 5분 (필요시 crontab/LaunchAgent에서 조정)
- **충돌 해결**: rclone은 최신 파일로 덮어쓰기 (timestamp 기준)
- **대역폭**: GCP Free Tier 1GB/month 고려 (현재 ~50MB 사용)
- **보안**: `.env` 파일은 동기화하지 않음 (각 서버에 별도 관리)

## 완료 체크리스트

- [ ] rclone 설치 완료 (GCP)
- [ ] Google Drive OAuth 인증 완료
- [ ] `rclone lsd gdrive:` 성공
- [ ] 동기화 스크립트 실행 테스트 성공
- [ ] GCP crontab 등록 완료
- [ ] Mac LaunchAgent 등록 완료
- [ ] GCP → Mac 테스트 파일 전송 확인
- [ ] Mac → GCP 테스트 파일 전송 확인
- [ ] chat_memory 실시간 동기화 확인
