# 하이브리드 지능망: Zero-Cost 최적화 전략

**날짜**: 2026-02-15
**전략**: 맥북 (전투기) + VM (정찰기) + Cloud Run (레이더)
**비용**: **$0/월** (Google 무료 플랜 100% 활용)

---

## I. 전략 개요

### 핵심 개념

"맥북은 '전투기', VM은 '정찰기', Cloud Run은 '레이더'다."

이 시스템은 Google Cloud 무료 플랜의 한계치를 1원도 넘기지 않으면서, 맥북과 클라우드가 서로의 빈틈을 메우는 **무결점 하이브리드 배치**입니다.

### 역할 분담 (Zero-Cost Matrix)

| 환경 | 하드웨어 | 핵심 역할 | 가동 전략 |
|------|---------|---------|----------|
| **맥북 (Local)** | M1/M2 High Spec | Heavy Lifting: 고부하 AI 분석, 대용량 파일 처리, 전체 시스템 빌드 | 주간/수동: 사용자가 깨어 있을 때 풀파워 가동 |
| **무료 VM (GCP)** | e2-micro (RAM 1GB) | Night Guard: 트렌드 크롤링, RSS 모니터링, 맥북 부재 시 상태 유지 | 24/7/자동: 미국(us-west1) 지역에서 상시 감시 |
| **Cloud Run** | Serverless | Gatekeeper: 텔레그램 봇 응답, 외부 웹훅(Webhook) 수신 | 이벤트 기반: 메시지가 올 때만 잠깐 깨어남 |

---

## II. 아키텍처 다이어그램

```
┌────────────────────────────────────────────────────┐
│      Google Cloud Scheduler (트리거 발사기)         │
│      ─────────────────────────────────────────     │
│      • 09:00 daily → Cloud Run (content)           │
│      • 06:00 daily → Cloud Run (trends)            │
│      • 00:00 weekly → VM (evolution)               │
└──────────┬─────────────────────┬───────────────────┘
           │                     │
           ▼                     ▼
┌──────────────────────┐  ┌──────────────────────┐
│  Cloud Run           │  │  VM (e2-micro)       │
│  (서버리스, 즉시)     │  │  (24/7 무료)         │
│  ─────────────────   │  │  ─────────────────   │
│  ✓ Telegram webhook  │  │  ✓ Night Guard       │
│  ✓ 빠른 AI 응답      │  │  ✓ 트렌드 크롤링     │
│  ✓ Health check      │  │  ✓ 주권 확인         │
│  ✓ 컨텐츠 아이디어   │  │  ✓ Heartbeat 갱신    │
│  ✓ 트렌드 분석       │  │  ✓ 관찰 모드         │
│  (1-2분 작업)        │  │  (10분+ 작업)        │
└──────────┬───────────┘  └──────────┬───────────┘
           │                         │
           └──────────┬──────────────┘
                      ▼
           ┌────────────────────────┐
           │  Google Drive (Hub)     │
           │  ────────────────────   │
           │  • sync_state.json      │
           │  • Handshake 프로토콜   │
           │  • 주권 확인            │
           │  • 결과물 저장          │
           └────────────┬───────────┘
                        ▲
                        │
           ┌────────────┴───────────┐
           │  MacBook (전투기)       │
           │  ────────────────────   │
           │  • 5-Agent 병렬 (11초)  │
           │  • Multimodal 처리      │
           │  • 복잡한 전략 분석     │
           │  • Heartbeat 송신       │
           └────────────────────────┘
```

---

## III. The Handshake 프로토콜 (주권 확인)

### 개념

맥북과 VM이 충돌 없이 '바톤'을 넘기는 지능 이관 메커니즘.

### 중앙 신경계: `sync_state.json`

```json
{
  "active_node": "macbook",
  "last_heartbeat": "2026-02-15T15:30:00Z",
  "pending_handover": false,
  "node_history": [
    {"from": "macbook", "to": "gcp_vm", "timestamp": "...", "reason": "timeout"},
    {"from": "gcp_vm", "to": "macbook", "timestamp": "...", "reason": "recovery"}
  ],
  "health": {
    "macbook": "online",
    "gcp_vm": "standby"
  }
}
```

### 주권 확인 로직

