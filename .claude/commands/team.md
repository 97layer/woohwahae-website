---
description: 팀 모드 — 리더 기억 보존 + 팀원(워커) 자동 교체 오케스트레이션
---

# /team — Team Mode Orchestration

리더(메인 세션)가 기억을 유지하면서 워커(서브에이전트)를 순차/병렬로 자동 교체하는 워크플로우.

---

## 핵심 원칙

```
리더 = 메인 Claude Code 세션
    - state.md + team_queue.json 소유
    - 모든 태스크 결과를 수집해 메모리 갱신
    - 워커에게 최소 컨텍스트만 전달 (토큰 절약)

워커 = Task 서브에이전트
    - 브리핑만 받아 단일 태스크 실행
    - 완료 후 result_summary + learnings 반환
    - 자동 종료 (재사용 없음)
```

---

## 시작 전 체크리스트

```bash
# 1. 큐 상태 확인
cat knowledge/system/team_queue.json

# 2. 다음 태스크 브리핑 미리보기
python3 core/scripts/team_briefing.py next
```

---

## 팀 큐 설정 방법

`knowledge/system/team_queue.json` 수정:

```json
{
  "session_id": "team-[날짜]-[주제]",
  "leader_agent": "claude-sonnet-[날짜]",
  "tasks": [
    {
      "id": "T001",
      "status": "pending",
      "priority": 1,
      "type": "code|design|content|infra",
      "title": "태스크 제목",
      "objective": "달성해야 할 구체적 결과 (1-2문장)",
      "context_keys": ["infra", "pipeline"],
      "files_to_read": ["core/system/pipeline_orchestrator.py"],
      "constraints": ["state.md 수정 금지", "VM 직접 접근 금지"],
      "output_format": "수정된 파일 + 변경 요약"
    }
  ]
}
```

**context_keys 옵션**: `infra` / `pipeline` / `web` / `content` / `design` / `agent`

---

## 오케스트레이션 루프

리더가 아래 패턴을 반복한다:

### Step 1 — 브리핑 생성
```bash
python3 core/scripts/team_briefing.py next
```
→ 다음 pending 태스크의 최소 컨텍스트 브리핑 출력

### Step 2 — 워커 스폰
Task 도구를 사용해 subagent_type=Bash 또는 general-purpose로 스폰:
- prompt = 브리핑 내용 그대로 전달
- 리더 컨텍스트 **복붙 금지** — 브리핑만 전달
- 워커에게 `result_summary`와 `learnings` 반드시 포함하도록 지시

### Step 3 — 결과 수집
워커가 반환한 결과에서 추출:
- `result_summary`: 한 줄 결과
- `learnings`: 리더에게 유용한 인사이트
- `changed_files`: 수정된 파일 목록

### Step 4 — 리더 메모리 갱신
```bash
python3 core/scripts/team_briefing.py complete T001 "결과 요약" "인사이트1" "인사이트2"
```
→ team_queue.json 업데이트 + 체크포인트 기록

### Step 5 — state.md 반영
워커 learnings 중 시스템 전체에 영향 있는 것만 state.md에 추가.
→ 다음 워커가 state.md 읽으면 갱신된 정보 자동 획득

### Step 6 — 다음 워커로 이동
Step 1로 돌아감. `ALL_DONE` 출력되면 종료.

---

## 병렬 실행 (독립 태스크)

의존성 없는 태스크는 동시에 스폰:
```
priority: 1 → T001, T002 동시 실행
priority: 2 → T001/T002 완료 후 T003 실행
```

Task 도구 두 번 동시 호출하면 됨.

---

## 워커 브리핑 예시 (자동 생성됨)

```
# 워커 브리핑 — T001: 파이프라인 CE→AD 체인 수정

## 목표
pipeline_orchestrator.py에서 CE 완료 후 AD 자동 호출되도록 수정.

## 관련 컨텍스트 (필요 부분만)
[state.md에서 파이프라인 관련 섹션만 추출]

## 읽어야 할 파일
- core/system/pipeline_orchestrator.py

## 제약 조건
- state.md 직접 수정 금지
- VM 배포 금지 (리더가 처리)

## 출력 형식
수정된 파일 + 변경 내용 요약 + learnings
```

---

## 완료 후

```bash
# 핸드오프 (전체 팀 세션 종료 시)
/handoff
```

---

## 빠른 시작 예시

```bash
# 1. 큐에 태스크 추가 (team_queue.json 편집)
# 2. 시작
python3 core/scripts/team_briefing.py next
# 3. 출력된 브리핑으로 워커 스폰 (Task 도구)
# 4. 결과 받으면
python3 core/scripts/team_briefing.py complete T001 "AD 체인 연결 완료" "CE task 완료 시 task_type=ad로 재큐 필요"
# 5. 반복
python3 core/scripts/team_briefing.py next
```
