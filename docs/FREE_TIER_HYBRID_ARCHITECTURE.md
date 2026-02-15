# 무료 티어 하이브리드 아키텍처 구현 가능성 분석

**날짜**: 2026-02-15
**결론**: ✅ **구현 가능** (Google Cloud 무료 티어 범위 내)

---

## I. Google Cloud Always Free Tier 제공 항목

### 1. Compute Engine
- **e2-micro 인스턴스**: 1개 (월 730시간)
  - vCPU: 0.25-2 (버스트 가능)
  - RAM: 1GB
  - 디스크: 30GB Standard Persistent Disk
  - **리전 제한**: us-west1, us-central1, us-east1 only

### 2. Cloud Run
- 월 200만 요청
- 360,000 GB-초 메모리
- 180,000 vCPU-초
- (현재 Telegram bot 배포 중: `layer97os-bot`)

### 3. Cloud Storage
- 5GB Regional Storage (us-west1, us-central1, us-east1)
- 5,000 Class A operations/월
- 50,000 Class B operations/월

### 4. Network
- 북미 내 1GB 아웃바운드 트래픽 (중국/호주 제외)

---

## II. 97layerOS 하이브리드 아키텍처 리소스 소비량

### 현재 구성

```
MacBook (Primary)
├─ Podman Container (97layeros-telegram-bot)
│  ├─ 5-Agent System (SA, AD, CE, CD, TD)
│  ├─ Multimodal Processing (Gemini Flash/Vision)
│  ├─ Memory Manager + Gardener
│  └─ Anti-Gravity Protocol
├─ Hybrid Sync Daemon (5분마다)
└─ Google Drive API

Google Cloud Run (24/7)
├─ Telegram Webhook Receiver
└─ Request Router → MacBook/VM

Google Drive (Central Hub)
├─ Bidirectional Sync
└─ Heartbeat Status File
```

### 추가 필요: VM Standby

```
Google VM (Standby, e2-micro)
├─ Heartbeat Monitor (1분마다)
├─ MacBook 오프라인 감지 (5분 타임아웃)
└─ 동일 Podman Container 실행 (경량 모드)
```

---

## III. 무료 티어 소비량 계산

### A. Compute Engine (VM)

**시나리오 1: 항상 실행** (비추천)
- 월 730시간 = 24시간 x 30.4일
- 무료 티어: 730시간 → ✅ 커버됨
- **문제**: RAM 1GB로 5-agent 병렬 처리 버거움

**시나리오 2: 스마트 Standby** (권장)
- MacBook 온라인 (80% 시간): VM 중지
- MacBook 오프라인 (20% 시간): VM 실행
- 월 146시간 (730 x 0.2) → ✅ 무료 티어 충분
- **해결책**: 오프라인 시에만 VM 활성화

**예상 비용**: **$0/월** (무료 티어 범위 내)

---

### B. Cloud Run (Telegram Bot)

**현재 사용량**:
- 하루 100 요청 x 30일 = 3,000 요청/월
- 무료 티어: 200만 요청/월
- **사용률**: 0.15%

**예상 비용**: **$0/월**

---

### C. Cloud Storage (Google Drive API)

**현재 사용량**:
- 97layerOS 프로젝트 크기: ~500MB (추정)
- Sync 트래픽: 하루 10MB x 30 = 300MB/월
- 개인 Google Drive: 15GB 무료

**예상 비용**: **$0/월**

---

### D. Network Egress

**트래픽 분석**:
- Cloud Run → VM: 내부 트래픽 (무료)
- VM → Google Drive API: HTTPS 아웃바운드
- 월 300MB (sync) + 100MB (webhook) = 400MB
- 무료 티어: 1GB/월

**예상 비용**: **$0/월**

---

## IV. RAM 1GB 제약 해결 방안

### 문제
e2-micro의 1GB RAM으로 5-agent 병렬 처리는 메모리 압박이 심함.

### 해결책: 경량 모드 전환

**MacBook (Primary Mode)**:
```python
# 병렬 처리 (11초 응답)
async def process_request():
    sa_task = strategy_analyst.analyze()
    ad_task = art_director.design()
    results = await asyncio.gather(sa_task, ad_task)
    ce_result = chief_editor.compile(results)
    return creative_director.finalize(ce_result)
```

**VM (Standby Mode)**:
```python
# 순차 처리 (55초 응답, 하지만 작동함)
def process_request_sequential():
    sa_result = strategy_analyst.analyze()
    ad_result = art_director.design()
    ce_result = chief_editor.compile([sa_result, ad_result])
    return creative_director.finalize(ce_result)
```