```python
# execution/system/hybrid_sync.py

def claim_ownership(node: NodeType, timeout_minutes: int = 10) -> bool:
    """
    주권 요청 (The Handshake)

    Case 1: 이미 본인이 주권 보유 → True (Heartbeat 갱신)
    Case 2: 타임아웃 발생 (10분 무응답) → 주권 이관 → True
    Case 3: 타 노드 활성 → False (관찰 모드)
    """
```

### 동작 시나리오

#### 시나리오 1: 평소 (맥북 활성)
1. 맥북: `claim_ownership("macbook")` → True
2. 맥북: Heartbeat 갱신 (매 5분)
3. VM: `claim_ownership("gcp_vm")` → False (관찰 모드)

#### 시나리오 2: 맥북 오프라인 (10분 경과)
1. VM: Heartbeat 체크 → 10분 경과 감지
2. VM: 주권 이관 (`gcp_vm`으로 승격)
3. VM: 트렌드 크롤링 시작

#### 시나리오 3: 맥북 복귀
1. 맥북: `claim_ownership("macbook")` → Heartbeat 확인
2. 맥북: 타임아웃 없음 → VM이 주권 보유 중
3. VM: 다음 주기에 맥북 Heartbeat 감지 → 주권 반환

---

## IV. Low-Power Mode (RAM 1GB 최적화)

### 문제

GCP 무료 VM (e2-micro)의 RAM 1GB로는 5-agent 병렬 처리가 불가능.

### 해결책

환경 감지 후 자동으로 Low-Power Mode 활성화.

```python
# libs/core_config.py

def detect_environment() -> str:
    if Path("/etc/google_compute_engine").exists():
        return "GCP_VM"
    elif Path("/.dockerenv").exists():
        return "CLOUD_CONTAINER"
    else:
        return "MACBOOK"

ENVIRONMENT = detect_environment()

if ENVIRONMENT == "GCP_VM":
    # Night Guard 모드
    PROCESSING_MODE = "sequential"       # 순차 처리
    MAX_CONCURRENT_AGENTS = 1            # 1개만 실행
    ENABLE_MULTIMODAL = False            # 이미지 분석 비활성화
    MEMORY_LIMIT_MB = 700                # 메모리 제한
    AI_MODEL_PREFERENCE = "gemini-1.5-flash"  # 경량 모델
else:
    # Full Power 모드
    PROCESSING_MODE = "parallel"         # 병렬 처리
    MAX_CONCURRENT_AGENTS = 5            # 5-agent 동시
    ENABLE_MULTIMODAL = True             # 이미지 분석
    MEMORY_LIMIT_MB = None
    AI_MODEL_PREFERENCE = "gemini-1.5-pro"
```

### Swap Memory 2GB

```bash
# VM 초기화 시 자동 생성 (deployment/init_nightguard.sh)
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

---

## V. US-West Strategy (무료 리전)

### 무료 VM 조건

Google Cloud 무료 플랜의 e2-micro는 **특정 리전에서만 무료**입니다.

#### 무료 리전 (반드시 준수)
- `us-west1` (오리건) ✅ **권장**
- `us-central1` (아이오와)
- `us-east1` (사우스캐롤라이나)

#### 유료 리전 (비용 발생)
- `asia-northeast3` (서울) ❌
- `asia-northeast1` (도쿄) ❌
- 기타 모든 리전 ❌

### VM 생성 명령어

```bash
# deployment/create_nightguard.sh

gcloud compute instances create 97layer-nightguard \
  --zone=us-west1-b \               # ✅ 무료 리전
  --machine-type=e2-micro \         # ✅ 무료 머신
  --boot-disk-size=30GB \           # ✅ 무료 범위
  --image-family=ubuntu-minimal-2204-lts  # ✅ 경량 OS
```

---

## VI. 배치 실행 가이드

### Step 1: VM 생성

```bash
cd /Users/97layer/97layerOS/deployment
./create_nightguard.sh
```

**예상 시간**: 2-3분

### Step 2: VM SSH 접속

```bash
gcloud compute ssh 97layer-nightguard --zone=us-west1-b
```

### Step 3: 97layerOS 복사 (맥북 → VM)

```bash
# 맥북에서 실행
gcloud compute scp --recurse /Users/97layer/97layerOS \
  97layer-nightguard:~/ --zone=us-west1-b
