# rclone 빠른 설정 가이드 (5분 완료)

## ✅ rclone 설치 완료
rclone이 이미 설치되었습니다: `~/bin/rclone`

---

## 🔐 Google Drive 인증 (Mac에서 실행)

터미널에서 다음 명령어 실행:

```bash
~/bin/rclone config
```

### 대화형 설정 과정:

**1. New remote 생성**
```
No remotes found, make a new one?
n/s/q> n
```

**2. Remote 이름 입력**
```
name> gdrive
```

**3. Storage 타입 선택**
```
Type of storage to configure.
Choose a number from below, or type in your own value.
...
XX / Google Drive
   \ (drive)
...
Storage> drive
```

**4. Client ID (기본값 사용)**
```
client_id> (Enter - 그냥 엔터)
```

**5. Client Secret (기본값 사용)**
```
client_secret> (Enter - 그냥 엔터)
```

**6. Scope 선택 (Full access)**
```
scope> 1
```

**7. Root folder ID (기본값)**
```
root_folder_id> (Enter)
```

**8. Service account (사용 안 함)**
```
service_account_file> (Enter)
```

**9. Advanced config (사용 안 함)**
```
Edit advanced config?
y/n> n
```

**10. Auto config (브라우저 인증 사용)**
```
Use auto config?
 * Say Y if not sure
 * Say N if you are working on a remote or headless machine

y/n> y
```

→ **브라우저가 자동으로 열립니다**
→ **Google 계정 로그인**
→ **rclone 권한 승인**
→ **"Success! All done." 메시지 확인**

**11. Shared Drive (사용 안 함)**
```
Configure this as a Shared Drive (Team Drive)?
y/n> n
```

**12. 설정 확인**
```
y/e/d> y
```

**13. 종료**
```
e/n/d/r/c/s/q> q
```

---

## ✅ 인증 확인

```bash
~/bin/rclone lsd gdrive:
```

성공 시 Google Drive의 폴더 목록이 보입니다.

---

## 🚀 GCP로 설정 파일 전송

```bash
# Mac의 rclone 설정을 GCP로 복사
scp -i ~/.ssh/id_ed25519_gcp ~/.config/rclone/rclone.conf skyto5339@35.184.30.182:/tmp/
```

---

## 🖥️ GCP에서 설정 (브라우저 SSH)

GCP 브라우저 SSH에서:

```bash
# rclone 설치
curl https://rclone.org/install.sh | sudo bash

# 설정 디렉토리 생성
mkdir -p ~/.config/rclone

# Mac에서 전송한 설정 파일 이동
mv /tmp/rclone.conf ~/.config/rclone/
chmod 600 ~/.config/rclone/rclone.conf

# 테스트
rclone lsd gdrive:

# 97layerOS 폴더 확인 및 생성
rclone lsd gdrive: | grep 97layerOS || rclone mkdir gdrive:97layerOS

# 하위 폴더 생성
rclone mkdir gdrive:97layerOS/knowledge
rclone mkdir gdrive:97layerOS/directives
rclone mkdir gdrive:97layerOS/execution
rclone mkdir gdrive:97layerOS/libs

# 동기화 스크립트 실행 권한
cd ~/97layerOS
chmod +x execution/ops/sync_from_gcp_to_gdrive.sh

# 첫 동기화 테스트
./execution/ops/sync_from_gcp_to_gdrive.sh
```

---

## ⏰ 자동 동기화 설정

### GCP: Crontab (5분마다)

```bash
crontab -e
```

추가:
```
*/5 * * * * /home/skyto5339/97layerOS/execution/ops/sync_from_gcp_to_gdrive.sh >> /tmp/gdrive_sync.log 2>&1
```

저장 후:
```bash
crontab -l  # 확인
```

### Mac: LaunchAgent (5분마다)

```bash
# 실행 권한
chmod +x /Users/97layer/97layerOS/execution/ops/sync_from_gdrive_to_mac.sh

# LaunchAgent 생성
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

# 활성화
launchctl load ~/Library/LaunchAgents/com.97layer.gdrive-to-mac-sync.plist

# 확인
launchctl list | grep gdrive
```

---

## 🧪 동기화 테스트

### GCP → Mac 테스트

GCP에서:
```bash
echo "Test from GCP at $(date)" > ~/97layerOS/knowledge/test_gcp_sync.txt
./execution/ops/sync_from_gcp_to_gdrive.sh
```

5-10분 후 Mac에서:
```bash
cat /Users/97layer/97layerOS/knowledge/test_gcp_sync.txt
```

### Mac → GCP 테스트

Mac에서:
```bash
echo "Test from Mac at $(date)" > /Users/97layer/97layerOS/knowledge/test_mac_sync.txt
python3 execution/ops/sync_to_gdrive.py
```

5분 후 Google Drive 확인:
```bash
~/bin/rclone cat gdrive:97layerOS/knowledge/test_mac_sync.txt
```

---

## 📊 로그 모니터링

**Mac:**
```bash
tail -f /tmp/gdrive_to_mac_sync.log
```

**GCP:**
```bash
tail -f /tmp/gdrive_sync.log
```

---

## 완료 체크리스트

- [ ] Mac: `~/bin/rclone config` 완료
- [ ] Mac: `~/bin/rclone lsd gdrive:` 성공
- [ ] GCP: rclone 설치 완료
- [ ] GCP: rclone.conf 전송 완료
- [ ] GCP: `rclone lsd gdrive:` 성공
- [ ] GCP: 동기화 스크립트 테스트 성공
- [ ] GCP: crontab 등록 완료
- [ ] Mac: LaunchAgent 등록 완료
- [ ] GCP → Mac 테스트 파일 확인
- [ ] Mac → GCP 테스트 파일 확인

---

## 🎯 예상 소요 시간

- Mac rclone 인증: 2분
- GCP 설정: 2분
- 자동화 설정: 1분
- **총 5분**
