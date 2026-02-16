# 97layerOS 배포 가이드

> 통합 배포 문서 — Last Updated: 2026-02-16
> 기존 루트 배포 문서 8개 통합본

---

## 1. 환경 요구사항

| 환경 | 역할 | 사용자 |
|---|---|---|
| 로컬 MacBook | 코드 작성, Git, 배포 트리거 | 97layer |
| GCP VM (layer97-nightguard) | 24/7 Telegram Bot 실행 | skyto5339_gmail_com |
| Podman 컨테이너 | Python 실행, MCP CLI | - |

**GCP VM 접속 정보:**
- VM 이름: `layer97-nightguard`
- SSH alias: `97layer-vm`
- External IP: GCP Console → Compute Engine → VM instances에서 확인

---

## 2. 빠른 시작 (Quick Start)

```bash
# 로컬 MacBook에서 실행
cd /Users/97layer/97layerOS
./deploy.sh
```

**배포 흐름:**
```
로컬 MacBook (./deploy.sh)
  → SSH 접속 (자동)
  → 파일 업로드 (rsync)
  → GCP VM 압축 해제
  → systemd 재시작
  → Telegram Bot 재시작 완료
```

---

## 3. SSH 초기 설정 (최초 1회)

```bash
# ~/.ssh/config 설정
cat >> ~/.ssh/config << 'EOF'

Host 97layer-vm
    HostName YOUR_VM_EXTERNAL_IP
    User skyto5339_gmail_com
    IdentityFile ~/.ssh/google_compute_engine
    ServerAliveInterval 60
EOF

# 키 권한 설정
chmod 600 ~/.ssh/google_compute_engine

# 연결 테스트
ssh 97layer-vm "echo 'SSH OK'"
```

GCP SSH 키 없을 경우:
```bash
gcloud compute config-ssh
```

---

## 4. 수동 배포 (gcloud / SSH 없이)

### 4-1. 패키지 생성 (로컬)

```bash
tar -czf ~/97layer-manual-deploy.tar.gz \
    core/ directives/ knowledge/docs/ knowledge/agent_hub/ \
    requirements.txt start_telegram.sh start_monitor.sh

ls -lh ~/97layer-manual-deploy.tar.gz
```

### 4-2. 업로드 (GCP Console)

1. GCP Console → Compute Engine → VM instances
2. `layer97-nightguard` → SSH
3. SSH 창 우상단 톱니바퀴 → Upload file → `97layer-manual-deploy.tar.gz`

### 4-3. VM에서 적용

```bash
cd ~/97layerOS

# 백업
tar -czf ~/97layer-backup-$(date +%Y%m%d-%H%M%S).tar.gz core/ directives/ knowledge/ 2>/dev/null || true

# 압축 해제 및 재시작
tar -xzf ~/97layer-manual-deploy.tar.gz
sudo systemctl restart 97layer-telegram
sudo systemctl status 97layer-telegram --no-pager
```

---

## 5. CI/CD 자동화 (GitHub Actions)

위치: `.github/workflows/deploy.yml`

```yaml
name: Deploy to GCP VM

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Deploy to VM
        uses: appleboy/ssh-action@v0.1.7
        with:
          host: ${{ secrets.VM_EXTERNAL_IP }}
          username: skyto5339_gmail_com
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd ~/97layerOS
            git pull origin main
            sudo systemctl restart 97layer-telegram
```

GitHub Secrets 필요: `VM_EXTERNAL_IP`, `SSH_PRIVATE_KEY`

---

## 6. systemd 서비스 등록 (GCP VM)

### 6-1. 서비스 파일 복사 (USERNAME_PLACEHOLDER 치환)

```bash
# GCP VM에서 실행
GCP_USER=$(whoami)  # e.g. skyto5339_gmail_com

# 서비스 파일 복사 + 사용자명 치환
for f in telegram nightguard ecosystem; do
    sed "s/USERNAME_PLACEHOLDER/$GCP_USER/g" \
        ~/97layerOS/knowledge/docs/deployment/97layer-${f}.service \
        | sudo tee /etc/systemd/system/97layer-${f}.service > /dev/null
done

# systemd 리로드
sudo systemctl daemon-reload
```

### 6-2. Telegram 봇 서비스 활성화 (기존)

```bash
sudo systemctl enable 97layer-telegram
sudo systemctl start 97layer-telegram
sudo systemctl status 97layer-telegram --no-pager
```

### 6-3. Nightguard V2 서비스 활성화

```bash
sudo systemctl enable 97layer-nightguard
sudo systemctl start 97layer-nightguard
sudo systemctl status 97layer-nightguard --no-pager
```

### 6-4. (선택) Ecosystem 전체 서비스 활성화

```bash
# SA/AD/CE 에이전트 포함 전체 스택
sudo systemctl enable 97layer-ecosystem
sudo systemctl start 97layer-ecosystem
sudo systemctl status 97layer-ecosystem --no-pager
```

### 6-5. 서비스 관리 명령

```bash
# 상태 확인 (전체)
sudo systemctl status 97layer-telegram 97layer-nightguard

# 로그 실시간 확인
sudo journalctl -u 97layer-nightguard -f

# 재시작
sudo systemctl restart 97layer-telegram
sudo systemctl restart 97layer-nightguard

# 시간대 설정 (한국 표준시)
sudo timedatectl set-timezone Asia/Seoul
```

### 6-6. .infra/logs 디렉토리 생성 (최초 1회)

```bash
mkdir -p ~/97layerOS/.infra/logs
```

---

## 7. 배포 체크리스트

```
[ ] SSH 연결 확인: ssh 97layer-vm "echo OK"
[ ] .env 파일 VM에 존재 확인
[ ] requirements.txt 의존성 설치 확인
[ ] systemd 서비스 active 상태 확인
[ ] Telegram Bot 응답 확인 (/status)
[ ] Nightguard V2 실행 확인
```

---

## 8. 긴급 복구

```bash
# 자동 복구
python core/system/handoff.py --emergency-recovery

# 수동 롤백
ssh 97layer-vm
cd ~/97layerOS
tar -xzf ~/97layer-backup-YYYYMMDD-HHMMSS.tar.gz
sudo systemctl restart 97layer-telegram
```
