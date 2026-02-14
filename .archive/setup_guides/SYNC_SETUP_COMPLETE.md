# 양방향 동기화 시스템 설정 완료 🔄

## ✅ 완료된 작업

### 1. Mac → Google Drive 동기화
- **스크립트**: [execution/ops/sync_to_gdrive.py](execution/ops/sync_to_gdrive.py)
- **경로 수정**: `~/내 드라이브(skyto5339@gmail.com)/97layerOS`
- **테스트 결과**: ✅ 7개 성공
- **자동화**: 기존 시스템에서 이미 실행 중

### 2. Google Drive → Mac 동기화
- **스크립트**: [execution/ops/sync_from_gdrive_to_mac.sh](execution/ops/sync_from_gdrive_to_mac.sh)
- **LaunchAgent**: `com.97layer.gdrive-to-mac-sync` (PID: 25628)
- **주기**: 5분마다 자동 실행
- **테스트 결과**: ✅ 정상 작동
- **로그**: `/tmp/gdrive_to_mac_sync.log`

### 3. GCP → Google Drive 동기화 준비
- **Python 스크립트**: [execution/ops/sync_from_gcp_to_gdrive.py](execution/ops/sync_from_gcp_to_gdrive.py)
- **배포 패키지**: `/tmp/97layerOS_sync_update.tar.gz` (5.1KB)
- **가이드**: [deploy_sync_to_gcp.md](deploy_sync_to_gcp.md)

---

## 🎯 현재 상태

### 작동 중:
- ✅ Mac → Google Drive (자동, 5분마다)
- ✅ Google Drive → Mac (LaunchAgent, 5분마다)

### 대기 중:
- ⏳ GCP → Google Drive (배포 필요)

---

## 📦 GCP 배포 방법 (간단 요약)

### 방법 1: SCP로 전송 (터미널)

```bash
scp -i ~/.ssh/id_ed25519_gcp /tmp/97layerOS_sync_update.tar.gz skyto5339@35.184.30.182:/tmp/
```

### 방법 2: GCP 브라우저 SSH 업로드

1. GCP Console → SSH 버튼
2. 톱니바퀴 아이콘 → "Upload file"
3. `/tmp/97layerOS_sync_update.tar.gz` 선택

### 방법 3: 제가 명령어 실행 (자동화)

지금 바로 제가 scp로 전송할 수 있습니다. 원하시면 "전송해줘"라고 말씀해주세요.

---

## 🚀 GCP 설치 명령어 (전송 후)

GCP 브라우저 SSH에서:

```bash
cd ~
tar xzf /tmp/97layerOS_sync_update.tar.gz -C ~/
cd ~/97layerOS
source .venv/bin/activate
pip install -q google-api-python-client google-auth-httplib2 google-auth-oauthlib
chmod +x execution/ops/sync_from_gcp_to_gdrive.py
```

Mac에서 인증 토큰 생성 후 전송:
```bash
cd /Users/97layer/97layerOS
python3 execution/ops/sync_from_gcp_to_gdrive.py
# 브라우저 인증 완료 후
scp -i ~/.ssh/id_ed25519_gcp gdrive_token.pickle skyto5339@35.184.30.182:~/97layerOS/
```

GCP에서 Cron 등록:
```bash
crontab -e
# 추가:
*/5 * * * * cd /home/skyto5339/97layerOS && /home/skyto5339/97layerOS/.venv/bin/python /home/skyto5339/97layerOS/execution/ops/sync_from_gcp_to_gdrive.py >> /tmp/gcp_gdrive_sync.log 2>&1
```

---

## 🧪 동기화 테스트 확인

### Mac에서 확인:
```bash
# 로그 확인
tail -f /tmp/gdrive_to_mac_sync.log

# chat_memory 확인 (GCP 대화가 보여야 함)
tail -50 knowledge/chat_memory/7565534667.json
```

### GCP에서 확인:
```bash
tail -f /tmp/gcp_gdrive_sync.log
```

---

## 📊 동기화 흐름

```
┌─────────────┐
│  Mac Local  │
│             │
└──────┬──────┘
       │
       │ sync_to_gdrive.py (5분)
       │ sync_from_gdrive_to_mac.sh (5분)
       ▼
┌──────────────────────────┐
│   Google Drive           │
│   ~/내 드라이브(...)/     │
│                          │
│   📁 97layerOS/          │◀──────┐
│      ├─ knowledge/       │       │
│      ├─ directives/      │       │
│      ├─ execution/       │       │
│      └─ libs/            │       │
└──────────────────────────┘       │
                                   │
                    sync_from_gcp_to_gdrive.py (5분)
                                   │
                           ┌───────┴────────┐
                           │   GCP Server   │
                           │  35.184.30.182 │
                           │                │
                           │  Technical ✅  │
                           │  Telegram ✅   │
                           └────────────────┘
```

---

## 🔍 문제 해결

### WOOHWAHAE 대화가 Mac에 안 보임
**원인**: GCP → Google Drive 동기화가 아직 설정 안 됨
**해결**: 위의 GCP 배포 방법 실행

### Google Drive 경로 오류
**원인**: Google Drive Desktop 경로가 변경됨
**해결**: 이미 수정 완료 (`~/내 드라이브(skyto5339@gmail.com)/`)

### LaunchAgent 작동 안 함
**확인**:
```bash
launchctl list | grep gdrive
tail -f /tmp/gdrive_to_mac_sync_error.log
```

---

## ✅ 완료 체크리스트

- [x] rclone 설치 시도 (Google Drive Desktop으로 대체)
- [x] Google Drive 경로 확인 및 수정
- [x] Mac → Google Drive 동기화 테스트
- [x] Google Drive → Mac LaunchAgent 설정
- [x] GCP 동기화 스크립트 작성
- [x] GCP 배포 패키지 생성
- [ ] GCP에 패키지 전송
- [ ] GCP에서 Google API 패키지 설치
- [ ] Mac에서 인증 토큰 생성 및 전송
- [ ] GCP Cron 등록
- [ ] 양방향 동기화 최종 테스트

---

## 🎉 결과

**Mac → Google Drive → Mac 순환 동기화 작동 중!**

GCP만 연결하면 완전한 3-way 동기화 시스템 완성됩니다.

---

**다음 명령:**
- "전송해줘" - 제가 GCP에 파일 전송해드립니다
- "GCP 설정 도와줘" - GCP 설정 단계 안내
- "테스트해줘" - 현재 동기화 상태 테스트
