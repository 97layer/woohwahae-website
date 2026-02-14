# Systemd Services Installation Guide

## 목적 (Purpose)
GCP 서버가 재부팅되어도 자동으로 Technical Daemon과 Telegram Daemon이 실행되도록 systemd 서비스로 등록합니다.

## 설치 방법 (Installation Steps)

### 1단계: 서비스 파일 업로드
GCP 브라우저 SSH에서:

```bash
# 홈 디렉토리로 이동
cd ~/97layerOS

# 업로드 확인
ls -la | grep -E "97layer_.*\.service|install_systemd"
```

예상 출력:
```
-rw-r--r--  1 skyto5339 skyto5339   489 Feb 14 00:09 97layer_technical.service
-rw-r--r--  1 skyto5339 skyto5339   487 Feb 14 00:09 97layer_telegram.service
-rwxr-xr-x  1 skyto5339 skyto5339  1428 Feb 14 00:09 install_systemd_services.sh
```

### 2단계: 설치 스크립트 실행
```bash
chmod +x install_systemd_services.sh
./install_systemd_services.sh
```

### 3단계: 서비스 상태 확인
```bash
# Technical Daemon 상태
sudo systemctl status 97layer_technical.service

# Telegram Daemon 상태
sudo systemctl status 97layer_telegram.service

# 로그 실시간 확인
sudo journalctl -u 97layer_technical.service -f
sudo journalctl -u 97layer_telegram.service -f
```

## 서비스 관리 명령어 (Service Management Commands)

### 시작/중지/재시작
```bash
# 시작
sudo systemctl start 97layer_technical.service
sudo systemctl start 97layer_telegram.service

# 중지
sudo systemctl stop 97layer_technical.service
sudo systemctl stop 97layer_telegram.service

# 재시작
sudo systemctl restart 97layer_technical.service
sudo systemctl restart 97layer_telegram.service
```

### 자동 시작 설정
```bash
# 부팅 시 자동 시작 활성화
sudo systemctl enable 97layer_technical.service
sudo systemctl enable 97layer_telegram.service

# 자동 시작 비활성화
sudo systemctl disable 97layer_technical.service
sudo systemctl disable 97layer_telegram.service
```

### 로그 확인
```bash
# 전체 로그 확인
sudo journalctl -u 97layer_technical.service
sudo journalctl -u 97layer_telegram.service

# 최근 50줄
sudo journalctl -u 97layer_technical.service -n 50

# 실시간 로그 (Ctrl+C로 종료)
sudo journalctl -u 97layer_technical.service -f

# 오늘 로그만
sudo journalctl -u 97layer_technical.service --since today

# 특정 시간 이후
sudo journalctl -u 97layer_technical.service --since "2026-02-14 08:00:00"
```

## 재부팅 테스트 (Reboot Test)

```bash
# GCP 서버 재부팅
sudo reboot

# 재부팅 후 SSH 재접속하여 확인
ps aux | grep -E "technical_daemon|telegram_daemon" | grep -v grep
sudo systemctl status 97layer_technical.service
sudo systemctl status 97layer_telegram.service
```

## 문제 해결 (Troubleshooting)

### 서비스가 시작되지 않을 때
```bash
# 상세 상태 확인
sudo systemctl status 97layer_technical.service -l

# 로그에서 에러 찾기
sudo journalctl -u 97layer_technical.service --since "10 minutes ago" | grep -i error

# 설정 파일 검증
sudo systemd-analyze verify /etc/systemd/system/97layer_technical.service
```

### 환경 변수 문제
서비스 파일에 환경 변수가 정의되어 있지만, `.env` 파일이 필요한 경우:
```bash
# .env 파일 위치 확인
ls -la ~/97layerOS/.env

# 내용 확인 (민감 정보 주의)
cat ~/97layerOS/.env | grep -E "BOT_TOKEN|GEMINI"
```

### 권한 문제
```bash
# 서비스 파일 권한 확인
ls -l /etc/systemd/system/97layer_*.service

# 실행 파일 권한 확인
ls -l ~/97layerOS/execution/technical_daemon.py
ls -l ~/97layerOS/execution/telegram_daemon.py

# venv 활성화 테스트
source ~/97layerOS/.venv/bin/activate
python -c "import google.generativeai; print('OK')"
```

## 설치 후 체크리스트 (Post-Installation Checklist)

- [ ] `sudo systemctl status 97layer_technical.service` → active (running)
- [ ] `sudo systemctl status 97layer_telegram.service` → active (running)
- [ ] `ps aux | grep technical_daemon` → 프로세스 확인됨
- [ ] `ps aux | grep telegram_daemon` → 프로세스 확인됨
- [ ] `sudo journalctl -u 97layer_technical.service -n 20` → 에러 없음
- [ ] `sudo journalctl -u 97layer_telegram.service -n 20` → 에러 없음
- [ ] Telegram으로 `/status` 전송 → 응답 확인
- [ ] `sudo reboot` 후 재접속 → 서비스 자동 시작 확인

## 백업/복구 (Backup/Recovery)

### 서비스 파일 백업
```bash
sudo cp /etc/systemd/system/97layer_technical.service ~/97layer_technical.service.backup
sudo cp /etc/systemd/system/97layer_telegram.service ~/97layer_telegram.service.backup
```

### 서비스 완전 제거
```bash
# 서비스 중지 및 비활성화
sudo systemctl stop 97layer_technical.service 97layer_telegram.service
sudo systemctl disable 97layer_technical.service 97layer_telegram.service

# 서비스 파일 삭제
sudo rm /etc/systemd/system/97layer_technical.service
sudo rm /etc/systemd/system/97layer_telegram.service

# systemd 재로드
sudo systemctl daemon-reload
```

## 참고사항 (Notes)

- **Restart Policy**: 서비스가 실패하면 10초 후 자동 재시작됩니다 (`Restart=always`, `RestartSec=10`)
- **로그 저장**: 로그는 systemd journal에 저장되며, `/tmp/`의 nohup 로그는 더 이상 사용되지 않습니다
- **성능**: systemd는 nohup보다 안정적이고 리소스 효율적입니다
- **모니터링**: `systemctl status`로 언제든지 서비스 상태를 확인할 수 있습니다