```

**예상 시간**: 5-10분 (프로젝트 크기에 따라)

### Step 4: VM 초기화

```bash
# VM 내부에서 실행
cd ~/97layerOS/deployment
chmod +x init_nightguard.sh
./init_nightguard.sh
```

**작업 내역**:
- Swap 2GB 생성
- Python 3.10+ 설치
- 의존성 설치
- systemd 서비스 등록
- Night Guard 가동

### Step 5: Cloud Scheduler 설정

```bash
# 맥북에서 실행
cd /Users/97layer/97layerOS/deployment
./setup_scheduler.sh
```

**생성되는 Job**:
1. `daily-content`: 매일 09:00 컨텐츠 아이디어 (Cloud Run)
2. `daily-trends`: 매일 06:00 트렌드 분석 (Cloud Run)
3. `weekly-evolution`: 매주 일요일 00:00 Gardener (VM)

### Step 6: 테스트

#### 6.1 맥북 Heartbeat 확인

```bash
python3 -c "
from execution.system.hybrid_sync import HybridSync
sync = HybridSync()
print(f'환경: {sync.location}')
print(f'노드 타입: {sync.get_node_type()}')
result = sync.claim_ownership(sync.get_node_type())
print(f'주권 획득: {result}')
"
```

#### 6.2 VM Night Guard 로그 확인

```bash
# VM에서 실행
sudo journalctl -u 97layeros-nightguard -f
```

예상 로그:
```
[2026-02-15 15:30:00] ○ 관찰 모드 (맥북 활성) (Cycle #1)
[2026-02-15 15:35:00] ○ 관찰 모드 (맥북 활성) (Cycle #2)
```

#### 6.3 맥북 오프라인 시뮬레이션

맥북에서 Heartbeat 중지 → 10분 대기 → VM 로그 확인:

```
[2026-02-15 15:45:00] ✓ Night Guard 활성화 (Cycle #3)
[Handshake] macbook 타임아웃 (0:10:02) → gcp_vm로 주권 이관
✅ 주권 이관: macbook → gcp_vm
🔍 트렌드 크롤링 시작...
✅ 트렌드 크롤링 완료: 3개 항목
```

---

## VII. 무료 플랜 검증

| 항목 | 무료 한도 | 예상 사용량 | 사용률 | 비용 |
|------|----------|-----------|-------|------|
| **VM (e2-micro)** | 730시간/월 | 730시간 (24/7) | 100% | **$0** |
| **Cloud Run** | 200만 요청 | 3,090 요청 | 0.15% | **$0** |
| **Cloud Scheduler** | 3 job 무료 | 3 job | 100% | **$0** |
| **Cloud Storage** | 5GB | 500MB | 10% | **$0** |
| **Network Egress** | 1GB/월 | 400MB | 40% | **$0** |
| **총합** | - | - | - | **$0/월** |

### 무료 플랜 주의사항

1. **VM 리전**: 반드시 us-west1/us-central1/us-east1 사용
2. **머신 타입**: e2-micro만 무료 (다른 타입은 유료)
3. **디스크**: 30GB Standard까지 무료
4. **Cloud Scheduler**: 3 job 초과 시 $0.10/job/월
5. **Network**: 1GB 초과 시 $0.12/GB

---

## VIII. 유지보수 및 모니터링

### VM 상태 확인

```bash
# VM 목록
gcloud compute instances list

# VM 상세 정보
gcloud compute instances describe 97layer-nightguard --zone=us-west1-b

# VM SSH 접속
gcloud compute ssh 97layer-nightguard --zone=us-west1-b
```

### Night Guard 서비스 관리

```bash
# VM 내부에서 실행

# 상태 확인
sudo systemctl status 97layeros-nightguard

# 로그 확인 (실시간)
sudo journalctl -u 97layeros-nightguard -f

# 재시작
sudo systemctl restart 97layeros-nightguard

# 중지
sudo systemctl stop 97layeros-nightguard

# 시작
sudo systemctl start 97layeros-nightguard
```

### Cloud Scheduler Job 관리

```bash
# Job 목록
gcloud scheduler jobs list --location=us-central1

# Job 수동 실행 (테스트)
gcloud scheduler jobs run daily-content --location=us-central1

# Job 일시 정지
gcloud scheduler jobs pause daily-content --location=us-central1

# Job 재개
gcloud scheduler jobs resume daily-content --location=us-central1

# Job 삭제
gcloud scheduler jobs delete daily-content --location=us-central1
```

### 무료 플랜 사용량 모니터링

Google Cloud Console:
1. [https://console.cloud.google.com/billing](https://console.cloud.google.com/billing)
2. 프로젝트 선택: `layer97os`
3. "보고서" 탭 → 이번 달 비용 확인

---

## IX. 문제 해결 (Troubleshooting)

### 문제 1: VM 생성 실패

**증상**: `gcloud compute instances create` 오류

**해결책**:
```bash
# Compute Engine API 활성화 확인
gcloud services enable compute.googleapis.com

# 프로젝트 ID 확인
gcloud config get-value project

# 리전 확인 (us-west1이 맞는지)
gcloud compute zones list | grep us-west1
```

### 문제 2: Night Guard 서비스 시작 실패

**증상**: `sudo systemctl status 97layeros-nightguard` → failed

**해결책**:
```bash
# 로그 확인
sudo journalctl -u 97layeros-nightguard -n 50

# Python 경로 확인
which python3

# 의존성 재설치
cd ~/97layerOS
pip3 install -r requirements.txt

# systemd 파일 권한 확인
ls -la /etc/systemd/system/97layeros-nightguard.service
```

### 문제 3: Handshake 주권 충돌

**증상**: 맥북과 VM이 동시에 작동

**해결책**:
```bash
# sync_state.json 수동 리셋
cd /Users/97layer/97layerOS
cat > knowledge/system/sync_state.json << EOF
{
  "active_node": "macbook",
  "last_heartbeat": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "pending_handover": false,
  "node_history": [],
  "health": {"macbook": "online", "gcp_vm": "standby"}
}
EOF

# VM 재시작
gcloud compute ssh 97layer-nightguard --zone=us-west1-b \
  --command="sudo systemctl restart 97layeros-nightguard"
```

---

## X. 다음 단계

### 완료 체크리스트

- [ ] VM 생성 (us-west1-b)
- [ ] 97layerOS 복사 (맥북 → VM)
- [ ] VM 초기화 (Swap, systemd)
- [ ] Night Guard 가동 확인
- [ ] Cloud Scheduler 3 job 생성
- [ ] 맥북 Heartbeat 테스트
- [ ] VM 자동 승격 테스트 (10분 타임아웃)
- [ ] 무료 플랜 사용량 모니터링 설정

### 향후 확장 가능성

1. **Cloud Run 엔드포인트 추가**:
   - `/scheduled/content` - 컨텐츠 아이디어 생성
   - `/scheduled/trends` - 트렌드 분석 리포트

2. **VM Flask 앱 구축** (선택):
   - `/scheduled/evolution` - Gardener 진화 사이클
   - `/job/heavy` - 대용량 데이터 처리

3. **자동화 확장**:
   - 매일 자정: 스냅샷 백업
   - 30분마다: 클립보드 아카이브
   - 주간 리포트 생성

---

## XI. 결론

**하이브리드 지능망: Zero-Cost 최적화** 전략은 Google Cloud 무료 플랜을 최대한 활용하여 **$0/월 비용**으로 24/7 자동화 시스템을 구축하는 전략입니다.

### 핵심 장점

1. ✅ **완전 무료**: VM 24/7 + Cloud Run + Scheduler 모두 무료 플랜 내
2. ✅ **무결점 전환**: Handshake 프로토콜로 맥북-VM 충돌 방지
3. ✅ **자동 복구**: 맥북 오프라인 시 VM 자동 승격
4. ✅ **메모리 최적화**: Low-Power Mode로 RAM 1GB 극복
5. ✅ **확장 가능**: Cloud Scheduler로 자동화 추가 용이

### 최종 아키텍처

**"맥북(전투기) + VM(정찰기) + Cloud Run(레이더)"**

이 3-Layer 하이브리드 지능망으로 97layerOS는 24시간 내내 트렌드를 감시하면서도, 맥북을 켰을 때만 강력한 성능(5-Agent 병렬)으로 결과를 뽑아내는 **비용 효율 1,000%의 시스템**을 갖게 됩니다.

---

**문서 작성**: 2026-02-15
**작성자**: Claude (97layerOS Technical Director)
**버전**: 1.0.0
