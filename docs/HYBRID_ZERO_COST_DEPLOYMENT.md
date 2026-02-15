# 하이브리드 지능망: Zero-Cost 최적화 전략

**버전**: 1.0.0
**날짜**: 2026-02-15
**전략**: 맥북(전투기) + VM(정찰기) + Cloud Run(레이더)
**비용**: **$0/월** (Google 무료 플랜 100% 활용)

---

## I. 전략 개요

### 핵심 개념
"맥북은 전투기, VM은 정찰기, Cloud Run은 레이더다."

97layerOS를 Google Cloud 무료 플랜(Free Tier) 범위 내에서 24/7 운영하는 하이브리드 배치 전략.

### 역할 분담

| 환경 | 하드웨어 | 핵심 역할 | 가동 전략 |
|------|----------|----------|----------|
| **맥북 (전투기)** | M1/M2 High Spec | Heavy Lifting: 고부하 AI 분석, 5-Agent 병렬 처리 | 주간/수동: 사용자 작업 시 풀파워 |
| **VM (정찰기)** | e2-micro (RAM 1GB) | Night Guard: 트렌드 감시, 맥북 부재 시 상태 유지 | 24/7/자동: us-west1 상시 감시 |
| **Cloud Run (레이더)** | Serverless | Gatekeeper: Telegram 봇, 외부 webhook 수신 | 이벤트 기반: 요청 시만 활성화 |

---

## II. 인프라 상태 확인

### 현재 배포 상태

#### ✅ Cloud Run (레이더)
```bash
gcloud run services list
```
- **서비스명**: `telegram-bot`
- **리전**: `asia-northeast3` (서울)
- **URL**: `https://telegram-bot-514569077225.asia-northeast3.run.app`
- **상태**: 활성화 ✅

#### ❌ GCP VM (정찰기)
```bash
gcloud compute instances list
```
- **결과**: 0개 (생성 필요)

---

## III. 구현 완료 항목

### 1. Handshake 프로토콜 ✅
**파일**: `execution/system/hybrid_sync.py`

**기능**:
- `claim_ownership(node, timeout_minutes)` - 주권 확인 메서드
- 맥북 10분 무응답 → VM 자동 승격
- `sync_state.json`에 `active_node`, `last_heartbeat` 기록

**주권 확인 로직**:
```python
# 맥북 온라인
sync_state.json: {"active_node": "macbook"}
→ VM은 관찰 모드

# 맥북 10분 무응답
→ VM이 "gcp_vm"으로 자동 승격
→ 트렌드 크롤링 시작

# 맥북 복귀
→ VM이 주권 반환 → 관찰 모드 전환
```

---

### 2. Low-Power Mode (메모리 최적화) ✅
**파일**: `libs/core_config.py`

**GCP_VM 모드 설정**:
```python
ENVIRONMENT = "GCP_VM"
PROCESSING_MODE = "sequential"          # 순차 처리
MAX_CONCURRENT_AGENTS = 1               # 1개만 실행
ENABLE_MULTIMODAL = False               # 이미지 분석 비활성화
MEMORY_LIMIT_MB = 700                   # 700MB 제한
AI_MODEL_PREFERENCE = "gemini-1.5-flash"  # 경량 모델
```

**환경 감지**:
```python
def detect_environment() -> str:
    if Path("/etc/google_compute_engine").exists():
        return "GCP_VM"
    elif Path("/.dockerenv").exists():
        return "CLOUD_CONTAINER"
    else:
        return "MACBOOK"
```

---

### 3. Night Guard 데몬 ✅
**파일**: `execution/system/nightguard_daemon.py`

**기능**:
- 5분마다 주권 확인 (`claim_ownership`)
- 주권 획득 시:
  - Google Drive 동기화 (최신 상태 pull)
  - 트렌드 크롤링 실행
  - 상태 보고 (Telegram 알림)
- 관찰 모드 시:
  - 대기 (맥북 활성 중)

**실행 예시**:
```python
# VM에서 실행
python3 execution/system/nightguard_daemon.py

# 출력:
# [2026-02-15 15:30:00] ✓ Night Guard 활성화 (Cycle #1)
# [2026-02-15 15:35:00] ○ 관찰 모드 (맥북 활성) (Cycle #2)
```

