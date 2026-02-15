# 🚀 GCP VM 배포 명령어 (복사-붙여넣기용)

> **중요**: 각 단계별로 명령어를 **순서대로** 복사해서 VM에 붙여넣으세요.
>
> **GCP Console 접속**: https://console.cloud.google.com/compute/instances

---

## 📝 사전 준비

### 1. 로컬 컴퓨터에서 배포 파일 확인
```bash
ls -lh ~/97layer-deploy.tar.gz
# 결과: 169KB 파일이 있어야 함
```

---

## 🔧 GCP VM에서 실행할 명령어

### **Step 1: 프로젝트 디렉토리 생성**
```bash
cd ~
mkdir -p 97layerOS
cd 97layerOS
pwd
```
**예상 출력**: `/home/YOUR_USERNAME/97layerOS`

---

### **Step 2: 배포 파일 업로드 대기**

**⚠️ 여기서 멈춤! 다음 작업 필요:**

1. **GCP Console SSH 창에서**:
   - 우측 상단 ⚙️ (설정) 버튼 클릭
   - "Upload file" 선택
   - `~/97layer-deploy.tar.gz` 파일 선택
   - 업로드 완료 대기

2. **업로드 확인**:
```bash
ls -lh ~/97layer-deploy.tar.gz
```
**예상 출력**: `169K ... 97layer-deploy.tar.gz`

---

### **Step 3: 압축 해제**
```bash
cd ~/97layerOS
tar -xzf ~/97layer-deploy.tar.gz
ls -la
```
**예상 출력**: `core/`, `directives/`, `knowledge/`, `requirements.txt` 등

---

### **Step 4: Python 3.11 설치**
```bash
sudo apt update && sudo apt install -y python3.11 python3.11-venv python3-pip git
python3.11 --version
```
**예상 출력**: `Python 3.11.x`
**소요 시간**: 2-3분

---

### **Step 5: Virtual Environment 생성**
```bash
cd ~/97layerOS
python3.11 -m venv .venv
source .venv/bin/activate
which python3
```
**예상 출력**: `/home/YOUR_USERNAME/97layerOS/.venv/bin/python3`

---

### **Step 6: Dependencies 설치**
```bash
pip install --upgrade pip
pip install -r requirements.txt
```
**소요 시간**: 5-10분 (인내심 필요!)
**예상 출력**: 많은 패키지 설치 메시지

---

### **Step 7: Import 테스트**
```bash
python3 -c "from core.daemons.telegram_secretary import TelegramSecretary; print('✅ Import OK')"
```
**예상 출력**: `✅ Import OK`
**❌ 에러 나면**: 저에게 에러 메시지 복사해서 보내주세요

---

### **Step 8: .env 파일 생성**
```bash
cd ~/97layerOS
cat > .env << 'EOFENV'
TELEGRAM_BOT_TOKEN=8501568801:AAE-3fBl-p6uZcmrdsWSRQuz_eg8yDADwjI
GOOGLE_API_KEY=AIzaSyCGgHVPjEEI3OI3tSNW3SSHNbZuYpHrH-g
GEMINI_API_KEY=AIzaSyCGgHVPjEEI3OI3tSNW3SSHNbZuYpHrH-g
ANTHROPIC_API_KEY=sk-ant-api03-RfIvjE0-M0iN_3f76vY6S9_Fm2p6X6y5X9_Fm2p6X6y5X9_Fm2p6X6y5X9_Fm2p6X6y5UQAA
TZ=Asia/Seoul
EOFENV

chmod 600 .env
cat .env
```
**예상 출력**: API keys가 표시됨

---

### **Step 9: 로그 디렉토리 생성**
```bash
mkdir -p ~/97layerOS/logs
ls -la logs/
```

---

### **Step 10: Foreground 테스트 (중요!)**
```bash
cd ~/97layerOS
source .venv/bin/activate
python3 core/daemons/telegram_secretary.py
```

