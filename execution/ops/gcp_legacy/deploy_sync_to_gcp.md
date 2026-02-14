# GCP 동기화 시스템 배포 가이드

## 📦 준비 완료
파일: `/tmp/97layerOS_sync_update.tar.gz` (5.1KB)

---

## 🚀 GCP 브라우저 SSH에서 실행

### 1단계: 파일 업로드
GCP 브라우저 SSH에서 "톱니바퀴 아이콘" → "Upload file" 클릭
→ `/tmp/97layerOS_sync_update.tar.gz` 선택

또는 Mac 터미널에서 scp:
```bash
scp -i ~/.ssh/id_ed25519_gcp /tmp/97layerOS_sync_update.tar.gz skyto5339@35.184.30.182:/tmp/
```

### 2단계: GCP에서 압축 해제 및 설정

GCP 브라우저 SSH에서:

```bash
cd ~
tar xzf /tmp/97layerOS_sync_update.tar.gz -C ~/
cd ~/97layerOS

# Python 패키지 설치
source .venv/bin/activate
pip install -q google-api-python-client google-auth-httplib2 google-auth-oauthlib

# 실행 권한 부여
chmod +x execution/ops/sync_from_gcp_to_gdrive.py

# 테스트 (첫 실행 시 브라우저 인증 필요 - 실패 가능)
python execution/ops/sync_from_gcp_to_gdrive.py
```

### 3단계: Cron 설정 (5분마다 자동 동기화)

```bash
crontab -e
```

추가:
```
*/5 * * * * cd /home/skyto5339/97layerOS && /home/skyto5339/97layerOS/.venv/bin/python /home/skyto5339/97layerOS/execution/ops/sync_from_gcp_to_gdrive.py >> /tmp/gcp_gdrive_sync.log 2>&1
```

저장 후:
```bash
crontab -l  # 확인
```

---

## ⚠️ 중요: Google Drive API 인증 문제

GCP는 헤드리스 서버라 브라우저 인증이 어렵습니다.

**해결 방법:**

### 방법 1: Mac에서 인증 후 토큰 전송 (권장)

Mac에서:
```bash
cd /Users/97layer/97layerOS
python3 execution/ops/sync_from_gcp_to_gdrive.py
# 브라우저 인증 완료 → gdrive_token.pickle 생성됨

# GCP로 전송
scp -i ~/.ssh/id_ed25519_gcp gdrive_token.pickle skyto5339@35.184.30.182:~/97layerOS/
```

GCP에서:
```bash
cd ~/97layerOS
python execution/ops/sync_from_gcp_to_gdrive.py
# 이제 토큰이 있어서 바로 동기화됨
```

### 방법 2: rclone 사용 (대안)

bash 스크립트 사용:
```bash
cd ~/97layerOS
./execution/ops/sync_from_gcp_to_gdrive.sh
```

---

## 📊 로그 확인

```bash
tail -f /tmp/gcp_gdrive_sync.log
```

---

## ✅ 완료 체크리스트

- [ ] sync_update 패키지 GCP 업로드
- [ ] Python 패키지 설치 (google-api-python-client 등)
- [ ] Mac에서 gdrive_token.pickle 생성 및 전송
- [ ] GCP에서 동기화 스크립트 테스트 성공
- [ ] Crontab 등록 완료
- [ ] 로그에서 정상 동작 확인
