# 역할 분담 하이브리드 아키텍처

**날짜**: 2026-02-15
**전략**: MacBook (무거운 작업) + VM (가벼운 자동화) 분리

---

## I. 핵심 개념

### 기존 문제점
- Standby 방식: VM이 MacBook 오프라인 시에만 대체 역할
- RAM 1GB로 5-agent 병렬 처리는 버거움
- MacBook 꺼지면 모든 자동화도 중단

### 새로운 접근
- **MacBook**: 무거운 AI 처리 (사용자 요청 즉시 응답)
- **VM**: 가벼운 자동화 (24/7 루틴 작업)
- **각자 역할 명확**: 경쟁이 아닌 협업

---

## II. 아키텍처 다이어그램

```
┌─────────────────────────────────────────────────────┐
│         Google Cloud Run (Webhook Receiver)         │
│  - Telegram 메시지 수신                              │
│  - 작업 유형 판단 (heavy vs light)                   │
│  - MacBook/VM 상태 확인 (heartbeat)                  │
└─────────────────┬───────────────────────────────────┘
                  │
         ┌────────┴────────┐
         │                 │
    [Heavy Task]      [Light Task]
         │                 │
         ▼                 ▼
┌─────────────────┐  ┌──────────────────┐
│  MacBook         │  │  VM (e2-micro)   │
│  (온/오프 가능)   │  │  (24/7 운영)     │
│  ─────────────  │  │  ──────────────  │
│  • 5-Agent      │  │  • Cron Jobs     │
│  • Multimodal   │  │  • Content Gen   │
│  • 병렬 처리     │  │  • Gardener      │
│  • 11초 응답     │  │  • Clipboard     │
│  • 즉시 대화     │  │  • Backups       │
│  • RAM 16GB+    │  │  • RAM 1GB       │
└─────────────────┘  └──────────────────┘
         │                 │
         └────────┬────────┘
                  ▼
       ┌──────────────────────┐
       │   Google Drive (Hub)  │
       │   - Bidirectional     │
       │   - Heartbeat         │
       │   - Task Queue        │
       └──────────────────────┘
```

---

## III. 역할 분담 상세

### A. MacBook (Heavy Processing)

**책임**: 사용자와의 실시간 대화, 복잡한 AI 처리

**처리 작업**:
1. **5-Agent 병렬 처리**
   - SA + AD 동시 분석 (11초 응답)
   - CE → CD 순차 처리

2. **Multimodal Processing**
   - 이미지 분석 (Vision API)
   - 영상 컨텐츠 기획

3. **복잡한 전략 분석**
   - WOOHWAHAE 브랜드 아이덴티티 분석
   - 장기 전략 수립

4. **즉시 응답 필요 작업**
   - 사용자 질문 (대화형)
   - 긴급 요청

**작동 방식**:
- 사용자가 MacBook 앞에 있을 때: 즉시 처리
- MacBook 오프라인: Cloud Run이 큐에 추가 → 복귀 시 처리

---

### B. VM (Light Automation)

**책임**: 24/7 루틴 작업, 자동화, 시스템 유지보수

**처리 작업**:

#### 1. 스케줄된 컨텐츠 생성
```python
# execution/vm/scheduled_content.py
# 매일 오전 9시 실행
def daily_content_ideation():
    """최근 활동 기반 인스타그램 컨텐츠 아이디어 생성"""
    context = memory_manager.get_recent_context(hours=24)
    prompt = f"""
    최근 활동: {context}

    WOOHWAHAE 브랜드 철학에 맞는 인스타그램 컨텐츠 3가지 제안:
    - Slow (자신의 파동)
    - 실용적 미학
    - 과정 중시
    """
    ideas = ai_engine.generate_response(prompt)
    save_to_drive("content_ideas/{today}.md", ideas)
```

#### 2. Gardener 진화 사이클
```python
# execution/vm/weekly_evolution.py
# 매주 일요일 자정 실행
def weekly_system_evolution():
    """시스템 자가 진화 (directive 업데이트)"""
    gardener.run_cycle(days=7)
    gardener.analyze_and_heal()
```