---

### 4. VM 배치 스크립트 ✅

#### A. VM 생성 스크립트
**파일**: `deployment/create_nightguard.sh`

```bash
#!/bin/bash
# US-West1 무료 리전에 e2-micro 생성

gcloud compute instances create 97layer-nightguard \
  --zone=us-west1-b \
  --machine-type=e2-micro \
  --boot-disk-size=30GB \
  --image-family=ubuntu-minimal-2204-lts \
  --tags=97layer-nightguard \
  --scopes=cloud-platform
```

**실행**:
```bash
cd /Users/97layer/97layerOS/deployment
chmod +x create_nightguard.sh
./create_nightguard.sh
```

---

#### B. VM 초기화 스크립트
**파일**: `deployment/init_nightguard.sh`

**작업 내용**:
1. **Swap 2GB 생성** (RAM 1GB 극복)
   ```bash
   sudo fallocate -l 2G /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   ```

2. **Python 3.10+ 설치**
3. **Podman 설치** (경량 컨테이너)
4. **97layerOS 클론**
5. **환경변수 설정** (`.env`)
6. **의존성 설치** (`requirements.txt`)
7. **systemd 서비스 등록**

**실행** (VM SSH 접속 후):
```bash
cd ~/97layerOS/deployment
chmod +x init_nightguard.sh
./init_nightguard.sh
```

---

#### C. systemd 서비스 파일
**파일**: `deployment/97layeros-nightguard.service`

```ini
[Unit]
Description=97LAYER Night Guard (정찰기)
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/97layerOS
ExecStart=/usr/bin/python3 /home/ubuntu/97layerOS/execution/system/nightguard_daemon.py
Restart=always
RestartSec=10
Environment="ENVIRONMENT=GCP_VM"
Environment="PROCESSING_MODE=sequential"

# 리소스 제한 (RAM 700MB)
MemoryMax=700M
MemoryHigh=600M

[Install]
WantedBy=multi-user.target
```

**관리 명령**:
```bash
# 상태 확인
sudo systemctl status 97layeros-nightguard

# 로그 확인
sudo journalctl -u 97layeros-nightguard -f

# 재시작
sudo systemctl restart 97layeros-nightguard
```

---

## IV. 무료 플랜 검증

### Google Cloud Always Free Tier 제공 항목

| 항목 | 무료 한도 | 예상 사용량 | 사용률 | 비용 |
|------|----------|-----------|-------|------|
| **VM (e2-micro)** | 730시간/월 | 730시간 (24/7) | 100% | **$0** |
| **Cloud Run** | 200만 요청 | 3,090 요청 | 0.15% | **$0** |
| **Cloud Scheduler** | 3 job 무료 | 0개 (추가 예정) | 0% | **$0** |
| **Cloud Storage** | 5GB | 500MB | 10% | **$0** |
| **Network Egress** | 1GB | 400MB | 40% | **$0** |
| **총합** | - | - | - | **$0/월** |

### 중요 제약사항

1. **리전 제한**: us-west1, us-central1, us-east1만 무료
2. **서울 리전 불가**: asia-northeast3는 유료
3. **Swap 필수**: RAM 1GB는 부족 → 2GB Swap 필요
4. **경량 OS**: Ubuntu Minimal 권장 (GUI 없음)

---

## V. 배치 절차 (Step-by-Step)

### Phase 1: 맥북에서 사전 준비 ✅

이미 완료된 항목:
- [x] Handshake 프로토콜 (`hybrid_sync.py`)
- [x] Low-Power Mode (`core_config.py`)
- [x] Night Guard 데몬 (`nightguard_daemon.py`)
- [x] VM 배치 스크립트 (3개)

---

### Phase 2: VM 생성

```bash
# 1. 맥북에서 VM 생성 스크립트 실행
cd /Users/97layer/97layerOS/deployment
chmod +x create_nightguard.sh
./create_nightguard.sh

# 2. 생성 확인
gcloud compute instances list

# 출력 예시:
# NAME                ZONE        MACHINE_TYPE  STATUS
# 97layer-nightguard  us-west1-b  e2-micro      RUNNING

# 3. 외부 IP 확인
gcloud compute instances describe 97layer-nightguard \
  --zone=us-west1-b \
  --format="get(networkInterfaces[0].accessConfigs[0].natIP)"
```

