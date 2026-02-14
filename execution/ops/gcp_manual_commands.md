# GCP 수동 명령어 가이드
배포 스크립트 대신 수동으로 설정하려면 이 가이드를 따르세요.

---

## 1. 한 번만 실행 (초기 설정)

### SSH 접속
```bash
ssh username@gcp-instance-ip
cd ~/97layerOS
```

### systemd 서비스 생성

#### Master Controller 서비스
```bash
sudo nano /etc/systemd/system/97layer-master.service
```

내용:
```ini
[Unit]
Description=97layerOS Master Controller
After=network.target

[Service]
Type=simple
User=username
WorkingDirectory=/home/username/97layerOS
ExecStart=/usr/bin/python3 /home/username/97layerOS/execution/ops/master_controller.py start_all
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

#### Cycle Manager 서비스
```bash
sudo nano /etc/systemd/system/97layer-cycle.service
```

내용:
```ini
[Unit]
Description=97layerOS Cycle Manager
After=network.target

[Service]
Type=simple
User=username
WorkingDirectory=/home/username/97layerOS
ExecStart=/usr/bin/python3 /home/username/97layerOS/execution/cycle_manager.py
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### 서비스 활성화 및 시작
```bash
# systemd 리로드
sudo systemctl daemon-reload

# 자동 시작 활성화
sudo systemctl enable 97layer-master.service
sudo systemctl enable 97layer-cycle.service

# 서비스 시작
sudo systemctl start 97layer-master.service
sudo systemctl start 97layer-cycle.service
```

---

## 2. 상태 확인 (언제든지)

### 서비스 상태
```bash
# Master Controller
sudo systemctl status 97layer-master

# Cycle Manager
sudo systemctl status 97layer-cycle
```

### 로그 확인
```bash
# 실시간 로그
sudo journalctl -u 97layer-master -f
sudo journalctl -u 97layer-cycle -f

# 최근 로그 (100줄)
sudo journalctl -u 97layer-master -n 100
```

### 프로세스 확인
```bash
ps aux | grep -E "(telegram|junction|cycle)" | grep -v grep
```

---

## 3. 서비스 제어

### 재시작
```bash
sudo systemctl restart 97layer-master
sudo systemctl restart 97layer-cycle
```

### 중지
```bash
sudo systemctl stop 97layer-master
sudo systemctl stop 97layer-cycle
```

### 자동 시작 비활성화
```bash
sudo systemctl disable 97layer-master
sudo systemctl disable 97layer-cycle
```

---

## 4. 코드 업데이트 (맥북에서)

### 파일 동기화
```bash
# 맥북에서
cd /Users/97layer/97layerOS
rsync -avz --exclude '.git' --exclude '__pycache__' \
    ./ username@gcp-ip:~/97layerOS/

# GCP에서 서비스 재시작
ssh username@gcp-ip
sudo systemctl restart 97layer-master
sudo systemctl restart 97layer-cycle
```

---

## 5. 트러블슈팅

### 서비스가 시작 안 됨
```bash
# 에러 로그 확인
sudo journalctl -u 97layer-master -n 50 --no-pager

# Python 경로 확인
which python3

# 권한 확인
ls -la ~/97layerOS/execution/
```

### 의존성 에러
```bash
cd ~/97layerOS
pip3 install -r requirements.txt

# 또는 수동 설치
pip3 install asyncio aiohttp python-telegram-bot \
    google-generativeai anthropic schedule psutil python-dotenv
```

### 메모리 부족
```bash
# 메모리 사용량 확인
free -h

# 프로세스별 메모리
ps aux --sort=-%mem | head -10

# 서비스 재시작
sudo systemctl restart 97layer-master
```

---

## 6. 완전 자동화 확인

### 재부팅 테스트
```bash
# GCP 인스턴스 재부팅
sudo reboot

# 2분 후 접속
ssh username@gcp-ip

# 서비스 자동 시작 확인
sudo systemctl status 97layer-master
sudo systemctl status 97layer-cycle

# 프로세스 확인
ps aux | grep telegram
```

### 예상 결과
```
✅ 97layer-master.service - active (running)
✅ 97layer-cycle.service - active (running)
✅ async_telegram_daemon.py 실행 중
✅ cycle_manager.py 실행 중
```

---

## 7. 맥북에서 원격 제어

### GCP Management API
```bash
# 상태 확인
curl http://gcp-ip:8888/status

# Async Telegram 재시작
curl -X POST http://gcp-ip:8888/restart_async
```

---

**완료 후**: 텔레그램 메시지 전송 → 자동 처리 확인