#### 3. 클립보드 아카이브
```python
# execution/system/clipboard_sentinel.py (기존)
# 30분마다 실행 (이미 구현됨)
```

#### 4. 스냅샷 백업
```python
# execution/system/snapshot_daemon.py (기존)
# 매일 자정 실행 (이미 구현됨)
```

#### 5. 로그 정리 및 분석
```python
# execution/vm/log_analysis.py
# 매일 자정 실행
def analyze_daily_logs():
    """에러 패턴 분석 및 리포트 생성"""
    error_logger.analyze_patterns()
    generate_health_report()
```

#### 6. Drive Sync 모니터링
```python
# execution/vm/sync_monitor.py
# 5분마다 실행
def monitor_sync_health():
    """MacBook-Drive 동기화 상태 체크"""
    check_heartbeat()
    validate_file_integrity()
```

**작동 방식**:
- VM은 항상 켜져 있음 (무료 티어 730시간/월)
- RAM 1GB로 충분 (자동화는 메모리 적게 씀)
- MacBook 상태와 무관하게 독립 실행

---

## IV. 작업 라우팅 로직

### Cloud Run의 판단 기준

```python
# execution/cloud_run/router.py

def route_task(message: str, user_id: str) -> str:
    """작업을 MacBook 또는 VM으로 라우팅"""

    # 1. 작업 유형 분석
    task_type = analyze_task_type(message)

    # 2. MacBook 상태 확인
    macbook_online = check_heartbeat("macbook")

    # 3. 라우팅 결정
    if task_type == "HEAVY":
        if macbook_online:
            return route_to_macbook(message)
        else:
            return queue_for_macbook(message)
            # 응답: "작업을 큐에 추가했습니다. MacBook 복귀 시 처리됩니다."

    elif task_type == "LIGHT":
        return route_to_vm(message)

    elif task_type == "SCHEDULED":
        return schedule_on_vm(message)
```

### 작업 분류 기준

**HEAVY (→ MacBook)**:
- "분석해줘", "전략 수립", "브랜드 아이덴티티"
- 이미지 분석 요청
- 5-agent 협업 필요 작업
- 즉시 응답 필요 (사용자 대기 중)

**LIGHT (→ VM)**:
- "일정 등록", "로그 확인", "상태 체크"
- 단순 조회 작업
- 파일 정리

**SCHEDULED (→ VM cron)**:
- "매일 오전 9시 컨텐츠 아이디어 생성"
- "매주 일요일 시스템 진화"
- "30분마다 클립보드 백업"

---

## V. 무료 티어 비용 분석

### VM 24/7 운영 시

| 항목 | 무료 티어 한도 | 사용량 | 비용 |
|------|---------------|--------|------|
| e2-micro VM | 730시간/월 | 730시간/월 (100%) | $0 |
| Cloud Run | 200만 요청 | 3,000 요청 | $0 |
| Cloud Storage | 5GB | 500MB | $0 |
| Network Egress | 1GB | 400MB | $0 |
| **총합** | - | - | **$0/월** |

**결론**: ✅ VM을 24/7 돌려도 무료 티어 범위 내

---

## VI. RAM 1GB 최적화

### VM에서 실행되는 작업의 메모리 프로파일

```python
# 스케줄 작업 메모리 사용량 추정
┌────────────────────────┬──────────┐
│ 작업                    │ 메모리    │
├────────────────────────┼──────────┤
│ Python runtime         │ 50MB     │
│ AI Engine (순차 1개)    │ 200MB    │
│ Memory Manager         │ 100MB    │
│ Google Drive API       │ 50MB     │
│ Gardener (1 agent)     │ 150MB    │
│ OS + Container         │ 200MB    │
├────────────────────────┼──────────┤
│ **총합**               │ 750MB    │
│ **여유**               │ 250MB    │
└────────────────────────┴──────────┘
```

**최적화 전략**:
1. **순차 처리**: 한 번에 1개 agent만 실행
2. **Swap 활성화**: 30GB 디스크 중 2GB swap
3. **메모리 정리**: 작업 완료 후 `gc.collect()`

---

## VII. 구현 로드맵

### Phase 1: VM 인스턴스 생성