---

### Phase 3: VM 초기화

```bash
# 1. VM SSH 접속
gcloud compute ssh 97layer-nightguard --zone=us-west1-b

# 2. 97layerOS 복사 (맥북에서 VM으로)
# 방법 A: gcloud scp (권장)
gcloud compute scp --recurse \
  /Users/97layer/97layerOS \
  97layer-nightguard:~/ \
  --zone=us-west1-b

# 방법 B: Git clone (SSH 키 필요)
git clone git@github.com:your-org/97layerOS.git

# 3. 환경변수 설정 (VM에서)
cd ~/97layerOS
cat > .env << EOF
ENVIRONMENT=GCP_VM
PROCESSING_MODE=sequential
TELEGRAM_BOT_TOKEN=your_token
GEMINI_API_KEY=your_key
ANTHROPIC_API_KEY=your_key
EOF

# 4. 초기화 스크립트 실행
cd deployment
chmod +x init_nightguard.sh
./init_nightguard.sh

# 5. 상태 확인
sudo systemctl status 97layeros-nightguard
```

---

### Phase 4: Handshake 테스트

#### A. 맥북에서 주권 확인
```bash
cd /Users/97layer/97layerOS
python3 -c "
from execution.system.hybrid_sync import HybridSync
sync = HybridSync()
print(f'환경: {sync.location}')
print(f'노드 타입: {sync.get_node_type()}')
has_ownership = sync.claim_ownership('macbook', timeout_minutes=10)
print(f'주권: {has_ownership}')
"

# 출력 예시:
# 환경: LOCAL_MAC
# 노드 타입: macbook
# 주권: True
```

#### B. VM에서 관찰 모드 확인
```bash
# VM SSH 접속
gcloud compute ssh 97layer-nightguard --zone=us-west1-b

# 로그 확인
sudo journalctl -u 97layeros-nightguard -f

# 출력 예시:
# [2026-02-15 15:30:00] ○ 관찰 모드 (맥북 활성) (Cycle #1)
```

#### C. 맥북 오프라인 → VM 승격 테스트
```bash
# 1. 맥북에서 Handshake 중단 (10분 대기)
# (아무것도 하지 않음)

# 2. VM 로그 확인 (10분 후)
sudo journalctl -u 97layeros-nightguard -f

# 출력 예시:
# [2026-02-15 15:40:00] [Handshake] macbook 타임아웃 → gcp_vm으로 주권 이관
# [2026-02-15 15:40:01] ✓ Night Guard 활성화 (Cycle #3)
# [2026-02-15 15:40:05] ✅ Google Drive → VM 동기화 완료
# [2026-02-15 15:40:10] ✅ 트렌드 0개 수집
```

---

## VI. 운영 가이드

### 일상 운영

**맥북 사용 시** (전투기 모드):
```bash
# 아무것도 하지 않아도 자동으로 주권 유지
# VM은 관찰 모드로 자동 전환
```

**맥북 꺼둘 때** (정찰기 모드):
```bash
# 10분 후 VM이 자동으로 주권 획득
# 트렌드 감시, 상태 유지 계속
```

---

### 모니터링

#### A. VM 상태 확인
```bash
# SSH 접속
gcloud compute ssh 97layer-nightguard --zone=us-west1-b

# 서비스 상태
sudo systemctl status 97layeros-nightguard

# 실시간 로그
sudo journalctl -u 97layeros-nightguard -f

# Swap 확인 (2GB 활성화 확인)
free -h

# 메모리 사용량
ps aux --sort=-%mem | head
```

#### B. 주권 상태 확인
```bash
# 맥북 또는 VM에서
cat knowledge/system/sync_state.json | python3 -m json.tool

# 출력 예시:
# {
#   "active_node": "macbook",
#   "last_heartbeat": "2026-02-15T15:30:00",
#   "health": {
#     "macbook": "online",
#     "gcp_vm": "standby"
#   }
# }
```

---

### 문제 해결

#### 1. VM이 주권을 획득하지 못함
**증상**: 맥북 오프라인인데도 VM이 관찰 모드