**환경변수로 모드 전환**:
```bash
# MacBook
PROCESSING_MODE=parallel

# VM
PROCESSING_MODE=sequential
ENABLE_SWAP=true
```

---

## V. 스마트 Standby 구현 로직

### Heartbeat 시스템

**MacBook (Primary)**:
```python
# execution/system/heartbeat_sender.py
def send_heartbeat():
    """5분마다 Google Drive에 heartbeat 파일 업데이트"""
    heartbeat = {
        "location": "LOCAL_MAC",
        "timestamp": datetime.now().isoformat(),
        "status": "active"
    }
    drive_service.update_file("heartbeat.json", heartbeat)
```

**VM (Standby)**:
```python
# execution/system/heartbeat_monitor.py
def monitor_primary():
    """1분마다 heartbeat 체크"""
    heartbeat = drive_service.get_file("heartbeat.json")
    last_seen = datetime.fromisoformat(heartbeat["timestamp"])

    if (datetime.now() - last_seen).total_seconds() > 300:  # 5분
        logger.warning("Primary offline. Activating standby...")
        activate_vm_mode()
    else:
        if is_vm_active():
            logger.info("Primary recovered. Deactivating standby...")
            deactivate_vm_mode()
```

---

## VI. 구현 로드맵

### Phase 1: gcloud 환경 점검 ✅
- [x] Compute Engine API 활성화
- [x] 프로젝트 ID 확인: `layer97os`
- [x] 인증 확인: `skyto5339@gmail.com`

### Phase 2: VM 인스턴스 생성
```bash
gcloud compute instances create 97layeros-standby \
  --zone=us-west1-b \
  --machine-type=e2-micro \
  --boot-disk-size=30GB \
  --boot-disk-type=pd-standard \
  --image-family=cos-stable \
  --image-project=cos-cloud \
  --tags=97layer-standby \
  --metadata=enable-oslogin=true
```

### Phase 3: Heartbeat 시스템 구현
- [ ] `execution/system/heartbeat_sender.py` 작성
- [ ] `execution/system/heartbeat_monitor.py` 작성
- [ ] Google Drive에 `heartbeat.json` 파일 생성
- [ ] MacBook Launchd daemon 추가

### Phase 4: 경량 모드 환경변수
- [ ] `libs/core_config.py`에 `PROCESSING_MODE` 추가
- [ ] Agent orchestration 로직 수정 (parallel vs sequential)
- [ ] VM startup script에 환경변수 주입

### Phase 5: 테스트
- [ ] MacBook 정상 작동 시 VM 대기 확인
- [ ] MacBook 네트워크 차단 → VM 활성화 확인
- [ ] MacBook 복구 → VM 비활성화 확인
- [ ] 무료 티어 한도 모니터링

---

## VII. 대안 (VM 없이)

만약 VM 설정이 복잡하다면:

### Option A: Cloud Run Only (Degraded Mode)
```
MacBook 오프라인 시:
└─ Telegram bot → "현재 오프라인 상태입니다. 잠시 후 다시 시도해주세요."
```

### Option B: Minimal Response Mode
```
Cloud Run에서 간단한 응답만:
├─ FAQ 응답 (사전 정의된)
├─ 상태 확인
└─ "고급 AI 처리는 MacBook 복귀 후 가능합니다"
```

---

## VIII. 비용 요약

| 항목 | 무료 티어 한도 | 예상 사용량 | 비용 |
|------|---------------|------------|------|
| e2-micro VM | 730시간/월 | 146시간/월 (20%) | $0 |
| Cloud Run | 200만 요청 | 3,000 요청 | $0 |
| Cloud Storage | 5GB | 500MB | $0 |
| Network Egress | 1GB | 400MB | $0 |
| **총합** | - | - | **$0/월** |

---

## IX. 권장 사항

1. **즉시 구현 가능**: 무료 티어 충분
2. **스마트 Standby 필수**: VM 항상 실행은 RAM 부족
3. **경량 모드 전환**: VM에서는 순차 처리
4. **모니터링 설정**: GCP 무료 티어 한도 알림 설정
5. **Python 3.10+ 업그레이드**: gcloud CLI 호환성

---

## X. 다음 단계

1. VM 인스턴스 생성 여부 확인 필요
2. Heartbeat 시스템 구현 시작
3. 경량 모드 환경변수 추가

**결론**: ✅ **무료 티어 내에서 완전히 구현 가능합니다.**