**예상 출력**:
```
✅ TelegramSecretary initialized
🤖 Bot started, waiting for messages...
```

**이제 Telegram에서 테스트**:
1. 봇에게 `/status` 메시지 전송
2. 응답이 오는지 확인

**✅ 응답 받으면**: `Ctrl+C` 눌러서 종료
**❌ 응답 없으면**: 저에게 로그 복사해서 보내주세요

---

### **Step 11: Systemd Service 설정**
```bash
cd ~/97layerOS

# Service 파일 준비
sed "s/USERNAME_PLACEHOLDER/$(whoami)/g" deployment/97layer-telegram.service > /tmp/97layer-telegram.service

# Service 설치
sudo mv /tmp/97layer-telegram.service /etc/systemd/system/97layer-telegram.service

# 확인
cat /etc/systemd/system/97layer-telegram.service | head -15
```

**예상 출력**: Service 파일 내용이 표시됨

---

### **Step 12: Service 시작**
```bash
sudo systemctl daemon-reload
sudo systemctl enable 97layer-telegram
sudo systemctl start 97layer-telegram
```

---

### **Step 13: Service 상태 확인**
```bash
sudo systemctl status 97layer-telegram
```

**예상 출력**:
```
● 97layer-telegram.service - 97layerOS Telegram Executive Secretary
   Loaded: loaded (/etc/systemd/system/97layer-telegram.service; enabled)
   Active: active (running) since ...
```

**✅ "active (running)" 보이면 성공!**
**❌ "failed" 보이면**: 저에게 알려주세요

---

### **Step 14: 실시간 로그 확인**
```bash
journalctl -u 97layer-telegram -f
```

**예상 출력**: Bot 로그가 실시간으로 표시됨

**종료**: `Ctrl+C`

---

### **Step 15: Telegram에서 최종 테스트**

Telegram 봇에게 다음 명령어 테스트:
- `/status` - 시스템 상태
- `/help` - 도움말
- 아무 텍스트 - 자동 신호 포착

**모두 응답하면 ✅ 배포 완료!**

---

## 🎉 배포 완료!

### 확인 사항
- [x] VM에 코드 배포됨
- [x] Python 환경 설정됨
- [x] Foreground 테스트 통과
- [x] Systemd service 실행 중
- [x] Telegram bot 응답함

### 관리 명령어

**Service 관리**:
```bash
# 재시작
sudo systemctl restart 97layer-telegram

# 중지
sudo systemctl stop 97layer-telegram

# 시작
sudo systemctl start 97layer-telegram

# 상태 확인
sudo systemctl status 97layer-telegram
```

**로그 확인**:
```bash
# 실시간
journalctl -u 97layer-telegram -f

# 마지막 50줄
journalctl -u 97layer-telegram -n 50

# 오늘 로그
journalctl -u 97layer-telegram --since today
```

**메모리 확인**:
```bash
free -h
ps aux | grep telegram_secretary
```

---

## 🚨 트러블슈팅

### Bot이 응답 안 함

**1. Service 상태 확인**:
```bash
sudo systemctl status 97layer-telegram
```

**2. 로그 확인**:
```bash
journalctl -u 97layer-telegram -n 100
```

**3. 수동 실행 테스트**:
```bash
cd ~/97layerOS
source .venv/bin/activate
python3 core/daemons/telegram_secretary.py
```

### 메모리 부족
```bash
free -h
# Available이 100MB 이하면 문제
sudo systemctl restart 97layer-telegram
```

### Import 에러
```bash
cd ~/97layerOS
source .venv/bin/activate
pip install -r requirements.txt
```

---

## 📊 성공 기준

✅ **확인 완료**:
- Telegram bot이 `/status` 명령에 응답함
- Service가 `active (running)` 상태
- 메모리 사용 < 200MB
- 로그에 critical error 없음

🎯 **다음 단계**: Multi-agent 통합!

---

> **문제 발생 시**: 해당 단계의 출력을 복사해서 저에게 보내주세요. 바로 해결해드리겠습니다!