```bash
gcloud compute instances create 97layeros-automation \
  --zone=us-west1-b \
  --machine-type=e2-micro \
  --boot-disk-size=30GB \
  --boot-disk-type=pd-standard \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --tags=97layer-automation \
  --metadata=enable-oslogin=true
```

### Phase 2: VM 환경 설정

```bash
# SSH 접속
gcloud compute ssh 97layeros-automation --zone=us-west1-b

# 필수 패키지 설치
sudo apt update
sudo apt install -y python3.10 python3-pip podman

# 프로젝트 clone
git clone https://github.com/97layer/97layerOS.git
cd 97layerOS

# 환경변수 설정
cat > .env << EOF
PROCESSING_MODE=sequential
ENABLE_SWAP=true
LOCATION=GCP_VM
EOF

# 의존성 설치
pip3 install -r requirements.txt
```

### Phase 3: Cron Jobs 설정

```bash
# VM crontab 설정
crontab -e

# 추가:
0 9 * * * /usr/bin/python3 /home/user/97layerOS/execution/vm/scheduled_content.py
0 0 * * 0 /usr/bin/python3 /home/user/97layerOS/execution/vm/weekly_evolution.py
*/30 * * * * /usr/bin/python3 /home/user/97layerOS/execution/system/clipboard_sentinel.py
0 0 * * * /usr/bin/python3 /home/user/97layerOS/execution/system/snapshot_daemon.py
*/5 * * * * /usr/bin/python3 /home/user/97layerOS/execution/vm/sync_monitor.py
```

### Phase 4: Cloud Run 라우터 업데이트

```python
# execution/cloud_run/main.py 수정
# route_task() 함수 추가 (위 IV. 참조)
```

### Phase 5: Heartbeat 시스템

```python
# MacBook에서 실행 (기존 hybrid_sync.py 활용)
# VM에서 실행 (sync_monitor.py)
```

---

## VIII. 예상 시나리오

### 시나리오 1: 평일 낮 (MacBook 온라인)
```
09:00 - VM: 컨텐츠 아이디어 3개 생성 → Drive 저장
10:30 - 사용자: "이 중 첫 번째 아이디어 상세 기획해줘"
10:30 - Cloud Run → MacBook (HEAVY)
10:30 - MacBook: 5-agent 병렬 처리 (11초)
10:31 - 응답 완료 ✅
```

### 시나리오 2: 외출 중 (MacBook 오프라인)
```
09:00 - VM: 컨텐츠 아이디어 생성 (정상 작동)
12:00 - 사용자: "브랜드 전략 분석해줘"
12:00 - Cloud Run → MacBook 오프라인 확인
12:00 - 응답: "작업을 큐에 추가했습니다. MacBook 복귀 시 처리됩니다."
12:30 - VM: 클립보드 백업 (정상 작동)
18:00 - MacBook 온라인 복귀
18:00 - 큐 작업 자동 처리 → 결과 Drive 저장 → Telegram 알림
```

### 시나리오 3: 주말 자정 (MacBook 꺼짐)
```
00:00 - VM: Gardener 진화 사이클 실행
00:15 - VM: 스냅샷 백업 생성
00:30 - VM: 로그 분석 리포트 생성
01:00 - 모든 자동화 완료 → Drive 저장
```

---

## IX. 장점

1. **MacBook 자유**: 꺼도 자동화는 계속 작동
2. **무료 티어**: VM 24/7 돌려도 $0/월
3. **역할 명확**: 무거운 vs 가벼운 작업 분리
4. **RAM 효율**: VM 1GB로 충분 (순차 처리)
5. **응답 속도**: MacBook 온라인 시 11초 유지
6. **자동화 보장**: 컨텐츠 생성, 진화, 백업 중단 없음

---

## X. 다음 단계

- [ ] VM 인스턴스 생성
- [ ] Cron jobs 스크립트 작성 (`execution/vm/` 폴더)
- [ ] Cloud Run 라우터 로직 구현
- [ ] Heartbeat 시스템 테스트
- [ ] 1주일 테스트 운영

**최종 결론**: ✅ **무료 티어 내에서 완전한 자동화 + 고성능 AI 처리 동시 달성**