**원인**: `sync_state.json`이 Google Drive와 동기화되지 않음

**해결**:
```bash
# VM에서 수동 동기화
cd ~/97layerOS
python3 execution/system/hybrid_sync.py pull

# sync_state.json 확인
cat knowledge/system/sync_state.json
```

#### 2. VM 메모리 부족 (Out of Memory)
**증상**: Night Guard 크래시

**원인**: Swap 미활성화 또는 메모리 누수

**해결**:
```bash
# Swap 확인
free -h
# Swap이 0이면 재생성
sudo swapon /swapfile

# 서비스 재시작
sudo systemctl restart 97layeros-nightguard
```

#### 3. Cloud Run 요청 실패
**증상**: Telegram 봇 무응답

**원인**: Cloud Run 서비스 중지 또는 환경변수 누락

**해결**:
```bash
# 서비스 상태 확인
gcloud run services list

# 재배포
cd /Users/97layer/97layerOS/deployment
./deploy_google_cloud.sh
```

---

## VII. 비용 모니터링

### GCP 콘솔에서 확인

1. **Compute Engine**:
   - `https://console.cloud.google.com/compute/instances`
   - e2-micro가 us-west1-b에 있는지 확인
   - 730시간/월 이내 → $0

2. **Cloud Run**:
   - `https://console.cloud.google.com/run`
   - 요청 수 200만 이내 확인
   - $0

3. **Cloud Storage**:
   - `https://console.cloud.google.com/storage`
   - 5GB 이내 확인
   - $0

### 알림 설정

```bash
# 청구 알림 설정 (무료 플랜 초과 시 알림)
gcloud billing budgets create \
  --billing-account=YOUR_BILLING_ACCOUNT_ID \
  --display-name="97layerOS Free Tier Alert" \
  --budget-amount=1 \
  --threshold-rule=percent=50,basis=current-spend
```

---

## VIII. 향후 확장

### Cloud Scheduler 추가 (선택 사항)

**목적**: 정기 작업 자동화

**예시**:
- 매일 09:00: 컨텐츠 아이디어 생성
- 매주 일요일: Gardener 진화 사이클

**설정**:
```bash
# Job 생성 (3개까지 무료)
gcloud scheduler jobs create http daily-content \
  --schedule="0 9 * * *" \
  --uri="https://telegram-bot-xxx.run.app/scheduled/content" \
  --http-method=POST
```

---

## IX. 체크리스트

### 배치 전 확인
- [ ] GCP 프로젝트 생성 (`layer97os`)
- [ ] Compute Engine API 활성화
- [ ] 환경변수 준비 (TELEGRAM_BOT_TOKEN, GEMINI_API_KEY)
- [ ] 맥북에서 코드 최신화 (`git pull`)

### 배치 완료 확인
- [ ] VM 생성 완료 (us-west1-b, e2-micro)
- [ ] Swap 2GB 활성화 (`free -h`)
- [ ] Night Guard 서비스 실행 중 (`systemctl status`)
- [ ] Handshake 테스트 통과 (맥북 → VM 주권 이관)
- [ ] Cloud Run 정상 작동 (Telegram 봇 응답)

### 운영 중 확인 (주 1회)
- [ ] VM 상태 확인 (`gcloud compute instances list`)
- [ ] 무료 플랜 한도 확인 (GCP 콘솔)
- [ ] 로그 확인 (에러 없는지)
- [ ] sync_state.json 주권 상태 확인

---

## X. 요약

✅ **완료된 구현**:
1. Handshake 프로토콜 (주권 확인)
2. Low-Power Mode (RAM 1GB 최적화)
3. Night Guard 데몬 (24/7 정찰기)
4. VM 배치 스크립트 (3개)
5. systemd 서비스 (자동 시작)

🚀 **다음 단계**:
1. VM 생성 실행 (`./deployment/create_nightguard.sh`)
2. VM 초기화 (`./deployment/init_nightguard.sh`)
3. Handshake 테스트 (맥북 오프라인 → VM 승격 확인)

💰 **비용**: **$0/월** (무료 플랜 100% 활용)

---

**문의 및 지원**:
- GitHub Issues: `https://github.com/97layer/97layerOS/issues`
- Telegram: `@97layerOS_bot`
